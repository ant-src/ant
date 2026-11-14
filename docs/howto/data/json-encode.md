<!-- status: reviewed -->

# Encode values as JSON

Turn the records, arrays, and maps of a program into JSON text for a file, a log, or another
program.

1. Describe the data with records. Each field becomes a property of the same name, in the order the
   record type lists its fields.
2. Call [Json.stringify](../../lib/Json/stringify.md) on the value. It returns one line of text.
3. Pass `indent:` for text that people read.
4. Convert what has no JSON form, such as `Bytes` or a map with `int` keys, before you encode.

## Encode a record

This script saves a report to `report.json`. `Json.stringify` writes nested records and arrays as
nested objects and arrays, and a `null` field as `null`. The text ends without a newline, so add one
before you write it with [File.write](../../lib/File/write.md):

```ant
type Item = { sku: string, quantity: int, price: float64 }
type Report = { customer: string, coupon: string | null, items: Array[Item] }

let report = Report{
  customer: "Ann",
  coupon: null,
  items: [Item{ sku: "tea", quantity: 2, price: 4.5 }, Item{ sku: "mug", quantity: 1, price: 12.0 }],
}
try File.write("report.json", Json.stringify(report, indent: "  ") + "\n")
print(try File.read("report.json"))
```

```text
{
  "customer": "Ann",
  "coupon": null,
  "items": [
    {
      "sku": "tea",
      "quantity": 2,
      "price": 4.5
    },
    {
      "sku": "mug",
      "quantity": 1,
      "price": 12.0
    }
  ]
}
```

An `int` is written as `2` and a `float64` keeps its `.0`, as `String.from` writes them. The
`indent` is the text of one level: spaces or a tab, such as `"\t"`.

## Write one value per line

Leave out `indent` for compact text. It has no line breaks, so each value takes one line, as JSON
Lines needs:

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

## Encode a map, a set, or a time value

A map with `string` keys becomes an object, a set an array of its members, and a time value its
string:

```ant
let start = try Time.parseInstant("2026-10-22T14:03:05Z")
let stock = Map{ "tea": 12, "mug": 0 }
let tags = Set.of(["new", "sale"])
println(Json.stringify(Map{ "start": start, "timeout": 90s }))
println(Json.stringify(stock))
println(Json.stringify(tags))
```

```text
{"start":"2026-10-22T14:03:05Z","timeout":"PT90S"}
{"tea":12,"mug":0}
["new","sale"]
```

[Time.parseInstant](../../lib/Time/parseInstant.md) and [Set.of](../../lib/Set/of.md) build the
values here.

A key or value of a named primitive over `string` is written without its tag, so keep such keys as
they are; only keys of other types need converting:

```ant
type UserId = string

let visits = Map{ UserId("u-1"): 1, UserId("u-2"): 3 }
println(Json.stringify(visits))
println(Json.stringify(Map{ "owner": UserId("u-1") }))
```

```text
{"u-1":1,"u-2":3}
{"owner":"u-1"}
```

## Convert what has no JSON form

`Json.stringify` panics on a value with no JSON form, wherever it sits in the value. Convert it
first:

- Encode `Bytes` as a string with [Encoding.toBase64](../../lib/Encoding/toBase64.md).
- Turn the keys of a map into strings with [Map.entries](../../lib/Map/entries.md),
  [Array.map](../../lib/Array/map.md), and [Map.fromEntries](../../lib/Map/fromEntries.md). A JSON
  property name is a string, so an `int` key panics.
- Render `Html` with [String.from](../../lib/String/from.md).

```ant
type Upload = { name: string, data: string, scores: Map[string, string] }

let bytes = "hi".toBytes()
let ranks = Map{ 1: "gold", 2: "silver" }
let upload = Upload{
  name: "hi.txt",
  data: Encoding.toBase64(bytes),
  scores: ranks
    | Map.entries()
    | Array.map({ Entry{ key: String.from(x.key), value: x.value } })
    | Map.fromEntries(),
}
println(Json.stringify(upload))
```

