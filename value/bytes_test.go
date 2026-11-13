package value

import "testing"

func bs(s string) *Bytes { return CopyBytes([]byte(s)) }

func TestBytes(t *testing.T) {
	b := bs("\x89PNG")
	for _, tc := range []struct {
		v    Value
		want string
	}{
		{b, "Bytes(89 50 4e 47)"},
		{bs(""), "Bytes()"},
		{bs("\x00\xff"), "Bytes(00 ff)"},
		{NewArray([]Value{b, "P"}), `[Bytes(89 50 4e 47), "P"]`},
	} {
		if got := Render(tc.v); got != tc.want {
			t.Errorf("Render(%#v) = %s, want %s", tc.v, got, tc.want)
		}
	}
	if !Equal(b, bs("\x89PNG")) || Equal(b, bs("\x89PN")) {
		t.Error("Bytes compare by content")
	}
	if Equal(bs("PNG"), "PNG") || Equal("PNG", bs("PNG")) {
		t.Error("Bytes never equal a string with the same bytes")
	}
	if Hash(b) != Hash(bs("\x89PNG")) {
		t.Error("equal Bytes hash alike")
	}
	if Hash(bs("a")) == Hash(bs("b")) {
		t.Error("Bytes hash by their contents, so map keys spread")
	}
	if Hash(bs("PNG")) == Hash("PNG") {
		t.Error("Bytes and a string with the same bytes hash apart, by their kinds")
	}
	if k, ok := KindByName("Bytes"); !ok || k != KindBytes || KindOf(b) != KindBytes || k.IsPrimitive() {
		t.Errorf("KindByName(Bytes) = %v, %v; KindOf = %v", k, ok, KindOf(b))
	}
	if Describe(b) != "Bytes" {
		t.Errorf("Describe = %s", Describe(b))
	}
	if ff := bs("\xff"); TextIn(ff) != ff {
		t.Error("TextIn decodes strings, not Bytes")
	}
	m := NewMap()
	m.Set(bs("k"), int64(1))
	m.Set("k", int64(2))
	if v, _ := m.Get(bs("k")); v != int64(1) || m.Len() != 2 {
		t.Errorf("a map keeps Bytes and string keys apart: %s", Render(m))
	}
}

func TestBytesGrow(t *testing.T) {
	b := NewBytes(nil)
	for i := 0; i < 1000; i++ {
		b.Push(byte(i))
	}
	if b.Len() != 1000 {
		t.Fatalf("Len = %d", b.Len())
	}
	for i := 0; i < 1000; i++ {
		if c, _ := b.Get(i); c != byte(i) {
			t.Fatalf("Get(%d) = %d", i, c)
		}
	}
	b.Append(b.View())
	if b.Len() != 2000 {
		t.Fatalf("self-append: Len = %d", b.Len())
	}
	if c, _ := b.Get(1999); c != byte(999%256) {
		t.Fatalf("self-append: last = %d", c)
	}
	if !b.Truncate(1) || b.Truncate(5) {
		t.Fatal("Truncate keeps 0 ..= length")
	}
	b.Push(7)
	if Render(b) != "Bytes(00 07)" {
		t.Fatalf("after truncate and push: %s", Render(b))
	}
	if c, ok := b.Pop(); !ok || c != 7 || b.Len() != 1 {
		t.Fatalf("Pop = %d, %v", c, ok)
	}
	if _, ok := b.Get(1); ok {
		t.Fatal("Get past the length")
	}
}
