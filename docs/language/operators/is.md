<!-- status: reviewed -->

# is

`x is T` is `true` when the value of `x` belongs to the member `T` of its type, and `false`
otherwise. It tells the members of a union apart at run time, and as a condition it narrows `x` to
`T`.

## Details

- `T` is a name without type arguments: a primitive type, `null`, `void`, `Bytes`, `Html`, `Array`,
  `Map`, `Set`, `func`, a named record or primitive type, a time type, or `Regex.Pattern`.
- The static type of `x` must be a union with `T` as a member, an anonymous record type that `T` is
  assignable to, or an interface that `T` satisfies. A test that is always true or always false is a
  compile error.
- In the `then` branch `x` has type `T`. In the `else` branch a union loses `T`, and an interface
  stays the interface. Only a local variable, a parameter, or a top-level `let` narrows.
- `is` binds like `<`: tighter than `==`, `&&`, and `||`, looser than `+`.

## Basic examples

Test the member a value holds:

```ant
let id: int | string = 42
id is int    // true
id is string // false
```

Narrow a union of records in an `if`:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func area(s: Shape): float64 {
  if s is Circle { 3.14159 * s.radius * s.radius } else { s.width * s.height }
}

println(area(Circle{ radius: 1.0 }))
println(area(Rect{ width: 2.0, height: 3.0 }))
```

```text
3.14159
6.0
```

## Scope

### Primitives and `null`

Each numeric type is its own member, so `int` and `float64` stay apart:

```ant
let n: int | float64 = 1
n is int     // true
n is float64 // false
let name: string | null = null
name is null   // true
name is string // false
```

### Containers and functions

`Array`, `Map`, `Set`, and `func` take no type arguments in the test:

```ant
let v: Array[int] | Map[string, int] | func(int): int = [1, 2]
v is Array // true
v is Map   // false
v is func  // false
```

### Exhaustive chains

A chain of tests that covers every member needs no final `else`. Adding a member to the union later
makes the chain fail to check until it handles the new member.

```ant
type Pending = {}
type Shipped = { tracking: string }
type Delivered = { signedBy: string }
type Order = Pending | Shipped | Delivered

func status(o: Order): string {
  if o is Pending { "waiting" }
  else if o is Shipped { "on the way: ${o.tracking}" }
  else if o is Delivered { "signed by ${o.signedBy}" }
}

println(status(Pending{}))
println(status(Shipped{ tracking: "TR-7" }))
```

```text
waiting
on the way: TR-7
```

### Interfaces

On an interface, `is` narrows the `then` branch to the record; the `else` branch stays a `Shape`.

```ant
interface Shape {
  area(s: Self): float64
}

type Circle = { radius: float64 }
type Square = { side: float64 }

func Circle.area(c: Circle): float64 { 3.0 * c.radius * c.radius }
func Square.area(s: Square): float64 { s.side * s.side }

func describe(s: Shape): string {
  if s is Circle { "circle of radius ${s.radius}" } else { "area ${s.area()}" }
}

println(describe(Circle{ radius: 1.0 }))
println(describe(Square{ side: 2.0 }))
```

```text
circle of radius 1.0
area 4.0
```

### Named primitive types

A named primitive type has a tag of its own, so `is` tells it from its base type:

```ant
type UserId = string
let who: UserId | string = UserId("u-1")
who is UserId // true
who is string // false
```

### `Result`

`Ok` and `Err` are members like any other:

```ant
let parsed = Json.parse("[1, 2]")
parsed is Ok  // true
parsed is Err // false
```

### Generic types

A generic named type is tested by its name alone, and the branch gets the instantiation the static
type allows:

```ant
type Box[T] = { value: T }

let found: Box[int] | null = Box{ value: 41 }
if found is Box {
  println(found.value + 1)
}
```

```text
42
```

### Guards

With `return` in the branch, the narrowing holds for the rest of the block:

```ant
func total(input: int | string | null): int {
  if input is null { return 0 }
  if input is string { return input.length() }
  input * 2
}

println(total(null))
println(total("four"))
println(total(21))
```

`assert` narrows the same way:

```ant
func next(x: int | string): int {
  assert(x is int)
  x + 1
}

println(next(1))
```

```text
2
```

```text
0
4
42
```

### `!`, `&&`, and `||`

`is` binds like a comparison, so it combines with `&&` and `||` without parentheses. The narrowing
flows through them: the right side of `&&` sees the tested member, and the right side of `||` sees
the rest. `!` swaps the branches: after `!(s is Circle)` the right side sees a `Rect`.

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func describe(s: Shape): string {
  if s is Circle && s.radius > 1.0 { return "big circle" }
  if !(s is Circle) && s.width == s.height { return "square" }
  if s is Circle || s.width == s.height { return "small or square" }
  "rectangle"
}

println(describe(Circle{ radius: 2.0 }))
println(describe(Circle{ radius: 0.5 }))
println(describe(Rect{ width: 2.0, height: 2.0 }))
println(describe(Rect{ width: 1.0, height: 2.0 }))
```

