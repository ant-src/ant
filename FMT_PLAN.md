# Plan: `ant fmt`

The contract is `spec/fmt.md`; this file is how to get there. Each step is one commit that leaves
`go test ./...` green. Delete this file when the last step lands.

## Approach

A token printer, not a tree printer. The AST drops comments, blank lines, parentheses, and the
spelling of literals (`IntLit` keeps only the value, `RegexLit` the normalized source), and adding
all of that to the tree would touch the resolver, the checker, and the interpreter for a tool that
needs none of it. The formatter instead:

1. lexes the source in a mode that also keeps comments and each token's byte span;
2. parses it as usual, which rejects a bad file and records one fact the tokens alone cannot give:
   what each opening bracket is (a block, a list, or the `{` of a typed literal);
3. walks the tokens with a stack of open brackets and prints each one with the spacing of
   `fmt.md` 4.7, inserting or dropping only newlines, statement semicolons, and trailing commas.

Literal spelling comes for free by slicing the source at each token's span. A template string is
one token whose interpolations carry their own tokens; the printer copies its text pieces from the
source and prints each interpolation's tokens by the same rules, flat, with no space inside `${`
and `}` (`fmt.md` 4.9). The meaning guarantee (`fmt.md` section 2) becomes a cheap,
mechanical check: lex the input and the output, drop the layout tokens, compare.

## Steps

### 1. Lexer: a layout mode

- `token.Token` gains `Off, End int`, the byte span in the source, set for every token, including
  the tokens inside a template's interpolations, whose offsets are already absolute because the
  same lexer scans them. A template's `Part`s gain the spans of their text pieces.
- A `lexer.Layout` option that also returns the comments, each with its span and whether it began
  its line, and the newline count between consecutive tokens, counted in the source rather than
  from the significant `Newline` tokens. Newlines inside block comments and template text are not
  counted, since they never make a block or a list broken (`fmt.md` 4.3, 4.4). The normal mode
  stays as it is and as fast.
- Tests: spans slice back to the exact source text for every token kind, on the lexer fuzz corpus.

### 2. Parser: bracket roles

- The parser records, keyed by the opening bracket's offset, a role: `Block` (from `enterBlock`),
  `List` (from `enterList`), or `TypedBrace` (the `{` of `Point{`, `Map{`, `Geometry.Point{`), and
  for a list whether a trailing comma is allowed after its last item. The parser sets that flag at
  the place it accepts a trailing comma, so a parenthesized expression, an index, and a
  destructuring that ends in `...rest` come out as "no comma" without a list of exceptions. It is
  filled only when the caller asks for it, through a new
  `parser.ParseLayout(path, src) (*ast.File, Roles, error)`; `Parse` is unchanged.
- Tests: one table case per row of section 5, Blocks (each way a brace is decided), plus `select`
  clauses, record types, destructuring, and import lists, and a trailing-comma case for each list
  kind: allowed after call arguments and literals, refused after a parenthesized expression, an
  index, and a rest position.

### 3. `internal/format`: the printer

- `format.Source(path string, src []byte) ([]byte, error)`, in its own package so that `cmd/ant`,
  the tests, and later an editor integration share it.
- One pass over the tokens. A pre-pass decides for each bracket whether it is flat or broken
  (`fmt.md` 4.3, 4.4) from the newline counts between its direct items. The main pass keeps a
  stack of `{role, broken, lineIndent}` and, per statement and per list item, the indent of the
  line it started on. Any later line of the same statement or item is a continuation (4.5) and
  gets one level more, whether the lexer suppressed the newline (outside brackets) or the parser
  ignored it (inside a list). A line that starts with `else` takes the indent of the line that
  holds its `if` instead (4.6).
- Spacing is one function `space(prev, cur token, ctx) int`, written as the table in 4.7, in the
  same order, so the code and the spec can be read side by side.
- Comments: own-line comments take the indent of the next code line in the same bracket, trailing
  comments keep their source column, counted in code points, or fall back to one space (4.8). A
  comma is always printed before a trailing comment, never after it (4.4), and a `;` followed only
  by a comment is dropped (4.3).
