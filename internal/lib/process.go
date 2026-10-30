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

// Commands (stdlib.md, Cmd): a Cmd says what to run and a verb how, Cmd.run in the foreground
// with the program's streams, Cmd.read and Cmd.capture captured, Cmd.start in the background with
// any handles. A Cmd.Pipeline runs its commands at once, each one's output the next one's input.
// Every verb starts a Process, so the blocking ones are a start, the captures, and a wait.

// CmdTag is Cmd: { program: string, args: Array[string], dir: string | null,
// env: Map[string, string] | null }, with the defaults the prelude declares.
var CmdTag = value.NewTag("Cmd", value.RecordTag, []string{"program", "args", "dir", "env"})

// PipelineTag is Cmd.Pipeline: { cmds: Array[Cmd] }.
var PipelineTag = value.NewTag("Cmd.Pipeline", value.RecordTag, []string{"cmds"})

// OutputTag is Cmd.Output: { status: int, statuses: Array[int], stdout: string, stderr: string,
// stdoutBytes: Bytes, stderrBytes: Bytes }, each output as text and as the children wrote it.
var OutputTag = value.NewTag("Cmd.Output", value.RecordTag, []string{"status", "statuses", "stdout", "stderr", "stdoutBytes", "stderrBytes"})

// PipeTag is Os.Pipe: { read: File.Handle, write: File.Handle }.
var PipeTag = value.NewTag("Os.Pipe", value.RecordTag, []string{"read", "write"})

func init() {
	CmdTag.Defaults = interp.NativeDefaults{
		nil,
		func() value.Value { return value.NewArray([]value.Value{}) },
		func() value.Value { return value.Null },
		func() value.Value { return value.Null },
	}
	register("Cmd", func() value.Value {
		withFuncs(CmdTag, cmdFuncs())
		CmdTag.SetFunc("Pipeline", PipelineTag)
		CmdTag.SetFunc("Output", OutputTag)
		CmdTag.SetFunc("Process", value.NewTag("Cmd.Process", value.SyncTag, nil))
		return CmdTag
	})
}

// waitDelay is how long a wait goes on, after the child exited, for the copies into and out of
// it: a process the child started may hold a pipe open long after.
const waitDelay = time.Second

// Process is a started command, one child per stage: Cmd.Process. A goroutine waits for them
// from the start, so they are reaped whether or not the program waits, and records the outcome
// every Cmd.wait returns.
type Process struct {
	cmds     []*exec.Cmd
	programs []string // each stage's program as its Cmd names it, for messages
	group    bool     // the stages share a process group of their own, which signals reach
	mu       sync.Mutex
	killed   context.Context // the context whose end made the runtime kill it; nil when none did
	done     chan struct{}   // closed once the outcome is recorded
	status   int             // the pipeline's status (pipelineStatus)
	failed   int             // the stage that set status, or -1
	statuses []int
	err      error // what failed beside the exit status, such as a copy
}

// SyncName names the process for rendering.
func (p *Process) SyncName() string { return "Cmd.Process" }

func processArg(g *interp.Goroutine, who string, v value.Value) *Process {
	return typedArg[*Process](g, who, "a Cmd.Process", v)
}

// last is the stage whose output and pid the process's are.
func (p *Process) last() *exec.Cmd { return p.cmds[len(p.cmds)-1] }

// signal sends sig to every stage, through their group when they share one. It is
// os.ErrProcessDone, and nothing is sent, once every stage has exited.
func (p *Process) signal(sig os.Signal) error { return signalStages(p.cmds, p.group, sig) }

// signalStages sends sig to the started stages cmds, through their group when they share one.
func signalStages(cmds []*exec.Cmd, group bool, sig os.Signal) error {
	procs := make([]*os.Process, len(cmds))
	for i, c := range cmds {
		procs[i] = c.Process
	}
	if group {
		return signalGroup(procs, sig)
	}
	return signalEach(procs, sig)
}

// signalEach sends sig to each of ps alone. It is nil when one got it, os.ErrProcessDone when
// every one has exited, and otherwise the first other failure.
func signalEach(ps []*os.Process, sig os.Signal) error {
	err := os.ErrProcessDone
	for _, p := range ps {
		switch e := p.Signal(sig); {
		case e == nil:
			err = nil
		case !errors.Is(e, os.ErrProcessDone) && errors.Is(err, os.ErrProcessDone):
			err = e
		}
	}
	return err
}

// cancel kills the stages because ctx ended, unless they have exited, in which case their status
// stands: the race is decided by whether the kill was sent.
func (p *Process) cancel(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.killed == nil && p.signal(os.Kill) == nil {
		p.killed = ctx
	}
}

