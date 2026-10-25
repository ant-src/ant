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

// TestDeepCycles compares rings longer than one goroutine's share of a walk: a cycle that crosses
// a hop to a fresh goroutine must still end the comparison, with the same result.
func TestDeepCycles(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(64 << 20))
	node := NewTag("Node", RecordTag, []string{"next", "v"})
	ring := func(n int, last int64) Value {
		head := NewRecord(node, []Value{Null, int64(0)})
		tail := head
		for i := 1; i < n; i++ {
			v := int64(i)
			if i == n-1 {
				v = last
			}
			r := NewRecord(node, []Value{Null, v})
			tail.Set(0, r)
			tail = r
		}
		tail.Set(0, head)
		return head
	}
	n := 3*stackChunk + 5
	if !Equal(ring(n, 1), ring(n, 1)) {
		t.Error("two equal deep rings are unequal")
	}
	if Equal(ring(n, 1), ring(n, 2)) {
		t.Error("rings differing in the last node are equal")
	}
	if s := Render(ring(n, 1)); !strings.Contains(s, "next: ...") {
		t.Errorf("a deep ring renders without ...")
	}
}
