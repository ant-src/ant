package format

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/lexer"
	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/parser"
)

// corpusDirs hold the modules of the repository: every file in them parses as a module, or as a
// declaration file when its name says so (parser.ModeOf), the prelude among them. tests/syntax
// does not parse by design, and tests/stdlib holds tables the harness reads one input per line,
// where a line may start with a regular expression literal that the whole file would read as a
// division (fmt.md section 4.3).
var corpusDirs = []string{"../../examples", "../../tests/run", "../../tests/unchecked", "../../tests/check", "../../internal/prelude"}

func corpus(t *testing.T) []string {
	t.Helper()
	var files []string
	for _, dir := range corpusDirs {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(path, ".ant") {
				files = append(files, path)
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(files) < 100 {
		t.Fatalf("only %d corpus files", len(files))
	}
	return files
}

// TestCorpusProperties checks the guarantees of fmt.md section 2 on every module in the
// repository.
func TestCorpusProperties(t *testing.T) {
	for _, path := range corpus(t) {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parser.Parse(path, string(src), parser.ModeOf(path)); err != nil {
			continue // a check program may hold a syntax error on purpose
		}
		checkProperties(t, path, string(src))
	}
}

// TestFormatted fails on any module of the repository that ant fmt would change, which is how
// continuous integration runs ant fmt --check (fmt.md section 1). Run ant fmt over the corpus
// directories to fix it.
func TestFormatted(t *testing.T) {
	for _, path := range corpus(t) {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parser.Parse(path, string(src), parser.ModeOf(path)); err != nil {
			continue
		}
		if out, err := Source(path, src); err != nil || string(out) != string(src) {
			t.Errorf("%s is not formatted; run ant fmt on it", strings.TrimPrefix(path, "../../"))
		}
	}
}

// checkProperties formats src and checks the output: the same tokens and the same tree, the
// same comments, idempotence, and the layout rules every file obeys.
func checkProperties(t *testing.T, path, src string) {
	t.Helper()
	out, err := Source(path, []byte(src))
	if err != nil {
		t.Errorf("%s: %v", path, err)
		return
	}
	got := string(out)
	before, err := parser.Parse(path, src, parser.ModeOf(path))
	if err != nil {
		t.Errorf("%s: %v", path, err)
		return
	}
	after, err := parser.Parse(path, got, parser.ModeOf(path))
	if err != nil {
		t.Errorf("%s: the output does not parse: %v\n%s", path, err, got)
		return
	}
	if ast.Sprint(before) != ast.Sprint(after) {
		t.Errorf("%s: the tree changed\n%s", path, got)
		return
	}
	if a, b := significant(t, src), significant(t, got); !slices.Equal(a, b) {
		t.Errorf("%s: the tokens changed\nbefore %q\nafter  %q", path, a, b)
	}
	if a, b := commentTexts(t, src), commentTexts(t, got); !slices.Equal(a, b) {
		t.Errorf("%s: the comments changed\nbefore %q\nafter  %q", path, a, b)
	}
	again, err := Source(path, out)
	if err != nil || string(again) != got {
		t.Errorf("%s: formatting the output changes it (err %v)\nfirst:\n%s\nsecond:\n%s", path, err, got, again)
	}
	start := 0
	for i, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		end := start + len(line)
		if strings.TrimRight(line, " \t") != line && !verbatim(got, end) {
			t.Errorf("%s:%d: trailing whitespace", path, i+1)
		}
		if strings.HasPrefix(strings.TrimLeft(line, " "), "\t") && !verbatim(got, start) && !inInterpolation(got, start) {
			t.Errorf("%s:%d: a tab in the indentation", path, i+1)
		}
		start = end + 1
	}
	if got != "" && (!strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\n\n")) {
		t.Errorf("%s: does not end with exactly one newline", path)
	}
}

// verbatim reports whether the byte at off lies inside the text of a string or a raw string or
// inside a block comment, which are kept as written (fmt.md 4.8, 4.9). The code of an
// interpolation is not.
func verbatim(src string, off int) bool {
	toks, comments, err := lexer.LexLayout(src, true)
	if err != nil {
		return false
	}
	var inText func([]token.Token) bool
	inText = func(toks []token.Token) bool {
		for _, t := range toks {
			if t.Kind == token.String && t.Off < off && off < t.End {
				return true
			}
			for _, part := range t.Parts {
				if part.Tokens == nil && part.Off <= off && off < part.End || inText(part.Tokens) {
					return true
				}
			}
		}
		return false
	}
	if inText(toks) {
		return true
	}
	for _, t := range toks {
		if t.Kind == token.Html && t.Off < off && off < t.End && !inMarkupCode(t.Markup, off) {
			return true
		}
	}
	for _, c := range comments {
		if c.Off < off && off < c.End {
			return true
		}
	}
	return false
}

// inInterpolation reports whether the line starting at off starts inside the code of a string's
// interpolation, whose indentation belongs to the string and is kept as written (fmt.md 4.9).
func inInterpolation(src string, off int) bool {
	toks, _, err := lexer.LexLayout(src, true)
	if err != nil {
		return false
	}
	for _, t := range toks {
		for _, part := range t.Parts {
			if part.Tokens != nil && part.Off < off && off < part.End {
				return true
			}
		}
		if t.Kind == token.Html && inMarkupCode(t.Markup, off) {
			return true
		}
	}
	return false
}

// significant lists the tokens of src as written, without newlines, semicolons, and the commas
// directly before a closing bracket; a template is its text and the tokens of its
// interpolations.
func significant(t *testing.T, src string) []string {
	t.Helper()
	toks, _, err := lexer.LexLayout(src, true)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	var walk func([]token.Token)
	walk = func(toks []token.Token) {
		for i, tok := range toks {
			switch {
			case tok.Kind == token.Newline || tok.Kind == token.EOF || tok.Is(";"):
				continue
			case tok.Is(","):
				j := i + 1
				for j < len(toks) && toks[j].Kind == token.Newline {
					j++
				}
				if j < len(toks) && (toks[j].Is(")") || toks[j].Is("]") || toks[j].Is("}")) {
					continue
				}
			case tok.Kind == token.Html:
				walkMarkup(tok.Markup, &out, walk)
				continue
			case tok.Kind == token.Template:
				for _, part := range tok.Parts {
					if part.Tokens == nil {
						out = append(out, "text:"+lineFeeds(src[part.Off:part.End]))
					} else {
						out = append(out, "${")
						walk(part.Tokens)
						out = append(out, "}")
					}
				}
				continue
			case tok.Kind == token.String:
				out = append(out, lineFeeds(src[tok.Off:tok.End]))
				continue
			}
			out = append(out, src[tok.Off:tok.End])
		}
	}
	walk(toks)
	return out
}

// lineFeeds writes every line break of a string's source \n, which keeps its value (fmt.md 4.1).
func lineFeeds(s string) string {
	return strings.ReplaceAll(lexer.NormalizeBreaks(s), "\r\n", "\n")
}

// commentTexts lists the comments of src, with the whitespace at the end of each line removed
// and every line break written \n.
func commentTexts(t *testing.T, src string) []string {
	t.Helper()
	_, comments, err := lexer.LexLayout(src, true)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, c := range comments {
		lines := strings.Split(lexer.NormalizeBreaks(src)[c.Off:c.End], "\n")
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], " \t\r")
		}
		out = append(out, strings.Join(lines, "\n"))
	}
	return out
}

