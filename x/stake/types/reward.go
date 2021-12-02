package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Sharer struct {
	AccAddr sdk.AccAddress
	Shares  sdk.Dec
}

type Reward struct {
	AccAddr sdk.AccAddress
	Shares  sdk.Dec
	Amount  int64
}

type StoredReward struct {
	Validator sdk.ValAddress
	AccAddr   sdk.AccAddress
	Shares    sdk.Dec
	Amount    int64
}

func MustUnmarshalReward(cdc *codec.Codec, value []byte) StoredReward {
	reward, err := UnmarshalReward(cdc, value)
	if err != nil {
		panic(err)
	}
	return reward
}

func UnmarshalReward(cdc *codec.Codec, value []byte) (reward StoredReward, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &reward)
	return reward, err
}

func MustMarshalRewards(cdc *codec.Codec, rewards []StoredReward) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(rewards)
}

func MustUnmarshalRewards(cdc *codec.Codec, value []byte) []StoredReward {
	rewards, err := UnmarshalRewards(cdc, value)
	if err != nil {
		panic(err)
	}
	return rewards
}

func UnmarshalRewards(cdc *codec.Codec, value []byte) (rewards []StoredReward, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &rewards)
	return rewards, err
}

func MustMarshalValDistAddr(cdc *codec.Codec, valDistAddrMap map[string]string) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(valDistAddrMap)
}

func MustUnmarshalValDistAddr(cdc *codec.Codec, value []byte) (valDistAddrMap map[string]string) {
	err := cdc.UnmarshalBinaryLengthPrefixed(value, &valDistAddrMap)
	if err != nil {
		panic(err)
	}
	return valDistAddrMap
}
