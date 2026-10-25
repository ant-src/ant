package temporal

import (
	"math"
	"strings"
	"testing"
)

func dur(t *testing.T, s string) Duration {
	t.Helper()
	d, err := ParseDuration(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func zdt(t *testing.T, s string) ZonedDateTime {
	t.Helper()
	z, err := ParseZoned(s, Compatible)
	if err != nil {
		t.Fatal(err)
	}
	return z
}

func date(t *testing.T, s string) PlainDate {
	t.Helper()
	d, err := ParsePlainDate(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func check(t *testing.T, what string, got any, err error, want string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: %v", what, err)
		return
	}
	if s, ok := got.(interface{ String() string }); ok {
		if s.String() != want {
			t.Errorf("%s: %s, want %s", what, s, want)
		}
	}
}

func TestAdd(t *testing.T) {
	combined := Duration{Years: 1, Months: 2, Weeks: 3, Days: 4, Hours: 5, Minutes: 6, Seconds: 7, Milliseconds: 8}
	r, err := zdt(t, "2021-11-01T12:34:56-04:00[America/New_York]").Add(combined)
	check(t, "combined", r, err, "2023-01-26T17:41:03.008-05:00[America/New_York]")
	r, err = zdt(t, "2024-03-09T02:05:00-05:00[America/New_York]").Add(dur(t, "P1D"))
	check(t, "a day across the gap", r, err, "2024-03-10T03:05:00-04:00[America/New_York]")
	r, err = zdt(t, "2024-11-02T01:00:00-04:00[America/New_York]").Add(dur(t, "P1D"))
	check(t, "a day across the overlap", r, err, "2024-11-03T01:00:00-04:00[America/New_York]")
	r, err = zdt(t, "2024-11-03T01:00:00-04:00[America/New_York]").Add(dur(t, "PT1H"))
	check(t, "an hour across the overlap", r, err, "2024-11-03T01:00:00-05:00[America/New_York]")
	back, _ := dur(t, "P1D").Negate()
	r, err = zdt(t, "2024-11-03T01:00:00-04:00[America/New_York]").Add(back)
	check(t, "a day back", r, err, "2024-11-02T01:00:00-04:00[America/New_York]")

	d, err := date(t, "2024-08-31").Add(dur(t, "P1M"))
	check(t, "August 31 plus a month", d, err, "2024-09-30")
	d, err = date(t, "2019-01-31").Add(dur(t, "P1M"))
	check(t, "January 31 plus a month", d, err, "2019-02-28")
	d, err = date(t, "2024-01-31").Add(dur(t, "P1M1D"))
	check(t, "a month and a day", d, err, "2024-03-01")
	d, err = date(t, "2024-01-31").Add(dur(t, "P1Y1W"))
	check(t, "a year and a week", d, err, "2025-02-07")
	d, err = date(t, "2024-01-31").Add(dur(t, "PT47H"))
	check(t, "47 hours is a day", d, err, "2024-02-01")
	d, err = date(t, "2024-01-31").Add(dur(t, "-PT25H"))
	check(t, "minus 25 hours", d, err, "2024-01-30")

	tm, err := (PlainTime{Hour: 23}).Add(dur(t, "PT2H"))
	check(t, "time wraps", tm, err, "01:00:00")
	tm, err = (PlainTime{Hour: 23}).Add(dur(t, "P1DT2H"))
	check(t, "time ignores days", tm, err, "01:00:00")
	dt, err := dt(t, "2026-10-22T23:00").Add(dur(t, "PT2H"))
	check(t, "date-time carries", dt, err, "2026-10-23T01:00:00")
	dt, err = dt.Add(dur(t, "-P1M"))
	check(t, "date-time back a month", dt, err, "2026-09-23T01:00:00")

	i, _ := ParseInstant("2026-10-22T14:00:00Z")
	if _, err := i.Add(dur(t, "P1D")); err == nil {
		t.Error("a day added to an instant")
	}
	j, err := i.Add(dur(t, "PT1H30M"))
	check(t, "instant plus 90 minutes", j, err, "2026-10-22T15:30:00Z")
}

func TestUntil(t *testing.T) {
	start := zdt(t, "2024-11-03T01:00:00-04:00[America/New_York]")
	end := zdt(t, "2024-11-04T01:00:00-05:00[America/New_York]")
	r, err := start.Until(end, Hour)
	check(t, "across the overlap", r, err, "PT25H")
	r, err = start.Until(end, Day)
	check(t, "across the overlap in days", r, err, "P1D")
	r, err = end.Until(start, Day)
	check(t, "back across the overlap", r, err, "-P1D")
	start2 := zdt(t, "2024-03-10T01:00:00-05:00[America/New_York]")
	end2 := zdt(t, "2024-03-11T01:00:00-04:00[America/New_York]")
	r, err = start2.Until(end2, Hour)
	check(t, "across the gap", r, err, "PT23H")
	r, err = start2.Until(end2, Day)
	check(t, "across the gap in days", r, err, "P1D")
	lima := zdt(t, "2024-11-04T01:00:00-05:00[America/Lima]")
	r, err = start.Until(lima, Hour)
	check(t, "to Lima in hours", r, err, "PT25H")
	if _, err := start.Until(lima, Day); err == nil {
		t.Error("days between zones")
	}
	r, err = start.Until(lima.WithZone(start.Zone()), Day)
	check(t, "Lima seen in New York", r, err, "P1D")
	r, err = start.WithZone(lima.Zone()).Until(lima, Day)
	check(t, "New York seen in Lima", r, err, "P1DT1H")

	mar31, feb29 := date(t, "2024-03-31"), date(t, "2024-02-29")
	r, err = mar31.Until(feb29, Month)
	check(t, "March 31 until February 29", r, err, "-P1M")
	r, err = feb29.Until(mar31, Month)
	check(t, "February 29 until March 31", r, err, "P1M2D")
	r, err = date(t, "2024-01-01").Until(date(t, "2024-02-01"), Month)
	check(t, "a month", r, err, "P1M")
	r, err = date(t, "2022-01-01").Until(date(t, "2022-01-28"), Day)
	check(t, "27 days", r, err, "P27D")
	r, err = date(t, "2022-01-01").Until(date(t, "2022-01-28"), Week)
	check(t, "weeks", r, err, "P3W6D")
	r, err = date(t, "2020-02-29").Until(date(t, "2024-02-28"), Year)
	check(t, "years with a leap day", r, err, "P3Y11M30D")
	// At a month's end a month has passed only when the day is reached, as Temporal has it.
	r, err = date(t, "2023-01-31").Until(date(t, "2023-02-28"), Month)
	check(t, "January 31 until February 28", r, err, "P28D")
	r, err = date(t, "2024-01-31").Until(date(t, "2024-02-29"), Month)
	check(t, "January 31 until February 29", r, err, "P29D")
	r, err = date(t, "2024-03-31").Until(date(t, "2024-04-30"), Month)
	check(t, "March 31 until April 30", r, err, "P30D")
	r, err = date(t, "2024-02-29").Until(date(t, "2025-02-28"), Year)
	check(t, "a leap day until the next February 28", r, err, "P11M30D")
	r, err = date(t, "2023-02-28").Until(date(t, "2023-01-31"), Month)
	check(t, "February 28 back to January 31", r, err, "-P28D")
	r, err = date(t, "2023-01-30").Until(date(t, "2023-02-28"), Month)
	check(t, "January 30 until February 28", r, err, "P29D")
	r, err = date(t, "2023-01-28").Until(date(t, "2023-02-28"), Month)
	check(t, "January 28 until February 28", r, err, "P1M")
	r, err = dur(t, "P29D").Round(0, false, Month, true, 1, HalfExpand, date(t, "2024-01-31"))
	check(t, "29 days from January 31 in months", r, err, "P29D")
	if _, err := mar31.Until(feb29, Hour); err == nil {
		t.Error("hours between dates")
	}

	a, b := dt(t, "2026-10-22T10:00"), dt(t, "2026-10-24T09:00")
	r, err = a.Until(b, Day)
	check(t, "date-times in days", r, err, "P1DT23H")
	r, err = a.Until(b, Hour)
	check(t, "date-times in hours", r, err, "PT47H")
	r, err = b.Until(a, Day)
	check(t, "date-times back", r, err, "-P1DT23H")

	i, _ := ParseInstant("2026-10-22T14:00:00Z")
	j, _ := ParseInstant("2026-10-22T15:30:00Z")
	r, err = i.Until(j, Second)
	check(t, "instants in seconds", r, err, "PT5400S")
	r, err = i.Until(j, Hour)
	check(t, "instants in hours", r, err, "PT1H30M")
	if _, err := i.Until(j, Day); err == nil {
		t.Error("days between instants")
	}
	r, err = (PlainTime{Hour: 23}).Until(PlainTime{Hour: 1}, Hour)
	check(t, "times", r, err, "-PT22H")
}

func TestWith(t *testing.T) {
	three := 3
	r, err := zdt(t, "2024-11-02T01:05:00-04:00[America/New_York]").With(Fields{Day: &three}, Compatible)
	check(t, "with from before the overlap", r, err, "2024-11-03T01:05:00-04:00[America/New_York]")
	r, err = zdt(t, "2024-11-04T01:05:00-05:00[America/New_York]").With(Fields{Day: &three}, Compatible)
	check(t, "with from after the overlap", r, err, "2024-11-03T01:05:00-05:00[America/New_York]")
	ten := 10
	hour := 2
	minute := 30
	r, err = zdt(t, "2024-03-09T12:00:00-05:00[America/New_York]").With(Fields{Day: &ten, Hour: &hour, Minute: &minute}, Earlier)
	check(t, "with into the gap, earlier", r, err, "2024-03-10T01:30:00-05:00[America/New_York]")
	two, thirteen, zero := 2, 13, 0
	d, err := date(t, "2024-01-31").With(Fields{Month: &two})
	check(t, "constrained day", d, err, "2024-02-29")
	d, err = date(t, "2024-01-31").With(Fields{Month: &thirteen})
	check(t, "constrained month", d, err, "2024-12-31")
	if _, err := date(t, "2024-01-31").With(Fields{Day: &zero}); err == nil {
		t.Error("day 0 is not constrained up, as in Temporal")
	}
	if _, err := date(t, "2024-01-31").With(Fields{Hour: &two}); err == nil {
		t.Error("an hour on a date")
	}
	big := 25
	tm, err := (PlainTime{Hour: 1}).With(Fields{Hour: &big})
	check(t, "clamped hour", tm, err, "23:00:00")
	if _, err := (PlainTime{}).With(Fields{Year: &two}); err == nil {
		t.Error("a year on a time")
	}
	pdt, err := dt(t, "2024-01-31T10:00").With(Fields{Month: &two, Minute: &big})
	check(t, "date-time with", pdt, err, "2024-02-29T10:25:00")
}

func TestRound(t *testing.T) {
	base := dt(t, "2021-07-01T12:34:56.123456789")
	r, err := base.Round(Millisecond, 1, HalfExpand)
	check(t, "to the millisecond", r, err, "2021-07-01T12:34:56.123")
	r, err = base.Round(Minute, 30, HalfExpand)
	check(t, "to the half hour", r, err, "2021-07-01T12:30:00")
	r, err = base.Round(Day, 1, Ceil)
	check(t, "up to the day", r, err, "2021-07-02T00:00:00")
	r, err = base.Round(Hour, 1, Floor)
	check(t, "down to the hour", r, err, "2021-07-01T12:00:00")
	z, err := zdt(t, "2024-03-10T01:00:00-05:00[America/New_York]").Round(Hour, 2, HalfExpand)
	check(t, "into the gap", z, err, "2024-03-10T03:00:00-04:00[America/New_York]")
	z, err = zdt(t, "2021-07-01T12:00:00-04:00[America/New_York]").Round(Day, 1, Ceil)
	check(t, "zoned up to the day", z, err, "2021-07-02T00:00:00-04:00[America/New_York]")
	z, err = zdt(t, "2024-11-03T12:00:00-05:00[America/New_York]").Round(Day, 1, HalfExpand)
	check(t, "a 25-hour day, past half", z, err, "2024-11-04T00:00:00-05:00[America/New_York]")
	z, err = zdt(t, "2024-11-03T01:30:00-05:00[America/New_York]").Round(Hour, 1, HalfExpand)
	check(t, "keeps the offset in the overlap", z, err, "2024-11-03T02:00:00-05:00[America/New_York]")
	i, _ := ParseInstant("2026-10-22T14:03:05.5Z")
	j, err := i.Round(Second, 1, HalfExpand)
	check(t, "instant to the second", j, err, "2026-10-22T14:03:06Z")
	j, err = i.Round(Minute, 15, HalfExpand)
	check(t, "instant to the quarter hour", j, err, "2026-10-22T14:00:00Z")
	j, err = i.Round(Hour, 1, HalfEven)
	check(t, "instant half even", j, err, "2026-10-22T14:00:00Z")
	tm, err := (PlainTime{Hour: 23, Minute: 59, Second: 59, Nano: 900_000_000}).Round(Second, 1, HalfExpand)
	check(t, "time wraps at midnight", tm, err, "00:00:00")
	for _, bad := range []error{
		func() error { _, err := base.Round(Hour, 5, HalfExpand); return err }(),
		func() error { _, err := base.Round(Day, 2, HalfExpand); return err }(),
		func() error { _, err := base.Round(Month, 1, HalfExpand); return err }(),
		func() error { _, err := (PlainTime{}).Round(Day, 1, HalfExpand); return err }(),
		func() error { _, err := i.Round(Day, 1, HalfExpand); return err }(),
	} {
		if bad == nil {
			t.Error("a bad rounding was accepted")
		}
	}
}

func TestDurationText(t *testing.T) {
	cases := map[string]Duration{
		"PT1H30M15S":       {Hours: 1, Minutes: 30, Seconds: 15},
		"PT9.58S":          {Seconds: 9, Milliseconds: 580},
		"PT0S":             {},
		"-PT1H":            {Hours: -1},
		"PT1.5S":           {Milliseconds: 1500},
		"P2D":              {Days: 2},
		"PT0.000000001S":   {Nanoseconds: 1},
		"P1Y2M3W4DT5H6M7S": {Years: 1, Months: 2, Weeks: 3, Days: 4, Hours: 5, Minutes: 6, Seconds: 7},
		"-P1M":             {Months: -1},
	}
	for want, d := range cases {
		if d.String() != want {
			t.Errorf("%v renders as %s, want %s", d, d, want)
		}
	}
	parsed := map[string]Duration{
		"P1Y2M3DT4H5M6.5S": {Years: 1, Months: 2, Days: 3, Hours: 4, Minutes: 5, Seconds: 6, Milliseconds: 500},
		"PT1.5H":           {Hours: 1, Minutes: 30},
		"PT0.5M":           {Seconds: 30},
		"-PT1H":            {Hours: -1},
		"PT0S":             {},
		"P1W":              {Weeks: 1},
		"+P1D":             {Days: 1},
		"PT0.000000001S":   {Nanoseconds: 1},
		"P1DT36H":          {Days: 1, Hours: 36},
	}
	for s, want := range parsed {
		got, err := ParseDuration(s)
		if err != nil || got != want {
			t.Errorf("%s parses to %v (%v), want %v", s, got, err, want)
		}
	}
	for _, bad := range []string{"", "P", "PT", "1H", "P1H", "PT1.5H30M", "P1D2Y", "P1.5D", "PT1S1M", "P1DT", "PT1.1234567890S", "P-1D", "PT1H.5S"} {
		if _, err := ParseDuration(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	if _, err := NewDuration(1, 0, 0, -1, 0, 0, 0, 0, 0, 0); err == nil {
		t.Error("mixed signs")
	}
	if _, err := NewDuration(1<<32, 0, 0, 0, 0, 0, 0, 0, 0, 0); err == nil {
		t.Error("2^32 years")
	}
	if _, err := NewDuration(0, 0, 0, 0, 1<<53, 0, 0, 0, 0, 0); err == nil {
		t.Error("2^53 hours")
	}
	if _, err := NewDuration(0, 0, 0, 0, 0, 0, 1<<53-1, 0, 0, 999_999_999); err != nil {
		t.Error("the largest time part", err)
	}
	a, b := dur(t, "PT90M"), dur(t, "PT1H30M")
	neg, _ := a.Negate()
	abs, _ := neg.Abs()
	if a.Key() == b.Key() || a.Sign() != 1 || neg.Sign() != -1 || abs != a || a.Part(Minute) != 90 {
		t.Error("key, sign, negate, abs, part")
	}
	// -2^63 nanoseconds is a duration whose negation has no field to hold it.
	if least, err := NewDuration(0, 0, 0, 0, 0, 0, 0, 0, 0, math.MinInt64); err != nil {
		t.Error("-2^63 nanoseconds", err)
	} else if n, err := least.Negate(); err == nil {
		t.Errorf("-2^63 nanoseconds negated to %v", n)
	}
	if (Duration{}).Sign() != 0 || (Duration{}).defaultLargestUnit() != Nanosecond {
		t.Error("the zero duration")
	}
}

func TestDurationRound(t *testing.T) {
	round := func(s string, smallest Unit, hasSmallest bool, largest Unit, hasLargest bool, inc int64, mode RoundingMode, rel RelativeTo) (Duration, error) {
		return dur(t, s).Round(smallest, hasSmallest, largest, hasLargest, inc, mode, rel)
	}
	r, err := round("PT1H30M15S", Minute, true, 0, false, 1, HalfExpand, nil)
	check(t, "to minutes", r, err, "PT1H30M")
	r, err = round("P3DT1H41M5S", 0, false, Hour, true, 1, HalfExpand, nil)
	check(t, "largest hours", r, err, "PT73H41M5S")
	r, err = round("P1DT1H30M", Hour, true, Hour, true, 1, Floor, nil)
	check(t, "whole hours down", r, err, "PT25H")
	r, err = round("PT1H17M", Minute, true, 0, false, 15, HalfExpand, nil)
	check(t, "quarter hours", r, err, "PT1H15M")
	r, err = round("PT80M90S", 0, false, Minute, true, 1, HalfExpand, nil)
	check(t, "balanced to minutes", r, err, "PT81M30S")
	r, err = round("PT80M90S", 0, false, Hour, true, 1, HalfExpand, nil)
	check(t, "balanced to hours", r, err, "PT1H21M30S")
	r, err = round("PT80M30S", Nanosecond, true, 0, false, 1, HalfExpand, nil)
	check(t, "unchanged", r, err, "PT80M30S")
	r, err = round("PT48H", 0, false, Day, true, 1, HalfExpand, nil)
	check(t, "to days", r, err, "P2D")
	r, err = round("PT36H", Day, true, 0, false, 1, HalfExpand, nil)
	check(t, "rounded to days", r, err, "P2D")
	if _, err := round("P370D", 0, false, Year, true, 1, HalfExpand, nil); err == nil {
		t.Error("years without a point")
	}
	if _, err := round("P1M", Day, true, 0, false, 1, HalfExpand, nil); err == nil {
		t.Error("months without a point")
	}
	if _, err := round("PT1H", 0, false, 0, false, 1, HalfExpand, nil); err == nil {
		t.Error("no unit")
	}
	if _, err := round("PT1H", Hour, true, Minute, true, 1, HalfExpand, nil); err == nil {
		t.Error("largest below smallest")
	}
	r, err = round("P370D", 0, false, Year, true, 1, HalfExpand, date(t, "2019-01-01"))
	check(t, "370 days from 2019", r, err, "P1Y5D")
	r, err = round("P370D", 0, false, Year, true, 1, HalfExpand, date(t, "2020-01-01"))
	check(t, "370 days from a leap year", r, err, "P1Y4D")
	la := zdt(t, "2020-03-08T00:00:00-08:00[America/Los_Angeles]")
	r, err = round("PT48H", 0, false, Day, true, 1, HalfExpand, la)
	check(t, "48 hours across a 23-hour day", r, err, "P2DT1H")
	r, err = round("P1M1DT1H", Day, true, Day, true, 1, HalfExpand, date(t, "2022-01-01"))
	check(t, "a month and a day in days", r, err, "P32D")
	r, err = round("P1M", Day, true, Day, true, 1, HalfExpand, date(t, "2022-02-01"))
	check(t, "February in days", r, err, "P28D")
	r, err = round("P45D", Month, true, 0, false, 1, HalfExpand, date(t, "2021-01-01"))
	check(t, "45 days to months", r, err, "P2M")
	r, err = round("P45D", Month, true, 0, false, 1, Trunc, date(t, "2021-01-01"))
	check(t, "45 days truncated to months", r, err, "P1M")
	r, err = round("P1Y11M", Year, true, 0, false, 1, HalfExpand, date(t, "2021-01-01"))
	check(t, "bubbled up to a year", r, err, "P2Y")
	r, err = round("P13M", 0, false, Year, true, 1, HalfExpand, date(t, "2021-01-01"))
	check(t, "balanced to years", r, err, "P1Y1M")
	r, err = round("P1DT25H", Hour, true, 0, false, 1, HalfExpand, la)
	check(t, "zoned hours", r, err, "P2DT1H") // the hours are exact, after the day in the zone

	s, err := AddDurations(dur(t, "PT26H45M"), dur(t, "PT30M"))
	check(t, "sum", s, err, "PT27H15M")
	s, err = AddDurations(dur(t, "PT80M90S"), dur(t, "PT100M15S"))
	check(t, "sum of minutes", s, err, "PT181M45S")
	s2, err := s.Round(0, false, Hour, true, 1, HalfExpand, nil)
	check(t, "sum balanced", s2, err, "PT3H1M45S")
	s, err = AddDurations(dur(t, "PT48H"), dur(t, "PT24H"))
	check(t, "sum in hours", s, err, "PT72H")
	s2, err = s.Round(0, false, Day, true, 1, HalfExpand, nil)
	check(t, "sum in days", s2, err, "P3D")
	s2, err = s.Round(0, false, Day, true, 1, HalfExpand, la)
	check(t, "sum in zoned days", s2, err, "P3DT1H")
	s, err = AddDurations(dur(t, "P1D"), dur(t, "PT25H"))
	check(t, "days plus hours", s, err, "P2DT1H")
	if _, err := AddDurations(dur(t, "P1M"), dur(t, "P1D")); err == nil {
		t.Error("months added")
	}
}

func TestDurationTotalAndCompare(t *testing.T) {
	f, err := dur(t, "PT1H30M").Total(Minute, nil)
	if err != nil || f != 90 {
		t.Errorf("minutes: %v %v", f, err)
	}
	f, err = dur(t, "PT1H30M").Total(Hour, nil)
	if err != nil || f != 1.5 {
		t.Errorf("hours: %v %v", f, err)
	}
	f, err = dur(t, "P1M").Total(Day, date(t, "2021-01-01"))
	if err != nil || f != 31 {
		t.Errorf("a month in days: %v %v", f, err)
	}
	f, err = dur(t, "P45D").Total(Month, date(t, "2021-01-01"))
	if err != nil || f != 1.5 {
		t.Errorf("45 days in months: %v %v", f, err)
	}
	f, err = dur(t, "P1D").Total(Hour, zdt(t, "2024-03-10T00:00:00-05:00[America/New_York]"))
	if err != nil || f != 23 {
		t.Errorf("a short day in hours: %v %v", f, err)
	}
	f, err = dur(t, "PT36H").Total(Day, zdt(t, "2024-03-10T00:00:00-05:00[America/New_York]"))
	if err != nil || f != 1+13.0/24 { // a 23-hour day, then 13 hours of a 24-hour one
		t.Errorf("36 hours in zoned days: %v %v", f, err)
	}
	if _, err := dur(t, "P1M").Total(Day, nil); err == nil {
		t.Error("months without a point")
	}
	if _, err := dur(t, "P1D").Total(Month, nil); err == nil {
		t.Error("months without a point")
	}
	c, err := CompareDurations(dur(t, "PT1H"), dur(t, "PT60M"), nil)
	if err != nil || c != 0 {
		t.Errorf("an hour and sixty minutes: %d %v", c, err)
	}
	c, err = CompareDurations(dur(t, "P1M"), dur(t, "P30D"), date(t, "2021-01-01"))
	if err != nil || c != 1 {
		t.Errorf("a month and thirty days: %d %v", c, err)
	}
	c, err = CompareDurations(dur(t, "P1D"), dur(t, "PT24H"), zdt(t, "2020-03-08T00:00:00-08:00[America/Los_Angeles]"))
	if err != nil || c != -1 {
		t.Errorf("a short day and 24 hours: %d %v", c, err)
	}
	c, err = CompareDurations(dur(t, "P1D"), dur(t, "PT24H"), nil)
	if err != nil || c != 0 {
		t.Errorf("a day and 24 hours: %d %v", c, err)
	}
	if _, err := CompareDurations(dur(t, "P1M"), dur(t, "P30D"), nil); err == nil {
		t.Error("months compared without a point")
	}
}

func TestTransitions(t *testing.T) {
	z := zdt(t, "2024-01-01T00:00:00-05:00[America/New_York]")
	n, ok := z.NextTransition()
	if !ok || n.String() != "2024-03-10T03:00:00-04:00[America/New_York]" {
		t.Errorf("next: %v %v", n, ok)
	}
	n2, ok := n.NextTransition()
	if !ok || n2.String() != "2024-11-03T01:00:00-05:00[America/New_York]" {
		t.Errorf("next after next: %v %v", n2, ok)
	}
	p, ok := z.PreviousTransition()
	if !ok || p.String() != "2023-11-05T01:00:00-05:00[America/New_York]" {
		t.Errorf("previous: %v %v", p, ok)
	}
	p2, ok := n.PreviousTransition()
	if !ok || p2.String() != "2023-11-05T01:00:00-05:00[America/New_York]" {
		t.Errorf("previous from a transition: %v %v", p2, ok)
	}
	if _, ok := zdt(t, "2024-01-01T00:00:00+00:00[UTC]").NextTransition(); ok {
		t.Error("UTC has a transition")
	}
	if _, ok := zdt(t, "2024-01-01T00:00:00+05:30[+05:30]").PreviousTransition(); ok {
		t.Error("an offset zone has a transition")
	}
	far := zdt(t, "2100-06-01T00:00:00-04:00[America/New_York]")
	if n, ok := far.NextTransition(); !ok || !strings.HasPrefix(n.String(), "2100-11-07T01:00:00-05:00") {
		t.Errorf("a transition from the rule beyond the table: %v %v", n, ok)
	}
	// Go's bounds lag a day at the end of a leap year beyond the table; the search steps over.
	autumn := zdt(t, "2040-10-28T03:00:00+02:00[Europe/Vilnius]")
	if n, ok := autumn.NextTransition(); !ok || n.String() != "2041-03-31T04:00:00+03:00[Europe/Vilnius]" {
		t.Errorf("a transition after a leap year's autumn: %v %v", n, ok)
	}
	if p, ok := zdt(t, "2041-01-01T12:00:00+02:00[Europe/Vilnius]").PreviousTransition(); !ok || p.String() != "2040-10-28T03:00:00+02:00[Europe/Vilnius]" {
		t.Errorf("a transition before a leap year's end: %v %v", p, ok)
	}
}

func TestRoundingModes(t *testing.T) {
	type row struct {
		x    int64
		want [9]int64 // Ceil Floor Expand Trunc HalfCeil HalfFloor HalfExpand HalfTrunc HalfEven
	}
	rows := []row{
		{25, [9]int64{30, 20, 30, 20, 30, 20, 30, 20, 20}},
		{-25, [9]int64{-20, -30, -30, -20, -20, -30, -30, -20, -20}},
		{35, [9]int64{40, 30, 40, 30, 40, 30, 40, 30, 40}},
		{26, [9]int64{30, 20, 30, 20, 30, 30, 30, 30, 30}},
		{24, [9]int64{30, 20, 30, 20, 20, 20, 20, 20, 20}},
		{-24, [9]int64{-20, -30, -30, -20, -20, -20, -20, -20, -20}},
		{30, [9]int64{30, 30, 30, 30, 30, 30, 30, 30, 30}},
		{-30, [9]int64{-30, -30, -30, -30, -30, -30, -30, -30, -30}},
	}
	for _, r := range rows {
		for m := Ceil; m <= HalfEven; m++ {
			if got := roundInt(r.x, 10, m); got != r.want[m] {
				t.Errorf("round %d by mode %d: %d, want %d", r.x, m, got, r.want[m])
			}
		}
	}
}

func TestBubbleAndNudge(t *testing.T) {
	round := func(s string, smallest Unit, hasSmallest bool, largest Unit, hasLargest bool, mode RoundingMode, rel RelativeTo) (Duration, error) {
		return dur(t, s).Round(smallest, hasSmallest, largest, hasLargest, 1, mode, rel)
	}
	jan := date(t, "2021-01-01")
	r, err := round("P11M29D", Month, true, 0, false, HalfExpand, jan)
	check(t, "months stop at months", r, err, "P12M")
	r, err = round("P11M29D", Month, true, Year, true, HalfExpand, jan)
	check(t, "months bubble into a year", r, err, "P1Y")
	r, err = round("-P11M29D", Month, true, Year, true, HalfExpand, jan)
	check(t, "months bubble into a year, backward", r, err, "-P1Y")
	r, err = round("P1Y1M", Month, true, Year, true, HalfExpand, jan)
	check(t, "no bubbling", r, err, "P1Y1M")
	r, err = round("P1Y11M20D", Month, true, Year, true, HalfExpand, jan)
	check(t, "two years", r, err, "P2Y")
	r, err = round("P1W6DT23H", Day, true, Week, true, HalfExpand, jan)
	check(t, "days bubble into weeks", r, err, "P2W")
	r, err = round("P1W6DT1H", Day, true, Week, true, HalfExpand, jan)
	check(t, "days stay below a week", r, err, "P1W6D")
	r, err = round("P1Y200D", Year, true, 0, false, HalfExpand, jan)
	check(t, "to years", r, err, "P2Y")
	r, err = round("P1Y100D", Year, true, 0, false, HalfExpand, jan)
	check(t, "to years, down", r, err, "P1Y")
	r, err = round("P10D", Week, true, 0, false, HalfExpand, jan)
	check(t, "ten days in weeks", r, err, "P1W")
	r, err = round("P11D", Week, true, 0, false, HalfExpand, jan)
	check(t, "eleven days in weeks", r, err, "P2W")
	r, err = round("P1M10D", Week, true, 0, false, HalfExpand, jan)
	check(t, "weeks after a month", r, err, "P1M1W")
	ny := zdt(t, "2024-03-10T00:00:00-05:00[America/New_York]")
	r, err = round("PT36H", Day, true, 0, false, HalfExpand, ny)
	check(t, "zoned days", r, err, "P2D")
	r, err = round("PT36H", Day, true, 0, false, Trunc, ny)
	check(t, "zoned days truncated", r, err, "P1D")
	la := zdt(t, "2020-03-08T00:00:00-08:00[America/Los_Angeles]")
	r, err = round("P1DT23H59M", Hour, true, 0, false, HalfExpand, la)
	check(t, "rounded beyond the day", r, err, "P2D")
	r, err = round("-P1DT23H59M", Hour, true, 0, false, HalfExpand, la)
	check(t, "rounded beyond the day, backward", r, err, "-P2D")
	r, err = round("P1DT12H", Hour, true, 0, false, HalfExpand, la)
	check(t, "within the day", r, err, "P1DT12H")
	r, err = round("PT25H", Minute, true, Day, true, HalfExpand, la)
	check(t, "zoned minutes balanced to days", r, err, "P1DT2H")
	r, err = round("P1Y", Nanosecond, true, Month, true, HalfExpand, jan)
	check(t, "a year in months", r, err, "P12M")
	r, err = round("P2Y", Month, true, Year, true, HalfExpand, ny)
	check(t, "zoned years", r, err, "P2Y")
	r, err = round("P1M", Day, true, 0, false, HalfExpand, ny)
	check(t, "a zoned month stays a month", r, err, "P1M")
	r, err = round("P1M", Day, true, Day, true, HalfExpand, ny)
	check(t, "a zoned month in days", r, err, "P31D")

	f, err := dur(t, "P1Y").Total(Month, jan)
	if err != nil || f != 12 {
		t.Errorf("a year in months: %v %v", f, err)
	}
	f, err = dur(t, "PT36H").Total(Day, jan)
	if err != nil || f != 1.5 {
		t.Errorf("36 hours in days: %v %v", f, err)
	}
	f, err = dur(t, "P1DT12H").Total(Nanosecond, jan)
	if err != nil || f != 129_600_000_000_000 {
		t.Errorf("in nanoseconds: %v %v", f, err)
	}
	f, err = dur(t, "PT0S").Total(Hour, jan)
	if err != nil || f != 0 {
		t.Errorf("zero: %v %v", f, err)
	}
	f, err = dur(t, "P1Y6M").Total(Year, jan)
	if err != nil || f < 1.49 || f > 1.5 {
		t.Errorf("a year and a half: %v %v", f, err)
	}
	f, err = dur(t, "PT3H").Total(Minute, ny)
	if err != nil || f != 180 {
		t.Errorf("zoned minutes: %v %v", f, err)
	}
	f, err = dur(t, "P1M").Total(Week, ny)
	if err != nil || f < 4.4 || f > 4.5 {
		t.Errorf("a zoned month in weeks: %v %v", f, err)
	}
	c, err := CompareDurations(dur(t, "PT1H"), dur(t, "PT2H"), ny)
	if err != nil || c != -1 {
		t.Errorf("zoned time-only compare: %d %v", c, err)
	}
	c, err = CompareDurations(dur(t, "P1Y"), dur(t, "P12M"), jan)
	if err != nil || c != 0 {
		t.Errorf("a year and twelve months: %d %v", c, err)
	}
	c, err = CompareDurations(dur(t, "P2D"), dur(t, "P1D"), jan)
	if err != nil || c != 1 {
		t.Errorf("days with a date: %d %v", c, err)
	}
}

func TestArithmeticErrors(t *testing.T) {
	last, _ := ParseInstant("+275760-09-13T00:00:00Z")
	if _, err := last.Add(dur(t, "PT1H")); err == nil {
		t.Error("an instant past the range")
	}
	if _, err := zdt(t, "+275760-09-12T00:00:00+00:00[UTC]").Add(dur(t, "P2D")); err == nil {
		t.Error("a zoned date-time past the range")
	}
	if _, err := zdt(t, "+275760-09-12T00:00:00+00:00[UTC]").Add(dur(t, "PT48H")); err == nil {
		t.Error("a zoned date-time past the range by hours")
	}
	if _, err := date(t, "+275760-09-13").Add(dur(t, "P2D")); err == nil {
		t.Error("a date past the range")
	}
	if _, err := dt(t, "+275760-09-13T00:00").Add(dur(t, "P1Y")); err == nil {
		t.Error("a date-time past the range")
	}
	million := 1_000_000
	if _, err := date(t, "2024-01-01").With(Fields{Year: &million}); err == nil {
		t.Error("a year past the range")
	}
	if _, err := dt(t, "2024-01-01T00:00").With(Fields{Year: &million}); err == nil {
		t.Error("a date-time year past the range")
	}
	if _, err := zdt(t, "2024-01-01T00:00:00+00:00[UTC]").With(Fields{Year: &million}, Compatible); err == nil {
		t.Error("a zoned year past the range")
	}
	if _, err := dt(t, "+275760-09-14T23:00").Round(Hour, 2, HalfExpand); err == nil {
		t.Error("rounding past the range")
	}
	if _, err := dur(t, "PT9007199254740991S").Round(0, false, Nanosecond, true, 1, HalfExpand, nil); err == nil {
		t.Error("nanoseconds that do not fit an int")
	}
	if _, err := dur(t, "PT1H").Round(Hour, true, 0, false, 5, HalfExpand, nil); err == nil {
		t.Error("an increment of five hours")
	}
	if _, err := dur(t, "P1Y1M").Round(Month, true, 0, false, 2, HalfExpand, date(t, "2021-01-01")); err == nil {
		t.Error("an increment of two months with a larger unit")
	}
	if r, err := dur(t, "P3M").Round(Month, true, Month, true, 2, HalfExpand, date(t, "2021-01-01")); err != nil || r.String() != "P4M" {
		t.Errorf("an increment of two months: %v %v", r, err)
	}
	if _, err := dur(t, "PT1H").Round(Hour, true, 0, false, 0, HalfExpand, nil); err == nil {
		t.Error("a zero increment")
	}
	if _, err := (PlainTime{Hour: 1}).Until(PlainTime{}, Day); err == nil {
		t.Error("days between times")
	}
	first := zdt(t, "1800-01-01T00:00:00-04:56[America/New_York]")
	if _, ok := first.PreviousTransition(); ok {
		t.Error("a transition before the first")
	}
	if n, ok := first.NextTransition(); !ok || n.Date().Year != 1883 {
		t.Errorf("the first transition: %v %v", n, ok)
	}
	if d, err := NewDuration(0, 0, 0, 0, 0, 0, 0, 0, 0, -5); err != nil {
		t.Error("-5 nanoseconds", err)
	} else if abs, err := d.Abs(); err != nil || abs.Nanoseconds != 5 {
		t.Error("abs of a negative duration")
	}
	if (dateDur{}).sign() != 0 || (internalDur{time: zeroTime}).sign() != 0 || (internalDur{date: dateDur{days: -1}, time: zeroTime}).sign() != -1 {
		t.Error("signs of the internal forms")
	}
}
