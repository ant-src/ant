<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.captureBytes

```ant signature
func Cmd.captureBytes(c: Cmd | Cmd.Pipeline, input: string | Bytes = "", ctx: Context | null = null): Result[Cmd.OutputBytes, Error]
```

Cmd.captureBytes is [`Cmd.capture`](capture.md) that keeps standard output as the command wrote it,
for a command whose output is binary; standard error stays text.

<!-- end generated -->

## Details

- `stdout` holds the last stage's standard output as `Bytes`, exactly as written: a final newline
  stays, and bytes that are not UTF-8 stay as they are.
- `stderr` stays text, as in [Cmd.capture](capture.md): bytes that are not UTF-8 become U+FFFD. To
  keep standard error as bytes, start the command with [Cmd.start](start.md) and a handle.
- A nonzero status is `Ok`. `Err` means a stage could not be started or run to its end, or `ctx`
  ended first: before the start nothing runs, and during the run every stage is killed.
- `input` is what the first stage reads: a string as its UTF-8, `Bytes` as they are. By default the
  command reads nothing.
- Both outputs are held in memory until the command exits.

## Basic examples

Run a command whose output is binary, and read its status, its bytes, and its standard error. The
exit status 3 is `Ok`:

```ant
let script = "printf '\\377\\330\\377'; echo 'warning: truncated' >&2; exit 3"
let out = try Cmd.captureBytes(Cmd.new("sh", ["-c", script]))
println(out.status)
println(out.stdout)
print(out.stderr)
```

```text
3
Bytes(ff d8 ff)
warning: truncated
```

Give the command its standard input. `gzip` output starts with the bytes `1f 8b`:

```ant
let out = try Cmd.captureBytes(Cmd.new("gzip", ["-c", "-n"]), "hello, ant\n")
println(out.stdout.slice(0, 2))
```

```text
Bytes(1f 8b)
```

## Scope

The whole `Cmd.OutputBytes` record. A byte `ff` stays in `stdout` and becomes U+FFFD in `stderr`:

```ant
let script = "printf 'a\\377'; printf 'b\\377\\n' >&2"
println(try Cmd.captureBytes(Cmd.new("sh", ["-c", script])))
```

```text
Cmd.OutputBytes{ status: 0, statuses: [0], stdout: Bytes(61 ff), stderr: "b�\n" }
```

A command that writes nothing gives empty `Bytes`:

```ant
let out = try Cmd.captureBytes(Cmd.new("true"))
println(out.stdout)
println(out.stdout.length())
```

```text
Bytes()
0
```

`input` as `Bytes` reaches the command unchanged:

```ant
let data = Bytes.fromArray([0, 255, 10])
let out = try Cmd.captureBytes(Cmd.new("cat"), data)
println(out.stdout)
```

```text
Bytes(00 ff 0a)
```

A pipeline from [Cmd.pipe](pipe.md) is captured as one command: `stdout` is what the last stage
wrote, and `statuses` holds each stage's status:

```ant
let lastTwo = Cmd.pipe([
  Cmd.new("printf", ["\\001\\002\\003\\004"]),
  Cmd.new("tail", ["-c", "2"]),
])
println(try Cmd.captureBytes(lastTwo))
```

```text
Cmd.OutputBytes{ status: 0, statuses: [0, 0], stdout: Bytes(03 04), stderr: "" }
```

A program that is not found is `Err`:

```ant
println(Cmd.captureBytes(Cmd.new("no-such-program")))
```

```text
Err{ error: Error{ message: "exec: \"no-such-program\": executable file not found in $PATH" } }
```

Call it with a dot, or in a pipe with the command first:

```ant
let hello = Cmd.new("printf", ["hello"])
println((try hello.captureBytes()).stdout.length())
println(hello | Cmd.captureBytes() | Result.map({ x.stdout }))
```

```text
5
Ok{ value: Bytes(68 65 6c 6c 6f) }
```

## Named arguments

### input

`input` defaults to `""`, so a command that reads standard input sees it end at once:

```ant
println(Cmd.new("cat").captureBytes() | Result.map({ x.stdout }))
println(Cmd.new("cat").captureBytes(input: "abc") | Result.map({ x.stdout }))
```

