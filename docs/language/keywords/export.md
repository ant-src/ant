<!-- status: reviewed -->

# `export`

`export` makes a top-level declaration of a library visible to other modules. A declaration without
it is private to its file.

## Details

- `export` goes before a top-level `var`, `let`, `func`, `type`, or `interface`, and nowhere else.
- Only a library exports. A file with `export` cannot have a top-level statement, since a script
  cannot be imported.
- An exported function states its return type. A function exported on a type needs the type exported
  too.
- An imported `var` reads the exporting module's current value, but only that module can assign it.
- Privacy applies to names, not to data. A test file sees every declaration of the modules it
  imports, exported or not.

## Basic examples

A library marks what other files may use:

```ant [geometry.ant]
export type Point = { x: float64, y: float64 }

export func Point.distance(a: Point, b: Point): float64 {
  Math.sqrt(square(a.x - b.x) + square(a.y - b.y))
}

func square(n: float64): float64 {
  n * n
}
```

A script imports the exported names. The functions exported on `Point` come with it:

```ant
import "./geometry.ant" { Point }

let p = Point{ x: 3.0, y: 4.0 }
println(p.distance(Point{ x: 0.0, y: 0.0 }))
```

```text
5.0
```

`square` has no `export`, so no other module can import it:

```ant
import "./geometry.ant" { square } // error: square is not exported by ./geometry.ant
```

## Scope

### Variables and constants

`export var` and `export let` export a module's state. An imported `var` is live: each read sees the
module's current value.

```ant [counter.ant]
export let LIMIT = 3

export var count = 0

export func bump(): void {
  if count >= LIMIT {
    return
  }
  count += 1
}
```

```ant
import "./counter.ant" { LIMIT, count, bump }

println(count)
for _ in Iter.range(0, 5) {
  bump()
}
println("${count} of ${LIMIT}")
```

```text
0
3 of 3
```

### Namespace imports

A module imported under one name is a namespace of its exports, read with `.`:

```ant
import "./counter.ant" Counter

Counter.bump()
println(Counter.count)
```

```text
1
```

### Interfaces

`export interface` exports an interface. A type satisfies it outside its module only through
exported functions: here `Circle.area` is private, so another module cannot use a `Circle` as a
`Shape`.

```ant [shapes.ant]
export interface Shape {
  area(s: Self): float64
}

export type Square = { side: float64 }

export func Square.area(s: Square): float64 {
  s.side * s.side
}

export type Circle = { radius: float64 }

func Circle.area(c: Circle): float64 {
  Math.PI * c.radius * c.radius
}

export func total(shapes: Array[Shape]): float64 {
  (shapes | Array.map({ x.area() }) | Array.sum()) ?? 0.0
}
```

```ant
import "./shapes.ant" { Shape, Square, total }

let squares: Array[Shape] = [Square{ side: 2.0 }, Square{ side: 3.0 }]
println(total(squares))
```

```text
13.0
```

```ant
import "./shapes.ant" { Shape, Circle }

let s: Shape = Circle{ radius: 1.0 }
// error: cannot use Circle where Shape is expected: Circle has no function area of type func(Circle): float64
```

### Private types

A field of an exported type may have a private type. The importer reads the field but cannot name
its type:

```ant [session.ant]
type Token = { value: string }

export type Session = { token: Token }

export func Session.start(): Session {
  Session{ token: Token{ value: "abc" } }
}
```

```ant
import "./session.ant" { Session }

let s = Session.start()
println(s.token.value)
```

```text
abc
```

```ant
import "./session.ant" { Token } // error: Token is not exported by ./session.ant
```

### Test files

A test file imports private declarations as it imports exports, so it can test them directly:

<!-- test: cmd ant test session_test.ant -->

```ant [session_test.ant]
import "./session.ant" { Session, Token }

test "start issues a fresh token" {
  assert(Session.start().token == Token{ value: "abc" })
}
```

## Possible issues

A script cannot export. One top-level statement makes the file a script:

```ant
export let answer = 42
println(answer)
// error: a module with exports cannot have top-level statements
```

An exported function must state its return type, so the module's interface reads without its bodies:

```ant
export func double(n: int) { n * 2 }
// error: the exported function double must state its return type
```

A function exported on a private type is a compile error:

```ant
type Point = { x: int, y: int }

export func Point.origin(): Point { Point{ x: 0, y: 0 } }
// error: Point.origin is exported but Point is not
```

Imports are read-only. Only the module that declares an exported `var` assigns it, so change it
through an exported function such as `bump`:

```ant
import "./counter.ant" Counter
import "./counter.ant" { count }

count = 5         // error: cannot assign to the imported variable count
Counter.count = 5 // error: cannot assign to Counter.count: imports are read-only
```

`export` works only at the top level:

```ant
func answer(): int {
  export let x = 42 // error: 'export' is only allowed at the top level
  x
}
```

There is no re-export. A module that imports `bump` does not pass it on to its own importers, and
`export` takes a declaration, not a name:

```ant [twice.ant]
import "./counter.ant" { bump }

export func twice(): void {
  bump()
  bump()
}
```

```ant
import "./twice.ant" { twice, bump } // error: bump is not exported by ./twice.ant
```

```ant
import "./counter.ant" { bump }

export bump
// error: expected var, let, func, type, or interface after export, got identifier "bump"
```

## See also

[`import`](import.md), [`func`](func.md), [`type`](type.md), [`interface`](interface.md),
[`let`](let.md), [`var`](var.md), [`test`](test.md), [Modules](../../tour/19-modules.md),
[`ant test`](../../tools/test.md)
