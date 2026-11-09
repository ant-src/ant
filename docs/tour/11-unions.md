<!-- status: reviewed -->

# Unions

A union is a type whose values belong to one of several types. You write it with `|`, and you find
out which member a value holds with `is`.

```ant
type Id = int | string

func describe(id: Id): string {
  if id is int { "number ${id}" } else { "text ${id}" }
}

type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func area(s: Shape): float64 {
  if s is Circle { 3.14 * s.radius * s.radius }
  else if s is Rect { s.width * s.height }
}

type Red = {}
type Yellow = {}
type Green = {}
type Light = Red | Yellow | Green

func next(light: Light): Light {
  if light is Red { Green{} }
  else if light is Green { Yellow{} }
  else if light is Yellow { Red{} }
}

println(describe(42))
println(describe("a7"))

let shapes: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
for s in shapes {
  println(area(s))
}
println(shapes.filter({ x is Rect }).length())

let lights: Array[Light] = [Red{}, Green{}, Yellow{}]
for light in lights {
  println("${light} -> ${next(light)}")
}
```

```text
number 42
text a7
3.14
6.0
1
Red{} -> Green{}
Green{} -> Yellow{}
Yellow{} -> Red{}
```

`Id` holds an `int` or a `string`. Inside `if id is int`, the type checker knows `id` is an `int`;
in the `else` branch it knows `id` is the rest of the union, a `string`. This is narrowing: a test
on a local variable or a parameter changes its type in the branch it guards.

`Shape` is a tagged union, a union of named records. Every record value carries its type's name as a
tag, and `s is Circle` tests that tag. In the branch, `s` is a `Circle`, so `s.radius` is there to
read. Ant has no pattern matching: `is` with `if` is how you take a union of records apart. `is`
works anywhere a `bool` does, as in the lambda passed to `filter`.

The chain in `area` has no final `else`. After `s is Circle` and `s is Rect`, no member of `Shape`
is left, so the type checker knows the chain covers every case and gives it type `float64`.

`Light` is an enumeration: a union of empty records. `Red{}` builds a value of it, and `Red` is also
a type of its own.

## Try it

Add a member to `Light` and leave `next` as it is. The chain no longer covers every member, so
`next` fails to check until you handle the new one:

```ant
type Red = {}
type Yellow = {}
type Green = {}
type Blue = {}
type Light = Red | Yellow | Green | Blue

func next(light: Light): Light {
  if light is Red { Green{} }
  else if light is Green { Yellow{} }
  else if light is Yellow { Red{} }
  // error: an if without else has type void when the condition is false, which cannot be used where Light is expected
}
```

A field that every member declares, such as a `name` on both `Circle` and `Rect`, reads straight
through the union: `s.name` is a `string`.

```ant
type Circle = { name: string, radius: float64 }
type Rect = { name: string, width: float64, height: float64 }
type Shape = Circle | Rect

func label(s: Shape): string { s.name }

println(label(Rect{ name: "box", width: 2.0, height: 3.0 }))
```

```text
box
```

Read a field that only some members declare, and the type checker asks you to narrow first:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func radius(s: Shape): float64 {
  s.radius // error: Shape has no field radius on every member: narrow the value first
}
```

`is` must be able to tell the members of a union apart at run time:

```ant
type Ids = Array[int] | Array[string] // error: a union may contain each discriminator at most once: Array[int] and Array[string] are both an array
```

`is` sees only that a value is an array, not what it holds, so write `Array[int | string]` instead.

Read more in the [unions chapter](../guide/unions.md) of the guide, and in the reference pages for
[`is`](../language/operators/is.md), [`if`](../language/keywords/if.md), and
[`type`](../language/keywords/type.md). Next, [null](12-null.md) and the nullable type `T | null`, a
union with `null`.
