<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Csv.parse

```ant signature
func Csv.parse(text: string, sep: string = ","): Result[Array[Array[string]], Error]
```

Csv.parse parses text as RFC 4180 CSV with sep between fields. Each row keeps as many fields as
written, nothing is trimmed, and empty lines are dropped; it is Err, naming the line, on a quote
inside an unquoted field or a quoted field that never closes. It panics if sep is not one code point
other than ", \\r, \\n, U+0000, and U+FFFD.

<!-- end generated -->

## Details

- Rows may differ in length; `Csv.parse` does not check them against the first row.
- An empty `text`, or one of only line breaks, gives `Ok{ value: [] }`.
- A line break inside a quoted field reads as `\n`, whether the text had `\n` or `\r\n`.
- A bad `sep` panics before any text is read, since the separator is part of the program, not the
  data. Malformed text is `Err`.

## Basic examples

Parse two rows. A quoted field can hold the separator:

```ant
Csv.parse("a,\"b,c\"\n1,2\n") // Ok{ value: [["a", "b,c"], ["1", "2"]] }
```

The last line needs no line break:

```ant
Csv.parse("name,age\nAnn,31") // Ok{ value: [["name", "age"], ["Ann", "31"]] }
```

## Scope

An empty text has no rows, and empty lines between rows are dropped:

```ant
Csv.parse("")             // Ok{ value: [] }
Csv.parse("a,b\n\n\nc,d") // Ok{ value: [["a", "b"], ["c", "d"]] }
```

Rows end at `\n` or `\r\n`:

```ant
Csv.parse("a,b\r\nc,d\r\n") // Ok{ value: [["a", "b"], ["c", "d"]] }
```

A doubled quote inside a quoted field is one quote:

```ant
Csv.parse("\"say \"\"hi\"\"\",x") // Ok{ value: [["say \"hi\"", "x"]] }
```

A quoted field can span lines, and its line break reads as `\n`:

```ant
Csv.parse("\"two\r\nlines\",x") // Ok{ value: [["two\nlines", "x"]] }
```

Each row keeps the fields it has, and empty fields stay as `""`:

```ant
Csv.parse("a,b,c\n1\n2,,") // Ok{ value: [["a", "b", "c"], ["1"], ["2", "", ""]] }
Csv.parse(",")             // Ok{ value: [["", ""]] }
Csv.parse("\"\"")          // Ok{ value: [[""]] }
```

White space around fields is kept:

```ant
Csv.parse(" a , b ") // Ok{ value: [[" a ", " b "]] }
```

Fields hold any Unicode text:

```ant
Csv.parse("city,name\nZürich,🐜") // Ok{ value: [["city", "name"], ["Zürich", "🐜"]] }
```

Call it in a pipe to work on the rows inside the `Result`:

```ant
Csv.parse("a,b\n1,2") | Result.map({ x.length() }) // Ok{ value: 2 }
```

## Named arguments

`sep` sets the separator between fields, such as `;` for spreadsheets in many European locales:

```ant
Csv.parse("name;price\nTea;1,50", sep: ";") // Ok{ value: [["name", "price"], ["Tea", "1,50"]] }
```

Tab-separated text uses `"\t"`:

```ant
Csv.parse("id\tname\n7\tAnn", sep: "\t") // Ok{ value: [["id", "name"], ["7", "Ann"]] }
```

Any single code point works, not only ASCII:

```ant
Csv.parse("a→b→c", sep: "→") // Ok{ value: [["a", "b", "c"]] }
```

With another separator, a comma is ordinary text:

```ant
Csv.parse("a,b|c", sep: "|") // Ok{ value: [["a,b", "c"]] }
```

## Applications

Read a file, skip the header row, and turn each row into a record:

<!-- test: file people.csv
name,age
Ann,31
Bob,27
-->

```ant
type Person = { name: string, age: int }

func parsePeople(text: string): Result[Array[Person], Error] {
  let rows = try Csv.parse(text)
  if rows.isEmpty() { return Result.fail("no header row") }
  let people: Array[Person] = []
  for row in rows.slice(1, rows.length()) {
    if row.length() != 2 {
      return Result.fail("want 2 fields, got ${row.length()}")
    }
    let age = row[1].toInt()
    if age == null {
      return Result.fail("age of ${row[0]} is not a number")
    }
    people.push(Person{ name: row[0], age })
  }
  Result.ok(people)
}

let text = Result.unwrap(File.read("people.csv"))
for p in Result.unwrap(parsePeople(text)) {
  println("${p.name} is ${p.age}")
}
```

