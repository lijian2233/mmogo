package logger

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type asyncWriter struct {
	dir      string
	file     string
	fd       *os.File
	writer   *bufio.Writer
	bytePool *sync.Pool
	msgQueue chan *message
	timer    *time.Ticker
	getFile  func() string
	ctx      context.Context
	cancel   context.CancelFunc
	end      chan bool
}

type message struct {
	prefix   string
	format   string
	args     []interface{}
	ignoreLF bool
}

func newWriter(dir string, getFile func() string) (writer *asyncWriter, err error) {
	writer = &asyncWriter{
		dir:      dir,
		getFile:  getFile,
		bytePool: &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
		msgQueue: make(chan *message, 8192),
		timer:    time.NewTicker(time.Second),
		end:      make(chan bool, 1),
	}
	writer.ctx, writer.cancel = context.WithCancel(context.Background())

	if err = os.MkdirAll(writer.dir, 0755); err != nil {
		return
	}

	writer.refresh()
	writer.fd, err = os.OpenFile(writer.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return
	}

	writer.writer = bufio.NewWriter(writer.fd)

	go writer.flush()
	go writer.start()

	return
}

func (l *asyncWriter) refresh() bool {
	oldFile := l.file
	l.file = l.getFile()
	return l.file != oldFile
}

func (l *asyncWriter) start() {
	for msg := range l.msgQueue {
		if msg == nil {
			_ = l.writer.Flush()

			if l.refresh() {
				_ = l.fd.Close()
				l.fd, _ = os.OpenFile(l.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				l.writer.Reset(l.fd)
			}
		} else {
			_, _ = l.writer.Write(l.bytes(msg))
		}
	}

	l.end <- true
}

func (l *asyncWriter) bytes(msg *message) []byte {
	w := l.bytePool.Get().(*bytes.Buffer)

	defer func() {
		recover()
		w.Reset()
		l.bytePool.Put(w)
	}()

	if len(msg.prefix) > 0 {
		_, _ = fmt.Fprintf(w, msg.prefix)
	}

	if len(msg.format) == 0 {
		for i := 0; i < len(msg.args); i++ {
			if i > 0 {
				w.Write([]byte{' '})
			}

			_, _ = fmt.Fprint(w, msg.args[i])
		}
	} else {
		_, _ = fmt.Fprintf(w, msg.format, msg.args...)
	}

	if !msg.ignoreLF {
		_, _ = fmt.Fprintf(w, "\n")
	}

	b := make([]byte, w.Len())
	copy(b, w.Bytes())

	return b
}

func (l *asyncWriter) flush() {
	for range l.timer.C {
		l.msgQueue <- nil
	}
}

func (l *asyncWriter) write(msg *message) {
	select {
	case <-l.ctx.Done():
	default:
		l.msgQueue <- msg
	}
}

func (l *asyncWriter) close() {
	l.cancel()
	l.timer.Stop()
	l.msgQueue <- nil
	close(l.msgQueue)
	<-l.end
}
