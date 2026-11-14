<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.OutputBytes

```ant signature
type Cmd.OutputBytes = { status: int, statuses: Array[int], stdout: Bytes, stderr: string }
```

Cmd.OutputBytes is what [`Cmd.captureBytes`](../captureBytes.md) collects: as
[`Cmd.Output`](Output.md), with standard output as the last stage wrote it.

<!-- end generated -->

## Details

- `stdout` holds every byte the command wrote, final newline included, with nothing decoded or
  replaced. `stderr` is text, as in [Cmd.Output](Output.md): bytes there that are not valid UTF-8
  become U+FFFD. To keep standard error as bytes, run the command with [Cmd.start](../start.md) and
  a handle for `stderr`.
- A nonzero status is not an error: [Cmd.captureBytes](../captureBytes.md) returns `Ok` with it in
  `status`. It is `Err` only when the command could not be started or run to its end.
- `statuses` has one entry per stage, in pipeline order, and `status` is that of the rightmost stage
  that failed, or 0, as in [Cmd.Output](Output.md).

## Basic examples

Capture a command's status, its standard output as bytes, and its standard error as text with
[Cmd.captureBytes](../captureBytes.md):

```ant
let out = try Cmd.captureBytes(Cmd.new("sh", ["-c", "printf '\\001\\377'; echo oops >&2; exit 3"]))
println(out.status)
println(out.stdout)
println(out.stderr.trim())
```

```text
3
Bytes(01 ff)
oops
```

A `Cmd.OutputBytes` prints as a record:

```ant
println(try Cmd.captureBytes(Cmd.new("echo", ["hi"])))
```

```text
Cmd.OutputBytes{ status: 0, statuses: [0], stdout: Bytes(68 69 0a), stderr: "" }
```

## Scope

A command that writes nothing has an empty `stdout` and `stderr`:

```ant
let out = try Cmd.captureBytes(Cmd.new("true"))
println(out.stdout.length())
println(out.stderr.isEmpty())
```

```text
0
true
```

Standard error stays text, so a byte there that is not UTF-8 is U+FFFD:

```ant
let out = try Cmd.captureBytes(Cmd.new("sh", ["-c", "printf '\\377' >&2"]))
println(out.stderr == "\u{FFFD}")
```

```text
true
```

A pipeline has one entry in `statuses` per stage, and `stdout` is what the last stage wrote:

```ant
let out = try Cmd.captureBytes(Cmd.pipe([
  Cmd.new("printf", ["\\001\\002"]),
  Cmd.new("sh", ["-c", "cat; exit 4"]),
]))
println(out.status)
println(out.statuses)
println(out.stdout)
```

```text
4
[0, 4]
Bytes(01 02)
```

`Cmd.OutputBytes` is a record type, so a literal builds one, for example to test a function that
takes an output without running a command:

```ant
func describe(out: Cmd.OutputBytes): string {
  if out.status != 0 {
    return "failed (${out.status}): ${out.stderr.trim()}"
  }
  "${out.stdout.length()} bytes"
}

let ok = Cmd.OutputBytes{
  status: 0,
  statuses: [0],
  stdout: Bytes.fromArray([0x1f, 0x8b, 0x08]),
  stderr: "",
}
println(describe(ok))
println(describe(Cmd.OutputBytes{ status: 1, statuses: [1], stdout: Bytes.new(), stderr: "bad\n" }))
```

```text
3 bytes
failed (1): bad
```

## Applications

Compress text with `gzip`, check the format's magic number, and feed the bytes back to decompress
them:

```ant
let packed = try Cmd.captureBytes(Cmd.new("gzip", ["-c"]), input: "hello, ant\n")
assert(packed.status == 0)
println(packed.stdout.startsWith(Bytes.fromArray([0x1f, 0x8b])))

let unpacked = try Cmd.captureBytes(Cmd.new("gzip", ["-dc"]), input: packed.stdout)
print(String.fromBytes(unpacked.stdout) ?? "not text")
```

```text
true
hello, ant
```

Fingerprint what a command printed, whatever it is, and report a failure with its standard error:

```ant
func fingerprint(c: Cmd): Result[string, Error] {
  let out = try Cmd.captureBytes(c)
  if out.status != 0 {
    return Result.fail("${c.program}: ${out.stderr.trim()}")
  }
  Result.ok(out.stdout | Crypto.sha256() | Encoding.toHex() | String.slice(0, 12))
}

println(fingerprint(Cmd.new("printf", ["\\000\\001\\002"])))
println(fingerprint(Cmd.new("sh", ["-c", "echo 'no data' >&2; exit 1"])))
```

```text
Ok{ value: "ae4b3280e56e" }
Err{ error: Error{ message: "sh: no data" } }
```

Treat the output as text only when it is valid UTF-8:

```ant
for script in ["echo café", "printf '\\377\\376'"] {
  let out = try Cmd.captureBytes(Cmd.new("sh", ["-c", script]))
  let text = String.fromBytes(out.stdout)
  println(if text != null { "text: ${text.trim()}" } else { "binary: ${out.stdout}" })
}
```

```text
text: café
binary: Bytes(ff fe)
```

## Properties and relations

[Cmd.capture](../capture.md) gives `stdout` decoded with U+FFFD for bad bytes, and the same
`status`, `statuses`, and `stderr`:

```ant
for script in ["echo plain", "printf 'caf\\303\\251'", "printf '\\377'; exit 2", "echo e >&2"] {
  let c = Cmd.new("sh", ["-c", script])
  let raw = try Cmd.captureBytes(c)
  let text = try Cmd.capture(c)
  assert(String.fromBytesLossy(raw.stdout) == text.stdout)
  assert(raw.stderr == text.stderr)
  assert(raw.status == text.status && raw.statuses == text.statuses)
}
```

On status 0, [Cmd.readBytes](../readBytes.md) returns `stdout` as it is, final newline included:

```ant
let c = Cmd.new("printf", ["a\\n\\000\\n"])
let out = try Cmd.captureBytes(c)
assert(out.status == 0)
assert(try Cmd.readBytes(c) == out.stdout)
assert(out.stdout.length() == 4)
```

## Possible issues

`stdout` is `Bytes`, not a string, so comparing it with a string does not compile. Decode it with
[String.fromBytes](../../String/fromBytes.md), or compare it with bytes:

```ant
let out = try Cmd.captureBytes(Cmd.new("echo", ["yes"]))
println(out.stdout == "yes\n") // error: Bytes and string cannot be compared with ==: neither is assignable to the other
```

```ant
let out = try Cmd.captureBytes(Cmd.new("echo", ["yes"]))
println(String.fromBytes(out.stdout) == "yes\n")
println(out.stdout == "yes\n".toBytes())
```

```text
true
true
```

`try` does not stop on a failing command, since a nonzero status is still `Ok`. Check `status`
before using `stdout`:

```ant
let out = try Cmd.captureBytes(Cmd.new("sh", ["-c", "printf partial; exit 1"]))
println("status ${out.status}, ${out.stdout.length()} bytes")
```

```text
status 1, 7 bytes
```

## See also

[Cmd.captureBytes](../captureBytes.md) · [Cmd.Output](Output.md) · [Cmd.capture](../capture.md) ·
[Cmd.readBytes](../readBytes.md) · [Cmd.start](../start.md) · [Bytes](../../Bytes/index.md) ·
[String.fromBytes](../../String/fromBytes.md)
