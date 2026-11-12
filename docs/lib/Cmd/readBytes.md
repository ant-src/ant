<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.readBytes

```ant signature
func Cmd.readBytes(c: Cmd | Cmd.Pipeline, input: string | Bytes = "", ctx: Context | null = null): Result[Bytes, Error]
```

Cmd.readBytes is [`Cmd.read`](read.md) that returns standard output as the command wrote it.

<!-- end generated -->

## Details

- The output comes back exactly as the command wrote it: a final newline is kept, and bytes that are
  not valid UTF-8 stay as they are. [`Cmd.read`](read.md) removes one final newline and turns
  invalid bytes into U+FFFD.
- A nonzero status is `Err`, with the message `<program>: exit status <n>`, followed by standard
  error when the command wrote any. Use [`Cmd.capture`](capture.md) when a nonzero status is an
  answer, not a failure.
- Standard error is captured, not passed through: on success it is dropped, and on a nonzero status
  it follows the message in the `Err`. Use [`Cmd.capture`](capture.md) to keep it.
- `input` is the command's standard input: a string as its UTF-8, `Bytes` as they are.
- The whole output is held in memory. For a large or endless output, use [`Cmd.start`](start.md)
  with a pipe and read it as it comes.

## Basic examples

Run a command and take its standard output as bytes. `Bytes` prints as hexadecimal, and the final
newline is there:

```ant
let bytes = try Cmd.readBytes(Cmd.new("printf", ["hi\\n"]))
println(bytes)
println(bytes.length())
```

```text
Bytes(68 69 0a)
3
```

Give the command its standard input:

```ant
println(Cmd.readBytes(Cmd.new("tr", ["a-z", "A-Z"]), input: "ant\n"))
```

```text
Ok{ value: Bytes(41 4e 54 0a) }
```

## Scope

Bytes that are not valid UTF-8 come back unchanged, where [`Cmd.read`](read.md) replaces them with
U+FFFD. `printf` writes the octal escape `\351` as the single byte `e9`:

```ant
let cafe = Cmd.new("printf", ["caf\\351"])
println(cafe.readBytes())
println(cafe.read())
```

```text
Ok{ value: Bytes(63 61 66 e9) }
Ok{ value: "caf�" }
```

`input` takes `Bytes` too, passed to the command as they are:

```ant
let data = Bytes.fromArray([0, 255, 10])
println(Cmd.new("cat").readBytes(input: data))
```

```text
Ok{ value: Bytes(00 ff 0a) }
```

A string `input` arrives as its UTF-8 encoding:

```ant
println(Cmd.new("cat").readBytes(input: "é"))
```

```text
Ok{ value: Bytes(c3 a9) }
```

Line endings stay as written, `\r\n` included:

```ant
println(Cmd.new("printf", ["ok\\r\\n"]).readBytes())
```

```text
Ok{ value: Bytes(6f 6b 0d 0a) }
```

A nonzero status is an `Err` that carries standard error:

```ant
println(Cmd.new("sh", ["-c", "echo disk full >&2; exit 3"]).readBytes())
```

```text
Err{ error: Error{ message: "sh: exit status 3: disk full" } }
```

Standard error is captured even when the command succeeds, and then it is dropped:

```ant
println(Cmd.new("sh", ["-c", "echo out; echo warn >&2"]).readBytes())
```

```text
Ok{ value: Bytes(6f 75 74 0a) }
```

A program that cannot be found is an `Err` too:

```ant
println(Cmd.new("no-such-program").readBytes())
```

```text
Err{ error: Error{ message: "exec: \"no-such-program\": executable file not found in $PATH" } }
```

A pipeline reads the output of its last stage:

```ant
let firstTwo = Cmd.pipe([
  Cmd.new("printf", ["\\377\\330\\377\\340"]),
  Cmd.new("head", ["-c", "2"]),
])
println(firstTwo.readBytes())
```

```text
Ok{ value: Bytes(ff d8) }
```

A pipeline fails when any stage fails, as under bash's `pipefail`; the `Err` names the stage's
program:

```ant
println(Cmd.pipe([Cmd.new("false"), Cmd.new("cat")]).readBytes())
```

```text
Err{ error: Error{ message: "false: exit status 1" } }
```

Call it in a pipe, with the command first:

```ant
let size = Cmd.new("printf", ["hello"]) | Cmd.readBytes() | Result.map(Bytes.length)
println(size)
```

