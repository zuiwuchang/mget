package get

import (
	"net/http"
	"strings"

	"github.com/zuiwuchang/mget/cmd/internal/get/worker"
	"github.com/zuiwuchang/mget/utils"
)

type Worker struct {
	*worker.Worker
	Status string
}

func NewWorker(id int64, rely worker.Rely) *Worker {
	return &Worker{
		Worker: worker.New(id, rely),
	}
}
func (m *Manager) DeleteWorker(worker *worker.Worker) {
	m.m.Lock()
	for i, w := range m.ready {
		if w.Worker == worker {
			m.ready = append(m.ready[:i], m.ready[i+1:]...)
			break
		}
	}
	m.postStatus(true)
	m.updateWorkerStatus()
	m.m.Unlock()
}
func (m *Manager) WorkerStatus(worker *worker.Worker, status string) {
	m.m.Lock()
	for _, w := range m.ready {
		if w.Worker == worker {
			w.Status = status
			m.updateWorkerStatus()
			break
		}
	}
	m.m.Unlock()
}
func (m *Manager) updateWorkerStatus() {
	strs := make([]string, len(m.ready))
	for i, w := range m.ready {
		strs[i] = w.Status
	}
	m.view.SetWorker(strings.Join(strs, "\n"))
}
func (m *Manager) Block() utils.Size {
	return m.conf.Block
}
func (m *Manager) WriteStatus(n int64, net bool) {
	m.m.Lock()
	m.statusDownload += utils.Size(n)
	if net {
		m.statistics.Push(n)
	}
	m.postStatus(true)
	m.m.Unlock()
}
func (m *Manager) GetRequest() (req *http.Request, e error) {
	return m.conf.NewRequestWithContext(m.ctx, http.MethodGet, m.conf.URL, nil)
}
func (m *Manager) Do(req *http.Request) (resp *http.Response, e error) {
	return m.conf.Do(req)
}
func (m *Manager) Finish() <-chan struct{} {
	return m.finish
}
