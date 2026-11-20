//go:build !js

package interp

// maxDepth is the runtime's limit on a goroutine's call depth (ant-lang.md); depth_js.go has
// the one on js/wasm.
const maxDepth = 10000
