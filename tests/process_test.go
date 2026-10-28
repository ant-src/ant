//go:build unix

package tests

import (
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ant-src/ant"
)

func loadSource(t *testing.T, src string) *ant.Program {
	t.Helper()
	prog, err := ant.Load("main.ant", ant.WithSource("main.ant", src), ant.WithFS(fstest.MapFS{}))
	if err != nil {
		t.Fatal(err)
	}
	return prog
}

// TestProcessHostWriters checks children against a host whose streams are not files: what the
// program and its children write arrives in program order, and a child left running writes
// nothing into the host's writers once Run has returned, which the race detector would see.
func TestProcessHostWriters(t *testing.T) {
	prog := loadSource(t, `print("a")
Os.run("printf", ["b"])
print("c")
File.writeText(Os.stdout(), "d")
Os.wait(Result.unwrap(Os.start("printf", ["e"])))
print("f")
File.writeText(Os.stderr(), "1")
Os.run("sh", ["-c", "printf 2 >&2"])
File.writeText(Os.stderr(), "3")
Os.start("sh", ["-c", "sleep 0.2; head -c 100000 /dev/zero; head -c 100000 /dev/zero >&2"])
`)
	var out, errOut bytes.Buffer
	if err := prog.Run(context.Background(), ant.WithStdout(&out), ant.WithStderr(&errOut)); err != nil {
		t.Fatal(err)
	}
	if out.String() != "abcdef" || errOut.String() != "123" {
		t.Fatalf("printed %q and %q", out.String(), errOut.String())
	}
	for i := 0; i < 5; i++ {
		out.Reset()
		errOut.Reset()
		time.Sleep(100 * time.Millisecond)
	}
	if out.Len() != 0 || errOut.Len() != 0 {
		t.Fatalf("a child wrote %d and %d bytes after the run", out.Len(), errOut.Len())
	}
}

// TestProcessSharedStdin checks that children given a host's standard input that is not a file
// read it one at a time, beside Os.stdin.
func TestProcessSharedStdin(t *testing.T) {
	prog := loadSource(t, `var a = Result.unwrap(Os.start("cat", stdin: Os.stdin(), stdout: null))
var b = Result.unwrap(Os.start("cat", stdin: Os.stdin(), stdout: null))
println(File.readLine(Os.stdin()) is Ok)
Os.wait(a)
Os.wait(b)
`)
	var out bytes.Buffer
	in := bytes.NewReader(bytes.Repeat([]byte("line\n"), 10000))
	if err := prog.Run(context.Background(), ant.WithStdout(&out), ant.WithStdin(in)); err != nil {
		t.Fatal(err)
	}
}

// TestProcessInstanceClose checks that closing an instance, as a host cancelling it, kills the
// children its functions started.
func TestProcessInstanceClose(t *testing.T) {
	prog := loadSource(t, "func spawn(): int { Os.pid(Result.unwrap(Os.start(\"sleep\", [\"30\"]))) }\n")
	inst, err := prog.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	f, _ := inst.Func("spawn")
	v, err := f.Call()
	if err != nil {
		t.Fatal(err)
	}
	pid := strconv.FormatInt(v.(int64), 10)
	inst.Close()
	time.Sleep(100 * time.Millisecond)
	if exec.Command("kill", "-0", pid).Run() == nil {
		exec.Command("kill", "-9", pid).Run()
		t.Fatalf("the child %s outlived the instance", pid)
	}
}
