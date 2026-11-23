<!-- status: reviewed -->

# Ant for Python developers

Ant keeps what makes Python good for scripts: a file of statements that runs top to bottom, no
`main`, and a standard library for files, processes, and JSON. It differs in three places. Types are
static, and `ant check` checks them before you run. Expected failures are values, not exceptions.
Concurrency is goroutines, not threads or `asyncio`.

## A script, side by side

This Python script prints the three most common words of the file named on its command line:

```python
import sys
from collections import Counter

text = open(sys.argv[1]).read()
counts = Counter(text.lower().split())
for word, n in counts.most_common(3):
    print(f"{word}: {n}")
```

The same script in Ant, run as `ant run words.ant notes.txt`:

<!-- test: file notes.txt
the ant and the bee
an ant on the hill
the ant
-->
<!-- test: args notes.txt -->

```ant
let text = try File.read(Os.args()[0])
let counts: Map[string, int] = Map{}
for word in text.toLower().fields() {
  counts[word] = (counts[word] ?? 0) + 1
}
let top = counts.entries() | Array.sortBy({ -x.value }) | Iter.fromArray() | Iter.take(3)
for { key, value } in top {
  println("${key}: ${value}")
}
```

```text
the: 4
ant: 3
and: 1
```

- `Os.args()` holds only the arguments: index 0 is the first one, not the script path.
- `File.read` returns a `Result`. `try` unwraps it, or ends the script with the error on standard
  error and status 1.
- A map read such as `counts[word]` is `int | null`, so `?? 0` supplies the missing count.
- `let` names a value that never changes, and `var` one you reassign. An undeclared name is an
  error, not a new variable.
- `"${key}"` interpolates, as an f-string does.

## Equivalents

### The language

| Python                                   | Ant                                                                |
| ---------------------------------------- | ------------------------------------------------------------------ |
| `n = 0`                                  | `var n = 0`, or `let n = 0` when it never changes                  |
| `None`                                   | `null`                                                             |
| `a and not b`                            | `a && !b`                                                          |
| `x if cond else y`                       | `if cond { x } else { y }`                                         |
| `def area(w, h=1): return w * h`         | `func area(w: int, h: int = 1): int { w * h }`                     |
| `area(3, h=2)`                           | `area(3, h: 2)`                                                    |
| `lambda x: x * 2`                        | `{ x * 2 }`                                                        |
| `@dataclass class Point: x: int; y: int` | `type Point = { x: int, y: int }`                                  |
| `Point(1, 2)`                            | `Point{ x: 1, y: 2 }`                                              |
| `def norm(self): ...` inside the class   | `func Point.norm(p: Point): float64 { ... }`, called as `p.norm()` |
| `isinstance(p, Point)`                   | `p is Point`, when `p`'s type is a union such as `Shape`           |
| `str(n)`, `int(s)`                       | `"${n}"`, `s.toInt()`, which is `null` when not a number           |
| `float(n) / 2`                           | `float64(n) / 2.0`                                                 |
| `a, b = b, a`                            | `[a, b] = [b, a]`                                                  |
| `raise ValueError("bad")`                | `return Result.fail("bad")`                                        |
| `assert n > 0`                           | `assert(n > 0)`                                                    |
| `from util import f`                     | `import "./util.ant" { f }`                                        |
| `def test_f(): ...` with pytest          | `test "f" { ... }` with `ant test`                                 |

### Strings

| Python                      | Ant                                                                |
| --------------------------- | ------------------------------------------------------------------ |
| `f"{name}: {n}"`            | `"${name}: ${n}"`                                                  |
| `"""..."""`                 | `"..."`, which may span lines                                      |
| `r"C:\temp"`                | `` `C:\temp` ``                                                    |
| `len(s)`                    | `len(s)` or `s.length()`, both counting code points as Python does |
| `s.strip()`, `s.lower()`    | `s.trim()`, `s.toLower()`                                          |
| `s.split(",")`, `s.split()` | `s.split(",")`, `s.fields()`                                       |
| `s.splitlines()`            | `s.lines()`                                                        |
| `", ".join(parts)`          | `String.join(parts, ", ")`                                         |
| `s[1:3]`                    | `s.slice(1, 3)`, which panics when a bound is past the end         |
| `f"{x:.2f}"`                | `"${String.fromFloat(x, 2)}"`                                      |
| `f"{n:08x}"`                | `String.padStart(String.fromInt(n, 16), 8, "0")`                   |
| `s.find("ab")`              | `s.indexOf("ab")`, which is `null` when absent, not -1             |
| `"ab" in s`                 | `s.contains("ab")`                                                 |
| `s.startswith("#")`         | `s.startsWith("#")`                                                |

### Collections

