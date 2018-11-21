package types

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
)

// NOTE: never use new(Dec) or else we will panic unmarshalling into the
// nil embedded big.Int
type Dec struct {
	int64 `json:"int"`
}

// number of decimal places
const (
	Precision = 8

	// bytes required to represent the above precision
	// ceil(log2(9999999999))
	DecimalPrecisionBits = 34
)

var (
	precisionReuse       = new(big.Int).Exp(big.NewInt(10), big.NewInt(Precision), nil).Int64()
	fivePrecision        = precisionReuse / 2
	precisionMultipliers []int64
	zeroInt              = big.NewInt(0)
	oneInt               = big.NewInt(1)
	tenInt               = big.NewInt(10)
)

// Set precision multipliers
func init() {
	precisionMultipliers = make([]int64, Precision+1)
	for i := 0; i <= Precision; i++ {
		precisionMultipliers[i] = calcPrecisionMultiplier(int64(i))
	}
}

func precisionInt() int64 {
	return precisionReuse
}

// nolint - common values
func ZeroDec() Dec { return Dec{0} }
func OneDec() Dec  { return Dec{precisionInt()} }

// calculate the precision multiplier
func calcPrecisionMultiplier(prec int64) int64 {
	if prec > Precision {
		panic(fmt.Sprintf("too much precision, maximum %v, provided %v", Precision, prec))
	}
	zerosToAdd := Precision - prec
	multiplier := new(big.Int).Exp(tenInt, big.NewInt(zerosToAdd), nil).Int64()
	return multiplier
}

// get the precision multiplier, do not mutate result
func precisionMultiplier(prec int64) int64 {
	if prec > Precision {
		panic(fmt.Sprintf("too much precision, maximum %v, provided %v", Precision, prec))
	}
	return precisionMultipliers[prec]
}

//______________________________________________________________________________________________

// create a new Dec from integer assuming whole number
func NewDec(i int64) Dec {
	return NewDecWithPrec(i, 0)
}

// create a new Dec from integer with decimal place at prec
// CONTRACT: prec <= Precision
func NewDecWithPrec(i, prec int64) Dec {
	if i == 0 {
		return Dec{0}
	}
	c := i * precisionMultiplier(prec)
	if c/i != precisionMultiplier(prec) {
		panic("Int overflow")
	}
	return Dec{c}
}

// create a new Dec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewDecFromBigInt(i int64) Dec {
	return NewDecFromBigIntWithPrec(i, 0)
}

// create a new Dec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewDecFromBigIntWithPrec(i int64, prec int64) Dec {
	return NewDecWithPrec(i, prec)
}

// create a new Dec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewDecFromInt(i int64) Dec {
	return NewDecFromIntWithPrec(i, 0)
}

// create a new Dec from big integer with decimal place at prec
// CONTRACT: prec <= Precision
func NewDecFromIntWithPrec(i int64, prec int64) Dec {
	return NewDecWithPrec(i, prec)
}

// create a decimal from an input decimal string.
// valid must come in the form:
//   (-) whole integers (.) decimal integers
// examples of acceptable input include:
//   -123.456
//   456.7890
//   345
//   -456789
//
// NOTE - An error will return if more decimal places
// are provided in the string than the constant Precision.
//
// CONTRACT - This function does not mutate the input str.
func NewDecFromStr(str string) (d Dec, err Error) {
	if len(str) == 0 {
		return d, ErrUnknownRequest("decimal string is empty")
	}

	// first extract any negative symbol
	neg := false
	if str[0] == '-' {
		neg = true
		str = str[1:]
	}

	if len(str) == 0 {
		return d, ErrUnknownRequest("decimal string is empty")
	}

	strs := strings.Split(str, ".")
	lenDecs := 0
	combinedStr := strs[0]
	if len(strs) == 2 {
		lenDecs = len(strs[1])
		if lenDecs == 0 || len(combinedStr) == 0 {
			return d, ErrUnknownRequest("bad decimal length")
		}
		combinedStr = combinedStr + strs[1]
	} else if len(strs) > 2 {
		return d, ErrUnknownRequest("too many periods to be a decimal string")
	}

	if lenDecs > Precision {
		return d, ErrUnknownRequest(
			fmt.Sprintf("too much precision, maximum %v, len decimal %v", Precision, lenDecs))
	}

	// add some extra zero's to correct to the Precision factor
	zerosToAdd := Precision - lenDecs
	zeros := fmt.Sprintf(`%0`+strconv.Itoa(zerosToAdd)+`s`, "")
	combinedStr = combinedStr + zeros

	combined, parseErr := strconv.ParseInt(combinedStr, 10, 64)
	if parseErr != nil {
		return d, ErrUnknownRequest(fmt.Sprintf("bad string to integer conversion, combinedStr: %v, error: %v", combinedStr, err))
	}
	if neg {
		combined = -combined
	}
	return Dec{combined}, nil
}

