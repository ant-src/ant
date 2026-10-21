package value

import (
	"math"
	"sync"
	"testing"
)

type testFunc struct{ name string }

func (f *testFunc) FuncName() string { return f.name }

var point = NewTag("Point", RecordTag, []string{"x", "y"})
var userID = NewPrimitiveTag("UserId", KindString)

func rec(tag *Tag, vals ...Value) *Record { return NewRecord(tag, vals) }

func TestRender(t *testing.T) {
	tenth, fifth := 0.1, 0.2 // variables, so the sum rounds at run time
	cases := []struct {
		v    Value
		want string
	}{
		{int64(42), "42"}, {int64(-7), "-7"}, {int8(-128), "-128"}, {uint8(255), "255"},
		{3.5, "3.5"}, {2.0, "2.0"}, {-0.5, "-0.5"}, {tenth + fifth, "0.30000000000000004"},
		{1.0 / 3.0, "0.3333333333333333"}, {float32(0.1), "0.1"}, {100.0, "100.0"},
		{math.Inf(1), "Infinity"}, {math.Inf(-1), "-Infinity"}, {math.NaN(), "NaN"},
		{"plain", "plain"}, {"", ""}, {true, "true"}, {false, "false"}, {Null, "null"}, {Void, "void"},
		{NewArray([]Value{int64(1), int64(2)}), "[1, 2]"},
		{NewArray(nil), "[]"},
		{NewArray([]Value{"a", "tab\there", "q\"", "b\\s", "\x01"}), `["a", "tab\there", "q\"", "b\\s", "\u{1}"]`},
		{NewArray([]Value{true, Null, Void}), "[true, null, void]"},
		{rec(point, 1.0, 2.0), "Point{ x: 1.0, y: 2.0 }"},
		{rec(NewTag("Empty", RecordTag, nil)), "Empty{}"},
		{&Tagged{Tag: userID, Payload: "u-1"}, `UserId("u-1")`},
		{&testFunc{"f"}, "func"}, {&Constructor{Tag: userID}, "UserId"}, {point, "Point"},
		{NewMutex(), "Mutex"}, {NewWaitGroup(), "WaitGroup"}, {NewAtomic(int64(1)), "Atomic"},
	}
	for _, c := range cases {
		if got := Render(c.v); got != c.want {
			t.Errorf("Render(%v) = %q, want %q", c.v, got, c.want)
		}
	}
	m := NewMap()
	if got := Render(m); got != "Map{}" {
		t.Errorf("empty map: %q", got)
	}
	m.Set("a", int64(1))
	m.Set(int64(7), "seven")
	m.Set(rec(point, 0.0, 0.0), "origin")
	if got, want := Render(m), `Map{ "a": 1, 7: "seven", Point{ x: 0.0, y: 0.0 }: "origin" }`; got != want {
		t.Errorf("map: %q, want %q", got, want)
	}
	// cycles
	node := NewTag("Node", RecordTag, []string{"value", "next"})
	a := rec(node, int64(1), Null)
	a.Set(1, a)
	if got := Render(a); got != "Node{ value: 1, next: ... }" {
		t.Errorf("cyclic record: %q", got)
	}
	cell := NewTag("Cell", RecordTag, []string{"items"})
	c := rec(cell, NewArray(nil))
	items, _ := c.Field("items")
	items.(*Array).Push(c)
	if got := Render(c); got != "Cell{ items: [...] }" {
		t.Errorf("cyclic array: %q", got)
	}
	if got := Render(NewArray([]Value{c, c})); got != "[Cell{ items: [...] }, Cell{ items: [...] }]" {
		t.Errorf("twice: %q", got)
	}
}

