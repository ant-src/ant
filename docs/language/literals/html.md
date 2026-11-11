<!-- status: reviewed -->

# HTML expressions

An HTML expression writes markup in code: `<p>Hello, {name}</p>` is a value of type `Html`.

## Details

- `{expr}` inserts a value: a `string` becomes text, which the renderer escapes, a number becomes
  its text, `void` inserts nothing, an `Html` value goes in unchanged, and an array inserts each
  element. Any other type, `null` and `bool` among them, is a compile error.
- An attribute takes literal text in double quotes, or a value in braces: a `string`, a number, or
  `null`, which leaves the attribute out. A `bool` works only on a boolean attribute such as
  `checked` or `disabled`.
- `<script>`, `<style>`, `style`, `srcdoc`, and every attribute whose name starts with `on` take
  only literal text. An interpolated URL attribute such as `href` or `src` keeps a relative URL and
  the schemes `http`, `https`, `mailto`, and `tel`, and replaces any other with `about:invalid#ant`.
- White space that contains a line break becomes one space between content and is dropped at the
  start and end of an element, outside `<pre>`, `<textarea>`, `<script>`, and `<style>`. The source
  can be indented freely, and the output carries no indentation.
- An HTML expression that is a statement of a block but not its last item, or any statement at a
  script's top level, is a syntax error, since its value is discarded.
- `String.from` renders the markup, and `println` prints it.

## Basic examples

An element with an interpolated string. The renderer escapes the text:

```ant
let name = "Ann & Bob"
println(<p class="lead">Hello, {name}.</p>)
```

```text
<p class="lead">Hello, Ann &amp; Bob.</p>
```

A void element is written self-closing and renders without the slash:

```ant
println(<img src="/logo.png" alt="Ant" />)
```

```text
<img src="/logo.png" alt="Ant">
```

`map` is the loop, and an `if` without `else` is the conditional:

```ant
let fruits = ["apple", "pear"]
let loggedIn = false
println(<ul>{fruits.map({ <li>{x}</li> })}</ul>)
println(<nav>{if loggedIn { <a href="/account">Account</a> }}</nav>)
```

```text
<ul><li>apple</li><li>pear</li></ul>
<nav></nav>
```

## Scope

### Elements

Any element other than a void one written self-closing renders with an end tag, which is what SVG
wants too. Names keep their case, as SVG needs for `viewBox`, and custom elements work:

```ant
println(<div />)
println(<svg viewBox="0 0 10 10"><circle r="4" /></svg>)
println(<my-widget size="2" />)
```

```text
<div></div>
<svg viewBox="0 0 10 10"><circle r="4"></circle></svg>
<my-widget size="2"></my-widget>
```

### Text and character references

Text between tags is written as in an `.html` file, so `//` is text. Character references are
decoded, and the renderer escapes the characters again. An `&` that does not begin a reference is a
plain `&`, and a reference needs its `;`:

```ant
println(<p>Tom &amp; Jerry &#8212; &lt;3</p>)
println(<p>see https://ant.dev</p>)
println(<p>R&D, &copy=2</p>)
```

```text
<p>Tom &amp; Jerry — &lt;3</p>
<p>see https://ant.dev</p>
<p>R&amp;D, &amp;copy=2</p>
```

A literal brace is a numeric reference or an interpolated string:

```ant
println(<code>&#123;x&#125; {"{"}</code>)
```

```text
<code>{x} {</code>
```

### Interpolated values

Strings and numbers become text, and every character that matters to HTML is escaped. An `Html`
value goes in as markup:

```ant
let count = 3
let price = 2.5
let icon = <b>!</b>
println(<p>{count} items at {price}</p>)
println(<p>{"<b>it's \"bold\"</b>"}</p>)
println(<p>{icon} Careful</p>)
```

```text
<p>3 items at 2.5</p>
<p>&lt;b&gt;it&#39;s &quot;bold&quot;&lt;/b&gt;</p>
<p><b>!</b> Careful</p>
```

An array inserts its elements with nothing between them, at any depth:

```ant
println(<p>{["a", "b"]}{[[1, 2], [3]]}</p>)
```

```text
<p>ab123</p>
```

A nullable value needs a decision about what `null` shows:

```ant
let user: string | null = null
println(<p>Hello, {user ?? "guest"}</p>)
```

```text
<p>Hello, guest</p>
```