// outcome is what a wait returns once the stages are done: Err with the reason of the context
// that killed them, Err when a copy failed, and otherwise Ok of the status.
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

// pipelineStatus is a pipeline's status from its stages': the rightmost that failed, as under
// bash's pipefail, and the stage that set it, or 0 and -1 when none did. A stage before the last
// that a broken pipe ended does not count, since only a later stage that stopped reading can end
// it so: yes | head -n 1 succeeds.
func pipelineStatus(cmds []*exec.Cmd, statuses []int) (status, stage int) {
	for i := len(statuses) - 1; i >= 0; i-- {
		if statuses[i] == 0 {
			continue
		}
		if i < len(statuses)-1 && brokenPipe != nil && endedBy(cmds[i].ProcessState) == brokenPipe {
			continue
		}
		return statuses[i], i
	}
	return 0, -1
}

// failure is the message of a Cmd.read whose pipeline failed: the program of the stage that set
// the status, with its place when another stage runs the same program, the status, and what the
// stages wrote to standard error.
func (p *Process) failure(stderr []byte) string {
	program := p.programs[p.failed]
	same := 0
	for _, q := range p.programs {
		if q == program {
			same++
		}
	}
	if same > 1 {
		program += fmt.Sprintf(" (stage %d)", p.failed+1)
	}
	msg := fmt.Sprintf("%s: exit status %d", program, p.status)
	if e := strings.TrimRightFunc(string(stderr), unicode.IsSpace); e != "" {
		msg += ": " + e
	}
	return msg
}

