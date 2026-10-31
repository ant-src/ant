package interp

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/value"
)

// The test runner of section 10, Tests: a test file's top-level tests and their subtests, each
// body on a goroutine of its own, the goroutines a test owns, and the events test.md reports.

// TestConfig configures RunTests.
type TestConfig struct {
	// File labels the events: the test file's path as the runner found it.
	File string
	// Run and Skip are the -run and -skip patterns, one regular expression per level of a
	// full name; nil runs every test and skips none.
	Run, Skip []*regexp.Regexp
	// Parallel is the most tests that run at once after Test.parallel; zero means one.
	Parallel int
	// Stop, when set, is set by the first failure and stops tests from starting (-failfast).
	// It may be shared by the runs of several files.
	Stop *atomic.Bool
	// Report receives every event, from any goroutine, one at a time.
	Report func(TestEvent)
}

// TestEventKind is what a TestEvent reports.
type TestEventKind int

const (
	TestRun    TestEventKind = iota // a test starts
	TestOutput                      // a test printed Output
	TestPass
	TestFail // Message holds the failures, one report each, or "" for one that failed by its subtests alone
	TestSkip // Message holds the reason
)

// TestEvent is one event of a test's run.
type TestEvent struct {
	Kind    TestEventKind
	File    string
	Name    string // the full name, as "parse/0x10"
	Output  string
	Message string
	Elapsed time.Duration
	// Counted is set on an outcome a summary counts: a test without subtests, or one with
	// subtests that failed on its own.
	Counted bool
}

// testFile is the state the tests of one file share.
type testFile struct {
	in   *Instance
	cfg  TestConfig
	sem  chan struct{}
	repo sync.Mutex // Report is called under it
}

func (tf *testFile) report(e TestEvent) {
	e.File = tf.cfg.File
	tf.repo.Lock()
	defer tf.repo.Unlock()
	if tf.cfg.Report != nil {
		tf.cfg.Report(e)
	}
}

// testRun is one test as it runs, or the file itself, the parent of its top-level tests.
type testRun struct {
	file   *testFile
	parent *testRun
	name   string // its own name; "" for the file
	full   string
	level  int // 0 for a top-level test
	start  time.Time

	mu       sync.Mutex
	failures []string
	self     bool // failed on its own, not only by a subtest
	failed   bool
	skipped  string // the reason, once skipped
	isSkip   bool
	subs     map[string]int
	hasSubs  bool
	owned    map[*Goroutine]bool // the goroutines it owns still running
	tempDirs []string

	body     *Goroutine     // the goroutine running its body
	ctx      *value.Context // Test.context, done when the body ends
	kill     context.Context
	killStop context.CancelFunc
	killed   atomic.Bool

	parallel     bool
	wentParallel chan struct{}  // closed when the test calls Test.parallel
	release      chan struct{}  // closed when its sequential subtests have ended: parallel ones may run
	children     sync.WaitGroup // its parallel subtests
}

// skipSignal unwinds a goroutine that called Test.skip.
type skipSignal struct{}

// RunTests runs a test file's program: it initializes the modules, then runs the root's tests
// in source order. It returns nil when the tests ran, whatever their outcomes; a *PanicError
// when an initializer panicked; an *ExitError when the program called Os.exit; and the
// context's error when ctx ended the run.
func RunTests(ctx context.Context, prog *loader.Program, cfg Config, tc TestConfig) error {
	in := New(ctx, prog, cfg)
	if tc.Parallel < 1 {
		tc.Parallel = 1
	}
	tf := &testFile{in: in, cfg: tc, sem: make(chan struct{}, tc.Parallel)}
	g := in.newGoroutine(true)
	if g.d != nil {
		g.d.register(g)
	}
	in.active.Add(1)
	go func() {
		defer in.active.Done()
		defer in.recoverRoot(g)
		if g.d != nil {
			g.enterAnt()
		}
		in.initModules(g)
		root := in.modules[prog.Root.Index]
		file := &testRun{file: tf, level: -1, release: make(chan struct{})}
		for _, t := range prog.Root.File.Tests {
			g.runTest(file, t, t.Name.(*ast.StringLit).Value, root.root)
		}
		close(file.release)
		file.children.Wait()
	}()
	return in.wait()
}

