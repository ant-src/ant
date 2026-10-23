package tests

import (
	"bytes"
	"context"
	"io"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ant-src/ant"
)

// runWithLateStdin runs src with standard input that delivers text after a delay and then
// ends, and returns what the program printed.
func runWithLateStdin(t *testing.T, src, text string, delay time.Duration) string {
	t.Helper()
	prog, err := ant.Load("main.ant", ant.WithSource("main.ant", src), ant.WithFS(fstest.MapFS{}))
	if err != nil {
		t.Fatal(err)
	}
	pr, pw := io.Pipe()
	go func() {
		time.Sleep(delay)
		io.WriteString(pw, text)
		pw.Close()
	}()
	var out bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := prog.Run(ctx, ant.WithStdin(pr), ant.WithStdout(&out)); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

// TestStdinReadResumes checks that a read of standard input ended by its context leaves the
// stream readable and loses nothing: the line that arrives later is taken whole by the next
// read, and a pipe that gave up leaves the line to the read that follows it.
func TestStdinReadResumes(t *testing.T) {
	t.Run("readLine", func(t *testing.T) {
		out := runWithLateStdin(t, `
func main(): void {
  print(File.readLine(Os.stdin(), Context.new(Time.duration(milliseconds: 50))))
  print(File.readLine(Os.stdin()))
  print(File.readLine(Os.stdin()))
  print(File.readLine(Os.stdin()))
}
`, "hello\nworld\n", 200*time.Millisecond)
		want := "Err{ error: Error{ message: \"timeout\" } }\nOk{ value: \"hello\" }\nOk{ value: \"world\" }\nOk{ value: null }\n"
		if out != want {
			t.Fatalf("printed %q, want %q", out, want)
		}
	})
	t.Run("pipe", func(t *testing.T) {
		out := runWithLateStdin(t, `
func main(): void {
  print(File.pipe(Os.stdin(), Os.stdout(), Context.new(Time.duration(milliseconds: 50))))
  print(File.readLine(Os.stdin()))
  print(File.readAll(Os.stdin()))
}
`, "late\nrest\n", 200*time.Millisecond)
		want := "Err{ error: Error{ message: \"timeout\" } }\nOk{ value: \"late\" }\nOk{ value: \"rest\\n\" }\n"
		if out != want {
			t.Fatalf("printed %q, want %q", out, want)
		}
	})
}
