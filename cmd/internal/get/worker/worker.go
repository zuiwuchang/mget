package worker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/zuiwuchang/mget/cmd/internal/db"
	"github.com/zuiwuchang/mget/utils"
)

type Rely interface {
	Context() context.Context
	GetChannel() <-chan *db.Task
	DeleteWorker(*Worker)
	WorkerStatus(w *Worker, str string)
	ExitWithError(e error)
	WriteStatus(n int64, net bool)
	Block() utils.Size
	Finish() <-chan struct{}

	GetRequest() (req *http.Request, e error)
	Do(req *http.Request) (resp *http.Response, e error)
}
type Worker struct {
	ID         int64
	rely       Rely
	statistics *utils.Statistics
}

func New(id int64, rely Rely) *Worker {
	return &Worker{
		ID:         id,
		rely:       rely,
		statistics: utils.NewStatistics(time.Second * 5),
	}
}
func (w *Worker) Serve() {
	finished := false
	defer func() {
		if !finished {
			w.delete()
		}
	}()
	done := w.rely.Context().Done()
	ch := w.rely.GetChannel()
	finish := w.rely.Finish()
	w.postIDLE()
	for {
		select {
		case <-done:
			return
		case <-finish:
			finished = true
			return
		case t := <-ch:
			if t == nil {
				return
			} else {
				e := w.serve(t)
				if e != nil {
					w.rely.ExitWithError(e)
					return
				}
			}
		}
	}
}
func (w *Worker) delete() {
	w.rely.DeleteWorker(w)
}
func (w *Worker) postIDLE() {
	w.rely.WorkerStatus(w, fmt.Sprintf(`worker-%v: IDLE`, w.ID))
}
func (w *Worker) postStart(t *db.Task) {
	w.rely.WorkerStatus(w, fmt.Sprintf(`worker-%v: Start step: %v offset: %s download: 0b/%s`,
		w.ID,
		t.ID,
		t.Offset, t.Num,
	))
}
func (w *Worker) postStatus(status string, t *db.Task, download utils.Size) {
	var md string
	if download != 0 {
		speed := w.statistics.Speed()
		if speed != 0 {
			md += fmt.Sprintf(` [%s/s]`, utils.Size(speed))
			if download < t.Num {
				duration := time.Second * time.Duration(t.Num-download) / time.Duration(speed)
				md += fmt.Sprintf(` %s ETA`, duration)
			}
		}
	}
	w.rely.WorkerStatus(w, fmt.Sprintf(`worker-%v: %s step: %v offset: %s download: %s/%s%s`,
		w.ID, status,
		t.ID,
		t.Offset, download, t.Num,
		md,
	))
}
func (w *Worker) serve(t *db.Task) (e error) {
	w.postStart(t)
	db := db.DefaultDB()
	num, e := db.GetSize(t.ID)
	if e != nil {
		return
	} else if num == t.Num {
		w.rely.WriteStatus(int64(num), false)
		w.postStatus(`Finish`, t, num)
		return
	} else if num > t.Num {
		e = fmt.Errorf(`%v db.num(%s) > num(%s)`, t.ID, num, t.Num)
		return
	} else if num < 0 {
		e = fmt.Errorf(`%v db.num(%s) < 0`, t.ID, num)
		return
	} else if num > 0 {
		w.rely.WriteStatus(int64(num), false)
	}
	f, e := os.OpenFile(db.Temp, os.O_WRONLY, 0666)
	if e != nil {
		return
	}
	_, e = f.Seek(int64(t.Offset+num), io.SeekStart)
	if e != nil {
		f.Close()
		return
	}
	e = w.downloadRange(t, &Writer{
		t:    t,
		w:    w,
		f:    f,
		db:   db,
		size: int64(num),
	}, num, w.rely.Block())
	f.Close()
	return
}
func (w *Worker) downloadRange(t *db.Task, writer io.Writer, num, block utils.Size) (e error) {
	w.postStatus(`Get`, t, num)
	req, e := w.rely.GetRequest()
	if e != nil {
		return
	}
	offset := int64(t.Offset + num)
	size := int64(t.Num - num)
	end := offset + size - 1
	req.Header.Set(`Range`, fmt.Sprintf(`bytes=%v-%v`, offset, end))
	resp, e := w.rely.Do(req)
	if e != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent {
		e = fmt.Errorf(`not StatusPartialContent: %v %v`, resp.StatusCode, resp.Status)
		return
	}
	n, e := io.Copy(writer, io.LimitReader(resp.Body, size))
	if e != nil {
		if e == io.EOF && n == size {
			e = nil
		}
		return
	}
	return
}

type Writer struct {
	t    *db.Task
	w    *Worker
	f    *os.File
	db   *db.DB
	size int64
}

func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.f.Write(p)
	if err != nil {
		w.setSize()
		return
	}
	if n != 0 {
		w.w.statistics.Push(int64(n))
		w.size += int64(n)
		err = w.setSize()
		if err == nil {
			w.w.rely.WriteStatus(int64(n), true)
			w.update()
		}
	}
	return
}
func (w *Writer) update() {
	size := utils.Size(w.size)
	if size == w.t.Num {
		w.w.postStatus(`Finish`, w.t, size)
	} else {
		w.w.postStatus(`Get`, w.t, size)
	}
}
func (w *Writer) setSize() (e error) {
	return w.db.SetSize(w.t.ID, w.size)
}
