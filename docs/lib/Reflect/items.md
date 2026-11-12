<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.items

```ant signature
func Reflect.items(d: Reflect.Value): Array[Reflect.Value]
```

Reflect.items returns an array's elements or a set's members, each boxed at the element type, in a
new array read when it is called. It panics if d's type is neither.

<!-- end generated -->

## Details

- Each box has the type the container's type gives its elements. When that type is a union, the box
  takes the member the element belongs to, so `Reflect.type` of an item is never a union.
- The result is a new array, read once at the call. Adding to the container afterward does not
  change it, but the boxes refer to the original elements, not copies.
- It panics on any other type, including a map, `string`, and `Bytes`. Use
  [`Reflect.entries`](entries.md) for a map.

## Basic examples

Box each element of an array:

```ant
let d = Reflect.valueOf(type(Array[int]), [1, 2, 3])
Reflect.items(d) // [Reflect.Value(1: int), Reflect.Value(2: int), Reflect.Value(3: int)]
```

Box each member of a set:

```ant
let d = Reflect.valueOf(type(Set[string]), Set.of(["a", "b"]))
Reflect.items(d) // [Reflect.Value("a": string), Reflect.Value("b": string)]
```

Take the values back out with [`Reflect.to`](to.md):

```ant
let d = Reflect.valueOf(type(Array[int]), [3, 1, 2])
Reflect.items(d) | Array.map({ Reflect.to(x, type(int)) }) // [3, 1, 2]
```

## Scope

An empty array gives an empty array:

```ant
Reflect.valueOf(type(Array[string]), []).items() // []
```

Duplicates in a set collapse before it is read:

```ant
Reflect.valueOf(type(Set[int]), Set.of([2, 1, 2])).items() // [Reflect.Value(2: int), Reflect.Value(1: int)]
```

The elements of a nested array are boxed arrays, which `Reflect.items` reads in turn:

```ant
let d = Reflect.valueOf(type(Array[Array[int]]), [[1], [2, 3]])
d.items()                                     // [Reflect.Value([1]: Array[int]), Reflect.Value([2, 3]: Array[int])]
d.items() | Array.map({ x.items().length() }) // [1, 2]
```

An element type that is a union resolves to each element's own member:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
let shapes: Array[Circle | Square] = [Circle{ r: 1.0 }, Square{ side: 2.0 }]
let d = Reflect.valueOf(type(Array[Circle | Square]), shapes)
d.items() | Array.map(Reflect.type) // [Circle, Square]
```

A `null` element of an `Array[T | null]` is boxed at `null`:

```ant
let d = Reflect.valueOf(type(Array[int | null]), [1, null])
d.items() | Array.map(Reflect.type) // [int, null]
```

Elements of a named primitive type keep their tag:

```ant
type Celsius = float64
let d = Reflect.valueOf(type(Array[Celsius]), [Celsius(21.5)])
d.items() // [Reflect.Value(Celsius(21.5): Celsius)]
```

A box made at a nullable array type holds the array itself, so it can be read:

```ant
let xs: Array[int] | null = [4]
Reflect.valueOf(type(Array[int] | null), xs).items() // [Reflect.Value(4: int)]
```

## Applications

Add up the `int` elements of an array or a set, read through [`Reflect.scalar`](scalar.md):

```ant
func total(d: Reflect.Value): int {
  d.items() | Array.reduce(0, func(sum: int, item: Reflect.Value): int {
    let v = item.scalar()
    if v is int { sum + v } else { sum }
  })
}

println(total(Reflect.valueOf(type(Array[int]), [1, 2, 3])))
println(total(Reflect.valueOf(type(Set[int]), Set.of([10, 20]))))
```

```text
6
30
```

Render any array, set, or scalar, recursing into containers by their shape:

```ant
func render(d: Reflect.Value): string {
  let s = Reflect.shape(Reflect.type(d))
  if s is Reflect.Array || s is Reflect.Set {
    return "<" + (d.items() | Array.map(render) | String.join(", ")) + ">"
  }
  String.from(d.scalar())
}

let nested = [Set.of(["a", "b"]), Set.of([])]
println(render(Reflect.valueOf(type(Array[Set[string]]), nested)))
```

```text
<<a, b>, <>>
```

Walk a structure that refers to itself, keeping the boxes on the path and testing each child against
them with [`Reflect.same`](same.md):

```ant
type Node = { name: string, children: Array[Node] }

