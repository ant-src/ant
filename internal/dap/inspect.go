package dap

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// maxValue is how many characters of a value a variable shows (dap.md section 3, Variables).
const maxValue = 200

func (s *server) stackTrace(req *request) {
	var args struct {
		ThreadID   int `json:"threadId"`
		StartFrame int `json:"startFrame"`
		Levels     int `json:"levels"`
	}
	json.Unmarshal(req.Arguments, &args)
	if s.d == nil {
		s.fail(req, "the program is not running")
		return
	}
	frames, ok := s.d.Frames(args.ThreadID)
	if !ok {
		s.fail(req, "the program is running, or there is no such goroutine")
		return
	}
	total := len(frames)
	frames = frames[min(args.StartFrame, total):]
	if args.Levels > 0 && args.Levels < len(frames) {
		frames = frames[:args.Levels]
	}
	out := []map[string]any{}
	s.hmu.Lock()
	for _, f := range frames {
		s.frames = append(s.frames, f)
		m := s.moduleOf(f.Path)
		sf := map[string]any{
			"id":     len(s.frames),
			"name":   f.Function,
			"line":   s.line(f.Pos.Line),
			"column": s.col(m, f.Pos),
		}
		if m != nil {
			sf["source"] = map[string]any{"name": f.Path, "path": m.abs}
		}
		out = append(out, sf)
	}
	s.hmu.Unlock()
	s.reply(req, map[string]any{"stackFrames": out, "totalFrames": total})
}

func (s *server) scopes(req *request) {
	var args struct {
		FrameID int `json:"frameId"`
	}
	json.Unmarshal(req.Arguments, &args)
	s.hmu.Lock()
	if args.FrameID < 1 || args.FrameID > len(s.frames) {
		s.hmu.Unlock()
		s.fail(req, "no such frame; the program has resumed since")
		return
	}
	f := s.frames[args.FrameID-1]
	s.hmu.Unlock()
	locals, closure, module := s.d.Scopes(f)
	out := []map[string]any{}
	for _, sc := range []struct {
		name string
		hint string
		vars []interp.Var
	}{{"Locals", "locals", locals}, {"Closure", "", closure}, {"Module", "", module}} {
		if len(sc.vars) == 0 {
			continue
		}
		items := make([]item, len(sc.vars))
		for i, v := range sc.vars {
			items[i] = item{name: v.Name, value: v.Value}
			if v.Settable() {
				items[i].set = func(x value.Value) error { return s.d.SetVar(v, x) }
			}
		}
		m := map[string]any{"name": sc.name, "variablesReference": s.handle(handle{items: items, frame: f}), "namedVariables": len(sc.vars), "expensive": false}
		if sc.hint != "" {
			m["presentationHint"] = sc.hint
		}
		out = append(out, m)
	}
	s.reply(req, map[string]any{"scopes": out})
}

// item is a variable as the adapter lists it: a binding of a scope, or a part of a value.
type item struct {
	name  string
	value value.Value
	set   func(value.Value) error // nil when it cannot be set
}

// handle is what a variablesReference names: a scope's variables or a value's parts, with the
// frame they were listed for, where setVariable evaluates.
type handle struct {
	items []item
	value value.Value
	frame interp.StackFrame
}

// handle returns a variablesReference for h.
func (s *server) handle(h handle) int {
	s.hmu.Lock()
	defer s.hmu.Unlock()
	s.handles = append(s.handles, h)
	return len(s.handles)
}

// lookup returns the handle a variablesReference names.
func (s *server) lookup(ref int) (handle, bool) {
	s.hmu.Lock()
	defer s.hmu.Unlock()
	if ref < 1 || ref > len(s.handles) {
		return handle{}, false
	}
	return s.handles[ref-1], true
}

// items lists a handle's variables, the indexed ones from start, count of them when count is
// above zero.
func (h handle) list(start, count int) []item {
	if h.items != nil {
		return h.items
	}
	return children(h.value, start, count)
}

func (s *server) variables(req *request) {
	var args struct {
		Ref   int `json:"variablesReference"`
		Start int `json:"start"`
		Count int `json:"count"`
	}
	json.Unmarshal(req.Arguments, &args)
	h, found := s.lookup(args.Ref)
	if !found {
		s.fail(req, "no such variable; the program has resumed since")
		return
	}
	out := []map[string]any{}
	for _, it := range h.list(args.Start, args.Count) {
		out = append(out, s.variable(it, h.frame))
	}
	s.reply(req, map[string]any{"variables": out})
}

