<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random.new

```ant signature
func Random.new(seed: int | null = null): Random
```

Random.new returns a generator seeded with seed. Without a seed, or with null, it is seeded from the
host's secure source, as [`Crypto.randomBytes`](../Crypto/randomBytes.md) is, and its values differ
from run to run.

<!-- end generated -->

## Basic examples

```ant
let r = Random.new(42)
Random.int(r, 0, 100) // 22
```