- Templates: text pieces are copied from the source byte for byte; interpolation tokens go through
  the same spacing function, with source newlines kept and not reindented (4.9).
- Golden tests in `internal/format/testdata`: pairs `name.in.ant` / `name.out.ant`, one per rule
  and example of `fmt.md`, each starting with a `// spec: fmt 4.4` tag naming the section it
  proves.
  The harness also formats every `.out.ant` and requires it unchanged.

### 4. Properties over the whole corpus

A test that runs `format.Source` over every module in `examples/` and `tests/` (skipping
`tests/syntax`, whose files do not parse, and `tests/stdlib`, which are not modules) and checks:

- **tokens:** the output's tokens, without newlines, statement semicolons and trailing commas,
  equal the input's, spelling included, with a template compared by its text pieces and the tokens
  of its interpolations;
- **tree:** `ast.Sprint` of the parsed output equals `ast.Sprint` of the parsed input;
- **comments:** the same comment texts in the same order, after removing trailing whitespace from
  each line and turning `\r\n` into `\n`, the two changes `fmt.md` section 2 allows;
- **idempotence:** formatting the output changes nothing;
- **layout:** no tab in indentation, no trailing whitespace, one final newline.

The same properties as `FuzzFormat`, seeded from the parser's fuzz corpus: any input that parses
must satisfy them.

### 5. The command

- `ant fmt <path>...`, `ant fmt --check <path>...`, and `ant fmt -` in `cmd/ant/main.go`, with the
  usage text and the package comment updated, and the exit statuses of `fmt.md` section 1.
- Directory walk with `filepath.WalkDir`, lexical order, dot entries skipped, links not followed.
- In-place writes go to a temporary file in the same directory and are renamed over the original,
  keeping its mode; an unchanged file is not written at all.
- Tests in `cmd/ant/main_test.go`: each row of the exit-status table, a syntax error that leaves the
  file byte-for-byte untouched while the other files are formatted, `--check` output and its
  writing nothing, stdin to stdout with `<stdin>` in errors, and a directory with a dot directory
  and a non-`.ant` file in it.

### 6. Spec coverage

- `tests/spec_test.go` learns a third tag form, `// spec: fmt 4.4`, beside `stdlib Json`, that
  names a section of `spec/fmt.md`, and its scan learns `internal/format/testdata`.
  `TestSpecCoverage` fails when a subsection of `fmt.md` section 4 has no golden test, as it does
  for `ant-lang.md` headings today. Sections 1, 2, 3, and 5 are exempt: section 1 is tested in
  `cmd/ant`, section 2 by the properties of step 4, and sections 3 and 5 say what does not
  happen.

### 7. Format the repository

- Run `ant fmt` over `examples/` and the module directories of `tests/` (`run`, `unchecked`,
  `check`, and their `lib` folders) in one commit with nothing else in it, then `go test ./...`.
  `tests/stdlib` holds top-level statements, which `ant fmt` rejects, and stays as it is. In
  `check/`, only same-line `// ERR:` markers tie an error to a position, so blank-line changes
  are safe; only splitting `a; b // ERR: ...` can move an error away from its marker. Fix any such
  program by hand, in the same commit.
- CI runs the layout check as a Go test, `TestFormatted`, that calls `format.Source` on the same
  files and fails on any that would change. It needs no built binary.
- The README gains a line for `ant fmt` in its command list.

## Risks

- **Brace roles drifting from the parser.** The roles come from the parser's own `enterBlock` and
  `enterList` calls, not from a second classifier, so they cannot disagree with how the program
  parses.
- **Newline significance.** The printer only moves newlines where section 2 already ignores them
  (inside lists, after continuation tokens) or turns `;` into one between statements. The token
  property in step 4 and the fuzz target catch any slip, because a newline that changed meaning
  would change the parsed tree.
- **Trailing-comment columns.** Keeping the source column is simple and keeps the corpus's aligned
  `// =>` columns, but a reindented block can leave its comments ragged. Accept that in the first
  version; revisit only if it shows up in real code.
- **Templates.** Interpolations are formatted but never reindented, so a multi-line interpolation
  inside a multi-line template keeps its author's indentation. Reindenting it would mean deciding
  how the code relates to the string's own lines, and a template's text must not change.