### Attributes

A value in braces is a string or a number, and `null` leaves the attribute out. A name alone is an
attribute with an empty value, on any element:

```ant
let id = 42
let title: string | null = null
println(<a href="/items" data-id={id} title={title}>Item</a>)
println(<input name="q" required />)
```

```text
<a href="/items" data-id="42">Item</a>
<input name="q" required>
```

On a boolean attribute, `true` gives the bare name and `false` leaves it out:

```ant
let done = true
println(<input type="checkbox" checked={done} disabled={!done} />)
```

```text
<input type="checkbox" checked>
```

`class` also takes an array: the non-empty strings joined with one space, `null` and `void` dropped,
and no attribute when nothing remains:

```ant
let active = true
println(<li class={["tab", if active { "active" }]}>Home</li>)
println(<li class={[if !active { "active" }, null, ""]}>Home</li>)
```

```text
<li class="tab active">Home</li>
<li>Home</li>
```

### Spread attributes

`{...attrs}` spreads a map of type `Html.Attributes` in the map's order. When two give one name, the
later value wins, so a later `null` removes an attribute:

```ant
let extra: Html.Attributes = Map{ "data-id": "7", disabled: true }
println(<button class="btn" {...extra}>Save</button>)
println(<button disabled {...Map{ disabled: null }} />)
```

```text
<button class="btn" data-id="7" disabled>Save</button>
<button></button>
```

A component forwards extra attributes through a parameter of type `Html.Attributes` that defaults to
an empty map:

```ant
func button(label: string, attrs: Html.Attributes = Map{}): Html {
  <button class="btn" {...attrs}>{label}</button>
}

println(button("Save"))
println(button("Delete", attrs: Map{ "data-id": "7", disabled: true }))
```

```text
<button class="btn">Save</button>
<button class="btn" data-id="7" disabled>Delete</button>
```

### URL attributes

An interpolated URL with a scheme other than `http`, `https`, `mailto`, or `tel` is replaced, and
the page still renders. `srcset` is checked candidate by candidate, and the descriptors are kept:

```ant
let link = "javascript:alert(1)"
let sources = "small.png 1x, javascript:x 2x"
println(<a href={link}>Go</a>)
println(<a href={"/search?q=ant"}>Search</a>)
println(<a href={"mailto:ann@example.com"}>Mail</a>)
println(<img srcset={sources} />)
```

```text
<a href="about:invalid#ant">Go</a>
<a href="/search?q=ant">Search</a>
<a href="mailto:ann@example.com">Mail</a>
<img srcset="small.png 1x, about:invalid#ant 2x">
```

### Scripts and styles

`<script>` and `<style>` hold literal text up to their end tag, braces and `<` included, and it is
written out as is. Data reaches a script through an attribute, which is escaped like any other:

```ant
let points = [1, 2, 3]
println(<script>if (a < b) { go() }</script>)
println(<style>p { color: red }</style>)
println(<div id="chart" data-points={Json.stringify(points)}></div>)
```

```text
<script>if (a < b) { go() }</script>
<style>p { color: red }</style>
<div id="chart" data-points="[1,2,3]"></div>
```

### Fragments

`<>...</>` is an `Html` value with no element around it. Inserted into content, its nodes are
spliced in, and `<></>` is the empty `Html`:

```ant
let items = <><li>a</li><li>b</li></>
println(<ul>{items}</ul>)
println(String.from(<></>).isEmpty())
```

```text
<ul><li>a</li><li>b</li></ul>
true
```

### Comments

A comment in markup is `{/* ... */}`, and renders nothing:

```ant
println(<p>{/* todo: translate */}Hello</p>)
```

```text
<p>Hello</p>
```

### White space

A line break with its indentation becomes one space between two pieces of content, and disappears at
the start and the end of an element:

```ant
let list = <ul>
  <li>a</li>
  <li>b</li>
</ul>
println(list)
```

```text
<ul><li>a</li> <li>b</li></ul>
```

Inside `<pre>` the text is kept exactly, except for a line break right after the start tag:

```ant
let poem = <pre>
roses are red,
  violets are blue</pre>
println(poem)
```

```text
<pre>roses are red,
  violets are blue</pre>
```

### Components

A component is a function that returns `Html`. Its parameters are its properties, and a parameter of
type `Html` takes children:

