<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.makeArray

```ant signature
func Reflect.makeArray(t: Reflect.ErasedType, items: Array[Reflect.Value]): Reflect.Value
```

Reflect.makeArray returns a new array of the array type t holding the values of items. It panics if
t is not an array type, or an item's type is not assignable to the element type.

<!-- end generated -->

## Details

- The result is a new array, so later changes to `items` do not change it. The elements are not
  copied: a record or array item is the same value the item boxes.
- Each item's type must be assignable to the element type. A number is not converted: box it at the
  element type with [`Reflect.makeScalar`](makeScalar.md) first.
- `t` may be an alias of an array type. Any other type, `Array[T] | null` included, panics.

## Basic examples

Build an `Array[int]` from boxed items:

```ant
let t = Reflect.erase(type(Array[int]))
let one = Reflect.valueOf(type(int), 1)
let two = Reflect.valueOf(type(int), 2)
Reflect.makeArray(t, [one, two]) // Reflect.Value([1, 2]: Array[int])
```

Take the array out with [`Reflect.to`](to.md):

```ant
let t = Reflect.erase(type(Array[string]))
let d = Reflect.makeArray(t, [Reflect.valueOf(type(string), "ant")])
Reflect.to(d, type(Array[string])) // ["ant"]
```

## Scope

With no items, the array is empty but keeps its type:

```ant
let empty = Reflect.makeArray(Reflect.erase(type(Array[int])), [])
empty               // Reflect.Value([]: Array[int])
Reflect.type(empty) // Array[int]
```

An element type that is a union takes an item of any of its members:

```ant
let t = Reflect.erase(type(Array[int | null]))
let items = [Reflect.valueOf(type(int), 1), Reflect.valueOf(type(null), null)]
Reflect.makeArray(t, items) // Reflect.Value([1, null]: Array[int | null])
```

Records are items like any other value:

```ant
type Point = { x: int, y: int }
let t = Reflect.erase(type(Array[Point]))
let p = Reflect.valueOf(type(Point), Point{ x: 1, y: 2 })
Reflect.makeArray(t, [p]) // Reflect.Value([Point{ x: 1, y: 2 }]: Array[Point])
```

An item may itself be an array built by reflection:

```ant
let inner = Reflect.erase(type(Array[int]))
let outer = Reflect.erase(type(Array[Array[int]]))
let a = Reflect.makeArray(inner, [Reflect.valueOf(type(int), 1)])
let b = Reflect.makeArray(inner, [])
Reflect.makeArray(outer, [a, b]) // Reflect.Value([[1], []]: Array[Array[int]])
```

An alias of an array type is an array type:

```ant
type Names = Array[string]
let t = Reflect.erase(type(Names))
Reflect.makeArray(t, [Reflect.valueOf(type(string), "Ann")]) // Reflect.Value(["Ann"]: Array[string])
```

Build the array type from an element type held as a type value with [`Reflect.arrayOf`](arrayOf.md):

```ant
func arrayOfOne[T](t: Reflect.Type[T], v: T): Reflect.Value {
  let at = Reflect.erase(Reflect.arrayOf(t))
  Reflect.makeArray(at, [Reflect.valueOf(t, v)])
}

arrayOfOne(type(bool), true) // Reflect.Value([true]: Array[bool])
```

## Applications

Parse a comma-separated list into an array of any number element type, checking each number against
the element type with [`Reflect.makeScalar`](makeScalar.md):

```ant
func parseList(t: Reflect.ErasedType, text: string): Reflect.Value | null {
  let s = Reflect.shape(t)
  if !(s is Reflect.Array) { return null }
  let items: Array[Reflect.Value] = []
  for part in text.split(",") {
    let n = part.trim().toInt() ?? return null
    items.push(Reflect.makeScalar(s.element, n) ?? return null)
  }
  Reflect.makeArray(t, items)
}

println(parseList(Reflect.erase(type(Array[int8])), "1, 2, 3"))
println(parseList(Reflect.erase(type(Array[int8])), "1, 300"))
println(parseList(Reflect.erase(type(Array[float64])), "1, 2"))
```

```text
Reflect.Value([1, 2, 3]: Array[int8])
null
Reflect.Value([1.0, 2.0]: Array[float64])
```

Reverse an array whose element type is known only at run time:

```ant
func reversed(d: Reflect.Value): Reflect.Value {
  Reflect.makeArray(Reflect.type(d), Reflect.items(d).reverse())
}

println(reversed(Reflect.valueOf(type(Array[string]), ["a", "b", "c"])))
println(reversed(Reflect.valueOf(type(Array[float64]), [1.5, 2.5])))
```

