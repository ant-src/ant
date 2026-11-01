//go:build unix

package lib

import (
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/ant-src/ant/value"
)

// TestNoWaitOnFIFO: File.touch sets a FIFO's times and File.copy refuses one as its source, and
// neither opens it, which would wait for the other end (stdlib.md, File: Times, Copies).
func TestNoWaitOnFIFO(t *testing.T) {
	fifo := filepath.Join(t.TempDir(), "fifo")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Skip("no FIFOs here:", err)
	}
	done := make(chan [2]string, 1)
	go func() {
		touched := value.Render(touch(fifo, time.Unix(0, 0)))
		copied := value.Render(copyFile(fifo, fifo+".copy", true))
		done <- [2]string{touched, copied}
	}()
	select {
	case got := <-done:
		if got[0] != "Ok{ value: void }" {
			t.Errorf("touch = %s", got[0])
		}
		if !strings.Contains(got[1], "not a regular file") {
			t.Errorf("copy = %s, want Err for a FIFO", got[1])
		}
	case <-time.After(5 * time.Second):
		t.Fatal("touch or copy waited on a FIFO")
	}
}
