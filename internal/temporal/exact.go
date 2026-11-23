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

// FromGoDuration is a Go duration as hours, minutes, seconds, milliseconds, microseconds and
// nanoseconds, each below the next unit, as Go prints one. Every Go duration is in range.
func FromGoDuration(d time.Duration) Duration {
	ns := int64(d)
	part := func(unit int64) int64 {
		n := ns / unit
		ns -= n * unit
		return n
	}
	return Duration{
		Hours: part(int64(time.Hour)), Minutes: part(int64(time.Minute)), Seconds: part(int64(time.Second)),
		Milliseconds: part(int64(time.Millisecond)), Microseconds: part(int64(time.Microsecond)), Nanoseconds: ns,
	}
}

// GoTime is the moment as a Go time, in the zone's location.
func (z ZonedDateTime) GoTime() time.Time { return z.inst.goTime(z.zone) }
