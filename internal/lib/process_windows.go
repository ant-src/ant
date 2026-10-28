//go:build windows

package lib

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/ant-src/ant/internal/interp"
)

// terminate is the signal Os.stop sends first; Windows ends a process one way.
var terminate os.Signal = os.Kill

// signalNamed maps the names Os.kill takes to the one way Windows has: SIGTERM and SIGKILL both
// terminate the process, with status 1.
func signalNamed(name string) (os.Signal, error) {
	switch name {
	case "SIGTERM", "SIGKILL":
		return os.Kill, nil
	case "SIGINT", "SIGHUP", "SIGQUIT", "SIGUSR1", "SIGUSR2":
		return nil, fmt.Errorf("the signal %s is not supported on Windows", name)
	}
	return nil, fmt.Errorf("unknown signal %s", name)
}

// setGroup puts the child in a process group of its own, which Ctrl-C does not reach.
func setGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

// signalGroup signals the child alone: Windows has no signal for a process group.
func signalGroup(p *os.Process, sig os.Signal) error { return p.Signal(sig) }

func exitStatus(ps *os.ProcessState) int { return ps.ExitCode() }

func endedBy(ps *os.ProcessState) os.Signal { return nil }

func isTerminal(f *os.File) bool { return false }

// lookPath is Go's exec.LookPath, over the process's PATH and PATHEXT.
func lookPath(in *interp.Instance, program string) (string, error) { return exec.LookPath(program) }
