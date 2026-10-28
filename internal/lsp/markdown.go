package lsp

import (
	"regexp"
	"strings"

	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/internal/types"
)

// docMarkdown renders a doc comment by the conventions of doc.md section 2: paragraphs separated
// by empty lines, lines indented beyond the comment's own space as a code block, and a
// bracketed name that known says is a declaration as code; any other text is plain, so the
// characters Markdown would read are escaped.
func docMarkdown(doc string, known func(string) bool) string {
	if doc == "" {
		return ""
	}
	var out []string
	var para, codeLines []string
	flush := func() {
		if len(codeLines) > 0 {
			out = append(out, code(strings.Join(dedent(codeLines), "\n")))
			codeLines = nil
		}
		if len(para) > 0 {
			out = append(out, prose(strings.Join(para, " "), known))
			para = nil
		}
	}
	for _, line := range strings.Split(doc, "\n") {
		switch {
		case strings.TrimSpace(line) == "":
			flush()
		case strings.HasPrefix(line, " "):
			if len(para) > 0 {
				flush()
			}
			codeLines = append(codeLines, line)
		default:
			if len(codeLines) > 0 {
				flush()
			}
			para = append(para, strings.TrimSpace(line))
		}
	}
	flush()
	return strings.Join(out, "\n\n")
}

// dedent removes the indentation the lines share.
func dedent(lines []string) []string {
	indent := -1
	for _, l := range lines {
		n := len(l) - len(strings.TrimLeft(l, " "))
		if indent < 0 || n < indent {
			indent = n
		}
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = l[min(indent, len(l)):]
	}
	return out
}

var bracketedName = regexp.MustCompile(`\[([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)?)\]`)

// prose renders a paragraph of plain text as Markdown: a bracketed name that known says is a
// declaration becomes code, since the editor cannot link to it from a hover, and the characters
// Markdown would read, as in `a * b`, are escaped. A bracketed text that names nothing stays as
// written.
func prose(s string, known func(string) bool) string {
	var sb strings.Builder
	last := 0
	for _, m := range bracketedName.FindAllStringSubmatchIndex(s, -1) {
		name := s[m[2]:m[3]]
		if known == nil || !known(name) {
			continue
		}
		sb.WriteString(escapeMarkdown(s[last:m[0]]))
		sb.WriteString("`" + name + "`")
		last = m[1]
	}
	sb.WriteString(escapeMarkdown(s[last:]))
	return sb.String()
}

// escapeMarkdown escapes the characters that would format plain text.
func escapeMarkdown(s string) string {
	if !strings.ContainsAny(s, "\\`*_{}[]()#+-!<>|~") {
		return s
	}
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '`', '*', '_', '[', ']', '<', '>', '#', '|', '~':
			sb.WriteByte('\\')
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// summary is the first sentence of a doc: up to the first period followed by a space or a line
// break (doc.md section 2).
func summary(doc string, known func(string) bool) string {
	doc = strings.TrimSpace(doc)
	for i := 0; i < len(doc); i++ {
		if doc[i] == '.' && (i+1 == len(doc) || doc[i+1] == ' ' || doc[i+1] == '\n') {
			return prose(strings.ReplaceAll(doc[:i+1], "\n", " "), known)
		}
	}
	if i := strings.IndexByte(doc, '\n'); i >= 0 {
		return prose(doc[:i], known)
	}
	return prose(doc, known)
}

// describe renders an object for hover: a code block that says what it is, and its doc.
func (a *analysis) describe(o *check.Object) (head, doc string) {
	switch o.Kind {
	case check.Func:
		head = o.Sig()
		if head == "" {
			head = "func " + a.signatureLabel(o)
		}
		return head, o.Doc()
	case check.TypeKind:
		if head = o.Sig(); head == "" {
			head = "type " + o.Name
		}
		return head, o.Doc()
	case check.Const:
		return o.Sig(), o.Doc()
	case check.Var:
		return "var " + o.Name + ": " + typeString(o.Type()), o.Doc()
	case check.Param:
		return o.Name + ": " + typeString(o.Type()), "A parameter."
	case check.Field:
		owner := ""
		if o.Owner != "" {
			owner = "A field of " + o.Owner + "."
		}
		return o.Name + ": " + typeString(o.Type()), owner
	case check.TypeParam:
		head = o.Name
		if p, ok := o.Type().(types.Param); ok && p.Decl.Bound != nil {
			head += ": " + typeString(p.Decl.Bound)
		}
		return head, "A type parameter."
	case check.Namespace:
		_, nsDoc := prelude.NamespaceDoc(o.Name)
		return o.Name, nsDoc
	case check.Module:
		head = o.Name
		if f := a.file(o.Path); f != nil {
			return "import " + quote(a.importPath(o)) + " " + o.Name, f.Doc
		}
		return head, ""
	}
	return o.Name, ""
}

// importPath is the path an import of a module object was written with, or its file path.
func (a *analysis) importPath(o *check.Object) string {
	for _, m := range a.modules {
		for _, imp := range m.File.Imports {
			if imp.Namespace == o.Name && imp.Module >= 0 && imp.Module < len(a.prog.Modules) && a.prog.Modules[imp.Module].File.Path == o.Path {
				return imp.Path
			}
		}
	}
	return o.Path
}

func quote(s string) string { return `"` + s + `"` }

// typeString renders a type, or ? for one the checker could not determine.
func typeString(t types.Type) string {
	if t == nil {
		return "?"
	}
	return types.String(t)
}

// signatureLabel renders a function's signature from its type: name(a: T, b: U = …): R, with
// the type parameters in brackets.
func (a *analysis) signatureLabel(o *check.Object) string {
	fn, ok := o.Type().(types.Func)
	if !ok {
		return o.Name
	}
	name := o.Name
	if o.Owner != "" {
		name = o.Owner + "." + name
	}
	var sb strings.Builder
	sb.WriteString(name)
	if tps := o.TypeParams(); len(tps) > 0 {
		sb.WriteString("[")
		for i, p := range tps {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(p.Name)
			if p.Bound != nil {
				sb.WriteString(": " + types.String(p.Bound))
			}
		}
		sb.WriteString("]")
	}
	sb.WriteString("(")
	sb.WriteString(strings.Join(paramLabels(fn, fn.Names, o.Defaults()), ", "))
	sb.WriteString("): ")
	sb.WriteString(typeString(fn.Return))
	return sb.String()
}

// paramLabels renders each parameter of a function type as name: T, with = … after one that has
// a default.
func paramLabels(fn types.Func, names []string, defaults []bool) []string {
	out := make([]string, len(fn.Params))
	for i, p := range fn.Params {
		label := types.String(p)
		if i < len(names) && names[i] != "" {
			label = names[i] + ": " + label
		}
		if i < len(defaults) && defaults[i] || i >= fn.Required && len(defaults) == 0 {
			label += " = …"
		}
		out[i] = label
	}
	return out
}
