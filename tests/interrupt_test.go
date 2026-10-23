package tests

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ant-src/ant"
)

// runWithInterrupts runs src with interrupts sent at the given delays, and returns what the
// program printed and how the run ended.
func runWithInterrupts(t *testing.T, src string, delays ...time.Duration) (string, error) {
	t.Helper()
	prog, err := ant.Load("main.ant", ant.WithSource("main.ant", src), ant.WithFS(fstest.MapFS{}))
	if err != nil {
		t.Fatal(err)
	}
	interrupts := make(chan struct{})
	go func() {
		for _, d := range delays {
			time.Sleep(d)
			interrupts <- struct{}{}
		}
	}()
	var out bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = prog.Run(ctx, ant.WithStdout(&out), ant.WithInterrupts(interrupts))
	return out.String(), err
}

// TestInterrupts checks the rule of Os.interrupt: a program that asked gets the first
// interrupt as its context's reason and goes on; a program that never asked is ended by the
// first; a program that asked and ignores it is ended by the second.
func TestInterrupts(t *testing.T) {
	t.Run("delivered", func(t *testing.T) {
		out, err := runWithInterrupts(t, `
func main(): void {
  var stop = Os.interrupt()
  Context.wait(stop)
  print(Context.reason(stop) ?? "none")
  print("drained")
}
`, 50*time.Millisecond)
		if err != nil {
			t.Fatalf("the run ended with %v; the first interrupt should have gone to the program", err)
		}
		if strings.TrimSpace(out) != "interrupt\ndrained" {
			t.Fatalf("printed %q", out)
		}
	})
	t.Run("never asked", func(t *testing.T) {
		out, err := runWithInterrupts(t, `
func main(): void {
  Time.sleep(5000)
  print("slept")
}
`, 50*time.Millisecond)
		if err == nil || strings.Contains(out, "slept") {
			t.Fatalf("the first interrupt should have ended the run: err %v, printed %q", err, out)
		}
		if !errors.Is(err, ant.ErrInterrupted) || ant.ExitStatus(err) != 130 {
			t.Fatalf("an interrupt that ends the run is ErrInterrupted with status 130, got %v (%d)", err, ant.ExitStatus(err))
		}
	})
	t.Run("ignored", func(t *testing.T) {
		out, err := runWithInterrupts(t, `
func main(): void {
  var stop = Os.interrupt()
  Time.sleep(5000)
  print(Context.isDone(stop))
}
`, 50*time.Millisecond, 100*time.Millisecond)
		if err == nil || strings.Contains(out, "true") {
			t.Fatalf("the second interrupt should have ended the run: err %v, printed %q", err, out)
		}
	})
}
