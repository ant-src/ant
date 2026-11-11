<!-- status: reviewed -->

# null

`null` is the only value of type `null`. It marks an absent value. `null` is a keyword.

## Details

- A type holds `null` only when it says so: `string | null` can hold `null`, `string` cannot. There
  is no `Option` wrapper and no `undefined`.

## Basic examples

The literal, and a test for it:

```ant
null         // null
null == null // true
```

A nullable variable, and `??` for a fallback:

```ant
let nickname: string | null = null
nickname ?? "anonymous" // "anonymous"
```

## Scope

### Function results

A function that may find nothing returns `T | null`. Here `find` gives `null` when no user matches,
and `?.` passes it on:

```ant
type User = { id: int, name: string }

func findUser(users: Array[User], id: int): User | null {
  users.find({ x.id == id })
}

let users = [User{ id: 1, name: "Ann" }, User{ id: 2, name: "Bob" }]
println(findUser(users, 2)?.name)
println(findUser(users, 7)?.name)
```

```text
Bob
null
```

### Narrowing

After a test against `null`, the checker knows the variable is not `null`:

```ant
func greet(name: string | null): string {
  if name == null { return "Hello, stranger" }
  "Hello, ${name.trim()}"
}

println(greet(null))
println(greet(" Ann "))
```

```text
Hello, stranger
Hello, Ann
```

A test narrows both branches of an `if`: after `name != null`, the `then` branch sees a `string` and
the `else` branch sees `null`. `name is null` narrows the same way as `name == null`:

```ant
func greet(name: string | null): string {
  if name != null { name.trim() } else { "stranger" }
}

greet(" Ann ") // "Ann"
greet(null)    // "stranger"
```

### Default arguments and record fields

`null` is a common default for an optional parameter or field. A field never narrows, so `label`
copies `c.email` into a `let` and tests the copy:

```ant
type Contact = { name: string, email: string | null = null }

func label(c: Contact, prefix: string | null = null): string {
  let name = if prefix == null { c.name } else { "${prefix} ${c.name}" }
  let email = c.email
  if email == null { name } else { "${name} <${email}>" }
}

println(label(Contact{ name: "Ann" }))
println(label(Contact{ name: "Bob", email: "bob@example.com" }, prefix: "Dr."))
```

```text
Ann
Dr. Bob <bob@example.com>
```

### Collections

An array, a map, or a set can hold `null` when its element type includes it:

```ant
let scores = [90, null, 75]
scores                          // [90, null, 75]
scores.filter({ x != null })    // [90, 75]
Map{ (null): "none", 1: "one" } // Map{ null: "none", 1: "one" }
Set.of([null, 1])               // Set.of([null, 1])
```

### Rendering

`null` renders as `null` in a string and when printed:

```ant
let middle: string | null = null
"middle: ${middle}" // "middle: null"
String.from(null)   // "null"
```

## Possible issues

### A missing key and a stored `null` look the same

Indexing a map gives `null` both for a missing key and for a key that holds `null`. Ask the map with
`has`:

```ant
let settings = Map{ "theme": null, "lang": "en" }
settings["theme"]     // null
settings["font"]      // null
settings.has("theme") // true
settings.has("font")  // false
```

### A plain type cannot hold `null`

```ant
let name: string = null // error: cannot use null where string is expected
```

### A nullable value needs narrowing first

A field or function of `T` is not available on `T | null`:

```ant
let name: string | null = "Ann"
let size = name.length()
// error: string | null has no field or function length on every member: narrow the value first
```

Test it, or use `?.` or `??`:

```ant
let name: string | null = "Ann"
name?.length()        // 3
(name ?? "").length() // 3
```

### A field is not narrowed

Narrowing works only for a local variable, a parameter, or a top-level `let`. A test on a field such
as `c.email` does not narrow it:

```ant
type Contact = { name: string, email: string | null }

func emailLength(c: Contact): int {
  if c.email == null { return 0 }
  c.email.length()
  // error: string | null has no field or function length on every member: narrow the value first
}
```

Copy the field into a `let` and test the copy, or use `?.`:

```ant
type Contact = { name: string, email: string | null }

func emailLength(c: Contact): int {
  let email = c.email
  if email == null { return 0 }
  email.length()
}

func emailLength2(c: Contact): int {
  c.email?.length() ?? 0
}

println(emailLength(Contact{ name: "Ann", email: "ann@example.com" }))
println(emailLength2(Contact{ name: "Bob", email: null }))
```

```text
15
0
```

### Testing a non-nullable value against `null`

`==` needs one side's type to be assignable to the other's, so a value that cannot be `null` cannot
be compared with `null`:

```ant
let s = "a"
if s == null { println("empty") }
// error: string and null cannot be compared with ==: neither is assignable to the other
```

### `var x = null` has type `null`

Without an annotation, the variable can hold nothing but `null`:

```ant
var cached = null
cached = "data" // error: cannot use string where null is expected
```

Annotate the nullable type:

```ant
var cached: string | null = null
cached = "data"
cached // "data"
```

### `??` on a value that cannot be `null`

`??` and `?.` need a nullable left side; on any other type they are dead code:

```ant
let name = "Ann"
let shown = name ?? "anonymous" // error: ?? needs a nullable left side, got string
```

### A bare `null` key is a string

```ant
Map{ null: 1 }   // Map{ "null": 1 }
Map{ (null): 1 } // Map{ null: 1 }
```

### `null` is not a variable name

```ant
let null = 0 // error: 'null' cannot be a variable name
```

## See also

- [Null](../../guide/null.md)
- [?? (null coalescing)](../operators/null-coalescing.md)
- [?. (optional chaining)](../operators/optional-chaining.md)
- [Equality operators](../operators/equality.md)
- [is](../operators/is.md)
- [void](../types/void.md)
- [Map.has](../../lib/Map/has.md)
- [true and false](booleans.md)
