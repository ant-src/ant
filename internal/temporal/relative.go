package temporal

import (
	"errors"
	"fmt"
	"math/big"
)

// Rounding, totals and comparisons of durations relative to a point, as Temporal's
// RoundRelativeDuration has them: a calendar unit is measured by placing the duration's bounds
// on the time line around the point.

// RelativeTo is the point a calendar unit is measured from: a ZonedDateTime, whose days are
// real days, or a PlainDate, whose days are 24 hours; nil is none.
type RelativeTo interface{ relative() }

func (ZonedDateTime) relative() {}
func (PlainDate) relative()     {}

// epochNsFor is a wall-clock time's epoch nanoseconds in a zone, or as UTC without one.
func epochNsFor(dt PlainDateTime, z *Zone) (*big.Int, error) {
	if z == nil {
		return utcNs(dt), nil
	}
	zd, err := ToZoned(dt, z, Compatible)
	if err != nil {
		return nil, err
	}
	return zd.inst.ns(), nil
}

type nudgeResult struct {
	duration  internalDur
	nudgedNs  *big.Int
	didExpand bool
	total     float64 // with an increment of 1: the exact count of the unit, fraction included
}

type nudgeWindow struct {
	r1, r2                     int64
	startNs, endNs             *big.Int
	startDuration, endDuration dateDur
}

func computeNudgeWindow(sign int, d internalDur, originNs *big.Int, dt PlainDateTime, z *Zone, inc int64, unit Unit, additionalShift bool) (nudgeWindow, error) {
	var w nudgeWindow
	s := int64(sign)
	switch unit {
	case Year:
		years := d.date.years / inc * inc
		w.r1 = years
		if additionalShift {
			w.r1 = years + inc*s
		}
		w.r2 = w.r1 + inc*s
		w.startDuration = dateDur{years: w.r1}
		w.endDuration = dateDur{years: w.r2}
	case Month:
		months := d.date.months / inc * inc
		w.r1 = months
		if additionalShift {
			w.r1 = months + inc*s
		}
		w.r2 = w.r1 + inc*s
		w.startDuration = dateDur{d.date.years, w.r1, 0, 0}
		w.endDuration = dateDur{d.date.years, w.r2, 0, 0}
	case Week:
		yearsMonths := dateDur{d.date.years, d.date.months, 0, 0}
		weeksStart, err := dateAdd(dt.date, yearsMonths)
		if err != nil {
			return w, err
		}
		weeksEnd, err := addDays(weeksStart, d.date.days)
		if err != nil {
			return w, err
		}
		until := dateUntil(weeksStart, weeksEnd, Week)
		weeks := (d.date.weeks + until.weeks) / inc * inc
		w.r1 = weeks
		w.r2 = weeks + inc*s
		w.startDuration = dateDur{d.date.years, d.date.months, w.r1, 0}
		w.endDuration = dateDur{d.date.years, d.date.months, w.r2, 0}
	default: // Day
		days := d.date.days / inc * inc
		w.r1 = days
		w.r2 = days + inc*s
		w.startDuration = dateDur{d.date.years, d.date.months, d.date.weeks, w.r1}
		w.endDuration = dateDur{d.date.years, d.date.months, d.date.weeks, w.r2}
	}
	if w.startDuration.sign() == 0 {
		w.startNs = originNs
	} else {
		start, err := dateAdd(dt.date, w.startDuration)
		if err != nil {
			return w, err
		}
		if w.startNs, err = epochNsFor(PlainDateTime{start, dt.time}, z); err != nil {
			return w, err
		}
	}
	end, err := dateAdd(dt.date, w.endDuration)
	if err != nil {
		return w, err
	}
	if w.endNs, err = epochNsFor(PlainDateTime{end, dt.time}, z); err != nil {
		return w, err
	}
	return w, nil
}

func within(sign int, lo, x, hi *big.Int) bool {
	if sign == 1 {
		return lo.Cmp(x) <= 0 && x.Cmp(hi) <= 0
	}
	return hi.Cmp(x) <= 0 && x.Cmp(lo) <= 0
}

