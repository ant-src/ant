<!-- status: reviewed -->

# Ant for Python developers

Ant keeps what makes Python good for scripts: a file of statements that runs from top to bottom, no
`main`, and a standard library for files, processes, and JSON. It differs in a few places, and this
page is about those. It does not try to cover the whole language. The [Tour](../tour/01-hello.md)
does that.

## A first script

This Python script counts the words in the file named on its command line:

```python
import sys
from collections import Counter

text = open(sys.argv[1]).read()
counts = Counter(text.lower().split())
for word, n in counts.most_common():
    print(f"{word}: {n}")
```

The same script in Ant, run as `ant run words.ant notes.txt`:

<!-- test: file notes.txt
The ant and the bee
the ant
-->
<!-- test: args notes.txt -->

```ant
let text = try File.read(Os.args()[0])
let counts = text.toLower().fields().countBy({ x })

for { key, value } in counts.entries().sortBy({ -x.value }) {
  println("${key}: ${value}")
}
```

```text
the: 3
ant: 2
and: 1
bee: 1
```

A few things to notice:

- `let` names a value that never changes, and `var` one that you reassign. Using a name you never
  declared is an error, not a new variable.
- `File.read` can fail, so `try` goes in front of it. If the file is missing, the script stops there
  and prints the error.
- `Os.args()` holds only the arguments. Index 0 is the first one, not the script's path.
- `countBy` plays the part of `Counter`: a map from each word to how often it occurs, in the order
  the words first appear.
- `{ -x.value }` is a short function, like a `lambda`. Its one parameter is always named `x`.
- `"${key}"` works like an f-string.

## What carries over

| Python                           | Ant                                            |
| -------------------------------- | ---------------------------------------------- |
| `None`                           | `null`                                         |
| `a and not b`                    | `a && !b`                                      |
| `x if cond else y`               | `if cond { x } else { y }`                     |
| `def area(w, h=1): return w * h` | `func area(w: int, h: int = 1): int { w * h }` |
| `area(3, h=2)`                   | `area(3, h: 2)`                                |
| `lambda x: x * 2`                | `{ x * 2 }`                                    |
| `f"{name}: {n}"`                 | `"${name}: ${n}"`                              |
| `len(s)`, `s.strip()`            | `s.length()`, `s.trim()`                       |
| `xs.append(3)`                   | `xs.push(3)`                                   |
| `[x * 2 for x in xs if x > 0]`   | `xs.filter({ x > 0 }).map({ x * 2 })`          |
| `for i, x in enumerate(xs)`      | `for { index, value } in xs.entries()`         |
| `d = {"a": 1}`, `d.get(k, 0)`    | `let d = Map{ a: 1 }`, `d[k] ?? 0`             |
| `open(p).read()`                 | `try File.read(p)`                             |
| `json.loads(text)`               | `try Json.parse(text)`                         |
| `from util import f`             | `import "./util.ant" { f }`                    |

## What works differently

### Types are checked before you run

Every parameter has a type, and `ant check` tells you about a mismatch before the program runs:

```ant
func double(n: int): int { n * 2 }

println(double("21")) // error: cannot use string where int is expected
```

`ant run` does not check types, so run `ant check` too, or let your editor run it with the
[Ant extension](/vscode). Ant infers the type of a variable from its value, so most lines carry no
annotation.

### Errors are values, not exceptions

A function that can fail returns a `Result`, so you see the failure where you make the call. Where
Python catches an exception:

```python
try:
    text = open("settings.ini").read()
except OSError:
    text = ""
```

Ant reads the `Result`:

```ant
let text = File.read("settings.ini") | Result.unwrapOr("")
println(text.isEmpty())
```

```text
true
```

`try` passes an error on to the caller, as an exception would travel up, but only where you write
it. `panic` exists for bugs, not for control flow. See [Errors](../guide/errors.md).

### Missing values are `null`, not exceptions

Where Python raises `KeyError` or `ValueError`, Ant gives you `null`, and the type says so:

```ant
let ages = Map{ ann: 31 }
ages["bob"]      // null
ages["bob"] ?? 0 // 0
"42".toInt()     // 42
"4x".toInt()     // null
[1, 2, 3].last() // 3
```

A condition must be a `bool`, so write `if name != null`, not `if name:`. Indexes do not count from
the end: use `xs.last()` instead of `xs[-1]`. See [Null](../guide/null.md).

### Numbers have a fixed size

An `int` has 64 bits. Dividing two `int`s gives an `int`, as `//` does in Python, and an `int` never
turns into a float on its own:

```ant
7 / 2            // 3
float64(7) / 2.0 // 3.5
```

An `int` that grows past its limit panics instead of growing.

### No `async`

A call that waits, such as reading a file or sleeping, pauses only its own goroutine. There is no
`async` and no `await`: ordinary functions run concurrently as they are. To run a function over a
list in parallel, see [process a list in parallel](../howto/concurrency/parallel-map.md).

## Where to go next

- The [Tour](../tour/01-hello.md) walks through the language one small program at a time.
- [Files](../lib/guides/files.md), [System](../lib/guides/system.md), and
  [Data formats](../lib/guides/data-formats.md) cover the standard library for scripts.
- How-to: [read and write a file](../howto/files/read-write.md) and
  [run a command](../howto/processes/run.md).
- [Coming from JavaScript](javascript.md) and [Coming from Go](go.md).
