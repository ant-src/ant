<!-- status: reviewed -->

# Playground

The [playground](/playground) runs Ant in your browser, with nothing to install. Type a program in
`main.ant` and press Run, or Ctrl+Enter (⌘+Enter on a Mac). The output panel shows what the program
prints, then its exit status and how long it ran.

```ant
let names = ["Ada", "Grace", "Barbara"]
for name in names {
  println("Hello, ${name}")
}
```

```text
Hello, Ada
Hello, Grace
Hello, Barbara
```

The program runs as a script, as `ant run main.ant` would run it: the top-level code runs from top
to bottom, and the exit status is 1 when it panics or returns an `Err`, and 2 when it cannot load,
such as after a syntax error or an import of a missing module. While a program runs, the Run button
becomes Stop, which ends it, as Esc does outside the editor. A program that prints more than a
megabyte is stopped too.

## Read input

The Input tab is the program's standard input. Type lines there and read them with `Os.stdin()`:

<!-- test: stdin
3
4
5
-->

```ant
let input = Os.stdin()
var total = 0
while true {
  let line = try input.readLine() ?? break
  total += line.trim().toInt() ?? 0
}
println("total: ${total}")
```

```text
total: 12
```

At the end of the input, `readLine` returns `null`. With the Input tab empty, the first read returns
`null`.

## Split a program into modules

The `+` button next to the tabs adds a module. Import it from `main.ant` by its path relative to
`main.ant`, such as `./greet.ant`:

```ant [greet.ant]
export func greet(name: string): string { "Hello, ${name}" }
```

```ant
import "./greet.ant" { greet }

println(greet("Ant"))
```

```text
Hello, Ant
```

The program always starts at `main.ant`, so that tab has no rename or delete. Double-click any other
tab to rename it, and press its × to delete it.

## What a program can reach

The playground runs the program inside the browser page, so it reaches less than `ant run` does:

- **Files.** The modules in the tabs are the program's source, not files it can open. `File` calls
  that reach the file system, such as `File.read` and `File.open`, return an `Err`, and
  `File.exists` returns `false`. The standard streams work: `Os.stdin()` reads the Input tab, and
  `print` writes to the output panel.
- **Processes.** `Cmd` cannot start a program; it returns an `Err`.
- **The network.** `Http` requests go through the browser, which lets a page reach only the servers
  that allow requests from other sites. `Net` connections and listeners, and an `Http` server,
  return an `Err`.
- **Arguments and environment.** `Os.args()` is empty, and `Os.env` returns `null` for every name.

Handle the `Result` as you would anywhere: `try` ends the script with the error and exit status 1,
and an `is Err` test lets the program go on:

```ant
if File.read("notes.txt") is Err {
  println("no notes.txt to read")
}
```

```text
no notes.txt to read
```

Write a program that needs files, processes, or its own server in a file on your machine, and run it
with `ant run` (see [Install](install.md)).

## Edit

The editor checks the program as you type, as `ant check` does, and underlines each compile error.
Run does not wait for the checker: like `ant run`, it runs a program that has compile errors, so fix
the underlines first.

- Hover a name to see its type and its documentation.
- F12 jumps to a definition, in this module or another.
- Hold Ctrl+Alt to show the types `let` and `var` infer.
- Format, or Ctrl+S (⌘+S), formats the module as `ant fmt` does.
- A position in the output, such as `main.ant:3:5` in a panic's trace, links to that line.

The Examples menu opens a sample program, from a first `println` to a small calculator that reads
its input.

## Share

Share copies a link to the program: every module and the input. The code travels in the link itself,
after the `#`, so whoever opens it sees your code without a server storing it. Editing a shared
program makes it your own copy.

The browser keeps the code you write between visits, so a reload shows the program you left. Image
opens your code as a picture, to post where a link does not fit.
