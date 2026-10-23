package tests

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/value"
)

// TestGenerated builds random typed expressions from a small grammar and checks that the
// checker and the runtime agree. For each case with expression e of type T:
//
//  1. a program binding e to a variable of type T checks clean;
//  2. evaluating e yields a value whose kind matches T, or an Ant panic from a value-dependent
//     row of the panics table;
//  3. after one literal is swapped for a literal of another kind, either the checker reports
//     an error or the run raises no kind panic.
//
// ANT_SEED fixes the seed and ANT_GENERATED the number of cases; a failure prints both and the
// program.
func TestGenerated(t *testing.T) {
	seed := time.Now().UnixNano()
	if s := os.Getenv("ANT_SEED"); s != "" {
		seed, _ = strconv.ParseInt(s, 10, 64)
	}
	cases := 1500
	if testing.Short() {
		cases = 300
	}
	if s := os.Getenv("ANT_GENERATED"); s != "" {
		cases, _ = strconv.Atoi(s)
	}
	t.Logf("ANT_SEED=%d ANT_GENERATED=%d", seed, cases)
	rng := rand.New(rand.NewSource(seed))

	session, err := ant.NewSession(context.Background(), ant.WithStdout(io.Discard), ant.WithStderr(io.Discard))
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	for _, line := range strings.Split(strings.TrimSpace(genDecls), "\n") {
		if _, err := session.Eval(line); err != nil {
			t.Fatalf("declaring %q: %v", line, err)
		}
	}
	g := &generator{rng: rng}
	for i := 0; i < cases; i++ {
		typ := g.pickType()
		expr := g.gen(typ, 3)
		prog := genDecls + "\nfunc main(): void {\n  var x: " + typ.String() + " = " + expr + "\n  print(x)\n}\n"
		fail := func(format string, args ...any) {
			t.Helper()
			t.Fatalf("case %d (ANT_SEED=%d):\n%s\n%s", i, seed, fmt.Sprintf(format, args...), prog)
		}
		if diags := checkProgram(t, prog); len(diags) > 0 {
			fail("the checker rejects a generated program:\n%s", diags)
		}
		v, err := session.Eval(expr)
		if err != nil {
			var pe *ant.PanicError
			if !asPanic(err, &pe) || !allowedPanic(pe.Message) {
				fail("evaluating %s: %v", expr, err)
			}
		} else if !kindMatches(typ, v) {
			fail("the value %s of %s is not a %s", value.Render(v), expr, typ)
		}
		mutated, ok := g.mutate(expr)
		if !ok {
			continue
		}
		mprog := strings.Replace(prog, expr, mutated, 1)
		if diags := checkProgram(t, mprog); len(diags) > 0 {
			continue // rejected, as a mutation usually is
		}
		if _, err := session.Eval(mutated); err != nil {
			var pe *ant.PanicError
			if !asPanic(err, &pe) || !allowedPanic(pe.Message) {
				fail("soundness: the checker accepts the mutation %s, which panics: %v", mutated, err)
			}
		}
	}
}

const genDecls = `
type R1 = { a: int, b: string }
type R2 = { n: float64 }
func inc(x: int): int { x + 1 }
func neg(b: bool): bool { !b }
func len2(s: string): int { String.length(s) }
`