//______________________________________________________________________________________________
//nolint
func (d Dec) IsNil() bool       { return false }               // is decimal nil
func (d Dec) IsZero() bool      { return d.int64 == 0 }        // is equal to zero
func (d Dec) Equal(d2 Dec) bool { return d.int64 == d2.int64 } // equal decimals
func (d Dec) GT(d2 Dec) bool    { return d.int64 > d2.int64 }  // greater than
func (d Dec) GTE(d2 Dec) bool   { return d.int64 >= d2.int64 } // greater than or equal
func (d Dec) LT(d2 Dec) bool    { return d.int64 < d2.int64 }  // less than
func (d Dec) LTE(d2 Dec) bool   { return d.int64 <= d2.int64 } // less than or equal
func (d Dec) Neg() Dec          { return Dec{-d.int64} }       // reverse the decimal sign
func (d Dec) Abs() Dec {
	if d.int64 < 0 {
		return d.Neg()
	}
	return d
}

func (d Dec) Set(v int64) Dec {
	d.int64 = v
	return d
}

// addition
func (d Dec) Add(d2 Dec) Dec {
	c := d.int64 + d2.int64
	if (c > d.int64) != (d2.int64 > 0) {
		panic("Int overflow")
	}
	return Dec{c}
}

// subtraction
func (d Dec) Sub(d2 Dec) Dec {
	c := d.int64 - d2.int64
	if (c < d.int64) != (d2.int64 > 0) {
		panic("Int overflow")
	}
	return Dec{c}
}

// multiplication
func (d Dec) Mul(d2 Dec) Dec {
	mul := new(big.Int).Mul(big.NewInt(d.int64), big.NewInt(d2.int64))
	chopped := chopPrecisionAndRound(mul)

	if !chopped.IsInt64() {
		panic("Int overflow")
	}
	return Dec{chopped.Int64()}
}

// multiplication
func (d Dec) MulInt(i int64) Dec {
	mul := new(big.Int).Mul(big.NewInt(d.int64), big.NewInt(i))

	if !mul.IsInt64() {
		panic("Int overflow")
	}
	return Dec{mul.Int64()}
}

// quotient
func (d Dec) Quo(d2 Dec) Dec {
	if d2.IsZero() {
		panic("Dived can not be zero")
	}
	// multiply precision twice
	mul := new(big.Int).Mul(big.NewInt(d.int64), big.NewInt(precisionReuse))
	mul.Mul(mul, big.NewInt(precisionReuse))

	quo := new(big.Int).Quo(mul, big.NewInt(d2.int64))
	chopped := chopPrecisionAndRound(quo)

	if !chopped.IsInt64() {
		panic("Int overflow")
	}
	return Dec{chopped.Int64()}
}

// quotient
func (d Dec) QuoInt(i int64) Dec {
	mul := d.int64 / i
	return Dec{mul}
}

// is integer, e.g. decimals are zero
func (d Dec) IsInteger() bool {
	return d.int64%precisionReuse == 0
}

func (d Dec) String() string {
	s := strconv.FormatInt(d.int64, 10)
	bz := []byte(s)
	var bzWDec []byte
	inputSize := len(bz)
	// TODO: Remove trailing zeros
	// case 1, purely decimal
	if inputSize <= 8 {
		bzWDec = make([]byte, 10)
		// 0. prefix
		bzWDec[0] = byte('0')
		bzWDec[1] = byte('.')
		// set relevant digits to 0
		for i := 0; i < 8-inputSize; i++ {
			bzWDec[i+2] = byte('0')
		}
		// set last few digits
		copy(bzWDec[2+(8-inputSize):], bz)
	} else {
		// inputSize + 1 to account for the decimal point that is being added
		bzWDec = make([]byte, inputSize+1)
		copy(bzWDec, bz[:inputSize-8])
		bzWDec[inputSize-8] = byte('.')
		copy(bzWDec[inputSize-7:], bz[inputSize-8:])
	}
	return string(bzWDec)
}

