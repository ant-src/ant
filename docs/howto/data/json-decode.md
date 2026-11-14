<!-- status: reviewed -->

# Decode JSON into records

Turn a JSON document into your own record types, and report where the input does not fit.

1. Parse the text with [Json.parse](../../lib/Json/parse.md). It returns a
   [Json.Value](../../lib/Json/types/Value.md): a tree of `null`, `bool`, `float64`, `string`,
   arrays, and maps.
2. Write a decoder for each record type. Narrow each value with
   [`is`](../../language/operators/is.md), convert numbers with
   [Math.toInt](../../lib/Math/toInt.md), and build the record.
3. Return an `Err` from [Result.fail](../../lib/Result/fail.md) for input that does not fit, and
   pass it up with [`try`](../../language/keywords/try.md).
4. Pass a path along, such as `$.servers[1].port`, so the error says where the input is wrong.

## Write a helper per JSON kind

Each helper checks one kind of value and returns it narrowed, or an `Err` that names the path. Put
them in a module of their own, `decode.ant`:

```ant [decode.ant]
export func asObject(v: Json.Value, path: string): Result[Map[string, Json.Value], Error] {
  if !(v is Map) {
    return Result.fail("${path}: want an object")
  }
  Result.ok(v)
}

export func asString(v: Json.Value, path: string): Result[string, Error] {
  if !(v is string) {
    return Result.fail("${path}: want a string")
  }
  Result.ok(v)
}

export func asBool(v: Json.Value, path: string): Result[bool, Error] {
  if !(v is bool) {
    return Result.fail("${path}: want a boolean")
  }
  Result.ok(v)
}

export func asFloat(v: Json.Value, path: string): Result[float64, Error] {
  if !(v is float64) {
    return Result.fail("${path}: want a number")
  }
  Result.ok(v)
}

export func asInt(v: Json.Value, path: string): Result[int, Error] {
  let n = if v is float64 { Math.toInt(v) } else { null }
  if n == null {
    return Result.fail("${path}: want an integer")
  }
  Result.ok(n)
}

export func asArray[T](
  v: Json.Value,
  path: string,
  decode: func(Json.Value, string): Result[T, Error],
): Result[Array[T], Error] {
  if !(v is Array) {
    return Result.fail("${path}: want an array")
  }
  let items: Array[T] = []
  for { index, value } in v.entries() {
    items.push(try decode(value, "${path}[${index}]"))
  }
  Result.ok(items)
}

export func asMap[T](
  v: Json.Value,
  path: string,
  decode: func(Json.Value, string): Result[T, Error],
): Result[Map[string, T], Error] {
  let obj = try asObject(v, path)
  let items: Map[string, T] = Map{}
  for { key, value } in obj {
    items.set(key, try decode(value, "${path}.${key}"))
  }
  Result.ok(items)
}
```

`asBool`, `asFloat` and `asString` cover the remaining JSON kinds. Every JSON number is a `float64`,
so `asFloat` returns it as is and `asInt` converts it with `Math.toInt`, which returns `null` for a
fraction or a number out of the `int` range. `asArray` takes the decoder for its elements and adds
each element's index to the path; [Array.entries](../../lib/Array/entries.md) gives the index and
[Array.push](../../lib/Array/push.md) collects the results. `asMap` does the same for an object with
arbitrary keys: it loops over the entries, adds each key to the path, and collects the results with
[Map.set](../../lib/Map/set.md).

## Decode one record

Read each field of the object by its key and decode it with a helper. A missing key reads as `null`,
so the same check rejects it. Keys the record does not have are ignored:

```ant
import "./decode.ant" { asInt, asObject, asString }

type Server = { host: string, port: int }

func decodeServer(v: Json.Value, path: string): Result[Server, Error] {
  let obj = try asObject(v, path)
  let host = try asString(obj["host"], "${path}.host")
  let port = try asInt(obj["port"], "${path}.port")
  Result.ok(Server{ host, port })
}

for input in [
  `{"host": "a.example", "port": 8080, "zone": "eu"}`,
  `{"host": "a.example", "port": "8080"}`,
  `{"port": 8080}`,
  `[1, 2]`,
] {
  println(decodeServer(try Json.parse(input), "$"))
}
```

```text
Ok{ value: Server{ host: "a.example", port: 8080 } }
Err{ error: Error{ message: "$.port: want an integer" } }
Err{ error: Error{ message: "$.host: want a string" } }
Err{ error: Error{ message: "$: want an object" } }
```

## Decode nested arrays and records

