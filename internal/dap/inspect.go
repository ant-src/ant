package dap

import (
	"encoding/json"
	"fmt"
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
		m := map[string]any{"name": sc.name, "variablesReference": s.handle(sc.vars), "namedVariables": len(sc.vars), "expensive": false}
		if sc.hint != "" {
			m["presentationHint"] = sc.hint
		}
		out = append(out, m)
	}
	s.reply(req, map[string]any{"scopes": out})
}

// handle returns a variablesReference for a scope's variables or a value's children.
func (s *server) handle(v any) int {
	s.hmu.Lock()
	defer s.hmu.Unlock()
	s.handles = append(s.handles, v)
	return len(s.handles)
}

func (s *server) variables(req *request) {
	var args struct {
		Ref   int `json:"variablesReference"`
		Start int `json:"start"`
		Count int `json:"count"`
	}
	json.Unmarshal(req.Arguments, &args)
	s.hmu.Lock()
	if args.Ref < 1 || args.Ref > len(s.handles) {
		s.hmu.Unlock()
		s.fail(req, "no such variable; the program has resumed since")
		return
	}
	h := s.handles[args.Ref-1]
	s.hmu.Unlock()
	var vars []interp.Var
	switch h := h.(type) {
	case []interp.Var:
		vars = h
	case value.Value:
		vars = children(h, args.Start, args.Count)
	}
	out := []map[string]any{}
	for _, v := range vars {
		out = append(out, s.variable(v))
	}
	s.reply(req, map[string]any{"variables": out})
}

// variable describes one variable, with a reference to its children when it has any.
func (s *server) variable(v interp.Var) map[string]any {
	text := value.RenderQuoted(v.Value)
	if r := []rune(text); len(r) > maxValue {
		text = string(r[:maxValue]) + "…"
	}
	m := map[string]any{"name": v.Name, "value": text, "type": typeName(v.Value), "variablesReference": 0}
	switch x := v.Value.(type) {
	case *value.Record:
		if x.Len() > 0 {
			m["variablesReference"] = s.handle(v.Value)
			m["namedVariables"] = x.Len()
		}
	case *value.Array:
		if x.Len() > 0 {
			m["variablesReference"] = s.handle(v.Value)
			m["indexedVariables"] = x.Len()
		}
	case *value.Set:
		if x.Len() > 0 {
			m["variablesReference"] = s.handle(v.Value)
			m["indexedVariables"] = x.Len()
		}
	case value.Bytes:
		if len(x) > 0 {
			m["variablesReference"] = s.handle(v.Value)
			m["indexedVariables"] = len(x)
		}
	case *value.Map:
		if x.Len() > 0 {
			m["variablesReference"] = s.handle(v.Value)
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

// children lists a value's children, the indexed ones from start, count of them when count is
// above zero.
func children(v value.Value, start, count int) []interp.Var {
	var out []interp.Var
	page := func(n int, at func(i int) value.Value) {
		end := n
		if count > 0 {
			end = min(n, start+count)
		}
		for i := min(start, n); i < end; i++ {
			out = append(out, interp.Var{Name: fmt.Sprintf("[%d]", i), Value: at(i)})
		}
	}
	switch x := v.(type) {
	case *value.Record:
		for i, f := range x.Tag.Fields {
			if i < x.Len() {
				out = append(out, interp.Var{Name: f, Value: x.Get(i)})
			}
		}
	case *value.Array:
		page(x.Len(), func(i int) value.Value { e, _ := x.Get(i); return e })
	case value.Bytes:
		page(len(x), func(i int) value.Value { return x[i] })
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
		page(len(members), func(i int) value.Value { return members[i] })
	case *value.Map:
		var cursor uint64
		for {
			k, e, seq, ok := x.Next(cursor)
			if !ok {
				break
			}
			out = append(out, interp.Var{Name: value.Render(k), Value: e})
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
