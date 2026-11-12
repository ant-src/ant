<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.args

```ant signature
func Os.args(): Array[string]
```

Os.args returns the arguments after the program path: ant run main.ant a b gives \["a", "b"\]. Each
call returns a new array, which the caller may change without changing the arguments. It is \[\] in
the REPL.

<!-- end generated -->

## Details

- The array holds only the arguments, not the program path, so the first argument is at index 0. Use
  [Os.program](program.md) for the path.
- `Os.args` does not parse anything: an option such as `-v` or `--out=x.txt` arrives as a plain
  string, in the order given.
- Each argument is decoded as UTF-8, and each ill-formed piece becomes U+FFFD. `Os.args` has no
  `Bytes` form, so an argument that is not UTF-8 cannot be read as it was given.

## Basic examples

Print the arguments a program receives, here run as `ant run main.ant red green`:

<!-- test: args red green -->

```ant
println(Os.args())
```

```text
["red", "green"]
```

Index the array for one argument. Test the length first, since an index past the end panics:

<!-- test: args Ann -->

```ant
let args = Os.args()
if args.length() != 1 {
  eprintln("usage: greet <name>")
  Os.exit(2)
}
println("Hello, ${args[0]}")
```

```text
Hello, Ann
```

## Scope

Without arguments, the array is empty:

```ant
let args = Os.args()
println(args)
println(args.isEmpty())
```

```text
[]
true
```

Options and plain words arrive alike, as strings, in the order given:

<!-- test: args -v --out=report.txt data.csv -->

```ant
for arg in Os.args() {
  println(arg)
}
```

```text
-v
--out=report.txt
data.csv
```

Arguments are Unicode strings, so string functions count characters, not bytes:

<!-- test: args café 東京 -->

```ant
for arg in Os.args() {
  println("${arg}: ${arg.length()}")
}
```

```text
café: 4
東京: 2
```

## Applications

Add up the numbers given as arguments, and reject one that is not a number:

<!-- test: args 3 4 5 -->

```ant
var total = 0
for arg in Os.args() {
  let n = arg.toInt()
  if n == null {
    eprintln("not a number: ${arg}")
    Os.exit(2)
  }
  total += n
}
println(total)
```

```text
12
```

Separate flags from file names:

<!-- test: args -v notes.txt todo.txt -->

```ant
let args = Os.args()
let verbose = args.contains("-v")
let files = args.filter({ !x.startsWith("-") })
println("verbose: ${verbose}")
println("files: ${files}")
```

```text
verbose: true
files: ["notes.txt", "todo.txt"]
```

Choose a subcommand by the first argument, and pass the rest on:

<!-- test: args add milk eggs -->

```ant
let args = Os.args()
if args.isEmpty() {
  eprintln("usage: ${Path.base(Os.program())} add <item>... | list")
  Os.exit(2)
}
let rest = args.slice(1, args.length())
if args[0] == "add" {
  println("adding ${rest | String.join(", ")}")
} else if args[0] == "list" {
  println("listing")
} else {
  eprintln("unknown command: ${args[0]}")
  Os.exit(2)
}
```

```text
adding milk, eggs
```

Count the lines of each file named on the command line. `try` ends the script with the error when a
file cannot be read:

<!-- test: file a.txt
one
two
-->
<!-- test: file b.txt
three
-->
<!-- test: args a.txt b.txt -->

```ant
var total = 0
for path in Os.args() {
  let text = try File.read(path)
  let n = text.lines().length()
  total += n
  println("${n} ${path}")
}
println("${total} total")
```

```text
2 a.txt
1 b.txt
3 total
```

## Properties and relations

The arguments are exactly the words after the program path, which [Os.program](program.md) gives on
its own:

<!-- test: args a b c -->

```ant
assert(Os.args() == ["a", "b", "c"])
assert(Os.program() == "main.ant")
```

## Possible issues

Unlike `argv` in C or `sys.argv` in Python, the array does not start with the program path. Index 0
is the first argument, and with no arguments it panics:

```ant
let file = Os.args()[0] // panic: index 0 out of range for an array of length 0
println(file)
```

Use [Array.first](../Array/first.md) for an argument that may be missing:

```ant
let file = Os.args().first() ?? "data.csv"
println(file)
```

```text
data.csv
```

Every argument is a string. Convert a number with [String.toInt](../String/toInt.md), which gives
`null` for text that is not one. Give a default for a missing argument, and reject one that is not a
number:

```ant
let count: int = Os.args()[0] // error: cannot use string where int is expected
```

<!-- test: args 3 -->

```ant
let arg = Os.args().first()
let count = if arg == null { 1 } else { arg.toInt() }
if count == null {
  eprintln("not a number: ${arg}")
  Os.exit(2)
}
println(count * 2)
```

```text
6
```

## See also

[Os.program](program.md), [Os.exit](exit.md), [Os.env](env.md), [Os.stdin](stdin.md),
[Array.first](../Array/first.md), [String.toInt](../String/toInt.md),
[eprintln](../prelude/eprintln.md)
