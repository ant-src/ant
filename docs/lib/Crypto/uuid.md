<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Crypto.uuid

```ant signature
func Crypto.uuid(): string
```

Crypto.uuid returns a random version 4 UUID, in lower case with hyphens.

<!-- end generated -->

## Details

- Each call returns a new id from the operating system's secure source of randomness. No
  `Random.new` seed affects it, so a test cannot fix the value it returns.

## Basic examples

Make an id. Its value differs on every run, so this example checks its shape:

```ant
Crypto.uuid().length() // 36
```

Two calls give two different ids:

```ant
Crypto.uuid() == Crypto.uuid() // false
```

Print one:

<!-- test: check-only the id is random -->

```ant
println(Crypto.uuid())
```

The program prints a line such as `43f692cc-2d9e-4718-94cc-98b1b5aec6d9`.

## Scope

The five groups hold 8, 4, 4, 4, and 12 hex digits:

```ant
Crypto.uuid().split("-").map({ x.length() }) // [8, 4, 4, 4, 12]
```

The first digit of the third group is always `4`, the version:

```ant
Crypto.uuid().slice(14, 15) // "4"
```

The first digit of the fourth group is one of `8`, `9`, `a`, or `b`, the variant:

```ant
"89ab".contains(Crypto.uuid().slice(19, 20)) // true
```

Every letter is lower case:

```ant
let id = Crypto.uuid()
id == id.toLower() // true
```

## Applications

Give each record its own id with a field default. A default is evaluated afresh for each literal, so
two tasks with the same title still get different ids:

```ant
type Task = { id: string = Crypto.uuid(), title: string }
let a = Task{ title: "write docs" }
let b = Task{ title: "write docs" }
a.title == b.title // true
a.id == b.id       // false
```

Keep records in a map keyed by their ids:

```ant
type User = { id: string, name: string }
let users: Map[string, User] = Map{}
for name in ["Ann", "Bob"] {
  let user = User{ id: Crypto.uuid(), name }
  users[user.id] = user
}
let names = users.values()
  | Array.map({ x.name })
  | Array.sort()
  | String.join(", ")
println(names)
```

```text
Ann, Bob
```

Give a scratch file a name that no other run is likely to pick. `replace: false` makes a clash an
`Err` instead of an overwrite, and `defer` removes the file when done:

```ant
func scratch(text: string): Result[string, Error] {
  let path = "upload-${Crypto.uuid()}.tmp"
  try File.write(path, text, replace: false)
  defer { File.remove(path) }
  File.read(path)
}

println(scratch("hello").unwrap())
```

```text
hello
```

Check that a string from outside has the shape `Crypto.uuid` produces before using it as an id:

```ant
let shape = `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
let pattern = Regex.compile(shape).unwrap()

func parseId(text: string): Result[string, Error] {
  if !pattern.test(text) {
    return Result.fail("not a version 4 UUID: ${text}")
  }
  Result.ok(text)
}

println(parseId(Crypto.uuid()).isOk())
println(parseId("42"))
```

```text
true
Err{ error: Error{ message: "not a version 4 UUID: 42" } }
```

## Properties and relations

Without its hyphens an id is 16 bytes in hex. The high four bits of byte 6 hold the version, 4, and
the top two bits of byte 8 hold the variant, `10` in binary:

```ant
let id = Crypto.uuid()
let bytes = Encoding.fromHex(id.replace("-", "")).unwrap()
assert(bytes.length() == 16)
assert(Bits.shiftRight(bytes[6], 4) == byte(4))
assert(Bits.and(bytes[8], byte(0xc0)) == byte(0x80))
```

A thousand calls give a thousand different ids:

```ant
let ids = Array.range(0, 1000).map({ Crypto.uuid() })
assert(Set.of(ids).size() == 1000)
```

## Possible issues

An id is random, so sorting ids does not sort records by the time they were made. Store the creation
time in a field of its own for that.

`Crypto.uuid` always gives lower case, but an id from another system may come in upper case, and
`==` compares strings exactly. Compare with `String.equalFold`:

```ant
let stored = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
let received = "F47AC10B-58CC-4372-A567-0E02B2C3D479"
stored == received         // false
stored.equalFold(received) // true
```

A `let` evaluates `Crypto.uuid()` once. Call it again, or use a field default, for each value that
needs its own id:

```ant
let id = Crypto.uuid()
let ids = [id, id]
ids[0] == ids[1] // true
```

## See also

[Crypto.randomBytes](randomBytes.md), [Encoding.toHex](../Encoding/toHex.md),
[File.tempFile](../File/tempFile.md), [Random.new](../Random/new.md),
[String.equalFold](../String/equalFold.md)
