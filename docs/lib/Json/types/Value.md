<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Json.Value

```ant signature
type Json.Value = null | bool | float64 | string | Array[Json.Value] | Map[string, Json.Value]
```

Json.Value is a JSON value: null, a boolean, a number, a string, an array, or an object. Every
number is a float64; [`Math.toInt`](../../Math/toInt.md) converts one to an int when it is whole and
in range.

<!-- end generated -->

## Details

- A `Json.Value` cannot be indexed or iterated until `is` narrows it. `is` tests `Array` and `Map`
  without type arguments, and the elements and values inside are again `Json.Value`.
- A JSON number is a `float64`, so compare it with `1.0`, not `1`, or convert it with `Math.toInt`.
- Reading a key of an object gives `null` both when the key is missing and when its value is JSON
  `null`. `Map.has` tells the two apart.
- An `Array[string]` or a `Map[string, float64]` is not a `Json.Value`, since type arguments must
  match exactly. A literal written where a `Json.Value` is expected builds one directly.
- A `Json.Value` is for reading JSON whose shape is not known. [`Json.stringify`](../stringify.md)
  takes any value, so typed data is written directly, with no conversion.

## Basic examples

`Json.parse` returns a `Json.Value` inside a `Result`:

```ant
Json.parse("{\"name\": \"Ann\", \"age\": 31}") // Ok{ value: Map{ "name": "Ann", "age": 31.0 } }
Json.parse("[true, null, \"x\"]")              // Ok{ value: [true, null, "x"] }
```

Narrow a `Json.Value` with `is` to use it:

```ant
func describe(v: Json.Value): string {
  if v is null {
    "null"
  } else if v is bool {
    "boolean ${v}"
  } else if v is float64 {
    "number ${v}"
  } else if v is string {
    "string ${v}"
  } else if v is Array {
    "array of ${v.length()}"
  } else if v is Map {
    "object with ${v.size()} keys"
  }
}

describe(Result.unwrap(Json.parse("2.5")))                  // "number 2.5"
describe(Result.unwrap(Json.parse("[1, 2, 3]")))            // "array of 3"
describe(Result.unwrap(Json.parse("{\"a\": 1, \"b\": 2}"))) // "object with 2 keys"
```

The chain in `describe` has no final `else`: after six tests no member is left, so the checker knows
it is complete.

## Scope

Each of the six members is a `Json.Value`:

```ant
let values: Array[Json.Value] = [null, true, 1.5, "text", [1.0, "a"], Map{ "k": false }]
values.length() // 6
values[4]       // [1.0, "a"]
values[5]       // Map{ "k": false }
```

Arrays and objects nest to any depth, and each level is a `Json.Value` again:

```ant
let doc: Json.Value = Map{
  "name": "ant",
  "versions": [Map{ "tag": "v0.1.0", "stable": true }],
  "license": null,
}
Json.stringify(doc) // "{\"name\":\"ant\",\"versions\":[{\"tag\":\"v0.1.0\",\"stable\":true}],\"license\":null}"
```

After `v is Map`, `v` is a `Map[string, Json.Value]`, so its keys, values, and dot calls are
available:

```ant
func keys(v: Json.Value): Array[string] {
  if v is Map { v.keys() } else { [] }
}

keys(Result.unwrap(Json.parse("{\"b\": 2, \"a\": 1}"))) // ["b", "a"]
keys(Result.unwrap(Json.parse("[1, 2]")))               // []
```

After `v is Array`, `v` is an `Array[Json.Value]`, and every element needs its own test:

```ant
func countNumbers(v: Json.Value): int {
  if v is Array { v.filter({ x is float64 }).length() } else { 0 }
}

countNumbers(Result.unwrap(Json.parse("[1, \"two\", 3]"))) // 2
```

A missing key and a stored `null` read the same; `has` tells them apart:

```ant
let obj: Map[string, Json.Value] = Map{ "a": null }
obj["a"]     // null
obj["b"]     // null
obj.has("a") // true
obj.has("b") // false
```

A whole number converts to an `int` with `Math.toInt`, which gives `null` for a fraction:

```ant
func port(v: Json.Value): int | null {
  if v is float64 { Math.toInt(v) } else { null }
}

port(Result.unwrap(Json.parse("8080")))   // 8080
port(Result.unwrap(Json.parse("80.5")))   // null
port(Result.unwrap(Json.parse("\"80\""))) // null
```

## Applications

Decode a document into records, narrowing each value and passing a path along so an error says where
it is:

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

func decodeServers(v: Json.Value): Result[Array[Server], Error] {
  let list = if v is Map { v["servers"] } else { null }
  if !(list is Array) {
    return Result.fail("$.servers: want an array")
  }
  let servers: Array[Server] = []
  for { index, value } in list.entries() {
    servers.push(try decodeServer(value, "$.servers[${index}]"))
  }
  Result.ok(servers)
}

