<!-- status: reviewed -->

# Records

A record groups named fields under a type name. Declare the type with `type`, then build values with
a literal that names it.

```ant
type Point = { x: int, y: int }

type Server = {
  addr: string,
  workers: int = 4,
  tags: Array[string] = [],
}

let p = Point{ x: 1, y: 2 }
let moved = Point{ ...p, x: 10 }
println(p)
println(moved)
let x = 3
let y = 4
println(Point{ x, y })

let local = Server{ addr: ":8080" }
let busy = Server{ addr: ":9090", workers: 16, tags: ["prod"] }
println(local)
println(busy.workers)

let { addr, workers: w } = busy
println("${addr} runs ${w} workers")

let alias = p
alias.y = 20
println(p.y)
println(p == Point{ x: 1, y: 20 })
println(Record.same(p, alias))

func describe(s: { addr: string }): string { "server at ${s.addr}" }
println(describe(local))

type UserId = string
let id = UserId("u-1")
println(string(id))
```

```text
Point{ x: 1, y: 2 }
Point{ x: 10, y: 2 }
Point{ x: 3, y: 4 }
Server{ addr: ":8080", workers: 4, tags: [] }
16
:9090 runs 16 workers
20
true
true
server at :8080
u-1
```

`Point{ ...p, x: 10 }` copies every field of `p` into a new record, and the later `x` wins. A bare
name in a literal, `Point{ x, y }`, is short for `Point{ x: x, y: y }`, the same shorthand
destructuring uses.

`Server` gives two of its fields a default after `=`. A literal may leave those out, and each one it
leaves out gets its default: `local` has 4 workers and its own empty `tags` array. `addr` has no
default, so every literal must give it.

`let { addr, workers: w } = busy` destructures the record: it binds the field `addr` under its own
name and the field `workers` as `w`. Fields you do not name are ignored.

Fields are mutable, and a record is shared, not copied: `alias` and `p` are one record, so changing
`alias.y` changes `p.y`. `let` only stops you from reassigning the name. `==` compares the type and
the fields, and [`Record.same`](../lib/Record/same.md) tells whether two names hold the same record.

`describe` takes `{ addr: string }`, an anonymous record type, so it accepts any record with an
`addr` field, such as a `Server`.

`type UserId = string` makes a new type over `string`. `UserId("u-1")` builds one, and `string(id)`
gives the string back. A `UserId` goes only where a `UserId` is expected, not where a `string` is.

The type checker catches a missing field and a misspelled one:

```ant
type Server = { addr: string, workers: int = 4 }

let a = Server{ workers: 8 }             // error: Server literal is missing the field addr
let b = Server{ addr: ":80", worker: 8 } // error: Server has no field worker
```

Records in depth are in the [Records](../guide/records.md) chapter of the guide; see also
[record literals](../language/literals/records.md) and [`type`](../language/keywords/type.md). Next:
[unions](11-unions.md).
