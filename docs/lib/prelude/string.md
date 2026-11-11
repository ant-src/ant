<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# string

```ant signature
func string(x: string): string
```

string returns the string x holds: x is a string or a named primitive over string, which it unwraps.
It does not format other values; [`String.from`](../String/from.md) does.

<!-- end generated -->

## Details

- `string` does not format or convert other kinds. A number, a `bool`, or `null` is a compile error.
  Use [String.from](../String/from.md) to render any value as text, and
  [String.toInt](../String/toInt.md) to read a number from text.
- A named primitive type over `string` is not a `string`. Library functions such as
  [String.length](../String/length.md) need `string(x)` to use its value.

## Basic examples

Unwrap a value of a named primitive type over `string`:

```ant
type UserId = string
let id = UserId("u-1")
id         // UserId("u-1")
string(id) // "u-1"
```

On a plain `string`, `string` is the identity:

```ant
string("héllo") // "héllo"
string("")      // ""
```

## Scope

Unwrap before calling a `String` function, directly or through `.`:

```ant
type Email = string
let e = Email("Ann@Example.com")
String.toLower(string(e)) // "ann@example.com"
string(e).length()        // 15
```

`string` is a value, so it can be passed to a function. It unwraps every element:

```ant
type Tag = string
let tags = [Tag("ant"), Tag("docs")]
tags.map(string) // ["ant", "docs"]
```

In a pipe:

```ant
type Email = string
let e = Email("Ann@Example.com")
e | string() | String.toLower() // "ann@example.com"
```

The keys of a map:

```ant
type Slug = string
let pages = Map{ Slug("home"): 1, Slug("about"): 2 }
pages | Map.keys() | Array.map(string) | Array.sort() // ["about", "home"]
```

## Applications

Build a URL path from a tagged identifier:

```ant
type UserId = string

func profileUrl(id: UserId): string {
  "/users/" + string(id)
}

println(profileUrl(UserId("u-42")))
```

```text
/users/u-42
```

Join tagged values into one line of text:

```ant
type Tag = string
let tags = [Tag("go"), Tag("ant"), Tag("docs")]
println(tags | Array.map(string) | String.join(", "))
```

```text
go, ant, docs
```

## Properties and relations

`string` gives back the payload a constructor wrapped, and the identity on a plain `string`:

```ant
type Name = string
for s in ["", "Ann", "日本"] {
  assert(string(Name(s)) == s)
  assert(string(s) == s)
}
```

`len` unwraps on its own, so it agrees with `String.length` of the payload:

```ant
type Code = string
let c = Code("ß-1")
assert(len(c) == String.length(string(c)))
```

## Possible issues

`string` does not render a number. That is a compile error; use `String.from` or interpolation:

```ant
let s = string(42)
// error: string converts only a string or a named primitive type over string, got int
```

```ant
String.from(42) // "42"
"${42}"         // "42"
```

`string` does not unwrap a named primitive type over another kind:

```ant
type Count = int
let s = string(Count(3))
// error: string converts only a string or a named primitive type over string, got Count
```

A tagged value is not accepted where a `string` is expected. Unwrap it first:

```ant
type UserId = string
let id = UserId("u-1")
let n = String.length(id) // error: cannot use UserId where string is expected
```

Its own namespace has no `String` functions, so `.` does not reach them either:

```ant
type UserId = string
let id = UserId("u-1")
let n = id.length() // error: UserId has no field or function length
```

A tagged value is not equal to a plain `string`, and comparing them is a compile error:

```ant
type UserId = string
let same = UserId("u-1") == "u-1"
// error: UserId and string cannot be compared with ==: neither is assignable to the other
```

Interpolation and `String.from` render a tagged value with its tag. Unwrap it for the bare text:

```ant
type UserId = string
let id = UserId("u-1")
"${id}"         // "UserId(\"u-1\")"
"${string(id)}" // "u-1"
```

Markup does not take a named primitive type inside `{...}`. Unwrap it:

```ant
type UserId = string
let id = UserId("u-1")
let h = <p>{id}</p>
// error: markup takes a string, a number, void, Html, or an array of them, got UserId; convert it, as {string(x)}
```

```ant
type UserId = string
let id = UserId("u-1")
println(<p>User {string(id)}</p>)
```

```text
<p>User u-1</p>
```

A nullable `string` must be narrowed first, with `??` or a `null` check:

```ant
let maybe: string | null = null
let s = string(maybe)
// error: string converts only a string or a named primitive type over string, got string | null
```

```ant
let maybe: string | null = null
maybe ?? "" // ""
```

## See also

[bool](bool.md), [int](int.md), [len](len.md), [String.from](../String/from.md),
[String.toInt](../String/toInt.md), [string type](../../language/types/string.md),
[String literals](../../language/literals/strings.md), [type](../../language/keywords/type.md)
