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

func TestServeEndsWithStatus0WhenTheClientClosesAfterTerminated(t *testing.T) {
	dir := t.TempDir()
	main := filepath.Join(dir, "main.ant")
	os.WriteFile(main, []byte("println(1)\n"), 0o644)
	c := newClient(t)
	ok(t, c.request("initialize", map[string]any{"adapterID": "ant"}))
	ok(t, c.request("launch", map[string]any{"program": main}))
	c.event("initialized")
	ok(t, c.request("configurationDone", nil))
	c.event("terminated")
	c.w.(io.Closer).Close() // as nvim-dap does: no disconnect after terminated
	if s := <-c.status; s != 0 {
		t.Errorf("status %d, want 0", s)
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

func TestEvaluateAndLogpoints(t *testing.T) {
	dir := write(t, map[string]string{"main.ant": `var xs = [1, 2, 3]
var sum = 0
for x in xs {
  sum += x
}
println(sum)
`})
	main := filepath.Join(dir, "main.ant")
	c := newClient(t)
	caps := ok(t, c.request("initialize", nil))
	for _, k := range []string{"supportsEvaluateForHovers", "supportsConditionalBreakpoints", "supportsHitConditionalBreakpoints", "supportsLogPoints"} {
		if caps[k] != true {
			t.Errorf("%s: %v", k, caps[k])
		}
	}
	ok(t, c.request("launch", map[string]any{"program": main}))
	bps := ok(t, c.request("setBreakpoints", map[string]any{
		"source": map[string]any{"path": main},
		"breakpoints": []map[string]any{
			{"line": 4, "logMessage": "x is {x}"},
			{"line": 6, "condition": "sum > 5"},
			{"line": 2, "hitCondition": "sometimes"},
		},
	}))["breakpoints"].([]any)
	if b := bps[2].(map[string]any); b["verified"] != false || !strings.Contains(b["message"].(string), "hit condition") {
		t.Errorf("a bad hit condition: %v", b)
	}
	ok(t, c.request("configurationDone", nil))
	stop := c.event("stopped")
	if got := c.output("console"); got != "x is 1\nx is 2\nx is 3\n" {
		t.Errorf("logpoint output %q", got)
	}
	frame := ok(t, c.request("stackTrace", map[string]any{"threadId": stop["threadId"]}))["stackFrames"].([]any)[0].(map[string]any)
	res := ok(t, c.request("evaluate", map[string]any{"expression": "xs | Array.map({ x * sum })", "frameId": frame["id"], "context": "repl"}))
	if res["result"] != "[6, 12, 18]" || res["indexedVariables"] != 3.0 {
		t.Errorf("evaluate: %v", res)
	}
	if r := c.request("evaluate", map[string]any{"expression": "len(xs)", "frameId": frame["id"], "context": "hover"}); r["success"] != false {
		t.Errorf("hover called a function: %v", r)
	}
	if res := ok(t, c.request("evaluate", map[string]any{"expression": "sum", "frameId": frame["id"], "context": "hover"})); res["result"] != "6" {
		t.Errorf("hover: %v", res)
	}
	ok(t, c.request("continue", nil))
	c.event("terminated")
}

func TestSetVariableAndFunctionBreakpoints(t *testing.T) {
	dir := write(t, map[string]string{"main.ant": `type Point = { x: int, y: int }

func Point.sum(p: Point): int {
  var s = p.x + p.y
  return s
}

var p = Point { x: 1, y: 2 }
var scores = Map{ "ada": 1 }
var n = 10
println(Point.sum(p) + n + scores["ada"])
`})
	main := filepath.Join(dir, "main.ant")
	c := newClient(t)
	caps := ok(t, c.request("initialize", nil))
	if caps["supportsSetVariable"] != true || caps["supportsFunctionBreakpoints"] != true {
		t.Errorf("capabilities: %v", caps)
	}
	ok(t, c.request("launch", map[string]any{"program": main}))
	fbs := ok(t, c.request("setFunctionBreakpoints", map[string]any{"breakpoints": []map[string]any{{"name": "Point.sum"}, {"name": "nothing"}}}))["breakpoints"].([]any)
	if b := fbs[0].(map[string]any); b["verified"] != true || b["line"] != 4.0 {
		t.Errorf("Point.sum: %v", b)
	}
	if b := fbs[1].(map[string]any); b["verified"] != false {
		t.Errorf("nothing: %v", b)
	}
	ok(t, c.request("configurationDone", nil))
	stop := c.event("stopped")
	if stop["reason"] != "function breakpoint" {
		t.Fatalf("stopped: %v", stop)
	}
	frames := ok(t, c.request("stackTrace", map[string]any{"threadId": stop["threadId"]}))["stackFrames"].([]any)
	scope := func(frame any, name string) float64 {
		for _, sc := range ok(t, c.request("scopes", map[string]any{"frameId": frame}))["scopes"].([]any) {
			if sc.(map[string]any)["name"] == name {
				return sc.(map[string]any)["variablesReference"].(float64)
			}
		}
		t.Fatalf("no %s scope", name)
		return 0
	}
	vars := func(ref any) map[string]map[string]any {
		out := map[string]map[string]any{}
		for _, v := range ok(t, c.request("variables", map[string]any{"variablesReference": ref}))["variables"].([]any) {
			out[v.(map[string]any)["name"].(string)] = v.(map[string]any)
		}
		return out
	}
	locals := scope(frames[0].(map[string]any)["id"], "Locals")
	// a field of the parameter, which the caller's p shares
	pRef := vars(locals)["p"]["variablesReference"]
	if r := ok(t, c.request("setVariable", map[string]any{"variablesReference": pRef, "name": "y", "value": "p.x * 100"})); r["value"] != "100" {
		t.Errorf("p.y: %v", r)
	}
	module := scope(frames[1].(map[string]any)["id"], "Module")
	if r := ok(t, c.request("setVariable", map[string]any{"variablesReference": module, "name": "n", "value": "n + 1"})); r["value"] != "11" {
		t.Errorf("n: %v", r)
	}
	scoresRef := vars(module)["scores"]["variablesReference"]
	ok(t, c.request("setVariable", map[string]any{"variablesReference": scoresRef, "name": "ada", "value": "5"}))
	if r := c.request("setVariable", map[string]any{"variablesReference": module, "name": "n", "value": "nope"}); r["success"] != false {
		t.Errorf("an expression that fails set n: %v", r)
	}
	ok(t, c.request("continue", nil))
	c.event("terminated")
	// sum is 1 + 100, then n is 11 and scores["ada"] is 5
	if got := c.output("stdout"); got != "117\n" {
		t.Errorf("output %q", got)
	}
}

func TestLaunchRelativeProgram(t *testing.T) {
	// Os.program is the resolved path, which still names the program once the adapter has moved
	// into its directory.
	dir := write(t, map[string]string{"main.ant": "println(1)\nprintln(Path.isAbs(Os.program()) && File.exists(Os.program()))\n"})
	parent := filepath.Dir(dir)
	rel := filepath.Join(filepath.Base(dir), "main.ant")
	wd, _ := os.Getwd()
	if err := os.Chdir(parent); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(wd) })
	c := newClient(t)
	ok(t, c.request("initialize", nil))
	ok(t, c.request("launch", map[string]any{"program": rel})) // relative to the adapter's directory, with no cwd
	ok(t, c.request("configurationDone", nil))
	c.event("terminated")
	if got := c.output("stdout"); got != "1\ntrue\n" {
		t.Errorf("stdout: %q", got)
	}
}