```text
{"name":"hi.txt","data":"aGk=","scores":{"1":"gold","2":"silver"}}
```

`NaN` and the infinities have no JSON form either:

```ant
let ratio = 0.0 / 0.0
println(Json.stringify(Map{ "ratio": ratio })) // panic: Json.stringify: NaN has no JSON form
```

## Mark the member of a union

A union member is written as the record it holds, without its tag, so a reader cannot tell
`Circle{ radius: 1.0 }` from another record with a `radius`. Add a field with a default that names
the member:

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

## Choose other property names

`Json.stringify` names each property after its field. When the format wants other names for one
type, declare a record with the names the format uses and copy the values into it:

```ant
type User = { firstName: string, lastName: string }
type UserJson = { first_name: string, last_name: string }

let u = User{ firstName: "Ann", lastName: "Lee" }
println(Json.stringify(UserJson{ first_name: u.firstName, last_name: u.lastName }))
```

```text
{"first_name":"Ann","last_name":"Lee"}
```

For a policy that applies to every record, such as names in snake case and no `null` fields, write
an encoder over [Reflect](../../lib/Reflect/index.md). It walks the value along its type, one level
at a time: [Reflect.shape](../../lib/Reflect/shape.md) says what a type is,
[Reflect.field](../../lib/Reflect/field.md), [Reflect.items](../../lib/Reflect/items.md), and
[Reflect.payload](../../lib/Reflect/payload.md) read its parts, and
[Reflect.scalar](../../lib/Reflect/scalar.md) gives a leaf for `Json.stringify` to write. The caller
passes the type as `type(Customer)`, and [Reflect.valueOf](../../lib/Reflect/valueOf.md) boxes the
value with it:

```ant
type Address = { streetName: string, postCode: string | null }
type Customer = { firstName: string, nickName: string | null, addresses: Array[Address] }

func snakeCase(name: string): string {
  name.chars()
    | Array.map({ if x.isUpper() { "_" + x.toLower() } else { x } })
    | String.concat()
}

func isNull(d: Reflect.Value): bool {
  Reflect.shape(Reflect.type(d)) is Reflect.Null
}

func encode(d: Reflect.Value): string {
  let s = Reflect.shape(Reflect.type(d))
  if s is Reflect.Record {
    let members = s.fields
      | Array.map({ Entry{ key: x.name, value: Reflect.field(d, x.name) } })
      | Array.filter({ !isNull(x.value) })
      | Array.map({ Json.stringify(snakeCase(x.key)) + ":" + encode(x.value) })
      | String.join(",")
    return "{" + members + "}"
  }
  if s is Reflect.Array {
    return "[" + (Reflect.items(d) | Array.map(encode) | String.join(",")) + "]"
  }
  if s is Reflect.Tagged {
    return encode(Reflect.payload(d))
  }
  Json.stringify(Reflect.scalar(d))
}

func toJson[T](t: Reflect.Type[T], value: T): string {
  encode(Reflect.valueOf(t, value))
}

let c = Customer{
  firstName: "Ann",
  nickName: null,
  addresses: [Address{ streetName: "Main St", postCode: null }],
}
println(toJson(type(Customer), c))
```

```text
{"first_name":"Ann","addresses":[{"street_name":"Main St"}]}
```

`Reflect.type` gives the type of the value a field holds, never a union, so a `null` in a
`string | null` field has the `Reflect.Null` shape. `Reflect.scalar` panics on any other shape, such
as a map; add a branch for each shape your data has, such as `Reflect.Map` with
[Reflect.entries](../../lib/Reflect/entries.md).

## See also

[Json.stringify](../../lib/Json/stringify.md) · [Json.Value](../../lib/Json/types/Value.md) ·
[File.write](../../lib/File/write.md) · [Encoding.toBase64](../../lib/Encoding/toBase64.md) ·
[Reflect](../../lib/Reflect/index.md) · [Decode JSON into records](json-decode.md)
