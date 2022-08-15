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

type CrossStakeSynPackageFromBSC struct {
	EventType   types.CrossStakeEventType
	ParamsBytes []byte
}

func DeserializeCrossStakeSynPackage(serializedPackage []byte) (interface{}, error) {
	var pack1 CrossStakeSynPackageFromBSC
	err := rlp.DecodeBytes(serializedPackage, &pack1)
	if err != nil {
		return nil, err
	}
	switch pack1.EventType {
	case types.CrossStakeTypeDelegate:
		var pack2 types.CrossStakeDelegateSynPackage
		err := rlp.DecodeBytes(pack1.ParamsBytes, &pack2)
		if err != nil {
			return nil, err
		}
		return &pack2, nil
	case types.CrossStakeTypeUndelegate:
		var pack2 types.CrossStakeUndelegateSynPackage
		err := rlp.DecodeBytes(pack1.ParamsBytes, &pack2)
		if err != nil {
			return nil, err
		}
		return &pack2, nil
	case types.CrossStakeTypeRedelegate:
		var pack2 types.CrossStakeRedelegateSynPackage
		err := rlp.DecodeBytes(pack1.ParamsBytes, &pack2)
		if err != nil {
			return nil, err
		}
		return &pack2, nil
	default:
		return nil, fmt.Errorf("unrecognized cross stake event type: %d", pack1.EventType)
	}
}

func DeserializeCrossStakeRefundPackage(serializedPackage []byte) (*types.CrossStakeRefundPackage, error) {
	var pack types.CrossStakeRefundPackage
	err := rlp.DecodeBytes(serializedPackage, &pack)
	if err != nil {
		return nil, err
	}
	return &pack, nil
}

func DeserializeCrossStakeFailAckPackage(serializedPackage []byte) (interface{}, error) {
	deserializeFuncSet := []func(serializedPackage []byte) (interface{}, error){
		func(serializedPackage []byte) (interface{}, error) {
			var pack types.CrossStakeDistributeRewardSynPackage
			err := rlp.DecodeBytes(serializedPackage, &pack)
			if err != nil {
				return nil, err
			}
			if pack.EventType != types.CrossStakeTypeDistributeReward {
				return nil, fmt.Errorf("wrong cross stake event type")
			}
			return &pack, nil
		},
		func(serializedPackage []byte) (interface{}, error) {
			var pack types.CrossStakeDistributeUndelegatedSynPackage
			err := rlp.DecodeBytes(serializedPackage, &pack)
			if err != nil {
				return nil, err
			}
			if pack.EventType != types.CrossStakeTypeDistributeUndelegated {
				return nil, fmt.Errorf("wrong cross stake event type")
			}
			return &pack, nil
		},
	}

	var pack interface{}
	var err error
	for _, deserializeFunc := range deserializeFuncSet {
		pack, err = deserializeFunc(serializedPackage)
		if err == nil {
			return pack, nil
		}
	}
	return nil, err
}
