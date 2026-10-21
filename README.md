# Ant

A Go implementation of [Ant](https://github.com/ant-src/spec): a tree-walking interpreter, a
type checker, and a REPL, usable as a command and as a Go module.

Work in progress. `PLAN.md` describes the design and the order of work.

## Command

```
go install github.com/ant-src/ant/cmd/ant@latest

ant run main.ant [args...]    # run a program; exit 0, 1 on a panic, 2 when it cannot load
ant check main.ant            # type-check it and every module it imports; exit 1 on errors
ant                           # a REPL
```

## Library

```go
prog, err := ant.Load("main.ant")
if err != nil { ... }                         // *ant.LoadError: syntax error, missing module, cycle
err = prog.Run(ctx, ant.WithArgs(args))       // nil, *ant.PanicError, or *ant.ExitError
diags := prog.Check()                         // []ant.Diagnostic
```

The library never exits the process or writes anywhere but the writers it is given.

## License

MIT