func checkProgram(t *testing.T, src string) string {
	t.Helper()
	prog, err := ant.Load("gen.ant", ant.WithSource("gen.ant", src), ant.WithFS(fstest.MapFS{}))
	if err != nil {
		return err.Error()
	}
	var sb strings.Builder
	for _, d := range prog.Check() {
		sb.WriteString(d.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}

func asPanic(err error, pe **ant.PanicError) bool {
	p, ok := err.(*ant.PanicError)
	if ok {
		*pe = p
	}
	return ok
}

// allowedPanic reports whether a panic message belongs to a panic a checked program can reach: arithmetic overflow, a zero
// divisor, or an index out of range, which a checked program may reach.
func allowedPanic(msg string) bool {
	for _, s := range []string{"overflow", "zero", "out of range", "does not fit", "fewer"} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

// ---- types

type gtype struct {
	kind    string // int, float64, string, bool, null, array, record, union
	elem    *gtype
	name    string
	members []*gtype
}

var (
	gInt    = &gtype{kind: "int"}
	gFloat  = &gtype{kind: "float64"}
	gString = &gtype{kind: "string"}
	gBool   = &gtype{kind: "bool"}
	gNull   = &gtype{kind: "null"}
	gR1     = &gtype{kind: "record", name: "R1"}
	gR2     = &gtype{kind: "record", name: "R2"}
)

func (t *gtype) String() string {
	switch t.kind {
	case "array":
		return "Array[" + t.elem.String() + "]"
	case "record":
		return t.name
	case "union":
		var parts []string
		for _, m := range t.members {
			parts = append(parts, m.String())
		}
		return strings.Join(parts, " | ")
	}
	return t.kind
}

func (t *gtype) isNumber() bool { return t.kind == "int" || t.kind == "float64" }

func (t *gtype) base() *gtype { // the T of T | null
	if t.kind == "union" && len(t.members) == 2 && t.members[1] == gNull {
		return t.members[0]
	}
	return nil
}

func nullable(t *gtype) *gtype { return &gtype{kind: "union", members: []*gtype{t, gNull}} }
func array(t *gtype) *gtype    { return &gtype{kind: "array", elem: t} }

var gShapes = &gtype{kind: "union", members: []*gtype{gR1, gR2}}

// ---- generation

type generator struct {
	rng *rand.Rand
}

func (g *generator) pick(n int) int { return g.rng.Intn(n) }

func (g *generator) pickType() *gtype {
	switch g.pick(10) {
	case 0, 1:
		return gInt
	case 2:
		return gFloat
	case 3:
		return gString
	case 4:
		return gBool
	case 5:
		return array(g.pickBase())
	case 6:
		return gR1
	case 7:
		return nullable(g.pickBase())
	case 8:
		return gShapes
	}
	return gR2
}

func (g *generator) pickBase() *gtype {
	return []*gtype{gInt, gFloat, gString, gBool}[g.pick(4)]
}

func (g *generator) literal(t *gtype) string {
	switch t.kind {
	case "int":
		return strconv.Itoa(g.pick(7) - 1)
	case "float64":
		return strconv.FormatFloat(float64(g.pick(9))/2-1, 'f', 1, 64)
	case "string":
		return `"` + []string{"", "a", "bc", "héllo"}[g.pick(4)] + `"`
	case "bool":
		return []string{"true", "false"}[g.pick(2)]
	case "null":
		return "null"
	case "array":
		n := 1 + g.pick(2) // never empty: an empty literal needs an expected type
		var elems []string
		for i := 0; i < n; i++ {
			elems = append(elems, g.literal(t.elem))
		}
		return "[" + strings.Join(elems, ", ") + "]"
	case "record":
		if t.name == "R1" {
			return "R1{ a: " + g.literal(gInt) + ", b: " + g.literal(gString) + " }"
		}
		return "R2{ n: " + g.literal(gFloat) + " }"
	case "union":
		return g.literal(t.members[g.pick(len(t.members))])
	}
	panic("unknown type " + t.kind)
}

// gen returns an expression of type t.
func (g *generator) gen(t *gtype, depth int) string {
	if depth <= 0 {
		return g.literal(t)
	}
	d := depth - 1
	cond := func() string { return g.gen(gBool, d) }
	ifElse := func(a, b string) string { return "(if " + cond() + " { " + a + " } else { " + b + " })" }
	switch t.kind {
	case "int":
		switch g.pick(9) {
		case 0:
			return g.literal(t)
		case 1:
			return "(" + g.gen(t, d) + " " + []string{"+", "-", "*", "/", "%"}[g.pick(5)] + " " + g.gen(t, d) + ")"
		case 2:
			return "Array.length(" + g.gen(array(g.pickBase()), d) + ")"
		case 3:
			return "inc(" + g.gen(t, d) + ")"
		case 4:
			return g.gen(gR1, d) + ".a"
		case 5:
			return ifElse(g.gen(t, d), g.gen(t, d))
		case 6:
			return "(" + g.nullableTyped(t, d, false) + " ?? " + g.gen(t, d) + ")"
		case 7:
			return "-" + g.gen(t, d)
		}
		return "len2(" + g.gen(gString, d) + ")"
	case "float64":
		switch g.pick(6) {
		case 0:
			return g.literal(t)
		case 1:
			return "(" + g.gen(t, d) + " " + []string{"+", "-", "*", "/"}[g.pick(4)] + " " + g.gen(t, d) + ")"
		case 2:
			return g.gen(gR2, d) + ".n"
		case 3:
			return ifElse(g.gen(t, d), g.gen(t, d))
		case 4:
			return "(" + g.nullableTyped(t, d, false) + " ?? " + g.gen(t, d) + ")"
		}
		return "float64(" + g.gen(gInt, d) + ")"
	case "string":
		switch g.pick(6) {
		case 0:
			return g.literal(t)
		case 1:
			return "(" + g.gen(t, d) + " + " + g.gen(t, d) + ")"
		case 2:
			return g.gen(gR1, d) + ".b"
		case 3:
			return ifElse(g.gen(t, d), g.gen(t, d))
		case 4:
			return "String.from(" + g.gen(g.pickType(), d) + ")"
		}
		return "`${" + g.gen(g.pickBase(), d) + "}`"
	case "bool":
		switch g.pick(9) {
		case 0:
			return g.literal(t)
		case 1:
			o := g.pickOrdered()
			return "(" + g.gen(o, d) + " " + []string{"<", "<=", ">", ">="}[g.pick(4)] + " " + g.gen(o, d) + ")"
		case 2:
			// one side carries the full type, so the other is assignable to it
			e := g.pickType()
			l := g.gen(e, d)
			switch {
			case e.base() != nil:
				l = g.nullableTyped(e.base(), d, true)
			case e == gShapes:
				l = g.unionTyped(d)
			}
			return "(" + l + " " + []string{"==", "!="}[g.pick(2)] + " " + g.gen(e, d) + ")"
		case 3:
			return "(" + g.gen(t, d) + " " + []string{"&&", "||"}[g.pick(2)] + " " + g.gen(t, d) + ")"
		case 4:
			return "!" + g.gen(t, d)
		case 5:
			// the operand must have the union as its static type, not one member of it
			return "(" + g.unionTyped(d) + " is " + []string{"R1", "R2"}[g.pick(2)] + ")"
		case 6:
			return "neg(" + g.gen(t, d) + ")"
		case 7:
			return ifElse(g.gen(t, d), g.gen(t, d))
		}
		return "Array.isEmpty(" + g.gen(array(g.pickBase()), d) + ")"
	case "null":
		return "null"
	case "array":
		switch g.pick(5) {
		case 0, 1:
			return "[" + g.gen(t.elem, d) + ", " + g.gen(t.elem, d) + "]"
		case 2:
			return "Array.concat(" + g.gen(t, d) + ", " + g.gen(t, d) + ")"
		case 3:
			return "Array.reverse(" + g.gen(t, d) + ")"
		}
		return ifElse(g.gen(t, d), g.gen(t, d))
	case "record":
		switch g.pick(3) {
		case 0:
			return ifElse(g.gen(t, d), g.gen(t, d))
		case 1:
			if t.name == "R1" {
				return "R1{ a: " + g.gen(gInt, d) + ", b: " + g.gen(gString, d) + " }"
			}
			return "R2{ n: " + g.gen(gFloat, d) + " }"
		}
		return g.literal(t)
	case "union":
		if b := t.base(); b != nil {
			if g.pick(5) == 0 {
				return g.gen(b, d) // a T is a T | null, though not a nullable operand for ??
			}
			return g.nullableTyped(b, d, false)
		}
		// R1 | R2
		switch g.pick(3) {
		case 0:
			return ifElse(g.gen(gR1, d), g.gen(gR2, d))
		case 1:
			return g.gen(t.members[g.pick(len(t.members))], d)
		}
		return "(" + g.gen(array(t), 0) + "[0])"
	}
	panic("unknown type " + t.kind)
}

func (g *generator) pickOrdered() *gtype { return []*gtype{gInt, gFloat, gString}[g.pick(3)] }

// nullableTyped returns an expression whose static type includes null, as ?? needs; with full
// set, the type is b | null itself rather than possibly plain null.
func (g *generator) nullableTyped(b *gtype, d int, full bool) string {
	switch g.pick(4) {
	case 0:
		if !full {
			return "null"
		}
		return "(if " + g.gen(gBool, d) + " { null } else { " + g.gen(b, d) + " })"
	case 1:
		return "(if " + g.gen(gBool, d) + " { " + g.gen(b, d) + " } else { null })"
	case 2:
		return "Array.get(" + g.gen(array(b), d) + ", " + g.gen(gInt, d) + ")"
	}
	return "Array.find(" + g.gen(array(b), d) + ", { " + g.gen(gBool, 0) + " })"
}

// unionTyped returns an expression whose static type is R1 | R2, as an is test needs.
func (g *generator) unionTyped(d int) string {
	if g.pick(2) == 0 {
		return "(if " + g.gen(gBool, d) + " { " + g.gen(gR1, d) + " } else { " + g.gen(gR2, d) + " })"
	}
	// a literal with both members has the union as its element type
	return "[" + g.literal(gR1) + ", " + g.literal(gR2) + "][" + strconv.Itoa(g.pick(2)) + "]"
}

var (
	intLit    = regexp.MustCompile(`(^|[^\w."])(\d+)(?:[^\w.]|$)`)
	stringLit = regexp.MustCompile(`"[^"]*"`)
	boolLit   = regexp.MustCompile(`\b(true|false)\b`)
)

// mutate swaps one literal of expr for a literal of another kind.
func (g *generator) mutate(expr string) (string, bool) {
	type hit struct {
		start, end int
		repl       string
	}
	var hits []hit
	for _, m := range intLit.FindAllStringSubmatchIndex(expr, -1) {
		hits = append(hits, hit{m[4], m[5], []string{`"s"`, "true", "1.5", "null"}[g.pick(4)]})
	}
	for _, m := range stringLit.FindAllStringIndex(expr, -1) {
		hits = append(hits, hit{m[0], m[1], []string{"7", "false", "2.0", "null"}[g.pick(4)]})
	}
	for _, m := range boolLit.FindAllStringIndex(expr, -1) {
		hits = append(hits, hit{m[0], m[1], []string{"0", `"t"`, "null"}[g.pick(3)]})
	}
	if len(hits) == 0 {
		return "", false
	}
	h := hits[g.pick(len(hits))]
	return expr[:h.start] + h.repl + expr[h.end:], true
}

// ---- kinds

func kindMatches(t *gtype, v value.Value) bool {
	switch t.kind {
	case "int":
		return value.KindOf(v) == value.KindInt64
	case "float64":
		return value.KindOf(v) == value.KindFloat64
	case "string":
		return value.KindOf(v) == value.KindString
	case "bool":
		return value.KindOf(v) == value.KindBool
	case "null":
		return value.KindOf(v) == value.KindNull
	case "array":
		a, ok := v.(*value.Array)
		if !ok {
			return false
		}
		for _, e := range a.Values() {
			if !kindMatches(t.elem, e) {
				return false
			}
		}
		return true
	case "record":
		r, ok := v.(*value.Record)
		return ok && r.Tag.Name == t.name
	case "union":
		for _, m := range t.members {
			if kindMatches(m, v) {
				return true
			}
		}
		return false
	}
	return false
}