```ant
func card(title: string, children: Html, footer: Html | null = null): Html {
  <section class="card">
    <h2>{title}</h2>
    {children}
    {if footer != null { <footer>{footer}</footer> }}
  </section>
}

println(card(title: "Done", children: <p>All clear.</p>, footer: <a href="/archive">Archive</a>))
```

```text
<section class="card"><h2>Done</h2> <p>All clear.</p> <footer><a href="/archive">Archive</a></footer></section>
```

A function returns several siblings in a fragment:

```ant
func field(label: string, name: string): Html {
  <>
    <label for={name}>{label}</label>
    <input id={name} name={name} />
  </>
}

println(<form>{field("Email", "email")}</form>)
```

```text
<form><label for="email">Email</label> <input id="email" name="email"></form>
```

### Reading the tree

`Html.nodes` returns the top-level nodes, with names as written and text decoded:

```ant
let link = <a href="/">Home &amp; away</a>
let node = Html.nodes(link)[0]
if node is Html.Element {
  println(node.name)
  println(node.attributes)
}
```

```text
a
[Html.Attribute{ name: "href", value: "/" }]
```

## Applications

### A task list

Map records to list items, with a guard for the empty case:

```ant
type Task = { title: string, done: bool }

func taskItem(task: Task): Html {
  <li class={["task", if task.done { "done" }]}>
    <input type="checkbox" checked={task.done} disabled />
    {task.title}
  </li>
}

func taskList(tasks: Array[Task]): Html {
  if tasks.isEmpty() { return <p>Nothing to do.</p> }
  <ul>{tasks.map(taskItem)}</ul>
}

println(taskList([Task{ title: "Write docs", done: true }, Task{ title: "Fix <br>", done: false }]))
println(taskList([]))
```

```text
<ul><li class="task done"><input type="checkbox" checked disabled> Write docs</li><li class="task"><input type="checkbox" disabled> Fix &lt;br&gt;</li></ul>
<p>Nothing to do.</p>
```

### A table from records

```ant
type Row = { name: string, score: int }

func scoreTable(rows: Array[Row]): Html {
  <table>
    <tr><th>Name</th><th>Score</th></tr>
    {rows.map({ <tr><td>{x.name}</td><td>{x.score}</td></tr> })}
  </table>
}

println(scoreTable([Row{ name: "Ann", score: 9 }, Row{ name: "Bob", score: 7 }]))
```

```text
<table><tr><th>Name</th><th>Score</th></tr> <tr><td>Ann</td><td>9</td></tr><tr><td>Bob</td><td>7</td></tr></table>
```

### A whole page

`String.from` renders the markup, so string interpolation builds a response body with its doctype:

```ant
func page(title: string, body: Html): string {
  let html = <html lang="en">
    <head><title>{title}</title></head>
    <body>{body}</body>
  </html>
  "<!doctype html>${html}"
}

println(page("Home", <h1>Welcome</h1>))
```

```text
<!doctype html><html lang="en"><head><title>Home</title></head> <body><h1>Welcome</h1></body></html>
```

## Properties and relations

Two values built differently but meaning the same are equal, so a test compares a component's output
with an HTML expression:

```ant
func badge(n: int): Html { <b>{n}</b> }

assert(badge(3) == <b>3</b>)
assert(<p>{"a"}{"b"}</p> == <p>ab</p>)
assert(<p>{<>x</>}</p> == <p>x</p>)
assert(<i>{if false { "hidden" }}</i> == <i></i>)
```

`Html.unsafe` renders as the markup it holds, but a raw node is not equal to the element it
describes:

```ant
let raw = Html.unsafe("<b>x</b>")
assert(String.from(raw) == String.from(<b>x</b>))
assert(raw != <b>x</b>)
```

## Possible issues

### `null` and `bool` are not content

A nullable value is a case to decide, and a `bool` has no text:

```ant
let user: string | null = null
let greeting = <p>{user}</p>
// error: markup takes a string, a number, void, Html, or an array of them, got string | null; decide what null shows, as {x ?? "none"}, or use an if
```

```ant
let ok = true
let status = <p>{ok}</p>
// error: markup takes a string, a number, void, Html, or an array of them, got bool
```

A named primitive type needs a conversion:

```ant
type UserId = string
let id = UserId("u-1")
let label = <p>{id}</p>
// error: markup takes a string, a number, void, Html, or an array of them, got UserId; convert it, as {string(x)}
```

