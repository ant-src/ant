// Package dap is ant dap, the debug adapter (spec/dap.md): it serves the Debug Adapter Protocol
// on a stream and runs one program under the interpreter's debugger, in its own process.
package dap

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/frame"
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/lib"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/parser"
)

// Serve serves one debug session on in and out until the client disconnects, logging to log.
// It returns the command's exit status (dap.md section 1).
func Serve(ctx context.Context, in io.Reader, out, log io.Writer) int {
	s := &server{r: bufio.NewReader(in), w: out, log: log, lines1: true, cols1: true}
	s.ctx, s.cancel = context.WithCancel(ctx)
	defer s.cancel()
	for {
		body, err := frame.Read(s.r)
		if err != nil {
			if !errors.Is(err, frame.ErrClosed) {
				fmt.Fprintln(log, "ant dap:", err)
			}
			s.end()
			return 1
		}
		var req request
		if err := json.Unmarshal(body, &req); err != nil || req.Type != "request" {
			fmt.Fprintln(log, "ant dap: not a request:", string(body))
			continue
		}
		if s.serve(&req) {
			return 0
		}
	}
}

type request struct {
	Seq       int             `json:"seq"`
	Type      string          `json:"type"`
	Command   string          `json:"command"`
	Arguments json.RawMessage `json:"arguments"`
}

type server struct {
	r   *bufio.Reader
	w   io.Writer
	log io.Writer

	wmu sync.Mutex // one writer at a time
	seq int

	lines1, cols1 bool // the client counts lines and columns from 1

	ctx    context.Context
	cancel context.CancelFunc

	d          *interp.Debugger
	prog       *loader.Program
	cfg        interp.Config
	modules    map[string]*module // by absolute path
	interrupts chan struct{}
	terminated int
	done       chan struct{} // closed when the run has ended
	started    bool

	// handles are the frames and variable containers of the current stop, reset on resume.
	hmu     sync.Mutex
	frames  []interp.StackFrame
	handles []handle
}

// module is one module of the program, with what positions and breakpoints need.
type module struct {
	file  *ast.File
	path  string // the module's path as the interpreter names it
	abs   string
	lines []string
	stops []interp.StopPoint
}

// ---- messages

func (s *server) send(v map[string]any) {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	s.seq++
	v["seq"] = s.seq
	body, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintln(s.log, "ant dap:", err)
		return
	}
	if err := frame.Write(s.w, body); err != nil {
		fmt.Fprintln(s.log, "ant dap:", err)
	}
}

func (s *server) reply(req *request, body any) {
	m := map[string]any{"type": "response", "request_seq": req.Seq, "command": req.Command, "success": true}
	if body != nil {
		m["body"] = body
	}
	s.send(m)
}

func (s *server) fail(req *request, msg string) {
	s.send(map[string]any{"type": "response", "request_seq": req.Seq, "command": req.Command, "success": false, "message": msg})
}

func (s *server) event(name string, body any) {
	m := map[string]any{"type": "event", "event": name}
	if body != nil {
		m["body"] = body
	}
	s.send(m)
}

func (s *server) output(category, text string) {
	s.event("output", map[string]any{"category": category, "output": text})
}

// ---- requests

