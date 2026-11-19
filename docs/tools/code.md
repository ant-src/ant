<!-- status: draft -->

# ant code

`ant code` answers from a terminal the questions an editor answers when you hover over a name or
jump to its declaration: where a name is declared, where it is used, what type it has, which
declarations a module has, and how to rename one. It runs the same analysis as the language server,
on the files as they are on disk, and prints plain text or JSON.

```sh
ant code def     [flags] <loc>          # where a name is declared
ant code refs    [flags] <loc>          # every use of a declaration
ant code type    [flags] <loc>          # the type and doc of a name or an expression
ant code symbols [flags] [path...]      # the declarations of modules
ant code rename  [flags] <loc> <name>   # rename a declaration and its uses
```

## Details

- A location `<loc>` is one of `file:line:col`, `file:line:name`, `file:Name`, or a bare `Name`.
  Lines and columns count from 1, and a column counts bytes, as `ant check` prints them. `Name` may
  be `Owner.name`: a function on a type, a field of a record type, or a function of the standard
  library such as `Array.map`. Case matters.
- `file:line:name` finds the name on that line, so you need not count columns. A position on white
  space, in a comment, or just after a name selects nothing.
- Every subcommand takes `-json`, which prints one JSON object instead of text, and `-root dir`, the
  workspace, which is the current directory by default. `refs` takes `-decl`, `symbols` takes
  `-match text` and `-all`, and `rename` takes `-diff` and `-force`. Flags come before the operands.
- The workspace is every `.ant` file under the root, skipping directories whose names start with a
  dot and `node_modules`. `refs` and `rename` search it for the modules that import a declaration.
- Text output is one location per line, `path:line:col: text`, as `ant check` and `grep -n` print
  them. Counts, warnings, and errors go to standard error.
- `ant code` parses and type-checks, and never runs your code: no initializer, function, or test.
- The exit status is 0 for an answer, 1 when nothing is found, a location is ambiguous, or a rename
  is refused, and 2 when the usage is wrong, a file cannot be read, parsed, or written, or an answer
  may be incomplete.

## Basic examples

A shop keeps its money in one module:

```ant [shop/money.ant]
// Money is an amount in cents.
export type Money = { cents: int }

// Money.add returns the sum of a and b.
export func Money.add(a: Money, b: Money): Money {
  Money{ cents: a.cents + b.cents }
}

// format writes m as euros and cents, such as 3.50.
export func format(m: Money): string {
  let cents = String.from(m.cents % 100).padStart(2, "0")
  "${m.cents / 100}.${cents}"
}
```

Find where the `add` on line 9 of the cart is declared:

```sh
ant code def shop/cart.ant:9:add
```

<!-- test: cmd ant code def shop/cart.ant:9:add -->

```ant [shop/cart.ant]
import "./money.ant" { Money, format }

export type Item = { name: string, price: Money }

// total adds up the prices of items.
export func total(items: Array[Item]): Money {
  var sum = Money{ cents: 0 }
  for item in items {
    sum = sum.add(item.price)
  }
  sum
}

export func receipt(items: Array[Item]): string {
  let lines = items | Array.map({ "${x.name} ${format(x.price)}" })
  String.join(lines, "\n")
}
```

```text
shop/money.ant:5:19: func Money.add(a: Money, b: Money): Money
```

`def` prints the location of the declaration and its signature on one line.

List every use of `format` in the workspace, the script that imports it included:

```sh
ant code refs shop/money.ant:format
```

<!-- test: cmd ant code refs shop/money.ant:format -->

```ant [shop.ant]
import "./shop/cart.ant" { Item, receipt, total }
import "./shop/money.ant" { Money, format }

let items = [
  Item{ name: "tea", price: Money{ cents: 350 } },
  Item{ name: "cake", price: Money{ cents: 425 } },
]
println(receipt(items))
println("total ${format(total(items))}")
```

