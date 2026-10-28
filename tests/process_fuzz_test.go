//go:build unix

package tests

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
	"time"
	"unicode/utf8"

	"github.com/ant-src/ant"
)

// processChildren are the programs a generated script starts: quick ones, ones that wait for
// input, ones that ignore SIGTERM, ones that leave a grandchild holding the outputs.
var processChildren = []string{
	`"true"`,
	`"false"`,
	`"cat"`,
	`"sleep", ["0.05"]`,
	`"sh", ["-c", "echo out; echo err >&2"]`,
	`"sh", ["-c", "head -c 70000 /dev/zero"]`,
	`"sh", ["-c", "trap '' TERM; sleep 0.2"]`,
	`"sh", ["-c", "sleep 0.3 & echo bg"]`,
	`"sh", ["-c", "kill -9 $$"]`,
	`"sh", ["-c", "read line; echo got $line"]`,
	`"ant-no-such-program-xyz"`,
}

var processSignals = []string{"SIGTERM", "SIGKILL", "SIGINT", "SIGHUP", "SIGQUIT", "SIGUSR1", "SIGNOPE"}

// genProcessProgram turns ops into a well-typed script that exercises the process calls on
// three pipes and a file, with every wait bounded, and that stops every child at the end.
func genProcessProgram(ops []byte, dir string) string {
	var b strings.Builder
	b.WriteString("var pipes = [Result.unwrap(Os.openPipe()), Result.unwrap(Os.openPipe()), Result.unwrap(Os.openPipe())]\n")
	fmt.Fprintf(&b, "var file = Result.unwrap(File.create(%q))\n", dir+"/f.txt")
	b.WriteString("var procs: Array[Os.Process] = []\n")
	b.WriteString("func ms(n: int): Time.Duration { Time.duration(milliseconds: n) }\n")
	b.WriteString("func started(r: Result[Os.Process, Error]): void { if r is Ok { Array.push(procs, r.value) } else { println(r.error.message) } }\n")
	next := func(i *int) int {
		if *i >= len(ops) {
			return 0
		}
		v := int(ops[*i])
		*i++
		return v
	}
	input := func(c int) string {
		switch c % 5 {
		case 0:
			return "null"
		case 1, 2:
			return fmt.Sprintf("pipes[%d].read", c%3)
		case 3:
			return "Os.stdin()"
		}
		return "file"
	}
	output := func(c int) string {
		switch c % 6 {
		case 0:
			return "null"
		case 1, 2:
			return fmt.Sprintf("pipes[%d].write", c%3)
		case 3:
			return "Os.stdout()"
		case 4:
			return "Os.stderr()"
		}
		return "file"
	}
	proc := func(c int) string {
		return fmt.Sprintf("procs[%d %% Array.length(procs)]", c)
	}
	for i := 0; i < len(ops) && i < 400; {
		op := next(&i)
		switch op % 12 {
		case 0, 1:
			child := processChildren[next(&i)%len(processChildren)]
			ctx := ""
			if c := next(&i); c%3 == 0 {
				ctx = fmt.Sprintf(", ctx: Context.new(timeout: ms(%d))", c%200+1)
			}
			fmt.Fprintf(&b, "started(Os.start(%s, stdin: %s, stdout: %s, stderr: %s%s))\n",
				child, input(next(&i)), output(next(&i)), output(next(&i)), ctx)
		case 2:
			fmt.Fprintf(&b, "if Array.length(procs) > 0 { println(Os.wait(%s, ctx: Context.new(timeout: ms(%d))) is Ok) }\n", proc(next(&i)), next(&i)%100+1)
		case 3:
			fmt.Fprintf(&b, "if Array.length(procs) > 0 { println(Os.kill(%s, %q)) }\n", proc(next(&i)), processSignals[next(&i)%len(processSignals)])
		case 4:
			fmt.Fprintf(&b, "if Array.length(procs) > 0 { println(Os.stop(%s, grace: ms(%d)) is Ok) }\n", proc(next(&i)), next(&i)%60-10)
		case 5:
			c := next(&i)
			end := "read"
			if c%2 == 1 {
				end = "write"
			}
			fmt.Fprintf(&b, "println(File.close(pipes[%d].%s))\n", c%3, end)
		case 6:
			fmt.Fprintf(&b, "go { var r = File.readAll(pipes[%d].read, ctx: Context.new(timeout: ms(300))) }\n", next(&i)%3)
		case 7:
			fmt.Fprintf(&b, "go { var w = File.writeText(pipes[%d].write, \"line\\n\") }\n", next(&i)%3)
		case 8:
			child := processChildren[next(&i)%len(processChildren)]
			fmt.Fprintf(&b, "println(Os.exec(%s, stdin: \"in\\n\", ctx: Context.new(timeout: ms(%d))) is Ok)\n", child, next(&i)%300+1)
		case 9:
			child := processChildren[next(&i)%len(processChildren)]
			fmt.Fprintf(&b, "println(Os.output(%s, ctx: Context.new(timeout: ms(%d))) is Ok)\n", child, next(&i)%300+1)
		case 10:
			child := processChildren[next(&i)%len(processChildren)]
			fmt.Fprintf(&b, "println(Os.run(%s, ctx: Context.new(timeout: ms(%d))) is Ok)\n", child, next(&i)%300+1)
		case 11:
			fmt.Fprintf(&b, "if Array.length(procs) > 0 { println(Os.pid(%s) > 0) }\n", proc(next(&i)))
		}
	}
	b.WriteString("for p in procs { Os.stop(p, grace: ms(20)) }\n")
	b.WriteString("println(\"end\")\n")
	return b.String()
}

