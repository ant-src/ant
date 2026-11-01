package lib

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ant-src/ant/value"
)

// The lock tests run on every platform CI has (stdlib.md, File: Locks), Windows included, where
// the lock is LockFileEx's rather than flock's.

// lockHandle is a lock's handle, or nil when lockFile gave null.
func lockHandle(t *testing.T, v value.Value) *Handle {
	t.Helper()
	r, isRecord := v.(*value.Record)
	if !isRecord || r.Tag != value.OkTag {
		t.Fatalf("lock = %s, want Ok", value.Render(v))
	}
	inner, _ := r.Field("value")
	h, _ := inner.(*Handle)
	return h
}

// waitLock waits up to 10 seconds for the lock: a child forked by another test can hold a copy of
// a closed handle until it execs.
func waitLock(t *testing.T, path string) *Handle {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	h := lockHandle(t, lockFile(ctx, path, true))
	if h == nil {
		t.Fatal("lock gave null")
	}
	return h
}

// TestLockInProcess: a second handle conflicts with the first in the same process; lock waits
// until its context ends; closing the holder releases the lock.
func TestLockInProcess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.lock")
	first := lockHandle(t, lockFile(context.Background(), path, false))
	if first == nil {
		t.Fatal("tryLock of a free lock gave null")
	}
	if h := lockHandle(t, lockFile(context.Background(), path, false)); h != nil {
		t.Fatal("tryLock of a held lock gave a handle")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	// A Go context's reason reads "cancelled"; tests/run/file_lock.ant checks an Ant timeout's.
	if got := value.Render(lockFile(ctx, path, true)); !strings.HasPrefix(got, "Err{") {
		t.Fatalf("lock of a held lock = %s, want Err when the context ends", got)
	}
	if err := first.close(); err != nil {
		t.Fatal(err)
	}
	waitLock(t, path).close()
}

// TestLockHelper is the other process of TestLockAcrossProcesses: it takes the lock named by
// ANT_LOCK_HELPER, says so, and holds it until its standard input closes.
func TestLockHelper(t *testing.T) {
	path := os.Getenv("ANT_LOCK_HELPER")
	if path == "" {
		t.Skip("only as TestLockAcrossProcesses's helper")
	}
	h := lockHandle(t, lockFile(context.Background(), path, false))
	if h == nil {
		os.Stdout.WriteString("held\n")
		return
	}
	os.Stdout.WriteString("locked\n")
	io.Copy(io.Discard, os.Stdin)
	h.close()
}

// startHolder starts a process that holds the lock at path, and waits until it does.
func startHolder(t *testing.T, path string) (*exec.Cmd, io.WriteCloser) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestLockHelper$")
	cmd.Env = append(os.Environ(), "ANT_LOCK_HELPER="+path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil || line != "locked\n" {
		cmd.Process.Kill()
		cmd.Wait()
		t.Fatalf("helper said %q, %v; want locked", line, err)
	}
	return cmd, stdin
}

// TestLockAcrossProcesses: a lock another process holds is held here too; it is released when
// that process closes it and when that process is killed; and it is advisory on Unix but keeps
// other processes from reading the file on Windows.
func TestLockAcrossProcesses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.lock")
	if err := os.WriteFile(path, []byte("pid"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd, stdin := startHolder(t, path)
	if h := lockHandle(t, lockFile(context.Background(), path, false)); h != nil {
		t.Fatal("tryLock of a lock another process holds gave a handle")
	}
	_, err := os.ReadFile(path)
	if runtime.GOOS == "windows" && err == nil {
		t.Error("another process read a file locked on Windows, where the lock is mandatory")
	}
	if runtime.GOOS != "windows" && err != nil {
		t.Errorf("reading a file locked on Unix, where the lock is advisory: %v", err)
	}
	stdin.Close()
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	waitLock(t, path).close()

	cmd, _ = startHolder(t, path)
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	cmd.Wait()
	waitLock(t, path).close()
}
