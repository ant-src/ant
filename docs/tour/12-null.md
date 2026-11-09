<!-- status: reviewed -->

# Null

A value is `null` only where its type says so. `User | null` is a user or nothing; a plain `User` is
always a user.

```ant
type Address = { city: string }
type User = { id: int, name: string, address: Address | null }

let users = [
  User{ id: 1, name: "Ann", address: Address{ city: "Oslo" } },
  User{ id: 2, name: "Bob", address: null },
]

func find(id: int): User | null {
  users.find({ x.id == id })
}

func greet(id: int, user: User | null): string {
  if user == null {
    return "no user ${id}"
  }
  "hello, ${user.name}"
}

for id in [1, 2, 3] {
  let user = find(id)
  let city = user?.address?.city ?? "unknown"
  println("${greet(id, user)} from ${city}")
}

let ages = Map{ "Ann": 31 }
let age = ages["Bob"] ?? 0
println(age + 1)
```

```text
hello, Ann from Oslo
hello, Bob from unknown
no user 3 from unknown
1
```

`find` returns `User | null`, as [`Array.find`](../lib/Array/find.md) does when nothing matches. In
`greet`, the test `user == null` ends in `return`, so below it the type checker knows `user` is a
`User` and lets you read `user.name`. This is narrowing: after a test on a local or a parameter, its
type shrinks to what the test proved.

`?.` reads a field only when the left side is not null, and gives `null` otherwise, so
`user?.address?.city` is a `string | null`. `??` replaces `null` with the value on its right, here
`"unknown"`, and the result is a plain `string`. Reading a missing key from a map gives `null` too,
and `?? 0` turns it into an `int` you can add to.

Skip the test and the type checker stops you:

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }

func find(id: int): User | null {
  null
}

let user = find(1)
println(user.name)      // error: User | null has no field name: narrow the value first
let name: string = null // error: cannot use null where string is expected

func show(user: User) {
  if user.address != null {
    println(user.address.city) // error: Address | null has no field city: narrow the value first
  }
}
```

A field such as `user.address` does not narrow. Bind it to a `let` first, or use `?.` and `??`.

The guide chapter [Null](../guide/null.md) covers where narrowing holds and where it does not. The
operators have their own pages: [`?.`](../language/operators/optional-chaining.md) and
[`??`](../language/operators/null-coalescing.md).

Next: [Errors](13-errors.md).
