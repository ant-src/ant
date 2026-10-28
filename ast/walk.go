package ast

// Start is where n begins on the page: its Position, except for a postfix link, an operator, an
// is test, an assignment, or a pipe, whose Position is the operator and whose first operand comes
// before it.
func Start(n Node) Pos {
	switch x := n.(type) {
	case *Field:
		return Start(x.X)
	case *Index:
		return Start(x.X)
	case *Call:
		if x.Pipe && len(x.Args) > 0 {
			return Start(x.Args[0].Value)
		}
		return Start(x.Fn)
	case *Binary:
		return Start(x.X)
	case *Is:
		return Start(x.X)
	case *Assign:
		return Start(x.Target)
	case *RecordLit:
		return Start(x.Head)
	case *TemplatePart:
		if x.Expr != nil && x.Pos == (Pos{}) {
			return Start(x.Expr)
		}
	}
	return n.Position()
}

// Before reports whether a comes before b.
func (a Pos) Before(b Pos) bool { return a.Line < b.Line || a.Line == b.Line && a.Col < b.Col }

// Contains reports whether p lies in the span of n, from its Start inclusive to its EndPosition
// exclusive.
func Contains(n Node, p Pos) bool {
	return !p.Before(Start(n)) && p.Before(n.EndPosition())
}

