<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Html

```ant signature
type Html
```

Html is the value of an HTML expression, \<p\>\{name\}\</p\>: an immutable list of nodes, each an
element, a text node, or raw markup from [`Html.unsafe`](unsafe.md). It compares by its nodes with
==, and String.from renders it as markup, escaping text and attribute values.

<!-- end generated -->

An `Html` value is markup that a program builds with HTML expressions, compares, and renders. The
`Html` namespace holds the two functions that work on it: one reads its nodes, and one turns a
trusted string into markup. `Html` is also the type, for a component's return type or a record
field.

## Details

- `Html` has no constructor. An HTML expression makes it, such as `<p>{name}</p>` or the fragment
  `<>...</>`, and [`Html.unsafe`](unsafe.md) makes it from a string of trusted markup.
- A value is kept in one normal form: fragments are spliced in, `void` leaves nothing, adjacent text
  is joined, and empty text is dropped. So two values built differently from the same elements and
  text are `==`, and an `Html` works as a map key.
- `String.from` renders the markup, and so do `println` and `${...}` in a string. Text and attribute
  values are escaped on the way out.
- `len`, indexing, and `for` do not take `Html`. Read its nodes with [`Html.nodes`](nodes.md).
  `Json.stringify` panics on it.

## Basic examples

An HTML expression is a value of type `Html`. Interpolated text is escaped when it renders:

```ant
let name = "Tom & <Jerry>"
let greeting = <p class="greet">Hello, {name}</p>
println(greeting)
```

```text
<p class="greet">Hello, Tom &amp; &lt;Jerry&gt;</p>
```

`String.from` gives the markup as a string:

```ant
String.from(<b>bold</b>) // "<b>bold</b>"
String.from(<br />)      // "<br>"
```

A function that returns `Html` is a component:

```ant
func badge(count: int): Html {
  <span class="badge">{count}</span>
}

println(<h2>Inbox {badge(3)}</h2>)
```

```text
<h2>Inbox <span class="badge">3</span></h2>
```

## Scope

The empty fragment is the empty `Html`. It renders as nothing, and as `<></>` inside a container:

```ant
String.from(<></>) // ""
[<></>, <b>x</b>]  // [<></>, <b>x</b>]
```

An array of `Html` inserts its elements in order, nothing between them:

```ant
let items = ["milk", "eggs"] | Array.map({ <li>{x}</li> })
String.from(<ul>{items}</ul>) // "<ul><li>milk</li><li>eggs</li></ul>"
```

Quotes and apostrophes in an attribute value are escaped:

```ant
String.from(<p title={"it's \"new\""}>x</p>) // "<p title=\"it&#39;s &quot;new&quot;\">x</p>"
```

A string embeds the rendered markup, which is how a page gets its doctype:

```ant
let page = <html><body>Hi</body></html>
"<!doctype html>${page}" // "<!doctype html><html><body>Hi</body></html>"
```

`Html` is a map key, and two values meaning the same find one entry:

```ant
let hits: Map[Html, int] = Map{}
hits[<b>x</b>] = 1
hits[<b>{"x"}</b>] // 1
```

A record field can hold `Html`:

```ant
type Section = { title: string, body: Html }

let intro = Section{ title: "Intro", body: <p>Ants live in colonies.</p> }
println(<section><h2>{intro.title}</h2>{intro.body}</section>)
```

```text
<section><h2>Intro</h2><p>Ants live in colonies.</p></section>
```

## Functions and types by job

### Making markup from a string

- [Html.unsafe](unsafe.md): a string of trusted markup as `Html`, rendered unescaped

### Reading the nodes

- [Html.nodes](nodes.md): the top-level nodes, in order
- [Html.Node](types/Node.md): one node, an `Html.Element`, an `Html.Text`, or an `Html.Raw`
- [Html.Attribute](types/Attribute.md): one attribute of an element, a name and a value

### Spreading attributes

- [Html.Attributes](types/Attributes.md): the map an attribute spread `{...attrs}` takes

## Applications

### Test a component

Compare a component's output with the HTML expression it should give. The `$` is literal text, and
`{whole}` interpolates:

```ant
func price(cents: int): Html {
  let whole = cents / 100
  let rest = String.from(cents % 100).padStart(2, "0")
  <span class="price">${whole}.{rest}</span>
}

assert(price(1250) == <span class="price">$12.50</span>)
assert(price(5) == <span class="price">$0.05</span>)
```

### Collect the text of a tree

`Html.nodes` reads one level; a recursive function reads the whole tree:

```ant
func nodeText(node: Html.Node): string {
  if node is Html.Text { return node.text }
  if node is Html.Element { return textOf(node.children) }
  ""
}

func textOf(h: Html): string {
  Html.nodes(h) | Array.map(nodeText) | String.join("")
}

println(textOf(<p>Ants &amp; <em>bees</em> work.</p>))
```

```text
Ants & bees work.
```

### Count the elements

```ant
func countElements(h: Html): int {
  var count = 0
  for node in Html.nodes(h) {
    if node is Html.Element {
      count += 1 + countElements(node.children)
    }
  }
  count
}

println(countElements(<article><h1>Title</h1><p>One <em>two</em></p></article>))
```

```text
4
```

## Properties and relations

The normal form makes these pairs equal:

```ant
assert(<p>{"a"}{"b"}</p> == <p>ab</p>)
assert(<p>{<>x</>}</p> == <p>x</p>)
assert(<p>{""}</p> == <p></p>)
assert(<i>{if false { "hidden" }}</i> == <i></i>)
```

Rendering text escapes it, and `Html.nodes` gives it back decoded:

```ant
let text = "1 < 2 & 3 > 2"
let h = <>{text}</>
assert(String.from(h) == "1 &lt; 2 &amp; 3 &gt; 2")
assert(Html.nodes(h)[0] == Html.Text{ text: text })
```

## Possible issues

### `Html` is not a collection

`len`, `for`, and indexing do not take `Html`:

```ant
let h = <><li>a</li><li>b</li></>
let n = len(h)
// error: len needs an array, a map, a set, a string, Bytes, or a channel, got Html
for x in h {
  // error: for needs an array, a map, a set, a string, Bytes, or a sequence, got Html
  println(x)
}
let first = h[0] // error: Html cannot be indexed
```

Read the nodes as an array with `Html.nodes`:

```ant
let h = <><li>a</li><li>b</li></>
Html.nodes(h).length() // 2
Html.nodes(h)[0]       // Html.Element{ name: "li", attributes: [], children: a }
```

### A string is not markup

A string never converts to `Html`. Interpolate it as text, or pass trusted markup through
`Html.unsafe`:

```ant
let h: Html = "<b>x</b>" // error: cannot use string where Html is expected
```

### A node record is not markup

`Html.nodes` returns copies as records, and no function turns a record back into `Html`. An
`Html.Element` in content is a compile error:

```ant
let e = Html.Element{ name: "b", attributes: [], children: <>x</> }
let page = <div>{e}</div>
// error: markup takes a string, a number, void, Html, or an array of them, got Html.Element
```

### Raw markup is not equal to the markup it spells

A raw node from `Html.unsafe` renders as the markup it holds, but it is not `==` to an HTML
expression that builds the same markup:

```ant
let raw = Html.unsafe("<b>x</b>")
String.from(raw) == String.from(<b>x</b>) // true
raw == <b>x</b>                           // false
```

Compare rendered strings when a value holds raw markup.

### `Json.stringify` panics

JSON has no form for `Html`. Render it with `String.from` and put the string in the JSON:

```ant
Json.stringify(<p>hi</p>)
// panic: Json.stringify: Html has no JSON form; write String.from of it
```

```ant
Json.stringify(String.from(<p>hi</p>)) // "\"<p>hi</p>\""
```

## See also

- [Html.unsafe](unsafe.md): trusted markup from a string
- [Html.nodes](nodes.md): the top-level nodes of a value
- [String.from](../String/from.md): render a value as a string
- [HTML expressions](../../language/literals/html.md): the syntax that builds `Html`
- [HTML expressions guide](../../guide/html.md): how the parts fit together
- [Render HTML pages](../../howto/network/html-pages.md): serve markup from an HTTP handler
