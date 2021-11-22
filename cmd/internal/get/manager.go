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

type Manager struct {
	ctx      context.Context
	cancel   context.CancelFunc
	conf     *metadata.Configure
	status   metadata.Status
	m        sync.Mutex
	view     *view.View
	workers  int
	ch       chan *db.Task
	ready    []*Worker
	workerID int64
	wait     sync.WaitGroup

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
		ch:     make(chan *db.Task),
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
		if e := m.produce(); e != nil {
			m.ExitWithError(e)
		}
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
	m.workerID++
	w := NewWorker(m.workerID, m)
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
func (m *Manager) GetChannel() <-chan *db.Task {
	return m.ch
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
	m.updateWorkerStatus()
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
	m.updateWorkerStatus()
}
func (m *Manager) produce() (e error) {
	modified, size, e := m.conf.GetMetadata(m.ctx)
	if e != nil {
		return
	}
	m.statusSize = utils.Size(size)
	var block int64 = int64(m.conf.Block)
	steps := (size + block - 1) / block
	m.statusSteps = steps
	log.Infof(`Metadata: size=%s steps=%v modified=%s`, m.statusSize, steps, modified)
	m.postStatus(false)

	d, e := db.OpenDB(m.conf.Output)
	if e != nil {
		return
	}
	log.Info(`open db: `, d.Filename)
	e = d.Load(m.statusSize, m.conf.Block, modified)
	if e != nil {
		return
	}
	m.status = metadata.StatusDownload
	m.postStatus(false)

	var (
		offset, num, id int64
	)
	for offset < size {
		id++
		if offset+block > size {
			num = size - offset
		} else {
			num = block
		}
		select {
		case m.ch <- &db.Task{
			ID:     id,
			Offset: utils.Size(offset),
			Num:    utils.Size(num),
		}:
		case <-m.ctx.Done():
			return
		}
		offset += num
	}
	return
}
func (m *Manager) ExitWithError(e error) {
	m.m.Lock()
	if m.status < metadata.StatusError {
		m.status = metadata.StatusError
		m.view.Update(func(g *gocui.Gui) error {
			return e
		})
	}
	m.m.Unlock()
}
