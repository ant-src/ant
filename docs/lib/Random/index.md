<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random

```ant signature
type Random
```

Random is a seeded generator of pseudo-random numbers. The same seed gives the same values in the
same order on every implementation of Ant, so a simulation, a test, or a game can be run again
exactly. Every function but [`Random.new`](new.md) draws from the generator and changes it, so
goroutines that share one without a lock race; [`Random.fork`](fork.md) gives each its own.
Generators compare by identity. None is fit for keys or tokens, which take
[`Crypto.randomBytes`](../Crypto/randomBytes.md).

<!-- end generated -->
