package get

import (
	"context"

	"github.com/zuiwuchang/mget/cmd/internal/log"
)

type WorkerRely interface {
	Context() context.Context
	GetChannel() <-chan *Task
	deleteWorker(*Worker)
}
type Worker struct {
	rely   WorkerRely
	ctx    context.Context
	cancel context.CancelFunc
}

func NewWorker(rely WorkerRely) *Worker {
	ctx, cancel := context.WithCancel(rely.Context())
	return &Worker{
		ctx:    ctx,
		cancel: cancel,
		rely:   rely,
	}
}
func (w *Worker) Serve() {
	defer w.delete()
	done := w.ctx.Done()
	ch := w.rely.GetChannel()
	for {
		select {
		case <-done:
			return
		case t := <-ch:
			if t == nil {
				return
			} else {
				w.serve(t)
			}
		}
	}
}
func (w *Worker) delete() {
	w.rely.deleteWorker(w)
}
func (w *Worker) serve(t *Task) {
	log.Info(t)
}
