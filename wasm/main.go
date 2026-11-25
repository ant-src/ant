//go:build js && wasm

// Command wasm is Ant compiled to WebAssembly for JavaScript hosts, browsers and Node: the
// @ant-lang/wasm package loads it. It formats and checks sources, runs programs whose modules
// it is given and the tests of test files, and serves the language server and the debug adapter over messages instead of a
// stream.
//
// The host passes the name of a global function as the first argument; main calls it once
// with the API object and then waits for calls.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall/js"
	"testing/fstest"
	"time"
	_ "time/tzdata" // the zone database, which js/wasm has no file system to read
	"unicode/utf16"

	"ant-lang.org"
	"ant-lang.org/internal/batch"
	"ant-lang.org/internal/dap"
	"ant-lang.org/internal/format"
	"ant-lang.org/internal/frame"
	"ant-lang.org/internal/lsp"
)

func main() {
	if len(os.Args) < 2 {
		panic("ant.wasm: the name of the ready callback is missing")
	}
	api := js.Global().Get("Object").New()
	api.Set("format", js.FuncOf(formatSource))
	api.Set("check", js.FuncOf(check))
	api.Set("run", js.FuncOf(run))
	api.Set("test", js.FuncOf(test))
	api.Set("languageServer", js.FuncOf(languageServer))
	api.Set("debugAdapter", js.FuncOf(debugAdapter))
	js.Global().Get(os.Args[1]).Invoke(api)
	select {}
}

// formatSource is format(path, src): an object with the formatted text, or with the parser's
// error. Nothing here throws, since a Go panic would end the whole module.
func formatSource(this js.Value, args []js.Value) any {
	r := js.Global().Get("Object").New()
	out, err := format.Source(args[0].String(), []byte(args[1].String()))
	if err != nil {
		r.Set("error", err.Error())
	} else {
		r.Set("text", string(out))
	}
	return r
}

// check is check(path, files): the diagnostics of loading and checking the program rooted at
// path, whose modules are files, an object from path to source. Lines and columns are 1-based;
// columns count UTF-16 code units, as JavaScript strings do.
func check(this js.Value, args []js.Value) any {
	path, files := root(args[0].String()), sources(args[1])
	prog, err := ant.Load(path, ant.WithFS(mapFS(files)))
	var diags []ant.Diagnostic
	if le, ok := err.(*ant.LoadError); ok {
		diags = le.Diagnostics
	} else if err != nil {
		diags = []ant.Diagnostic{{Pos: ant.Position{Path: path}, Message: err.Error()}}
	} else {
		diags = prog.Check()
	}
	return diagnostics(diags, files)
}

// run is run(options): it runs a program and returns a Promise of its exit status. The
// options are path and files, as check takes, and the optional args (an array of strings), env
// (an object), stdin (a string), stdout and stderr (functions given each write as a
// Uint8Array), and signal (an AbortSignal, which interrupts the run as Ctrl-C does). An abort
// reaches the run only when the program waits, as on a channel or a timer, since JavaScript runs
// only then; a host stops a program that never waits by ending its worker.
func run(this js.Value, args []js.Value) any {
	opts := args[0]
	path, files := root(opts.Get("path").String()), sources(opts.Get("files"))
	// The program's output goes to the page in batches: a call into JavaScript per print would
	// make a print loop crawl.
	sinks := map[string]io.Writer{"stdout": writer(opts.Get("stdout")), "stderr": writer(opts.Get("stderr"))}
	out := batch.New(outputDelay, func(stream string, p []byte) { sinks[stream].Write(p) })
	stdout, stderr := out.Stream("stdout"), out.Stream("stderr")
	return promise(func() (any, error) {
		defer out.Flush()
		prog, err := ant.Load(path, ant.WithFS(mapFS(files)))
		if err != nil {
			if le, ok := err.(*ant.LoadError); ok {
				return result(ant.ExitStatus(err), diagnostics(le.Diagnostics, files)), nil
			}
			fmt.Fprintln(stderr, err)
			return result(ant.ExitStatus(err), js.Global().Get("Array").New()), nil
		}
		runOpts := []ant.Option{
			ant.WithStdout(stdout),
			ant.WithStderr(stderr),
			ant.WithStdin(strings.NewReader(str(opts.Get("stdin")))),
			ant.WithArgs(strs(opts.Get("args"))),
			ant.WithEnvironment(env(opts.Get("env"))),
		}
		if signal := opts.Get("signal"); truthy(signal) {
			interrupts := make(chan struct{}, 2)
			defer onAbort(signal, func() {
				select {
				case interrupts <- struct{}{}:
				default:
				}
			})()
			runOpts = append(runOpts, ant.WithInterrupts(interrupts))
		}
		err = prog.Run(context.Background(), runOpts...)
		if err != nil && !reported(err) {
			fmt.Fprintln(stderr, err) // as the ant command does
		}
		return result(ant.ExitStatus(err), js.Global().Get("Array").New()), nil
	})
}

