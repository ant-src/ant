package tests

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ant-src/ant"
)

// The programs under ../examples are written to be read and run by hand, so they carry no
// trailer: the tests here run each one as a user would, with arguments, standard input,
// environment and files, and for the servers, clients that talk to them.

// run is how one run of an example ended.
type run struct {
	out, err string
	status   int
}

// exampleRun says what to give an example.
type exampleRun struct {
	args  []string
	stdin string
	env   map[string]string
}

// runExample runs examples/<name>.ant to its end.
func runExample(t *testing.T, name string, in exampleRun) run {
	t.Helper()
	prog := loadExample(t, name)
	var stdout, stderr lockedBuffer
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	opts := []ant.Option{ant.WithStdin(strings.NewReader(in.stdin)), ant.WithStdout(&stdout),
		ant.WithStderr(&stderr), ant.WithArgs(in.args)}
	if in.env != nil {
		opts = append(opts, ant.WithEnv(func(name string) (string, bool) { v, ok := in.env[name]; return v, ok }))
	}
	err := prog.Run(ctx, opts...)
	return finish(err, &stdout, &stderr)
}

func loadExample(t *testing.T, name string) *ant.Program {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "examples", name+".ant"))
	if err != nil {
		t.Fatal(err)
	}
	prog, err := ant.Load(abs)
	if err != nil {
		t.Fatal(err)
	}
	return prog
}

func finish(err error, stdout, stderr *lockedBuffer) run {
	switch err.(type) {
	case nil, *ant.PanicError, *ant.ExitError:
	default:
		stderr.WriteString(err.Error() + "\n")
	}
	return run{out: stdout.String(), err: stderr.String(), status: ant.ExitStatus(err)}
}

// server is a server example running in the background.
type server struct {
	addr      string
	interrupt chan struct{}
	done      chan run
}

var listening = regexp.MustCompile(`listening on (\S+)`)

// startExample starts a server example with the port 0 and waits until it prints the address it
// took.
func startExample(t *testing.T, name string) *server {
	t.Helper()
	prog := loadExample(t, name)
	s := &server{interrupt: make(chan struct{}, 1), done: make(chan run, 1)}
	var stdout, stderr lockedBuffer
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	go func() {
		err := prog.Run(ctx, ant.WithStdout(&stdout), ant.WithStderr(&stderr),
			ant.WithArgs([]string{"0"}), ant.WithInterrupts(s.interrupt))
		s.done <- finish(err, &stdout, &stderr)
	}()
	for deadline := time.Now().Add(10 * time.Second); s.addr == ""; {
		select {
		case r := <-s.done:
			t.Fatalf("%s ended before it listened: status %d\n%s%s", name, r.status, r.out, r.err)
		default:
		}
		if m := listening.FindStringSubmatch(stdout.String()); m != nil {
			s.addr = m[1]
		} else if time.Now().After(deadline) {
			t.Fatalf("%s did not say where it listens", name)
		} else {
			time.Sleep(5 * time.Millisecond)
		}
	}
	return s
}

// stop sends the server Ctrl-C and waits for it to end.
func (s *server) stop(t *testing.T) run {
	t.Helper()
	s.interrupt <- struct{}{}
	select {
	case r := <-s.done:
		return r
	case <-time.After(10 * time.Second):
		t.Fatal("the server did not stop on an interrupt")
		return run{}
	}
}

// expect compares a run with the output, standard error substrings and status it should have.
func (r run) expect(t *testing.T, out string, status int, errs ...string) {
	t.Helper()
	if r.status != status {
		t.Errorf("exit status %d, want %d\nstderr:\n%s", r.status, status, r.err)
	}
	if r.out != out {
		t.Errorf("stdout:\n%s\nwant:\n%s", r.out, out)
	}
	for _, e := range errs {
		if !strings.Contains(r.err, e) {
			t.Errorf("stderr lacks %q:\n%s", e, r.err)
		}
	}
}

func lines(ls ...string) string { return strings.Join(ls, "\n") + "\n" }

func TestExampleEcho(t *testing.T) {
	s := startExample(t, "echo")
	type reply struct {
		name string
		got  []string
		err  error
	}
	replies := make(chan reply, 3)
	for _, name := range []string{"ann", "bo", "cy"} {
		go func() {
			r := reply{name: name}
			defer func() { replies <- r }()
			c, err := net.DialTimeout("tcp", s.addr, 5*time.Second)
			if err != nil {
				r.err = err
				return
			}
			defer c.Close()
			c.SetDeadline(time.Now().Add(5 * time.Second))
			in := bufio.NewScanner(c)
			for _, line := range []string{"hello", "bye"} {
				if _, r.err = c.Write([]byte(name + ": " + line + "\n")); r.err != nil {
					return
				}
				if !in.Scan() {
					r.err = in.Err()
					return
				}
				r.got = append(r.got, in.Text())
			}
		}()
	}
	for range 3 {
		r := <-replies
		if r.err != nil {
			t.Errorf("%s: %v", r.name, r.err)
			continue
		}
		up := strings.ToUpper(r.name)
		if want := []string{up + ": HELLO", up + ": BYE"}; strings.Join(r.got, "|") != strings.Join(want, "|") {
			t.Errorf("%s got %q, want %q", r.name, r.got, want)
		}
	}
	// A client still connected when the server stops is closed with it.
	idle, err := net.Dial("tcp", s.addr)
	if err != nil {
		t.Fatal(err)
	}
	defer idle.Close()
	s.stop(t).expect(t, lines("listening on "+s.addr, "stopped"), 0)
}

