<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Html.Node

```ant signature
type Html.Node = Html.Element | Html.Text | Html.Raw
```

Html.Node is one node of an Html value, as [`Html.nodes`](../nodes.md) returns it.

```ant signature
type Html.Element = { name: string, attributes: Array[Html.Attribute], children: Html }
```

Html.Element is an element: its name as written, its attributes in order, and its children.

```ant signature
type Html.Text = { text: string }
```

Html.Text is a text node, holding its text decoded.

```ant signature
type Html.Raw = { markup: string }
```

Html.Raw is a node of trusted markup, made by [`Html.unsafe`](../unsafe.md).

<!-- end generated -->

## Details

- A value of type `Html` is read with [`Html.nodes`](../nodes.md), one level at a time; an element's
  `children` is an `Html`, read with `Html.nodes` again.
- Names and attributes are as written, attributes in their final order, a boolean attribute that is
  on has the empty value, and text is decoded: `&amp;` reads as `&`.
- The records are copies. Changing one does not change the `Html` it came from, and building one
  does not make markup: no function turns an `Html.Element` record into `Html`.
- A raw node from [`Html.unsafe`](../unsafe.md) is opaque: its markup is a string, not elements.
- `Html.Node` is a closed union. A chain of `is` tests that covers all three members needs no final
  `else`, and a member added later breaks it on purpose.

## Basic examples

`Html.nodes` returns the top-level nodes, each one member of `Html.Node`:

```ant
Html.nodes(<p>hi</p>)           // [Html.Element{ name: "p", attributes: [], children: hi }]
Html.nodes(<>hi</>)             // [Html.Text{ text: "hi" }]
Html.nodes(Html.unsafe("<hr>")) // [Html.Raw{ markup: "<hr>" }]
```

Narrow a node with `is` to read the fields of its member:

```ant
let link = <a href="/docs">Docs</a>
let node = Html.nodes(link)[0]
if node is Html.Element {
  println(node.name)
  println(node.attributes)
  println(node.children)
}
```

```text
a
[Html.Attribute{ name: "href", value: "/docs" }]
Docs
```

Handle every member with a chain of `is` tests:

```ant
func describe(n: Html.Node): string {
  if n is Html.Element { "element ${n.name}" }
  else if n is Html.Text { "text ${n.text}" }
  else if n is Html.Raw { "raw ${n.markup}" }
}

Html.nodes(<>a<br />{Html.unsafe("<hr>")}</>).map(describe) // ["text a", "element br", "raw <hr>"]
```

## Scope

### Elements

An element's children are an `Html`, which `Html.nodes` reads one level further down:

```ant
let list = <ul><li>one</li><li>two</li></ul>
let ul = Html.nodes(list)[0]
if ul is Html.Element {
  for item in Html.nodes(ul.children) {
    println(item)
  }
}
```

```text
Html.Element{ name: "li", attributes: [], children: one }
Html.Element{ name: "li", attributes: [], children: two }
```

A line break between tags becomes one space of text, so a list written over several lines has text
nodes between its items. Skip them with an `is` test, or `filter({ x is Html.Element })`, when you
want only the elements:

```ant
let list = <ul>
  <li>one</li>
  <li>two</li>
</ul>
let ul = Html.nodes(list)[0]
if ul is Html.Element {
  for item in Html.nodes(ul.children) {
    println(item)
  }
}
```

```text
Html.Element{ name: "li", attributes: [], children: one }
Html.Text{ text: " " }
Html.Element{ name: "li", attributes: [], children: two }
```

Names keep their case as written, and a boolean attribute that is on reads as the empty value:

```ant
let done = true
let shapes = [<svg viewBox="0 0 8 8" />, <input type="checkbox" checked={done} disabled={!done} />]
for h in shapes {
  let node = Html.nodes(h)[0]
  if node is Html.Element {
    println("${node.name}: ${node.attributes}")
  }
}
```

```text
svg: [Html.Attribute{ name: "viewBox", value: "0 0 8 8" }]
input: [Html.Attribute{ name: "type", value: "checkbox" }, Html.Attribute{ name: "checked", value: "" }]
```

### Text

Text is decoded, and adjacent text joins into one node, whether written or interpolated:

```ant
Html.nodes(<>Tom &amp; Jerry</>) // [Html.Text{ text: "Tom & Jerry" }]
Html.nodes(<>{"a"}{"b"}c</>)     // [Html.Text{ text: "abc" }]
Html.nodes(<>{"<b>"}</>)         // [Html.Text{ text: "<b>" }]
Html.nodes(<>{42}</>)            // [Html.Text{ text: "42" }]
```

### Raw markup

A raw node holds its string as it is; the markup inside it is not parsed into elements:

```ant
Html.nodes(Html.unsafe("<b>x</b>")) // [Html.Raw{ markup: "<b>x</b>" }]
```

