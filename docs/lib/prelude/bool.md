<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# bool

```ant signature
func bool(x: bool): bool
```

bool returns the bool x holds: x is a bool or a named primitive over bool, which it unwraps.

<!-- end generated -->

## Details

- `bool` does not convert other kinds. There is no truthiness: a number, a string, or `null` is a
  compile error. Compare instead, such as `n != 0` or `s == "true"`.
- A named primitive type over `bool` is not a `bool`. A condition, `!`, `&&`, `||`, and `assert`
  need `bool(x)` to use its value.

## Basic examples

Unwrap a value of a named primitive type over `bool`:

```ant
type Verified = bool
let v = Verified(true)
v       // Verified(true)
bool(v) // true
```

On a plain `bool`, `bool` is the identity:

```ant
bool(true)  // true
bool(false) // false
```

## Scope

`bool` is a value, so it can be passed to a function. It unwraps every element:

```ant
type Enabled = bool
let flags = [Enabled(true), Enabled(false), Enabled(true)]
Array.map(flags, bool) // [true, false, true]
```

As a predicate, it keeps the tagged elements whose payload is `true`:

```ant
type Enabled = bool
let flags = [Enabled(true), Enabled(false), Enabled(true)]
Array.filter(flags, bool)     // [Enabled(true), Enabled(true)]
Array.countWhere(flags, bool) // 2
```

## Applications

Use a tagged flag in a condition:

```ant
type Verified = bool
type User = { name: string, verified: Verified }

func badge(u: User): string {
  if bool(u.verified) { "${u.name} (verified)" } else { u.name }
}

println(badge(User{ name: "Ann", verified: Verified(true) }))
println(badge(User{ name: "Bob", verified: Verified(false) }))
```

```text
Ann (verified)
Bob
```

Count the features that are on in a settings map:

```ant
type On = bool
let features = Map{ "search": On(true), "export": On(false), "sync": On(true) }
let on = Array.countWhere(Map.values(features), bool)
println("${on} of ${len(features)} features on")
```

```text
2 of 3 features on
```

## Properties and relations

`bool` gives back the payload a constructor wrapped:

```ant
type Flag = bool
for b in [true, false] {
  assert(bool(Flag(b)) == b)
}
```

Two tagged values are equal exactly when their payloads are:

```ant
type Flag = bool
let a = Flag(true)
let b = Flag(false)
assert((a == b) == (bool(a) == bool(b)))
```

## Possible issues

`bool` is not a truthiness test. A number is a compile error; compare it instead:

```ant
let n = 3
let some = bool(n) // error: bool converts only a bool or a named primitive type over bool, got int
```

```ant
let n = 3
n != 0 // true
```

`bool` does not parse a string. Compare the text:

```ant
let s = "true"
let b = bool(s) // error: bool converts only a bool or a named primitive type over bool, got string
```

```ant
let s = "true"
s == "true" // true
```

A tagged value cannot be a condition or an operand of `&&`, `||`, or `!`. Unwrap it first:

```ant
type Enabled = bool
let e = Enabled(true)
if e { println("on") } // error: cannot use Enabled where bool is expected
```

```ant
type Enabled = bool
let e = Enabled(true)
bool(e) && !false // true
```

You cannot compare a tagged value with a plain `bool` with `==`; unwrap it first:

```ant
type Enabled = bool
let same = Enabled(true) == true
// error: Enabled and bool cannot be compared with ==: neither is assignable to the other
```

```ant
type Enabled = bool
bool(Enabled(true)) == true // true
```

Printing or interpolating a tagged value shows the constructor; unwrap it to get `true`/`false`:

```ant
type Verified = bool
"${Verified(true)}" // "Verified(true)"
```

```ant
type Verified = bool
"${bool(Verified(true))}" // "true"
```

A nullable `bool` must be narrowed first, with `??` or a `null` check:

```ant
let maybe: bool | null = null
let b = bool(maybe)
// error: bool converts only a bool or a named primitive type over bool, got bool | null
```

```ant
let maybe: bool | null = null
maybe ?? false // false
```

## See also

[string](string.md), [int](int.md), [bool type](../../language/types/bool.md),
[Boolean literals](../../language/literals/booleans.md),
[logical operators](../../language/operators/logical.md), [type](../../language/keywords/type.md),
[Random.bool](../Random/bool.md)
