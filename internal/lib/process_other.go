//go:build !unix && !windows

package lib

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/ant-src/ant/internal/interp"
)

// terminate is the signal Cmd.stop sends first; a platform without processes, such as js/wasm,
// never reaches it, since starting a command fails.
var terminate os.Signal = os.Kill

// brokenPipe is nil: there is no signal for a writer whose reader is gone.
var brokenPipe os.Signal

func signalNamed(name string) (os.Signal, error) {
	return nil, fmt.Errorf("the signal %s is not supported on this platform", name)
}

func setGroup(cmd *exec.Cmd, leader int) {}

func signalGroup(ps []*os.Process, sig os.Signal) error { return signalEach(ps, sig) }

func exitStatus(ps *os.ProcessState) int { return ps.ExitCode() }

func endedBy(ps *os.ProcessState) os.Signal { return nil }

func isTerminal(f *os.File) bool { return false }

func lookPath(in *interp.Instance, program string) (string, error) {
	return "", errors.New("running programs is not supported on this platform")
}
