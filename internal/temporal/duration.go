package temporal

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Unit is a unit of time, largest first.
type Unit int

const (
	Year Unit = iota
	Month
	Week
	Day
	Hour
	Minute
	Second
	Millisecond
	Microsecond
	Nanosecond
)

var unitNames = [...]string{"year", "month", "week", "day", "hour", "minute", "second", "millisecond", "microsecond", "nanosecond"}

func (u Unit) String() string { return unitNames[u] }

func largerUnit(a, b Unit) Unit {
	if a < b {
		return a
	}
	return b
}

func isCalendarUnit(u Unit) bool { return u <= Week }
func isDateUnit(u Unit) bool     { return u <= Day }

var nsPerUnit = map[Unit]int64{
	Day: 86_400_000_000_000, Hour: 3_600_000_000_000, Minute: 60_000_000_000, Second: 1_000_000_000,
	Millisecond: 1_000_000, Microsecond: 1000, Nanosecond: 1,
}

// RoundingMode is one of Temporal's nine.
type RoundingMode int

const (
	Ceil RoundingMode = iota
	Floor
	Expand
	Trunc
	HalfCeil
	HalfFloor
	HalfExpand
	HalfTrunc
	HalfEven
)

// Duration is years to nanoseconds, all of one sign, kept as given.
type Duration struct {
	Years, Months, Weeks, Days              int64
	Hours, Minutes, Seconds                 int64
	Milliseconds, Microseconds, Nanoseconds int64
}

// NewDuration makes a duration, rejecting mixed signs and the fields Temporal's range
// excludes: years, months, weeks and days at or beyond 2^32, and a time part, days included
// at 24 hours, at or beyond 2^53 seconds.
func NewDuration(years, months, weeks, days, hours, minutes, seconds, milliseconds, microseconds, nanoseconds int64) (Duration, error) {
	d := Duration{years, months, weeks, days, hours, minutes, seconds, milliseconds, microseconds, nanoseconds}
	if err := d.validate(); err != nil {
		return Duration{}, err
	}
	return d, nil
}

func (d Duration) fields() [10]int64 {
	return [10]int64{d.Years, d.Months, d.Weeks, d.Days, d.Hours, d.Minutes, d.Seconds, d.Milliseconds, d.Microseconds, d.Nanoseconds}
}

func (d Duration) validate() error {
	sign := 0
	for _, f := range d.fields() {
		s := sgn64(f)
		if s != 0 {
			if sign != 0 && s != sign {
				return errors.New("a duration's fields must all have one sign")
			}
			sign = s
		}
	}
	for _, f := range []int64{d.Years, d.Months, d.Weeks, d.Days} {
		if f >= 1<<32 || f <= -(1<<32) {
			return errors.New("years, months, weeks and days must be below 2^32")
		}
	}
	if _, err := d.internal24(); err != nil {
		return err
	}
	return nil
}

// Sign is -1, 0 or 1.
func (d Duration) Sign() int {
	for _, f := range d.fields() {
		if s := sgn64(f); s != 0 {
			return s
		}
	}
	return 0
}

// Negate flips every field; a field of -2^63, which a subsecond field may hold, has no
// positive counterpart that fits an int, and is an error.
func (d Duration) Negate() (Duration, error) {
	for _, f := range d.fields() {
		if f == math.MinInt64 {
			return Duration{}, errors.New("the negated duration has a field that does not fit an int")
		}
	}
	return Duration{-d.Years, -d.Months, -d.Weeks, -d.Days, -d.Hours, -d.Minutes, -d.Seconds, -d.Milliseconds, -d.Microseconds, -d.Nanoseconds}, nil
}

// Abs makes every field non-negative, an error where Negate is.
func (d Duration) Abs() (Duration, error) {
	if d.Sign() < 0 {
		return d.Negate()
	}
	return d, nil
}

// Part is the field of a unit, as given.
func (d Duration) Part(u Unit) int64 { return d.fields()[u] }

// defaultLargestUnit is the largest nonzero unit, Nanosecond for a zero duration.
func (d Duration) defaultLargestUnit() Unit {
	for i, f := range d.fields() {
		if f != 0 {
			return Unit(i)
		}
	}
	return Nanosecond
}

// TimeType is the name of the type's tag.
func (Duration) TimeType() string { return "Time.Duration" }

