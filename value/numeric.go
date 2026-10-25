package value

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
)

// Arithmetic and comparison follow section 12: both operands must be numbers of one kind, or
// two strings for + and the comparisons, or two tagged values of one tag, which are unwrapped
// and rewrapped. Integer results that do not fit their kind are errors, as are zero divisors.

var errOverflow = errors.New("integer overflow")
var errDivZero = errors.New("division by zero")

// Binary applies an arithmetic operator: + - * / %.
func Binary(op string, a, b Value) (Value, error) {
	if ta, ok := a.(*Tagged); ok {
		tb, ok := b.(*Tagged)
		if !ok || ta.Tag != tb.Tag {
			return nil, operandError(op, a, b)
		}
		v, err := Binary(op, ta.Payload, tb.Payload)
		if err != nil {
			return nil, err
		}
		return &Tagged{Tag: ta.Tag, Payload: v}, nil
	}
	if _, ok := b.(*Tagged); ok {
		return nil, operandError(op, a, b)
	}
	switch x := a.(type) {
	case string:
		if op != "+" {
			return nil, operandError(op, a, b)
		}
		return binary(op, x, a, b, func(_ string, x, y string) (Value, error) { return x + y, nil })
	case int64:
		return binary(op, x, a, b, int64Op)
	case float64:
		return binary(op, x, a, b, floatOp[float64])
	case float32:
		return binary(op, x, a, b, floatOp[float32])
	case int8:
		return binary(op, x, a, b, signedOp[int8])
	case int16:
		return binary(op, x, a, b, signedOp[int16])
	case int32:
		return binary(op, x, a, b, signedOp[int32])
	case uint8:
		return binary(op, x, a, b, unsignedOp[uint8])
	case uint16:
		return binary(op, x, a, b, unsignedOp[uint16])
	case uint32:
		return binary(op, x, a, b, unsignedOp[uint32])
	case uint64:
		return binary(op, x, a, b, uint64Op)
	}
	return nil, operandError(op, a, b)
}

// binary applies f to x, a's payload, and b's, which must have x's kind.
func binary[T Value](op string, x T, a, b Value, f func(string, T, T) (Value, error)) (Value, error) {
	y, ok := b.(T)
	if !ok {
		return nil, operandError(op, a, b)
	}
	return f(op, x, y)
}

func operandError(op string, a, b Value) error {
	return fmt.Errorf("%s needs two numbers of one kind%s, got %s and %s", op, stringsToo(op), Describe(a), Describe(b))
}

func stringsToo(op string) string {
	switch op {
	case "+", "<", "<=", ">", ">=":
		return " or two strings"
	}
	return ""
}

// Describe names a value's kind, with its tag for a tagged value or a record.
func Describe(v Value) string {
	switch x := v.(type) {
	case *Tagged:
		return x.Tag.Name
	case *Record:
		return x.Tag.Name
	case Time:
		return x.TimeType()
	case *Regex:
		return "Regex.Pattern"
	}
	return KindOf(v).String()
}

func floatOp[T float32 | float64](op string, x, y T) (Value, error) {
	switch op {
	case "+":
		return x + y, nil
	case "-":
		return x - y, nil
	case "*":
		return x * y, nil
	case "/":
		return x / y, nil
	case "%":
		return T(math.Mod(float64(x), float64(y))), nil
	}
	panic("unknown operator " + op)
}

// signedOp computes in int64 and checks that the result fits T, for the kinds narrower than 64
// bits.
func signedOp[T int8 | int16 | int32](op string, a, b T) (Value, error) {
	x, y := int64(a), int64(b)
	var r int64
	switch op {
	case "+":
		r = x + y
	case "-":
		r = x - y
	case "*":
		r = x * y
	case "/":
		if y == 0 {
			return nil, errDivZero
		}
		r = x / y
	case "%":
		if y == 0 {
			return nil, errDivZero
		}
		r = x % y
	}
	if int64(T(r)) != r {
		return nil, errOverflow
	}
	return T(r), nil
}

// unsignedOp computes in uint64 and checks that the result fits T, for the kinds narrower than
// 64 bits.
func unsignedOp[T uint8 | uint16 | uint32](op string, a, b T) (Value, error) {
	x, y := uint64(a), uint64(b)
	var r uint64
	switch op {
	case "+":
		r = x + y
	case "-":
		if y > x {
			return nil, errOverflow
		}
		r = x - y
	case "*":
		r = x * y
	case "/":
		if y == 0 {
			return nil, errDivZero
		}
		r = x / y
	case "%":
		if y == 0 {
			return nil, errDivZero
		}
		r = x % y
	}
	if uint64(T(r)) != r {
		return nil, errOverflow
	}
	return T(r), nil
}

