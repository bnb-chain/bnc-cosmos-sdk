package ibc

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

func createTestInput(t *testing.T, isCheckTx bool) (sdk.Context, Keeper) {
	keyIBC := sdk.NewKVStoreKey("ibc")
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyIBC, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	mode := sdk.RunTxModeDeliver
	if isCheckTx {
		mode = sdk.RunTxModeCheck
	}

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, mode, log.NewNopLogger())
	ibcKeeper := NewKeeper(keyIBC, DefaultCodespace)

	return ctx, ibcKeeper
}

func TestKeeper(t *testing.T) {
	sourceChainID := sdk.IbcChainID(0x0001)

	destChainName := "bsc"
	destChainID := sdk.IbcChainID(0x000f)

	channelName := "transfer"
	channelID := sdk.IbcChannelID(0x01)

	SetSrcIbcChainID(sourceChainID)
	require.NoError(t, RegisterDestChain(destChainName, destChainID))
	require.NoError(t, RegisterChannel(channelName, channelID))

	ctx, keeper := createTestInput(t, true)


	value := []byte{0x00}
	sequence, err := keeper.CreateIBCPackage(ctx, destChainName, channelName, value)
	require.NoError(t, err)
	require.Equal(t, uint64(0), sequence)

	value = []byte{0x00, 0x01}
	sequence, err = keeper.CreateIBCPackage(ctx, destChainName, channelName, value)
	require.NoError(t, err)
	require.Equal(t, uint64(1), sequence)
	value = []byte{0x00, 0x01, 0x02}
	sequence, err = keeper.CreateIBCPackage(ctx, destChainName, channelName, value)
	require.NoError(t, err)
	require.Equal(t, uint64(2), sequence)
	value = []byte{0x00, 0x01, 0x02, 0x03}
	sequence, err = keeper.CreateIBCPackage(ctx, destChainName, channelName, value)
	require.NoError(t, err)
	require.Equal(t, uint64(3), sequence)
	value = []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	sequence, err = keeper.CreateIBCPackage(ctx, destChainName, channelName, value)
	require.NoError(t, err)
	require.Equal(t, uint64(4), sequence)


	keeper.CleanupIBCPackage(ctx, destChainName, channelName, 3)

	ibcPackage, sdkErr := keeper.GetIBCPackage(ctx, destChainName, channelName, 0)
	require.NoError(t, sdkErr)
	require.Nil(t, ibcPackage)
	ibcPackage, sdkErr = keeper.GetIBCPackage(ctx, destChainName, channelName, 1)
	require.NoError(t, sdkErr)
	require.Nil(t, ibcPackage)
	ibcPackage, sdkErr = keeper.GetIBCPackage(ctx, destChainName, channelName, 2)
	require.NoError(t, sdkErr)
	require.Nil(t, ibcPackage)
	ibcPackage, sdkErr = keeper.GetIBCPackage(ctx, destChainName, channelName, 3)
	require.NoError(t, sdkErr)
	require.Nil(t, ibcPackage)
	ibcPackage, sdkErr = keeper.GetIBCPackage(ctx, destChainName, channelName, 4)
	require.NoError(t, sdkErr)
	require.NotNil(t, ibcPackage)

	require.NoError(t, RegisterDestChain("btc", sdk.IbcChainID(0x0002)))
	sequence, err = keeper.CreateIBCPackage(ctx, "btc", channelName, value)
	require.NoError(t, err)
	require.Equal(t, uint64(0), sequence)

	require.NoError(t, RegisterChannel("mockChannel", sdk.IbcChannelID(2)))
	sequence, err = keeper.CreateIBCPackage(ctx, destChainName, "mockChannel", value)
	require.NoError(t, err)
	require.Equal(t, uint64(0), sequence)
	require.Equal(t, uint64(0), sequence)

}
