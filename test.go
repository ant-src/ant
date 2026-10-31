package ant

import (
	"context"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/parser"
)

// TestConfig configures Program.Test.
type TestConfig struct {
	// File labels the events: the test file's path as the runner found it.
	File string
	// Run and Skip hold one regular expression per level of a test's full name (the -run and
	// -skip flags of ant test); nil runs every test and skips none.
	Run, Skip []*regexp.Regexp
	// Parallel is the most tests that run at once after Test.parallel; zero means one.
	Parallel int
	// Stop, when set, is set by the first failure and stops tests from starting. It may be shared
	// by the runs of several files.
	Stop *atomic.Bool
	// Report receives every event, one at a time, from any goroutine.
	Report func(TestEvent)
}

// TestEventKind is what a TestEvent reports.
type TestEventKind int

const (
	TestRun    TestEventKind = iota // a test starts
	TestOutput                      // a test printed Output
	TestPass                        // a test passed
	TestFail                        // a test failed; Message holds its failures, "" when only subtests failed
	TestSkip                        // a test was skipped; Message holds the reason
)

// TestEvent is one event of a test run (test.md).
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

// IsTestFile reports whether path names a test file, whose name ends in _test.ant.
func IsTestFile(path string) bool { return parser.IsTestFile(path) }

// Test runs the tests of a program whose root is a test file (section 10, Tests): it
// initializes the modules, then runs each test, reporting events to cfg.Report. It returns nil
// when the tests ran, whatever their outcomes, a *PanicError when an initializer panicked, an
// *ExitError when the program called Os.exit, and ctx.Err() when the context ended first.
func (p *Program) Test(ctx context.Context, cfg TestConfig, opts ...Option) error {
	report := cfg.Report
	tc := interp.TestConfig{File: cfg.File, Run: cfg.Run, Skip: cfg.Skip, Parallel: cfg.Parallel, Stop: cfg.Stop,
		Report: func(e interp.TestEvent) {
			if report != nil {
				report(TestEvent{Kind: TestEventKind(e.Kind), File: e.File, Name: e.Name, Output: e.Output,
					Message: e.Message, Elapsed: e.Elapsed, Counted: e.Counted})
			}
		}}
	return convertRunError(interp.RunTests(ctx, p.prog, p.cfg.with(opts).run(p.prelude), tc))
}

// TestPattern compiles a -run or -skip pattern of ant test: one regular expression per level of
// a full name, split on each / that is not escaped with a backslash or inside brackets or
// parentheses.
func TestPattern(s string) ([]*regexp.Regexp, error) { return interp.TestPattern(s) }
