<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.makeSet

```ant signature
func Reflect.makeSet(t: Reflect.ErasedType, members: Array[Reflect.Value]): Reflect.Value
```

Reflect.makeSet returns a new set of the set type t holding the values of members, duplicates
collapsed. It panics if t is not a set type, or a member's type is not assignable to the member
type.

<!-- end generated -->

## Details

- Duplicates collapse as in `Set.of`: the set keeps the first of each `==` member, in the order of
  `members`.
- Each member's type must be assignable to the member type, so an `int` box goes into a
  `Set[int | null]`.

## Basic examples

Box the members, then build a `Set[int]` from them:

```ant
let t = Reflect.erase(type(Set[int]))
let one = Reflect.valueOf(type(int), 1)
let two = Reflect.valueOf(type(int), 2)
Reflect.makeSet(t, [one, two, one]) // Reflect.Value(Set.of([1, 2]): Set[int])
```

Take the set out with `Reflect.to`:

```ant
let t = Reflect.erase(type(Set[string]))
let members = ["b", "a", "b"].map({ Reflect.valueOf(type(string), x) })
Reflect.to(Reflect.makeSet(t, members), type(Set[string])) // Set.of(["b", "a"])
```

## Scope

No members give an empty set of the type `t`:

```ant
Reflect.makeSet(Reflect.erase(type(Set[string])), []) // Reflect.Value(Set.of([]): Set[string])
```

The set type may come from `Reflect.setOf`, for a member type held as a type value:

```ant
let member = type(float64)
let t = Reflect.erase(Reflect.setOf(member))
Reflect.makeSet(t, [Reflect.valueOf(member, 1.5)]) // Reflect.Value(Set.of([1.5]): Set[float64])
```

A union member type takes a box of any of its members. Each box keeps its own type:

```ant
let t = Reflect.erase(type(Set[int | null]))
let members = [Reflect.valueOf(type(int), 7), Reflect.valueOf(type(null), null)]
let d = Reflect.makeSet(t, members)
d                                   // Reflect.Value(Set.of([7, null]): Set[int | null])
d.items() | Array.map(Reflect.type) // [int, null]
```

Records collapse by contents, as they do in any set:

```ant
type Point = { x: int, y: int }
let t = Reflect.erase(type(Set[Point]))
let a = Reflect.valueOf(type(Point), Point{ x: 1, y: 2 })
let b = Reflect.valueOf(type(Point), Point{ x: 1, y: 2 })
Reflect.makeSet(t, [a, b]) // Reflect.Value(Set.of([Point{ x: 1, y: 2 }]): Set[Point])
```

A named primitive member type needs boxes of that type, which `Reflect.makeTagged` makes:

```ant
type Tag = string
let members = ["new", "sale"].map({
  Reflect.makeTagged(Reflect.erase(type(Tag)), Reflect.valueOf(type(string), x))
})
Reflect.makeSet(Reflect.erase(type(Set[Tag])), members)
// Reflect.Value(Set.of([Tag("new"), Tag("sale")]): Set[Tag])
```

## Applications

Decode a list of leaves into whatever set type a slot asks for. `Reflect.makeScalar` checks and
converts each leaf, so the decoder needs no branch per member type:

```ant
func decodeSet(t: Reflect.ErasedType, raw: Array[Reflect.Scalar]): Reflect.Value | null {
  let s = t.shape()
  if !(s is Reflect.Set) {
    return null
  }
  let members: Array[Reflect.Value] = []
  for v in raw {
    let m = Reflect.makeScalar(s.member, v)
    if m == null {
      return null
    }
    members.push(m)
  }
  Reflect.makeSet(t, members)
}

let ports = Reflect.erase(type(Set[uint16]))
println(decodeSet(ports, [80, 443, 80]))
println(decodeSet(ports, [80, -1]))
println(decodeSet(Reflect.erase(type(Set[string])), ["ant", "go"]))
```

```text
Reflect.Value(Set.of([80, 443]): Set[uint16])
null
Reflect.Value(Set.of(["ant", "go"]): Set[string])
```

Trim every string in a value of any array or set type. Rebuilding a set collapses the strings that
become equal:

```ant
func trimAll(d: Reflect.Value): Reflect.Value {
  let t = d.type()
  let s = t.shape()
  if s is Reflect.String {
    return Reflect.valueOf(type(string), d.to(type(string)).trim())
  }
  if s is Reflect.Array {
    return Reflect.makeArray(t, d.items().map(trimAll))
  }
  if s is Reflect.Set {
    return Reflect.makeSet(t, d.items().map(trimAll))
  }
  d
}

let groups = [Set.of([" ant", "ant ", "go"]), Set.of(["go "])]
let clean = trimAll(Reflect.valueOf(type(Array[Set[string]]), groups))
println(Reflect.to(clean, type(Array[Set[string]])))
```

```text
[Set.of(["ant", "go"]), Set.of(["go"])]
```

## Properties and relations

`Reflect.makeSet` over the boxed values builds the set `Set.of` builds from the values:

```ant
let xs = [3, 1, 3, 2, 1]
let d = Reflect.makeSet(Reflect.erase(type(Set[int])), xs.map({ Reflect.valueOf(type(int), x) }))
assert(d.to(type(Set[int])) == Set.of(xs))
```

Rebuilding a set from its own type and `Reflect.items` gives an equal set, but a new one:

```ant
let d = Reflect.valueOf(type(Set[string]), Set.of(["a", "b"]))
let copy = Reflect.makeSet(d.type(), d.items())
assert(copy == d)
assert(!Reflect.same(copy, d))
```

## Possible issues

A type that is not a set type panics, even when the members would fit:

```ant
let one = Reflect.valueOf(type(int), 1)
Reflect.makeSet(Reflect.erase(type(Array[int])), [one])
// panic: Reflect.makeSet needs a set type, got Array[int]
```

A member whose type is not assignable to the member type panics. Test it first with
`Reflect.assignable` when the input may not fit:

```ant
let t = Reflect.erase(type(Set[int]))
let s = Reflect.valueOf(type(string), "1")
Reflect.makeSet(t, [s])
// panic: Reflect.makeSet: a member of type string is not assignable to the member type int of Set[int]
Reflect.assignable(s.type(), Reflect.erase(type(int))) // false
```

A named primitive is not its base type, so a plain `string` box does not go into a `Set[Tag]`. Box
it with `Reflect.makeTagged`:

```ant
type Tag = string
let t = Reflect.erase(type(Set[Tag]))
Reflect.makeSet(t, [Reflect.valueOf(type(string), "new")])
// panic: Reflect.makeSet: a member of type string is not assignable to the member type Tag of Set[Tag]
```

`NaN` is not `==` to itself, so `NaN` members never collapse, as in `Set.of`:

```ant
let members = [Math.NaN, Math.NaN].map({ Reflect.valueOf(type(float64), x) })
Reflect.makeSet(Reflect.erase(type(Set[float64])), members)
// Reflect.Value(Set.of([NaN, NaN]): Set[float64])
```

## See also

[Reflect.setOf](setOf.md) · [Reflect.items](items.md) · [Reflect.makeArray](makeArray.md) ·
[Reflect.makeMap](makeMap.md) · [Reflect.makeScalar](makeScalar.md) ·
[Reflect.makeTagged](makeTagged.md) · [Reflect.assignable](assignable.md) · [Set.of](../Set/of.md)
