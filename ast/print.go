package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ant-src/ant/value"
)

// Sprint renders a node as an S-expression, for tests and debugging.
func Sprint(n Node) string {
	var sb strings.Builder
	sprint(&sb, n)
	return sb.String()
}

func sprint(sb *strings.Builder, n Node) {
	switch x := n.(type) {
	case nil:
		sb.WriteString("nil")
	case *File:
		sb.WriteString("(file")
		for _, imp := range x.Imports {
			sb.WriteString(" ")
			sprint(sb, imp)
		}
		for _, d := range x.Decls {
			sb.WriteString(" ")
			sprint(sb, d)
		}
		for _, s := range x.Stmts {
			if _, ok := s.(*VarDecl); ok {
				continue
			}
			sb.WriteString(" ")
			sprint(sb, s)
		}
		sb.WriteString(")")
	case *Import:
		fmt.Fprintf(sb, "(import %q", x.Path)
		if x.Namespace != "" {
			sb.WriteString(" " + x.Namespace)
		} else {
			sb.WriteString(" {")
			for i, n := range x.Names {
				if i > 0 {
					sb.WriteString(" ")
				}
				sb.WriteString(n.Name)
				if n.Alias != n.Name {
					sb.WriteString(":" + n.Alias)
				}
			}
			sb.WriteString("}")
		}
		sb.WriteString(")")
	case *VarDecl:
		sb.WriteString("(var ")
		if x.Export {
			sb.WriteString("export ")
		}
		sprint(sb, x.Target)
		if x.Type != nil {
			sb.WriteString(":")
			sprint(sb, x.Type)
		}
		sb.WriteString(" ")
		sprint(sb, x.Init)
		sb.WriteString(")")
	case *FuncDecl:
		sb.WriteString("(func ")
		if x.Export {
			sb.WriteString("export ")
		}
		if x.Owner != "" {
			sb.WriteString(x.Owner + ".")
		}
		sb.WriteString(x.Name)
		typeParams(sb, x.TypeParams)
		params(sb, x.Params)
		if x.Return != nil {
			sb.WriteString(":")
			sprint(sb, x.Return)
		}
		if x.Body != nil {
			sb.WriteString(" ")
			sprint(sb, x.Body)
		}
		sb.WriteString(")")
	case *TypeDecl:
		sb.WriteString("(type ")
		if x.Export {
			sb.WriteString("export ")
		}
		if x.Owner != "" {
			sb.WriteString(x.Owner + ".")
		}
		sb.WriteString(x.Name)
		typeParams(sb, x.TypeParams)
		if x.Body != nil {
			sb.WriteString(" ")
			sprint(sb, x.Body)
		}
		sb.WriteString(")")
	case *TypeName:
		sb.WriteString(strings.Join(x.Path, "."))
		if len(x.Args) > 0 {
			sb.WriteString("[")
			for i, a := range x.Args {
				if i > 0 {
					sb.WriteString(" ")
				}
				sprint(sb, a)
			}
			sb.WriteString("]")
		}
	case *RecordType:
		sb.WriteString("{")
		for i, f := range x.Fields {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(f.Name + ":")
			sprint(sb, f.Type)
		}
		sb.WriteString("}")
	case *FuncType:
		sb.WriteString("(func (")
		for i, p := range x.Params {
			if i > 0 {
				sb.WriteString(" ")
			}
			if p.Name != "" {
				sb.WriteString(p.Name + ":")
			}
			sprint(sb, p.Type)
		}
		sb.WriteString("):")
		sprint(sb, x.Return)
		sb.WriteString(")")
	case *UnionType:
		sb.WriteString("(| ")
		for i, m := range x.Members {
			if i > 0 {
				sb.WriteString(" ")
			}
			sprint(sb, m)
		}
		sb.WriteString(")")
	case *NameTarget:
		sb.WriteString(x.Name)
	case *RecordTarget:
		sb.WriteString("{")
		for i, f := range x.Fields {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(f.Name)
			if nt, ok := f.Target.(*NameTarget); !ok || nt.Name != f.Name {
				sb.WriteString(":")
				sprint(sb, f.Target)
			}
		}
		sb.WriteString("}")
	case *ArrayTarget:
		sb.WriteString("[")
		for i, e := range x.Elems {
			if i > 0 {
				sb.WriteString(" ")
			}
			sprint(sb, e)
		}
		if x.Rest != nil {
			if len(x.Elems) > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString("..." + x.Rest.Name)
		}
		sb.WriteString("]")
	case *Ident:
		sb.WriteString(x.Name)
	case *IntLit:
		sb.WriteString(strconv.FormatInt(x.Value, 10))
	case *FloatLit:
		sb.WriteString(strconv.FormatFloat(x.Value, 'g', -1, 64))
	case *StringLit:
		sb.WriteString(strconv.Quote(x.Value))
	case *RegexLit:
		sb.WriteString(value.RegexLiteral(x.Source, x.Flags))
	case *BoolLit:
		sb.WriteString(strconv.FormatBool(x.Value))
	case *NullLit:
		sb.WriteString("null")
	case *VoidLit:
		sb.WriteString("void")
	case *Template:
		sb.WriteString("(template")
		for _, part := range x.Parts {
			sb.WriteString(" ")
			if part.Expr != nil {
				sprint(sb, part.Expr)
			} else {
				sb.WriteString(strconv.Quote(part.Text))
			}
		}
		sb.WriteString(")")
	case *Html:
		sb.WriteString("(html")
		sprintHtml(sb, x.Content)
		sb.WriteString(")")
	case *RecordLit:
		sb.WriteString("(")
		sprint(sb, x.Head)
		for _, e := range x.Entries {
			sb.WriteString(" ")
			if e.Spread != nil {
				sb.WriteString("...")
				sprint(sb, e.Spread)
			} else {
				sb.WriteString(e.Name + ":")
				sprint(sb, e.Value)
			}
		}
		sb.WriteString(")")
	case *MapLit:
		sb.WriteString("(Map")
		for _, e := range x.Entries {
			sb.WriteString(" ")
			sprint(sb, e.Key)
			sb.WriteString(":")
			sprint(sb, e.Value)
		}
		sb.WriteString(")")
	case *ArrayLit:
		sb.WriteString("[")
		for i, e := range x.Elems {
			if i > 0 {
				sb.WriteString(" ")
			}
			if e.Spread {
				sb.WriteString("...")
			}
			sprint(sb, e.Value)
		}
		sb.WriteString("]")
	case *Lambda:
		sb.WriteString("(lambda")
		stmts(sb, x.Body.Stmts)
		sb.WriteString(")")
	case *FuncLit:
		sb.WriteString("(func")
		params(sb, x.Params)
		if x.Return != nil {
			sb.WriteString(":")
			sprint(sb, x.Return)
		}
		sb.WriteString(" ")
		sprint(sb, x.Body)
		sb.WriteString(")")
	case *Field:
		op := "."
		if x.Optional {
			op = "?."
		}
		sb.WriteString("(" + op + " ")
		sprint(sb, x.X)
		sb.WriteString(" " + x.Name + ")")
	case *Index:
		op := "[]"
		if x.Optional {
			op = "?.[]"
		}
		sb.WriteString("(" + op + " ")
		sprint(sb, x.X)
		sb.WriteString(" ")
		sprint(sb, x.Index)
		sb.WriteString(")")
	case *Call:
		switch {
		case x.Pipe:
			sb.WriteString("(pipe ")
		case x.Optional:
			sb.WriteString("(?.call ")
		default:
			sb.WriteString("(call ")
		}
		sprint(sb, x.Fn)
		for _, a := range x.Args {
			sb.WriteString(" ")
			if a.Name != "" {
				sb.WriteString(a.Name + ":")
			}
			sprint(sb, a.Value)
		}
		sb.WriteString(")")
	case *Unary:
		sb.WriteString("(" + x.Op + " ")
		sprint(sb, x.X)
		sb.WriteString(")")
	case *Binary:
		sb.WriteString("(" + x.Op + " ")
		sprint(sb, x.X)
		sb.WriteString(" ")
		sprint(sb, x.Y)
		sb.WriteString(")")
	case *Is:
		sb.WriteString("(is ")
		sprint(sb, x.X)
		sb.WriteString(" ")
		if x.Path != nil {
			sprint(sb, x.Path)
		} else {
			sb.WriteString(x.Kind)
		}
		sb.WriteString(")")
	case *Assign:
		sb.WriteString("(" + x.Op + " ")
		sprint(sb, x.Target)
		sb.WriteString(" ")
		sprint(sb, x.Value)
		sb.WriteString(")")
	case *AssignArray:
		sb.WriteString("[")
		for i, e := range x.Elems {
			if i > 0 {
				sb.WriteString(" ")
			}
			sprint(sb, e)
		}
		if x.Rest != nil {
			if len(x.Elems) > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString("...")
			sprint(sb, x.Rest)
		}
		sb.WriteString("]")
	case *If:
		sb.WriteString("(if ")
		sprint(sb, x.Cond)
		sb.WriteString(" ")
		sprint(sb, x.Then)
		if x.Else != nil {
			sb.WriteString(" ")
			sprint(sb, x.Else)
		}
		sb.WriteString(")")
	case *While:
		sb.WriteString("(while ")
		sprint(sb, x.Cond)
		sb.WriteString(" ")
		sprint(sb, x.Body)
		sb.WriteString(")")
	case *For:
		sb.WriteString("(for ")
		sprint(sb, x.Target)
		sb.WriteString(" ")
		sprint(sb, x.Iter)
		sb.WriteString(" ")
		sprint(sb, x.Body)
		sb.WriteString(")")
	case *Select:
		sb.WriteString("(select")
		for _, c := range x.Clauses {
			sb.WriteString(" (")
			if c.Target != nil {
				sb.WriteString("var ")
				sprint(sb, c.Target)
				sb.WriteString(" = ")
			}
			if c.Head == nil {
				sb.WriteString("else")
			} else {
				sprint(sb, c.Head)
			}
			sb.WriteString(" ")
			sprint(sb, c.Body)
			sb.WriteString(")")
		}
		sb.WriteString(")")
	case *Go:
		sb.WriteString("(go ")
		sprint(sb, x.Body)
		sb.WriteString(")")
	case *Defer:
		sb.WriteString("(defer ")
		sprint(sb, x.Body)
		sb.WriteString(")")
	case *Return:
		sb.WriteString("(return")
		if x.Value != nil {
			sb.WriteString(" ")
			sprint(sb, x.Value)
		}
		sb.WriteString(")")
	case *Break:
		sb.WriteString("(break)")
	case *Continue:
		sb.WriteString("(continue)")
	case *Placeholder:
		sb.WriteString("...")
	case *Block:
		sb.WriteString("{")
		stmts(sb, x.Stmts)
		sb.WriteString("}")
	default:
		fmt.Fprintf(sb, "%T", n)
	}
}

