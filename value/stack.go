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

// Ancestors holds the containers a walk is inside, the enclosing ones of the value at hand, as a
// set, so that a deep value is walked in linear time and a cyclic one ends: rendering writes ...
// where Enter fails, and Json.stringify panics.
type Ancestors map[any]bool

// Enter adds v, or reports false when v is already being walked: a cycle.
func (a Ancestors) Enter(v any) bool {
	if a[v] {
		return false
	}
	a[v] = true
	return true
}

// Leave removes v once its walk is done.
func (a Ancestors) Leave(v any) { delete(a, v) }