// onAbort calls f when the AbortSignal aborts, at once when it has already. The function it
// returns stops listening.
func onAbort(signal js.Value, f func()) (stop func()) {
	listener := js.FuncOf(func(js.Value, []js.Value) any {
		f()
		return nil
	})
	signal.Call("addEventListener", "abort", listener)
	if signal.Get("aborted").Bool() {
		f()
	}
	return func() {
		signal.Call("removeEventListener", "abort", listener)
		listener.Release()
	}
}

// test is test(options): it runs the tests of test files as ant test does (spec/test.md) and
// returns a Promise of the command's exit status, with the checker's errors when there are any,
// in which case no test ran. The options are files, as check takes them, paths (the test files
// to run, in order), the optional run and skip (patterns, as the -run and -skip flags), failfast,
// env, stdin, stdout and stderr, as run takes them, signal (an AbortSignal, which ends the run),
// and onEvent, a function given each event of ant test -json as a JSON string. What a test
// prints is in its events; stdout and stderr get what the modules print as they initialize.
func test(this js.Value, args []js.Value) any {
	opts := args[0]
	files, paths := sources(opts.Get("files")), strs(opts.Get("paths"))
	onEvent := opts.Get("onEvent")
	sinks := map[string]io.Writer{"stdout": writer(opts.Get("stdout")), "stderr": writer(opts.Get("stderr"))}
	out := batch.New(outputDelay, func(stream string, p []byte) { sinks[stream].Write(p) })
	stdout, stderr := out.Stream("stdout"), out.Stream("stderr")
	return promise(func() (any, error) {
		defer out.Flush()
		for i, path := range paths {
			paths[i] = root(path)
			if !ant.IsTestFile(path) {
				return nil, fmt.Errorf("%s is not a test file, whose name ends in _test.ant", path)
			}
		}
		run, err := ant.TestPattern(str(opts.Get("run")))
		if err != nil {
			return nil, fmt.Errorf("run: %s", err)
		}
		skip, err := ant.TestPattern(str(opts.Get("skip")))
		if err != nil {
			return nil, fmt.Errorf("skip: %s", err)
		}
		fsys := ant.WithFS(mapFS(files))
		// A green run means the test files check (test.md section 1): an error runs nothing.
		if diags, loaded := ant.Check(paths, fsys); len(ant.Errors(diags)) > 0 || !loaded {
			status := 1
			if !loaded {
				status = 2
			}
			return result(status, diagnostics(ant.Errors(diags), files)), nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if signal := opts.Get("signal"); truthy(signal) {
			defer onAbort(signal, cancel)()
		}
		var mu sync.Mutex
		var passed, failed, skipped int
		emit := func(e testEvent) {
			out.Flush() // what the modules printed before the event shows before it
			if onEvent.Type() == js.TypeFunction {
				b, _ := json.Marshal(e)
				call(onEvent, string(b))
			}
		}
		var halt atomic.Bool
		cfg := ant.TestConfig{Run: run, Skip: skip, Parallel: runtime.GOMAXPROCS(0), Report: func(e ant.TestEvent) {
			mu.Lock()
			defer mu.Unlock()
			if ctx.Err() != nil {
				return
			}
			je := testEvent{File: e.File, Test: e.Name}
			secs := e.Elapsed.Seconds()
			switch e.Kind {
			case ant.TestRun:
				je.Action = "run"
			case ant.TestOutput:
				je.Action, je.Output = "output", e.Output
			case ant.TestPass:
				je.Action, je.Elapsed = "pass", &secs
			case ant.TestSkip:
				je.Action, je.Elapsed, je.Reason = "skip", &secs, e.Message
			case ant.TestFail:
				je.Action, je.Elapsed, je.Message = "fail", &secs, e.Message
			}
			if e.Counted {
				switch e.Kind {
				case ant.TestPass:
					passed++
				case ant.TestFail:
					failed++
				case ant.TestSkip:
					skipped++
				}
			}
			emit(je)
		}}
		failfast := opts.Get("failfast").Truthy()
		if failfast {
			cfg.Stop = &halt
		}
		start := time.Now()
		for _, file := range paths {
			if ctx.Err() != nil || failfast && halt.Load() {
				break
			}
			prog, err := ant.Load(file, fsys)
			if err != nil {
				return nil, err // the files loaded a moment ago
			}
			cfg.File = file
			err = prog.Test(ctx, cfg, ant.WithStdout(stdout), ant.WithStderr(stderr),
				ant.WithStdin(strings.NewReader(str(opts.Get("stdin")))), ant.WithEnvironment(env(opts.Get("env"))))
			var pe *ant.PanicError
			var ee *ant.ExitError
			switch {
			case err == nil:
			case errors.As(err, &ee):
				return result(ee.Status, js.Global().Get("Array").New()), nil
			case errors.As(err, &pe):
				// the runtime has reported the panic of an initializer; the file's tests did not run
				mu.Lock()
				failed++
				mu.Unlock()
				halt.Store(true)
			case ctx.Err() != nil:
			default:
				return nil, err
			}
		}
		if ctx.Err() != nil {
			return result(130, js.Global().Get("Array").New()), nil
		}
		mu.Lock()
		defer mu.Unlock()
		secs := time.Since(start).Seconds()
		emit(testEvent{Action: "done", Passed: &passed, Failed: &failed, Skipped: &skipped, Elapsed: &secs})
		status := 0
		if failed > 0 {
			status = 1
		}
		return result(status, js.Global().Get("Array").New()), nil
	})
}

// testEvent is an event of ant test -json (test.md section 2).
type testEvent struct {
	Action  string   `json:"action"`
	File    string   `json:"file,omitempty"`
	Test    string   `json:"test,omitempty"`
	Output  string   `json:"output,omitempty"`
	Elapsed *float64 `json:"elapsed,omitempty"`
	Reason  string   `json:"reason,omitempty"`
	Message string   `json:"message,omitempty"`
	Passed  *int     `json:"passed,omitempty"`
	Failed  *int     `json:"failed,omitempty"`
	Skipped *int     `json:"skipped,omitempty"`
}

// outputDelay is the longest the program's output waits for a batch to fill.
const outputDelay = 50 * time.Millisecond

// reported says whether the runtime has already said what err says, or there is nothing to say:
// a panic and a returned Err are on stderr, and an exit or an interrupt says nothing.
func reported(err error) bool {
	var pe *ant.PanicError
	var ee *ant.ExitError
	var re *ant.ResultError
	return errors.As(err, &pe) || errors.As(err, &ee) || errors.As(err, &re) || errors.Is(err, ant.ErrInterrupted)
}

func result(status int, diags any) js.Value {
	r := js.Global().Get("Object").New()
	r.Set("exitCode", status)
	r.Set("diagnostics", diags)
	return r
}

// languageServer is languageServer(onMessage, onLog): it starts a language server
// and returns a connection to it. The server's messages reach onMessage as JSON strings, and its
// log lines reach onLog when it is given. The connection's send takes a message as a JSON
// string, and close ends the server as the end of its input stream does.
func languageServer(this js.Value, args []js.Value) any {
	logw := logWriter(args, 1)
	return connection(args[0], func(in io.Reader, out io.Writer) {
		lsp.Serve(context.Background(), in, out, logw)
	})
}

// debugAdapter is debugAdapter(files, onMessage, onLog): it starts a debug adapter for
// one session on a program whose modules are files, as run takes them, and returns a connection
// to it, as languageServer does. The launch request's program and the sources of breakpoints
// are paths among files. A message reaches the adapter only when the program waits or is
// stopped, since JavaScript runs only then; a host stops a program that never waits by ending
// its worker.
func debugAdapter(this js.Value, args []js.Value) any {
	fsys := mapFS(sources(args[0]))
	logw := logWriter(args, 2)
	return connection(args[1], func(in io.Reader, out io.Writer) {
		dap.ServeFS(context.Background(), fsys, in, out, logw)
	})
}

// logWriter passes each line a server logs to the function that is argument i, when it is given.
func logWriter(args []js.Value, i int) io.Writer {
	if len(args) > i && truthy(args[i]) {
		return lineWriter(args[i])
	}
	return io.Discard
}

// connection runs serve, a server of framed messages on a stream, and returns a connection to
// it: the server's messages reach onMessage as JSON strings, send takes a message as a JSON
// string, and close ends the server as the end of its input stream does.
func connection(onMessage js.Value, serve func(in io.Reader, out io.Writer)) js.Value {
	in := &queue{}
	in.cond.L = &in.mu
	outr, outw := io.Pipe()
	go func() {
		serve(in, outw)
		outw.Close()
	}()
	go func() {
		r := bufio.NewReaderSize(outr, 1<<20)
		for {
			body, err := frame.Read(r)
			if err != nil {
				return
			}
			call(onMessage, string(body))
		}
	}()
	conn := js.Global().Get("Object").New()
	conn.Set("send", js.FuncOf(func(this js.Value, args []js.Value) any {
		var b strings.Builder
		frame.Write(&b, []byte(args[0].String()))
		in.write([]byte(b.String()))
		return nil
	}))
	conn.Set("close", js.FuncOf(func(js.Value, []js.Value) any {
		in.close()
		return nil
	}))
	return conn
}

// queue is the server's input stream: send appends to it without blocking, since a JavaScript
// call must not wait for Go, and the server reads it.
type queue struct {
	mu     sync.Mutex
	cond   sync.Cond
	buf    []byte
	closed bool
}

func (q *queue) write(p []byte) {
	q.mu.Lock()
	q.buf = append(q.buf, p...)
	q.mu.Unlock()
	q.cond.Signal()
}

func (q *queue) close() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	q.cond.Signal()
}