// serve serves one request and reports whether the session is over.
func (s *server) serve(req *request) bool {
	switch req.Command {
	case "initialize":
		var args struct {
			LinesStartAt1   *bool `json:"linesStartAt1"`
			ColumnsStartAt1 *bool `json:"columnsStartAt1"`
		}
		json.Unmarshal(req.Arguments, &args)
		if args.LinesStartAt1 != nil {
			s.lines1 = *args.LinesStartAt1
		}
		if args.ColumnsStartAt1 != nil {
			s.cols1 = *args.ColumnsStartAt1
		}
		s.reply(req, map[string]any{
			"supportsConfigurationDoneRequest":  true,
			"supportsTerminateRequest":          true,
			"supportsExceptionInfoRequest":      true,
			"supportsEvaluateForHovers":         true,
			"supportsConditionalBreakpoints":    true,
			"supportsHitConditionalBreakpoints": true,
			"supportsLogPoints":                 true,
			"supportsSetVariable":               true,
			"supportsFunctionBreakpoints":       true,
			"exceptionBreakpointFilters": []map[string]any{
				{"filter": "uncaught", "label": "Uncaught panics", "default": true},
				{"filter": "all", "label": "All panics"},
			},
		})
	case "launch":
		if err := s.launch(req.Arguments); err != nil {
			s.fail(req, err.Error())
			return false
		}
		s.reply(req, nil)
		s.event("initialized", nil)
	case "setBreakpoints":
		s.setBreakpoints(req)
	case "setExceptionBreakpoints":
		var args struct {
			Filters []string `json:"filters"`
		}
		json.Unmarshal(req.Arguments, &args)
		var uncaught, all bool
		for _, f := range args.Filters {
			uncaught = uncaught || f == "uncaught"
			all = all || f == "all"
		}
		if s.d != nil {
			s.d.SetPanicFilters(uncaught, all)
		}
		s.reply(req, nil)
	case "configurationDone":
		s.reply(req, nil)
		s.start()
	case "threads":
		threads := []map[string]any{}
		if s.d != nil {
			for _, g := range s.d.Goroutines() {
				threads = append(threads, map[string]any{"id": g.ID, "name": g.Name})
			}
		}
		s.reply(req, map[string]any{"threads": threads})
	case "stackTrace":
		s.stackTrace(req)
	case "scopes":
		s.scopes(req)
	case "variables":
		s.variables(req)
	case "continue":
		s.reply(req, map[string]any{"allThreadsContinued": true})
		s.resume(func() { s.d.Continue() })
	case "next", "stepIn", "stepOut":
		var args struct {
			ThreadID int `json:"threadId"`
		}
		json.Unmarshal(req.Arguments, &args)
		kind := map[string]interp.StepKind{"next": interp.StepOver, "stepIn": interp.StepIn, "stepOut": interp.StepOut}[req.Command]
		s.reply(req, nil)
		s.resume(func() { s.d.Step(args.ThreadID, kind) })
	case "pause":
		var args struct {
			ThreadID int `json:"threadId"`
		}
		json.Unmarshal(req.Arguments, &args)
		s.reply(req, nil)
		if s.d != nil {
			s.d.Pause(args.ThreadID)
		}
	case "exceptionInfo":
		s.exceptionInfo(req)
	case "evaluate":
		s.evaluate(req)
	case "setVariable":
		s.setVariable(req)
	case "setFunctionBreakpoints":
		s.setFunctionBreakpoints(req)
	case "terminate":
		s.reply(req, nil)
		s.terminate()
	case "disconnect":
		s.end()
		s.reply(req, nil)
		return true
	default:
		s.fail(req, fmt.Sprintf("ant dap does not serve %s", req.Command))
	}
	return false
}

// resume runs f, which resumes the program, once the handles of the stop are dropped.
func (s *server) resume(f func()) {
	if s.d == nil {
		return
	}
	s.hmu.Lock()
	s.frames, s.handles = nil, nil
	s.hmu.Unlock()
	f()
}

