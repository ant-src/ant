<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.match

```ant signature
func Path.match(pattern: string, path: string, hidden: bool = false): Result[bool, Error]
```

Path.match reports whether the whole of path matches the glob pattern, element by element. \* is any
run of characters within an element, ? one character, \[a-z\] one of a set, \[^a-z\] or \[!a-z\] one
outside it, and \*\* as a whole element zero or more elements (one or more as the last). On Unix \\
escapes; on Windows it separates, and \[\*\] escapes. A wildcard does not match a dot that begins an
element unless hidden is true. It looks only at the text and is Err on a malformed pattern.

<!-- end generated -->

## Details

- `Path.match` looks only at the text and reads no file.
- A malformed pattern is `Err`, whatever the path, so a check against `""` validates a pattern. A
  pattern that is well formed is always `Ok`, with `true` or `false`.
- On Windows both `/` and `\` separate elements, so a pattern written with `/` works on every
  system.

## Basic examples

Match a path against a glob pattern:

```ant
Path.match("*.txt", "notes.txt")        // Ok{ value: true }
Path.match("src/*.ant", "src/main.ant") // Ok{ value: true }
Path.match("src/*.ant", "lib/main.ant") // Ok{ value: false }
```

`**` as a whole element matches any number of elements, zero included:

```ant
Path.match("**/*.ant", "src/main.ant") // Ok{ value: true }
Path.match("**/*.ant", "main.ant")     // Ok{ value: true }
```

A wildcard skips a name that begins with a dot:

```ant
Path.match("*", ".env") // Ok{ value: false }
```

## Scope

`*` matches any run of characters within one element, the empty run included, but not a separator:

```ant
Path.match("*.ant", ".ant")      // Ok{ value: false }
Path.match("a*", "a")            // Ok{ value: true }
Path.match("*.ant", "src/a.ant") // Ok{ value: false }
```

`?` matches exactly one character, which is one code point, so `é` is one:

```ant
Path.match("?.md", "é.md")  // Ok{ value: true }
Path.match("??.md", "é.md") // Ok{ value: false }
```

A set in brackets matches one character of the set, or of a range:

```ant
Path.match("log[0-9].txt", "log7.txt") // Ok{ value: true }
Path.match("[abc].txt", "d.txt")       // Ok{ value: false }
```

`[^...]` and `[!...]` match one character outside the set:

```ant
Path.match("[^a-c]og", "dog") // Ok{ value: true }
Path.match("[!a-c]og", "bog") // Ok{ value: false }
```

A backslash escapes the next character on Unix, `-` and `]` in a set included:

```ant
Path.match("what\\?.txt", "what?.txt") // Ok{ value: true }
Path.match("what\\?.txt", "whatX.txt") // Ok{ value: false }
Path.match("[\\-\\]]", "]")            // Ok{ value: true }
```

`**` as the last element matches one or more elements, so `src/**` matches everything below `src`,
but not `src` itself:

```ant
Path.match("src/**", "src/a/b/c.ant") // Ok{ value: true }
Path.match("src/**", "src")           // Ok{ value: false }
```

`**` between two elements matches zero or more elements:

```ant
Path.match("a/**/z", "a/z")     // Ok{ value: true }
Path.match("a/**/z", "a/b/c/z") // Ok{ value: true }
```

`**` inside an element, not alone in it, is a `*`:

```ant
Path.match("a**z", "abcz")  // Ok{ value: true }
Path.match("a**z", "ab/cz") // Ok{ value: false }
```

A pattern element that begins with a literal `.` matches names that begin with a dot:

```ant
Path.match(".*", ".env")        // Ok{ value: true }
Path.match("**/.git", "a/.git") // Ok{ value: true }
```

`.` and `..` are matched only by themselves:

```ant
Path.match("*", "..")      // Ok{ value: false }
Path.match("../*", "../a") // Ok{ value: true }
```

A pattern that ends in a separator matches a path that does:

```ant
Path.match("*/", "src/") // Ok{ value: true }
Path.match("*/", "src")  // Ok{ value: false }
```

A malformed pattern is `Err`, with a message that says what is wrong:

```ant
Path.match("[a-z", "b")  // Err{ error: Error{ message: "malformed pattern [a-z: [ without ]" } }
Path.match("[z-a]", "b") // Err{ error: Error{ message: "malformed pattern [z-a]: range z-a ends before it starts" } }
```

## Named arguments

### hidden

With `hidden: true`, wildcards match names that begin with a dot too:

```ant
Path.match("*", ".env")                            // Ok{ value: false }
Path.match("*", ".env", hidden: true)              // Ok{ value: true }
Path.match("**/*.ant", ".git/x.ant", hidden: true) // Ok{ value: true }
```

`.` and `..` still match only themselves:

```ant
Path.match("*", "..", hidden: true) // Ok{ value: false }
```

## Applications

Keep the paths that match a pattern, checking the pattern once:

```ant
func matching(pattern: string, paths: Array[string]): Result[Array[string], Error] {
  try Path.match(pattern, "")
  Result.ok(paths.filter({ Path.match(pattern, x) == Ok{ value: true } }))
}

