<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Html.Attributes

```ant signature
type Html.Attributes = Map[string, string | bool | null]
```

Html.Attributes is what an attribute spread \{...attrs\} takes.

<!-- end generated -->

## Details

- A spread writes the entries in the map's order, each as if written there as `name={value}`: a
  string is the value, `null` leaves the attribute out, and a `bool` works only on a boolean
  attribute such as `disabled` or `hidden`; on any other name it panics.
- Where a spread and another attribute give one name, compared ignoring ASCII case, the later value
  wins and the attribute keeps the earlier position. A later `null` removes the attribute.
- The names are data, so they are checked when the element is built: a name that is not an attribute
  name, `key`, or a name that runs code (`on...`, `style`, `srcdoc`) panics.
- A URL attribute such as `href` or `src` is checked as when written: a scheme other than `http`,
  `https`, `mailto`, or `tel` becomes `about:invalid#ant`.
- The value type has no numbers: write `String.from(n)` or `"${n}"`.

## Basic examples

Spread a map of attributes onto an element:

```ant
let attrs: Html.Attributes = Map{ id: "save", "data-id": "7" }
println(<button {...attrs}>Save</button>)
```

```text
<button id="save" data-id="7">Save</button>
```

A component takes extra attributes through a parameter of type `Html.Attributes` that defaults to an
empty map:

```ant
func button(label: string, attrs: Html.Attributes = Map{}): Html {
  <button class="btn" {...attrs}>{label}</button>
}

println(button("Save"))
println(button("Delete", attrs: Map{ "data-confirm": "Delete it?", disabled: true }))
```

```text
<button class="btn">Save</button>
<button class="btn" data-confirm="Delete it?" disabled>Delete</button>
```

## Scope

A string is the value as it is, and the renderer escapes it:

```ant
println(<p {...Map{ title: "Tom & \"Jerry\"" }} />)
```

```text
<p title="Tom &amp; &quot;Jerry&quot;"></p>
```

On a boolean attribute, `true` gives the bare name and `false` leaves it out:

```ant
let open = false
println(<details {...Map{ open: open, hidden: true }} />)
```

```text
<details hidden></details>
```

`null` leaves the attribute out, so a nullable value gives an optional attribute:

```ant
let tooltip: string | null = null
let attrs: Html.Attributes = Map{ title: tooltip, lang: "en" }
println(<p {...attrs}>Hej</p>)
```

```text
<p lang="en">Hej</p>
```

An empty map adds nothing:

```ant
let none: Html.Attributes = Map{}
println(<hr {...none} />)
```

```text
<hr>
```

The entries keep the map's order, after the attributes written before the spread and before the ones
written after it:

```ant
let attrs: Html.Attributes = Map{ b: "2", a: "1" }
println(<p z="0" {...attrs} y="3" />)
```

```text
<p z="0" b="2" a="1" y="3"></p>
```

A later attribute or spread overrides an earlier one with the same name, in the earlier position:

```ant
let defaults: Html.Attributes = Map{ type: "button", class: "btn" }
println(<button {...defaults} type="submit">Send</button>)
println(<button disabled {...Map{ disabled: null }} />)
```

```text
<button type="submit" class="btn">Send</button>
<button></button>
```

Since a name is compared ignoring ASCII case, `Title` and `title` are one attribute:

```ant
println(<p title="a" {...Map{ Title: null }} />)
```

```text
<p></p>
```

## Applications

### Optional attributes from settings

Fill the map with `if` expressions; a `null` drops the attribute:

<!-- test: before
type Field = { name: string, label: string, required: bool, hint: string | null }
-->

```ant
func input(field: Field): Html {
  let attrs: Html.Attributes = Map{
    "aria-describedby": if field.hint != null { "${field.name}-hint" } else { null },
    required: field.required,
  }
  <label>{field.label} <input name={field.name} {...attrs} /></label>
}

println(input(Field{ name: "email", label: "Email", required: true, hint: "work address" }))
println(input(Field{ name: "nick", label: "Nickname", required: false, hint: null }))
```

```text
<label>Email <input name="email" aria-describedby="email-hint" required></label>
<label>Nickname <input name="nick"></label>
```

### Data attributes from a record

Build `data-` attributes for a script to read, converting numbers to strings:

```ant
type Product = { id: int, sku: string, price: float64 }

func card(p: Product): Html {
  let data: Html.Attributes = Map{
    "data-id": String.from(p.id),
    "data-sku": p.sku,
    "data-price": String.from(p.price),
  }
  <article class="product" {...data}>{p.sku}</article>
}

println(card(Product{ id: 7, sku: "ANT-1", price: 9.5 }))
```