// TestFormatCorpusDiff writes the formatted corpus to $ANT_FMT_OUT, for reading the output by
// eye while the printer is developed.
func TestFormatCorpusDiff(t *testing.T) {
	dir := os.Getenv("ANT_FMT_OUT")
	if dir == "" {
		t.Skip("ANT_FMT_OUT not set")
	}
	for _, path := range corpus(t) {
		src, _ := os.ReadFile(path)
		out, err := Source(path, src)
		if err != nil {
			continue
		}
		dst := filepath.Join(dir, strings.TrimPrefix(path, "../../"))
		os.MkdirAll(filepath.Dir(dst), 0o755)
		os.WriteFile(dst, out, 0o644)
	}
}

// walkMarkup lists the significant pieces of an HTML expression: its tags, attributes and decoded
// text, and the tokens of its code.
func walkMarkup(m *token.Markup, out *[]string, walk func([]token.Token)) {
	for _, n := range m.Nodes {
		switch n.Kind {
		case token.MarkupText:
			*out = append(*out, "text:"+n.Text)
		case token.MarkupExpr:
			*out = append(*out, "{")
			walk(n.Tokens)
			*out = append(*out, "}")
		case token.MarkupElement:
			*out = append(*out, "<"+n.Name)
			for _, a := range n.Attrs {
				switch a.Kind {
				case token.AttrExpr, token.AttrSpread:
					*out = append(*out, a.Name+"={")
					walk(a.Tokens)
					*out = append(*out, "}")
				default:
					*out = append(*out, a.Name+"="+a.Text)
				}
			}
			walkMarkup(n.Children, out, walk)
			*out = append(*out, "</"+n.Name+">")
		}
	}
}

// inMarkupCode reports whether off lies inside the code of an interpolation, attribute value, or
// spread of an HTML expression, whose indentation is kept as written (fmt.md 4.10).
func inMarkupCode(m *token.Markup, off int) bool {
	for _, n := range m.Nodes {
		switch n.Kind {
		case token.MarkupExpr:
			if n.Off < off && off < n.End {
				return true
			}
		case token.MarkupElement:
			for _, a := range n.Attrs {
				if (a.Kind == token.AttrExpr || a.Kind == token.AttrSpread) && a.Off < off && off < a.End {
					return true
				}
			}
			if inMarkupCode(n.Children, off) {
				return true
			}
		}
	}
	return false
}
