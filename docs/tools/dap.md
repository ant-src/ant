<!-- status: reviewed -->

# ant dap

`ant dap` is the debugger. It speaks the Debug Adapter Protocol over standard input and standard
output, so an editor runs a program with breakpoints, steps through it, and shows its goroutines,
stacks, and variables.

You do not run it yourself. The Ant plugins start it for you:

- [VS Code](/vscode): press F5, no `launch.json` needed. See
  [Debug in VS Code](../howto/debugging/vscode.md).
- [IntelliJ IDEA](/intellij): debug a run configuration. See
  [Debug in IntelliJ IDEA](../howto/debugging/intellij.md).
- [Neovim](/nvim): `:Ant debug` with nvim-dap. See [Debug in Neovim](../howto/debugging/neovim.md).

The rest of this page is for connecting another editor.

## Usage

```sh
ant dap
```

`ant dap` takes no arguments. It writes only protocol messages to standard output, each framed with
a `Content-Length` header as the Language Server Protocol frames them, and logs to standard error.
One `ant dap` process serves one client, one debug session, and one program. `ant dap` runs the
program it launches; it does not attach to a running one.

| Exit status | When                                                                      |
| ----------- | ------------------------------------------------------------------------- |
| `0`         | the client sent `disconnect`, or closed the connection after `terminated` |
| `1`         | the connection closed before `terminated` without `disconnect`            |
| `2`         | the usage is wrong, such as `ant dap main.ant`                            |

The program's own exit status arrives in the `exited` event.

## A session

1. Send `initialize`. The response lists the capabilities.
2. Send `launch` with the attributes below. The adapter loads the program as `ant run` does. If it
   does not load, the response is an error with `ant run`'s message.
3. Wait for the `initialized` event, then send your breakpoints and panic filters.
4. Send `configurationDone`. The program starts.
5. When the program ends, the adapter sends `exited` with its exit status, then `terminated`.

| Attribute     | Meaning                                              | Default                 |
| ------------- | ---------------------------------------------------- | ----------------------- |
| `program`     | the file to run, as given to `ant run`               | required                |
| `args`        | the program's arguments, as `Os.args()` returns them | `[]`                    |
| `cwd`         | the working directory                                | the program's directory |
| `env`         | variables added to the program's environment         | `{}`                    |
| `stdin`       | a file to use as the program's standard input        | none                    |
| `stopOnEntry` | stop before the first statement of the root module   | `false`                 |
| `test`        | a `-run` pattern for the tests of a test file        | every test              |

A minimal `launch` request:

```json
{
  "seq": 2,
  "type": "request",
  "command": "launch",
  "arguments": { "program": "main.ant", "cwd": "/home/me/shop" }
}
```

- A relative `program` or `stdin` is resolved against `cwd`.
- Without `stdin`, the program's standard input is at end of file. The program has no terminal, so a
  program that reads keys as they are typed runs with `ant run` instead.
- A `program` that is a test file runs its tests as `ant test` does, with `test` as the `-run`
  pattern. Each test's start and result arrive as `console` output, as `ant test -v` prints them.
  `test` with any other program is an error. The exit status in `exited` is `ant test`'s: 1 when a
  test failed.
- What the program writes arrives as `output` events of category `stdout` and `stderr`, at most 50
  milliseconds after the write, and always before the program stops.
- `terminate` interrupts the program, as the first Ctrl-C does under `ant run`; a second `terminate`
  cancels it. `disconnect` cancels the program if it still runs, and ends `ant dap`.

## Where a program stops

The program stops only at a stop point: before a statement of any block, before a top-level `var` or
`let` of a library module, and before each stage of a pipeline of two or more stages, at its `|`.

```ant
func total(prices: Array[int]): int {
  var sum = 0 // stop point: a statement of a function body
  for p in prices {
    sum += p // stop point: a statement of a loop body
  }
  sum
}

let prices = [120, 80, 45] // stop point: a statement of the script
let discounted = prices
  | Array.filter({ x > 50 })  // stop point: a stage, at its |
  | Array.map({ x * 9 / 10 }) // stop point: a stage, at its |
println(total(discounted))
```

```text
180
```

