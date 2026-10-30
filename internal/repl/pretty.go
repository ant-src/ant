package repl

import (
	"fmt"
	"strings"

	"github.com/ant-src/ant/value"
)

// prettyMaxItems is how many elements of an array, a map, or a set a result shows.
const prettyMaxItems = 100

// pretty renders a result for the terminal: as String.from does, colored by th, and, when it
// does not fit in width columns, with each element of a record, an array, a map, or a set on
// a line of its own, indented. A long array, map, or set shows its first prettyMaxItems.
func (th *theme) pretty(v value.Value, width int) string {
	if s, ok := v.(string); ok {
		return s // a string prints as print prints it
	}
	p := &printer{th: th, seen: map[value.Value]bool{}}
	return p.render(v, 0, width)
}

type printer struct {
	th   *theme
	seen map[value.Value]bool // the containers being rendered, which a cycle meets again
}

// render renders v, indented by indent spaces after its first line, in width columns.
func (p *printer) render(v value.Value, indent, width int) string {
	flat := p.flat(v, false)
	if stringWidth(flat) <= width-indent || !container(v) {
		return p.flat(v, true)
	}
	if p.seen[v] {
		return "..."
	}
	p.seen[v] = true
	defer delete(p.seen, v)
	open, close, items := p.parts(v)
	if len(items) == 0 {
		return p.flat(v, true)
	}
	pad := strings.Repeat(" ", indent+2)
	var sb strings.Builder
	sb.WriteString(open + "\n")
	for _, it := range items {
		sb.WriteString(pad)
		if it.more != "" {
			sb.WriteString(p.color(clsComment, it.more))
		} else {
			if it.name != "" || it.key != nil {
				sb.WriteString(it.keyText(p, indent+2, width))
				sb.WriteString(": ")
			}
			sb.WriteString(p.render(it.val, indent+2, width))
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(strings.Repeat(" ", indent) + close)
	return sb.String()
}

// item is an element of a container: a value with a field name or a map key, or the count of
// the elements not shown.
type item struct {
	name string
	key  value.Value // a map's key, or nil
	val  value.Value
	more string
}

func (it item) keyText(p *printer, indent, width int) string {
	if it.name != "" {
		return it.name
	}
	return p.render(it.key, indent, width)
}

func container(v value.Value) bool {
	switch x := v.(type) {
	case *value.Record:
		return x.Len() > 0
	case *value.Array, *value.Map, *value.Set:
		return true
	}
	return false
}

// parts splits a container into its brackets and its elements, colored.
func (p *printer) parts(v value.Value) (open, close string, items []item) {
	limit := func(n int) int { return min(n, prettyMaxItems) }
	more := func(n int) []item {
		if n > prettyMaxItems {
			return []item{{more: fmt.Sprintf("… %d more", n-prettyMaxItems)}}
		}
		return nil
	}
	switch x := v.(type) {
	case *value.Record:
		for i, name := range x.Tag.Fields {
			items = append(items, item{name: name, val: x.Get(i)})
		}
		return p.color(clsType, x.Tag.Name) + "{", "}", items
	case *value.Array:
		xs := x.Values()
		for _, e := range xs[:limit(len(xs))] {
			items = append(items, item{val: e})
		}
		return "[", "]", append(items, more(len(xs))...)
	case *value.Map:
		es := x.Entries()
		for _, e := range es[:limit(len(es))] {
			items = append(items, item{key: e.Key, val: e.Value})
		}
		return p.color(clsType, "Map") + "{", "}", append(items, more(len(es))...)
	case *value.Set:
		xs := x.Values()
		for _, e := range xs[:limit(len(xs))] {
			items = append(items, item{val: e})
		}
		return p.color(clsType, "Set") + ".of([", "])", append(items, more(len(xs))...)
	}
	return "", "", nil
}

// flat renders v on one line, as String.from renders it inside a container, colored when
// colored is set, and with the long containers cut short.
func (p *printer) flat(v value.Value, colored bool) string {
	var sb strings.Builder
	p.flatTo(&sb, v, colored)
	return sb.String()
}

func (p *printer) flatTo(sb *strings.Builder, v value.Value, colored bool) {
	paint := func(c class, s string) {
		if colored {
			s = p.color(c, s)
		}
		sb.WriteString(s)
	}
	if container(v) || isRecord(v) {
		if p.seen[v] {
			sb.WriteString("...")
			return
		}
		p.seen[v] = true
		defer delete(p.seen, v)
	}
	switch x := v.(type) {
	case int64, int8, int16, int32, uint8, uint16, uint32, uint64, float64, float32:
		paint(clsNumber, value.RenderQuoted(v))
	case string:
		paint(clsString, value.RenderQuoted(v))
	case bool:
		paint(clsConstant, value.RenderQuoted(v))
	case *value.Record:
		paint(clsType, x.Tag.Name)
		if x.Len() == 0 {
			sb.WriteString("{}")
			return
		}
		sb.WriteString("{ ")
		for i, name := range x.Tag.Fields {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(name + ": ")
			p.flatTo(sb, x.Get(i), colored)
		}
		sb.WriteString(" }")
	case *value.Array, *value.Map, *value.Set:
		open, close, items := p.parts(v)
		if !colored {
			open, close, _ = (&printer{th: &plainTheme}).parts(v)
		}
		if _, isMap := v.(*value.Map); isMap && len(items) > 0 {
			open += " "
			close = " " + close
		}
		sb.WriteString(open)
		for i, it := range items {
			if i > 0 {
				sb.WriteString(", ")
			}
			switch {
			case it.more != "":
				paint(clsComment, it.more)
			case it.key != nil:
				p.flatTo(sb, it.key, colored)
				sb.WriteString(": ")
				p.flatTo(sb, it.val, colored)
			default:
				p.flatTo(sb, it.val, colored)
			}
		}
		sb.WriteString(close)
	default:
		if v == value.Null || v == value.Void {
			paint(clsConstant, value.RenderQuoted(v))
			return
		}
		sb.WriteString(value.RenderQuoted(v))
	}
}

// isRecord reports whether v is a record, empty or not, which a cycle can meet.
func isRecord(v value.Value) bool {
	_, ok := v.(*value.Record)
	return ok
}

// color paints s with the class's color, or leaves it when the theme has none.
func (p *printer) color(c class, s string) string {
	if p.th.class[c] == "" {
		return s
	}
	return p.th.class[c] + s + reset
}
