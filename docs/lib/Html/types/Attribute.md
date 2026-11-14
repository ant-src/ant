<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Html.Attribute

```ant signature
type Html.Attribute = { name: string, value: string }
```

Html.Attribute is an attribute of an element; a boolean attribute that is on has the empty value.

<!-- end generated -->

## Details

- An `Html.Attribute` is how [Html.nodes](../nodes.md) reports an attribute: the `attributes` of an
  [Html.Element](Node.md) list them in their final order, after spreads and repeated names are
  resolved.
- A number gives its text, a boolean attribute that is on, or one written without a value, gives
  `""`, and character references are decoded: `"&amp;"` in the source is `"&"`.
- An attribute left out, by `null` or by `false` on a boolean attribute, is not in the list.
- `name` keeps its case as written: `viewBox` stays `viewBox`, and `CLASS` stays `CLASS`.
- The records are copies. Changing them does not change the `Html` they came from, and no function
  builds `Html` from them.

## Basic examples

Read the attributes of an element with [Html.nodes](../nodes.md):

```ant
let node = Html.nodes(<a href="/docs" title="Read the docs">Docs</a>)[0]
if node is Html.Element {
  for attr in node.attributes {
    println("${attr.name} = ${attr.value}")
  }
}
```

```text
href = /docs
title = Read the docs
```

An `Html.Attribute` is a record, so a literal builds one, and it compares by its fields:

```ant
let id = Html.Attribute{ name: "id", value: "main" }
id                                                // Html.Attribute{ name: "id", value: "main" }
id == Html.Attribute{ name: "id", value: "main" } // true
```

## Scope

The cases below read attributes with this helper:

```ant
func attributesOf(h: Html): Array[Html.Attribute] {
  let node = Html.nodes(h)[0]
  if node is Html.Element { node.attributes } else { [] }
}

attributesOf(<input type="text" />) // [Html.Attribute{ name: "type", value: "text" }]
attributesOf(<p>no attributes</p>)  // []
```

A boolean attribute that is on, and any attribute written without a value, has the empty value. A
boolean attribute set to `false` and any attribute set to `null` are left out:

<!-- test: before
func attributesOf(h: Html): Array[Html.Attribute] {
  let node = Html.nodes(h)[0]
  if node is Html.Element { node.attributes } else { [] }
}
-->

```ant
attributesOf(<input checked={true} disabled={false} required />) // [Html.Attribute{ name: "checked", value: "" }, Html.Attribute{ name: "required", value: "" }]
attributesOf(<p id={null}>x</p>)                                 // []
attributesOf(<div hidden="until-found"></div>)                   // [Html.Attribute{ name: "hidden", value: "until-found" }]
```

A number becomes its text, and a `class` array becomes its non-empty strings joined with a space:

<!-- test: before
func attributesOf(h: Html): Array[Html.Attribute] {
  let node = Html.nodes(h)[0]
  if node is Html.Element { node.attributes } else { [] }
}
-->

```ant
attributesOf(<img width={640} />)                           // [Html.Attribute{ name: "width", value: "640" }]
attributesOf(<li class={["task", null, "", "done"]}>x</li>) // [Html.Attribute{ name: "class", value: "task done" }]
attributesOf(<li class={[null, ""]}>x</li>)                 // []
```

Character references are decoded, in a literal value as in text:

<!-- test: before
func attributesOf(h: Html): Array[Html.Attribute] {
  let node = Html.nodes(h)[0]
  if node is Html.Element { node.attributes } else { [] }
}
-->

```ant
attributesOf(<a href="/search?q=ant&amp;page=2">x</a>) // [Html.Attribute{ name: "href", value: "/search?q=ant&page=2" }]
```

An interpolated URL whose scheme is not `http`, `https`, `mailto`, or `tel` is replaced when the
element is built, so `value` holds the replacement:

<!-- test: before
func attributesOf(h: Html): Array[Html.Attribute] {
  let node = Html.nodes(h)[0]
  if node is Html.Element { node.attributes } else { [] }
}
-->

```ant
let url = "javascript:alert(1)"
attributesOf(<a href={url}>x</a>) // [Html.Attribute{ name: "href", value: "about:invalid#ant" }]
```

A spread gives one attribute per entry left after `null` values are dropped, in the map's order:

<!-- test: before
func attributesOf(h: Html): Array[Html.Attribute] {
  let node = Html.nodes(h)[0]
  if node is Html.Element { node.attributes } else { [] }
}
-->

```ant
let extra: Html.Attributes = Map{ "aria-label": "Close", hidden: true, title: null }
attributesOf(<button class="icon" {...extra} />) // [Html.Attribute{ name: "class", value: "icon" }, Html.Attribute{ name: "aria-label", value: "Close" }, Html.Attribute{ name: "hidden", value: "" }]
```

## Applications

Find one attribute by name:

```ant
func attr(h: Html, name: string): string | null {
  let node = Html.nodes(h)[0]
  if !(node is Html.Element) { return null }
  node.attributes.find({ x.name == name })?.value
}

let logo = <img src="/logo.png" alt="Ant" />
println(attr(logo, "src"))
println(attr(logo, "width"))
```

```text
/logo.png
null
```

Collect every link target in a tree, reading each element's children with [Html.nodes](../nodes.md)
again:

```ant
func links(h: Html): Array[string] {
  Html.nodes(h) | Array.flatMap(func(node: Html.Node): Array[string] {
    if !(node is Html.Element) { return [] }
    let href = node.attributes.find({ x.name == "href" })
    let own: Array[string] = if href != null { [href.value] } else { [] }
    own.concat(links(node.children))
  })
}

let nav = <nav>
  <a href="/">Home</a>
  <ul>
    <li><a href="/docs">Docs</a></li>
    <li><a href="/blog">Blog</a></li>
  </ul>
</nav>
println(links(nav))
```

```text
["/", "/docs", "/blog"]
```

Turn the attributes of one element into an [Html.Attributes](Attributes.md) map and spread them on
another:

```ant
func attributesOf(h: Html): Array[Html.Attribute] {
  let node = Html.nodes(h)[0]
  if node is Html.Element { node.attributes } else { [] }
}

func toSpread(attrs: Array[Html.Attribute]): Html.Attributes {
  let spread: Html.Attributes = Map{}
  for a in attrs {
    spread[a.name] = a.value
  }
  spread
}

let field = <input type="email" name="email" required />
println(<input {...toSpread(attributesOf(field))} placeholder="you@example.com" />)
```

```text
<input type="email" name="email" required placeholder="you@example.com">
```

## Properties and relations

`value` holds the string the source gave, unescaped; rendering escapes it:

```ant
let title = "Tom & \"Jerry\""
let input = <input title={title} />
let node = Html.nodes(input)[0]
if node is Html.Element {
  assert(node.attributes == [Html.Attribute{ name: "title", value: title }])
}
assert(String.from(input) == "<input title=\"Tom &amp; &quot;Jerry&quot;\">")
```

A name repeated through a spread appears once, with the later value, at the earlier position:

```ant
let node = Html.nodes(<a href="/old" title="Home" {...Map{ href: "/new" }}>x</a>)[0]
if node is Html.Element {
  assert(node.attributes == [
    Html.Attribute{ name: "href", value: "/new" },
    Html.Attribute{ name: "title", value: "Home" },
  ])
}
```

An empty value renders as the bare name, so `alt=""` and `alt` build the same value:

```ant
assert(<img alt="" /> == <img alt />)
```

## Possible issues

A boolean attribute that is on has the value `""`, not `"true"`. Test that it is present:

```ant
let node = Html.nodes(<input type="checkbox" checked={true} />)[0]
if node is Html.Element {
  println(node.attributes.some({ x.name == "checked" && x.value == "true" }))
  println(node.attributes.some({ x.name == "checked" }))
}
```

```text
false
true
```

`name` keeps the case it was written in, so comparing it with `==` misses `CLASS`. Compare with
[String.equalFold](../../String/equalFold.md), as browsers fold case on HTML names:

```ant
let node = Html.nodes(<p CLASS="lead">x</p>)[0]
if node is Html.Element {
  println(node.attributes.find({ x.name == "class" }))
  println(node.attributes.find({ x.name.equalFold("class") })?.value)
}
```

```text
null
lead
```

Changing the records does not change the `Html` they came from:

```ant
let link = <a href="/old">Home</a>
let node = Html.nodes(link)[0]
if node is Html.Element {
  node.attributes[0].value = "/new"
  println(node.attributes[0].value)
}
println(link)
```

```text
/new
<a href="/old">Home</a>
```

A spread rejects attributes that run code, so copying the attributes of an element with `style`, an
`on` handler, or `srcdoc` panics. Skip them first, for example with
`attrs | Array.filter({ !x.name.startsWith("on") && x.name != "style" && x.name != "srcdoc" })`:

```ant
func attributesOf(h: Html): Array[Html.Attribute] {
  let node = Html.nodes(h)[0]
  if node is Html.Element { node.attributes } else { [] }
}

func toSpread(attrs: Array[Html.Attribute]): Html.Attributes {
  let spread: Html.Attributes = Map{}
  for a in attrs {
    spread[a.name] = a.value
  }
  spread
}

let styled = <div style="color: red">x</div>
println(<p {...toSpread(attributesOf(styled))}>y</p>) // panic: style takes only literal text, so a spread cannot give it
```

## See also

[Html.nodes](../nodes.md) · [Html.Node](Node.md) · [Html.Attributes](Attributes.md) ·
[Html](../index.md) · [HTML expressions](../../../language/literals/html.md)
