<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex

Regex holds the functions of regular expressions, written as a literal /.../flags or compiled from
data with [`Regex.compile`](compile.md). The syntax is RE2, so a match takes time linear in the
input; indices into the searched string count code points.

<!-- end generated -->