func stmts(sb *strings.Builder, ss []Stmt) {
	for i, s := range ss {
		if i > 0 {
			sb.WriteString(";")
		}
		sb.WriteString(" ")
		sprint(sb, s)
	}
	if len(ss) > 0 {
		sb.WriteString(" ")
	}
}

func typeParams(sb *strings.Builder, tps []TypeParam) {
	if len(tps) == 0 {
		return
	}
	sb.WriteString("[")
	for i, tp := range tps {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(tp.Name)
		if tp.Bound != nil {
			sb.WriteString(":")
			sprint(sb, tp.Bound)
		}
	}
	sb.WriteString("]")
}

func params(sb *strings.Builder, ps []Param) {
	sb.WriteString("(")
	for i, p := range ps {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(p.Name + ":")
		sprint(sb, p.Type)
		if p.Default != nil {
			sb.WriteString("=")
			sprint(sb, p.Default)
		}
	}
	sb.WriteString(")")
}

func sprintHtml(sb *strings.Builder, content []HtmlNode) {
	for _, n := range content {
		sb.WriteString(" ")
		switch {
		case n.Element != nil:
			sb.WriteString("(<" + n.Element.Name)
			for _, a := range n.Element.Attrs {
				sb.WriteString(" ")
				switch {
				case a.Spread:
					sb.WriteString("...")
					sprint(sb, a.Expr)
				case a.Expr != nil:
					sb.WriteString(a.Name + "=")
					sprint(sb, a.Expr)
				default:
					sb.WriteString(a.Name + "=" + strconv.Quote(a.Text))
				}
			}
			sb.WriteString(">")
			sprintHtml(sb, n.Element.Content)
			sb.WriteString(")")
		case n.Expr != nil:
			sprint(sb, n.Expr)
		default:
			sb.WriteString(strconv.Quote(n.Text))
		}
	}
}
