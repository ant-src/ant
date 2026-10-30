//go:build js && wasm

// Command wasm is Ant compiled to WebAssembly for JavaScript hosts, browsers and Node: the
// @ant-lang/wasm package loads it. It formats and checks sources, runs programs whose modules
// it is given, and serves the language server over messages instead of a stream.
//
// The host passes the name of a global function as the first argument; main calls it once
// with the API object and then waits for calls.
package main

import (
	"bufio"
	"context"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"
	"syscall/js"
	"testing/fstest"
	"unicode/utf16"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/internal/format"
	"github.com/ant-src/ant/internal/frame"
	"github.com/ant-src/ant/internal/lsp"
)

func main() {
	if len(os.Args) < 2 {
		panic("ant.wasm: the name of the ready callback is missing")
	}
	api := js.Global().Get("Object").New()
	api.Set("format", js.FuncOf(formatSource))
	api.Set("check", js.FuncOf(check))
	api.Set("run", js.FuncOf(run))
	api.Set("languageServer", js.FuncOf(languageServer))
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
	path, files := args[0].String(), sources(args[1])
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
// Uint8Array), and signal (an AbortSignal, which interrupts the run as Ctrl-C does).
func run(this js.Value, args []js.Value) any {
	opts := args[0]
	path, files := opts.Get("path").String(), sources(opts.Get("files"))
	return promise(func() (any, error) {
		prog, err := ant.Load(path, ant.WithFS(mapFS(files)))
		if err != nil {
			if le, ok := err.(*ant.LoadError); ok {
				return result(ant.ExitStatus(err), diagnostics(le.Diagnostics, files)), nil
			}
			return nil, err
		}
		runOpts := []ant.Option{
			ant.WithStdout(writer(opts.Get("stdout"))),
			ant.WithStderr(writer(opts.Get("stderr"))),
			ant.WithStdin(strings.NewReader(str(opts.Get("stdin")))),
			ant.WithArgs(strs(opts.Get("args"))),
			ant.WithEnvironment(env(opts.Get("env"))),
		}
		if signal := opts.Get("signal"); truthy(signal) {
			interrupts := make(chan struct{}, 2)
			onAbort := js.FuncOf(func(js.Value, []js.Value) any {
				select {
				case interrupts <- struct{}{}:
				default:
				}
				return nil
			})
			defer onAbort.Release()
			signal.Call("addEventListener", "abort", onAbort)
			defer signal.Call("removeEventListener", "abort", onAbort)
			if signal.Get("aborted").Bool() {
				interrupts <- struct{}{}
			}
			runOpts = append(runOpts, ant.WithInterrupts(interrupts))
		}
		err = prog.Run(context.Background(), runOpts...)
		return result(ant.ExitStatus(err), js.Global().Get("Array").New()), nil
	})
}

func result(status int, diags any) js.Value {
	r := js.Global().Get("Object").New()
	r.Set("exitCode", status)
	r.Set("diagnostics", diags)
	return r
}

// languageServer is languageServer(onMessage, onLog): it starts a language server (spec/lsp.md)
// and returns a connection to it. The server's messages reach onMessage as JSON strings, and its
// log lines reach onLog when it is given. The connection's send takes a message as a JSON
// string, and close ends the server as the end of its input stream does.
func languageServer(this js.Value, args []js.Value) any {
	onMessage := args[0]
	var logw io.Writer = io.Discard
	if len(args) > 1 && truthy(args[1]) {
		logw = lineWriter(args[1])
	}
	in := &queue{}
	in.cond.L = &in.mu
	outr, outw := io.Pipe()
	go func() {
		lsp.Serve(context.Background(), in, outw, logw)
		outw.Close()
	}()
	go func() {
		r := bufio.NewReaderSize(outr, 1<<20)
		for {
			body, err := frame.Read(r)
			if err != nil {
				return
			}
			onMessage.Invoke(string(body))
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
				reject.Invoke(js.Global().Get("Error").New(err.Error()))
				return
			}
			resolve.Invoke(v)
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
		f.Invoke(a)
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
			f.Invoke(string(pending[:i]))
			pending = pending[i+1:]
		}
	})
}

type writeFunc func(p []byte)

func (w writeFunc) Write(p []byte) (int, error) {
	w(p)
	return len(p), nil
}

// mapFS is the file system a program is loaded from: its modules, by path.
func mapFS(files map[string]string) fs.FS {
	fsys := fstest.MapFS{}
	for path, src := range files {
		fsys[strings.TrimPrefix(path, "/")] = &fstest.MapFile{Data: []byte(src)}
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
