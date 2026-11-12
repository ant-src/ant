<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.ErasedType

```ant signature
type Reflect.ErasedType
```

Reflect.ErasedType is a type value whose static type is not known: what
[`Reflect.erase`](../erase.md) returns, and what [`Reflect.shape`](../shape.md) takes apart. Code
that walks a type recurses on erased types, since the element type of an Array\[E\] is not known
statically one level down.

<!-- end generated -->

## Details

- Three things make a `Reflect.ErasedType`: [`Reflect.erase`](../erase.md) of a type value,
  [`Reflect.type`](../type.md) of a boxed value, and the children of a shape that
  [`Reflect.shape`](../shape.md) returns.
- [`Reflect.shape`](../shape.md), [`Reflect.assignable`](../assignable.md), and the `Reflect.make*`
  builders take an erased type.
- Two are `==` when they denote the same type, whichever of these made them, and they hash the same,
  so an erased type works as a map key. `String.from` renders the type in annotation syntax.
- No function turns it back into a `Reflect.Type[T]`. The static type comes back only at the end,
  when [`Reflect.to`](../to.md) unboxes a value with a typed `type(T)`.
- It has no `is` discriminator: a union holds at most one of `Reflect.ErasedType`,
  `Reflect.Type[T]`, `Reflect.Decl`, and `Reflect.Value`, beside `null`.
- `Json.stringify` panics on it.

## Basic examples

Erase a type value, and the result renders as the type:

```ant
let t = Reflect.erase(type(Map[string, Array[int]]))
t // Map[string, Array[int]]
```

Take it apart with `Reflect.shape`. The children are erased types too:

```ant
let t = Reflect.erase(type(Map[string, Array[int]]))
Reflect.shape(t) // Reflect.Map{ key: string, value: Array[int] }
```

## Scope

`Reflect.type` gives the erased type a value was boxed at. A union is resolved to the value's
member:

```ant
type Celsius = float64
Reflect.type(Reflect.valueOf(type(int | null), 42))            // int
Reflect.type(Reflect.valueOf(type(Celsius), Celsius(21.5)))    // Celsius
Reflect.type(Reflect.valueOf(type(Array[string]), ["a", "b"])) // Array[string]
```

Every child of a shape is erased: a set's member, a union's members, a function's parameters and
result, and likewise an array's element, a map's key and value, and a record's fields and type
arguments:

```ant
Reflect.shape(Reflect.erase(type(Set[string])))             // Reflect.Set{ member: string }
Reflect.shape(Reflect.erase(type(int | string | null)))     // Reflect.Union{ members: [int, string, null] }
Reflect.shape(Reflect.erase(type(func(int, string): bool))) // Reflect.Func{ params: [int, string], result: bool }
```

An array or a map holds erased types of different static types side by side, which `Reflect.Type`
values cannot do:

```ant
let ts: Array[Reflect.ErasedType] = [Reflect.erase(type(int)), Reflect.erase(type(Array[string]))]
ts // [int, Array[string]]
```

## Applications

Choose a form widget for each field of a record from a table keyed by erased types, with a fallback
for types the table does not list:

```ant
type Signup = { email: string, age: int, newsletter: bool, birthday: Time.PlainDate | null }

let widgets: Map[Reflect.ErasedType, string] = Map{
  Reflect.erase(type(int)): "number",
  Reflect.erase(type(bool)): "checkbox",
  Reflect.erase(type(Time.PlainDate | null)): "date",
}

let s = Reflect.shape(Reflect.erase(type(Signup)))
if s is Reflect.Record {
  for f in s.fields {
    println("${f.name}: ${widgets.get(f.type) ?? "text"}")
  }
}
```

```text
email: text
age: number
newsletter: checkbox
birthday: date
```

List every type a recursive type reaches. A shape describes one level, so a walk that records the
erased types it has seen visits each one once and ends:

```ant
type Tree = { label: string, children: Array[Tree], parent: Tree | null }

func children(t: Reflect.ErasedType): Array[Reflect.ErasedType] {
  let s = Reflect.shape(t)
  if s is Reflect.Record {
    s.fields.map({ x.type })
  } else if s is Reflect.Array {
    [s.element]
  } else if s is Reflect.Union {
    s.members
  } else {
    []
  }
}

func visit(t: Reflect.ErasedType, seen: Set[Reflect.ErasedType]): void {
  if !seen.add(t) {
    return
  }
  for c in children(t) {
    visit(c, seen)
  }
}

let seen: Set[Reflect.ErasedType] = Set.of([])
visit(Reflect.erase(type(Tree)), seen)
println(seen)
```

```text
Set.of([Tree, string, Array[Tree], Tree | null, null])
```

## Properties and relations

An erased type from `Reflect.erase`, from `Reflect.type`, or from a shape is `==` to the others when
they denote the same type:

```ant
type Box[T] = { value: T, note: string | null }
let t = Reflect.erase(type(Box[Array[int]]))
let boxed = Reflect.valueOf(type(Box[Array[int]]), Box{ value: [1], note: null })
assert(Reflect.type(boxed) == t)

let s = Reflect.shape(t)
if s is Reflect.Record {
  assert(s.args[0] == Reflect.erase(type(Array[int])))
  assert(s.fields[0].type == s.args[0])
}
```

`Reflect.assignable` compares two erased types by the assignability rules:

```ant
Reflect.assignable(Reflect.erase(type(int)), Reflect.erase(type(int | null))) // true
Reflect.assignable(Reflect.erase(type(int | null)), Reflect.erase(type(int))) // false
```

Erasing keeps the rendering of the type value:

```ant
let t = type(int | null)
assert(String.from(Reflect.erase(t)) == String.from(t))
```

## Possible issues

`Reflect.to` takes a typed `Reflect.Type[T]`, never an erased type, since the result needs a static
type:

```ant
let v = Reflect.valueOf(type(int), 1)
let t = Reflect.erase(type(int))
println(Reflect.to(v, t)) // error: cannot use Reflect.ErasedType where Reflect.Type[T] is expected
```

A `Reflect.Type[T]` is not assignable to a `Reflect.ErasedType`. Write the `Reflect.erase` call:

```ant
let t: Reflect.ErasedType = type(int) // error: cannot use Reflect.Type[int] where Reflect.ErasedType is expected
```

A union cannot hold an erased type beside another reflection type, since `is` cannot tell them
apart:

```ant
let x: Reflect.ErasedType | Reflect.Decl = Reflect.erase(type(int))
// error: a union may contain at most one type that is cannot test, since it cannot tell Reflect.ErasedType from Reflect.Decl
```

An erased type has no JSON form:

```ant
Json.stringify(Reflect.erase(type(int))) // panic: Json.stringify: a type has no JSON form
```

## See also

[Reflect](../index.md) · [Reflect.erase](../erase.md) · [Reflect.Type](Type.md) ·
[Reflect.shape](../shape.md) · [Reflect.Shape](Shape.md) · [Reflect.type](../type.md) ·
[Reflect.to](../to.md) · [Reflect.assignable](../assignable.md)
