package lexer

import (
	"testing"

	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/value"
)

// TestQuotedSpan checks that the text of a \Q...\E span is kept as written by the normalization
// of section 11, and that a pattern renders as a literal that lexes and normalizes back to the
// same source, closing and reopening a span around a / or a line break.
func TestQuotedSpan(t *testing.T) {
	for _, c := range []struct {
		src, norm, lit, match string
	}{
		{`^\Q\/\E$`, `^\Q\/\E$`, `/^\Q\\E\/\Q\E$/`, `\/`},
		{"^\\Qa\nb\\E$", "^\\Qa\nb\\E$", `/^\Qa\E\n\Qb\E$/`, "a\nb"},
		{"^\\Qa\rb\\E$", "^\\Qa\rb\\E$", `/^\Qa\E\r\Qb\E$/`, "a\rb"},
		{`^\Qa/b\E$`, `^\Qa/b\E$`, `/^\Qa\E\/\Qb\E$/`, `a/b`},
		{`^\Qa\E\/\Qb\E$`, `^\Qa/b\E$`, `/^\Qa\E\/\Qb\E$/`, `a/b`},
		{`^\Qa\E\n\Qb\E$`, "^\\Qa\nb\\E$", `/^\Qa\E\n\Qb\E$/`, "a\nb"},
		{`\Q[/`, `\Q[/\E`, `/\Q[\E\/\Q\E/`, `[/`},
		{`\Qa\`, `\Qa\\E`, `/\Qa\\E/`, `a\`},
		{`\Q\\E\/`, `\Q\\E/`, `/\Q\\E\//`, `\/`},
		{`a\/b\Q.\E`, `a/b\Q.\E`, `/a\/b\Q.\E/`, `a/b.`},
	} {
		re, err := value.NewRegex(c.src, "")
		if err != nil {
			t.Errorf("%q: %v", c.src, err)
			continue
		}
		if re.Source != c.norm {
			t.Errorf("%q: normalized to %q, want %q", c.src, re.Source, c.norm)
		}
		if !re.Re.MatchString(c.match) {
			t.Errorf("%q does not match %q", c.src, c.match)
		}
		lit := re.String()
		if lit != c.lit {
			t.Errorf("%q: renders as %s, want %s", c.src, lit, c.lit)
		}
		toks, err := Lex(lit)
		if err != nil || toks[0].Kind != token.Regex {
			t.Errorf("%s does not lex as one literal: %v %v", lit, toks, err)
			continue
		}
		back, err := value.NewRegex(toks[0].Text, toks[0].Flags)
		if err != nil || back.Source != re.Source {
			t.Errorf("%s reads back as %v, %v, want %q", lit, back, err, re.Source)
		}
	}
}