func TestSequenceBodyFrameIsInTheLoopsModule(t *testing.T) {
	dir := write(t, map[string]string{
		"main.ant": `import "./lib.ant" { count }

var seen = 0
for n in count {
  seen += n
}
println(seen)
`,
		"lib.ant": `export func count(yield: func(int): bool) {
  for i in [1, 2] {
    if !yield(i) { return }
  }
}
`,
	})
	main := filepath.Join(dir, "main.ant")
	c := newClient(t)
	ok(t, c.request("initialize", nil))
	ok(t, c.request("launch", map[string]any{"program": main}))
	ok(t, c.request("setBreakpoints", map[string]any{"source": map[string]any{"path": main}, "breakpoints": []map[string]any{{"line": 5}}}))
	ok(t, c.request("configurationDone", nil))
	stop := c.event("stopped")
	frames := ok(t, c.request("stackTrace", map[string]any{"threadId": stop["threadId"]}))["stackFrames"].([]any)
	top := frames[0].(map[string]any)
	if top["name"] != "count" || top["line"] != 5.0 || top["source"].(map[string]any)["path"] != main {
		t.Errorf("the body's frame: %v", top)
	}
	ok(t, c.request("setBreakpoints", map[string]any{"source": map[string]any{"path": main}, "breakpoints": []map[string]any{}}))
	ok(t, c.request("continue", nil))
	c.event("terminated")
	if got := c.output("stdout"); got != "3\n" {
		t.Errorf("stdout: %q", got)
	}
}

