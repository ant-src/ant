<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel.new

```ant signature
func Channel.new[T](capacity: int = 0): Channel[T]
```

Channel.new returns a new open channel that buffers up to capacity values. A capacity of 0, the
default, makes a rendezvous channel, whose send waits for a receiver. It panics if capacity < 0 or
is above 2^31 - 1.

<!-- end generated -->
