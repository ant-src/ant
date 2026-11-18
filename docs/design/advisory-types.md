<!-- status: reviewed -->

# Types are advisory

Ant is statically typed, and its interpreter ignores the annotations. `ant check` reads every
annotation and reports compile errors; `ant run` parses the annotations and discards them. This page
explains why Ant splits the work that way, what each command guarantees, and what the split costs.

## Two commands, two jobs

```ant
type User = { name: string, age: int }

let users = [
  User{ name: "Ann", age: 31 },
  User{ name: "Bob", age: "27" }, // error: cannot use string where int is expected
]
println("${users.length()} users")
let adults = users.filter({ x.age >= 18 })
println(adults)
```

`ant check` rejects the program at the field that is wrong, before anything runs. `ant run` does not
look at types, so it runs the program until an operation gets values it cannot use:

```sh
ant run main.ant
```

```text
2 users
panic: >= needs two numbers of one kind or two strings, got string and int
  at func (main.ant:8:35)
  at module main.ant (main.ant:8:26)
```

The checker points at line 5, where the mistake is. The interpreter stops at line 8, where the wrong
value is first used, after the work in between has run. Both catch the bug, but only one says where
it was made.

The checker reads the whole program, including code nothing calls, and finds every compile error on
every path. The interpreter sees one run, on one input, and a mistake on a path that run does not
take stays hidden:

```ant
func greet(name: string): string {
  if name.isEmpty() {
    return defualt // error: defualt is not defined
  }
  "hello, ${name}"
}

println(greet("Ann"))
```

`ant run` prints `hello, Ann` and exits with status 0, because an unknown name panics only when the
program reaches it.

## What the interpreter keeps

A running program has values, and each value knows its kind: `int`, `string`, array, record, and so
on. A record and a value of a named primitive type also carry a tag, the name of their `type`
declaration. A `type` declaration is the one thing the interpreter keeps of the type language: its
tag, a record's field names and defaults, and its body, which only `type(T)` reads. Annotations, a
function's type parameters, and interface declarations are gone by the time the program runs.

Every rule of `ant run` is defined on those values. `+` looks at the kinds of its operands, a field
read looks up the field in the record it gets, and `is` compares a tag:

```ant
type Meters = float64

func describe(v: Meters | float64): string {
  if v is Meters { "meters" } else { "a plain number" }
}

describe(Meters(3.0)) // "meters"
describe(3.0)         // "a plain number"
```

`is` can test `Meters` because the value carries the `Meters` tag. It cannot test what a value has
no record of. A type alias names no tag a value carries, and an array does not record its element
type, so `is` takes neither:

```ant
type Id = int | string

func show(v: Id | bool): string {
  if v is Id { "an id" } else { "a flag" }
  // error: Id is an alias and has no run-time identity; test its members instead
}
```

Test the members instead, `v is int || v is string`, or narrow `v is bool` first.

## Why the interpreter ignores types

### A program means what its code says

No evaluation rule looks at an annotation: only a type expression, `type(T)`, turns a written type
into a value. An annotation never converts a value and never changes what an operator or a call
does. Reading a line of Ant, you know what it does from its expressions and the values they get. No
annotation around it can change that.

The checker holds the annotations to the same rule. Since the interpreter never sees
`let ratio: float64 = 1`, the checker cannot let the annotation turn the `1` into a float:

```ant
type Meters = float64

let ratio: float64 = 1   // error: cannot use int where float64 is expected
let height: Meters = 3.0 // error: cannot use float64 where Meters is expected
```

Write the value with the type it has: `1.0`, or `float64(n)` for an `int` in a variable, and
`Meters(3.0)` to give a number the `Meters` tag. A conversion you can see is one the interpreter can
run.

### Running starts at once

`ant run` does no static work beyond parsing. It loads the modules a program imports, so a syntax
error anywhere stops the program before any of it runs, and then it runs. Inference, generic bounds,
and narrowing all live in the checker, and so does their cost: starting a script costs parsing, not
checking.

The parser asks the checker nothing either. The grammar reads a program the same way whatever its
types, so the interpreter runs what it parsed with no checker at all.

### Checked programs fail only on values

Advisory does not mean weak. `ant check` rules out every panic that comes from a wrong kind: an
operand of the wrong type, a call of something that is not a function, a field a record does not
have. A program that passes the checker can panic only in ways that depend on the values it meets at
run time:

```ant
func average(xs: Array[int]): int {
  (xs.sum() ?? 0) / xs.length()
}

println(average([3, 5, 7]))
let empty = average([]) // panic: division by zero
println(empty)
```

```text
5
```

Panics such as an index out of range, an integer division by zero, an overflow, a failed `assert`,
and `Result.unwrap` on an `Err` remain. The checker cannot know which arrays will be empty, so that
guard is your job. Whether every value has the kind each operation needs, the checker settles before
the program starts.

The `?? 0` in `average` is there because `sum` returns `null` on an empty array; the last section
says why.

## What it costs

- **You run the checker yourself.** No flag makes `ant run` check first, and `ant test` does not
  check either. Run `ant check` in your editor and in continuous integration, on each script and
  each test file.
- **A wrong annotation can go unnoticed.** In a program nobody checked, a wrong annotation fails
  only where a value is misused, and sometimes nowhere:

  ```ant
  let port: int = "8080" // error: cannot use string where int is expected
  println("listening on ${port}")
  ```

  `ant run` prints `listening on 8080` and exits with status 0, though `port` holds a string.

- **Conversions are always written.** An integer literal is always an `int`, a float literal a
  `float64`, and a string literal never carries a tag. An annotation does not change that, so a
  value of another numeric type or of a named primitive type needs a call such as `byte(255)` or
  `Meters(3.0)`.
- **`is` sees only kinds and tags.** It tests `Array`, not `Array[int]`, and it cannot test an alias
  or an interface. So a union holds at most one array type, whatever its element type, and likewise
  one map and one set type.
- **Generic code cannot make a zero.** The interpreter never knows a type parameter, so `Array.sum`,
  `Array.min`, and `Array.max` return `null` on an empty array instead of a zero of the right kind.
  Write `xs.sum() ?? 0`, or `?? 0.0` for floats.
- **Types leave little at run time.** A function value cannot report its signature, and a type
  parameter is bound to nothing while the program runs, so `type(T)` on a function's own type
  parameter `T` is a compile error. The function takes a `Reflect.Type[T]` parameter instead, and
  its caller writes `type(Config)`, with a name the interpreter keeps.

## See also

The [types guide](../guide/types.md) introduces the checker, and
[Types at run time](../guide/runtime-types.md) shows what survives into a running program and how
`type(T)` and `Reflect` use it. The tools: [ant check](../tools/check.md),
[ant run](../tools/run.md), and [ant test](../tools/test.md). The design notes
[No implicit coercion, no truthiness](no-coercion.md) and [Why no exceptions](errors-are-values.md)
cover related decisions.
