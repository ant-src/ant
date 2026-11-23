package value

import (
	"time"

	"github.com/ant-src/ant/internal/temporal"
)

// A host reads and builds the two time values that have a Go form: a Time.Instant is a
// time.Time and a Time.Duration a time.Duration. The others, a zone and the plain date and time
// types, stay opaque: a host passes them back.

// NewInstant returns the Time.Instant of t, a moment whatever t's location. It is an error
// beyond the range of an instant, 100 million days either side of the epoch.
func NewInstant(t time.Time) (Time, error) {
	return temporal.NewInstant(t.Unix(), int64(t.Nanosecond()))
}

// NewDuration returns d as a Time.Duration of hours, minutes, seconds, milliseconds,
// microseconds and nanoseconds, each below the next unit, as Go prints a duration: 90 seconds
// is 1m30s. Ant's == compares durations field by field, so this one is not == to the literal
// 90s; Time.compareDuration and Time.total compare lengths.
func NewDuration(d time.Duration) Time { return temporal.FromGoDuration(d) }

// GoTime returns the moment of a Time.Instant, in UTC, or of a Time.ZonedDateTime, in its zone.
// It is false for any other value.
func GoTime(v Value) (time.Time, bool) {
	switch t := v.(type) {
	case temporal.Instant:
		return t.GoTime(), true
	case temporal.ZonedDateTime:
		return t.GoTime(), true
	}
	return time.Time{}, false
}

// GoDuration returns a Time.Duration as a Go duration, as Time.total takes one without a
// relativeTo: a day is 24 hours. It is false for a duration with years, months or weeks, which
// have no fixed length, and for any other value. A duration beyond Go's range, about 292 years
// either way, is that range.
func GoDuration(v Value) (time.Duration, bool) {
	if d, ok := v.(temporal.Duration); ok {
		return d.Exact()
	}
	return 0, false
}
