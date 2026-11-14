package temporal

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// The short notation of durations: the duration literal of the language (2h30m, 1.5s), and Go's
// time.ParseDuration strings, which Time.parseDuration takes besides ISO 8601. The two share the
// units and the spreading of a fraction, not the grammar.

// shortUnit is the unit a short notation suffix names; the literal takes only ASCII suffixes.
func shortUnit(suffix string, micro bool) (Unit, bool) {
	switch suffix {
	case "h":
		return Hour, true
	case "m":
		return Minute, true
	case "s":
		return Second, true
	case "ms":
		return Millisecond, true
	case "us":
		return Microsecond, true
	case "ns":
		return Nanosecond, true
	case "µs", "μs": // U+00B5 and U+03BC, as Go takes them
		return Microsecond, micro
	}
	return 0, false
}

var errFieldRange = errors.New("a field of the duration does not fit an int")

// addTo adds v to field u, failing where the field would leave int64.
func addTo(f *[10]int64, u Unit, v int64) error {
	s := f[u] + v
	if v > 0 && s < f[u] {
		return errFieldRange
	}
	f[u] = s
	return nil
}

// addComponent adds whole units of u and a decimal fraction of one to f, the fraction spread into
// the smaller fields, largest first, as ISO 8601's parser spreads one. With exact, a fraction that
// is not a whole number of nanoseconds is an error; otherwise what is below a nanosecond is
// dropped. whole and frac are ASCII digits; either may be empty.
func addComponent(f *[10]int64, u Unit, whole, frac string, exact bool) error {
	if whole != "" {
		v, err := strconv.ParseInt(whole, 10, 64)
		if err != nil {
			return errFieldRange
		}
		if err := addTo(f, u, v); err != nil {
			return err
		}
	}
	if strings.Trim(frac, "0") == "" {
		return nil
	}
	num, _ := new(big.Int).SetString(frac, 10)
	den := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(len(frac))), nil)
	ns, rem := new(big.Int).QuoRem(num.Mul(num, big.NewInt(nsPerUnit[u])), den, new(big.Int))
	if exact && rem.Sign() != 0 {
		return errors.New("not a whole number of nanoseconds")
	}
	// ns is below one unit of u, so it fits an int64
	rest := ns.Int64()
	for v := u + 1; v <= Nanosecond; v++ {
		if err := addTo(f, v, rest/nsPerUnit[v]); err != nil {
			return err
		}
		rest %= nsPerUnit[v]
	}
	return nil
}

// DurationLiteral is the value of a duration literal as written in source, underscores included:
// one or more components of a decimal number and a unit of h, m, s, ms, us or ns, the units each
// once and in descending order, a fraction only in the last component, and a fraction that comes
// out as a whole number of nanoseconds. The error says what is wrong, for a syntax error.
func DurationLiteral(s string) (Duration, error) {
	var f [10]int64
	next := Unit(Hour) // the largest unit still allowed
	fraction := false
	for i := 0; i < len(s); {
		if fraction {
			return Duration{}, errors.New("only the last component of a duration literal can have a fraction")
		}
		start := i
		for i < len(s) && (isDigit(s[i]) || s[i] == '_' || s[i] == '.') {
			i++
		}
		number := s[start:i]
		us := i
		for i < len(s) && !isDigit(s[i]) && s[i] != '_' && s[i] != '.' {
			i++
		}
		suffix := s[us:i]
		whole, frac, hasPoint := strings.Cut(number, ".")
		if !decimalSpelling(whole) || hasPoint && !digitRun(frac) {
			return Duration{}, fmt.Errorf("malformed number %s in a duration literal", number)
		}
		if suffix == "" {
			return Duration{}, fmt.Errorf("the number %s in a duration literal needs a unit", number)
		}
		u, ok := shortUnit(suffix, false)
		if !ok {
			switch suffix {
			case "d", "w", "y", "mo":
				return Duration{}, errors.New("there are no duration literals for days, weeks, months or years: write Time.duration(days: 1) and the like")
			}
			return Duration{}, fmt.Errorf("malformed number literal: %s is not a duration unit, which are h, m, s, ms, us and ns", suffix)
		}
		if u < next {
			return Duration{}, errors.New("the units of a duration literal come in descending order, each at most once")
		}
		next = u + 1
		fraction = hasPoint
		frac = strings.ReplaceAll(frac, "_", "")
		if err := addComponent(&f, u, strings.ReplaceAll(whole, "_", ""), frac, true); err != nil {
			if err == errFieldRange {
				return Duration{}, errors.New("duration literal out of range")
			}
			return Duration{}, fmt.Errorf("duration literal %s is %s", s, err)
		}
	}
	d, err := NewDuration(f[0], f[1], f[2], f[3], f[4], f[5], f[6], f[7], f[8], f[9])
	if err != nil {
		return Duration{}, fmt.Errorf("duration literal out of range: %s", err)
	}
	return d, nil
}

// digitRun is digit { ["_"] digit }.
func digitRun(s string) bool {
	if s == "" || !isDigit(s[0]) || !isDigit(s[len(s)-1]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isDigit(s[i]) && !(s[i] == '_' && isDigit(s[i-1]) && isDigit(s[i+1])) {
			return false
		}
	}
	return true
}

// decimalSpelling is an Ant decimal integer: a digit run that starts with 0 only when it is 0.
func decimalSpelling(s string) bool {
	return digitRun(s) && (s[0] != '0' || len(s) == 1)
}

// parseGoDuration reads Go's time.ParseDuration notation: an optional sign, then "0" or one or
// more components of a decimal number and a unit, in any order and repeating, any of them with a
// fraction. Each component adds to its own field.
func parseGoDuration(s string) (Duration, error) {
	bad := func() (Duration, error) { return Duration{}, fmt.Errorf("not a duration: %q", s) }
	rest := s
	neg := false
	if rest != "" && (rest[0] == '-' || rest[0] == '+') {
		neg = rest[0] == '-'
		rest = rest[1:]
	}
	if rest == "0" {
		return Duration{}, nil
	}
	if rest == "" {
		return bad()
	}
	var f [10]int64
	for rest != "" {
		i := 0
		for i < len(rest) && isDigit(rest[i]) {
			i++
		}
		whole := rest[:i]
		rest = rest[i:]
		var frac string
		if rest != "" && rest[0] == '.' {
			j := 1
			for j < len(rest) && isDigit(rest[j]) {
				j++
			}
			frac = rest[1:j]
			rest = rest[j:]
		}
		if whole == "" && frac == "" {
			return bad()
		}
		k := 0
		for k < len(rest) && rest[k] != '.' && !isDigit(rest[k]) {
			k++
		}
		u, ok := shortUnit(rest[:k], true)
		if !ok {
			return bad()
		}
		rest = rest[k:]
		if err := addComponent(&f, u, whole, frac, false); err != nil {
			return Duration{}, fmt.Errorf("duration out of range: %q", s)
		}
	}
	if neg {
		for i := range f {
			f[i] = -f[i]
		}
	}
	return NewDuration(f[0], f[1], f[2], f[3], f[4], f[5], f[6], f[7], f[8], f[9])
}
