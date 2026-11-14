<!-- status: reviewed -->

# Standard library

The standard library is a set of namespaces in the prelude, so every module sees them without an
`import`. Each namespace holds functions, and some also hold types and constants. A function takes
its subject as the first argument, so you can call it through its namespace, through `.` on the
value, or at the end of a pipe:

```ant
let words = ["ant", "bee", "cat"]
Array.map(words, { x.toUpper() })  // ["ANT", "BEE", "CAT"]
words.map({ x.toUpper() })         // ["ANT", "BEE", "CAT"]
words | Array.map({ x.toUpper() }) // ["ANT", "BEE", "CAT"]
```

A function marked `mutates` changes its first argument; every other function leaves its arguments
untouched and returns a new value. In the namespaces that talk to the world outside the program, an
operation that can fail returns a `Result` with an `Error` rather than panicking on bad input.

Some namespaces are types as well, such as `Result`, `Bytes`, and `Mutex`: the name is a type in an
annotation and a namespace in an expression. Others, such as `Os`, `Time`, and `Reflect`, are only
namespaces, reached through `.`. A type a namespace declares is written through it, as in
`Http.Response` or `Time.Instant`, and a value of that type reaches the namespace's functions
through `.`, as in `h.writeLine(line)`.

## Core

- [Prelude](prelude/index.md): the names that stand on their own: `print`, `println`, `eprint`,
  `eprintln`, `panic`, `assert`, `len`, the conversions, and the shared types such as `Entry`,
  `Seq`, and `Number`.
- [Array](Array/index.md): the functions of arrays, ordered and growable.
- [Map](Map/index.md): the functions of maps, hash maps that keep insertion order.
- [Set](Set/index.md): the functions of sets, hash sets that keep insertion order.
- [Iter](Iter/index.md): lazy sequences, which produce their elements on demand.
- [String](String/index.md): the functions of strings, immutable Unicode text.
- [Regex](Regex/index.md): regular expressions, written as a literal or compiled from data.
- [Bytes](Bytes/index.md): mutable, growable sequences of bytes, for binary data.
- [Html](Html/index.md): the value of an HTML expression, an immutable list of nodes.
- [Math](Math/index.md): numeric functions and constants.
- [Bits](Bits/index.md): bitwise operations and wrapping arithmetic over the integer types.
- [Random](Random/index.md): seeded generators of pseudo-random numbers.
- [Time](Time/index.md): instants, dates, clock times, zones, and durations.
- [Record](Record/index.md): the functions of record values.

## Errors

- [Result](Result/index.md): the outcome of an operation that can fail, `Ok` with a value or `Err`
  with an error.
- [Error](Error/index.md): the error the standard library returns in an `Err`, and wrapping it with
  context.
- [Panic](Panic/index.md): catching a panic as a value, and raising it again.

## Concurrency

- [Channel](Channel/index.md): carries values from goroutine to goroutine.
- [Context](Context/index.md): bounds a wait with a deadline or ends it with a cancellation.
- [Parallel](Parallel/index.md): runs work on goroutines and waits for all of it.
- [WaitGroup](WaitGroup/index.md): waits until a group of goroutines has finished.
- [Mutex](Mutex/index.md): a mutual-exclusion lock.
- [RwMutex](RwMutex/index.md): a lock held by many readers or one writer.
- [Once](Once/index.md): runs a function exactly once, however many goroutines ask.
- [Cond](Cond/index.md): a condition variable that goroutines wait on until another signals.
- [Atomic](Atomic/index.md): a cell holding one value, read and written atomically.

## Files and the system

- [File](File/index.md): the file system, open files, and the standard streams.
- [Path](Path/index.md): builds and takes apart paths as text, without touching the disk.
- [Os](Os/index.md): the process: its arguments, environment, standard streams, exit, and
  interrupts.
- [Cmd](Cmd/index.md): runs other programs, alone or in a pipeline.

## Network

- [Http](Http/index.md): an HTTP client and server.
- [Net](Net/index.md): TCP listeners and connections.
- [Url](Url/index.md): takes URLs and query strings apart and puts them back together.

## Data formats

- [Json](Json/index.md): converts between JSON text and values.
- [Csv](Csv/index.md): converts between CSV text and rows of fields.
- [Encoding](Encoding/index.md): converts bytes to and from base64 and hex text.
- [Crypto](Crypto/index.md): cryptographic hashes, HMACs, a constant-time comparison, and secure
  randomness.

## Tooling

- [Test](Test/index.md): what a test body needs beyond `assert`, for `ant test`.
- [Reflect](Reflect/index.md): types as values, taken apart and used to read and build values.

## Topic guides

Each guide shows how the namespaces of one area work together and when to pick which.

- [Collections](guides/collections.md): arrays, maps, sets, and sequences.
- [Working with text](guides/text.md): strings, regular expressions, and bytes.
- [Numbers and math](guides/numbers.md): the number types, conversions, and `Math` and `Bits`.
- [Dates and times](guides/time.md): instants, wall-clock times, zones, and durations.
- [Errors](guides/errors.md): `Result` for expected failure, panics for bugs.
- [Concurrency](guides/concurrency.md): goroutines, channels, contexts, and locks.
- [Files and paths](guides/files.md): reading, writing, and walking the file system.
- [Processes and the OS](guides/system.md): the program's own process and the programs it runs.
- [Network](guides/network.md): HTTP, TCP, and URLs.
- [Data formats](guides/data-formats.md): JSON, CSV, encodings, and hashes.
- [Testing](guides/testing.md): test files, `assert`, and `ant test`.
- [Reflection](guides/reflection.md): typed decoding and encoding with `Reflect`.
