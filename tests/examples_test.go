package tests

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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

func TestExampleApi(t *testing.T) {
	s := startExample(t, "api")
	base := "http://" + s.addr
	for _, c := range []struct {
		method, path, body string
		status             int
		reply              string
	}{
		{"GET", "/health", "", 200, `{"ok": true}`},
		{"GET", "/items", "", 200, `[]`},
		{"POST", "/items", `{"name": "ant"}`, 201, `{"id":1,"name":"ant"}`},
		{"POST", "/items", `{"name": "bee"}`, 201, `{"id":2,"name":"bee"}`},
		{"POST", "/items", `{"name": 7}`, 400, `{"error": "a name is required"}`},
		{"POST", "/items", `not json`, 400, ``},
		{"GET", "/items", "", 200, `[{"id":1,"name":"ant"},{"id":2,"name":"bee"}]`},
		{"GET", "/nowhere", "", 404, `{"error": "no such route"}`},
	} {
		status, body := call(t, c.method, base+c.path, "application/json", c.body)
		if status != c.status || c.reply != "" && body != c.reply {
			t.Errorf("%s %s %s: %d %s, want %d %s", c.method, c.path, c.body, status, body, c.status, c.reply)
		}
	}
	s.stop(t).expect(t, lines("listening on "+s.addr, "stopped: Ok{ value: void } (interrupt)"), 0)
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
		{name: "logstat", in: exampleRun{stdin: "127.0.0.1 - - [10/Oct/2026:13:55:36 +0000] \"GET /index.html HTTP/1.1\" 200 2326\n127.0.0.1 - frank [10/Oct/2026:13:55:37 +0000] \"GET /missing HTTP/1.1\" 404 512\n10.0.0.5 - - [10/Oct/2026:13:55:38 +0000] \"POST /api/items HTTP/1.1\" 200 74\n10.0.0.5 - - [10/Oct/2026:13:55:39 +0000] \"GET /favicon.ico HTTP/1.1\" 304 -\nnot a log line\n"}, out: lines("4 requests, 1 unparsed", "status   count   share    mean", "200          2  50.00%  1200.0", "304          1  25.00%     0.0", "404          1  25.00%   512.0"), status: 0, err: []string{}},
		{name: "nbody", in: exampleRun{}, out: lines("-0.169075164", "-0.169087605", "sun moved to 0.007413153, 0.003009803, -0.000266704"), status: 0, err: []string{}},
		{name: "queue", in: exampleRun{}, out: lines("took 60 items, sum 6630, largest 220", "queue drained: true", "2 of 3 workers finished"), status: 0, err: []string{"worker 1 failed"}},
		{name: "runner", in: exampleRun{stdin: "[{\"name\": \"version\", \"program\": \"go\", \"args\": [\"version\"]},\n {\"name\": \"nonsense\", \"program\": \"go\", \"args\": [\"nonsense\"]},\n {\"name\": \"missing\", \"program\": \"ant-no-such-program-xyz\"},\n 7]\n"}, out: lines("version: ok, 1 line of output", "nonsense: failed with status 2", "missing: could not start", "skipped: a task must be an object", "1 of 4 tasks succeeded"), status: 1, err: []string{"took"}},
		{name: "shapes", in: exampleRun{}, out: lines("circle: area 3.142, perimeter 6.283", "rect: area 6.0, perimeter 10.0", "triangle: area 6.0, perimeter 12.0", "rect: area 1.0, perimeter 4.0", "largest: rect", "[\"rect\", \"circle\", \"rect\", \"triangle\"]", "points made: 4 true", "Point{ x: 1.0, y: 2.0 }", "2.23606797749979", "points made: 5", "1"), status: 0, err: []string{}},
		{name: "sift", in: exampleRun{args: []string{"ant", "time"}, stdin: "Ant has time values\nno match here\nTIME and ANT again\n"}, out: lines("1: Ant has time values", "3: TIME and ANT again"), status: 0, err: []string{"2 of 3 lines"}},
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

func TestExampleBf(t *testing.T) {
	dir := t.TempDir()
	write := func(name, src string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	hello := write("hello.bf", "++++++++[>++++[>++>+++>+++>+<<<<-]>+>+>->>+[<]<-]>>.>---.+++++++..+++.>>.<-.<.+++.------.--------.>>+.>++.")
	runExample(t, "bf", exampleRun{args: []string{hello}}).expect(t, "Hello World!\n", 0)
	abc := write("abc.bf", "a comment, ignored\n++++++++[>++++++++<-]>+.+.+.")
	runExample(t, "bf", exampleRun{args: []string{abc}}).expect(t, "ABC", 0)
	cat := write("cat.bf", ",[.,]")
	runExample(t, "bf", exampleRun{args: []string{cat}, stdin: "ant"}).expect(t, "ant", 0)
	wrap := write("wrap.bf", "-.")
	runExample(t, "bf", exampleRun{args: []string{wrap}}).expect(t, "ÿ", 0)
	nested := write("nested.bf", "+++[>+++[>+++<-]<-]>>++++++.")
	runExample(t, "bf", exampleRun{args: []string{nested}}).expect(t, "!", 0)
	open := write("open.bf", "[[]")
	runExample(t, "bf", exampleRun{args: []string{open}}).expect(t, "", 1, "[ without ] at 0")
	shut := write("shut.bf", "]")
	runExample(t, "bf", exampleRun{args: []string{shut}}).expect(t, "", 1, "] without [ at 0")
	runExample(t, "bf", exampleRun{}).expect(t, "", 2, "usage:")
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

func TestExampleWordfreq(t *testing.T) {
	dir := t.TempDir()
	var args []string
	for name, text := range map[string]string{
		"fox.txt":   "The quick brown fox jumps over the lazy dog. The dog sleeps; the fox doesn't.",
		"rain.txt":  "The rain in Spain stays mainly in the plain. In the plain the rain is plain.",
		"song.txt":  "Row, row, row your boat gently down the stream. The stream, the stream!",
		"empty.txt": "",
	} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		args = append(args, path)
	}
	args = append(args, filepath.Join(dir, "missing.txt"))
	runExample(t, "wordfreq", exampleRun{args: args}).expect(t, lines(
		"    11 the",
		"     3 in",
		"     3 plain",
		"     3 row",
		"     3 stream",
		"     2 dog",
		"     2 fox",
		"     2 rain",
		"     1 boat",
		"     1 brown",
		"4 files, 23 distinct words, 44 in all",
	), 0, "skipped: ", "missing.txt")
	runExample(t, "wordfreq", exampleRun{}).expect(t, "", 2, "usage:")
}

func TestExampleConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	for _, c := range []struct {
		json, out, err string
	}{
		{`{
			"name": "api",
			"servers": [
				{ "host": "a.example", "port": 443, "tls": true },
				{ "host": "b.example", "port": 8080 }
			],
			"limits": { "rps": 100, "burst": 250 }
		}`, lines(`api: 2 servers, limits Map{ "rps": 100, "burst": 250 }`, "  a.example:443 (tls)", "  b.example:8080"), ""},
		{`{ "name": "bare", "servers": [] }`, "bare: 0 servers, limits Map{}\n", ""},
		{`{ "name": "oops", "servers": [ { "host": "a", "port": 80 }, { "host": "b", "port": "81" } ] }`, "", "$.servers[1].port: expected a number, got a string"},
		{`{ "name": "frac", "servers": [ { "host": "a", "port": 80.5 } ] }`, "", "$.servers[0].port: expected an integer, got 80.5"},
		{`{ "name": "range", "servers": [ { "host": "a", "port": 70000 } ] }`, "", "$.servers[0].port: 70000 is out of range"},
		{`{ "name": "nohost", "servers": [ { "port": 1 } ] }`, "", "$.servers[0]: missing field host"},
		{`{ "name": 7, "servers": [] }`, "", "$.name: expected a string, got a number"},
		{`{ "servers": [] }`, "", "$: missing field name"},
		{`{ "name": "limits", "servers": [], "limits": { "rps": true } }`, "", "$.limits.rps: expected a number, got a bool"},
		{`[1, 2, 3]`, "", "$: expected an object, got an array"},
		{`{ "name": "broken"`, "", "unexpected end of JSON input"},
	} {
		if err := os.WriteFile(path, []byte(c.json), 0o644); err != nil {
			t.Fatal(err)
		}
		status := 0
		if c.err != "" {
			status = 1
		}
		runExample(t, "config", exampleRun{args: []string{path}}).expect(t, c.out, status, c.err)
	}
	runExample(t, "config", exampleRun{args: []string{filepath.Join(t.TempDir(), "none.json")}}).expect(t, "", 1, "none.json")
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

func TestExamplePool(t *testing.T) {
	base := echoServer(t)
	paths := []string{"/hello", "/status/404", "/slow", "/status/500", "/hello"}
	var args []string
	for _, p := range paths {
		args = append(args, base+p)
	}
	runExample(t, "pool", exampleRun{args: args}).expect(t, lines(
		"200  "+base+"/hello",
		"404  "+base+"/status/404",
		"timeout  "+base+"/slow",
		"500  "+base+"/status/500",
		"200  "+base+"/hello",
	), 0)
}

func TestExampleWebhook(t *testing.T) {
	var got []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mac := hmac.New(sha256.New, []byte("shared-secret"))
		mac.Write(body)
		want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if r.Method != "POST" || r.URL.Path != "/hook" || !hmac.Equal([]byte(r.Header.Get("X-Signature")), []byte(want)) {
			http.Error(w, "bad signature", http.StatusUnauthorized)
			return
		}
		got = body
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	runExample(t, "webhook", exampleRun{args: []string{srv.URL + "/hook"}}).expect(t, "delivered: 204\n", 0)
	if string(got) != `{"id":"evt_1","kind":"payment","amount":12.5}` {
		t.Errorf("the receiver got %s", got)
	}
	wrong := exampleRun{args: []string{srv.URL + "/hook"}, env: map[string]string{"WEBHOOK_SECRET": "guess"}}
	runExample(t, "webhook", wrong).expect(t, "delivered: 401\n", 1)
	runExample(t, "webhook", exampleRun{args: []string{"http://127.0.0.1:1/hook"}}).expect(t, "", 1, "delivery failed: ")
}