```text
Reflect.Value(["c", "b", "a"]: Array[string])
Reflect.Value([2.5, 1.5]: Array[float64])
```

## Properties and relations

Rebuilding an array from its own type and items gives an equal value, but a different array:

```ant
let d = Reflect.valueOf(type(Array[int]), [1, 2, 3])
let copy = Reflect.makeArray(Reflect.type(d), Reflect.items(d))
assert(copy == d)
assert(!Reflect.same(copy, d))
```

The array does not follow later changes to `items`:

```ant
let t = Reflect.erase(type(Array[int]))
let items = [Reflect.valueOf(type(int), 1)]
let d = Reflect.makeArray(t, items)
items.push(Reflect.valueOf(type(int), 2))
assert(Reflect.items(d).length() == 1)
```

The elements are not copied: a record in the array is the record the item boxed, so a change to it
shows through the array:

```ant
type P = { x: int }
let p = P{ x: 1 }
let d = Reflect.makeArray(Reflect.erase(type(Array[P])), [Reflect.valueOf(type(P), p)])
p.x = 5
assert(Reflect.to(d, type(Array[P]))[0].x == 5)
```

[`Reflect.items`](items.md) gives back the items, boxed at the element type:

```ant
let t = Reflect.erase(type(Array[string]))
let items = [Reflect.valueOf(type(string), "a"), Reflect.valueOf(type(string), "b")]
assert(Reflect.items(Reflect.makeArray(t, items)) == items)
```

## Possible issues

An item whose type is not assignable to the element type panics. An `int` is not converted to a
`float64`:

```ant
let t = Reflect.erase(type(Array[float64]))
Reflect.makeArray(t, [Reflect.valueOf(type(int), 1)])
// panic: Reflect.makeArray: an item of type int is not assignable to the element type float64 of Array[float64]
```

Box the number at the element type with [`Reflect.makeScalar`](makeScalar.md) instead, which
converts it:

```ant
let t = Reflect.erase(type(Array[float64]))
let one = Reflect.makeScalar(Reflect.erase(type(float64)), 1) ?? panic("not a float64")
Reflect.makeArray(t, [one]) // Reflect.Value([1.0]: Array[float64])
```

To keep the panic for bugs, test the contract first with [`Reflect.shape`](shape.md) and
[`Reflect.assignable`](assignable.md):

```ant
func tryMakeArray(t: Reflect.ErasedType, items: Array[Reflect.Value]): Reflect.Value | null {
  let s = Reflect.shape(t)
  if !(s is Reflect.Array) { return null }
  if !items.every({ Reflect.assignable(Reflect.type(x), s.element) }) { return null }
  Reflect.makeArray(t, items)
}

let t = Reflect.erase(type(Array[float64]))
tryMakeArray(t, [Reflect.valueOf(type(int), 1)])       // null
tryMakeArray(t, [Reflect.valueOf(type(float64), 1.5)]) // Reflect.Value([1.5]: Array[float64])
```

A type that is not an array type panics, even a set type or a nullable array type:

```ant
Reflect.makeArray(Reflect.erase(type(Set[int])), [])
// panic: Reflect.makeArray needs an array type, got Set[int]
```

```ant
Reflect.makeArray(Reflect.erase(type(Array[int] | null)), [])
// panic: Reflect.makeArray needs an array type, got Array[int] | null
```

`t` is an erased type, and the items are boxes, not bare values:

```ant
let a = Reflect.makeArray(type(Array[int]), [])
// error: cannot use Reflect.Type[Array[int]] where Reflect.ErasedType is expected
let b = Reflect.makeArray(Reflect.erase(type(Array[int])), [Reflect.valueOf(type(int), 1), 2])
// error: cannot use int where Reflect.Value is expected
```

Taking the array out at a type it does not have panics:

```ant
let d = Reflect.makeArray(Reflect.erase(type(Array[int])), [])
Reflect.to(d, type(Array[string])) // panic: Reflect.to: Array[int] is not assignable to Array[string]
```

## See also

[Reflect.items](items.md) · [Reflect.arrayOf](arrayOf.md) · [Reflect.makeSet](makeSet.md) ·
[Reflect.makeMap](makeMap.md) · [Reflect.makeScalar](makeScalar.md) · [Reflect.to](to.md) ·
[Reflect.shape](shape.md) · [Reflect.assignable](assignable.md)
