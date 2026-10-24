package value

import (
	"math"
	"strconv"
	"strings"
	"unicode"
)

// Render is String.from (section 11): the rendering print, println and template strings use.
func Render(v Value) string {
	var sb strings.Builder
	render(&sb, v, true, nil)
	return sb.String()
}

// RenderQuoted renders v as it would appear inside a container: strings quoted with escapes.
func RenderQuoted(v Value) string {
	var sb strings.Builder
	render(&sb, v, false, nil)
	return sb.String()
}

func render(sb *strings.Builder, v Value, top bool, stack []any) {
	switch v.(type) {
	case *Record, *Array, *Map, *Set:
		// A container inside itself renders as ... at the inner occurrence.
		if onStack(stack, v) {
			sb.WriteString("...")
			return
		}
		stack = append(stack, v)
	}
	switch x := v.(type) {
	case int64, int8, int16, int32:
		sb.WriteString(strconv.FormatInt(toInt64(v), 10))
	case uint8, uint16, uint32, uint64:
		sb.WriteString(strconv.FormatUint(toUint64(v), 10))
	case float64:
		sb.WriteString(FormatFloat(x, 64))
	case float32:
		sb.WriteString(FormatFloat(float64(x), 32))
	case string:
		if top {
			sb.WriteString(x)
		} else {
			quote(sb, x)
		}
	case bool:
		sb.WriteString(strconv.FormatBool(x))
	case nullType:
		sb.WriteString("null")
	case voidType:
		sb.WriteString("void")
	case *Tagged:
		sb.WriteString(x.Tag.Name)
		sb.WriteByte('(')
		render(sb, x.Payload, false, stack)
		sb.WriteByte(')')
	case *Record:
		sb.WriteString(x.Tag.Name)
		if x.Len() == 0 {
			sb.WriteString("{}")
			return
		}
		sb.WriteString("{ ")
		for i, name := range x.Tag.Fields {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(name)
			sb.WriteString(": ")
			render(sb, x.Get(i), false, stack)
		}
		sb.WriteString(" }")
	case *Array:
		xs := x.Values()
		sb.WriteByte('[')
		for i, e := range xs {
			if i > 0 {
				sb.WriteString(", ")
			}
			render(sb, e, false, stack)
		}
		sb.WriteByte(']')
	case *Map:
		es := x.Entries()
		if len(es) == 0 {
			sb.WriteString("Map{}")
			return
		}
		sb.WriteString("Map{ ")
		for i, e := range es {
			if i > 0 {
				sb.WriteString(", ")
			}
			render(sb, e.Key, false, stack)
			sb.WriteString(": ")
			render(sb, e.Value, false, stack)
		}
		sb.WriteString(" }")
	case *Set:
		sb.WriteString("Set.of([")
		for i, e := range x.Values() {
			if i > 0 {
				sb.WriteString(", ")
			}
			render(sb, e, false, stack)
		}
		sb.WriteString("])")
	case *Constructor:
		sb.WriteString(x.Tag.Name)
	case *Tag:
		sb.WriteString(x.Name)
	case Callable:
		sb.WriteString("func")
	case Sync:
		sb.WriteString(x.SyncName())
	case Time:
		sb.WriteString(x.String())
	case *Regex:
		sb.WriteString(x.String())
	case uninitType:
		sb.WriteString("<uninitialized>")
	default:
		sb.WriteString("<invalid>")
	}
}

func onStack(stack []any, v any) bool {
	for _, s := range stack {
		if s == v {
			return true
		}
	}
	return false
}

// FormatFloat renders a float as section 11 says: the shortest decimal that reads back to the
// same value, with .0 when integral, and in JavaScript's exponent form, 1e+21 or 1.5e-7, when the
// magnitude is at or above 1e21 or below 1e-6. bits is 32 for a float32.
func FormatFloat(f float64, bits int) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	}
	if abs := math.Abs(f); abs != 0 && (abs < 1e-6 || abs >= 1e21) {
		s := strconv.FormatFloat(f, 'e', -1, bits)
		// Go pads the exponent to two digits; JavaScript does not: 1e-7, not 1e-07.
		if i := strings.LastIndexAny(s, "+-"); i > 0 && len(s) > i+2 && s[i+1] == '0' {
			s = s[:i+1] + s[i+2:]
		}
		return s
	}
	s := strconv.FormatFloat(f, 'f', -1, bits)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

// quote writes s as a string literal: every code point that String.isPrint rejects, such as a
// control character, a non-breaking space, or a zero-width joiner, as an escape, so the rendering
// shows every code point there is (section 11, String).
func quote(sb *strings.Builder, s string) {
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\n':
			sb.WriteString(`\n`)
		case '\t':
			sb.WriteString(`\t`)
		case '\r':
			sb.WriteString(`\r`)
		default:
			if !unicode.IsPrint(r) {
				sb.WriteString(`\u{`)
				sb.WriteString(strconv.FormatInt(int64(r), 16))
				sb.WriteByte('}')
			} else {
				sb.WriteRune(r)
			}
		}
	}
	sb.WriteByte('"')
}
