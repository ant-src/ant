package interp

import (
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
	if a.Bad != "" {
		g.panicf(x.Pos, "%s", a.Bad)
	}
	rec := &assertRec{a: a, vals: make([]value.Value, len(a.Parts))}
	saved := g.rec
	g.rec = rec
	v, c := func() (value.Value, ctl) {
		defer func() { g.rec = saved }()
		return g.evalExpr(a.Cond, env)
	}()
	if c != ctlNone {
		return v, c
	}
	ok, isBool := v.(bool)
	if !isBool {
		g.panicf(a.Cond.Position(), "an assert condition must be a bool, got %s", value.Describe(v))
	}
	if ok {
		return value.Void, ctlNone
	}
	msg := ""
	if a.Message != nil {
		m, c := g.eval(a.Message, env)
		if c != ctlNone {
			return m, c
		}
		s, isString := m.(string)
		if !isString {
			g.panicf(a.Message.Position(), "assert needs a string message, got %s", value.Describe(m))
		}
		msg = s
	}
	g.panicf(x.Pos, "%s", rec.report(msg))
	return nil, ctlNone
}

// report is the panic message of a failed assert (section 8, assert).
func (rec *assertRec) report(msg string) string {
	a := rec.a
	var sb strings.Builder
	sb.WriteString("assertion failed: ")
	sb.WriteString(format.Line(a.Text))
	if a.Op != "" && rec.vals[a.X] != nil {
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
