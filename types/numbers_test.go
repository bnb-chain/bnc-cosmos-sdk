package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMulDivDec(t *testing.T) {
	a := NewDec(2000000000)
	b := NewDec(300000000)
	c := NewDec(1500000000)
	r, ok := MulDivDec(a, b, c)
	require.True(t, ok)
	require.EqualValues(t, 400000000, r.RawInt())

	a = NewDec(2000000000000000)
	b = NewDec(3000000000000)
	c = NewDec(1500000000)
	r, ok = MulDivDec(a, b, c)
	require.True(t, ok)
	require.EqualValues(t, 4000000000000000000, r.RawInt())
}

func TestMulDivDecWithExtraDecimal(t *testing.T) {
	// 8/7 = 1.14285714,285714...
	a := NewDec(2e8)
	b := NewDec(4e8)
	c := NewDec(7e8)
	afterRoundDown, extraDecimalValue := MulDivDecWithExtraDecimal(a, b, c, 1)
	require.EqualValues(t, 114285714, afterRoundDown)
	require.EqualValues(t, 2, extraDecimalValue)
	afterRoundDown, extraDecimalValue = MulDivDecWithExtraDecimal(a, b, c, 6)
	require.EqualValues(t, 114285714, afterRoundDown)
	require.EqualValues(t, 285714, extraDecimalValue)
	// 800/7 = 114.28571428,5714...
	a = NewDec(20e8)
	b = NewDec(40e8)
	c = NewDec(7e8)
	afterRoundDown, extraDecimalValue = MulDivDecWithExtraDecimal(a, b, c, 1)
	require.EqualValues(t, 11428571428, afterRoundDown)
	require.EqualValues(t, 5, extraDecimalValue)
	afterRoundDown, extraDecimalValue = MulDivDecWithExtraDecimal(a, b, c, 4)
	require.EqualValues(t, 11428571428, afterRoundDown)
	require.EqualValues(t, 5714, extraDecimalValue)
	// 8000/7 = 1142.85714285,714...
	a = NewDec(200e8)
	b = NewDec(40e8)
	c = NewDec(7e8)
	afterRoundDown, extraDecimalValue = MulDivDecWithExtraDecimal(a, b, c, 1)
	require.EqualValues(t, 114285714285, afterRoundDown)
	require.EqualValues(t, 7, extraDecimalValue)
	afterRoundDown, extraDecimalValue = MulDivDecWithExtraDecimal(a, b, c, 3)
	require.EqualValues(t, 114285714285, afterRoundDown)
	require.EqualValues(t, 714, extraDecimalValue)
}