// Key is the canonical form == compares: the ten fields as given.
func (d Duration) Key() string {
	return fmt.Sprint(d.fields())
}

// String is the ISO 8601 form, the fields as given and the subsecond fields folded into the
// seconds: P1Y2M3DT4H5M6.5S, -PT1H, PT0S.
func (d Duration) String() string {
	var sb strings.Builder
	if d.Sign() < 0 {
		sb.WriteByte('-')
	}
	sb.WriteByte('P')
	for _, p := range []struct {
		v int64
		c byte
	}{{d.Years, 'Y'}, {d.Months, 'M'}, {d.Weeks, 'W'}, {d.Days, 'D'}} {
		if p.v != 0 {
			sb.WriteString(strconv.FormatInt(abs64(p.v), 10))
			sb.WriteByte(p.c)
		}
	}
	var tp strings.Builder
	if d.Hours != 0 {
		tp.WriteString(strconv.FormatInt(abs64(d.Hours), 10) + "H")
	}
	if d.Minutes != 0 {
		tp.WriteString(strconv.FormatInt(abs64(d.Minutes), 10) + "M")
	}
	secs := new(big.Int).Mul(big.NewInt(d.Seconds), big.NewInt(nanosPerSecond))
	secs.Add(secs, new(big.Int).Mul(big.NewInt(d.Milliseconds), big.NewInt(1_000_000)))
	secs.Add(secs, new(big.Int).Mul(big.NewInt(d.Microseconds), big.NewInt(1000)))
	secs.Add(secs, big.NewInt(d.Nanoseconds))
	secs.Abs(secs)
	if secs.Sign() != 0 || d.defaultLargestUnit() >= Second {
		whole, frac := new(big.Int).QuoRem(secs, big.NewInt(nanosPerSecond), new(big.Int))
		tp.WriteString(whole.String())
		if frac.Sign() != 0 {
			tp.WriteString("." + strings.TrimRight(fmt.Sprintf("%09d", frac.Int64()), "0"))
		}
		tp.WriteByte('S')
	}
	if tp.Len() > 0 {
		sb.WriteByte('T')
		sb.WriteString(tp.String())
	}
	return sb.String()
}

// ParseDuration reads ISO 8601's PnYnMnWnDTnHnMnS with an optional sign, where only the last
// time component may carry a fraction, which is spread over the smaller units.
func ParseDuration(s string) (Duration, error) {
	bad := func() (Duration, error) { return Duration{}, fmt.Errorf("not a duration: %q", s) }
	i := 0
	sign := int64(1)
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		if s[i] == '-' {
			sign = -1
		}
		i++
	}
	if i >= len(s) || (s[i] != 'P' && s[i] != 'p') {
		return bad()
	}
	i++
	var f [10]int64
	var fracNs int64 // the fraction of the last time component, in nanoseconds of that unit
	fracUnit := -1
	seen := 0
	inTime := false
	any := false
	order := map[byte]int{'Y': 0, 'M': 1, 'W': 2, 'D': 3, 'H': 4, 'S': 6}
	for i < len(s) {
		c := s[i]
		if c == 'T' || c == 't' {
			if inTime {
				return bad()
			}
			inTime = true
			i++
			if i >= len(s) {
				return bad() // T with nothing after it
			}
			continue
		}
		start := i
		for i < len(s) && isDigit(s[i]) {
			i++
		}
		if start == i {
			return bad()
		}
		digits := s[start:i]
		var fraction string
		if i < len(s) && (s[i] == '.' || s[i] == ',') {
			i++
			fs := i
			for i < len(s) && isDigit(s[i]) {
				i++
			}
			fraction = s[fs:i]
			if fraction == "" || len(fraction) > 9 {
				return bad()
			}
		}
		if i >= len(s) {
			return bad()
		}
		designator := s[i]
		i++
		if designator >= 'a' && designator <= 'z' {
			designator -= 'a' - 'A'
		}
		var idx int
		switch {
		case designator == 'M' && inTime:
			idx = 5
		case designator == 'M':
			idx = 1
		default:
			var ok bool
			idx, ok = order[designator]
			if !ok || (idx >= 4) != inTime {
				return bad()
			}
		}
		if idx < seen || (fracUnit >= 0) {
			return bad() // out of order, or a component after a fraction
		}
		seen = idx + 1
		v, err := strconv.ParseInt(digits, 10, 64)
		if err != nil {
			return bad()
		}
		f[idx] = v
		if fraction != "" {
			if !inTime {
				return bad()
			}
			n, _ := strconv.ParseInt(fraction+strings.Repeat("0", 9-len(fraction)), 10, 64)
			fracNs = n
			fracUnit = idx
		}
		any = true
	}
	if !any {
		return bad()
	}
	// Spread a fraction of hours or minutes or seconds over the smaller units, as Temporal's
	// parser does: the fraction is of the unit's length in nanoseconds.
	if fracUnit >= 0 {
		var unitNs int64
		switch fracUnit {
		case 4:
			unitNs = 3_600_000_000_000
		case 5:
			unitNs = 60_000_000_000
		default:
			unitNs = nanosPerSecond
		}
		rem := new(big.Int).Mul(big.NewInt(fracNs), big.NewInt(unitNs))
		rem.Quo(rem, big.NewInt(nanosPerSecond)) // nanoseconds of the fraction
		total := rem.Int64()
		if fracUnit == 4 {
			f[5] = total / 60_000_000_000
			total %= 60_000_000_000
		}
		if fracUnit <= 5 {
			f[6] = total / nanosPerSecond
			total %= nanosPerSecond
		}
		f[7] = total / 1_000_000
		f[8] = total / 1000 % 1000
		f[9] = total % 1000
	}
	for i := range f {
		f[i] *= sign
	}
	return NewDuration(f[0], f[1], f[2], f[3], f[4], f[5], f[6], f[7], f[8], f[9])
}

