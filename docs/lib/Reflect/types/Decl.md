<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.Decl

```ant signature
type Reflect.Decl
```

Reflect.Decl is a named type's declaration without its type arguments: Box\[int\] and Box\[string\]
share one. Two are == when they are the same declaration, and one renders as the declared name. It
is the key for anything that applies to every instantiation of a generic type.

<!-- end generated -->

## Details

- No literal or constructor makes a `Reflect.Decl`. You read one from the `decl` field of a
  `Reflect.Record`, `Reflect.Tagged`, or `Reflect.Opaque` shape. An anonymous record type has none,
  so its `decl` is `null`.
- It compares by identity and hashes, so it works as a `Map` key or `Set` member, and does nothing
  else. It renders as the name written after `type`, whatever name an import gives the type; a
  prelude declaration renders with its owner, as `Panic.Frame`.
- `is` cannot test it, and `Json.stringify` panics on it.

## Basic examples

Read the declaration of a record type from its shape:

```ant
type Box[T] = { value: T }

let s = Reflect.shape(Reflect.erase(type(Box[int])))
if s is Reflect.Record {
  println(s.decl)
}
```

```text
Box
```

`Box[int]` and `Box[string]` are different types with one declaration:

```ant
type Box[T] = { value: T }

func declOf(t: Reflect.ErasedType): Reflect.Decl | null {
  let s = Reflect.shape(t)
  if s is Reflect.Record || s is Reflect.Tagged || s is Reflect.Opaque { s.decl } else { null }
}

let ints = Reflect.erase(type(Box[int]))
let strings = Reflect.erase(type(Box[string]))
println(ints == strings)
println(declOf(ints) == declOf(strings))
```

```text
false
true
```

## Scope

Every named type has a declaration: a named record type, a named primitive type, or an opaque type,
from your program or from the prelude. Types with no declaration, such as `int`, an anonymous record
type, or a union such as `Result[int, Error]`, give `null`; each member of a union has its own:

```ant
type Box[T] = { value: T }
type Id[T] = string
type User = { id: Id[User], name: string }

func declOf(t: Reflect.ErasedType): Reflect.Decl | null {
  let s = Reflect.shape(t)
  if s is Reflect.Record || s is Reflect.Tagged || s is Reflect.Opaque { s.decl } else { null }
}

println(declOf(Reflect.erase(type(Box[float64]))))
println(declOf(Reflect.erase(type(Id[User]))))
println(declOf(Reflect.erase(type(Channel[int]))))
println(declOf(Reflect.erase(type(Panic.Frame))))
println(declOf(Reflect.erase(type(Time.Instant))))
println(declOf(Reflect.erase(type({ x: int, y: int }))))
println(declOf(Reflect.erase(type(int))))
println(declOf(Reflect.erase(type(Result[int, Error]))))
println(declOf(Reflect.erase(type(Ok[int]))))
```

```text
Box
Id
Channel
Panic.Frame
Time.Instant
null
null
null
Ok
```

A type declared in a block has one declaration for the whole run, however often the block runs:

```ant
func configType(): Reflect.ErasedType {
  type Config = { name: string }
  Reflect.erase(type(Config))
}

let a = Reflect.shape(configType())
let b = Reflect.shape(configType())
if a is Reflect.Record && b is Reflect.Record {
  println(a.decl == b.decl)
}
```

```text
true
```

An import renames the type in your source, not its declaration, which renders as the name written
after `type`:

```ant [geo.ant]
export type Circle = { r: float64 }
```

```ant
import "./geo.ant" { Circle: Round }
import "./geo.ant" Geo

let a = Reflect.shape(Reflect.erase(type(Round)))
let b = Reflect.shape(Reflect.erase(type(Geo.Circle)))
if a is Reflect.Record && b is Reflect.Record {
  println(a.decl)
  println(a.decl == b.decl)
}
```

```text
Circle
true
```

## Applications

Count how many instantiations of each generic type a list of types holds:

```ant
type Box[T] = { value: T }
type Pair[A, B] = { first: A, second: B }

let types = [
  Reflect.erase(type(Box[int])),
  Reflect.erase(type(Pair[int, string])),
  Reflect.erase(type(Box[string])),
  Reflect.erase(type(Box[Box[int]])),
]

let counts: Map[Reflect.Decl, int] = Map{}
for t in types {
  let s = Reflect.shape(t)
  let decl = if s is Reflect.Record { s.decl } else { null }
  if decl == null {
    continue
  }
  counts[decl] = (counts.get(decl) ?? 0) + 1
}
println(counts)
```

```text
Map{ Box: 3, Pair: 1 }
```

Give every instantiation of a generic type one encoding. Here every `Box[T]`, whatever its `T`,
encodes as its value, and every `Id[T]` as its string:

```ant
type Box[T] = { value: T }
type Id[T] = string
type User = { id: Id[User], name: Box[string], age: Box[int] }

func declOf(t: Reflect.ErasedType): Reflect.Decl | null {
  let s = Reflect.shape(t)
  if s is Reflect.Record || s is Reflect.Tagged || s is Reflect.Opaque { s.decl } else { null }
}

let transparent = Set.of([
  declOf(Reflect.erase(type(Box[int]))),
  declOf(Reflect.erase(type(Id[int]))),
])

func encode(d: Reflect.Value): string {
  let t = Reflect.type(d)
  let s = Reflect.shape(t)
  if transparent.has(declOf(t)) {
    if s is Reflect.Tagged {
      return encode(Reflect.payload(d))
    }
    return encode(Reflect.field(d, "value"))
  }
  if s is Reflect.Record {
    return s.fields.map({ "${x.name}=${encode(Reflect.field(d, x.name))}" }) | String.join(" ")
  }
  String.from(Reflect.scalar(d))
}

let ann = User{ id: Id("u-1"), name: Box{ value: "Ann" }, age: Box{ value: 31 } }
println(encode(Reflect.valueOf(type(User), ann)))
```

```text
id=u-1 name=Ann age=31
```

## Properties and relations

The declaration ignores type arguments, which the shape keeps in `args`. `String.from` gives the
declared name:

```ant
type Pair[A, B] = { first: A, second: B }

let s = Reflect.shape(Reflect.erase(type(Pair[int, string])))
let t = Reflect.shape(Reflect.erase(type(Pair[bool, bool])))
if s is Reflect.Record && t is Reflect.Record {
  assert(s.decl == t.decl)
  assert(s.args != t.args)
  assert(String.from(s.decl) == "Pair")
}
```

## Possible issues

A name is not unique. Two blocks may each declare a `Config`: the declarations render alike and are
not `==`. Key a table by the `Reflect.Decl`, not by its name:

```ant
func first(): Reflect.ErasedType {
  type Config = { name: string }
  Reflect.erase(type(Config))
}

func second(): Reflect.ErasedType {
  type Config = { name: string }
  Reflect.erase(type(Config))
}

let a = Reflect.shape(first())
let b = Reflect.shape(second())
if a is Reflect.Record && b is Reflect.Record {
  println("${a.decl} ${b.decl}")
  println(a.decl == b.decl)
}
```

```text
Config Config
false
```

`is` cannot test a `Reflect.Decl`. Narrow a union that holds one by testing its other members:

```ant
func describe(d: Reflect.Decl | int): string {
  if d is Reflect.Decl { "a declaration" } else { "a number" }
  // error: is cannot test Reflect.Decl; narrow by testing the other members
}
```

A declaration has no JSON form. Write its name with `String.from` instead:

```ant
type Box[T] = { value: T }
let s = Reflect.shape(Reflect.erase(type(Box[int])))
let d = if s is Reflect.Record { s.decl } else { null }
Json.stringify(d) // panic: Json.stringify: a declaration has no JSON form
String.from(d)    // "Box"
```

## See also

[Reflect.shape](../shape.md), [Reflect.Shape](Shape.md), [Reflect.Type](Type.md),
[Reflect.ErasedType](ErasedType.md), [Reflect.erase](../erase.md)