// runTest runs test t, a child of parent named name, with env as its body's enclosing scope,
// and returns when it has ended or called Test.parallel.
func (g *Goroutine) runTest(parent *testRun, node *ast.Test, name string, env *Env) {
	tf := parent.file
	if stop := tf.cfg.Stop; stop != nil && stop.Load() {
		return
	}
	full := name
	if parent.full != "" {
		full = parent.full + "/" + name
	}
	t := &testRun{file: tf, parent: parent, name: name, full: full, level: parent.level + 1,
		owned: map[*Goroutine]bool{}, wentParallel: make(chan struct{}), release: make(chan struct{})}
	if !t.selected() {
		return
	}
	parent.mu.Lock()
	parent.hasSubs = true
	parent.mu.Unlock()
	in := g.in
	t.kill, t.killStop = context.WithCancel(in.ctx)
	t.ctx = value.NewContext(in.ctx, 0)
	t.start = time.Now()
	tf.report(TestEvent{Kind: TestRun, Name: full})
	done := make(chan struct{})
	ng := in.newGoroutine(false)
	ng.test, ng.startPath, ng.startPos = t, env.mod.path, node.Pos
	t.body = ng
	if ng.d != nil {
		ng.d.register(ng)
	}
	in.active.Add(1)
	go func() {
		defer in.active.Done()
		defer close(done)
		if ng.d != nil {
			defer ng.d.unregister(ng)
			ng.enterAnt()
		}
		t.runBody(ng, node, env)
		if t.parallel {
			<-tf.sem
		}
		close(t.release)
		t.children.Wait()
		t.end()
		if t.parallel {
			parent.children.Done()
		}
	}()
	select {
	case <-done:
	case <-t.wentParallel:
	case <-in.ctx.Done():
		panic(&cancelSignal{})
	}
}

// selected reports whether the -run and -skip patterns let t run.
func (t *testRun) selected() bool {
	names := t.names()
	if run := t.file.cfg.Run; t.level < len(run) && !run[t.level].MatchString(t.name) {
		return false
	}
	skip := t.file.cfg.Skip
	if len(skip) == 0 || len(names) < len(skip) {
		return true
	}
	for i, re := range skip {
		if !re.MatchString(names[i]) {
			return true
		}
	}
	return false
}

// names lists the names from the top-level test down to t.
func (t *testRun) names() []string {
	var out []string
	for x := t; x != nil && x.level >= 0; x = x.parent {
		out = append(out, x.name)
	}
	slices.Reverse(out)
	return out
}

// runBody runs the test's body on g, as Panic.catch runs a function (section 12, Running a
// test): a panic that leaves it is the test's failure, and an Err it returns is too.
func (t *testRun) runBody(g *Goroutine, node *ast.Test, env *Env) {
	in := g.in
	defer func() {
		t.ctx.Cancel(value.ErrCancelled)
		r := recover()
		switch x := r.(type) {
		case nil:
		case *AntPanic:
			t.fail(true, panicReport(x))
		case *skipSignal:
		case *exitSignal:
			in.finish(&ExitError{Status: x.status})
		case *cancelSignal:
		default:
			panic(r)
		}
	}()
	g.push(`test "`+t.name+`"`, env.mod.path, node.Pos)
	benv := newEnv(&node.Scope.BlockScope, env)
	if g.d != nil {
		benv.frame = g.stack[len(g.stack)-1].id
	}
	v, c := g.evalBody(node.Body, benv, func() { t.ctx.Cancel(value.ErrCancelled) })
	if c != ctlReturn {
		return
	}
	if rec, ok := v.(*value.Record); ok && rec.Tag == value.ErrTag {
		t.fail(true, errorMessage(rec.Get(0)))
	}
}

// evalBody runs a test's body block in env, calling before as the block exits and before its
// deferred blocks run, whether it ends normally, by return, or by a panic.
func (g *Goroutine) evalBody(b *ast.Block, env *Env, before func()) (v value.Value, c ctl) {
	if b.Scope != nil {
		env = newEnv(b.Scope, env)
	}
	if len(b.Types) > 0 {
		g.in.bindLocalTypes(b, env)
	}
	frame := &deferFrame{}
	g.defers = append(g.defers, frame)
	defer func() {
		g.defers = g.defers[:len(g.defers)-1]
		before()
		r := recover()
		if _, exiting := r.(*exitSignal); exiting {
			panic(r)
		}
		for i := len(frame.list) - 1; i >= 0; i-- {
			if p := g.runDeferred(frame.list[i]); p != nil {
				if _, exiting := p.(*exitSignal); exiting {
					panic(p)
				}
				r = p
			}
		}
		if r != nil {
			panic(r)
		}
	}()
	return g.evalStmts(b.Stmts, env)
}

