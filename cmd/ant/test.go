package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ant-lang.org"
	"ant-lang.org/ast"
	"ant-lang.org/internal/temporal"
	"ant-lang.org/parser"
	"ant-lang.org/value"
)

// testCommand is ant test (spec/test.md): it finds the test files under its paths, runs each
// as a program of its own, and reports the tests that failed.
func testCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("ant test", flag.ContinueOnError)
	fl.SetOutput(stderr)
	runPat := fl.String("run", "", "run only the tests whose full name matches `regexp`")
	listPat := fl.String("list", "", "list the top-level tests whose name matches `regexp`, and run none")
	skipPat := fl.String("skip", "", "leave out the tests whose full name matches `regexp`")
	verbose := fl.Bool("v", false, "print every test as it starts and ends, with its output")
	jsonOut := fl.Bool("json", false, "print events as JSON lines")
	failfast := fl.Bool("failfast", false, "start no test after the first failure")
	nocheck := fl.Bool("nocheck", false, "run the tests without type-checking the test files first")
	parallel := fl.Int("parallel", runtime.GOMAXPROCS(0), "the most tests that run at once after Test.parallel()")
	timeout := new(time.Duration)
	*timeout = 10 * time.Minute
	fl.Var((*testTimeout)(timeout), "timeout", "end the run after `d`, as Time.parseDuration reads it; 0 means no limit")
	fl.Usage = func() {
		fmt.Fprintln(stderr, "usage: ant test [flags] [path ...]")
		fl.PrintDefaults()
	}
	if err := fl.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	run, err := ant.TestPattern(*runPat)
	if err != nil {
		fmt.Fprintf(stderr, "ant test: -run: %s\n", err)
		return 2
	}
	skip, err := ant.TestPattern(*skipPat)
	if err != nil {
		fmt.Fprintf(stderr, "ant test: -skip: %s\n", err)
		return 2
	}
	var list *regexp.Regexp
	if *listPat != "" {
		re, err := value.NewRegex(*listPat, "")
		if err != nil {
			fmt.Fprintf(stderr, "ant test: -list: %s\n", err)
			return 2
		}
		list = re.Re
	}
	if *parallel < 1 {
		fmt.Fprintln(stderr, "ant test: -parallel must be at least 1")
		return 2
	}
	files, err := testSources.files(fl.Args())
	if err != nil {
		fmt.Fprintf(stderr, "ant test: %s\n", err)
		return 2
	}
	if !*nocheck {
		// A green run means the test files check (test.md section 1): an error runs nothing.
		// warnings are ant check's to report; the test report stays the tests' (test.md section 2)
		if diags, loaded := ant.Check(files); len(ant.Errors(diags)) > 0 || !loaded {
			diags = ant.Errors(diags)
			if *jsonOut {
				writeJSONCheckErrors(stdout, diags)
			} else {
				printDiagnostics(stderr, diags)
			}
			if !loaded {
				return 2
			}
			return 1
		}
	}
	if list != nil {
		return listTests(files, list, *jsonOut, stdout, stderr)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	var timedOut atomic.Bool
	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
		timer := time.AfterFunc(*timeout, func() {
			timedOut.Store(true)
			cancel()
		})
		defer timer.Stop()
	}

	rep := &reporter{w: stdout, verbose: *verbose, json: *jsonOut, output: map[string]*strings.Builder{},
		running: map[string]runningTest{}}
	var halt atomic.Bool
	cfg := ant.TestConfig{Run: run, Skip: skip, Parallel: *parallel, Report: rep.event}
	if *failfast {
		cfg.Stop = &halt
	}
	start := time.Now()
	status := 0
	// A file that cannot be loaded when it is run is reported as ant run reports one, or under
	// -json as "error" events (test.md section 4); the other files still run.
	loadFailed := func(err error) {
		status = 2
		var le *ant.LoadError
		if *jsonOut && errors.As(err, &le) && len(le.Diagnostics) > 0 {
			writeJSONCheckErrors(stdout, le.Diagnostics)
			return
		}
		printLoadError(stderr, err)
	}
	for _, file := range files {
		if ctx.Err() != nil || *failfast && halt.Load() {
			break
		}
		prog, err := ant.Load(file)
		if err != nil {
			loadFailed(err)
			continue
		}
		cfg.File = file
		// Under -json, standard output holds the events alone; what the program prints
		// outside a test goes to standard error.
		progOut := stdout
		if *jsonOut {
			progOut = stderr
		}
		err = prog.Test(ctx, cfg, ant.WithStdin(stdin), ant.WithStdout(progOut), ant.WithStderr(stderr))
		var pe *ant.PanicError
		var ee *ant.ExitError
		switch {
		case err == nil:
		case errors.As(err, &ee):
			return ee.Status
		case errors.As(err, &pe):
			// the runtime has reported the panic of an initializer; the file's tests did not run
			rep.mu.Lock()
			rep.failed++
			rep.mu.Unlock()
			if *failfast {
				halt.Store(true)
			}
		case ctx.Err() != nil:
		default:
			loadFailed(err)
		}
	}
	if timedOut.Load() {
		msg := fmt.Sprintf("test timed out after %s", *timeout)
		fmt.Fprintln(stderr, msg)
		running := rep.stop()
		if len(running) > 0 {
			fmt.Fprintln(stderr, "running tests:")
			for _, t := range running {
				fmt.Fprintf(stderr, "    %s: %s\n", t.file, t.name)
			}
		}
		if *jsonOut {
			rep.timedOut(running, msg, time.Since(start))
		}
		return 1
	}
	if ctx.Err() != nil {
		return 130
	}
	rep.done(time.Since(start), status == 2)
	if (*runPat != "" || *skipPat != "") && status == 0 && rep.passed+rep.failed+rep.skipped == 0 {
		fmt.Fprintln(stderr, "ant test: warning: no tests to run")
	}
	if status == 0 && rep.failed > 0 {
		status = 1
	}
	return status
}

