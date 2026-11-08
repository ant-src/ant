<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array

Array holds the functions of the built-in array type Array[T], dense and growable. xs[i] reads and
xs[i] = v writes, both panicking when i is out of range; indices are ints from 0, and ranges are
half-open, from inclusive and to exclusive. A function whose doc says it mutates xs changes it in
place; every other returns a new array.

<!-- end generated -->
