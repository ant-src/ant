package lib

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// The processes of the Os namespace (stdlib.md, Os): Os.run in the foreground with the program's
// streams, Os.exec and Os.output captured, Os.start in the background with any handles, and the
// pipes that join them. Every call starts a Process, so the blocking ones are a start, the
// captures, and a wait.

// OutputTag is Os.Output: { status: int, stdout: string, stderr: string, stdoutBytes: Bytes,
// stderrBytes: Bytes }, each output as text and as the child wrote it.
var OutputTag = value.NewTag("Os.Output", value.RecordTag, []string{"status", "stdout", "stderr", "stdoutBytes", "stderrBytes"})

// PipeTag is Os.Pipe: { read: File.Handle, write: File.Handle }.
var PipeTag = value.NewTag("Os.Pipe", value.RecordTag, []string{"read", "write"})

// waitDelay is how long a wait goes on, after the child exited, for the copies into and out of
// it: a process the child started may hold a pipe open long after.
const waitDelay = time.Second

// Process is a started child: Os.Process. A goroutine waits for it from the start, so it is
// reaped whether or not the program waits, and records the outcome every Os.wait returns.
type Process struct {
	cmd    *exec.Cmd
	group  bool // it leads a process group of its own, which its signals reach
	mu     sync.Mutex
	killed context.Context // the context whose end made the runtime kill it; nil when none did
	done   chan struct{}   // closed once the outcome is recorded
	status int
	err    error // what failed beside the exit status, such as a copy
}

// SyncName names the process for rendering.
func (p *Process) SyncName() string { return "Os.Process" }

func processArg(g *interp.Goroutine, who string, v value.Value) *Process {
	return typedArg[*Process](g, who, "an Os.Process", v)
}

// signal sends sig to the child, and to its group when it leads one. It is os.ErrProcessDone,
// and nothing is sent, once the child has exited.
func (p *Process) signal(sig os.Signal) error {
	if p.group {
		return signalGroup(p.cmd.Process, sig)
	}
	return p.cmd.Process.Signal(sig)
}

// cancel kills the child because ctx ended, unless it has exited, in which case its status
// stands: the race is decided by whether the kill was sent.
func (p *Process) cancel(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.killed == nil && p.signal(os.Kill) == nil {
		p.killed = ctx
	}
}

// outcome is what a wait returns once the child is done: Err with the reason of the context
// that killed it, Err when a copy failed, and otherwise Ok of its status.
func (p *Process) outcome() value.Value {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch {
	case p.killed != nil:
		return gaveUp(p.killed)
	case p.err != nil:
		return errResult(p.err)
	}
	return value.Ok(int64(p.status))
}

// exited reports whether the outcome is recorded.
func (p *Process) exited() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}

// command builds the child that a call's first four arguments describe: program, looked up in
// the run's PATH when it has no path separator, args, dir, and env as the whole environment.
func command(g *interp.Goroutine, who string, a []value.Value) (*exec.Cmd, value.Value) {
	in := g.Instance()
	program := stringArg(g, who, a[0])
	args := []string{program}
	if given(a[1]) {
		for _, v := range arrayArg(g, who, a[1]).Values() {
			args = append(args, stringArg(g, who, v))
		}
	}
	// A path is taken as it is, relative to dir when it is relative, as Go's exec.Command does.
	path := program
	if filepath.Base(program) == program {
		var err error
		if path, err = lookPath(in, program); err != nil {
			return nil, errResult(err)
		}
	}
	cmd := &exec.Cmd{Path: path, Args: args, Env: in.Environ(), WaitDelay: waitDelay}
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
	return cmd, nil
}

// input is what a child reads through a copy the runtime runs, rather than as a file: the copy
// writes r into a pipe the child reads, and ends when the child is done, through dl's read
// deadline or its stop channel when the source can wait; release runs once it has ended. A copy from the host's own
// standard input, which nothing can interrupt, is left behind rather than waited for.
type input struct {
	r       io.Reader
	dl      deadliner
	stop    chan struct{} // closed to end a copy from a pump
	abandon bool
	release func()
}

func (f *input) end() {
	if f.release != nil {
		f.release()
	}
}

