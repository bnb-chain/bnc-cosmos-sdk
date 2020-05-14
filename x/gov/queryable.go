package gov

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// query endpoints supported by the governance Querier
const (
	QueryProposals = "proposals"
	QueryProposal  = "proposal"
	QueryDeposits  = "deposits"
	QueryDeposit   = "deposit"
	QueryVotes     = "votes"
	QueryVote      = "vote"
	QueryTally     = "tally"

	ParsedRequestKey = "ParsedRequest"
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryProposals:
			p := new(QueryProposalsParams)
			ctx, err = RequestPrepare(ctx, keeper, req, p)
			if err != nil {
				return res, err
			}
			return queryProposals(ctx, path[1:], req, keeper)
		case QueryProposal:
			p := new(QueryProposalParams)
			ctx, err = RequestPrepare(ctx, keeper, req, p)
			if err != nil {
				return res, err
			}
			return queryProposal(ctx, path[1:], req, keeper)
		case QueryDeposits:
			p := new(QueryDepositsParams)
			ctx, err = RequestPrepare(ctx, keeper, req, p)
			if err != nil {
				return res, err
			}
			return queryDeposits(ctx, path[1:], req, keeper)
		case QueryDeposit:
			p := new(QueryDepositParams)
			ctx, err = RequestPrepare(ctx, keeper, req, p)
			if err != nil {
				return res, err
			}
			return queryDeposit(ctx, path[1:], req, keeper)
		case QueryVotes:
			p := new(QueryVotesParams)
			ctx, err = RequestPrepare(ctx, keeper, req, p)
			if err != nil {
				return res, err
			}
			return queryVotes(ctx, path[1:], req, keeper)
		case QueryVote:
			p := new(QueryVoteParams)
			ctx, err = RequestPrepare(ctx, keeper, req, p)
			if err != nil {
				return res, err
			}
			return queryVote(ctx, path[1:], req, keeper)
		case QueryTally:
			p := new(QueryTallyParams)
			ctx, err = RequestPrepare(ctx, keeper, req, p)
			if err != nil {
				return res, err
			}
			return queryTally(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown gov query endpoint")
		}
	}
}

// Params for query 'custom/gov/proposal'
type QueryProposalParams struct {
	BaseParams
	ProposalID int64
}

// nolint: unparam
func queryProposal(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	iParam := ctx.Value(ParsedRequestKey)
	if iParam == nil {
		return nil, sdk.ErrUnknownRequest("missing request data")
	}
	params, ok := iParam.(*QueryProposalParams)
	if !ok {
		return nil, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}

	proposal := keeper.GetProposal(ctx, params.ProposalID)
	if proposal == nil {
		return nil, ErrUnknownProposal(DefaultCodespace, params.ProposalID)
	}

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, proposal)
	if err2 != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err2.Error()))
	}
	return bz, nil
}

// Params for query 'custom/gov/deposit'
type QueryDepositParams struct {
	BaseParams
	ProposalID int64
	Depositer  sdk.AccAddress
}

// nolint: unparam
func queryDeposit(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	iParam := ctx.Value(ParsedRequestKey)
	if iParam == nil {
		return nil, sdk.ErrUnknownRequest("missing request data")
	}
	params, ok := iParam.(*QueryDepositParams)
	if !ok {
		return nil, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}

	deposit, _ := keeper.GetDeposit(ctx, params.ProposalID, params.Depositer)
	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, deposit)
	if err2 != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err2.Error()))
	}
	return bz, nil
}

// Params for query 'custom/gov/vote'
type QueryVoteParams struct {
	BaseParams
	ProposalID int64
	Voter      sdk.AccAddress
}

// nolint: unparam
func queryVote(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	iParam := ctx.Value(ParsedRequestKey)
	if iParam == nil {
		return nil, sdk.ErrUnknownRequest("missing request data")
	}
	params, ok := iParam.(*QueryVoteParams)
	if !ok {
		return nil, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}

	vote, _ := keeper.GetVote(ctx, params.ProposalID, params.Voter)
	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, vote)
	if err2 != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err2.Error()))
	}
	return bz, nil
}

// Params for query 'custom/gov/deposits'
type QueryDepositsParams struct {
	BaseParams
	ProposalID int64
}

