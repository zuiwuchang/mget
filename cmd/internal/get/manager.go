package get

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/zuiwuchang/mget/cmd/internal/db"
	"github.com/zuiwuchang/mget/cmd/internal/log"
	"github.com/zuiwuchang/mget/cmd/internal/metadata"
	"github.com/zuiwuchang/mget/cmd/internal/view"
	"github.com/zuiwuchang/mget/utils"
)

type Task struct {
	ID     int
	Offset int64
	Num    int64
}
type Manager struct {
	ctx     context.Context
	cancel  context.CancelFunc
	conf    *metadata.Configure
	status  metadata.Status
	m       sync.Mutex
	view    *view.View
	workers int
	ch      chan *Task
	ready   []*Worker
	wait    sync.WaitGroup

	statusSize     utils.Size
	statusDownload utils.Size
	statusSteps    int64
}

func NewManager(ctx context.Context, conf *metadata.Configure) *Manager {
	ctx, cancel := context.WithCancel(ctx)
	return &Manager{
		ctx:    ctx,
		cancel: cancel,
		conf:   conf,
		ch:     make(chan *Task),
	}
}
func (m *Manager) ConfigureView() string {
	return strings.TrimRight(m.conf.String(), "\n")
}
func (m *Manager) Serve() (e error) {
	v := view.New(m)
	if e != nil {
		return
	}
	m.view = v

	e = v.Init()
	if e == nil {
		defer v.Close()
		e = m.init()
		if e == nil {
			m.postStatus(true)
			e = v.MainLoop()
		}
	}
	m.cancel()
	if m.status == metadata.StatusSuccess {
		e = nil
	}
	m.wait.Wait()
	return
}
func (m *Manager) init() (e error) {
	m.status = metadata.StatusInit
	log.Info(`Status: `, m.status)
	m.workers = m.conf.Worker
	for i := 0; i < m.workers; i++ {
		m.createWorker()
	}

	m.wait.Add(1)
	go func() {
		defer m.wait.Done()
		m.produce()
	}()
	return
}
func (m *Manager) postStatus(safe bool) {
	if m.view == nil {
		return
	}
	if !safe {
		m.m.Lock()
		defer m.m.Unlock()
	}

	md := ``
	if m.statusSteps != 0 {
		md += fmt.Sprintf(` steps: %v`, m.statusSteps)
	}
	if m.statusSize != 0 {
		md += fmt.Sprintf(` download: %s/%s`, m.statusDownload, m.statusSize)
	}

	body := fmt.Sprintf(`status: %s worker: %v/%v%s`, m.status, m.workers, len(m.ready), md)
	m.view.SetStatus(body)
}
func (m *Manager) createWorker() {
	w := NewWorker(m)
	m.wait.Add(1)
	m.ready = append(m.ready, w)
	go func() {
		defer m.wait.Done()
		w.Serve()
	}()
}
func (m *Manager) Context() context.Context {
	return m.ctx
}
func (m *Manager) GetChannel() <-chan *Task {
	return m.ch
}
func (m *Manager) deleteWorker(worker *Worker) {
	m.m.Lock()
	for i, w := range m.ready {
		if w == worker {
			m.ready = append(m.ready[:i], m.ready[i+1:]...)
			break
		}
	}
	m.postStatus(true)
	m.m.Unlock()
}
func (m *Manager) Increase() {
	m.m.Lock()
	defer m.m.Unlock()
	if m.status > metadata.StatusError {
		return
	} else if m.workers == metadata.MaxWorkers {
		return
	}

	m.workers++
	m.createWorker()
	m.postStatus(true)
}
func (m *Manager) Reduce() {
	m.m.Lock()
	defer m.m.Unlock()
	if m.status > metadata.StatusError {
		return
	} else if m.workers == 1 {
		return
	}

	m.workers--
	go func() {
		select {
		case m.ch <- nil:
		case <-m.ctx.Done():
		}
	}()
	m.postStatus(true)
}
func (m *Manager) produce() {
	modified, size, e := m.conf.GetMetadata(m.ctx)
	if e != nil {
		m.exitWithError(e)
		return
	}
	m.statusSize = utils.Size(size)
	var block int64 = int64(m.conf.Block)
	m.statusSteps = (size + block - 1) / block
	log.Infof(`Metadata: size=%s steps=%v modified=%s`, m.statusSize, m.statusSteps, modified)
	m.postStatus(false)

	db, e := db.OpenDB(m.conf.Output)
	if e != nil {
		m.exitWithError(e)
		return
	}
	log.Info(`open db: `, db.Filename)
}
func (m *Manager) exitWithError(e error) {
	m.m.Lock()
	if m.status < metadata.StatusError {
		m.status = metadata.StatusError
		m.view.Update(func(g *gocui.Gui) error {
			return e
		})
	}
	m.m.Unlock()
}
