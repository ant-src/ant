<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.hostname

```ant signature
func Os.hostname(): Result[string, Error]
```

Os.hostname returns the name the host's kernel reports, as hostname prints it, Go's os.Hostname. It
is not looked up in DNS, so it may or may not hold a domain. It is Err only when the host cannot
report one.

<!-- end generated -->

## Basic examples

Print the result as it comes back:

<!-- test: check-only the host name differs between machines -->

```ant
println(Os.hostname())
```

In a script, `try` unwraps the name, or ends the script with the error when the name cannot be read:

<!-- test: check-only the host name differs between machines -->

```ant
let host = try Os.hostname()
println("running on ${host}")
```

## Scope

### A fallback name

`Result.unwrapOr` gives a name to use when the host's name cannot be read, so the program goes on:

<!-- test: check-only the host name differs between machines -->

```ant
let host = Os.hostname().unwrapOr("unknown")
println("[${host}] started")
```

### Inside a function

In a function that returns a `Result`, `try` passes the error on to the caller. Wrap it first to say
which step failed:

<!-- test: check-only the host name and the process id differ between runs -->

```ant
func instanceId(): Result[string, Error] {
  let host = try (Os.hostname() | Result.wrapErr("read the host name"))
  Result.ok("${host}:${Os.pid()}")
}

println(instanceId())
```

## Applications

### Short host name

A host name may be fully qualified. Keep the part before the first dot:

```ant
func shortName(host: string): string { host.cut(".")?.before ?? host }

shortName("web-1.example.com") // "web-1"
shortName("laptop")            // "laptop"
```

Then apply it to this host:

<!-- test: before
func shortName(host: string): string { host.cut(".")?.before ?? host }
-->
<!-- test: check-only the host name differs between machines -->

```ant
let host = try Os.hostname()
println(shortName(host))
```

### A file name per host

Machines that share a directory each write their own file when the name includes the host:

<!-- test: check-only the host name differs between machines -->

```ant
let host = try Os.hostname()
let path = "status-${host}.txt"
try File.write(path, "ok\n")
println(try File.read(path))
```

### Choose a role by host

An `if` expression picks a setting from the host's name:

<!-- test: check-only the host name differs between machines -->

```ant
let host = Os.hostname().unwrapOr("")
let role = if host.startsWith("build-") { "builder" } else { "worker" }
println("${host}: ${role}")
```

## Possible issues

### The result is not a string

`Os.hostname` returns a `Result`, which a `string` does not accept:

```ant
let host: string = Os.hostname() // error: cannot use Result[string, Error] where string is expected
```

### `try` binds before a pipe

`try` applies to the call right after it, so `try Os.hostname() | Result.wrapErr(…)` pipes the
unwrapped string into `Result.wrapErr`. Parenthesize the pipeline, as `instanceId` does above:

```ant
func label(): Result[string, Error] {
  let host = try Os.hostname() | Result.wrapErr("read the host name")
  // error: cannot use string where Result[T, Error] is expected
  Result.ok(host)
}
```

## See also

[Os.platform](platform.md), [Os.arch](arch.md), [Os.pid](pid.md), [Os.uid](uid.md),
[Os.env](env.md), [Result.unwrapOr](../Result/unwrapOr.md), [Result.wrapErr](../Result/wrapErr.md)
