<!-- status: reviewed -->

# Keywords and reserved words

Keywords are the words of Ant's grammar:

```text
break continue defer else export false for func go if import
in interface is let null return select true try type var void while
```

Reserved words are kept free so that they stay available and so that habits from JavaScript produce
an error instead of an identifier. `Self` is the one with a use today: it names the satisfying type
inside an `interface`.

```text
as async await catch class const enum finally function
match Self switch this throw undefined
```

## Details

- A keyword or reserved word cannot name a variable, a function, a type, or a parameter.
- A field can have any name. After `.` or `?.`, and before `:` in a record type, a record or map
  literal, or a record destructuring, a keyword or reserved word is a plain name.
- `Self` has one use: in the signatures of an `interface`, it names the type that satisfies it.
- In a test file, `test` is a keyword as well. Everywhere else it is a plain name.
- Identifiers are case-sensitive, so `Class` and `Type` are ordinary names.

## Basic examples

A keyword cannot name a variable:

```ant
let type = "click" // error: 'type' cannot be a variable name
```

A reserved word cannot either:

```ant
let class = "button" // error: 'class' cannot be a variable name
```

A keyword works as a field name:

```ant
type Event = { type: string, target: string }
let event = Event{ type: "click", target: "save" }
event.type // "click"
```

## Scope

### Field access

After `.` and `?.`, a keyword or reserved word is a field name:

```ant
type Label = { class: string, for: string }
let label: Label | null = Label{ class: "hint", for: "email" }
label?.class // "hint"
label?.for   // "email"
```

### Map keys

Before `:` in a map literal, a keyword is a key. The key of `null:` is the string `"null"`:

```ant
let flags = Map{ class: "wide", null: "empty", true: "yes" }
flags.keys()  // ["class", "null", "true"]
flags["null"] // "empty"
```

### Destructuring

A record destructuring takes a field named by a keyword when it binds the field to another name:

```ant
type Event = { type: string, target: string }
let event = Event{ type: "submit", target: "form" }
let { type: kind, target } = event
"${kind} on ${target}" // "submit on form"
```

### `type` as a declaration and as an expression

The token after `type` decides what it begins. A name begins a type declaration, `(` begins a type
expression:

```ant
type Point = { x: int, y: int }
type(Point) // Point
```

### `test` outside a test file

Outside a test file, `test` is a plain name:

```ant
let test = "ab" + "c"
test.length() // 3
```

## Applications

### Records that mirror JSON

Data from JavaScript often has fields named `type` or `class`. A record type spells them as they
are:

```ant
type Message = { type: string, text: string }

let messages = [
  Message{ type: "info", text: "saved" },
  Message{ type: "error", text: "disk full" },
  Message{ type: "error", text: "timeout" },
]
let errors = messages | Array.filter({ x.type == "error" }) | Array.map({ x.text })
println(errors | String.join(", "))
```

```text
disk full, timeout
```

## Possible issues

### Destructuring shorthand

The shorthand `{ type }` would bind a variable named `type`, which cannot exist:

```ant
type Event = { type: string }
let event = Event{ type: "click" }
let { type } = event // error: 'type' is a keyword and cannot be bound: write { type: name }
```

Name the variable: `let { type: kind } = event`.

### Map shorthand

`Map{ type }` would read a variable named `type`. Write the key and the value:

```ant
let kind = "click"
let event = Map{ type } // error: 'type' is a keyword: as a key it needs ':'
```

`Map{ type: kind }` builds the map.

### Parameter names

A keyword cannot name a parameter, so it cannot name a named argument either:

```ant
func render(type: string): string { "<${type}>" }
// error: 'type' is a keyword and cannot be a parameter name
```

Pick another word, such as `kind` or `tag`.

### JavaScript habits

`const`, `function`, `as`, and `undefined` are reserved, so each is an error where JavaScript would
accept it:

```ant
const limit = 10 // error: 'const' is a reserved word
```

```ant
let n = 1
let count = n as int // error: 'as' is a reserved word
```

```ant
let missing = undefined // error: 'undefined' is a reserved word
```

Ant writes `let` for `const`, `func` for `function`, and `null` for `undefined`.

### `Self` outside an interface

```ant
let x: Self = 1 // error: 'Self' is a reserved word
```

### Prelude names

`Array`, `Map`, `println`, `panic`, and the other standard library names are ordinary identifiers,
not keywords. The prelude binds them, and no declaration may take one:

```ant
let println = "hi" // error: println is a prelude name and cannot be redeclared
```

## See also

[Identifiers](identifiers.md), [type](../keywords/type.md), [interface](../keywords/interface.md),
[Self](../types/Self.md), [test](../keywords/test.md), [Map literals](../literals/maps.md),
[Record literals](../literals/records.md), [null](../literals/null.md)
