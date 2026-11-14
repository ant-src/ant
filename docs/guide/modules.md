<!-- status: reviewed -->

# Modules

Every `.ant` file except a declaration file is a module. A program is one script, the file you run,
and the libraries it imports. There are no packages and no search path: an `import` names a file by
its path, relative to the file that holds the `import`.

This chapter builds one small program, a word counter, and uses it to show how the parts fit.

## Scripts and libraries

A file that only declares things is a library. `export` marks the names other files may use:

```ant [text.ant]
//! Text counts the words of a text.

export type Count = { word: string, n: int }

export func Count.label(c: Count): string {
  "${c.word} ${c.n}"
}

export func counts(text: string): Array[Count] {
  words(text)
    | Array.groupBy({ x })
    | Map.entries()
    | Array.map({ Count{ word: x.key, n: x.value.length() } })
    | Array.sortBy({ -x.n })
}

func words(text: string): Array[string] {
  text.toLower().fields()
}
```

A file with a statement at its top level is a script. Here the `for` loop makes it one:

<!-- test: file poem.txt
the cat and the hat
The end
-->
<!-- test: args poem.txt -->

```ant
//! Prints the words of each file given as an argument, most common first.

import "./text.ant" Text

for path in Os.args() {
  let content = try File.read(path)
  let line = Text.counts(content) | Array.map({ x.label() }) | String.join(", ")
  println("${path}: ${line}")
}
```

```text
poem.txt: the 3, cat 1, and 1, hat 1, end 1
```

Save the script as `main.ant` next to `text.ant` and run `ant run main.ant poem.txt`, or
`ant main.ant poem.txt` for short. `Os.args()` holds the arguments after the script's name.

The top level of a script runs like the body of a function. `return` ends the program early, `defer`
runs at the end, and `try` returns an `Err` from it. A script that returns an `Err` prints the error
to standard error and exits with status 1. Here the second file is missing, so the program prints
the first line and then stops with `open missing.txt: no such file or directory`:

<!-- test: args poem.txt missing.txt -->
<!-- test: exit 1 -->

```ant
import "./text.ant" Text

for path in Os.args() {
  let content = try File.read(path)
  let line = Text.counts(content) | Array.map({ x.label() }) | String.join(", ")
  println("${path}: ${line}")
}
```

```text
poem.txt: the 3, cat 1, and 1, hat 1, end 1
```

The split is strict, in both directions. Nothing can import a script, not even a test file, and a
script cannot export. So put the logic in libraries and keep the script thin: it reads the
arguments, calls the libraries, and prints. Everything worth testing then lives where a test file
can import it.

One stray statement turns a library into a script. In Python or JavaScript a module often prints or
registers something when it loads; in Ant that statement makes the file a script, and its exports
are an error:

```ant
export func greet(name: string): string {
  "Hello, ${name}"
}

println("greet loaded")
// error: a module with exports cannot have top-level statements; this statement makes it a script
```

A library's `var` and `let` initializers still run when it loads, and may call any function. Keep
them to building values.

## Imports and exports

An `import` names a file and then says how to bind it. Braces pick out names, the way you
destructure a record. A bare name binds the whole module as a namespace:

```ant
import "./text.ant" { Count, counts }
import "./text.ant" Text

let best: Count = counts("to be or not to be")[0]
println(best.label())
println(Text.counts("a b b").length())
```

```text
to 2
2
```

Both imports name one module, so `Count` and `Text.Count` are the same type. Pick by how the code
reads:

- **Braces** for a few names you use often, and for types you write in annotations.
- **A namespace** when the names read better with their origin, as `Text.counts` does, or when you
  use many of them.
- **`name: alias`** inside braces when two modules export the same name. Every name in a file has
  one origin, so two imports cannot bind the same name, and a top-level declaration cannot reuse an
  imported one.

Renaming an import changes only the local name. The functions declared on a renamed type follow the
alias:

```ant
import "./text.ant" { counts: countWords, Count: C }

let best: C = countWords("a b b")[0]
println(C.label(best))
```

