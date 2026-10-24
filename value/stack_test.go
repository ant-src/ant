package value

import (
	"runtime/debug"
	"strings"
	"testing"
)

// TestDeepValues compares and renders lists nested far deeper than one small Go stack holds, so
// the walks must move to fresh goroutines as they go down (section 9: no crash).
func TestDeepValues(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(64 << 20))
	node := NewTag("Node", RecordTag, []string{"next", "v"})
	list := func(n int) Value {
		var l Value = Null
		for i := 0; i < n; i++ {
			l = NewRecord(node, []Value{l, int64(i)})
		}
		return l
	}
	a, b := list(300_000), list(300_000)
	if !Equal(a, b) {
		t.Error("two equal deep lists are unequal")
	}
	if Equal(a, NewRecord(node, []Value{b, int64(0)})) {
		t.Error("lists of different lengths are equal")
	}
	if s := Render(a); !strings.HasPrefix(s, "Node{ next: Node{") {
		t.Errorf("rendering starts %.40q", s)
	}
}