// listTests answers -list: the top-level tests of files whose name matches re, in source order,
// one per line as "file: name", or under -json as "list" events. A file that does not parse lists
// nothing, of any file (test.md section 1).
func listTests(files []string, re *regexp.Regexp, jsonOut bool, stdout, stderr io.Writer) int {
	parsed := make([]*ast.File, len(files))
	status := 0
	for i, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(stderr, "ant test: %s\n", err)
			status = 2
			continue
		}
		if parsed[i], err = parser.Parse(file, string(src), parser.Program); err != nil {
			fmt.Fprintf(stderr, "%s:%s\n", file, err)
			status = 2
		}
	}
	if status != 0 {
		return status
	}
	for i, file := range files {
		for _, t := range parsed[i].Tests {
			name := t.Name.(*ast.StringLit).Value
			if !re.MatchString(name) {
				continue
			}
			if jsonOut {
				b, _ := json.Marshal(jsonEvent{Action: "list", File: file, Test: name})
				stdout.Write(append(b, '\n'))
			} else {
				fmt.Fprintf(stdout, "%s: %s\n", file, name)
			}
		}
	}
	return 0
}

// testSources are the files ant test takes: the test files the paths name, or those under a
// directory; a file named directly that is not a test file is an error.
var testSources = sources{keep: ant.IsTestFile, named: func(path string) error {
	if !ant.IsTestFile(path) {
		return fmt.Errorf("%s is not a test file, whose name ends in _test.ant", path)
	}
	return nil
}}

// reporter writes the report of test.md section 2, or its JSON events.
type reporter struct {
	mu      sync.Mutex
	w       io.Writer
	verbose bool
	json    bool
	// output collects each running test's output, by file and name
	output  map[string]*strings.Builder
	running map[string]runningTest
	// stopped is set when the time ran out: what the tests report after that is dropped
	stopped bool

	passed, failed, skipped int
}

// runningTest is a test that started and has not ended.
type runningTest struct {
	file, name, parent string
	start              time.Time
}

