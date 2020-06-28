package oracle

import (
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"strconv"

	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/fees"
	"github.com/cosmos/cosmos-sdk/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/x/sidechain"
)

func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.ClaimMsg:
			return handleClaimMsg(ctx, keeper, msg)
		default:
			errMsg := "Unrecognized oracle msg type"
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleClaimMsg(ctx sdk.Context, oracleKeeper Keeper, msg ClaimMsg) sdk.Result {
	claim := NewClaim(types.GetClaimId(msg.ChainId, types.RelayPackagesChannelId, msg.Sequence),
		sdk.ValAddress(msg.ValidatorAddress), hex.EncodeToString(msg.Payload))

	sequence := oracleKeeper.ScKeeper.GetReceiveSequence(ctx, msg.ChainId, types.RelayPackagesChannelId)
	if sequence != msg.Sequence {
		return types.ErrInvalidSequence(fmt.Sprintf("current sequence of channel %d is %d", types.RelayPackagesChannelId, sequence)).Result()
	}

	prophecy, sdkErr := oracleKeeper.ProcessClaim(ctx, claim)
	if sdkErr != nil {
		return sdkErr.Result()
	}

	if prophecy.Status.Text == types.FailedStatusText {
		oracleKeeper.DeleteProphecy(ctx, prophecy.ID)
		return sdk.Result{}
	}

	if prophecy.Status.Text != types.SuccessStatusText {
		return sdk.Result{}
	}

	packages := types.Packages{}
	err := rlp.DecodeBytes(msg.Payload, &packages)
	if err != nil {
		return types.ErrInvalidPayload("decode packages error").Result()
	}

	events := make([]sdk.Event, 0, len(packages))
	for _, pack := range packages {
		event, sdkErr := handlePackage(ctx, oracleKeeper, msg.ChainId, &pack)
		if sdkErr != nil {
			// only do log, but let reset package get chance to execute.
			ctx.Logger().With("module", "oracle").Error(fmt.Sprintf("failed to process package channel %d, sequence %d, error %v", pack.ChannelId, pack.Sequence, sdkErr))
		}
		events = append(events, event)
	}

	// delete prophecy when execute claim success
	oracleKeeper.DeleteProphecy(ctx, prophecy.ID)
	oracleKeeper.ScKeeper.IncrReceiveSequence(ctx, msg.ChainId, types.RelayPackagesChannelId)

	return sdk.Result{
		Events: events,
	}
}