// command builds the child a Cmd describes: program, looked up in the run's PATH when it has no
// path separator, args, dir, and env as the whole environment.
func command(g *interp.Goroutine, who string, c *value.Record) (*exec.Cmd, value.Value) {
	in := g.Instance()
	program := stringArg(g, who, c.Get(0))
	args := []string{program}
	if given(c.Get(1)) {
		for _, v := range arrayArg(g, who, c.Get(1)).Values() {
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
	if given(c.Get(2)) {
		cmd.Dir = stringArg(g, who, c.Get(2))
	}
	if given(c.Get(3)) {
		env := []string{}
		for _, e := range mapArg(g, who, c.Get(3)).Entries() {
			env = append(env, stringArg(g, who, e.Key)+"="+stringArg(g, who, e.Value))
		}
		cmd.Env = env
	}
	return cmd, nil
}

// cmdStages reads a Cmd or a Cmd.Pipeline as its stages, one Cmd record each. An empty pipeline
// panics, as Cmd.pipe does.
func cmdStages(g *interp.Goroutine, who string, v value.Value) []*value.Record {
	r, ok := v.(*value.Record)
	switch {
	case ok && r.Tag == CmdTag:
		return []*value.Record{r}
	case ok && r.Tag == PipelineTag:
		cmds := arrayArg(g, who, r.Get(0)).Values()
		if len(cmds) == 0 {
			g.Panic("%s needs a pipeline of at least one command", who)
		}
		stages := make([]*value.Record, len(cmds))
		for i, c := range cmds {
			cr, ok := c.(*value.Record)
			if !ok || cr.Tag != CmdTag {
				g.Panic("%s needs a pipeline of Cmd values, got %s", who, value.Describe(c))
			}
			stages[i] = cr
		}
		return stages
	}
	g.Panic("%s needs a Cmd or a Cmd.Pipeline, got %s", who, value.Describe(v))
	return nil
}

// commands builds the children of a Cmd or a Cmd.Pipeline, and the programs they name.
func commands(g *interp.Goroutine, who string, v value.Value) ([]*exec.Cmd, []string, value.Value) {
	stages := cmdStages(g, who, v)
	cmds := make([]*exec.Cmd, len(stages))
	programs := make([]string, len(stages))
	for i, c := range stages {
		cmd, fail := command(g, who, c)
		if fail != nil {
			return nil, nil, fail
		}
		cmds[i], programs[i] = cmd, stringArg(g, who, c.Get(0))
	}
	return cmds, programs, nil
}

// lockedWriter serializes the writes of several children's copies into one writer.
type lockedWriter struct {
	mu *sync.Mutex
	w  io.Writer
}

func (l lockedWriter) Write(b []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(b)
}

// serialize makes the outputs of every stage that are not files write one at a time, since each
// stage copies into its own from a goroutine of its own.
func serialize(cmds []*exec.Cmd) {
	mu := &sync.Mutex{}
	wrap := func(w io.Writer) io.Writer {
		switch w.(type) {
		case nil, *os.File:
			return w
		}
		return lockedWriter{mu, w}
	}
	for _, c := range cmds {
		c.Stdout, c.Stderr = wrap(c.Stdout), wrap(c.Stderr)
	}
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

// startProcess starts the stages, flushing standard output first, each one's standard output
// joined to the next one's input, and returns the process, with a goroutine that reaps them and
// then ends the copy of feed into the first, when there is one. The caller has set the first
// stage's input, the last one's output, and every standard error. The stages are killed when ctx
// ends, when ctx is not nil, and when a cancellation ends the run. group puts them in a process
// group of their own. When a stage cannot start, those started are killed and reaped.
func startProcess(in *interp.Instance, cmds []*exec.Cmd, programs []string, ctx context.Context, group bool, feed *input) (*Process, value.Value) {
	if ctx != nil && ctx.Err() != nil {
		if feed != nil {
			feed.end()
		}
		return nil, gaveUp(ctx)
	}
	var ends []*os.File // the pipe ends the children hold, closed here once they started
	closeEnds := func() {
		for _, f := range ends {
			f.Close()
		}
	}
	var pw *os.File
	fail := func(err error) (*Process, value.Value) {
		closeEnds()
		if pw != nil {
			pw.Close()
		}
		if feed != nil {
			feed.end()
		}
		return nil, errResult(err)
	}
	if feed != nil {
		pr, w, err := os.Pipe()
		if err != nil {
			return fail(err)
		}
		cmds[0].Stdin, pw = pr, w
		ends = append(ends, pr)
	}
	for i := 0; i < len(cmds)-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			return fail(err)
		}
		cmds[i].Stdout, cmds[i+1].Stdin = w, r
		ends = append(ends, r, w)
	}
	serialize(cmds)
	in.Flush() // what the program printed precedes what the children print
	for i, cmd := range cmds {
		if group {
			leader := 0
			if i > 0 {
				leader = cmds[0].Process.Pid
			}
			setGroup(cmd, leader)
		}
		if err := cmd.Start(); err != nil {
			signalStages(cmds[:i], group, os.Kill) // with their group, what they started too
			waitAll(cmds[:i])
			return fail(err)
		}
	}
	closeEnds()
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
	p := &Process{cmds: cmds, programs: programs, group: group, done: make(chan struct{}), failed: -1}
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
		errs := waitAll(cmds)
		statuses := make([]int, len(cmds))
		var waitErr error
		for i, cmd := range cmds {
			if cmd.ProcessState != nil {
				statuses[i] = exitStatus(cmd.ProcessState)
			}
			var exit *exec.ExitError
			if err := errs[i]; err != nil && waitErr == nil && !errors.As(err, &exit) && !errors.Is(err, exec.ErrWaitDelay) {
				waitErr = err
			}
		}
		stopAfter()
		removeStop()
		endFeed()
		p.mu.Lock()
		p.statuses = statuses
		p.status, p.failed = pipelineStatus(cmds, statuses)
		p.err = waitErr
		p.mu.Unlock()
		close(p.done)
	}()
	return p, nil
}

// waitAll waits for every stage at once, so each one's wait for its copies (WaitDelay) runs
// beside the others' rather than after them, and returns each one's error.
func waitAll(cmds []*exec.Cmd) []error {
	errs := make([]error, len(cmds))
	var wg sync.WaitGroup
	for i, cmd := range cmds {
		wg.Go(func() { errs[i] = cmd.Wait() })
	}
	wg.Wait()
	return errs
}

// childInput gives the first stage of Cmd.start the handle h as its standard input: a standard stream
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

// capture runs a command for Cmd.read, Cmd.readBytes and Cmd.capture, arguments c, input and ctx:
// standard output and error captured, in the program's process group, the call blocking until
// every stage is done. fail is the Err of a stage that could not start or of the end of ctx.
func capture(g *interp.Goroutine, who string, a []value.Value) (p *Process, stdout, stderr []byte, fail value.Value) {
	cmds, programs, fail := commands(g, who, a[0])
	if fail != nil {
		return nil, nil, nil, fail
	}
	var feed *input
	if given(a[1]) {
		feed = &input{r: strings.NewReader(textOrBytes(g, who, a[1]))}
	}
	var out, errOut bytes.Buffer
	cmds[len(cmds)-1].Stdout = &out
	for _, cmd := range cmds {
		cmd.Stderr = &errOut
	}
	p, fail = startProcess(g.Instance(), cmds, programs, waitContext(g, who, a[2]), false, feed)
	if fail != nil {
		return nil, nil, nil, fail
	}
	<-p.done
	g.AfterWait()
	if r := p.outcome(); r.(*value.Record).Tag == value.ErrTag {
		return nil, nil, nil, r
	}
	return p, out.Bytes(), errOut.Bytes(), nil
}

