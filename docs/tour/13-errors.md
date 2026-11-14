<!-- status: reviewed -->

# Errors

A function that can fail returns a `Result`: either `Ok` with a value or `Err` with an error. The
failure is a value like any other. A `Result` is not its value, so you unwrap it before you use the
value.

```ant
type Order = { item: string, qty: int }

func parseQty(text: string): Result[int, Error] {
  let n = text.toInt()
  if n == null {
    return Result.fail("\"${text}\" is not a number")
  }
  if n <= 0 {
    return Result.fail("${n} is not positive")
  }
  Result.ok(n)
}

func parseOrder(line: string): Result[Order, Error] {
  let parts = line.split(" x")
  if parts.length() != 2 {
    return Result.fail("no quantity in \"${line}\"")
  }
  let qty = try (parseQty(parts[1]) | Result.wrapErr("quantity of ${parts[0]}"))
  Result.ok(Order{ item: parts[0], qty })
}

for line in ["apples x3", "pears x0", "plums", "figs xmany"] {
  let order = parseOrder(line)
  if order is Ok {
    println("${order.value.qty} ${order.value.item}")
  } else {
    println("skipped: ${order.error.message}")
  }
}

let orders = try (["apples x3", "pears x2"] | Array.map(parseOrder) | Result.all())
println(orders.map({ x.qty }).sum())
```

```text
3 apples
skipped: quantity of pears: 0 is not positive
skipped: no quantity in "plums"
skipped: quantity of figs: "many" is not a number
5
```

`Result[int, Error]` is a tagged union of two records, `Ok{ value }` and `Err{ error }`. `Result.ok`
and `Result.err` build them. `Error` is the standard library's error type, a record with a
`message`, and `Result.fail("...")` is short for `Result.err(Error{ message: "..." })`. You can use
any type for the error, such as a union of your own records.

`try` unwraps a `Result`. On `Ok` it gives the value, so `qty` is a plain `int`. On `Err` it returns
that error from the enclosing function right away. The loop takes the result apart with `is`, as any
union: after `order is Ok` the type checker knows `order.value` is there.

`Result.wrapErr` puts context in front of an error's message, joined by `: `. `parseQty` knows the
text it read; `parseOrder` knows which item it was parsing. Each adds what it knows, and the final
message reads as one line. Messages are lowercase with no closing punctuation, so they read well
inside a longer one. The parentheses make `try` apply to the whole pipeline, not only to its first
call.

The last `try` is in the script's top-level code. `Result.all` turns an array of results into one
result: the first `Err`, or `Ok` of every value.

When a failure has an obvious fallback, `Result.unwrapOr` gives the value or the fallback:
`Result.unwrapOr(parseQty("x"), 1)` is `1`. `Result.unwrap` gives the value and panics on an `Err`.
It is for failures that are bugs, where `try` is for failures you expect. The next page covers
panics.

Use `Error` for messages a person reads, and a type of your own for errors code branches on. A union
of records lets the caller tell the kinds apart with `is`:

```ant
type Empty = {}
type BadDigit = { at: int }

func parseCount(text: string): Result[int, Empty | BadDigit] {
  if text == "" {
    return Result.err(Empty{})
  }
  let bad = text.chars().findIndex({ !x.isDigit() })
  if bad != null {
    return Result.err(BadDigit{ at: bad })
  }
  Result.ok(text.toInt() ?? 0)
}

for text in ["42", "", "4x2"] {
  let r = parseCount(text)
  if r is Ok {
    println(r.value)
  } else {
    let error = r.error
    if error is Empty {
      println("nothing to count")
    } else {
      println("bad digit at ${error.at}")
    }
  }
}
```

```text
42
nothing to count
bad digit at 1
```

The error is bound to a local first, because a field path such as `r.error` does not narrow.

## Try it

A `try` at the top level of a script ends the script on an `Err`. It prints the error to standard
error and exits with status 1:

<!-- test: exit 1 -->

```ant
func half(n: int): Result[int, Error] {
  if n % 2 != 0 {
    return Result.fail("${n} is odd")
  }
  Result.ok(n / 2)
}

println(try half(8))
println(try (half(7) | Result.wrapErr("halve 7")))
println("not reached")
```

```text
4
```

The second `try` stops the script with `halve 7: 7 is odd`, and the last line never runs.

`try` returns the error from the function it is in, so that function must return a `Result` too. And
a `Result` is not its value: you unwrap it before you use it.

```ant
func half(n: int): Result[int, Error] {
  Result.ok(n / 2)
}

func quarter(n: int): int {
  let h = try half(n) // error: try would return Err[Error], but the function returns int
  h / 2
}

let q: int = half(8) // error: cannot use Result[int, Error] where int is expected
```

The guide chapter [Errors](../guide/errors.md) explains when to return a `Result` and when to panic.
The reference has pages for [`try`](../language/keywords/try.md),
[`Result`](../lib/Result/index.md), [`Result.wrapErr`](../lib/Result/wrapErr.md),
[`Result.all`](../lib/Result/all.md), [`Result.unwrapOr`](../lib/Result/unwrapOr.md),
[`Result.unwrap`](../lib/Result/unwrap.md), and [`Error.wrap`](../lib/Error/wrap.md).
