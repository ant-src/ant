<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.string

```ant signature
func Cmd.string(c: Cmd | Cmd.Pipeline): string
```

Cmd.string returns c as a line for a log: the program and its arguments, each quoted as a POSIX
shell needs it, and a pipeline's stages joined by " \| ". A word is left bare when it is not empty
and is only ASCII letters, digits and \_@%+=:,./-, = excepted in the program's word; any other word
is put in single quotes, with each ' in it written '\\'', so '' is the empty word and a line break
stays inside the quotes. env and dir are left out, so a secret in env never reaches the log. It
panics on a pipeline of no commands.

<!-- end generated -->

## Basic examples

Write a command as a line for a log:

```ant
Cmd.string(Cmd.new("go", ["build", "./..."])) // "go build ./..."
```

An argument with a space is quoted, so the line shows where each argument ends:

```ant
Cmd.string(Cmd.new("git", ["commit", "-m", "fix the parser"])) // "git commit -m 'fix the parser'"
```

A pipeline's stages are joined by `|`:

```ant
let lines = Cmd.pipe([Cmd.new("git", ["log", "--oneline"]), Cmd.new("wc", ["-l"])])
Cmd.string(lines) // "git log --oneline | wc -l"
```

## Scope

A command with no arguments is its program alone:

```ant
Cmd.new("ls").string() // "ls"
```

An empty argument is quoted, so it does not vanish from the line:

```ant
Cmd.new("grep", ["", "notes.txt"]).string() // "grep '' notes.txt"
```

Characters a shell would expand or split on are quoted, so the line shows the argument the program
gets, not what a shell would make of it:

```ant
Cmd.new("echo", ["$HOME", "*.go", "a;b"]).string() // "echo '$HOME' '*.go' 'a;b'"
```

An argument with a single quote is quoted so a shell reads the quote back as part of the argument:

```ant
Cmd.new("echo", ["it's"]).string() // "echo 'it'\\''s'"
```

A non-ASCII argument is quoted:

```ant
Cmd.new("echo", ["héllo", "日本"]).string() // "echo 'héllo' '日本'"
```

A program whose path has a space is quoted like an argument:

```ant
Cmd.new("/opt/my tools/lint", ["main.go"]).string() // "'/opt/my tools/lint' main.go"
```

An argument a shell takes as it is, such as an option or a path, stays bare:

```ant
Cmd.new("go", ["test", "-run=TestDocs", "./tests"]).string() // "go test -run=TestDocs ./tests"
```

A `Cmd` literal works as well as one from `Cmd.new`:

```ant
Cmd{ program: "make", args: ["all"] }.string() // "make all"
```

A pipeline of one stage is that stage's line:

```ant
Cmd.pipe([Cmd.new("ls", ["-l"])]).string() // "ls -l"
```

Call it in a pipe after building the command:

```ant
Cmd.new("tar", ["czf", "backup.tgz", "notes"]) | Cmd.string() // "tar czf backup.tgz notes"
```

## Applications

Log each command before running it:

```ant
func runLogged(c: Cmd): Result[int, Error] {
  println("+ ${c.string()}")
  c.run()
}

let status = try runLogged(Cmd.new("echo", ["hello, world"]))
println("status ${status}")
```

```text
+ echo 'hello, world'
hello, world
status 0
```

A dry run prints what a script would do and runs nothing:

```ant
let dryRun = true
let steps = [
  Cmd.new("go", ["vet", "./..."]),
  Cmd.new("go", ["test", "./..."]),
  Cmd.new("git", ["tag", "-m", "release v1", "v1.0.0"]),
]

for step in steps {
  if dryRun {
    println(step.string())
    continue
  }
  try step.run()
}
```

```text
go vet ./...
go test ./...
git tag -m 'release v1' v1.0.0
```

Put the command in the message of an error, so the reader sees what failed:

```ant
func readCmd(c: Cmd): Result[string, Error] {
  Cmd.read(c) | Result.wrapErr("run ${c.string()}")
}

let checked = readCmd(Cmd.new("test", ["-d", "/no such directory"]))
if checked is Err {
  println(checked.error.message)
}
```

```text
run test -d '/no such directory': test: exit status 1
```

## Properties and relations

The line is quoted as a POSIX shell needs it, so `sh -c` runs it with the same arguments:

```ant
let c = Cmd.new("printf", ["%s|", "a b", "it's", "", "$HOME"])
let direct = try c.read()
let viaShell = try Cmd.new("sh", ["-c", c.string()]).read()
assert(direct == viaShell)
println(direct)
```

```text
a b|it's||$HOME|
```

The line of a pipeline is the lines of its stages joined by `|`:

```ant
let stages = [Cmd.new("sort"), Cmd.new("uniq", ["-c"])]
let line = Cmd.pipe(stages).string()
assert(stages.every({ line.contains(x.string()) }))
assert(line.contains("|"))
```

Two commands that differ only in `dir` or `env` have the same line:

```ant
let plain = Cmd.new("make", ["deploy"])
let secret = Cmd.new("make", ["deploy"], dir: "/srv/app", env: Map{ "TOKEN": "s3cr3t" })
assert(plain.string() == secret.string())
assert(plain != secret)
```

## Possible issues

A `Cmd` printed as a value shows its record, `env` included, so a secret reaches the log. Log
`Cmd.string` instead:

```ant
let c = Cmd.new("deploy", env: Map{ "TOKEN": "s3cr3t" })
"${c}"     // "Cmd{ program: \"deploy\", args: [], dir: null, env: Map{ \"TOKEN\": \"s3cr3t\" } }"
c.string() // "deploy"
```

Since `dir` and `env` are left out, pasting the line into a shell can run the command somewhere
else, or with another environment, than the script did. Log `dir` next to the line when it matters:

```ant
let build = Cmd.new("make", dir: "/srv/app")
"in ${build.dir ?? "."}: ${build.string()}" // "in /srv/app: make"
```

An argument with a newline keeps the newline inside its quotes, so the line for a log spans two
lines. When one line per command matters, replace newlines before logging, or print `args` as a
value:

```ant
let c = Cmd.new("echo", ["a\nb"])
c.string()  // "echo 'a\nb'"
"${c.args}" // "[\"a\\nb\"]"
```

## See also

[Cmd.new](new.md), [Cmd.pipe](pipe.md), [Cmd.run](run.md), [Cmd.read](read.md),
[Cmd.Pipeline](types/Pipeline.md), [Result.wrapErr](../Result/wrapErr.md)
