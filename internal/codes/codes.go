// Package codes names the rules a diagnostic of ant check can report (spec/check.md section 3).
// A code names a rule, not a message: several messages share one, and a message may be reworded
// without its code changing.
package codes

// Code is the name of a rule, as ant check prints it in brackets and -json writes it.
type Code string

const (
	Syntax           Code = "syntax"
	MissingModule    Code = "missing-module"
	Import           Code = "import"
	Undefined        Code = "undefined"
	NotExported      Code = "not-exported"
	Redeclared       Code = "redeclared"
	DeclarationOrder Code = "declaration-order"
	Misplaced        Code = "misplaced"
	WrongKind        Code = "wrong-kind"
	TypeMismatch     Code = "type-mismatch"
	NoMember         Code = "no-member"
	NeedsNarrowing   Code = "needs-narrowing"
	NotNullable      Code = "not-nullable"
	Operands         Code = "operands"
	Arguments        Code = "arguments"
	NeedsAnnotation  Code = "needs-annotation"
	NoCommonType     Code = "no-common-type"
	Bound            Code = "bound"
	InvalidUnion     Code = "invalid-union"
	TypeDeclaration  Code = "type-declaration"
	Spread           Code = "spread"
	MissingField     Code = "missing-field"
	MissingElse      Code = "missing-else"
	Try              Code = "try"
	Assign           Code = "assign"
	Lambda           Code = "lambda"
	IsTest           Code = "is-test"
	UselessTest      Code = "useless-test"
	DurationEquality Code = "duration-equality"
	Markup           Code = "markup"
	Conversion       Code = "conversion"
	VoidValue        Code = "void-value"
	Reflect          Code = "reflect"
	Internal         Code = "internal"

	// The warnings (spec/check.md section 4).
	UnusedResult       Code = "unused-result"
	UnusedImport       Code = "unused-import"
	UnusedVariable     Code = "unused-variable"
	NamedInterpolation Code = "named-interpolation"
)

// All is every code, in the order of the catalog.
var All = []Code{
	Syntax, MissingModule, Import, Undefined, NotExported, Redeclared, DeclarationOrder, Misplaced,
	WrongKind, TypeMismatch, NoMember, NeedsNarrowing, NotNullable, Operands, Arguments,
	NeedsAnnotation, NoCommonType, Bound, InvalidUnion, TypeDeclaration, Spread, MissingField,
	MissingElse, Try, Assign, Lambda, IsTest, UselessTest, DurationEquality, Markup, Conversion,
	VoidValue, Reflect, Internal,
	UnusedResult, UnusedImport, UnusedVariable, NamedInterpolation,
}

// Known reports whether c is a code of the catalog.
func Known(c Code) bool {
	for _, k := range All {
		if k == c {
			return true
		}
	}
	return false
}