func nudgeToCalendarUnit(sign int, d internalDur, originNs, destNs *big.Int, dt PlainDateTime, z *Zone, inc int64, unit Unit, mode RoundingMode) (nudgeResult, error) {
	didExpand := false
	w, err := computeNudgeWindow(sign, d, originNs, dt, z, inc, unit, false)
	if err != nil {
		return nudgeResult{}, err
	}
	if !within(sign, w.startNs, destNs, w.endNs) {
		if w, err = computeNudgeWindow(sign, d, originNs, dt, z, inc, unit, true); err != nil {
			return nudgeResult{}, err
		}
		if !within(sign, w.startNs, destNs, w.endNs) {
			return nudgeResult{}, fmt.Errorf("a %s was zero days long", unit)
		}
		didExpand = true
	}
	if w.endNs.Cmp(w.startNs) == 0 {
		return nudgeResult{}, fmt.Errorf("a %s was zero days long", unit)
	}
	numerator := new(big.Int).Sub(destNs, w.startNs)
	denominator := new(big.Int).Sub(w.endNs, w.startNs)
	cmp := new(big.Int).Abs(new(big.Int).Mul(numerator, big.NewInt(2))).Cmp(new(big.Int).Abs(denominator))
	even := (abs64(w.r1)/inc)%2 == 0
	absR1, absR2 := big.NewInt(abs64(w.r1)), big.NewInt(abs64(w.r2))
	var roundedUnit *big.Int
	switch {
	case numerator.Sign() == 0:
		roundedUnit = absR1
	case numerator.Cmp(denominator) == 0:
		roundedUnit = absR2
	default:
		roundedUnit = applyUnsigned(absR1, absR2, cmp, even, unsignedModeOf(mode, sign < 0))
	}
	var total float64
	if inc == 1 {
		fake := new(big.Int).Mul(denominator, big.NewInt(w.r1))
		fake.Add(fake, new(big.Int).Mul(numerator, big.NewInt(int64(sign))))
		total, _ = new(big.Rat).SetFrac(fake, denominator).Float64()
	}
	expanded := roundedUnit.Cmp(absR2) == 0
	didExpand = didExpand || expanded
	res := nudgeResult{duration: internalDur{w.startDuration, zeroTime}, nudgedNs: w.startNs, didExpand: didExpand, total: total}
	if expanded {
		res.duration = internalDur{w.endDuration, zeroTime}
	}
	if didExpand {
		res.nudgedNs = w.endNs
	}
	return res, nil
}

func nudgeToZonedTime(sign int, d internalDur, dt PlainDateTime, z *Zone, inc int64, unit Unit, mode RoundingMode) (nudgeResult, error) {
	start, err := dateAdd(dt.date, d.date)
	if err != nil {
		return nudgeResult{}, err
	}
	endDate, err := addDays(start, int64(sign))
	if err != nil {
		return nudgeResult{}, err
	}
	startNs, err := epochNsFor(PlainDateTime{start, dt.time}, z)
	if err != nil {
		return nudgeResult{}, err
	}
	endNs, err := epochNsFor(PlainDateTime{endDate, dt.time}, z)
	if err != nil {
		return nudgeResult{}, err
	}
	daySpan := timeDur{new(big.Int).Sub(endNs, startNs)}
	if daySpan.sign() != sign {
		return nudgeResult{}, errors.New("the zone gave inconsistent instants")
	}
	unitInc := new(big.Int).Mul(big.NewInt(nsPerUnit[unit]), big.NewInt(inc))
	rounded, err := d.time.round(unitInc, mode)
	if err != nil {
		return nudgeResult{}, err
	}
	beyond, err := rounded.sub(daySpan)
	if err != nil {
		return nudgeResult{}, err
	}
	didRoundBeyondDay := beyond.sign() != -sign
	var dayDelta int64
	var nudgedNs *big.Int
	if didRoundBeyondDay {
		dayDelta = int64(sign)
		if rounded, err = beyond.round(unitInc, mode); err != nil {
			return nudgeResult{}, err
		}
		nudgedNs = new(big.Int).Add(endNs, rounded.ns)
	} else {
		nudgedNs = new(big.Int).Add(startNs, rounded.ns)
	}
	date := d.date
	date.days += dayDelta
	return nudgeResult{duration: internalDur{date, rounded}, nudgedNs: nudgedNs, didExpand: didRoundBeyondDay}, nil
}

