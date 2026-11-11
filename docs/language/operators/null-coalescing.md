<!-- status: reviewed -->

# ??

`a ?? b` yields `a`, unless `a` is `null`; then it evaluates `b` and yields that.

## Details

- `??` short-circuits: the right side runs only when the left side is `null`.
- The left side must have a type that includes `null`. On any other type, `??` is a compile error.
- The right side may be any expression, including `return`, `break`, and `panic(...)`. After
  `let v = x ?? return ...`, `v` is not nullable.
- `??` binds looser than arithmetic, comparison, and equality, and tighter than `|`. It cannot be
  mixed with `&&`, `||`, or a pipe stage without parentheses.

## Basic examples

A fallback for a nullable value:

```ant
let nickname: string | null = null
nickname ?? "anonymous" // "anonymous"
let name: string | null = "Ann"
name ?? "anonymous" // "Ann"
```

A default for a missing map key:

```ant
let ages = Map{ "Ann": 31, "Bob": 27 }
ages["Ann"] ?? 0 // 31
ages["Cy"] ?? 0  // 0
```

## Scope

### Chains

`??` chains: the first value that is not `null` wins.

```ant
let flag: string | null = null
let env: string | null = "staging"
flag ?? env ?? "production" // "staging"
```

### A nullable right side

When the right side can be `null` too, so can the result:

```ant
let primary: string | null = null
let backup: string | null = null
primary ?? backup                  // null
(primary ?? backup) ?? "localhost" // "localhost"
```

### After `?.`

`?.` turns a missing link into `null`, and `??` replaces it:

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }
let ann = User{ name: "Ann", address: Address{ city: "Oslo" } }
let bob = User{ name: "Bob", address: null }
ann.address?.city ?? "unknown" // "Oslo"
bob.address?.city ?? "unknown" // "unknown"
```

### Empty collections and numbers

Without an expected type, the right side takes the left side's type without `null`, so an empty
literal needs no annotation:

```ant
let groups = Map{ 1: [10, 20] }
groups[1] ?? [] // [10, 20]
groups[2] ?? [] // []
```

`Array.sum`, `Array.min`, and `Array.max` return `null` for an empty array; `??` gives the zero:

```ant
let empty: Array[float64] = []
[1.5, 2.0].sum() ?? 0.0 // 3.5
empty.sum() ?? 0.0      // 0.0
```

### `return`, `break`, and `panic`

A right side of type `never` leaves instead of producing a value. The name on the left is then not
nullable:

```ant
func tens(xs: Array[int]): int {
  let first = xs.first() ?? return -1
  first * 10
}

println(tens([4, 5]))
println(tens([]))
```

```text
40
-1
```

`?? break` ends a loop when a closed channel gives `null`:

```ant
let jobs: Channel[string] = Channel.new(2)
jobs.send("build")
jobs.send("test")
jobs.close()

while true {
  let job = jobs.receive() ?? break
  println("running ${job}")
}
```

```text
running build
running test
```

`?? panic(...)` unwraps a value that must be there:

```ant
let port: int | null = null
let p = port ?? panic("no port configured") // panic: no port configured
```

### Short-circuit

The right side does not run when the left side has a value:

```ant
func fallback(): string {
  println("fallback ran")
  "default"
}

let set: string | null = "set"
let unset: string | null = null
println(set ?? fallback())
println(unset ?? fallback())
```

```text
set
fallback ran
default
```

## Applications

### Counting with a map

Start a missing count at zero:

```ant
let words = "the cat and the hat".split(" ")
let counts: Map[string, int] = Map{}
for w in words {
  counts[w] = (counts[w] ?? 0) + 1
}
println(counts["the"])
println(counts["dog"] ?? 0)
```

```text
2
0
```

### Failing early with a `Result`

Turn a missing value into an error and return it:

```ant
type Account = { id: string, balance: int }

func balance(accounts: Map[string, Account], id: string): Result[int, Error] {
  let account = accounts[id] ?? return Result.fail("no account ${id}")
  Result.ok(account.balance)
}

let accounts = Map{ "a1": Account{ id: "a1", balance: 50 } }
println(balance(accounts, "a1"))
println(balance(accounts, "b2"))
```

```text
Ok{ value: 50 }
Err{ error: Error{ message: "no account b2" } }
```

### Optional settings

Fill in defaults for the fields a user left out:

```ant
type Settings = { host: string | null = null, port: int | null = null }

func address(s: Settings): string {
  "${s.host ?? "localhost"}:${s.port ?? 8080}"
}

println(address(Settings{}))
println(address(Settings{ port: 9000 }))
```

```text
localhost:8080
localhost:9000
```

## Properties and relations

`a ?? b` is the `if` expression that tests `a` against `null`:

```ant
let values: Array[int | null] = [3, null]
for a in values {
  assert((a ?? 0) == (if a != null { a } else { 0 }))
}
```

## Possible issues

### A left side that cannot be `null`

```ant
let name = "Ann"
let shown = name ?? "anonymous" // error: ?? needs a nullable left side, got string
```

### A stored `null` and a missing key give the same result

`??` cannot tell them apart. Ask the map with `has` when the difference matters:

```ant
let theme = Map{ "dark": null, "light": "#fff" }
theme["dark"] ?? "none"  // "none"
theme["sepia"] ?? "none" // "none"
theme.has("dark")        // true
theme.has("sepia")       // false
```

### Comparison binds tighter

`n ?? 0 > 5` is `n ?? (0 > 5)`, an `int | bool`, not a test:

```ant
let n: int | null = 3
let big = n ?? 0 > 5
if big { println("big") } // error: cannot use int | bool where bool is expected
```

Parenthesize the `??`:

```ant
let n: int | null = 3
(n ?? 0) > 5 // false
```

### Mixing with `&&` and `||`

```ant
let ready: bool | null = null
let start = ready ?? false && true // error: '??' cannot mix with '&&' or '||' without parentheses
```

### Mixing with a pipe

`|` binds looser than `??`, so a fallback after a pipeline needs parentheses:

```ant
let xs = [1, 2, 3]
let total = xs | Array.sum() ?? 0 // error: '|' binds looser than '??', so parenthesize the pipe, as in (a | f()) ?? b
```

Parenthesize the pipe:

```ant
let xs = [1, 2, 3]
(xs | Array.sum()) ?? 0 // 6
```

### `return` takes everything to its right

`x ?? return a + b` returns `a + b`, not `a`. A plain `?? return 0` ends at the line:

```ant
func double(x: int | null): int {
  let v = x ?? return 0
  v * 2
}

println(double(4))
println(double(null))
```

```text
8
0
```

To end the `return` earlier, parenthesize it.

### A nullable fallback keeps the result nullable

```ant
let a: string | null = null
let b: string = a ?? null // error: cannot use null where string is expected
```

## See also

- [?. (optional chaining)](optional-chaining.md)
- [null](../literals/null.md)
- [Null](../../guide/null.md)
- [Equality operators](equality.md)
- [Logical operators](logical.md)
- [return](../keywords/return.md)
- [panic](../../lib/prelude/panic.md)
- [Map.get](../../lib/Map/get.md)
- [Array.sum](../../lib/Array/sum.md)
