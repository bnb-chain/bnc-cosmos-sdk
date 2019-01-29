package gov_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

func TestTickExpiredDepositPeriod(t *testing.T) {
	mapp, ck, keeper, stakeKeeper, addrs, pubKeys, _ := getMockApp(t, 10)

	validator := stake.NewValidator(sdk.ValAddress(addrs[0]), pubKeys[0], stake.Description{})
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(sdk.RunTxModeDeliver, abci.Header{ProposerAddress: pubKeys[0].Address()})

	// create validator
	stakeKeeper.SetValidator(ctx, validator)
	stakeKeeper.SetValidatorByConsAddr(ctx, validator)

	govHandler := gov.NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))

	newProposalMsg := gov.NewMsgSubmitProposal("Test", "test", gov.ProposalTypeText, addrs[1], sdk.Coins{sdk.NewCoin(gov.DefaultDepositDenom, 1000e8)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	gov.EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	gov.EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositProcedure(ctx).MaxDepositPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.True(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))
	gov.EndBlocker(ctx, keeper)

	validatorCoins := ck.GetCoins(ctx, addrs[0])
	// check distribute deposits to proposer
	require.Equal(t, validatorCoins, sdk.Coins{sdk.NewCoin(gov.DefaultDepositDenom, 6000e8)})

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))
}

func TestTickMultipleExpiredDepositPeriod(t *testing.T) {
	mapp, _, keeper, stakeKeeper, addrs, pubKeys, _ := getMockApp(t, 10)

	validator := stake.NewValidator(sdk.ValAddress(addrs[0]), pubKeys[0], stake.Description{})
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(sdk.RunTxModeDeliver, abci.Header{ProposerAddress: pubKeys[0].Address()})

	// create validator
	stakeKeeper.SetValidator(ctx, validator)
	stakeKeeper.SetValidatorByConsAddr(ctx, validator)

	govHandler := gov.NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))

	newProposalMsg := gov.NewMsgSubmitProposal("Test", "test", gov.ProposalTypeText, addrs[0], sdk.Coins{sdk.NewCoin(gov.DefaultDepositDenom, 1000e8)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	gov.EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(2) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	gov.EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))

	newProposalMsg2 := gov.NewMsgSubmitProposal("Test2", "test2", gov.ProposalTypeText, addrs[1], sdk.Coins{sdk.NewCoin(gov.DefaultDepositDenom, 5)})
	res = govHandler(ctx, newProposalMsg2)
	require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositProcedure(ctx).MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.True(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))
	gov.EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(5) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.True(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))
	gov.EndBlocker(ctx, keeper)
	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))
}

func TestTickPassedDepositPeriod(t *testing.T) {
	mapp, _, keeper, _, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(sdk.RunTxModeDeliver, abci.Header{})
	govHandler := gov.NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))
	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopActiveProposalQueue(ctx, keeper))

	newProposalMsg := gov.NewMsgSubmitProposal("Test", "test", gov.ProposalTypeText, addrs[0], sdk.Coins{sdk.NewCoin(gov.DefaultDepositDenom, 1000e8)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	proposalID, _ := strconv.Atoi(string(res.Data))

	gov.EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	gov.EndBlocker(ctx, keeper)
	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))

	newDepositMsg := gov.NewMsgDeposit(addrs[1], int64(proposalID), sdk.Coins{sdk.NewCoin(gov.DefaultDepositDenom, 1000e8)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	require.NotNil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.True(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))
	require.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))

	gov.EndBlocker(ctx, keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))
	require.NotNil(t, keeper.ActiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopActiveProposalQueue(ctx, keeper))

}

