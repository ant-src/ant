<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Csv

Csv converts between CSV text, as RFC 4180 describes it, and rows of fields.

<!-- end generated -->

## Details

- Every field is a `string`, both ways. Convert numbers yourself, with
  [String.toInt](../String/toInt.md) after parsing and interpolation before writing.
- A header row is an ordinary row: `Csv.parse` returns it first, and neither function checks that
  rows have the same number of fields.
- Malformed text is an `Err` from [Csv.parse](parse.md), naming the line. A separator that is not
  one code point other than a quote, `\r`, `\n`, U+0000, and U+FFFD panics in both functions, since
  the separator is part of the program, not the data.
- Csv works on text in memory. Read and write files with [File.read](../File/read.md) and
  [File.write](../File/write.md).

## Basic examples

Parse CSV text into rows of fields with `Csv.parse`:

```ant
Csv.parse("name,age\nAnn,31\nBob,27") // Ok{ value: [["name", "age"], ["Ann", "31"], ["Bob", "27"]] }
```

Write rows as CSV text with `Csv.stringify`. Each row ends with a line break:

```ant
Csv.stringify([["name", "age"], ["Ann", "31"]]) // "name,age\nAnn,31\n"
```

A field that holds the separator, a quote, or a line break, or begins with white space, is quoted on
the way out and unquoted on the way in:

```ant
let rows = [["city", "note"], ["Lima, Peru", "said \"hi\""]]
print(Csv.stringify(rows))
println(Csv.parse(Csv.stringify(rows)))
```

```text
city,note
"Lima, Peru","said ""hi"""
Ok{ value: [["city", "note"], ["Lima, Peru", "said \"hi\""]] }
```

## Scope

Both functions take `sep:`, for semicolon- or tab-separated text:

```ant
Csv.parse("name;price\nTea;1,50", sep: ";")              // Ok{ value: [["name", "price"], ["Tea", "1,50"]] }
Csv.stringify([["id", "name"], ["7", "Ann"]], sep: "\t") // "id\tname\n7\tAnn\n"
```

Rows may have different numbers of fields, and empty fields stay `""`:

```ant
Csv.parse("a,b,c\n1\n2,,") // Ok{ value: [["a", "b", "c"], ["1"], ["2", "", ""]] }
```

`Csv.parse` reads `\n` and `\r\n` line endings, and drops empty lines. `Csv.stringify` writes `\n`:

```ant
Csv.parse("a,b\r\n\r\nc,d\r\n") // Ok{ value: [["a", "b"], ["c", "d"]] }
```

No text gives no rows, and no rows give no text:

```ant
Csv.parse("")     // Ok{ value: [] }
Csv.stringify([]) // ""
```

A field with leading white space is quoted, and a row of one empty field is written `""`, so
`Csv.parse` keeps it:

```ant
Csv.stringify([[" x", "y "], [""]]) // "\" x\",y \n\"\"\n"
```

## Functions by job

- [Csv.parse](parse.md): the rows of CSV text, or an `Err` naming the line of a stray or unclosed
  quote
- [Csv.stringify](stringify.md): CSV text for rows, a line per row, quoting fields that need it

## Applications

Read a file, skip its header, and total a column for each group:

<!-- test: file sales.csv
region,amount
north,120
south,80
north,45
-->

```ant
func totals(rows: Array[Array[string]]): Result[Map[string, int], Error] {
  let sums: Map[string, int] = Map{}
  for row in rows.slice(1, rows.length()) {
    let amount = row.get(1)?.toInt()
    if amount == null {
      return Result.fail("bad row ${row}")
    }
    sums[row[0]] = (sums.get(row[0]) ?? 0) + amount
  }
  Result.ok(sums)
}

let text = try File.read("sales.csv")
let rows = try (Csv.parse(text) | Result.wrapErr("sales.csv"))
println(try totals(rows))
```

```text
Map{ "north": 165, "south": 80 }
```

Convert tab-separated text to CSV:

```ant
let tsv = "id\tcity\n1\tLima, Peru\n2\tOslo\n"
let rows = try Csv.parse(tsv, sep: "\t")
print(Csv.stringify(rows))
```

```text
id,city
1,"Lima, Peru"
2,Oslo
```

Write records as CSV with a header row:

```ant
type Item = { name: string, qty: int }

let items = [Item{ name: "apple", qty: 3 }, Item{ name: "fig, dried", qty: 12 }]
let rows = [["name", "qty"]].concat(items.map({ [x.name, "${x.qty}"] }))
print(Csv.stringify(rows))
```

```text
name,qty
apple,3
"fig, dried",12
```

## Properties and relations

`Csv.parse` reads back the rows `Csv.stringify` writes, with the same separator, as long as no row
is empty:

```ant
let rows = [["a", "b, c"], ["two\nlines", " x"], [""]]
assert(Result.unwrap(Csv.parse(Csv.stringify(rows))) == rows)
assert(Result.unwrap(Csv.parse(Csv.stringify(rows, sep: ";"), sep: ";")) == rows)
```

## Possible issues

Fields stay strings, so numbers compare as text until you convert them:

```ant
let row = Result.unwrap(Csv.parse("9,10"))[0]
row[0] < row[1]                               // false
(row[0].toInt() ?? 0) < (row[1].toInt() ?? 0) // true
```

A quote inside an unquoted field is an `Err`, not a panic:

```ant
Csv.parse("id,note\n1,say \"hi\"") // Err{ error: Error{ message: "parse error on line 2, column 7: bare \" in non-quoted-field" } }
```

A separator of more than one code point panics:

```ant
Csv.parse("a, b", sep: ", ") // panic: Csv.parse needs a separator of one character other than a quote, a line break, U+0000 or U+FFFD, got ", "
```

A row with no fields is written as an empty line, which `Csv.parse` drops:

```ant
Csv.parse(Csv.stringify([["a"], [], ["b"]])) // Ok{ value: [["a"], ["b"]] }
```

## See also

[Csv.parse](parse.md), [Csv.stringify](stringify.md), [Json](../Json/index.md),
[File.read](../File/read.md), [Result.wrapErr](../Result/wrapErr.md),
[String.toInt](../String/toInt.md)
