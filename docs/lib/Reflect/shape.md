<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.shape

```ant signature
func Reflect.shape(t: Reflect.ErasedType): Reflect.Shape
```

Reflect.shape returns one level of t's structure: aliases expanded, unions flattened, and the
children as erased type values, so a recursive type is a finite graph that a program walks as deep
as it needs.

<!-- end generated -->

## Details

- Pass a child, such as an element, a field's type, or a union member, back to `Reflect.shape` to
  see its structure.
- An alias gives the shape of the type it stands for. A named record type gives `Reflect.Record`
  with its `decl`, and a named primitive type gives `Reflect.Tagged` with its `decl` and `base`.
- A union gives `Reflect.Union` with its members flattened and deduplicated, in the order written.
- Each call returns a new description that nothing else holds, so you may keep it or change it.

## Basic examples

```ant
Reflect.shape(Reflect.erase(type(Array[int]))) // Reflect.Array{ element: int }
Reflect.shape(Reflect.erase(type(int)))        // Reflect.Int{ bits: 64, signed: true }
Reflect.shape(Reflect.erase(type(int | null))) // Reflect.Union{ members: [int, null] }
```

Narrow the result with `is` before you read its fields:

```ant
type Point = { x: float64, y: float64 }
let s = Reflect.shape(Reflect.erase(type(Point)))
if s is Reflect.Record {
  for f in s.fields {
    println("${f.name}: ${f.type}")
  }
}
```

```text
x: float64
y: float64
```

## Scope

### Scalars

Each integer type gives its width and sign, and each float type its width. `int` is `int64` and
`byte` is `uint8`:

```ant
Reflect.shape(Reflect.erase(type(byte)))    // Reflect.Int{ bits: 8, signed: false }
Reflect.shape(Reflect.erase(type(int32)))   // Reflect.Int{ bits: 32, signed: true }
Reflect.shape(Reflect.erase(type(float32))) // Reflect.Float{ bits: 32 }
Reflect.shape(Reflect.erase(type(string)))  // Reflect.String{}
Reflect.shape(Reflect.erase(type(bool)))    // Reflect.Bool{}
Reflect.shape(Reflect.erase(type(null)))    // Reflect.Null{}
Reflect.shape(Reflect.erase(type(void)))    // Reflect.Void{}
Reflect.shape(Reflect.erase(type(never)))   // Reflect.Never{}
```

### Containers and functions

```ant
Reflect.shape(Reflect.erase(type(Map[string, int])))        // Reflect.Map{ key: string, value: int }
Reflect.shape(Reflect.erase(type(Set[string])))             // Reflect.Set{ member: string }
Reflect.shape(Reflect.erase(type(func(int, string): bool))) // Reflect.Func{ params: [int, string], result: bool }
```

### Records

A record shape holds the declaration, the type arguments, and the fields in declaration order. A
field's type has the type arguments substituted, and `hasDefault` says whether the declaration gives
the field a default:

```ant
type Box[T] = { value: T }
type Server = { host: string, port: int = 8080 }
Reflect.shape(Reflect.erase(type(Box[int]))) // Reflect.Record{ decl: Box, args: [int], fields: [Reflect.Field{ name: "value", type: int, hasDefault: false }] }
Reflect.shape(Reflect.erase(type(Server)))   // Reflect.Record{ decl: Server, args: [], fields: [Reflect.Field{ name: "host", type: string, hasDefault: false }, Reflect.Field{ name: "port", type: int, hasDefault: true }] }
```

An anonymous record type has no declaration, so `decl` is `null`:

```ant
Reflect.shape(Reflect.erase(type({ x: int }))) // Reflect.Record{ decl: null, args: [], fields: [Reflect.Field{ name: "x", type: int, hasDefault: false }] }
```

A recursive record type stops at one level: the field holds the erased type, not its expansion:

```ant
type Tree = { value: int, children: Array[Tree] }
Reflect.shape(Reflect.erase(type(Tree))) // Reflect.Record{ decl: Tree, args: [], fields: [Reflect.Field{ name: "value", type: int, hasDefault: false }, Reflect.Field{ name: "children", type: Array[Tree], hasDefault: false }] }
```

### Named primitive types

A primitive declaration is a new type, which gives `Reflect.Tagged` with its base. `Id[User]` keeps
its type argument in `args`:

```ant
type User = { name: string }
type Id[T] = int
type Celsius = float64
Reflect.shape(Reflect.erase(type(Celsius)))  // Reflect.Tagged{ decl: Celsius, args: [], base: float64 }
Reflect.shape(Reflect.erase(type(Id[User]))) // Reflect.Tagged{ decl: Id, args: [User], base: int }
```

### Aliases and unions

An alias gives the shape of its body. Nested unions flatten, and a repeated member appears once:

```ant
type Ids = Array[int]
type Circle = { r: float64 }
type Rect = { w: float64, h: float64 }
type Figure = Circle | Rect
Reflect.shape(Reflect.erase(type(Ids)))                         // Reflect.Array{ element: int }
Reflect.shape(Reflect.erase(type(Figure | null)))               // Reflect.Union{ members: [Circle, Rect, null] }
Reflect.shape(Reflect.erase(type(int | (string | null) | int))) // Reflect.Union{ members: [int, string, null] }
```

`Result` is a union of two records, and `Json.Value` a recursive union:

```ant
Reflect.shape(Reflect.erase(type(Result[int, Error]))) // Reflect.Union{ members: [Ok[int], Err[Error]] }
Reflect.shape(Reflect.erase(type(Json.Value)))         // Reflect.Union{ members: [null, bool, float64, string, Array[Json.Value], Map[string, Json.Value]] }
```

### Opaque types

Types whose values reflection does not take apart give `Reflect.Opaque`, with their declaration and
type arguments:

```ant
Reflect.shape(Reflect.erase(type(Bytes)))              // Reflect.Opaque{ decl: Bytes, args: [] }
Reflect.shape(Reflect.erase(type(Html)))               // Reflect.Opaque{ decl: Html, args: [] }
Reflect.shape(Reflect.erase(type(Regex.Pattern)))      // Reflect.Opaque{ decl: Regex.Pattern, args: [] }
Reflect.shape(Reflect.erase(type(Channel[int])))       // Reflect.Opaque{ decl: Channel, args: [int] }
Reflect.shape(Reflect.erase(type(Time.Instant)))       // Reflect.Opaque{ decl: Time.Instant, args: [] }
Reflect.shape(Reflect.erase(type(Reflect.ErasedType))) // Reflect.Opaque{ decl: Reflect.ErasedType, args: [] }
```

### Through `.` and pipes

```ant
let t = Reflect.erase(type(Set[string]))
t.shape()                                             // Reflect.Set{ member: string }
type(Array[bool]) | Reflect.erase() | Reflect.shape() // Reflect.Array{ element: bool }
```

## Applications

List the required fields of a record type, the ones a literal must give:

```ant
type Server = { host: string, port: int = 8080, debug: bool = false, name: string | null }

func required(t: Reflect.ErasedType): Array[string] {
  let s = Reflect.shape(t)
  if !(s is Reflect.Record) {
    return []
  }
  s.fields.filter({ !x.hasDefault }).map({ x.name })
}

println(required(Reflect.erase(type(Server))))
```

```text
["host", "name"]
```

Ask whether a type accepts `null`, by looking for a `Reflect.Null` member:

```ant
func acceptsNull(t: Reflect.ErasedType): bool {
  let s = Reflect.shape(t)
  s is Reflect.Null || (s is Reflect.Union && s.members.some({ Reflect.shape(x) is Reflect.Null }))
}

println(acceptsNull(Reflect.erase(type(string | null))))
println(acceptsNull(Reflect.erase(type(string))))
println(acceptsNull(Reflect.erase(type(Json.Value))))
```

