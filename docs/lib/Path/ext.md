<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.ext

```ant signature
func Path.ext(p: string): string
```

Path.ext returns the extension of p's last element, from its last dot, or "" when it has none.

<!-- end generated -->

## Details

- The letters are kept as written, so `.TXT` and `.txt` are different extensions.
- `Path.ext` works on the text alone and never looks at the file system.

## Basic examples

Get the extension of a file name, with its dot:

```ant
Path.ext("a/b.tar.gz") // ".gz"
Path.ext("main.ant")   // ".ant"
Path.ext("a/b")        // ""
```

## Scope

An empty path, and a name with no dot, give `""`:

```ant
Path.ext("")         // ""
Path.ext("Makefile") // ""
```

A dot in a directory name is not an extension:

```ant
Path.ext("v1.2/readme")     // ""
Path.ext("v1.2/readme.txt") // ".txt"
```

An absolute path works the same as a relative one:

```ant
Path.ext("/etc/nginx/nginx.conf") // ".conf"
```

Letters outside ASCII are kept as they are:

```ant
Path.ext("notes/résumé.pdf") // ".pdf"
Path.ext("data.été")         // ".été"
```

Call it in a pipe:

```ant
"src/main.ant" | Path.ext() // ".ant"
```

## Applications

Keep the Ant sources of a list of files:

```ant
let files = ["main.ant", "README.md", "lib/util.ant", "Makefile"]
files.filter({ Path.ext(x) == ".ant" }) // ["main.ant", "lib/util.ant"]
```

Match an extension whatever its case, by lowering it first:

```ant
func isImage(name: string): bool {
  let ext = Path.ext(name).toLower()
  ext == ".png" || ext == ".jpg" || ext == ".gif"
}

isImage("Photo.JPG") // true
isImage("photo.txt") // false
```

Strip the extension to get the name without it:

```ant
func stem(p: string): string {
  Path.base(p).trimSuffix(Path.ext(p))
}

stem("docs/guide.md")   // "guide"
stem("dist/app.tar.gz") // "app.tar"
stem("bin/ant")         // "ant"
```

Swap the extension of a path, keeping its directory:

```ant
func withExt(p: string, ext: string): string {
  p.trimSuffix(Path.ext(p)) + ext
}

withExt("src/main.ant", ".js") // "src/main.js"
withExt("README", ".md")       // "README.md"
```

Group files by extension:

```ant
let files = ["a.ant", "b.md", "c.ant", "Makefile"]
files.groupBy(Path.ext) // Map{ ".ant": ["a.ant", "c.ant"], ".md": ["b.md"], "": ["Makefile"] }
```

## Properties and relations

The extension is `""` or starts with a dot, and the last element ends with it:

```ant
for p in ["a/b.tar.gz", "a/b", "main.ant", "/etc/hosts"] {
  let ext = Path.ext(p)
  assert(ext.isEmpty() || ext.startsWith("."))
  assert(Path.base(p).endsWith(ext))
}
```

For a path that does not end in a separator, the extension is the extension of its last element:

```ant
let p = "photos/2024.06/beach.jpeg"
assert(Path.ext(p) == Path.ext(Path.base(p)))
```

## Possible issues

A double extension such as `.tar.gz` gives only its last part. Test the whole suffix with
`String.endsWith` instead:

```ant
Path.ext("dist/app.tar.gz")           // ".gz"
"dist/app.tar.gz".endsWith(".tar.gz") // true
```

A path that ends in a separator has no extension, even though its last element has one:

```ant
Path.ext("photos/album.d/")            // ""
Path.ext(Path.base("photos/album.d/")) // ".d"
```

A path is a plain `string`, and a call through `.` looks in `String`, so it does not find
`Path.ext`. Use the namespace form or a pipe:

```ant
let name = "notes.md"
let ext = name.ext() // error: string has no field or function ext
```

## See also

[Path.base](base.md), [Path.dir](dir.md), [Path.join](join.md), [Path.match](match.md),
[String.trimSuffix](../String/trimSuffix.md), [String.endsWith](../String/endsWith.md),
[Path](index.md)
