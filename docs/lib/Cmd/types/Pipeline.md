<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.Pipeline

```ant signature
type Cmd.Pipeline = { cmds: Array[Cmd] }
```

Cmd.Pipeline is commands run at once, each one's standard output the next one's standard input, made
by [`Cmd.pipe`](../pipe.md). Every function of Cmd that runs a Cmd runs a pipeline too: the first
stage reads its input, the last writes its output, and every stage writes standard error. Its status
is that of the rightmost stage that failed, or 0; a stage before the last that SIGPIPE ended does
not count, since only a later stage that stopped reading ends it so. A pipeline of no commands
panics wherever it is run.

<!-- end generated -->

## Details

- A `Cmd.Pipeline` is a record. [Cmd.pipe](../pipe.md) makes one, and so does a literal. Its `cmds`
  is an ordinary array: a change to it shows the next time the pipeline runs.
- `Cmd.pipe` keeps a copy of the array it is given, so a later change to that array changes no
  pipeline; change `p.cmds` instead.
- Every function that takes a `Cmd` to run takes a `Cmd.Pipeline` too. A stage is a `Cmd`, never
  another pipeline; spread one pipeline's `cmds` into a new one to join them.

## Basic examples

[Cmd.pipe](../pipe.md) returns a `Cmd.Pipeline`, whose `cmds` are the stages in order:

```ant
let p = Cmd.pipe([Cmd.new("echo", ["hi"]), Cmd.new("tr", ["a-z", "A-Z"])])
p.cmds.length()   // 2
p.cmds[1].program // "tr"
p.string()        // "echo hi | tr a-z A-Z"
```

Run it as you would run a `Cmd`:

```ant
let p = Cmd.pipe([Cmd.new("echo", ["hi"]), Cmd.new("tr", ["a-z", "A-Z"])])
println(try Cmd.read(p))
```

```text
HI
```

## Scope

### A literal

A `Cmd.Pipeline` literal makes the same pipeline as [Cmd.pipe](../pipe.md):

```ant
let p = Cmd.Pipeline{ cmds: [Cmd.new("echo", ["hi"]), Cmd.new("rev")] }
println(try p.read())
```

```text
ih
```

### Changing the stages

`cmds` is an array, so push a stage onto it or change a stage's arguments in place:

```ant
let p = Cmd.pipe([Cmd.new("printf", ["b\\na\\nc\\n"]), Cmd.new("sort")])
println(try p.read())
p.cmds[1].args.push("-r")
p.cmds.push(Cmd.new("head", ["-n", "1"]))
println(p.string())
println(try p.read())
```

```text
a
b
c
printf 'b\na\nc\n' | sort -r | head -n 1
c
```

### Input to the first stage

`input:` feeds the first stage when the pipeline runs with [Cmd.read](../read.md),
[Cmd.readBytes](../readBytes.md) or [Cmd.capture](../capture.md):

```ant
let p = Cmd.pipe([Cmd.new("tr", ["a-z", "A-Z"]), Cmd.new("rev")])
p.read(input: "hello") // Ok{ value: "OLLEH" }
```

### Every way to run

[Cmd.capture](../capture.md) keeps each stage's status in `statuses`, and [Cmd.start](../start.md)
runs the pipeline in the background:

```ant
let p = Cmd.pipe([Cmd.new("echo", ["one two"]), Cmd.new("tr", ["a-z", "A-Z"])])
let out = try p.capture()
println(out.stdout.trim())
println(out.statuses)
let started = try p.start(stdout: null)
println(try started.wait())
println(try p.run())
```

```text
ONE TWO
[0, 0]
0
ONE TWO
0
```

### Taking a command or a pipeline

A function that takes `Cmd | Cmd.Pipeline` accepts either, and `is` tells them apart:

```ant
func stages(c: Cmd | Cmd.Pipeline): int {
  if c is Cmd.Pipeline { c.cmds.length() } else { 1 }
}

println(stages(Cmd.new("ls")))
println(stages(Cmd.pipe([Cmd.new("ls"), Cmd.new("wc", ["-l"])])))
```

```text
1
2
```

## Applications

### Steps of a build

Keep commands and pipelines in one list, log each, and stop at the first that fails:

```ant
type Step = { name: string, run: Cmd | Cmd.Pipeline }

func runAll(steps: Array[Step]): Result[void, Error] {
  for step in steps {
    println("${step.name}: ${step.run.string()}")
    try step.run.read().wrapErr(step.name)
  }
  Result.ok(void)
}

let steps = [
  Step{ name: "greet", run: Cmd.new("echo", ["hi"]) },
  Step{ name: "check", run: Cmd.pipe([Cmd.new("echo", ["ant"]), Cmd.new("grep", ["bee"])]) },
  Step{ name: "never", run: Cmd.new("echo", ["unreached"]) },
]
println(runAll(steps))
```

