// Package value defines Ant's run-time values (section 12): their kinds, the tags of named
// types, records, tagged primitives, arrays, maps, functions, and sync objects, together with
// equality, hashing, rendering, arithmetic, and conversions.
//
// A Value is a Go interface value holding one of: the ten numeric Go types int8 … uint64,
// float32, float64; bool; string; the sentinels Null and Void; *Array; *Map; *Record; *Tagged;
// a Callable (functions, defined by the interpreter, and *Constructor); *Tag; or a sync object.
//
// Every slot that two goroutines can reach, a record field, an array element, a module
// variable, or a captured local, is a Slot: one machine word, an atomic pointer to an immutable
// cell. That is what makes a data race memory-safe (section 9).
package value

import (
	"sync/atomic"
)

// Value is any Ant value.
type Value = any

type nullType struct{}
type voidType struct{}
type uninitType struct{}

// Null is the value null.
var Null Value = nullType{}

// Void is the value void.
var Void Value = voidType{}

// Uninitialized marks a slot whose initializer has not run; reading it is a panic.
var Uninitialized Value = uninitType{}

// Kind classifies a value (section 12, Values).
type Kind int

const (
	KindInvalid Kind = iota
	KindInt8
	KindInt16
	KindInt32
	KindInt64
	KindUint8
	KindUint16
	KindUint32
	KindUint64
	KindFloat32
	KindFloat64
	KindBool
	KindString
	KindNull
	KindVoid
	KindArray
	KindMap
	KindRecord
	KindTagged
	KindFunc
	KindTag
	KindSync
	KindTime
)

var kindNames = [...]string{
	KindInvalid: "invalid", KindInt8: "int8", KindInt16: "int16", KindInt32: "int32",
	KindInt64: "int", KindUint8: "byte", KindUint16: "uint16", KindUint32: "uint32",
	KindUint64: "uint64", KindFloat32: "float32", KindFloat64: "float64", KindBool: "bool",
	KindString: "string", KindNull: "null", KindVoid: "void", KindArray: "array", KindMap: "map",
	KindRecord: "record", KindTagged: "tagged primitive", KindFunc: "func", KindTag: "type",
	KindSync: "sync object", KindTime: "time value",
}

func (k Kind) String() string { return kindNames[k] }

// IsNumber reports whether k is one of the ten numeric kinds.
func (k Kind) IsNumber() bool { return k >= KindInt8 && k <= KindFloat64 }

// IsInteger reports whether k is one of the eight integer kinds.
func (k Kind) IsInteger() bool { return k >= KindInt8 && k <= KindUint64 }

// IsSigned reports whether k is a signed integer kind.
func (k Kind) IsSigned() bool { return k >= KindInt8 && k <= KindInt64 }

// IsFloat reports whether k is a float kind.
func (k Kind) IsFloat() bool { return k == KindFloat32 || k == KindFloat64 }

// IsPrimitive reports whether k is one of the twelve primitive kinds a named type may wrap.
func (k Kind) IsPrimitive() bool { return k >= KindInt8 && k <= KindString }

// Bits is the width of an integer kind.
func (k Kind) Bits() int {
	switch k {
	case KindInt8, KindUint8:
		return 8
	case KindInt16, KindUint16:
		return 16
	case KindInt32, KindUint32:
		return 32
	case KindInt64, KindUint64:
		return 64
	}
	return 0
}

// KindByName maps a prelude kind name, including int and byte, to its kind.
func KindByName(name string) (Kind, bool) {
	switch name {
	case "int8":
		return KindInt8, true
	case "int16":
		return KindInt16, true
	case "int32":
		return KindInt32, true
	case "int64", "int":
		return KindInt64, true
	case "uint8", "byte":
		return KindUint8, true
	case "uint16":
		return KindUint16, true
	case "uint32":
		return KindUint32, true
	case "uint64":
		return KindUint64, true
	case "float32":
		return KindFloat32, true
	case "float64":
		return KindFloat64, true
	case "bool":
		return KindBool, true
	case "string":
		return KindString, true
	case "null":
		return KindNull, true
	case "void":
		return KindVoid, true
	case "Array":
		return KindArray, true
	case "Map":
		return KindMap, true
	case "func":
		return KindFunc, true
	}
	return KindInvalid, false
}

// KindOf returns the kind of v.
func KindOf(v Value) Kind {
	switch v.(type) {
	case int64:
		return KindInt64
	case float64:
		return KindFloat64
	case string:
		return KindString
	case bool:
		return KindBool
	case nullType:
		return KindNull
	case voidType:
		return KindVoid
	case int8:
		return KindInt8
	case int16:
		return KindInt16
	case int32:
		return KindInt32
	case uint8:
		return KindUint8
	case uint16:
		return KindUint16
	case uint32:
		return KindUint32
	case uint64:
		return KindUint64
	case float32:
		return KindFloat32
	case *Array:
		return KindArray
	case *Map:
		return KindMap
	case *Record:
		return KindRecord
	case *Tagged:
		return KindTagged
	case *Tag:
		return KindTag
	case Callable:
		return KindFunc
	case Sync:
		return KindSync
	case Time:
		return KindTime
	}
	return KindInvalid
}

