<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.program

```ant signature
func Os.program(): string
```

Os.program returns the path of the program's root module as the command was given it, as a shell's
$0 is: ./tool.ant when a \#! line runs it, tool.ant after ant run tool.ant. It is "" in the REPL.

<!-- end generated -->

## Details

- The path is the one the command named, as it was typed: `Os.program` does not make it absolute or
  clean it, so `ant run ./tool.ant` gives `"./tool.ant"`.
- It names the script, the file `ant run` started, even when a module that the script imports calls
  it.

## Basic examples

Print the path of the running program, here run as `ant run main.ant`:

```ant
println(Os.program())
```

```text
main.ant
```

Run as `ant run ./main.ant`, the path keeps the `./`:

<!-- test: cmd ant run ./main.ant -->

```ant
println(Os.program())
```

```text
./main.ant
```

## Scope

A program in another directory is named with that directory, relative to where the command ran:

<!-- test: cmd ant run tools/report.ant -->

```ant [tools/report.ant]
println(Os.program())
println(Path.base(Os.program()))
```

```text
tools/report.ant
report.ant
```

An imported module sees the path of the program that imports it, not its own:

```ant [log.ant]
// prefix returns msg prefixed with the program's name.
export func prefix(msg: string): string { "${Path.base(Os.program())}: ${msg}" }

// name returns the path the program was started with.
export func name(): string { Os.program() }
```

```ant
import "./log.ant" { name, prefix }

println(prefix("starting"))
println(name())
```

```text
main.ant: starting
main.ant
```

## Applications

Name the script in its usage line, as it was run, so the line is right however the file is named:

<!-- test: exit 2 -->

```ant
let args = Os.args()
if args.isEmpty() {
  eprintln("usage: ${Path.base(Os.program())} <file>...")
  Os.exit(2)
}
println(args)
```

Prefix error messages with the program's name, as Unix tools do, so a message read in a log or a
pipeline says where it came from:

<!-- test: exit 1 -->

```ant
func countLines(path: string): Result[int, Error] {
  let text = try File.read(path)
  Result.ok(text.lines().length())
}

let lines = countLines("missing.txt")
if lines is Err {
  eprintln("${Path.base(Os.program())}: ${lines.error.message}")
  Os.exit(1)
}
println(lines.value)
```

Read a file kept next to the script, wherever the command runs from. [Path.dir](../Path/dir.md) of
the program path is the script's directory:

<!-- test: file tools/words.txt
ant
bee
-->
<!-- test: cmd ant run tools/count.ant -->

```ant [tools/count.ant]
let dir = Path.dir(Os.program())
let words = try File.read(Path.join([dir, "words.txt"]))
println("${words.lines().length()} words in ${dir}")
```

```text
2 words in tools
```

## Properties and relations

The program path is not one of the [Os.args](args.md), and it is relative when the command named a
relative path:

<!-- test: args a b -->

```ant
assert(Os.program() == "main.ant")
assert(Os.args() == ["a", "b"])
assert(!Os.args().contains(Os.program()))
assert(!Path.isAbs(Os.program()))
```

## Possible issues

The path is relative to the directory the command ran from. A child started in another directory,
with `dir:`, reads it against that directory instead and finds a different file or none. Make the
path absolute first with [Path.abs](../Path/abs.md):

```ant
assert(!Path.isAbs(Os.program()))
let script = try Path.abs(Os.program())
assert(Path.isAbs(script))
```

Unlike `argv[0]` in C or `sys.argv[0]` in Python, `Os.args()[0]` is the first argument, not the
program path. Use `Os.program` for the path:

<!-- test: args input.txt -->

```ant
println(Os.args()[0])
println(Os.program())
```

```text
input.txt
main.ant
```

## See also

[Os.args](args.md) · [Os.exit](exit.md) · [Os.cwd](cwd.md) · [Path.base](../Path/base.md) ·
[Path.dir](../Path/dir.md) · [Path.abs](../Path/abs.md) · [eprintln](../prelude/eprintln.md)
