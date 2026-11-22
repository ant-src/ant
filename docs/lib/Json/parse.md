<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Json.parse

```ant signature
func Json.parse(text: string): Result[Json.Value, Error]
```

Json.parse parses text as one JSON value. It is Err when text is not valid JSON, repeats a name in
an object, escapes a surrogate outside a pair, nests arrays and objects deeper than 10000 (1000 on
WebAssembly), or holds a number too large for a float64.

<!-- end generated -->

## Details

- Every number becomes a `float64`, so `1` parses as `1.0`. Convert one with
  [Math.toInt](../Math/toInt.md) where an `int` is needed.
- An object becomes a `Map` with its keys in the order of the text. An object that repeats a name is
  an `Err`, since readers that disagree on which value counts would read one text two ways.
- White space may surround the value; anything else after it is an `Err`. Comments, trailing commas,
  single quotes, and `NaN` are not JSON and are `Err` too.

## Basic examples

```ant
Json.parse(`{"name": "Ann", "tags": ["admin"]}`) // Ok{ value: Map{ "name": "Ann", "tags": ["admin"] } }
Json.parse("[1, 2.5]")                           // Ok{ value: [1.0, 2.5] }
```

Text that is not JSON is an `Err` that says what went wrong and where:

```ant
Json.parse("[1, 2,]") // Err{ error: Error{ message: "expected a value, found ']' at line 1, column 7" } }
```

## Scope

Each kind of JSON value has its own Ant type in `Json.Value`:

```ant
Json.parse("null")         // Ok{ value: null }
Json.parse("true")         // Ok{ value: true }
Json.parse("-12.5e1")      // Ok{ value: -125.0 }
Json.parse(`"hi"`)         // Ok{ value: "hi" }
Json.parse("[]")           // Ok{ value: [] }
Json.parse("{}")           // Ok{ value: Map{} }
Json.parse(`[1, "a", []]`) // Ok{ value: [1.0, "a", []] }
```

White space around the value is allowed, line breaks included:

```ant
Json.parse("\n  42  \n") // Ok{ value: 42.0 }
```

Keys keep the order of the text:

```ant
Json.parse(`{"b": 1, "a": 2}`) // Ok{ value: Map{ "b": 1.0, "a": 2.0 } }
```

A name may appear once in an object. The error points at the second one:

```ant
Json.parse(`{"a": 1, "b": 2, "a": 3}`) // Err{ error: Error{ message: "duplicate name \"a\" at line 1, column 18" } }
```

Escapes decode to the characters they name, a surrogate pair to one code point:

```ant
Json.parse(`"caf\u00e9 \ud83d\ude00"`) // Ok{ value: "café 😀" }
Json.parse(`"tab\there"`)              // Ok{ value: "tab\there" }
```

A surrogate escape outside a pair encodes no text and is an `Err`:

```ant
Json.parse(`"\ud800"`) // Err{ error: Error{ message: "unpaired surrogate \\ud800 at line 1, column 2" } }
```

A number too small for a `float64` rounds to zero; one too large is an `Err`:

```ant
Json.parse("1e-400") // Ok{ value: 0.0 }
Json.parse("1e400")  // Err{ error: Error{ message: "number 1e400 is too large for a float64" } }
```

Text that stops early, and text that goes on after the value, each have their own message:

```ant
Json.parse("")      // Err{ error: Error{ message: "unexpected end of JSON input" } }
Json.parse(`{"a":`) // Err{ error: Error{ message: "unexpected end of JSON input" } }
Json.parse("1 2")   // Err{ error: Error{ message: "unexpected text after the JSON value" } }
```

Any other error names the line and column:

```ant
Json.parse("{\n  \"on\": yes\n}") // Err{ error: Error{ message: "expected a value, found 'y' at line 2, column 9" } }
```

In a pipe, the text is the first argument:

```ant
"  [1, 2]  ".trim() | Json.parse() | Result.isOk() // true
```

## Applications

Read a JSON file and add context to either failure with `try` and
[Result.wrapErr](../Result/wrapErr.md):

<!-- test: file config.json
{"name": "api", "port": 8080}
-->

```ant
func loadConfig(path: string): Result[Json.Value, Error] {
  let text = try File.read(path)
  Json.parse(text) | Result.wrapErr("parse ${path}")
}

println(loadConfig("config.json"))
println(loadConfig("missing.json") is Err)
```

```text
Ok{ value: Map{ "name": "api", "port": 8080.0 } }
true
```

Decode a document into records. Narrow each value with `is`, convert numbers with
[Math.toInt](../Math/toInt.md), and pass a path along so an error says where it is:

```ant
type Server = { host: string, port: int }

func decodeServer(v: Json.Value, path: string): Result[Server, Error] {
  if !(v is Map) {
    return Result.fail("${path}: want an object")
  }
  let host = v["host"]
  if !(host is string) {
    return Result.fail("${path}.host: want a string")
  }
  let port = v["port"]
  let n = if port is float64 { Math.toInt(port) } else { null }
  if n == null {
    return Result.fail("${path}.port: want an integer")
  }
  Result.ok(Server{ host, port: n })
}

func decodeServers(text: string): Result[Array[Server], Error] {
  let doc = try Json.parse(text)
  if !(doc is Array) {
    return Result.fail("$: want an array")
  }
  let servers: Array[Server] = []
  for x in doc.entries() {
    servers.push(try decodeServer(x.value, "$[${x.index}]"))
  }
  Result.ok(servers)
}

println(decodeServers(`[{"host": "a", "port": 80}, {"host": "b", "port": 8080}]`))
println(decodeServers(`[{"host": "a", "port": 80}, {"host": "b", "port": 80.5}]`))
println(decodeServers(`[{"host": "a"}]`))
println(decodeServers(`{"host": "a"}`))
```

```text
Ok{ value: [Server{ host: "a", port: 80 }, Server{ host: "b", port: 8080 }] }
Err{ error: Error{ message: "$[1].port: want an integer" } }
Err{ error: Error{ message: "$[0].port: want an integer" } }
Err{ error: Error{ message: "$: want an array" } }
```

Sum the numbers of an array, skipping the other values:

```ant
func total(text: string): Result[float64, Error] {
  let doc = try Json.parse(text)
  if !(doc is Array) {
    return Result.fail("want an array")
  }
  var sum = 0.0
  for v in doc {
    if v is float64 {
      sum += v
    }
  }
  Result.ok(sum)
}

total(`[1, "two", 3.5, null]`) // Ok{ value: 4.5 }
```

## Properties and relations

[Json.stringify](stringify.md) writes text that parses back to the same value:

```ant
let v: Json.Value = Map{ "a": [1.0, true, null], "b": "x" }
let back = v | Json.stringify() | Json.parse()
assert(back.unwrap() == v)
```

Parsing and writing again drops the white space and keeps the key order:

```ant
Json.parse("{ \"b\" : [ 1 , 2 ] ,\n \"a\" : true }").unwrap() | Json.stringify() // "{\"b\":[1.0,2.0],\"a\":true}"
```

## Possible issues

The result is a `Json.Value`, a union. Narrow it with `is` before you index it:

```ant
let doc = Json.parse(`{"a": 1}`).unwrap()
let a = doc["a"] // error: Json.Value cannot be indexed
```

```ant
let doc = Json.parse(`{"a": 1}`).unwrap()
let a = if doc is Map { doc["a"] } else { null }
a // 1.0
```

Numbers are `float64`, so comparing one with an `int` is a compile error. Compare with a `float64`,
or convert with [Math.toInt](../Math/toInt.md):

```ant
let n = Json.parse("3").unwrap()
let three = n == 3 // error: Json.Value and int cannot be compared with ==: neither is assignable to the other
```

```ant
let n = Json.parse("3").unwrap()
n == 3.0                                        // true
if n is float64 { Math.toInt(n) } else { null } // 3
```

An integer beyond 2^53 loses precision, as every `float64` does:

```ant
Json.parse("9007199254740993") // Ok{ value: 9007199254740992.0 }
```

A missing key and a key whose value is `null` both index to `null`. Use [Map.has](../Map/has.md) to
tell them apart:

```ant
let doc = Json.parse(`{"a": null}`).unwrap()
let m: Map[string, Json.Value] = if doc is Map { doc } else { Map{} }
m["a"]     // null
m["b"]     // null
m.has("a") // true
m.has("b") // false
```

Common extensions of JSON are not JSON, and each is an `Err`:

```ant
Json.parse("{'a': 1}")   // Err{ error: Error{ message: "expected a string key, found '\\'' at line 1, column 2" } }
Json.parse("// note\n1") // Err{ error: Error{ message: "expected a value, found '/' at line 1, column 1" } }
Json.parse("NaN")        // Err{ error: Error{ message: "expected a value, found 'N' at line 1, column 1" } }
```

`Result.unwrap` on malformed text panics. Use `try` or test the result where the text comes from
outside:

```ant
Json.parse("{").unwrap() // panic: Result.unwrap on an Err: Error{ message: "unexpected end of JSON input" }
```

## See also

[Json.stringify](stringify.md) · [Json.Value](types/Value.md) · [Math.toInt](../Math/toInt.md) ·
[Result.wrapErr](../Result/wrapErr.md) · [try](../../language/keywords/try.md) ·
[is](../../language/operators/is.md) · [Errors guide](../../guide/errors.md)