func (q *queue) Read(p []byte) (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.buf) == 0 && !q.closed {
		q.cond.Wait()
	}
	if len(q.buf) == 0 {
		return 0, io.EOF
	}
	n := copy(p, q.buf)
	q.buf = q.buf[n:]
	return n, nil
}

// promise runs f on a goroutine of its own, since Ant code may block, and returns a Promise
// of its result.
func promise(f func() (any, error)) js.Value {
	var handler js.Func
	handler = js.FuncOf(func(this js.Value, args []js.Value) any {
		resolve, reject := args[0], args[1]
		go func() {
			defer handler.Release()
			v, err := f()
			if err != nil {
				call(reject, js.Global().Get("Error").New(err.Error()))
				return
			}
			call(resolve, v)
		}()
		return nil
	})
	return js.Global().Get("Promise").New(handler)
}

// writer passes each write to the function f as a Uint8Array, or discards it when f is not a
// function.
func writer(f js.Value) io.Writer {
	if f.Type() != js.TypeFunction {
		return io.Discard
	}
	return writeFunc(func(p []byte) {
		a := js.Global().Get("Uint8Array").New(len(p))
		js.CopyBytesToJS(a, p)
		call(f, a)
	})
}

// lineWriter passes each line written, without its newline, to the function f.
func lineWriter(f js.Value) io.Writer {
	var mu sync.Mutex
	var pending []byte
	return writeFunc(func(p []byte) {
		mu.Lock()
		defer mu.Unlock()
		pending = append(pending, p...)
		for {
			i := strings.IndexByte(string(pending), '\n')
			if i < 0 {
				return
			}
			call(f, string(pending[:i]))
			pending = pending[i+1:]
		}
	})
}

