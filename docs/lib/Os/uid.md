<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.uid

```ant signature
func Os.uid(): int | null
```

Os.uid returns the effective user id, 0 for root, or null on Windows and WebAssembly, which have
none.

<!-- end generated -->

## Details

- The id is the effective one, the id that decides what the process may touch.
- It is null on Windows and on other hosts without user ids, such as WebAssembly. [Os.gid](gid.md)
  is null on the same hosts.

## Basic examples

Print the id of the user the program runs as:

<!-- test: check-only the id differs between users -->

```ant
println(Os.uid())
```

Ask whether the program runs as root:

<!-- test: check-only the answer depends on the user -->

```ant
let root = Os.uid() == 0
println(if root { "running as root" } else { "running as a regular user" })
```

## Scope

### Handling `null`

Comparing with `==` needs no `null` check, since `null == 0` is false. Arithmetic does: `??` gives a
value to use where there is no id:

<!-- test: check-only the id differs between users -->

```ant
let uid = Os.uid() ?? -1
println("uid ${uid}")
```

A guard narrows the result to `int` for the rest of the block:

<!-- test: check-only the id differs between users -->

```ant
let uid = Os.uid()
if uid == null {
  println("no user ids on this platform")
  return
}
println("uid ${uid}, next ${uid + 1}")
```

### One label per case

An `if` expression gives each case its own label:

<!-- test: check-only the id differs between users -->

```ant
let uid = Os.uid()
let who = if uid == null { "no user id" } else if uid == 0 { "root" } else { "user ${uid}" }
println(who)
```

### As a function value

```ant
let current: func(): int | null = Os.uid
assert(current() == Os.uid())
```

## Applications

### Refuse to run as root

A tool that writes into the user's files can stop before it creates files root owns:

<!-- test: check-only the outcome depends on the user -->

```ant
if Os.uid() == 0 {
  eprintln("do not run this tool as root")
  Os.exit(1)
}
println("ok")
```

### Require root

An installer that writes system directories checks for the opposite:

<!-- test: check-only the outcome depends on the user -->

```ant
if Os.uid() != 0 {
  eprintln("run this installer with sudo")
  Os.exit(1)
}
println("installing")
```

### Tag log lines with the user

Two users running the same program write lines that can be told apart:

<!-- test: check-only the id differs between users -->

```ant
func log(message: string) {
  let uid = Os.uid()
  let user = if uid == null { "-" } else { "${uid}" }
  eprintln("[uid ${user}] ${message}")
}

log("starting")
```

### Check who owns a file

[File.stat](../File/stat.md) reports a file's owner as `uid`, null on the same hosts, so guard
first:

<!-- test: check-only the answer depends on the user and the file -->

```ant
let uid = Os.uid()
if uid == null {
  println("no user ids on this platform")
  return
}
let info = try File.stat("/etc/hosts")
println(if info.uid == uid { "mine" } else { "owned by ${info.uid}" })
```

## Properties and relations

On Unix the id is what `id -u` prints for the same process user:

```ant
let printed = try Cmd.read(Cmd.new("id", ["-u"]))
assert(printed.trim() == "${Os.uid()}")
```

## Possible issues

### The result may be `null`

The result is `int | null`, so it does not go where an `int` is expected:

```ant
let next: int = Os.uid() + 1 // error: + needs both operands of one type, got int | null and int
```

### `!= 0` is true where there are no user ids

`Os.uid() != 0` holds on Windows and WebAssembly too, where there is no root to compare with, so the
installer above stops there. Check for `null` first when the program runs on such hosts:

<!-- test: check-only the outcome depends on the platform and the user -->

```ant
let uid = Os.uid()
let privileged = uid == null || uid == 0
println(privileged)
```

## See also

[Os.gid](gid.md) · [File.stat](../File/stat.md) · [File.chown](../File/chown.md) · [Os.pid](pid.md)
· [Os.home](home.md) · [Os.platform](platform.md) · [Os](index.md)