// startProcess starts cmd, flushing standard output first, and returns the child, with a
// goroutine that reaps it and then ends the copy of feed into it, when there is one. The child
// is killed when ctx ends, when ctx is not nil, and when a cancellation ends the run. group puts
// it in a process group of its own.
func startProcess(in *interp.Instance, cmd *exec.Cmd, ctx context.Context, group bool, feed *input) (*Process, value.Value) {
	if ctx != nil && ctx.Err() != nil {
		if feed != nil {
			feed.end()
		}
		return nil, gaveUp(ctx)
	}
	if group {
		setGroup(cmd)
	}
	var pw *os.File
	if feed != nil {
		pr, w, err := os.Pipe()
		if err != nil {
			feed.end()
			return nil, errResult(err)
		}
		cmd.Stdin, pw = pr, w
		defer pr.Close() // the child has its own copy once it starts
	}
	in.Flush() // what the program printed precedes what the child prints
	if err := cmd.Start(); err != nil {
		if feed != nil {
			pw.Close()
			feed.end()
		}
		return nil, errResult(err)
	}
	endFeed := func() {}
	if feed != nil {
		copied := make(chan struct{})
		go func() {
			io.Copy(pw, feed.r) // a child that exits early ends it with a broken pipe
			pw.Close()
			close(copied)
		}()
		endFeed = func() {
			if feed.dl != nil {
				feed.dl.SetReadDeadline(time.Now())
			}
			if feed.stop != nil {
				close(feed.stop)
			}
			if !feed.abandon {
				<-copied
			}
			if feed.dl != nil {
				feed.dl.SetReadDeadline(time.Time{})
			}
			feed.end()
		}
	}
	p := &Process{cmd: cmd, group: group, done: make(chan struct{})}
	stopAfter := func() bool { return false }
	if ctx != nil {
		stopAfter = context.AfterFunc(ctx, func() { p.cancel(ctx) })
	}
	// The run's end does not wait for the AfterFunc: a child that a cancellation, or the end of
	// its context, should kill is killed before the run returns.
	removeStop := in.OnStop(func(aborted bool) {
		switch {
		case aborted:
			p.cancel(in.Context())
		case ctx != nil && ctx.Err() != nil:
			p.cancel(ctx)
		}
	})
	go func() {
		err := cmd.Wait()
		stopAfter()
		removeStop()
		endFeed()
		p.mu.Lock()
		if cmd.ProcessState != nil {
			p.status = exitStatus(cmd.ProcessState)
		}
		var exit *exec.ExitError
		if err != nil && !errors.As(err, &exit) && !errors.Is(err, exec.ErrWaitDelay) {
			p.err = err
		}
		p.mu.Unlock()
		close(p.done)
	}()
	return p, nil
}

// childInput gives a child of Os.start the handle h as its standard input: a standard stream
// as the run's own, a file with nothing read ahead as it is, and anything else as a copy, with
// the handle's read slot held until the copy ends, so the program does not read it meanwhile.
// A pipe end handed over as it is reads through a pump from then on, since starting the child
// makes it blocking and a deadline no longer ends a read of it.
func childInput(in *interp.Instance, h *Handle) (file *os.File, feed *input, fail value.Value) {
	if h.r == nil || h.isClosed() {
		return nil, nil, value.Err(value.ErrorValue(errNotReadable))
	}
	if h.std {
		return stdinFor(in)
	}
	ctx := in.Context()
	select {
	case h.reads <- struct{}{}:
	case <-ctx.Done():
		return nil, nil, gaveUp(ctx)
	}
	if f := h.file(); f != nil && h.pending == "" && h.r.Buffered() == 0 {
		if h.via != nil && h.dl != nil {
			p := &pump{src: f, done: ctx.Done()}
			h.via.r, h.pump, h.dl = p, p, nil
		}
		<-h.reads
		return f, nil, nil
	}
	copied := &input{r: handleReader{h}, dl: h.dl, release: func() { <-h.reads }}
	if p := h.pump; p != nil {
		// A pipe end handed over before reads through a pump, which the copy's end stops.
		stop := make(chan struct{})
		p.cancel, copied.stop = stop, stop
		copied.release = func() {
			p.cancel = nil
			<-h.reads
		}
	}
	return nil, copied, nil
}

// stdinFor is the run's standard input as a child is given it: the file as it is, and a host's
// reader through a copy that is left behind when the child is done.
func stdinFor(in *interp.Instance) (*os.File, *input, value.Value) {
	switch r := in.ChildStdin().(type) {
	case nil:
		return nil, nil, nil
	case *os.File:
		return r, nil, nil
	default:
		return nil, &input{r: r, abandon: true}, nil
	}
}

// childOutput gives a child the handle h as an output, flushed first, so what the program wrote
// to it comes first: a standard stream as the run's own, a file as it is, anything else through
// a copy the runtime writes into the handle.
func childOutput(h *Handle) (io.Writer, value.Value) {
	if h.w == nil || h.isClosed() {
		return nil, value.Err(value.ErrorValue(errNotWritable))
	}
	h.writes.Lock()
	err := h.w.Flush()
	h.writes.Unlock()
	switch {
	case err != nil:
		return nil, errResult(err)
	case h.std:
		return h.childOut, nil
	}
	if f := h.file(); f != nil {
		return f, nil
	}
	return handleWriter{h}, nil
}

