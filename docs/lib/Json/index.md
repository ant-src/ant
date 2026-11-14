<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Json

Json converts between JSON text and values.

<!-- end generated -->

## Details

- [Json.parse](parse.md) reads text into a [Json.Value](types/Value.md), a union of `null`, `bool`,
  `float64`, `string`, arrays, and maps. Narrow it with `is` before you index it. Malformed text is
  an `Err`, not a panic.
- Every JSON number parses as a `float64`, so `1` reads back as `1.0`. Convert with
  [Math.toInt](../Math/toInt.md) where you need an `int`.
- [Json.stringify](stringify.md) takes any value, typed data included: a record becomes an object
  without its tag, a set an array, a time value its string, and a named primitive its payload.
- `Json.stringify` panics on a value with no JSON form: `NaN`, an infinity, `Bytes`, `Html`, a
  function, a reference such as a `Mutex`, a value that contains itself, or a map key that is not a
  `string` or a named primitive over `string`.
- Json has no typed decoder. Turn a `Json.Value` into records with a function per type that narrows
  with `is` and returns a `Result`, or write a generic one over [Reflect](../Reflect/index.md).

## Basic examples

Parse JSON text with `Json.parse`:

```ant
Json.parse(`{"name": "Ann", "tags": ["admin"]}`) // Ok{ value: Map{ "name": "Ann", "tags": ["admin"] } }
Json.parse("[1, 2,]")                            // Err{ error: Error{ message: "expected a value, found ']' at line 1, column 7" } }
```

Write any value as JSON text with `Json.stringify`:

```ant
type User = { name: string, age: int }
Json.stringify(User{ name: "Ann", age: 31 }) // "{\"name\":\"Ann\",\"age\":31}"
Json.stringify(Map{ "ids": [1, 2] })         // "{\"ids\":[1,2]}"
```

Pass `indent` to spread the text over lines:

```ant
type User = { name: string, tags: Array[string] }
println(Json.stringify(User{ name: "Ann", tags: ["admin"] }, indent: "  "))
```

```text
{
  "name": "Ann",
  "tags": [
    "admin"
  ]
}
```

## Scope

Each kind of JSON text parses to its own member of `Json.Value`:

```ant
Json.parse(`[null, true, 2, "s", [], {}]`) // Ok{ value: [null, true, 2.0, "s", [], Map{}] }
```

Typed values write as the JSON they hold, and read back as plain `Json.Value`:

```ant
type Sku = string
type Item = { sku: Sku, tags: Set[string], qty: int, note: string | null }
let item = Item{ sku: Sku("a-1"), tags: Set.of(["new"]), qty: 2, note: null }
let text = Json.stringify(item)
text                      // "{\"sku\":\"a-1\",\"tags\":[\"new\"],\"qty\":2,\"note\":null}"
Json.parse(text).unwrap() // Map{ "sku": "a-1", "tags": ["new"], "qty": 2.0, "note": null }
```

A `Result` is a tagged union of records, so it writes as the fields of its member:

```ant
let done: Result[int, Error] = Result.ok(3)
let failed: Result[int, Error] = Result.fail("bad")
Json.stringify(done)   // "{\"value\":3}"
Json.stringify(failed) // "{\"error\":{\"message\":\"bad\"}}"
```

## Functions by job

### Reading

- [Json.parse](parse.md): one JSON text as a `Json.Value`, or an `Err` that says what went wrong and
  where

### Writing

- [Json.stringify](stringify.md): the JSON text of any value, compact or indented with `indent:`

### Types

- [Json.Value](types/Value.md): a JSON value of unknown shape, narrowed with `is`

## Applications

### Load a config file into a record

Read the file, parse it, and build the record, with `try` for each step that can fail and
[Result.wrapErr](../Result/wrapErr.md) to name the file:

<!-- test: file config.json
{"name": "api", "port": 8080, "debug": true}
-->

```ant
type Config = { name: string, port: int, debug: bool = false }

func decodeConfig(v: Json.Value): Result[Config, Error] {
  if !(v is Map) {
    return Result.fail("want an object")
  }
  let name = v["name"]
  if !(name is string) {
    return Result.fail("name: want a string")
  }
  let port = v["port"]
  let n = if port is float64 { Math.toInt(port) } else { null }
  if n == null {
    return Result.fail("port: want an integer")
  }
  let debug = v["debug"]
  Result.ok(Config{ name, port: n, debug: debug is bool && debug })
}

func loadConfig(path: string): Result[Config, Error] {
  let text = try File.read(path)
  let doc = try Json.parse(text)
  decodeConfig(doc) | Result.wrapErr(path)
}

println(loadConfig("config.json"))
```

```text
Ok{ value: Config{ name: "api", port: 8080, debug: true } }
```

### Read JSON Lines

Parse each line on its own, and name the line of the first bad one:

```ant
func total(text: string): Result[float64, Error] {
  var sum = 0.0
  for { index, value } in text.lines().entries() {
    let doc = try (Json.parse(value) | Result.wrapErr("line ${index + 1}"))
    let ms = if doc is Map { doc["ms"] } else { null }
    if ms is float64 {
      sum += ms
    }
  }
  Result.ok(sum)
}

println(total("{\"ms\": 12}\n{\"ms\": 30.5}"))
println(total("{\"ms\": 12}\n{\"ms\": }"))
```

```text
Ok{ value: 42.5 }
Err{ error: Error{ message: "line 2: expected a value, found '}' at line 1, column 8" } }
```

### Decode nested records with a path in the error

Pass each decoder the path of the value it reads, and extend it for each field and element, so an
error names the bad value's place in the document:

```ant
type Server = { port: int }
type Fleet = { servers: Array[Server] }

func decodeServer(v: Json.Value, path: string): Result[Server, Error] {
  if !(v is Map) {
    return Result.fail("${path}: want an object")
  }
  let port = v["port"]
  let n = if port is float64 { Math.toInt(port) } else { null }
  if n == null {
    return Result.fail("${path}.port: want an integer")
  }
  Result.ok(Server{ port: n })
}

func decodeFleet(v: Json.Value, path: string): Result[Fleet, Error] {
  if !(v is Map) {
    return Result.fail("${path}: want an object")
  }
  let list = v["servers"]
  if !(list is Array) {
    return Result.fail("${path}.servers: want an array")
  }
  let servers: Array[Server] = []
  for { index, value } in list.entries() {
    servers.push(try decodeServer(value, "${path}.servers[${index}]"))
  }
  Result.ok(Fleet{ servers })
}

let doc = Json.parse(`{"servers": [{"port": 80}, {"port": "x"}]}`).unwrap()
println(decodeFleet(doc, "$"))
```

```text
Err{ error: Error{ message: "$.servers[1].port: want an integer" } }
```

### Decode a tagged union

JSON has no tags, so a document names the member in a field of its own. Test that field to pick the
record:

```ant
type Circle = { radius: float64 }
type Square = { side: float64 }
type Shape = Circle | Square

func decodeShape(v: Json.Value): Result[Shape, Error] {
  if !(v is Map) {
    return Result.fail("want an object")
  }
  let kind = v["kind"]
  let size = v["size"]
  if !(size is float64) {
    return Result.fail("size: want a number")
  }
  if kind == "circle" {
    return Result.ok(Circle{ radius: size })
  }
  if kind == "square" {
    return Result.ok(Square{ side: size })
  }
  Result.fail("kind: want \"circle\" or \"square\", got ${Json.stringify(kind)}")
}

let text = `[{"kind": "circle", "size": 1}, {"kind": "square", "size": 2}, {"kind": "star", "size": 3}]`
let doc = Json.parse(text).unwrap()
if doc is Array {
  for v in doc {
    println(decodeShape(v))
  }
}
```

```text
Ok{ value: Circle{ radius: 1.0 } }
Ok{ value: Square{ side: 2.0 } }
Err{ error: Error{ message: "kind: want \"circle\" or \"square\", got \"star\"" } }
```

### Change one field of a document

Edit the map in place and write it back. The other keys keep their order:

```ant
let doc = Json.parse(`{"version": 1, "name": "ant", "private": true}`).unwrap()
if doc is Map {
  doc["version"] = 2.0
  doc.remove("private")
}
println(Json.stringify(doc, indent: "  "))
```

```text
{
  "version": 2.0,
  "name": "ant"
}
```

## Properties and relations

A `Json.Value` without `NaN` or an infinity survives a round trip, with or without `indent`:

```ant
let v: Json.Value = Map{ "a": [1.0, null, "é"], "b": Map{ "c": true } }
assert(Json.parse(Json.stringify(v)).unwrap() == v)
assert(Json.parse(Json.stringify(v, indent: "\t")).unwrap() == v)
```

A record writes the same text as a map from its field names to its values:

```ant
type Point = { x: float64, y: float64 }
assert(Json.stringify(Point{ x: 1.0, y: 2.0 }) == Json.stringify(Map{ "x": 1.0, "y": 2.0 }))
```

## Possible issues

### `Json.parse` does not build records

It returns a `Json.Value`, which is not your record type:

```ant
type Config = { name: string, port: int }
let c: Config = Json.parse(`{"name": "a", "port": 1}`).unwrap()
// error: cannot use Result[Json.Value, Error] where Result[Config, E] is expected
```

### A `Json.Value` must be narrowed

```ant
let doc = Json.parse(`{"a": 1}`).unwrap()
let a = doc["a"] // error: Json.Value cannot be indexed
```

### Integers come back as floats

```ant
let text = Json.stringify(Map{ "n": 1 })
text                      // "{\"n\":1}"
Json.parse(text).unwrap() // Map{ "n": 1.0 }
```

### The tag is lost

`Json.stringify` writes a record without its tag, so the members of an enumeration all write as
`{}`. Write a string for each member instead:

```ant
type Low = {}
type High = {}
type Level = Low | High
let levels: Array[Level] = [Low{}, High{}]
Json.stringify(levels)                                                // "[{},{}]"
Json.stringify(levels.map({ if x is Low { "low" } else { "high" } })) // "[\"low\",\"high\"]"
```

### Values with no JSON form panic

```ant
Json.stringify(Map{ 1: "gold" }) // panic: Json.stringify: a map key must be a string, got int
Json.stringify([0.0 / 0.0])      // panic: Json.stringify: NaN has no JSON form
Json.stringify("hi".toBytes())   // panic: Json.stringify: Bytes has no JSON form; write Encoding.toBase64 of it
```

### `unwrap` on text from outside

Malformed text makes `Result.unwrap` panic. Use `try` where the text comes from a file, a request,
or a user:

```ant
Json.parse("{").unwrap() // panic: Result.unwrap on an Err: Error{ message: "unexpected end of JSON input" }
```

## See also

[Json.parse](parse.md) · [Json.stringify](stringify.md) · [Json.Value](types/Value.md) ·
[Math.toInt](../Math/toInt.md) · [Reflect](../Reflect/index.md) · [Csv](../Csv/index.md) ·
[Encoding.toBase64](../Encoding/toBase64.md) · [Result.wrapErr](../Result/wrapErr.md) ·
[Errors guide](../../guide/errors.md)