```text
<article class="product" data-id="7" data-sku="ANT-1" data-price="9.5">ANT-1</article>
```

### Layered defaults

Spread a component's defaults and then the caller's attributes, so the caller's entries win:

```ant
func link(href: string, label: string, attrs: Html.Attributes = Map{}): Html {
  let defaults: Html.Attributes = Map{ rel: "noopener", target: "_blank" }
  <a href={href} {...defaults} {...attrs}>{label}</a>
}

println(link("https://ant-lang.org", "Ant"))
println(link("/docs", "Docs", attrs: Map{ target: null }))
```

```text
<a href="https://ant-lang.org" rel="noopener" target="_blank">Ant</a>
<a href="/docs" rel="noopener">Docs</a>
```

## Properties and relations

A spread is the same as writing each entry as an attribute:

```ant
let attrs: Html.Attributes = Map{ id: "a", hidden: true, title: null }
assert(<p {...attrs} /> == <p id="a" hidden={true} title={null} />)
assert(<p {...attrs} /> == <p id="a" hidden />)
```

[Html.nodes](../nodes.md) reads the attributes back in their final order, as
[Html.Attribute](Attribute.md) records; a boolean attribute that is on has the empty value:

```ant
let attrs: Html.Attributes = Map{ id: "a", hidden: true }
let node = Html.nodes(<p {...attrs} />)[0]
if node is Html.Element {
  println(node.attributes)
}
```

```text
[Html.Attribute{ name: "id", value: "a" }, Html.Attribute{ name: "hidden", value: "" }]
```

## Possible issues

### A map literal without a type

Type arguments are invariant, so a map whose type is inferred from its entries alone, such as
`Map[string, string]`, is not an `Html.Attributes`. Annotate the binding, or write the map inside
the spread or as the argument, where the expected type applies:

```ant
let attrs = Map{ id: "a" }
let p = <p {...attrs} />
// error: cannot use Map[string, string] where Map[string, string | bool | null] is expected
```

### Names that run code

A spread cannot give an event handler, `style`, or `srcdoc`:

```ant
let attrs: Html.Attributes = Map{ onclick: "save()" }
let button = <button {...attrs}>Save</button>
// panic: onclick takes only literal text, so a spread cannot give it
```

### A bool on an attribute that is not boolean

Many attributes take the words `"true"` and `"false"`; write the string:

```ant
let attrs: Html.Attributes = Map{ "aria-hidden": true }
let icon = <span {...attrs} />
// panic: aria-hidden is not a boolean attribute, so it takes a string, not a bool
```

Choose the string with an `if` expression. A custom attribute that is only on or off takes `""` for
on and `null` for off:

```ant
let open = false
let attrs: Html.Attributes = Map{
  "aria-hidden": if open { "false" } else { "true" },
  "data-active": if open { "" } else { null },
}
println(<span {...attrs} />)
```

```text
<span aria-hidden="true"></span>
```

### Class lists in a spread

Written directly, `class={[...]}` takes an array, but a spread's values are `string | bool | null`,
so an array is a compile error:

```ant
let primary = true
let attrs: Html.Attributes = Map{ class: ["btn", if primary { "primary" }] }
// error: cannot use Array[string | void] where string | bool | null is expected
```

Give the whole class string, or write `class={[...]}` next to the spread. `String.join` takes an
`Array[string]`, and an `if` without `else` is `void` when false, so the array above cannot be
joined as it is:

```ant
let primary = true
let attrs: Html.Attributes = Map{ class: if primary { "btn primary" } else { "btn" } }
println(<button {...attrs}>Save</button>)
println(<button class={["btn", if primary { "primary" }]} {...Map{ id: "save" }}>Save</button>)
```

```text
<button class="btn primary">Save</button>
<button class="btn primary" id="save">Save</button>
```

### Names that are not attribute names

A name from data must still be an attribute name, and `key` is reserved:

```ant
let attrs: Html.Attributes = Map{ "data id": "7" }
let p = <p {...attrs} /> // panic: "data id" is not an attribute name
```

```ant
let attrs: Html.Attributes = Map{ key: "7" }
let li = <li {...attrs} /> // panic: key is reserved, so a spread cannot give it
```

### Unsafe URLs

A URL attribute from a spread is checked as when written, so a `javascript:` URL is replaced:

```ant
let attrs: Html.Attributes = Map{ href: "javascript:alert(1)" }
println(<a {...attrs}>Home</a>)
```

```text
<a href="about:invalid#ant">Home</a>
```

## See also

[Html.Attribute](Attribute.md) · [Html.Node](Node.md) · [Html.nodes](../nodes.md) ·
[HTML expressions](../../../language/literals/html.md) · [Map.merge](../../Map/merge.md)
