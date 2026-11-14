<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.touch

```ant signature
func File.touch(path: string, time: Time.Instant | null = null): Result[void, Error]
```

File.touch sets the modified and access times of the file at path to time, or to now when time is
null, creating an empty file when nothing is there. It never truncates, and it follows a link.

<!-- end generated -->

## Details

- A file that exists keeps its contents; only its times change. A new file is empty and gets mode
  `0o666` less the process's umask.
- At a symbolic link it sets the times of the target. At a dangling link it creates the target.
- The access time gets the same instant, but `File.Info` does not hold it, so only `modified` reads
  back.
- [File.stat](stat.md) reads the time back as `modified`, rounded as the file system keeps times: to
  2 seconds on FAT, to the second on HFS+, to the nanosecond on ext4 and APFS.

## Basic examples

Create an empty stamp file, or bring its modified time up to now:

```ant
let before = Time.now()
try File.touch("stamp")
let info = try File.stat("stamp")
println(info.size)
println(Time.compare(info.modified, Time.subtract(before, 2s)) >= 0)
```

```text
0
true
```

Set the modified time to a fixed instant:

```ant
let release = try Time.parseInstant("2026-01-15T09:30:00Z")
try File.touch("notes.txt", release)
println((try File.stat("notes.txt")).modified)
```

```text
2026-01-15T09:30:00Z
```

## Scope

A file that exists keeps its contents:

```ant
try File.write("todo.txt", "buy milk\n")
try File.touch("todo.txt")
print(try File.read("todo.txt"))
```

```text
buy milk
```

A directory takes a new time too:

```ant
let t = try Time.parseInstant("2025-12-31T23:59:59Z")
try File.mkdir("build")
try File.touch("build", t)
println((try File.stat("build")).modified)
```

```text
2025-12-31T23:59:59Z
```

Through a symbolic link, `File.touch` sets the times of the file the link points to, and leaves the
link's own time alone:

```ant
let t = try Time.parseInstant("2020-05-01T00:00:00Z")
try File.write("config.json", "{}")
try File.symlink("config.json", "current.json")
try File.touch("current.json", t)
println((try File.stat("config.json")).modified)
println((try File.stat("current.json", follow: false)).modified == t)
```

```text
2020-05-01T00:00:00Z
false
```

Any `Time.Instant` works, the Unix epoch and times before it included:

```ant
try File.touch("old", Time.fromEpochMillis(0))
println((try File.stat("old")).modified)
try File.touch("older", Time.fromEpochMillis(-86_400_000))
println((try File.stat("older")).modified)
```

```text
1970-01-01T00:00:00Z
1969-12-31T00:00:00Z
```

## Named arguments

`time:` names the instant to set. Leave it out, or pass `null`, for now:

```ant
let t = try Time.parseInstant("2026-03-01T12:00:00Z")
try File.touch("a.txt", time: t)
println((try File.stat("a.txt")).modified)
try File.touch("a.txt", time: null)
println((try File.stat("a.txt")).modified == t)
```

```text
2026-03-01T12:00:00Z
false
```

## Applications

Rebuild only when a source is newer than the stamp of the last build, as `make` does:

```ant
func stale(source: string, stamp: string): Result[bool, Error] {
  if !File.exists(stamp) { return Ok{ value: true } }
  let src = try File.stat(source)
  let built = try File.stat(stamp)
  Ok{ value: Time.compare(src.modified, built.modified) > 0 }
}

let monday = try Time.parseInstant("2026-11-09T10:00:00Z")
let tuesday = try Time.parseInstant("2026-11-10T10:00:00Z")
try File.write("site.md", "# Hello\n")
try File.touch("site.md", monday)
println(try stale("site.md", ".built"))
try File.touch(".built", tuesday)
println(try stale("site.md", ".built"))
try File.touch("site.md")
println(try stale("site.md", ".built"))
```

```text
true
false
true
```

[File.copy](copy.md) does not copy times. Give the copy the time of the original:

```ant
let t = try Time.parseInstant("2024-07-04T08:00:00Z")
try File.write("photo.jpg", "jpeg")
try File.touch("photo.jpg", t)
try File.copy("photo.jpg", "backup.jpg")
try File.touch("backup.jpg", (try File.stat("photo.jpg")).modified)
println((try File.stat("backup.jpg")).modified)
```

```text
2024-07-04T08:00:00Z
```

Find the logs older than a week. Backdated with `File.touch`, the example does not depend on the
clock:

```ant
let now = try Time.parseInstant("2026-11-12T00:00:00Z")
try File.mkdir("logs")
for days in [1, 10, 30] {
  try File.touch("logs/${days}d.log", Time.subtract(now, Time.duration(hours: days * 24)))
}

let cutoff = Time.subtract(now, 168h)
for path in try File.glob("logs/*.log") {
  let info = try File.stat(path)
  if Time.compare(info.modified, cutoff) < 0 {
    println(path)
  }
}
```

```text
logs/10d.log
logs/30d.log
```

Say which step failed with [Result.wrapErr](../Result/wrapErr.md):

```ant
let r = File.touch("cache/stamp") | Result.wrapErr("mark the cache fresh")
println(r)
```

```text
Err{ error: Error{ message: "mark the cache fresh: open cache/stamp: no such file or directory" } }
```

## Properties and relations

[File.stat](stat.md) reads back the instant `File.touch` set, and the size does not change:

```ant
try File.write("f.txt", "data")
for ms in [0, 1_000, 1_700_000_000_000, -5_000] {
  let t = Time.fromEpochMillis(ms)
  try File.touch("f.txt", t)
  let info = try File.stat("f.txt")
  assert(info.modified == t)
  assert(info.size == 4)
}
```

## Possible issues

`time` is a `Time.Instant`. Turn a zoned time into one with [Time.toInstant](../Time/toInstant.md):

```ant
let when = try Time.parseZoned("2026-01-15T09:30:00+01:00[Europe/Paris]")
try File.touch("stamp", when)
// error: cannot use Time.ZonedDateTime where Time.Instant | null is expected
```

```ant
let when = try Time.parseZoned("2026-01-15T09:30:00+01:00[Europe/Paris]")
try File.touch("stamp", Time.toInstant(when))
println((try File.stat("stamp")).modified)
```

```text
2026-01-15T08:30:00Z
```

`out` does not exist yet, so the touch is `Err`; make it with [File.mkdir](mkdir.md) first:

```ant
println(File.touch("out/stamp"))
try File.mkdir("out")
println(File.touch("out/stamp"))
```

```text
Err{ error: Error{ message: "open out/stamp: no such file or directory" } }
Ok{ value: void }
```

A dangling link is followed too, so `File.touch` creates the file it points to:

```ant
try File.symlink("target.txt", "link.txt")
println(File.exists("target.txt"))
try File.touch("link.txt")
println(File.exists("target.txt"))
```

```text
false
true
```

## See also

[File.stat](stat.md) · [File.Info](types/Info.md) · [File.write](write.md) ·
[File.create](create.md) · [File.copy](copy.md) · [File.chmod](chmod.md) ·
[Time.now](../Time/now.md) · [Time.parseInstant](../Time/parseInstant.md)
