package types

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ant-src/ant/value"
)

func decls() (point, boxDecl, id, shape *Decl, tParam *ParamDecl) {
	point = &Decl{Name: "Point", Kind: RecordDecl, Fields: []Field{{"x", Float64Type}, {"y", Float64Type}}}
	tParam = &ParamDecl{Name: "T"}
	boxDecl = &Decl{Name: "Box", Kind: RecordDecl, Params: []*ParamDecl{tParam}}
	boxDecl.Fields = []Field{{"value", Param{tParam}}}
	id = &Decl{Name: "UserId", Kind: PrimitiveDecl, Base: value.KindString}
	shape = &Decl{Name: "Shape", Kind: AliasDecl}
	shape.Body = Union{Members: []Type{Named{Decl: point}, Named{Decl: id}}}
	return
}

func TestStringAndEqual(t *testing.T) {
	point, boxDecl, id, shape, tParam := decls()
	box := func(arg Type) Type { return Named{Decl: boxDecl, Args: []Type{arg}} }
	cases := map[string]Type{
		"int":                           IntType,
		"byte":                          ByteType,
		"null":                          Null{},
		"void":                          Void{},
		"never":                         Never{},
		"Array[string]":                 Array{StringType},
		"Map[string, int]":              Map{StringType, IntType},
		"func(int, string): bool":       Func{Params: []Type{IntType, StringType}, Return: BoolType, Required: 2},
		"func(int, string = ...): bool": Func{Params: []Type{IntType, StringType}, Return: BoolType, Required: 1},
		"{ x: float64, y: float64 }":    Record{Fields: point.Fields},
		"{}":                            Record{},
		"Point":                         Named{Decl: point},
		"Box[int]":                      box(IntType),
		"Box[Box[T]]":                   box(box(Param{tParam})),
		"UserId":                        Named{Decl: id},
		"Shape":                         Named{Decl: shape},
		"int | null":                    Union{Members: []Type{IntType, Null{}}},
		"(func(): int) | null":          Union{Members: []Type{Func{Return: IntType}, Null{}}},
		"T":                             Param{tParam},
	}
	for want, typ := range cases {
		if got := String(typ); got != want {
			t.Errorf("String(%#v) = %q, want %q", typ, got, want)
		}
		if !Equal(typ, typ) {
			t.Errorf("%s is not equal to itself", want)
		}
	}
	if !strings.Contains(String(Invalid{}), "invalid") && String(Invalid{}) != "?" {
		t.Logf("Invalid renders as %q", String(Invalid{}))
	}
	if Equal(IntType, Float64Type) || Equal(box(IntType), box(StringType)) || Equal(Array{IntType}, Array{StringType}) ||
		Equal(Named{Decl: point}, Record{Fields: point.Fields}) || Equal(Func{Return: IntType}, Func{Params: []Type{IntType}, Return: IntType, Required: 1}) {
		t.Error("distinct types compare equal")
	}
	if !Equal(Union{Members: []Type{IntType, StringType}}, Union{Members: []Type{StringType, IntType}}) {
		t.Error("unions compare as sets")
	}
	if !Equal(Named{Decl: shape}, shape.Body) {
		t.Error("an alias equals its body")
	}
	if Equal(Map{StringType, IntType}, Map{IntType, StringType}) {
		t.Error("map key and value are ordered")
	}
}

func TestAssignable(t *testing.T) {
	point, boxDecl, id, shape, _ := decls()
	box := func(arg Type) Type { return Named{Decl: boxDecl, Args: []Type{arg}} }
	nullable := func(x Type) Type { return Union{Members: []Type{x, Null{}}} }
	yes := [][2]Type{
		{IntType, IntType},
		{Never{}, IntType},
		{Null{}, nullable(IntType)},
		{IntType, nullable(IntType)},
		{Union{Members: []Type{IntType, Null{}}}, Union{Members: []Type{IntType, StringType, Null{}}}},
		{Named{Decl: point}, Record{Fields: []Field{{"x", Float64Type}}}},
		{Named{Decl: point}, Record{}},
		{Named{Decl: point}, Named{Decl: shape}},
		{Named{Decl: shape}, Union{Members: []Type{Named{Decl: point}, Named{Decl: id}}}},
		{box(IntType), box(IntType)},
		{Func{Params: []Type{Named{Decl: shape}}, Return: Named{Decl: point}, Required: 1}, Func{Params: []Type{Named{Decl: point}}, Return: Named{Decl: shape}, Required: 1}},
		{Func{Params: []Type{IntType, StringType}, Return: IntType, Required: 1}, Func{Params: []Type{IntType}, Return: IntType, Required: 1}},
		{Invalid{}, IntType},
		{IntType, Invalid{}},
		{Void{}, Void{}},
	}
	no := [][2]Type{
		{IntType, Float64Type},
		{ByteType, IntType},
		{Named{Decl: id}, StringType},
		{StringType, Named{Decl: id}},
		{Null{}, IntType},
		{Void{}, IntType},
		{IntType, Void{}},
		{nullable(IntType), IntType},
		{Record{Fields: []Field{{"x", Float64Type}}}, Named{Decl: point}},
		{Named{Decl: point}, Record{Fields: []Field{{"x", nullable(Float64Type)}}}},
		{Named{Decl: point}, Record{Fields: []Field{{"z", Float64Type}}}},
		{box(IntType), box(StringType)},
		{Array{Named{Decl: point}}, Array{Named{Decl: shape}}},
		{Func{Params: []Type{Named{Decl: point}}, Return: IntType, Required: 1}, Func{Params: []Type{Named{Decl: shape}}, Return: IntType, Required: 1}},
		{Func{Params: []Type{IntType, StringType}, Return: IntType, Required: 2}, Func{Params: []Type{IntType}, Return: IntType, Required: 1}},
		{Func{Return: IntType}, Func{Return: Void{}}},
	}
	for _, c := range yes {
		if !Assignable(c[0], c[1]) {
			t.Errorf("%s should be assignable to %s", String(c[0]), String(c[1]))
		}
	}
	for _, c := range no {
		if Assignable(c[0], c[1]) {
			t.Errorf("%s should not be assignable to %s", String(c[0]), String(c[1]))
		}
	}
	if f := FieldsOf(Named{Decl: point}); len(f) != 2 || f[0].Name != "x" {
		t.Errorf("FieldsOf(Point) = %v", f)
	}
	if FieldsOf(IntType) != nil {
		t.Error("FieldsOf(int)")
	}
}