// ---- Slots

type cell struct {
	v Value
}

// Slot is a one-word shared location holding a Value. A store publishes an immutable cell; a
// read copies the value out of the cell it finds. Racy stores and loads never tear.
type Slot struct {
	p atomic.Pointer[cell]
}

// Load returns the slot's value, or Uninitialized when nothing was stored.
func (s *Slot) Load() Value {
	c := s.p.Load()
	if c == nil {
		return Uninitialized
	}
	return c.v
}

// Store sets the slot's value.
func (s *Slot) Store(v Value) { s.p.Store(cellOf(v)) }

// cellOf returns an immutable cell holding v. Cells are never written after they are made, so
// one may serve any number of slots: null, void, the booleans and the ints from minCellInt to
// maxCellInt have one each, and storing them allocates nothing.
func cellOf(v Value) *cell {
	switch x := v.(type) {
	case int64:
		if x >= minCellInt && x <= maxCellInt {
			return &intCells[x-minCellInt]
		}
	case bool:
		if x {
			return &trueCell
		}
		return &falseCell
	case nullType:
		return &nullCell
	case voidType:
		return &voidCell
	}
	return &cell{v: v}
}

const (
	minCellInt = -128
	maxCellInt = 1023
)

var (
	intCells                                [maxCellInt - minCellInt + 1]cell
	trueCell, falseCell, nullCell, voidCell = cell{v: true}, cell{v: false}, cell{v: Null}, cell{v: Void}
)

func init() {
	for i := range intCells {
		intCells[i].v = int64(i + minCellInt)
	}
}

// CompareAndSwap stores v when the slot holds a value equal (by ==) to expected, and reports
// whether it did. It is lock-free: every store publishes a fresh cell, so the cell pointer
// identifies the version the comparison saw.
func (s *Slot) CompareAndSwap(expected, v Value) bool {
	for {
		c := s.p.Load()
		var cur Value = Uninitialized
		if c != nil {
			cur = c.v
		}
		if !Equal(cur, expected) {
			return false
		}
		if s.p.CompareAndSwap(c, &cell{v: v}) {
			return true
		}
	}
}

// Modify replaces the slot's value with f(old) atomically and returns the new value. f may be
// called more than once when stores race; when f returns an error nothing is stored.
func (s *Slot) Modify(f func(old Value) (Value, error)) (Value, error) {
	for {
		c := s.p.Load()
		var cur Value = Uninitialized
		if c != nil {
			cur = c.v
		}
		nv, err := f(cur)
		if err != nil {
			return nil, err
		}
		if s.p.CompareAndSwap(c, &cell{v: nv}) {
			return nv, nil
		}
	}
}

// ---- Tags

// TagKind says what a type declaration bound its name to.
type TagKind int

const (
	RecordTag    TagKind = iota // a record type: literals and is use it
	AliasTag                    // an alias: neither literals nor is accept it
	PrimitiveTag                // a named primitive type: its Constructor carries it
	SyncTag                     // Mutex and the other sync types: a namespace for their functions
	TimeTag                     // a time type of stdlib.md: is tests a time value's type by its name
)

// Tag is the run-time identity of a named type: what a type declaration binds. A record tag
// lists its field names in declaration order. Every tag carries the functions declared on its
// type as fields (section 6).
type Tag struct {
	Name   string
	Kind   TagKind
	Fields []string
	Base   Kind // PrimitiveTag: the base kind
	index  map[string]int
	funcs  map[string]Value
}

// NewTag creates a tag. fields is used for record tags only.
func NewTag(name string, kind TagKind, fields []string) *Tag {
	t := &Tag{Name: name, Kind: kind, Fields: fields, index: make(map[string]int, len(fields))}
	for i, f := range fields {
		t.index[f] = i
	}
	return t
}

// NewPrimitiveTag creates the tag of a named primitive type over base.
func NewPrimitiveTag(name string, base Kind) *Tag {
	t := NewTag(name, PrimitiveTag, nil)
	t.Base = base
	return t
}

// FieldIndex returns the position of a field of a record tag.
func (t *Tag) FieldIndex(name string) (int, bool) {
	i, ok := t.index[name]
	return i, ok
}

// SetFunc attaches a function declared on the type. It is called while loading, before any
// goroutine can read the tag.
func (t *Tag) SetFunc(name string, fn Value) {
	if t.funcs == nil {
		t.funcs = map[string]Value{}
	}
	t.funcs[name] = fn
}

// Func returns a function declared on the type.
func (t *Tag) Func(name string) (Value, bool) {
	fn, ok := t.funcs[name]
	return fn, ok
}

// Funcs lists the functions declared on the type.
func (t *Tag) Funcs() map[string]Value { return t.funcs }

// ---- Records