func sgn64(v int64) int {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	}
	return 0
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// ---- the internal form: a date part and an exact time part in big nanoseconds

// timeDur is an exact span of time in nanoseconds, bounded as Temporal's time durations are.
type timeDur struct{ ns *big.Int }

var (
	bigNsPerSecond = big.NewInt(nanosPerSecond)
	bigNsPerDay    = big.NewInt(86_400 * nanosPerSecond)
	maxTimeNs      = func() *big.Int { v, _ := new(big.Int).SetString("9007199254740991999999999", 10); return v }()
	zeroTime       = timeDur{new(big.Int)}
)

func tdOf(ns *big.Int) (timeDur, error) {
	if new(big.Int).Abs(ns).Cmp(maxTimeNs) > 0 {
		return timeDur{}, errors.New("the time part of a duration must be below 2^53 seconds")
	}
	return timeDur{ns}, nil
}

func tdComponents(h, m, s, ms, us, ns int64) (timeDur, error) {
	t := big.NewInt(ns)
	t.Add(t, new(big.Int).Mul(big.NewInt(us), big.NewInt(1000)))
	t.Add(t, new(big.Int).Mul(big.NewInt(ms), big.NewInt(1_000_000)))
	t.Add(t, new(big.Int).Mul(big.NewInt(s), bigNsPerSecond))
	t.Add(t, new(big.Int).Mul(big.NewInt(m), big.NewInt(60*nanosPerSecond)))
	t.Add(t, new(big.Int).Mul(big.NewInt(h), big.NewInt(3600*nanosPerSecond)))
	return tdOf(t)
}

func (t timeDur) add24HourDays(days int64) (timeDur, error) {
	return tdOf(new(big.Int).Add(t.ns, new(big.Int).Mul(big.NewInt(days), bigNsPerDay)))
}

func (t timeDur) add(o timeDur) (timeDur, error) { return tdOf(new(big.Int).Add(t.ns, o.ns)) }
func (t timeDur) sub(o timeDur) (timeDur, error) { return tdOf(new(big.Int).Sub(t.ns, o.ns)) }
func (t timeDur) sign() int                      { return t.ns.Sign() }
func (t timeDur) cmp(o timeDur) int              { return t.ns.Cmp(o.ns) }

// round rounds to a multiple of inc nanoseconds by mode, as TimeDuration.round.
func (t timeDur) round(inc *big.Int, mode RoundingMode) (timeDur, error) {
	return tdOf(roundBig(t.ns, inc, mode))
}

// divmod divides by n nanoseconds: the whole quotient and the remainder, both toward zero.
func (t timeDur) divmod(n *big.Int) (*big.Int, timeDur) {
	q, r := new(big.Int).QuoRem(t.ns, n, new(big.Int))
	return q, timeDur{r}
}

