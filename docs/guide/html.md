<!-- status: reviewed -->

# HTML expressions

Ant writes markup in code. `<p>Hello, {name}</p>` is an expression, and its value has the type
`Html`. You build pages from functions that return `Html`, put values into markup with braces, and
render the result with `String.from` or `println`. The renderer escapes every string you put in, so
data stays text unless you say otherwise.

This chapter shows how the parts fit: elements and text, values in content and attributes,
components, what escaping covers, `Html.unsafe`, and serving a page. The
[HTML literal](../language/literals/html.md) reference lists every rule.

## Markup is a value

```ant
let name = "Ann & Bob"
let greeting = <p class="lead">Hello, {name}.</p>
println(greeting)
println(greeting == <p class="lead">Hello, Ann &amp; Bob.</p>)
```

```text
<p class="lead">Hello, Ann &amp; Bob.</p>
true
```

`greeting` is not a string. It is a tree of nodes: an element named `p`, its attributes, and its
children. Text between tags is written as in an `.html` file, so `&amp;` in the source is the
character `&`, and the renderer writes it as `&amp;` again. Two `Html` values are equal when they
hold the same nodes, however you built them.

An element is a start tag, content, and a matching end tag. A void element, such as `<br />`,
`<img />`, or `<input />`, is written self-closing and renders without the slash. Any other element
written self-closing renders with its end tag, as SVG wants:

```ant
println(<p>one<br />two</p>)
println(<svg viewBox="0 0 10 10"><circle r="4" /></svg>)
```

```text
<p>one<br>two</p>
<svg viewBox="0 0 10 10"><circle r="4"></circle></svg>
```

Indent the source as you indent code. White space that contains a line break disappears at the start
and end of an element's content and becomes one space elsewhere, so the output carries no
indentation:

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

`<pre>` and `<textarea>` keep their text exactly, as the browser shows it.

## Putting values into content

Braces put a value into content. What it becomes depends on its type:

```ant
let title = "Fix <script> tag"
let count = 3
let icon = <b>!</b>
println(<p>{icon} {title}: {count} left</p>)
```

```text
<p><b>!</b> Fix &lt;script&gt; tag: 3 left</p>
```

- A string becomes a text node, escaped when rendered.
- A number becomes its text.
- An `Html` value goes in as markup.
- `void` puts in nothing, so an `if` without `else` is the conditional.
- An array puts in each element, so `map` is the loop.

`{` always starts an interpolation. Write a literal brace as `&#123;` and `&#125;`, or as `{"{"}`:

```ant
println(<code>&#123;x&#125;</code>)
println(<code>{"{"}y{"}"}</code>)
```

```text
<code>{x}</code>
<code>{y}</code>
```

```ant
let tasks = ["Buy milk", "Call Bob"]
let signedIn = false
println(<ul>{tasks.map({ <li>{x}</li> })}</ul>)
println(<nav>{if signedIn { <a href="/account">Account</a> }}</nav>)
println(<p>{if tasks.length() == 1 { "1 task" } else { "${tasks.length()} tasks" }}</p>)
```

```text
<ul><li>Buy milk</li><li>Call Bob</li></ul>
<nav></nav>
<p>2 tasks</p>
```

Anything else is a compile error: `null`, `bool`, a record, a map, and a named type such as
`UserId`, which goes in converted, as `{string(id)}`. A nullable value on a page is a case you have
not decided yet, so the type checker asks you to decide it:

```ant
type User = { name: string, nickname: string | null }

func badge(user: User): Html {
  <span>{user.nickname}</span>
  // error: markup takes a string, a number, void, Html, or an array of them, got string | null; decide what null shows, as {x ?? "none"}, or use an if
}
```

Write the decision: `{user.nickname ?? user.name}`, or an `if` that leaves the element out. A record
goes in through the fields you want to show, `{task.title}`, and a `bool` through the words it
means, `{if done { "done" } else { "open" }}`.

## Attributes

An attribute is literal text in double quotes, or a value in braces:

```ant
let field = "email"
let value: string | null = null
println(<input type="email" name={field} value={value} required />)
```

```text
<input type="email" name="email" required>
```

