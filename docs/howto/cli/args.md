<!-- status: reviewed -->

# Read command-line arguments

This script takes a count and some words, and prints the words that many times. Save it as
`repeat.ant` and run it with `ant run repeat.ant 3 hello world`:

<!-- test: cmd ant run repeat.ant 3 hello world -->

```ant [repeat.ant]
let args = Os.args()
if args.length() < 2 {
  eprintln("usage: ${Path.base(Os.program())} <count> <word>...")
  Os.exit(2)
}
let count = args[0].toInt()
if count == null || count < 1 {
  eprintln("not a positive count: ${args[0]}")
  Os.exit(2)
}
let line = args.slice(1, args.length()) | String.join(" ")
for _ in Iter.range(0, count) {
  println(line)
}
```

```text
hello world
hello world
hello world
```

1. Call [Os.args](../../lib/Os/args.md). It returns the arguments after the script path as an
   `Array[string]`, so `args[0]` is the first argument, here `"3"`.
2. Check the number of arguments before you index the array: an index past the end panics. When
   arguments are missing, write a usage line to standard error with
   [eprintln](../../lib/prelude/eprintln.md) and end with status 2 through
   [Os.exit](../../lib/Os/exit.md). [Os.program](../../lib/Os/program.md) gives the script path as
   it was run, and [Path.base](../../lib/Path/base.md) keeps its last element.
3. Convert a numeric argument with [String.toInt](../../lib/String/toInt.md). It returns `null` for
   text that is not an integer, so reject that case with the same exit status.
4. Take the remaining arguments with [Array.slice](../../lib/Array/slice.md) and join them with
   [String.join](../../lib/String/join.md). [Iter.range](../../lib/Iter/range.md) repeats the line.

Run it without arguments to see the usage line:

```sh
ant run repeat.ant
```

```text
usage: repeat.ant <count> <word>...
```

The script exits with status 2. A count of `0` or `x` prints `not a positive count: 0` (or `x`) to
standard error and exits with status 2 as well.

## Make an argument optional

Use [Array.first](../../lib/Array/first.md), which returns `null` for an empty array, and give the
default after `??`:

<!-- test: args Ann -->

```ant
let name = Os.args().first() ?? "world"
println("Hello, ${name}")
```

```text
Hello, Ann
```

Without arguments, the script prints `Hello, world`.

## Check for a flag

Test for the flag with [Array.contains](../../lib/Array/contains.md), and drop it from the rest of
the arguments with [Array.filter](../../lib/Array/filter.md):

<!-- test: args -v a.txt b.txt -->

```ant
let args = Os.args()
let verbose = args.contains("-v")
let files = args.filter({ x != "-v" })
println("${verbose} ${files}")
```

```text
true ["a.txt", "b.txt"]
```

## See also

[Os.args](../../lib/Os/args.md), [Exit with a status](exit-status.md),
[Run a script as a command](shebang.md),
[Read environment variables and standard directories](env.md)
