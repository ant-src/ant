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

var unitNames = []string{"Year", "Month", "Week", "Day", "Hour", "Minute", "Second", "Millisecond", "Microsecond", "Nanosecond"}
var modeNames = []string{"Ceil", "Floor", "Expand", "Trunc", "HalfCeil", "HalfFloor", "HalfExpand", "HalfTrunc", "HalfEven"}

func enumTable(names []string) map[string]int {
	out := map[string]int{}
	for i, n := range names {
		out["Time."+n] = i
	}
	return out
}

var units, modes = enumTable(unitNames), enumTable(modeNames)

// unitArg reads a Time.Unit; fallback is the default, or -1 when the argument is required.
func unitArg(g *interp.Goroutine, who string, v value.Value, fallback int) (temporal.Unit, bool) {
	if v == nil || v == value.Null {
		if fallback < 0 {
			if v == nil {
				return 0, false
			}
			g.Panic("%s needs a Time.Unit, got null", who)
		}
		return temporal.Unit(fallback), true
	}
	return temporal.Unit(enumArg(g, who, "Time.Unit", v, units, 0)), true
}

func modeArg(g *interp.Goroutine, who string, v value.Value) temporal.RoundingMode {
	return temporal.RoundingMode(enumArg(g, who, "Time.RoundingMode", v, modes, int(temporal.HalfExpand)))
}

func durationArg(g *interp.Goroutine, who string, v value.Value) temporal.Duration {
	return timeArg[temporal.Duration](g, who, v)
}

// relativeArg reads relativeTo: a zoned date-time, a date, or a date-time taken as its date.
func relativeArg(g *interp.Goroutine, who string, v value.Value) temporal.RelativeTo {
	switch x := v.(type) {
	case nil:
		return nil
	case temporal.ZonedDateTime:
		return x
	case temporal.PlainDate:
		return x
	case temporal.PlainDateTime:
		return x.Date()
	}
	if v == value.Null {
		return nil
	}
	g.Panic("%s needs a Time.ZonedDateTime, Time.PlainDate or Time.PlainDateTime as relativeTo, got %s", who, value.Describe(v))
	return nil
}

func optIntPtr(g *interp.Goroutine, who string, v value.Value) *int {
	if v == nil || v == value.Null {
		return nil
	}
	n := int(intArg(g, who, v))
	return &n
}

func optInc(g *interp.Goroutine, who string, v value.Value) int64 {
	if v == nil {
		return 1
	}
	return intArg(g, who, v)
}

// defaultLargest is until's default largest unit per type.
func defaultLargest(v value.Value) int {
	switch v.(type) {
	case temporal.Instant:
		return int(temporal.Second)
	case temporal.ZonedDateTime, temporal.PlainTime:
		return int(temporal.Hour)
	}
	return int(temporal.Day)
}

func addPoint(g *interp.Goroutine, who string, t value.Value, d temporal.Duration) value.Value {
	var out value.Value
	var err error
	switch x := t.(type) {
	case temporal.Instant:
		out, err = x.Add(d)
	case temporal.ZonedDateTime:
		out, err = x.Add(d)
	case temporal.PlainDate:
		out, err = x.Add(d)
	case temporal.PlainTime:
		out, err = x.Add(d)
	case temporal.PlainDateTime:
		out, err = x.Add(d)
	default:
		g.Panic("%s needs a time value, got %s", who, value.Describe(t))
	}
	must(g, who, err)
	return out
}