func nudgeToDayOrTime(d internalDur, destNs *big.Int, largest Unit, inc int64, smallest Unit, mode RoundingMode) (nudgeResult, error) {
	td, err := d.time.add24HourDays(d.date.days)
	if err != nil {
		return nudgeResult{}, err
	}
	rounded, err := td.round(new(big.Int).Mul(big.NewInt(nsPerUnit[smallest]), big.NewInt(inc)), mode)
	if err != nil {
		return nudgeResult{}, err
	}
	diff, _ := rounded.sub(td)
	wholeDays, _ := td.divmod(bigNsPerDay)
	roundedWholeDays, _ := rounded.divmod(bigNsPerDay)
	deltaDays := new(big.Int).Sub(roundedWholeDays, wholeDays)
	didExpand := deltaDays.Sign() == td.sign() && deltaDays.Sign() != 0
	nudgedNs := new(big.Int).Add(destNs, diff.ns)
	var days int64
	remainder := rounded
	if isDateUnit(largest) {
		days = roundedWholeDays.Int64()
		if remainder, err = rounded.add24HourDays(-days); err != nil {
			return nudgeResult{}, err
		}
	}
	date := d.date
	date.days = days
	return nudgeResult{duration: internalDur{date, remainder}, nudgedNs: nudgedNs, didExpand: didExpand}, nil
}

func bubbleRelativeDuration(sign int, d internalDur, nudgedNs *big.Int, dt PlainDateTime, z *Zone, largest, smallest Unit) (internalDur, error) {
	if smallest == largest {
		return d, nil
	}
	for unit := smallest - 1; unit >= largest; unit-- {
		if unit == Week && largest != Week {
			continue
		}
		var end dateDur
		switch unit {
		case Year:
			end = dateDur{years: d.date.years + int64(sign)}
		case Month:
			end = dateDur{d.date.years, d.date.months + int64(sign), 0, 0}
		case Week:
			end = dateDur{d.date.years, d.date.months, d.date.weeks + int64(sign), 0}
		default:
			return d, nil
		}
		endDate, err := dateAdd(dt.date, end)
		if err != nil {
			return internalDur{}, err
		}
		endNs, err := epochNsFor(PlainDateTime{endDate, dt.time}, z)
		if err != nil {
			return internalDur{}, err
		}
		if nudgedNs.Cmp(endNs) != -sign {
			d = internalDur{end, zeroTime}
		} else {
			break
		}
	}
	return d, nil
}

func roundRelativeDuration(d internalDur, originNs, destNs *big.Int, dt PlainDateTime, z *Zone, largest Unit, inc int64, smallest Unit, mode RoundingMode) (internalDur, error) {
	irregular := isCalendarUnit(smallest) || (z != nil && smallest == Day)
	sign := 1
	if d.sign() < 0 {
		sign = -1
	}
	var res nudgeResult
	var err error
	switch {
	case irregular:
		res, err = nudgeToCalendarUnit(sign, d, originNs, destNs, dt, z, inc, smallest, mode)
	case z != nil:
		res, err = nudgeToZonedTime(sign, d, dt, z, inc, smallest, mode)
	default:
		res, err = nudgeToDayOrTime(d, destNs, largest, inc, smallest, mode)
	}
	if err != nil {
		return internalDur{}, err
	}
	d = res.duration
	if res.didExpand && smallest != Week {
		if d, err = bubbleRelativeDuration(sign, d, res.nudgedNs, dt, z, largest, largerUnit(smallest, Day)); err != nil {
			return internalDur{}, err
		}
	}
	return d, nil
}

func totalRelativeDuration(d internalDur, originNs, destNs *big.Int, dt PlainDateTime, z *Zone, unit Unit) (float64, error) {
	if isCalendarUnit(unit) || (z != nil && unit == Day) {
		sign := 1
		if d.sign() < 0 {
			sign = -1
		}
		res, err := nudgeToCalendarUnit(sign, d, originNs, destNs, dt, z, 1, unit, Trunc)
		if err != nil {
			return 0, err
		}
		return res.total, nil
	}
	td, err := d.time.add24HourDays(d.date.days)
	if err != nil {
		return 0, err
	}
	return td.fdiv(big.NewInt(nsPerUnit[unit])), nil
}

func differencePlainDateTimeWithRounding(dt1, dt2 PlainDateTime, largest Unit, inc int64, smallest Unit, mode RoundingMode) (internalDur, error) {
	if dt1.Compare(dt2) == 0 {
		return internalDur{time: zeroTime}, nil
	}
	d, err := differenceISODateTime(dt1, dt2, largest)
	if err != nil {
		return internalDur{}, err
	}
	if smallest == Nanosecond && inc == 1 {
		return d, nil
	}
	return roundRelativeDuration(d, utcNs(dt1), utcNs(dt2), dt1, nil, largest, inc, smallest, mode)
}

func differencePlainDateTimeWithTotal(dt1, dt2 PlainDateTime, unit Unit) (float64, error) {
	if dt1.Compare(dt2) == 0 {
		return 0, nil
	}
	d, err := differenceISODateTime(dt1, dt2, unit)
	if err != nil {
		return 0, err
	}
	if unit == Nanosecond {
		f, _ := new(big.Float).SetInt(d.time.ns).Float64()
		return f, nil
	}
	return totalRelativeDuration(d, utcNs(dt1), utcNs(dt2), dt1, nil, unit)
}

