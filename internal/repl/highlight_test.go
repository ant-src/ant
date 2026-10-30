package repl

import (
	"strings"
	"testing"
)

func TestClasses(t *testing.T) {
	// each test marks the class of every byte: k keyword, c constant, s string, n number,
	// r regex, / comment, N namespace, f function, m markup, b bad, . plain
	tests := []struct{ src, want string }{
		{`var x = 1`, "kkk . . n"},
		{`func f(a: int): string { "" }`, "kkkk f... NNN.. NNNNNN . ss ."},
		{`int(x)`, "fff..."},
		{`if x == null { "a" }`, "kk . .. cccc . sss ."},
		{`println("hi") // say`, "fffffff.ssss. //////"},
		{"xs | Array.map({ x * 2 })", ".. . NNNNN.fff.. . . n .."},
		{`"a ${n + 1} b"`, "sssss....nssss"}, // the interpolation's tokens keep their classes
		{"`a ${n} \\`", "ssssssssss"},        // a raw string has no interpolations
		{`/ab+/g`, "rrrrrr"},
		{`<p class={c}>{n}</p>`, "mmmmmmmmmm.mmm.mmmmm"},
		{`class`, "bbbbb"},
		{`"unfinished`, "sssssssssss"},
		{`Http.Reply{ status: 200 }`, "NNNN.NNNNN. ....... nnn ."},
		{"/* a\nb */ x", "///////// ."},
	}
	for _, tc := range tests {
		cls := classes(tc.src)
		var sb strings.Builder
		for i := range tc.src {
			sb.WriteByte(".kcsnr/NfmbG"[cls[i]])
		}
		// a space in want stands for a plain byte too
		got := sb.String()
		want := strings.ReplaceAll(tc.want, " ", ".")
		if got != want {
			t.Errorf("%q:\n got %s\nwant %s", tc.src, got, want)
		}
	}
}

func TestPaint(t *testing.T) {
	src := "var x = 1"
	var sb strings.Builder
	ansiTheme.paint(&sb, src, classes(src), 0)
	if got := sb.String(); got != "\x1b[35mvar\x1b[0m x = \x1b[33m1\x1b[0m" {
		t.Errorf("paint: %q", got)
	}
	sb.Reset()
	plainTheme.paint(&sb, src, classes(src), 0)
	if sb.String() != src {
		t.Errorf("plain paint: %q", sb.String())
	}
	// a window of the source keeps the classes of its bytes
	sb.Reset()
	ansiTheme.paint(&sb, "x = 1", classes(src), 4)
	if got := sb.String(); got != "x = \x1b[33m1\x1b[0m" {
		t.Errorf("paint from 4: %q", got)
	}
}