```text
Ok{ value: Bytes() }
Ok{ value: Bytes(61 62 63) }
```

### ctx

`ctx:` bounds the run. When the context is done first, every stage is killed and the result is `Err`
with the context's reason:

```ant
let ctx = Context.new(timeout: 100ms)
println(Cmd.new("sleep", ["5"]).captureBytes(ctx: ctx))
```

```text
Err{ error: Error{ message: "timeout" } }
```

## Applications

Decompress data with `gzip`, and turn a nonzero status into an `Err` that says which status it was:

```ant
func gunzip(data: Bytes): Result[Bytes, Error] {
  let out = try Cmd.new("gzip", ["-d", "-c"]).captureBytes(input: data)
  if out.status != 0 {
    return Result.fail("gunzip: status ${out.status}")
  }
  Result.ok(out.stdout)
}

let packed = try Cmd.new("gzip", ["-c", "-n"]).readBytes(input: "hello, ant\n")
print(String.fromBytesLossy(try gunzip(packed)))
println(gunzip(Bytes.fromArray([1, 2, 3])))
```

```text
hello, ant
Err{ error: Error{ message: "gunzip: status 1" } }
```

Keep a tool's binary output and pass its warnings on. The digest covers the output only, never the
warnings:

```ant
let script = "printf 'payload'; echo 'warning: old format' >&2"
let out = try Cmd.captureBytes(Cmd.new("sh", ["-c", script]))
for line in out.stderr.lines() {
  println("tool: ${line}")
}
println(out.stdout | Crypto.sha256() | Encoding.toHex())
```

```text
tool: warning: old format
239f59ed55e737c77147cf55ad0c1b030b6d7ee748a7426952f9b852d5a935e5
```

## Properties and relations

On status 0, `stdout` is what [Cmd.readBytes](readBytes.md) returns:

```ant
let c = Cmd.new("printf", ["a\\nb\\n"])
let out = try Cmd.captureBytes(c)
assert(out.status == 0)
assert(out.stdout == (try Cmd.readBytes(c)))
```

[Cmd.capture](capture.md) gives the same status and standard error, and `stdout` decoded with
[String.fromBytesLossy](../String/fromBytesLossy.md):

```ant
let c = Cmd.new("sh", ["-c", "printf 'caf\\351'; echo note >&2; exit 4"])
let raw = try Cmd.captureBytes(c)
let text = try Cmd.capture(c)
assert(raw.status == text.status)
assert(raw.stderr == text.stderr)
assert(String.fromBytesLossy(raw.stdout) == text.stdout)
```

## Possible issues

`stdout` is `Bytes`, so comparing it with a string is a compile error. Compare it with the string's
bytes, or decode it with [String.fromBytes](../String/fromBytes.md):

```ant
let out = try Cmd.captureBytes(Cmd.new("echo", ["hi"]))
println(out.stdout == "hi\n")
// error: Bytes and string cannot be compared with ==: neither is assignable to the other
```

```ant
let out = try Cmd.captureBytes(Cmd.new("echo", ["hi"]))
println(out.stdout == "hi\n".toBytes())
println(String.fromBytes(out.stdout) == "hi\n")
```

```text
true
true
```

`try` passes a failed command through: a nonzero status is `Ok`. Check `status`, or use
[Cmd.readBytes](readBytes.md) when a nonzero status should be `Err`:

```ant
let out = try Cmd.captureBytes(Cmd.new("sh", ["-c", "printf partial; exit 1"]))
println("${out.status} ${out.stdout}")
println(Cmd.new("sh", ["-c", "printf partial; exit 1"]).readBytes())
```

```text
1 Bytes(70 61 72 74 69 61 6c)
Err{ error: Error{ message: "sh: exit status 1" } }
```

## See also

[Cmd.capture](capture.md) · [Cmd.readBytes](readBytes.md) · [Cmd.read](read.md) ·
[Cmd.start](start.md) · [Cmd.pipe](pipe.md) · [Cmd.OutputBytes](types/OutputBytes.md) ·
[String.fromBytes](../String/fromBytes.md) · [String.fromBytesLossy](../String/fromBytesLossy.md)