func names(d: Reflect.Value, path: Array[Reflect.Value]): Array[string] {
  if path.some({ Reflect.same(x, d) }) {
    return ["..."]
  }
  let inner = [...path, d]
  let name = d.field("name").to(type(string))
  let kids = d.field("children").items()
  [name, ...kids | Array.flatMap({ names(x, inner) })]
}

let root = Node{ name: "root", children: [] }
let leaf = Node{ name: "leaf", children: [] }
root.children.push(leaf)
leaf.children.push(root)
println(names(Reflect.valueOf(type(Node), root), []))
```

```text
["root", "leaf", "..."]
```

## Properties and relations

The result has one box per element, and unboxing them gives the array back:

```ant
let xs = [3, 1, 2]
let d = Reflect.valueOf(type(Array[int]), xs)
assert(d.items().length() == xs.length())
assert((d.items() | Array.map({ Reflect.to(x, type(int)) })) == xs)
```

[`Reflect.makeArray`](makeArray.md) is its inverse, and builds a new array equal to the original:

```ant
let d = Reflect.valueOf(type(Array[int]), [3, 1, 2])
let copy = Reflect.makeArray(Reflect.type(d), d.items())
assert(copy == d)
assert(!Reflect.same(copy, d))
```

For a set, the items are the members in the order [`Set.values`](../Set/values.md) gives:

```ant
let s = Set.of(["b", "a", "c"])
let d = Reflect.valueOf(type(Set[string]), s)
assert((d.items() | Array.map({ Reflect.to(x, type(string)) })) == s.values())
```

A record element's box refers to the original record, so [`Reflect.same`](same.md) reports it:

```ant
type Point = { x: int, y: int }
let ps = [Point{ x: 1, y: 2 }]
let first = Reflect.valueOf(type(Array[Point]), ps).items()[0]
assert(Reflect.same(first, Reflect.valueOf(type(Point), ps[0])))
```

## Possible issues

The result is read when `Reflect.items` is called. An element pushed later is not in it:

```ant
let xs = [1, 2]
let d = Reflect.valueOf(type(Array[int]), xs)
let before = d.items()
xs.push(3)
before.length()    // 2
d.items().length() // 3
```

A record taken out of a box is the original record, so changing it changes the array:

```ant
type Point = { x: int, y: int }
let ps = [Point{ x: 1, y: 2 }]
let p = Reflect.to(Reflect.valueOf(type(Array[Point]), ps).items()[0], type(Point))
p.x = 9
ps // [Point{ x: 9, y: 2 }]
```

A map is not an array or a set:

```ant
Reflect.items(Reflect.valueOf(type(Map[string, int]), Map{ a: 1 }))
// panic: Reflect.items needs a value of an array or set type, got one of type Map[string, int]
```

A string is not one either; reflection treats it as a leaf:

```ant
Reflect.items(Reflect.valueOf(type(string), "abc"))
// panic: Reflect.items needs a value of an array or set type, got one of type string
```

A box made at a nullable array type holds `null` when the value is `null`, and `Reflect.items`
panics on it:

```ant
let xs: Array[int] | null = null
Reflect.valueOf(type(Array[int] | null), xs).items()
// panic: Reflect.items needs a value of an array or set type, got one of type null
```

Test the shape first, as `count` does below, when the type is not known:

```ant
func count(d: Reflect.Value): int | null {
  let s = Reflect.shape(Reflect.type(d))
  if !(s is Reflect.Array || s is Reflect.Set) {
    return null
  }
  d.items().length()
}

count(Reflect.valueOf(type(Array[int]), [1, 2])) // 2
count(Reflect.valueOf(type(string), "abc"))      // null
```

## See also

[Reflect.entries](entries.md) · [Reflect.field](field.md) · [Reflect.payload](payload.md) ·
[Reflect.scalar](scalar.md) · [Reflect.makeArray](makeArray.md) · [Reflect.makeSet](makeSet.md) ·
[Reflect.same](same.md) · [Reflect.shape](shape.md) · [Reflect.to](to.md) ·
[Reflect.valueOf](valueOf.md) · [Set.values](../Set/values.md)