// fdiv is the quotient by n nanoseconds as a float.
func (t timeDur) fdiv(n *big.Int) float64 {
	f, _ := new(big.Rat).SetFrac(t.ns, n).Float64()
	return f
}

// secAndSubsec splits into whole seconds and the rest, both toward zero.
func (t timeDur) secAndSubsec() (sec, subsec int64) {
	q, r := new(big.Int).QuoRem(t.ns, bigNsPerSecond, new(big.Int))
	return q.Int64(), r.Int64()
}

// ns of an instant and back.
func (i Instant) ns() *big.Int {
	return new(big.Int).Add(new(big.Int).Mul(big.NewInt(i.sec), bigNsPerSecond), big.NewInt(int64(i.nsec)))
}

func instantFromNs(ns *big.Int) (Instant, error) {
	q, r := new(big.Int).DivMod(ns, bigNsPerSecond, new(big.Int)) // Euclidean: r in [0, 1e9)
	if !q.IsInt64() {
		return Instant{}, ErrRange
	}
	return NewInstant(q.Int64(), r.Int64())
}

func (t timeDur) addToInstant(i Instant) (Instant, error) {
	return instantFromNs(new(big.Int).Add(i.ns(), t.ns))
}

// utcNs is a wall-clock date-time's epoch nanoseconds were it in UTC.
func utcNs(dt PlainDateTime) *big.Int {
	sec, nsec := dt.asUTC()
	return new(big.Int).Add(new(big.Int).Mul(big.NewInt(sec), bigNsPerSecond), big.NewInt(nsec))
}

type dateDur struct{ years, months, weeks, days int64 }

func (d dateDur) sign() int {
	for _, f := range []int64{d.years, d.months, d.weeks, d.days} {
		if s := sgn64(f); s != 0 {
			return s
		}
	}
	return 0
}

type internalDur struct {
	date dateDur
	time timeDur
}

func (id internalDur) sign() int {
	if s := id.date.sign(); s != 0 {
		return s
	}
	return id.time.sign()
}

// internal keeps the days in the date part.
func (d Duration) internal() (internalDur, error) {
	t, err := tdComponents(d.Hours, d.Minutes, d.Seconds, d.Milliseconds, d.Microseconds, d.Nanoseconds)
	if err != nil {
		return internalDur{}, err
	}
	return internalDur{dateDur{d.Years, d.Months, d.Weeks, d.Days}, t}, nil
}

// internal24 folds the days into the time part at 24 hours each.
func (d Duration) internal24() (internalDur, error) {
	id, err := d.internal()
	if err != nil {
		return internalDur{}, err
	}
	if id.time, err = id.time.add24HourDays(d.Days); err != nil {
		return internalDur{}, err
	}
	id.date.days = 0
	return id, nil
}

// fromInternal balances the time part into fields up to largest, as
// TemporalDurationFromInternal; a field that does not fit an int is an error.
func fromInternal(id internalDur, largest Unit) (Duration, error) {
	sign := int64(id.time.sign())
	absNs := new(big.Int).Abs(id.time.ns)
	var days, hours, minutes, seconds, millis, micros, nanos int64
	split := func(unit Unit) (*big.Int, error) {
		q, r := new(big.Int).QuoRem(absNs, big.NewInt(nsPerUnit[unit]), new(big.Int))
		absNs = r
		if !q.IsInt64() {
			return nil, errors.New("a balanced field does not fit an int")
		}
		return q, nil
	}
	take := func(unit Unit, into *int64) error {
		q, err := split(unit)
		if err != nil {
			return err
		}
		*into = q.Int64()
		return nil
	}
	order := []struct {
		unit Unit
		into *int64
	}{{Day, &days}, {Hour, &hours}, {Minute, &minutes}, {Second, &seconds}, {Millisecond, &millis}, {Microsecond, &micros}, {Nanosecond, &nanos}}
	start := largest // extract from the larger of largest and a day downward
	if start < Day {
		start = Day
	}
	for _, o := range order {
		if o.unit < start {
			continue
		}
		if err := take(o.unit, o.into); err != nil {
			return Duration{}, err
		}
	}
	return NewDuration(id.date.years, id.date.months, id.date.weeks, id.date.days+sign*days,
		sign*hours, sign*minutes, sign*seconds, sign*millis, sign*micros, sign*nanos)
}
