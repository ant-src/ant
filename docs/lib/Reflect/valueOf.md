<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.valueOf

```ant signature
func Reflect.valueOf[T](t: Reflect.Type[T], value: T): Reflect.Value
```

Reflect.valueOf boxes value with its type t. When t is a union, the box's type is the member value
belongs to, found from its kind or tag. The type comes first so that it fixes T before the value is
checked against it: Reflect.valueOf(type(Array\[int\]), \[\]) checks.

<!-- end generated -->

## Details

- The box's type is never a union or an alias. At a union it is the member the value belongs to, and
  at an alias it is the type the alias stands for.
- At an anonymous record type the box's type is that anonymous type, not the record's own named
  type, so readers see only the fields it lists.
- The box holds the value itself, not a copy. A later change to a boxed array, map, set, or record
  shows through the box.

## Basic examples

Box a value with its type:

```ant
Reflect.valueOf(type(int), 42)        // Reflect.Value(42: int)
Reflect.valueOf(type(string), "ant")  // Reflect.Value("ant": string)
Reflect.valueOf(type(Array[int]), []) // Reflect.Value([]: Array[int])
```

At a union, the box takes the member the value belongs to:

```ant
Reflect.valueOf(type(int | null), 42)   // Reflect.Value(42: int)
Reflect.valueOf(type(int | null), null) // Reflect.Value(null: null)
```

## Scope

### Numbers

Each number type is its own type, so the same digits box differently:

```ant
Reflect.valueOf(type(int), 1)                // Reflect.Value(1: int)
Reflect.valueOf(type(byte), byte(1))         // Reflect.Value(1: byte)
Reflect.valueOf(type(float64), 1.0)          // Reflect.Value(1.0: float64)
Reflect.valueOf(type(float32), float32(1.5)) // Reflect.Value(1.5: float32)
```

### Containers

The type argument fixes the element type, so an empty literal or one with fewer member types checks:

```ant
Reflect.valueOf(type(Array[int | null]), [1, 2])       // Reflect.Value([1, 2]: Array[int | null])
Reflect.valueOf(type(Map[string, int]), Map{ "a": 1 }) // Reflect.Value(Map{ "a": 1 }: Map[string, int])
Reflect.valueOf(type(Set[string]), Set.of(["a"]))      // Reflect.Value(Set.of(["a"]): Set[string])
Reflect.valueOf(type(Map[string, int]), Map{})         // Reflect.Value(Map{}: Map[string, int])
```

### Records and named primitive types

```ant
type User = { name: string, age: int }
type UserId = int
Reflect.valueOf(type(User), User{ name: "Ann", age: 31 }) // Reflect.Value(User{ name: "Ann", age: 31 }: User)
Reflect.valueOf(type(UserId), UserId(7))                  // Reflect.Value(UserId(7): UserId)
```

### Tagged unions

A union of records resolves by tag:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
let box = Reflect.valueOf(type(Shape), Circle{ r: 1.0 })
box               // Reflect.Value(Circle{ r: 1.0 }: Circle)
Reflect.type(box) // Circle
```

`Result` is a union of the `Ok` and `Err` records and resolves the same way:

```ant
Reflect.valueOf(type(Result[int, Error]), Ok{ value: 1 }) // Reflect.Value(Ok{ value: 1 }: Ok[int])
```

### Aliases

An alias expands to the type it stands for, here a member of the recursive `Json.Value`:

```ant
let doc = Json.parse(`{"a": [1]}`).unwrap()
let box = Reflect.valueOf(type(Json.Value), doc)
Reflect.type(box) // Map[string, Json.Value]
```

### Anonymous record types

The box keeps the anonymous type it was given. The value is still the whole `User`, but readers
reach only `name`:

```ant
type User = { name: string, age: int }
let box = Reflect.valueOf(type({ name: string }), User{ name: "Ann", age: 31 })
box                        // Reflect.Value(User{ name: "Ann", age: 31 }: { name: string })
Reflect.field(box, "name") // Reflect.Value("Ann": string)
```

### Function values

A function value has no type of its own at run time, so the type you write is the one it gets:

```ant
Reflect.valueOf(type(func(int): int), { x * 2 }) // Reflect.Value(func: func(int): int)
```

### `void`

```ant
func reset() {}
Reflect.valueOf(type(void), reset()) // Reflect.Value(void: void)
```

## Applications

### An entry point for a generic encoder

A walker works on boxes, one level at a time. A generic function takes the caller's type value and
the value together, boxes them once, and hands the box to the walker:

```ant
type Person = { name: string, age: int, tags: Array[string], nick: string | null }

