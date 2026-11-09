<!-- status: reviewed -->

# Where next

You know the language now. Most of what you write from here calls the standard library: a set of
namespaces such as `Array`, `String`, `Time`, and `Json`, always there, with no import.

```ant
type Visit = { user: string, day: Time.PlainDate }

let log = `ann,2026-11-02
bob,2026-11-02
ann,2026-11-05
cid,2026-11-07`

// Csv splits the text into rows, Time reads the dates.
let rows = Csv.parse(log).unwrap()
let visits = rows.map({ Visit{ user: x[0], day: Time.parsePlainDate(x[1]).unwrap() } })

// Array groups the visits, Map counts each group, Json writes the counts.
let byUser = visits.groupBy({ x.user })
let counts = byUser.map({ x.value.length() })
println(Json.stringify(counts, indent: "  "))

// Time measures the span from the first visit to the last.
let first = visits[0].day
let last = visits[visits.length() - 1].day
println("${first} to ${last}: ${first.until(last)}")

// String joins and formats the names.
println(byUser.keys().sort() | String.join(", ") | String.toUpper())
```

```text
{
  "ann": 2,
  "bob": 1,
  "cid": 1
}
2026-11-02 to 2026-11-07: P5D
ANN, BOB, CID
```

Six namespaces, and every call has the shape you met in the tour: the subject goes first, so `|`
reaches each function, and `.` reaches the functions of the subject's own type.
`visits.groupBy(...)` is [`Array.groupBy`](../lib/Array/groupBy.md), `byUser.map(...)` is
[`Map.map`](../lib/Map/map.md), and `first.until(last)` is [`Time.until`](../lib/Time/until.md). An
array's `.` reaches `Array`, so `String.join` takes the names through `|`. The other calls are
[`Csv.parse`](../lib/Csv/parse.md), [`Time.parsePlainDate`](../lib/Time/parsePlainDate.md),
[`Json.stringify`](../lib/Json/stringify.md), and [`String.join`](../lib/String/join.md).

## Look a name up

`ant doc` prints the signature and doc comment of any name in the standard library, the same text
your editor shows on hover:

```sh
ant doc Array.groupBy
```

```text
func Array.groupBy[T, K](xs: Array[T], key: func(T): K): Map[K, Array[T]]

    Array.groupBy returns a map from each key to the elements of xs with that key. Keys are in
    first-seen order and each group keeps the order of xs.

        Array.groupBy([1, 2, 3, 4, 5], { x % 2 })   // Map{ 1: [1, 3, 5], 0: [2, 4] }
```

A namespace alone, such as `ant doc Csv`, prints its doc and the signature of everything in it. Give
`ant doc` a file, as in `ant doc geo.ant`, and it documents your own module the same way.

## Try it

- Run `ant doc Map` and find a function that keeps only the users with more than one visit.
- Print the day of the week of each visit with [`Time.dayOfWeek`](../lib/Time/dayOfWeek.md).
- Write the counts back out as CSV with [`Csv.stringify`](../lib/Csv/stringify.md).

## Where to go

- [Language guide](../guide/types.md): one chapter per area, such as [errors](../guide/errors.md),
  [concurrency](../guide/concurrency.md), and [modules](../guide/modules.md), with when to pick
  what.
- [How-to guides](../howto/files/read-write.md): one task per page, such as
  [decoding JSON](../howto/data/json-decode.md), [running a program](../howto/processes/run.md), or
  [serving HTTP](../howto/network/http-server.md).
- [Standard library](../lib/index.md): a page for every function and type, and topic guides such as
  [text](../lib/guides/text.md), [files](../lib/guides/files.md), and [time](../lib/guides/time.md).
- [Language reference](../language/index.md): every keyword, operator, literal, and built-in type.
- Coming from [JavaScript](../coming-from/javascript.md), [Go](../coming-from/go.md), or
  [Python](../coming-from/python.md): the equivalents, and the differences that bite.
- [Design notes](../design/errors-are-values.md): why Ant is the way it is, one decision per page.

The tools come in the same binary as `ant run`: [`ant check`](../tools/check.md),
[`ant fmt`](../tools/fmt.md), [`ant test`](../tools/test.md), and [`ant doc`](../tools/doc.md). The
plugins for [VS Code](/vscode), [IntelliJ IDEA](/intellij), and [Neovim](/nvim) connect your editor
to the language server, `ant lsp`, and the debugger, `ant dap`.
