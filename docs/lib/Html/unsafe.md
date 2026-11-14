<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Html.unsafe

```ant signature
func Html.unsafe(s: string): Html
```

Html.unsafe returns s as markup, which String.from writes unescaped. It is the one way to turn a
string into markup: pass only markup the program trusts, such as a sanitizer's output.

<!-- end generated -->

## Details

- The result holds one raw node, an `Html.Raw` with `s` as its `markup`. The renderer writes it as
  it is: no escaping, and no URL check.
- `s` is not parsed or validated. A raw node is opaque: `Html.nodes` sees the raw node, not the
  elements its markup describes, and it is not `==` to the elements or text its markup spells.
- `Html.unsafe("")` is the empty `Html`, `==` to `<></>`.

## Basic examples

Turn a string into markup, which renders unescaped:

```ant
String.from(Html.unsafe("<b>x</b>")) // "<b>x</b>"
```

The same string interpolated as text is escaped:

```ant
String.from(<>{"<b>x</b>"}</>) // "&lt;b&gt;x&lt;/b&gt;"
```

Insert the result into an HTML expression like any other `Html` value:

```ant
let icon = Html.unsafe("<svg></svg>")
String.from(<button>{icon} Save</button>) // "<button><svg></svg> Save</button>"
```

## Scope

The result is one `Html.Raw` node holding the string as written:

```ant
Html.nodes(Html.unsafe("<b>x</b>")) // [Html.Raw{ markup: "<b>x</b>" }]
```

Character references stay as written; the renderer does not decode or escape them:

```ant
String.from(<p>{Html.unsafe("a &amp; b")}</p>) // "<p>a &amp; b</p>"
```

Plain text without tags is markup too, so a `<` in it reaches the page as it is:

```ant
String.from(Html.unsafe("1 < 2")) // "1 < 2"
```

Two raw nodes side by side stay two nodes. Unlike text nodes, they are not joined:

```ant
let both = <>{Html.unsafe("a")}{Html.unsafe("b")}</>
both                      // ab
Html.nodes(both).length() // 2
```

In a pipeline, the string is the first argument:

```ant
"<hr>" | Html.unsafe() // <hr>
```

## Applications

Escape user text first, then add the few tags a program allows. The result is trusted, because every
`<` from the input is already `&lt;`:

```ant
func emphasize(s: string): Html {
  let escaped = String.from(<>{s}</>)
  Html.unsafe(/\*([^*]+)\*/.replace(escaped, "<em>$1</em>"))
}

println(<p>{emphasize("a *very* <b>bold</b> claim")}</p>)
```

```text
<p>a <em>very</em> &lt;b&gt;bold&lt;/b&gt; claim</p>
```

Inline an SVG file shipped with the program:

<!-- test: file logo.svg
<svg viewBox="0 0 10 10"><circle cx="5" cy="5" r="4" /></svg>
-->

```ant
let svg = try File.read("logo.svg")
println(<a href="/" class="logo">{Html.unsafe(svg.trim())}</a>)
```

```text
<a href="/" class="logo"><svg viewBox="0 0 10 10"><circle cx="5" cy="5" r="4" /></svg></a>
```

## Properties and relations

Rendering gives back the string, whatever it holds:

```ant
for s in ["<b>x</b>", "1 < 2", "&amp;", "🐜", "<p>unclosed"] {
  assert(String.from(Html.unsafe(s)) == s)
}
```

Two raw nodes are equal when their strings are equal, even when they would render the same in a
browser:

```ant
assert(Html.unsafe("<br>") == Html.unsafe("<br>"))
assert(Html.unsafe("<br>") != Html.unsafe("<br/>"))
```

A raw node renders like the HTML expression it spells, but the two are not equal. Compare rendered
strings instead:

```ant
let raw = Html.unsafe("<b>x</b>")
assert(String.from(raw) == String.from(<b>x</b>))
assert(raw != <b>x</b>)
```

## Possible issues

`Html.unsafe` on input the program does not control lets that input run script in the page.
Interpolate the string instead, and the renderer escapes it:

```ant
let comment = "<img src=x onerror=alert(1)>"
println(<p>{comment}</p>)
println(<p>{Html.unsafe(comment)}</p>)
```

```text
<p>&lt;img src=x onerror=alert(1)&gt;</p>
<p><img src=x onerror=alert(1)></p>
```

The URL check of HTML expressions does not reach inside a raw node. An `href` written in
`Html.unsafe` is kept, whatever its scheme:

```ant
let url = "javascript:alert(1)"
String.from(<a href={url}>go</a>)                     // "<a href=\"about:invalid#ant\">go</a>"
String.from(Html.unsafe("<a href=\"${url}\">go</a>")) // "<a href=\"javascript:alert(1)\">go</a>"
```

The markup is not checked. An unclosed tag leaks into the markup around it:

```ant
String.from(<div>{Html.unsafe("<p>unclosed")}</div>) // "<div><p>unclosed</div>"
```

`Html.nodes` does not look inside a raw node, so code that walks a tree for its elements does not
find the ones in raw markup:

```ant
Html.nodes(<div>{Html.unsafe("<a href=\"/\">home</a>")}</div>)[0] is Html.Element // true
Html.nodes(Html.unsafe("<a href=\"/\">home</a>"))[0] is Html.Element              // false
```

`<title>` and `<textarea>` outside `<svg>` and `<math>` hold text only, and an attribute takes a
string, a number, or `null`, so a raw node is a compile error there:

```ant
let title = <title>{Html.unsafe("Tasks &middot; Ant")}</title>
// error: <title> and <textarea> hold text only, so they cannot hold Html
let link = <a title={Html.unsafe("x")}>y</a>
// error: attribute title takes a string, a number, or null, got Html
```

## See also

[Html](index.md) · [Html.nodes](nodes.md) · [Html.Node](types/Node.md) ·
[String.from](../String/from.md) · [HTML expressions](../../language/literals/html.md) ·
[HTML expressions guide](../../guide/html.md)
