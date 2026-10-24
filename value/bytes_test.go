package value

import "testing"

func TestBytes(t *testing.T) {
	b := Bytes("\x89PNG")
	for _, tc := range []struct {
		v    Value
		want string
	}{
		{b, "Bytes(89 50 4e 47)"},
		{Bytes(""), "Bytes()"},
		{Bytes("\x00\xff"), "Bytes(00 ff)"},
		{NewArray([]Value{b, "P"}), `[Bytes(89 50 4e 47), "P"]`},
	} {
		if got := Render(tc.v); got != tc.want {
			t.Errorf("Render(%#v) = %s, want %s", tc.v, got, tc.want)
		}
	}
	if !Equal(b, Bytes("\x89PNG")) || Equal(b, Bytes("\x89PN")) {
		t.Error("Bytes compare by content")
	}
	if Equal(Bytes("PNG"), "PNG") || Equal("PNG", Bytes("PNG")) {
		t.Error("Bytes never equal a string with the same bytes")
	}
	if Hash(b) != Hash(Bytes("\x89PNG")) {
		t.Error("equal Bytes hash alike")
	}
	if Hash(Bytes("a")) == Hash(Bytes("b")) {
		t.Error("Bytes hash by their contents, so map keys spread")
	}
	if Hash(Bytes("PNG")) == Hash("PNG") {
		t.Error("Bytes and a string with the same bytes hash apart, by their kinds")
	}
	if k, ok := KindByName("Bytes"); !ok || k != KindBytes || KindOf(b) != KindBytes || k.IsPrimitive() {
		t.Errorf("KindByName(Bytes) = %v, %v; KindOf = %v", k, ok, KindOf(b))
	}
	if Describe(b) != "Bytes" {
		t.Errorf("Describe = %s", Describe(b))
	}
	if TextIn(Bytes("\xff")) != Bytes("\xff") {
		t.Error("TextIn decodes strings, not Bytes")
	}
	m := NewMap()
	m.Set(Bytes("k"), int64(1))
	m.Set("k", int64(2))
	if v, _ := m.Get(Bytes("k")); v != int64(1) || m.Len() != 2 {
		t.Errorf("a map keeps Bytes and string keys apart: %s", Render(m))
	}
}