func TestEqualAndHash(t *testing.T) {
	same := [][2]Value{
		{int64(1), int64(1)}, {"a", "a"}, {1.5, 1.5}, {0.0, math.Copysign(0, -1)}, {Null, Null}, {Void, Void},
		{NewArray([]Value{int64(1), "x"}), NewArray([]Value{int64(1), "x"})},
		{rec(point, 1.0, 2.0), rec(point, 1.0, 2.0)},
		{&Tagged{userID, "u"}, &Tagged{userID, "u"}},
		{uint8(3), uint8(3)}, {float32(1.5), float32(1.5)},
	}
	for _, p := range same {
		if !Equal(p[0], p[1]) {
			t.Errorf("Equal(%v, %v) = false", p[0], p[1])
		}
		if Hash(p[0]) != Hash(p[1]) {
			t.Errorf("Hash(%v) != Hash(%v)", p[0], p[1])
		}
	}
	other := NewTag("Point", RecordTag, []string{"x", "y"})
	different := [][2]Value{
		{int64(1), int8(1)}, {int64(1), 1.0}, {"1", int64(1)}, {Null, Void}, {math.NaN(), math.NaN()},
		{rec(point, 1.0, 2.0), rec(other, 1.0, 2.0)}, {rec(point, 1.0, 2.0), rec(point, 1.0, 3.0)},
		{NewArray([]Value{int64(1)}), NewArray([]Value{int64(1), int64(2)})},
		{&Tagged{userID, "u"}, "u"}, {&testFunc{"f"}, &testFunc{"f"}}, {NewMutex(), NewMutex()},
	}
	for _, p := range different {
		if Equal(p[0], p[1]) {
			t.Errorf("Equal(%v, %v) = true", p[0], p[1])
		}
	}
	m1, m2 := NewMap(), NewMap()
	m1.Set("a", int64(1))
	m1.Set("b", int64(2))
	m2.Set("b", int64(2))
	m2.Set("a", int64(1))
	if !Equal(m1, m2) || Hash(m1) != Hash(m2) {
		t.Error("maps with the same entries in another order should be equal")
	}
	// cycles compare equal in finite time
	node := NewTag("Node", RecordTag, []string{"value", "next"})
	a := rec(node, int64(1), Null)
	a.Set(1, a)
	b := rec(node, int64(1), Null)
	b.Set(1, b)
	if !Equal(a, b) {
		t.Error("cyclic records should be equal")
	}
	if Hash(a) != Hash(b) {
		t.Error("cyclic records should hash alike")
	}
	m := NewMap()
	m.Set(a, "cycle")
	if v, ok := m.Get(b); !ok || v != "cycle" {
		t.Error("a cyclic key should be found through an equal key")
	}
}

func TestMapOrder(t *testing.T) {
	m := NewMap()
	m.Set("c", int64(3))
	m.Set("a", int64(1))
	m.Set("b", int64(2))
	m.Set("a", int64(10))
	m.Set("d", int64(4))
	if got := Render(m); got != `Map{ "c": 3, "a": 10, "b": 2, "d": 4 }` {
		t.Errorf("after replace: %q", got)
	}
	m.Remove("c")
	m.Set("c", int64(30))
	if got := Render(m); got != `Map{ "a": 10, "b": 2, "d": 4, "c": 30 }` {
		t.Errorf("after remove and reinsert: %q", got)
	}
	// iteration sees insertions and skips removals
	var seen []string
	var cursor uint64
	for {
		k, _, seq, ok := m.Next(cursor)
		if !ok {
			break
		}
		cursor = seq
		seen = append(seen, k.(string))
		if k == "a" {
			m.Set("e", int64(5))
			m.Remove("b")
		}
	}
	if got := len(seen); got != 4 || seen[0] != "a" || seen[1] != "d" || seen[2] != "c" || seen[3] != "e" {
		t.Errorf("iteration: %v", seen)
	}
	// compaction keeps order and lookups
	big := NewMap()
	for i := int64(0); i < 100; i++ {
		big.Set(i, i)
	}
	for i := int64(0); i < 90; i++ {
		big.Remove(i)
	}
	if big.Len() != 10 {
		t.Errorf("len after removals: %d", big.Len())
	}
	if v, ok := big.Get(int64(95)); !ok || v != int64(95) {
		t.Error("lookup after compaction")
	}
	if keys := big.Keys(); keys[0] != int64(90) || keys[9] != int64(99) {
		t.Errorf("keys after compaction: %v", keys)
	}
}

