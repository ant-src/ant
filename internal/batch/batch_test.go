package batch

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

type log struct {
	mu   sync.Mutex
	sent []string
}

func (l *log) send(stream string, p []byte) {
	l.mu.Lock()
	l.sent = append(l.sent, stream+":"+string(p))
	l.mu.Unlock()
}

func (l *log) get() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.sent...)
}

// TestOrder: writes to one stream coalesce, and a write to another sends the first stream's
// batch before it, so the order written is kept.
func TestOrder(t *testing.T) {
	l := &log{}
	w := New(time.Hour, l.send)
	io.WriteString(w.Stream("out"), "a")
	io.WriteString(w.Stream("out"), "b")
	io.WriteString(w.Stream("err"), "c")
	io.WriteString(w.Stream("out"), "d")
	w.Flush()
	if got := fmt.Sprint(l.get()); got != "[out:ab err:c out:d]" {
		t.Errorf("sent %s", got)
	}
}

// TestSizeAndDelay: a batch is sent once it holds Size bytes, and after the delay otherwise.
func TestSizeAndDelay(t *testing.T) {
	l := &log{}
	w := New(20*time.Millisecond, l.send)
	io.WriteString(w.Stream("out"), strings.Repeat("x", Size))
	if n := len(l.get()); n != 1 {
		t.Fatalf("%d batches after Size bytes, want 1", n)
	}
	io.WriteString(w.Stream("out"), "y")
	deadline := time.Now().Add(2 * time.Second)
	for len(l.get()) < 2 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if got := l.get(); len(got) != 2 || got[1] != "out:y" {
		t.Errorf("sent %q after the delay", got)
	}
}