// setVariable gives a variable a new value (dap.md section 6, Setting variables).
func (s *server) setVariable(req *request) {
	var args struct {
		Ref   int    `json:"variablesReference"`
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	json.Unmarshal(req.Arguments, &args)
	h, found := s.lookup(args.Ref)
	if !found {
		s.fail(req, "no such variable; the program has resumed since")
		return
	}
	var target *item
	for _, it := range h.list(0, 0) {
		if it.name == args.Name {
			target = &it
			break
		}
	}
	if target == nil || target.set == nil {
		s.fail(req, fmt.Sprintf("%s cannot be set", args.Name))
		return
	}
	v, err := s.d.Evaluate(h.frame, args.Value)
	if err != nil {
		s.fail(req, err.Error())
		return
	}
	if err := target.set(v); err != nil {
		s.fail(req, err.Error())
		return
	}
	shown := s.variable(item{name: args.Name, value: v}, h.frame)
	body := map[string]any{"value": shown["value"], "type": shown["type"], "variablesReference": shown["variablesReference"]}
	for _, k := range []string{"namedVariables", "indexedVariables"} {
		if n, ok := shown[k]; ok {
			body[k] = n
		}
	}
	s.reply(req, body)
}

// variable describes one variable, with a reference to its children when it has any.
func (s *server) variable(v item, frame interp.StackFrame) map[string]any {
	text := value.RenderQuoted(v.value)
	if r := []rune(text); len(r) > maxValue {
		text = string(r[:maxValue]) + "…"
	}
	m := map[string]any{"name": v.name, "value": text, "type": typeName(v.value), "variablesReference": 0}
	ref := func() int { return s.handle(handle{value: v.value, frame: frame}) }
	switch x := v.value.(type) {
	case *value.Record:
		if x.Len() > 0 {
			m["variablesReference"] = ref()
			m["namedVariables"] = x.Len()
		}
	case *value.Array:
		if x.Len() > 0 {
			m["variablesReference"] = ref()
			m["indexedVariables"] = x.Len()
		}
	case *value.Set:
		if x.Len() > 0 {
			m["variablesReference"] = ref()
			m["indexedVariables"] = x.Len()
		}
	case value.Bytes:
		if len(x) > 0 {
			m["variablesReference"] = ref()
			m["indexedVariables"] = len(x)
		}
	case *value.Map:
		if x.Len() > 0 {
			m["variablesReference"] = ref()
			m["namedVariables"] = x.Len()
		}
	}
	return m
}

// typeName is a value's runtime type (dap.md section 3, Variables).
func typeName(v value.Value) string {
	switch x := v.(type) {
	case *value.Record:
		return x.Tag.Name
	case *value.Tagged:
		return x.Tag.Name
	}
	return value.KindOf(v).String()
}

// children lists a value's parts, the indexed ones from start, count of them when count is
// above zero, with what sets those that can be set.
func children(v value.Value, start, count int) []item {
	var out []item
	page := func(n int, at func(i int) value.Value, set func(i int) func(value.Value) error) {
		end := n
		if count > 0 {
			end = min(n, start+count)
		}
		for i := min(start, n); i < end; i++ {
			it := item{name: fmt.Sprintf("[%d]", i), value: at(i)}
			if set != nil {
				it.set = set(i)
			}
			out = append(out, it)
		}
	}
	switch x := v.(type) {
	case *value.Record:
		for i, f := range x.Tag.Fields {
			if i < x.Len() {
				out = append(out, item{name: f, value: x.Get(i), set: func(nv value.Value) error { x.Set(i, nv); return nil }})
			}
		}
	case *value.Array:
		page(x.Len(), func(i int) value.Value { e, _ := x.Get(i); return e }, func(i int) func(value.Value) error {
			return func(nv value.Value) error {
				if !x.Set(i, nv) {
					return fmt.Errorf("index %d is out of range now", i)
				}
				return nil
			}
		})
	case value.Bytes:
		page(len(x), func(i int) value.Value { return x[i] }, nil)
	case *value.Set:
		var members []value.Value
		var cursor uint64
		for {
			m, seq, ok := x.Next(cursor)
			if !ok {
				break
			}
			members = append(members, m)
			cursor = seq
		}
		page(len(members), func(i int) value.Value { return members[i] }, nil)
	case *value.Map:
		var cursor uint64
		for {
			k, e, seq, ok := x.Next(cursor)
			if !ok {
				break
			}
			out = append(out, item{name: value.Render(k), value: e, set: func(nv value.Value) error { x.Set(k, nv); return nil }})
			cursor = seq
		}
	}
	return out
}

func (s *server) exceptionInfo(req *request) {
	var p *interp.AntPanic
	if s.d != nil {
		p = s.d.Panic()
	}
	if p == nil {
		s.fail(req, "the program is not stopped at a panic")
		return
	}
	var sb strings.Builder
	sb.WriteString("panic: " + p.Message + "\n")
	for _, f := range p.Trace {
		fmt.Fprintf(&sb, "  at %s (%s:%d:%d)\n", f.Function, f.Path, f.Pos.Line, f.Pos.Col)
	}
	s.reply(req, map[string]any{
		"exceptionId": "panic",
		"description": p.Message,
		"breakMode":   "always",
		"details":     map[string]any{"message": p.Message, "stackTrace": sb.String()},
	})
}

// fieldPath is what hover evaluates: a name, or a name followed by fields.
var fieldPath = regexp.MustCompile(`^\s*[A-Za-z_][A-Za-z0-9_]*(\s*\.\s*[A-Za-z_][A-Za-z0-9_]*)*\s*$`)

func (s *server) evaluate(req *request) {
	var args struct {
		Expression string `json:"expression"`
		FrameID    int    `json:"frameId"`
		Context    string `json:"context"`
	}
	json.Unmarshal(req.Arguments, &args)
	if args.Context == "hover" && !fieldPath.MatchString(args.Expression) {
		s.fail(req, "hover shows only a name or a field path")
		return
	}
	s.hmu.Lock()
	if s.d == nil || args.FrameID < 1 || args.FrameID > len(s.frames) {
		s.hmu.Unlock()
		s.fail(req, "the program is not stopped in a frame")
		return
	}
	f := s.frames[args.FrameID-1]
	s.hmu.Unlock()
	v, err := s.d.Evaluate(f, args.Expression)
	if err != nil {
		s.fail(req, err.Error())
		return
	}
	shown := s.variable(item{name: args.Expression, value: v}, f)
	body := map[string]any{"result": shown["value"], "type": shown["type"], "variablesReference": shown["variablesReference"]}
	for _, k := range []string{"namedVariables", "indexedVariables"} {
		if n, ok := shown[k]; ok {
			body[k] = n
		}
	}
	s.reply(req, body)
}
