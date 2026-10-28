package lsp

import (
	"sort"
	"strings"

	"github.com/ant-src/ant/ast"
)

// foldingRanges answers textDocument/foldingRange: blocks, lists, markup elements, comment runs,
// and the imports (spec/lsp.md section 4).
func (s *Server) foldingRanges(p *FoldingRangeParams) []FoldingRange {
	d := s.doc(p.TextDocument.URI)
	if d == nil {
		return nil
	}
	a := s.analyze(d)
	m := a.module()
	out := []FoldingRange{}
	add := func(from, to ast.Pos, kind string) {
		// The last line folded is the one before the closer when the closer stands alone.
		start, end := from.Line-1, to.Line-1
		if end > start && to.Col <= 2 {
			end--
		}
		if end > start {
			out = append(out, FoldingRange{StartLine: start, EndLine: end, Kind: kind})
		}
	}
	ast.Inspect(m.File, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Block:
			add(x.Pos, x.End, "region")
		case *ast.Call:
			if len(x.Args) > 0 {
				add(x.Pos, x.End, "region")
			}
		case *ast.ArrayLit, *ast.MapLit, *ast.RecordType:
			add(n.Position(), n.EndPosition(), "region")
		case *ast.RecordLit:
			add(x.Head.EndPosition(), x.End, "region")
		case *ast.FuncDecl:
			if x.Body == nil && len(x.Params) > 0 {
				add(x.Pos, x.End, "region")
			}
		case *ast.HtmlElement:
			if x.Close != (ast.Pos{}) {
				add(x.Pos, x.Close, "region")
			} else {
				add(x.Pos, x.End, "region")
			}
		case *ast.Template:
			add(x.Pos, x.End, "region")
		case *ast.Select:
			add(x.Pos, x.End, "region")
		}
		return true
	})
	if imports := m.File.Imports; len(imports) > 1 {
		first, last := imports[0], imports[len(imports)-1]
		out = append(out, FoldingRange{StartLine: first.Pos.Line - 1, EndLine: last.End.Line - 1, Kind: "imports"})
	}
	if m.Layout != nil {
		// runs of line comments on consecutive lines, and block comments spanning lines
		run := -1
		var runStart int
		flush := func(endLine int) {
			if run >= 0 && endLine > runStart {
				out = append(out, FoldingRange{StartLine: runStart - 1, EndLine: endLine - 1, Kind: "comment"})
			}
			run = -1
		}
		for _, c := range m.Layout.Comments {
			if strings.HasPrefix(c.Text, "/*") {
				flush(run)
				if c.EndPos.Line > c.Pos.Line {
					out = append(out, FoldingRange{StartLine: c.Pos.Line - 1, EndLine: c.EndPos.Line - 1, Kind: "comment"})
				}
				continue
			}
			if run >= 0 && c.Pos.Line == run+1 {
				run = c.Pos.Line
				continue
			}
			flush(run)
			run, runStart = c.Pos.Line, c.Pos.Line
		}
		flush(run)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].StartLine != out[j].StartLine {
			return out[i].StartLine < out[j].StartLine
		}
		return out[i].EndLine > out[j].EndLine
	})
	return out
}
