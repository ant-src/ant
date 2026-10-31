package interp

import (
	"slices"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/format"
	"github.com/ant-src/ant/value"
)

// assertRec holds the values of the parts of an assert's condition (section 8, assert), in the
// order their evaluation finished.
type assertRec struct {
	a     *ast.Assert
	vals  []value.Value // by part, nil until it ran
	order []int         // the parts in the order they finished
}

// evalRecorded evaluates e while an assert's condition is evaluated, keeping its value when it
// is a part of that condition.
func (g *Goroutine) evalRecorded(e ast.Expr, env *Env) (value.Value, ctl) {
	rec := g.rec
	i, ok := rec.a.Index[e]
	v, c := g.evalExpr(e, env)
	if ok && c == ctlNone && rec.vals[i] == nil {
		rec.vals[i] = v
		rec.order = append(rec.order, i)
	}
	return v, c
}

// assert evaluates a call of assert: the condition, recording its parts, and on failure the
// message, then the panic with the report.
func (g *Goroutine) assert(x *ast.Call, env *Env) (value.Value, ctl) {
	a := x.Assert
	slots := g.assertSlots(x)
	var rec *assertRec
	if len(a.Parts) > 0 {
		rec = g.newRec(a)
	}
	v, c := g.evalCond(a.Cond, env, rec)
	if c != ctlNone {
		g.freeRec(rec)
		return v, c
	}
	ok, isBool := v.(bool)
	if !isBool {
		g.panicf(a.Cond.Position(), "an assert condition must be a bool, got %s", value.Describe(v))
	}
	if ok {
		g.freeRec(rec)
		return value.Void, ctlNone
	}
	msg := ""
	if slots[1] >= 0 {
		message := x.Args[slots[1]].Value
		m, c := g.eval(message, env)
		if c != ctlNone {
			return m, c
		}
		s, isString := m.(string)
		if !isString {
			g.panicf(message.Position(), "assert needs a string message, got %s", value.Describe(m))
		}
		msg = s
	}
	g.panicf(x.Pos, "%s", report(a, rec, msg))
	return nil, ctlNone
}

// assertParams are assert's parameters, for matching a call's arguments as a builtin's are.
var assertParams = []string{"cond", "message"}

// assertSlots is the argument that fills each of assert's parameters, -1 for a message left out.
// It panics as a call of a builtin with these parameters would.
func (g *Goroutine) assertSlots(x *ast.Call) [2]int {
	var names []string
	for i, arg := range x.Args {
		if arg.Name != "" && names == nil {
			names = make([]string, len(x.Args))
		}
		if names != nil {
			names[i] = arg.Name
		}
	}
	params := g.matchBuiltin("assert", assertParams, 1, names, len(x.Args), x.Pos)
	slots := [2]int{-1, -1}
	for i, p := range params {
		slots[p] = i
	}
	return slots
}

// evalCond evaluates an assert's condition, recording into rec when it is not nil.
func (g *Goroutine) evalCond(cond ast.Expr, env *Env, rec *assertRec) (value.Value, ctl) {
	saved := g.rec
	g.rec = rec
	defer g.restoreRec(saved)
	return g.evalExpr(cond, env)
}

func (g *Goroutine) restoreRec(r *assertRec) { g.rec = r }

// newRec returns an empty recording for a's parts, reusing one a passing assert gave back.
func (g *Goroutine) newRec(a *ast.Assert) *assertRec {
	var rec *assertRec
	if n := len(g.spare); n > 0 {
		rec = g.spare[n-1]
		g.spare = g.spare[:n-1]
	} else {
		rec = &assertRec{}
	}
	rec.a = a
	rec.vals = slices.Grow(rec.vals[:0], len(a.Parts))[:len(a.Parts)]
	clear(rec.vals)
	rec.order = rec.order[:0]
	return rec
}

// freeRec gives back a recording that is no longer needed.
func (g *Goroutine) freeRec(rec *assertRec) {
	if rec != nil {
		rec.a = nil
		clear(rec.vals)
		g.spare = append(g.spare, rec)
	}
}

// report is the panic message of a failed assert (section 8, assert). rec is nil when the
// condition has no parts.
func report(a *ast.Assert, rec *assertRec, msg string) string {
	if rec == nil {
		rec = &assertRec{}
	}
	var sb strings.Builder
	sb.WriteString("assertion failed: ")
	sb.WriteString(format.Line(a.Text))
	if a.Op != "" && rec.vals != nil && rec.vals[a.X] != nil {
		l := rec.vals[a.X]
		switch {
		case a.Op == "is":
			sb.WriteString(" (" + show(l) + " is " + value.Describe(l) + ")")
		case rec.vals[a.Y] == nil:
		case a.Op == "!=" && isNullLit(a.Parts[a.Y].Expr):
			sb.WriteString(" (" + format.Line(a.Parts[a.X].Text) + " is null)")
		case a.Op == "!=" && isNullLit(a.Parts[a.X].Expr):
			sb.WriteString(" (" + format.Line(a.Parts[a.Y].Text) + " is null)")
		default:
			sb.WriteString(" (" + show(l) + " " + a.Op + " " + show(rec.vals[a.Y]) + ")")
		}
	}
	if msg != "" {
		sb.WriteString(": " + msg)
	}
	seen := map[string]bool{}
	for _, i := range rec.order {
		p := a.Parts[i]
		v := rec.vals[i]
		if !p.Shown || value.KindOf(v) == value.KindFunc {
			continue
		}
		text := format.Line(p.Text)
		if seen[text] {
			continue
		}
		seen[text] = true
		sb.WriteString("\n  " + text + " = " + show(v))
	}
	return sb.String()
}

// show renders a value in a report as String.from renders it inside an array, so that a string
// is quoted.
func show(v value.Value) string { return value.RenderQuoted(v) }

func isNullLit(e ast.Expr) bool {
	_, ok := e.(*ast.NullLit)
	return ok
}
