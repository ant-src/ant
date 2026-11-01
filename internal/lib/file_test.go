package lib

import (
	"errors"
	"testing"

	"github.com/ant-src/ant/value"
)

// failingWriter accepts nothing, as a full disk does once the buffer is written out.
type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) { return 0, errors.New("no space left on device") }

// countingCloser counts its closes.
type countingCloser struct{ n int }

func (c *countingCloser) Close() error { c.n++; return nil }

// TestWriteFailsAtOnce: a write to a handle that is not buffered reaches the writer before it
// returns, so its error is the write's, and the close has nothing left to report.
func TestWriteFailsAtOnce(t *testing.T) {
	c := &countingCloser{}
	h := newHandle(c, nil, failingWriter{})
	if got := value.Render(h.write("lost\n")); got == "Ok{ value: void }" {
		t.Fatalf("write = %s, want Err", got)
	}
	if err := h.close(); err != nil || c.n != 1 {
		t.Fatalf("close = %v after %d closes, want nil after 1", err, c.n)
	}
}

// TestBufferedKeepsError: a write that entered the buffer and failed when it was written out is
// reported by every later write and flush and by the first close; the handle under is closed all
// the same, and a second close does nothing and reports nothing.
func TestBufferedKeepsError(t *testing.T) {
	c := &countingCloser{}
	h := buffered(newHandle(c, nil, failingWriter{}), 4)
	if got := value.Render(h.write("ab")); got != "Ok{ value: void }" {
		t.Fatalf("write = %s, want Ok", got)
	}
	if err := h.flush(); err == nil || err.Error() != "no space left on device" {
		t.Fatalf("flush = %v, want the write's error", err)
	}
	if got := value.Render(h.write("c")); got == "Ok{ value: void }" {
		t.Fatalf("write after a failed flush = %s, want the kept Err", got)
	}
	if err := h.close(); err == nil || err.Error() != "no space left on device" {
		t.Fatalf("close = %v, want the kept error", err)
	}
	if c.n != 1 {
		t.Fatalf("closed %d times after a failed flush, want 1", c.n)
	}
	if err := h.close(); err != nil || c.n != 1 {
		t.Fatalf("second close = %v after %d closes, want nil after 1", err, c.n)
	}
	if got := value.Render(h.write("late\n")); got == "Ok{ value: void }" {
		t.Fatalf("write after close = %s, want Err", got)
	}
}

// blockingWriter waits in Write until its handle is closed, as a write to a peer that stopped
// reading does.
type blockingWriter struct{ entered, closed chan struct{} }

func (w blockingWriter) Write(p []byte) (int, error) {
	close(w.entered)
	<-w.closed
	return 0, errors.New("use of closed connection")
}

type closerFunc func() error

func (f closerFunc) Close() error { return f() }

// TestCloseEndsWaitingBufferedWrite: a close while a flush of the buffered handle waits on
// another goroutine does not wait behind it: it closes the handle under at once, which ends the
// flush with an Err.
func TestCloseEndsWaitingBufferedWrite(t *testing.T) {
	w := blockingWriter{entered: make(chan struct{}), closed: make(chan struct{})}
	h := buffered(newHandle(closerFunc(func() error { close(w.closed); return nil }), nil, w), 16)
	h.write("request")
	flushed := make(chan error)
	go func() { flushed <- h.flush() }()
	<-w.entered
	if err := h.close(); err != nil {
		t.Fatalf("close = %v, want nil", err)
	}
	if err := <-flushed; err == nil {
		t.Fatal("the waiting flush returned nil, want an error")
	}
}
