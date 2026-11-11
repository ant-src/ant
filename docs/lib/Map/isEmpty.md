<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.isEmpty

```ant signature
func Map.isEmpty[K, V](m: Map[K, V]): bool
```

Map.isEmpty reports whether m has no entries.

<!-- end generated -->

## Details

- An entry whose value is `null` is still an entry, so a map that holds one is not empty.

## Basic examples

```ant
Map.isEmpty(Map{ ada: 36, linus: 54 }) // false
let none: Map[string, int] = Map{}
Map.isEmpty(none) // true
```

Call it through `.` on any map:

```ant
let ages = Map{ ada: 36 }
ages.isEmpty() // false
```

## Scope

An entry whose value is `null` keeps the map from being empty:

```ant
let seen: Map[string, int | null] = Map{ ada: null }
seen.isEmpty() // false
```

The answer follows the map as it changes:

```ant
let counts: Map[string, int] = Map{}
counts.isEmpty() // true
counts["go"] = 1
counts.isEmpty()    // false
counts.remove("go") // 1
counts.isEmpty()    // true
```

The result of another function, such as `Map.filter`, is often what you test:

```ant
Map{ ada: 36, linus: 54 }.filter({ x.value > 60 }).isEmpty() // true
```

In a pipeline, the map is the first argument:

```ant
Map{ ada: 36 } | Map.isEmpty() // false
```

As a function value, it maps an array of maps to an array of `bool`s:

```ant
let empty: Map[string, int] = Map{}
[Map{ a: 1 }, empty].map(Map.isEmpty) // [false, true]
```

For a map that may be `null`, `?.` yields `null` instead of calling:

```ant
let headers: Map[string, string] | null = null
headers?.isEmpty()         // null
headers?.isEmpty() ?? true // true
```

## Applications

Print a word count, or a note when there are no words:

```ant
func report(text: string) {
  let counts: Map[string, int] = Map{}
  for word in text.fields() {
    counts[word] = (counts[word] ?? 0) + 1
  }
  if counts.isEmpty() {
    println("no words")
    return
  }
  for e in counts.entries() {
    println("${e.key}: ${e.value}")
  }
}

report("to be or not to be")
report("   ")
```

```text
to: 2
be: 2
or: 1
not: 1
no words
```

Report the missing fields of a form, or accept it:

```ant
func missing(form: Map[string, string], required: Array[string]): Map[string, string] {
  let problems: Map[string, string] = Map{}
  for name in required {
    if (form[name] ?? "").isEmpty() {
      problems[name] = "required"
    }
  }
  problems
}

let form = Map{ name: "Ann", email: "" }
let problems = missing(form, ["name", "email"])
println(if problems.isEmpty() { "ok" } else { "fix: ${problems.keys()}" })
```

```text
fix: ["email"]
```

## Properties and relations

`Map.isEmpty` agrees with `Map.size`, `len`, `Map.keys`, and `==` against an annotated empty map:

```ant
let none: Map[string, int] = Map{}
let cases: Array[Map[string, int]] = [Map{}, Map{ a: 1 }, Map{ a: 1, b: 2 }]
for m in cases {
  assert(m.isEmpty() == (m.size() == 0))
  assert(m.isEmpty() == (len(m) == 0))
  assert(m.isEmpty() == m.keys().isEmpty())
  assert(m.isEmpty() == (m == none))
}
```

## Possible issues

A map is not a condition. Ask `Map.isEmpty` explicitly:

```ant
let m = Map{ a: 1 }
if m { // error: cannot use Map[string, int] where bool is expected
  println("has entries")
}
```

An empty map literal has no key and value types to infer. Annotate it:

```ant
println(Map{}.isEmpty()) // error: cannot infer the key and value types of an empty map; annotate it
```

An operand of `==` gives no expected type, so `m == Map{}` cannot infer the types either. Use
`Map.isEmpty`:

```ant
let m = Map{ a: 1 }
println(m == Map{}) // error: cannot infer the key and value types of an empty map; annotate it
```

A nullable map must be narrowed or reached through `?.` first:

```ant
let m: Map[string, int] | null = null
println(Map.isEmpty(m)) // error: cannot use Map[string, int] | null where Map[K, V] is expected
```

`Map.isEmpty` takes only maps. Arrays, strings, and sets have their own:

```ant
println(Map.isEmpty(Set.of([1]))) // error: cannot use Set[int] where Map[K, V] is expected
```

## See also

[Map.size](size.md) · [len](../prelude/len.md) · [Map.has](has.md) · [Map.remove](remove.md) ·
[Map.filter](filter.md) · [Array.isEmpty](../Array/isEmpty.md) · [Set.isEmpty](../Set/isEmpty.md) ·
[String.isEmpty](../String/isEmpty.md)