func TestArray(t *testing.T) {
	a := NewArray(nil)
	for i := int64(0); i < 10; i++ {
		a.Push(i)
	}
	if a.Len() != 10 {
		t.Fatal("len")
	}
	if v, ok := a.Get(9); !ok || v != int64(9) {
		t.Error("get")
	}
	if _, ok := a.Get(10); ok {
		t.Error("get out of range")
	}
	if !a.Set(0, "x") || a.Set(10, "y") {
		t.Error("set")
	}
	v, ok := a.Pop()
	if !ok || v != int64(9) || a.Len() != 9 {
		t.Error("pop")
	}
	// racing pushes and reads never crash and lose no structure
	b := NewArray(nil)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				b.Push(int64(i))
				b.Get(i)
				b.Set(i/2, int64(i))
				_ = b.Values()
			}
		}()
	}
	wg.Wait()
	if b.Len() != 8000 {
		t.Errorf("len after racing pushes: %d", b.Len())
	}
}

func TestArithmetic(t *testing.T) {
	ok := []struct {
		op   string
		a, b Value
		want Value
	}{
		{"+", int64(1), int64(2), int64(3)}, {"-", int64(1), int64(2), int64(-1)}, {"*", int64(3), int64(4), int64(12)},
		{"/", int64(7), int64(2), int64(3)}, {"/", int64(-7), int64(2), int64(-3)}, {"%", int64(-7), int64(2), int64(-1)},
		{"%", 7.5, 2.0, 1.5}, {"%", -7.5, 2.0, -1.5}, {"/", 1.0, 0.0, math.Inf(1)},
		{"+", "a", "b", "ab"}, {"+", uint8(200), uint8(55), uint8(255)}, {"-", int8(-128), int8(-1), int8(-127)},
		{"*", float32(1.5), float32(2), float32(3)},
		{"+", &Tagged{userID, "a"}, &Tagged{userID, "b"}, &Tagged{userID, "ab"}},
	}
	for _, c := range ok {
		got, err := Binary(c.op, c.a, c.b)
		if err != nil {
			t.Errorf("%v %s %v: %v", c.a, c.op, c.b, err)
			continue
		}
		if !Equal(got, c.want) {
			t.Errorf("%v %s %v = %v, want %v", c.a, c.op, c.b, got, c.want)
		}
	}
	bad := []struct {
		op   string
		a, b Value
		want string
	}{
		{"+", int64(math.MaxInt64), int64(1), "overflow"}, {"-", int64(math.MinInt64), int64(1), "overflow"},
		{"*", int64(math.MinInt64), int64(-1), "overflow"}, {"/", int64(math.MinInt64), int64(-1), "overflow"},
		{"/", int64(1), int64(0), "division by zero"}, {"%", int64(1), int64(0), "division by zero"},
		{"+", uint8(200), uint8(100), "overflow"}, {"-", uint8(1), uint8(2), "overflow"},
		{"*", uint64(1 << 63), uint64(2), "overflow"}, {"+", int8(127), int8(1), "overflow"},
		{"+", int64(1), 1.0, "one kind"}, {"-", "a", "b", "one kind"}, {"+", int64(1), int8(1), "one kind"},
		{"+", &Tagged{userID, "a"}, "b", "one kind"},
	}
	for _, c := range bad {
		if _, err := Binary(c.op, c.a, c.b); err == nil || !contains(err.Error(), c.want) {
			t.Errorf("%v %s %v: err %v, want %q", c.a, c.op, c.b, err, c.want)
		}
	}
	if v, _ := Neg(int64(5)); v != int64(-5) {
		t.Error("neg")
	}
	if _, err := Neg(int64(math.MinInt64)); err == nil {
		t.Error("neg overflow")
	}
	cmps := []struct {
		op   string
		a, b Value
		want bool
	}{
		{"<", int64(1), int64(2), true}, {">=", "b", "a", true}, {"<", "a", "ab", true},
		{"<", math.NaN(), 1.0, false}, {">=", math.NaN(), 1.0, false}, {"<=", 1.0, 1.0, true},
		{"<", &Tagged{userID, "a"}, &Tagged{userID, "b"}, true},
	}
	for _, c := range cmps {
		got, err := Compare(c.op, c.a, c.b)
		if err != nil || got != c.want {
			t.Errorf("%v %s %v = %v, %v; want %v", c.a, c.op, c.b, got, err, c.want)
		}
	}
	if _, err := Compare("<", int64(1), 2.0); err == nil {
		t.Error("compare across kinds")
	}
}