let paths = ["src/main.ant", "src/main_test.ant", "tests/io_test.ant", "README.md"]
println(try matching("**/*_test.ant", paths))
```

```text
["src/main_test.ant", "tests/io_test.ant"]
```

Ignore a path when any pattern of an ignore list matches it:

```ant
func ignored(path: string, patterns: Array[string]): Result[bool, Error] {
  for pattern in patterns {
    if try Path.match(pattern, path) {
      return Result.ok(true)
    }
  }
  Result.ok(false)
}

let patterns = ["build/**", "**/*.tmp", ".*"]
for path in ["build/out.js", "src/a.ant", "src/b.tmp", ".env"] {
  let skip = try ignored(path, patterns)
  println("${path}: ${if skip { "ignored" } else { "kept" }}")
}
```

```text
build/out.js: ignored
src/a.ant: kept
src/b.tmp: ignored
.env: ignored
```

Check a pattern a user gives before using it:

```ant
func checkPattern(pattern: string): Result[string, Error] {
  try (Path.match(pattern, "") | Result.wrapErr("--include"))
  Result.ok(pattern)
}

println(checkPattern("*.ant"))
println(checkPattern("[*.ant"))
```

```text
Ok{ value: "*.ant" }
Err{ error: Error{ message: "--include: malformed pattern [*.ant: [ without ]" } }
```

Match a name alone, wherever the file is, with `Path.base`. The pattern is a literal that is well
formed, so `Result.unwrap` cannot panic here:

```ant
let paths = ["docs/guide.md", "README.md", "src/main.ant"]
paths.filter({ Result.unwrap(Path.match("*.md", Path.base(x))) }) // ["docs/guide.md", "README.md"]
```

Pick the files of a directory tree from `File.walk`:

<!-- test: file app/src/main.ant
println("main")
-->
<!-- test: file app/src/main_test.ant
test "main" {}
-->
<!-- test: file app/README.md
# App
-->

```ant
for path in try File.walk("app") {
  if try Path.match("src/*.ant", path) {
    println(path)
  }
}
```

```text
src/main.ant
src/main_test.ant
```

## Properties and relations

On Unix, a pattern without wildcards or escapes matches only itself:

```ant
let path = "docs/lib/Path/match.md"
assert(try Path.match(path, path))
assert(!(try Path.match(path, "docs/lib/Path/join.md")))
```

For a pattern that does not end in a separator, on a tree without links, `File.glob` returns the
paths from `File.walk` that `Path.match` accepts, in the same order:

<!-- test: file tree/a.txt
a
-->
<!-- test: file tree/sub/b.txt
b
-->
<!-- test: file tree/sub/c.md
c
-->

```ant
let pattern = "**/*.txt"
let globbed = try File.glob(pattern, dir: "tree")
let walked = (try File.walk("tree")).filter({ Path.match(pattern, x) == Ok{ value: true } })
assert(globbed == walked)
println(globbed)
```

```text
["a.txt", "sub/b.txt"]
```

## Possible issues

`Path.match` does not clean `path`, so a leading `./` or a doubled separator stops a match. Clean
the path first:

```ant
Path.match("*.txt", "./a.txt")             // Ok{ value: false }
Path.match("*.txt", Path.clean("./a.txt")) // Ok{ value: true }
```

`*` stays within one element. Use `**/` to match at any depth:

```ant
Path.match("*.ant", "src/main.ant")    // Ok{ value: false }
Path.match("**/*.ant", "src/main.ant") // Ok{ value: true }
```

`**` does not take a name that begins with a dot, so it does not descend into `.git`:

```ant
Path.match("**/*.ant", ".git/hooks/x.ant") // Ok{ value: false }
```

Matching is case-sensitive on every system, even where the file system ignores case:

```ant
Path.match("*.TXT", "a.txt")          // Ok{ value: false }
Path.match("*.[tT][xX][tT]", "a.TXT") // Ok{ value: true }
```

The result is a `Result`, so comparing it with `true` does not compile:

```ant
let hit = Path.match("*.md", "README.md") == true
// error: Result[bool, Error] and bool cannot be compared with ==: neither is assignable to the other
```

Unwrap it with `try`, or compare it with `Ok{ value: true }`, which counts a malformed pattern as no
match:

```ant
Path.match("*.md", "README.md") == Ok{ value: true }  // true
Path.match("[*.md", "README.md") == Ok{ value: true } // false
```

## See also

[File.glob](../File/glob.md), [File.walk](../File/walk.md), [Path.base](base.md),
[Path.clean](clean.md), [Path.join](join.md), [String.endsWith](../String/endsWith.md)
