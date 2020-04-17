package types

import (
	"errors"
	"math/big"
)

const (
	ErrZeroDividend = "Dividend is zero "
	ErrIntOverflow  = "Int Overflow "
)

func Mul64(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	if (c < 0) == ((a < 0) != (b < 0)) {
		if c/b == a {
			return c, true
		}
	}
	return c, false
}

func MulQuoDec(a, b, c Dec) (Dec, error) {
	if c.IsZero() {
		return Dec{}, errors.New(ErrZeroDividend)
	}
	r, ok := Mul64(a.RawInt(), b.RawInt())
	if !ok {
		var bi big.Int
		bi.Quo(bi.Mul(big.NewInt(a.RawInt()), big.NewInt(b.RawInt())), big.NewInt(c.RawInt()))
		if !bi.IsInt64() {
			return Dec{}, errors.New(ErrIntOverflow)
		}
		return NewDec(bi.Int64()), nil
	}
	return NewDec(r / c.RawInt()), nil
}

// calculate a * b / c, getting the extra decimal digital as result of extraDecimalValue. For example:
// 0.00000003 * 2 / 0.00000004 = 0.000000015,
// assume that decimal place number of Dec is 8, and the extraDecimalPlace was given 1, then
// we take the 8th decimal place value '1' as afterRoundDown, and extra decimal value(9th) '5' as extraDecimalValue
func MulQuoDecWithExtraDecimal(a, b, c Dec, extraDecimalPlace int) (afterRoundDown int64, extraDecimalValue int) {
	extra := int64(Pow(10, extraDecimalPlace))
	product, ok := Mul64(a.RawInt(), b.RawInt())
	if !ok { // int64 exceed
		return mulQuoBig64WithExtraDecimal(big.NewInt(a.RawInt()), big.NewInt(b.RawInt()), big.NewInt(c.RawInt()), big.NewInt(extra))
	} else {
		if product, ok = Mul64(product, extra); !ok {
			return mulQuoBig64WithExtraDecimal(big.NewInt(a.RawInt()), big.NewInt(b.RawInt()), big.NewInt(c.RawInt()), big.NewInt(extra))
		}
		resultOfAddDecimalPlace := product / c.RawInt()
		afterRoundDown = resultOfAddDecimalPlace / extra
		extraDecimalValue = int(resultOfAddDecimalPlace % extra)
		return afterRoundDown, extraDecimalValue
	}
}

func mulQuoBig64WithExtraDecimal(a, b, c, extra *big.Int) (afterRoundDown int64, extraDecimalValue int) {
	product := MulBigInt64(MulBigInt64(a, b), extra)
	result := QuoBigInt64(product, c)

	expectedDecimalValueBig := &big.Int{}
	afterRoundDownBig, expectedDecimalValueBig := result.QuoRem(result, extra, expectedDecimalValueBig)
	afterRoundDown = afterRoundDownBig.Int64()
	extraDecimalValue = int(expectedDecimalValueBig.Int64())
	return afterRoundDown, extraDecimalValue
}

func MulBigInt64(a, b *big.Int) *big.Int {
	var bi big.Int
	bi.Mul(a, b)
	return &bi
}

func QuoBigInt64(x, y *big.Int) *big.Int {
	var bi big.Int
	bi.Quo(x, y)
	return &bi
}

func Pow(x, n int) int {
	ret := 1
	for n != 0 {
		if n%2 != 0 {
			ret = ret * x
		}
		n /= 2
		x = x * x
	}
	return ret
}