func TestServeRejectsWhatIsNotARequestAndEndsWithStatus1(t *testing.T) {
	c := newClient(t)
	frame.Write(c.w, []byte("not json"))
	frame.Write(c.w, []byte(`{"seq": 1, "type": "event", "event": "x"}`))
	if r := c.request("nothing", nil); r["success"] != false || !strings.Contains(r["message"].(string), "does not serve") {
		t.Errorf("an unknown command: %v", r)
	}
	c.w.(io.Closer).Close() // the connection ends without disconnect
	if s := <-c.status; s != 1 {
		t.Errorf("status %d, want 1", s)
	}
}

func TestRequestsBeforeLaunchAndWhileRunning(t *testing.T) {
	dir := write(t, map[string]string{"main.ant": `var n = 0
while true {
  n += 1
}
`})
	main := filepath.Join(dir, "main.ant")
	c := newClient(t)
	ok(t, c.request("initialize", map[string]any{"linesStartAt1": false, "columnsStartAt1": false}))
	bps := ok(t, c.request("setBreakpoints", map[string]any{"source": map[string]any{"path": main}, "lines": []int{2}}))["breakpoints"].([]any)
	if b := bps[0].(map[string]any); b["verified"] != false || !strings.Contains(b["message"].(string), "not launched") {
		t.Errorf("a breakpoint before launch: %v", b)
	}
	if r := c.request("stackTrace", map[string]any{"threadId": 1}); r["success"] != false {
		t.Errorf("stackTrace before launch: %v", r)
	}
	if r := c.request("exceptionInfo", map[string]any{"threadId": 1}); r["success"] != false {
		t.Errorf("exceptionInfo before launch: %v", r)
	}
	if r := c.request("evaluate", map[string]any{"expression": "1", "frameId": 1}); r["success"] != false {
		t.Errorf("evaluate before launch: %v", r)
	}
	if threads := ok(t, c.request("threads", nil))["threads"].([]any); len(threads) != 0 {
		t.Errorf("threads before launch: %v", threads)
	}
	ok(t, c.request("setExceptionBreakpoints", map[string]any{"filters": []string{"uncaught", "all"}}))
	ok(t, c.request("terminate", nil)) // nothing to terminate yet

	if r := c.request("launch", map[string]any{"program": 7}); r["success"] != false {
		t.Errorf("launch with a malformed argument: %v", r)
	}
	if r := c.request("launch", map[string]any{}); r["success"] != false || !strings.Contains(r["message"].(string), "needs a program") {
		t.Errorf("launch without a program: %v", r)
	}
	if r := c.request("launch", map[string]any{"program": main, "stdin": filepath.Join(dir, "missing.txt")}); r["success"] != false {
		t.Errorf("launch with a missing stdin: %v", r)
	}
	if r := c.request("launch", map[string]any{"program": main, "cwd": filepath.Join(dir, "nowhere")}); r["success"] != false {
		t.Errorf("launch with a missing cwd: %v", r)
	}
	ok(t, c.request("launch", map[string]any{"program": main}))
	if r := c.request("launch", map[string]any{"program": main}); r["success"] != false || !strings.Contains(r["message"].(string), "already launched") {
		t.Errorf("a second launch: %v", r)
	}
	// lines and columns from 0: line 2 is the third line, the loop body
	bps = ok(t, c.request("setBreakpoints", map[string]any{"source": map[string]any{"path": main}, "lines": []int{2}}))["breakpoints"].([]any)
	if b := bps[0].(map[string]any); b["verified"] != true || b["line"] != 2.0 || b["column"] != 2.0 {
		t.Errorf("a breakpoint counted from 0: %v", b)
	}
	bps = ok(t, c.request("setBreakpoints", map[string]any{"source": map[string]any{"path": filepath.Join(dir, "other.ant")}, "lines": []int{0}}))["breakpoints"].([]any)
	if b := bps[0].(map[string]any); b["verified"] != false || !strings.Contains(b["message"].(string), "not a module") {
		t.Errorf("a breakpoint in another file: %v", b)
	}
	ok(t, c.request("configurationDone", nil))
	stop := c.event("stopped")
	frames := ok(t, c.request("stackTrace", map[string]any{"threadId": stop["threadId"], "startFrame": 0, "levels": 1}))["stackFrames"].([]any)
	if top := frames[0].(map[string]any); top["line"] != 2.0 {
		t.Errorf("a frame counted from 0: %v", top)
	}
	if r := c.request("scopes", map[string]any{"frameId": 99}); r["success"] != false {
		t.Errorf("scopes of no frame: %v", r)
	}
	if r := c.request("variables", map[string]any{"variablesReference": 99}); r["success"] != false {
		t.Errorf("variables of no reference: %v", r)
	}
	if r := c.request("setVariable", map[string]any{"variablesReference": 99, "name": "n", "value": "1"}); r["success"] != false {
		t.Errorf("setVariable of no reference: %v", r)
	}
	if r := c.request("exceptionInfo", map[string]any{"threadId": stop["threadId"]}); r["success"] != false {
		t.Errorf("exceptionInfo at a breakpoint: %v", r)
	}
	ok(t, c.request("setBreakpoints", map[string]any{"source": map[string]any{"path": main}, "breakpoints": []map[string]any{}}))
	ok(t, c.request("stepIn", map[string]any{"threadId": stop["threadId"]}))
	c.event("stopped")
	ok(t, c.request("stepOut", map[string]any{"threadId": stop["threadId"]})) // the module body has no caller: the program runs on
	time.Sleep(20 * time.Millisecond)
	if r := c.request("stackTrace", map[string]any{"threadId": 1}); r["success"] != false {
		t.Errorf("stackTrace while running: %v", r)
	}
	if threads := ok(t, c.request("threads", nil))["threads"].([]any); len(threads) != 1 {
		t.Errorf("threads while running: %v", threads)
	}
	ok(t, c.request("pause", map[string]any{"threadId": 1}))
	if stop := c.event("stopped"); stop["reason"] != "pause" {
		t.Errorf("after pause: %v", stop)
	}
	ok(t, c.request("terminate", nil)) // the first interrupt ends a program that did not ask for one
	if code := c.event("exited")["exitCode"]; code != 130.0 {
		t.Errorf("exit code after terminate: %v", code)
	}
	c.event("terminated")
	ok(t, c.request("disconnect", nil))
}

