<!-- status: reviewed -->

# `import`

`import` brings the exports of another module into a file, either as chosen names or as one
namespace.

## Details

- The path is relative to the importing file and includes the extension. There are no packages: a
  program is its script and the files it imports.
- Imports come first in a file, before any declaration or statement.
- An imported name is read-only, and live: it reads the exporting module's current value each time.
- Each module is initialized once, after the modules it imports, however many files import it. Two
  modules that import each other, directly or through others, are an import cycle, which is an
  error.
- Only a library can be imported, and a test file only by another test file. A file with a top-level
  statement is a script, and no file can import it.

## Basic examples

A library exports names with [`export`](export.md):

```ant [geometry.ant]
//! Points on a plane.

export type Point = { x: float64, y: float64 }

export func Point.distance(a: Point, b: Point): float64 {
  Math.sqrt(square(a.x - b.x) + square(a.y - b.y))
}

export let origin = Point{ x: 0.0, y: 0.0 }

func square(n: float64): float64 {
  n * n
}
```

Braces pick names out of the module, the way you destructure a record:

```ant
import "./geometry.ant" { Point, origin }

let p = Point{ x: 3.0, y: 4.0 }
println(p.distance(origin))
```

```text
5.0
```

A bare name binds the whole module as a namespace, and `.` reads its exports:

```ant
import "./geometry.ant" Geometry

let p = Geometry.Point{ x: 3.0, y: 4.0 }
println(p.distance(Geometry.origin))
```

```text
5.0
```

## Scope

### Renaming

`name: alias` binds an export under another name. Importing a type brings the exported functions
declared on it along, under the new name too:

```ant
import "./geometry.ant" { Point: P, origin: zero }

let p = P{ x: 6.0, y: 8.0 }
println(P.distance(p, zero))
```

```text
10.0
```

### One module, several imports

Two imports of one file name one module, so its types are the same type under every name:

```ant
import "./geometry.ant" { Point }
import "./geometry.ant" Geometry

let q: Geometry.Point = Point{ x: 3.0, y: 4.0 }
println(Point.distance(q, Geometry.origin))
```

```text
5.0
```

### Paths in other directories

A path is relative to the file that holds the `import`, not to the script. Here `text/words.ant`
imports its neighbor as `./case.ant`:

```ant [text/case.ant]
export func shout(s: string): string {
  s.toUpper()
}
```

```ant [text/words.ant]
import "./case.ant" { shout }

export func loud(words: Array[string]): string {
  words | Array.map(shout) | String.join(" ")
}
```

```ant
import "./text/words.ant" { loud }

println(loud(["hello", "ant"]))
```

```text
HELLO ANT
```

### Imported variables

A module may export a `var`. The importer reads its current value each time, but only the exporting
module can assign it, here through `bump`:

```ant [counter.ant]
export var count = 0

export func bump(): void {
  count += 1
}
```

```ant
import "./counter.ant" { count, bump }

println(count)
bump()
bump()
println(count)
```

```text
0
2
```

### Initialization order

A module is initialized after everything it imports, depth first in source order, and only once.
Here `b.ant` runs its initializers before `a.ant`, and the second import of `b.ant` does not run
them again:

```ant [b.ant]
export let b = loaded("b")

func loaded(name: string): string {
  println("init ${name}")
  name
}
```

```ant [a.ant]
import "./b.ant" { b }

export let a = loaded("a")

func loaded(name: string): string {
  println("init ${name}")
  name
}
```

```ant
import "./a.ant" { a }
import "./b.ant" { b }

println("main ${a} ${b}")
```

```text
init b
init a
main a b
```

### Private names in test files

A test file may import the private names of any module, such as `square` here. Nothing but a test
file can import a test file, so the private names stay private in the program:

<!-- test: cmd ant test . -->

```ant [geometry_test.ant]
import "./geometry.ant" { square }

test "square" {
  assert(square(3.0) == 9.0)
}
```

## Possible issues

A name without `export` is private to its module, by braces and by `.` alike:

```ant
import "./geometry.ant" { square } // error: square is not exported by ./geometry.ant
import "./geometry.ant" Geometry

println(Geometry.square(2.0)) // error: square is not exported by Geometry
```

The path needs its extension:

```ant
import "./counter" { count } // error: cannot find module "./counter"
```

Imports are read-only, through braces and through a namespace:

```ant
import "./counter.ant" { count }
import "./geometry.ant" Geometry

count = 5 // error: cannot assign to the imported variable count
Geometry.origin = Geometry.Point{ x: 1.0, y: 1.0 }
// error: cannot assign to Geometry.origin: imports are read-only
```

Every name in a file has one origin. Two imports cannot bind the same name, and a top-level
declaration cannot reuse an imported one; rename the import instead:

```ant
import "./counter.ant" { count }
import "./geometry.ant" { origin: count } // error: count is already bound by an import

let total = count
```

```ant
import "./counter.ant" { count }

let count = 1 // error: count is already bound by an import
```

An `import` after a declaration is an error:

```ant
let start = 1
import "./counter.ant" { count } // error: imports must come before declarations
```

A file with a top-level statement is a script, and cannot be imported. Move the code worth sharing
into a library:

```ant [hello.ant]
println("hi")
```

```ant
import "./hello.ant" Hello
// error: cannot import "./hello.ant": it is a script, made one by the statement at hello.ant:1:1
```

A test file is a library, but only another test file can import it. Put helpers that the program
needs in a plain library:

```ant [helpers_test.ant]
export func twice(n: int): int {
  n * 2
}
```

```ant
import "./helpers_test.ant" { twice }
// error: cannot import "./helpers_test.ant": it is a test file, which only another test file can import
```

A module's imports are not passed on: there is no re-export. Here `shapes.ant` imports `Point` but
does not export it:

```ant [shapes.ant]
import "./geometry.ant" { Point }

export func make(x: float64, y: float64): Point {
  Point{ x, y }
}
```

```ant
import "./shapes.ant" { Point, make } // error: Point is not exported by ./shapes.ant
```

Import `Point` from `geometry.ant` itself, next to `make` from `shapes.ant`.

Two modules that import each other are an import cycle, which `ant` rejects before any module runs.
When `x.ant` starts with `import "./y.ant" { y }` and `y.ant` with `import "./x.ant" { x }`, a
program that imports `x.ant` fails with `y.ant:1:1: import cycle through x.ant`. Move the
declarations both need into a third module that both import.

## See also

[`export`](export.md), [`test`](test.md), [Modules](../../tour/19-modules.md),
[Modules guide](../../guide/modules.md), [`ant run`](../../tools/run.md)
