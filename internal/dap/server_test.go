package dap

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ant-src/ant/internal/frame"
)

// client drives a Serve over pipes, as an editor would.
type client struct {
	t      *testing.T
	w      io.Writer
	msgs   chan map[string]any
	seq    int
	events []map[string]any // events read while waiting for something else
	status chan int
}

func newClient(t *testing.T) *client {
	t.Helper()
	wd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(wd) }) // launch changes the working directory
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	c := &client{t: t, w: inW, msgs: make(chan map[string]any, 100), status: make(chan int, 1)}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		c.status <- Serve(ctx, inR, outW, io.Discard)
		outW.Close()
	}()
	go func() {
		r := bufio.NewReader(outR)
		for {
			body, err := frame.Read(r)
			if err != nil {
				close(c.msgs)
				return
			}
			var m map[string]any
			json.Unmarshal(body, &m)
			c.msgs <- m
		}
	}()
	return c
}

func (c *client) next() map[string]any {
	c.t.Helper()
	select {
	case m, ok := <-c.msgs:
		if !ok {
			c.t.Fatal("the adapter closed the connection")
		}
		return m
	case <-time.After(5 * time.Second):
		c.t.Fatal("no message from the adapter")
	}
	return nil
}

// request sends a request and returns its response, keeping the events that come before it.
func (c *client) request(command string, args any) map[string]any {
	c.t.Helper()
	c.seq++
	body, _ := json.Marshal(map[string]any{"seq": c.seq, "type": "request", "command": command, "arguments": args})
	if err := frame.Write(c.w, body); err != nil {
		c.t.Fatal(err)
	}
	for {
		m := c.next()
		if m["type"] == "response" && int(m["request_seq"].(float64)) == c.seq {
			return m
		}
		c.events = append(c.events, m)
	}
}

// event waits for the event name, returning its body.
func (c *client) event(name string) map[string]any {
	c.t.Helper()
	for i, e := range c.events {
		if e["event"] == name {
			c.events = append(c.events[:i:i], c.events[i+1:]...)
			body, _ := e["body"].(map[string]any)
			return body
		}
	}
	for {
		m := c.next()
		if m["type"] == "event" && m["event"] == name {
			body, _ := m["body"].(map[string]any)
			return body
		}
		c.events = append(c.events, m)
	}
}

// output is every output event's text so far, of the category.
func (c *client) output(category string) string {
	var sb strings.Builder
	for _, e := range c.events {
		if b, ok := e["body"].(map[string]any); ok && e["event"] == "output" && b["category"] == category {
			sb.WriteString(b["output"].(string))
		}
	}
	return sb.String()
}

func body(m map[string]any) map[string]any {
	b, _ := m["body"].(map[string]any)
	return b
}

func ok(t *testing.T, m map[string]any) map[string]any {
	t.Helper()
	if m["success"] != true {
		t.Fatalf("%s failed: %v", m["command"], m["message"])
	}
	return body(m)
}