// FuzzProcess runs generated scripts of process calls. Each must check cleanly, end within the
// deadline, and print its last line.
func FuzzProcess(f *testing.F) {
	f.Add([]byte{0, 3, 1, 1, 1, 2, 0, 50})
	f.Add([]byte{0, 2, 0, 1, 3, 3, 5, 1, 6, 1, 2, 0, 90, 4, 0, 30})
	f.Add([]byte{1, 5, 0, 0, 2, 2, 6, 2, 5, 2, 7, 0, 3, 0, 1, 4, 0, 0})
	f.Add([]byte{0, 9, 1, 2, 3, 0, 7, 1, 5, 1, 2, 0, 99, 8, 9, 50, 9, 3, 20, 10, 4, 100})
	f.Add([]byte{0, 6, 0, 0, 0, 0, 4, 0, 5, 3, 0, 1, 0, 7, 3, 3, 3, 2, 0, 5})
	f.Fuzz(func(t *testing.T, ops []byte) {
		src := genProcessProgram(ops, t.TempDir())
		prog, err := ant.Load("main.ant", ant.WithSource("main.ant", src), ant.WithFS(fstest.MapFS{}))
		if err != nil {
			t.Fatalf("%v\n%s", err, src)
		}
		if diags := prog.Check(); len(diags) > 0 {
			t.Fatalf("%v\n%s", diags, src)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		var out, errOut bytes.Buffer
		err = prog.Run(ctx, ant.WithStdout(&out), ant.WithStderr(&errOut), ant.WithStdin(strings.NewReader("typed\n")))
		if errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("the script did not end\n%s", src)
		}
		if err != nil {
			t.Fatalf("%v\nstderr: %s\n%s", err, errOut.String(), src)
		}
		if !strings.HasSuffix(out.String(), "end\n") {
			t.Fatalf("printed %q\n%s", out.String(), src)
		}
	})
}

// roundTrip is a script that hands its arguments to a child, which writes each back followed by
// a NUL, and then reads the first as Os.output's text.
const roundTrip = `var args = ["-c", "printf '%s\\0' \"$@\"", "sh"]
for a in Os.args() { Array.push(args, a) }
var all = Os.exec("sh", args)
if all is Err { println("ERR") } else { File.writeText(Os.stdout(), all.value.stdout) }
var first = Os.output("printf", ["%s", Os.args()[0]])
if first is Err { println("ERR") } else { File.writeText(Os.stdout(), "|" + first.value + "|") }
`

// FuzzProcessArgs checks that arguments reach a child byte for byte, that a NUL in one is Err,
// and that Os.output removes one final newline and nothing else.
func FuzzProcessArgs(f *testing.F) {
	f.Add("plain", "two words")
	f.Add("trailing\n\n", "\r\n")
	f.Add("quote'\"$HOME`ls`", "")
	f.Add("nul\x00inside", "x")
	f.Add("-n", "ünïcødé 🐜")
	prog, err := ant.Load("main.ant", ant.WithSource("main.ant", roundTrip), ant.WithFS(fstest.MapFS{}))
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		if !utf8.ValidString(a) || !utf8.ValidString(b) {
			return
		}
		var out bytes.Buffer
		if err := prog.Run(context.Background(), ant.WithStdout(&out), ant.WithArgs([]string{a, b})); err != nil {
			t.Fatal(err)
		}
		want := a + "\x00" + b + "\x00"
		if strings.ContainsRune(a+b, 0) {
			want = "ERR\n"
		}
		if strings.ContainsRune(a, 0) {
			want += "ERR\n"
		} else {
			text := strings.TrimSuffix(a, "\n")
			if strings.HasSuffix(a, "\r\n") {
				text = strings.TrimSuffix(a, "\r\n")
			}
			want += "|" + text + "|"
		}
		if out.String() != want {
			t.Fatalf("args %q %q: got %q, want %q", a, b, out.String(), want)
		}
	})
}