func TestTickPassedVotingPeriodRejected(t *testing.T) {
	mapp, ck, keeper, stakeKeeper, addrs, pubKeys, _ := getMockApp(t, 10)

	validator := stake.NewValidator(sdk.ValAddress(addrs[0]), pubKeys[0], stake.Description{})
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(sdk.RunTxModeDeliver, abci.Header{ProposerAddress: pubKeys[0].Address()})

	// create validator
	stakeKeeper.SetValidator(ctx, validator)
	stakeKeeper.SetValidatorByConsAddr(ctx, validator)

	govHandler := gov.NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))
	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopActiveProposalQueue(ctx, keeper))

	newProposalMsg := gov.NewMsgSubmitProposal("Test", "test", gov.ProposalTypeText, addrs[0], sdk.Coins{sdk.NewCoin(gov.DefaultDepositDenom, 1000e8)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	proposalID, _ := strconv.Atoi(string(res.Data))

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newDepositMsg := gov.NewMsgDeposit(addrs[1], int64(proposalID), sdk.Coins{sdk.NewCoin(gov.DefaultDepositDenom, 1000e8)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())
	gov.EndBlocker(ctx, keeper)

	// pass voting period
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetVotingProcedure(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	require.True(t, gov.ShouldPopActiveProposalQueue(ctx, keeper))
	depositsIterator := keeper.GetDeposits(ctx, int64(proposalID))
	require.True(t, depositsIterator.Valid())
	depositsIterator.Close()
	require.Equal(t, gov.StatusVotingPeriod, keeper.GetProposal(ctx, int64(proposalID)).GetStatus())

	gov.EndBlocker(ctx, keeper)

	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	depositsIterator = keeper.GetDeposits(ctx, int64(proposalID))
	require.False(t, depositsIterator.Valid())
	depositsIterator.Close()
	require.Equal(t, gov.StatusRejected, keeper.GetProposal(ctx, int64(proposalID)).GetStatus())
	require.True(t, keeper.GetProposal(ctx, int64(proposalID)).GetTallyResult().Equals(gov.EmptyTallyResult()))

	// check distribute deposits to proposer
	validatorCoins := ck.GetCoins(ctx, addrs[0])
	require.Equal(t, validatorCoins, sdk.Coins{sdk.NewCoin(gov.DefaultDepositDenom, 6000e8)})
}

func TestTickPassedVotingPeriodPassed(t *testing.T) {
	mapp, ck, keeper, stakeKeeper, addrs, pubKeys, _ := getMockApp(t, 3)

	validator0 := stake.NewValidator(sdk.ValAddress(addrs[0]), pubKeys[0], stake.Description{})
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(sdk.RunTxModeDeliver, abci.Header{ProposerAddress: pubKeys[0].Address()})

	// create and delegate validator
	stakeKeeper.SetValidator(ctx, validator0)
	stakeKeeper.SetValidatorByConsAddr(ctx, validator0)
	stakeKeeper.Delegate(ctx, sdk.AccAddress(addrs[2]), sdk.NewCoin(gov.DefaultDepositDenom, 1000), validator0, true)

	stakeKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validator0, _ = stakeKeeper.GetValidator(ctx, validator0.OperatorAddr)

	govHandler := gov.NewHandler(keeper)

	require.Nil(t, keeper.InactiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopInactiveProposalQueue(ctx, keeper))
	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	require.False(t, gov.ShouldPopActiveProposalQueue(ctx, keeper))

	newProposalMsg := gov.NewMsgSubmitProposal("Test", "test", gov.ProposalTypeText, addrs[0], sdk.Coins{sdk.NewCoin(gov.DefaultDepositDenom, 1000e8)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	proposalID, _ := strconv.Atoi(string(res.Data))

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newDepositMsg := gov.NewMsgDeposit(addrs[1], int64(proposalID), sdk.Coins{sdk.NewCoin(gov.DefaultDepositDenom, 1000e8)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())
	gov.EndBlocker(ctx, keeper)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)
	newVoteMsg := gov.NewMsgVote(addrs[0], int64(proposalID), gov.OptionYes)
	res = govHandler(ctx, newVoteMsg)
	println(res.Log)
	require.True(t, res.IsOK())
	gov.EndBlocker(ctx, keeper)

	// pass voting period
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetVotingProcedure(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	require.True(t, gov.ShouldPopActiveProposalQueue(ctx, keeper))
	depositsIterator := keeper.GetDeposits(ctx, int64(proposalID))
	require.True(t, depositsIterator.Valid())
	depositsIterator.Close()
	require.Equal(t, gov.StatusVotingPeriod, keeper.GetProposal(ctx, int64(proposalID)).GetStatus())

	gov.EndBlocker(ctx, keeper)

	require.Nil(t, keeper.ActiveProposalQueuePeek(ctx))
	depositsIterator = keeper.GetDeposits(ctx, int64(proposalID))
	require.False(t, depositsIterator.Valid())
	depositsIterator.Close()
	require.Equal(t, gov.StatusPassed, keeper.GetProposal(ctx, int64(proposalID)).GetStatus())

	// check refund deposits
	validatorCoins := ck.GetCoins(ctx, addrs[0])
	require.Equal(t, validatorCoins, sdk.Coins{sdk.NewCoin(gov.DefaultDepositDenom, 5000e8)})
}
