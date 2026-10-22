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
			fn("Json.parse", "text", func(g *interp.Goroutine, a []value.Value) value.Value {
				dec := json.NewDecoder(strings.NewReader(stringArg(g, "Json.parse", a[0])))
				v, err := parseJSON(dec)
				if err != nil {
					return value.Err(value.ErrorValue(parseError(err)))
				}
				if _, err := dec.Token(); err != io.EOF {
					return value.Err(value.ErrorValue("unexpected text after the JSON value"))
				}
				return value.Ok(v)
			}),
			fn("Json.stringify", "value", func(g *interp.Goroutine, a []value.Value) value.Value {
				var sb strings.Builder
				if !stringify(&sb, a[0], nil) {
					g.Panic("Json.stringify: cyclic value")
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

// stringify writes v as JSON. It returns false on a cyclic value, which has no JSON form.
func stringify(sb *strings.Builder, v value.Value, stack []any) bool {
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
			return false
		}
		stack = append(stack, x)
		sb.WriteByte('{')
		for i, name := range x.Tag.Fields {
			if i > 0 {
				sb.WriteByte(',')
			}
			writeJSONString(sb, name)
			sb.WriteByte(':')
			if !stringify(sb, x.Get(i), stack) {
				return false
			}
		}
		sb.WriteByte('}')
	case *value.Array:
		if onStack(stack, x) {
			return false
		}
		stack = append(stack, x)
		sb.WriteByte('[')
		for i, e := range x.Values() {
			if i > 0 {
				sb.WriteByte(',')
			}
			if !stringify(sb, e, stack) {
				return false
			}
		}
		sb.WriteByte(']')
	case *value.Map:
		if onStack(stack, x) {
			return false
		}
		stack = append(stack, x)
		sb.WriteByte('{')
		for i, e := range x.Entries() {
			if i > 0 {
				sb.WriteByte(',')
			}
			if s, ok := e.Key.(string); ok {
				writeJSONString(sb, s)
			} else {
				writeJSONString(sb, value.Render(e.Key))
			}
			sb.WriteByte(':')
			if !stringify(sb, e.Value, stack) {
				return false
			}
		}
		sb.WriteByte('}')
	default:
		sb.WriteString("null")
	}
	return true
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
