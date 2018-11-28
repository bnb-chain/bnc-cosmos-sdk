package gov

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

func TestTickExpiredDepositPeriod(t *testing.T) {
	mapp, keeper, stakeKeeper, addrs, pubKeys, _ := getMockApp(t, 10)

	validator := stake.NewValidator(sdk.ValAddress(addrs[0]), pubKeys[0], stake.Description{})
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(sdk.RunTxModeDeliver, abci.Header{ProposerAddress: pubKeys[0].Address()})

	// create validator
	stakeKeeper.SetValidator(ctx, validator)
	stakeKeeper.SetValidatorByConsAddr(ctx, validator)

	govHandler := NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[1], sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 1000e8)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositProcedure(ctx).MaxDepositPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.True(t, shouldPopInactiveProposalQueue(ctx, keeper))
	EndBlocker(ctx, keeper)

	validatorCoins := keeper.ck.GetCoins(ctx, addrs[0])
	// check distribute deposits to proposer
	require.Equal(t, validatorCoins, sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 6000e8)})

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
}

func TestTickMultipleExpiredDepositPeriod(t *testing.T) {
	mapp, keeper, stakeKeeper, addrs, pubKeys, _ := getMockApp(t, 10)

	validator := stake.NewValidator(sdk.ValAddress(addrs[0]), pubKeys[0], stake.Description{})
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(sdk.RunTxModeDeliver, abci.Header{ProposerAddress: pubKeys[0].Address()})

	// create validator
	stakeKeeper.SetValidator(ctx, validator)
	stakeKeeper.SetValidatorByConsAddr(ctx, validator)

	govHandler := NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 1000e8)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(2) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newProposalMsg2 := NewMsgSubmitProposal("Test2", "test2", ProposalTypeText, addrs[1], sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 5)})
	res = govHandler(ctx, newProposalMsg2)
	require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositProcedure(ctx).MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.True(t, shouldPopInactiveProposalQueue(ctx, keeper))
	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(5) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.True(t, shouldPopInactiveProposalQueue(ctx, keeper))
	EndBlocker(ctx, keeper)
	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
}

func TestTickPassedDepositPeriod(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(sdk.RunTxModeDeliver, abci.Header{})
	govHandler := NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	require.False(t, shouldPopActiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 1000e8)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID int64
	keeper.cdc.UnmarshalBinaryBare(res.Data, &proposalID)

	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))

	newDepositMsg := NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 1000e8)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.True(t, shouldPopInactiveProposalQueue(ctx, keeper))
	require.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))

	EndBlocker(ctx, keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
	require.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))
	require.False(t, shouldPopActiveProposalQueue(ctx, keeper))

}

func TestTickPassedVotingPeriodRejected(t *testing.T) {
	mapp, keeper, stakeKeeper, addrs, pubKeys, _ := getMockApp(t, 10)

	validator := stake.NewValidator(sdk.ValAddress(addrs[0]), pubKeys[0], stake.Description{})
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(sdk.RunTxModeDeliver, abci.Header{ProposerAddress: pubKeys[0].Address()})

	// create validator
	stakeKeeper.SetValidator(ctx, validator)
	stakeKeeper.SetValidatorByConsAddr(ctx, validator)

	govHandler := NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	require.False(t, shouldPopActiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 1000e8)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID int64
	keeper.cdc.UnmarshalBinaryBare(res.Data, &proposalID)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newDepositMsg := NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 1000e8)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())
	EndBlocker(ctx, keeper)

	// pass voting period
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetVotingProcedure(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	require.True(t, shouldPopActiveProposalQueue(ctx, keeper))
	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	require.True(t, depositsIterator.Valid())
	depositsIterator.Close()
	require.Equal(t, StatusVotingPeriod, keeper.GetProposal(ctx, proposalID).GetStatus())

	EndBlocker(ctx, keeper)

	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	depositsIterator = keeper.GetDeposits(ctx, proposalID)
	require.False(t, depositsIterator.Valid())
	depositsIterator.Close()
	require.Equal(t, StatusRejected, keeper.GetProposal(ctx, proposalID).GetStatus())
	require.True(t, keeper.GetProposal(ctx, proposalID).GetTallyResult().Equals(EmptyTallyResult()))

	// check distribute deposits to proposer
	validatorCoins := keeper.ck.GetCoins(ctx, addrs[0])
	require.Equal(t, validatorCoins, sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 6000e8)})
}

func TestTickPassedVotingPeriodPassed(t *testing.T) {
	mapp, keeper, stakeKeeper, addrs, pubKeys, _ := getMockApp(t, 3)

	validator0 := stake.NewValidator(sdk.ValAddress(addrs[0]), pubKeys[0], stake.Description{})
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(sdk.RunTxModeDeliver, abci.Header{ProposerAddress: pubKeys[0].Address()})

	// create and delegate validator
	stakeKeeper.SetValidator(ctx, validator0)
	stakeKeeper.SetValidatorByConsAddr(ctx, validator0)
	stakeKeeper.Delegate(ctx, sdk.AccAddress(addrs[2]), sdk.NewCoin(DefaultDepositDenom, 1000), validator0, true)

	stakeKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validator0, _ = stakeKeeper.GetValidator(ctx, validator0.OperatorAddr)

	govHandler := NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, shouldPopInactiveProposalQueue(ctx, keeper))
	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	require.False(t, shouldPopActiveProposalQueue(ctx, keeper))

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 1000e8)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID int64
	keeper.cdc.UnmarshalBinaryBare(res.Data, &proposalID)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newDepositMsg := NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 1000e8)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())
	EndBlocker(ctx, keeper)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)
	newVoteMsg := NewMsgVote(addrs[0], proposalID, OptionYes)
	res = govHandler(ctx, newVoteMsg)
	println(res.Log)
	require.True(t, res.IsOK())
	EndBlocker(ctx, keeper)

	// pass voting period
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetVotingProcedure(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	require.True(t, shouldPopActiveProposalQueue(ctx, keeper))
	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	require.True(t, depositsIterator.Valid())
	depositsIterator.Close()
	require.Equal(t, StatusVotingPeriod, keeper.GetProposal(ctx, proposalID).GetStatus())

	EndBlocker(ctx, keeper)

	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	depositsIterator = keeper.GetDeposits(ctx, proposalID)
	require.False(t, depositsIterator.Valid())
	depositsIterator.Close()
	require.Equal(t, StatusPassed, keeper.GetProposal(ctx, proposalID).GetStatus())

	// check refund deposits
	validatorCoins := keeper.ck.GetCoins(ctx, addrs[0])
	require.Equal(t, validatorCoins, sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 5000e8)})
}