// call makes one request and returns the status and body.
func call(t *testing.T, method, url, contentType, body string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	client := http.Client{Timeout: 5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, string(data)
}

func TestExampleGuestbook(t *testing.T) {
	s := startExample(t, "guestbook")
	base := "http://" + s.addr
	sign := func(fields url.Values) int {
		status, _ := call(t, "POST", base+"/sign", "application/x-www-form-urlencoded", fields.Encode())
		return status
	}
	if status, body := call(t, "GET", base+"/", "", ""); status != 200 || !strings.Contains(body, "Nobody has signed yet.") {
		t.Errorf("empty book: %d %s", status, body)
	}
	if got := sign(url.Values{"name": {"Ada"}, "message": {"Hello & welcome"}, "website": {"https://ada.dev"}}); got != 303 {
		t.Errorf("signing: %d, want 303", got)
	}
	if got := sign(url.Values{"name": {"<script>alert(1)</script>"}, "message": {"hi"}, "website": {"javascript:alert(1)"}}); got != 303 {
		t.Errorf("signing as a script: %d, want 303", got)
	}
	if got := sign(url.Values{"name": {"  "}, "message": {"no name"}}); got != 400 {
		t.Errorf("signing without a name: %d, want 400", got)
	}
	status, body := call(t, "GET", base+"/", "", "")
	for _, want := range []string{
		"<title>Guestbook (2)</title>",
		`<li class="entry"><b><a href="https://ada.dev">Ada</a></b> wrote: Hello &amp; welcome</li>`,
		`<li class="entry"><b><a href="about:invalid#ant">&lt;script&gt;alert(1)&lt;/script&gt;</a></b> wrote: hi</li>`,
	} {
		if status != 200 || !strings.Contains(body, want) {
			t.Errorf("the page lacks %s:\n%s", want, body)
		}
	}
	if !strings.HasPrefix(body, "<!doctype html><html lang=\"en\">") {
		t.Errorf("the page does not start with its doctype:\n%s", body)
	}
	if status, body := call(t, "GET", base+"/nowhere?x=<b>", "", ""); status != 404 || !strings.Contains(body, "<code>/nowhere</code>") {
		t.Errorf("a missing page: %d %s", status, body)
	}
	s.stop(t).expect(t, lines("listening on "+s.addr), 0)
}

// The examples that run to their end and print, each with the input of its Try line.
func TestExamplePrograms(t *testing.T) {
	for _, c := range []struct {
		name   string
		in     exampleRun
		out    string
		status int
		err    []string
	}{
		{name: "agenda", in: exampleRun{args: []string{"Europe/Stockholm", "2024-10-30"}, stdin: "[{\"title\": \"Standup\", \"when\": \"2024-11-01T09:00\", \"zone\": \"America/New_York\"},\n {\"title\": \"Retro\", \"when\": \"2024-11-05T09:00\", \"zone\": \"America/New_York\"},\n {\"title\": \"Planning\", \"when\": \"2024-11-04T17:00\", \"zone\": \"Asia/Tokyo\"},\n {\"title\": \"Ghost\", \"when\": \"2024-11-02T10:00\", \"zone\": \"Mars/Olympus\"}]\n"}, out: lines("Fri 2024-11-01 14:00 +01:00  Standup (09:00 in America/New_York), in 2 days", "Mon 2024-11-04 09:00 +01:00  Planning (17:00 in Asia/Tokyo), in 5 days", "Tue 2024-11-05 15:00 +01:00  Retro (09:00 in America/New_York), in 6 days"), status: 0, err: []string{"skipped: unknown time zone"}},
		{name: "bank", in: exampleRun{}, out: lines("transfers: 4000", "audits: 2000, mismatches: 0", "total 10000 of 10000, balances [1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000]", "rates loaded 1 time(s), leaders elected 1", "false"), status: 0, err: []string{}},
		{name: "shapes", in: exampleRun{}, out: lines("circle: area 3.142, perimeter 6.283", "rect: area 6.0, perimeter 10.0", "triangle: area 6.0, perimeter 12.0", "rect: area 1.0, perimeter 4.0", "largest: rect", "[\"rect\", \"circle\", \"rect\", \"triangle\"]", "points made: 4 true", "Point{ x: 1.0, y: 2.0 }", "2.23606797749979", "points made: 5", "1"), status: 0, err: []string{}},
		{name: "grep", in: exampleRun{args: []string{"ant", "time"}, stdin: "Ant has time values\nno match here\nTIME and ANT again\n"}, out: lines("1: Ant has time values", "3: TIME and ANT again"), status: 0, err: []string{"2 of 3 lines"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			runExample(t, c.name, c.in).expect(t, c.out, c.status, c.err...)
		})
	}
}

func TestExampleCalc(t *testing.T) {
	in := lines("1 + 2 * 3", "(1 + 2) * 3", "", "-pi * -(e + 1)", "10 / 4", "1 / (pi - pi)",
		"1 +", "2 * (3", "3 $ 4", "z + 1", "1 2", "1.5.2")
	runExample(t, "calc", exampleRun{stdin: in}).expect(t, lines(
		"(1.0 + (2.0 * 3.0)) => 7.0",
		"((1.0 + 2.0) * 3.0) => 9.0",
		"((-pi) * (-(e + 1.0))) => 11.681326876263359",
		"(10.0 / 4.0) => 2.5",
		"(1.0 / (pi - pi)) => error: division by zero",
		"1 + => error: unexpected end of input",
		"2 * (3 => error: expected )",
		"3 $ 4 => error: unexpected character $ at 2",
		"(z + 1.0) => error: unknown variable z",
		"1 2 => error: unexpected 2.0 after the expression",
		"1.5.2 => error: bad number 1.5.2",
	), 0)
}

func TestExampleTodo(t *testing.T) {
	env := exampleRun{env: map[string]string{"TODO_FILE": filepath.Join(t.TempDir(), "todo.json")}}
	for _, step := range []struct {
		command string
		out     string
		status  int
		err     string
	}{
		{"", "nothing to do\n", 0, ""},
		{"add buy milk", "added 1\n", 0, ""},
		{"add write tests", "added 2\n", 0, ""},
		{"add", "", 1, "error: a title is required"},
		{"list", lines("[ ] 1 buy milk", "[ ] 2 write tests"), 0, ""},
		{"done 1", "done 1\n", 0, ""},
		{"done 7", "", 1, "error: no task 7"},
		{"done x", "", 1, "error: x is not an id"},
		{"done", "", 1, "error: an id is required"},
		{"remove 2", "removed 2\n", 0, ""},
		{"list", "[x] 1 buy milk\n", 0, ""},
		{"frob", "", 1, "error: unknown command frob"},
		{"clear", "cleared\n", 0, ""},
		{"list", "nothing to do\n", 0, ""},
	} {
		in := env
		in.args = strings.Fields(step.command)
		r := runExample(t, "todo", in)
		if r.out != step.out || r.status != step.status || !strings.Contains(r.err, step.err) {
			t.Errorf("todo %s: status %d, stdout %q, stderr %q; want %d, %q, %q",
				step.command, r.status, r.out, r.err, step.status, step.out, step.err)
		}
	}
}

func TestExampleFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repo":
			if r.Header.Get("Accept") != "application/json" {
				http.Error(w, "json only", http.StatusNotAcceptable)
				return
			}
			io.WriteString(w, `{"name": "ant", "owner": {"login": "ada"}, "tags": ["lang", "tree"], "stars": 7}`)
		case "/broken":
			io.WriteString(w, `{"name": `)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	repo := srv.URL + "/repo"
	for _, c := range []struct {
		args   []string
		out    string
		status int
		err    string
	}{
		{[]string{repo, "owner.login"}, "ada\n", 0, ""},
		{[]string{repo, "tags.1"}, "tree\n", 0, ""},
		{[]string{repo, "stars"}, "7.0\n", 0, ""},
		{[]string{repo, "owner"}, lines("{", `  "login": "ada"`, "}"), 0, ""},
		{[]string{repo, "owner.name"}, "", 1, "no field name"},
		{[]string{repo, "tags.2"}, "", 1, "no index 2"},
		{[]string{repo, "tags.first"}, "", 1, "first is not an index"},
		{[]string{repo, "stars.count"}, "", 1, "cannot look up count in 7"},
		{[]string{srv.URL + "/nowhere"}, "", 1, "status 404"},
		{[]string{srv.URL + "/broken"}, "", 1, "unexpected end of JSON input"},
		{[]string{"http://127.0.0.1:1/"}, "", 1, "http://127.0.0.1:1/: "},
		{nil, "", 2, "usage:"},
	} {
		runExample(t, "fetch", exampleRun{args: c.args}).expect(t, c.out, c.status, c.err)
	}
}
