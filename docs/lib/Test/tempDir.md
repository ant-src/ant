<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Test.tempDir

```ant signature
func Test.tempDir(): string
```

Test.tempDir returns a new directory, removed with its contents when the current test ends. Panics
if the directory cannot be made.

<!-- end generated -->

## Details

- Each call makes another directory, empty and with mode `0o700`, as
  [File.tempDir](../File/tempDir.md) makes one.
- The directory is removed after the test's deferred blocks have run and its goroutines have ended
  or been ended, whether the test passed, failed, or was skipped. A `defer` in the test can still
  use it.
- In a subtest it makes a directory for the subtest, removed when the subtest ends; on a goroutine
  the test owns, one for that test.

## Basic examples

Write a file into a fresh directory and read it back. Nothing has to remove it:

<!-- test: cmd ant test basic_test.ant -->

```ant [basic_test.ant]
test "write and read back" {
  let dir = Test.tempDir()
  let path = Path.join([dir, "notes.txt"])
  try File.write(path, "hello")
  assert(try File.read(path) == "hello")
}
```

```text
ok    1 passed (0.00s)
```

## Scope

### Helpers

A helper function calls `Test.tempDir` without being given a handle to the test. Here `writeConfig`
puts a file in a new directory and returns its path, so each test reads its own:

```ant [config.ant]
export type Config = { name: string, port: int }

export func load(path: string): Result[Config, Error] {
  let text = try File.read(path).wrapErr("load ${path}")
  let lines = text.trim().split("\n")
  let port = lines[1].toInt() ?? return Result.fail("load ${path}: bad port")
  Ok{ value: Config{ name: lines[0], port } }
}
```

<!-- test: cmd ant test config_test.ant -->

```ant [config_test.ant]
import "./config.ant" { load }

func writeConfig(text: string): string {
  let path = Path.join([Test.tempDir(), "app.conf"])
  Result.unwrap(File.write(path, text))
  path
}

test "load reads the name and port" {
  let config = try load(writeConfig("web\n8080\n"))
  assert(config.name == "web")
  assert(config.port == 8080)
}

test "load rejects a bad port" {
  assert(load(writeConfig("web\nhttp\n")) is Err)
}

test "load of a missing file is Err" {
  let path = Path.join([Test.tempDir(), "absent.conf"])
  assert(load(path) is Err)
}
```

```text
ok    3 passed (0.00s)
```

### Subtests and goroutines

In a subtest, `Test.tempDir` makes a directory for the subtest, removed when it ends, while the
parent's directory stays. A goroutine the test owns gets one for that test:

<!-- test: cmd ant test scope_test.ant -->

```ant [scope_test.ant]
test "subtests" {
  let parent = Test.tempDir()
  var inner = ""
  test "inner" {
    inner = Test.tempDir()
    assert(inner != parent)
  }
  assert(!File.exists(inner))
  assert(File.exists(parent))
}

test "goroutine" {
  let paths: Channel[string] = Channel.new()
  go { paths.send(Test.tempDir()) }
  let dir = paths.receive() ?? ""
  assert(File.exists(dir))
}
```

```text
ok    2 passed (0.00s)
```

## Applications

A deferred block that checks what the test left behind still finds the directory, since it is
removed only after the deferred blocks run:

<!-- test: cmd ant test log_test.ant -->

```ant [log_test.ant]
func appendLog(path: string, line: string): Result[void, Error] {
  let log = try File.append(path)
  defer { log.close() }
  log.writeLine(line)
}

test "every step is logged" {
  let log = Path.join([Test.tempDir(), "steps.log"])
  defer {
    let lines = Result.unwrap(File.read(log)).trim().split("\n")
    assert(lines == ["start", "stop"])
  }
  try appendLog(log, "start")
  try appendLog(log, "stop")
}
```

```text
ok    1 passed (0.00s)
```

## Properties and relations

Each call makes a new, empty directory with mode `0o700`. Once the test ends, the directory is gone,
even when the test failed:

<!-- test: cmd ant test props_test.ant -->
<!-- test: exit 1 -->

```ant [props_test.ant]
var saved = ""

test "a new private directory" {
  let dir = Test.tempDir()
  let info = try File.stat(dir)
  assert(info.isDir)
  assert(info.mode == 0o700)
  let entries = try File.list(dir)
  assert(entries.isEmpty())
  assert(Test.tempDir() != dir)
}

test "fails halfway" {
  saved = Test.tempDir()
  try File.write(Path.join([saved, "partial.txt"]), "half")
  assert(false, "stopped")
}

test "the failed test left nothing" {
  assert(!File.exists(saved))
}
```

```text
--- FAIL: props_test.ant: fails halfway (0.00s)
    assertion failed: false: stopped
      at test "fails halfway" (props_test.ant:16:9)
FAIL  1 failed, 2 passed (0.00s)
```

## Possible issues

`Test.tempDir` works only while a test runs. In a script, or in a module initialized before the
tests start, it panics. Use [File.tempDir](../File/tempDir.md) there, and remove the directory
yourself:

```ant
let dir = Test.tempDir() // panic: Test.tempDir: not in a test
```

A path kept past the end of its test names a directory that no longer exists. Make the directory in
each test that needs it rather than sharing one across tests.

## See also

[File.tempDir](../File/tempDir.md) · [File.tempFile](../File/tempFile.md) ·
[File.removeAll](../File/removeAll.md) · [Test.context](context.md) · [Test.skip](skip.md) ·
[test](../../language/keywords/test.md)
