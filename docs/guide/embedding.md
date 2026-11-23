<!-- status: reviewed -->

# Embedding Ant in a Go program

A Go program can run Ant scripts and give them functions of its own. That Go program is the
**host**. The Go package [`github.com/ant-src/ant`](https://pkg.go.dev/github.com/ant-src/ant)
loads, checks, and runs a script, and a **host namespace** is a set of Go functions that the script
calls as it calls the standard library.

This chapter explains how the parts of a host fit, on one small example: a host that reads lines
from its standard input and hands each one to a script. The whole example is in
[`examples/host`](https://github.com/ant-src/ant/tree/master/examples/host), and the code on this
page is copied from it.

## The three parts

A host has three parts:

- A **declaration file**, `app.d.ant`, declares the namespace in Ant: its types and the signatures
  of its functions.
- A **script**, `shout.ant`, calls the namespace.
- The **host**, in Go, implements each declared function and runs the script.

The script decides what happens and when. It starts, calls into the host, and the host calls back
into it. The host does not look for a `main` function in the script.

## The declaration file

The declaration file is the contract between the script and the host:

```ant [app.d.ant]
// App
//
// App is the namespace of the example host, which feeds a script the lines of its standard
// input.

// App.Event is one line of input, and when the host read it.
type App.Event = { line: string, at: Time.Instant }

// App.run calls handle with each line of standard input, until the input ends or handle
// returns false.
func App.run(handle: func(event: App.Event): bool): void

// App.since returns how long ago at was, by the host's clock.
func App.since(at: Time.Instant): Time.Duration
```

Every declaration names its owner, `App`, and a function has a signature and no body. The comment
above each declaration is its documentation, which the editor shows on hover. The chapter on
[modules](modules.md#declaration-files) describes the format.

The script calls `App` as it calls `File` or `Time`:

<!-- test: cmd ant check shout.ant -->

```ant [shout.ant]
var count = 0

func handle(event: App.Event): bool {
  if event.line == "quit" {
    return false
  }
  count += 1
  println("${count}: ${event.line.toUpper()}")
  Time.compareDuration(App.since(event.at), 1m) < 0
}

App.run(handle)
println("${count} lines")
```

The declaration file belongs in the directory of the script. `ant check`, your editor, and
`ant code` read it from there, so the script type-checks, completes, and jumps to a declaration with
no host running. `ant check shout.ant` passes here, although nothing implements `App` yet.

## Loading, checking, and running

The host needs two packages, which `go get github.com/ant-src/ant` adds to its module:

```go
import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/value"
)
```

It embeds the script and the declaration file, and loads the script from them:

```go
// The script, and the declaration file that holds the signatures of App.
//
//go:embed shout.ant app.d.ant
var files embed.FS

func main() {
	os.Exit(run(files, "shout.ant", os.Stdin, os.Stdout, os.Stderr))
}

// run loads the script at path in fsys and runs it, and returns the exit status.
func run(fsys fs.FS, path string, stdin io.Reader, stdout, stderr io.Writer) int {
	prog, err := ant.Load(path,
		ant.WithFS(fsys),
		ant.WithNamespace(app(stdin, stderr)),
		ant.WithoutNamespaces("File", "Http", "Net", "Os"))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return ant.ExitStatus(err)
	}
	if errs := ant.Errors(prog.Check()); len(errs) > 0 {
		for _, d := range errs {
			fmt.Fprintln(stderr, d)
		}
		return 1
	}
	err = prog.Run(context.Background(), ant.WithStdout(stdout), ant.WithStderr(stderr))
	// The run has reported a panic and a returned Err itself. A program it refused to start
	// it has not.
	var refused *ant.LoadError
	if errors.As(err, &refused) {
		fmt.Fprintln(stderr, err)
	}
	return ant.ExitStatus(err)
}
```

**Loading.** `ant.Load` parses the script and every module it imports. With `WithFS` it reads them
from the embedded files. Without it, `Load` reads the path from disk, which is how a host runs
scripts its users write.

**Namespaces.** `WithNamespace` adds `App`. `WithoutNamespaces` removes namespaces of the standard
library: here the ones that reach files, the network, and the process. Removing `Os` removes `Cmd`
with it. A host that keeps `Os` keeps `Cmd`, which runs any command, so removing `File` alone
protects nothing. A script that names a removed namespace gets a compile error from `Check`,
`File is not defined`, and a panic at the use when it runs unchecked.

**Checking.** `Load` does not type-check, and the interpreter erases types, so a script with a
compile error runs, and fails late or not at all. `prog.Check` type-checks the program. A host that
runs scripts written by others calls it first and stops on an error, as `run` does.

**Running.** `prog.Run` runs the script and returns when it ends. The package writes only to the
writers you give it and never exits the process; `WithStdin`, `WithArgs`, and `WithEnv` give the
script its input, its arguments, and its environment. `ant.ExitStatus` turns the error into the exit
status the `ant` command uses:

| Status | The error from `Load` or `Run`                                                |
| ------ | ----------------------------------------------------------------------------- |
| `0`    | none: the script ended                                                        |
| `1`    | the script panicked, or returned an `Err`; or the context of the run ended    |
| `2`    | the program could not be loaded, or `Run` refused to start it                 |
| `130`  | the run was interrupted                                                       |
| other  | the status the script gave `Os.exit`, when the host leaves `Os` in the script |

The run writes a panic and a returned `Err` to the script's standard error itself. It does not write
the error of a program it refused to start, so `run` prints that one. `Run` refuses a program that
reads a name which only a declaration file declares, as a script written for another host does.

## Host functions

A namespace is its name, its signatures, and one Go function for each declared function:

```go
// app is the App namespace: its signatures, and a Go function for each.
func app(stdin io.Reader, stderr io.Writer) ant.Namespace {
	signatures, _ := files.ReadFile("app.d.ant")
	return ant.Namespace{
		Name:       "App",
		Signatures: string(signatures),
		Funcs: map[string]ant.HostFunc{
			"run": func(rt ant.Runtime, args []value.Value) (value.Value, error) {
				return value.Void, feed(rt, args[0], stdin, stderr)
			},
			"since": func(rt ant.Runtime, args []value.Value) (value.Value, error) {
				at, _ := value.GoTime(args[0])
				return value.NewDuration(time.Since(at)), nil
			},
		},
	}
}
```

The signatures are the text of the declaration file, so the host and the tools read the same
declarations. The host's own signatures are the ones its scripts are checked against. For a script
on disk, a declaration file beside it is then not used for `App`, so a stale copy lets `ant check`
pass a script the host rejects. The checker still reports a syntax error in that file.

**Arguments.** A host function receives its arguments in the order the signature lists them, each in
the Go representation of its Ant type:

| Ant                               | Go                                                         |
| --------------------------------- | ---------------------------------------------------------- |
| `int`                             | `int64`                                                    |
| `float64`, `string`, `bool`       | `float64`, `string`, `bool`                                |
| `null`, `void`                    | `value.Null`, `value.Void`                                 |
| `Array[T]`, `Map[K, V]`, `Set[T]` | `*value.Array`, `*value.Map`, `*value.Set`                 |
| a record                          | `*value.Record`                                            |
| a function                        | an opaque value, for `rt.Call`                             |
| `Time.Instant`, `Time.Duration`   | an opaque value, for `value.GoTime` and `value.GoDuration` |

The [package documentation](https://pkg.go.dev/github.com/ant-src/ant#hdr-Values) has the full
table. `prog.Check` has checked every call against the signatures, so a host function asserts the
types of its arguments without testing them.

**Results.** A function returns a value in the same representations, and an `int` result is an
`int64`, never a Go `int`. A union needs no wrapper: a function declared to return
`App.Event | null` returns the record or `value.Null`.

**Failures.** A failure the script handles is a `Result` in the signature. The function returns
`value.Ok(v)`, or `value.Err(value.ErrorValue("message"))` for a `Result[T, Error]`: `value.Err`
wraps what it is given as it is, so a bare string there is an `Err` whose `error` is not an `Error`.
A Go `error` returned from a host function becomes an Ant panic with the error's message, which is
for bugs. A Go panic of the host's own is not converted: it ends the process, as in any Go program.

**State.** A script may start goroutines, and they run in parallel, so two of them may call a host
function at the same time. A host function that reads or writes Go state shared between calls locks
it.

## Records and tags

`App.run` hands the script an `App.Event`. A record is its fields and a **tag**, which names its
type, and the tag must be the one the loaded program holds for that type. `rt.Tag` returns it:

```go
// feed calls handle, an Ant function, with an App.Event for each line of stdin.
func feed(rt ant.Runtime, handle value.Value, stdin io.Reader, stderr io.Writer) error {
	event, _ := rt.Tag("App.Event")
	loop := &Loop{Log: stderr}
	var unwound any
	var failed error
	loop.Handle = func(line string) {
		defer func() {
			if r := recover(); r != nil {
				unwound = r
				loop.Stop()
			}
		}()
		at, err := value.NewInstant(time.Now())
		if err != nil {
			failed = err
			loop.Stop()
			return
		}
		more := rt.Call(handle, value.NewRecord(event, []value.Value{line, at}))
		if !more.(bool) {
			loop.Stop()
		}
	}
	err := loop.Run(stdin)
	if unwound != nil {
		panic(unwound)
	}
	return errors.Join(failed, err)
}
```

`rt.Tag` takes the name of a record type that a namespace declares, and its second result is false
for any other name. `value.NewRecord` takes the tag and the fields in the order the type declares
them.

`value.NewTag` makes a new tag, and a new tag is a new type. A record built with one prints like an
`App.Event` and reads like one, and fails every `is App.Event` test with no error, because tags
compare by identity.

## Time values

`value.NewInstant` makes a `Time.Instant` from a Go `time.Time`, and reports an error for a time
outside the range an instant holds. `value.NewDuration` makes a `Time.Duration` from a
`time.Duration`. `value.GoTime` and `value.GoDuration` convert back. Their second result is false
for a value of another type, and `value.GoDuration` is also false for a duration with years, months,
or weeks, which has no fixed length.

A duration from `value.NewDuration` is balanced: 90 seconds arrive as one minute and thirty seconds.
`==` compares durations field by field, so that value equals `1m30s` and not `90s`. The script
compares with `Time.compareDuration`, which compares lengths.

## Calls back into the script

`handle` arrives as an argument like any other. `rt.Call` calls it with positional arguments and
returns its result, in the same Go representations: `handle` returns a `bool`.

Three things end a call early: a panic in the script, `Os.exit`, and the end of a run that was
cancelled or interrupted. Each unwinds through `rt.Call` and through the host function as a Go
panic. A host function lets it pass, and cleans up with `defer`. The run then ends as it would
without the host function, unless the script catches the panic with `Panic.catch` and goes on: the
host's state must be sound after an unwind.

Code that recovers panics must not stand in the way. In the example, `Loop` stands in for a web
server or a terminal UI framework: it owns the loop, calls a handler, and recovers a panic of the
handler, so that one bad event does not end the host. Left alone, it would swallow the script's
panic, and the script would go on after its own failure. So `feed` recovers first, inside the
handler, keeps the recovered value, and stops the loop. When the loop has returned, `feed` panics
with the same value. A script that panics in `handle` then ends the host as it ends `ant run`: with
the panic and its stack trace on standard error, and status 1.

The same applies when the handler runs on another Go goroutine, where an unrecovered panic ends the
process: recover there, carry the value back, and panic with it in the host function.

`rt.Call` takes one caller at a time, and only while the host function that received `rt` has not
returned. Another goroutine may call it while the host function waits for that goroutine. Two calls
at once on one `Runtime` are a data race, so a host with concurrent events sends them to one place
and calls the script from there.

## The example, run

```sh
printf 'hello\nant\nquit\nnot read\n' | go run ./examples/host
```

```text
1: HELLO
2: ANT
2 lines
```

## Scripts outside the host

The declaration file is all the tools need. A script that calls `App` passes `ant check` with
`app.d.ant` beside it, and your editor completes and documents `App.run`. `ant run` refuses such a
script with a load error and status 2, since the file declares `App` and does not implement it: the
script runs only in the host. A host ships `app.d.ant` for the people who write scripts for it.

## See also

- [The Go API](https://pkg.go.dev/github.com/ant-src/ant): every option, `Program.Start` and
  `Instance.Func` for calling the functions of an Ant library from Go, and `Session` for evaluating
  inputs one at a time
- [Modules](modules.md#declaration-files): the declaration file format
- [`ant check`](../tools/check.md) and [`ant run`](../tools/run.md): what each does with a
  declaration file beside a script