- `name="text"` is literal text. It has no interpolation: to build a value, write
  `href={"/users/${id}"}`.
- `name={expr}` takes a string or a number. `null` leaves the attribute out, which is how an
  optional attribute is written.
- `name` alone is the attribute with an empty value.

A boolean attribute, such as `checked`, `disabled`, `hidden`, `required`, or `selected`, also takes
a `bool`: `true` writes the attribute and `false` leaves it out. Any other attribute takes no
`bool`, because many of them take the words `"true"` and `"false"` and mean the opposite of what
presence would. Write the words:

```ant
let done = true
let open = false
println(<input type="checkbox" checked={done} disabled={!done} />)
println(<div aria-hidden={if open { "false" } else { "true" }}>menu</div>)
```

```text
<input type="checkbox" checked>
<div aria-hidden="true">menu</div>
```

`class` also takes an array. The non-empty strings are joined with one space, and `null`, `void`,
and empty strings drop out, so an `if` without `else` adds a class conditionally:

```ant
let active = true
let size: string | null = null
println(<li class={["tab", if active { "active" }, size]}>Home</li>)
```

```text
<li class="tab active">Home</li>
```

To pass attributes through a component, spread a map of type `Html.Attributes`. A later attribute of
the same name wins, and `null` removes one:

```ant
func button(label: string, attrs: Html.Attributes = Map{}): Html {
  <button type="button" class="btn" {...attrs}>{label}</button>
}

println(button("Save"))
println(button("Send", attrs: Map{ type: "submit", disabled: true, "data-form": "signup" }))
```

```text
<button type="button" class="btn">Save</button>
<button type="submit" class="btn" disabled data-form="signup">Send</button>
```

A spread's names are data, so the runtime checks them when the element is built. A name that runs
code, such as `onclick`, panics there.

## Components

A component is a function that returns `Html`. Parameters are its properties, named arguments name
them, defaults make them optional, and a parameter of type `Html` takes children:

```ant
type Task = { title: string, done: bool }

func card(title: string, children: Html, footer: Html | null = null): Html {
  <section class="card">
    <h2>{title}</h2>
    {children}
    {if footer != null { <footer>{footer}</footer> }}
  </section>
}

func taskItem(task: Task): Html {
  <li class={["task", if task.done { "done" }]}>{task.title}</li>
}

let tasks = [Task{ title: "Buy milk", done: true }, Task{ title: "Call Bob", done: false }]
let open = tasks.filter({ !x.done }).length()
println(card(
  title: "Today",
  children: <ul>{tasks.map(taskItem)}</ul>,
  footer: <>{open} open</>,
))
```

```text
<section class="card"><h2>Today</h2> <ul><li class="task done">Buy milk</li><li class="task">Call Bob</li></ul> <footer>1 open</footer></section>
```

A component is a plain function, so everything functions do works here: `tasks.map(taskItem)` passes
it by name, and a call without `footer:` takes the default and renders no footer.

A function returns one value, and two elements side by side are two values. Wrap siblings in a
fragment, `<>...</>`, which is `Html` with no element around it. Inserting a fragment splices its
nodes in, and `<></>` is the empty `Html`:

```ant
func header(title: string): Html {
  <h1>{title}</h1> // error: this HTML expression's value is discarded; put siblings in a fragment <>...</>
  <p class="subtitle">Ant</p>
}
```

```ant
func header(title: string): Html {
  <>
    <h1>{title}</h1>
    <p class="subtitle">Ant</p>
  </>
}

println(<header>{header("Docs")}</header>)
```

```text
<header><h1>Docs</h1> <p class="subtitle">Ant</p></header>
```

Since `==` compares nodes, a test checks a component against the markup you expect, written out:

```ant
func badge(count: int): Html {
  if count == 0 { return <></> }
  <span class="badge">{count}</span>
}

assert(badge(0) == <></>)
assert(badge(3) == <span class="badge">3</span>)
```

## What escaping covers

The renderer escapes `&`, `<`, `>`, `"`, and `'` in every text node and attribute value. That makes
a string safe as text and as a quoted attribute value, wherever it came from:

```ant
let comment = "<img src=x onerror=alert(1)>"
println(<p title={comment}>{comment}</p>)
```

```text
<p title="&lt;img src=x onerror=alert(1)&gt;">&lt;img src=x onerror=alert(1)&gt;</p>
```

Escaping does not make a string safe as JavaScript, as CSS, or as a URL, so Ant takes those places
away from interpolation:

- `<script>` and `<style>` hold literal text, written as in an `.html` file, braces and all.
- Attributes that run code, every name that starts with `on`, `style`, and `srcdoc`, take only
  literal text.
- An interpolated URL attribute, such as `href` or `src`, keeps a relative URL and the schemes
  `http`, `https`, `mailto`, and `tel`. Any other scheme, `javascript:` and `data:` among them,
  becomes `about:invalid#ant`, and the page still renders.

```ant
let link = "javascript:alert(1)"
println(<a href={link}>home</a>)
println(<a href={"/users/42"}>Ann</a>)
```

```text
<a href="about:invalid#ant">home</a>
<a href="/users/42">Ann</a>
```

To hand data to a script, put it in a `data-` attribute, which is escaped like any other, and read
it from the script:

```ant
let points = [3, 1, 4]
println(<div id="chart" data-points={Json.stringify(points)}></div>)
```

```text
<div id="chart" data-points="[3,1,4]"></div>
```

Escaping is a floor, not a guarantee. Literal text you write is trusted, a URL with a safe scheme
can still point somewhere hostile, and `Html.unsafe` passes anything through. Ant does not validate
HTML either: a `<div>` inside a `<p>` is the browser's to repair.

## Html.unsafe

`Html.unsafe(s)` is the one way to turn a string into markup. The renderer writes it as it is:

```ant
let logo = Html.unsafe("<svg viewBox=\"0 0 8 8\"><circle r=\"4\" /></svg>")
println(<a href="/">{logo} Home</a>)
```

```text
<a href="/"><svg viewBox="0 0 8 8"><circle r="4" /></svg> Home</a>
```

Use it for markup your program trusts: an SVG file shipped with the program, or the output of a
sanitizer or a Markdown renderer. Never pass it a string from a user, a request, or a database you
do not control. Because it is the only door, a search for `Html.unsafe` finds every place a string
becomes markup.

The result is opaque: it renders as the markup it spells, but it is not equal to the same markup
written as an expression, and `Html.nodes` sees one raw node, not elements.

```ant
let raw = Html.unsafe("<b>x</b>")
println(raw == <b>x</b>)
println(String.from(raw) == String.from(<b>x</b>))
```

```text
false
true
```

When you have a string that holds HTML and want to show it, you do not need `Html.unsafe`: put it in
braces, and the reader sees the tags as text.

## Serving a page

`Html` is a value, not a response. To serve it, render it into the body of an `Http.Reply`. The
doctype is not markup, so it goes into the string, with the `Content-Type` header beside it:

```ant
func layout(title: string, children: Html): Html {
  <html lang="en">
    <head>
      <meta charset="utf-8" />
      <title>{title}</title>
    </head>
    <body>{children}</body>
  </html>
}

func page(status: int, title: string, children: Html): Http.Reply {
  Http.Reply{
    status,
    headers: Map{ "Content-Type": "text/html; charset=utf-8" },
    body: "<!doctype html>${layout(title, children)}",
  }
}

func handle(req: Http.Request): Http.Reply {
  if req.path != "/" {
    return page(404, "Not found", <p>No page at <code>{req.path}</code>.</p>)
  }
  page(200, "Home", <h1>Welcome</h1>)
}

let listener = try Net.listen("127.0.0.1:0")
let base = "http://${listener.address()}"
let stop = Context.new()
go { _ = Http.serve(listener, handle, stop) }

let home = try Http.get("${base}/")
println(home.headers["Content-Type"])
println(home.text())
let missing = try Http.get("${base}/%3Cscript%3E")
println("${missing.status} ${missing.text()}")
stop.cancel()
```