### An empty value

An empty `Html` has no nodes, and `void` or an empty text adds none:

```ant
Html.nodes(<></>)                       // []
Html.nodes(<>{""}{if false { "x" }}</>) // []
```

### Testing the member

A value of type `Html.Node` is exactly one member, which `is` tests:

```ant
let n: Html.Node = Html.Text{ text: "hi" }
n is Html.Text    // true
n is Html.Element // false
```

## Applications

### Text content

Collect the text of a tree, walking into every element:

```ant
func textOf(h: Html): string {
  Html.nodes(h)
    | Array.map(func(n: Html.Node): string {
      if n is Html.Element { textOf(n.children) }
      else if n is Html.Text { n.text }
      else { "" }
    })
    | String.join("")
}

println(textOf(<article><h1>Tom &amp; Jerry</h1><p>A <em>short</em> film.</p></article>))
```

```text
Tom & JerryA short film.
```

### Links on a page

Find the `href` of every `<a>`, at any depth:

```ant
func links(h: Html): Array[string] {
  Html.nodes(h) | Array.flatMap(func(n: Html.Node): Array[string] {
    if !(n is Html.Element) { return [] }
    let inner = links(n.children)
    if n.name != "a" { return inner }
    let href = n.attributes.find({ x.name == "href" })
    if href == null { inner } else { [href.value, ...inner] }
  })
}

let page = <nav>
  <a href="/">Home</a>
  <ul><li><a href="/docs">Docs</a></li><li><a>No link</a></li></ul>
</nav>
println(links(page))
```

```text
["/", "/docs"]
```

### Testing a component's structure

Check the shape of a component's output without comparing its whole markup:

```ant
func badge(count: int): Html {
  <span class="badge" title={"${count} unread"}>{count}</span>
}

let root = Html.nodes(badge(3))[0]
assert(root is Html.Element)
assert(root.name == "span")
assert(root.attributes.some({ x.name == "class" && x.value == "badge" }))
assert(root.children == <>3</>)
println("ok")
```

```text
ok
```

## Properties and relations

An element's `children` is the `Html` written inside it:

```ant
let body = <>Hello, <b>world</b></>
let node = Html.nodes(<p>{body}</p>)[0]
if node is Html.Element {
  assert(node.children == body)
}
```

Nodes compare by their fields, so a node equals the record that describes it:

```ant
let node = Html.nodes(<b>x</b>)[0]
assert(node == Html.Element{ name: "b", attributes: [], children: <>x</> })
assert(Html.nodes(<>x</>)[0] == Html.Text{ text: "x" })
```

## Possible issues

A field read on the union, before narrowing, is a compile error:

```ant
let n = Html.nodes(<p>hi</p>)[0]
println(n.name) // error: Html.Node has no field name on every member: narrow the value first
```

A record is not markup. An `Html.Element` record does not convert to `Html`, and `String.from`
renders it as a record:

```ant
let e = Html.Element{ name: "b", attributes: [], children: <>x</> }
let h: Html = e // error: cannot use Html.Element where Html is expected
```

```ant
let e = Html.Element{ name: "b", attributes: [], children: <>x</> }
String.from(e) // "Html.Element{ name: \"b\", attributes: [], children: x }"
```

Changing a node changes the copy, not the `Html` it was read from:

```ant
let link = <a href="/old">x</a>
let node = Html.nodes(link)[0]
if node is Html.Element {
  node.attributes[0].value = "/new"
  println(node.attributes)
}
println(link)
```

```text
[Html.Attribute{ name: "href", value: "/new" }]
<a href="/old">x</a>
```

Raw markup is not parsed, so a walk over the tree does not see inside it, and a raw node is not
equal to the same markup written as an expression:

```ant
Html.nodes(Html.unsafe("<a href=\"/\">x</a>"))[0] is Html.Element // false
Html.unsafe("<b>x</b>") == <b>x</b>                               // false
```

A chain over the members that leaves one out is a compile error, since the chain then has no value
for that member:

```ant
func describe(n: Html.Node): string {
  if n is Html.Element { "element" }
  else if n is Html.Text { "text" }
  // error: an if without else has type void when the condition is false, which cannot be used where string is expected
}
```

When a chain covers every member, it needs no `else` and cannot have one:

```ant
func describe(n: Html.Node): string {
  if n is Html.Element { "element ${n.name}" }
  else if n is Html.Text { "text ${n.text}" }
  else if n is Html.Raw { "raw ${n.markup}" }
  else { "unknown" } // error: this else cannot run: the tests before it cover every case
}
```

## See also

[Html.nodes](../nodes.md) · [Html.unsafe](../unsafe.md) · [Html.Attribute](Attribute.md) ·
[Html.Attributes](Attributes.md) · [Html](../index.md) ·
[HTML expressions](../../../language/literals/html.md) · [is](../../../language/operators/is.md)
