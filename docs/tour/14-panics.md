<!-- status: reviewed -->

# Panics

A `Result` is for failures a caller expects and handles. A panic is for bugs: a state the program
believed impossible. It stops the goroutine on the spot. This program settles a few bills, and one
of them is broken.

```ant
type Bill = { name: string, total: int, people: int }

func share(bill: Bill): int {
  assert(bill.people > 0, "a bill needs someone to pay it")
  bill.total / bill.people
}

func settle(bill: Bill) {
  println("open ${bill.name}")
  defer { println("filed ${bill.name}") }
  defer { println("close ${bill.name}") }
  println("each pays ${share(bill)}")
}

let bills = [
  Bill{ name: "dinner", total: 90, people: 3 },
  Bill{ name: "lunch", total: 40, people: 0 },
  Bill{ name: "coffee", total: 12, people: 4 },
]

for bill in bills {
  let result = Panic.catch({ settle(bill) })
  if result is Err {
    println("${bill.name} failed: ${result.error.message}")
  }
}
```

```text
open dinner
each pays 30
close dinner
filed dinner
open lunch
close lunch
filed lunch
lunch failed: assertion failed: bill.people > 0 (0 > 0): a bill needs someone to pay it
open coffee
each pays 3
close coffee
filed coffee
```

`assert` states what `share` believes about its input. When the condition is `false`, it panics with
a message that quotes the condition, the values it compared, and your reason. The reason is
optional: `assert(bill.total >= 0)` works too. `assert` is always on; no flag turns it off.

`defer` schedules a block to run when the enclosing block exits: at its end, by `return`, `break`,
or `continue`, or by a panic. Only `Os.exit` skips deferred blocks. The lunch bill panics before
`settle` prints its share, and `close lunch` still appears. `settle` defers two blocks, and they run
in reverse order: `close` before `filed`.

`Panic.catch` calls a function and returns a `Result`: `Ok` with the function's value, or `Err` with
the panic, once the deferred blocks inside have run. The panic is a record with a `message` and a
`trace`. Catching at the edge of each bill keeps one broken bill from stopping the others.

Without `Panic.catch`, a panic in the main goroutine ends the program. Many operations panic on
their own, such as an index past the end of an array:

```ant
func pick(names: Array[string], i: int): string {
  defer { println("pick ${i} done") }
  names[i] // panic: index 2 out of range for an array of length 2
}

let names = ["Ann", "Bob"]
println(pick(names, 1))
println(pick(names, 2))
println("never printed")
```

```text
pick 1 done
Bob
pick 2 done
```

The deferred block runs as the panic leaves `pick`, and then `ant run` writes the message and a
stack trace to standard error and exits with a non-zero status:

```text
panic: index 2 out of range for an array of length 2
  at pick (main.ant:3:8)
  at module main.ant (main.ant:8:13)
```

Integer overflow, integer division by zero, and a lossy numeric conversion panic the same way. Call
`panic` yourself to stop on a state the program believes impossible:

```ant
func price(menu: Map[string, int], item: string): int {
  // Callers ask only for items on the menu, so a missing price is a bug: panic.
  menu[item] ?? panic("no price for ${item}") // panic: no price for soup
}

let menu = Map{ tea: 3, cake: 5 }
println(price(menu, "tea"))
println(price(menu, "soup"))
```

```text
3
```

If the program can recover by doing something else, return a `Result`; if the only fix is to change
the code, panic.

The guide chapter [Errors](../guide/errors.md) covers when to pick a `Result` and when to panic. The
reference has pages for [`panic`](../lib/prelude/panic.md), [`assert`](../lib/prelude/assert.md),
[`defer`](../language/keywords/defer.md), [`Panic.catch`](../lib/Panic/catch.md), and
[`Panic.raise`](../lib/Panic/raise.md), which lets a panic you caught go on.

Next: [Generics](15-generics.md).
