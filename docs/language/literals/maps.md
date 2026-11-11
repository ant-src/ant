<!-- status: reviewed -->

# Map literals

`Map{ key: value, ... }` builds a new `Map[K, V]` with the given entries, in the order written.

## Details

- A key that is a single word, an identifier or a keyword, is the string of that word: `Map{ a: 1 }`
  has the key `"a"`. Any other key is an expression, evaluated as written. Write a variable as a key
  in parentheses: `Map{ (name): 1 }`.
- A bare word with no `:` is shorthand: `Map{ name }` means `Map{ "name": name }`.
- A key repeated in one literal keeps the position of its first entry and the value of its last.
- Each evaluation builds a new map. An empty literal `Map{}` needs an expected type, such as an
  annotation.
- In the header of an `if`, `while`, or `for`, put a map literal in parentheses.

## Basic examples

Word keys are strings:

```ant
Map{ ada: 36, linus: 54 } // Map{ "ada": 36, "linus": 54 }
```

Read a value by its key. A missing key gives `null`:

```ant
let ages = Map{ ada: 36, linus: 54 }
ages["ada"]    // 36
ages["bjarne"] // null
```

An empty map takes its key and value types from an annotation:

```ant
let empty: Map[string, int] = Map{}
empty // Map{}
```

## Scope

### Keywords as keys

A keyword before `:` is a plain word too, so it gives a string key:

```ant
Map{ type: "button", class: "primary", null: 0 } // Map{ "type": "button", "class": "primary", "null": 0 }
```

### Quoted keys

A string literal gives a key that is not a single word:

```ant
Map{ "Content-Type": "text/html", "two words": "ok" } // Map{ "Content-Type": "text/html", "two words": "ok" }
```

### Computed keys

A parenthesized identifier is the variable's value. Any key that is not a single word is evaluated
as written, such as a field access or a call:

```ant
type User = { name: string }
let lang = "go"
let user = User{ name: "ann" }
Map{ (lang): 2009, user.name: 1, "x".toUpper(): 2 } // Map{ "go": 2009, "ann": 1, "X": 2 }
```

### Number keys

A number key gives a map with number keys:

```ant
let squares = Map{ 1: 1, 2: 4, 3: 9 }
squares[2] // 4
```

### Record and array keys

Keys compare the way `==` does, so a record or an array can be a key:

```ant
type Point = { x: int, y: int }
let labels = Map{ Point{ x: 0, y: 0 }: "origin", Point{ x: 1, y: 0 }: "east" }
labels[Point{ x: 1, y: 0 }] // "east"
let paths = Map{ ["usr", "bin"]: 2, ["etc"]: 1 }
paths[["usr", "bin"]] // 2
```

### Shorthand

A bare name makes an entry whose key is the name and whose value is the variable:

```ant
let host = "localhost"
let port = 8080
Map{ host, port: port + 1 } // Map{ "host": "localhost", "port": 8081 }
```

### Values of any type

Values can be arrays, records, or other maps:

```ant
let greetings = Map{ en: Map{ hi: "hello" }, fr: Map{ hi: "bonjour" } }
greetings["fr"]?.["hi"] // "bonjour"
let primes = Map{ small: [2, 3, 5], large: [97] }
primes["small"]?.length() // 3
```

### Expected types

An annotation fixes the key and value types, and each entry is checked against them:

```ant
let limits: Map[string, int | null] = Map{ cpu: 4, memory: null }
limits // Map{ "cpu": 4, "memory": null }
```

### Many lines

A literal may span lines, with a trailing comma after the last entry:

```ant
let codes = Map{
  ok: 200,
  notFound: 404,
  teapot: 418,
}
codes.keys() // ["ok", "notFound", "teapot"]
```

### Repeated keys

A later entry replaces the value of an earlier one with the same key, and the key keeps its first
position:

```ant
Map{ a: 1, b: 2, a: 3 } // Map{ "a": 3, "b": 2 }
```

### In a header

In an `if`, `while`, or `for` header, parenthesize the literal:

```ant
for e in (Map{ red: "#f00", green: "#0f0" }) {
  println("${e.key} ${e.value}")
}
```

```text
red #f00
green #0f0
```

## Applications

### Lookup table

A literal replaces a chain of `if` tests with data:

```ant
let plural = Map{ child: "children", mouse: "mice", person: "people" }

func pluralize(word: string): string {
  plural[word] ?? "${word}s"
}

println(["child", "cat", "person"] | Array.map(pluralize))
```

