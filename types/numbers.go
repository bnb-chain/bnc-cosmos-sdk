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

func MulQuoDec(a, b, c Dec) (Dec, bool) {
	r, ok := Mul64(a.RawInt(), b.RawInt())
	if !ok {
		var bi big.Int
		bi.Quo(bi.Mul(big.NewInt(a.RawInt()), big.NewInt(b.RawInt())), big.NewInt(c.RawInt()))
		if !bi.IsInt64() {
			return Dec{}, false
		}
		return NewDec(bi.Int64()), true
	}
	return NewDec(r / c.RawInt()), true
}