### `bool` on an attribute that is not boolean

Many attributes take the words `"true"` and `"false"`, and presence alone would mean the opposite:

```ant
let open = false
let panel = <div aria-hidden={!open}></div>
// error: aria-hidden is not a boolean attribute, so it takes a string, not a bool; write the words, as {if on { "true" } else { "false" }}
```

Write the string:

```ant
let open = false
println(<div aria-hidden={if open { "false" } else { "true" }}></div>)
```

```text
<div aria-hidden="true"></div>
```

A custom attribute that is present or absent takes the empty string or `null`:

```ant
let open = true
println(<div data-open={if open { "" } else { null }}></div>)
```

```text
<div data-open></div>
```

### Attributes that run code

Event handlers and `style` take only literal text:

```ant
let handler = "save()"
let button = <button onclick={handler}>Save</button>
// error: onclick takes only literal text, since it runs code; pass data in a data- attribute that a script reads, or use a class
```

A spread checks its names when the element is built, and panics:

```ant
let attrs: Html.Attributes = Map{ onclick: "save()" }
let button = <button {...attrs}>Save</button>
// panic: onclick takes only literal text, so a spread cannot give it
```

### `key` is reserved

`key` is reserved, so a list item has no key:

```ant
let xs = ["a"]
let list = <ul>{xs.map({ <li key={x}>{x}</li> })}</ul> // error: key is reserved for a later client renderer
```

### Single quotes

An attribute value is written in double quotes:

```ant
let p = <p class='lead'>x</p> // error: an attribute value is written in double quotes
```

### Interpolation inside quotes

Quoted attribute text is literal:

```ant
let name = "Ann"
let greeting = <p title="hi {name}">x</p>
// error: { in a quoted attribute value; interpolate with title={...}, or write &#123;
```

Interpolate the whole value instead:

```ant
let name = "Ann"
println(<p title={"hi ${name}"}>x</p>)
```

```text
<p title="hi Ann">x</p>
```

### Named character references

Only `&amp;`, `&lt;`, `&gt;`, `&quot;`, `&apos;`, and `&nbsp;` are named. Write any other character
itself, as in `<p>© 2026</p>`, or as a numeric reference:

```ant
let footer = <p>&copy; 2026</p>
// error: &copy; is not one of the named references &amp; &lt; &gt; &quot; &apos; &nbsp;; write the character itself or a numeric reference
```

### A lone closing brace

```ant
let text = <p>a } b</p> // error: a } in text is written &#125; or {"}"}
```

### HTML comments

```ant
let note = <p><!-- c --></p> // error: an HTML comment is written {/* ... */} in markup
```

### Siblings need a fragment

An HTML expression that is not the last statement of its block is discarded, which is a syntax
error. Wrap the siblings in `<>...</>`:

```ant
func header(): Html {
  <h1>Title</h1>
  // error: this HTML expression's value is discarded; put siblings in a fragment <>...</>
  <p>Subtitle</p>
}
```

### Void elements and end tags

```ant
let br = <br></br> // error: <br> is a void element; write <br />
```

An end tag repeats its start tag's name exactly:

```ant
let box = <div></DIV> // error: </DIV> does not close <div>
```

### `<title>` and `<textarea>` hold text only

```ant
let title = <title>{<b>Home</b>}</title>
// error: <title> and <textarea> hold text only, so they cannot hold Html
```

### `Html.unsafe` bypasses escaping

`Html.unsafe` writes its string unescaped, and literal URLs are never checked. Pass only markup the
program trusts:

```ant
let comment = "<img src=x onerror=alert(1)>"
println(<p>{comment}</p>)
println(<p>{Html.unsafe(comment)}</p>)
```

```text
<p>&lt;img src=x onerror=alert(1)&gt;</p>
<p><img src=x onerror=alert(1)></p>
```

## See also

[Html](../../lib/Html/index.md) · [Html.unsafe](../../lib/Html/unsafe.md) ·
[Html.nodes](../../lib/Html/nodes.md) · [Html.Attributes](../../lib/Html/types/Attributes.md) ·
[Html.Node](../../lib/Html/types/Node.md) · [String.from](../../lib/String/from.md) ·
[Array.map](../../lib/Array/map.md) · [Strings](strings.md) · [if](../keywords/if.md) ·
[HTML expressions guide](../../guide/html.md)