```text
text/html; charset=utf-8
<!doctype html><html lang="en"><head><meta charset="utf-8"> <title>Home</title></head> <body><h1>Welcome</h1></body></html>
404 <!doctype html><html lang="en"><head><meta charset="utf-8"> <title>Not found</title></head> <body><p>No page at <code>/&lt;script&gt;</code>.</p></body></html>
```

`layout` is a component like any other: every page passes its content as children. The 404 page puts
the request's path into markup, and the renderer escapes it, so a path that spells a tag reaches the
browser as text. `${...}` in a string renders an `Html` value as `String.from` does, so the body is
the doctype and the page.

`title` is a `string` because `<title>` and `<textarea>` hold text only: they take strings, numbers,
`void`, and arrays of them, but not `Html`.

A real server passes `Os.interrupt()` as the context and runs until Ctrl-C; see
[Http.serve](../lib/Http/serve.md).

## Coming from JSX and templates

Ant markup looks like JSX and HTML templates, and differs from both in a few places.

**Attribute names are HTML's.** Ant writes a name as you spell it and translates nothing, so
`className` renders as `className`, which the browser does not read as `class`. Write `class` and
`for`. `class` takes an array for the conditional classes that `clsx` builds elsewhere.

**`&&` is not the conditional.** `{ok && <b>x</b>}` mixes a `bool` and `Html`. Write
`{if ok { <b>x</b> }}`:

```ant
func flag(ok: bool): Html {
  <p>{ok && <b>yes</b>}</p>
  // error: markup takes a string, a number, void, Html, or an array of them, got bool
}
```

**Event handlers and styles are literal text.** `onclick={handler}` and `style={css}` are compile
errors, since escaping cannot make data safe as code. Put behavior in a script that reads `data-`
attributes, and style in classes:

```ant
func swatch(color: string): Html {
  <div style={"color: ${color}"}>sample</div>
  // error: style takes only literal text, since it runs code; pass data in a data- attribute that a script reads, or use a class
}
```

**A quoted attribute does not interpolate.** Templates often allow `href="/u/{{id}}"`. In Ant the
quotes hold literal text, and a brace there is a compile error; put the whole value in braces:

```ant
func profile(id: string): Html {
  <a href="/u/{id}">profile</a>
  // error: { in a quoted attribute value; interpolate with href={...}, or write &#123;
}
```

```ant
let id = "42"
println(<a href={"/u/${id}"}>profile</a>)
```

```text
<a href="/u/42">profile</a>
```

**Void elements close themselves.** HTML allows `<br>`; Ant asks for `<br />`. A comment in markup
is `{/* ... */}`, not `<!-- ... -->`. Attribute values take double quotes only.

```ant
func lines(): Html {
  <p>one<br>two</p> // error: <br> is a void element; write <br />
}
```

```ant
func note(): Html {
  <p><!-- every page -->Hi</p> // error: an HTML comment is written {/* ... */} in markup
}
```

**`key` is reserved.** A list built with `map` needs no `key`, and `key="..."` or `key={...}` is a
compile error:

```ant
func items(xs: Array[string]): Html {
  <ul>{xs.map({ <li key={x}>{x}</li> })}</ul> // error: key is reserved for a later client renderer
}
```

**Only six named references exist.** They are `&amp;`, `&lt;`, `&gt;`, `&quot;`, `&apos;`, and
`&nbsp;`. Write any other character as itself or as a numeric reference:

```ant
func footer(): Html {
  <p>&copy; 2026</p>
  // error: &copy; is not one of the named references &amp; &lt; &gt; &quot; &apos; &nbsp;; write the character itself or a numeric reference
}
```

```ant
println(<p>© 2026 &#8212; Ant</p>)
```

```text
<p>© 2026 — Ant</p>
```

**Strings are not markup.** Building HTML with string concatenation and inserting the result shows
the tags as text, because a string is always text. Build `Html` with expressions and components
instead, and keep `Html.unsafe` for markup you trust.

## See also

- [HTML literal](../language/literals/html.md): every rule of the syntax.
- [Html](../lib/Html/index.md): the type, reading nodes, and `Html.Attributes`.
- [Html.unsafe](../lib/Html/unsafe.md): trusted markup from a string.
- [Http.serve](../lib/Http/serve.md): serving requests.
