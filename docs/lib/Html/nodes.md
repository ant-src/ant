<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Html.nodes

```ant signature
func Html.nodes(h: Html): Array[Html.Node]
```

Html.nodes returns the top-level nodes of h, in order. An element's children are read with
Html.nodes again.

<!-- end generated -->

## Details

- The value is in its normal form: fragments are spliced in, adjacent text is joined into one text
  node, and empty text is dropped. A line break between two tags in the source leaves a text node
  holding one space.
- Character references written in the markup are decoded, so `&amp;` reads as `&`. An interpolated
  string is kept as it is. A boolean attribute that is on has the empty value, and one that is off
  is not in the list.
- The records are copies. Changing one leaves the `Html` value as it was, and no function turns
  records back into `Html`.

## Basic examples

An element gives one `Html.Element`, with its name, its attributes, and its children:

```ant
Html.nodes(<p>hi</p>) // [Html.Element{ name: "p", attributes: [], children: hi }]
```

A fragment gives its nodes in order:

```ant
Html.nodes(<><h1>Title</h1>Intro</>) // [Html.Element{ name: "h1", attributes: [], children: Title }, Html.Text{ text: "Intro" }]
```

Narrow a node with `is` before reading its fields:

```ant
let link = <a href="/about">About</a>
let node = Html.nodes(link)[0]
if node is Html.Element {
  println(node.name)
  println(node.attributes)
}
```

```text
a
[Html.Attribute{ name: "href", value: "/about" }]
```

## Scope

The empty `Html` has no nodes:

```ant
Html.nodes(<></>) // []
```

Text and interpolated values become one text node, since adjacent text is joined:

```ant
let name = "Ann"
Html.nodes(<>Hello, {name}! You have {3} tasks.</>) // [Html.Text{ text: "Hello, Ann! You have 3 tasks." }]
```

Character references are decoded in text and in attribute values. An interpolated string is kept as
it is:

```ant
Html.nodes(<>Tom &amp; Jerry&#33;</>)         // [Html.Text{ text: "Tom & Jerry!" }]
Html.nodes(<abbr title="&quot;HTML&quot;" />) // [Html.Element{ name: "abbr", attributes: [Html.Attribute{ name: "title", value: "\"HTML\"" }], children: <></> }]
Html.nodes(<>{"Tom &amp; Jerry"}</>)          // [Html.Text{ text: "Tom &amp; Jerry" }]
```

A boolean attribute that is on has the empty value. One that is `false` is left out:

```ant
Html.nodes(<input checked={true} disabled={false} required />) // [Html.Element{ name: "input", attributes: [Html.Attribute{ name: "checked", value: "" }, Html.Attribute{ name: "required", value: "" }], children: <></> }]
```

Attributes come in their final order. A spread adds its entries where it stands, and a `null` value
leaves its attribute out:

```ant
let extra: Html.Attributes = Map{ id: "main", title: null }
Html.nodes(<div class="box" {...extra} />) // [Html.Element{ name: "div", attributes: [Html.Attribute{ name: "class", value: "box" }, Html.Attribute{ name: "id", value: "main" }], children: <></> }]
```

Names keep the case they were written in:

```ant
Html.nodes(<svg viewBox="0 0 10 10" />) // [Html.Element{ name: "svg", attributes: [Html.Attribute{ name: "viewBox", value: "0 0 10 10" }], children: <></> }]
```

Markup from `Html.unsafe` is one raw node. It is not parsed into elements:

```ant
Html.nodes(Html.unsafe("<b>bold</b> text")) // [Html.Raw{ markup: "<b>bold</b> text" }]
```

An array interpolated into content gives a node per element:

```ant
let items = ["a", "b"].map({ <li>{x}</li> })
Html.nodes(<>{items}</>) // [Html.Element{ name: "li", attributes: [], children: a }, Html.Element{ name: "li", attributes: [], children: b }]
```

Read the children of an element with `Html.nodes` again:

```ant
let list = Html.nodes(<ul><li>one</li><li>two</li></ul>)[0]
if list is Html.Element {
  println(Html.nodes(list.children).length())
}
```

```text
2
```

Call it through `.` or in a pipe:

```ant
let para = <p>hi</p>
para.nodes()        // [Html.Element{ name: "p", attributes: [], children: hi }]
para | Html.nodes() // [Html.Element{ name: "p", attributes: [], children: hi }]
```

## Applications

Collect the text of a tree, walking every element. The chain of `is` tests covers every kind of
node, so the checker accepts it without an `else`:

```ant
func text(h: Html): string {
  Html.nodes(h).map(nodeText) | String.join("")
}

func nodeText(node: Html.Node): string {
  if node is Html.Text { node.text }
  else if node is Html.Element { text(node.children) }
  else if node is Html.Raw { "" }
}

println(text(<article><h1>News</h1><p>Ant &amp; <b>friends</b></p></article>))
```

```text
NewsAnt & friends
```

List every link target on a page:

```ant
func hrefs(h: Html): Array[string] {
  Html.nodes(h).flatMap(func(node: Html.Node): Array[string] {
    if !(node is Html.Element) { return [] }
    let own = node.attributes.filter({ x.name == "href" }).map({ x.value })
    own.concat(hrefs(node.children))
  })
}

let page = <nav>
  <a href="/">Home</a>
  <ul><li><a href="/docs">Docs</a></li></ul>
</nav>
println(hrefs(page))
```

```text
["/", "/docs"]
```

Check one attribute of a component's output in a test, without spelling out the whole markup:

```ant
type Task = { title: string, done: bool }

func taskItem(task: Task): Html {
  <li class={["task", if task.done { "done" }]}>{task.title}</li>
}

func classOf(h: Html): string | null {
  let node = Html.nodes(h)[0]
  if !(node is Html.Element) { return null }
  node.attributes.find({ x.name == "class" })?.value
}

assert(classOf(taskItem(Task{ title: "Ship", done: true })) == "task done")
assert(classOf(taskItem(Task{ title: "Plan", done: false })) == "task")
```

Count the elements in a tree by name:

```ant
func countTags(h: Html, counts: Map[string, int]): void {
  for node in Html.nodes(h) {
    if node is Html.Element {
      counts[node.name] = (counts[node.name] ?? 0) + 1
      countTags(node.children, counts)
    }
  }
}

let counts: Map[string, int] = Map{}
countTags(<div><p>a <b>b</b></p><p>c</p></div>, counts)
println(counts)
```

```text
Map{ "div": 1, "p": 2, "b": 1 }
```

## Properties and relations

Two `Html` values are equal exactly when their nodes are:

```ant
let a = <p>{"a"}{"b"}</p>
let b = <p>ab</p>
assert(a == b)
assert(Html.nodes(a) == Html.nodes(b))
```

A fragment's nodes are its parts' nodes in order, when no text meets text at a seam:

```ant
let head = <h1>Title</h1>
let body = <p>Text</p>
assert(Html.nodes(<>{head}{body}</>) == Html.nodes(head).concat(Html.nodes(body)))
```

The records are copies, so changing one leaves the `Html` value as it was:

```ant
let link = <a href="/old">Old</a>
let node = Html.nodes(link)[0]
if node is Html.Element {
  node.attributes[0].value = "/new"
}
println(link)
```

```text
<a href="/old">Old</a>
```

## Possible issues

A line break between tags leaves a text node of one space, so the second node below is text, not the
second item:

```ant
let list = <ul>
  <li>a</li>
  <li>b</li>
</ul>
let ul = Html.nodes(list)[0]
if ul is Html.Element {
  println(Html.nodes(ul.children))
}
```

```text
[Html.Element{ name: "li", attributes: [], children: a }, Html.Text{ text: " " }, Html.Element{ name: "li", attributes: [], children: b }]
```

Skip the text when only the elements matter:

```ant
let list = <ul>
  <li>a</li>
  <li>b</li>
</ul>
let ul = Html.nodes(list)[0]
if ul is Html.Element {
  println(Html.nodes(ul.children).countWhere({ x is Html.Element }))
}
```

```text
2
```

`Html` itself cannot be looped over or indexed. Call `Html.nodes` first:

```ant
let list = <ul><li>a</li></ul>
for node in list { // error: for needs an array, a map, a set, a string, Bytes, or a sequence, got Html
  println(node)
}
let first = list[0] // error: Html cannot be indexed
```

An `Html.Element` record is not markup, and content does not take it:

```ant
let el = Html.Element{ name: "b", attributes: [], children: <>x</> }
let out = <div>{el}</div> // error: markup takes a string, a number, void, Html, or an array of them, got Html.Element
```

A raw node is opaque. It renders like the markup it holds, but it is neither equal to that markup
nor split into elements:

```ant
let raw = Html.unsafe("<b>x</b>")
assert(String.from(raw) == String.from(<b>x</b>))
assert(raw != <b>x</b>)
let expected: Array[Html.Node] = [Html.Raw{ markup: "<b>x</b>" }]
assert(Html.nodes(raw) == expected)
```

## See also

[Html.unsafe](unsafe.md), [Html.Node](types/Node.md), [Html.Attribute](types/Attribute.md),
[Html](index.md), [HTML expressions](../../language/literals/html.md)