//     ____
//  __|    |__   "chop 'em
//       ` \     round!"
// ___||  ~  _     -bankers
// |         |      __
// |       | |   __|__|__
// |_____:  /   | $$$    |
//              |________|

// nolint - go-cyclo
// Remove a Precision amount of rightmost digits and perform bankers rounding
// on the remainder (gaussian rounding) on the digits which have been removed.
//
// Mutates the input. Use the non-mutative version if that is undesired
func chopPrecisionAndRound(d *big.Int) *big.Int {

	// remove the negative and add it back when returning
	if d.Sign() == -1 {
		// make d positive, compute chopped value, and then un-mutate d
		d = d.Neg(d)
		d = chopPrecisionAndRound(d)
		d = d.Neg(d)
		return d
	}

	// get the trucated quotient and remainder
	quo, rem := d, big.NewInt(0)
	quo, rem = quo.QuoRem(d, big.NewInt(precisionReuse), rem)

	if rem.Sign() == 0 { // remainder is zero
		return quo
	}

	switch rem.Cmp(big.NewInt(fivePrecision)) {
	case -1:
		return quo
	case 1:
		return quo.Add(quo, oneInt)
	default: // bankers rounding must take place
		// always round to an even number
		if quo.Bit(0) == 0 {
			return quo
		}
		return quo.Add(quo, oneInt)
	}
}

func chopPrecisionAndRoundNonMutative(d *big.Int) *big.Int {
	tmp := new(big.Int).Set(d)
	return chopPrecisionAndRound(tmp)
}

// RoundInt64 rounds the decimal using bankers rounding
func (d Dec) RoundInt64() int64 {
	chopped := chopPrecisionAndRoundNonMutative(big.NewInt(d.int64))
	if !chopped.IsInt64() {
		panic("Int64() out of bound")
	}
	return chopped.Int64()
}

// RoundInt round the decimal using bankers rounding
func (d Dec) RoundInt() int64 {
	return d.RoundInt64()
}

//___________________________________________________________________________________

// similar to chopPrecisionAndRound, but always rounds down
func chopPrecisionAndTruncate(d int64) int64 {
	return d / precisionReuse
}

func chopPrecisionAndTruncateNonMutative(d int64) int64 {
	return chopPrecisionAndTruncate(d)
}

// TruncateInt64 truncates the decimals from the number and returns an int64
func (d Dec) TruncateInt64() int64 {
	return chopPrecisionAndTruncateNonMutative(d.int64)
}

// TruncateInt truncates the decimals from the number and returns an Int
func (d Dec) TruncateInt() int64 {
	return chopPrecisionAndTruncateNonMutative(d.int64)
}

//___________________________________________________________________________________

// wraps d.MarshalText()
func (d Dec) MarshalAmino() (int64, error) {
	return d.int64, nil
}

func (d Dec) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatInt(d.int64, 10)), nil
}

func (d *Dec) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 64)
	d.int64 = v
	return err
}

// requires a valid JSON string - strings quotes and calls UnmarshalText
func (d *Dec) UnmarshalAmino(v int64) (err error) {
	d.int64 = v
	return nil
}

// MarshalJSON marshals the decimal
func (d Dec) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON defines custom decoding scheme
func (d *Dec) UnmarshalJSON(bz []byte) error {
	var text string
	err := json.Unmarshal(bz, &text)
	if err != nil {
		return err
	}
	// TODO: Reuse dec allocation
	newDec, err := NewDecFromStr(text)
	if err != nil {
		return err
	}
	d.int64 = newDec.int64
	return nil
}

//___________________________________________________________________________________
// helpers

// test if two decimal arrays are equal
func DecsEqual(d1s, d2s []Dec) bool {
	if len(d1s) != len(d2s) {
		return false
	}

	for i, d1 := range d1s {
		if !d1.Equal(d2s[i]) {
			return false
		}
	}
	return true
}

// minimum decimal between two
func MinDec(d1, d2 Dec) Dec {
	if d1.LT(d2) {
		return d1
	}
	return d2
}

// maximum decimal between two
func MaxDec(d1, d2 Dec) Dec {
	if d1.LT(d2) {
		return d2
	}
	return d1
}

// intended to be used with require/assert:  require.True(DecEq(...))
func DecEq(t *testing.T, exp, got Dec) (*testing.T, bool, string, string, string) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}