```text
shop.ant:2:36: import "./shop/money.ant" { Money, format }
shop.ant:9:18: println("total ${format(total(items))}")
shop/cart.ant:1:31: import "./money.ant" { Money, format }
shop/cart.ant:15:48: let lines = items | Array.map({ "${x.name} ${format(x.price)}" })
```

`refs` leaves the declaration out unless you add `-decl`, and prints `4 references in 2 files` on
standard error.

Ask what type the checker inferred for a name:

```sh
ant code type words.ant:2:groups
```

<!-- test: cmd ant code type words.ant:2:groups -->

```ant [words.ant]
let words = "the ant and the bee and the ant".split(" ")
let groups = words | Array.groupBy({ x }) | Map.entries()
for group in groups {
  println("${group.key} ${group.value.length()}")
}
```

```text
let groups: Array[Entry[string, Array[string]]]
```

## Scope

### Locations

A position names the identifier it falls on. `file:Name` names a declaration of a module, private or
exported, and a bare name searches the top-level declarations of the workspace and the standard
library:

```sh
ant code type shop/money.ant:Money.add
```

<!-- test: cmd ant code type shop/money.ant:Money.add -->

```ant [shop/tax.ant]
import "./money.ant" { Money }

// vat is the tax on m, at 20 percent.
export func vat(m: Money): Money {
  Money{ cents: m.cents / 5 }
}
```

```text
func Money.add(a: Money, b: Money): Money
    Money.add returns the sum of a and b.
```

`type` prints the signature, then the doc comment indented by four spaces.

### Ambiguous locations

When a location matches more than one binding, `ant code` answers about none of them. It lists the
candidates and exits with status 1, so that you can ask again with one of them:

```sh
ant code def -json parse
```

<!-- test: cmd ant code def -json parse -->
<!-- test: exit 1 -->
<!-- test: file query.ant
// parse reads a query string.
func parse(text: string): Array[string] {
  text.split("&")
}
-->

```ant [config.ant]
// parse reads a line of key=value.
func parse(line: string): string | null {
  line.cut("=")?.before
}

func parseAll(lines: Array[string]): Array[string | null] {
  lines | Array.map({ parse(x) })
}
```

```text
{"error":{"code":"ambiguous","message":"parse names 2 declarations","candidates":[{"path":"config.ant","line":2,"column":6,"start":{"line":2,"column":6},"end":{"line":2,"column":11}},{"path":"query.ant","line":2,"column":6,"start":{"line":2,"column":6},"end":{"line":2,"column":11}}]}}
```

Both `config.ant` and `query.ant` declare a private `parse`. Without `-json`, the candidates go to
standard error with their lines.

### Expressions

On an operator or a literal, `type` describes the innermost expression around it. On the second `|`
of a pipeline, that is the pipeline up to the step after it:

```sh
ant code type votes.ant:2:42
```

<!-- test: cmd ant code type votes.ant:2:42 -->

```ant [votes.ant]
let votes = ["yes", "no", "yes"]
let tally = votes | Array.groupBy({ x }) | Map.entries() | Array.map({ x.key })
println(tally)
```

```text
Array[Entry[string, Array[string]]]
```

A position on a keyword, in a comment, or on white space has no type, and neither has a name the
checker could not resolve. `type` exits with status 1 rather than guess.

### Symbols

`symbols` lists the declarations of the modules under each path, the current directory by default.
Each line is the location of the name and the signature as written, with `export` for an export:

```sh
ant code symbols shop
```

<!-- test: cmd ant code symbols shop -->

```ant [shop/discount.ant]
import "./money.ant" { Money }

// half takes 50 percent off m.
export func half(m: Money): Money {
  Money{ cents: m.cents / 2 }
}
```

