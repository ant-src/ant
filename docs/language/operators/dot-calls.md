<!-- status: reviewed -->

# Calls through .

`x.f(a, b)` calls the field `f` of `x` when `x` has one. Otherwise it calls the function `f` of the
namespace of `x`'s type with `x` as the first argument, so `p.norm()` is `Point.norm(p)` and
`name.trim()` is `String.trim(name)`.

## Details

- A field wins. Only when `x` has no field `f` does the call go to the namespace.
- The namespace comes from the type of `x`: its own type for a record or a named primitive type;
  `String`, `Bytes`, `Html`, `Array`, `Map`, or `Set` for those values; `Result` for `Ok` and `Err`;
  the prelude namespace that declares a prelude type, such as `File` for `File.Handle`; and `Iter`
  for a function, so sequences chain. Numbers, `bool`, `null`, and `void` have none.
- The function must take a value of `x`'s type as its first parameter. A function that takes none,
  such as `Point.origin()`, or that takes something else first, is not reached through a value.
- `x.f` without a call reads a field. A namespace function is reached through a value only by
  calling it; write `{ p.norm() }` or `Point.norm` for the function value.
- The call is checked as the direct call it stands for, with the same inference and the same errors.

## Basic examples

A function declared on a type is called through a value of that type:

```ant
type Point = { x: float64, y: float64 }
func Point.norm(p: Point): float64 { Math.sqrt(p.x * p.x + p.y * p.y) }
let p = Point{ x: 3.0, y: 4.0 }
p.norm()      // 5.0
Point.norm(p) // 5.0
```

Prelude values call the functions of their namespace the same way:

```ant
"  ant ".trim()       // "ant"
String.trim("  ant ") // "ant"
[3, 1, 2].sort()      // [1, 2, 3]
Array.sort([3, 1, 2]) // [1, 2, 3]
```

The value goes first, and the other arguments follow:

```ant
"a,b,c".split(",")     // ["a", "b", "c"]
"ant".padStart(5, ".") // "..ant"
```

## Scope

### Chains

Each call returns a value, and the next `.` looks in that value's namespace:

```ant
"  Hello, Ant ".trim().toLower().split(", ") // ["hello", "ant"]
[5, 3, 8, 1].filter({ x > 2 }).sort()        // [3, 5, 8]
```

A chain may switch namespaces. Here `String.split` returns an array, so `.map` is `Array.map`:

```ant
"1 2 3".split(" ").map({ x.toInt() ?? 0 }) // [1, 2, 3]
```

### Fields that hold functions

A field that holds a function is called through `.` like any field; `g` itself is not passed as an
argument:

```ant
type Greeter = { greet: func(string): string }
let g = Greeter{ greet: { "Hello, ${x}" } }
g.greet("Ann") // "Hello, Ann"
```

### Generic functions on types

A generic function infers its type arguments from the receiver as from any first argument:

```ant
type Box[T] = { value: T }
func Box.map[T, U](b: Box[T], f: func(T): U): Box[U] { Box{ value: f(b.value) } }
Box{ value: 2 }.map({ x * 10 }).map({ "${x}!" }) // Box{ value: "20!" }
```

### Named arguments

Named arguments work through `.` as in the direct call:

```ant
"7".padStart(3, pad: "0") // "007"
```

They also work through a union whose members share a namespace, so `r.unwrapOr(fallback: 0)` works
on a `Result`, since `Ok` and `Err` both use `Result`:

```ant
let r: Result[int, Error] = Result.fail("no number")
r.unwrapOr(fallback: 0) // 0
```

Through a union of several namespaces, an interface, or a type parameter the call is positional,
since each implementation may name its parameters differently. Pass the argument by position:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
func Circle.area(c: Circle, scale: float64): float64 { Math.PI * c.radius * c.radius * scale }
func Rect.area(r: Rect, scale: float64): float64 { r.width * r.height * scale }
let shapes: Array[Circle | Rect] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
let big = shapes.map({ x.area(scale: 2.0) }) // error: a call through Circle | Rect is positional
```

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
func Circle.area(c: Circle, scale: float64): float64 { Math.PI * c.radius * c.radius * scale }
func Rect.area(r: Rect, scale: float64): float64 { r.width * r.height * scale }
let shapes: Array[Circle | Rect] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
shapes.map({ x.area(2.0) }) // [6.283185307179586, 12.0]
```

### Results

`Ok` and `Err` both use the `Result` namespace, so a `Result` value calls `Result` functions:

```ant
let parsed: Result[int, Error] = Result.ok(5)
let failed: Result[int, Error] = Result.fail("no number")
parsed.unwrapOr(0) // 5
failed.unwrapOr(0) // 0
```

### Sequences

A function value uses the `Iter` namespace, so a sequence chains lazily:

```ant
Iter.range(0, 100).filter({ x % 7 == 0 }).take(3).toArray() // [0, 7, 14]
```

The receiver's type picks the function: on an array `.map` is `Array.map`, which builds an array; on
a sequence it is `Iter.map`, which runs only as the sequence is read:

```ant
let seen: Array[int] = []
let doubled = Iter.fromArray([1, 2, 3]).map({
  seen.push(x)
  x * 2
})
println(seen)
println(doubled.take(1).toArray())
println(seen)
```

```text
[]
[2]
[1]
```

### Handles

A prelude type such as `File.Handle` uses its namespace, `File`:

```ant
func save(path: string, lines: Array[string]): Result[void, Error] {
  let file = try File.create(path)
  defer { file.close() }
  for line in lines {
    try file.writeLine(line)
  }
  file.close()
}
```

### Unions

A union whose members have different namespaces calls `f` on whichever member the value is. Every
member must have an `f`:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
func Circle.area(c: Circle): float64 { Math.PI * c.radius * c.radius }
func Rect.area(r: Rect): float64 { r.width * r.height }
let shapes: Array[Circle | Rect] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
shapes.map({ x.area() }) // [3.141592653589793, 6.0]
```

### Interfaces

Through an interface, `.` calls the function of the value's own type:

```ant
interface Shape {
  area(s: Self): float64
}
type Square = { side: float64 }
func Square.area(s: Square): float64 { s.side * s.side }
func total(shapes: Array[Shape]): float64 { shapes.map({ x.area() }).sum() ?? 0.0 }
total([Square{ side: 2.0 }, Square{ side: 1.0 }]) // 5.0
```

### Optional chaining

`?.` calls through `.` only when the value is not `null`:

```ant
let env = Map{ "HOME": " /home/ann " }
env.get("HOME")?.trim() // "/home/ann"
env.get("USER")?.trim() // null
```

### Module namespaces

An imported module is a namespace whose functions are fields, so `Geometry.distance(p, q)` reads a
field and calls it. A value of a type from the module calls the type's functions through `.`:

```ant [geometry.ant]
export type Point = { x: float64, y: float64 }

export func Point.norm(p: Point): float64 { Math.sqrt(p.x * p.x + p.y * p.y) }

export func distance(a: Point, b: Point): float64 {
  Point{ x: a.x - b.x, y: a.y - b.y }.norm()
}
```

```ant
import "./geometry.ant" Geometry

