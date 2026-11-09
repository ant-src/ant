<!-- status: reviewed -->

# Tests

A test is a `test` block in a file whose name ends in `_test.ant`. Inside it, `assert` checks what
your code does, and `ant test` runs every test and reports the ones that fail.

Here is a small library, `words.ant`, that counts words:

```ant [words.ant]
// Counts how often each word appears in text, ignoring case.
export func count(text: string): Map[string, int] {
  let counts: Map[string, int] = Map{}
  for w in split(text) {
    counts[w] = (counts[w] ?? 0) + 1
  }
  counts
}

func split(text: string): Array[string] {
  text.toLower().fields()
}
```

And here are its tests, in `words_test.ant` next to it:

<!-- test: cmd ant test . -->

```ant [words_test.ant]
import "./words.ant" { count, split }

test "count adds up repeated words" {
  let counts = count("the cat and the hat")
  assert(counts["the"] == 2)
  assert(counts["cat"] == 1)
  assert(counts["dog"] == null)
}

test "count ignores case" {
  assert(count("Go go GO")["go"] == 3)
}

type Case = { name: string, text: string, want: Array[string] }

test "split" {
  let cases = [
    Case{ name: "empty", text: "", want: [] },
    Case{ name: "one word", text: "one", want: ["one"] },
    Case{ name: "extra spaces", text: "  two   words ", want: ["two", "words"] },
  ]
  for c in cases {
    test "${c.name}" {
      assert(split(c.text) == c.want)
    }
  }
}
```

Run `ant test` in the directory. It finds `words_test.ant`, runs its tests one after another, and
prints one line when they all pass:

```sh
ant test
```

```text
ok    5 passed (0.00s)
```

`split` is not exported, yet the test file imports it. A test file may import the private names of
the module it tests, so you can test a helper without exporting it. Nothing else can import a test
file, so the private names stay private in your program.

The test `"split"` is a table: a `test` statement inside a loop runs one subtest per row, named
after the row. Each subtest passes or fails on its own, and its full name joins the names with `/`.
That is why the run counts five tests: the first two, and the three rows of `"split"`.

Now break a row. Change the last `want` to `["two words"]` and run `ant test` again:

```text
--- FAIL: words_test.ant: split/extra spaces (0.00s)
    assertion failed: split(c.text) == c.want (["two", "words"] == ["two words"])
      c.text = "  two   words "
      at test "extra spaces" (words_test.ant:24:13)
FAIL  1 failed, 4 passed (0.00s)
```

The report names the failing subtest, `split/extra spaces`. A failed `assert` quotes the condition,
shows both sides of the comparison, and lists the values it was computed from, so you rarely need a
message of your own. The other rows still run, and `ant test` exits with status 1. To run only the
failing row while you fix it, pass `-run`:

```sh
ant test -run 'split/extra'
```

`ant run` refuses a test file: tests run only under `ant test`. Like `ant run`, `ant test` does not
type-check; run `ant check words_test.ant` to check the tests and the code they import.

Use `try` in a test body: an `Err` it returns fails the test. To check that code panics, wrap it in
[`Panic.catch`](../lib/Panic/catch.md) and assert that the result is an `Err`. A test body can also
skip itself, run in parallel with others, or get a temporary directory, with the functions of
[Test](../lib/Test/index.md).

The [Testing](../guide/testing.md) chapter shows how the parts fit together. The reference has the
[`test`](../language/keywords/test.md) keyword, [`assert`](../lib/prelude/assert.md), and the
[`ant test`](../tools/test.md) command with all its flags.

Next: [HTML](21-html.md).