type writeFunc func(p []byte)

func (w writeFunc) Write(p []byte) (int, error) {
	w(p)
	return len(p), nil
}

// call calls the JavaScript function f. Invoke panics when f throws, and a panic would end the
// whole module, every instance of the API with it, so the exception goes to the console instead.
func call(f js.Value, args ...any) {
	defer func() {
		if r := recover(); r != nil {
			js.Global().Get("console").Call("error", fmt.Sprintf("ant.wasm: a callback threw: %v", r))
		}
	}()
	f.Invoke(args...)
}

// root is the path of a program's root module in the file system mapFS makes.
func root(path string) string { return strings.TrimPrefix(path, "/") }

// mapFS is the file system a program is loaded from: its modules, by path.
func mapFS(files map[string]string) fs.FS {
	fsys := fstest.MapFS{}
	for path, src := range files {
		fsys[root(path)] = &fstest.MapFile{Data: []byte(src)}
	}
	return fsys
}

func sources(v js.Value) map[string]string {
	out := map[string]string{}
	if !truthy(v) {
		return out
	}
	keys := js.Global().Get("Object").Call("keys", v)
	for i := range keys.Length() {
		k := keys.Index(i).String()
		out[k] = v.Get(k).String()
	}
	return out
}

func env(v js.Value) map[string]string { return sources(v) }

