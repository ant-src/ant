package lib

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// OutputTag is Os.Output: { status: int, stdout: string, stderr: string }.
var OutputTag = value.NewTag("Os.Output", value.RecordTag, []string{"status", "stdout", "stderr"})

// given reports whether an optional argument was passed and is not null.
func given(v value.Value) bool { return v != nil && v != value.Null }

// execProcess is Os.exec: no shell, the outputs captured, Err only when the child could not be
// started or run to its end, a nonzero exit as Ok with its status.
func execProcess(g *interp.Goroutine, a []value.Value) value.Value {
	const who = "Os.exec"
	in := g.Instance()
	program := stringArg(g, who, a[0])
	var args []string
	if given(a[1]) {
		for _, v := range arrayArg(g, who, a[1]).Values() {
			args = append(args, stringArg(g, who, v))
		}
	}
	ctx := waitContext(g, optionalContext(g, who, a[5]))
	cmd := exec.CommandContext(ctx, program, args...)
	cmd.WaitDelay = time.Second // a child that ignores the kill does not hold the run
	cmd.Env = in.Environ()      // the run's environment, which a host may have replaced
	if given(a[2]) {
		cmd.Dir = stringArg(g, who, a[2])
	}
	if given(a[3]) {
		env := []string{}
		for _, e := range mapArg(g, who, a[3]).Entries() {
			env = append(env, stringArg(g, who, e.Key)+"="+stringArg(g, who, e.Value))
		}
		cmd.Env = env
	}
	if given(a[4]) {
		cmd.Stdin = strings.NewReader(stringArg(g, who, a[4]))
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	in.Flush() // what the program printed precedes what it prints from the child
	err := cmd.Run()
	status := 0
	if err != nil {
		if ctx.Err() != nil {
			return gaveUp(ctx) // the child was killed, or never started
		}
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			return errResult(err)
		}
		status = exit.ExitCode()
	}
	return value.Ok(value.NewRecord(OutputTag, []value.Value{int64(status), stdout.String(), stderr.String()}))
}

func init() {
	register("Os", func() value.Value {
		return namespace("Os",
			constant("Output", OutputTag),
			fn("Os.args", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				args := g.Instance().Args()
				out := make([]value.Value, len(args))
				for i, s := range args {
					out[i] = s
				}
				return value.NewArray(out)
			}),
			fn("Os.env", "name", func(g *interp.Goroutine, a []value.Value) value.Value {
				v, ok := g.Instance().LookupEnv(stringArg(g, "Os.env", a[0]))
				if !ok {
					return value.Null
				}
				return v
			}),
			fn("Os.environ", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				m := value.NewMap()
				for _, kv := range g.Instance().Environ() {
					k, v, _ := strings.Cut(kv, "=")
					m.Set(k, v)
				}
				return m
			}),
			fn("Os.cwd", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				dir, err := os.Getwd()
				if err != nil {
					return errResult(err)
				}
				return value.Ok(dir)
			}),
			fn("Os.exit", "status", func(g *interp.Goroutine, a []value.Value) value.Value {
				g.Instance().Exit(int(intArg(g, "Os.exit", a[0])))
				return nil
			}),
			fn("Os.exec", "program, args?, dir?, env?, stdin?, ctx?", execProcess),
			// The standard streams are one handle each per run, so that a program reading its
			// input from two places shares one buffer.
			fn("Os.stdin", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				in := g.Instance()
				return in.Keep("Os.stdin", func() any {
					return &Handle{r: bufio.NewReader(in.Stdin()), std: true, in: in}
				}).(*Handle)
			}),
			fn("Os.stdout", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				in := g.Instance()
				return in.Keep("Os.stdout", func() any {
					return &Handle{w: bufio.NewWriter(in.Stdout()), std: true}
				}).(*Handle)
			}),
			fn("Os.stderr", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				in := g.Instance()
				return in.Keep("Os.stderr", func() any {
					return &Handle{w: bufio.NewWriter(in.Stderr()), std: true}
				}).(*Handle)
			}),
		)
	})
}
