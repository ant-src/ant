//go:build unix

package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

// quit is SIGQUIT, which Os.run holds back with SIGINT.
var quit os.Signal = syscall.SIGQUIT

// dieBy ends the command by sig, at its default disposition, so the parent sees a process the
// signal ended and a shell running it in a loop stops too (Cracauer's "wait and cooperative
// exit"). It returns only when the signal did not end the process.
func dieBy(sig os.Signal) {
	signal.Reset(sig)
	syscall.Kill(os.Getpid(), sig.(syscall.Signal))
	time.Sleep(time.Second)
}
