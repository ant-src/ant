<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.isAbs

```ant signature
func Path.isAbs(p: string): bool
```

Path.isAbs reports whether p is an absolute path.

<!-- end generated -->

## Details

- It looks only at the text: it does not touch the file system, so a path that does not exist can be
  absolute.
- It follows the host's convention. The examples on this page run on Unix, where an absolute path
  begins with `/`.

## Basic examples

```ant
Path.isAbs("/usr/local/bin") // true
Path.isAbs("notes.txt")      // false
```

## Scope

Relative forms, with or without a leading `.` or `..`, are all relative:

```ant
Path.isAbs("src/main.ant") // false
Path.isAbs("./main.ant")   // false
Path.isAbs("../lib")       // false
Path.isAbs("")             // false
```

The root and paths that are not clean are absolute when they begin at the root:

```ant
Path.isAbs("/")          // true
Path.isAbs("//srv/data") // true
Path.isAbs("/../etc")    // true
```

A Windows path with a drive letter is relative on Unix, where `C:` is an ordinary name:

```ant
Path.isAbs("C:\\Users") // false
```

As a function value, it filters an array of paths:

```ant
["/etc/hosts", "notes.txt", "/var/log"].filter(Path.isAbs) // ["/etc/hosts", "/var/log"]
```

## Named arguments

The parameter is named `p`:

```ant
Path.isAbs(p: "/home") // true
```

## Applications

Resolve a path from a configuration file against the directory the file is in, and keep an absolute
one as it is:

```ant
func resolve(dir: string, p: string): string {
  if Path.isAbs(p) { p } else { Path.join([dir, p]) }
}

resolve("/etc/app", "cache")      // "/etc/app/cache"
resolve("/etc/app", "/var/cache") // "/var/cache"
resolve("/etc/app", "../shared")  // "/etc/shared"
```

Reject a setting that must be absolute before using it:

```ant
func dataDir(setting: string): Result[string, Error] {
  if !Path.isAbs(setting) {
    return Result.fail("data dir must be absolute, got ${setting}")
  }
  Result.ok(Path.clean(setting))
}

println(dataDir("/srv/data/"))
println(dataDir("data"))
```

```text
Ok{ value: "/srv/data" }
Err{ error: Error{ message: "data dir must be absolute, got data" } }
```

## Properties and relations

[`Path.abs`](abs.md) always gives an absolute path, [`Path.clean`](clean.md) keeps a path absolute
or relative, and [`Path.join`](join.md) is absolute when its first part is:

```ant
let paths = ["/a/b", "a/b", "./a", "../a", "/", "", "/../a"]
for p in paths {
  assert(Path.isAbs(try Path.abs(p)))
  assert(Path.isAbs(Path.clean(p)) == Path.isAbs(p))
  assert(Path.isAbs(Path.join(["/root", p])))
}
```

## Possible issues

`~` is not expanded: a path that starts with it is relative.

```ant
Path.isAbs("~/notes.txt") // false
```

[`Path.rel`](rel.md) fails when one path is absolute and the other is not. Check both first, or make
both absolute with [`Path.abs`](abs.md):

```ant
Path.rel("/srv", "data").isOk() // false
```

## See also

[Path.abs](abs.md) · [Path.clean](clean.md) · [Path.join](join.md) · [Path.rel](rel.md) ·
[Path](index.md)
