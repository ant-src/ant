<!-- status: reviewed -->

# Read and change modes, owners, and times

Check and set who owns a file, who may read, write, and run it, and when it last changed.

1. Read the mode and the modified time with [File.stat](../../lib/File/stat.md). Both are fields of
   the [File.Info](../../lib/File/types/Info.md) it returns.
2. Set the mode with [File.chmod](../../lib/File/chmod.md). It sets every bit, so build the new mode
   from the old one with [Bits.or](../../lib/Bits/or.md) and [Bits.and](../../lib/Bits/and.md) to
   change only a few.
3. Read the owner and group ids from the `uid` and `gid` fields, and change them with
   [File.chown](../../lib/File/chown.md).
4. Set the modified time with [File.touch](../../lib/File/touch.md).
5. Compare two times with [Time.compare](../../lib/Time/compare.md).

## Read the mode and the modified time

Call `File.stat` and read `mode` and `modified`. Print a mode in octal with
[String.fromInt](../../lib/String/fromInt.md) and `base: 8`:

<!-- test: before
try File.write("report.txt", "total: 42\n")
try File.chmod("report.txt", 0o640)
try File.touch("report.txt", try Time.parseInstant("2026-11-02T08:15:00Z"))
-->

```ant
let info = try File.stat("report.txt")
println(String.fromInt(info.mode, base: 8))
println(info.modified)
```

```text
640
2026-11-02T08:15:00Z
```

`modified` is a [Time.Instant](../../lib/Time/types/Instant.md), with no zone. Show it as a reader's
clock does with [Time.inZone](../../lib/Time/inZone.md):

<!-- test: before
try File.write("report.txt", "total: 42\n")
try File.touch("report.txt", try Time.parseInstant("2026-11-02T08:15:00Z"))
-->

```ant
let info = try File.stat("report.txt")
let zone = try Time.zone("Europe/Stockholm")
println(Time.inZone(info.modified, zone))
```

```text
2026-11-02T09:15:00+01:00[Europe/Stockholm]
```

At a symbolic link, `File.stat` describes the file the link points to. Pass `follow: false` to
describe the link itself; its own mode is whatever the host reports, so do not rely on it.

## Test a permission bit

Mask the mode with `Bits.and` and compare with zero. `0o100` is the owner's execute bit, `0o002` the
write bit for others:

<!-- test: before
try File.write("deploy.sh", "echo deploying\n")
try File.chmod("deploy.sh", 0o755)
-->

```ant
let mode = (try File.stat("deploy.sh")).mode
let runnable = Bits.and(mode, 0o100) != 0
let worldWritable = Bits.and(mode, 0o002) != 0
println("runnable: ${runnable}, world-writable: ${worldWritable}")
```

```text
runnable: true, world-writable: false
```