func TestVariablesOfEveryKind(t *testing.T) {
	dir := write(t, map[string]string{"main.ant": `type Id = int

var set = Set.of([3, 4])
var bytes = String.toBytes("hi")
var m = Map{ "k": [1] }
var id = Id(7)
var empty = []
var text = String.repeat("x", 300)
println(id)
`})
	main := filepath.Join(dir, "main.ant")
	c := newClient(t)
	ok(t, c.request("initialize", nil))
	ok(t, c.request("launch", map[string]any{"program": main}))
	// a blank line binds to the next stop point of the module body
	bps := ok(t, c.request("setBreakpoints", map[string]any{"source": map[string]any{"path": main}, "breakpoints": []map[string]any{{"line": 2}}}))["breakpoints"].([]any)
	if b := bps[0].(map[string]any); b["verified"] != true || b["line"] != 3.0 {
		t.Errorf("a breakpoint on a blank line: %v", b)
	}
	ok(t, c.request("setBreakpoints", map[string]any{"source": map[string]any{"path": main}, "breakpoints": []map[string]any{{"line": 9}}}))
	fbs := ok(t, c.request("setFunctionBreakpoints", map[string]any{"breakpoints": []map[string]any{{"name": "nothing", "condition": "1 +"}}}))["breakpoints"].([]any)
	if b := fbs[0].(map[string]any); b["verified"] != false {
		t.Errorf("a function breakpoint on nothing: %v", b)
	}
	ok(t, c.request("configurationDone", nil))
	stop := c.event("stopped")
	frame := ok(t, c.request("stackTrace", map[string]any{"threadId": stop["threadId"]}))["stackFrames"].([]any)[0].(map[string]any)
	scopes := ok(t, c.request("scopes", map[string]any{"frameId": frame["id"]}))["scopes"].([]any)
	vars := map[string]map[string]any{}
	for _, v := range ok(t, c.request("variables", map[string]any{"variablesReference": scopes[0].(map[string]any)["variablesReference"]}))["variables"].([]any) {
		vars[v.(map[string]any)["name"].(string)] = v.(map[string]any)
	}
	if v := vars["id"]; v["type"] != "Id" || v["value"] != "Id(7)" {
		t.Errorf("id: %v", v)
	}
	if v := vars["empty"]; v["variablesReference"] != 0.0 {
		t.Errorf("an empty array expands: %v", v)
	}
	if v := vars["text"]; len([]rune(v["value"].(string))) != maxValue+1 {
		t.Errorf("a long string is not cut: %d runes", len([]rune(v["value"].(string))))
	}
	readOnly := func(v map[string]any) bool {
		hint, _ := v["presentationHint"].(map[string]any)
		attrs, _ := hint["attributes"].([]any)
		return len(attrs) == 1 && attrs[0] == "readOnly"
	}
	for name, typ := range map[string]string{"set": "set", "bytes": "Bytes"} {
		v := vars[name]
		if v["type"] != typ || v["indexedVariables"] != 2.0 {
			t.Fatalf("%s: %v", name, v)
		}
		members := ok(t, c.request("variables", map[string]any{"variablesReference": v["variablesReference"], "start": 1, "count": 5}))["variables"].([]any)
		if len(members) != 1 || members[0].(map[string]any)["name"] != "[1]" || !readOnly(members[0].(map[string]any)) {
			t.Errorf("%s[1:]: %v", name, members)
		}
		if r := c.request("setVariable", map[string]any{"variablesReference": v["variablesReference"], "name": "[1]", "value": "9"}); r["success"] != false {
			t.Errorf("a member of %s was set: %v", name, r)
		}
	}
	entries := ok(t, c.request("variables", map[string]any{"variablesReference": vars["m"]["variablesReference"]}))["variables"].([]any)
	if e := entries[0].(map[string]any); e["name"] != "k" || e["value"] != "[1]" || readOnly(e) {
		t.Errorf("the map's entry: %v", entries)
	}
	ok(t, c.request("continue", nil))
	c.event("terminated")
}

func TestLaunchDeclarationFile(t *testing.T) {
	dir := write(t, map[string]string{"lib.d.ant": "func double(n: int): int\n"})
	c := newClient(t)
	ok(t, c.request("initialize", nil))
	r := c.request("launch", map[string]any{"program": filepath.Join(dir, "lib.d.ant")})
	if r["success"] != false || !strings.Contains(r["message"].(string), "declaration file") {
		t.Errorf("a declaration file launched: %v", r)
	}
}
