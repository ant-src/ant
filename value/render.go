package value

import (
	"math"
	"strconv"
	"strings"
)

// Render is String.from (section 11): the rendering print and template strings use.
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
	switch x := v.(type) {
	case int64:
		sb.WriteString(strconv.FormatInt(x, 10))
	case int8:
		sb.WriteString(strconv.FormatInt(int64(x), 10))
	case int16:
		sb.WriteString(strconv.FormatInt(int64(x), 10))
	case int32:
		sb.WriteString(strconv.FormatInt(int64(x), 10))
	case uint8:
		sb.WriteString(strconv.FormatUint(uint64(x), 10))
	case uint16:
		sb.WriteString(strconv.FormatUint(uint64(x), 10))
	case uint32:
		sb.WriteString(strconv.FormatUint(uint64(x), 10))
	case uint64:
		sb.WriteString(strconv.FormatUint(x, 10))
	case float64:
		sb.WriteString(formatFloat(x, 64))
	case float32:
		sb.WriteString(formatFloat(float64(x), 32))
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
		if onStack(stack, x) {
			sb.WriteString("...")
			return
		}
		stack = append(stack, x)
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
		if onStack(stack, x) {
			sb.WriteString("[...]")
			return
		}
		stack = append(stack, x)
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
		if onStack(stack, x) {
			sb.WriteString("Map{...}")
			return
		}
		stack = append(stack, x)
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
	case *Constructor:
		sb.WriteString(x.Tag.Name)
	case *Tag:
		sb.WriteString(x.Name)
	case Callable:
		sb.WriteString("func")
	case Sync:
		sb.WriteString(x.SyncName())
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

func formatFloat(f float64, bits int) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	}
	s := strconv.FormatFloat(f, 'f', -1, bits)
	if !strings.ContainsAny(s, ".") {
		s += ".0"
	}
	return s
}

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
			if r < 0x20 || r == 0x7f {
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