// Record is a record value: a tag plus its fields in the tag's order. A module's scope is a
// record too.
type Record struct {
	Tag    *Tag
	fields []Slot
}

// NewRecord creates a record with the given field values, in the tag's order. A nil entry
// leaves the field uninitialized.
func NewRecord(tag *Tag, fields []Value) *Record {
	r := NewEmptyRecord(tag)
	for i, v := range fields {
		if v != nil {
			r.fields[i].Store(v)
		}
	}
	return r
}

// Records with up to four fields are allocated together with their fields.
type (
	record1 struct {
		Record
		slots [1]Slot
	}
	record2 struct {
		Record
		slots [2]Slot
	}
	record3 struct {
		Record
		slots [3]Slot
	}
	record4 struct {
		Record
		slots [4]Slot
	}
)

// NewEmptyRecord creates a record of tag with every field uninitialized.
func NewEmptyRecord(tag *Tag) *Record {
	var r *Record
	switch len(tag.Fields) {
	case 1:
		x := &record1{}
		x.fields, r = x.slots[:], &x.Record
	case 2:
		x := &record2{}
		x.fields, r = x.slots[:], &x.Record
	case 3:
		x := &record3{}
		x.fields, r = x.slots[:], &x.Record
	case 4:
		x := &record4{}
		x.fields, r = x.slots[:], &x.Record
	default:
		r = &Record{fields: make([]Slot, len(tag.Fields))}
	}
	r.Tag = tag
	return r
}

// IsSet reports whether field i was ever stored.
func (r *Record) IsSet(i int) bool { return r.fields[i].p.Load() != nil }

// Len is the number of fields.
func (r *Record) Len() int { return len(r.fields) }

// Get reads the field at position i.
func (r *Record) Get(i int) Value { return r.fields[i].Load() }

// Set writes the field at position i.
func (r *Record) Set(i int, v Value) { r.fields[i].Store(v) }

// Slot returns the field's slot, for sharing a module variable with its importers.
func (r *Record) Slot(i int) *Slot { return &r.fields[i] }

// Field reads a field by name.
func (r *Record) Field(name string) (Value, bool) {
	i, ok := r.Tag.FieldIndex(name)
	if !ok {
		return nil, false
	}
	return r.fields[i].Load(), true
}

// SetField writes a field by name and reports whether the record has it.
func (r *Record) SetField(name string, v Value) bool {
	i, ok := r.Tag.FieldIndex(name)
	if !ok {
		return false
	}
	r.fields[i].Store(v)
	return true
}

// Values copies the fields out, in order.
func (r *Record) Values() []Value {
	out := make([]Value, len(r.fields))
	for i := range r.fields {
		out[i] = r.fields[i].Load()
	}
	return out
}

// ---- Tagged primitives and constructors

// Tagged is a value of a named primitive type: a primitive payload carrying the type's tag.
type Tagged struct {
	Tag     *Tag
	Payload Value
}

// Callable is implemented by every function value. The interpreter defines the concrete
// function types; the value package needs only the name.
type Callable interface {
	FuncName() string
}

// Constructor is the function value a named primitive type binds: it builds Tagged values and
// carries the type's functions as fields.
type Constructor struct {
	Tag *Tag
}

// FuncName returns the type's name.
func (c *Constructor) FuncName() string { return c.Tag.Name }

// Sync is implemented by the sync objects.
type Sync interface {
	SyncName() string
}

// Time is implemented by the time values of stdlib.md's Time namespace: immutable, compared by
// value, rendered as their RFC 9557 string or, for a zone, its identifier. Two time values are
// equal when they have the same type and the same key, which names the fields == compares.
type Time interface {
	TimeType() string // the name of its type's tag, such as "Time.Instant"
	Key() string      // a canonical form of the fields == compares
	String() string
}

// ---- Prelude tags

// The prelude's record types (section 11).
var (
	OkTag      = NewTag("Ok", RecordTag, []string{"value"})
	ErrTag     = NewTag("Err", RecordTag, []string{"error"})
	ErrorTag   = NewTag("Error", RecordTag, []string{"message"})
	EntryTag   = NewTag("Entry", RecordTag, []string{"key", "value"})
	IndexedTag = NewTag("Indexed", RecordTag, []string{"index", "value"})
)

// Ok builds Ok{ value: v }.
func Ok(v Value) *Record { return NewRecord(OkTag, []Value{v}) }

// Err builds Err{ error: e }.
func Err(e Value) *Record { return NewRecord(ErrTag, []Value{e}) }

// ErrorValue builds Error{ message: msg }.
func ErrorValue(msg string) *Record { return NewRecord(ErrorTag, []Value{msg}) }

// Entry builds Entry{ key: k, value: v }.
func Entry(k, v Value) *Record { return NewRecord(EntryTag, []Value{k, v}) }

// Indexed builds Indexed{ index: i, value: v }.
func Indexed(i int64, v Value) *Record { return NewRecord(IndexedTag, []Value{i, v}) }
