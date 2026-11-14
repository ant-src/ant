<!-- status: reviewed -->

# Modules

Every file is a module. A file that only declares things is a library; `export` decides which of its
names other files may use.

```ant [geometry.ant]
//! Points on a plane.

export type Point = { x: float64, y: float64 }

export func Point.distance(a: Point, b: Point): float64 {
  Math.sqrt(square(a.x - b.x) + square(a.y - b.y))
}

export func origin(): Point {
  Point{ x: 0.0, y: 0.0 }
}

func square(n: float64): float64 {
  n * n
}
```

A file with a statement at its top level, such as a call to `println`, is a script. A script is the
program you run, and it imports the libraries it needs:

```ant
import "./geometry.ant" { Point, origin }
import "./geometry.ant" Geometry

let p = Point{ x: 3.0, y: 4.0 }
let q: Geometry.Point = origin()
println(p.distance(q))
println(Point.distance(p, Geometry.origin()))
```

```text
5.0
5.0
```

Save the script as `main.ant` next to `geometry.ant` and run `ant run main.ant`. The path in an
`import` is relative to the importing file, extension included. There are no packages: a program is
the script and the files it imports.

The first `import` picks names out of the module with braces, the way you destructure a record, and
`{ origin: zero }` would rename one. The second binds the whole module to one name, `Geometry`, and
you reach its exports with `.`. Importing `Point` brings the exported functions declared on it
along, so both `p.distance(...)` and `Point.distance(...)` work. Both imports name one module, so
`Geometry.Point` and `Point` are the same type.

`square` has no `export`, so it stays private to `geometry.ant`:

```ant
import "./geometry.ant" { square } // error: square is not exported by ./geometry.ant
```

A script cannot export anything, since nothing can import it:

```ant
export let answer = 42
println(answer)
// error: a module with exports cannot have top-level statements
```

The reverse holds too: one top-level statement turns a file into a script, and no other file can
import it. A helper file with a stray `println` is a script:

```ant [hello.ant]
println("hi")
```

```ant
import "./hello.ant" Hello // error: it is a script
```

The guide chapter [Modules](../guide/modules.md) covers namespaces, how modules are initialized, and
how to lay out a project. The keywords have their own pages:
[`import`](../language/keywords/import.md) and [`export`](../language/keywords/export.md).
[`ant run`](../tools/run.md) runs a script.
