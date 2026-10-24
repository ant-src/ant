package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// Json (stdlib.md): parse produces Json.Value trees with float64 numbers and maps in document
// order; stringify renders any value, records as objects without their tag.

func init() {
	register("Json", func() value.Value {
		return namespace("Json",
			constant("Value", value.NewTag("Json.Value", value.AliasTag, nil)),
			fn("Json.parse", "text", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				dec := json.NewDecoder(strings.NewReader(stringArg(g, who, a[0])))
				v, err := parseJSON(dec)
				if err != nil {
					return value.Err(value.ErrorValue(parseError(err)))
				}
				if _, err := dec.Token(); err != io.EOF {
					return value.Err(value.ErrorValue("unexpected text after the JSON value"))
				}
				return value.Ok(v)
			}),
			fn("Json.stringify", "value", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				var sb strings.Builder
				if msg := stringify(&sb, a[0], nil); msg != "" {
					g.Panic("Json.stringify: %s", msg)
				}
				return sb.String()
			}),
		)
	})
}

// parseError is the message of a decoding error. Truncated input gets one message whatever Go
// version built the binary: encoding/json reports an unexpected end as io.EOF, as
// io.ErrUnexpectedEOF or as a SyntaxError depending on where the text stops and on the version.
func parseError(err error) string {
	var syn *json.SyntaxError
	if err == io.EOF || err == io.ErrUnexpectedEOF ||
		(errors.As(err, &syn) && strings.HasPrefix(syn.Error(), "unexpected end")) {
		return "unexpected end of JSON input"
	}
	return err.Error()
}

func parseJSON(dec *json.Decoder) (value.Value, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	return parseFrom(dec, tok)
}

func parseFrom(dec *json.Decoder, tok json.Token) (value.Value, error) {
	switch t := tok.(type) {
	case nil:
		return value.Null, nil
	case bool:
		return t, nil
	case float64:
		return t, nil
	case string:
		return t, nil
	case json.Delim:
		switch t {
		case '[':
			var elems []value.Value
			for dec.More() {
				v, err := parseJSON(dec)
				if err != nil {
					return nil, err
				}
				elems = append(elems, v)
			}
			if _, err := dec.Token(); err != nil {
				return nil, err
			}
			return value.NewArray(elems), nil
		case '{':
			m := value.NewMap()
			for dec.More() {
				k, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, ok := k.(string)
				if !ok {
					return nil, fmt.Errorf("JSON object keys must be strings")
				}
				v, err := parseJSON(dec)
				if err != nil {
					return nil, err
				}
				m.Set(key, v)
			}
			if _, err := dec.Token(); err != nil {
				return nil, err
			}
			return m, nil
		}
	}
	return nil, fmt.Errorf("unexpected JSON token %v", tok)
}

// stringify writes v as JSON. It returns why v has no JSON form, a cycle or Bytes, or "" when it
// has one.
func stringify(sb *strings.Builder, v value.Value, stack []any) string {
	switch x := v.(type) {
	case int64, int8, int16, int32, uint8, uint16, uint32, uint64:
		sb.WriteString(value.Render(x))
	case float64:
		writeFloat(sb, x, 64)
	case float32:
		writeFloat(sb, float64(x), 32)
	case string:
		writeJSONString(sb, x)
	case bool:
		sb.WriteString(strconv.FormatBool(x))
	case *value.Tagged:
		return stringify(sb, x.Payload, stack)
	case *value.Record:
		if onStack(stack, x) {
			return "cyclic value"
		}
		stack = append(stack, x)
		sb.WriteByte('{')
		for i, name := range x.Tag.Fields {
			if i > 0 {
				sb.WriteByte(',')
			}
			writeJSONString(sb, name)
			sb.WriteByte(':')
			if msg := stringify(sb, x.Get(i), stack); msg != "" {
				return msg
			}
		}
		sb.WriteByte('}')
	case *value.Array, *value.Set:
		if onStack(stack, x) {
			return "cyclic value"
		}
		stack = append(stack, x)
		var xs []value.Value
		if a, ok := x.(*value.Array); ok {
			xs = a.Values()
		} else {
			xs = x.(*value.Set).Values()
		}
		sb.WriteByte('[')
		for i, e := range xs {
			if i > 0 {
				sb.WriteByte(',')
			}
			if msg := stringify(sb, e, stack); msg != "" {
				return msg
			}
		}
		sb.WriteByte(']')
	case *value.Map:
		if onStack(stack, x) {
			return "cyclic value"
		}
		stack = append(stack, x)
		sb.WriteByte('{')
		for i, e := range x.Entries() {
			if i > 0 {
				sb.WriteByte(',')
			}
			key := e.Key
			if t, ok := key.(*value.Tagged); ok {
				key = t.Payload // a named primitive key, without its tag as a value is written
			}
			if holdsBytes(key, nil) {
				return bytesNoJSON
			}
			if s, ok := key.(string); ok {
				writeJSONString(sb, s)
			} else {
				writeJSONString(sb, value.Render(key))
			}
			sb.WriteByte(':')
			if msg := stringify(sb, e.Value, stack); msg != "" {
				return msg
			}
		}
		sb.WriteByte('}')
	case value.Bytes:
		return bytesNoJSON
	case value.Time:
		writeJSONString(sb, x.String())
	case *value.Regex:
		writeJSONString(sb, x.String())
	default:
		sb.WriteString("null")
	}
	return ""
}

// bytesNoJSON is why Json.stringify panics on Bytes: JSON has no form for them (stdlib.md, Json).
const bytesNoJSON = "Bytes has no JSON form; write Encoding.toBase64 of it"

// holdsBytes reports whether v is Bytes or contains them, for a map key, which stringify writes
// through value.Render rather than taking apart. stack holds the containers being searched, so a
// cyclic key terminates.
func holdsBytes(v value.Value, stack []any) bool {
	var parts []value.Value
	switch x := v.(type) {
	case value.Bytes:
		return true
	case *value.Tagged:
		return holdsBytes(x.Payload, stack)
	case *value.Array:
		parts = x.Values()
	case *value.Set:
		parts = x.Values()
	case *value.Record:
		for i := 0; i < x.Len(); i++ {
			parts = append(parts, x.Get(i))
		}
	case *value.Map:
		for _, e := range x.Entries() {
			parts = append(parts, e.Key, e.Value)
		}
	default:
		return false
	}
	if onStack(stack, v) {
		return false
	}
	stack = append(stack, v)
	for _, p := range parts {
		if holdsBytes(p, stack) {
			return true
		}
	}
	return false
}

func onStack(stack []any, v any) bool {
	for _, s := range stack {
		if s == v {
			return true
		}
	}
	return false
}

func writeFloat(sb *strings.Builder, f float64, bits int) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		sb.WriteString("null")
		return
	}
	sb.WriteString(value.FormatFloat(f, bits))
}

func writeJSONString(sb *strings.Builder, s string) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.Encode(s)
	sb.WriteString(strings.TrimRight(buf.String(), "\n"))
}