| Python                                             | Ant                                                                           |
| -------------------------------------------------- | ----------------------------------------------------------------------------- |
| `xs.append(3)`                                     | `xs.push(3)`                                                                  |
| `xs[-1]`                                           | `xs.last()`, which is `null` when `xs` is empty                               |
| `xs.sort()`, `sorted(xs)`                          | `let sorted = xs.sort()`, which returns a new array and leaves `xs` unchanged |
| `x in xs`                                          | `xs.contains(x)`                                                              |
| `[x * 2 for x in xs if x > 0]`                     | `xs \| Array.filter({ x > 0 }) \| Array.map({ x * 2 })`                       |
| `sorted(words, key=len)`                           | `words.sortBy({ x.length() })`                                                |
| `sum(xs)`                                          | `xs.sum() ?? 0`                                                               |
| `any(x > 0 for x in xs)`, `all(x > 0 for x in xs)` | `xs.some({ x > 0 })`, `xs.every({ x > 0 })`                                   |
| `for i, x in enumerate(xs)`                        | `for { index, value } in xs.entries()`                                        |
| `range(10)`                                        | `Array.range(0, 10)`, or lazily `Iter.range(0, 10)`                           |
| `d = {"a": 1}`                                     | `let d = Map{ "a": 1 }`                                                       |
| `d.get(k, 0)`                                      | `d[k] ?? 0`                                                                   |
| `k in d`                                           | `d.has(k)`                                                                    |
| `for k, v in d.items()`                            | `for { key, value } in d`                                                     |
| `set(xs)`                                          | `Set.of(xs)`                                                                  |

### Files, processes, and JSON

| Python                                                  | Ant                                                    |
| ------------------------------------------------------- | ------------------------------------------------------ |
| `open(p).read()`                                        | `try File.read(p)`                                     |
| `open(p, "w").write(s)`                                 | `try File.write(p, s)`                                 |
| `with open(p) as f:`                                    | `let f = try File.open(p)`, then `defer { f.close() }` |
| `f.flush(); os.fsync(f.fileno())`                       | `try f.sync()`                                         |
| `os.chown(p, uid, gid)`                                 | `try File.chown(p, uid, gid)`                          |
| `os.path.join(a, b)`                                    | `Path.join([a, b])`                                    |
| `os.path.exists(p)`                                     | `File.exists(p)`                                       |
| `os.listdir(d)`, `os.scandir(d)`                        | `try File.list(d)`, `try File.readDir(d)`              |
| `glob.glob("*.txt")`, `os.walk(d)`                      | `try File.glob("*.txt")`, `try File.walk(d)`           |
| `sys.argv[1:]`                                          | `Os.args()`                                            |
| `os.environ.get("HOME")`                                | `Os.env("HOME")`                                       |
| `sys.exit(2)`                                           | `Os.exit(2)`                                           |
| `subprocess.run(["git", "status"]).returncode`          | `try Cmd.run(Cmd.new("git", ["status"]))`              |
| `subprocess.check_output(["git", "status"], text=True)` | `try Cmd.read(Cmd.new("git", ["status"]))`             |
| `json.loads(text)`                                      | `try Json.parse(text)`                                 |
| `json.dumps(value, indent=2)`                           | `Json.stringify(value, indent: "  ")`                  |

A close hands what was written to the OS but does not wait for the disk. `f.sync()` flushes and then
has the OS write the file to storage, as `os.fsync` does.

`Cmd` runs the program with no shell, so each argument reaches it as written. `Cmd.run` returns the
exit status; `Cmd.read` returns the output without its final line break, and a nonzero status is an
`Err`. `Json.parse` gives a `Json.Value`, which you narrow with `is` before you read it, and every
JSON number in it is a `float64`.

### Concurrency

| Python                                               | Ant                                                  |
| ---------------------------------------------------- | ---------------------------------------------------- |
| `threading.Thread(target=f).start()`                 | `go { f() }`                                         |
| `pool.map(f, xs)` in a `ThreadPoolExecutor`          | `Parallel.map(xs, f, limit: 8)`                      |
| `ProcessPoolExecutor().map(f, xs)`, `os.cpu_count()` | `Parallel.map(xs, f, limit: Os.cpus())`, `Os.cpus()` |
| `asyncio.gather(a(), b())`                           | `Parallel.run([{ ra = a(x) }, { rb = b(x) }])`       |
| `async def` and `await`                              | ordinary functions and calls                         |
| `queue.Queue()`                                      | `let q: Channel[Job] = Channel.new(100)`             |
| `threading.Lock()`                                   | `Mutex.new()`                                        |
| `time.sleep(0.5)`                                    | `Time.sleep(500ms)`                                  |

The callbacks of `Parallel.run` receive the context as `x` and return nothing, so results go into
variables declared before the call. A channel's capacity bounds it, where a `queue.Queue()` grows
without limit: a `send` to a full channel waits, and `Channel.new()` with no capacity is a
rendezvous, whose `send` waits until a receiver takes the value. The declared type gives
`Channel.new` its element type, which its arguments do not fix.

