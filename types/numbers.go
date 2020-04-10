package types

import (
	"math/big"
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

func MulBig64(a, b *big.Int) *big.Int {
	var bi big.Int
	bi.Mul(a, b)
	return &bi
}

func DivBig64(x, y *big.Int) *big.Int {
	var bi big.Int
	bi.Div(x, y)
	return &bi
}

func MulDivDec(a, b, c Dec) Dec {
	r, ok := Mul64(a.RawInt(), b.RawInt())
	if !ok {
		var bi big.Int
		bi.Div(bi.Mul(big.NewInt(a.RawInt()), big.NewInt(b.RawInt())), big.NewInt(c.RawInt()))
		if !bi.IsInt64() {
			panic("int64 overflow")
		}
		return NewDec(bi.Int64())
	}
	return NewDec(r / c.RawInt())
}

func MulDivDecWithExtraDecimal(a, b, c Dec, extraDecimalPlace int) (afterRoundDown int64, extraDecimalValue int) {
	extra := int64(Pow(10, extraDecimalPlace))
	product, ok := Mul64(a.RawInt(), b.RawInt())
	if !ok { // int64 exceed
		return mulDivBig64WithExtraDecimal(big.NewInt(a.RawInt()), big.NewInt(b.RawInt()), big.NewInt(c.RawInt()), big.NewInt(extra))
	} else {
		if product, ok = Mul64(product, extra); !ok {
			return mulDivBig64WithExtraDecimal(big.NewInt(a.RawInt()), big.NewInt(b.RawInt()), big.NewInt(c.RawInt()), big.NewInt(extra))
		}
		resultOfAddDecimalPlace := product / c.RawInt()
		afterRoundDown = resultOfAddDecimalPlace / extra
		extraDecimalValue = int(resultOfAddDecimalPlace % extra)
		return afterRoundDown, extraDecimalValue
	}
}

func mulDivBig64WithExtraDecimal(a, b, c, extra *big.Int) (afterRoundDown int64, extraDecimalValue int) {
	product := MulBig64(MulBig64(a, b), extra)
	result := DivBig64(product, c)

	expectedDecimalValueBig := &big.Int{}
	afterRoundDownBig, expectedDecimalValueBig := result.QuoRem(result, extra, expectedDecimalValueBig)
	afterRoundDown = afterRoundDownBig.Int64()
	extraDecimalValue = int(expectedDecimalValueBig.Int64())
	return afterRoundDown, extraDecimalValue
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
