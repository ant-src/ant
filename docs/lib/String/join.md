<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.join

```ant signature
func String.join(parts: Array[string], sep: string): string
```

String.join returns the strings in parts joined with sep between each two.

<!-- end generated -->

## Details

- `String.join` leaves `parts` unchanged and returns a new string
  ([spec](https://github.com/ant-src/spec/blob/master/ant-lang.md#11-standard-library)).
- Its subject is an array, and a value reaches its own type's namespace, so `parts.join(",")` looks
  for `Array.join`, which does not exist. Call `String.join(parts, ",")` or use a pipe
  ([spec](https://github.com/ant-src/spec/blob/master/ant-lang.md#calls-through-)).

## Basic examples

Join words with a comma and a space:

```ant
String.join(["ant", "bee", "wasp"], ", ") // "ant, bee, wasp"
```

Join lines with a line break:

```ant
println(String.join(["first", "second", "third"], "\n"))
```

```text
first
second
third
```

## Scope

Join an empty array and an array of one part:

```ant
var none: Array[string] = []
String.join(none, ",")     // ""
String.join(["only"], ",") // "only"
```

An empty part still counts as a part, so the separators next to it stay:

```ant
String.join(["a", "", "b"], ",") // "a,,b"
String.join(["", ""], ",")       // ","
String.join([""], ",")           // ""
```

The separator can be any string: empty, several code points long, or in any script:

```ant
String.join(["a", "b", "c"], "")     // "abc"
String.join(["a", "b", "c"], " -> ") // "a -> b -> c"
String.join(["日本", "🐜"], "·")        // "日本·🐜"
```

Pass the parts with a pipe, which puts the array in as the first argument:

```ant
["x", "y", "z"] | String.join("/") // "x/y/z"
```

The array can come from any expression, such as a map over numbers:

```ant
String.join(Array.map([1, 2, 3], { String.from(x) }), " + ") // "1 + 2 + 3"
```

## Named arguments

Name the arguments `parts` and `sep`, in either order:

```ant
String.join(parts: ["x", "y"], sep: "+") // "x+y"
String.join(sep: "+", parts: ["x", "y"]) // "x+y"
```

## Applications

Replace one separator with another:

```ant
String.join(String.split("2026-11-08", "-"), "/") // "2026/11/08"
```

Collapse every run of white space to one space:

```ant
String.join(String.fields("  too   many\tspaces \n"), " ") // "too many spaces"
```

Write a row of comma-separated values from a record:

```ant
type Ant = { name: string, legs: int, queen: bool }

func row(a: Ant): string {
  String.join([a.name, String.from(a.legs), String.from(a.queen)], ",")
}

println(row(Ant{ name: "Formica", legs: 6, queen: false }))
```

```text
Formica,6,false
```

List names in a sentence, with "and" before the last one:

```ant
func listing(names: Array[string]): string {
  var n = len(names)
  if n < 2 {
    return String.join(names, "")
  }
  String.join(Array.slice(names, 0, n - 1), ", ") + " and " + names[n - 1]
}

println(listing(["Ann"]))
println(listing(["Ann", "Bob"]))
println(listing(["Ann", "Bob", "Cy"]))
```

```text
Ann
Ann and Bob
Ann, Bob and Cy
```

## Properties and relations

Joining with `""` matches [String.concat](concat.md) here:

```ant
var parts = ["a", "bc", "", "d"]
assert(String.join(parts, "") == String.concat(parts))
```

Splitting this string with [String.split](split.md) and joining it back with the same separator
gives the string again:

```ant
var s = "a,b,,c"
assert(String.join(String.split(s, ","), ",") == s)
```

Here the length of the result is the sum of the lengths of the parts, plus one separator fewer than
there are parts:

```ant
var parts = ["é", "日本", "", "🐜"]
var sep = ", "
var total = Array.sum(Array.map(parts, { String.length(x) })) ?? 0
assert(String.length(String.join(parts, sep)) == total + (len(parts) - 1) * String.length(sep))
```

## Possible issues

`parts.join(",")` looks in the `Array` namespace, which has no `join`, so the call is a compile
error. Call `String.join(parts, ",")` or use a pipe:

<!-- test: error Array[string] has no field or function join -->

```ant
var s = ["a", "b"].join(",")
```

A string is not the subject either: `String.join` takes the array first:

<!-- test: error String.join is not a function on string: it takes Array[string] first -->

```ant
var s = ",".join(["a", "b"])
```

The separator has no default. Leaving it out is a compile error; pass `""` or use
[String.concat](concat.md):

<!-- test: error String.join needs an argument for sep -->

```ant
var s = String.join(["a", "b"])
```

Every part must be a string. Convert other values with [String.from](from.md) first:

<!-- test: error cannot use int where string is expected -->

```ant
var s = String.join(["a", 1], ",")
```

Interpolating an array does not join it: it renders the array with its brackets and quotes:

```ant
var xs = ["a", "b"]
"${xs}"                    // "[\"a\", \"b\"]"
"${String.join(xs, ", ")}" // "a, b"
```

[String.split](split.md) does not always undo `String.join`. In these two cases the split gives back
different parts:

```ant
String.split(String.join(["a,b", "c"], ","), ",") // ["a", "b", "c"]
var none: Array[string] = []
String.split(String.join(none, ","), ",") // [""]
```

## See also

[String.concat](concat.md), [String.split](split.md), [String.fields](fields.md),
[String.lines](lines.md), [String.repeat](repeat.md), [String.from](from.md),
[Path.join](../Path/join.md)
