// Package batch coalesces what a run writes for a host to which each write is costly, such as a
// debug adapter sending an event per write or a page posted a message per write. The language
// writes each print at once (stdlib.md, File); batching is the host's choice.
package batch

import (
	"io"
	"sync"
	"time"
)

// Size is how many bytes a batch holds before it is sent without waiting for the delay.
const Size = 4096

// Writer collects writes to several streams and sends them in batches, in the order written: a
// write to another stream sends what the first one holds first. A batch is sent once it holds
// Size bytes, delay after its first write, and at Flush.
type Writer struct {
	mu    sync.Mutex
	send  func(stream string, p []byte)
	delay time.Duration
	cur   string
	buf   []byte
	timer *time.Timer
}

// New returns a Writer that hands each batch to send, under its lock, so send must not write to
// the Writer.
func New(delay time.Duration, send func(stream string, p []byte)) *Writer {
	return &Writer{send: send, delay: delay}
}

// Stream returns the writer of one stream.
func (w *Writer) Stream(name string) io.Writer { return stream{w, name} }

type stream struct {
	w    *Writer
	name string
}

func (s stream) Write(p []byte) (int, error) {
	s.w.write(s.name, p)
	return len(p), nil
}

func (w *Writer) write(name string, p []byte) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.buf) > 0 && w.cur != name {
		w.flushLocked()
	}
	w.cur = name
	w.buf = append(w.buf, p...)
	switch {
	case len(w.buf) >= Size:
		w.flushLocked()
	case w.timer == nil:
		w.timer = time.AfterFunc(w.delay, w.Flush)
	}
}

// Flush sends what is held now.
func (w *Writer) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.flushLocked()
}

func (w *Writer) flushLocked() {
	if w.timer != nil {
		w.timer.Stop()
		w.timer = nil
	}
	if len(w.buf) == 0 {
		return
	}
	p := w.buf
	w.buf = nil
	w.send(w.cur, p)
}