func handlePackage(ctx sdk.Context, oracleKeeper Keeper, chainId sdk.IbcChainID, pack *types.Package) (sdk.Event, sdk.Error) {
	logger := ctx.Logger().With("module", "x/oracle")
	// increase claim type sequence
	oracleKeeper.ScKeeper.IncrReceiveSequence(ctx, chainId, pack.ChannelId)

	crossChainApp := oracleKeeper.ScKeeper.GetCrossChainApp(ctx, pack.ChannelId)

	if crossChainApp == nil {
		return sdk.Event{}, types.ErrChannelNotRegistered(fmt.Sprintf("channel %d not registered", pack.ChannelId))
	}

	sequence := oracleKeeper.ScKeeper.GetReceiveSequence(ctx, chainId, pack.ChannelId)
	if sequence != pack.Sequence {
		return sdk.Event{}, types.ErrInvalidSequence(fmt.Sprintf("current sequence of channel %d is %d", pack.ChannelId, sequence))
	}

	packageType, relayFee, err := sidechain.DecodePackageHeader(pack.Payload)
	if err != nil {
		return sdk.Event{}, types.ErrInvalidPayloadHeader(err.Error())
	}

	if !sdk.IsValidCrossChainPackageType(packageType) {
		return sdk.Event{}, types.ErrInvalidPackageType()
	}

	feeAmount := relayFee.Int64()
	if feeAmount < 0 {
		return sdk.Event{}, types.ErrFeeOverflow("relayFee overflow")
	}

	fee := sdk.Coins{sdk.Coin{Denom: sdk.NativeTokenSymbol, Amount: feeAmount}}
	_, _, sdkErr := oracleKeeper.BkKeeper.SubtractCoins(ctx, sdk.PegAccount, fee)
	if sdkErr != nil {
		return sdk.Event{}, sdkErr
	}

	if ctx.IsDeliverTx() {
		// add changed accounts
		oracleKeeper.Pool.AddAddrs([]sdk.AccAddress{sdk.PegAccount})

		// add fee
		fees.Pool.AddAndCommitFee(
			fmt.Sprintf("cross_communication:%d:%d:%v", pack.ChannelId, pack.Sequence, packageType),
			sdk.Fee{
				Tokens: fee,
				Type:   sdk.FeeForProposer,
			},
		)
	}

	cacheCtx, write := ctx.CacheContext()
	crash, result := executeClaim(cacheCtx, crossChainApp, pack.Payload, packageType)
	if result.IsOk() {
		write()
	} else if ctx.IsDeliverTx() {
		oracleKeeper.Metrics.ErrNumOfChannels.With("channel_id", fmt.Sprintf("%d", pack.ChannelId)).Add(1)
	}

	// write ack package
	if packageType == sdk.SynCrossChainPackageType {
		if crash {
			_, err := oracleKeeper.IbcKeeper.CreateRawIBCPackageById(ctx, chainId,
				pack.ChannelId, sdk.FailAckCrossChainPackageType, pack.Payload)
			if err != nil {
				logger.Error("failed to write FailAckCrossChainPackage", "err", err)
			}
		} else {
			if len(result.Payload) != 0 {
				_, err := oracleKeeper.IbcKeeper.CreateRawIBCPackageById(ctx, chainId,
					pack.ChannelId, sdk.AckCrossChainPackageType, result.Payload)
				if err != nil {
					logger.Error("failed to write AckCrossChainPackage", "err", err)
				}
			}
		}
	}

	resultTags := sdk.NewTags(
		sdk.GetPegOutTagName(sdk.NativeTokenSymbol), []byte(strconv.FormatInt(feeAmount, 10)),
		types.ClaimResultCode, []byte(strconv.FormatInt(int64(result.Code()), 10)),
		types.ClaimResultMsg, []byte(result.Msg()),
		types.ClaimPackageType, []byte(strconv.FormatInt(int64(packageType), 10)),
		// The following tags are for index
		types.ClaimChannel, []byte{uint8(pack.ChannelId)},
		types.ClaimSequence, []byte(strconv.FormatUint(pack.Sequence, 10)),
	)

	if result.Tags != nil {
		resultTags = resultTags.AppendTags(result.Tags)
	}

	event := sdk.Event{
		Type:       types.EventTypeClaim,
		Attributes: resultTags,
	}

	return event, nil
}

func executeClaim(ctx sdk.Context, app sdk.CrossChainApplication, payload []byte, packageType sdk.CrossChainPackageType) (crash bool, result sdk.ExecuteResult) {
	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("recovered: %v\nstack:\n%v", r, string(debug.Stack()))
			logger := ctx.Logger().With("module", "oracle")
			logger.Error("execute claim panic", "err_log", log)
			crash = true
			result = sdk.ExecuteResult{
				Err: sdk.ErrInternal(fmt.Sprintf("execute claim failed: %v", r)),
			}
		}
	}()

	switch packageType {
	case sdk.SynCrossChainPackageType:
		result = app.ExecuteSynPackage(ctx, payload[sidechain.PackageHeaderLength:])
	case sdk.AckCrossChainPackageType:
		result = app.ExecuteAckPackage(ctx, payload[sidechain.PackageHeaderLength:])
	case sdk.FailAckCrossChainPackageType:
		result = app.ExecuteFailAckPackage(ctx, payload[sidechain.PackageHeaderLength:])
	default:
		panic(fmt.Sprintf("receive unexpected package type %d", packageType))
	}
	return
}
