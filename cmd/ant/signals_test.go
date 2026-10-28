//go:build unix

package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestMain lets a test run the command itself in a child process, so a signal can end it.
func TestMain(m *testing.M) {
	if os.Getenv("ANT_TEST_COMMAND") == "1" {
		os.Args = append([]string{"ant"}, os.Args[len(os.Args)-1])
		main()
	}
	os.Exit(m.Run())
}

// TestSignals checks how the command ends: by the signal that interrupted the run, so a shell
// loop running it stops too; and, while Os.run has a child in the foreground, by a Ctrl-C only
// when the child dies of it. The terminal's Ctrl-C reaches the whole foreground process group,
// which the test does by signalling the group.
func TestSignals(t *testing.T) {
	cases := []struct {
		name   string
		src    string
		sig    syscall.Signal
		ended  syscall.Signal // the signal that ends the command, or 0 when it exits
		status int
		out    string
	}{
		{"interrupt", `Time.sleep(Time.duration(seconds: 10))` + "\n", syscall.SIGINT, syscall.SIGINT, 0, ""},
		{"terminate", `Time.sleep(Time.duration(seconds: 10))` + "\n", syscall.SIGTERM, syscall.SIGTERM, 0, ""},
		{"child dies of Ctrl-C", `println(Os.run("sleep", ["10"]))` + "\nprintln(\"after\")\n", syscall.SIGINT, syscall.SIGINT, 0, ""},
		{"child handles Ctrl-C", `println(Os.run("sh", ["-c", "trap 'echo handled' INT; sleep 1; exit 0"]))` + "\nprintln(\"after\")\n",
			syscall.SIGINT, 0, 0, "handled\nOk{ value: 0 }\nafter\n"},
		{"program asked", "var stop = Os.interrupt()\n" + `println(Os.run("sleep", ["10"]))` + "\nprintln(Context.isDone(stop))\n",
			syscall.SIGINT, 0, 0, "Ok{ value: 130 }\ntrue\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			script := filepath.Join(t.TempDir(), "main.ant")
			if err := os.WriteFile(script, []byte(c.src), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^$", script)
			cmd.Env = append(os.Environ(), "ANT_TEST_COMMAND=1")
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			var out bytes.Buffer
			cmd.Stdout = &out
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			time.Sleep(300 * time.Millisecond)
			syscall.Kill(-cmd.Process.Pid, c.sig)
			err := cmd.Wait()
			var exit *exec.ExitError
			if err != nil && !errors.As(err, &exit) {
				t.Fatal(err)
			}
			ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
			switch {
			case c.ended != 0 && (!ws.Signaled() || ws.Signal() != c.ended):
				t.Errorf("the command should have ended by %v, got %v", c.ended, ws)
			case c.ended == 0 && (ws.Signaled() || ws.ExitStatus() != c.status):
				t.Errorf("the command should have exited with %d, got %v", c.status, ws)
			}
			if out.String() != c.out {
				t.Errorf("printed %q, want %q", out.String(), c.out)
			}
		})
	}
}
