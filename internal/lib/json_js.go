//go:build js

package lib

// maxJSONDepth is how many arrays and objects Json.parse lets nest (stdlib.md, Json). On js/wasm
// the walks of a value, as == and printing, run on the JavaScript engine's stack, which a value
// some 2000 deep fills in a browser's worker, and overflowing it ends the whole instance.
const maxJSONDepth = 1000
