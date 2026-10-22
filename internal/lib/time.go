package lib

import (
	"time"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/temporal"
	"github.com/ant-src/ant/value"
)

// The Time namespace (stdlib.md), over internal/temporal. Time is computation, so a field that
// cannot be panics as the core library would, while parsing returns a Result. The seven types
// are tags of kind TimeTag, which is tests by name; the option types are enumerations, empty
// record types the builtins read by their tag.

var origin = time.Now()

func timeTag(name string) *value.Tag { return value.NewTag("Time."+name, value.TimeTag, nil) }
func enumTag(name string) *value.Tag { return value.NewTag("Time."+name, value.RecordTag, nil) }

// timeArg takes an argument of one time type.
func timeArg[T value.Time](g *interp.Goroutine, who string, v value.Value) T {
	t, ok := v.(T)
	if !ok {
		var zero T
		g.Panic("%s needs a %s, got %s", who, zero.TimeType(), value.Describe(v))
	}
	return t
}

func datedArg(g *interp.Goroutine, who string, v value.Value) temporal.Dated {
	if d, ok := v.(temporal.Dated); ok {
		return d
	}
	g.Panic("%s needs a Time.ZonedDateTime, Time.PlainDate or Time.PlainDateTime, got %s", who, value.Describe(v))
	return nil
}

func timedArg(g *interp.Goroutine, who string, v value.Value) temporal.Timed {
	if t, ok := v.(temporal.Timed); ok {
		return t
	}
	g.Panic("%s needs a Time.ZonedDateTime, Time.PlainTime or Time.PlainDateTime, got %s", who, value.Describe(v))
	return nil
}

// instantArg takes an Instant or the instant of a ZonedDateTime.
func instantArg(g *interp.Goroutine, who string, v value.Value) temporal.Instant {
	switch x := v.(type) {
	case temporal.Instant:
		return x
	case temporal.ZonedDateTime:
		return x.Instant()
	}
	g.Panic("%s needs a Time.Instant or Time.ZonedDateTime, got %s", who, value.Describe(v))
	return temporal.Instant{}
}

// wallArg takes a ZonedDateTime or a PlainDateTime as a wall-clock date and time.
func wallArg(g *interp.Goroutine, who string, v value.Value) temporal.PlainDateTime {
	switch x := v.(type) {
	case temporal.ZonedDateTime:
		return x.PlainDateTime()
	case temporal.PlainDateTime:
		return x
	}
	g.Panic("%s needs a Time.ZonedDateTime or Time.PlainDateTime, got %s", who, value.Describe(v))
	return temporal.PlainDateTime{}
}

// enumArg reads an enumeration value by its tag's name; nil, an omitted argument, is the default.
func enumArg(g *interp.Goroutine, who, what string, v value.Value, names map[string]int, fallback int) int {
	if v == nil {
		return fallback
	}
	if r, ok := v.(*value.Record); ok {
		if n, ok := names[r.Tag.Name]; ok {
			return n
		}
	}
	g.Panic("%s needs a %s, got %s", who, what, value.Describe(v))
	return 0
}

var disambiguations = map[string]int{
	"Time.Compatible": int(temporal.Compatible), "Time.Earlier": int(temporal.Earlier), "Time.Later": int(temporal.Later),
}

func disambiguationArg(g *interp.Goroutine, who string, v value.Value) temporal.Disambiguation {
	return temporal.Disambiguation(enumArg(g, who, "Time.Disambiguation", v, disambiguations, int(temporal.Compatible)))
}

// timeResult wraps a parse: Ok with the value, or Err with the message.
func timeResult(v value.Time, err error) value.Value {
	if err != nil {
		return errResult(err)
	}
	return value.Ok(v)
}

// must panics on an error, for the ◆ of a value that cannot be.
func must(g *interp.Goroutine, who string, err error) {
	if err != nil {
		g.Panic("%s: %v", who, err)
	}
}

func optInt(g *interp.Goroutine, who string, v value.Value) int {
	if v == nil {
		return 0
	}
	return int(intArg(g, who, v))
}

func compareTime(g *interp.Goroutine, a, b value.Value) int {
	const who = "Time.compare"
	switch x := a.(type) {
	case temporal.Instant:
		return x.Compare(timeArg[temporal.Instant](g, who, b))
	case temporal.ZonedDateTime:
		return x.Compare(timeArg[temporal.ZonedDateTime](g, who, b))
	case temporal.PlainDate:
		return x.Compare(timeArg[temporal.PlainDate](g, who, b))
	case temporal.PlainTime:
		return x.Compare(timeArg[temporal.PlainTime](g, who, b))
	case temporal.PlainDateTime:
		return x.Compare(timeArg[temporal.PlainDateTime](g, who, b))
	}
	g.Panic("%s needs two time values of one type, got %s", who, value.Describe(a))
	return 0
}