```text
b 2
```

Importing a type brings the exported functions declared on it along. You need no import at all to
call one through `.`: a value finds the functions of its own type, wherever the value came from.
Here the script imports only `counts`, and `label` still works:

```ant
import "./text.ant" { counts }

let best = counts("to be or not to be")[0]
println(best.label())
```

```text
to 2
```

A function that is not declared on a type is reached through its module, never through `.` on a
value. `counts(s)` or `Text.counts(s)` works; `s.counts()` looks in the namespace of `string`, which
is `String`.

A name without `export` is private to its module. Privacy applies to names, not to values: an
exported function may return a value of a private type, which the importer holds but cannot name.

```ant
import "./text.ant" { words } // error: words is not exported by ./text.ant
```

Imports are live and read-only. An imported `var` reads the exporting module's current value each
time, but only that module can assign it; export a function that changes it. There is no re-export:
a module that imports `Count` does not pass it on, so each file imports a name from the module that
declares it. The [`import`](../language/keywords/import.md) and
[`export`](../language/keywords/export.md) pages show each of these with its error.

## Namespaces

`Text.counts`, `Count.label`, and `Array.map` are the same operation: a field read on a namespace.
Three things are namespaces, and nothing else:

- **A type.** `Count` holds the functions declared on it, such as `Count.label`.
- **A module imported under a name.** `Text` holds the module's exports.
- **A namespace of the prelude.** `Array`, `String`, `Os`, `Time`, and the rest hold the standard
  library.

A namespace is not a value. You cannot store one in a variable or pass it to a function:

```ant
import "./text.ant" Text

let t = Text // error: Text is a namespace, not a value
```

To pass a module's behavior around, pass its functions, such as `Text.counts`, or declare an
[interface](interfaces.md).

The prelude's names are in every module, and you cannot shadow them, not even in an inner scope. A
local named `len`, `int`, or `Array` is an error, so these names mean the same thing everywhere:

```ant
func count(xs: Array[string]): int {
  let len = xs.length() // error: len is a prelude name and cannot be redeclared
  len
}
```

## How a program starts

`ant run` loads the script's imports depth first, in source order, and initializes each module once,
after everything it imports. A module imported by several files still runs its initializers once.
The script runs last.

Inside a module, `func` and `type` declarations can appear in any order, so functions may call each
other freely. The other items, `var` and `let` initializers and a script's statements, run in source
order, and one that refers to a later variable is an error:

```ant
println(limit)
// error: limit is declared later in the module; a top-level initializer or statement may refer only to the variables above it
let limit = 3
```

The type checker does not follow calls, so a function that reads the variable gets past it, and the
read panics instead:

```ant
func show(): void {
  println(limit) // panic: limit is read before its initializer has run
}

show()
let limit = 3
```

Two modules that import each other, directly or through others, are an import cycle, which `ant`
reports before any module runs. Move what both need into a third module that both import.

## Private imports in test files

A file whose name ends in `_test.ant` is a test file. Its imports reach every top-level declaration
of a module, exported or not, so a test can call `words` directly:

<!-- test: cmd ant test . -->

```ant [text_test.ant]
import "./text.ant" { counts, words }
import "./helpers_test.ant" { labels }

test "words ignores case" {
  assert(words("The the") == ["the", "the"])
}

test "counts puts the most common word first" {
  assert(labels(counts("a b b")) == ["b 2", "a 1"])
}
```

```text
ok    2 passed (0.00s)
```

Only a test file can import a test file, so nothing private leaks into a program. Helpers that
several test files share go in a test file of their own, such as `helpers_test.ant`:

```ant [helpers_test.ant]
import "./text.ant" { Count }

export func labels(cs: Array[Count]): Array[string] {
  cs.map({ x.label() })
}
```

Since a script cannot be imported, a test cannot reach its code. That is one more reason to keep
scripts thin. The [Testing](testing.md) chapter covers tests themselves.

## Declaration files