`Os.cpus()` is not `os.cpu_count()`, which counts every CPU of the host. It is the number of CPUs
the program may use at once: fewer under a container's CPU limit, and the `GOMAXPROCS` environment
variable's value when that is a positive number. It is the limit for CPU-bound work, and it can
change while the program runs, so read it when you size the work.

## Differences that bite

### Types are checked by `ant check`

Every parameter has a type, and `ant check` reports a mismatch before the program runs. `ant run`
does not check, so run `ant check` in your editor and CI. The checker rejects habits that Python
allows: an array as a condition, `1 == 1.0`, `s[0]` on a string, a value of another type assigned to
a name. See [Types](../guide/types.md#mistakes-to-avoid),
[Strings](../guide/strings.md#habits-from-other-languages), and
[Variables](../guide/values.md#coming-from-other-languages).

Numbers are fixed-size. An `int` has 64 bits and overflows with a panic instead of growing. Python's
`/` always gives a float; Ant's `/` on two `int`s gives an `int` that truncates toward zero, and `%`
takes the sign of the dividend, where Python's `//` rounds down and its `%` takes the sign of the
divisor:

```ant
7 / 2            // 3
-7 / 2           // -3
-7 % 2           // -1
float64(7) / 2.0 // 3.5
```

```ant
let big = 9223372036854775807
big + 1 // panic: integer overflow
```

### Errors are values, not exceptions

A function that can fail returns a `Result`, so you see the failure at the call. Where Python
catches an `OSError`, Ant reads the `Result`:

```python
try:
    text = open("settings.ini").read()
except OSError:
    text = ""
```

```ant
let text = File.read("settings.ini") | Result.unwrapOr("")
println(text.isEmpty())
```

```text
true
```

`try` passes the error to the caller, and `panic` is for bugs, not for control flow. See
[Errors](../guide/errors.md#coming-from-other-languages).

### `null` only where the type says

A `string` is never `null`; a `string | null` may be, and you narrow it before you use it. A map
read, `xs.last()`, and `s.toInt()` return a nullable value, where Python raises `KeyError`,
`IndexError`, or `ValueError`. A condition must be a `bool`, so write `if name != null`, not
`if name:`. Indexes and slices do not count from the end or clamp: `xs[-1]` and `xs.slice(0, 10)` on
a shorter array panic; use `xs.last()` or `Iter.take`. See
[Null](../guide/null.md#coming-from-other-languages).

### Two Python traps are gone

A default value runs on every call, so a default array is new each time. Each loop iteration binds a
fresh variable, so a closure created in the loop keeps that iteration's value:

```ant
func add(n: int, into: Array[int] = []): Array[int] {
  into.push(n)
  into
}

println(add(1))
println(add(2))

let getters: Array[func(): int] = []
for i in [1, 2, 3] {
  getters.push({ i })
}
println(getters.map({ x() }))
```

```text
[1]
[2]
[1, 2, 3]
```

### Modules are files

An import names a file, extension included, and a file is either a script or a library: there is no
`if __name__ == "__main__"`. A library marks what other files may import with `export`, as in
`export func f()`; everything else is private. See
[Modules](../guide/modules.md#coming-from-other-languages).

### Goroutines instead of threads and `asyncio`

There is no `async` and no `await`. A blocking call, such as reading a file or sleeping, blocks only
its goroutine, so an ordinary function runs concurrently as it is. `Parallel.map` runs it on many
inputs and waits for all of them:

```ant
func fetch(name: string, ctx: Context): string {
  Time.sleep(10ms, ctx: ctx)
  name.toUpper()
}

println(Parallel.map(["ann", "bob", "eve"], fetch, limit: 3))
```

```text
["ANN", "BOB", "EVE"]
```

The function takes the element and a `Context`, which it passes to its blocking calls.
`Parallel.map` returns `null` only when it is cancelled before every element starts, so the result
needs `??` before you use it as an array.

The program exits when the script ends, even with goroutines still running, and a `go` block returns
no result. See [Concurrency](../guide/concurrency.md#pitfalls) and
[No async](../design/no-async.md).

## See also

- [Tour](../tour/01-hello.md): the language one step at a time
- [Idiomatic Ant](../guide/idiomatic.md)
- [Files](../lib/guides/files.md), [System](../lib/guides/system.md), and
  [Data formats](../lib/guides/data-formats.md) in the standard library
- How-to: [read and write a whole file](../howto/files/read-write.md),
  [read a large file line by line](../howto/files/lines.md),
  [run a command](../howto/processes/run.md),
  [process a list in parallel](../howto/concurrency/parallel-map.md)
- [Coming from JavaScript](javascript.md) and [Coming from Go](go.md)