// file returns the file behind h, which a child can be given as it is, or nil.
func (h *Handle) file() *os.File {
	if h.conn {
		return nil
	}
	f, _ := h.c.(*os.File)
	return f
}

// viaReader is a pipe's reading end as its handle reads it, so the source can become a pump,
// under the read slot, once the end is handed to a child.
type viaReader struct{ r io.Reader }

func (v *viaReader) Read(b []byte) (int, error) { return v.r.Read(b) }

// handleReader reads what h read ahead and then the rest of its input, for a copy into a child.
type handleReader struct{ h *Handle }

func (r handleReader) Read(b []byte) (int, error) {
	if r.h.pending != "" {
		n := copy(b, r.h.pending)
		r.h.pending = r.h.pending[n:]
		return n, nil
	}
	return r.h.r.Read(b)
}

// handleWriter writes a child's output into h, flushed as it comes.
type handleWriter struct{ h *Handle }

func (w handleWriter) Write(b []byte) (int, error) {
	w.h.writes.Lock()
	defer w.h.writes.Unlock()
	n, err := w.h.w.Write(b)
	if err == nil {
		err = w.h.w.Flush()
	}
	return n, err
}

// capture runs a child for Os.exec and Os.output, arguments program, args, dir, env, stdin and
// ctx: both outputs captured, in the program's process group, the call blocking until the child
// is done. fail is the Err of a child that could not start or that ctx ended.
func capture(g *interp.Goroutine, who string, a []value.Value) (status int, stdout, stderr []byte, fail value.Value) {
	cmd, fail := command(g, who, a)
	if fail != nil {
		return 0, nil, nil, fail
	}
	var feed *input
	if given(a[4]) {
		feed = &input{r: strings.NewReader(textOrBytes(g, who, a[4]))}
	}
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	p, fail := startProcess(g.Instance(), cmd, waitContext(g, who, a[5]), false, feed)
	if fail != nil {
		return 0, nil, nil, fail
	}
	<-p.done
	g.AfterWait()
	if r := p.outcome(); r.(*value.Record).Tag == value.ErrTag {
		return 0, nil, nil, r
	}
	return p.status, out.Bytes(), errOut.Bytes(), nil
}

// execProcess is Os.exec: the status and both outputs, a nonzero status being Ok.
func execProcess(g *interp.Goroutine, who string, a []value.Value) value.Value {
	status, stdout, stderr, fail := capture(g, who, a)
	if fail != nil {
		return fail
	}
	outText, outBytes := received(stdout)
	errText, errBytes := received(stderr)
	return value.Ok(value.NewRecord(OutputTag, []value.Value{int64(status), outText, errText, outBytes, errBytes}))
}

// outputProcess is Os.output: standard output as text without one final newline, and Err on a
// nonzero status, with standard error in the message.
func outputProcess(g *interp.Goroutine, who string, a []value.Value) value.Value {
	status, stdout, stderr, fail := capture(g, who, a)
	if fail != nil {
		return fail
	}
	if status != 0 {
		msg := fmt.Sprintf("%s: exit status %d", stringArg(g, who, a[0]), status)
		if e := strings.TrimRightFunc(string(stderr), unicode.IsSpace); e != "" {
			msg += ": " + e
		}
		return value.Err(value.ErrorValue(value.Text(msg)))
	}
	if bytes.HasSuffix(stdout, []byte("\r\n")) {
		stdout = stdout[:len(stdout)-2]
	} else if bytes.HasSuffix(stdout, []byte("\n")) {
		stdout = stdout[:len(stdout)-1]
	}
	text, _ := received(stdout)
	return value.Ok(text)
}

// runProcess is Os.run: the child in the foreground with the program's three streams, the host
// told, so it can hold back its keyboard signals until the child exits.
func runProcess(g *interp.Goroutine, who string, a []value.Value) value.Value {
	in := g.Instance()
	cmd, fail := command(g, who, a)
	if fail != nil {
		return fail
	}
	cmd.Stdout, cmd.Stderr = in.ChildStdout(), in.ChildStderr()
	file, feed, _ := stdinFor(in)
	if file != nil {
		cmd.Stdin = file
	}
	release := in.Foreground()
	p, fail := startProcess(in, cmd, waitContext(g, who, a[4]), false, feed)
	if fail != nil {
		release(nil)
		return fail
	}
	<-p.done
	release(endedBy(cmd.ProcessState))
	in.Flush()
	g.AfterWait()
	return p.outcome()
}