```text
["children", "cats", "people"]
```

### Dispatch on a command

Map names to functions and call the one a user picks:

```ant
let commands: Map[string, func(string): string] = Map{
  upper: { x.toUpper() },
  lower: { x.toLower() },
}

func run(name: string, input: string): string {
  commands[name]?.(input) ?? "unknown command: ${name}"
}

println(run("upper", "ant"))
println(run("sort", "ant"))
```

```text
ANT
unknown command: sort
```

### Default options

A map literal works as a default value, and a fresh map is built on every call that omits it. The
function appends to the `accept` header of whatever map it receives, so a reused default would grow
on the second call:

```ant
func request(url: string, headers: Map[string, string] = Map{ accept: "text/html" }): string {
  headers["accept"] = "${headers["accept"] ?? "text/plain"}, */*"
  "${url} ${headers}"
}

println(request("/a"))
println(request("/a"))
println(request("/b", headers: Map{ accept: "application/json" }))
```

```text
/a Map{ "accept": "text/html, */*" }
/a Map{ "accept": "text/html, */*" }
/b Map{ "accept": "application/json, */*" }
```

## Properties and relations

The shorthand is the same as a quoted key with the variable as its value, and a word key is the same
as the quoted word:

```ant
let name = "ann"
assert(Map{ name } == Map{ "name": name })
assert(Map{ id: 1 } == Map{ "id": 1 })
```

Two maps are `==` when they hold the same entries, whatever their order, though the order shows when
you iterate:

```ant
let a = Map{ x: 1, y: 2 }
let b = Map{ y: 2, x: 1 }
assert(a == b)
assert(a.keys() == ["x", "y"])
assert(b.keys() == ["y", "x"])
```

Each evaluation of a literal builds a new map:

```ant
func counters(): Map[string, int] { Map{ hits: 0 } }

let first = counters()
first["hits"] = 5
assert(counters()["hits"] == 0)
assert(!Map.same(counters(), counters()))
```

## Possible issues

### A bare word is a string, not a variable

A word before `:` is never read as a variable. Parenthesize it to use its value:

```ant
let key = "id"
Map{ key: 1 }   // Map{ "key": 1 }
Map{ (key): 1 } // Map{ "id": 1 }
```

### An empty map needs a type

`Map{}` alone has no key or value type:

```ant
let cache = Map{} // error: cannot infer the key and value types of an empty map; annotate it
```

### Unparenthesized in a header

In a header, the `{` of a literal would begin the body:

```ant
let colors = Map{ red: "#f00" }
if colors == Map{ red: "#f00" } { println("same") }
// error: a record or map literal in the if header must be parenthesized
```

### No spread

A map literal takes only entries. Combine maps with `Map.merge`, whose second map wins:

```ant
let defaults = Map{ color: "red", size: "m" }
let custom = Map{ ...defaults, size: "l" } // error: spread has no meaning in a map literal
```

```ant
let defaults = Map{ color: "red", size: "m" }
defaults.merge(Map{ size: "l" }) // Map{ "color": "red", "size": "l" }
```

### Keyword shorthand

The shorthand reads a variable of the same name, and no variable can be named by a keyword:

```ant
let m = Map{ type } // error: 'type' is a keyword: as a key it needs ':'
```

### Mutating a key

A key is hashed when the entry is inserted. Changing an array or record after using it as a key is a
program error, and the entry may become unreachable:

```ant
let path = ["usr"]
let sizes = Map{ (path): 10 }
path.push("bin")
println(sizes[["usr"]])
```

```text
null
```

The map shares `path`, so its key is now `["usr", "bin"]` and the lookup by `["usr"]` misses. The
entry may also be unreachable by its new value.

Use a value you do not change as a key, or copy it first, as `Map{ [...path]: 10 }`.

## See also

[Map](../types/Map.md) · [Map.merge](../../lib/Map/merge.md) ·
[Map.fromEntries](../../lib/Map/fromEntries.md) · [Map.get](../../lib/Map/get.md) ·
[Map.has](../../lib/Map/has.md) · [Entry](../../lib/Map/types/Entry.md) ·
[Record literals](records.md) · [Array literals](arrays.md) · [Set literals](sets.md) ·
[Indexing](../operators/indexing.md) · [Collections](../../guide/collections.md)