```text
Ok{ value: 5 }
```

## Named arguments

### input

`input` defaults to `""`, so a command that reads standard input sees it end at once:

```ant
println(Cmd.new("cat").readBytes())
println(Cmd.new("cat").readBytes(input: "abc"))
```

```text
Ok{ value: Bytes() }
Ok{ value: Bytes(61 62 63) }
```

### ctx

A context bounds the command. When it ends first, the command is killed and the result is `Err` with
the context's reason:

```ant
let ctx = Context.new(timeout: Time.duration(milliseconds: 100))
println(Cmd.new("sleep", ["5"]).readBytes(ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
```

A context that is already done starts nothing:

```ant
let ctx = Context.new()
ctx.cancel()
println(Cmd.new("echo", ["never"]).readBytes(ctx: ctx))
```

```text
Err{ error: Error{ message: "cancelled" } }
```

## Applications

Compute the SHA-256 digest of what a command prints:

```ant
let digest = try Cmd.new("printf", ["abc"]).readBytes()
  | Crypto.sha256()
  | Encoding.toHex()
println(digest)
```

```text
ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
```

Compress text with `gzip` and decompress it again. The compressed data is binary, so it goes through
`Bytes` both ways:

```ant
let compressed = try Cmd.new("gzip", ["-c", "-n"]).readBytes(input: "hello, ant\n")
println(compressed.startsWith(Bytes.fromArray([0x1f, 0x8b])))
println(Cmd.new("gzip", ["-d", "-c"]).read(input: compressed))
```

```text
true
Ok{ value: "hello, ant" }
```

Decode the output strictly, and refuse it when it is not valid UTF-8:

```ant
func readUtf8(c: Cmd): Result[string, Error] {
  let bytes = try c.readBytes()
  let text = String.fromBytes(bytes)
  if text == null { return Result.fail("${c.string()}: output is not valid UTF-8") }
  Result.ok(text)
}

println(readUtf8(Cmd.new("printf", ["h\\303\\251"])))
println(readUtf8(Cmd.new("printf", ["h\\351"])))
```

```text
Ok{ value: "hé" }
Err{ error: Error{ message: "printf 'h\\351': output is not valid UTF-8" } }
```

## Properties and relations

`Cmd.readBytes` gives what [`Cmd.capture`](capture.md) keeps in `stdoutBytes` when the status is 0:

```ant
let lines = Cmd.new("printf", ["a\\nb\\n"])
let output = try lines.capture()
assert(output.status == 0)
assert(output.stdoutBytes == (try lines.readBytes()))
```

[`Cmd.read`](read.md) is the same output decoded with
[`String.fromBytesLossy`](../String/fromBytesLossy.md), less one final newline:

```ant
let hello = Cmd.new("printf", ["hello\\n"])
let text = String.fromBytesLossy(try hello.readBytes())
assert(text == (try hello.read()) + "\n")
```

## Possible issues

A command that reports its answer through a nonzero status is an `Err`. `grep` exits 1 when nothing
matches; use [`Cmd.capture`](capture.md) to read the status instead:

```ant
let grep = Cmd.new("grep", ["x"])
println(grep.readBytes(input: "abc\n"))
println(grep.capture(input: "abc\n") | Result.map({ x.status }))
```

```text
Err{ error: Error{ message: "grep: exit status 1" } }
Ok{ value: 1 }
```

There is no shell: arguments arrive as written, so `*` and `|` are not expanded. Use
[`Cmd.pipe`](pipe.md) for a pipeline:

```ant
println(Cmd.new("echo", ["*"]).readBytes())
println(Cmd.new("echo", ["a", "|", "tr", "a", "b"]).readBytes())
println(Cmd.pipe([Cmd.new("echo", ["a"]), Cmd.new("tr", ["a", "b"])]).readBytes())
```

```text
Ok{ value: Bytes(2a 0a) }
Ok{ value: Bytes(61 20 7c 20 74 72 20 61 20 62 0a) }
Ok{ value: Bytes(62 0a) }
```

## See also

[Cmd.read](read.md) · [Cmd.capture](capture.md) · [Cmd.run](run.md) · [Cmd.start](start.md) ·
[Cmd.pipe](pipe.md) · [File.readBytes](../File/readBytes.md) ·
[String.fromBytes](../String/fromBytes.md)
