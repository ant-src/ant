<!-- status: reviewed -->

# Generics

A type parameter lets one function or type work for many types. You declare it in square brackets
after the name, and the type checker works out what it stands for at each call.

```ant
type Pair[A, B] = { first: A, second: B }

func swap[A, B](p: Pair[A, B]): Pair[B, A] {
  Pair{ first: p.second, second: p.first }
}

func repeat[T](value: T, times: int): Array[T] {
  let out: Array[T] = []
  for _ in Array.range(0, times) {
    out.push(value)
  }
  out
}

func largest[T: Ordered](xs: Array[T]): T | null {
  var best: T | null = null
  for x in xs {
    if best == null || x > best {
      best = x
    }
  }
  best
}

func names[T: { name: string }](xs: Array[T]): Array[string] {
  xs.map({ x.name })
}

type User = { name: string, age: int }
type City = { name: string, population: int }

let p = Pair{ first: "Ann", second: 31 }
println(swap(p))
println(repeat("ab", 3))
println(repeat(0.5, 2))
println(largest([3, 9, 4]))
println(largest(["pear", "apple"]))
println(names([User{ name: "Ann", age: 31 }]))
println(names([City{ name: "Oslo", population: 700000 }]))
```

```text
Pair{ first: 31, second: "Ann" }
["ab", "ab", "ab"]
[0.5, 0.5]
9
pear
["Ann"]
["Oslo"]
```

`Pair[A, B]` is a generic record type, and `swap` is a generic function: it takes a `Pair[A, B]` and
returns a `Pair[B, A]`. You never write the type arguments at a call. In `swap(p)`, the type checker
reads `A = string` and `B = int` off the argument, and in `repeat("ab", 3)` it sets `T` to `string`.
The literal `Pair{ first: "Ann", second: 31 }` infers its type arguments the same way.

Without a bound, `T` could be any type, so the body of `repeat` can store it, pass it along, and
compare it with `==`, but not read fields or apply `>`. `[T: Ordered]` bounds `T` to the numbers and
`string`, the types `>` compares, which is what lets `largest` write `x > best`. A record bound such
as `{ name: string }` accepts every record type that has a `name: string` field, so `names` reads
`x.name` from users and cities alike. Each call picks one member of the bound: `T` is `int` in one
call of `largest` and `string` in the next.

## Try it

Each marked line fails to check. Two arguments for one `T` must have the same type, a bound admits
only its members, and an empty array gives the type checker no element type to infer:

```ant
func bigger[T: Ordered](a: T, b: T): T {
  if a > b { a } else { b }
}

println(bigger(1, 2.0)) // error: cannot use float64 where int is expected
println(bigger(true, false))
// error: T would be bool, which is not a member of its bound Ordered
let zs = [] // error: cannot infer the element type of an empty array; annotate it
let ys: Array[string] = []
println(ys)
```

The annotation on `ys` fixes the element type, so its empty array checks. A generic type with
different type arguments is a different type, even when one element type fits the other:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

let circles = [Circle{ radius: 1.0 }]
let shapes: Array[Shape] = circles
// error: cannot use Array[Circle] where Array[Shape] is expected
let widened: Array[Shape] = [...circles]
```

Spreading the circles into a new array builds an `Array[Shape]`; the expected type widens each
element.

Read more in the [generics chapter](../guide/generics.md) of the guide, and in the reference pages
for [`type`](../language/keywords/type.md), [`func`](../language/keywords/func.md), and
[`Array`](../language/types/Array.md). Next, [interfaces](16-interfaces.md), bounds that list the
functions a type must have.