let p = Geometry.Point{ x: 3.0, y: 4.0 }
let q = Geometry.Point{ x: 0.0, y: 0.0 }
println(p.norm())
println(Geometry.distance(p, q))
```

```text
5.0
5.0
```

## Applications

Normalize user input in one chain:

```ant
func slug(title: string): string {
  title.toLower().fields() | String.join("-")
}
slug("  Hello   Ant World ") // "hello-ant-world"
```

Keep a type's operations next to its values:

```ant
type Money = { cents: int }
func Money.add(a: Money, b: Money): Money { Money{ cents: a.cents + b.cents } }
func Money.format(m: Money): string {
  let cents = "${m.cents % 100}".padStart(2, pad: "0")
  "${m.cents / 100}.${cents} USD"
}
let price = Money{ cents: 1999 }
price.add(Money{ cents: 6 }).format() // "20.05 USD"
```

## Properties and relations

`x.f(a)` and `T.f(x, a)` call the same function, and so does the pipe `x | T.f(a)`:

```ant
type Point = { x: float64, y: float64 }
func Point.scale(p: Point, k: float64): Point { Point{ x: p.x * k, y: p.y * k } }
let p = Point{ x: 1.0, y: 2.0 }
assert(p.scale(3.0) == Point.scale(p, 3.0))
assert(p.scale(3.0) == (p | Point.scale(3.0)))
assert(" a ".trim() == String.trim(" a "))
```

With a pipe, `a | x.f(b)` is `x.f(a, b)`: the receiver stays first and the piped value becomes the
second argument, though `a` still evaluates first:

```ant
assert((5 | "ant".padStart(".")) == "ant".padStart(5, "."))
println(5 | "ant".padStart("."))
```

```text
..ant
```

## Possible issues

### No function of that name

A misspelled name, or a function the namespace does not have, is a compile error. Arrays have no
`join`; end the chain with a pipe to `String.join` instead:

```ant
type Point = { x: float64, y: float64 }
func Point.norm(p: Point): float64 { Math.sqrt(p.x * p.x + p.y * p.y) }
let p = Point{ x: 3.0, y: 4.0 }
let n = p.nrom()             // error: Point has no field or function nrom
let s = ["a", "b"].join(",") // error: Array[string] has no field or function join
```

### Numbers have no namespace

Numbers, `bool`, and `null` have no namespace, so nothing is called through them. Call the function
by its namespace, such as `Math.abs(n)`:

```ant
let n = -5
let a = n.abs() // error: int has no field or function abs
```

### Nullable values

A `string | null` is not a `string`. Narrow it first, or use `?.`:

```ant
let env = Map{ "HOME": "/home/ann" }
let home = env.get("HOME")
let n = home.length()
// error: string | null has no field or function length on every member: narrow the value first
```

### A function that does not take the value first

A function with no parameters is called through its type's name, not through a value:

```ant
type Point = { x: float64, y: float64 }
func Point.origin(): Point { Point{ x: 0.0, y: 0.0 } }
let p = Point{ x: 3.0, y: 4.0 }
let o = p.origin() // error: Point.origin takes no argument, so it is not a function on Point
```

A namespace that mixes receivers offers through a value only the functions that take it first.
`File.write` takes a path, so a handle cannot call it; use `writeLine` or another handle function:

```ant
func save(path: string): Result[void, Error] {
  let file = try File.create(path)
  defer { file.close() }
  file.write("text") // error: File.write is not a function on File.Handle: it takes string first
}
```

A function value is a sequence only when its type is one, so `Iter` functions do not apply to any
other function:

```ant
let handler = func(n: int): string { "${n}" }
let m = handler.map({ x })
// error: Iter.map is not a function on func(int): string: it takes Seq[T] first
```

### Reading the function without calling it

`p.norm` without parentheses reads a field, and `Point` has no field `norm`. Write a lambda or the
namespace form for the function value:

```ant
type Point = { x: float64, y: float64 }
func Point.norm(p: Point): float64 { Math.sqrt(p.x * p.x + p.y * p.y) }
let p = Point{ x: 3.0, y: 4.0 }
let f = p.norm // error: Point has no field norm
```

```ant
type Point = { x: float64, y: float64 }
func Point.norm(p: Point): float64 { Math.sqrt(p.x * p.x + p.y * p.y) }
let points = [Point{ x: 3.0, y: 4.0 }, Point{ x: 0.0, y: 1.0 }]
points.map(Point.norm)   // [5.0, 1.0]
points.map({ x.norm() }) // [5.0, 1.0]
```

### Functions on a union alias

A function declared on an alias of a union is reached only through the alias's name. A value carries
its member's type, and `.` looks there:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
func Shape.area(s: Shape): float64 {
  if s is Circle { Math.PI * s.radius * s.radius } else { s.width * s.height }
}
let s: Shape = Circle{ radius: 1.0 }
let a = s.area()
// error: Shape has no field or function area on every member: narrow the value first
```

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
func Shape.area(s: Shape): float64 {
  if s is Circle { Math.PI * s.radius * s.radius } else { s.width * s.height }
}
let s: Shape = Circle{ radius: 1.0 }
Shape.area(s) // 3.141592653589793
```

### Named primitive types use their own namespace

A value of a named primitive type looks in its own type, not in its base. Convert to the base first:

```ant
type UserId = string
let id = UserId("u-42")
let n = id.length() // error: UserId has no field or function length
```

```ant
type UserId = string
let id = UserId("u-42")
string(id).length() // 4
```

### A field and a function of one name

A type cannot have both, since `x.f` could then mean either:

```ant
type Pt = { x: int, scale: int }
func Pt.scale(p: Pt): int { p.x * p.scale } // error: Pt.scale shares its name with a field of Pt
```

## See also

[Pipe](pipe.md) · [Optional chaining](optional-chaining.md) · [func](../keywords/func.md) ·
[interface](../keywords/interface.md) · [Functions](../../guide/functions.md) ·
[Idiomatic Ant](../../guide/idiomatic.md) · [Calls tour](../../tour/07-calls.md)
