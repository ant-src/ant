// Command host is the example of docs/guide/embedding.md: a Go program that runs an Ant script
// and gives it a namespace of its own, App.
package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"ant-lang.org"
	"ant-lang.org/value"
)

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