```text
true
false
true
```

Name the machine type behind a type, going through named primitives to their base:

```ant
type Celsius = float64

func machineType(t: Reflect.ErasedType): string {
  let s = Reflect.shape(t)
  if s is Reflect.Int {
    let prefix = if s.signed { "int" } else { "uint" }
    "${prefix}${s.bits}"
  } else if s is Reflect.Float {
    "float${s.bits}"
  } else if s is Reflect.Tagged {
    machineType(s.base)
  } else {
    String.from(t)
  }
}

let types = [
  Reflect.erase(type(int)),
  Reflect.erase(type(byte)),
  Reflect.erase(type(Celsius)),
  Reflect.erase(type(string)),
]
for t in types {
  println(machineType(t))
}
```

```text
int64
uint8
float64
string
```

Write a TypeScript interface for a record type. Each field's type is translated one level at a time,
and a named record is referred to by name, so the recursive `replies` field ends the walk:

```ant
type UserId = int
type Post = { id: UserId, title: string, tags: Array[string], replies: Array[Post], author: string | null }

func ts(t: Reflect.ErasedType): string {
  let s = Reflect.shape(t)
  if s is Reflect.Null {
    "null"
  } else if s is Reflect.Bool {
    "boolean"
  } else if s is Reflect.Int || s is Reflect.Float {
    "number"
  } else if s is Reflect.String {
    "string"
  } else if s is Reflect.Tagged {
    ts(s.base)
  } else if s is Reflect.Array {
    "${ts(s.element)}[]"
  } else if s is Reflect.Map {
    "Record<${ts(s.key)}, ${ts(s.value)}>"
  } else if s is Reflect.Union {
    s.members | Array.map(ts) | String.join(" | ")
  } else if s is Reflect.Record && s.decl != null {
    String.from(s.decl)
  } else {
    panic("no TypeScript type for ${t}")
  }
}

func iface(t: Reflect.ErasedType): string {
  let s = Reflect.shape(t)
  if !(s is Reflect.Record) {
    panic("${t} is not a record type")
  }
  let fields = s.fields | Array.map({ "  ${x.name}: ${ts(x.type)};" }) | String.join("\n")
  "interface ${t} {\n${fields}\n}"
}

println(iface(Reflect.erase(type(Post))))
```

```text
interface Post {
  id: number;
  title: string;
  tags: string[];
  replies: Post[];
  author: string | null;
}
```

Collect every type a type reaches. A recursive type is a finite graph, and a set of the types seen
so far stops the walk at the first repeat:

```ant
type Tag = { name: string }
type Post = { title: string, tags: Set[Tag], replies: Array[Post], author: string | null }

func children(t: Reflect.ErasedType): Array[Reflect.ErasedType] {
  let s = Reflect.shape(t)
  if s is Reflect.Record {
    s.fields.map({ x.type })
  } else if s is Reflect.Array {
    [s.element]
  } else if s is Reflect.Set {
    [s.member]
  } else if s is Reflect.Map {
    [s.key, s.value]
  } else if s is Reflect.Union {
    s.members
  } else {
    []
  }
}

func reach(t: Reflect.ErasedType, seen: Set[Reflect.ErasedType]): void {
  if !seen.add(t) {
    return
  }
  for c in children(t) {
    reach(c, seen)
  }
}

let seen: Set[Reflect.ErasedType] = Set.of([])
reach(Reflect.erase(type(Post)), seen)
println(seen.values())
```

```text
[Post, string, Set[Tag], Tag, Array[Post], string | null, null]
```

## Properties and relations

An alias and its body give equal shapes, and so do `int` and `int64`:

```ant
type Ids = Array[int]
assert(Reflect.shape(Reflect.erase(type(Ids))) == Reflect.shape(Reflect.erase(type(Array[int]))))
assert(Reflect.shape(Reflect.erase(type(int))) == Reflect.shape(Reflect.erase(type(int64))))
```

