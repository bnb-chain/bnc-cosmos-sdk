package types

import (
	"encoding/binary"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IbcValidator struct {
	ConsAddr []byte
	FeeAddr  []byte
	DistAddr sdk.AccAddress
	Power    int64
}

// {20 bytes consensusAddress} + {20 bytes feeAddress} + {20 bytes distributionAddress} + {8 bytes voting power}
func (v *IbcValidator) Serialize() ([]byte, error) {
	consAddrLen, feeAddrLen, distAddrLen, powerLen:= len(v.ConsAddr), len(v.FeeAddr), len(v.DistAddr), 8
	if consAddrLen == 0 || feeAddrLen == 0 || distAddrLen == 0 || v.Power == 0 {
		return nil, errors.New("not all the IbcValidator fields are filled")
	}
	totalLen := consAddrLen + feeAddrLen + distAddrLen + powerLen
	result := make([]byte, totalLen)
	copy(result[:consAddrLen], v.ConsAddr)
	copy(result[consAddrLen:consAddrLen+feeAddrLen], v.FeeAddr)
	copy(result[consAddrLen+feeAddrLen:totalLen-powerLen], v.DistAddr)
	binary.BigEndian.PutUint64(result[totalLen-powerLen:], uint64(v.Power))
	return result, nil
}

type IbcValidatorSet []IbcValidator

func (vs IbcValidatorSet) Serialize() ([]byte, error) {
	vsLen := len(vs)
	if vsLen == 0 {
		return nil, errors.New("empty validator set")
	}

	v0 := vs[0]
	consAddrLen, feeAddrLen, distAddrLen, powerLen:= len(v0.ConsAddr), len(v0.FeeAddr), sdk.AddrLen, 8
	eachLen := consAddrLen + feeAddrLen + distAddrLen + powerLen
	result := make([]byte, vsLen*eachLen)
	for i:= range vs {
		v := vs[i]
		if len(v.ConsAddr) != consAddrLen ||
			len(v.FeeAddr) != feeAddrLen ||
			len(v.DistAddr) != distAddrLen ||
			v.Power == 0 {
			return nil, errors.New("not all validators' fields are complete")
		}
		start := i*eachLen
		copy(result[start: start+consAddrLen], vs[i].ConsAddr)
		copy(result[start+consAddrLen: start+consAddrLen+feeAddrLen], vs[i].FeeAddr)
		copy(result[start+consAddrLen+feeAddrLen: start+eachLen-powerLen], vs[i].DistAddr)
		binary.BigEndian.PutUint64(result[start+eachLen-powerLen:start+eachLen], uint64(vs[i].Power))
	}
	return result, nil
}