func int64Op(op string, x, y int64) (Value, error) {
	switch op {
	case "+":
		r := x + y
		if (r > x) != (y > 0) {
			return nil, errOverflow
		}
		return r, nil
	case "-":
		r := x - y
		if (r < x) != (y > 0) {
			return nil, errOverflow
		}
		return r, nil
	case "*":
		if x == 0 || y == 0 {
			return int64(0), nil
		}
		r := x * y
		if r/y != x || (x == -1 && y == math.MinInt64) || (y == -1 && x == math.MinInt64) {
			return nil, errOverflow
		}
		return r, nil
	case "/":
		if y == 0 {
			return nil, errDivZero
		}
		if x == math.MinInt64 && y == -1 {
			return nil, errOverflow
		}
		return x / y, nil
	case "%":
		if y == 0 {
			return nil, errDivZero
		}
		if y == -1 {
			return int64(0), nil
		}
		return x % y, nil
	}
	panic("unknown operator " + op)
}

func uint64Op(op string, x, y uint64) (Value, error) {
	switch op {
	case "+":
		r, carry := bits.Add64(x, y, 0)
		if carry != 0 {
			return nil, errOverflow
		}
		return r, nil
	case "-":
		if y > x {
			return nil, errOverflow
		}
		return x - y, nil
	case "*":
		hi, lo := bits.Mul64(x, y)
		if hi != 0 {
			return nil, errOverflow
		}
		return lo, nil
	case "/":
		if y == 0 {
			return nil, errDivZero
		}
		return x / y, nil
	case "%":
		if y == 0 {
			return nil, errDivZero
		}
		return x % y, nil
	}
	panic("unknown operator " + op)
}

// Neg is unary minus.
func Neg(a Value) (Value, error) {
	if t, ok := a.(*Tagged); ok {
		v, err := Neg(t.Payload)
		if err != nil {
			return nil, err
		}
		return &Tagged{Tag: t.Tag, Payload: v}, nil
	}
	switch x := a.(type) {
	case int64:
		if x == math.MinInt64 {
			return nil, errOverflow
		}
		return -x, nil
	case float64:
		return -x, nil
	case float32:
		return -x, nil
	case int8:
		if x == math.MinInt8 {
			return nil, errOverflow
		}
		return -x, nil
	case int16:
		if x == math.MinInt16 {
			return nil, errOverflow
		}
		return -x, nil
	case int32:
		if x == math.MinInt32 {
			return nil, errOverflow
		}
		return -x, nil
	case uint8, uint16, uint32, uint64:
		if isZero(a) {
			return a, nil
		}
		return nil, errOverflow
	}
	return nil, fmt.Errorf("unary - needs a number, got %s", Describe(a))
}

func isZero(v Value) bool {
	switch x := v.(type) {
	case uint8:
		return x == 0
	case uint16:
		return x == 0
	case uint32:
		return x == 0
	case uint64:
		return x == 0
	}
	return false
}

// Compare applies < <= > >=.
func Compare(op string, a, b Value) (bool, error) {
	if ta, ok := a.(*Tagged); ok {
		tb, ok := b.(*Tagged)
		if !ok || ta.Tag != tb.Tag {
			return false, operandError(op, a, b)
		}
		return Compare(op, ta.Payload, tb.Payload)
	}
	if _, ok := b.(*Tagged); ok {
		return false, operandError(op, a, b)
	}
	var c int // -1, 0, 1; 2 for unordered floats
	var err error
	switch x := a.(type) {
	case string:
		c, err = order(op, x, a, b, compareStrings)
	case int64:
		c, err = order(op, x, a, b, cmp[int64])
	case float64:
		c, err = order(op, x, a, b, cmpFloat[float64])
	case float32:
		c, err = order(op, x, a, b, cmpFloat[float32])
	case int8:
		c, err = order(op, x, a, b, cmp[int8])
	case int16:
		c, err = order(op, x, a, b, cmp[int16])
	case int32:
		c, err = order(op, x, a, b, cmp[int32])
	case uint8:
		c, err = order(op, x, a, b, cmp[uint8])
	case uint16:
		c, err = order(op, x, a, b, cmp[uint16])
	case uint32:
		c, err = order(op, x, a, b, cmp[uint32])
	case uint64:
		c, err = order(op, x, a, b, cmp[uint64])
	default:
		return false, operandError(op, a, b)
	}
	if err != nil {
		return false, err
	}
	if c == 2 {
		return false, nil
	}
	switch op {
	case "<":
		return c < 0, nil
	case "<=":
		return c <= 0, nil
	case ">":
		return c > 0, nil
	case ">=":
		return c >= 0, nil
	}
	panic("unknown operator " + op)
}

func cmp[T int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64](x, y T) int {
	switch {
	case x < y:
		return -1
	case x > y:
		return 1
	}
	return 0
}

// order compares x, a's payload, with b's, which must have x's kind.
func order[T Value](op string, x T, a, b Value, f func(T, T) int) (int, error) {
	y, ok := b.(T)
	if !ok {
		return 0, operandError(op, a, b)
	}
	return f(x, y), nil
}

