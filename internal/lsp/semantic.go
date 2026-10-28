package lsp

import (
	"sort"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
)

// The semantic token types and modifiers the server emits (spec/lsp.md section 4).
var semanticLegend = SemanticTokensLegend{
	TokenTypes:     []string{"namespace", "type", "typeParameter", "parameter", "variable", "property", "function", "method", "element", "attribute"},
	TokenModifiers: []string{"declaration", "defaultLibrary", "readonly"},
}

const (
	tokNamespace = iota
	tokType
	tokTypeParam
	tokParameter
	tokVariable
	tokProperty
	tokFunction
	tokMethod
	tokElement
	tokAttribute
)

const (
	modDeclaration    = 1 << 0
	modDefaultLibrary = 1 << 1
	modReadonly       = 1 << 2
)

type semToken struct {
	pos    ast.Pos
	length int
	typ    int
	mods   int
}

// semanticTokens answers textDocument/semanticTokens/full: what each identifier is.
func (s *Server) semanticTokens(p *SemanticTokensParams) *SemanticTokens {
	d := s.doc(p.TextDocument.URI)
	if d == nil {
		return nil
	}
	a := s.analyze(d)
	in := a.info
	var toks []semToken
	add := func(pos ast.Pos, name string, typ, mods int) {
		if name != "" && pos != (ast.Pos{}) {
			toks = append(toks, semToken{pos, len(name), typ, mods})
		}
	}
	object := func(pos ast.Pos, name string, o *check.Object, decl bool) {
		if o == nil {
			return
		}
		typ, mods := -1, 0
		switch o.Kind {
		case check.Namespace, check.Module:
			typ = tokNamespace
		case check.TypeKind:
			typ = tokType
		case check.TypeParam:
			typ = tokTypeParam
		case check.Param:
			typ = tokParameter
		case check.Var:
			typ = tokVariable
		case check.Field:
			typ = tokProperty
		case check.Func:
			typ = tokFunction
			if o.Owner != "" {
				typ = tokMethod
			}
		case check.Const:
			typ, mods = tokVariable, modReadonly
		}
		if typ < 0 {
			return
		}
		if decl {
			mods |= modDeclaration
		}
		if o.Path == "prelude.ant" {
			mods |= modDefaultLibrary
		}
		add(pos, name, typ, mods)
	}
	ast.Inspect(a.module().File, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			object(x.Pos, x.Name, in.Objects[x], false)
		case *ast.Field:
			object(x.NamePos, x.Name, in.Members[x], false)
		case *ast.TypeName:
			for i, o := range in.TypeNames[x] {
				if i < len(x.PathPos) {
					object(x.PathPos[i], x.Path[i], o, false)
				}
			}
		case *ast.RecordEntry:
			if x.Name != "" {
				object(x.Pos, x.Name, in.Entries[x], false)
			}
		case *ast.Call:
			if ci := in.Calls[x]; ci != nil {
				for _, arg := range x.Args {
					if arg.Name != "" {
						add(arg.Pos, arg.Name, tokParameter, 0)
					}
				}
			}
		case *ast.FuncDecl:
			object(x.NamePos, x.Name, in.Decls[x], true)
			if x.Owner != "" {
				object(x.OwnerPos, x.Owner, a.named(a.module().File.Path, x.Owner), false)
			}
		case *ast.TypeDecl:
			object(x.NamePos, x.Name, in.Decls[x], true)
		case *ast.TypeParam:
			object(x.Pos, x.Name, in.Decls[x], true)
		case *ast.NameTarget:
			object(x.Pos, x.Name, in.Decls[x], true)
		case *ast.Param:
			object(x.Pos, x.Name, in.Decls[x], true)
		case *ast.FieldType:
			object(x.Pos, x.Name, in.Decls[x], true)
		case *ast.ImportName:
			object(x.Pos, x.Name, in.Decls[x], false)
			if x.Alias != x.Name {
				object(x.AliasPos, x.Alias, in.Decls[x], true)
			}
		case *ast.Import:
			if x.Namespace != "" {
				add(x.NamespacePos, x.Namespace, tokNamespace, modDeclaration)
			}
		case *ast.HtmlElement:
			add(x.NamePos, x.Name, tokElement, 0)
			if x.Close != (ast.Pos{}) {
				add(ast.Pos{Line: x.Close.Line, Col: x.Close.Col + 2}, x.Name, tokElement, 0)
			}
		case *ast.HtmlAttr:
			if !x.Spread {
				add(x.Pos, x.Name, tokAttribute, 0)
			}
		}
		return true
	})
	sort.SliceStable(toks, func(i, j int) bool { return toks[i].pos.Before(toks[j].pos) })
	data := make([]uint32, 0, len(toks)*5)
	var prev Position
	for i, t := range toks {
		if i > 0 && t.pos == toks[i-1].pos {
			continue
		}
		start := d.position(d.offsetOf(t.pos), s.utf16)
		end := d.position(d.offsetOf(t.pos)+t.length, s.utf16)
		if end.Line != start.Line {
			continue
		}
		deltaLine := start.Line - prev.Line
		deltaChar := start.Character
		if deltaLine == 0 {
			deltaChar = start.Character - prev.Character
		}
		data = append(data, uint32(deltaLine), uint32(deltaChar), uint32(end.Character-start.Character), uint32(t.typ), uint32(t.mods))
		prev = start
	}
	return &SemanticTokens{Data: data}
}