```text
greet: echo hi
check: echo ant | grep bee
Err{ error: Error{ message: "check: grep: exit status 1" } }
```

### Extend a pipeline

Spread the stages of one pipeline into a new one to add a stage, and leave the first unchanged:

```ant
let words = Cmd.pipe([Cmd.new("echo", ["to be or not to be"]), Cmd.new("tr", [" ", "\\n"])])
let distinct = Cmd.pipe([...words.cmds, Cmd.new("sort", ["-u"]), Cmd.new("wc", ["-l"])])
println(try words.read())
let count = try distinct.read()
println(count.trim())
```

```text
to
be
or
not
to
be
4
```

### Name the stages

List what a pipeline runs, one program per stage:

```ant
let p = Cmd.pipe([Cmd.new("git", ["log"]), Cmd.new("grep", ["fix"]), Cmd.new("wc", ["-l"])])
p.cmds.map({ x.program }) | String.join(" -> ") // "git -> grep -> wc"
```

## Properties and relations

[Cmd.pipe](../pipe.md) returns the literal with a copy of its array:

```ant
let cmds = [Cmd.new("echo", ["hi"]), Cmd.new("cat")]
assert(Cmd.pipe(cmds) == Cmd.Pipeline{ cmds: cmds })
```

Since the array is copied, pushing onto it afterward leaves the pipeline as it was:

```ant
let cmds = [Cmd.new("echo", ["hi"]), Cmd.new("cat")]
let p = Cmd.pipe(cmds)
cmds.push(Cmd.new("wc", ["-c"]))
assert(p.string() == "echo hi | cat")
```

[Cmd.string](../string.md) of a pipeline joins the strings of its stages with `" | "`:

```ant
let p = Cmd.pipe([Cmd.new("echo", ["a b"]), Cmd.new("tr", ["a-z", "A-Z"]), Cmd.new("cat")])
let joined = p.cmds.map({ x.string() }) | String.join(" | ")
assert(p.string() == joined)
```

The status is the rightmost failure, whatever the stages after it return:

```ant
let p = Cmd.pipe([
  Cmd.new("sh", ["-c", "exit 2"]),
  Cmd.new("sh", ["-c", "cat; exit 3"]),
  Cmd.new("cat"),
])
let out = try p.capture()
assert(out.status == 3)
assert(out.statuses == [2, 3, 0])
```

A stage before the last that SIGPIPE ended does not count. `yes` ends by SIGPIPE once `head` stops
reading; the pipeline still succeeds, and `statuses` keeps the 141:

```ant
let out = try Cmd.pipe([Cmd.new("yes"), Cmd.new("head", ["-n", "1"])]).capture()
assert(out.status == 0)
assert(out.statuses == [141, 0])
```

## Possible issues

### A pipeline is not a stage

`cmds` holds `Cmd` values only, so a pipeline does not nest in another:

```ant
let inner = Cmd.pipe([Cmd.new("echo", ["a"]), Cmd.new("cat")])
let outer = Cmd.pipe([inner, Cmd.new("wc", ["-l"])]) // error: cannot use Cmd.Pipeline where Cmd is expected
```

Spread its stages instead, as in [Extend a pipeline](#extend-a-pipeline).

### Which stage failed

The error names the program of the rightmost stage that failed, and adds its stage number, counting
from 1, when another stage runs the same program:

```ant
let p = Cmd.pipe([Cmd.new("echo", ["a"]), Cmd.new("grep", ["b"]), Cmd.new("grep", ["c"])])
p.read() // Err{ error: Error{ message: "grep (stage 3): exit status 1" } }
```

### An empty pipeline

A literal with no commands makes a pipeline, and the run panics:

```ant
let p = Cmd.Pipeline{ cmds: [] }
let out = p.capture() // panic: Cmd.capture needs a pipeline of at least one command
```

### Printing shows the environment

Printing a pipeline shows every field of every stage, `env` included. Log it with
[Cmd.string](../string.md), which leaves `env` and `dir` out:

```ant
let p = Cmd.pipe([
  Cmd.new("sh", ["-c", "deploy"], env: Map{ "TOKEN": "s3cret" }),
  Cmd.new("tee", ["deploy.log"]),
])
println(p)
println(p.string())
```

```text
Cmd.Pipeline{ cmds: [Cmd{ program: "sh", args: ["-c", "deploy"], dir: null, env: Map{ "TOKEN": "s3cret" } }, Cmd{ program: "tee", args: ["deploy.log"], dir: null, env: null }] }
sh -c deploy | tee deploy.log
```

## See also

[Cmd.pipe](../pipe.md) · [Cmd.string](../string.md) · [Cmd.read](../read.md) ·
[Cmd.capture](../capture.md) · [Cmd.run](../run.md) · [Cmd.start](../start.md) ·
[Cmd.Output](Output.md) · [Cmd](../index.md)
