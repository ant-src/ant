<!-- status: reviewed -->

# ant lsp

`ant lsp` is the language server. An editor starts it and talks to it over the Language Server
Protocol, and gets the errors of `ant check` as you type, completion, hover, navigation, rename, and
formatting with `ant fmt`.

You do not run it yourself when your editor has a plugin. The plugins for [VS Code](/vscode),
[IntelliJ IDEA](/intellij), and [Neovim](/nvim) start `ant lsp` and the debugger,
[`ant dap`](dap.md), for you. The rest of this page is for connecting another editor.

## Start the server

```sh
ant lsp
```

`ant lsp` takes no arguments. It speaks version 3.17 of the protocol on standard input and standard
output, framed with `Content-Length` headers, and writes its logs to standard error. It serves the
one client that started it, and one workspace: an editor with two workspaces starts two servers.

| Exit status | When                                                                     |
| ----------- | ------------------------------------------------------------------------ |
| `0`         | the client sent `shutdown` and then `exit`                               |
| `1`         | the client sent `exit` without `shutdown`, or the connection ended first |
| `2`         | the usage is wrong, such as `ant lsp main.ant`                           |

## Connect an editor

Register `ant lsp` as the server for `.ant` files. Beyond the standard protocol, a client handles
these:

- **Position encoding.** Offer `utf-8` in `positionEncodings` and the server counts characters in
  bytes, as the columns of `ant check` do. Otherwise it counts UTF-16 code units, the protocol's
  default.
- **File changes.** The text of an open document stands in for the file, saved or not. A module that
  is not open is read from disk at each analysis. The server does not watch the file system: a
  client that registers file watchers dynamically is asked to watch `**/*.ant`, and a change it
  reports through `workspace/didChangeWatchedFiles` is seen at once.
- **The prelude.** Going to the definition of a prelude name, such as `Array.map`, opens the
  read-only document `ant:/prelude.d.ant`. The client fetches its text with the request
  `ant/source`, which takes the URI and returns the text, and then sends requests inside it as in
  any file.
- **Test commands.** The code lenses of a test file carry the commands `ant.test.run` and
  `ant.test.debug`, which the client carries out (see
  [Run a test from its lens](#run-a-test-from-its-lens)).

## What the server offers

| Feature                     | What you get                                                            |
| --------------------------- | ----------------------------------------------------------------------- |
| Diagnostics                 | every error `ant check` reports, as you type                            |
| Completion                  | names, members, fields, named arguments, imports, types, and markup     |
| Hover                       | the kind, type, and doc comment of the name under the cursor            |
| Signature help              | the signature of the call the cursor is in, with the active parameter   |
| Definition                  | the declaration of a name, member, field, or import, prelude included   |
| References, highlight       | every use of a declaration in the file and the modules it imports       |
| Rename                      | a new name for a declaration and its uses, across importing modules     |
| Document, workspace symbols | the outline of a module, and the top-level declarations of the folders  |
| Formatting                  | `ant fmt` on the whole document                                         |
| Folding, selection ranges   | blocks, lists, markup, comment runs, imports; the enclosing expressions |
| Semantic tokens             | what each identifier is, so the editor can color it                     |
| Inlay hints                 | inferred types after `let` and `var`, and parameter names               |
| Code actions                | a missing type annotation, missing record fields, a missing import      |
| Code lenses                 | "run test" and "debug test" above each test                             |

After `|`, completion offers the functions that take the left side's type as their first argument,
so `filt` finds `Array.filter`. Where a name is not yet imported, completion and the code action
offer it from any module of the workspace that exports it. The edit adds the name to an existing
`import` of that module, writes `Namespace.name` when the module is imported as a namespace, and
otherwise adds a new `import`. Scripts, declaration files, and modules that would make an import
cycle are not offered.

The server analyzes each open file with the modules it imports, so references from a library do not
reach the files that import it. Rename does: renaming an export also renames its uses in every
module of the workspace that imports it. Rename refuses a keyword, a reserved word, a prelude name,
or a name already bound where the renamed name is visible, and does not rename record fields or
markup elements.

## Diagnostics

The server reports what `ant check` reports, with the same message, on the code the message is
about:

```ant
let total = 1 + "2" // error: + needs both operands of one type, got int and string
```

Every diagnostic is an error; there are no warnings. A syntax error does not stop the analysis: the
server reads on to the next statement, so completion and the outline keep working on a half-typed
line.

An error in an imported module shows in that module's document when it is open. The importing file
gets one diagnostic on its `import`, naming the module and its first error. An `import` of a missing
file, of a script, or in a cycle gets its error on the `import`.

The server publishes diagnostics when a document opens and after each change, and clears them when
it closes.

## Run a test from its lens

In a test file, the server puts two code lenses above each test, and above each subtest named by a
plain string. Both carry the document's URI and a `-run` pattern for [`ant test`](test.md) that
selects that test. In this file, the lens above `split` carries the pattern `^split$`, and for
`ant.test.run` the client runs `ant test -run` with the pattern on the document's file, which runs
`split` and its subtest, and not `join`:

<!-- test: cmd ant test -v -run ^split$ words_test.ant -->

```ant [words_test.ant]
test "split" {
  assert("a,b".split(",") == ["a", "b"])

  test "empty" {
    assert("".split(",") == [""])
  }
}

test "join" {
  assert(String.join(["a", "b"], ",") == "a,b")
}
```

```text
=== RUN   words_test.ant: split
=== RUN   words_test.ant: split/empty
--- PASS: words_test.ant: split/empty (0.00s)
--- PASS: words_test.ant: split (0.00s)
ok    1 passed (0.00s)
```

The pattern of the lens above `empty` names `split` and then `empty`, joined with `/`. Its part for
`empty` also matches `empty #2` and `empty #3`, so the lens of a subtest in a loop runs every run of
it. For `ant.test.debug` the client starts [`ant dap`](dap.md) with a `launch` whose `test` is the
pattern.

## What it does not do

- It does not run programs or tests; the editor runs `ant run` and `ant test`.
- It formats only a whole document, and leaves one that does not parse as it is. There is no range
  formatting and no formatting as you type.
- It has no call hierarchy and no type hierarchy, and no code lenses but those of tests.