func differenceZonedDateTimeWithRounding(i1, i2 Instant, z *Zone, largest Unit, inc int64, smallest Unit, mode RoundingMode) (internalDur, error) {
	if !isDateUnit(largest) {
		return differenceInstant(i1, i2, inc, smallest, mode)
	}
	d, err := differenceZonedDateTime(i1, i2, z, largest)
	if err != nil {
		return internalDur{}, err
	}
	if smallest == Nanosecond && inc == 1 {
		return d, nil
	}
	return roundRelativeDuration(d, i1.ns(), i2.ns(), i1.InZone(z).PlainDateTime(), z, largest, inc, smallest, mode)
}

func differenceZonedDateTimeWithTotal(i1, i2 Instant, z *Zone, unit Unit) (float64, error) {
	if !isDateUnit(unit) {
		return timeDur{new(big.Int).Sub(i2.ns(), i1.ns())}.fdiv(big.NewInt(nsPerUnit[unit])), nil
	}
	d, err := differenceZonedDateTime(i1, i2, z, unit)
	if err != nil {
		return 0, err
	}
	return totalRelativeDuration(d, i1.ns(), i2.ns(), i1.InZone(z).PlainDateTime(), z, unit)
}

// dateDurationDays is the days of a date duration counted from a date.
func dateDurationDays(dd dateDur, rel PlainDate) (int64, error) {
	ymw := dateDur{dd.years, dd.months, dd.weeks, 0}
	if ymw.sign() == 0 {
		return dd.days, nil
	}
	later, err := dateAdd(rel, ymw)
	if err != nil {
		return 0, err
	}
	return dd.days + later.days() - rel.days(), nil
}

// target is the point a duration reaches from a plain date, and the date-time of the date.
func plainTarget(d Duration, rel PlainDate) (origin, target PlainDateTime, err error) {
	id, err := d.internal24()
	if err != nil {
		return
	}
	delta, t := addTime(PlainTime{}, id.time)
	dd := id.date
	dd.days = delta
	targetDate, err := dateAdd(rel, dd)
	if err != nil {
		return
	}
	return PlainDateTime{date: rel}, PlainDateTime{targetDate, t}, nil
}

// AddDurations sums two durations with no years, months or weeks, days at 24 hours, balanced
// up to the larger of their largest units.
func AddDurations(a, b Duration) (Duration, error) {
	largest := largerUnit(a.defaultLargestUnit(), b.defaultLargestUnit())
	if isCalendarUnit(largest) {
		return Duration{}, errors.New("years, months and weeks cannot be added without a starting point; add each to a date")
	}
	ia, err := a.internal24()
	if err != nil {
		return Duration{}, err
	}
	ib, err := b.internal24()
	if err != nil {
		return Duration{}, err
	}
	sum, err := ia.time.add(ib.time)
	if err != nil {
		return Duration{}, err
	}
	return fromInternal(internalDur{time: sum}, largest)
}

// CompareDurations orders two durations by length: with a zoned point for any date unit, with
// a date for years, months or weeks, and days as 24 hours otherwise.
func CompareDurations(a, b Duration, rel RelativeTo) (int, error) {
	la, lb := a.defaultLargestUnit(), b.defaultLargestUnit()
	if rel == nil && (isCalendarUnit(la) || isCalendarUnit(lb)) {
		// before the shortcut for equal durations, so the panic does not depend on the values
		return 0, errors.New("years, months or weeks need a relativeTo point to compare")
	}
	if a.fields() == b.fields() {
		return 0, nil
	}
	ia, err := a.internal()
	if err != nil {
		return 0, err
	}
	ib, err := b.internal()
	if err != nil {
		return 0, err
	}
	if z, ok := rel.(ZonedDateTime); ok && (isDateUnit(la) || isDateUnit(lb)) {
		after1, err := z.Add(a)
		if err != nil {
			return 0, err
		}
		after2, err := z.Add(b)
		if err != nil {
			return 0, err
		}
		return after1.inst.Compare(after2.inst), nil
	}
	d1, d2 := ia.date.days, ib.date.days
	if isCalendarUnit(la) || isCalendarUnit(lb) {
		pd, ok := rel.(PlainDate)
		if !ok {
			return 0, errors.New("years, months or weeks need a relativeTo point to compare")
		}
		if d1, err = dateDurationDays(ia.date, pd); err != nil {
			return 0, err
		}
		if d2, err = dateDurationDays(ib.date, pd); err != nil {
			return 0, err
		}
	}
	t1, err := ia.time.add24HourDays(d1)
	if err != nil {
		return 0, err
	}
	t2, err := ib.time.add24HourDays(d2)
	if err != nil {
		return 0, err
	}
	return t1.cmp(t2), nil
}

