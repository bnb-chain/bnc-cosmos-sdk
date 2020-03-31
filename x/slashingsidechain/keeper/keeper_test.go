package keeper

import (
	"crypto/rand"
	"github.com/cosmos/cosmos-sdk/x/slashingsidechain/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestParams(t *testing.T) {
	ctx, keeper := CreateTestInput(t)
	defaultParams := types.DefaultParams()
	require.Equal(t, defaultParams.SlashAmount, keeper.SlashAmount(ctx))
	require.Equal(t, defaultParams.SubmitterReward, keeper.SubmitterReward(ctx))
	require.Equal(t, defaultParams.MaxEvidenceAge, keeper.MaxEvidenceAge(ctx))

	newParams := types.Params{
		SlashAmount:     1000000000,
		SubmitterReward: 500000000,
		MaxEvidenceAge:  60 * 60 * 24 * 7 * time.Second,
	}
	keeper.SetParams(ctx, newParams)
	require.Equal(t, newParams.SlashAmount, keeper.SlashAmount(ctx))
	require.Equal(t, newParams.SubmitterReward, keeper.SubmitterReward(ctx))
	require.Equal(t, newParams.MaxEvidenceAge, keeper.MaxEvidenceAge(ctx))
}

func TestSlashRecord(t *testing.T) {
	ctx, keeper := CreateTestInput(t)
	sideConsAddr := randomSideConsAddr()
	keeper.SetSlashRecord(ctx,sideConsAddr,100)
	require.NotNil(t,keeper.GetSlashRecord(ctx,sideConsAddr,100))
}

func randomSideConsAddr() []byte {
	bz := make([]byte,20)
	rand.Read(bz)
	return bz
}