let good = "{\"servers\": [{\"host\": \"a\", \"port\": 80}, {\"host\": \"b\", \"port\": 8080}]}"
let bad = "{\"servers\": [{\"host\": \"a\", \"port\": 80}, {\"host\": \"b\", \"port\": 80.5}]}"
for text in [good, bad] {
  println(Json.parse(text) | Result.unwrap() | decodeServers())
}
```

```text
Ok{ value: [Server{ host: "a", port: 80 }, Server{ host: "b", port: 8080 }] }
Err{ error: Error{ message: "$.servers[1].port: want an integer" } }
```

Rewrite a document of unknown shape, here hiding every `password` at any depth:

```ant
func redact(v: Json.Value): Json.Value {
  if v is Array {
    return v.map(redact)
  }
  if !(v is Map) {
    return v
  }
  v.map({ if x.key == "password" { "***" } else { redact(x.value) } })
}

let text = "{\"user\": \"ann\", \"password\": \"hunter2\", \"keys\": [{\"password\": \"x\"}]}"
let doc = try Json.parse(text)
println(redact(doc) | Json.stringify())
```

```text
{"user":"ann","password":"***","keys":[{"password":"***"}]}
```

Collect every string in a document, in order:

```ant
func strings(v: Json.Value): Array[string] {
  if v is string {
    [v]
  } else if v is Array {
    v.flatMap(strings)
  } else if v is Map {
    v.values().flatMap(strings)
  } else {
    []
  }
}

strings(Result.unwrap(Json.parse("{\"a\": [\"x\", 1, {\"b\": \"y\"}], \"c\": \"z\"}"))) // ["x", "y", "z"]
```

## Properties and relations

A `Json.Value` without `NaN` or an infinity survives `Json.stringify` and `Json.parse` unchanged:

```ant
let v: Json.Value = Map{ "a": [1.5, null, "é"], "b": Map{ "c": true } }
let back = v | Json.stringify() | Json.parse() | Result.unwrap()
assert(back == v)
```

Every parsed number is a `float64`, whole or not:

```ant
assert(Result.unwrap(Json.parse("1")) == 1.0)
assert(Result.unwrap(Json.parse("1e3")) == 1000.0)
```

## Possible issues

A `Json.Value` cannot be indexed before it is narrowed:

```ant
let v = Result.unwrap(Json.parse("{\"name\": \"Ann\"}"))
let name = v["name"] // error: Json.Value cannot be indexed
```

`is` takes `Array` or `Map` alone, without type arguments. Test the elements one by one instead:

```ant
let v: Json.Value = ["a"]
let ok = v is Array[string]
// error: 'is' tests a kind or a tag, written without type arguments: write is Array, then test the elements
```

A JSON number is never an `int`, so an `int` literal is not a `Json.Value`, and comparing with one
does not compile. Write `1.0`, or convert with `Math.toInt`:

```ant
let v: Json.Value = [1, 2.0] // error: cannot use int where Json.Value is expected
let one = Result.unwrap(Json.parse("1")) == 1
// error: Json.Value and int cannot be compared with ==: neither is assignable to the other
```

An array or map already typed with other type arguments is not a `Json.Value`. Spread an array into
a new one where a `Json.Value` is expected:

```ant
let tags = ["a", "b"]
let v: Json.Value = tags // error: cannot use Array[string] where Json.Value is expected
```

```ant
let tags = ["a", "b"]
let v: Json.Value = [...tags]
v // ["a", "b"]
```

A map literal has no spread. Map the entries of a typed map into a new one instead:

```ant
let prices = Map{ "tea": 2.5 }
let v: Json.Value = prices.map(func(e: Entry[string, float64]): Json.Value { e.value })
v // Map{ "tea": 2.5 }
```

A `Set` is not a member, though `Json.stringify` writes one as an array:

```ant
let s: Json.Value = Set.of(["a"]) // error: cannot use Set[string] where Json.Value is expected
```

Integers beyond 2^53 lose precision, since every number is a `float64`:

```ant
Json.parse("9007199254740993") // Ok{ value: 9007199254740992.0 }
```

Building a `Json.Value` by hand to write out typed data is not needed. `Json.stringify` writes a
record as an object and a `Set` as an array directly:

```ant
type Server = { host: string, port: int }
Json.stringify(Server{ host: "a", port: 80 }) // "{\"host\":\"a\",\"port\":80}"
```

## See also

[Json.parse](../parse.md), [Json.stringify](../stringify.md), [Math.toInt](../../Math/toInt.md),
[Map.has](../../Map/has.md), [is](../../../language/operators/is.md),
[Unions](../../../guide/unions.md)