Decode an array field with `asArray`, passing the decoder of its elements: `decodeServer` for an
array of records, `asString` for an array of strings. Decode an object with arbitrary keys into a
`Map` with `asMap` the same way. The path grows at each level, so an error deep in the document
names its place:

```ant
import "./decode.ant" { asArray, asInt, asMap, asObject, asString }

type Server = { host: string, port: int }
type Config = { name: string, servers: Array[Server], limits: Map[string, int] }

func decodeServer(v: Json.Value, path: string): Result[Server, Error] {
  let obj = try asObject(v, path)
  let host = try asString(obj["host"], "${path}.host")
  let port = try asInt(obj["port"], "${path}.port")
  Result.ok(Server{ host, port })
}

func decodeConfig(v: Json.Value, path: string): Result[Config, Error] {
  let obj = try asObject(v, path)
  let name = try asString(obj["name"], "${path}.name")
  let servers = try asArray(obj["servers"], "${path}.servers", decodeServer)
  let limits = try asMap(obj["limits"], "${path}.limits", asInt)
  Result.ok(Config{ name, servers, limits })
}

func parseConfig(input: string): Result[Config, Error] {
  decodeConfig(try Json.parse(input), "$")
}

println(parseConfig(`{"name": "prod", "servers": [{"host": "a", "port": 80}], "limits": {"cpu": 2, "mem": 512}}`))
println(parseConfig(`{"name": "prod", "servers": [{"host": "a", "port": 80}, {"host": "b", "port": 80.5}], "limits": {}}`))
println(parseConfig(`{"name": "prod", "servers": [], "limits": {"cpu": "two"}}`))
println(parseConfig(`{"name": "prod", "servers": {}}`))
println(parseConfig(`{"name": "prod"`))
```

```text
Ok{ value: Config{ name: "prod", servers: [Server{ host: "a", port: 80 }], limits: Map{ "cpu": 2, "mem": 512 } } }
Err{ error: Error{ message: "$.servers[1].port: want an integer" } }
Err{ error: Error{ message: "$.limits.cpu: want an integer" } }
Err{ error: Error{ message: "$.servers: want an array" } }
Err{ error: Error{ message: "unexpected end of JSON input" } }
```

## Fill in optional fields

Give an optional field a default in the record type, and build the record without it. Then test with
[Map.has](../../lib/Map/has.md) whether the object has the key, and set the field only when it does:

```ant
import "./decode.ant" { asArray, asBool, asInt, asObject, asString }

type Server = {
  host: string,
  port: int = 8080,
  enabled: bool = true,
  tags: Array[string] = [],
}

func decodeServer(v: Json.Value, path: string): Result[Server, Error] {
  let obj = try asObject(v, path)
  let server = Server{ host: try asString(obj["host"], "${path}.host") }
  if obj.has("port") {
    server.port = try asInt(obj["port"], "${path}.port")
  }
  if obj.has("enabled") {
    server.enabled = try asBool(obj["enabled"], "${path}.enabled")
  }
  if obj.has("tags") {
    server.tags = try asArray(obj["tags"], "${path}.tags", asString)
  }
  Result.ok(server)
}

for input in [
  `{"host": "a.example"}`,
  `{"host": "b.example", "port": 9000, "enabled": false, "tags": ["eu", "db"]}`,
  `{"host": "c.example", "tags": ["eu", 1]}`,
  `{"host": "d.example", "port": null}`,
] {
  println(decodeServer(try Json.parse(input), "$"))
}
```

```text
Ok{ value: Server{ host: "a.example", port: 8080, enabled: true, tags: [] } }
Ok{ value: Server{ host: "b.example", port: 9000, enabled: false, tags: ["eu", "db"] } }
Err{ error: Error{ message: "$.tags[1]: want a string" } }
Err{ error: Error{ message: "$.port: want an integer" } }
```

A key whose value is JSON `null` is present, so `"port": null` is an error here. Test
`obj["port"] != null` instead of `obj.has("port")` to treat it as missing.

That test also gives a nullable field: declare it with the type `string | null` and the default
`null`, and set it only when the value is not `null`. A missing key and a JSON `null` both leave it
`null`:

```ant
import "./decode.ant" { asObject, asString }

type User = { name: string, nickname: string | null = null }

func decodeUser(v: Json.Value, path: string): Result[User, Error] {
  let obj = try asObject(v, path)
  let user = User{ name: try asString(obj["name"], "${path}.name") }
  if obj["nickname"] != null {
    user.nickname = try asString(obj["nickname"], "${path}.nickname")
  }
  Result.ok(user)
}

for input in [
  `{"name": "Ann", "nickname": "annie"}`,
  `{"name": "Bob", "nickname": null}`,
  `{"name": "Cy"}`,
  `{"name": "Di", "nickname": 7}`,
] {
  println(decodeUser(try Json.parse(input), "$"))
}
```

```text
Ok{ value: User{ name: "Ann", nickname: "annie" } }
Ok{ value: User{ name: "Bob", nickname: null } }
Ok{ value: User{ name: "Cy", nickname: null } }
Err{ error: Error{ message: "$.nickname: want a string" } }
```

## Read the document from a file

Read the file with [File.read](../../lib/File/read.md), decode it, and add the file name to either
error with [Result.wrapErr](../../lib/Result/wrapErr.md):

<!-- test: file servers.json
{"host": "a.example", "port": "eighty"}
-->

```ant
import "./decode.ant" { asInt, asObject, asString }

type Server = { host: string, port: int }

func decodeServer(v: Json.Value, path: string): Result[Server, Error] {
  let obj = try asObject(v, path)
  let host = try asString(obj["host"], "${path}.host")
  let port = try asInt(obj["port"], "${path}.port")
  Result.ok(Server{ host, port })
}

func loadServer(path: string): Result[Server, Error] {
  let input = try File.read(path)
  let doc = try (Json.parse(input) | Result.wrapErr(path))
  decodeServer(doc, "$") | Result.wrapErr(path)
}

println(loadServer("servers.json"))
```

```text
Err{ error: Error{ message: "servers.json: $.port: want an integer" } }
```

## The whole program

This script reads `config.json` and prints one line per server. Run on a file that does not fit, it
writes the error to standard error and exits with status 1:

<!-- test: file config.json
{
  "name": "prod",
  "servers": [
    {"host": "a.example", "tags": ["eu"]},
    {"host": "b.example", "port": 9000}
  ]
}
-->

```ant
import "./decode.ant" { asArray, asInt, asObject, asString }

type Server = { host: string, port: int = 8080, tags: Array[string] = [] }
type Config = { name: string, servers: Array[Server] }

func decodeServer(v: Json.Value, path: string): Result[Server, Error] {
  let obj = try asObject(v, path)
  let server = Server{ host: try asString(obj["host"], "${path}.host") }
  if obj.has("port") {
    server.port = try asInt(obj["port"], "${path}.port")
  }
  if obj.has("tags") {
    server.tags = try asArray(obj["tags"], "${path}.tags", asString)
  }
  Result.ok(server)
}

func decodeConfig(v: Json.Value, path: string): Result[Config, Error] {
  let obj = try asObject(v, path)
  let name = try asString(obj["name"], "${path}.name")
  let servers = try asArray(obj["servers"], "${path}.servers", decodeServer)
  Result.ok(Config{ name, servers })
}

func loadConfig(path: string): Result[Config, Error] {
  let input = try File.read(path)
  let doc = try (Json.parse(input) | Result.wrapErr(path))
  decodeConfig(doc, "$") | Result.wrapErr(path)
}

let config = try loadConfig("config.json")
for server in config.servers {
  println("${config.name}: ${server.host}:${server.port} ${server.tags}")
}
```

```text
prod: a.example:8080 ["eu"]
prod: b.example:9000 []
```

## Decode many types with reflection

A decoder per type keeps every rule in plain sight, and the type checker checks each record literal.
When a program reads many record types, write one decoder that walks the target type instead, with
[Reflect.shape](../../lib/Reflect/shape.md), [Reflect.makeRecord](../../lib/Reflect/makeRecord.md),
and [Reflect.makeScalar](../../lib/Reflect/makeScalar.md).
[Types at run time](../../guide/runtime-types.md) builds one.

## See also

[Json.parse](../../lib/Json/parse.md) · [Json.Value](../../lib/Json/types/Value.md) ·
[Math.toInt](../../lib/Math/toInt.md) · [Map.has](../../lib/Map/has.md) ·
[Map.set](../../lib/Map/set.md) · [Array.entries](../../lib/Array/entries.md) ·
[Array.push](../../lib/Array/push.md) · [Result.ok](../../lib/Result/ok.md) ·
[Result.fail](../../lib/Result/fail.md) · [Result.wrapErr](../../lib/Result/wrapErr.md) ·
[File.read](../../lib/File/read.md) · [try](../../language/keywords/try.md) ·
[is](../../language/operators/is.md) · [import](../../language/keywords/import.md) ·
[Records](../../guide/records.md) · [Errors](../../guide/errors.md) ·
[Encode values as JSON](json-encode.md)
