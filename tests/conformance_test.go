// Package tests runs the programs under run/ through the library and compares what they do
// with what their trailers expect.
//
// Each program under run/ is a complete Ant program followed by a trailer that starts with a
// line reading "// expect" and continues with lines of these forms:
//
//	// out: <line>     one expected line of standard output, in order; "// out:" alone is an
//	                   empty line
//	// exit: <n>       the expected exit status, as the ant command would report it; default 0
//	// err: <text>     standard error must contain this text; may repeat
//	// args: a b c     command-line arguments, split on spaces; $TMP is replaced by a fresh
//	                   temporary directory and $HTTP by the URL of a test HTTP server that
//	                   echoes requests
//	// stdin: <line>   one line of standard input, in order; the program reads them through
//	                   Os.stdin()
//	// os: <goos>      the program runs only on this operating system, as runtime.GOOS names it,
//	                   and is skipped elsewhere; for host behavior such as Linux's /dev/full
//
// Files under run/lib are modules the programs import, not tests. Every program under run/ must
// also pass ant check, which is tested once the checker exists; programs under unchecked/ are
// the kind panics the checker would reject, and are only run.
package tests

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ant-src/ant"
)

type expectation struct {
	out   []string
	exit  int
	err   []string
	args  []string
	stdin string
	os    string
}

func parseTrailer(t *testing.T, path, src string) expectation {
	t.Helper()
	i := strings.LastIndex(src, "\n// expect\n")
	if i < 0 {
		t.Fatalf("%s: no \"// expect\" trailer", path)
	}
	var e expectation
	for _, line := range strings.Split(strings.TrimRight(src[i+len("\n// expect\n"):], "\n"), "\n") {
		key, rest, _ := strings.Cut(line, ":")
		rest = strings.TrimPrefix(rest, " ")
		switch key {
		case "// out":
			e.out = append(e.out, rest)
		case "// exit":
			n, err := strconv.Atoi(strings.TrimSpace(rest))
			if err != nil {
				t.Fatalf("%s: bad exit line %q", path, line)
			}
			e.exit = n
		case "// err":
			e.err = append(e.err, strings.TrimSpace(rest))
		case "// args":
			e.args = strings.Fields(rest)
		case "// stdin":
			e.stdin += rest + "\n"
		case "// os":
			e.os = strings.TrimSpace(rest)
		default:
			t.Fatalf("%s: unexpected trailer line %q", path, line)
		}
	}
	return e
}

func TestRun(t *testing.T) {
	paths, err := filepath.Glob("run/*.ant")
	if err != nil {
		t.Fatal(err)
	}
	unchecked, err := filepath.Glob("unchecked/*.ant")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, unchecked...)
	syntax, err := filepath.Glob("syntax/*.ant")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, syntax...)
	if len(paths) == 0 {
		t.Fatal("no programs under run/")
	}
	for _, path := range paths {
		t.Run(testName(path), func(t *testing.T) {
			t.Parallel()
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want := parseTrailer(t, path, string(src))
			if want.os != "" && want.os != runtime.GOOS {
				t.Skipf("runs only on %s", want.os)
			}
			for i, a := range want.args {
				switch a {
				case "$TMP":
					want.args[i] = t.TempDir()
				case "$HTTP":
					want.args[i] = echoServer(t)
				}
			}
			abs, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			var stdout, stderr lockedBuffer
			prog, err := ant.Load(abs)
			if err == nil {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				err = prog.Run(ctx, ant.WithStdin(strings.NewReader(want.stdin)),
					ant.WithStdout(&stdout), ant.WithStderr(&stderr), ant.WithArgs(want.args))
			}
			// The command prints load errors and unexpected errors itself; mirror that.
			switch err.(type) {
			case nil, *ant.PanicError, *ant.ExitError:
			default:
				stderr.WriteString(err.Error() + "\n")
			}
			status := ant.ExitStatus(err)
			if status != want.exit {
				t.Errorf("exit status %d, want %d\nstderr:\n%s", status, want.exit, stderr.String())
			}
			got := strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n")
			if stdout.Len() == 0 {
				got = nil
			}
			if strings.Join(got, "\n") != strings.Join(want.out, "\n") {
				t.Errorf("stdout:\n%s\nwant:\n%s", stdout.String(), strings.Join(want.out, "\n"))
			}
			for _, text := range want.err {
				if !strings.Contains(stderr.String(), text) {
					t.Errorf("stderr lacks %q:\n%s", text, stderr.String())
				}
			}
		})
	}
}

// testName names a program's subtest: its file name, prefixed by its directory unless that is
// run/.
func testName(path string) string {
	name := strings.TrimSuffix(filepath.Base(path), ".ant")
	dir := filepath.Base(filepath.Dir(path))
	if dir == "run" {
		return name
	}
	return dir + "/" + name
}

// lockedBuffer is a bytes.Buffer safe for concurrent writers.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) WriteString(s string) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.WriteString(s)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func (b *lockedBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Len()
}

// echoServer serves a small API for the Http programs: GET /hello answers "hello", GET
// /status/404 answers that status, /echo with any method answers the request body with the
// method in X-Method and the request's X-Token header copied into the response, and /slow
// answers only once the client has given up, or after three seconds, for the timeout.
func echoServer(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/hello":
			w.Header().Set("Content-Type", "text/plain")
			io.WriteString(w, "hello")
		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/status/"):
			code, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/status/"))
			w.WriteHeader(code)
			io.WriteString(w, "status")
		case r.URL.Path == "/echo":
			w.Header().Set("X-Method", r.Method)
			w.Header().Set("X-Token", r.Header.Get("X-Token"))
			body, _ := io.ReadAll(r.Body)
			w.Write(body)
		case r.URL.Path == "/slow":
			select {
			case <-r.Context().Done():
			case <-time.After(3 * time.Second):
			}
			io.WriteString(w, "slow")
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}
