package lib

import (
	"bytes"
	"encoding/json"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/ant-src/ant/value"
)

// TestStringifyDeep writes a value nested far deeper than one small Go stack holds, so the walk
// must move to fresh goroutines as it goes down (section 9: no crash), and a map keyed by one,
// which panics on the key without walking it.
func TestStringifyDeep(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(64 << 20))
	var deep value.Value = value.Bytes("x")
	for i := 0; i < 300_000; i++ {
		deep = value.NewArray([]value.Value{deep})
	}
	m := value.NewMap()
	m.Set(deep, int64(1))
	var sb strings.Builder
	if msg := stringify(&sb, m, value.Ancestors{}, ""); msg != "a map key must be a string, got array" {
		t.Errorf("a map keyed by deeply nested Bytes stringifies with %q", msg)
	}
	sb.Reset()
	if msg := stringify(&sb, deep, value.Ancestors{}, ""); msg != bytesNoJSON {
		t.Errorf("deeply nested Bytes stringifies with %q", msg)
	}
}

// TestWriteJSONString checks writeJSONString against encoding/json without HTML escaping, which
// it replaced and whose output it keeps.
func TestWriteJSONString(t *testing.T) {
	var all strings.Builder
	for r := rune(0); r < 0x80; r++ {
		all.WriteRune(r)
	}
	for _, s := range []string{
		"", all.String(), "\u00e9\U0001F600\u2028\u2029\u00a0\ufeff<>&", "\uFFFD", "a\xffb\xc3", "\x7f\u0080\u009f",
	} {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.Encode(s)
		var sb strings.Builder
		writeJSONString(&sb, s)
		if want := strings.TrimSuffix(buf.String(), "\n"); sb.String() != want {
			t.Errorf("writeJSONString(%q) = %s, want %s", s, sb.String(), want)
		}
	}
}