func write(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, src := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestSession(t *testing.T) {
	dir := write(t, map[string]string{
		"main.ant": `import "./lib.ant" { double }

type Point = { x: int, y: int }

var p = Point { x: 1, y: 2 }
var xs = [10, 20, 30]
println("before")
var d = double(len(xs))
println(d)
`,
		"lib.ant": `export func double(n: int): int {
  var r = n * 2
  return r
}
`,
	})
	main := filepath.Join(dir, "main.ant")
	lib := filepath.Join(dir, "lib.ant")
	c := newClient(t)
	caps := ok(t, c.request("initialize", map[string]any{"adapterID": "ant"}))
	if caps["supportsConfigurationDoneRequest"] != true {
		t.Errorf("capabilities: %v", caps)
	}
	ok(t, c.request("launch", map[string]any{"program": main}))
	c.event("initialized")
	bps := ok(t, c.request("setBreakpoints", map[string]any{
		"source":      map[string]any{"path": lib},
		"breakpoints": []map[string]any{{"line": 2}, {"line": 4}},
	}))["breakpoints"].([]any)
	if b := bps[0].(map[string]any); b["verified"] != true || b["line"] != 2.0 {
		t.Errorf("breakpoint on line 2: %v", b)
	}
	if b := bps[1].(map[string]any); b["verified"] != false {
		t.Errorf("breakpoint past the body: %v", b)
	}
	ok(t, c.request("setExceptionBreakpoints", map[string]any{"filters": []string{"uncaught"}}))
	ok(t, c.request("configurationDone", nil))

	stop := c.event("stopped")
	if stop["reason"] != "breakpoint" || stop["allThreadsStopped"] != true {
		t.Fatalf("stopped: %v", stop)
	}
	if got := c.output("stdout"); got != "before\n" {
		t.Errorf("output before the stop: %q", got)
	}
	tid := stop["threadId"]
	frames := ok(t, c.request("stackTrace", map[string]any{"threadId": tid}))["stackFrames"].([]any)
	top := frames[0].(map[string]any)
	if top["name"] != "double" || top["line"] != 2.0 || top["source"].(map[string]any)["path"] != lib {
		t.Fatalf("top frame: %v", top)
	}
	if caller := frames[1].(map[string]any); caller["line"] != 8.0 {
		t.Errorf("caller frame: %v", caller)
	}

	scopes := ok(t, c.request("scopes", map[string]any{"frameId": top["id"]}))["scopes"].([]any)
	locals := scopes[0].(map[string]any)
	vars := ok(t, c.request("variables", map[string]any{"variablesReference": locals["variablesReference"]}))["variables"].([]any)
	if len(vars) != 1 || vars[0].(map[string]any)["name"] != "n" || vars[0].(map[string]any)["value"] != "3" {
		t.Errorf("locals of double: %v", vars)
	}

	// The caller's module scope expands a record and an array.
	caller := frames[1].(map[string]any)
	scopes = ok(t, c.request("scopes", map[string]any{"frameId": caller["id"]}))["scopes"].([]any)
	var module map[string]any
	for _, s := range scopes {
		if s.(map[string]any)["name"] == "Module" {
			module = s.(map[string]any)
		}
	}
	if module == nil {
		t.Fatalf("no Module scope: %v", scopes)
	}
	vars = ok(t, c.request("variables", map[string]any{"variablesReference": module["variablesReference"]}))["variables"].([]any)
	byName := map[string]map[string]any{}
	for _, v := range vars {
		byName[v.(map[string]any)["name"].(string)] = v.(map[string]any)
	}
	if p := byName["p"]; p == nil || p["type"] != "Point" || p["value"] != "Point{ x: 1, y: 2 }" {
		t.Errorf("p: %v", p)
	}
	if _, ok := byName["d"]; ok {
		t.Errorf("d is shown before its initializer ran: %v", vars)
	}
	xs := byName["xs"]
	if xs == nil || xs["indexedVariables"] != 3.0 {
		t.Fatalf("xs: %v", xs)
	}
	elems := ok(t, c.request("variables", map[string]any{"variablesReference": xs["variablesReference"], "start": 1, "count": 2}))["variables"].([]any)
	if len(elems) != 2 || elems[0].(map[string]any)["name"] != "[1]" || elems[1].(map[string]any)["value"] != "30" {
		t.Errorf("xs[1:3]: %v", elems)
	}

	ok(t, c.request("next", map[string]any{"threadId": tid}))
	if stop := c.event("stopped"); stop["reason"] != "step" {
		t.Fatalf("after next: %v", stop)
	}
	ok(t, c.request("continue", map[string]any{"threadId": tid}))
	if code := c.event("exited")["exitCode"]; code != 0.0 {
		t.Errorf("exit code %v", code)
	}
	c.event("terminated")
	if got := c.output("stdout"); got != "before\n6\n" {
		t.Errorf("output: %q", got)
	}
	ok(t, c.request("disconnect", nil))
	if s := <-c.status; s != 0 {
		t.Errorf("status %d", s)
	}
}

func TestPanicStopAndExitCode(t *testing.T) {
	dir := write(t, map[string]string{"main.ant": `var xs = [1]
println(xs[3])
`})
	c := newClient(t)
	ok(t, c.request("initialize", nil))
	ok(t, c.request("launch", map[string]any{"program": filepath.Join(dir, "main.ant")}))
	ok(t, c.request("configurationDone", nil))
	stop := c.event("stopped")
	if stop["reason"] != "exception" {
		t.Fatalf("stopped: %v", stop)
	}
	info := ok(t, c.request("exceptionInfo", map[string]any{"threadId": stop["threadId"]}))
	if !strings.Contains(info["description"].(string), "out of range") {
		t.Errorf("exceptionInfo: %v", info)
	}
	ok(t, c.request("continue", nil))
	if code := c.event("exited")["exitCode"]; code != 1.0 {
		t.Errorf("exit code %v", code)
	}
	if !strings.Contains(c.output("stderr"), "panic: index 3 out of range") {
		t.Errorf("stderr: %q", c.output("stderr"))
	}
}

func TestLaunchErrorsAndStdin(t *testing.T) {
	dir := write(t, map[string]string{
		"bad.ant":  "var = 1\n",
		"echo.ant": "println(File.readAll(Os.stdin()))\nprintln(Os.args())\n",
		"in.txt":   "hello",
	})
	c := newClient(t)
	ok(t, c.request("initialize", nil))
	if r := c.request("launch", map[string]any{"program": filepath.Join(dir, "bad.ant")}); r["success"] != false {
		t.Errorf("a program with a syntax error launched: %v", r)
	}

	c = newClient(t)
	ok(t, c.request("initialize", nil))
	ok(t, c.request("launch", map[string]any{"program": "echo.ant", "cwd": dir, "stdin": "in.txt", "args": []string{"a"}}))
	ok(t, c.request("configurationDone", nil))
	c.event("terminated")
	if got := c.output("stdout"); !strings.Contains(got, `"hello"`) {
		t.Errorf("stdout: %q", got)
	}
	if got := c.output("stdout"); !strings.Contains(got, `["a"]`) {
		t.Errorf("args: %q", got)
	}
}
