<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Json.stringify

```ant signature
func Json.stringify[T](value: T, indent: string = ""): string
```

Json.stringify returns the compact JSON text of value. Records become objects without their tag,
sets arrays, and time values strings. It panics with "cyclic value" on a value that contains itself,
and panics on what has no JSON form: NaN and the infinities, Bytes, which a program writes with
[`Encoding.toBase64`](../Encoding/toBase64.md) instead, a function, void, a type, or a reference
such as a channel or a Random. A map key must be a string or a named primitive over one, which is
written without its tag; any other key panics, as do two keys that give one property name.

With an empty indent the text is one line. Otherwise each member of an array or object is on a line
of its own, indented by indent once per level, ": " follows a property name, and an empty array or
object stays \[\] or \{\}; the text ends without a newline. It panics on an indent that is not all
spaces and tabs.

<!-- end generated -->

## Details

- Numbers are written as `String.from` writes them: an `int` as `31`, a `float64` as `1.0` or
  `1e+21`.
- Each panic message starts `Json.stringify: `. `Html` has no JSON form either; render it with
  `String.from` first.

## Basic examples

Write a map of arrays as compact JSON:

```ant
Json.stringify(Map{ "a": [1, 2] }) // "{\"a\":[1,2]}"
```

Write a record as an object:

```ant
type User = { name: string, age: int }
Json.stringify(User{ name: "Ann", age: 31 }) // "{\"name\":\"Ann\",\"age\":31}"
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

### Scalars

A string, a number, a `bool`, or `null` on its own is a JSON text too:

```ant
Json.stringify("ant") // "\"ant\""
Json.stringify(42)    // "42"
Json.stringify(2.5)   // "2.5"
Json.stringify(true)  // "true"
Json.stringify(null)  // "null"
```

### Numbers

A `float64` keeps its `.0`, and a large one is written with an exponent:

```ant
Json.stringify([1, 1.0, -0.0, 1e21]) // "[1,1.0,-0.0,1e+21]"
```

### Strings

A string is quoted, with quotes, backslashes, and control characters escaped:

```ant
println(Json.stringify("say \"hi\"\tC:\\ é 😀"))
println(Json.stringify("\u{1}"))
```

```text
"say \"hi\"\tC:\\ é 😀"
"\u0001"
```

### Empty containers

```ant
let names: Array[string] = []
let counts: Map[string, int] = Map{}
Json.stringify(names)  // "[]"
Json.stringify(counts) // "{}"
```

### Records with `null` fields

A `null` field is written as `null`; it is never left out:

```ant
type Contact = { name: string, email: string | null }
Json.stringify(Contact{ name: "Bo", email: null }) // "{\"name\":\"Bo\",\"email\":null}"
```

### Nested records and unions

A union member is written as the value it holds. Nothing in the text says which member it was:

```ant
type Circle = { radius: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
type Drawing = { title: string, shapes: Array[Shape] }
let d = Drawing{ title: "demo", shapes: [Circle{ radius: 1.0 }, Square{ side: 2.0 }] }
println(Json.stringify(d))
```

```text
{"title":"demo","shapes":[{"radius":1.0},{"side":2.0}]}
```

### Named primitive types

A named primitive is written without its tag, as a value and as a map key:

```ant
type UserId = string
Json.stringify(UserId("u-1"))               // "\"u-1\""
Json.stringify(Map{ UserId("u-1"): "Ann" }) // "{\"u-1\":\"Ann\"}"
```

### Sets

A set becomes an array of its members, in the set's order:

```ant
Json.stringify(Set.of(["b", "a", "c"])) // "[\"b\",\"a\",\"c\"]"
```

### Time values

A time value becomes its string:

```ant
let start = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
Json.stringify(Map{ "start": start }) // "{\"start\":\"2026-10-22T14:03:05Z\"}"
Json.stringify(90s)                   // "\"PT90S\""
```

### Values of `Json.Value`

A tree from `Json.parse` writes back as JSON:

```ant
let tree = Json.parse("{ \"a\": [1, true, null] }").unwrap()
Json.stringify(tree) // "{\"a\":[1.0,true,null]}"
```

### Calls through pipes

A dot call looks in the receiver's namespace, such as `Array`, so it does not reach
`Json.stringify`. A pipe does:

```ant
[1, 2, 3] | Array.map({ x * 10 }) | Json.stringify() // "[10,20,30]"
```

## Named arguments

### indent

With the default `""`, the text is one line with nothing between its tokens:

```ant
Json.stringify(Map{ "a": [1, 2] }, indent: "") // "{\"a\":[1,2]}"
```

Any run of spaces or tabs puts each member on its own line, indented once per level, with `": "`
after each property name. An empty array or object stays on one line:

```ant
let doc = Map{ "ids": [1, 2], "none": [] }
println(Json.stringify(doc, indent: "    "))
```

```text
{
    "ids": [
        1,
        2
    ],
    "none": []
}
```

A tab indents too. The text ends without a newline, so the `.` is right after the brace:

```ant
print(Json.stringify(Map{ "ok": true }, indent: "\t"))
println(".")
```

```text
{
	"ok": true
}.
```

## Applications

### Write a settings file

Add the final newline a text file needs yourself:

```ant
type Settings = { theme: string, fontSize: int, recent: Array[string] }

func save(path: string, s: Settings): Result[void, Error] {
  File.write(path, Json.stringify(s, indent: "  ") + "\n")
}

let result = save("settings.json", Settings{ theme: "dark", fontSize: 14, recent: ["a.ant"] })
assert(result is Ok)
print(File.read("settings.json").unwrap())
```

```text
{
  "theme": "dark",
  "fontSize": 14,
  "recent": [
    "a.ant"
  ]
}
```

### Log events as JSON Lines

Compact text has no line breaks, so each value takes exactly one line:

```ant
type Event = { level: string, message: string, ms: int }

let events = [
  Event{ level: "info", message: "started", ms: 0 },
  Event{ level: "warn", message: "slow \"query\"", ms: 812 },
]
for e in events {
  println(Json.stringify(e))
}
```

```text
{"level":"info","message":"started","ms":0}
{"level":"warn","message":"slow \"query\"","ms":812}
```

### Hand data to a script in a page

An attribute escapes the text, so the JSON reaches the script intact:

```ant
let points = [[0, 1], [1, 3]]
println(<div id="chart" data-points={Json.stringify(points)}></div>)
```

```text
<div id="chart" data-points="[[0,1],[1,3]]"></div>
```

### Keep the member of a union

The text drops the tag, so a reader cannot tell the members apart. Add a field that names the
member:

```ant
type Circle = { kind: string = "circle", radius: float64 }
type Square = { kind: string = "square", side: float64 }
type Shape = Circle | Square
let shapes: Array[Shape] = [Circle{ radius: 1.0 }, Square{ side: 2.0 }]
println(Json.stringify(shapes))
```

```text
[{"kind":"circle","radius":1.0},{"kind":"square","side":2.0}]
```

### Send binary data

Encode `Bytes` as a string first:

```ant
type Upload = { name: string, data: string }
let bytes = "hi".toBytes()
let upload = Upload{ name: "hi.txt", data: Encoding.toBase64(bytes) }
println(Json.stringify(upload))
```

```text
{"name":"hi.txt","data":"aGk="}
```

## Properties and relations

`Json.parse` reads back what `Json.stringify` writes from a `Json.Value`, with or without `indent`:

```ant
let v: Json.Value = Map{ "a": [1.0, -2.5, null], "b": Map{ "c": "é\n" }, "d": true }
assert(Json.parse(Json.stringify(v)).unwrap() == v)
assert(Json.parse(Json.stringify(v, indent: "\t")).unwrap() == v)
```

A record writes the same text as a map from its field names to its field values:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
assert(Json.stringify(p) == Json.stringify(Map{ "x": 1.0, "y": 2.0 }))
```

A named primitive writes the same text as its payload, and a set the same as an array of its
members:

```ant
type Sku = string
assert(Json.stringify(Sku("a-1")) == Json.stringify("a-1"))
assert(Json.stringify(Set.of([3, 1])) == Json.stringify([3, 1]))
```

A value with the same object in two places is not cyclic, and is written twice:

```ant
let shared = [1, 2]
Json.stringify([shared, shared]) // "[[1,2],[1,2]]"
```

## Possible issues

### Integers read back as floats

`Json.parse` gives every number as a `float64`, so a round trip does not give back an `int`:

```ant
let text = Json.stringify(Map{ "n": 1 })
text                      // "{\"n\":1}"
Json.parse(text).unwrap() // Map{ "n": 1.0 }
```

### `NaN` and the infinities

They have no JSON form:

```ant
Json.stringify([1.0, 0.0 / 0.0]) // panic: Json.stringify: NaN has no JSON form
Json.stringify(1.0 / 0.0)        // panic: Json.stringify: Infinity has no JSON form
```

### Map keys that are not strings

A JSON property name is a string, so a map with `int` keys panics. Convert the keys yourself:

```ant
let scores = Map{ 1: "gold", 2: "silver" }
Json.stringify(scores) // panic: Json.stringify: a map key must be a string, got int
```

```ant
let scores = Map{ 1: "gold", 2: "silver" }
let byName = scores
  | Map.entries()
  | Array.map({ Entry{ key: String.from(x.key), value: x.value } })
  | Map.fromEntries()
println(Json.stringify(byName))
```

```text
{"1":"gold","2":"silver"}
```

Two keys that give one property name panic too:

```ant
type UserId = string
let m: Map[UserId | string, int] = Map{ UserId("a"): 1, "a": 2 }
Json.stringify(m)
// panic: Json.stringify: map keys UserId("a") and "a" both give the property name "a"
```

### `Bytes` and `Html`

Neither has a JSON form. Encode `Bytes` with `Encoding.toBase64`, and render `Html` with
`String.from`:

```ant
Json.stringify("hi".toBytes())
// panic: Json.stringify: Bytes has no JSON form; write Encoding.toBase64 of it
Json.stringify(<b>hi</b>)
// panic: Json.stringify: Html has no JSON form; write String.from of it
Json.stringify(String.from(<b>hi</b>)) // "\"<b>hi</b>\""
```

### Functions, types, and references

A function, `void`, a type, or a reference such as a `Mutex` panics, wherever it is in the value:

```ant
type Job = { name: string, run: func(): void }
Json.stringify(Job{ name: "a", run: {} })  // panic: Json.stringify: a function has no JSON form
Json.stringify([type(int)])                // panic: Json.stringify: a type has no JSON form
Json.stringify(Map{ "lock": Mutex.new() }) // panic: Json.stringify: Mutex has no JSON form
```

### Cyclic values

A value that contains itself panics, where `String.from` would write `...`:

```ant
type Node = { name: string, next: Node | null }
let n = Node{ name: "a", next: null }
n.next = n
Json.stringify(n) // panic: Json.stringify: cyclic value
```

### An indent that is not white space

`indent` is the text of one level, not a count or a prefix:

```ant
Json.stringify([1], indent: "--")
// panic: Json.stringify needs an indent of spaces and tabs, got "--"
```

## See also

[Json.parse](parse.md) · [Json.Value](types/Value.md) · [String.from](../String/from.md) ·
[Encoding.toBase64](../Encoding/toBase64.md) · [Reflect](../Reflect/index.md) ·
[File.write](../File/write.md)
