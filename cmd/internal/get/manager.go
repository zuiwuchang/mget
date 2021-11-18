package get

import "context"

type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc
	conf   *Configure
}

func NewManager(ctx context.Context, conf *Configure) *Manager {
	ctx, cancel := context.WithCancel(ctx)
	return &Manager{
		ctx:    ctx,
		cancel: cancel,
		conf:   conf,
	}
}
func (m *Manager) Serve() (e error) {
	gui := NewGUI(m.conf)
	e = gui.Serve()
	m.cancel()
	return
}
