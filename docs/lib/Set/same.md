<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.same

```ant signature
func Set.same[T](a: Set[T], b: Set[T]): bool
```

Set.same reports whether a and b are the same set, not merely equal ones.

<!-- end generated -->

## Details

- `==` compares members in any order, so use `Set.same` to tell a shared set from a copy.
- `Set.of`, `Set.union`, `Set.intersect`, and `Set.difference` always build a new set.

## Basic examples

A set is the same as itself, and as any other name for it:

```ant
let a = Set.of([1, 2])
let b = a
Set.same(a, a) // true
Set.same(a, b) // true
```

An equal set built separately is not the same set:

```ant
let a = Set.of([1, 2])
a == Set.of([2, 1])         // true
Set.same(a, Set.of([2, 1])) // false
```

## Scope

Two empty sets are equal, and still two sets:

```ant
let a: Set[string] = Set.of([])
let b: Set[string] = Set.of([])
a == b         // true
Set.same(a, b) // false
```

`Set.union`, `Set.intersect`, and `Set.difference` return a new set, even when the result has the
same members as an argument:

```ant
let a = Set.of([1, 2])
let none: Set[int] = Set.of([])
Set.same(a, a.union(none))      // false
Set.same(a, a.intersect(a))     // false
Set.same(a, a.difference(none)) // false
```

`Set.add` and `Set.remove` change the set in place, so it stays the same set:

```ant
let a = Set.of(["x"])
let b = a
a.add("y")     // true
a.remove("x")  // true
Set.same(a, b) // true
b              // Set.of(["y"])
```

A set held in a record field is shared by a spread of the record, since the spread copies the field,
which is a reference:

```ant
type User = { name: string, roles: Set[string] }
let u = User{ name: "ann", roles: Set.of(["admin"]) }
let v = User{ ...u, name: "bob" }
Set.same(u.roles, v.roles) // true
```

In a pipeline, the left side is `a`:

```ant
let a = Set.of([1])
let b = a
a | Set.same(b) // true
```

Call it through `.` on any set:

```ant
let a = Set.of(["red"])
let b = a
a.same(b) // true
```

Name the arguments, in any order:

```ant
let a = Set.of([1])
Set.same(a: a, b: a)           // true
Set.same(b: a, a: Set.of([1])) // false
```

## Applications

Guard a function that moves every member of one set into another against being given one set twice:

```ant
func moveAll(from: Set[string], to: Set[string]) {
  if Set.same(from, to) {
    return
  }
  for x in from.values() {
    to.add(x)
    from.remove(x)
  }
}

let pending = Set.of(["a", "b"])
let done = Set.of(["z"])
moveAll(pending, done)
println("${pending} ${done}")
moveAll(done, done)
println(done)
```

```text
Set.of([]) Set.of(["z", "a", "b"])
Set.of(["z", "a", "b"])
```

## Properties and relations

`Set.same` is reflexive and symmetric:

```ant
let a = Set.of([1])
let b = a
assert(Set.same(a, a))
assert(Set.same(a, b) == Set.same(b, a))
```

A set that holds a `NaN` is not `==` to itself, since floats compare by IEEE 754, but it is still
the same set:

```ant
let a = Set.of([Math.NaN])
a == a         // false
Set.same(a, a) // true
```

When two names refer to the same set, a change through one shows through the other. For a copy it
does not:

```ant
let a = Set.of([1])
let b = a
let copy = Set.of(a.values())
a.add(2)
assert(Set.same(a, b) && b == Set.of([1, 2]))
assert(!Set.same(a, copy) && copy == Set.of([1]))
```

`Reflect.same` reports the same for two reflected sets:

```ant
let a = Set.of([1])
let t = type(Set[int])
assert(Reflect.same(Reflect.valueOf(t, a), Reflect.valueOf(t, a)))
assert(!Reflect.same(Reflect.valueOf(t, a), Reflect.valueOf(t, Set.of([1]))))
```

## Possible issues

Both arguments have the one type `Set[T]`. A `Set[int]` and a `Set[int | null]` are different types,
so comparing them is a compile error:

```ant
let a = Set.of([1])
let b: Set[int | null] = Set.of([1])
println(Set.same(a, b)) // error: cannot use Set[int | null] where Set[int] is expected
```

## See also

[Array.same](../Array/same.md) · [Map.same](../Map/same.md) · [Record.same](../Record/same.md) ·
[Reflect.same](../Reflect/same.md) · [Set.of](of.md) · [Set.union](union.md) ·
[== operator](../../language/operators/equality.md)
