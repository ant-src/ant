package temporal

import (
	"errors"
	"fmt"
	"math/big"
	"time"
)

// Calendar and exact arithmetic, after Temporal: the date part of a duration moves the wall
// clock with out-of-range fields constrained, the time part moves the instant.

func balanceYearMonth(y, m int64) (int64, int) {
	m--
	y += floorDiv(m, 12)
	m = floorMod(m, 12)
	return y, int(m + 1)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// constrainDate clamps the month and the day into range.
func constrainDate(y int64, m, d int) (PlainDate, error) {
	m = clampInt(m, 1, 12)
	d = clampInt(d, 1, daysInMonth(y, m))
	if y < -999_999 || y > 999_999 {
		return PlainDate{}, ErrRange
	}
	return NewPlainDate(int(y), m, d)
}

func addDays(d PlainDate, n int64) (PlainDate, error) {
	days := d.days() + n
	if days < -maxDays-1 || days > maxDays+1 {
		return PlainDate{}, ErrRange
	}
	return dateFromDays(days), nil
}

// dateAdd adds a date duration: years, then months, with the day constrained, then weeks
// and days.
func dateAdd(d PlainDate, dd dateDur) (PlainDate, error) {
	y, m := balanceYearMonth(int64(d.Year)+dd.years, int64(d.Month)+dd.months)
	intermediate, err := constrainDate(y, m, d.Day)
	if err != nil {
		return PlainDate{}, err
	}
	return addDays(intermediate, dd.days+7*dd.weeks)
}

func compareSurpasses(sign int, y int64, m, day int, d2 PlainDate) bool {
	c, err := constrainDate(y, m, day)
	if err != nil {
		return true
	}
	return sign*c.Compare(d2) > 0
}

func isoDateSurpasses(sign int, base, d2 PlainDate, years, months int64) bool {
	y0 := int64(base.Year) + years
	if compareSurpasses(sign, y0, base.Month, base.Day, d2) {
		return true
	}
	if months == 0 {
		return false
	}
	y, m := balanceYearMonth(y0, int64(base.Month)+months)
	return compareSurpasses(sign, y, m, base.Day, d2)
}

// dateUntil is the ISO calendar's dateUntil: the difference from one to two with years and
// months counted from one, the day constrained.
func dateUntil(one, two PlainDate, largest Unit) dateDur {
	sign := -one.Compare(two)
	if sign == 0 {
		return dateDur{}
	}
	var years, months int64
	if largest == Year || largest == Month {
		candidateYears := int64(two.Year - one.Year)
		if candidateYears != 0 {
			candidateYears -= int64(sign)
		}
		for !isoDateSurpasses(sign, one, two, candidateYears, 0) {
			years = candidateYears
			candidateYears += int64(sign)
		}
		candidateMonths := int64(sign)
		for !isoDateSurpasses(sign, one, two, years, candidateMonths) {
			months = candidateMonths
			candidateMonths += int64(sign)
		}
		if largest == Month {
			months += years * 12
			years = 0
		}
	}
	y, m := balanceYearMonth(int64(one.Year)+years, int64(one.Month)+months)
	constrained, _ := constrainDate(y, m, one.Day)
	var weeks int64
	days := two.days() - constrained.days()
	if largest == Week {
		weeks = days / 7
		days %= 7
	}
	return dateDur{years, months, weeks, days}
}

func differenceTime(t1, t2 PlainTime) timeDur {
	return timeDur{big.NewInt(t2.nanosOfDay() - t1.nanosOfDay())}
}

// addTime adds an exact span to a wall-clock time: the time of day and the days carried.
func addTime(t PlainTime, td timeDur) (deltaDays int64, out PlainTime) {
	total := new(big.Int).Add(big.NewInt(t.nanosOfDay()), td.ns)
	q, r := new(big.Int).DivMod(total, bigNsPerDay, new(big.Int))
	return q.Int64(), timeFromNanosOfDay(r.Int64())
}

func differenceISODateTime(dt1, dt2 PlainDateTime, largest Unit) (internalDur, error) {
	td := differenceTime(dt1.time, dt2.time)
	timeSign := td.sign()
	dateSign := dt1.date.Compare(dt2.date)
	adjusted := dt2.date
	var err error
	if dateSign == timeSign && timeSign != 0 {
		if adjusted, err = addDays(adjusted, int64(timeSign)); err != nil {
			return internalDur{}, err
		}
		if td, err = td.add24HourDays(int64(-timeSign)); err != nil {
			return internalDur{}, err
		}
	}
	dateLargest := largerUnit(Day, largest)
	dd := dateUntil(dt1.date, adjusted, dateLargest)
	if largest != dateLargest {
		if td, err = td.add24HourDays(dd.days); err != nil {
			return internalDur{}, err
		}
		dd.days = 0
	}
	return internalDur{dd, td}, nil
}

func differenceZonedDateTime(i1, i2 Instant, z *Zone, largest Unit) (internalDur, error) {
	nsDiff := new(big.Int).Sub(i2.ns(), i1.ns())
	if nsDiff.Sign() == 0 {
		return internalDur{time: zeroTime}, nil
	}
	sign := 1
	if nsDiff.Sign() < 0 {
		sign = -1
	}
	start, end := i1.InZone(z).PlainDateTime(), i2.InZone(z).PlainDateTime()
	if start.date == end.date {
		return internalDur{time: timeDur{nsDiff}}, nil
	}
	// The forward direction allows a second day of correction for a push-forward compatible
	// disambiguation.
	maxDayCorrection := int64(1)
	if sign == 1 {
		maxDayCorrection = 2
	}
	dayCorrection := int64(0)
	td := differenceTime(start.time, end.time)
	if td.sign() == -sign {
		dayCorrection++
	}
	var intermediate PlainDateTime
	for ; dayCorrection <= maxDayCorrection; dayCorrection++ {
		date, err := addDays(end.date, -dayCorrection*int64(sign))
		if err != nil {
			return internalDur{}, err
		}
		intermediate = PlainDateTime{date, start.time}
		zi, err := ToZoned(intermediate, z, Compatible)
		if err != nil {
			return internalDur{}, err
		}
		td = timeDur{new(big.Int).Sub(i2.ns(), zi.inst.ns())}
		if td.sign() != -sign {
			break
		}
	}
	if dayCorrection > maxDayCorrection {
		return internalDur{}, errors.New("the difference needs more than two days of correction")
	}
	dd := dateUntil(start.date, intermediate.date, largerUnit(Day, largest))
	return internalDur{dd, td}, nil
}

func differenceInstant(i1, i2 Instant, inc int64, smallest Unit, mode RoundingMode) (internalDur, error) {
	td := timeDur{new(big.Int).Sub(i2.ns(), i1.ns())}
	td, err := td.round(new(big.Int).Mul(big.NewInt(nsPerUnit[smallest]), big.NewInt(inc)), mode)
	if err != nil {
		return internalDur{}, err
	}
	return internalDur{time: td}, nil
}

// ---- add and subtract

// Add moves the date by the duration's date part, the time part counted as whole days of 24
// hours with the remainder dropped.
func (d PlainDate) Add(dur Duration) (PlainDate, error) {
	id, err := dur.internal24()
	if err != nil {
		return PlainDate{}, err
	}
	sec, _ := id.time.secAndSubsec()
	id.date.days = sec / 86400
	return dateAdd(d, id.date)
}

// Add moves the time around the clock; the date part is ignored.
func (t PlainTime) Add(dur Duration) (PlainTime, error) {
	id, err := dur.internal24()
	if err != nil {
		return PlainTime{}, err
	}
	_, out := addTime(t, id.time)
	return out, nil
}

// Add adds the time part to the clock, carrying days, then the date part to the date.
func (dt PlainDateTime) Add(dur Duration) (PlainDateTime, error) {
	id, err := dur.internal24()
	if err != nil {
		return PlainDateTime{}, err
	}
	delta, t := addTime(dt.time, id.time)
	id.date.days += delta
	d, err := dateAdd(dt.date, id.date)
	if err != nil {
		return PlainDateTime{}, err
	}
	return PlainDateTime{d, t}, nil
}

// Add moves the instant by the time part; a date unit is an error, since an instant has no
// calendar.
func (i Instant) Add(dur Duration) (Instant, error) {
	if isDateUnit(dur.defaultLargestUnit()) && dur.Sign() != 0 {
		return Instant{}, fmt.Errorf("an instant has no calendar; %ss cannot be added to it", dur.defaultLargestUnit())
	}
	id, err := dur.internal24()
	if err != nil {
		return Instant{}, err
	}
	return id.time.addToInstant(i)
}

// Add adds the date part to the wall clock, places the result in the zone as Compatible, and
// adds the time part to that instant; a duration with no date part moves the instant alone.
func (z ZonedDateTime) Add(dur Duration) (ZonedDateTime, error) {
	id, err := dur.internal()
	if err != nil {
		return ZonedDateTime{}, err
	}
	if id.date.sign() == 0 {
		i, err := id.time.addToInstant(z.inst)
		if err != nil {
			return ZonedDateTime{}, err
		}
		return ZonedDateTime{i, z.zone}, nil
	}
	dt := z.PlainDateTime()
	added, err := dateAdd(dt.date, id.date)
	if err != nil {
		return ZonedDateTime{}, err
	}
	intermediate, err := ToZoned(PlainDateTime{added, dt.time}, z.zone, Compatible)
	if err != nil {
		return ZonedDateTime{}, err
	}
	i, err := id.time.addToInstant(intermediate.inst)
	if err != nil {
		return ZonedDateTime{}, err
	}
	return ZonedDateTime{i, z.zone}, nil
}

// ---- until and since

// Until is the duration from i to o, balanced up to largest, which is Hour or smaller.
func (i Instant) Until(o Instant, largest Unit) (Duration, error) {
	if largest < Hour {
		return Duration{}, fmt.Errorf("an instant has no %ss", largest)
	}
	id, err := differenceInstant(i, o, 1, Nanosecond, Trunc)
	if err != nil {
		return Duration{}, err
	}
	return fromInternal(id, largest)
}

// Until is the duration from t to o around the clock, balanced up to largest, Hour or smaller.
func (t PlainTime) Until(o PlainTime, largest Unit) (Duration, error) {
	if largest < Hour {
		return Duration{}, fmt.Errorf("a time has no %ss", largest)
	}
	return fromInternal(internalDur{time: differenceTime(t, o)}, largest)
}

// Until is the duration from d to o, in units of Day or larger, counted from d.
func (d PlainDate) Until(o PlainDate, largest Unit) (Duration, error) {
	if largest > Day {
		return Duration{}, fmt.Errorf("a date has no %ss", largest)
	}
	dd := dateUntil(d, o, largest)
	return NewDuration(dd.years, dd.months, dd.weeks, dd.days, 0, 0, 0, 0, 0, 0)
}

// Until is the duration from dt to o, with days of 24 hours.
func (dt PlainDateTime) Until(o PlainDateTime, largest Unit) (Duration, error) {
	id, err := differenceISODateTime(dt, o, largest)
	if err != nil {
		return Duration{}, err
	}
	return fromInternal(id, largest)
}

// Until is the duration from z to o: an exact difference for Hour and smaller, and for Day
// and larger the hybrid difference of Temporal, whose days are the zone's real days, which
// needs both in one zone.
func (z ZonedDateTime) Until(o ZonedDateTime, largest Unit) (Duration, error) {
	if !isDateUnit(largest) {
		id, err := differenceInstant(z.inst, o.inst, 1, Nanosecond, Trunc)
		if err != nil {
			return Duration{}, err
		}
		return fromInternal(id, largest)
	}
	if z.zone.Key() != o.zone.Key() {
		return Duration{}, fmt.Errorf("%s and %s are different zones; a difference in days or larger needs one zone", z.zone.id, o.zone.id)
	}
	id, err := differenceZonedDateTime(z.inst, o.inst, z.zone, largest)
	if err != nil {
		return Duration{}, err
	}
	return fromInternal(id, largest)
}

// ---- with

// Fields are the fields with may replace; a nil pointer keeps the value's own.
type Fields struct {
	Year, Month, Day                     *int
	Hour, Minute, Second                 *int
	Millisecond, Microsecond, Nanosecond *int
}

func (f Fields) hasDate() bool { return f.Year != nil || f.Month != nil || f.Day != nil }
func (f Fields) hasTime() bool {
	return f.Hour != nil || f.Minute != nil || f.Second != nil || f.Millisecond != nil || f.Microsecond != nil || f.Nanosecond != nil
}

func pick(p *int, v int) int {
	if p != nil {
		return *p
	}
	return v
}

// With replaces the given date fields, constrained; a time field is an error.
func (d PlainDate) With(f Fields) (PlainDate, error) {
	if f.hasTime() {
		return PlainDate{}, errors.New("a date has no time fields")
	}
	return constrainDate(int64(pick(f.Year, d.Year)), pick(f.Month, d.Month), pick(f.Day, d.Day))
}

// With replaces the given time fields, each clamped to its range; a date field is an error.
func (t PlainTime) With(f Fields) (PlainTime, error) {
	if f.hasDate() {
		return PlainTime{}, errors.New("a time has no date fields")
	}
	return PlainTime{
		clampInt(pick(f.Hour, t.Hour), 0, 23), clampInt(pick(f.Minute, t.Minute), 0, 59), clampInt(pick(f.Second, t.Second), 0, 59),
		clampInt(pick(f.Millisecond, t.Millisecond()), 0, 999)*1_000_000 + clampInt(pick(f.Microsecond, t.Microsecond()), 0, 999)*1000 +
			clampInt(pick(f.Nanosecond, t.Nanosecond()), 0, 999),
	}, nil
}

// With replaces the given fields, constrained.
func (dt PlainDateTime) With(f Fields) (PlainDateTime, error) {
	d, err := dt.date.With(Fields{Year: f.Year, Month: f.Month, Day: f.Day})
	if err != nil {
		return PlainDateTime{}, err
	}
	t, _ := dt.time.With(Fields{Hour: f.Hour, Minute: f.Minute, Second: f.Second, Millisecond: f.Millisecond, Microsecond: f.Microsecond, Nanosecond: f.Nanosecond})
	return PlainDateTime{d, t}, nil
}

// With replaces the given wall-clock fields, constrained, and places the result in the zone
// with the value's own offset when the zone has it at the new time, and by d otherwise.
func (z ZonedDateTime) With(f Fields, d Disambiguation) (ZonedDateTime, error) {
	dt, err := z.PlainDateTime().With(f)
	if err != nil {
		return ZonedDateTime{}, err
	}
	return interpretOffset(dt, z.zone, z.OffsetSeconds(), d)
}

// interpretOffset is Temporal's InterpretISODateTimeOffset with offset "prefer": the instant
// at which the zone has the given offset for the wall-clock time, else the disambiguation.
func interpretOffset(dt PlainDateTime, z *Zone, offset int, d Disambiguation) (ZonedDateTime, error) {
	ps, err := PossibleInstants(dt, z)
	if err != nil {
		return ZonedDateTime{}, err
	}
	for _, p := range ps {
		if z.offsetAt(p) == offset {
			return ZonedDateTime{p, z}, nil
		}
	}
	return ToZoned(dt, z, d)
}

// ---- round

// roundTime rounds a time of day to a multiple of inc units; the days carried out are returned.
func roundTime(t PlainTime, inc int64, unit Unit, mode RoundingMode) (deltaDays int64, out PlainTime) {
	var quantity int64
	switch unit {
	case Day, Hour:
		quantity = t.nanosOfDay()
	case Minute:
		quantity = t.nanosOfDay() - int64(t.Hour)*nsPerUnit[Hour]
	case Second:
		quantity = int64(t.Second)*nsPerUnit[Second] + int64(t.Nano)
	case Millisecond:
		quantity = int64(t.Nano)
	case Microsecond:
		quantity = int64(t.Nano % 1_000_000)
	default:
		quantity = int64(t.Nano % 1000)
	}
	rounded := roundInt(quantity, nsPerUnit[unit]*inc, mode)
	base := t.nanosOfDay() - quantity
	if unit == Day {
		return rounded / nsPerUnit[Day], PlainTime{}
	}
	total := base + rounded
	return total / nsPerUnit[Day], timeFromNanosOfDay(total % nsPerUnit[Day])
}

// Round rounds the time of day; a result of 24:00 wraps to 00:00.
func (t PlainTime) Round(unit Unit, inc int64, mode RoundingMode) (PlainTime, error) {
	if unit == Day {
		return PlainTime{}, errors.New("a time cannot be rounded to a day")
	}
	if err := validateIncrement(unit, inc); err != nil {
		return PlainTime{}, err
	}
	_, out := roundTime(t, inc, unit, mode)
	return out, nil
}

// Round rounds the time of day, carrying into the date.
func (dt PlainDateTime) Round(unit Unit, inc int64, mode RoundingMode) (PlainDateTime, error) {
	if err := validateIncrement(unit, inc); err != nil {
		return PlainDateTime{}, err
	}
	delta, t := roundTime(dt.time, inc, unit, mode)
	d, err := addDays(dt.date, delta)
	if err != nil {
		return PlainDateTime{}, err
	}
	return PlainDateTime{d, t}, nil
}

// Round rounds the nanoseconds since the epoch.
func (i Instant) Round(unit Unit, inc int64, mode RoundingMode) (Instant, error) {
	if unit <= Day {
		return Instant{}, fmt.Errorf("an instant cannot be rounded to a %s", unit)
	}
	if err := validateIncrement(unit, inc); err != nil {
		return Instant{}, err
	}
	return instantFromNs(roundBig(i.ns(), big.NewInt(nsPerUnit[unit]*inc), mode))
}

// Round rounds to a day by the fraction of the real day that has passed, choosing the start
// of this date or the next; to a smaller unit on the wall clock, placed in the zone with the
// offset kept when the zone has it.
func (z ZonedDateTime) Round(unit Unit, inc int64, mode RoundingMode) (ZonedDateTime, error) {
	if err := validateIncrement(unit, inc); err != nil {
		return ZonedDateTime{}, err
	}
	if unit == Day {
		start, err := z.StartOfDay()
		if err != nil {
			return ZonedDateTime{}, err
		}
		next, err := addDays(z.Date(), 1)
		if err != nil {
			return ZonedDateTime{}, err
		}
		end, err := ToZoned(PlainDateTime{date: next}, z.zone, Compatible)
		if err != nil {
			return ZonedDateTime{}, err
		}
		this := z.inst.ns()
		if this.Cmp(end.inst.ns()) >= 0 {
			this = new(big.Int).Sub(end.inst.ns(), big.NewInt(1))
		}
		dayLength := new(big.Int).Sub(end.inst.ns(), start.inst.ns())
		progress := new(big.Int).Sub(this, start.inst.ns())
		rounded := roundBig(progress, dayLength, mode)
		i, err := instantFromNs(new(big.Int).Add(start.inst.ns(), rounded))
		if err != nil {
			return ZonedDateTime{}, err
		}
		return ZonedDateTime{i, z.zone}, nil
	}
	dt, err := z.PlainDateTime().Round(unit, inc, mode)
	if err != nil {
		return ZonedDateTime{}, err
	}
	return interpretOffset(dt, z.zone, z.OffsetSeconds(), Compatible)
}

// ---- transitions

func (i Instant) goTime(z *Zone) time.Time { return time.Unix(i.sec, int64(i.nsec)).In(z.loc) }

func instantOfGo(t time.Time) (Instant, bool) {
	i, err := NewInstant(t.Unix(), int64(t.Nanosecond()))
	return i, err == nil
}

// NextTransition is the first instant after z at which the zone's offset changes, if any; a
// change of abbreviation alone is skipped.
func (z ZonedDateTime) NextTransition() (ZonedDateTime, bool) {
	if z.zone.fixed {
		return ZonedDateTime{}, false
	}
	t := z.inst.goTime(z.zone)
	for {
		_, end := t.ZoneBounds()
		if end.IsZero() {
			return ZonedDateTime{}, false
		}
		i, ok := instantOfGo(end)
		if !ok {
			return ZonedDateTime{}, false
		}
		before, _ := NewInstant(i.sec, int64(i.nsec)-1)
		if z.zone.offsetAt(i) != z.zone.offsetAt(before) {
			return ZonedDateTime{i, z.zone}, true
		}
		t = end
	}
}

// PreviousTransition is the last instant before z at which the offset changed, if any.
func (z ZonedDateTime) PreviousTransition() (ZonedDateTime, bool) {
	if z.zone.fixed {
		return ZonedDateTime{}, false
	}
	t := z.inst.goTime(z.zone).Add(-time.Nanosecond)
	for {
		start, _ := t.ZoneBounds()
		if start.IsZero() {
			return ZonedDateTime{}, false
		}
		i, ok := instantOfGo(start)
		if !ok {
			return ZonedDateTime{}, false
		}
		before, err := NewInstant(i.sec, int64(i.nsec)-1)
		if err != nil {
			return ZonedDateTime{}, false
		}
		if z.zone.offsetAt(i) != z.zone.offsetAt(before) {
			return ZonedDateTime{i, z.zone}, true
		}
		t = start.Add(-time.Nanosecond)
	}
}
