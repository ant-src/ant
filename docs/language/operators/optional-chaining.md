<!-- status: reviewed -->

# ?.

`?.` reads a field (`a?.b`), indexes (`a?.[i]`), or calls (`f?.(x)`) through a value that may be
`null`. When the left side is `null`, the result is `null` and the rest of the chain is skipped;
otherwise the operator acts as `.`, `[i]`, or a call.

## Details

- The left side's type must include `null`. On a type without `null`, `?.` is a compile error.
- The result's type adds `null` to the type of the plain access: `user?.name` is `string | null`.
- A `null` left side skips everything after it in the chain, arguments included: in
  `s?.trim().length()` neither `trim` nor `length` runs, and in `f?.(g())` the function `g` is not
  called.

## Basic examples

```ant
type User = { name: string }
let ann: User | null = User{ name: "Ann" }
let nobody: User | null = null
ann?.name    // "Ann"
nobody?.name // null
```

Pair it with `??` to replace `null` with a default:

```ant
type User = { name: string }
let nobody: User | null = null
nobody?.name ?? "guest" // "guest"
```

## Scope

### Chains of nullable links

Each link that may be `null` takes its own `?.`. The chain is `null` as soon as one link is:

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }
let ann: User | null = User{ name: "Ann", address: Address{ city: "Oslo" } }
let bob: User | null = User{ name: "Bob", address: null }
ann?.address?.city // "Oslo"
bob?.address?.city // null
```

### Dot calls

`x?.f()` calls a function of `x`'s namespace when `x` is not `null`. Links after it that cannot be
`null` use a plain `.`, since a `null` left side has already skipped them:

```ant
let input: string | null = "  ant  "
let missing: string | null = null
input?.trim()            // "ant"
input?.trim().length()   // 3
missing?.trim().length() // null
```

### Indexing

`a?.[i]` indexes an array, a map, or `Bytes` that may be `null`:

```ant
let xs: Array[int] | null = [10, 20]
let none: Array[int] | null = null
xs?.[1]   // 20
none?.[1] // null
let ages: Map[string, int] | null = Map{ "Ann": 31 }
ages?.["Ann"] // 31
```

### Calling a function value

`f?.(x)` calls a function value that may be `null`. Here an optional callback runs only when one is
set:

```ant
type Options = { onDone: (func(string): void) | null }
func save(opts: Options) {
  opts.onDone?.("saved")
}
save(Options{ onDone: { println("done: ${x}") } })
save(Options{ onDone: null })
```

```text
done: saved
```

A `null` function skips its arguments too:

```ant
func loud(): int {
  println("evaluated")
  1
}
let f: (func(int): int) | null = null
println(f?.(loud()))
```

```text
null
```

### Fields that do not narrow

A test on a field path such as `user.address` does not narrow it, since a call between the test and
the use may reassign the field. Read the field with `?.` instead:

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }
let user = User{ name: "Ann", address: Address{ city: "Oslo" } }
user.address?.city ?? "unknown" // "Oslo"
```

Or bind the field to a local, which does narrow:

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }
let user = User{ name: "Ann", address: Address{ city: "Oslo" } }
let address = user.address
if address != null {
  println(address.city)
}
```

```text
Oslo
```

### Across lines

A line that starts with `?.` or `.` continues the expression above, so a long chain can break before
each link:

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }
let users = [User{ name: "Ann", address: Address{ city: "Oslo" } }]
let city = users.find({ x.name == "Ann" })
  ?.address
  ?.city
  .toUpper()
println(city)
```

```text
OSLO
```

## Applications

Show a fallback for each missing link of a lookup:

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }
let users = [
  User{ name: "Ann", address: Address{ city: "Oslo" } },
  User{ name: "Bob", address: null },
]
for name in ["Ann", "Bob", "Cy"] {
  let city = users.find({ x.name == name })?.address?.city ?? "unknown"
  println("${name}: ${city}")
}
```

```text
Ann: Oslo
Bob: unknown
Cy: unknown
```

Read an optional setting with a default:

```ant
type Server = { port: int }
type Config = { server: Server | null }
func port(config: Config): int {
  config.server?.port ?? 8080
}
println(port(Config{ server: Server{ port: 3000 } }))
println(port(Config{ server: null }))
```

```text
3000
8080
```

## Properties and relations

`a?.b` is an `if` expression that tests for `null`:

```ant
type User = { name: string }
for user in [User{ name: "Ann" }, null] {
  let viaIf = if user == null { null } else { user.name }
  assert(user?.name == viaIf)
}
```

With `??`, `?.` gives a value that is never `null`:

```ant
let word: string | null = null
let size: int = word?.length() ?? 0
assert(size == 0)
```

## Possible issues

### A left side that cannot be `null`

`?.` needs a nullable left side. On a plain value, write `.`:

```ant
type User = { name: string }
let ann = User{ name: "Ann" }
println(ann?.name) // error: ?. needs a nullable left side, got User
```

The same holds inside a chain: after `user?.name`, a `null` user has already skipped the rest, so
the next link sees a `string`:

```ant
type User = { name: string }
let user: User | null = null
let size = user?.name?.length() // error: ?. needs a nullable left side, got string
```

```ant
type User = { name: string }
let user: User | null = null
user?.name.length() // null
```

### A nullable link written with `.`

A `.` after a field that may be `null` is a compile error. Use `?.` for that link too:

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }
let user: User | null = null
println(user?.address.city) // error: Address | null has no field city: narrow the value first
```

### The result is still nullable

The value of a chain may be `null`, so it does not fit a type without `null`. Add a default with
`??`, or test it with `if`:

```ant
type User = { name: string }
let user: User | null = User{ name: "Ann" }
let name: string = user?.name // error: cannot use string | null where string is expected
```

A pipe after a chain passes it the nullable value as well:

```ant
let input: string | null = " ant "
let size = input?.trim() | String.length()
// error: cannot use string | null where string is expected
```

### `null` from two places

`ages?.["Bob"]` is `null` both when `ages` is `null` and when the key is missing. Test the map with
`has` to tell the two apart:

```ant
let ages: Map[string, int] | null = Map{ "Ann": 31 }
let none: Map[string, int] | null = null
ages?.["Bob"]    // null
none?.["Bob"]    // null
ages?.has("Bob") // false
none?.has("Bob") // null
```

### Assigning through `?.`

A `?.` chain is not an assignment target. Test the value with `if`, then assign through `.`:

```ant
type User = { name: string }
var u: User | null = null
u?.name = "x" // error: the left side of '=' must be a variable, a field, an element, or an array destructuring
```

```ant
type User = { name: string }
var u: User | null = User{ name: "Ann" }
if u != null {
  u.name = "x"
}
println(u?.name)
```

```text
x
```

### Out of range

`?.` does not guard an index. An index past the end still panics; use `get` for a nullable element:

```ant
let xs: Array[int] | null = [1]
xs?.[5]    // panic: index 5 out of range for an array of length 1
xs?.get(5) // null
```

## See also

[?? operator](null-coalescing.md) · [Dot calls](dot-calls.md) · [Indexing](indexing.md) ·
[is](is.md) · [Equality](equality.md) · [Null guide](../../guide/null.md) ·
[Null tour](../../tour/12-null.md)
