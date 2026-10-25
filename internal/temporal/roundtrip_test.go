package temporal

import (
	"math"
	"testing"
)

// Every value renders as its RFC 9557 string and parses back from it (stdlib.md, Time, Text):
// a spread of values of every type, the ends of the range, sub-minute offsets of local mean
// time, both instants of a repeated hour, and negative and extreme durations.
func TestRenderParsesBack(t *testing.T) {
	must := func(v any, err error) any {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	date := func(y, m, d int) PlainDate { return must(NewPlainDate(y, m, d)).(PlainDate) }
	clock := func(h, m, s, ms, us, ns int) PlainTime {
		return must(NewPlainTime(h, m, s, ms, us, ns)).(PlainTime)
	}
	zone := func(id string) *Zone { return must(LoadZone(id)).(*Zone) }

	dates := []PlainDate{
		date(2026, 10, 22), date(0, 1, 1), date(9999, 12, 31), date(10000, 1, 1), date(-1, 12, 31),
		date(1, 1, 1), date(2024, 2, 29), date(-271821, 4, 19), date(275760, 9, 14),
		date(-100000, 1, 1), date(1900, 2, 28),
	}
	times := []PlainTime{
		{}, clock(23, 59, 59, 999, 999, 999), clock(14, 3, 5, 500, 0, 0), clock(0, 0, 0, 0, 0, 1),
		clock(12, 0, 0, 0, 1, 0), clock(1, 2, 3, 120, 0, 0),
	}
	for _, d := range dates {
		if got, err := ParsePlainDate(d.String()); err != nil || got != d {
			t.Errorf("PlainDate %s parses back as %v (%v)", d, got, err)
		}
		for _, tm := range times {
			dt := PlainDateTime{d, tm}
			if got, err := ParsePlainDateTime(dt.String()); err != nil || got != dt {
				t.Errorf("PlainDateTime %s parses back as %v (%v)", dt, got, err)
			}
		}
	}
	for _, tm := range times {
		if got, err := ParsePlainTime(tm.String()); err != nil || got != tm {
			t.Errorf("PlainTime %s parses back as %v (%v)", tm, got, err)
		}
	}

	instants := []Instant{
		{}, must(NewInstant(-maxEpochSeconds, 0)).(Instant), must(NewInstant(maxEpochSeconds, 0)).(Instant),
		must(NewInstant(maxEpochSeconds-1, 999_999_999)).(Instant), must(NewInstant(-1, 1)).(Instant),
		must(NewInstant(1_792_677_785, 500_000_000)).(Instant), must(NewInstant(253_402_300_800, 0)).(Instant),
		must(NewInstant(-62_167_219_201, 0)).(Instant), // the last second of year -1
		// both instants of New York's repeated hour, and either side of its skipped one
		must(NewInstant(1_730_611_800, 0)).(Instant), must(NewInstant(1_730_615_400, 0)).(Instant),
		must(NewInstant(1_710_053_999, 999_999_999)).(Instant), must(NewInstant(1_710_054_000, 0)).(Instant),
		// local mean time: New York -04:56:02, Berlin +00:53:28, Monrovia -00:44:30, Amsterdam
		// +00:19:32, Kolkata +05:21:10 and +05:53:20
		must(NewInstant(-5_364_662_400, 0)).(Instant), must(NewInstant(-3_000_000_000, 123)).(Instant),
		must(NewInstant(-2_208_988_800, 0)).(Instant), must(NewInstant(-1_000_000_000, 0)).(Instant),
		must(NewInstant(-500_000_000, 0)).(Instant),
	}
	zones := []*Zone{
		UTC, zone("America/New_York"), zone("Europe/Berlin"), zone("Africa/Monrovia"),
		zone("Europe/Amsterdam"), zone("Asia/Kolkata"), zone("Asia/Calcutta"), zone("Australia/Lord_Howe"),
		zone("Pacific/Kiritimati"), zone("+05:30"), zone("-23:59"), zone("+00:00"), zone("-0130"), zone("+14"),
		LocalZone(),
	}
	for _, z := range zones {
		if got, err := LoadZone(z.String()); err != nil || got.Key() != z.Key() || got.String() != z.String() {
			t.Errorf("Zone %s parses back as %v (%v)", z, got, err)
		}
	}
	subMinute := false
	for _, i := range instants {
		if got, err := ParseInstant(i.String()); err != nil || got != i {
			t.Errorf("Instant %s parses back as %v (%v)", i, got, err)
		}
		for _, z := range zones {
			zdt := i.InZone(z)
			if zdt.OffsetSeconds()%60 != 0 {
				subMinute = true
			}
			if got, err := ParseZoned(zdt.String(), Compatible); err != nil {
				t.Errorf("ZonedDateTime %s does not parse back: %v", zdt, err)
			} else if got.Key() != zdt.Key() || got.String() != zdt.String() {
				t.Errorf("ZonedDateTime %s parses back as %v", zdt, got)
			}
			for _, d := range []Disambiguation{Earlier, Later} {
				if got, err := ParseZoned(zdt.String(), d); err != nil || got.Key() != zdt.Key() {
					t.Errorf("ZonedDateTime %s does not parse back to itself by disambiguation %d: %v", zdt, d, err)
				}
			}
		}
	}
	if !subMinute {
		t.Error("no offset with seconds among the zoned values")
	}

	const maxCal = 1<<32 - 1
	const maxSec = 1<<53 - 1
	durations := []Duration{
		{}, {Years: 1, Months: 2, Weeks: 3, Days: 4, Hours: 5, Minutes: 6, Seconds: 7, Milliseconds: 8, Microseconds: 9, Nanoseconds: 10},
		{Hours: -1}, {Days: -1, Nanoseconds: -1}, {Minutes: 90}, {Weeks: 1}, {Nanoseconds: 1}, {Milliseconds: -500},
		{Years: maxCal, Months: maxCal, Weeks: maxCal, Days: maxCal}, {Years: -maxCal, Months: -maxCal, Weeks: -maxCal, Days: -maxCal},
		{Seconds: maxSec, Milliseconds: 999, Microseconds: 999, Nanoseconds: 999}, {Seconds: -maxSec},
		{Hours: maxSec / 3600}, {Minutes: -(maxSec / 60)},
		{Hours: 1, Seconds: 1}, {Years: 1, Seconds: 1},
	}
	for _, d := range durations {
		if _, err := NewDuration(d.Years, d.Months, d.Weeks, d.Days, d.Hours, d.Minutes, d.Seconds, d.Milliseconds, d.Microseconds, d.Nanoseconds); err != nil {
			t.Fatalf("%v: %v", d, err)
		}
		if got, err := ParseDuration(d.String()); err != nil || got != d {
			t.Errorf("Duration %s parses back as %v (%v)", d, got, err)
		}
	}
	// The subsecond fields fold into the seconds when rendered, so a duration with more than
	// 999 of one, or one of a field's own sign beyond a second, parses back to the same string
	// and length but with the fields spread out.
	for _, d := range []Duration{{Milliseconds: 1500}, {Microseconds: -2_000_001}, {Nanoseconds: math.MaxInt64 / 2048}} {
		if _, err := NewDuration(0, 0, 0, 0, 0, 0, 0, d.Milliseconds, d.Microseconds, d.Nanoseconds); err != nil {
			t.Fatal(err)
		}
		got, err := ParseDuration(d.String())
		if c, cerr := CompareDurations(got, d, nil); err != nil || cerr != nil || got.String() != d.String() || c != 0 {
			t.Errorf("Duration %s parses back as %v (%v)", d, got, err)
		}
	}
}
