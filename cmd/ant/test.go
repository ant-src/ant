package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/value"
)

// testCommand is ant test (spec/test.md): it finds the test files under its paths, runs each
// as a program of its own, and reports the tests that failed.
func testCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("ant test", flag.ContinueOnError)
	fl.SetOutput(stderr)
	runPat := fl.String("run", "", "run only the tests whose full name matches `regexp`")
	skipPat := fl.String("skip", "", "leave out the tests whose full name matches `regexp`")
	verbose := fl.Bool("v", false, "print every test as it starts and ends, with its output")
	jsonOut := fl.Bool("json", false, "print events as JSON lines")
	failfast := fl.Bool("failfast", false, "start no test after the first failure")
	parallel := fl.Int("parallel", runtime.NumCPU(), "the most tests that run at once after Test.parallel()")
	timeout := fl.Duration("timeout", 10*time.Minute, "end the run after `d`; 0 means no limit")
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
	run, err := namePattern(*runPat)
	if err != nil {
		fmt.Fprintf(stderr, "ant test: -run: %s\n", err)
		return 2
	}
	skip, err := namePattern(*skipPat)
	if err != nil {
		fmt.Fprintf(stderr, "ant test: -skip: %s\n", err)
		return 2
	}
	if *parallel < 1 {
		fmt.Fprintln(stderr, "ant test: -parallel must be at least 1")
		return 2
	}
	files, err := testFiles(fl.Args())
	if err != nil {
		fmt.Fprintf(stderr, "ant test: %s\n", err)
		return 2
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
		running: map[string]time.Time{}}
	var halt atomic.Bool
	cfg := ant.TestConfig{Run: run, Skip: skip, Parallel: *parallel, Report: rep.event}
	if *failfast {
		cfg.Stop = &halt
	}
	start := time.Now()
	status := 0
	for _, file := range files {
		if ctx.Err() != nil || *failfast && halt.Load() {
			break
		}
		prog, err := ant.Load(file)
		if err != nil {
			printLoadError(stderr, err)
			status = 2
			continue
		}
		cfg.File = file
		err = prog.Test(ctx, cfg, ant.WithStdin(stdin), ant.WithStdout(stdout), ant.WithStderr(stderr))
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
			fmt.Fprintln(stderr, err)
			status = 2
		}
	}
	if timedOut.Load() {
		fmt.Fprintf(stderr, "test timed out after %s\n", *timeout)
		if names := rep.stillRunning(); len(names) > 0 {
			fmt.Fprintln(stderr, "running tests:")
			for _, n := range names {
				fmt.Fprintf(stderr, "    %s\n", n)
			}
		}
		return 1
	}
	if ctx.Err() != nil {
		return 130
	}
	rep.done(time.Since(start), status == 2)
	if status == 0 && rep.failed > 0 {
		status = 1
	}
	return status
}

// namePattern splits a -run or -skip pattern on / into one regular expression per level.
func namePattern(s string) ([]*regexp.Regexp, error) {
	if s == "" {
		return nil, nil
	}
	var out []*regexp.Regexp
	for _, part := range strings.Split(s, "/") {
		re, err := value.NewRegex(part, "")
		if err != nil {
			return nil, err
		}
		out = append(out, re.Re)
	}
	return out, nil
}

// testFiles lists the test files the paths name: a test file itself, or every test file under a
// directory, in lexical order, leaving out directories whose names begin with a dot.
func testFiles(paths []string) ([]string, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if !ant.IsTestFile(p) {
				return nil, fmt.Errorf("%s is not a test file, whose name ends in _test.ant", p)
			}
			files = append(files, p)
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != p && strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if ant.IsTestFile(d.Name()) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

// reporter writes the report of test.md section 2, or its JSON events.
type reporter struct {
	mu      sync.Mutex
	w       io.Writer
	verbose bool
	json    bool
	// output collects each running test's output, by file and name
	output  map[string]*strings.Builder
	running map[string]time.Time

	passed, failed, skipped int
}

func (r *reporter) event(e ant.TestEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := e.File + ": " + e.Name
	switch e.Kind {
	case ant.TestRun:
		r.running[key] = time.Now()
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
		if !r.verbose && e.Message == "" && (out == nil || out.Len() == 0) {
			return // failed by a subtest, which is reported
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
	if r.passed > 0 || r.failed > 0 || r.skipped == 0 && r.passed == 0 {
		parts = append(parts, fmt.Sprintf("%d passed", r.passed))
	}
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

// stillRunning lists the tests that started and have not ended.
func (r *reporter) stillRunning() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var names []string
	for n := range r.running {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