// nolint: unparam
func queryDeposits(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	iParam := ctx.Value(ParsedRequestKey)
	if iParam == nil {
		return nil, sdk.ErrUnknownRequest("missing request data")
	}
	params, ok := iParam.(*QueryDepositsParams)
	if !ok {
		return nil, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}

	var deposits []Deposit
	depositsIterator := keeper.GetDeposits(ctx, params.ProposalID)
	defer depositsIterator.Close()
	for ; depositsIterator.Valid(); depositsIterator.Next() {
		deposit := Deposit{}
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(depositsIterator.Value(), &deposit)
		deposits = append(deposits, deposit)
	}

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, deposits)
	if err2 != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err2.Error()))
	}
	return bz, nil
}

// Params for query 'custom/gov/votes'
type QueryVotesParams struct {
	BaseParams
	ProposalID int64
}

// nolint: unparam
func queryVotes(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	iParam := ctx.Value(ParsedRequestKey)
	if iParam == nil {
		return nil, sdk.ErrUnknownRequest("missing request data")
	}
	params, ok := iParam.(*QueryVotesParams)
	if !ok {
		return nil, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}
	var votes []Vote
	votesIterator := keeper.GetVotes(ctx, params.ProposalID)
	defer votesIterator.Close()
	for ; votesIterator.Valid(); votesIterator.Next() {
		vote := Vote{}
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(votesIterator.Value(), &vote)
		votes = append(votes, vote)
	}

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, votes)
	if err2 != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err2.Error()))
	}
	return bz, nil
}

// Params for query 'custom/gov/proposals'
type QueryProposalsParams struct {
	BaseParams
	Voter              sdk.AccAddress
	Depositer          sdk.AccAddress
	ProposalStatus     ProposalStatus
	NumLatestProposals int64
}

// nolint: unparam
func queryProposals(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	iParam := ctx.Value(ParsedRequestKey)
	if iParam == nil {
		return nil, sdk.ErrUnknownRequest("missing request data")
	}
	params, ok := iParam.(*QueryProposalsParams)
	if !ok {
		return nil, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}

	proposals := keeper.GetProposalsFiltered(ctx, params.Voter, params.Depositer, params.ProposalStatus, params.NumLatestProposals)

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, proposals)
	if err2 != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err2.Error()))
	}
	return bz, nil
}

// Params for query 'custom/gov/tally'
type QueryTallyParams struct {
	BaseParams
	ProposalID int64
}

// nolint: unparam
func queryTally(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	iParam := ctx.Value(ParsedRequestKey)
	if iParam == nil {
		return nil, sdk.ErrUnknownRequest("missing request data")
	}
	params, ok := iParam.(*QueryTallyParams)
	if !ok {
		return nil, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}

	proposal := keeper.GetProposal(ctx, params.ProposalID)
	if proposal == nil {
		return nil, ErrUnknownProposal(DefaultCodespace, params.ProposalID)
	}

	var tallyResult TallyResult

	if proposal.GetStatus() == StatusDepositPeriod {
		tallyResult = EmptyTallyResult()
	} else if proposal.GetStatus() == StatusPassed || proposal.GetStatus() == StatusRejected {
		tallyResult = proposal.GetTallyResult()
	} else {
		_, _, tallyResult = Tally(ctx, keeper, proposal)
	}

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, tallyResult)
	if err2 != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err2.Error()))
	}
	return bz, nil
}

func RequestPrepare(ctx sdk.Context, k Keeper, req abci.RequestQuery, p SideChainIder) (newCtx sdk.Context, err sdk.Error) {
	if req.Data == nil || len(req.Data) == 0 {
		return ctx, nil
	}
	errRes := k.cdc.UnmarshalJSON(req.Data, p)
	if errRes != nil {
		return ctx, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("can not unmarshal request", errRes.Error()))
	}
	newCtx = ctx.WithValue(ParsedRequestKey, p)
	if len(p.GetSideChainId()) != 0 {
		newCtx, err = prepareSideChainCtx(newCtx, k, p.GetSideChainId())
		if err != nil {
			return newCtx, err
		}
	}
	return newCtx, nil
}

func prepareSideChainCtx(ctx sdk.Context, k Keeper, sideChainId string) (sdk.Context, sdk.Error) {
	scCtx, err := k.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId)
	if err != nil {
		return sdk.Context{}, types.ErrInvalidSideChainId(k.codespace)
	}
	return scCtx, nil
}

type BaseParams struct {
	SideChainId string
}

func (p BaseParams) GetSideChainId() string {
	return p.SideChainId
}

type SideChainIder interface {
	GetSideChainId() string
}

func NewBaseParams(sideChainId string) BaseParams {
	return BaseParams{
		SideChainId: sideChainId,
	}
}
