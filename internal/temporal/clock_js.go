//go:build js

package temporal

import "time"

// On js/wasm the wall clock moves in whole milliseconds, as Date.now does, while the monotonic
// clock behind timers moves in finer steps, as performance.now does. So the wall clock is read
// once and moved on by the monotonic clock: times read from both agree, and a tick of
// Time.ticker is never before the Time.now its period is counted from.
var clockBase = time.Now()

func wall(t time.Time) time.Time { return clockBase.Add(t.Sub(clockBase)) }
