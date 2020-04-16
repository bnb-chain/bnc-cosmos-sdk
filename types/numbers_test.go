package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMulDivDec(t *testing.T) {
	a := NewDec(2000000000)
	b := NewDec(300000000)
	c := NewDec(1500000000)
	r, ok := MulQuoDec(a, b, c)
	require.True(t, ok)
	require.EqualValues(t, 400000000, r.RawInt())

	a = NewDec(2000000000000000)
	b = NewDec(3000000000000)
	c = NewDec(1500000000)
	r, ok = MulQuoDec(a, b, c)
	require.True(t, ok)
	require.EqualValues(t, 4000000000000000000, r.RawInt())
}
