<!-- status: draft -->

# Embedding Ant in a Go program

A Go program can run Ant scripts and give them functions of its own. The program is the **host**.
The Go package [`github.com/ant-src/ant`](https://pkg.go.dev/github.com/ant-src/ant) loads, checks,
and runs a script, and a **host namespace** is a set of Go functions the script calls as it calls
the standard library.

This chapter builds one small host and explains each part. The host reads lines from its standard
input and hands each one to the script. The whole program is in
[`examples/host`](https://github.com/ant-src/ant/tree/master/examples/host), and the code on this
page is copied from it.

## The three parts

A host has three parts:

- A **declaration file**, `app.d.ant`, which declares the namespace in Ant: its types and the
  signatures of its functions.
- A **script**, `shout.ant`, which calls the namespace.
- The **Go program**, which implements each declared function and runs the script.

The script owns the program's flow. It starts, calls into the host, and the host calls back into it.
The host does not look for a `main` function in the script.

## Declare the namespace

The declaration file is the contract between the script and the host. Write it first:

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

Keep the declaration file in the directory of the script. `ant check`, your editor, and `ant code`
read it from there, so the script type-checks, completes, and jumps to a declaration with no host
running. `ant check shout.ant` passes here, although nothing implements `App` yet.

## Load, check, and run

The host embeds both files and loads the script from them:

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
		ant.WithoutNamespaces("File", "Http", "Os"))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if errs := ant.Errors(prog.Check()); len(errs) > 0 {
		for _, d := range errs {
			fmt.Fprintln(stderr, d)
		}
		return 1
	}
	err = prog.Run(context.Background(), ant.WithStdout(stdout), ant.WithStderr(stderr))
	return ant.ExitStatus(err)
}
```

- `ant.Load` parses the script and every module it imports. `WithFS` makes it read from the embedded
  files; without it, `Load` reads the path from disk, which is what a host does for scripts its
  users write.
- `WithNamespace` adds `App`. `WithoutNamespaces` removes parts of the standard library, so this
  script cannot touch files, the network, or the process.
- `Load` does not type-check. `prog.Check` does, and a host that runs scripts written by others
  calls it first: the interpreter erases types, so a script with a compile error would run until it
  panics.
- `prog.Run` runs the script and returns when it ends. The library writes only to the writers you
  give it and never exits the process. `ant.ExitStatus` turns the error into the status the `ant`
  command uses: 0, 1 for a panic or an `Err` the script returned, or the status the script gave
  `Os.exit`.

## Implement the functions

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
declarations. When the host provides `App` itself, its signatures are the ones the program uses, and
the copy of the file beside the script is not read for that name.

A host function receives its arguments in the order the signature lists them, each in the Go
representation of its Ant type: an `int` is an `int64`, a `string` is a `string`, an `Array` is a
`*value.Array`, a record is a `*value.Record`. The
[package documentation](https://pkg.go.dev/github.com/ant-src/ant#hdr-Values) has the full table.
`prog.Check` has checked every call against the signatures, so a function may assert the types of
its arguments.

A function returns a value in the same representations, `value.Void` for `void`. For a failure the
script should handle, declare a `Result` and return `value.Ok` or `value.Err`. A Go `error` returned
from a host function becomes an Ant panic with the error's message, which is for bugs.

## Build records with the program's tags

`App.run` hands the script an `App.Event`. A record is its fields and a **tag**, which names its
type, and the tag must be the one the program holds for that type. `rt.Tag` returns it:

```go
// feed calls handle, an Ant function, with an App.Event for each line of stdin.
func feed(rt ant.Runtime, handle value.Value, stdin io.Reader, stderr io.Writer) error {
	event, _ := rt.Tag("App.Event")
	loop := &Loop{Log: stderr}
	var unwound any
	loop.Handle = func(line string) {
		defer func() {
			if r := recover(); r != nil {
				unwound = r
				loop.Stop()
			}
		}()
		at, err := value.NewInstant(time.Now())
		if err != nil {
			panic(err)
		}
		more := rt.Call(handle, value.NewRecord(event, []value.Value{line, at}))
		if more == false {
			loop.Stop()
		}
	}
	err := loop.Run(stdin)
	if unwound != nil {
		panic(unwound)
	}
	return err
}
```

`value.NewRecord` takes the tag and the fields in the order the type declares them. Do not make a
tag of your own with `value.NewTag` for a declared type: the record would print like an `App.Event`
and fail every `is App.Event` test, with no error, because tags compare by identity.

`value.NewInstant` and `value.NewDuration` make a `Time.Instant` and a `Time.Duration` from Go's
`time.Time` and `time.Duration`, and `value.GoTime` and `value.GoDuration` convert back, as
`App.since` does.

## Call back into the script

`handle` arrives as an argument like any other. It is an opaque value, and `rt.Call` calls it with
positional arguments and returns its result.

Three things end a call early: a panic in the script, `Os.exit`, and the cancellation of the run.
Each unwinds through `rt.Call` and through your host function as a Go panic. Let it pass: the run
then ends as it would have without the host function. Use `defer` for cleanup.

This matters when a library stands between your host function and the callback. In the example,
`Loop` stands in for such a library, a web server or a terminal UI framework: it owns the loop,
calls your handler, and recovers a panic in a handler so that one bad event does not end the
program. Left alone, it would swallow the script's panic, and the script would go on after its own
failure. So `feed` recovers first, inside the handler, keeps what it recovered, stops the loop, and
panics with the same value once the loop has returned. Run with a script that panics in `handle`,
the host prints the script's panic and its stack trace and exits with status 1, as `ant run` does.

`rt.Call` takes one caller at a time, and only while the host function that received `rt` has not
returned. A library may call your handler from another goroutine if the host function waits for it.
Two calls at once on one `Runtime` are a data race, so a host with concurrent events sends them to
one place and calls the script from there.

## Run it

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
script with a load error, since the file declares `App` and does not implement it: the script runs
only in the host. Ship `app.d.ant` with the host, for the people who write scripts for it.

## See also

- [The Go API](https://pkg.go.dev/github.com/ant-src/ant): every option, `Program.Start` and
  `Instance.Func` for calling a library's functions from Go, and `Session` for evaluating inputs one
  at a time
- [Modules](modules.md#declaration-files): the declaration file format
- [`ant check`](../tools/check.md) and [`ant run`](../tools/run.md): what each does with a
  declaration file beside a script
