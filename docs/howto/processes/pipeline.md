<!-- status: reviewed -->

# Build a pipeline

This script prints the two addresses that appear most often in an access log, by running
`cut | sort | uniq -c | sort -rn | head` the way a shell's `|` does. Save it as `top.ant` and run it
with `ant run top.ant access.log`:

<!-- test: file access.log
10.0.0.2 GET /
10.0.0.1 GET /about
10.0.0.2 GET /login
10.0.0.3 GET /
10.0.0.2 POST /login
10.0.0.1 GET /
-->
<!-- test: cmd ant run top.ant access.log -->

```ant [top.ant]
if Os.args().isEmpty() {
  eprintln("usage: top <log>")
  Os.exit(2)
}
let top = Cmd.pipe([
  Cmd.new("cut", ["-d", " ", "-f", "1", Os.args()[0]]),
  Cmd.new("sort"),
  Cmd.new("uniq", ["-c"]),
  Cmd.new("sort", ["-rn"]),
  Cmd.new("head", ["-n", "2"]),
])
eprintln("running ${top.string()}")
let out = try Cmd.read(top)
for line in out.lines() {
  let fields = line.fields()
  println("${fields[1]}: ${fields[0]} requests")
}
```

```text
10.0.0.2: 3 requests
10.0.0.1: 2 requests
```

1. Build each stage with [Cmd.new](../../lib/Cmd/new.md): the program, then its arguments. No shell
   runs the stages, so give each argument as it is, without quotes.
2. Join the stages with [Cmd.pipe](../../lib/Cmd/pipe.md). Each stage's output is the next one's
   input. The stages run at once.
3. Log the pipeline with [Cmd.string](../../lib/Cmd/string.md), which quotes it as a shell would:
   `cut -d ' ' -f 1 access.log | sort | uniq -c | sort -rn | head -n 2`.
4. Run it with [Cmd.read](../../lib/Cmd/read.md), which returns what the last stage prints, without
   its final newline. When a stage fails, `try` ends the script with the stage's error and status 1:

   ```sh
   $ ant run top.ant missing.log
   running cut -d ' ' -f 1 missing.log | sort | uniq -c | sort -rn | head -n 2
   cut: exit status 1: cut: missing.log: No such file or directory
   ```

   When two stages run the same program, the error names the stage too, counting from 1:
   `sort (stage 4): exit status 2: ...`.

5. Split the output with [String.lines](../../lib/String/lines.md) and each line with
   [String.fields](../../lib/String/fields.md), which drops the padding `uniq -c` puts before the
   count.

## Feed the pipeline from the program

Pass text as `input:`. The first stage reads it:

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

## Check the status of each stage

The pipeline fails when any stage fails, as under bash's `pipefail`. `grep` exits with 1 when no
line matches, so [Cmd.read](../../lib/Cmd/read.md) returns an `Err` for an empty result:

```ant
let errors = Cmd.pipe([Cmd.new("grep", ["ERROR"]), Cmd.new("sort", ["-u"])])
println(Cmd.read(errors, input: "INFO start\n"))
```

```text
Err{ error: Error{ message: "grep: exit status 1", kind: null } }
```

To tell "no match" from a real failure, run the pipeline with
[Cmd.capture](../../lib/Cmd/capture.md). It returns `Ok` whatever the status, with the pipeline's
status in `status` and each stage's own in `statuses`:

```ant
func errorLines(log: string): Result[Array[string], Error] {
  let errors = Cmd.pipe([Cmd.new("grep", ["ERROR"]), Cmd.new("sort", ["-u"])])
  let out = try Cmd.capture(errors, input: log)
  if out.statuses == [1, 0] {
    return Result.ok([])
  }
  if out.status != 0 {
    return Result.fail("${errors.string()}: exit status ${out.status}: ${out.stderr.trim()}")
  }
  Result.ok(out.stdout.lines())
}

println(try errorLines("INFO start\nERROR disk full\nERROR disk full\n"))
println(try errorLines("INFO start\n"))
```

```text
["ERROR disk full"]
[]
```

A stage that SIGPIPE ends because a later stage stopped reading does not fail the pipeline: in
`yes | head -n 1`, `head` exits after one line, the broken pipe ends `yes`, and the pipeline's
status is 0, where bash's `pipefail` gives 141. The stage's own entry in `statuses` is still 141. On
Windows, which has no SIGPIPE, every nonzero stage counts.

## Show the output on the terminal

Run the pipeline with [Cmd.run](../../lib/Cmd/run.md) to give the last stage the program's standard
output, and the first stage its standard input. It returns `Ok` of the pipeline's status, nonzero
included:

```ant
let words = Cmd.pipe([Cmd.new("printf", ["pear\nfig\npear\n"]), Cmd.new("sort", ["-u"])])
let status = try Cmd.run(words)
println("status ${status}")
```

```text
fig
pear
status 0
```

To read the output line by line while the pipeline runs, start it with
[Cmd.start](../../lib/Cmd/start.md); see [Stream a command's output](stream.md).

## Build the stages from data

A pipeline is an array of `Cmd` values, so build it with [Array.map](../../lib/Array/map.md). This
one keeps the lines that contain every word. `Cmd.pipe` panics on an empty array, so check the words
first:

```ant
func matchingAll(words: Array[string]): Cmd.Pipeline {
  if words.isEmpty() {
    return Cmd.pipe([Cmd.new("cat")])
  }
  Cmd.pipe(words.map({ Cmd.new("grep", ["-F", x]) }))
}

let notes = "red apple\ngreen apple\nred cherry\n"
println(try Cmd.read(matchingAll(["red", "apple"]), input: notes))
println(try Cmd.read(matchingAll([]), input: notes))
```

```text
red apple
red apple
green apple
red cherry
```

## See also

[Cmd.new](../../lib/Cmd/new.md) · [Cmd.pipe](../../lib/Cmd/pipe.md) ·
[Cmd.string](../../lib/Cmd/string.md) · [Cmd.read](../../lib/Cmd/read.md) ·
[Cmd.capture](../../lib/Cmd/capture.md) · [Cmd.run](../../lib/Cmd/run.md) ·
[Cmd.start](../../lib/Cmd/start.md) · [Cmd.Pipeline](../../lib/Cmd/types/Pipeline.md) ·
[Os.args](../../lib/Os/args.md) · [String.lines](../../lib/String/lines.md) ·
[String.fields](../../lib/String/fields.md) · [Run a command](run.md) ·
[Stream a command's output](stream.md) · [Stop or time out a process](stop.md)
