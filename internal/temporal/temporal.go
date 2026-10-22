// Package temporal implements the time values of stdlib.md's Time namespace after JavaScript's
// Temporal: an Instant on the time line, a ZonedDateTime in a zone, the plain wall-clock types,
// zones, and durations, with nanosecond precision and Temporal's range of 100 million days each
// way from the epoch. It depends on nothing of the interpreter; the Time namespace wraps it.
// Every type is immutable and compared by value, and renders as its RFC 9557 string.
package temporal

import (
	"errors"
	"fmt"
	"math"
	"time"
)

const (
	nanosPerSecond  = 1_000_000_000
	secondsPerDay   = 86_400
	maxDays         = 100_000_000 // Temporal's range, each way from the epoch
	maxEpochSeconds = maxDays * secondsPerDay
)

// ErrRange is the error of a value outside Temporal's range.
var ErrRange = errors.New("outside the range of time values")

// Instant is an exact moment: nanoseconds since the Unix epoch, held as seconds and the
// nanosecond within the second, which spans the whole range.
type Instant struct {
	sec  int64
	nsec int32 // 0 to 999,999,999
}

// NewInstant makes an instant from seconds and nanoseconds since the epoch, the nanoseconds
// of any sign and size, and reports ErrRange beyond the range.
func NewInstant(sec, nsec int64) (Instant, error) {
	sec += floorDiv(nsec, nanosPerSecond)
	nsec = floorMod(nsec, nanosPerSecond)
	if sec < -maxEpochSeconds || sec > maxEpochSeconds || (sec == maxEpochSeconds && nsec > 0) {
		return Instant{}, ErrRange
	}
	return Instant{sec: sec, nsec: int32(nsec)}, nil
}

// Now is the current moment.
func Now() Instant {
	t := time.Now()
	return Instant{sec: t.Unix(), nsec: int32(t.Nanosecond())}
}

// FromEpochMillis makes an instant from milliseconds since the epoch.
func FromEpochMillis(ms int64) (Instant, error) {
	return NewInstant(floorDiv(ms, 1000), floorMod(ms, 1000)*1_000_000)
}

// FromEpochNanos makes an instant from nanoseconds since the epoch; every int64 is in range.
func FromEpochNanos(ns int64) Instant {
	i, _ := NewInstant(floorDiv(ns, nanosPerSecond), floorMod(ns, nanosPerSecond))
	return i
}

// EpochMillis is the milliseconds since the epoch, rounded down.
func (i Instant) EpochMillis() int64 { return i.sec*1000 + int64(i.nsec)/1_000_000 }

// EpochNanos is the nanoseconds since the epoch, or ErrRange when they do not fit an int64,
// which is outside 1677-09-21 to 2262-04-11.
func (i Instant) EpochNanos() (int64, error) {
	const limit = math.MaxInt64 / nanosPerSecond // the last whole second that fits, 2262-04-11T23:47:16
	const highNsec = math.MaxInt64 - limit*nanosPerSecond
	const lowNsec = math.MinInt64 + limit*nanosPerSecond + nanosPerSecond // within the second before -limit
	sec, nsec := i.sec, int64(i.nsec)
	switch {
	case sec > limit || sec < -limit-1:
		return 0, ErrRange
	case sec == limit:
		if nsec > highNsec {
			return 0, ErrRange
		}
	case sec == -limit-1:
		if nsec < lowNsec {
			return 0, ErrRange
		}
		return math.MinInt64 + (nsec - lowNsec), nil
	}
	return sec*nanosPerSecond + nsec, nil
}

// Compare orders instants: -1, 0 or 1.
func (i Instant) Compare(j Instant) int {
	switch {
	case i.sec < j.sec || i.sec == j.sec && i.nsec < j.nsec:
		return -1
	case i.sec > j.sec || i.sec == j.sec && i.nsec > j.nsec:
		return 1
	}
	return 0
}

// InZone is the instant seen in a zone.
func (i Instant) InZone(z *Zone) ZonedDateTime { return ZonedDateTime{inst: i, zone: z} }

// TimeType is the name of the type's tag.
func (Instant) TimeType() string { return "Time.Instant" }

// Key is the canonical form == compares.
func (i Instant) Key() string { return fmt.Sprintf("%d.%09d", i.sec, i.nsec) }

// String is the RFC 9557 form in UTC: 2026-10-22T14:03:05.5Z.
func (i Instant) String() string {
	return wallOf(i, 0).String() + "Z"
}

func floorDiv(a, b int64) int64 {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

func floorMod(a, b int64) int64 { return a - floorDiv(a, b)*b }