- A breakpoint on a line with no stop point, such as a blank line or a closing brace, moves to the
  next line that has one in the same function or module body, and the response reports that line.
  One that binds to nothing is reported unverified, with the reason.
- At a pipeline stage, the `Locals` scope starts with `(piped)`, the value flowing into the stage.
  Stopped at the `Array.filter` stage above, `(piped)` is `[120, 80, 45]`.
- When one goroutine stops, every goroutine stops, and `continue` resumes them all.
- The clock does not stop. A sleep or a deadline that passes while the program is stopped ends as
  soon as the program resumes.
- `next` steps over calls: it stops in the current frame, or in its caller once the frame returns.
  In a `for` loop, it goes to the next statement of the body, and from the end of the body into the
  next iteration.
- `stepIn` stops at the next stop point in any frame, and `stepOut` stops once the current frame
  returns. A step watches only its own goroutine; if that goroutine ends first, the program runs on.
  A step does not enter a `go` block; set a breakpoint inside it. `pause` stops every goroutine.
- A function breakpoint names a top-level function, such as `total` or `Point.dist`, and stops at
  the first statement of its body, in every module that declares it. It takes a `condition` and a
  `hitCondition` as a line breakpoint does.

## The program in the protocol

- Lines and columns start at 1 unless the client sets `linesStartAt1` or `columnsStartAt1` to
  `false` in `initialize`. A column counts UTF-16 code units. A source is an absolute path.
- Each goroutine is a thread. The main goroutine is thread 1, named `main`; the others are numbered
  from 2 and named `go at <file>:<line> #<id>`, after the `go` block that started them.
- A frame has up to three scopes: `Locals`, the bindings of its body; `Closure`, the bindings a
  lambda or `go` block captures; and `Module`, the module's top-level variables.
- A value shows as `String.from` writes it. Its type is the value's runtime type, not the declared
  one, since annotations are not kept at run time.

## Breakpoint expressions

A line breakpoint takes up to three expressions, written in Ant:

| Field          | Example                | Effect                                               |
| -------------- | ---------------------- | ---------------------------------------------------- |
| `condition`    | `p > 100`              | stop only when it is `true`                          |
| `hitCondition` | `>= 3`                 | stop from the third time on; also `N`, `> N`, `% N`  |
| `logMessage`   | `p = {p}, sum = {sum}` | do not stop; write the message to the console, go on |

On the `sum += p` line above, the logpoint `p = {p}, sum = {sum}` writes `p = 108, sum = 0` and then
`p = 72, sum = 108`. Write `{{` and `}}` for literal braces.

A condition that does not parse makes the breakpoint unverified. A condition that panics, names
nothing visible there, or is not a `bool` stops the program and writes the error to the console.

## Panics

Two filters choose which panics stop the program:

| Filter     | Stops at                                                 | Default |
| ---------- | -------------------------------------------------------- | ------- |
| `uncaught` | a panic that no `Panic.catch` on its goroutine will stop | on      |
| `all`      | every panic, also one a `Panic.catch` turns into a value | off     |

The program stops where the panic starts, before any `defer` block runs, so every frame's variables
are still there. `exceptionInfo` gives the message and the panic trace.

## Evaluating and setting values

`evaluate` runs one Ant expression in a stopped frame, such as `prices.length()` or
`discounted | Array.sum()`. It sees what the frame sees at its position: its locals, the bindings a
lambda or `go` block captures, the module's declarations and imports, and the prelude.

- Hover evaluates only a name or a field path, such as `order.address.city`, and never calls a
  function.
- An expression that waits on another goroutine, by a channel or a lock, is abandoned after one
  second, because every other goroutine stays stopped.
- A panic in the expression is the evaluation's error; the program does not stop for it.
- Assigning to a name in an expression is an error, but changing a field or an element changes the
  program's value. `setVariable` changes a `var` binding itself, in `Locals`, `Closure`, or
  `Module`.
- A `let`, a parameter, a loop variable, and `(piped)` are read-only.

No type is checked when you set a value, as `ant run` checks none: a value of the wrong type fails
where the program uses it.

## See also

- [ant run](run.md), which runs a program the way `ant dap` loads it
- [ant test](test.md), for the test files `ant dap` can run
- [ant lsp](lsp.md), the language server the same plugins start
- [Read a panic trace](../howto/debugging/trace.md)
