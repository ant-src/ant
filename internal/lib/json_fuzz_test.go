package lib

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ant-src/ant/value"
)

// FuzzJSON checks that the JSON decoder never panics and that a decoded value survives a round
// trip through stringify and parse unchanged.
func FuzzJSON(f *testing.F) {
	for _, seed := range []string{
		`{"a": [1, 2.5, "x", null, true], "b": {"c": {}}}`, `[]`, `""`, `1e400`, `-0`, `"é😀"`,
		`{"a":1}{"b":2}`, `[1,`, `nul`, `{"k": "v", "k": "w"}`, `"\ud800"`,
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, src string) {
		v, err := parseJSON(json.NewDecoder(strings.NewReader(src)))
		if err != nil {
			return
		}
		var sb strings.Builder
		if !stringify(&sb, v, nil) {
			t.Fatal("a decoded value cannot be cyclic")
		}
		again, err := parseJSON(json.NewDecoder(strings.NewReader(sb.String())))
		if err != nil {
			t.Fatalf("stringify gave %q, which does not parse: %v", sb.String(), err)
		}
		if !value.Equal(v, again) {
			t.Fatalf("round trip changed %s to %s", value.Render(v), value.Render(again))
		}
	})
}
