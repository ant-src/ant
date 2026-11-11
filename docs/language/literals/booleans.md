<!-- status: reviewed -->

# true and false

`true` and `false` are the two values of type `bool`. Both are keywords.

## Details

- `true` and `false` are lowercase. `True`, `TRUE`, and `False` are ordinary names, undefined unless
  you declare them.
- Ant has no truthiness. A condition, `!`, `&&`, and `||` take a `bool` only; compare a number or a
  string to get one.
- In a map literal, a bare `true:` key is the string `"true"`. Write `(true):` for the `bool` key.

## Basic examples

The two literals:

```ant
true  // true
false // false
```

Bind one and use it in a string:

```ant
let done = false
"done: ${done}" // "done: false"
```

## Scope

### Results of comparisons and logic

A comparison, an equality test, and the logical operators produce a `bool`:

```ant
let age = 20
age >= 18            // true
age == 21            // false
!true                // false
true && false        // false
age > 65 || age < 13 // false
```

### Fields and named arguments

A record field and a named argument take `true` or `false` like any other value:

```ant
type Options = { verbose: bool, dryRun: bool = false }

func describe(name: string, loud: bool = false): string {
  if loud { name.toUpper() } else { name }
}

let opts = Options{ verbose: true }
println(opts)
println(describe("ant", loud: true))
```

```text
Options{ verbose: true, dryRun: false }
ANT
```

### Collections

`bool` works as an array element, a map key, and a set member:

```ant
let flags = [true, false, true]
let labels = Map{ (true): "on", (false): "off" }
flags.countWhere({ x }) // 2
labels[flags[1]]        // "off"
Set.of(flags)           // Set.of([true, false])
```

### Testing with `is`

In a union, `is bool` tells a `bool` apart from the other members:

```ant
let setting: bool | string = true
setting is bool // true
```

## Possible issues

### A number is not a condition

Ant does not read `0`, `""`, or `null` as `false`:

```ant
let count = 3
if count { println("some") } // error: cannot use int where bool is expected
```

Compare instead:

```ant
let count = 3
if count > 0 { "some" } else { "none" } // "some"
```

### A bare `true` key is a string

A single word before `:` in a map literal is a string key, keywords included:

```ant
Map{ true: "yes" }   // Map{ "true": "yes" }
Map{ (true): "yes" } // Map{ true: "yes" }
```

### `bool` values have no order

`<` and the other ordering operators do not apply to `bool`:

```ant
let first = true < false
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to bool; narrow first
```

### `bool` values do not mix with numbers

`true` is not `1`:

```ant
let total = true + 1 // error: + needs both operands of one type, got bool and int
```

### `bool(x)` does not convert numbers or strings

`bool` converts only a `bool` or a named primitive type over `bool`:

```ant
let on = bool(1) // error: bool converts only a bool or a named primitive type over bool, got int
```

Compare a number or a string to get a `bool`:

```ant
let count = 3
let text = "true"
count != 0     // true
text == "true" // true
```

### Keywords are not variable names

`true` and `false` are keywords, so neither can name a variable:

```ant
let true = 1 // error: 'true' cannot be a variable name
```

## See also

- [bool](../types/bool.md)
- [bool(x)](../../lib/prelude/bool.md)
- [&&, ||, and !](../operators/logical.md)
- [Comparison operators](../operators/comparison.md)
- [if](../keywords/if.md)
- [null](null.md)
- [Random.bool](../../lib/Random/bool.md)