func init() {
	register("Time", func() value.Value {
		entries := []any{
			constant("Zone", timeTag("Zone")),
			constant("Instant", timeTag("Instant")),
			constant("ZonedDateTime", timeTag("ZonedDateTime")),
			constant("PlainDate", timeTag("PlainDate")),
			constant("PlainTime", timeTag("PlainTime")),
			constant("PlainDateTime", timeTag("PlainDateTime")),
			constant("Compatible", enumTag("Compatible")),
			constant("Earlier", enumTag("Earlier")),
			constant("Later", enumTag("Later")),
			constant("utc", temporal.UTC),

			fn("Time.zone", "id", func(g *interp.Goroutine, a []value.Value) value.Value {
				z, err := temporal.LoadZone(stringArg(g, "Time.zone", a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(z)
			}),
			fn("Time.localZone", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				return temporal.LocalZone()
			}),
			fn("Time.zoneId", "z", func(g *interp.Goroutine, a []value.Value) value.Value {
				return timeArg[*temporal.Zone](g, "Time.zoneId", a[0]).ID()
			}),

			fn("Time.now", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				return temporal.Now()
			}),
			fn("Time.nowIn", "z", func(g *interp.Goroutine, a []value.Value) value.Value {
				return temporal.Now().InZone(timeArg[*temporal.Zone](g, "Time.nowIn", a[0]))
			}),
			fn("Time.sleep", "ms", func(g *interp.Goroutine, a []value.Value) value.Value {
				ms := intArg(g, "Time.sleep", a[0])
				in := g.Instance()
				in.Flush()
				select {
				case <-time.After(time.Duration(ms) * time.Millisecond):
				case <-in.Context().Done():
				}
				return value.Void
			}),
			fn("Time.monotonic", "", func(g *interp.Goroutine, a []value.Value) value.Value {
				return time.Since(origin).Milliseconds()
			}),
			fn("Time.fromEpochMillis", "ms", func(g *interp.Goroutine, a []value.Value) value.Value {
				i, err := temporal.FromEpochMillis(intArg(g, "Time.fromEpochMillis", a[0]))
				must(g, "Time.fromEpochMillis", err)
				return i
			}),
			fn("Time.fromEpochNanos", "ns", func(g *interp.Goroutine, a []value.Value) value.Value {
				return temporal.FromEpochNanos(intArg(g, "Time.fromEpochNanos", a[0]))
			}),
			fn("Time.epochMillis", "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return instantArg(g, "Time.epochMillis", a[0]).EpochMillis()
			}),
			fn("Time.epochNanos", "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				ns, err := instantArg(g, "Time.epochNanos", a[0]).EpochNanos()
				if err != nil {
					g.Panic("Time.epochNanos: the nanoseconds do not fit an int; the instant is outside 1677 to 2262")
				}
				return ns
			}),

			fn("Time.plainDate", "year, month, day", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Time.plainDate"
				d, err := temporal.NewPlainDate(int(intArg(g, who, a[0])), int(intArg(g, who, a[1])), int(intArg(g, who, a[2])))
				must(g, who, err)
				return d
			}),
			fn("Time.plainTime", "hour?, minute?, second?, millisecond?, microsecond?, nanosecond?", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Time.plainTime"
				t, err := temporal.NewPlainTime(optInt(g, who, a[0]), optInt(g, who, a[1]), optInt(g, who, a[2]),
					optInt(g, who, a[3]), optInt(g, who, a[4]), optInt(g, who, a[5]))
				must(g, who, err)
				return t
			}),
			fn("Time.plainDateTime", "date, time", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Time.plainDateTime"
				return temporal.NewPlainDateTime(timeArg[temporal.PlainDate](g, who, a[0]), timeArg[temporal.PlainTime](g, who, a[1]))
			}),

			fn("Time.parseInstant", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return timeResult(temporal.ParseInstant(stringArg(g, "Time.parseInstant", a[0])))
			}),
			fn("Time.parseZoned", "s, disambiguation?", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Time.parseZoned"
				return timeResult(temporal.ParseZoned(stringArg(g, who, a[0]), disambiguationArg(g, who, a[1])))
			}),
			fn("Time.parsePlainDate", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return timeResult(temporal.ParsePlainDate(stringArg(g, "Time.parsePlainDate", a[0])))
			}),
			fn("Time.parsePlainTime", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return timeResult(temporal.ParsePlainTime(stringArg(g, "Time.parsePlainTime", a[0])))
			}),
			fn("Time.parsePlainDateTime", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return timeResult(temporal.ParsePlainDateTime(stringArg(g, "Time.parsePlainDateTime", a[0])))
			}),

			fn("Time.inZone", "t, z", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Time.inZone"
				return timeArg[temporal.Instant](g, who, a[0]).InZone(timeArg[*temporal.Zone](g, who, a[1]))
			}),
			fn("Time.withZone", "t, z", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Time.withZone"
				return timeArg[temporal.ZonedDateTime](g, who, a[0]).WithZone(timeArg[*temporal.Zone](g, who, a[1]))
			}),
			fn("Time.toZoned", "t, z, disambiguation?", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Time.toZoned"
				z, err := temporal.ToZoned(timeArg[temporal.PlainDateTime](g, who, a[0]), timeArg[*temporal.Zone](g, who, a[1]), disambiguationArg(g, who, a[2]))
				must(g, who, err)
				return z
			}),
			fn("Time.possibleInstants", "t, z", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Time.possibleInstants"
				ps, err := temporal.PossibleInstants(timeArg[temporal.PlainDateTime](g, who, a[0]), timeArg[*temporal.Zone](g, who, a[1]))
				must(g, who, err)
				out := make([]value.Value, len(ps))
				for i, p := range ps {
					out[i] = p
				}
				return value.NewArray(out)
			}),
			fn("Time.toInstant", "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return timeArg[temporal.ZonedDateTime](g, "Time.toInstant", a[0]).Instant()
			}),
			fn("Time.toPlainDate", "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return wallArg(g, "Time.toPlainDate", a[0]).Date()
			}),
			fn("Time.toPlainTime", "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return wallArg(g, "Time.toPlainTime", a[0]).Clock()
			}),
			fn("Time.toPlainDateTime", "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return timeArg[temporal.ZonedDateTime](g, "Time.toPlainDateTime", a[0]).PlainDateTime()
			}),
			fn("Time.startOfDay", "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Time.startOfDay"
				z, err := timeArg[temporal.ZonedDateTime](g, who, a[0]).StartOfDay()
				must(g, who, err)
				return z
			}),
			fn("Time.zoneOf", "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return timeArg[temporal.ZonedDateTime](g, "Time.zoneOf", a[0]).Zone()
			}),
			fn("Time.offset", "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return timeArg[temporal.ZonedDateTime](g, "Time.offset", a[0]).Offset()
			}),
			fn("Time.offsetSeconds", "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return int64(timeArg[temporal.ZonedDateTime](g, "Time.offsetSeconds", a[0]).OffsetSeconds())
			}),
			fn("Time.hoursInDay", "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Time.hoursInDay"
				h, err := timeArg[temporal.ZonedDateTime](g, who, a[0]).HoursInDay()
				must(g, who, err)
				return h
			}),
			fn("Time.compare", "a, b", func(g *interp.Goroutine, a []value.Value) value.Value {
				return int64(compareTime(g, a[0], a[1]))
			}),
		}
		dated := map[string]func(temporal.PlainDate) value.Value{
			"year":        func(d temporal.PlainDate) value.Value { return int64(d.Year) },
			"month":       func(d temporal.PlainDate) value.Value { return int64(d.Month) },
			"day":         func(d temporal.PlainDate) value.Value { return int64(d.Day) },
			"dayOfWeek":   func(d temporal.PlainDate) value.Value { return int64(d.DayOfWeek()) },
			"dayOfYear":   func(d temporal.PlainDate) value.Value { return int64(d.DayOfYear()) },
			"weekOfYear":  func(d temporal.PlainDate) value.Value { return int64(d.WeekOfYear()) },
			"yearOfWeek":  func(d temporal.PlainDate) value.Value { return int64(d.YearOfWeek()) },
			"daysInMonth": func(d temporal.PlainDate) value.Value { return int64(d.DaysInMonth()) },
			"daysInYear":  func(d temporal.PlainDate) value.Value { return int64(d.DaysInYear()) },
			"inLeapYear":  func(d temporal.PlainDate) value.Value { return d.InLeapYear() },
		}
		for _, name := range []string{"year", "month", "day", "dayOfWeek", "dayOfYear", "weekOfYear", "yearOfWeek", "daysInMonth", "daysInYear", "inLeapYear"} {
			field := dated[name]
			who := "Time." + name
			entries = append(entries, fn(who, "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return field(datedArg(g, who, a[0]).Date())
			}))
		}
		timed := map[string]func(temporal.PlainTime) value.Value{
			"hour":        func(t temporal.PlainTime) value.Value { return int64(t.Hour) },
			"minute":      func(t temporal.PlainTime) value.Value { return int64(t.Minute) },
			"second":      func(t temporal.PlainTime) value.Value { return int64(t.Second) },
			"millisecond": func(t temporal.PlainTime) value.Value { return int64(t.Millisecond()) },
			"microsecond": func(t temporal.PlainTime) value.Value { return int64(t.Microsecond()) },
			"nanosecond":  func(t temporal.PlainTime) value.Value { return int64(t.Nanosecond()) },
		}
		for _, name := range []string{"hour", "minute", "second", "millisecond", "microsecond", "nanosecond"} {
			field := timed[name]
			who := "Time." + name
			entries = append(entries, fn(who, "t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return field(timedArg(g, who, a[0]).Clock())
			}))
		}
		return namespace("Time", entries...)
	})
}
