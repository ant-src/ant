<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.pipe

```ant signature
func Cmd.pipe(cmds: Array[Cmd]): Cmd.Pipeline
```

Cmd.pipe returns the pipeline of cmds, in order, holding a copy of the array. It panics when cmds is
empty.

<!-- end generated -->

## Details

- `Cmd.pipe` runs nothing. [Cmd.run](run.md), [Cmd.read](read.md), [Cmd.readBytes](readBytes.md),
  [Cmd.capture](capture.md) and [Cmd.start](start.md) run a pipeline as they run a `Cmd`.
- The stages run at once. The first reads the pipeline's input, each stage's output is the next
  one's input, and the last writes the pipeline's output.
- The pipeline's status is that of the rightmost stage that failed, or 0. A stage before the last
  that SIGPIPE ended does not count.
- Every stage writes standard error where the call sends it. Under `Cmd.capture` the stages' lines
  interleave in `stderr`, in no promised order.
- Each stage keeps its own `dir` and `env`.

## Basic examples

Join two commands, as a shell's `|` does:

```ant
let listing = Cmd.pipe([Cmd.new("ls"), Cmd.new("sort", ["-r"])])
listing.string()      // "ls | sort -r"
listing.cmds.length() // 2
listing.cmds[1].args  // ["-r"]
```

Run the pipeline and read what its last stage prints:

```ant
let sorted = Cmd.pipe([Cmd.new("printf", ["pear\nfig\napple\n"]), Cmd.new("sort")])
println(try Cmd.read(sorted))
```

```text
apple
fig
pear
```

## Scope

A pipeline of one command runs that command:

```ant
let solo = Cmd.pipe([Cmd.new("echo", ["alone"])])
println(try Cmd.read(solo))
```

```text
alone
```

The `input` of `Cmd.read` goes to the first stage:

```ant
let shout = Cmd.pipe([Cmd.new("tr", ["a-z", "A-Z"]), Cmd.new("rev")])
println(try Cmd.read(shout, input: "hello"))
```

```text
OLLEH
```

Each stage runs in its own `dir` and with its own `env`:

```ant
let p = Cmd.pipe([
  Cmd.new("sh", ["-c", "echo $GREETING from $(pwd)"], dir: "/", env: Map{ "GREETING": "hi" }),
  Cmd.new("tr", ["a-z", "A-Z"]),
])
println(try Cmd.read(p))
```

```text
HI FROM /
```

Every stage writes to the one `stderr` of [Cmd.capture](capture.md), in no promised order, so the
example sorts the lines:

```ant
let p = Cmd.pipe([
  Cmd.new("sh", ["-c", "echo first >&2; echo data"]),
  Cmd.new("sh", ["-c", "cat >/dev/null; echo second >&2"]),
])
let out = try Cmd.capture(p)
println(out.stderr.lines().sort())
```

```text
["first", "second"]
```

[Cmd.capture](capture.md) gives the pipeline's status and each stage's own in `statuses`:

```ant
let p = Cmd.pipe([Cmd.new("cat"), Cmd.new("sh", ["-c", "cat; exit 3"])])
let out = try Cmd.capture(p, input: "data")
println(out.status)
println(out.statuses)
println(out.stdout)
```

```text
3
[0, 3]
data
```

## Applications

List the distinct paths of the GET requests:

```ant
let log = "GET /\nPOST /login\nGET /about\nGET /\n"
let paths = Cmd.pipe([
  Cmd.new("grep", ["^GET"]),
  Cmd.new("cut", ["-d", " ", "-f", "2"]),
  Cmd.new("sort", ["-u"]),
])
println(try Cmd.read(paths, input: log))
```

```text
/
/about
```

Build the stages from data: keep the lines that contain every word.

```ant
func matchingAll(words: Array[string]): Cmd.Pipeline {
  Cmd.pipe(words.map({ Cmd.new("grep", ["-F", x]) }))
}

let notes = "red apple\ngreen apple\nred cherry\n"
println(try Cmd.read(matchingAll(["red", "apple"]), input: notes))
```

```text
red apple
```

Log the pipeline before you run it:

```ant
let p = Cmd.pipe([Cmd.new("echo", ["it's done"]), Cmd.new("tr", ["a-z", "A-Z"])])
println("running ${p.string()}")
println(try Cmd.read(p))
```

```text
running echo 'it'\''s done' | tr a-z A-Z
IT'S DONE
```

## Properties and relations

The pipeline keeps the commands in order, and copies the array, so a later push changes nothing:

```ant
let echo = Cmd.new("echo", ["hi"])
let upper = Cmd.new("tr", ["a-z", "A-Z"])
let cmds = [echo]
let p = Cmd.pipe(cmds)
cmds.push(upper)
assert(p.cmds == [echo])
assert(try Cmd.read(p) == "hi")
```

A pipeline of one command reads what the command reads:

```ant
let greet = Cmd.new("echo", ["hello"])
assert(try Cmd.read(Cmd.pipe([greet])) == try Cmd.read(greet))
```

## Possible issues

`Cmd.pipe` panics on an empty array:

```ant
Cmd.pipe([]) // panic: Cmd.pipe needs at least one command
```

A `Cmd.Pipeline` literal can hold no commands, and such a pipeline panics wherever it runs:

```ant
let p = Cmd.Pipeline{ cmds: [] }
let out = Cmd.read(p) // panic: Cmd.read needs a pipeline of at least one command
```

No shell runs the commands, so a `|` in the arguments is an argument like any other:

```ant
println(try Cmd.read(Cmd.new("echo", ["b a", "|", "tr", "a-z", "A-Z"])))
```

```text
b a | tr a-z A-Z
```

A stage that fails fails the pipeline even when a later stage succeeds, so `false | cat` is nonzero:

```ant
let out = try Cmd.capture(Cmd.pipe([Cmd.new("false"), Cmd.new("cat")]))
println(out.status)
println(out.statuses)
```

```text
1
[1, 0]
```

SIGPIPE ends a stage whose reader stops reading, and a stage before the last that SIGPIPE ended does
not count, so `yes | head -n 1` succeeds. Its own status, 141, is still in `statuses`. A stage that
ignores SIGPIPE and exits nonzero still counts, and on Windows every nonzero stage counts:

```ant
let out = try Cmd.capture(Cmd.pipe([Cmd.new("yes"), Cmd.new("head", ["-n", "1"])]))
println(out.status)
println(out.statuses)
```

```text
0
[141, 0]
```

When two stages run the same program, the error of [Cmd.read](read.md) names the stage, counting
from 1:

```ant
let p = Cmd.pipe([Cmd.new("echo", ["ant"]), Cmd.new("grep", ["a"]), Cmd.new("grep", ["x"])])
println(Cmd.read(p))
```

```text
Err{ error: Error{ message: "grep (stage 3): exit status 1" } }
```

## See also

[Cmd.new](new.md) · [Cmd.string](string.md) · [Cmd.read](read.md) · [Cmd.capture](capture.md) ·
[Cmd.run](run.md) · [Cmd.start](start.md) · [Cmd.Pipeline](types/Pipeline.md) ·
[Os.openPipe](../Os/openPipe.md) · [File.pipe](../File/pipe.md)
