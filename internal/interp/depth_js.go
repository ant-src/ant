//go:build js

package interp

// maxDepth is the runtime's limit on a goroutine's call depth (ant-lang.md). On js/wasm the
// interpreter's Go frames live on the JavaScript engine's stack, which in a browser's worker
// holds some 650 to 800 calls of Ant; past it the engine ends the whole instance instead of
// panicking. So the limit leaves room below that.
const maxDepth = 500