A file whose name ends in `.d.ant` is a declaration file. It lists the signatures of names that are
bound outside the language, with their doc comments and without bodies. The prelude is one,
`prelude.d.ant`: it is what your editor opens when you jump to `Array.map`, and what
`ant doc Array.map` prints from.

You write one only when a host program that embeds Ant adds a namespace of its own to the prelude.
The declaration file then tells the type checker, `ant doc`, and editors what that namespace holds:

<!-- test: cmd ant doc store.d.ant Store.get -->

```ant [store.d.ant]
// Store
//
// Store is a key-value store the host program provides.

// Store.Handle is an open store.
type Store.Handle

// Store.open opens the store with the given name.
func Store.open(name: string): Result[Store.Handle, Error]

// Store.get returns the value stored under key, or null when there is none.
func Store.get(h: Store.Handle, key: string): string | null

// Store.LIMIT is the longest value a store holds, in bytes.
let Store.LIMIT: int
```

```text
func Store.get(h: Store.Handle, key: string): string | null

    Store.get returns the value stored under key, or null when there is none.
```

Each declaration names its owner, `Store` here, which makes it a namespace. A type without a body,
such as `Store.Handle`, is opaque: its values come only from the functions declared on it. There is
no `export`, since every declaration is visible, and no `import`. `ant check store.d.ant` checks
that the signatures are sound.

You cannot import a declaration file, and `ant run` refuses one:

```ant
import "./store.d.ant" Store
// error: cannot import "./store.d.ant": a declaration file declares names for the prelude and cannot be imported
```

## Laying out a project

A project is a directory of files that import each other by path. Nothing else decides its shape, so
this layout is a convention, not a rule:

```text
wordcount/
  main.ant            the script: arguments, output, exit status
  text.ant            a library: the logic
  text_test.ant       tests for text.ant, private names included
  helpers_test.ant    helpers shared by test files
  tools/
    top.ant           another script, sharing text.ant
```

- Keep a test file next to the library it tests. `ant test` with a directory runs every test file
  under it, so `ant test .` runs them all.
- A program may have several scripts, one per command, all importing the same libraries. A path is
  relative to the file that holds the `import`, so `tools/top.ant` reaches up with `..`:

  ```ant [tools/top.ant]
  import "../text.ant" { counts }

  let best = counts("to be or not to be")[0]
  println(best.label())
  ```

- Give a script `#!/usr/bin/env ant` as its first line and make it executable (`chmod +x main.ant`)
  to run it as `./main.ant`; see [Run a script as a command](../howto/cli/shebang.md).
- Start each library with a `//!` doc comment that says what it is for. `ant doc text.ant` prints it
  with the module's exports; see [Documenting code](docs.md).

## Coming from other languages

- **No packages.** There is no `node_modules`, `GOPATH`, or `sys.path`. An import is a path to a
  file, extension included: `"./text.ant"`, not `"./text"` or `"text"`.
- **No `if __name__ == "__main__"`.** A file is a script or a library, never both. Move the code you
  want to reuse into a library, and import it from a thin script.
- **No index files or barrels.** There is no re-export, so a module cannot gather the exports of
  others. Import each name from the module that declares it.
- **No `import *`.** Name what you import, or bind the module as a namespace.
- **No circular imports.** Python and JavaScript tolerate some cycles; Ant rejects all of them.
- **No load-time statements.** A top-level statement, such as a call that prints or registers
  something, makes the file a script, and a script cannot be imported. Keep initializers to building
  values; the script does the work.
- **No shadowed builtins.** `len`, `int`, `print`, and the standard namespaces cannot be redeclared
  anywhere.

## See also

[Modules in the tour](../tour/19-modules.md), [`import`](../language/keywords/import.md),
[`export`](../language/keywords/export.md), [`test`](../language/keywords/test.md),
[Testing](testing.md), [`ant run`](../tools/run.md), [`ant test`](../tools/test.md),
[`ant doc`](../tools/doc.md), [Run Ant from a Go program](../howto/embedding/run-from-go.md)