func (r *reporter) event(e ant.TestEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stopped {
		return
	}
	key := e.File + ": " + e.Name
	switch e.Kind {
	case ant.TestRun:
		r.running[key] = runningTest{e.File, e.Name, e.Parent, time.Now()}
	case ant.TestOutput:
	default:
		delete(r.running, key)
		if e.Counted {
			switch e.Kind {
			case ant.TestPass:
				r.passed++
			case ant.TestFail:
				r.failed++
			case ant.TestSkip:
				r.skipped++
			}
		}
	}
	if r.json {
		r.jsonEvent(e)
		return
	}
	switch e.Kind {
	case ant.TestRun:
		if r.verbose {
			fmt.Fprintf(r.w, "=== RUN   %s\n", key)
		}
	case ant.TestOutput:
		if r.verbose {
			r.w.Write([]byte(indent(e.Output)))
			return
		}
		b := r.output[key]
		if b == nil {
			b = &strings.Builder{}
			r.output[key] = b
		}
		b.WriteString(e.Output)
	case ant.TestPass:
		delete(r.output, key)
		if r.verbose {
			fmt.Fprintf(r.w, "--- PASS: %s (%.2fs)\n", key, e.Elapsed.Seconds())
		}
	case ant.TestSkip:
		delete(r.output, key)
		if r.verbose {
			fmt.Fprintf(r.w, "--- SKIP: %s (%.2fs)\n", key, e.Elapsed.Seconds())
			if e.Message != "" {
				r.w.Write([]byte(indent(e.Message + "\n")))
			}
		}
	case ant.TestFail:
		out := r.output[key]
		delete(r.output, key)
		if !r.verbose && e.Message == "" {
			return // failed by its subtests alone, which are reported (test.md section 2)
		}
		fmt.Fprintf(r.w, "--- FAIL: %s (%.2fs)\n", key, e.Elapsed.Seconds())
		if out != nil {
			r.w.Write([]byte(indent(out.String())))
		}
		if e.Message != "" {
			r.w.Write([]byte(indent(e.Message + "\n")))
		}
	}
}

// indent indents every line of s by four spaces, ending the last with a newline.
func indent(s string) string {
	if s == "" {
		return ""
	}
	s = strings.TrimSuffix(s, "\n")
	return "    " + strings.ReplaceAll(s, "\n", "\n    ") + "\n"
}

type jsonEvent struct {
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

func (r *reporter) jsonEvent(e ant.TestEvent) {
	je := jsonEvent{File: e.File, Test: e.Name}
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
	r.writeJSON(je)
}

func (r *reporter) writeJSON(je jsonEvent) {
	b, _ := json.Marshal(je)
	r.w.Write(append(b, '\n'))
}

// done writes the last line, or the JSON done event.
func (r *reporter) done(elapsed time.Duration, broken bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.json {
		secs := elapsed.Seconds()
		r.writeJSON(jsonEvent{Action: "done", Passed: &r.passed, Failed: &r.failed, Skipped: &r.skipped, Elapsed: &secs})
		return
	}
	var parts []string
	if r.failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", r.failed))
	}
	parts = append(parts, fmt.Sprintf("%d passed", r.passed)) // always, so a skip-only run says 0 passed
	if r.skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", r.skipped))
	}
	summary := strings.Join(parts, ", ")
	if r.passed+r.failed+r.skipped == 0 {
		summary = "no tests"
	}
	head := "ok  "
	if r.failed > 0 || broken {
		head = "FAIL"
	}
	fmt.Fprintf(r.w, "%s  %s (%.2fs)\n", head, summary, elapsed.Seconds())
}

// stop drops the events that come after it and lists the tests that started and have not ended,
// by file and name.
func (r *reporter) stop() []runningTest {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stopped = true
	var out []runningTest
	for _, t := range r.running {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].file != out[j].file {
			return out[i].file < out[j].file
		}
		return out[i].name < out[j].name
	})
	return out
}

// timedOut writes, under -json, a "fail" event for each test still running when the time ran
// out, then the "done" event, which counts as failed each of them with no subtest still running
// (test.md section 3).
func (r *reporter) timedOut(running []runningTest, msg string, elapsed time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range running {
		secs := time.Since(t.start).Seconds()
		r.writeJSON(jsonEvent{Action: "fail", File: t.file, Test: t.name, Elapsed: &secs, Message: msg})
		parent := false
		for _, u := range running {
			parent = parent || u.file == t.file && u.parent == t.name
		}
		if !parent {
			r.failed++
		}
	}
	secs := elapsed.Seconds()
	r.writeJSON(jsonEvent{Action: "done", Passed: &r.passed, Failed: &r.failed, Skipped: &r.skipped, Elapsed: &secs})
}

// testTimeout is the -timeout flag, read as Time.parseDuration reads a string: Go's notation, such
// as 30s or 1h30m, or ISO 8601 (test.md).
type testTimeout time.Duration

func (t *testTimeout) String() string { return time.Duration(*t).String() }

func (t *testTimeout) Set(s string) error {
	d, err := temporal.ParseDuration(s)
	if err != nil {
		return err
	}
	span, ok := d.Exact()
	if !ok {
		return errors.New("a timeout cannot have years, months or weeks")
	}
	if span < 0 {
		return errors.New("a timeout cannot be negative")
	}
	*t = testTimeout(span)
	return nil
}
