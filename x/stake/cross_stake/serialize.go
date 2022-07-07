package cross_stake

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const (
	CrossStakeErrCodeExpired       uint8 = 1
	CrossStakeErrValidatorNotFound uint8 = 2
	CrossStakeErrValidatorJailed   uint8 = 3
	CrossStakeErrBadDelegation     uint8 = 4
)

func DeserializeCrossStakeSynPackage(serializedPackage []byte) (types.CrossStakePackageType, error) {
	var eventType types.CrossStakePackageType
	switch {
	case serializedPackage[0] >= 192 && serializedPackage[0] <= 247:
		eventType = types.CrossStakePackageType(serializedPackage[1])
	case serializedPackage[0] >= 248:
		eventType = types.CrossStakePackageType(serializedPackage[2])
	default:
		return eventType, fmt.Errorf("wrong package length")
	}
	return eventType, nil
}
