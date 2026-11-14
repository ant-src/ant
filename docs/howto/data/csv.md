<!-- status: reviewed -->

# Read and write CSV

This script reads a list of orders from `orders.csv`, adds up the amount of each customer, and
writes the totals to `totals.csv`:

```csv [orders.csv]
customer,amount
Ann,30
"Bob, Jr.",12
Ann,5
```

Save it as `report.ant` next to `orders.csv` and run it with `ant run report.ant`:

<!-- test: cmd ant run report.ant -->

```ant [report.ant]
type Order = { customer: string, amount: int }

func parseOrder(row: Array[string]): Result[Order, Error] {
  if row.length() != 2 {
    return Result.fail("want 2 fields, got ${row.length()}: ${row}")
  }
  let amount = row[1].toInt()
  if amount == null {
    return Result.fail("amount is not an integer: ${row[1]}")
  }
  Result.ok(Order{ customer: row[0], amount })
}

func total(e: Entry[string, Array[Order]]): Array[string] {
  let sum = e.value.map({ x.amount }).sum() ?? 0
  [e.key, "${sum}"]
}

let text = try File.read("orders.csv")
let rows = try (Csv.parse(text) | Result.wrapErr("orders.csv"))
if rows.first() != ["customer", "amount"] {
  return Result.fail("orders.csv: want the header customer,amount")
}
let orders = try (rows.slice(1, rows.length()) | Array.map(parseOrder) | Result.all())
let totals = orders
  | Array.groupBy({ x.customer })
  | Map.entries()
  | Array.map(total)
try File.write("totals.csv", Csv.stringify([["customer", "total"]].concat(totals)))
print(try File.read("totals.csv"))
```

```text
customer,total
Ann,35
"Bob, Jr.",12
```

1. Read the file with [File.read](../../lib/File/read.md).
2. Split it into rows of fields with [Csv.parse](../../lib/Csv/parse.md). It returns every row as an
   `Array[string]`, the header included, with quotes removed: `"Bob, Jr."` is one field. Add the
   file name to a parse error with [Result.wrapErr](../../lib/Result/wrapErr.md).
3. Check the header row, then drop it with [Array.slice](../../lib/Array/slice.md).
4. Turn each row into a record. `Csv.parse` does not check that rows have the same number of fields,
   and every field is a string, so check the length and convert numbers with
   [String.toInt](../../lib/String/toInt.md). Return an error with
   [Result.fail](../../lib/Result/fail.md) for a row that does not fit. `Csv.parse` trims nothing,
   so `Ann, 30` gives the field `" 30"`, which `toInt` rejects; call `.trim()` on a field before
   converting it when the file has spaces after the separators.
5. Convert all rows with [Array.map](../../lib/Array/map.md), and stop at the first bad one with
   [Result.all](../../lib/Result/all.md).
6. Compute the result, here with [Array.groupBy](../../lib/Array/groupBy.md),
   [Map.entries](../../lib/Map/entries.md) and [Array.sum](../../lib/Array/sum.md), and turn each
   result back into an array of strings.
7. Put the header in front with [Array.concat](../../lib/Array/concat.md), and write the rows with
   [Csv.stringify](../../lib/Csv/stringify.md) and [File.write](../../lib/File/write.md).
   `Csv.stringify` quotes the fields that need it and ends every row with `\n`.

A bad row ends the script: `try` writes the error to standard error and exits with status 1. With
`Bob,x` in `orders.csv`, the script prints `amount is not an integer: x`.

## Read tab-separated values

Pass the separator as `sep:` to both functions. It is one character other than a quote, a line
break, U+0000 or U+FFFD, such as a tab or `;`; any other value panics, because the separator is part
of the program, not of the data. The file must have a header row, so check it before dropping it:

<!-- test: file scores.tsv
name	score
Ann	31
Bob	27
-->

```ant
let rows = try Csv.parse(try File.read("scores.tsv"), sep: "\t")
if rows.first() != ["name", "score"] {
  return Result.fail("scores.tsv: want the header name and score")
}
for row in rows.slice(1, rows.length()) {
  println("${row[0]}: ${row[1]}")
}
print(Csv.stringify(rows, sep: ";"))
```

```text
Ann: 31
Bob: 27
name;score
Ann;31
Bob;27
```

## Handle a malformed file

`Csv.parse` returns an error naming the line on a quote inside a field that does not start with a
quote, or on a quoted field that never closes:

<!-- test: file bad.csv
name,note
Ann,says "hi"
-->
<!-- test: exit 1 -->

```ant
let rows = try (Csv.parse(try File.read("bad.csv")) | Result.wrapErr("bad.csv"))
println(rows.length())
```

The script prints `bad.csv: parse error on line 2, column 10: bare " in non-quoted-field` to
standard error and exits with status 1. Write such a field quoted, with the inner quotes doubled:
`Ann,"says ""hi"""`. `Csv.stringify` writes it that way.

## See also

[Csv.parse](../../lib/Csv/parse.md), [Csv.stringify](../../lib/Csv/stringify.md),
[Decode JSON](json-decode.md), [Read standard input line by line](../cli/stdin.md)