func walk(d: Reflect.Value): string {
  let s = Reflect.shape(Reflect.type(d))
  if s is Reflect.Record {
    let parts = s.fields | Array.map({ "${x.name}=${walk(Reflect.field(d, x.name))}" })
    return "{${parts | String.join(" ")}}"
  }
  if s is Reflect.Array {
    let parts = Reflect.items(d) | Array.map(walk)
    return "[${parts | String.join(",")}]"
  }
  String.from(Reflect.scalar(d))
}

func render[T](t: Reflect.Type[T], value: T): string {
  walk(Reflect.valueOf(t, value))
}

let ann = Person{ name: "Ann", age: 31, tags: ["admin", "dev"], nick: null }
println(render(type(Person), ann))
println(render(type(Array[int | null]), [1, null]))
```

```text
{name=Ann age=31 tags=[admin,dev] nick=null}
[1,null]
```

### Values of different types in one collection

A `Reflect.Value` holds any type, so one array keeps values that share no static type, each with its
type:

```ant
type Port = int
let settings = [
  Reflect.valueOf(type(string), "localhost"),
  Reflect.valueOf(type(Port), Port(8080)),
  Reflect.valueOf(type(Array[string]), ["a", "b"]),
]
for d in settings {
  println("${Reflect.type(d)}: ${d}")
}
```

```text
string: Reflect.Value("localhost": string)
Port: Reflect.Value(Port(8080): Port)
Array[string]: Reflect.Value(["a", "b"]: Array[string])
```

## Properties and relations

[`Reflect.to`](to.md) at the same type gives the value back:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
assert(Reflect.to(Reflect.valueOf(type(Point), p), type(Point)) == p)
```

The box's type is `t` erased, unless `t` is a union or an alias:

```ant
assert(Reflect.type(Reflect.valueOf(type(Array[int]), [1])) == Reflect.erase(type(Array[int])))
assert(Reflect.type(Reflect.valueOf(type(int | null), 1)) == Reflect.erase(type(int)))
```

Boxing at a union gives the same box as boxing at the member:

```ant
assert(Reflect.valueOf(type(int | string), 1) == Reflect.valueOf(type(int), 1))
```

Two boxes are `==` when their types and values are, so the same digits at two number types differ:

```ant
assert(Reflect.valueOf(type(int), 1) != Reflect.valueOf(type(int8), int8(1)))
```

The box holds the value itself, so a later change shows through it:

```ant
let xs = [1, 2]
let box = Reflect.valueOf(type(Array[int]), xs)
xs.push(3)
assert(Reflect.items(box).length() == 3)
assert(Reflect.same(box, Reflect.valueOf(type(Array[int]), xs)))
```

## Possible issues

A number literal keeps its own type. An integer literal is an `int`, even where `t` is a float type:

```ant
let half = Reflect.valueOf(type(float64), 1)
// error: cannot use int where float64 is expected
```

A container's type arguments must match exactly. An `Array[int | null]` is not an `Array[int]`:

```ant
let xs: Array[int | null] = [1, 2]
let box = Reflect.valueOf(type(Array[int]), xs)
// error: cannot use Array[int | null] where Array[int] is expected
```

`type(T)` cannot name a type parameter, so a generic function cannot box its argument by itself.
Take a `Reflect.Type[T]` parameter, as `render` does above:

```ant
func box[T](x: T): Reflect.Value {
  Reflect.valueOf(type(T), x)
  // error: type(...) cannot mention T, a type parameter of the enclosing function or type, since nothing is bound to it at run time; take a Reflect.Type[T] parameter and build the type from it with Reflect.arrayOf, Reflect.setOf, Reflect.mapOf, or Reflect.nullable
}
```

To box a container of `T`, build its type from the parameter:

```ant
func boxAll[T](t: Reflect.Type[T], xs: Array[T]): Reflect.Value {
  Reflect.valueOf(Reflect.arrayOf(t), xs)
}
boxAll(type(int), [1, 2]) // Reflect.Value([1, 2]: Array[int])
```

A box made at an anonymous record type does not unbox at the record's own named type:

```ant
type User = { name: string, age: int }
let box = Reflect.valueOf(type({ name: string }), User{ name: "Ann", age: 31 })
Reflect.to(box, type(User)) // panic: Reflect.to: { name: string } is not assignable to User
```

## See also

[Reflect.to](to.md) · [Reflect.type](type.md) · [Reflect.makeScalar](makeScalar.md) ·
[Reflect.erase](erase.md) · [Reflect.shape](shape.md) · [Reflect.field](field.md) ·
[Reflect.items](items.md) · [Reflect.same](same.md) · [Reflect.Value](types/Value.md) ·
[Reflect.Type](types/Type.md)
