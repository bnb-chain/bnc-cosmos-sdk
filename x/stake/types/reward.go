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

func MustUnmarshalReward(cdc *codec.Codec, value []byte) Reward {
	reward, err := UnmarshalReward(cdc, value)
	if err != nil {
		panic(err)
	}
	return reward
}

func UnmarshalReward(cdc *codec.Codec, value []byte) (reward Reward, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &reward)
	return reward, err
}

func MustMarshalRewards(cdc *codec.Codec, rewards []Reward) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(rewards)
}

func MustUnmarshalRewards(cdc *codec.Codec, value []byte) []Reward {
	rewards, err := UnmarshalRewards(cdc, value)
	if err != nil {
		panic(err)
	}
	return rewards
}

func UnmarshalRewards(cdc *codec.Codec, value []byte) (rewards []Reward, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &rewards)
	return rewards, err
}