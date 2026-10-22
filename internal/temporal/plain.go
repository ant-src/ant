package temporal

import "fmt"

// PlainDate is a calendar date with no zone.
type PlainDate struct {
	Year, Month, Day int
}

// NewPlainDate makes a date, or reports an error for a month or day that is not there, or a
// date beyond the range, which is the range of instants and a day more each way.
func NewPlainDate(year, month, day int) (PlainDate, error) {
	if month < 1 || month > 12 || day < 1 || day > daysInMonth(int64(year), month) {
		return PlainDate{}, fmt.Errorf("no such date: %s-%02d-%02d", formatYear(int64(year)), month, day)
	}
	d := PlainDate{year, month, day}
	if days := d.days(); days < -maxDays-1 || days > maxDays+1 {
		return PlainDate{}, ErrRange
	}
	return d, nil
}

func (d PlainDate) days() int64 { return daysFromCivil(int64(d.Year), d.Month, d.Day) }

func dateFromDays(days int64) PlainDate {
	y, m, dd := civilFromDays(days)
	return PlainDate{int(y), m, dd}
}

// Date returns the date itself; it makes a PlainDate a Dated.
func (d PlainDate) Date() PlainDate { return d }

// Compare orders dates.
func (d PlainDate) Compare(e PlainDate) int { return cmp(d.days(), e.days()) }

// DayOfWeek is 1 Monday to 7 Sunday.
func (d PlainDate) DayOfWeek() int { return dayOfWeek(d.days()) }

// DayOfYear is 1 to 366.
func (d PlainDate) DayOfYear() int { return dayOfYear(int64(d.Year), d.Month, d.Day) }

// WeekOfYear is the ISO week, 1 to 53.
func (d PlainDate) WeekOfYear() int { _, w := isoWeek(int64(d.Year), d.Month, d.Day); return w }

// YearOfWeek is the year of the ISO week, which differs from Year around New Year.
func (d PlainDate) YearOfWeek() int { y, _ := isoWeek(int64(d.Year), d.Month, d.Day); return int(y) }

// DaysInMonth is 28 to 31.
func (d PlainDate) DaysInMonth() int { return daysInMonth(int64(d.Year), d.Month) }

// DaysInYear is 365 or 366.
func (d PlainDate) DaysInYear() int { return daysInYear(int64(d.Year)) }

// InLeapYear reports a leap year.
func (d PlainDate) InLeapYear() bool { return isLeap(int64(d.Year)) }

// TimeType is the name of the type's tag.
func (PlainDate) TimeType() string { return "Time.PlainDate" }

// Key is the canonical form == compares.
func (d PlainDate) Key() string { return d.String() }

// String is 2026-10-22.
func (d PlainDate) String() string {
	return fmt.Sprintf("%s-%02d-%02d", formatYear(int64(d.Year)), d.Month, d.Day)
}

// PlainTime is a wall-clock time with no date; Nano is the nanosecond within the second.
type PlainTime struct {
	Hour, Minute, Second, Nano int
}

// NewPlainTime makes a time from its six fields, each checked against its range.
func NewPlainTime(hour, minute, second, millisecond, microsecond, nanosecond int) (PlainTime, error) {
	for _, f := range []struct {
		name   string
		v, max int
	}{
		{"hour", hour, 23}, {"minute", minute, 59}, {"second", second, 59},
		{"millisecond", millisecond, 999}, {"microsecond", microsecond, 999}, {"nanosecond", nanosecond, 999},
	} {
		if f.v < 0 || f.v > f.max {
			return PlainTime{}, fmt.Errorf("%s out of range: %d", f.name, f.v)
		}
	}
	return PlainTime{hour, minute, second, millisecond*1_000_000 + microsecond*1000 + nanosecond}, nil
}

// Millisecond, Microsecond and Nanosecond are the three parts of Nano.
func (t PlainTime) Millisecond() int { return t.Nano / 1_000_000 }
func (t PlainTime) Microsecond() int { return t.Nano / 1000 % 1000 }
func (t PlainTime) Nanosecond() int  { return t.Nano % 1000 }

// Clock returns the time itself; it makes a PlainTime a Timed.
func (t PlainTime) Clock() PlainTime { return t }

func (t PlainTime) nanosOfDay() int64 {
	return (int64(t.Hour)*3600+int64(t.Minute)*60+int64(t.Second))*nanosPerSecond + int64(t.Nano)
}

func timeFromNanosOfDay(n int64) PlainTime {
	sec := n / nanosPerSecond
	return PlainTime{int(sec / 3600), int(sec / 60 % 60), int(sec % 60), int(n % nanosPerSecond)}
}

// Compare orders times.
func (t PlainTime) Compare(u PlainTime) int { return cmp(t.nanosOfDay(), u.nanosOfDay()) }

// TimeType is the name of the type's tag.
func (PlainTime) TimeType() string { return "Time.PlainTime" }

// Key is the canonical form == compares.
func (t PlainTime) Key() string { return t.String() }

// String is 14:03:05.123.
func (t PlainTime) String() string { return formatClock(t.Hour, t.Minute, t.Second, t.Nano) }

// PlainDateTime is a date and a time with no zone.
type PlainDateTime struct {
	date PlainDate
	time PlainTime
}

// NewPlainDateTime joins a date and a time.
func NewPlainDateTime(d PlainDate, t PlainTime) PlainDateTime { return PlainDateTime{d, t} }

// Date is the date part; it makes a PlainDateTime a Dated.
func (dt PlainDateTime) Date() PlainDate { return dt.date }

// Clock is the time part; it makes a PlainDateTime a Timed.
func (dt PlainDateTime) Clock() PlainTime { return dt.time }

// Compare orders date-times.
func (dt PlainDateTime) Compare(e PlainDateTime) int {
	if c := dt.date.Compare(e.date); c != 0 {
		return c
	}
	return dt.time.Compare(e.time)
}

// TimeType is the name of the type's tag.
func (PlainDateTime) TimeType() string { return "Time.PlainDateTime" }

// Key is the canonical form == compares.
func (dt PlainDateTime) Key() string { return dt.String() }

// String is 2026-10-22T14:03:05.
func (dt PlainDateTime) String() string { return dt.date.String() + "T" + dt.time.String() }

// asUTC is the date-time's seconds since the epoch were it in UTC, and its nanosecond.
func (dt PlainDateTime) asUTC() (sec int64, nsec int64) {
	n := dt.time.nanosOfDay()
	return dt.date.days()*secondsPerDay + n/nanosPerSecond, n % nanosPerSecond
}

// Dated is a value with a date: a ZonedDateTime, a PlainDate or a PlainDateTime.
type Dated interface{ Date() PlainDate }

// Timed is a value with a time: a ZonedDateTime, a PlainTime or a PlainDateTime.
type Timed interface{ Clock() PlainTime }

func cmp(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}