The children are the erased types themselves, so they compare with `==` against `Reflect.erase`:

```ant
type Tree = { value: int, children: Array[Tree] }
let s = Reflect.shape(Reflect.erase(Reflect.arrayOf(type(Tree))))
assert(s == Reflect.Array{ element: Reflect.erase(type(Tree)) })
```

Two calls return equal descriptions but never the same one, and changing one leaves the next call
untouched:

```ant
type Box[T] = { value: T }
let t = Reflect.erase(type(Box[int]))
let a = Reflect.shape(t)
let b = Reflect.shape(t)
assert(a == b)
if a is Reflect.Record && b is Reflect.Record {
  assert(!Record.same(a, b))
  a.fields.push(Reflect.Field{ name: "extra", type: t, hasDefault: false })
  assert(b.fields.length() == 1)
}
assert(Reflect.shape(t) == b)
```

`Box[int]` and `Box[string]` share a declaration and differ in their type arguments:

```ant
type Box[T] = { value: T }
let a = Reflect.shape(Reflect.erase(type(Box[int])))
let b = Reflect.shape(Reflect.erase(type(Box[string])))
if a is Reflect.Record && b is Reflect.Record {
  assert(a.decl == b.decl)
  assert(a.args != b.args)
}
```

## Possible issues

The result is a union of shapes, so reading a field before narrowing it is a compile error:

```ant
type Point = { x: float64, y: float64 }
let s = Reflect.shape(Reflect.erase(type(Point)))
println(s.fields) // error: Reflect.Shape has no field fields on every member: narrow the value first
```

`Reflect.shape` takes an erased type. Passing `type(T)` itself is a compile error:

```ant
println(Reflect.shape(type(int))) // error: cannot use Reflect.Type[int] where Reflect.ErasedType is expected
```

A nullable type is a union, not its non-null member, so a test for one shape misses it. Handle
`Reflect.Union` too:

```ant
Reflect.shape(Reflect.erase(type(string | null))) is Reflect.String // false
```

A named primitive type is `Reflect.Tagged`, not the shape of its base. Follow `base` to reach it:

```ant
type Celsius = float64
Reflect.shape(Reflect.erase(type(Celsius))) is Reflect.Float // false
```

Union members come in the order written. Type values compare members as a set, but shapes compare
their arrays in order, so equal types can give unequal shapes:

```ant
let a = Reflect.erase(type(int | null))
let b = Reflect.erase(type(null | int))
a == b                               // true
Reflect.shape(a) == Reflect.shape(b) // false
```

Test for a particular type by comparing type values, as in `t == Reflect.erase(type(Bytes))`, not by
its declared name. A name is not unique: two modules, or two blocks, may each declare a `Config`,
and a test such as `String.from(s.decl) == "Config"` matches both. To match every instantiation of a
generic type, compare `decl`, which compares declarations:

```ant
func declOf(t: Reflect.ErasedType): Reflect.Decl | null {
  let s = Reflect.shape(t)
  if s is Reflect.Record { s.decl } else { null }
}

func clientConfig(): Reflect.Decl | null {
  type Config = { name: string }
  declOf(Reflect.erase(type(Config)))
}

func serverConfig(): Reflect.Decl | null {
  type Config = { port: int }
  declOf(Reflect.erase(type(Config)))
}

let a = clientConfig()
let b = serverConfig()
println(String.from(a) == String.from(b))
println(a == b)
```

```text
true
false
```

## See also

[Reflect.Shape](types/Shape.md) · [Reflect.erase](erase.md) ·
[Reflect.ErasedType](types/ErasedType.md) · [Reflect.assignable](assignable.md) ·
[Reflect.type](type.md) · [Reflect.field](field.md) · [Reflect.makeRecord](makeRecord.md) ·
[Reflect](index.md)