func cmpFloat[T float32 | float64](x, y T) int {
	switch {
	case x < y:
		return -1
	case x > y:
		return 1
	case x == y:
		return 0
	}
	return 2
}

// compareStrings orders by code point, which for valid UTF-8 is byte order.
func compareStrings(x, y string) int {
	switch {
	case x < y:
		return -1
	case x > y:
		return 1
	}
	return 0
}

// ---- Conversions

// Convert applies one of the twelve conversion functions (section 12). A tagged argument is
// unwrapped first.
func Convert(to Kind, v Value) (Value, error) {
	if t, ok := v.(*Tagged); ok {
		v = t.Payload
	}
	switch to {
	case KindString:
		if s, ok := v.(string); ok {
			return s, nil
		}
		return nil, fmt.Errorf("string needs a string, got %s", Describe(v))
	case KindBool:
		if b, ok := v.(bool); ok {
			return b, nil
		}
		return nil, fmt.Errorf("bool needs a bool, got %s", Describe(v))
	}
	k := KindOf(v)
	if !k.IsNumber() {
		return nil, fmt.Errorf("%s needs a number, got %s", to, Describe(v))
	}
	if k.IsFloat() {
		f := toFloat64(v)
		if to.IsFloat() {
			if to == KindFloat32 {
				return float32(f), nil
			}
			return f, nil
		}
		return floatToInt(to, f)
	}
	// integer source
	if to.IsFloat() {
		if to == KindFloat32 {
			if k.IsSigned() {
				return float32(toInt64(v)), nil
			}
			return float32(toUint64(v)), nil
		}
		if k.IsSigned() {
			return float64(toInt64(v)), nil
		}
		return float64(toUint64(v)), nil
	}
	if k.IsSigned() {
		return intToInt(to, toInt64(v))
	}
	return uintToInt(to, toUint64(v))
}

func toFloat64(v Value) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	}
	panic("not a float")
}

func toInt64(v Value) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	}
	panic("not a signed integer")
}

func toUint64(v Value) uint64 {
	switch x := v.(type) {
	case uint64:
		return x
	case uint8:
		return uint64(x)
	case uint16:
		return uint64(x)
	case uint32:
		return uint64(x)
	}
	panic("not an unsigned integer")
}

func rangeError(to Kind, v Value) error {
	return fmt.Errorf("%s out of range for %s", Render(v), to)
}

func intToInt(to Kind, x int64) (Value, error) {
	switch to {
	case KindInt64:
		return x, nil
	case KindInt8:
		if x < math.MinInt8 || x > math.MaxInt8 {
			return nil, rangeError(to, x)
		}
		return int8(x), nil
	case KindInt16:
		if x < math.MinInt16 || x > math.MaxInt16 {
			return nil, rangeError(to, x)
		}
		return int16(x), nil
	case KindInt32:
		if x < math.MinInt32 || x > math.MaxInt32 {
			return nil, rangeError(to, x)
		}
		return int32(x), nil
	}
	if x < 0 {
		return nil, rangeError(to, x)
	}
	return uintToInt(to, uint64(x))
}

func uintToInt(to Kind, x uint64) (Value, error) {
	switch to {
	case KindUint64:
		return x, nil
	case KindUint8:
		if x > math.MaxUint8 {
			return nil, rangeError(to, x)
		}
		return uint8(x), nil
	case KindUint16:
		if x > math.MaxUint16 {
			return nil, rangeError(to, x)
		}
		return uint16(x), nil
	case KindUint32:
		if x > math.MaxUint32 {
			return nil, rangeError(to, x)
		}
		return uint32(x), nil
	case KindInt64:
		if x > math.MaxInt64 {
			return nil, rangeError(to, x)
		}
		return int64(x), nil
	}
	if x > math.MaxInt64 {
		return nil, rangeError(to, x)
	}
	return intToInt(to, int64(x))
}

// floatToInt requires an integral value in range; NaN and infinities are out of range, and a
// fraction is reported as one.
func floatToInt(to Kind, f float64) (Value, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil, rangeError(to, f)
	}
	if f != math.Trunc(f) {
		return nil, fmt.Errorf("%s has a fraction, so it does not convert to %s", Render(f), to)
	}
	// Bounds as floats: lo inclusive, hi exclusive, both exactly representable.
	var lo, hi float64
	switch to {
	case KindInt8:
		lo, hi = -128, 128
	case KindInt16:
		lo, hi = -32768, 32768
	case KindInt32:
		lo, hi = -2147483648, 2147483648
	case KindInt64:
		lo, hi = -9223372036854775808, 9223372036854775808
	case KindUint8:
		lo, hi = 0, 256
	case KindUint16:
		lo, hi = 0, 65536
	case KindUint32:
		lo, hi = 0, 4294967296
	case KindUint64:
		lo, hi = 0, 18446744073709551616
	}
	if f < lo || f >= hi {
		return nil, rangeError(to, f)
	}
	if to.IsSigned() {
		return intToInt(to, int64(f))
	}
	return uintToInt(to, uint64(f))
}
