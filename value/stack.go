package value

// stackChunk is how many levels of a nested value a walk takes on one goroutine's stack.
const stackChunk = 4096

// Nested runs f, the walk of a container depth levels deep. Every stackChunk levels it runs f on a
// fresh goroutine and waits for it, so that ==, rendering and Json.stringify of a value nested a
// million deep need no single Go stack deeper than the Go runtime allows (section 9: the
// interpreter never crashes). A panic in f is raised again in the caller.
func Nested(depth int, f func()) {
	if depth == 0 || depth%stackChunk != 0 {
		f()
		return
	}
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