func strs(v js.Value) []string {
	if !truthy(v) {
		return nil
	}
	out := make([]string, v.Length())
	for i := range out {
		out[i] = v.Index(i).String()
	}
	return out
}

func str(v js.Value) string {
	if v.Type() != js.TypeString {
		return ""
	}
	return v.String()
}

func truthy(v js.Value) bool { return !v.IsUndefined() && !v.IsNull() }

// diagnostics converts diagnostics to JavaScript objects, with columns in UTF-16 code units
// when the line's text is among files.
func diagnostics(diags []ant.Diagnostic, files map[string]string) js.Value {
	out := js.Global().Get("Array").New()
	for _, d := range diags {
		o := js.Global().Get("Object").New()
		o.Set("path", d.Pos.Path)
		o.Set("line", d.Pos.Line)
		o.Set("column", utf16Col(files, d.Pos))
		o.Set("message", d.Message)
		o.Set("code", d.Code)
		o.Set("severity", d.Severity)
		out.Call("push", o)
	}
	return out
}

// utf16Col is the 1-based UTF-16 column of a position whose column counts bytes.
func utf16Col(files map[string]string, pos ant.Position) int {
	src, ok := files[pos.Path]
	if !ok {
		src, ok = files["/"+pos.Path]
	}
	if !ok || pos.Line < 1 || pos.Col < 1 {
		return pos.Col
	}
	lines := strings.SplitN(src, "\n", pos.Line+1)
	if len(lines) < pos.Line {
		return pos.Col
	}
	line := lines[pos.Line-1]
	if pos.Col-1 > len(line) {
		return pos.Col
	}
	n := 0
	for _, r := range line[:pos.Col-1] {
		n += utf16.RuneLen(r) // an invalid byte decodes to U+FFFD, one unit
	}
	return n + 1
}