```text
big circle
small or square
square
rectangle
```

### In a lambda

`is` is an expression of type `bool`, so it works as a predicate:

```ant
let values: Array[int | string] = [1, "a", 2, "b"]
values.filter({ x is int })  // [1, 2]
values.some({ x is string }) // true
values.every({ x is int })   // false
```

## Applications

### Walk a JSON value

`Json.Value` is a union of `null`, `bool`, `float64`, `string`, `Array`, and `Map`, so a chain of
`is` tests visits every shape:

```ant
func depth(v: Json.Value): int {
  if v is Array {
    return 1 + ((v | Array.map(depth) | Array.max()) ?? 0)
  }
  if v is Map {
    return 1 + ((v | Map.values() | Array.map(depth) | Array.max()) ?? 0)
  }
  0
}

let doc = Json.parse(`{"a": [1, [2, 3]], "b": "x"}`) | Result.unwrap()
println(depth(doc))
```

```text
3
```

### Handle each error case

Tagged error types let the caller react to each case. The error is bound to a local first, since a
field path such as `r.error` does not narrow:

```ant
type NotFound = { key: string }
type Expired = { key: string, age: int }

func lookup(key: string): Result[string, NotFound | Expired] {
  if key == "old" { return Result.err(Expired{ key, age: 90 }) }
  if key != "name" { return Result.err(NotFound{ key }) }
  Result.ok("Ann")
}

for key in ["name", "old", "city"] {
  let r = lookup(key)
  if r is Ok {
    println("${key}: ${r.value}")
    continue
  }
  let error = r.error
  let reason = if error is Expired { "expired after ${error.age} days" } else { "missing" }
  println("${key}: ${reason}")
}
```

```text
name: Ann
old: expired after 90 days
city: missing
```

## Properties and relations

`x is null` and `x == null` agree, and `!(x is T)` is the `else` branch of `x is T`:

```ant
let values: Array[int | string | null] = [1, "a", null]
for v in values {
  assert((v is null) == (v == null))
  assert(!(v is int) == (v is string || v is null))
}
```

## Possible issues

### An alias cannot be tested

An alias has no run-time identity. Test its members instead:

```ant
type Id = int | string
let key: Id | null = 7
let known = key is Id // error: Id is an alias and has no run-time identity; test its members instead
```

### No type arguments

The runtime sees tags and kinds, not type arguments, so the target takes none:

```ant
let v: Array[int] | string = [1]
let b = v is Array[int] // error: expected a newline or ';' before '['
```

### A test that cannot fail, or cannot pass

The value's type must be a union with `T` as one member, or an interface or anonymous record type
that `T` fits:

```ant
let n = 3
let a = n is int // error: the test is always true: int is already int
let id: int | string = 1
let b = id is bool // error: int | string is never a bool, so the test is always false
```

### An else after a complete chain

Once the tests cover every member, nothing is left for an `else`. Drop it.

```ant
type Pending = {}
type Shipped = { tracking: string }
type Order = Pending | Shipped

func status(o: Order): string {
  if o is Pending { "waiting" } else if o is Shipped { "on the way" } else { "unknown" }
  // error: this else cannot run: the tests before it cover every case
}
```

### `filter` does not narrow the element type

The array that `filter` returns keeps the element type it had. Narrow each element again, or collect
the narrowed values:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

let shapes: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
let circles = shapes.filter({ x is Circle })
let first = circles[0].radius
// error: Shape has no field radius on every member: narrow the value first
```

### A field path does not narrow

Another call could change the field between the test and the use. Bind it to a local first:

```ant
type Order = { note: string | null }
let order = Order{ note: "fragile" }
if order.note is string {
  let size = order.note.length()
  // error: string | null has no field or function length on every member: narrow the value first
}
```

```ant
type Order = { note: string | null }
let order = Order{ note: "fragile" }
let note = order.note
if note is string {
  println(note.length())
}
```

```text
7
```

### Types without a discriminator

The sync types, `Random`, resources such as `File.Handle`, interfaces, and anonymous record types
cannot be the target. Narrow by testing the other members:

```ant
let lock: Mutex | null = null
let held = lock is Mutex // error: is cannot test Mutex; narrow by testing the other members
```

```ant
let lock: Mutex | null = Mutex.new()
if lock != null {
  lock.withLock({ println("locked") })
}
```

```text
locked
```

## See also

- [if](../keywords/if.md)
- [Equality operators](equality.md)
- [Comparison operators](comparison.md)
- [Logical operators](logical.md)
- [?? (null coalescing)](null-coalescing.md)
- [type](../keywords/type.md)
- [Unions](../../guide/unions.md)
- [Run-time types](../../guide/runtime-types.md)
- [Null](../../guide/null.md)
