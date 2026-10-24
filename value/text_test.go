package value

import (
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"os/exec"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestText(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"", ""},
		{"ant", "ant"},
		{"größe 🐜", "größe 🐜"},
		// Unicode's table 3-8: U+FFFD for each maximal subpart.
		{"\x61\xF1\x80\x80\xE1\x80\xC2\x62\x80\x63\x80\xBF\x64", "a\uFFFD\uFFFD\uFFFDb\uFFFDc\uFFFD\uFFFDd"},
		{"\xff\xfe", "\uFFFD\uFFFD"},
		{"\xe2\x82", "\uFFFD"},                           // a truncated €
		{"\xe2\x82a", "\uFFFDa"},                         // likewise, before a letter
		{"\xf0\x9f\x98", "\uFFFD"},                       // a truncated emoji
		{"\xed\xa0\x80", "\uFFFD\uFFFD\uFFFD"},           // an encoded surrogate: ED takes only 80..9F next
		{"\xc0\xaf", "\uFFFD\uFFFD"},                     // an overlong /
		{"\xe0\x80\x80", "\uFFFD\uFFFD\uFFFD"},           // an overlong NUL: E0 takes only A0..BF next
		{"\xf4\x90\x80\x80", "\uFFFD\uFFFD\uFFFD\uFFFD"}, // above 10FFFF
		{"\xf5", "\uFFFD"},
		{"a\x80", "a\uFFFD"},
	} {
		if got := Text(tc.in); got != tc.want {
			t.Errorf("Text(%+q) = %+q, want %+q", tc.in, got, tc.want)
		}
	}
}

// TestTextAgainstTextDecoder compares Text with the TextDecoder of Node, which follows the
// WHATWG Encoding standard, over random bytes biased toward the edges of the UTF-8 ranges. It
// skips when node is not on the path.
func TestTextAgainstTextDecoder(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not on the path")
	}
	edges := []byte{0x00, 0x41, 0x7F, 0x80, 0x8F, 0x90, 0x9F, 0xA0, 0xBF, 0xC0, 0xC1, 0xC2, 0xDF,
		0xE0, 0xE1, 0xEC, 0xED, 0xEE, 0xEF, 0xF0, 0xF1, 0xF3, 0xF4, 0xF5, 0xFF}
	rng := rand.New(rand.NewSource(1))
	inputs := make([]string, 20000)
	for i := range inputs {
		b := make([]byte, 1+rng.Intn(8))
		for j := range b {
			if rng.Intn(4) == 0 {
				b[j] = byte(rng.Intn(256))
			} else {
				b[j] = edges[rng.Intn(len(edges))]
			}
		}
		inputs[i] = hex.EncodeToString(b)
	}
	data, _ := json.Marshal(inputs)
	cmd := exec.Command(node, "-e", `
		const inputs = JSON.parse(require("fs").readFileSync(0, "utf8"));
		const dec = new TextDecoder("utf-8");
		process.stdout.write(JSON.stringify(inputs.map(h => dec.decode(Buffer.from(h, "hex")))));
	`)
	cmd.Stdin = strings.NewReader(string(data))
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("node: %v", err)
	}
	var want []string
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatal(err)
	}
	failures := 0
	for i, h := range inputs {
		raw, _ := hex.DecodeString(h)
		if got := Text(string(raw)); got != want[i] {
			if failures++; failures <= 20 {
				t.Errorf("Text(%+q) = %+q, TextDecoder %+q", raw, got, want[i])
			}
		}
	}
	if failures > 0 {
		t.Errorf("%d of %d inputs differ from TextDecoder", failures, len(inputs))
	}
}

// FuzzText checks that the result is UTF-8, that UTF-8 comes back unchanged, and that every
// byte of the input is either kept or under a U+FFFD.
func FuzzText(f *testing.F) {
	f.Add("\x61\xF1\x80\x80\xE1\x80\xC2\x62\x80\x63\x80\xBF\x64")
	f.Add("größe")
	f.Fuzz(func(t *testing.T, s string) {
		got := Text(s)
		if !utf8.ValidString(got) {
			t.Fatalf("Text(%+q) = %+q, not UTF-8", s, got)
		}
		if utf8.ValidString(s) && got != s {
			t.Fatalf("Text(%+q) changed UTF-8 text", s)
		}
		if Text(got) != got {
			t.Fatalf("Text is not idempotent on %+q", s)
		}
		if strings.ReplaceAll(got, "\uFFFD", "") != strings.ReplaceAll(Text(strings.ReplaceAll(s, "\uFFFD", "")), "\uFFFD", "") {
			t.Fatalf("Text(%+q) = %+q lost or invented valid text", s, got)
		}
	})
}

func TestTextIn(t *testing.T) {
	if got := TextIn("a\xffb"); got != "a�b" {
		t.Errorf("TextIn of a string = %+q", got)
	}
	if got := TextIn(int64(7)); got != int64(7) {
		t.Errorf("TextIn of an int = %v", got)
	}

	tag := NewTag("Row", RecordTag, []string{"name", "tags"})
	tags := NewArray([]Value{"ok", "x\xff"})
	row := NewRecord(tag, []Value{"caf\xe9", tags})
	if TextIn(row) != row {
		t.Fatal("TextIn replaced a record instead of changing it in place")
	}
	if got := Render(row); got != `Row{ name: "caf`+"�"+`", tags: ["ok", "x`+"�"+`"] }` {
		t.Errorf("record: %s", got)
	}

	m := NewMap()
	m.Set("a", "1")
	m.Set("k\xff", "2")
	m.Set("k\xfe", "3")
	m.Set("z", "v\xff")
	TextIn(m)
	if got := Render(m); got != `Map{ "a": "1", "k`+"�"+`": "3", "z": "v`+"�"+`" }` {
		t.Errorf("map: %s", got)
	}

	s := NewSet()
	s.Add("p\xff")
	s.Add("q")
	TextIn(s)
	if !s.Has("p�") || !s.Has("q") || s.Len() != 2 {
		t.Errorf("set: %s", Render(s))
	}

	id := NewTag("Id", RecordTag, nil)
	tagged := &Tagged{Tag: id, Payload: "u\xff"}
	if got, ok := TextIn(tagged).(*Tagged); !ok || got.Payload != "u�" || got.Tag != id {
		t.Errorf("tagged: %v", got)
	}

	cycle := NewArray([]Value{"c\xff"})
	cycle.Set(0, NewArray([]Value{cycle, "d\xff"}))
	TextIn(cycle) // terminates
	inner := cycle.Values()[0].(*Array).Values()
	if inner[1] != "d�" {
		t.Errorf("cycle: %+q", inner[1])
	}
}
