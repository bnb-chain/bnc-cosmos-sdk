package sidechain

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/sidechain/types"
)

const (
	SafeToleratePeriod = 2 * 7 * 24 * 60 * 60 * time.Second // 2 weeks
)

func (k *Keeper) getLastChanPermissionChanges(ctx sdk.Context) []types.ChanPermissionSetting {
	changes := make([]types.ChanPermissionSetting, 0)
	// It can still find the valid proposal if the block chain stop for SafeToleratePeriod time
	backPeriod := SafeToleratePeriod + gov.MaxVotingPeriod
	k.govKeeper.Iterate(ctx, nil, nil, gov.StatusNil, 0, true, func(proposal gov.Proposal) bool {
		if proposal.GetProposalType() == gov.ProposalTypeManageChanPermission {
			if ctx.BlockHeader().Time.Sub(proposal.GetVotingStartTime()) > backPeriod {
				return true
			}
			if proposal.GetStatus() != gov.StatusPassed {
				return false
			}

			proposal.SetStatus(gov.StatusExecuted)
			k.govKeeper.SetProposal(ctx, proposal)

			var setting types.ChanPermissionSetting
			strProposal := proposal.GetDescription()
			err := k.cdc.UnmarshalJSON([]byte(strProposal), &setting)
			if err != nil {
				ctx.Logger().With("module", "side_chain").Error("Get broken data when unmarshal ChanPermissionSetting msg, will skip.",
					"proposalId", proposal.GetProposalID(), "err", err)
				return false
			}
			if _, ok := k.cfg.destChainNameToID[setting.SideChainId]; !ok {
				ctx.Logger().With("module", "side_chain").Error("The SideChainId do not exist, will skip.",
					"proposalId", proposal.GetProposalID(), "setting", setting)
				return false
			}
			if _, ok := k.cfg.channelIDToName[setting.ChannelId]; !ok {
				ctx.Logger().With("module", "side_chain").Error("The ChannelId do not exist, will skip.",
					"proposalId", proposal.GetProposalID(), "setting", setting)
				return false
			}
			if err := setting.Check(); err != nil {
				ctx.Logger().With("module", "side_chain").Error("The ChanPermissionSetting proposal is invalid, will skip.",
					"proposalId", proposal.GetProposalID(), "setting", setting, "err", err)
				return false
			}
			changes = append(changes, setting)
		}
		return false
	})
	return changes
}

func (k *Keeper) SaveChannelSettingChangeToIbc(ctx sdk.Context, sideChainId sdk.ChainID, channelId sdk.ChannelID, permission sdk.ChannelPermission) (seq uint64, sdkErr sdk.Error) {
	valueBytes := []byte{byte(channelId), byte(permission)}
	ctx.Logger().Info("SaveChannelSettingChangeToIbc", "sideChainId", sideChainId, "channelId", channelId, "permission", permission, "valueBytes", valueBytes, "k", k)
	return k.sendParamChangeToIbc(ctx, sideChainId, EnableOrDisableChannelKey, valueBytes, CrossChainContractAddr)
}
