//go:build !js

package lib

// maxJSONDepth is how many arrays and objects Json.parse lets nest (stdlib.md, Json);
// json_js.go has the one on js/wasm.
const maxJSONDepth = 10000
