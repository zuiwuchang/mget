package get

import (
	"context"
	"fmt"
	"sync"

	"github.com/jroimartin/gocui"
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
	conf    *Configure
	status  Status
	m       sync.Mutex
	gui     *GUI
	workers int
	ch      chan *Task
	ready   []*Worker
	wait    sync.WaitGroup

	filterStatus *Filter
	statusSize   utils.Size
}

func NewManager(ctx context.Context, conf *Configure) *Manager {
	ctx, cancel := context.WithCancel(ctx)
	return &Manager{
		ctx:    ctx,
		cancel: cancel,
		conf:   conf,
		ch:     make(chan *Task),
	}
}
func (m *Manager) Serve() (e error) {
	gui := NewGUI(m.conf, m)
	if e != nil {
		return
	}
	m.gui = gui

	e = gui.Init()
	if gui.g != nil {
		defer gui.g.Close()
	}
	if e == nil {
		e = m.init()
		if e == nil {
			m.postStatus(true)
			e = gui.g.MainLoop()
		}
	}
	m.cancel()
	if m.status == StatusSuccess {
		e = nil
	}
	if m.filterStatus != nil {
		m.filterStatus.Close()
	}
	m.wait.Wait()
	return
}
func (m *Manager) init() (e error) {
	m.filterStatus = NewFilter(m.gui.g)
	m.wait.Add(1)
	go func() {
		defer m.wait.Done()
		m.filterStatus.Serve()
	}()

	m.status = StatusInit
	defaultLog.Push(`info`, `Status: %s`, m.status)
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
	if !safe {
		m.m.Lock()
		defer m.m.Unlock()
	}

	steps := ``

	body := fmt.Sprintf(`status: %s expect: %v ready: %v%s`, m.status, m.workers, len(m.ready), steps)
	m.filterStatus.Update(func(g *gocui.Gui) error {
		if m.gui != nil && m.gui.viewStatus != nil {
			m.gui.viewStatus.SetBody(body)
		}
		return nil
	})
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
	if m.status > StatusError {
		return
	} else if m.workers == MaxWorkers {
		return
	}

	m.workers++
	m.createWorker()
	m.postStatus(true)
}
func (m *Manager) Reduce() {
	m.m.Lock()
	defer m.m.Unlock()
	if m.status > StatusError {
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
	defaultLog.Push(`info`, `Metadata: size=%s modified=%s`, modified, m.statusSize)
}
func (m *Manager) exitWithError(e error) {
	m.m.Lock()
	if m.status < StatusError {
		m.status = StatusError
		m.gui.g.Update(func(g *gocui.Gui) error {
			return e
		})
	}
	m.m.Unlock()
}
