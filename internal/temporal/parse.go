package temporal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// An RFC 9557 string: ISO 8601's extended format with the zone and the calendar in brackets.
// Each parser takes the part its type needs and ignores the rest, as Temporal's from does.
type parsed struct {
	hasDate          bool
	year             int64
	month, day       int
	hasTime          bool
	hour, minute     int
	second, nano     int
	hasOffset        bool
	utc              bool // Z
	offset           int  // seconds
	offsetHasSeconds bool
	zone             string
	calendar         string
}

type scanner struct {
	s string
	i int
}

func (sc *scanner) peek() byte {
	if sc.i < len(sc.s) {
		return sc.s[sc.i]
	}
	return 0
}

func (sc *scanner) digits(n int) (int64, bool) {
	if sc.i+n > len(sc.s) {
		return 0, false
	}
	var v int64
	for _, c := range []byte(sc.s[sc.i : sc.i+n]) {
		if c < '0' || c > '9' {
			return 0, false
		}
		v = v*10 + int64(c-'0')
	}
	sc.i += n
	return v, true
}

func (sc *scanner) take(c byte) bool {
	if sc.peek() == c {
		sc.i++
		return true
	}
	return false
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// parseISO reads a date, a time, an offset and annotations, each optional, and rejects
// anything else.
func parseISO(s string) (parsed, error) {
	var p parsed
	sc := &scanner{s: s}
	bad := func() (parsed, error) { return parsed{}, fmt.Errorf("not a date or time: %q", s) }
	// A date, unless the string starts as a time: HH: or T.
	startsTime := len(s) >= 3 && isDigit(s[0]) && isDigit(s[1]) && s[2] == ':' || sc.peek() == 'T' || sc.peek() == 't'
	if !startsTime {
		sign := int64(1)
		var ok bool
		switch sc.peek() {
		case '+', '-':
			if sc.s[sc.i] == '-' {
				sign = -1
			}
			sc.i++
			p.year, ok = sc.digits(6)
		default:
			p.year, ok = sc.digits(4)
		}
		if !ok {
			return bad()
		}
		p.year *= sign
		if sign == -1 && p.year == 0 {
			return bad() // -000000 is not a year
		}
		if !sc.take('-') {
			return bad()
		}
		m, ok1 := sc.digits(2)
		if !sc.take('-') {
			return bad()
		}
		d, ok2 := sc.digits(2)
		if !ok1 || !ok2 {
			return bad()
		}
		p.hasDate, p.month, p.day = true, int(m), int(d)
		if sc.i == len(s) || sc.peek() == '[' {
			return finish(p, sc, s)
		}
		if !sc.take('T') && !sc.take('t') && !sc.take(' ') {
			return bad()
		}
	} else {
		sc.take('T')
		sc.take('t')
	}
	h, ok := sc.digits(2)
	if !ok || !sc.take(':') {
		return bad()
	}
	m, ok := sc.digits(2)
	if !ok {
		return bad()
	}
	p.hasTime, p.hour, p.minute = true, int(h), int(m)
	if sc.take(':') {
		sec, ok := sc.digits(2)
		if !ok {
			return bad()
		}
		p.second = int(sec)
		if sc.take('.') || sc.take(',') {
			n, ok := sc.fraction()
			if !ok {
				return bad()
			}
			p.nano = n
		}
	}
	switch sc.peek() {
	case 'Z', 'z':
		sc.i++
		p.hasOffset, p.utc = true, true
	case '+', '-':
		off, hasSec, ok := sc.offset()
		if !ok {
			return bad()
		}
		p.hasOffset, p.offset, p.offsetHasSeconds = true, off, hasSec
	}
	return finish(p, sc, s)
}

// fraction reads one to nine digits after the point as nanoseconds.
func (sc *scanner) fraction() (int, bool) {
	start := sc.i
	for sc.i < len(sc.s) && isDigit(sc.s[sc.i]) && sc.i-start < 9 {
		sc.i++
	}
	if sc.i == start {
		return 0, false
	}
	digits := sc.s[start:sc.i]
	n, _ := strconv.Atoi(digits + strings.Repeat("0", 9-len(digits)))
	return n, true
}

// offset reads ±HH, ±HH:MM, ±HHMM, ±HH:MM:SS or ±HH:MM:SS.fraction.
func (sc *scanner) offset() (seconds int, hasSeconds, ok bool) {
	sign := 1
	if sc.s[sc.i] == '-' {
		sign = -1
	}
	sc.i++
	h, ok := sc.digits(2)
	if !ok {
		return 0, false, false
	}
	var m, s int64
	colon := sc.take(':')
	if isDigit(sc.peek()) {
		if m, ok = sc.digits(2); !ok {
			return 0, false, false
		}
		if colon && sc.take(':') {
			if s, ok = sc.digits(2); !ok {
				return 0, false, false
			}
			hasSeconds = true
			if sc.take('.') || sc.take(',') {
				if _, ok = sc.fraction(); !ok {
					return 0, false, false
				}
			}
		}
	} else if colon {
		return 0, false, false
	}
	if h > 23 || m > 59 || s > 59 {
		return 0, false, false
	}
	return sign * int(h*3600+m*60+s), hasSeconds, true
}

// finish reads the bracketed annotations: the zone, then key=value pairs.
func finish(p parsed, sc *scanner, s string) (parsed, error) {
	for sc.i < len(s) {
		if !sc.take('[') {
			return parsed{}, fmt.Errorf("not a date or time: %q", s)
		}
		sc.take('!')
		end := strings.IndexByte(s[sc.i:], ']')
		if end < 0 {
			return parsed{}, fmt.Errorf("not a date or time: %q", s)
		}
		body := s[sc.i : sc.i+end]
		sc.i += end + 1
		if key, val, ok := strings.Cut(body, "="); ok {
			if key == "u-ca" {
				if p.calendar != "" {
					return parsed{}, fmt.Errorf("two calendars in %q", s)
				}
				p.calendar = val
			}
			continue
		}
		if p.zone != "" {
			return parsed{}, fmt.Errorf("two time zones in %q", s)
		}
		if body == "" {
			return parsed{}, fmt.Errorf("not a date or time: %q", s)
		}
		p.zone = body
	}
	if p.calendar != "" && !strings.EqualFold(p.calendar, "iso8601") {
		return parsed{}, fmt.Errorf("calendar %s is not supported; only iso8601 is", p.calendar)
	}
	return p, nil
}

func (p parsed) date() (PlainDate, error) {
	if p.year < -999_999 || p.year > 999_999 {
		return PlainDate{}, ErrRange
	}
	return NewPlainDate(int(p.year), p.month, p.day)
}

func (p parsed) clock() (PlainTime, error) {
	return NewPlainTime(p.hour, p.minute, p.second, p.nano/1_000_000, p.nano/1000%1000, p.nano%1000)
}

// plainZ rejects Z without a bracketed zone on a plain type: Z names a moment, not a
// wall-clock time.
func (p parsed) plainZ() error {
	if p.utc && p.zone == "" {
		return errors.New("Z names an exact time; a plain value needs a wall-clock time")
	}
	return nil
}

// ParseInstant reads a date, a time and Z or an offset; a bracketed zone is ignored.
func ParseInstant(s string) (Instant, error) {
	p, err := parseISO(s)
	if err != nil {
		return Instant{}, err
	}
	if !p.hasDate || !p.hasTime || !p.hasOffset {
		return Instant{}, fmt.Errorf("an instant needs a date, a time and an offset: %q", s)
	}
	dt, err := p.dateTime()
	if err != nil {
		return Instant{}, err
	}
	sec, nsec := dt.asUTC()
	if !p.utc {
		sec -= int64(p.offset)
	}
	return NewInstant(sec, nsec)
}

func (p parsed) dateTime() (PlainDateTime, error) {
	d, err := p.date()
	if err != nil {
		return PlainDateTime{}, err
	}
	t, err := p.clock()
	if err != nil {
		return PlainDateTime{}, err
	}
	return PlainDateTime{d, t}, nil
}

// ParseZoned reads a date, a time, a bracketed zone, and an offset or Z. Without a time it is
// the start of the day; without an offset the wall-clock time is resolved by d; Z gives the exact moment, shown in the zone; an offset
// the zone does not have at that time is an error.
func ParseZoned(s string, d Disambiguation) (ZonedDateTime, error) {
	p, err := parseISO(s)
	if err != nil {
		return ZonedDateTime{}, err
	}
	if !p.hasDate || p.zone == "" {
		return ZonedDateTime{}, fmt.Errorf("a zoned date-time needs a date and a [zone]: %q", s)
	}
	z, err := LoadZone(p.zone)
	if err != nil {
		return ZonedDateTime{}, err
	}
	dt, err := p.dateTime()
	if err != nil {
		return ZonedDateTime{}, err
	}
	if !p.hasTime {
		// a date alone is the start of that day, as in Temporal, not midnight disambiguated
		return startOfDate(dt.date, z)
	}
	if p.utc {
		sec, nsec := dt.asUTC()
		i, err := NewInstant(sec, nsec)
		if err != nil {
			return ZonedDateTime{}, err
		}
		return ZonedDateTime{i, z}, nil
	}
	if !p.hasOffset {
		return ToZoned(dt, z, d)
	}
	ps, err := PossibleInstants(dt, z)
	if err != nil {
		return ZonedDateTime{}, err
	}
	for _, i := range ps {
		off := z.offsetAt(i)
		if off == p.offset || !p.offsetHasSeconds && roundToMinute(off) == p.offset {
			return ZonedDateTime{i, z}, nil
		}
	}
	return ZonedDateTime{}, fmt.Errorf("the offset %s is not the offset of %s at %s", formatOffset(p.offset), z.id, dt)
}

func roundToMinute(off int) int {
	r := off % 60
	off -= r
	if r >= 30 {
		off += 60
	} else if r <= -30 {
		off -= 60
	}
	return off
}

// ParsePlainDate reads a date, alone or at the start of a longer string.
func ParsePlainDate(s string) (PlainDate, error) {
	p, err := parseISO(s)
	if err != nil {
		return PlainDate{}, err
	}
	if !p.hasDate {
		return PlainDate{}, fmt.Errorf("not a date: %q", s)
	}
	if err := p.plainZ(); err != nil {
		return PlainDate{}, err
	}
	if p.hasTime {
		if _, err := p.clock(); err != nil {
			return PlainDate{}, err
		}
	}
	return p.date()
}

// ParsePlainTime reads a time, alone or after a date.
func ParsePlainTime(s string) (PlainTime, error) {
	p, err := parseISO(s)
	if err != nil {
		return PlainTime{}, err
	}
	if !p.hasTime {
		return PlainTime{}, fmt.Errorf("not a time: %q", s)
	}
	if err := p.plainZ(); err != nil {
		return PlainTime{}, err
	}
	if p.hasDate {
		if _, err := p.date(); err != nil {
			return PlainTime{}, err
		}
	}
	return p.clock()
}

// ParsePlainDateTime reads a date and a time.
func ParsePlainDateTime(s string) (PlainDateTime, error) {
	p, err := parseISO(s)
	if err != nil {
		return PlainDateTime{}, err
	}
	if !p.hasDate || !p.hasTime {
		return PlainDateTime{}, fmt.Errorf("a date-time needs a date and a time: %q", s)
	}
	if err := p.plainZ(); err != nil {
		return PlainDateTime{}, err
	}
	return p.dateTime()
}