// startBackground is Os.start: the child in a process group of its own, its streams the
// handles given, standard input the null device and the outputs the program's by default.
func startBackground(g *interp.Goroutine, who string, a []value.Value) value.Value {
	in := g.Instance()
	cmd, fail := command(g, who, a)
	if fail != nil {
		return fail
	}
	var ctx context.Context
	if given(a[7]) {
		ctx = contextArg(g, who, a[7]).Ctx
	}
	var feed *input
	undo := func() {
		if feed != nil {
			feed.end()
		}
	}
	if v := a[4]; given(v) {
		file, f, fail := childInput(in, handleArg(g, who, v))
		if fail != nil {
			return fail
		}
		if file != nil && isTerminal(file) {
			return value.Err(value.ErrorValue("a child in the background cannot read the terminal; Os.run gives it the terminal"))
		}
		if file != nil {
			cmd.Stdin = file
		}
		feed = f
	}
	for i, v := range []value.Value{a[5], a[6]} {
		set, std := &cmd.Stdout, in.ChildStdout
		if i == 1 {
			set, std = &cmd.Stderr, in.ChildStderr
		}
		switch {
		case v == nil:
			*set = std()
		case given(v):
			w, fail := childOutput(handleArg(g, who, v))
			if fail != nil {
				undo()
				return fail
			}
			*set = w
		}
	}
	p, fail := startProcess(in, cmd, ctx, true, feed)
	if fail != nil {
		return fail
	}
	return value.Ok(p)
}

// waitProcess is Os.wait: the outcome once the child is done, or Err when ctx ends first.
func waitProcess(g *interp.Goroutine, who string, a []value.Value) value.Value {
	p := processArg(g, who, a[0])
	ctx := waitContext(g, who, a[1])
	select {
	case <-p.done:
	case <-ctx.Done():
		if !p.exited() { // a child done by then gives its outcome, whichever select saw first
			g.AfterWait()
			return gaveUp(ctx)
		}
	}
	g.AfterWait()
	return p.outcome()
}

// killProcess is Os.kill: the signal named, sent to the child's group, Ok when it has exited.
func killProcess(g *interp.Goroutine, who string, a []value.Value) value.Value {
	p := processArg(g, who, a[0])
	name := "SIGTERM"
	if given(a[1]) {
		name = stringArg(g, who, a[1])
	}
	sig, err := signalNamed(name)
	if err != nil {
		return errResult(err)
	}
	if err := p.signal(sig); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return errResult(err)
	}
	return value.Ok(value.Void)
}

// stopProcess is Os.stop: SIGTERM, SIGKILL when grace passes first, and the outcome.
func stopProcess(g *interp.Goroutine, who string, a []value.Value) value.Value {
	p := processArg(g, who, a[0])
	grace := 5 * time.Second
	if given(a[1]) {
		grace = spanArg(g, who, a[1])
	}
	run := g.Context()
	kill := func() {
		p.signal(os.Kill)
		select {
		case <-p.done:
		case <-run.Done():
		}
	}
	switch {
	case p.exited():
	case grace <= 0:
		kill() // no grace: SIGKILL alone, so the status says which ended it
	default:
		p.signal(terminate)
		timer := time.NewTimer(grace)
		select {
		case <-p.done:
		case <-timer.C:
			kill()
		case <-run.Done():
		}
		timer.Stop()
	}
	g.AfterWait()
	return p.outcome()
}

// openPipe is Os.openPipe: two handles, what is written to one read from the other. A read of
// the reading end waits, and a context ends it through the deadline; a write is flushed.
func openPipe(g *interp.Goroutine, who string, a []value.Value) value.Value {
	r, w, err := os.Pipe()
	if err != nil {
		return errResult(err)
	}
	via := &viaReader{r}
	read := newHandle(r, via, nil)
	read.dl, read.via = r, via
	write := newHandle(w, nil, w)
	write.pipe = true
	return value.Ok(value.NewRecord(PipeTag, []value.Value{read, write}))
}

// processEntries are the Os namespace's entries for processes.
func processEntries() []any {
	return []any{
		constant("Output", OutputTag),
		constant("Pipe", PipeTag),
		constant("Process", value.NewTag("Os.Process", value.SyncTag, nil)),
		fn("Os.exec", "program, args?, dir?, env?, stdin?, ctx?", execProcess),
		fn("Os.output", "program, args?, dir?, env?, stdin?, ctx?", outputProcess),
		fn("Os.run", "program, args?, dir?, env?, ctx?", runProcess),
		fn("Os.start", "program, args?, dir?, env?, stdin?, stdout?, stderr?, ctx?", startBackground),
		fn("Os.wait", "p, ctx?", waitProcess),
		fn("Os.kill", "p, signal?", killProcess),
		fn("Os.stop", "p, grace?", stopProcess),
		fn("Os.pid", "p", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			return int64(processArg(g, who, a[0]).cmd.Process.Pid)
		}),
		fn("Os.openPipe", "", openPipe),
		fn("Os.which", "program", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			path, err := lookPath(g.Instance(), stringArg(g, who, a[0]))
			if err != nil {
				return value.Null
			}
			return value.Text(path)
		}),
	}
}
