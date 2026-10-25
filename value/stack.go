package value

// stackChunk is how many levels of a nested value a walk takes on one goroutine's stack.
const stackChunk = 4096

// Deep reports whether a walk of a container depth levels down continues on a fresh goroutine,
// which it does every stackChunk levels, so that ==, rendering and Json.stringify of a value
// nested a million deep need no single Go stack deeper than the Go runtime allows (section 9: the
// interpreter never crashes).
func Deep(depth int) bool {
	return depth > 0 && depth%stackChunk == 0
}

// OnFreshStack runs f on a fresh goroutine and waits for it. A panic in f is raised again in the
// caller. Callers build f only once Deep says so, which keeps the closure off the common path.
// Anything f shares with the caller, such as the set of enclosing containers that ends a walk of
// a cycle, can no longer live on the stack, so hot walks take that set from a sync.Pool.
func OnFreshStack(f func()) {
	var p any
	done := make(chan struct{})
	go func() {
		defer func() {
			p = recover()
			close(done)
		}()
		f()
	}()
	<-done
	if p != nil {
		panic(p)
	}
}
