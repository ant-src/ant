<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.isTerminal

```ant signature
func File.isTerminal(f: File.Handle): bool
```

File.isTerminal reports whether f is a terminal, as \[\[ -t 1 \]\] asks of standard output. It is
false for a closed handle. A buffered handle answers for the handle under it.

<!-- end generated -->

## Details

- `File.isTerminal` is false for a file, a pipe, and a network connection.
- When a program runs with its output redirected or piped, as in `ant run tool.ant > out.txt` or
  `ant run tool.ant | less`, `Os.stdout()` is not a terminal. Each stream answers for itself, so
  `Os.stderr()` can still be one.

## Basic examples

Ask whether standard output is a terminal. The examples on this page run with their output captured,
so it is not:

```ant
println(File.isTerminal(Os.stdout()))
```

```text
false
```

A file is never a terminal:

```ant
let f = File.create("notes.txt").unwrap()
defer { File.close(f) }
println(File.isTerminal(f))
```

```text
false
```

## Scope

Each of the three standard streams answers for itself, through `.`:

```ant
println(Os.stdin().isTerminal())
println(Os.stdout().isTerminal())
println(Os.stderr().isTerminal())
```

```text
false
false
false
```

Both ends of a pipe and both ends of a network connection are not terminals:

```ant
let pipe = Os.openPipe().unwrap()
defer {
  File.close(pipe.read)
  File.close(pipe.write)
}
println(pipe.read.isTerminal())
println(pipe.write.isTerminal())

let l = Net.listen("127.0.0.1:0").unwrap()
defer { Net.close(l) }
let conn = Net.dial(l.address()).unwrap()
defer { File.close(conn) }
println(conn.isTerminal())
```

```text
false
false
false
```

A buffered handle answers for the handle under it:

```ant
let out = File.buffered(Os.stdout())
defer { File.close(out) }
println(out.isTerminal())
```

```text
false
```

## Applications

Color the output only when a person is watching it. Captured, as here, the program prints plain
text; at a terminal it prints green and red:

```ant
func paint(text: string, code: int): string {
  if !Os.stdout().isTerminal() {
    return text
  }
  "\u{1b}[${code}m${text}\u{1b}[0m"
}

println(paint("ok", 32))
println(paint("failed", 31))
```

```text
ok
failed
```

Prompt only when the input comes from a keyboard. With input piped in, the prompt would only clutter
the output:

<!-- test: stdin
Ann
-->

```ant
func ask(question: string): Result[string, Error] {
  if Os.stdin().isTerminal() {
    print("${question} ")
  }
  let line = try File.readLine(Os.stdin())
  Result.ok(line ?? "")
}

let name = ask("Your name?").unwrap()
println("Hello, ${name}")
```

```text
Hello, Ann
```

Show progress on standard error, redrawn in place with `\r`, only when standard error is a terminal.
The result still goes to standard output either way:

```ant
let steps = ["fetch", "build", "test"]
let live = Os.stderr().isTerminal()
var done = 0
for step in steps {
  done += 1
  if live {
    eprint("\r${done}/${steps.length()} ${step}")
  }
}
if live {
  eprintln("")
}
println("${done} steps done")
```

```text
3 steps done
```

## Properties and relations

`File.buffered` does not change the answer, and a file is never a terminal:

```ant
let f = File.create("log.txt").unwrap()
defer { File.close(f) }
assert(!f.isTerminal())
assert(File.buffered(f).isTerminal() == f.isTerminal())
assert(File.buffered(Os.stdout()).isTerminal() == Os.stdout().isTerminal())
```

## Possible issues

Ask about the handle you write to. Progress written to standard error but guarded by
`Os.stdout().isTerminal()` disappears when the output is redirected to a file, although the person
is still watching standard error; guard it with `Os.stderr().isTerminal()` instead, as in
Applications.

## See also

[File](index.md) · [File.buffered](buffered.md) · [File.readLine](readLine.md) ·
[Os.stdout](../Os/stdout.md) · [Os.stderr](../Os/stderr.md) · [Os.stdin](../Os/stdin.md) ·
[Os.openPipe](../Os/openPipe.md)