// errorMessage renders the error of an Err as a script's is written: an Error as its message,
// anything else through String.from (section 10, Scripts).
func errorMessage(err value.Value) string {
	if r, ok := err.(*value.Record); ok && r.Tag == value.ErrorTag {
		if m, ok := r.Get(0).(string); ok {
			return m
		}
	}
	return value.Render(err)
}

// panicReport is a panic as the runtime reports it, without the "panic: " prefix.
func panicReport(p *AntPanic) string {
	var sb strings.Builder
	sb.WriteString(p.Message)
	for i, f := range p.Trace {
		if n := len(p.Trace); n > 2*traceEnds+1 && i >= traceEnds && i < n-traceEnds {
			if i == traceEnds {
				fmt.Fprintf(&sb, "\n  ... %d more frames", n-2*traceEnds)
			}
			continue
		}
		fmt.Fprintf(&sb, "\n  at %s (%s:%d:%d)", f.Function, f.Path, f.Pos.Line, f.Pos.Col)
	}
	return sb.String()
}

// fail records a failure of t: self for one of its own, not only a subtest's. The parents fail
// with it.
func (t *testRun) fail(self bool, msg string) {
	t.mu.Lock()
	t.failed = true
	if self {
		t.self = true
		t.failures = append(t.failures, msg)
	}
	t.mu.Unlock()
	if stop := t.file.cfg.Stop; stop != nil {
		stop.Store(true)
	}
}

// subtestFailed marks the parents of a failed test failed, up to the file.
func (t *testRun) subtestFailed() {
	for p := t.parent; p != nil && p.level >= 0; p = p.parent {
		p.mu.Lock()
		p.failed = true
		p.mu.Unlock()
	}
}

// grace is how long a test waits, after its body, for the goroutines it owns.
const testGrace = 100 * time.Millisecond

// end finishes t after its body, its deferred blocks, and its parallel subtests: it waits the
// grace for the goroutines t owns, ends those that outlive it, removes its temporary
// directories, and reports the outcome.
func (t *testRun) end() {
	deadline := time.Now().Add(testGrace)
	for {
		t.mu.Lock()
		n := len(t.owned)
		t.mu.Unlock()
		if n == 0 || time.Now().After(deadline) || t.file.in.ctx.Err() != nil {
			break
		}
		time.Sleep(time.Millisecond)
	}
	t.mu.Lock()
	var leaks []*Goroutine
	for og := range t.owned {
		leaks = append(leaks, og)
	}
	t.mu.Unlock()
	if len(leaks) > 0 && t.file.in.ctx.Err() == nil {
		slices.SortFunc(leaks, func(a, b *Goroutine) int {
			if a.startPos.Line != b.startPos.Line {
				return a.startPos.Line - b.startPos.Line
			}
			return a.startPos.Col - b.startPos.Col
		})
		for _, og := range leaks {
			t.fail(true, fmt.Sprintf("goroutine started at %s:%d:%d outlived the test", og.startPath, og.startPos.Line, og.startPos.Col))
		}
	}
	t.killed.Store(true)
	t.killStop()
	for _, dir := range t.tempDirs {
		os.RemoveAll(dir)
	}
	if t.file.in.ctx.Err() != nil {
		return // the run ended: the test has no outcome
	}
	t.mu.Lock()
	e := TestEvent{Name: t.full, Elapsed: time.Since(t.start), Counted: !t.hasSubs || t.self}
	switch {
	case t.failed:
		e.Kind = TestFail
		e.Message = strings.Join(t.failures, "\n")
	case t.isSkip:
		e.Kind, e.Message = TestSkip, t.skipped
	default:
		e.Kind = TestPass
	}
	t.mu.Unlock()
	if e.Kind == TestFail {
		t.subtestFailed()
	}
	t.file.report(e)
}

// output collects what a goroutine of t printed.
func (t *testRun) output(s string) {
	if s == "" {
		return
	}
	t.file.report(TestEvent{Kind: TestOutput, Name: t.full, Output: s})
}

