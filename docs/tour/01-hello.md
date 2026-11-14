<!-- status: reviewed -->

# Hello, Ant

```ant
//! Greets the world, the first program of the tour.

// The top level runs from top to bottom.
println("Hello, Ant")

/* print writes no line break,
   so the next output continues the line. */
print("Hello, ")
println("world")

// println() with no argument writes an empty line.
println()

// Any value prints, not only strings.
println(6 * 7)
println([1, 2, 3])
```

```text
Hello, Ant
Hello, world

42
[1, 2, 3]
```

This file is a whole program. It has no `main` function: a file with statements at its top level is
a script, and `ant` runs those statements in order. The program ends after the last one.

## Run it

Save the program as `hello.ant` and run it:

```sh
ant run hello.ant
```

`ant hello.ant` does the same. Before you run a larger program, `ant check` reads the whole file and
reports compile errors without running anything, and `ant fmt` rewrites the file into the one layout
every Ant file shares:

```sh
ant check hello.ant
ant fmt hello.ant
```

## Comments

`//` comments to the end of the line, and `/* ... */` comments a span, which may cover several
lines.

The `//!` line at the top describes the whole file. It goes before any code, and `ant doc` and your
editor show it.

## Run it as a command

A first line that starts with `#!` lets the system run the file directly:

```ant
#!/usr/bin/env ant
println("Hello from a command")
```

```text
Hello from a command
```

```sh
chmod +x greet.ant
./greet.ant
```

## Try it

- Change the greeting to your name.
- Replace `println(6 * 7)` with `println(7 / 2)` and see what an integer division prints.
- Misspell `println` as `printn` and run `ant check` to see a compile error.

## Learn more

- [Modules](../guide/modules.md): scripts, libraries, and how a program grows past one file.
- [println](../lib/prelude/println.md) and [print](../lib/prelude/print.md) write to standard
  output; [eprintln](../lib/prelude/eprintln.md) writes to standard error.
- [Comments](../language/lexical/comments.md), [ant run](../tools/run.md), and
  [ant check](../tools/check.md).