```text
Ann is 31
Bob is 27
```

Add the file name to a parse error with `Result.wrapErr`:

<!-- test: file broken.csv
id,note
1,say "hi"
-->

```ant
func load(path: string): Result[Array[Array[string]], Error] {
  let text = try File.read(path)
  Csv.parse(text) | Result.wrapErr(path)
}

let rows = load("broken.csv")
if rows is Err {
  println(rows.error.message)
}
```

```text
broken.csv: parse error on line 2, column 7: bare " in non-quoted-field
```

Sum a column found by its header name:

```ant
func columnSum(rows: Array[Array[string]], name: string): Result[int, Error] {
  if rows.isEmpty() { return Result.fail("no header row") }
  let col = rows[0].indexOf(name)
  if col == null { return Result.fail("no column ${name}") }
  var total = 0
  for row in rows.slice(1, rows.length()) {
    let n = row[col].toInt()
    if n == null { return Result.fail("${name} ${row[col]} is not a number") }
    total += n
  }
  Result.ok(total)
}

let rows = Result.unwrap(Csv.parse("item,qty\napple,3\npear,5\nfig,2"))
println(columnSum(rows, "qty"))
println(columnSum(rows, "price"))
```

```text
Ok{ value: 10 }
Err{ error: Error{ message: "no column price" } }
```

## Properties and relations

[Csv.stringify](stringify.md) writes rows that `Csv.parse` reads back as they were, quotes, line
breaks, and leading spaces included:

```ant
let rows = [["name", "note"], ["Ann", "said \"hi\", then left"], ["Bob", "two\nlines"], [" x", ""]]
assert(Result.unwrap(Csv.parse(Csv.stringify(rows))) == rows)
```

The same holds with another separator, when both calls use it:

```ant
let rows = [["a;b", "c"], ["1", "2"]]
assert(Result.unwrap(Csv.parse(Csv.stringify(rows, sep: ";"), sep: ";")) == rows)
```

`\r\n` and `\n` line endings give the same rows:

```ant
assert(Csv.parse("a,b\r\nc,d\r\n") == Csv.parse("a,b\nc,d\n"))
```

## Possible issues

A quote inside a field that is not quoted is `Err`. The message names the line and column:

```ant
Csv.parse("id,note\n1,say \"hi\"") // Err{ error: Error{ message: "parse error on line 2, column 7: bare \" in non-quoted-field" } }
```

A space before an opening quote makes the field unquoted, so the quote is out of place:

```ant
Csv.parse("a, \"b\"") // Err{ error: Error{ message: "parse error on line 1, column 4: bare \" in non-quoted-field" } }
```

A quoted field that never closes is `Err`:

```ant
Csv.parse("a,b\n\"open\n") // Err{ error: Error{ message: "parse error on line 2, column 7: extraneous or missing \" in quoted-field" } }
```

A separator that is empty, longer than one code point, a quote, a line break, U+0000, or U+FFFD
panics, even on empty text:

```ant
Csv.parse("a;;b", sep: ";;") // panic: Csv.parse needs a separator of one character other than a quote, a line break, U+0000 or U+FFFD, got ";;"
Csv.parse("", sep: "")       // panic: Csv.parse needs a separator of one character other than a quote, a line break, U+0000 or U+FFFD, got ""
Csv.parse("a", sep: "\"")    // panic: Csv.parse needs a separator of one character other than a quote, a line break, U+0000 or U+FFFD, got "\""
```

Numbers stay strings. Compare them as numbers only after converting:

```ant
let row = Result.unwrap(Csv.parse("9,10"))[0]
row[0] < row[1]                               // false
(row[0].toInt() ?? 0) < (row[1].toInt() ?? 0) // true
```

A row with no fields is written as an empty line, which `Csv.parse` drops, so it does not survive a
round trip:

```ant
Csv.parse(Csv.stringify([["a"], [], ["b"]])) // Ok{ value: [["a"], ["b"]] }
```

## See also

[Csv.stringify](stringify.md), [Csv](index.md), [Result.wrapErr](../Result/wrapErr.md),
[Result.unwrap](../Result/unwrap.md), [File.read](../File/read.md),
[String.split](../String/split.md), [String.lines](../String/lines.md)
