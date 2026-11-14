<!-- status: reviewed -->

# HTML expressions

Markup is an expression in Ant. `<p>Hello, {name}</p>` is a value of type `Html`, which you build,
pass around, compare, and print like any other value.

```ant
type Task = { title: string, done: bool }

func item(task: Task): Html {
  <li class={["task", if task.done { "done" }]}>
    <input type="checkbox" checked={task.done} disabled />
    {task.title}
  </li>
}

func page(owner: string, tasks: Array[Task]): Html {
  <section>
    <h1>{owner}'s tasks</h1>
    {if tasks.length() == 0 {
      <p>Nothing to do.</p>
    } else {
      <ul>{tasks.map(item)}</ul>
    }}
  </section>
}

let tasks = [
  Task{ title: "Buy milk & eggs", done: true },
  Task{ title: "Fix <script> tag", done: false },
]
println(page("Ann", tasks))
println(page("Bob", []))

let first = <li class="task done">
  <input type="checkbox" checked disabled />
  Buy milk &amp; eggs
</li>
assert(item(tasks[0]) == first)

let link = "javascript:alert(1)"
println(<a href={link}>home</a>)
```

```text
<section><h1>Ann&#39;s tasks</h1> <ul><li class="task done"><input type="checkbox" checked disabled> Buy milk &amp; eggs</li><li class="task"><input type="checkbox" disabled> Fix &lt;script&gt; tag</li></ul></section>
<section><h1>Bob&#39;s tasks</h1> <p>Nothing to do.</p></section>
<a href="about:invalid#ant">home</a>
```

`item` and `page` are components: plain functions that return `Html`. Braces put a value into
markup. A string becomes text, and the renderer escapes it, so the title `Fix <script> tag` reaches
the page as text and never as a tag. An `Html` value goes in as it is, which is how `page` nests the
items. An array puts in each element, so `tasks.map(item)` is the loop, and an `if` picks between
two pieces of markup.

Attributes take values too. `checked={task.done}` writes `checked` when the task is done and leaves
it out when it is not. `class` takes an array and joins the strings in it; the `if` without `else`
adds `done` only for a finished task.

The source is indented to match the code, and the output carries none of it. White space with a line
break in it disappears at the start and the end of an element's content, and becomes one space
anywhere else.

Two `Html` values are equal when they hold the same nodes, so markup built from a template and
markup written out by hand compare equal, and `assert` can check a component's output against the
markup you expect. The last line shows a URL attribute: Ant replaces a link whose scheme is not
`http`, `https`, `mailto`, or `tel` with `about:invalid#ant`, so a `javascript:` link cannot run,
and the page still renders.

## Try it

Put a value into markup that has no single way to show, and the type checker stops you:

```ant
type Task = { title: string, done: bool }

func item(task: Task): Html {
  <li>{task}</li> // error: markup takes a string, a number, void, Html, or an array of them, got Task
}

func heading(name: string | null): Html {
  <h1>{name}</h1>
  // error: markup takes a string, a number, void, Html, or an array of them, got string | null; decide what null shows, as {x ?? "none"}, or use an if
}
```

Write the decision instead: `<li>{task.title}</li>`, and `<h1>{name ?? "guest"}</h1>`.

A component that returns two elements side by side does not compile:

```ant
func card(title: string, body: string): Html {
  <h2>{title}</h2> // error: this HTML expression's value is discarded; put siblings in a fragment <>...</>
  <p>{body}</p>
}
```

Wrap siblings in a fragment, `<>...</>`, which is `Html` with no element around it.

Code in markup is literal text only, so data cannot become JavaScript:

```ant
let handler = "alert(1)"
let button = <button onclick={handler}>Go</button>
// error: onclick takes only literal text, since it runs code; pass data in a data- attribute that a script reads, or use a class
```

The guide chapter [HTML expressions](../guide/html.md) covers attributes, fragments, white space,
and serving a page. The reference pages are the [HTML literal](../language/literals/html.md), the
[`Html`](../lib/Html/index.md) namespace, and [`Html.unsafe`](../lib/Html/unsafe.md), the one way to
insert markup from a string.