// Inspect walks the tree under n in source order, calling f for each node, children included
// when f returns true. Besides the expressions, statements, types, and targets, it visits the
// parts a node is made of that carry a position: imports' names, type and value parameters, the
// fields of a record type or destructuring, the entries of literals, call arguments, template
// parts, select clauses, and the nodes, elements, and attributes of markup.
func Inspect(n Node, f func(Node) bool) {
	if n == nil || !f(n) {
		return
	}
	switch x := n.(type) {
	case *File:
		for _, imp := range x.Imports {
			Inspect(imp, f)
		}
		// Decls and Stmts share the top-level vars; walk each item once, in source order.
		items := make([]Node, 0, len(x.Decls)+len(x.Stmts))
		for _, d := range x.Decls {
			items = append(items, d)
		}
		for _, s := range x.Stmts {
			if _, ok := s.(*VarDecl); !ok {
				items = append(items, s)
			}
		}
		sortNodes(items)
		for _, it := range items {
			Inspect(it, f)
		}
	case *Import:
		for i := range x.Names {
			Inspect(&x.Names[i], f)
		}
	case *ImportName:
	case *VarDecl:
		Inspect(x.Target, f)
		inspectType(x.Type, f)
		inspectExpr(x.Init, f)
	case *FuncDecl:
		for i := range x.TypeParams {
			Inspect(&x.TypeParams[i], f)
		}
		for i := range x.Params {
			Inspect(&x.Params[i], f)
		}
		inspectType(x.Return, f)
		if x.Body != nil {
			Inspect(x.Body, f)
		}
	case *TypeDecl:
		for i := range x.TypeParams {
			Inspect(&x.TypeParams[i], f)
		}
		inspectType(x.Body, f)
	case *TypeParam:
		inspectType(x.Bound, f)
	case *Param:
		inspectType(x.Type, f)
		inspectExpr(x.Default, f)
	case *TypeName:
		for _, a := range x.Args {
			Inspect(a, f)
		}
	case *RecordType:
		for i := range x.Fields {
			Inspect(&x.Fields[i], f)
		}
	case *FieldType:
		inspectType(x.Type, f)
		inspectExpr(x.Default, f)
	case *FuncType:
		for _, p := range x.Params {
			inspectType(p.Type, f)
		}
		inspectType(x.Return, f)
	case *UnionType:
		for _, m := range x.Members {
			Inspect(m, f)
		}
	case *NameTarget:
	case *RecordTarget:
		for i := range x.Fields {
			Inspect(&x.Fields[i], f)
		}
	case *RecordTargetField:
		Inspect(x.Target, f)
	case *ArrayTarget:
		for _, e := range x.Elems {
			Inspect(e, f)
		}
		if x.Rest != nil {
			Inspect(x.Rest, f)
		}
	case *Ident, *IntLit, *FloatLit, *StringLit, *BoolLit, *NullLit, *RegexLit, *VoidLit,
		*Break, *Continue, *Placeholder, *BadExpr, *BadStmt, *BadType:
	case *Template:
		for i := range x.Parts {
			Inspect(&x.Parts[i], f)
		}
	case *TemplatePart:
		inspectExpr(x.Expr, f)
	case *Html:
		for i := range x.Content {
			Inspect(&x.Content[i], f)
		}
	case *HtmlNode:
		switch {
		case x.Element != nil:
			Inspect(x.Element, f)
		case x.Expr != nil:
			Inspect(x.Expr, f)
		}
	case *HtmlElement:
		for i := range x.Attrs {
			Inspect(&x.Attrs[i], f)
		}
		for i := range x.Content {
			Inspect(&x.Content[i], f)
		}
	case *HtmlAttr:
		inspectExpr(x.Expr, f)
	case *RecordLit:
		Inspect(x.Head, f)
		for i := range x.Entries {
			Inspect(&x.Entries[i], f)
		}
	case *RecordEntry:
		inspectExpr(x.Value, f)
		inspectExpr(x.Spread, f)
	case *MapLit:
		for i := range x.Entries {
			Inspect(&x.Entries[i], f)
		}
	case *MapEntry:
		Inspect(x.Key, f)
		Inspect(x.Value, f)
	case *ArrayLit:
		for _, e := range x.Elems {
			Inspect(e.Value, f)
		}
	case *Lambda:
		Inspect(x.Body, f)
	case *FuncLit:
		for i := range x.Params {
			Inspect(&x.Params[i], f)
		}
		inspectType(x.Return, f)
		Inspect(x.Body, f)
	case *Field:
		Inspect(x.X, f)
	case *Index:
		Inspect(x.X, f)
		Inspect(x.Index, f)
	case *Call:
		if x.Pipe && len(x.Args) > 0 {
			Inspect(&x.Args[0], f)
			Inspect(x.Fn, f)
			for i := 1; i < len(x.Args); i++ {
				Inspect(&x.Args[i], f)
			}
			return
		}
		Inspect(x.Fn, f)
		for i := range x.Args {
			Inspect(&x.Args[i], f)
		}
	case *Arg:
		Inspect(x.Value, f)
	case *Unary:
		Inspect(x.X, f)
	case *Binary:
		Inspect(x.X, f)
		Inspect(x.Y, f)
	case *Is:
		Inspect(x.X, f)
		inspectExpr(x.Path, f)
	case *Assign:
		Inspect(x.Target, f)
		Inspect(x.Value, f)
	case *AssignArray:
		for _, e := range x.Elems {
			Inspect(e, f)
		}
		if x.Rest != nil {
			Inspect(x.Rest, f)
		}
	case *If:
		Inspect(x.Cond, f)
		Inspect(x.Then, f)
		if x.Else != nil {
			Inspect(x.Else, f)
		}
	case *While:
		Inspect(x.Cond, f)
		Inspect(x.Body, f)
	case *For:
		Inspect(x.Target, f)
		Inspect(x.Iter, f)
		Inspect(x.Body, f)
	case *Select:
		for _, c := range x.Clauses {
			Inspect(c, f)
		}
	case *SelectClause:
		if x.Target != nil {
			Inspect(x.Target, f)
		}
		if x.Head != nil {
			Inspect(x.Head, f)
		}
		Inspect(x.Body, f)
	case *Go:
		Inspect(x.Body, f)
	case *Defer:
		Inspect(x.Body, f)
	case *Return:
		inspectExpr(x.Value, f)
	case *Block:
		for _, s := range x.Stmts {
			Inspect(s, f)
		}
	}
}

// inspectExpr and inspectType visit an optional child, whose nil interface must not reach
// Inspect's type switch as a typed nil.
func inspectExpr(e Expr, f func(Node) bool) {
	if e != nil {
		Inspect(e, f)
	}
}

func inspectType(t Type, f func(Node) bool) {
	if t != nil {
		Inspect(t, f)
	}
}

// sortNodes orders nodes by position, stably; the top-level items are nearly sorted already.
func sortNodes(ns []Node) {
	for i := 1; i < len(ns); i++ {
		for j := i; j > 0 && Start(ns[j]).Before(Start(ns[j-1])); j-- {
			ns[j], ns[j-1] = ns[j-1], ns[j]
		}
	}
}

// Path returns the nodes that contain p, outermost first: the file, and then each node whose
// span holds p, down to the innermost. A node whose span is empty is never on the path.
func Path(f *File, p Pos) []Node {
	path := []Node{f}
	var walk func(n Node)
	walk = func(n Node) {
		Inspect(n, func(c Node) bool {
			if c == n {
				return true
			}
			if Contains(c, p) {
				path = append(path, c)
				walk(c)
				return false
			}
			return false
		})
	}
	walk(f)
	return path
}
