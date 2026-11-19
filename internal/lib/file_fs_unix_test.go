//go:build unix

package lib

import (
	"os"
	"path/filepath"
	"strconv"
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

// TestChownAndSync: File.chown to the owner's own ids changes nothing and is Ok, of a link itself
// with follow false; stat reports the ids; sync of a pipe is Err, as fsync is (stdlib.md, File).
func TestChownAndSync(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f")
	link := filepath.Join(dir, "l")
	if got := value.Render(writeFile(file, []byte("x"), true)); got != "Ok{ value: void }" {
		t.Fatal(got)
	}
	if err := os.Symlink(file, link); err != nil {
		t.Fatal(err)
	}
	uid, gid := int64(os.Geteuid()), int64(os.Getegid())
	for _, follow := range []bool{true, false} {
		if got := value.Render(chown(link, uid, gid, follow)); got != "Ok{ value: void }" {
			t.Errorf("chown follow=%v = %s", follow, got)
		}
	}
	info := value.Render(statFile(link, false))
	if !strings.Contains(info, "uid: "+strconv.FormatInt(uid, 10)) || !strings.Contains(info, "gid: "+strconv.FormatInt(gid, 10)) {
		t.Errorf("stat = %s, want uid %d and gid %d", info, uid, gid)
	}
	if got := value.Render(chown(file, 1<<32, -1, true)); !strings.Contains(got, "is not -1 or in 0 ..= 4294967294") {
		t.Errorf("chown of a too large id = %s", got)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	if err := newHandle(w, nil, w).sync(); err == nil {
		t.Error("sync of a pipe is nil, want an error")
	}
}