func TestUnionsAndSubst(t *testing.T) {
	point, boxDecl, id, shape, tParam := decls()
	box := func(arg Type) Type { return Named{Decl: boxDecl, Args: []Type{arg}} }
	u, err := Join(IntType, Null{}, Null{}, Never{}, IntType)
	if err != nil || String(u) != "int | null" {
		t.Errorf("Join flattens and deduplicates: %s %v", String(u), err)
	}
	if one, _ := Join(IntType, IntType); !Equal(one, IntType) {
		t.Error("a union of one member is that member")
	}
	if none, _ := Join(Never{}); !IsNever(none) {
		t.Error("a union of never is never")
	}
	for _, bad := range [][]Type{
		{Array{IntType}, Array{StringType}},
		{box(IntType), box(StringType)},
		{Record{}, Named{Decl: point}},
		{Record{Fields: []Field{{"a", IntType}}}, Record{Fields: []Field{{"b", IntType}}}},
		{Param{tParam}, Array{IntType}},
		{Param{tParam}, box(IntType)},
		{Named{Decl: &Decl{Name: "Mutex", Kind: OpaqueDecl}}, Named{Decl: &Decl{Name: "Cond", Kind: OpaqueDecl}}},
	} {
		if _, err := Join(bad...); err == nil {
			t.Errorf("%s | %s should be rejected", String(bad[0]), String(bad[1]))
		}
	}
	for _, fine := range [][]Type{
		{IntType, Float64Type},
		{Param{tParam}, Null{}, Named{Decl: point}},
		{Named{Decl: &Decl{Name: "Mutex", Kind: OpaqueDecl}}, Null{}},
		{Record{}, Null{}},
	} {
		if _, err := Join(fine...); err != nil {
			t.Errorf("%s: %v", String(Union{Members: fine}), err)
		}
	}
	shapeT := Named{Decl: shape}
	if m := Members(shapeT); len(m) != 2 {
		t.Errorf("Members(Shape) = %v", m)
	}
	if !Includes(shapeT, Named{Decl: id}) || Includes(shapeT, IntType) {
		t.Error("Includes")
	}
	if w := Without(shapeT, Named{Decl: id}); !Equal(w, Named{Decl: point}) {
		t.Errorf("Without = %s", String(w))
	}
	if w := Without(IntType, IntType); !IsNever(w) {
		t.Errorf("Without everything = %s", String(w))
	}
	subst := map[*ParamDecl]Type{tParam: StringType}
	got, err := Subst(box(Param{tParam}), subst)
	if err != nil || !Equal(got, box(StringType)) {
		t.Errorf("Subst: %s %v", String(got), err)
	}
	got, _ = Subst(Func{Params: []Type{Param{tParam}}, Return: Array{Param{tParam}}, Required: 1}, subst)
	if String(got) != "func(string): Array[string]" {
		t.Errorf("Subst into a func: %s", String(got))
	}
	if _, err := Subst(Union{Members: []Type{Param{tParam}, Array{IntType}}}, map[*ParamDecl]Type{tParam: Array{StringType}}); err == nil {
		t.Error("a substitution that forms an invalid union is an error")
	}
	if !Mentions(box(Param{tParam}), map[*ParamDecl]bool{tParam: true}) || Mentions(box(IntType), map[*ParamDecl]bool{tParam: true}) {
		t.Error("Mentions")
	}
	loop := &Decl{Name: "Loop", Kind: AliasDecl}
	loop.Body = Named{Decl: loop}
	func() {
		defer func() {
			if r := recover(); r == nil || !strings.Contains(fmt.Sprint(r), "internal error") {
				t.Errorf("an endless alias chain, which the checker rejects, is an internal error: %v", r)
			}
		}()
		Expand(Named{Decl: loop})
	}()
	if !Equal(Expand(Named{Decl: shape}), shape.Body) || !Equal(Expand(IntType), IntType) {
		t.Error("Expand")
	}
}
