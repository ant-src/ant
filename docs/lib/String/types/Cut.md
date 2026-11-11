<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.Cut

```ant signature
type String.Cut = { before: string, after: string }
```

String.Cut is the two sides of a string around a separator, as [`String.cut`](../cut.md) returns
them.

<!-- end generated -->

## Details

- [`String.cut`](../cut.md) returns a `String.Cut` when the separator occurs, with `before` the text
  in front of its first occurrence and `after` the text that follows it, and `null` when it does
  not.

## Basic examples

Get a `String.Cut` from `String.cut`, check it for `null`, and read its fields:

```ant
for line in ["name=ant", "oops"] {
  let c = line.cut("=")
  if c != null {
    println("${c.before} -> ${c.after}")
  } else {
    println("no '=' in ${line}")
  }
}
```

```text
name -> ant
no '=' in oops
```

Only the first separator splits; `after` keeps the later ones, and a separator at an end leaves an
empty field:

```ant
String.cut("k=v=w", "=") // String.Cut{ before: "k", after: "v=w" }
String.cut("=v", "=")    // String.Cut{ before: "", after: "v" }
```

Write one with a literal:

```ant
String.Cut{ before: "k", after: "v" } // String.Cut{ before: "k", after: "v" }
```

Destructure the fields into variables:

```ant
let { before, after } = String.cut("host:8080", ":") ?? return
println(before)
println(after)
```

```text
host
8080
```

## Scope

Rename the fields while destructuring:

```ant
let { before: key, after: value } = String.cut("lang=ant", "=") ?? String.Cut{ before: "", after: "" }
key   // "lang"
value // "ant"
```

Annotate a variable, a parameter, or a return type with `String.Cut`:

```ant
func pair(s: string, sep: string): String.Cut {
  String.cut(s, sep) ?? String.Cut{ before: s, after: "" }
}

let p: String.Cut = pair("a:b", ":")
println(p)
println(pair("a", ":"))
```

```text
String.Cut{ before: "a", after: "b" }
String.Cut{ before: "a", after: "" }
```

Assign to a field, as with any record:

```ant
let c = String.Cut{ before: "k", after: "v" }
c.after = "w"
c // String.Cut{ before: "k", after: "w" }
```

Copy one with a spread and change a field:

```ant
let c = String.Cut{ before: "k", after: "v" }
String.Cut{ ...c, before: "key" } // String.Cut{ before: "key", after: "v" }
```

Test a union member with `is`:

```ant
func describe(x: String.Cut | string): string {
  if x is String.Cut { "${x.before} | ${x.after}" } else { x }
}

println(describe(String.Cut{ before: "a", after: "b" }))
println(describe("whole"))
```

```text
a | b
whole
```

An array of results holds `String.Cut` records and `null`s:

```ant
["a=1", "b=2", "c"].map({ x.cut("=") }) // [String.Cut{ before: "a", after: "1" }, String.Cut{ before: "b", after: "2" }, null]
```

A function that takes an anonymous record type accepts a `String.Cut` with the fields it lists:

```ant
func head(x: { before: string }): string { x.before }

head(String.Cut{ before: "a", after: "b" }) // "a"
```

`String.from` and `Json.stringify` show both fields:

```ant
let c = String.Cut{ before: "k", after: "v" }
String.from(c)    // "String.Cut{ before: \"k\", after: \"v\" }"
Json.stringify(c) // "{\"before\":\"k\",\"after\":\"v\"}"
```

## Applications

Split `user@domain` addresses and group the users by domain:

```ant
let users: Map[string, Array[string]] = Map{}
for address in ["ann@a.org", "bob@b.org", "cat@a.org", "nobody"] {
  let { before: user, after: domain } = address.cut("@") ?? continue
  let list = users[domain] ?? []
  list.push(user)
  users[domain] = list
}
println(users)
```

```text
Map{ "a.org": ["ann", "cat"], "b.org": ["bob"] }
```

Return a `String.Cut` from your own function, here one that cuts at the last separator:

```ant
func cutLast(s: string, sep: string): String.Cut | null {
  let i = s.lastIndexOf(sep) ?? return null
  String.Cut{ before: s.slice(0, i), after: s.slice(i + sep.length(), s.length()) }
}

println(cutLast("archive.tar.gz", "."))
println(cutLast("README", "."))
```

```text
String.Cut{ before: "archive.tar", after: "gz" }
null
```

## Properties and relations

Two `String.Cut` records are equal when their fields are equal:

```ant
assert(String.cut("k=v", "=") == String.Cut{ before: "k", after: "v" })
assert(String.Cut{ before: "a", after: "b" } != String.Cut{ before: "b", after: "a" })
```

Assigning a `String.Cut` shares it; a spread makes a copy:

```ant
let a = String.Cut{ before: "k", after: "v" }
let b = a
let copy = String.Cut{ ...a }
b.after = "w"
assert(a.after == "w")
assert(copy.after == "v")
assert(Record.same(a, b))
assert(!Record.same(a, copy))
```

## Possible issues

The type lives in the `String` namespace. `Cut` alone is not a type:

```ant
let c: Cut = String.Cut{ before: "a", after: "b" } // error: unknown type Cut
```

A literal gives both fields:

```ant
let c = String.Cut{ before: "a" } // error: String.Cut literal is missing the field after
```

`String.cut` returns `String.Cut | null`, so reading a field before narrowing is a compile error:

```ant
let c = String.cut("a=b", "=")
println(c.before) // error: String.Cut | null has no field before: narrow the value first
```

A record of your own with the same fields is a different type. It does not compare with a
`String.Cut`, and a spread converts it:

```ant
type Pair = { before: string, after: string }
let p = Pair{ before: "k", after: "v" }
println(p == String.Cut{ before: "k", after: "v" })
// error: Pair and String.Cut cannot be compared with ==: neither is assignable to the other
```

```ant
type Pair = { before: string, after: string }
let p = Pair{ before: "k", after: "v" }
String.Cut{ ...p } == String.Cut{ before: "k", after: "v" } // true
```

## See also

[String.cut](../cut.md) · [String.split](../split.md) · [String.indexOf](../indexOf.md) ·
[String.lastIndexOf](../lastIndexOf.md) · [Indexed](../../Array/types/Indexed.md) ·
[Entry](../../Map/types/Entry.md)
