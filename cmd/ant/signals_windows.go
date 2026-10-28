//go:build windows

package main

import "os"

// quit is nil: Windows has no SIGQUIT to hold back.
var quit os.Signal

// dieBy does nothing: a Windows process cannot end by a signal, and exits with 130 instead.
func dieBy(sig os.Signal) {}
