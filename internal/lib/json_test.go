package lib

import (
	"runtime/debug"
	"strings"
	"testing"

	"github.com/ant-src/ant/value"
)

// TestStringifyDeep writes a value and checks a map key nested far deeper than one small Go stack
// holds, so both walks must move to fresh goroutines as they go down (section 9: no crash).
func TestStringifyDeep(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(64 << 20))
	var deep value.Value = value.Bytes("x")
	for i := 0; i < 300_000; i++ {
		deep = value.NewArray([]value.Value{deep})
	}
	m := value.NewMap()
	m.Set(deep, int64(1))
	var sb strings.Builder
	if msg := stringify(&sb, m, ancestors{}); msg != bytesNoJSON {
		t.Errorf("a map keyed by deeply nested Bytes stringifies with %q", msg)
	}
	sb.Reset()
	if msg := stringify(&sb, deep, ancestors{}); msg != bytesNoJSON {
		t.Errorf("deeply nested Bytes stringifies with %q", msg)
	}
}