// Total is the whole duration in a unit, fraction included.
func (d Duration) Total(unit Unit, rel RelativeTo) (float64, error) {
	switch r := rel.(type) {
	case ZonedDateTime:
		target, err := r.Add(d)
		if err != nil {
			return 0, err
		}
		return differenceZonedDateTimeWithTotal(r.inst, target.inst, r.zone, unit)
	case PlainDate:
		origin, target, err := plainTarget(d, r)
		if err != nil {
			return 0, err
		}
		return differencePlainDateTimeWithTotal(origin, target, unit)
	}
	if isCalendarUnit(d.defaultLargestUnit()) {
		return 0, fmt.Errorf("a total with %ss needs a relativeTo point", d.defaultLargestUnit())
	}
	if isCalendarUnit(unit) {
		return 0, fmt.Errorf("a total in %ss needs a relativeTo point", unit)
	}
	id, err := d.internal24()
	if err != nil {
		return 0, err
	}
	return id.time.fdiv(big.NewInt(nsPerUnit[unit])), nil
}

// Round balances the duration up to largest and rounds it to a multiple of inc of smallest;
// hasSmallest and hasLargest say which were given, and at least one must be.
func (d Duration) Round(smallest Unit, hasSmallest bool, largest Unit, hasLargest bool, inc int64, mode RoundingMode, rel RelativeTo) (Duration, error) {
	existing := d.defaultLargestUnit()
	if !hasSmallest {
		smallest = Nanosecond
	}
	defaultLargest := largerUnit(existing, smallest)
	if !hasLargest {
		largest = defaultLargest
	}
	if !hasSmallest && !hasLargest {
		return Duration{}, errors.New("roundDuration needs a smallest or a largest unit")
	}
	if largerUnit(largest, smallest) != largest {
		return Duration{}, fmt.Errorf("the largest unit, %ss, is smaller than the smallest, %ss", largest, smallest)
	}
	if inc < 1 {
		return Duration{}, fmt.Errorf("the increment must be positive, not %d", inc)
	}
	if smallest >= Hour {
		if err := validateIncrement(smallest, inc); err != nil {
			return Duration{}, err
		}
	} else if inc > 1 && largest != smallest {
		return Duration{}, errors.New("a date unit with an increment above 1 needs the largest unit to be the smallest")
	}
	switch r := rel.(type) {
	case ZonedDateTime:
		target, err := r.Add(d)
		if err != nil {
			return Duration{}, err
		}
		id, err := differenceZonedDateTimeWithRounding(r.inst, target.inst, r.zone, largest, inc, smallest, mode)
		if err != nil {
			return Duration{}, err
		}
		if isDateUnit(largest) {
			largest = Hour
		}
		return fromInternal(id, largest)
	case PlainDate:
		origin, target, err := plainTarget(d, r)
		if err != nil {
			return Duration{}, err
		}
		id, err := differencePlainDateTimeWithRounding(origin, target, largest, inc, smallest, mode)
		if err != nil {
			return Duration{}, err
		}
		return fromInternal(id, largest)
	}
	if isCalendarUnit(existing) {
		return Duration{}, fmt.Errorf("balancing %ss needs a relativeTo point", existing)
	}
	if isCalendarUnit(largest) {
		return Duration{}, fmt.Errorf("balancing to %ss needs a relativeTo point", largest)
	}
	id, err := d.internal24()
	if err != nil {
		return Duration{}, err
	}
	if smallest == Day {
		rounded, err := id.time.round(new(big.Int).Mul(bigNsPerDay, big.NewInt(inc)), mode)
		if err != nil {
			return Duration{}, err
		}
		days, _ := rounded.divmod(bigNsPerDay)
		if !days.IsInt64() {
			return Duration{}, errors.New("the days do not fit an int")
		}
		id = internalDur{dateDur{days: days.Int64()}, zeroTime}
	} else {
		if id.time, err = id.time.round(new(big.Int).Mul(big.NewInt(nsPerUnit[smallest]), big.NewInt(inc)), mode); err != nil {
			return Duration{}, err
		}
	}
	return fromInternal(id, largest)
}
