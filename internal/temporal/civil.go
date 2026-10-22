package temporal

// The proleptic Gregorian calendar, as ISO 8601 has it: days counted from 1970-01-01, with the
// algorithms of Howard Hinnant's "chrono-compatible low-level date algorithms".

// daysFromCivil is the days since 1970-01-01 of a date.
func daysFromCivil(y int64, m, d int) int64 {
	if m <= 2 {
		y--
	}
	era := floorDiv(y, 400)
	yoe := y - era*400                     // year of era, 0 to 399
	mp := int64((m + 9) % 12)              // month from March, 0 to 11
	doy := (153*mp+2)/5 + int64(d) - 1     // day of the year from March 1
	doe := yoe*365 + yoe/4 - yoe/100 + doy // day of the era
	return era*146097 + doe - 719468
}

// civilFromDays is the date of a day count since 1970-01-01.
func civilFromDays(z int64) (y int64, m, d int) {
	z += 719468
	era := floorDiv(z, 146097)
	doe := z - era*146097
	yoe := (doe - doe/1460 + doe/36524 - doe/146096) / 365
	y = yoe + era*400
	doy := doe - (365*yoe + yoe/4 - yoe/100)
	mp := (5*doy + 2) / 153
	d = int(doy - (153*mp+2)/5 + 1)
	m = int(mp + 3)
	if m > 12 {
		m -= 12
	}
	if m <= 2 {
		y++
	}
	return y, m, d
}

func isLeap(y int64) bool { return y%4 == 0 && (y%100 != 0 || y%400 == 0) }

func daysInMonth(y int64, m int) int {
	switch m {
	case 2:
		if isLeap(y) {
			return 29
		}
		return 28
	case 4, 6, 9, 11:
		return 30
	}
	return 31
}

func daysInYear(y int64) int {
	if isLeap(y) {
		return 366
	}
	return 365
}

// dayOfWeek is 1 Monday to 7 Sunday for a day count; 1970-01-01 was a Thursday.
func dayOfWeek(days int64) int { return int(floorMod(days+3, 7)) + 1 }

func dayOfYear(y int64, m, d int) int {
	return int(daysFromCivil(y, m, d)-daysFromCivil(y, 1, 1)) + 1
}

// weeksInYear is 52 or 53: a year has 53 ISO weeks when it starts on a Thursday, or on a
// Wednesday in a leap year.
func weeksInYear(y int64) int {
	jan1 := dayOfWeek(daysFromCivil(y, 1, 1))
	if jan1 == 4 || jan1 == 3 && isLeap(y) {
		return 53
	}
	return 52
}

// isoWeek is the ISO 8601 week and its year, which differs from the calendar year around
// New Year.
func isoWeek(y int64, m, d int) (weekYear int64, week int) {
	days := daysFromCivil(y, m, d)
	week = (dayOfYear(y, m, d) - dayOfWeek(days) + 10) / 7
	switch {
	case week < 1:
		return y - 1, weeksInYear(y - 1)
	case week > weeksInYear(y):
		return y + 1, 1
	}
	return y, week
}