// subtest runs the subtest statement e on g, the body goroutine of the current test.
func (g *Goroutine) subtest(e *ast.Test, env *Env) {
	t := g.test
	if t == nil || t.body != g {
		g.panicf(e.Pos, "a subtest runs only on its test's own goroutine")
	}
	v, c := g.eval(e.Name, env)
	if c != ctlNone {
		return
	}
	name, ok := v.(string)
	if !ok {
		g.panicf(e.Name.Position(), "a test's name must be a string, got %s", value.Describe(v))
	}
	if name == "" {
		g.panicf(e.Name.Position(), "a test's name cannot be empty")
	}
	t.mu.Lock()
	if t.subs == nil {
		t.subs = map[string]int{}
	}
	t.subs[name]++
	if n := t.subs[name]; n > 1 {
		name = fmt.Sprintf("%s #%d", name, n)
	}
	t.mu.Unlock()
	g.runTest(t, e, name, env)
}

// ---- goroutines a test owns

// own records that g, just started by a goroutine of t, belongs to t.
func (t *testRun) own(g *Goroutine) {
	t.mu.Lock()
	t.owned[g] = true
	t.mu.Unlock()
}

// disown records that g, a goroutine t owns, has ended.
func (t *testRun) disown(g *Goroutine) {
	t.mu.Lock()
	delete(t.owned, g)
	t.mu.Unlock()
}

// ownedEnded handles what ended a goroutine a test owns, other than its body: a panic fails the
// test, Test.skip skips it, and the test's end or the run's end ends it quietly. It reports
// whether r was handled.
func (t *testRun) ownedEnded(g *Goroutine, r any) bool {
	switch x := r.(type) {
	case *AntPanic:
		t.fail(true, fmt.Sprintf("goroutine started at %s:%d:%d panicked: %s", g.startPath, g.startPos.Line, g.startPos.Col, panicReport(x)))
		return true
	case *skipSignal:
		t.skip("")
		return true
	case *cancelSignal:
		return true
	}
	return false
}

func (t *testRun) skip(reason string) {
	t.mu.Lock()
	if !t.isSkip {
		t.isSkip, t.skipped = true, reason
	}
	t.mu.Unlock()
}

// ---- the Test namespace

// currentTest returns the test the goroutine belongs to, or panics as the Test functions do
// outside a test.
func (g *Goroutine) currentTest(who string) *testRun {
	if g.test == nil {
		g.Panic("%s: not in a test", who)
	}
	return g.test
}

// TestSkip ends the current test as skipped (Test.skip).
func (g *Goroutine) TestSkip(who, reason string) {
	t := g.currentTest(who)
	t.skip(reason)
	panic(&skipSignal{})
}

// TestParallel makes the current test a parallel one (Test.parallel): it returns once the
// parent's sequential tests have ended and a slot of -parallel is free.
func (g *Goroutine) TestParallel(who string) {
	t := g.currentTest(who)
	if t.body != g {
		g.Panic("%s: called from a goroutine the test started, not from its body", who)
	}
	if t.parallel {
		g.Panic("%s: called twice", who)
	}
	t.parallel = true
	t.parent.children.Add(1)
	close(t.wentParallel)
	done := g.in.ctx.Done()
	select {
	case <-t.parent.release:
	case <-done:
		panic(&cancelSignal{})
	}
	select {
	case t.file.sem <- struct{}{}:
	case <-done:
		panic(&cancelSignal{})
	}
	t.mu.Lock()
	t.start = time.Now() // a parallel test's time counts from when it may run
	t.mu.Unlock()
}

// TestTempDir makes a directory removed when the current test ends (Test.tempDir).
func (g *Goroutine) TestTempDir(who string) string {
	t := g.currentTest(who)
	dir, err := os.MkdirTemp("", "ant-test-")
	if err != nil {
		g.Panic("%s: %s", who, err)
	}
	t.mu.Lock()
	t.tempDirs = append(t.tempDirs, dir)
	t.mu.Unlock()
	return dir
}

// TestContext returns the current test's context, done when its body ends (Test.context).
func (g *Goroutine) TestContext(who string) *value.Context {
	return g.currentTest(who).ctx
}

// TestName returns the current test's full name (Test.name).
func (g *Goroutine) TestName(who string) string {
	return g.currentTest(who).full
}