The mode does not tell whether your program can write the file:
[the owner](#read-and-change-the-owner), root, access control lists, and read-only mounts decide
too. To find out, try the write and handle the `Err`.

## Change the mode

Give `File.chmod` the whole mode, written as an octal literal. The mode also holds the setuid,
setgid and sticky bits, `0o4000`, `0o2000` and `0o1000`; a mode outside `0 ..= 0o7777` is an `Err`:

```ant
try File.write("notes.txt", "draft")
try File.chmod("notes.txt", 0o600)
println(String.fromInt((try File.stat("notes.txt")).mode, base: 8))
```

```text
600
```

To add bits, as `chmod +x` does, `Bits.or` them into the current mode. To clear bits, as `chmod a-w`
does, `Bits.and` the mode with [Bits.not](../../lib/Bits/not.md) of them:

```ant
func addMode(path: string, bits: int): Result[void, Error] {
  let mode = (try File.stat(path)).mode
  File.chmod(path, Bits.or(mode, bits))
}

func clearMode(path: string, bits: int): Result[void, Error] {
  let mode = (try File.stat(path)).mode
  File.chmod(path, Bits.and(mode, Bits.not(bits)))
}

try File.write("build.sh", "make\n")
try File.chmod("build.sh", 0o640)
try addMode("build.sh", 0o111)
println(String.fromInt((try File.stat("build.sh")).mode, base: 8))
try clearMode("build.sh", 0o222)
println(String.fromInt((try File.stat("build.sh")).mode, base: 8))
```

```text
751
551
```

On Windows a mode only reflects the read-only attribute: `File.stat` reports `0o444` for a read-only
file and `0o666` otherwise, and `File.chmod` makes the file read-only when `0o200` is clear and
ignores every other bit.

A new file gets `0o666` less the process's umask, usually `0o644`, so do not count on it. Writing a
file that exists keeps its mode. To write a file nobody else may read, such as a token, do not write
it and then call `File.chmod`; [Replace a file without a partial write](replace-safely.md) shows a
file that is private from the start.

## Read and change the owner

`File.stat` reads the owner into `uid` and the group into `gid`. They are numeric ids, as `ls -n`
shows them, and `int | null`: on Windows and WebAssembly they are null, as
[Os.uid](../../lib/Os/uid.md) and [Os.gid](../../lib/Os/gid.md) are. Give `File.chown` the user id
and the group id; `-1` keeps one as it is. A new file may take its directory's group, as it does on
macOS, so this example sets the group first and then compares it:

```ant
try File.write("report.txt", "total: 42\n")
let info = try File.stat("report.txt")
println(info.uid == Os.uid())
try File.chown("report.txt", -1, Os.gid() ?? -1)
println((try File.stat("report.txt")).gid == Os.gid())
```

```text
true
true
```

Only root may give a file to another user, and an owner may set the group only to one it belongs to;
anything else is an `Err`. At a symbolic link `File.chown` changes the file the link points to; pass
`follow: false` to change the link itself. The ids are numbers, so look a user name up first with
`Cmd.read(Cmd.new("id", ["-u", name]))`. On Windows `File.chown` is an `Err`.

## Set the modified time

Pass an instant to `File.touch`. Without one, it sets the time to now. Either way it creates an
empty file when nothing is there, and never changes what a file holds:

```ant
let release = try Time.parseInstant("2026-01-15T09:30:00Z")
try File.write("CHANGES.txt", "v1.2\n")
try File.touch("CHANGES.txt", release)
println((try File.stat("CHANGES.txt")).modified)
print(try File.read("CHANGES.txt"))
```

```text
2026-01-15T09:30:00Z
v1.2
```

`File.stat` reads the time back rounded as the file system keeps it: to 2 seconds on FAT, to the
second on HFS+, and to the nanosecond on ext4 and APFS.

[File.copy](../../lib/File/copy.md) copies the mode but not the times or the owner. To keep the time
too, touch the copy with the original's:

```ant
let taken = try Time.parseInstant("2024-07-04T08:00:00Z")
try File.write("photo.jpg", "jpeg")
try File.touch("photo.jpg", taken)
try File.copy("photo.jpg", "backup.jpg")
try File.touch("backup.jpg", (try File.stat("photo.jpg")).modified)
println((try File.stat("backup.jpg")).modified)
```

```text
2024-07-04T08:00:00Z
```

## Compare modified times

`<` does not apply to times. Call `Time.compare`, which returns `-1`, `0`, or `1`, to see which file
is newer:

```ant
func newer(a: string, b: string): Result[bool, Error] {
  let left = try File.stat(a)
  let right = try File.stat(b)
  Result.ok(Time.compare(left.modified, right.modified) > 0)
}

try File.touch("site.md", try Time.parseInstant("2026-11-10T10:00:00Z"))
try File.touch("site.html", try Time.parseInstant("2026-11-09T10:00:00Z"))
println(try newer("site.md", "site.html"))
```

```text
true
```

To find files older than a cutoff, subtract a span from an instant with
[Time.subtract](../../lib/Time/subtract.md). Use [Time.now](../../lib/Time/now.md) as the instant in
a real script:

<!-- test: before
let day = try Time.parseInstant("2026-11-12T00:00:00Z")
try File.mkdir("logs")
try File.touch("logs/a.log", Time.subtract(day, 24h))
try File.touch("logs/b.log", Time.subtract(day, 240h))
-->

```ant
let now = try Time.parseInstant("2026-11-12T00:00:00Z")
let cutoff = Time.subtract(now, 168h)
for path in try File.glob("logs/*.log") {
  let info = try File.stat(path)
  if Time.compare(info.modified, cutoff) < 0 {
    println(path)
  }
}
```

```text
logs/b.log
```

## The whole program

This script prepares a directory of scripts for release: it makes every `.sh` file in `bin`
executable, takes write permission away from others, and lists the files changed since the last run,
which it marks with a stamp file:

<!-- test: before
try File.mkdir("bin")
try File.write("bin/build.sh", "make\n")
try File.write("bin/deploy.sh", "rsync\n")
try File.write("bin/README", "scripts\n")
try File.chmod("bin/build.sh", 0o666)
try File.chmod("bin/deploy.sh", 0o644)
try File.chmod("bin/README", 0o644)
try File.touch("bin/build.sh", try Time.parseInstant("2026-11-01T12:00:00Z"))
try File.touch("bin/deploy.sh", try Time.parseInstant("2026-11-13T12:00:00Z"))
try File.touch("bin/README", try Time.parseInstant("2026-10-01T12:00:00Z"))
try File.touch(".released", try Time.parseInstant("2026-11-05T12:00:00Z"))
-->

```ant
let stamp = ".released"
let since = (try File.stat(stamp)).modified

for name in try File.list("bin") {
  let path = Path.join(["bin", name])
  let info = try File.stat(path)
  if name.endsWith(".sh") {
    let mode = Bits.and(Bits.or(info.mode, 0o111), Bits.not(0o002))
    try (File.chmod(path, mode) | Result.wrapErr("release ${path}"))
    println("${path} ${String.fromInt(mode, base: 8)}")
  }
  if Time.compare(info.modified, since) > 0 {
    println("${path} changed")
  }
}

try File.touch(stamp)
```

```text
bin/build.sh 775
bin/deploy.sh 755
bin/deploy.sh changed
```

Besides the functions above, it uses [File.list](../../lib/File/list.md),
[Path.join](../../lib/Path/join.md), [String.endsWith](../../lib/String/endsWith.md), and
[Result.wrapErr](../../lib/Result/wrapErr.md).

## See also

[File.stat](../../lib/File/stat.md) · [File.Info](../../lib/File/types/Info.md) ·
[File.chmod](../../lib/File/chmod.md) · [File.chown](../../lib/File/chown.md) ·
[File.touch](../../lib/File/touch.md) · [Os.uid](../../lib/Os/uid.md) ·
[Os.gid](../../lib/Os/gid.md) · [File.copy](../../lib/File/copy.md) ·
[File.glob](../../lib/File/glob.md) · [Bits.and](../../lib/Bits/and.md) ·
[Bits.or](../../lib/Bits/or.md) · [Bits.not](../../lib/Bits/not.md) ·
[String.fromInt](../../lib/String/fromInt.md) · [Time.compare](../../lib/Time/compare.md) ·
[Time.inZone](../../lib/Time/inZone.md) · [Time.subtract](../../lib/Time/subtract.md) ·
[Time.parseInstant](../../lib/Time/parseInstant.md) ·
[Replace a file without a partial write](replace-safely.md) ·
[Walk a directory and match globs](walk-glob.md)
