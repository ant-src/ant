package temporal

import (
	"math"
	"strings"
	"testing"
)

func zone(t *testing.T, id string) *Zone {
	t.Helper()
	z, err := LoadZone(id)
	if err != nil {
		t.Fatal(err)
	}
	return z
}

func dt(t *testing.T, s string) PlainDateTime {
	t.Helper()
	v, err := ParsePlainDateTime(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// The gap and the overlap of New York in 2024, as Temporal's documentation has them.
func TestDisambiguation(t *testing.T) {
	ny := zone(t, "America/New_York")
	cases := []struct {
		wall string
		d    Disambiguation
		want string
	}{
		{"2024-03-10T02:30", Compatible, "2024-03-10T03:30:00-04:00[America/New_York]"},
		{"2024-03-10T02:30", Earlier, "2024-03-10T01:30:00-05:00[America/New_York]"},
		{"2024-03-10T02:30", Later, "2024-03-10T03:30:00-04:00[America/New_York]"},
		{"2024-11-03T01:30", Compatible, "2024-11-03T01:30:00-04:00[America/New_York]"},
		{"2024-11-03T01:30", Earlier, "2024-11-03T01:30:00-04:00[America/New_York]"},
		{"2024-11-03T01:30", Later, "2024-11-03T01:30:00-05:00[America/New_York]"},
		{"2024-06-01T12:00", Compatible, "2024-06-01T12:00:00-04:00[America/New_York]"},
	}
	for _, c := range cases {
		got, err := ToZoned(dt(t, c.wall), ny, c.d)
		if err != nil {
			t.Fatal(err)
		}
		if got.String() != c.want {
			t.Errorf("%s with %d: %s, want %s", c.wall, c.d, got, c.want)
		}
	}
	for wall, n := range map[string]int{"2024-03-10T02:30": 0, "2024-11-03T01:30": 2, "2024-06-01T12:00": 1} {
		ps, err := PossibleInstants(dt(t, wall), ny)
		if err != nil || len(ps) != n {
			t.Errorf("%s: %d possible instants, want %d (%v)", wall, len(ps), n, err)
		}
	}
}

func TestHoursInDay(t *testing.T) {
	cases := []struct {
		s    string
		want float64
	}{
		{"2024-03-10T01:58:00-05:00[America/New_York]", 23},
		{"2024-11-03T01:58:00-04:00[America/New_York]", 25},
		{"2024-11-04T01:58:00-05:00[America/New_York]", 24},
		{"2024-10-06T12:00:00+11:00[Australia/Lord_Howe]", 23.5},
		{"2025-04-06T12:00:00+10:30[Australia/Lord_Howe]", 24.5},
	}
	for _, c := range cases {
		z, err := ParseZoned(c.s, Compatible)
		if err != nil {
			t.Fatal(err)
		}
		if h, err := z.HoursInDay(); err != nil || h != c.want {
			t.Errorf("%s: %v hours (%v), want %v", c.s, h, err, c.want)
		}
	}
	z, _ := ParseZoned("2024-03-31T12:00[Europe/Berlin]", Compatible)
	if s, _ := z.StartOfDay(); s.String() != "2024-03-31T00:00:00+01:00[Europe/Berlin]" {
		t.Errorf("start of day: %s", s)
	}
	// Midnight in a gap: Santiago's spring forward happens at 24:00, so the day starts at 01:00.
	z, _ = ParseZoned("2024-09-08T12:00[America/Santiago]", Compatible)
	if s, _ := z.StartOfDay(); s.String() != "2024-09-08T01:00:00-03:00[America/Santiago]" {
		t.Errorf("start of a day beginning in a gap: %s", s)
	}
}

func TestParseZoned(t *testing.T) {
	cases := []struct {
		s    string
		want string // "" for an error
	}{
		{"2024-11-03T01:05:00[America/New_York]", "2024-11-03T01:05:00-04:00[America/New_York]"},
		{"2024-03-10T02:05:00[America/New_York]", "2024-03-10T03:05:00-04:00[America/New_York]"},
		{"2024-11-03T01:05:00-05:00[America/New_York]", "2024-11-03T01:05:00-05:00[America/New_York]"},
		{"2024-11-03T06:05:00Z[America/New_York]", "2024-11-03T01:05:00-05:00[America/New_York]"},
		{"2019-12-23T12:00:00-02:00[America/Sao_Paulo]", ""},
		{"2026-10-22T16:03:05.5+02:00[Europe/Stockholm]", "2026-10-22T16:03:05.5+02:00[Europe/Stockholm]"},
		{"2026-10-22T16:03:05+02:00[europe/stockholm][u-ca=iso8601]", "2026-10-22T16:03:05+02:00[Europe/Stockholm]"},
		{"2026-10-22T16:03:05+02:00[Europe/Stockholm][u-ca=chinese]", ""},
		{"2026-10-22T16:03:05+02:00", ""},
		{"2026-10-22[Asia/Calcutta]", "2026-10-22T00:00:00+05:30[Asia/Calcutta]"},
		{"2026-10-22T16:03:05+05:30[+05:30]", "2026-10-22T16:03:05+05:30[+05:30]"},
		{"-000100-01-01T00:00:00+00:00[UTC]", "-000100-01-01T00:00:00+00:00[UTC]"},
	}
	for _, c := range cases {
		got, err := ParseZoned(c.s, Compatible)
		switch {
		case c.want == "" && err == nil:
			t.Errorf("%s: parsed to %s, want an error", c.s, got)
		case c.want != "" && err != nil:
			t.Errorf("%s: %v", c.s, err)
		case c.want != "" && got.String() != c.want:
			t.Errorf("%s: %s, want %s", c.s, got, c.want)
		}
	}
}

func TestParsePlain(t *testing.T) {
	if d, err := ParsePlainDate("2024-11-03T01:00:00-04:00[America/New_York]"); err != nil || d.String() != "2024-11-03" {
		t.Errorf("date from a zoned string: %v %v", d, err)
	}
	if _, err := ParsePlainDate("2024-11-03T01:00:00Z"); err == nil {
		t.Error("Z accepted by a plain date")
	}
	if d, err := ParsePlainDate("2024-11-03T01:00:00Z[UTC]"); err != nil || d.String() != "2024-11-03" {
		t.Errorf("Z with a zone: %v %v", d, err)
	}
	for _, bad := range []string{"2026-10-22T25:00", "2026-10-22T14:60", "2026-10-22T14:03:60"} {
		if _, err := ParsePlainDate(bad); err == nil {
			t.Errorf("a date with the impossible time %q", bad)
		}
	}
	if _, err := ParsePlainDate("2024-02-30"); err == nil || !strings.Contains(err.Error(), "no such date") {
		t.Errorf("February 30: %v", err)
	}
	if tm, err := ParsePlainTime("14:03:05.123"); err != nil || tm.String() != "14:03:05.123" || tm.Millisecond() != 123 {
		t.Errorf("time: %v %v", tm, err)
	}
	if tm, err := ParsePlainTime("T14:03"); err != nil || tm.String() != "14:03:00" {
		t.Errorf("T time: %v %v", tm, err)
	}
	if tm, err := ParsePlainTime("2026-10-22 14:03:05.000000001+02:00"); err != nil || tm.Nanosecond() != 1 || tm.String() != "14:03:05.000000001" {
		t.Errorf("time with a date: %v %v", tm, err)
	}
	if _, err := ParsePlainTime("2026-10-22"); err == nil {
		t.Error("a bare date accepted as a time")
	}
	if d, err := ParsePlainDateTime("2026-10-22T14:03:05"); err != nil || d.String() != "2026-10-22T14:03:05" {
		t.Errorf("date-time: %v %v", d, err)
	}
	if _, err := ParsePlainDateTime("2026-10-22T24:00:00"); err == nil {
		t.Error("hour 24 accepted")
	}
	if d, err := ParsePlainDate("+100000-01-01"); err != nil || d.String() != "+100000-01-01" || d.Year != 100000 {
		t.Errorf("extended year: %v %v", d, err)
	}
	for _, bad := range []string{"", "2026", "2026-1-1", "2026-10-22T", "2026-10-22T14", "2026-10-22T14:03:05+2", "2026-10-22x", "2026-10-22[]", "2026-10-22[a][b]"} {
		if _, err := ParsePlainDate(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestInstant(t *testing.T) {
	i, err := ParseInstant("2026-10-22T14:03:05.5Z")
	if err != nil || i.String() != "2026-10-22T14:03:05.5Z" {
		t.Fatalf("instant: %v %v", i, err)
	}
	j, _ := ParseInstant("2026-10-22T16:03:05.5+02:00[Europe/Stockholm]")
	if i.Compare(j) != 0 || i.Key() != j.Key() {
		t.Errorf("the same instant at an offset: %s and %s", i, j)
	}
	if _, err := ParseInstant("2026-10-22T14:03:05"); err == nil {
		t.Error("an instant without an offset")
	}
	if i.EpochMillis() != 1792677785500 {
		t.Errorf("epoch millis %d", i.EpochMillis())
	}
	if ns, err := i.EpochNanos(); err != nil || ns != 1792677785500000000 {
		t.Errorf("epoch nanos %d %v", ns, err)
	}
	far, _ := ParseInstant("2300-01-01T00:00:00Z")
	if _, err := far.EpochNanos(); err == nil {
		t.Error("nanoseconds of 2300 fit an int")
	}
	if FromEpochNanos(-1).String() != "1969-12-31T23:59:59.999999999Z" {
		t.Errorf("negative nanos: %s", FromEpochNanos(-1))
	}
	if m, err := FromEpochMillis(-1); err != nil || m.EpochMillis() != -1 {
		t.Errorf("negative millis: %v %v", m, err)
	}
	if _, err := FromEpochMillis(maxDays*secondsPerDay*1000 + 1); err == nil {
		t.Error("beyond the range accepted")
	}
	if edge, err := FromEpochMillis(maxDays * secondsPerDay * 1000); err != nil || edge.String() != "+275760-09-13T00:00:00Z" {
		t.Errorf("the last instant: %v %v", edge, err)
	}
	if Now().InZone(UTC).Zone().ID() != "UTC" {
		t.Error("now in UTC")
	}
}

func TestZones(t *testing.T) {
	a, b := zone(t, "Asia/Calcutta"), zone(t, "asia/kolkata")
	if a.Key() != b.Key() || a.ID() != "Asia/Calcutta" || b.ID() != "Asia/Kolkata" {
		t.Errorf("aliases: %s %q, %s %q", a, a.Key(), b, b.Key())
	}
	if zone(t, "europe/STOCKHOLM").ID() != "Europe/Stockholm" {
		t.Error("case")
	}
	for _, bad := range []string{"", "Local", "Europe/Nowhere", "+24:00", "+5", "0530"} {
		if _, err := LoadZone(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	off := zone(t, "+0530")
	if off.ID() != "+05:30" || !off.Fixed() || off.Key() == zone(t, "Asia/Kolkata").Key() {
		t.Errorf("offset zone: %s %q", off, off.Key())
	}
	if zone(t, "-08").ID() != "-08:00" {
		t.Error("hours-only offset")
	}
	if UTC.ID() != "UTC" || UTC.Key() != zone(t, "Etc/UTC").Key() || UTC.Key() == zone(t, "+00:00").Key() {
		t.Errorf("UTC: %q", UTC.Key())
	}
	// Aliases follow CLDR: tzdata's post-1970 merges do not make two zones one.
	if zone(t, "Europe/Stockholm").Key() == zone(t, "Europe/Berlin").Key() {
		t.Error("Stockholm and Berlin are one zone")
	}
	if zone(t, "US/Eastern").Key() != zone(t, "America/New_York").Key() {
		t.Error("US/Eastern and New York are two zones")
	}
	if zone(t, "Etc/GMT").Key() == UTC.Key() {
		t.Error("GMT and UTC are one zone")
	}
	if zone(t, "Europe/Kiev").Key() != zone(t, "Europe/Kyiv").Key() {
		t.Error("Kiev and Kyiv are two zones")
	}
	t.Setenv("TZ", "Europe/Stockholm")
	if LocalZone().ID() != "Europe/Stockholm" {
		t.Errorf("local from TZ: %s", LocalZone())
	}
	t.Setenv("TZ", "Nowhere/Nothing")
	if LocalZone() == nil {
		t.Error("no local zone")
	}
}

func TestCivil(t *testing.T) {
	d, _ := NewPlainDate(2026, 10, 22)
	if d.DayOfWeek() != 4 || d.DayOfYear() != 295 || d.WeekOfYear() != 43 || d.DaysInMonth() != 31 || d.DaysInYear() != 365 || d.InLeapYear() {
		t.Errorf("2026-10-22: %d %d %d", d.DayOfWeek(), d.DayOfYear(), d.WeekOfYear())
	}
	for _, c := range []struct {
		s          string
		year, week int
	}{{"2021-01-01", 2020, 53}, {"2024-12-30", 2025, 1}, {"2024-12-29", 2024, 52}, {"2027-01-03", 2026, 53}} {
		d, _ := ParsePlainDate(c.s)
		if d.YearOfWeek() != c.year || d.WeekOfYear() != c.week {
			t.Errorf("%s: week %d of %d, want %d of %d", c.s, d.WeekOfYear(), d.YearOfWeek(), c.week, c.year)
		}
	}
	if _, err := NewPlainDate(2023, 2, 29); err == nil {
		t.Error("2023-02-29")
	}
	if d, err := NewPlainDate(2024, 2, 29); err != nil || !d.InLeapYear() {
		t.Error("2024-02-29")
	}
	if _, err := NewPlainDate(300000, 1, 1); err == nil {
		t.Error("year 300000 accepted")
	}
	// A year whose day count overflows int64 back into the range.
	if d, err := NewPlainDate(3989932118587117601, 3, 1); err == nil {
		t.Errorf("year 3989932118587117601 accepted: %v", d)
	}
	for _, days := range []int64{-maxDays, -719468, -1, 0, 1, 59, 60, 365, 11016, maxDays} {
		if got := dateFromDays(days).days(); got != days {
			t.Errorf("day %d round-trips to %d", days, got)
		}
	}
	if d := dateFromDays(-719528); d.String() != "0000-01-01" {
		t.Errorf("year zero: %s", d)
	}
	if _, err := NewPlainTime(23, 59, 59, 999, 999, 999); err != nil {
		t.Error(err)
	}
	if _, err := NewPlainTime(0, 60, 0, 0, 0, 0); err == nil {
		t.Error("minute 60")
	}
	a, _ := ParsePlainDateTime("2026-10-22T14:03:05")
	b, _ := ParsePlainDateTime("2026-10-22T14:03:06")
	if a.Compare(b) != -1 || b.Compare(a) != 1 || a.Compare(a) != 0 {
		t.Error("compare")
	}
}

func TestParseEdges(t *testing.T) {
	good := map[string]string{
		"2026-10-22T14:03:05.5+05:30":                       "2026-10-22T14:03:05.5",
		"2026-10-22 14:03:05+05:30":                         "2026-10-22T14:03:05",
		"2026-10-22T14:03:05+05:30:00":                      "2026-10-22T14:03:05",
		"2026-10-22T14:03:05+05:30:00.5":                    "2026-10-22T14:03:05",
		"2026-10-22T14:03:05+05:30[Europe/Stockholm]":       "2026-10-22T14:03:05",
		"2026-10-22T14:03:05Z[Europe/Stockholm]":            "2026-10-22T14:03:05",
		"2026-10-22T14:03:05.123456789-00:00[u-ca=ISO8601]": "2026-10-22T14:03:05.123456789",
		"+010000-01-01T00:00":                               "+010000-01-01T00:00:00",
	}
	for s, want := range good {
		if dt, err := ParsePlainDateTime(s); err != nil || dt.String() != want {
			t.Errorf("%s: %v %v, want %s", s, dt, err, want)
		}
	}
	bad := []string{
		"2026-10-22T14:03:05+5", "2026-10-22T14:03:05+05:3", "2026-10-22T14:03:05+05:30:6",
		"2026-10-22T14:03:05+05:30:60", "2026-10-22T14:03:05+24:00", "2026-10-22T14:03:05+05:",
		"2026-10-22T14:03:05.1234567890Z", "2026-10-22T14:03:05.Z", "2026-10-22T14:03:05+05:30:00.",
		"2026-10-22T14:03:05[u-ca=iso8601][u-ca=iso8601]", "2026-10-22T14:03:05[Europe/Stockholm][Europe/Oslo]",
		"-000000-01-01", "2026-10-22T14:03:05Z", "2026-10-22T14:03:05[Europe/Stockholm", "2026-10-22T14:03:05]",
		"2026-13-01T00:00", "+1000000-01-01", "2026-10-22T14:03:05Zx",
		// Only the grammar of stdlib.md: no lowercase t or z, no decimal comma, no ±HH or
		// ±HHMM offset, no critical flag, and no leap second.
		"2026-10-22t14:03:05", "2026-10-22T14:03:05z[Europe/Stockholm]", "2026-10-22T14:03:05,5",
		"2026-10-22T14:03:05+05:30:00,5", "2026-10-22T14:03:05+05", "2026-10-22T14:03:05+0530",
		"2026-10-22T14:03:05+05:30[!Europe/Stockholm]", "2026-10-22T14:03:05[!u-ca=iso8601]",
		"2026-10-22T23:59:60", "t14:03",
	}
	for _, s := range bad {
		if _, err := ParsePlainDateTime(s); err == nil {
			t.Errorf("%q accepted", s)
		}
	}
	if _, err := ParsePlainTime("14:03Z"); err == nil {
		t.Error("a time with Z and no zone")
	}
	if _, err := ParsePlainTime("2026-02-30T14:03"); err == nil {
		t.Error("a time after an impossible date")
	}
	if _, err := ParseInstant("2026-02-30T14:03Z"); err == nil {
		t.Error("an instant on an impossible date")
	}
	if _, err := ParseInstant("+275760-09-14T00:00:00Z"); err == nil {
		t.Error("an instant past the range")
	}
	if _, err := ParseZoned("2026-02-30T14:03[UTC]", Compatible); err == nil {
		t.Error("a zoned value on an impossible date")
	}
	if _, err := ParseZoned("2026-10-22T14:03[Mars/Olympus]", Compatible); err == nil {
		t.Error("an unknown zone")
	}
	// New York kept local mean time, -04:56:02, until 1883; an offset given to the minute
	// matches it rounded.
	if z, err := ParseZoned("1883-11-18T12:00:00-04:56[America/New_York]", Compatible); err != nil || z.Offset() != "-04:56:02" {
		t.Errorf("local mean time: %v %v", z, err)
	}
	if _, err := ParseZoned("1883-11-18T12:00:00-04:56:00[America/New_York]", Compatible); err == nil {
		t.Error("an offset with seconds that do not match")
	}
	// The string rounds a sub-minute offset to the minute, as RFC 9557 has no seconds; the
	// offset itself keeps them, and the string parses back to the same instant.
	lmt, err := ParseZoned("1874-12-07T19:33:28+00:53:28[Europe/Berlin]", Compatible)
	if err != nil || lmt.String() != "1874-12-07T19:33:28+00:53[Europe/Berlin]" || lmt.Offset() != "+00:53:28" {
		t.Errorf("local mean time in Berlin: %v %q %v", lmt, lmt.Offset(), err)
	}
	if back, err := ParseZoned(lmt.String(), Compatible); err != nil || back.Key() != lmt.Key() {
		t.Errorf("the rounded string does not parse back: %v %v", back, err)
	}
	for off, want := range map[int]int{-17762: -17760, 17762: 17760, 29: 0, 30: 60, -30: -60, 31: 60} {
		if got := roundToMinute(off); got != want {
			t.Errorf("roundToMinute(%d) = %d, want %d", off, got, want)
		}
	}
	if formatOffset(-17762) != "-04:56:02" || formatOffset(0) != "+00:00" {
		t.Error("offset format")
	}
}

func TestRangeEdges(t *testing.T) {
	last, _ := ParsePlainDateTime("+275760-09-13T00:00:00")
	if z, err := ToZoned(last, zone(t, "Asia/Tokyo"), Compatible); err != nil || z.String() != "+275760-09-13T00:00:00+09:00[Asia/Tokyo]" {
		t.Errorf("the last day in Tokyo: %v %v", z, err)
	}
	if _, err := ToZoned(last, zone(t, "Etc/GMT+12"), Compatible); err == nil {
		t.Error("the last day at -12:00 is past the range")
	}
	if _, err := PossibleInstants(last, zone(t, "-12:00")); err == nil {
		t.Error("possible instants past the range")
	}
	if ps, err := PossibleInstants(last, zone(t, "Pacific/Kiritimati")); err != nil || len(ps) != 1 {
		t.Errorf("the last day at +14:00: %v %v", ps, err)
	}
	first, _ := ParsePlainDate("-271821-04-20")
	if _, err := ToZoned(NewPlainDateTime(first, PlainTime{}), zone(t, "+14:00"), Compatible); err == nil {
		t.Error("the first day at +14:00 is before the range")
	}
	if _, err := NewInstant(maxEpochSeconds, 1); err == nil {
		t.Error("a nanosecond past the range")
	}
	if i, err := NewInstant(0, -1); err != nil || i.String() != "1969-12-31T23:59:59.999999999Z" {
		t.Errorf("negative nanoseconds: %v %v", i, err)
	}
	if TzdataVersion() == "" {
		t.Error("no tzdata version")
	}
	a, _ := ParsePlainDate("2026-10-22")
	b, _ := ParsePlainDate("2026-10-23")
	if a.Compare(b) != -1 || b.Compare(a) != 1 || a.Key() != "2026-10-22" || a.TimeType() != "Time.PlainDate" {
		t.Error("date compare and key")
	}
	x, _ := ParsePlainTime("10:00")
	y, _ := ParsePlainTime("09:00")
	if x.Compare(y) != 1 || y.Compare(x) != -1 || x.Key() != "10:00:00" || x.TimeType() != "Time.PlainTime" {
		t.Error("time compare and key")
	}
	dt, _ := ParsePlainDateTime("2026-10-22T10:00")
	if dt.Key() != "2026-10-22T10:00:00" || dt.TimeType() != "Time.PlainDateTime" || dt.Date() != a || dt.Clock() != x {
		t.Error("date-time key and parts")
	}
	i, _ := ParseInstant("2026-10-22T10:00Z")
	j, _ := ParseInstant("2026-10-22T10:00:00.000000001Z")
	if i.Compare(j) != -1 || j.Compare(i) != 1 {
		t.Error("instant compare by nanosecond")
	}
	if daysInYear(2024) != 366 || daysInYear(2100) != 365 {
		t.Error("days in year")
	}
	for _, ns := range []int64{math.MaxInt64, math.MinInt64, 0, -1, 1} {
		if got, err := FromEpochNanos(ns).EpochNanos(); err != nil || got != ns {
			t.Errorf("%d round-trips to %d (%v)", ns, got, err)
		}
	}
	high, _ := NewInstant(math.MaxInt64/nanosPerSecond, math.MaxInt64%nanosPerSecond+1)
	low, _ := NewInstant(math.MinInt64/nanosPerSecond-1, 0)
	for _, i := range []Instant{high, low} {
		if _, err := i.EpochNanos(); err == nil {
			t.Errorf("%s fits an int64", i)
		}
	}
}
