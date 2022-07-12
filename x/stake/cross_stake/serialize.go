package cross_stake

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const (
	CrossStakeErrValidatorNotFound uint8 = 1
	CrossStakeErrValidatorJailed   uint8 = 2
	CrossStakeErrBadDelegation     uint8 = 3
)

type CrossStakePackage struct {
	EventCode   types.CrossStakePackageType
	ParamsBytes []byte
}

func DeserializeCrossStakeSynPackage(serializedPackage []byte) (types.CrossStakePackageType, interface{}, error) {
	var pack1 CrossStakePackage
	err := rlp.DecodeBytes(serializedPackage, &pack1)
	if err != nil {
		return 0, nil, err
	}
	switch pack1.EventCode {
	case types.CrossStakeTypeDelegate:
		var pack2 types.CrossStakeDelegateSynPackage
		err := rlp.DecodeBytes(pack1.ParamsBytes, &pack2)
		if err != nil {
			return 0, nil, err
		}
		return pack1.EventCode, pack2, nil
	case types.CrossStakeTypeUndelegate:
		var pack2 types.CrossStakeUndelegateSynPackage
		err := rlp.DecodeBytes(pack1.ParamsBytes, &pack2)
		if err != nil {
			return 0, nil, err
		}
		return pack1.EventCode, pack2, nil
	case types.CrossStakeTypeRedelegate:
		var pack2 types.CrossStakeRedelegateSynPackage
		err := rlp.DecodeBytes(pack1.ParamsBytes, &pack2)
		if err != nil {
			return 0, nil, err
		}
		return pack1.EventCode, pack2, nil
	default:
		return 0, nil, fmt.Errorf("unrecognized package type")
	}
}
