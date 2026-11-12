<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.assignable

```ant signature
func Reflect.assignable(from: Reflect.ErasedType, to: Reflect.ErasedType): bool
```

Reflect.assignable reports whether a value of type from may be used where to is expected, on two
closed types. It is the relation the readers, the builders, and [`Reflect.to`](to.md) use, exposed
so that a library can test before calling a function that would panic.

<!-- end generated -->

## Details

- The first argument is the type you have and the second is the slot it goes into. Swapping them
  asks the opposite question.
- It applies the type checker's own rules, so it is `true` exactly where the checker accepts a
  variable of type `from` where `to` is expected.

## Basic examples

An `int` fits where `int | null` is expected, but not the other way around:

```ant
let n = Reflect.erase(type(int))
let maybe = Reflect.erase(type(int | null))
Reflect.assignable(n, maybe) // true
Reflect.assignable(maybe, n) // false
```

Every type is assignable to itself, and `int` is another name for `int64`:

```ant
Reflect.assignable(Reflect.erase(type(string)), Reflect.erase(type(string))) // true
Reflect.assignable(Reflect.erase(type(int)), Reflect.erase(type(int64)))     // true
```

## Scope

### Primitives

Numbers never widen, and a named primitive type is not its base:

```ant
type UserId = int
let n = Reflect.erase(type(int))
Reflect.assignable(Reflect.erase(type(int8)), n)   // false
Reflect.assignable(Reflect.erase(type(UserId)), n) // false
```

`never` fits everywhere. `null` fits only a type that includes it, and `void` only `void`:

```ant
let text = Reflect.erase(type(string))
let empty = Reflect.erase(type(null))
Reflect.assignable(Reflect.erase(type(never)), text)          // true
Reflect.assignable(empty, text)                               // false
Reflect.assignable(empty, Reflect.erase(type(string | null))) // true
Reflect.assignable(empty, Reflect.erase(type(void)))          // false
```

### Unions

A member fits its union. A union fits a type only when every one of its members does:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
let shape = Reflect.erase(type(Shape))
let maybeCircle = Reflect.erase(type(Circle | null))
Reflect.assignable(Reflect.erase(type(Circle)), shape)             // true
Reflect.assignable(maybeCircle, shape)                             // false
Reflect.assignable(maybeCircle, Reflect.erase(type(Shape | null))) // true
```

### Records

A named record type accepts only its own tag. An anonymous record type accepts any record with at
least its fields, each with exactly that field's type:

```ant
type User = { name: string, age: int }
type Pet = { name: string }
let user = Reflect.erase(type(User))
Reflect.assignable(user, Reflect.erase(type({ name: string })))        // true
Reflect.assignable(user, Reflect.erase(type({ name: string | null }))) // false
Reflect.assignable(user, Reflect.erase(type(Pet)))                     // false
```

### Generic types

Type arguments must be identical, even when one would fit the other:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
let circles = Reflect.erase(type(Array[Circle]))
Reflect.assignable(circles, Reflect.erase(type(Array[Shape])))  // false
Reflect.assignable(circles, Reflect.erase(type(Array[Circle]))) // true
Reflect.assignable(
  Reflect.erase(type(Result[int, Error])),
  Reflect.erase(type(Result[int | null, Error])),
) // false
```

### Function types

A function that takes a wider parameter fits a slot that passes a narrower one, and not the reverse:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
let anyShape = Reflect.erase(type(func(Shape): float64))
let onlyCircles = Reflect.erase(type(func(Circle): float64))
Reflect.assignable(anyShape, onlyCircles) // true
Reflect.assignable(onlyCircles, anyShape) // false
```

### Recursive aliases

A recursive alias is related by its structure, at any depth:

```ant
type F = func(): F | null
type G = func(): G | int | null
Reflect.assignable(Reflect.erase(type(F)), Reflect.erase(type(G))) // true
Reflect.assignable(Reflect.erase(type(G)), Reflect.erase(type(F))) // false
```

### Types from values and shapes

The type of a box comes from `Reflect.type`, and a child type comes from `Reflect.shape`. Both are
erased already:

```ant
let box = Reflect.valueOf(type(int | null), 42)
let s = Reflect.shape(Reflect.erase(type(Array[int | null])))
Reflect.type(box)                                                      // int
s is Reflect.Array && Reflect.assignable(Reflect.type(box), s.element) // true
```

## Applications

### Unbox without a panic

`Reflect.to` panics when the box does not fit. Test first to get `null` instead:

```ant
func cast[T](box: Reflect.Value, t: Reflect.Type[T]): T | null {
  if !Reflect.assignable(Reflect.type(box), Reflect.erase(t)) { return null }
  Reflect.to(box, t)
}

