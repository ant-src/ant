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

// TestCloseReportsFlush: a write that entered the buffer and failed when close flushed it is
// reported by close, the file is closed all the same, and a second close does nothing and reports
// nothing.
func TestCloseReportsFlush(t *testing.T) {
	c := &countingCloser{}
	h := newHandle(c, nil, failingWriter{})
	if got := value.Render(h.write("lost\n")); got != "Ok{ value: void }" {
		t.Fatalf("write = %s, want Ok", got)
	}
	if err := h.close(); err == nil || err.Error() != "no space left on device" {
		t.Fatalf("close = %v, want the flush error", err)
	}
	if c.n != 1 {
		t.Fatalf("closed %d times after a failed flush, want 1", c.n)
	}
	if err := h.close(); err != nil {
		t.Fatalf("second close = %v, want nil", err)
	}
	if c.n != 1 {
		t.Fatalf("closed %d times after a second close, want 1", c.n)
	}
	if got := value.Render(h.write("late\n")); got == "Ok{ value: void }" {
		t.Fatalf("write after close = %s, want Err", got)
	}
}
