<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.Type

```ant signature
type Reflect.Type[T]
```

Reflect.Type is the type T as a value, written type(T). It is opaque: [`Reflect.erase`](../erase.md)
gives the value with its static type forgotten, and [`Reflect.valueOf`](../valueOf.md) boxes a T
under it. Two type values are == when they denote the same type, aliases expanded, and a type value
renders as the type.

<!-- end generated -->

## Details

- `type(T)` makes the value. It takes a type, never a value: there is no way to ask a value for its
  type through it.
- `T` must be closed: inside a generic function, `type(Array[T])` is a compile error. Build such a
  type from a `Reflect.Type[T]` parameter with [Reflect.arrayOf](../arrayOf.md),
  [Reflect.setOf](../setOf.md), [Reflect.mapOf](../mapOf.md), or [Reflect.nullable](../nullable.md).
- `Reflect.Type[T]` is invariant, and a union holds at most one of `Reflect.Type[...]` and
  `Reflect.ErasedType`. Erase type values with [Reflect.erase](../erase.md) to mix them, compare
  them, or take them apart.
- A type value is immutable, hashes, and can be a map key. [Json.stringify](../../Json/stringify.md)
  panics on one.

## Basic examples

`type(T)` makes a `Reflect.Type[T]`, which renders as the type:

```ant
type Config = { name: string, port: int = 8080 }
type(Config)        // Config
type(Array[Config]) // Array[Config]
type(int | null)    // int | null
```

A function generic over `T` solves `T` from a `Reflect.Type[T]` argument, so
`Reflect.to(d, type(Array[int]))` returns an `Array[int]`:

```ant
let d = Reflect.valueOf(type(Array[int]), [1, 2, 3])
let xs = Reflect.to(d, type(Array[int]))
xs.length() // 3
```

## Scope

Any closed type an annotation can hold is a type value, unless it reaches an interface. It renders
in annotation syntax, with an alias by its name and `int64` and `uint8` as `int` and `byte`:

```ant
type Ids = Array[int]
type(Ids)                     // Ids
type(int64)                   // int
type(uint8)                   // byte
type(Map[string, Array[int]]) // Map[string, Array[int]]
type({ x: int, y: int })      // { x: int, y: int }
type(func(int, string): bool) // func(int, string): bool
type(Result[int, Error])      // Result[int, Error]
type(Json.Value)              // Json.Value
type(Bytes)                   // Bytes
type(null)                    // null
type(void)                    // void
type(never)                   // never
```

A recursive type needs no value of itself:

```ant
type Tree = { value: int, children: Array[Tree] }
type(Tree) // Tree
```

A type declared in a block is a type value too:

```ant
func pointType(): Reflect.ErasedType {
  type Point = { x: int, y: int }
  Reflect.erase(type(Point))
}

println(pointType())
```

```text
Point
```

A generic function holds a `Reflect.Type[T]` and builds larger types from it with the builders,
which keep the static type:

```ant
func listOf[T](t: Reflect.Type[T]): Reflect.Type[Array[T] | null] {
  t | Reflect.arrayOf() | Reflect.nullable()
}

println(listOf(type(string)))
```

```text
Array[string] | null
```

Call the functions that take a type value through `.`:

```ant
let t = type(Map[string, int])
t.erase()            // Map[string, int]
type(int).arrayOf()  // Array[int]
type(string).setOf() // Set[string]
```

## Applications

Read a setting at the type the caller asks for, and give `null` when the stored value does not fit
it:

```ant
func setting[T](settings: Map[string, Reflect.Value], key: string, t: Reflect.Type[T]): T | null {
  let d = settings.get(key)
  if d == null || !Reflect.assignable(Reflect.type(d), t.erase()) {
    return null
  }
  Reflect.to(d, t)
}

let settings = Map{
  "port": Reflect.valueOf(type(int), 8080),
  "host": Reflect.valueOf(type(string), "localhost"),
}
println(setting(settings, "port", type(int)))
println(setting(settings, "port", type(string)))
println(setting(settings, "host", type(string | null)))
println(setting(settings, "debug", type(bool)))
```

```text
8080
null
localhost
null
```

Keep one service per type, keyed by the erased type, and hand each back at its static type:

```ant
type Logger = { prefix: string }
type Clock = { now: func(): string }
type Services = Map[Reflect.ErasedType, Reflect.Value]

func provide[T](services: Services, t: Reflect.Type[T], service: T): void {
  services[t.erase()] = Reflect.valueOf(t, service)
}

func need[T](services: Services, t: Reflect.Type[T]): T {
  let d = services.get(t.erase())
  if d == null {
    panic("no service of type ${t}")
  }
  Reflect.to(d, t)
}

let services: Services = Map{}
provide(services, type(Logger), Logger{ prefix: "[app]" })
provide(services, type(Clock), Clock{ now: { "12:00" } })
let log = need(services, type(Logger))
println("${log.prefix} started at ${need(services, type(Clock)).now()}")
```

```text
[app] started at 12:00
```

## Properties and relations

Two type values are `==` when they denote the same type: aliases expand, `int` is `int64`, and union
members compare as a set:

```ant
type Ids = Array[int]
assert(type(Ids) == type(Array[int]))
assert(type(int) == type(int64))
assert(type(int | string) == type(string | int))
```

Equal type values hash alike, so a set keeps one of them:

```ant
Set.of([type(int), type(int64)]).size() // 1
```

`String.from` and interpolation give the rendering, which keeps an alias's name even where `==`
expands it:

```ant
type Ids = Array[int]
assert(String.from(type(Ids)) == "Ids")
assert("${type(Array[int])}" == "Array[int]")
```

## Possible issues

`type(...)` takes a type, not a value. Unlike `typeof` in other languages, it does not ask what a
value is:

```ant
let config = 1
println(type(config)) // error: type(...) takes a type, not a value; config is a value
```

A type name alone is not a value. Write `type(Config)`:

```ant
type Config = { name: string }
println(Config) // error: Config is a type, not a value; a type appears only as a literal head, after is, or before a dot
```

A generic type needs its arguments:

```ant
type Box[T] = { value: T }
println(type(Box)) // error: Box takes 1 type argument, got 0
```

Inside a generic function, `type(...)` cannot mention a type parameter. Take a `Reflect.Type[T]` and
use a builder such as [Reflect.arrayOf](../arrayOf.md):

```ant
func listOf[T](t: Reflect.Type[T]): Reflect.ErasedType {
  Reflect.erase(type(Array[T]))
  // error: type(...) cannot mention T, a type parameter of the enclosing function or type, since nothing is bound to it at run time; take a Reflect.Type[T] parameter and build the type from it with Reflect.arrayOf, Reflect.setOf, Reflect.mapOf, or Reflect.nullable
}
```

`Reflect.Type` is invariant: a `Reflect.Type[int]` is not a `Reflect.Type[int | null]`. Use
[Reflect.nullable](../nullable.md) to get one:

```ant
let t: Reflect.Type[int | null] = type(int)
// error: cannot use Reflect.Type[int] where Reflect.Type[int | null] is expected
```

Type values of two different types cannot be compared with `==`, even to learn that they differ.
Erase both sides:

```ant
println(type(int) == type(string))
// error: Reflect.Type[int] and Reflect.Type[string] cannot be compared with ==: neither is assignable to the other
```

```ant
Reflect.erase(type(int)) == Reflect.erase(type(string)) // false
```

An `if` cannot pick between two type values at run time, since `Reflect.Type[int64]` and
`Reflect.Type[int32]` are two types that no union may hold:

```ant
let wide = true
let t = if wide { type(int64) } else { type(int32) }
// error: the branches of this if do not form a valid type: a union may hold at most one type with no discriminator, and Reflect.Type[int] and Reflect.Type[int32] both have none; narrow first or give the whole an expected type
```

Erase both branches, and the `if` gives a `Reflect.ErasedType`:

```ant
let wide = true
let t = if wide { Reflect.erase(type(int64)) } else { Reflect.erase(type(int32)) }
t // int
```

A type value has no JSON form:

```ant
Json.stringify(type(int)) // panic: Json.stringify: a type has no JSON form
```

`type(...)` cannot reach an interface, directly or through a field:

```ant
interface Named {
  name(self: Self): string
}
type Holder = { item: Named }
println(type(Array[Holder]))
// error: type(...) cannot reach an interface: Named is one, and an interface has no run-time identity
```

## See also

[Reflect](../index.md) · [Reflect.ErasedType](ErasedType.md) · [Reflect.erase](../erase.md) ·
[Reflect.valueOf](../valueOf.md) · [Reflect.to](../to.md) · [Reflect.arrayOf](../arrayOf.md) ·
[Reflect.nullable](../nullable.md) · [Reflect.shape](../shape.md)