let box = Reflect.valueOf(type(string), "ant")
println(cast(box, type(string | null)))
println(cast(box, type(int)))
```

```text
ant
null
```

### Keep only the items that fit

`Reflect.makeArray` panics on an item that does not fit the element type. Filter the items first:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

let t = Reflect.erase(type(Array[Shape]))
let s = Reflect.shape(t)
if !(s is Reflect.Array) { return }
let items = [
  Reflect.valueOf(type(Circle), Circle{ radius: 1.0 }),
  Reflect.valueOf(type(string), "not a shape"),
  Reflect.valueOf(type(Rect), Rect{ width: 2.0, height: 3.0 }),
]
let shapes = items.filter({ Reflect.assignable(Reflect.type(x), s.element) })
println(Reflect.makeArray(t, shapes))
```

```text
Reflect.Value([Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]: Array[Shape])
```

### Pick a handler by type

Look up the first entry of a table whose type accepts the value:

```ant
type Handler = { accepts: Reflect.ErasedType, name: string }

let handlers = [
  Handler{ accepts: Reflect.erase(type(int | float64)), name: "number" },
  Handler{ accepts: Reflect.erase(type(string)), name: "text" },
  Handler{ accepts: Reflect.erase(type(null)), name: "empty" },
]

func handlerFor(box: Reflect.Value): string {
  let t = Reflect.type(box)
  handlers.find({ Reflect.assignable(t, x.accepts) })?.name ?? "unknown"
}

println(handlerFor(Reflect.valueOf(type(float64), 2.5)))
println(handlerFor(Reflect.valueOf(type(string | null), null)))
println(handlerFor(Reflect.valueOf(type(bool), true)))
```

```text
number
empty
unknown
```

## Properties and relations

`Reflect.to` succeeds exactly when `Reflect.assignable` holds for the box's type:

```ant
let box = Reflect.valueOf(type(int), 7)
assert(Reflect.assignable(Reflect.type(box), Reflect.erase(type(int | string))))
let v: int | string = Reflect.to(box, type(int | string))
assert(v == 7)
```

Every type fits itself, `never` fits every type, and a type fits its nullable form:

```ant
let types = [
  Reflect.erase(type(int)),
  Reflect.erase(type(Array[string])),
  Reflect.erase(type(func(int): bool)),
  Reflect.erase(type({ id: int })),
]
let bottom = Reflect.erase(type(never))
for t in types {
  assert(Reflect.assignable(t, t))
  assert(Reflect.assignable(bottom, t))
}
let n = type(Map[string, int])
assert(Reflect.assignable(Reflect.erase(n), Reflect.erase(Reflect.nullable(n))))
```

## Possible issues

A `Reflect.Type[T]` is not an erased type. Passing `type(T)` directly is a compile error:

```ant
Reflect.assignable(type(int), Reflect.erase(type(int | null)))
// error: cannot use Reflect.Type[int] where Reflect.ErasedType is expected
```

`type(T)` cannot reach an interface, so `Reflect.assignable` cannot test whether a type satisfies
one:

```ant
interface Show {
  show(x: Self): string
}
let t = Reflect.erase(type(Show))
// error: type(...) cannot reach an interface: Show is one, and an interface has no run-time identity
```

A box holding a union value has the type of the member the value belongs to, so the test is on that
member, not on the union the box was made with:

```ant
let box = Reflect.valueOf(type(int | string), 1)
Reflect.assignable(Reflect.type(box), Reflect.erase(type(int))) // true
```

## See also

[Reflect.to](to.md), [Reflect.type](type.md), [Reflect.shape](shape.md), [Reflect.erase](erase.md),
[Reflect.makeArray](makeArray.md), [Reflect.makeRecord](makeRecord.md), [Reflect.same](same.md)
