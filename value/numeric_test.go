package value

import (
	"math"
	"testing"
)

// TestBinaryKinds runs every arithmetic operator over every numeric kind: results keep the
// kind, overflow and zero divisors are errors for integers, floats follow IEEE 754, and
// mismatched kinds are errors.
func TestBinaryKinds(t *testing.T) {
	ints := []Value{int8(100), int16(30000), int32(2_000_000_000), int64(9_000_000_000_000_000_000),
		uint8(200), uint16(60000), uint32(4_000_000_000), uint64(18_000_000_000_000_000_000)}
	for _, a := range ints {
		kind := KindOf(a)
		sum, err := Binary("+", a, Convert1(kind, 1))
		if err != nil || KindOf(sum) != kind {
			t.Errorf("%s + 1: %v %v", kind, sum, err)
		}
		if _, err := Binary("*", a, Convert1(kind, 2)); err == nil {
			t.Errorf("%s * 2 should overflow", kind)
		}
		if _, err := Binary("/", a, Convert1(kind, 0)); err == nil {
			t.Errorf("%s / 0 should fail", kind)
		}
		if _, err := Binary("%", a, Convert1(kind, 0)); err == nil {
			t.Errorf("%s %% 0 should fail", kind)
		}
		q, _ := Binary("/", a, Convert1(kind, 3))
		r, _ := Binary("%", a, Convert1(kind, 3))
		back, _ := Binary("*", q, Convert1(kind, 3))
		back, _ = Binary("+", back, r)
		if !Equal(back, a) {
			t.Errorf("%s: q*3+r != a: %v", kind, back)
		}
		d, _ := Binary("-", a, a)
		if !Equal(d, Convert1(kind, 0)) {
			t.Errorf("%s - itself: %v", kind, d)
		}
	}
	for _, u := range []Value{uint8(0), uint16(0), uint32(0), uint64(0)} {
		if _, err := Binary("-", u, Convert1(KindOf(u), 1)); err == nil {
			t.Errorf("%s below zero should fail", KindOf(u))
		}
	}
	if _, err := Binary("-", int8(-128), int8(1)); err == nil {
		t.Error("int8 underflow")
	}
	if _, err := Binary("+", int64(1), float64(1)); err == nil {
		t.Error("mixed kinds")
	}
	if _, err := Binary("+", "a", 1); err == nil {
		t.Error("string and int")
	}
	if v, _ := Binary("+", "a", "b"); v != "ab" {
		t.Error("string concatenation")
	}
	if _, err := Binary("-", "a", "b"); err == nil {
		t.Error("string subtraction")
	}
	inf, _ := Binary("/", 1.0, 0.0)
	if inf != math.Inf(1) {
		t.Errorf("1.0 / 0.0 = %v", inf)
	}
	f32, _ := Binary("/", float32(1), float32(0))
	if f32 != float32(math.Inf(1)) {
		t.Errorf("float32 1 / 0 = %v", f32)
	}
	m, _ := Binary("%", -7.5, 2.0)
	if m != -1.5 {
		t.Errorf("-7.5 %% 2.0 = %v", m)
	}
	q, _ := Binary("/", int64(-7), int64(2))
	r, _ := Binary("%", int64(-7), int64(2))
	if q != int64(-3) || r != int64(-1) {
		t.Errorf("-7 / 2, -7 %% 2 = %v, %v", q, r)
	}
	if _, err := Binary("/", int64(math.MinInt64), int64(-1)); err == nil {
		t.Error("MinInt64 / -1 should overflow")
	}
	tagged := &Tagged{Tag: NewPrimitiveTag("Meters", KindFloat64), Payload: 2.0}
	sum, err := Binary("+", tagged, tagged)
	if err != nil || sum.(*Tagged).Payload != 4.0 || sum.(*Tagged).Tag != tagged.Tag {
		t.Errorf("tagged + tagged: %v %v", sum, err)
	}
	if _, err := Binary("+", tagged, 2.0); err == nil {
		t.Error("tagged + plain")
	}
	other := &Tagged{Tag: NewPrimitiveTag("Feet", KindFloat64), Payload: 2.0}
	if _, err := Binary("+", tagged, other); err == nil {
		t.Error("two tags")
	}
	if _, err := Binary("?", 1, 1); err == nil {
		t.Error("unknown operator")
	}
}

// Convert1 converts a small int to a kind for the tables.
func Convert1(kind Kind, n int) Value {
	v, err := Convert(kind, int64(n))
	if err != nil {
		panic(err)
	}
	return v
}

