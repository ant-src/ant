<!-- status: reviewed -->

# Variables

`let` names a value you never reassign, and `var` names one you reassign later. Edit the program and
run it again.

```ant
let language = "Ant"
var visits = 0
visits += 1
visits = visits * 10

var price: float64 = 2.5
price = price + 0.5

let guests: Array[string] = []
guests.push("Ada")
guests.push("Grace")

let [_, second] = guests

println("${language}: ${visits} visits at ${price}")
println("second ${second}, guests ${guests}")
```

```text
Ant: 10 visits at 3.0
second Grace, guests ["Ada", "Grace"]
```

Every declaration has a value. The type checker reads its type from that value: `language` is a
`string` and `visits` an `int`. Write the type after a colon to state it, as `price` does, or when
the value alone does not say, as the empty array for `guests` does.

`let` fixes the name, not the value behind it. You cannot point `guests` at another array, but you
can push onto the array it names. Ant has no `++` or `--`; use `+= 1` and `-= 1`.

`_` is a name that discards its value. Here it skips the first guest. You can assign to `_`, but you
cannot read it.

## Try it

Assign to a `let`, and `ant check` stops you before the program runs:

```ant
let limit = 100
limit = 200 // error: cannot assign to limit, a let; declare it with var to reassign it
```

A `var` keeps the type of its first value. To hold another type later, write the type it may hold:

```ant
var status = "idle"
status = 0 // error: cannot use int where string is expected
var answer: string | int = "unknown"
answer = 42
```

An empty array says nothing about its elements, so annotate it. And `_` takes values but never gives
one back:

```ant
let names = [] // error: cannot infer the element type of an empty array; annotate it
println(_)     // error: _ is write-only and cannot be read
```

## Learn more

For more on names, scope, and when to pick `let` or `var`, read
[Values and variables](../guide/values.md) and [Types](../guide/types.md). The reference pages are
[let](../language/keywords/let.md), [var](../language/keywords/var.md),
[identifiers](../language/lexical/identifiers.md), and [Array.push](../lib/Array/push.md).
