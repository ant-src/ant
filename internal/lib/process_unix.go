//go:build unix

package lib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ant-src/ant/internal/interp"
)

// terminate is the signal Os.stop sends first.
var terminate os.Signal = syscall.SIGTERM

// signals are the names Os.kill takes.
var signals = map[string]syscall.Signal{
	"SIGTERM": syscall.SIGTERM, "SIGKILL": syscall.SIGKILL, "SIGINT": syscall.SIGINT,
	"SIGHUP": syscall.SIGHUP, "SIGQUIT": syscall.SIGQUIT, "SIGUSR1": syscall.SIGUSR1,
	"SIGUSR2": syscall.SIGUSR2,
}

func signalNamed(name string) (os.Signal, error) {
	if sig, ok := signals[name]; ok {
		return sig, nil
	}
	return nil, fmt.Errorf("unknown signal %s", name)
}

// setGroup makes the child the leader of a process group of its own, outside the terminal's
// foreground group.
func setGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// signalGroup sends sig once to the process group p leads, the leader included, while the
// leader runs. Signal 0 asks Go whether it does, which knows the process and not just its id,
// so a reused id is never signalled.
func signalGroup(p *os.Process, sig os.Signal) error {
	if err := p.Signal(syscall.Signal(0)); err != nil {
		return err
	}
	if err := syscall.Kill(-p.Pid, sig.(syscall.Signal)); err != nil && err != syscall.ESRCH {
		return err
	}
	return nil
}

// exitStatus is a child's status as a shell reports it: 128 + n for a child signal n ended.
func exitStatus(ps *os.ProcessState) int {
	if ws, ok := ps.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return ps.ExitCode()
}

// endedBy is the signal that ended a child, or nil.
func endedBy(ps *os.ProcessState) os.Signal {
	if ps == nil {
		return nil
	}
	if ws, ok := ps.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return ws.Signal()
	}
	return nil
}

// lookPath finds program as Go's exec.LookPath does, in the run's own PATH rather than the
// process's: a name with a slash is taken as it is, and a PATH entry that is not absolute, "."
// included, finds nothing.
func lookPath(in *interp.Instance, program string) (string, error) {
	if strings.Contains(program, "/") {
		if executable(program) {
			return program, nil
		}
		return "", &exec.Error{Name: program, Err: os.ErrNotExist}
	}
	path, _ := in.LookupEnv("PATH")
	for _, dir := range filepath.SplitList(path) {
		if !filepath.IsAbs(dir) {
			continue
		}
		if p := filepath.Join(dir, program); executable(p) {
			return p, nil
		}
	}
	return "", &exec.Error{Name: program, Err: exec.ErrNotFound}
}

func executable(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Mode()&0o111 != 0
}
