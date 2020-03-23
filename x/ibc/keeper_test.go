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
	sourceChainID := sdk.CrossChainID(0x0001)
	destChainID := sdk.CrossChainID(0x000f)
	channelName := "transfer"
	channelID := sdk.ChannelID(0x01)
	sdk.SetSourceChainID(sourceChainID)
	err := sdk.RegisterNewCrossChainChannel(channelName, channelID)
	require.NoError(t, err)
	ctx, keeper := createTestInput(t, true)


	sequence := keeper.GetNextSequence(ctx, destChainID, channelName)
	require.Equal(t, uint64(0), sequence)

	value := []byte{0x00}
	err = keeper.CreateIBCPackage(ctx, destChainID, channelName, value)
	require.NoError(t, err)
	sequence = keeper.GetNextSequence(ctx, destChainID, channelName)
	require.Equal(t, uint64(1), sequence)

	value = []byte{0x00, 0x01}
	err = keeper.CreateIBCPackage(ctx, destChainID, channelName, value)
	require.NoError(t, err)
	value = []byte{0x00, 0x01, 0x02}
	err = keeper.CreateIBCPackage(ctx, destChainID, channelName, value)
	require.NoError(t, err)
	value = []byte{0x00, 0x01, 0x02, 0x03}
	err = keeper.CreateIBCPackage(ctx, destChainID, channelName, value)
	require.NoError(t, err)
	value = []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	err = keeper.CreateIBCPackage(ctx, destChainID, channelName, value)
	require.NoError(t, err)

	sequence = keeper.GetNextSequence(ctx, destChainID, channelName)
	require.Equal(t, uint64(5), sequence)

	keeper.CleanupIBCPackage(ctx, destChainID, channelName, 3)

	ibcPackage, err := keeper.GetIBCPackage(ctx, destChainID, channelName, 0)
	require.NoError(t, err)
	require.Nil(t, ibcPackage)
	ibcPackage, err = keeper.GetIBCPackage(ctx, destChainID, channelName, 1)
	require.NoError(t, err)
	require.Nil(t, ibcPackage)
	ibcPackage, err = keeper.GetIBCPackage(ctx, destChainID, channelName, 2)
	require.NoError(t, err)
	require.Nil(t, ibcPackage)
	ibcPackage, err = keeper.GetIBCPackage(ctx, destChainID, channelName, 3)
	require.NoError(t, err)
	require.Nil(t, ibcPackage)
	ibcPackage, err = keeper.GetIBCPackage(ctx, destChainID, channelName, 4)
	require.NoError(t, err)
	require.NotNil(t, ibcPackage)

	destChainID = sdk.CrossChainID(0x0002)
	channelID = sdk.ChannelID(0x01)
	sequence = keeper.GetNextSequence(ctx, destChainID, channelName)
	require.Equal(t, uint64(0), sequence)

	destChainID = sdk.CrossChainID(0x0001)
	channelID = sdk.ChannelID(0x02)
	sequence = keeper.GetNextSequence(ctx, destChainID, channelName)
	require.Equal(t, uint64(0), sequence)

}
