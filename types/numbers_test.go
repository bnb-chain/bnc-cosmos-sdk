package types

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMulQuoDec(t *testing.T) {
	a := NewDecWithoutFra(2)
	b := NewDecWithoutFra(4)
	c := NewDecWithoutFra(70)
	r, err := MulQuoDec(a, b, c)
	require.Nil(t, err, fmt.Sprintf("expected nil error, but returns %s ",err))
	require.EqualValues(t, 11428571, r.RawInt())

	a = NewDecWithoutFra(20)
	b = NewDecWithoutFra(3)
	c = NewDecWithoutFra(15)
	r, err = MulQuoDec(a, b, c)
	require.Nil(t, err, fmt.Sprintf("expected nil error, but returns %s ",err))
	require.EqualValues(t, 4e8, r.RawInt())

	a = NewDecWithoutFra(20000000)
	b = NewDecWithoutFra(30000)
	c = NewDecWithoutFra(15)
	r, err = MulQuoDec(a, b, c)
	require.Nil(t, err, fmt.Sprintf("expected nil error, but returns %s ",err))
	require.EqualValues(t, 40000000000e8, r.RawInt())

	c = NewDec(15)
	r, err = MulQuoDec(a, b, c)
	require.NotNil(t, err)
	require.EqualError(t,err,ErrIntOverflow)

	c = ZeroDec()
	r, err = MulQuoDec(a, b, c)
	require.NotNil(t, err)
	require.EqualError(t,err,ErrZeroDividend)

}

func TestMulDivDecWithExtraDecimal(t *testing.T) {
	// 8/7 = 1.14285714,285714...
	a := NewDec(2e8)
	b := NewDec(4e8)
	c := NewDec(7e8)
	afterRoundDown, extraDecimalValue := MulQuoDecWithExtraDecimal(a, b, c, 1)
	require.EqualValues(t, 114285714, afterRoundDown)
	require.EqualValues(t, 2, extraDecimalValue)
	afterRoundDown, extraDecimalValue = MulQuoDecWithExtraDecimal(a, b, c, 6)
	require.EqualValues(t, 114285714, afterRoundDown)
	require.EqualValues(t, 285714, extraDecimalValue)
	// 800/7 = 114.28571428,5714...
	a = NewDec(20e8)
	b = NewDec(40e8)
	c = NewDec(7e8)
	afterRoundDown, extraDecimalValue = MulQuoDecWithExtraDecimal(a, b, c, 1)
	require.EqualValues(t, 11428571428, afterRoundDown)
	require.EqualValues(t, 5, extraDecimalValue)
	afterRoundDown, extraDecimalValue = MulQuoDecWithExtraDecimal(a, b, c, 4)
	require.EqualValues(t, 11428571428, afterRoundDown)
	require.EqualValues(t, 5714, extraDecimalValue)
	// 8000/7 = 1142.85714285,714...
	a = NewDec(200e8)
	b = NewDec(40e8)
	c = NewDec(7e8)
	afterRoundDown, extraDecimalValue = MulQuoDecWithExtraDecimal(a, b, c, 1)
	require.EqualValues(t, 114285714285, afterRoundDown)
	require.EqualValues(t, 7, extraDecimalValue)
	afterRoundDown, extraDecimalValue = MulQuoDecWithExtraDecimal(a, b, c, 3)
	require.EqualValues(t, 114285714285, afterRoundDown)
	require.EqualValues(t, 714, extraDecimalValue)
}