```text
shop/cart.ant:3:13: export type Item = { name: string, price: Money }
shop/cart.ant:6:13: export func total(items: Array[Item]): Money
shop/cart.ant:14:13: export func receipt(items: Array[Item]): string
shop/discount.ant:4:13: export func half(m: Money): Money
shop/money.ant:2:13: export type Money = { cents: int }
shop/money.ant:5:19: export func Money.add(a: Money, b: Money): Money
shop/money.ant:10:13: export func format(m: Money): string
shop/tax.ant:4:13: export func vat(m: Money): Money
```

A function on a type follows the type. In a test file, the tests come last, each named as
`ant test -run` matches it. `-match text` keeps the declarations whose name contains `text`, in any
case, and `-all` adds the fields of record types and the declarations inside functions, with their
types.

### Renaming

`rename -diff` prints the change as a unified diff and writes nothing:

```sh
ant code rename -diff temperature.ant:3:above delta
```

<!-- test: cmd ant code rename -diff temperature.ant:3:above delta -->

```ant [temperature.ant]
// celsius converts degrees Fahrenheit to Celsius.
export func celsius(f: float64): float64 {
  let above = f - 32.0
  above * 5.0 / 9.0
}
```

```text
--- a/temperature.ant
+++ b/temperature.ant
@@ -1,5 +1,5 @@
 // celsius converts degrees Fahrenheit to Celsius.
 export func celsius(f: float64): float64 {
-  let above = f - 32.0
-  above * 5.0 / 9.0
+  let delta = f - 32.0
+  delta * 5.0 / 9.0
 }
```

Without `-diff`, `rename` writes each file it changes and prints how many names it changed in each.
Renaming an export changes it in every module of the workspace that imports it. A rename changes
only names: renaming `celsius` would leave its doc comment as it is. A record entry written short
keeps its field, so renaming the variable `value` turns `{ value }` into `{ value: amount }`.

### JSON

With `-json`, every answer is one JSON object. A location has the fields of an `ant check -json`
error without the message: `path`, `line`, `column`, and `start` and `end`, with `end` just past the
name:

```sh
ant code refs -json greet.ant:1:name
```

<!-- test: cmd ant code refs -json greet.ant:1:name -->

```ant [greet.ant]
func greet(name: string): string {
  "Hello, ${name}"
}

println(greet("Ann"))
```

```text
{"declaration":{"path":"greet.ant","line":1,"column":12,"start":{"line":1,"column":12},"end":{"line":1,"column":16}},"references":[{"location":{"path":"greet.ant","line":2,"column":13,"start":{"line":2,"column":13},"end":{"line":2,"column":17}},"alias":null}],"complete":true,"skipped":[]}
```

A request that fails prints `{"error": {"code": ..., "message": ...}}` instead, with a code such as
`not-found`, `ambiguous`, `refused`, or `check`.

## Applications

### Rename a type safely

Check who uses a name, review the rename, apply it, and check the result:

```sh
ant code refs shop/money.ant:Money            # who uses Money?
ant code rename -diff shop/money.ant:Money Amount
ant code rename shop/money.ant:Money Amount
ant check . && ant test .
```

### Read an inferred type in a script

`jq` picks fields out of the JSON:

```sh
ant code type -json words.ant:2:groups | jq -r .text
```

## Possible issues

- `rename` refuses a field of a record type, a name of the standard library, a namespace, and a new
  name that is a keyword or that a scope already binds. It never falls back to replacing text.
- `rename` refuses when a program it changes has compile errors, since the checker may miss a use it
  cannot type. Fix the errors first, or add `-force` to rename anyway.
- A file of the workspace that does not parse or cannot be read may import the declaration. `refs`
  then still prints what it found but names the file and exits with status 2, and `rename` refuses
  unless you add `-force`.
- A rename covers the workspace only. A module outside the root that imports the declaration, and a
  name kept in a string, stay as they are.
- `rename` writes nothing if a file changed after it read it. Run it again.
- A rename can make a line longer than `ant fmt` keeps it. Run `ant fmt` after.

## See also

[`ant check`](check.md), [`ant doc`](doc.md), [`ant fmt`](fmt.md), [`ant test`](test.md)
