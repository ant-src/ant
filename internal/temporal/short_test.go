package temporal

import (
	"strings"
	"testing"
	"time"
)

func TestDurationLiteral(t *testing.T) {
	good := map[string]Duration{
		"500ms":      {Milliseconds: 500},
		"2h30m":      {Hours: 2, Minutes: 30},
		"90m":        {Minutes: 90},
		"1.5h":       {Hours: 1, Minutes: 30},
		"1h30.5m":    {Hours: 1, Minutes: 30, Seconds: 30},
		"1.5s":       {Seconds: 1, Milliseconds: 500},
		"0.5ms":      {Microseconds: 500},
		"0s":         {},
		"1_000ms":    {Milliseconds: 1000},
		"2h30m15s":   {Hours: 2, Minutes: 30, Seconds: 15},
		"1h1ns":      {Hours: 1, Nanoseconds: 1},
		"1.000_5s":   {Seconds: 1, Microseconds: 500},
		"0.25us":     {Nanoseconds: 250},
		"1.0000001h": {Hours: 1, Microseconds: 360},
	}
	for s, want := range good {
		got, err := DurationLiteral(s)
		if err != nil || got != want {
			t.Errorf("%s: %v (%v), want %v", s, got, err, want)
		}
	}
	bad := map[string]string{
		"1m2h":                 "descending",
		"1h1h":                 "descending",
		"1.5h30m":              "last component",
		"1d":                   "Time.duration(days: 1)",
		"1mo":                  "Time.duration(days: 1)",
		"1M":                   "M is not a duration unit",
		"1min":                 "min is not a duration unit",
		"0.1ns":                "whole number of nanoseconds",
		"1h30":                 "needs a unit",
		"01s":                  "malformed number",
		"1__0s":                "malformed number",
		"1_s":                  "malformed number",
		"1h_30m":               "malformed number",
		"9999999999999999999h": "out of range",
		"3000000000000h":       "out of range",
	}
	for s, want := range bad {
		if _, err := DurationLiteral(s); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: %v, want an error with %q", s, err, want)
		}
	}
}

func TestParseGoDuration(t *testing.T) {
	good := map[string]Duration{
		"2h30m":          {Hours: 2, Minutes: 30},
		"1m2h":           {Hours: 2, Minutes: 1},
		"1h1h":           {Hours: 2},
		"1.5h30m":        {Hours: 1, Minutes: 60},
		"-500ms":         {Milliseconds: -500},
		"+5s":            {Seconds: 5},
		"1.5µs":          {Microseconds: 1, Nanoseconds: 500},
		"1.5μs":          {Microseconds: 1, Nanoseconds: 500},
		"0":              {},
		"-0":             {},
		"0s":             {},
		"0.1ns":          {},
		".5s":            {Milliseconds: 500},
		"5.s":            {Seconds: 5},
		"1h0m0s":         {Hours: 1},
		"PT2H30M":        {Hours: 2, Minutes: 30},
		"-PT1H":          {Hours: -1},
		"1.9999999999ns": {Nanoseconds: 1},
	}
	for s, want := range good {
		got, err := ParseDuration(s)
		if err != nil || got != want {
			t.Errorf("%q: %v (%v), want %v", s, got, err, want)
		}
	}
	for _, s := range []string{"", "+", "-", "bad", "5", "1d", "1h-30m", ".s", "1h 30m", "1_000ms", "--5s", "+-5s", "1e3s", "1H", "--PT1H", "9999999999999999999h", "3000000000000h"} {
		if d, err := ParseDuration(s); err == nil {
			t.Errorf("%q accepted as %v", s, d)
		}
	}
	// Every string Go's Duration.String prints parses to the same length.
	for _, gd := range []time.Duration{0, 1, 1500, 1500 * time.Microsecond, 90 * time.Minute, -2*time.Hour - 3*time.Nanosecond, 1<<63 - 1, -1 << 63} {
		d, err := ParseDuration(gd.String())
		if err != nil {
			t.Errorf("%s: %v", gd, err)
			continue
		}
		want, _ := NewDuration(0, 0, 0, 0, 0, 0, 0, 0, 0, int64(gd))
		if c, err := CompareDurations(d, want, nil); err != nil || c != 0 {
			t.Errorf("%s parses to %v", gd, d)
		}
	}
}