// captureCmd is Cmd.capture: the status, each stage's, and both outputs, a nonzero status being
// Ok.
func captureCmd(g *interp.Goroutine, who string, a []value.Value) value.Value {
	p, stdout, stderr, fail := capture(g, who, a)
	if fail != nil {
		return fail
	}
	statuses := make([]value.Value, len(p.statuses))
	for i, s := range p.statuses {
		statuses[i] = int64(s)
	}
	outText, outBytes := received(stdout)
	errText, errBytes := received(stderr)
	return value.Ok(value.NewRecord(OutputTag, []value.Value{int64(p.status), value.NewArray(statuses), outText, errText, outBytes, errBytes}))
}

// readCmd is Cmd.read: standard output as text without one final newline, and Err on a nonzero
// status, with standard error in the message.
func readCmd(g *interp.Goroutine, who string, a []value.Value) value.Value {
	p, stdout, stderr, fail := capture(g, who, a)
	if fail != nil {
		return fail
	}
	if p.status != 0 {
		return value.Err(value.ErrorValue(value.Text(p.failure(stderr))))
	}
	if bytes.HasSuffix(stdout, []byte("\r\n")) {
		stdout = stdout[:len(stdout)-2]
	} else if bytes.HasSuffix(stdout, []byte("\n")) {
		stdout = stdout[:len(stdout)-1]
	}
	text, _ := received(stdout)
	return value.Ok(text)
}

// readBytesCmd is Cmd.readBytes: standard output as the last stage wrote it, and Err on a
// nonzero status as Cmd.read gives it.
func readBytesCmd(g *interp.Goroutine, who string, a []value.Value) value.Value {
	p, stdout, stderr, fail := capture(g, who, a)
	if fail != nil {
		return fail
	}
	if p.status != 0 {
		return value.Err(value.ErrorValue(value.Text(p.failure(stderr))))
	}
	return value.Ok(bytesValue(stdout))
}

// runCmd is Cmd.run: the stages in the foreground, the first reading the program's standard
// input, the last writing its output, and the host told, so it can hold back its keyboard
// signals until they exit. The last stage's end decides whether a held interrupt is passed on.
func runCmd(g *interp.Goroutine, who string, a []value.Value) value.Value {
	in := g.Instance()
	cmds, programs, fail := commands(g, who, a[0])
	if fail != nil {
		return fail
	}
	cmds[len(cmds)-1].Stdout = in.ChildStdout()
	for _, cmd := range cmds {
		cmd.Stderr = in.ChildStderr()
	}
	file, feed, _ := stdinFor(in)
	if file != nil {
		cmds[0].Stdin = file
	}
	release := in.Foreground()
	p, fail := startProcess(in, cmds, programs, waitContext(g, who, a[1]), false, feed)
	if fail != nil {
		release(nil)
		return fail
	}
	<-p.done
	release(endedBy(p.last().ProcessState))
	in.Flush()
	g.AfterWait()
	return p.outcome()
}

