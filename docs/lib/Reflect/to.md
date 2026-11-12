<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.to

```ant signature
func Reflect.to[T](d: Reflect.Value, t: Reflect.Type[T]): T
```

Reflect.to returns the value in d as a T. It panics if d's type is not assignable to T, which is the
check an assignment would have made statically.

<!-- end generated -->

## Details

- A box never holds a union: it holds the member its value belongs to. Unbox at any type that member
  fits, including the union it was boxed at.
- It returns the stored value itself, not a copy. An array or record taken out is the one that was
  boxed, and changes to it show through every reference.
- Use `Reflect.assignable` to test before calling when a mismatch is expected input rather than a
  bug.

## Basic examples

Get the value back at the type it was boxed at:

```ant
let box = Reflect.valueOf(type(int), 42)
Reflect.to(box, type(int)) + 1 // 43
```

Any type the stored type is assignable to works too:

```ant
let box = Reflect.valueOf(type(int), 42)
Reflect.to(box, type(int | null))   // 42
Reflect.to(box, type(int | string)) // 42
```

## Scope

### Unions and null

A box made at a union holds the member, so it unboxes at that member or at the union:

```ant
let box = Reflect.valueOf(type(int | null), 42)
Reflect.to(box, type(int))        // 42
Reflect.to(box, type(int | null)) // 42
let empty = Reflect.valueOf(type(string | null), null)
Reflect.to(empty, type(null))          // null
Reflect.to(empty, type(string | null)) // null
```

### Records

A named record unboxes at its own type, or at an anonymous record type with some of its fields. The
value keeps its tag:

```ant
type User = { name: string, age: int }
let box = Reflect.valueOf(type(User), User{ name: "Ann", age: 31 })
Reflect.to(box, type(User)).age              // 31
Reflect.to(box, type({ name: string })).name // "Ann"
Reflect.to(box, type({ name: string }))      // User{ name: "Ann", age: 31 }
```

### Named primitive types

```ant
type UserId = int
let box = Reflect.valueOf(type(UserId), UserId(7))
Reflect.to(box, type(UserId)) // UserId(7)
```

### Function values

A function boxed at a function type comes back callable:

```ant
let box = Reflect.valueOf(type(func(int): int), { x * 2 })
let double = Reflect.to(box, type(func(int): int))
double(21) // 42
```

### Calls through `.` and pipes

The box is the first argument, so `.` and pipes reach it:

```ant
let box = Reflect.valueOf(type(string), "ant")
box.to(type(string))           // "ant"
box | Reflect.to(type(string)) // "ant"
```

## Applications

### A typed getter over a table of boxes

Store values of different types in one map, and read each back at the type the caller names.
`Reflect.assignable` turns a mismatch into `null`:

```ant
let settings: Map[string, Reflect.Value] = Map{
  "port": Reflect.valueOf(type(int), 8080),
  "host": Reflect.valueOf(type(string), "localhost"),
}

func setting[T](key: string, t: Reflect.Type[T]): T | null {
  let box = settings.get(key)
  if box == null { return null }
  if !Reflect.assignable(Reflect.type(box), Reflect.erase(t)) { return null }
  Reflect.to(box, t)
}

let port = setting("port", type(int)) ?? 80
println(port + 1)
println(setting("host", type(string)))
println(setting("host", type(int)))
println(setting("debug", type(bool)))
```

```text
8081
localhost
null
null
```

### Build at run time, return a typed result

A generic function works on erased types and boxes in between, and `Reflect.to` gives its caller a
static `Array[T]` at the end. The `is` test narrows `s` to `Reflect.Array` so `s.element`
type-checks; `Reflect.arrayOf` always gives an array type, so that `return null` never runs:

```ant
func convert[T](xs: Array[float64], t: Reflect.Type[T]): Array[T] | null {
  let at = Reflect.arrayOf(t)
  let s = Reflect.shape(Reflect.erase(at))
  if !(s is Reflect.Array) { return null }
  let items: Array[Reflect.Value] = []
  for x in xs {
    let item = Reflect.makeScalar(s.element, x)
    if item == null { return null }
    items.push(item)
  }
  Reflect.to(Reflect.makeArray(Reflect.erase(at), items), at)
}

println(convert([1.0, 2.0, 3.0], type(int8)))
println(convert([1.0, 2.5], type(int8)))
println(convert([0.5], type(float32)))
```

```text
[1, 2, 3]
null
[0.5]
```

## Properties and relations

`Reflect.to` undoes `Reflect.valueOf` at the same type:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
assert(Reflect.to(Reflect.valueOf(type(Point), p), type(Point)) == p)
assert(Reflect.to(Reflect.valueOf(type(string), "a"), type(string)) == "a")
```

It returns the boxed value itself, so a mutation through the result shows in the original:

```ant
let xs = [1, 2]
let ys = Reflect.to(Reflect.valueOf(type(Array[int]), xs), type(Array[int]))
ys.push(3)
assert(Array.same(xs, ys))
assert(xs == [1, 2, 3])
```

It succeeds exactly when `Reflect.assignable` holds between the box's type and the target:

```ant
let box = Reflect.valueOf(type(int), 7)
let target = type(int | string)
assert(Reflect.assignable(Reflect.type(box), Reflect.erase(target)))
assert(Reflect.to(box, target) == 7)
```

## Possible issues

Numbers never widen. A box made at `int` does not unbox as `float64`:

```ant
let box = Reflect.valueOf(type(int), 1)
Reflect.to(box, type(float64)) // panic: Reflect.to: int is not assignable to float64
```

A named primitive type is not its base type:

```ant
type UserId = int
let box = Reflect.valueOf(type(UserId), UserId(1))
Reflect.to(box, type(int)) // panic: Reflect.to: UserId is not assignable to int
```

Type arguments must match exactly. An `Array[int]` is not an `Array[int | null]`:

```ant
let box = Reflect.valueOf(type(Array[int]), [1])
Reflect.to(box, type(Array[int | null]))
// panic: Reflect.to: Array[int] is not assignable to Array[int | null]
```

A record boxed at an anonymous record type keeps that type, so it does not unbox at its own named
type:

```ant
type User = { name: string, age: int }
let box = Reflect.valueOf(type({ name: string }), User{ name: "Ann", age: 31 })
Reflect.to(box, type(User)) // panic: Reflect.to: { name: string } is not assignable to User
```

The type argument fixes the static result type, so a mismatch with the variable's annotation is a
compile error:

```ant
let box = Reflect.valueOf(type(int), 42)
let n: string = Reflect.to(box, type(int))
// error: cannot use Reflect.Type[int] where Reflect.Type[string] is expected
```

## See also

[Reflect.valueOf](valueOf.md) · [Reflect.assignable](assignable.md) · [Reflect.type](type.md) ·
[Reflect.erase](erase.md) · [Reflect.makeScalar](makeScalar.md) · [Reflect.makeArray](makeArray.md)
· [Reflect.Value](types/Value.md)
