//go:build !js

package temporal

import "time"

// wall is the moment of t on the wall clock; clock_js.go says why that is not always t.
func wall(t time.Time) time.Time { return t }
