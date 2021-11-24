package debug

import (
	"bufio"
	"io"
	"net"
	"sync"
	"time"
	"unsafe"
)

const (
	TCPAddr = `127.0.0.1:7000`
)

var defaultLog = &log{
	ch: make(chan []byte, 100),
}

func init() {
	go defaultLog.Serve()
}

type log struct {
	wc io.WriteCloser
	bw *bufio.Writer
	ch chan []byte
	m  sync.Mutex
}

func (l *log) WriteString(str string) (int, error) {
	b := *(*[]byte)(unsafe.Pointer(&str))
	return l.Write(b)
}
func (l *log) Write(b []byte) (int, error) {
	l.m.Lock()
	defer l.m.Unlock()
	for {
		select {
		case l.ch <- b:
			return len(b), nil
		default:
		}
		select {
		case <-l.ch:
		case l.ch <- b:
			return len(b), nil
		default:
		}
	}
}
func (l *log) Serve() {
	for {
		b := <-l.ch
		e := l.write(b)
		if e != nil {
			continue
		}
		ok := true
		for ok {
			select {
			case b = <-l.ch:
				e = l.write(b)
				if e != nil {
					ok = false
				}
			default:
				e = l.bw.Flush()
				if e != nil {
					l.wc.Close()
					l.wc = nil
				}
			}
		}
	}
}

func (l *log) write(b []byte) (e error) {
	w := l.getWriter()
	_, e = w.Write(b)
	if e != nil {
		l.wc.Close()
		l.wc = nil
	}
	return
}
func (l *log) getWriter() *bufio.Writer {
	if l.wc != nil {
		return l.bw
	}
	for {
		c, e := net.Dial(`tcp`, TCPAddr)
		if e != nil {
			time.Sleep(time.Second)
			continue
		}
		l.wc = c
		break
	}
	if l.bw == nil {
		l.bw = bufio.NewWriterSize(l.wc, 1024*32)
	} else {
		l.bw.Reset(l.wc)
	}
	return l.bw
}
