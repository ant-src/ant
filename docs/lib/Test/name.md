<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Test.name

```ant signature
func Test.name(): string
```

Test.name returns the current test's full name, its names from the top-level test down joined with
/, as "parse/0x10".

<!-- end generated -->

## Details

- The current test is the test whose body runs the call, or the test that owns the goroutine that
  makes it. A helper function called from a test gets that test's name without being given it.
- When two subtests of one test get the same name, the second is named with ` #2`, the third with
  ` #3`, and `Test.name` returns that name.
- The full name does not include the test file: tests in two files can have the same full name.
- Called outside any test, `Test.name` panics. That includes the initialization of a test file's
  top-level variables, which runs before any test.

## Basic examples

In a top-level test, `Test.name` returns the name the test declares:

<!-- test: cmd ant test basic_test.ant -->

```ant [basic_test.ant]
test "parse" {
  assert(Test.name() == "parse")
}
```

```text
ok    1 passed (0.00s)
```

In a subtest, it returns the full name, with the parent's name first:

<!-- test: cmd ant test sub_test.ant -->

```ant [sub_test.ant]
test "parse" {
  test "0x10" {
    assert(Test.name() == "parse/0x10")
  }
}
```

```text
ok    1 passed (0.00s)
```

## Scope

### Subtests from a table

A subtest's name may interpolate, and `Test.name` returns the name as it was evaluated. Under
`ant test -v`, each test's output appears below the `=== RUN` line that `ant test` prints for it,
after the file name. `ant test` puts the file name before the full name; `Test.name` leaves it out:

<!-- test: cmd ant test -v table_test.ant -->

```ant [table_test.ant]
test "parse" {
  for input in ["1", "0x10"] {
    test "${input}" {
      println(Test.name())
    }
  }
}
```

```text
=== RUN   table_test.ant: parse
=== RUN   table_test.ant: parse/1
    parse/1
--- PASS: table_test.ant: parse/1 (0.00s)
=== RUN   table_test.ant: parse/0x10
    parse/0x10
--- PASS: table_test.ant: parse/0x10 (0.00s)
--- PASS: table_test.ant: parse (0.00s)
ok    2 passed (0.00s)
```

### Nested subtests

Subtests nest to any depth, and the full name holds every level:

<!-- test: cmd ant test nested_test.ant -->

```ant [nested_test.ant]
test "server" {
  test "routes" {
    test "GET /users" {
      assert(Test.name() == "server/routes/GET /users")
    }
  }
  assert(Test.name() == "server")
}
```

```text
ok    1 passed (0.00s)
```

A name may itself hold a `/`, as `GET /users` does here, so the full name cannot always be split
back into its levels.

### Repeated names

When two subtests of one test get the same name, the second is named with ` #2`, and `Test.name`
returns that name:

<!-- test: cmd ant test repeat_test.ant -->

```ant [repeat_test.ant]
test "words" {
  let names: Array[string] = []
  for word in ["a", "b", "a"] {
    test "${word}" {
      names.push(Test.name())
    }
  }
  assert(names == ["words/a", "words/b", "words/a #2"])
}
```

```text
ok    3 passed (0.00s)
```

### Helper functions

A function called from a test reads the name of the test that calls it:

<!-- test: cmd ant test helper_test.ant -->

```ant [helper_test.ant]
func where(): string {
  "in ${Test.name()}"
}

test "first" {
  assert(where() == "in first")
}

test "second" {
  test "inner" {
    assert(where() == "in second/inner")
  }
}
```

```text
ok    2 passed (0.00s)
```

### Goroutines

A goroutine started in a test belongs to it, so `Test.name` there returns the test's name:

<!-- test: cmd ant test go_test.ant -->

```ant [go_test.ant]
test "worker" {
  let names: Channel[string] = Channel.new()
  go { names.send(Test.name()) }
  assert(names.receive() == "worker")
}
```

```text
ok    1 passed (0.00s)
```

## Applications

### Golden files

Keep the expected output of each test in a file named after it. A subtest's name holds `/`, so
replace it to keep the files in one directory:

<!-- test: file golden/greet_hello.txt
Hello, Ann
-->
<!-- test: file golden/greet_bye.txt
Goodbye, Ann
-->
<!-- test: cmd ant test golden_test.ant -->

```ant [golden_test.ant]
func golden(): string {
  let file = Test.name().replace("/", "_")
  File.read("golden/${file}.txt").unwrap().trim()
}

test "greet" {
  test "hello" {
    assert(golden() == "Hello, Ann")
  }
  test "bye" {
    assert(golden() == "Goodbye, Ann")
  }
}
```

```text
ok    2 passed (0.00s)
```

### Labeled log lines

A logging helper can tag each line with the test that wrote it, without each test passing its name:

<!-- test: cmd ant test log_test.ant -->
<!-- test: exit 1 -->

```ant [log_test.ant]
func log(message: string) {
  println("[${Test.name()}] ${message}")
}

test "half" {
  for n in [4, 7] {
    test "${n}" {
      log("halving ${n}")
      assert(n % 2 == 0)
    }
  }
}
```

```text
--- FAIL: log_test.ant: half/7 (0.00s)
    [half/7] halving 7
    assertion failed: n % 2 == 0 (1 == 0)
      n = 7
      at test "7" (log_test.ant:9:13)
FAIL  1 failed, 1 passed (0.00s)
```

## Properties and relations

A subtest's name starts with its parent's name and a `/`:

<!-- test: cmd ant test prefix_test.ant -->

```ant [prefix_test.ant]
test "outer" {
  let parent = Test.name()
  test "inner" {
    assert(Test.name() == "${parent}/inner")
  }
}
```

```text
ok    1 passed (0.00s)
```

## Possible issues

Outside a test there is no current test, and `Test.name` panics:

```ant
println(Test.name()) // panic: Test.name: not in a test
```

A test file's top-level variables are set before any test runs, so `Test.name` panics there too.
Call it inside the test body:

<!-- test: cmd ant test init_test.ant -->

```ant [init_test.ant]
let name = Test.name() // panic: Test.name: not in a test

test "t" {
  assert(name == "t")
}
```

```text
FAIL  1 failed, 0 passed (0.00s)
```

## See also

- [test](../../language/keywords/test.md): tests, subtests, and their names
- [Test.skip](skip.md): end the current test as skipped
- [Test.tempDir](tempDir.md): a directory for the current test
- [Test.context](context.md): a context that ends with the current test
- [ant test](../../tools/test.md): run tests by name with `-run`