func untilPoint(g *interp.Goroutine, who string, from, to value.Value, largest temporal.Unit) temporal.Duration {
	var out temporal.Duration
	var err error
	switch x := from.(type) {
	case temporal.Instant:
		out, err = x.Until(timeArg[temporal.Instant](g, who, to), largest)
	case temporal.ZonedDateTime:
		out, err = x.Until(timeArg[temporal.ZonedDateTime](g, who, to), largest)
	case temporal.PlainDate:
		out, err = x.Until(timeArg[temporal.PlainDate](g, who, to), largest)
	case temporal.PlainTime:
		out, err = x.Until(timeArg[temporal.PlainTime](g, who, to), largest)
	case temporal.PlainDateTime:
		out, err = x.Until(timeArg[temporal.PlainDateTime](g, who, to), largest)
	default:
		g.Panic("%s needs a time value, got %s", who, value.Describe(from))
	}
	must(g, who, err)
	return out
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

// must panics on an error, for a value that cannot be, such as a date that does not exist.
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
			constant("Duration", timeTag("Duration")),
			constant("Compatible", enumTag("Compatible")),
			constant("Earlier", enumTag("Earlier")),
			constant("Later", enumTag("Later")),
			constant("utc", temporal.UTC),

			fn("Time.zone", "id", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				z, err := temporal.LoadZone(stringArg(g, who, a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(z)
			}),
			fn("Time.localZone", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return temporal.LocalZone()
			}),
			fn("Time.zoneId", "z", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeArg[*temporal.Zone](g, who, a[0]).ID()
			}),

			fn("Time.now", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return temporal.Now()
			}),
			fn("Time.nowIn", "z", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return temporal.Now().InZone(timeArg[*temporal.Zone](g, who, a[0]))
			}),
			fn("Time.sleep", "d, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d := spanArg(g, who, a[0])
				ctx := waitContext(g, who, a[1])
				g.Instance().Flush()
				if ctx.Err() == nil && d > 0 {
					timer := time.NewTimer(d)
					defer timer.Stop()
					select {
					case <-timer.C:
					case <-ctx.Done():
					}
				}
				g.AfterWait()
				return value.Void
			}),
			fn("Time.ticker", "d, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d := spanArg(g, who, a[0])
				if d <= 0 {
					g.Panic("Time.ticker needs a period above zero, got %s", value.Render(a[0]))
				}
				ctx := waitContext(g, who, a[1])
				ticks := value.NewChan(1)
				// As Go's ticker: a beat counted from the call, one tick held, a tick that
				// finds it full dropped; the goroutine and its timer end with the context,
				// which is the run's when none was given.
				t := time.NewTicker(d)
				go func() {
					defer t.Stop()
					for {
						select {
						case <-t.C:
							select {
							case ticks.Raw() <- temporal.Now():
							default:
							}
						case <-ctx.Done():
							return
						}
					}
				}()
				return ticks
			}),
			fn("Time.monotonic", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return time.Since(origin).Milliseconds()
			}),
			fn("Time.fromEpochMillis", "ms", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				i, err := temporal.FromEpochMillis(intArg(g, who, a[0]))
				must(g, who, err)
				return i
			}),
			fn("Time.fromEpochNanos", "ns", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return temporal.FromEpochNanos(intArg(g, who, a[0]))
			}),
			fn("Time.epochMillis", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return instantArg(g, who, a[0]).EpochMillis()
			}),
			fn("Time.epochNanos", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				ns, err := instantArg(g, who, a[0]).EpochNanos()
				if err != nil {
					g.Panic("Time.epochNanos: the nanoseconds do not fit an int; the instant is outside 1677 to 2262")
				}
				return ns
			}),

			fn("Time.plainDate", "year, month, day", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d, err := temporal.NewPlainDate(int(intArg(g, who, a[0])), int(intArg(g, who, a[1])), int(intArg(g, who, a[2])))
				must(g, who, err)
				return d
			}),
			fn("Time.plainTime", "hour?, minute?, second?, millisecond?, microsecond?, nanosecond?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				t, err := temporal.NewPlainTime(optInt(g, who, a[0]), optInt(g, who, a[1]), optInt(g, who, a[2]),
					optInt(g, who, a[3]), optInt(g, who, a[4]), optInt(g, who, a[5]))
				must(g, who, err)
				return t
			}),
			fn("Time.plainDateTime", "date, time", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return temporal.NewPlainDateTime(timeArg[temporal.PlainDate](g, who, a[0]), timeArg[temporal.PlainTime](g, who, a[1]))
			}),

			fn("Time.duration", "years?, months?, weeks?, days?, hours?, minutes?, seconds?, milliseconds?, microseconds?, nanoseconds?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				var f [10]int64
				for i := range f {
					if a[i] != nil {
						f[i] = intArg(g, who, a[i])
					}
				}
				d, err := temporal.NewDuration(f[0], f[1], f[2], f[3], f[4], f[5], f[6], f[7], f[8], f[9])
				must(g, who, err)
				return d
			}),
			fn("Time.parseDuration", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeResult(temporal.ParseDuration(stringArg(g, who, a[0])))
			}),
			fn("Time.parseInstant", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeResult(temporal.ParseInstant(stringArg(g, who, a[0])))
			}),
			fn("Time.parseZoned", "s, disambiguation?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeResult(temporal.ParseZoned(stringArg(g, who, a[0]), disambiguationArg(g, who, a[1])))
			}),
			fn("Time.parsePlainDate", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeResult(temporal.ParsePlainDate(stringArg(g, who, a[0])))
			}),
			fn("Time.parsePlainTime", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeResult(temporal.ParsePlainTime(stringArg(g, who, a[0])))
			}),
			fn("Time.parsePlainDateTime", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeResult(temporal.ParsePlainDateTime(stringArg(g, who, a[0])))
			}),

			fn("Time.inZone", "t, z", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeArg[temporal.Instant](g, who, a[0]).InZone(timeArg[*temporal.Zone](g, who, a[1]))
			}),
			fn("Time.withZone", "t, z", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeArg[temporal.ZonedDateTime](g, who, a[0]).WithZone(timeArg[*temporal.Zone](g, who, a[1]))
			}),
			fn("Time.toZoned", "t, z, disambiguation?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				z, err := temporal.ToZoned(timeArg[temporal.PlainDateTime](g, who, a[0]), timeArg[*temporal.Zone](g, who, a[1]), disambiguationArg(g, who, a[2]))
				must(g, who, err)
				return z
			}),
			fn("Time.possibleInstants", "t, z", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				ps, err := temporal.PossibleInstants(timeArg[temporal.PlainDateTime](g, who, a[0]), timeArg[*temporal.Zone](g, who, a[1]))
				must(g, who, err)
				out := make([]value.Value, len(ps))
				for i, p := range ps {
					out[i] = p
				}
				return value.NewArray(out)
			}),
			fn("Time.toInstant", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeArg[temporal.ZonedDateTime](g, who, a[0]).Instant()
			}),
			fn("Time.toPlainDate", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return wallArg(g, who, a[0]).Date()
			}),
			fn("Time.toPlainTime", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return wallArg(g, who, a[0]).Clock()
			}),
			fn("Time.toPlainDateTime", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeArg[temporal.ZonedDateTime](g, who, a[0]).PlainDateTime()
			}),
			fn("Time.startOfDay", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				z, err := timeArg[temporal.ZonedDateTime](g, who, a[0]).StartOfDay()
				must(g, who, err)
				return z
			}),
			fn("Time.zoneOf", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeArg[temporal.ZonedDateTime](g, who, a[0]).Zone()
			}),
			fn("Time.offset", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return timeArg[temporal.ZonedDateTime](g, who, a[0]).Offset()
			}),
			fn("Time.offsetSeconds", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return int64(timeArg[temporal.ZonedDateTime](g, who, a[0]).OffsetSeconds())
			}),
			fn("Time.hoursInDay", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h, err := timeArg[temporal.ZonedDateTime](g, who, a[0]).HoursInDay()
				must(g, who, err)
				return h
			}),
			fn("Time.compare", "a, b", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return int64(compareTime(g, a[0], a[1]))
			}),
			fn("Time.nextTransition", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				if n, ok := timeArg[temporal.ZonedDateTime](g, who, a[0]).NextTransition(); ok {
					return n
				}
				return value.Null
			}),
			fn("Time.previousTransition", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				if p, ok := timeArg[temporal.ZonedDateTime](g, who, a[0]).PreviousTransition(); ok {
					return p
				}
				return value.Null
			}),
			fn("Time.with", "t, year?, month?, day?, hour?, minute?, second?, millisecond?, microsecond?, nanosecond?, disambiguation?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				f := temporal.Fields{
					Year: optIntPtr(g, who, a[1]), Month: optIntPtr(g, who, a[2]), Day: optIntPtr(g, who, a[3]),
					Hour: optIntPtr(g, who, a[4]), Minute: optIntPtr(g, who, a[5]), Second: optIntPtr(g, who, a[6]),
					Millisecond: optIntPtr(g, who, a[7]), Microsecond: optIntPtr(g, who, a[8]), Nanosecond: optIntPtr(g, who, a[9]),
				}
				var out value.Value
				var err error
				switch x := a[0].(type) {
				case temporal.ZonedDateTime:
					out, err = x.With(f, disambiguationArg(g, who, a[10]))
				case temporal.PlainDate:
					out, err = x.With(f)
				case temporal.PlainTime:
					out, err = x.With(f)
				case temporal.PlainDateTime:
					out, err = x.With(f)
				default:
					g.Panic("%s needs a Time.ZonedDateTime, Time.PlainDate, Time.PlainTime or Time.PlainDateTime, got %s", who, value.Describe(a[0]))
				}
				must(g, who, err)
				return out
			}),
			fn("Time.add", "t, d", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return addPoint(g, who, a[0], durationArg(g, who, a[1]))
			}),
			fn("Time.subtract", "t, d", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d, err := durationArg(g, who, a[1]).Negate()
				must(g, who, err)
				return addPoint(g, who, a[0], d)
			}),
			fn("Time.until", "from, to, largestUnit?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				largest, _ := unitArg(g, who, a[2], defaultLargest(a[0]))
				return untilPoint(g, who, a[0], a[1], largest)
			}),
			fn("Time.since", "t, other, largestUnit?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				largest, _ := unitArg(g, who, a[2], defaultLargest(a[0]))
				d, err := untilPoint(g, who, a[0], a[1], largest).Negate()
				must(g, who, err)
				return d
			}),
			fn("Time.round", "t, smallestUnit, increment?, mode?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				unit, _ := unitArg(g, who, a[1], -1)
				inc := optInc(g, who, a[2])
				mode := modeArg(g, who, a[3])
				var out value.Value
				var err error
				switch x := a[0].(type) {
				case temporal.Instant:
					out, err = x.Round(unit, inc, mode)
				case temporal.ZonedDateTime:
					out, err = x.Round(unit, inc, mode)
				case temporal.PlainTime:
					out, err = x.Round(unit, inc, mode)
				case temporal.PlainDateTime:
					out, err = x.Round(unit, inc, mode)
				default:
					g.Panic("%s needs a Time.Instant, Time.ZonedDateTime, Time.PlainTime or Time.PlainDateTime, got %s", who, value.Describe(a[0]))
				}
				must(g, who, err)
				return out
			}),
			fn("Time.part", "d, unit", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				unit, _ := unitArg(g, who, a[1], -1)
				return durationArg(g, who, a[0]).Part(unit)
			}),
			fn("Time.sign", "d", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return int64(durationArg(g, who, a[0]).Sign())
			}),
			fn("Time.negate", "d", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d, err := durationArg(g, who, a[0]).Negate()
				must(g, who, err)
				return d
			}),
			fn("Time.abs", "d", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d, err := durationArg(g, who, a[0]).Abs()
				must(g, who, err)
				return d
			}),
			fn("Time.addDuration", "a, b", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d, err := temporal.AddDurations(durationArg(g, who, a[0]), durationArg(g, who, a[1]))
				must(g, who, err)
				return d
			}),
			fn("Time.compareDuration", "a, b, relativeTo?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				c, err := temporal.CompareDurations(durationArg(g, who, a[0]), durationArg(g, who, a[1]), relativeArg(g, who, a[2]))
				must(g, who, err)
				return int64(c)
			}),
			fn("Time.total", "d, unit, relativeTo?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				unit, _ := unitArg(g, who, a[1], -1)
				f, err := durationArg(g, who, a[0]).Total(unit, relativeArg(g, who, a[2]))
				must(g, who, err)
				return f
			}),
			fn("Time.roundDuration", "d, smallestUnit?, largestUnit?, increment?, mode?, relativeTo?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				smallest, hasSmallest := unitArg(g, who, a[1], -1)
				largest, hasLargest := unitArg(g, who, a[2], -1)
				d, err := durationArg(g, who, a[0]).Round(smallest, hasSmallest, largest, hasLargest, optInc(g, who, a[3]), modeArg(g, who, a[4]), relativeArg(g, who, a[5]))
				must(g, who, err)
				return d
			}),
		}
		for _, name := range unitNames {
			entries = append(entries, constant(name, enumTag(name)))
		}
		for _, name := range modeNames {
			entries = append(entries, constant(name, enumTag(name)))
		}
		dated := []struct {
			name  string
			field func(temporal.PlainDate) value.Value
		}{
			{"year", func(d temporal.PlainDate) value.Value { return int64(d.Year) }},
			{"month", func(d temporal.PlainDate) value.Value { return int64(d.Month) }},
			{"day", func(d temporal.PlainDate) value.Value { return int64(d.Day) }},
			{"dayOfWeek", func(d temporal.PlainDate) value.Value { return int64(d.DayOfWeek()) }},
			{"dayOfYear", func(d temporal.PlainDate) value.Value { return int64(d.DayOfYear()) }},
			{"weekOfYear", func(d temporal.PlainDate) value.Value { return int64(d.WeekOfYear()) }},
			{"yearOfWeek", func(d temporal.PlainDate) value.Value { return int64(d.YearOfWeek()) }},
			{"daysInMonth", func(d temporal.PlainDate) value.Value { return int64(d.DaysInMonth()) }},
			{"daysInYear", func(d temporal.PlainDate) value.Value { return int64(d.DaysInYear()) }},
			{"inLeapYear", func(d temporal.PlainDate) value.Value { return d.InLeapYear() }},
		}
		for _, f := range dated {
			field, who := f.field, "Time."+f.name
			entries = append(entries, fn(who, "t", func(g *interp.Goroutine, _ string, a []value.Value) value.Value {
				return field(datedArg(g, who, a[0]).Date())
			}))
		}
		timed := []struct {
			name  string
			field func(temporal.PlainTime) value.Value
		}{
			{"hour", func(t temporal.PlainTime) value.Value { return int64(t.Hour) }},
			{"minute", func(t temporal.PlainTime) value.Value { return int64(t.Minute) }},
			{"second", func(t temporal.PlainTime) value.Value { return int64(t.Second) }},
			{"millisecond", func(t temporal.PlainTime) value.Value { return int64(t.Millisecond()) }},
			{"microsecond", func(t temporal.PlainTime) value.Value { return int64(t.Microsecond()) }},
			{"nanosecond", func(t temporal.PlainTime) value.Value { return int64(t.Nanosecond()) }},
		}
		for _, f := range timed {
			field, who := f.field, "Time."+f.name
			entries = append(entries, fn(who, "t", func(g *interp.Goroutine, _ string, a []value.Value) value.Value {
				return field(timedArg(g, who, a[0]).Clock())
			}))
		}
		return namespace("Time", entries...)
	})
}
