<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.cpus

```ant signature
func Os.cpus(): int
```

Os.cpus returns the number of CPUs the program may use at once: the host's logical CPUs, or fewer when a container's CPU limit or $GOMAXPROCS lowers it, and at least 1. It can change while the program runs, as a container's limit can. It is the limit to give [`Parallel.map`](../Parallel/map.md) for CPU-bound work.

<!-- end generated -->
