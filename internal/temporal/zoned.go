package temporal

import "sort"

// ZonedDateTime is an instant seen in a zone.
type ZonedDateTime struct {
	inst Instant
	zone *Zone
}

// Disambiguation says which instant a wall-clock time in a gap or an overlap means.
type Disambiguation int

const (
	Compatible Disambiguation = iota // the later instant in a gap, the earlier in an overlap
	Earlier
	Later
)

// Instant is the exact moment.
func (z ZonedDateTime) Instant() Instant { return z.inst }

// Zone is the zone.
func (z ZonedDateTime) Zone() *Zone { return z.zone }

// WithZone is the same instant in another zone.
func (z ZonedDateTime) WithZone(nz *Zone) ZonedDateTime { return ZonedDateTime{inst: z.inst, zone: nz} }

// OffsetSeconds is the zone's offset from UTC at the instant.
func (z ZonedDateTime) OffsetSeconds() int { return z.zone.offsetAt(z.inst) }

// Offset is the offset as ±HH:MM.
func (z ZonedDateTime) Offset() string { return formatOffset(z.OffsetSeconds()) }

// wallOf is the wall-clock date-time of an instant at an offset.
func wallOf(i Instant, offset int) PlainDateTime {
	local := i.sec + int64(offset)
	days := floorDiv(local, secondsPerDay)
	n := floorMod(local, secondsPerDay)*nanosPerSecond + int64(i.nsec)
	return PlainDateTime{dateFromDays(days), timeFromNanosOfDay(n)}
}

// PlainDateTime is the wall-clock date and time.
func (z ZonedDateTime) PlainDateTime() PlainDateTime { return wallOf(z.inst, z.OffsetSeconds()) }

// Date is the wall-clock date.
func (z ZonedDateTime) Date() PlainDate { return z.PlainDateTime().date }

// Clock is the wall-clock time.
func (z ZonedDateTime) Clock() PlainTime { return z.PlainDateTime().time }

// Compare orders by instant alone.
func (z ZonedDateTime) Compare(o ZonedDateTime) int { return z.inst.Compare(o.inst) }

// TimeType is the name of the type's tag.
func (ZonedDateTime) TimeType() string { return "Time.ZonedDateTime" }

// Key is the canonical form == compares: the instant and the zone.
func (z ZonedDateTime) Key() string { return z.inst.Key() + " " + z.zone.Key() }

// String is 2024-11-03T01:00:00-04:00[America/New_York]: the offset rounded to the minute,
// since RFC 9557 has no seconds in it, while Offset keeps them.
func (z ZonedDateTime) String() string {
	off := z.OffsetSeconds()
	return wallOf(z.inst, off).String() + formatOffset(roundToMinute(off)) + "[" + z.zone.id + "]"
}

// PossibleInstants is the instants a wall-clock time has in a zone: none in a gap, two in an
// overlap, one otherwise, in order. It is Temporal's GetPossibleEpochNanoseconds: the offsets
// in force a day before and a day after the time, each tried.
func PossibleInstants(dt PlainDateTime, z *Zone) ([]Instant, error) {
	sec, nsec := dt.asUTC()
	if z.fixed {
		i, err := NewInstant(sec-int64(z.offset), nsec)
		if err != nil {
			return nil, err
		}
		return []Instant{i}, nil
	}
	var out []Instant
	seen := map[int]bool{}
	for _, delta := range []int64{-secondsPerDay, secondsPerDay} {
		probe, err := NewInstant(sec+delta, nsec)
		if err != nil {
			probe = Instant{sec: clamp(sec+delta, -maxEpochSeconds, maxEpochSeconds)}
		}
		off := z.offsetAt(probe)
		if seen[off] {
			continue
		}
		seen[off] = true
		cand, err := NewInstant(sec-int64(off), nsec)
		if err != nil {
			return nil, err
		}
		if z.offsetAt(cand) == off {
			out = append(out, cand)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Compare(out[j]) < 0 })
	return out, nil
}

// ToZoned places a wall-clock time in a zone, resolving a gap or an overlap by d. In a gap the
// time is moved by the gap's length, back for Earlier and forward otherwise, as Temporal's
// DisambiguatePossibleEpochNanoseconds has it.
func ToZoned(dt PlainDateTime, z *Zone, d Disambiguation) (ZonedDateTime, error) {
	ps, err := PossibleInstants(dt, z)
	if err != nil {
		return ZonedDateTime{}, err
	}
	switch {
	case len(ps) == 1:
		return ZonedDateTime{ps[0], z}, nil
	case len(ps) > 1:
		if d == Later {
			return ZonedDateTime{ps[len(ps)-1], z}, nil
		}
		return ZonedDateTime{ps[0], z}, nil
	}
	sec, nsec := dt.asUTC()
	before, _ := NewInstant(sec-secondsPerDay, nsec)
	after, _ := NewInstant(sec+secondsPerDay, nsec)
	gap := int64(z.offsetAt(after) - z.offsetAt(before))
	if d == Earlier {
		shifted, err := NewInstant(sec-gap, nsec)
		if err != nil {
			return ZonedDateTime{}, err
		}
		ps, err = PossibleInstants(wallOf(shifted, 0), z)
		if err != nil || len(ps) == 0 {
			return ZonedDateTime{}, ErrRange
		}
		return ZonedDateTime{ps[0], z}, nil
	}
	shifted, err := NewInstant(sec+gap, nsec)
	if err != nil {
		return ZonedDateTime{}, err
	}
	ps, err = PossibleInstants(wallOf(shifted, 0), z)
	if err != nil || len(ps) == 0 {
		return ZonedDateTime{}, ErrRange
	}
	return ZonedDateTime{ps[len(ps)-1], z}, nil
}

// StartOfDay is the first instant of the value's date in its zone, which is not midnight on
// a day whose midnight falls in a gap.
func (z ZonedDateTime) StartOfDay() (ZonedDateTime, error) {
	return ToZoned(PlainDateTime{date: z.Date()}, z.zone, Compatible)
}

// HoursInDay is the length of the value's date in its zone: 24, 23 or 25 around a
// daylight-saving change, or 23.5 where the change is half an hour.
func (z ZonedDateTime) HoursInDay() (float64, error) {
	today, err := z.StartOfDay()
	if err != nil {
		return 0, err
	}
	next := dateFromDays(z.Date().days() + 1)
	tomorrow, err := ToZoned(PlainDateTime{date: next}, z.zone, Compatible)
	if err != nil {
		return 0, err
	}
	secs := float64(tomorrow.inst.sec-today.inst.sec) + float64(tomorrow.inst.nsec-today.inst.nsec)/nanosPerSecond
	return secs / 3600, nil
}

func clamp(v, lo, hi int64) int64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
