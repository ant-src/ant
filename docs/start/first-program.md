<!-- status: reviewed -->

# Your first program

In this tutorial you write a small program, check it, format it, test it, and try Ant in the REPL.
You need `ant` on your path; [Install](install.md) shows how to get it. Make an empty directory and
work in it.

## Write a script

Save this as `hello.ant` and run it with `ant run hello.ant`:

<!-- test: cmd ant run hello.ant -->

```ant [hello.ant]
let name = "Ant"
println("Hello, ${name}")
```

```text
Hello, Ant
```

The file has statements at its top level, so it is a script: `ant run` runs them from top to bottom,
and the program ends after the last one. There is no `main` function to write. `ant hello.ant` does
the same as `ant run hello.ant`.

## Check it

Now make a mistake. Change the last line so that it adds a number to a string:

```ant
let name = "Ant"
println("Hello, " + 2) // error: + needs both operands of one type, got string and int
```

Run `ant check`:

```sh
ant check hello.ant
```

```text
hello.ant:2:19: + needs both operands of one type, got string and int
   │
 2 │ println("Hello, " + 2)
   │         ^^^^^^^^^^^^^

1 error
```

`ant check` reads the whole file, and every file it imports, reports every compile error, and exits
with a non-zero status. It runs nothing. `ant run` does not check types: it runs the program and
panics only when it reaches the bad line, after the lines before it have run. Run `ant check` while
you write, as your editor does, and `ant run` when you want to see the program work.

Undo the change before you go on.

## Format it

Every Ant file shares one layout, and `ant fmt` writes it. Say you typed this into `shout.ant` in a
hurry:

```text
func shout(s:string):string{s.toUpper()+"!"}
for w in ["hello","ant"]{println( shout(w) )}
```

`ant fmt` rewrites the file in place:

```sh
ant fmt shout.ant
```

The file now reads like this, and runs as it did before:

```ant
func shout(s: string): string { s.toUpper() + "!" }
for w in ["hello", "ant"] { println(shout(w)) }
```

```text
HELLO!
ANT!
```

It changes only layout (spaces, line breaks, semicolons at the end of a statement, and trailing
commas), never what the program means. `ant fmt .` formats every `.ant` file in the directory and
below it, and `ant fmt --check .` lists the files that are not formatted and changes nothing, which
is what continuous integration runs.

## Take arguments

A script reads its command-line arguments with `Os.args()`. Save this as `greet.ant`, and run it
with two names, `ant run greet.ant Ada Grace`:

<!-- test: args Ada Grace -->

```ant
func greeting(name: string): string {
  let who = name.trim()
  if who.isEmpty() { "Hello, world" } else { "Hello, ${who}" }
}

for name in Os.args() {
  println(greeting(name))
}
```

```text
Hello, Ada
Hello, Grace
```

`greeting` is a function: its parameter and its result have types, and its last expression is its
value. The `if` is an expression too, so it gives one of the two strings.

## Write a test

A test lives in a test file, whose name ends in `_test.ant`, and imports the code it tests. A script
cannot be imported, so move `greeting` into a library of its own, `greeting.ant`. A file with no
statements at its top level is a library, and `export` makes a name visible to the files that import
it:

```ant [greeting.ant]
// Returns a greeting for name, or for the world when name is blank.
export func greeting(name: string): string {
  let who = name.trim()
  if who.isEmpty() { "Hello, world" } else { "Hello, ${who}" }
}
```

`greet.ant` imports it, and runs as before:

<!-- test: cmd ant run greet.ant Ada Grace -->

```ant [greet.ant]
import "./greeting.ant" { greeting }

for name in Os.args() {
  println(greeting(name))
}
```

```text
Hello, Ada
Hello, Grace
```

Now save the tests as `greeting_test.ant`, next to the library. Each `test` block has a name, and
`assert` checks a condition. Run `ant test`:

<!-- test: cmd ant test . -->

```ant [greeting_test.ant]
import "./greeting.ant" { greeting }

test "greets by name" {
  assert(greeting("Ada") == "Hello, Ada")
}

test "trims the name" {
  assert(greeting("  Ada ") == "Hello, Ada")
}

test "greets the world without a name" {
  assert(greeting("") == "Hello, world")
}
```

```text
ok    3 passed (0.00s)
```

`ant test` finds every test file in the directory and below it, runs its tests, and prints one line
when they all pass.

Break the library to see a failure: change `"Hello, world"` to `"Hi, world"` and run `ant test`
again:

```sh
ant test
```

```text
--- FAIL: greeting_test.ant: greets the world without a name (0.00s)
    assertion failed: greeting("") == "Hello, world" ("Hi, world" == "Hello, world")
      at test "greets the world without a name" (greeting_test.ant:12:9)
FAIL  1 failed, 2 passed (0.00s)
```

The report names the test, quotes the condition, and shows both sides of the comparison. `ant test`
then exits with status 1, so a script or continuous integration sees the failure. Put the greeting
back before you go on.

Unlike `ant run`, `ant test` checks types first: it checks the tests and the library they import,
and runs nothing when the checker reports an error.

## Try the REPL

Run `ant` with no file to start the REPL:

```sh
ant
```

Type an input after the `>` prompt. The REPL runs it and prints its value:

```text
> let xs = [3, 1, 2]
> xs.sort()
[1, 2, 3]
> _.length()
3
> 7 / 2
3
```

A declaration such as `let` prints nothing and stays for the inputs after it. A name bound by `let`
stays bound for the whole session: you cannot declare it again or assign it. Use `var` for a name
you want to declare again as you experiment. `_` is the value the last input printed. The REPL is a
quick way to try a function before you put it in a file.

Type `:quit` to leave the REPL, or press Ctrl-D on an empty line. Type `:help` to see the commands
and editing keys.

## Next

- The [Tour](../tour/01-hello.md) teaches the language one concept at a time.
- [Idiomatic Ant](../guide/idiomatic.md) shows how Ant code is written.
- [Modules](../guide/modules.md) and [Testing](../guide/testing.md) explain scripts, libraries, and
  tests in depth.
- [ant run](../tools/run.md), [ant check](../tools/check.md), [ant fmt](../tools/fmt.md), and
  [ant test](../tools/test.md) list every flag.
