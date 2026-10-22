package temporal

import (
	"fmt"
	"math/big"
)

// Rounding as Temporal has it: a quantity is rounded to a multiple of an increment by a mode,
// and the sign-dependent modes are turned into unsigned ones for the magnitude.

type unsignedMode int

const (
	umZero unsignedMode = iota
	umInfinity
	umHalfZero
	umHalfInfinity
	umHalfEven
)

func unsignedModeOf(m RoundingMode, negative bool) unsignedMode {
	switch m {
	case Ceil:
		if negative {
			return umZero
		}
		return umInfinity
	case Floor:
		if negative {
			return umInfinity
		}
		return umZero
	case Expand:
		return umInfinity
	case Trunc:
		return umZero
	case HalfCeil:
		if negative {
			return umHalfZero
		}
		return umHalfInfinity
	case HalfFloor:
		if negative {
			return umHalfInfinity
		}
		return umHalfZero
	case HalfExpand:
		return umHalfInfinity
	case HalfTrunc:
		return umHalfZero
	}
	return umHalfEven
}

// applyUnsigned picks r1 or r2 for a magnitude between them; cmp compares twice the remainder
// with the increment.
func applyUnsigned(r1, r2 *big.Int, cmp int, even bool, m unsignedMode) *big.Int {
	switch m {
	case umZero:
		return r1
	case umInfinity:
		return r2
	case umHalfZero:
		if cmp > 0 {
			return r2
		}
		return r1
	case umHalfInfinity:
		if cmp < 0 {
			return r1
		}
		return r2
	}
	switch {
	case cmp < 0:
		return r1
	case cmp > 0:
		return r2
	case even:
		return r1
	}
	return r2
}

// roundBig rounds x to a multiple of inc by mode.
func roundBig(x, inc *big.Int, mode RoundingMode) *big.Int {
	q, r := new(big.Int).QuoRem(x, inc, new(big.Int))
	negative := x.Sign() < 0
	r1 := new(big.Int).Mul(new(big.Int).Abs(q), inc)
	if r.Sign() == 0 {
		if negative {
			r1.Neg(r1)
		}
		return r1
	}
	r2 := new(big.Int).Add(r1, inc)
	cmp := new(big.Int).Abs(new(big.Int).Mul(r, big.NewInt(2))).Cmp(inc)
	rounded := applyUnsigned(r1, r2, cmp, q.Bit(0) == 0, unsignedModeOf(mode, negative))
	if negative {
		rounded = new(big.Int).Neg(rounded)
	}
	return rounded
}

// roundInt rounds x to a multiple of inc by mode.
func roundInt(x, inc int64, mode RoundingMode) int64 {
	return roundBig(big.NewInt(x), big.NewInt(inc), mode).Int64()
}

// validateIncrement checks a rounding increment for a unit of a date-time: 1 for a day, else a
// divisor of the unit's count in the next larger unit, and smaller than it.
func validateIncrement(unit Unit, inc int64) error {
	if inc < 1 {
		return fmt.Errorf("the increment must be positive, not %d", inc)
	}
	var max int64
	switch unit {
	case Day:
		max = 1
	case Hour:
		max = 24
	case Minute, Second:
		max = 60
	case Millisecond, Microsecond, Nanosecond:
		max = 1000
	default:
		return fmt.Errorf("%ss cannot be rounded to", unit)
	}
	if unit == Day {
		if inc != 1 {
			return fmt.Errorf("the increment for days must be 1, not %d", inc)
		}
		return nil
	}
	if inc >= max || max%inc != 0 {
		return fmt.Errorf("the increment for %ss must divide %d and be smaller, not %d", unit, max, inc)
	}
	return nil
}