func TestNegAndCompare(t *testing.T) {
	for _, v := range []Value{int8(1), int16(1), int32(1), int64(1), float32(1), float64(1)} {
		n, err := Neg(v)
		if err != nil || KindOf(n) != KindOf(v) {
			t.Errorf("-%v: %v %v", v, n, err)
		}
		z, _ := Binary("+", n, v)
		if !Equal(z, Convert1(KindOf(v), 0)) {
			t.Errorf("-%v + %v = %v", v, v, z)
		}
	}
	if _, err := Neg(uint8(1)); err == nil {
		t.Error("negating an unsigned value")
	}
	if _, err := Neg(int8(-128)); err == nil {
		t.Error("negating int8 min")
	}
	if _, err := Neg("s"); err == nil {
		t.Error("negating a string")
	}
	for _, pair := range [][2]Value{{int8(1), int8(2)}, {uint16(1), uint16(2)}, {uint64(1), uint64(2)}, {float32(1), float32(2)}, {1.0, 2.0}, {"a", "b"}, {int64(1), int64(2)}} {
		a, b := pair[0], pair[1]
		for op, want := range map[string]bool{"<": true, "<=": true, ">": false, ">=": false} {
			got, err := Compare(op, a, b)
			if err != nil || got != want {
				t.Errorf("%v %s %v = %v %v", a, op, b, got, err)
			}
		}
		eq, _ := Compare("<=", a, a)
		if !eq {
			t.Errorf("%v <= itself", a)
		}
	}
	if _, err := Compare("<", int64(1), 1.0); err == nil {
		t.Error("comparing kinds")
	}
	if _, err := Compare("<", true, false); err == nil {
		t.Error("comparing bools")
	}
	nan := math.NaN()
	if lt, _ := Compare("<", nan, 1.0); lt {
		t.Error("NaN < 1")
	}
	tagged := &Tagged{Tag: NewPrimitiveTag("Count", KindInt64), Payload: int64(1)}
	bigger := &Tagged{Tag: tagged.Tag, Payload: int64(2)}
	if lt, err := Compare("<", tagged, bigger); err != nil || !lt {
		t.Errorf("tagged compare: %v %v", lt, err)
	}
	if _, err := Compare("<", tagged, int64(2)); err == nil {
		t.Error("tagged against plain")
	}
	if _, err := Compare("?", 1, 1); err == nil {
		t.Error("unknown comparison")
	}
}

func TestConvertMatrix(t *testing.T) {
	kinds := []Kind{KindInt8, KindInt16, KindInt32, KindInt64, KindUint8, KindUint16, KindUint32, KindUint64, KindFloat32, KindFloat64}
	for _, from := range kinds {
		for _, to := range kinds {
			v, err := Convert(to, Convert1(from, 7))
			if err != nil || KindOf(v) != to {
				t.Errorf("%s -> %s of 7: %v %v", from, to, v, err)
			}
			if f, _ := Convert(KindFloat64, v); f != 7.0 {
				t.Errorf("%s -> %s of 7 reads back %v", from, to, f)
			}
		}
	}
	bad := []struct {
		to Kind
		v  Value
	}{
		{KindInt8, int64(128)}, {KindInt8, int64(-129)}, {KindUint8, int64(-1)}, {KindUint8, int64(256)},
		{KindInt16, int64(40000)}, {KindUint16, int64(70000)}, {KindInt32, int64(1 << 40)}, {KindUint32, int64(-5)},
		{KindUint64, int64(-1)}, {KindInt64, uint64(1 << 63)}, {KindInt64, 1.5}, {KindInt64, math.NaN()},
		{KindInt64, math.Inf(1)}, {KindInt64, 1e30}, {KindInt8, float32(300)}, {KindUint8, -0.5},
		{KindInt64, "7"}, {KindFloat64, true},
	}
	for _, c := range bad {
		if _, err := Convert(c.to, c.v); err == nil {
			t.Errorf("%v -> %s should fail", c.v, c.to)
		}
	}
	if v, _ := Convert(KindFloat32, 16777217.0); v != float32(16777216) {
		t.Errorf("float32 rounding: %v", v)
	}
	if v, _ := Convert(KindFloat32, 1e40); v != float32(math.Inf(1)) {
		t.Errorf("float32 overflow: %v", v)
	}
	if v, _ := Convert(KindInt64, -0.0); v != int64(0) {
		t.Errorf("int(-0.0) = %v", v)
	}
	if v, _ := Convert(KindString, "s"); v != "s" {
		t.Error("string identity")
	}
	if v, _ := Convert(KindBool, true); v != true {
		t.Error("bool identity")
	}
	if _, err := Convert(KindString, 1); err == nil {
		t.Error("string of int")
	}
	tagged := &Tagged{Tag: NewPrimitiveTag("Count", KindInt64), Payload: int64(3)}
	if v, err := Convert(KindFloat64, tagged); err != nil || v != 3.0 {
		t.Errorf("unwrapping a tagged number: %v %v", v, err)
	}
	if v, _ := Convert(KindString, &Tagged{Tag: NewPrimitiveTag("Id", KindString), Payload: "x"}); v != "x" {
		t.Error("unwrapping a tagged string")
	}
}

func TestDescribe(t *testing.T) {
	for v, want := range map[Value]string{int64(1): "int", "s": "string", true: "bool", Null: "null", Void: "void", float32(1): "float32", uint8(1): "byte"} {
		if got := Describe(v); got != want {
			t.Errorf("Describe(%v) = %q, want %q", v, got, want)
		}
	}
	if got := Describe(NewArray(nil)); got != "array" {
		t.Errorf("array: %q", got)
	}
	tagged := &Tagged{Tag: NewPrimitiveTag("Count", KindInt64), Payload: int64(3)}
	if got := Describe(tagged); got != "Count" {
		t.Errorf("tagged: %q", got)
	}
	for _, k := range []Kind{KindInt8, KindUint64} {
		if !k.IsInteger() || !k.IsNumber() {
			t.Errorf("%s should be an integer kind", k)
		}
	}
	if KindFloat64.IsInteger() || KindString.IsNumber() {
		t.Error("float or string as integer")
	}
	if KindInt8.Bits() != 8 || KindUint64.Bits() != 64 || KindFloat32.Bits() != 0 {
		t.Error("Bits")
	}
}
