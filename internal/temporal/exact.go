package temporal

import (
	"math"
	"time"
)

// Exact is the duration as a Go duration, taken as Time.total takes one without a relativeTo:
// hours and smaller units exact and a day 24 hours. It is false when the duration has years,
// months or weeks, which have no fixed length. A duration beyond Go's range, about 292 years
// either way, is taken as that range.
func (d Duration) Exact() (time.Duration, bool) {
	if d.Years != 0 || d.Months != 0 || d.Weeks != 0 {
		return 0, false
	}
	seconds := float64(d.Days)*86400 + float64(d.Hours)*3600 + float64(d.Minutes)*60 + float64(d.Seconds) +
		float64(d.Milliseconds)/1e3 + float64(d.Microseconds)/1e6 + float64(d.Nanoseconds)/1e9
	ns := seconds * 1e9
	switch {
	case ns >= math.MaxInt64:
		return time.Duration(math.MaxInt64), true
	case ns <= math.MinInt64:
		return time.Duration(math.MinInt64), true
	}
	return time.Duration(ns), true
}