// launch loads the program (dap.md section 2).
func (s *server) launch(raw json.RawMessage) error {
	var args struct {
		Program     string            `json:"program"`
		Args        []string          `json:"args"`
		Cwd         string            `json:"cwd"`
		Env         map[string]string `json:"env"`
		Stdin       string            `json:"stdin"`
		StopOnEntry bool              `json:"stopOnEntry"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return err
	}
	if s.prog != nil {
		return errors.New("a program is already launched")
	}
	if args.Program == "" {
		return errors.New("launch needs a program")
	}
	// A relative program is resolved against cwd, which defaults to the program's directory
	// (dap.md section 2); without a cwd, a relative program is relative to the adapter's.
	cwd, program := args.Cwd, args.Program
	if cwd == "" {
		program, _ = filepath.Abs(program)
		cwd = filepath.Dir(program)
	} else {
		cwd, _ = filepath.Abs(cwd)
		if !filepath.IsAbs(program) {
			program = filepath.Join(cwd, program)
		}
	}
	if err := os.Chdir(cwd); err != nil {
		return err
	}
	prog, diags := loader.Load(program, loader.Config{Mode: parser.Program})
	if len(diags) > 0 {
		msgs := make([]string, len(diags))
		for i, d := range diags {
			msgs[i] = d.String()
		}
		return errors.New(strings.Join(msgs, "\n"))
	}
	if prog.Root.File.Declarations {
		return fmt.Errorf("%s: a declaration file declares names for the prelude and cannot be run", prog.Root.File.Path)
	}
	var stdin io.Reader
	if args.Stdin != "" {
		f, err := os.Open(args.Stdin)
		if err != nil {
			return err // before anything of the launch is kept, so that the next one can succeed
		}
		stdin = f
	}
	s.prog = prog
	s.modules = map[string]*module{}
	for _, m := range prog.Modules {
		abs, _ := filepath.Abs(m.File.Path)
		src, _ := os.ReadFile(m.File.Path)
		s.modules[abs] = &module{file: m.File, path: m.File.Path, abs: abs, lines: strings.Split(string(src), "\n"), stops: interp.StopPoints(m.File)}
	}

	env := map[string]string{}
	for _, kv := range os.Environ() {
		k, v, _ := strings.Cut(kv, "=")
		env[k] = v
	}
	for k, v := range args.Env {
		env[k] = v
	}
	s.interrupts = make(chan struct{}, 2)
	s.d = interp.NewDebugger()
	s.d.StopOnEntry = args.StopOnEntry
	s.d.OnStop = s.stopped
	s.d.OnLog = func(text string) { s.output("console", text) }
	s.d.OnGoroutine = func(id int, started bool) {
		reason := "exited"
		if started {
			reason = "started"
		}
		s.event("thread", map[string]any{"reason": reason, "threadId": id})
	}
	s.cfg = interp.Config{
		Prelude: lib.Prelude(),
		Stdin:   stdin,
		Stdout:  writer{s, "stdout"},
		Stderr:  writer{s, "stderr"},
		Args:    args.Args,
		Program: args.Program,
		Env: func(name string) (string, bool) {
			v, ok := env[name]
			return v, ok
		},
		Environ: func() []string {
			out := make([]string, 0, len(env))
			for k, v := range env {
				out = append(out, k+"="+v)
			}
			return out
		},
		Interrupts: s.interrupts,
		Debugger:   s.d,
	}
	return nil
}

// writer sends what the program writes as output events.
type writer struct {
	s        *server
	category string
}

func (w writer) Write(p []byte) (int, error) {
	w.s.output(w.category, string(p))
	return len(p), nil
}

// start runs the program, once the client has configured it.
func (s *server) start() {
	if s.prog == nil || s.started {
		return
	}
	s.started = true
	s.done = make(chan struct{})
	go func() {
		tick := time.NewTicker(100 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				s.d.Flush()
			case <-s.done:
				return
			}
		}
	}()
	go func() {
		err := interp.Run(s.ctx, s.prog, s.cfg)
		close(s.done)
		s.event("exited", map[string]any{"exitCode": exitStatus(err)})
		s.event("terminated", nil)
	}()
}

// exitStatus is the status ant run ends with for a run's error.
func exitStatus(err error) int {
	var ee *interp.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &ee):
		return ee.Status
	case errors.Is(err, interp.ErrInterrupted):
		return 130
	}
	return 1
}

// terminate interrupts the program, and cancels the run at the second (dap.md section 2).
func (s *server) terminate() {
	if !s.started {
		return
	}
	s.terminated++
	if s.terminated > 1 {
		s.cancel()
	} else {
		select {
		case s.interrupts <- struct{}{}:
		default:
		}
	}
	s.resume(func() { s.d.Continue() })
}

// end cancels the run, if it started, and waits for it.
func (s *server) end() {
	s.cancel()
	if !s.started {
		return
	}
	s.resume(func() { s.d.Continue() })
	select {
	case <-s.done:
	case <-time.After(2 * time.Second):
	}
}

// stopped reports a stop of the run.
func (s *server) stopped(st interp.Stop) {
	body := map[string]any{"reason": st.Reason, "threadId": st.Goroutine, "allThreadsStopped": true}
	if st.Panic != nil {
		body["description"] = "panic: " + st.Panic.Message
		body["text"] = st.Panic.Message
	}
	s.event("stopped", body)
}

// ---- breakpoints

func (s *server) setBreakpoints(req *request) {
	type sourceBreakpoint struct {
		Line         int    `json:"line"`
		Condition    string `json:"condition"`
		HitCondition string `json:"hitCondition"`
		LogMessage   string `json:"logMessage"`
	}
	var args struct {
		Source struct {
			Path string `json:"path"`
		} `json:"source"`
		Breakpoints []sourceBreakpoint `json:"breakpoints"`
		Lines       []int              `json:"lines"`
	}
	json.Unmarshal(req.Arguments, &args)
	asked := args.Breakpoints
	if asked == nil {
		for _, l := range args.Lines {
			asked = append(asked, sourceBreakpoint{Line: l})
		}
	}
	abs, _ := filepath.Abs(args.Source.Path)
	m := s.modules[abs]
	out := make([]map[string]any, len(asked))
	var set []interp.Breakpoint
	var setAt []int // the index in out of each breakpoint set
	for i, b := range asked {
		line := b.Line
		if !s.lines1 {
			line++
		}
		switch sp, ok := bind(m, line); {
		case m == nil && s.prog == nil:
			out[i] = map[string]any{"verified": false, "line": b.Line, "message": "the program is not launched yet"}
		case m == nil:
			out[i] = map[string]any{"verified": false, "line": b.Line, "message": "the file is not a module of the program"}
		case !ok:
			out[i] = map[string]any{"verified": false, "line": b.Line, "message": "no statement here or below in the same body"}
		default:
			set = append(set, interp.Breakpoint{Pos: sp.Pos, Condition: b.Condition, HitCondition: b.HitCondition, LogMessage: b.LogMessage})
			setAt = append(setAt, i)
			out[i] = map[string]any{"verified": true, "line": s.line(sp.Pos.Line), "column": s.col(m, sp.Pos)}
		}
	}
	if m != nil {
		for k, err := range s.d.SetBreakpoints(m.path, set) {
			if err != nil {
				out[setAt[k]]["verified"] = false
				out[setAt[k]]["message"] = err.Error()
			}
		}
	}
	s.reply(req, map[string]any{"breakpoints": out})
}

func (s *server) setFunctionBreakpoints(req *request) {
	var args struct {
		Breakpoints []struct {
			Name         string `json:"name"`
			Condition    string `json:"condition"`
			HitCondition string `json:"hitCondition"`
		} `json:"breakpoints"`
	}
	json.Unmarshal(req.Arguments, &args)
	out := make([]map[string]any, len(args.Breakpoints))
	var set []interp.FunctionBreakpoint
	var setAt []int
	for i, b := range args.Breakpoints {
		out[i] = map[string]any{"verified": false, "message": fmt.Sprintf("no function %s with a statement in its body", b.Name)}
		for _, m := range s.sortedModules() {
			sp, ok := bindFunction(m, strings.TrimSpace(b.Name))
			if !ok {
				continue
			}
			set = append(set, interp.FunctionBreakpoint{Path: m.path, Breakpoint: interp.Breakpoint{Pos: sp.Pos, Condition: b.Condition, HitCondition: b.HitCondition}})
			setAt = append(setAt, i)
			if out[i]["verified"] == false {
				out[i] = map[string]any{"verified": true, "line": s.line(sp.Pos.Line), "column": s.col(m, sp.Pos), "source": map[string]any{"name": m.path, "path": m.abs}}
			}
		}
	}
	if s.d != nil {
		for k, err := range s.d.SetFunctionBreakpoints(set) {
			if err != nil {
				out[setAt[k]] = map[string]any{"verified": false, "message": err.Error()}
			}
		}
	}
	s.reply(req, map[string]any{"breakpoints": out})
}

// sortedModules lists the program's modules by path, so that responses do not depend on map order.
func (s *server) sortedModules() []*module {
	ms := make([]*module, 0, len(s.modules))
	for _, m := range s.modules {
		ms = append(ms, m)
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].abs < ms[j].abs })
	return ms
}

// bindFunction finds the first stop point of the body of the top-level function name, written
// as name or Type.name, in m.
func bindFunction(m *module, name string) (interp.StopPoint, bool) {
	for _, d := range m.file.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		full := fd.Name
		if fd.Owner != "" {
			full = fd.Owner + "." + fd.Name
		}
		if full != name {
			continue
		}
		for _, sp := range m.stops {
			if sp.Body == ast.Node(fd) {
				return sp, true
			}
		}
	}
	return interp.StopPoint{}, false
}

// bind finds the stop point a breakpoint on line binds to: the first that starts on the line,
// or else on the next line that has one, in the body the line is in (dap.md section 4).
func bind(m *module, line int) (interp.StopPoint, bool) {
	if m == nil {
		return interp.StopPoint{}, false
	}
	for _, sp := range m.stops {
		if sp.Pos.Line == line {
			return sp, true
		}
	}
	// The innermost body that spans the line.
	var body ast.Node
	size := 0
	for _, sp := range m.stops {
		if n := spanOf(sp.Body, len(m.lines)); n > 0 && spans(sp.Body, line) && (body == nil || n < size) {
			body, size = sp.Body, n
		}
	}
	for _, sp := range m.stops {
		if sp.Pos.Line > line && sp.Body == body {
			return sp, true
		}
	}
	return interp.StopPoint{}, false
}

// spanOf is how many lines n spans; a file spans them all, and one more.
func spanOf(n ast.Node, lines int) int {
	if _, file := n.(*ast.File); file {
		return lines + 1
	}
	return n.EndPosition().Line - ast.Start(n).Line + 1
}

func spans(n ast.Node, line int) bool {
	if _, file := n.(*ast.File); file {
		return true
	}
	return ast.Start(n).Line <= line && line <= n.EndPosition().Line
}

// ---- positions

func (s *server) line(l int) int {
	if s.lines1 {
		return l
	}
	return l - 1
}

// col converts a byte column to the client's UTF-16 column.
func (s *server) col(m *module, p ast.Pos) int {
	c := p.Col
	if m != nil && p.Line-1 < len(m.lines) {
		line := m.lines[p.Line-1]
		if b := p.Col - 1; b <= len(line) {
			c = 1
			for _, r := range line[:b] {
				c++
				if r > 0xFFFF {
					c++
				}
			}
		}
	}
	if s.cols1 {
		return c
	}
	return c - 1
}

func (s *server) moduleOf(path string) *module {
	abs, _ := filepath.Abs(path)
	return s.modules[abs]
}
