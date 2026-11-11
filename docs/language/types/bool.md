<!-- status: reviewed -->

# bool

`bool` is the type of the two values `true` and `false`. Conditions, the logical operators,
comparisons, and predicates all use it.

## Details

- `bool` has no truthiness. A condition, `!`, `&&`, `||`, and `assert` take a `bool` only, and a
  number, a string, or a nullable value there is a compile error: give a `bool | null` a default
  with `??` first.
- `bool` does not convert to or from a number or a string. `int(true)` is a compile error, and no
  `String` function parses `"true"`: compare instead, such as `n != 0` or `s == "true"`.
- `bool` values have no order. `==` and `!=` compare them; `<` and `Array.sort` do not apply.
- `bool` renders as `true` or `false`, in a string, through `println`, and inside a container.

## Basic examples

A literal, a comparison, and a function that answers a question all give a `bool`:

```ant
let ready: bool = true
ready           // true
3 > 2           // true
"ant".isEmpty() // false
```

`bool` in a string:

```ant
let done = false
"done: ${done}" // "done: false"
```

## Scope

### Predicates

A function that returns `bool` is a predicate. Its type, `func(T): bool`, is what `filter`, `some`,
`every`, and `find` take. To negate a predicate, put `!` inside a lambda that calls it:

```ant
func isEven(n: int): bool { n % 2 == 0 }

let xs = [1, 2, 3, 4]
xs.filter(isEven)         // [2, 4]
xs.filter({ !isEven(x) }) // [1, 3]
xs.some({ x > 3 })        // true
xs.countWhere(isEven)     // 2
```

### Record fields and named arguments

A `bool` parameter often has a default, so a caller names it only to change it:

```ant
func greet(name: string, shout: bool = false): string {
  if shout { name.toUpper() } else { name }
}

greet("ann")              // "ann"
greet("ann", shout: true) // "ANN"
```

For several flags, an options record with `bool` fields keeps the call short:

```ant
type Options = { verbose: bool = false, dryRun: bool = false }

func run(task: string, opts: Options = Options{}): string {
  let prefix = if opts.dryRun { "would run" } else { "running" }
  "${prefix} ${task}"
}

println(run("build"))
println(run("deploy", opts: Options{ dryRun: true }))
```

```text
running build
would run deploy
```

### Collections

`bool` works as an array element, a map key, and a set member. `groupBy` on a predicate splits an
array into the elements it accepts and the rest:

```ant
let groups = [1, 2, 3, 4, 5].groupBy({ x % 2 == 0 })
groups[true]                // [2, 4]
groups[false]               // [1, 3, 5]
Set.of([true, true, false]) // Set.of([true, false])
```

### Nullable booleans

`bool | null` has three states. `??` picks the default for the missing one:

```ant
let settings: Map[string, bool] = Map{ "color": false }
settings["color"] ?? true // false
settings["debug"] ?? true // true
```

### In a union

`is bool` picks the `bool` member of a union, such as a JSON value:

```ant
func describe(v: Json.Value): string {
  if v is bool { return if v { "yes" } else { "no" } }
  String.from(v)
}

let parsed = Json.parse(`[true, 1, false]`) | Result.unwrap()
if parsed is Array {
  println(parsed.map(describe) | String.join(", "))
}
```

```text
yes, 1.0, no
```

### Named types over `bool`

A `type` declaration over `bool` makes a distinct type. Its constructor tags a value, and `bool(x)`
gives the plain `bool` back:

```ant
type Consent = bool

let accepted = Consent(true)
accepted                  // Consent(true)
bool(accepted)            // true
accepted == Consent(true) // true
```

## Applications

### Toggling a flag

```ant
var lightOn = false
for _ in Iter.range(0, 3) {
  lightOn = !lightOn
  println(lightOn)
}
```

```text
true
false
true
```

### Reading a yes-or-no answer

`String` has no function that parses a `bool`, so compare the text:

```ant
func answer(text: string): bool | null {
  let word = text.trim().toLower()
  if word == "yes" || word == "y" { return true }
  if word == "no" || word == "n" { return false }
  null
}

answer(" Yes ") // true
answer("n")     // false
answer("maybe") // null
```

## Properties and relations

Double negation gives the value back, and comparing a `bool` with `true` gives the same `bool`:

```ant
for b in [true, false] {
  assert(!!b == b)
  assert((b == true) == b)
  assert((b != false) == b)
  assert(String.from(b) == (if b { "true" } else { "false" }))
}
println("ok")
```

```text
ok
```

## Possible issues

### A number is not a condition

```ant
let count = 3
if count { println("some") } // error: cannot use int where bool is expected
```

Compare instead:

```ant
let count = 3
count > 0 // true
```

### `bool` does not convert to a number

```ant
let n = int(true) // error: int converts a number or a named primitive type over a number, got bool
```

Count with `countWhere`, or choose with `if`:

```ant
let flags = [true, false, true]
flags.countWhere({ x })      // 2
if flags[1] { 1 } else { 0 } // 0
```

### `bool` values cannot be sorted

`Array.sort` needs an ordered type. Sort by an `int` key instead:

```ant
let flags = [true, false, true]
let sorted = flags.sort() // error: T would be bool, which is not a member of its bound Ordered
```

```ant
let flags = [true, false, true]
flags.sortBy({ if x { 1 } else { 0 } }) // [false, true, true]
```

### A nullable `bool` is not a condition

```ant
let flag: bool | null = null
if flag { println("on") } // error: cannot use bool | null where bool is expected
```

```ant
let flag: bool | null = null
flag ?? false // false
```

### A named type over `bool` is not a `bool`

```ant
type Consent = bool
let accepted = Consent(true)
if accepted { println("ok") } // error: cannot use Consent where bool is expected
```

Unwrap it with `bool(accepted)`.

## See also

- [true and false](../literals/booleans.md)
- [&&, ||, and !](../operators/logical.md)
- [Comparison operators](../operators/comparison.md)
- [Equality operators](../operators/equality.md)
- [if](../keywords/if.md)
- [bool(x)](../../lib/prelude/bool.md)
- [String.from](../../lib/String/from.md)
- [Random.bool](../../lib/Random/bool.md)
- [Array.countWhere](../../lib/Array/countWhere.md)
- [Null](../../guide/null.md)
