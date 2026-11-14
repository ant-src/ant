<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Csv.stringify

```ant signature
func Csv.stringify(rows: Array[Array[string]], sep: string = ","): string
```

Csv.stringify writes rows as CSV with sep between fields, a line per row ending in \\n. A field is
quoted when it holds sep, a quote, or a line break, or begins with white space, and a row of one
empty field is written "". It panics if sep is not a separator [`Csv.parse`](parse.md) accepts.

<!-- end generated -->

## Details

- The last row ends with `\n` too.
- Rows are written as given: two rows may have different numbers of fields.
- A row with no fields is an empty line, which [`Csv.parse`](parse.md) drops.

## Basic examples

Write rows as CSV text:

```ant
Csv.stringify([["a", "b,c"], ["1", "2"]]) // "a,\"b,c\"\n1,2\n"
```

Print the text to see the lines it holds:

```ant
print(Csv.stringify([["name", "city"], ["Ann", "Oslo"], ["Bob", "Lima, Peru"]]))
```

```text
name,city
Ann,Oslo
Bob,"Lima, Peru"
```

## Scope

No rows give the empty string:

```ant
Csv.stringify([]) // ""
```

A field with a quote is quoted, and each quote in it doubled:

```ant
Csv.stringify([["say \"hi\""]]) // "\"say \"\"hi\"\"\"\n"
```

A field with a line break is quoted, so it stays one field:

```ant
Csv.stringify([["line one\nline two", "x"]]) // "\"line one\nline two\",x\n"
```

A field that begins with white space is quoted; one that only ends with it is not:

```ant
Csv.stringify([[" left", "right ", "\ttab"]]) // "\" left\",right ,\"\ttab\"\n"
```

Empty fields are written as nothing between separators:

```ant
Csv.stringify([["a", "", "c"], ["", ""]]) // "a,,c\n,\n"
```

A row of one empty field is written `""`, so it is not an empty line:

```ant
Csv.stringify([["a"], [""], ["b"]]) // "a\n\"\"\nb\n"
```

Rows may have different numbers of fields:

```ant
Csv.stringify([["id"], ["1", "Ann", "admin"]]) // "id\n1,Ann,admin\n"
```

Unicode text needs no quoting:

```ant
Csv.stringify([["日本", "café"]]) // "日本,café\n"
```

Call it in a pipe at the end of building rows:

```ant
[[3, 4], [5, 12]]
  | Array.map({ x.map({ "${x}" }) })
  | Csv.stringify() // "3,4\n5,12\n"
```

## Named arguments

### sep

`sep` puts another separator between fields, and a field that holds it is quoted:

```ant
Csv.stringify([["a;b", "c,d"]], sep: ";") // "\"a;b\";c,d\n"
```

A tab writes tab-separated values:

```ant
Csv.stringify([["name", "note"], ["Ann", "a\tb"]], sep: "\t") // "name\tnote\nAnn\t\"a\tb\"\n"
```

Any one code point other than a quote, `\r`, `\n`, U+0000, and U+FFFD works:

```ant
Csv.stringify([["x", "y"]], sep: "→") // "x→y\n"
```

## Applications

Export records with a header row:

```ant
type Person = { name: string, age: int }

let people = [Person{ name: "Ann", age: 31 }, Person{ name: "Bob, Jr.", age: 27 }]
let rows = [["name", "age"]].concat(people.map({ [x.name, "${x.age}"] }))
print(Csv.stringify(rows))
```

```text
name,age
Ann,31
"Bob, Jr.",27
```

Save a report to a file and read it back:

```ant
let rows = [["day", "visits"], ["Mon", "120"], ["Tue", "98"]]
try File.write("visits.csv", Csv.stringify(rows))
let back = try Csv.parse(try File.read("visits.csv"))
println(back[1])
```

```text
["Mon", "120"]
```

Write a table for a spreadsheet that expects semicolons, as many do where the comma is the decimal
mark:

```ant
let prices = Map{ "tea": 2.5, "coffee": 3.25 }
let rows = prices.entries().map({ [x.key, "${x.value}".replace(".", ",")] })
print(Csv.stringify(rows, sep: ";"))
```

```text
tea;2,5
coffee;3,25
```

## Properties and relations

[`Csv.parse`](parse.md) reads back the rows `Csv.stringify` writes, with the same separator:

```ant
let rows = [["name", "note"], ["Ann", "said \"hi\", left"], ["", " x"], [""]]
let back = try Csv.parse(Csv.stringify(rows))
let piped = try Csv.parse(Csv.stringify(rows, sep: "|"), sep: "|")
assert(back == rows)
assert(piped == rows)
```

The text has one line per row when no field holds a line break:

```ant
let rows = [["a", "b"], ["c, d", "e"], ["f"]]
let text = Csv.stringify(rows)
assert(text.endsWith("\n"))
assert(text.split("\n").length() == rows.length() + 1)
```

## Possible issues

A separator must be one code point, so a two-code-point one such as `", "` panics:

```ant
Csv.stringify([["a", "b"]], sep: ", ")
// panic: Csv.stringify needs a separator of one character other than a quote, a line break, U+0000 or U+FFFD, got ", "
```

The separator is checked first, so a bad one panics even with no rows:

```ant
Csv.stringify([], sep: "")
// panic: Csv.stringify needs a separator of one character other than a quote, a line break, U+0000 or U+FFFD, got ""
```

A row with no fields is written as an empty line, and `Csv.parse` drops empty lines, so the row is
lost on the way back:

```ant
let rows = [["a"], [], ["b"]]
Csv.stringify(rows)            // "a\n\nb\n"
Csv.parse(Csv.stringify(rows)) // Ok{ value: [["a"], ["b"]] }
```

`Csv.parse` reads a line break inside a quoted field as `\n`, so a field with `\r\n` comes back with
`\n`:

```ant
Csv.parse(Csv.stringify([["a\r\nb"]])) // Ok{ value: [["a\nb"]] }
```

## See also

[Csv.parse](parse.md), [Json.stringify](../Json/stringify.md), [String.join](../String/join.md),
[File.write](../File/write.md)