// startCmd is Cmd.start: the stages in a process group of their own, the first reading the
// stdin handle, the last writing stdout, and every one writing stderr; standard input is the
// null device and the outputs are the program's by default.
func startCmd(g *interp.Goroutine, who string, a []value.Value) value.Value {
	in := g.Instance()
	cmds, programs, fail := commands(g, who, a[0])
	if fail != nil {
		return fail
	}
	var ctx context.Context
	if given(a[4]) {
		ctx = contextArg(g, who, a[4]).Ctx
	}
	var feed *input
	undo := func() {
		if feed != nil {
			feed.end()
		}
	}
	if v := a[1]; given(v) {
		file, f, fail := childInput(in, handleArg(g, who, v))
		if fail != nil {
			return fail
		}
		if file != nil && isTerminal(file) {
			return value.Err(value.ErrorValue("a child in the background cannot read the terminal; Cmd.run gives it the terminal"))
		}
		if file != nil {
			cmds[0].Stdin = file
		}
		feed = f
	}
	var stdout, stderr io.Writer
	for i, v := range []value.Value{a[2], a[3]} {
		set, std := &stdout, in.ChildStdout
		if i == 1 {
			set, std = &stderr, in.ChildStderr
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
	cmds[len(cmds)-1].Stdout = stdout
	for _, cmd := range cmds {
		cmd.Stderr = stderr
	}
	p, fail := startProcess(in, cmds, programs, ctx, true, feed)
	if fail != nil {
		return fail
	}
	return value.Ok(p)
}

// waitCmd is Cmd.wait: the outcome once the stages are done, or Err when ctx ends first.
func waitCmd(g *interp.Goroutine, who string, a []value.Value) value.Value {
	p := processArg(g, who, a[0])
	ctx := waitContext(g, who, a[1])
	select {
	case <-p.done:
	case <-ctx.Done():
		if !p.exited() { // a process done by then gives its outcome, whichever select saw first
			g.AfterWait()
			return gaveUp(ctx)
		}
	}
	g.AfterWait()
	return p.outcome()
}

// killCmd is Cmd.kill: the signal named, sent to the stages' group, Ok when they have exited.
func killCmd(g *interp.Goroutine, who string, a []value.Value) value.Value {
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

// stopCmd is Cmd.stop: SIGTERM, SIGKILL when grace passes first, and the outcome.
func stopCmd(g *interp.Goroutine, who string, a []value.Value) value.Value {
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
	if !p.exited() { // the run ended first, which kills the stages; they have no status yet
		return gaveUp(run)
	}
	return p.outcome()
}

// newCmd is Cmd.new, the literal with its defaults.
func newCmd(g *interp.Goroutine, who string, a []value.Value) value.Value {
	defaults := CmdTag.Defaults.(interp.NativeDefaults)
	vals := []value.Value{a[0], a[1], a[2], a[3]}
	for i, v := range vals {
		if v == nil {
			vals[i] = defaults[i]()
		}
	}
	return value.NewRecord(CmdTag, vals)
}

// pipeCmd is Cmd.pipe: the commands as a pipeline, which needs one at least. The pipeline holds
// a copy of the array, so changing the array later does not change it.
func pipeCmd(g *interp.Goroutine, who string, a []value.Value) value.Value {
	cmds := arrayArg(g, who, a[0]).Values()
	if len(cmds) == 0 {
		g.Panic("%s needs at least one command", who)
	}
	return value.NewRecord(PipelineTag, []value.Value{value.NewArray(append([]value.Value(nil), cmds...))})
}

// stringCmd is Cmd.string: the program and its arguments as a POSIX shell would need them
// quoted, and a pipeline's stages joined by " | ". env and dir are left out.
func stringCmd(g *interp.Goroutine, who string, a []value.Value) value.Value {
	var b strings.Builder
	for i, c := range cmdStages(g, who, a[0]) {
		if i > 0 {
			b.WriteString(" | ")
		}
		b.WriteString(shellQuote(stringArg(g, who, c.Get(0)), true))
		for _, v := range arrayArg(g, who, c.Get(1)).Values() {
			b.WriteByte(' ')
			b.WriteString(shellQuote(stringArg(g, who, v), false))
		}
	}
	return value.Text(b.String())
}

// shellQuote is s as one word of a POSIX shell: as it is when no character of it is special,
// and in single quotes otherwise, each single quote in it closing them, escaped, and reopening.
// A program's word quotes = too, which a shell would read as an assignment there.
func shellQuote(s string, program bool) string {
	plain := s != ""
	safe := "_@%+=:,./-"
	if program {
		safe = "_@%+:,./-"
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune(safe, r)) {
			plain = false
			break
		}
	}
	if plain {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
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

// cmdFuncs are the functions of the Cmd type.
func cmdFuncs() []*interp.Builtin {
	return []*interp.Builtin{
		fn("Cmd.new", "program, args?, dir?, env?", newCmd),
		fn("Cmd.pipe", "cmds", pipeCmd),
		fn("Cmd.string", "c", stringCmd),
		fn("Cmd.run", "c, ctx?", runCmd),
		fn("Cmd.read", "c, input?, ctx?", readCmd),
		fn("Cmd.readBytes", "c, input?, ctx?", readBytesCmd),
		fn("Cmd.capture", "c, input?, ctx?", captureCmd),
		fn("Cmd.start", "c, stdin?, stdout?, stderr?, ctx?", startCmd),
		fn("Cmd.wait", "p, ctx?", waitCmd),
		fn("Cmd.kill", "p, signal?", killCmd),
		fn("Cmd.stop", "p, grace?", stopCmd),
		fn("Cmd.pid", "p", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			return int64(processArg(g, who, a[0]).last().Process.Pid)
		}),
	}
}

// processEntries are the Os namespace's entries for pipes and programs.
func processEntries() []any {
	return []any{
		constant("Pipe", PipeTag),
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