func TestConvert(t *testing.T) {
	ok := []struct {
		to   Kind
		v    Value
		want Value
	}{
		{KindInt8, int64(-128), int8(-128)}, {KindUint8, int64(255), uint8(255)}, {KindInt64, uint64(1 << 62), int64(1 << 62)},
		{KindFloat64, int64(7), 7.0}, {KindFloat32, int64(16777217), float32(16777216)},
		{KindInt64, 3.0, int64(3)}, {KindInt64, -0.0, int64(0)}, {KindUint8, 255.0, uint8(255)},
		{KindFloat32, 0.1, float32(0.1)}, {KindFloat64, float32(0.5), 0.5},
		{KindString, "s", "s"}, {KindBool, true, true},
		{KindString, &Tagged{userID, "u"}, "u"}, {KindInt64, &Tagged{NewPrimitiveTag("Count", KindInt64), int64(4)}, int64(4)},
		{KindInt64, 9223372036854775807.0, nil}, // 2^63 rounds up: out of range, checked below
	}
	for _, c := range ok {
		got, err := Convert(c.to, c.v)
		if c.want == nil {
			if err == nil {
				t.Errorf("Convert(%s, %v) should fail", c.to, c.v)
			}
			continue
		}
		if err != nil || !Equal(got, c.want) {
			t.Errorf("Convert(%s, %v) = %v, %v; want %v", c.to, c.v, got, err, c.want)
		}
	}
	bad := []struct {
		to Kind
		v  Value
	}{
		{KindInt8, int64(300)}, {KindUint8, int64(-1)}, {KindInt64, 1.5}, {KindInt64, math.NaN()},
		{KindInt64, math.Inf(1)}, {KindUint64, -1.0}, {KindInt32, "1"}, {KindString, int64(1)}, {KindBool, "true"},
		{KindInt64, uint64(1 << 63)},
	}
	for _, c := range bad {
		if _, err := Convert(c.to, c.v); err == nil {
			t.Errorf("Convert(%s, %v) should fail", c.to, c.v)
		}
	}
}

func TestSlotAndSync(t *testing.T) {
	var s Slot
	if s.Load() != Uninitialized {
		t.Error("fresh slot")
	}
	s.Store(int64(1))
	if !s.CompareAndSwap(int64(1), int64(2)) || s.Load() != int64(2) {
		t.Error("cas")
	}
	if s.CompareAndSwap(int64(1), int64(3)) {
		t.Error("cas should fail")
	}
	var wg sync.WaitGroup
	var a Atomic
	a.Slot.Store(int64(0))
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				a.Slot.Modify(func(old Value) (Value, error) { return old.(int64) + 1, nil })
			}
		}()
	}
	wg.Wait()
	if a.Slot.Load() != int64(8000) {
		t.Errorf("atomic adds: %v", a.Slot.Load())
	}
	m := NewMutex()
	if err := m.Unlock(); err == nil {
		t.Error("unlock of unlocked")
	}
	m.Lock()
	if !m.Locked() {
		t.Error("locked")
	}
	if err := m.Unlock(); err != nil {
		t.Error(err)
	}
	w := NewWaitGroup()
	if err := w.Done(); err == nil {
		t.Error("done at zero")
	}
	w.Add(2)
	go func() { w.Done(); w.Done() }()
	w.Wait()
	o := NewOnce()
	n := 0
	for i := 0; i < 3; i++ {
		o.Do(func() error { n++; return nil })
	}
	if n != 1 {
		t.Errorf("once ran %d times", n)
	}
	c := NewCond(m)
	if err := c.Wait(); err == nil {
		t.Error("wait without lock")
	}
	m.Lock()
	go func() {
		m.Lock()
		m.Unlock()
		c.Signal()
	}()
	if err := c.Wait(); err != nil {
		t.Error(err)
	}
	if !m.Locked() {
		t.Error("mutex should be reacquired after wait")
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
