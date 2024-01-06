package cli

import (
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// nolint
const (
	FlagAddressDelegator             = "address-delegator"
	FlagAddressValidator             = "validator"
	FlagAddressValidatorSrc          = "addr-validator-source"
	FlagAddressValidatorDst          = "addr-validator-dest"
	FlagAddressSmartChainValidator   = "address-smart-chain-validator"
	FlagAddressSmartChainBeneficiary = "address-smart-chain-beneficiary"
	FlagPubKey                       = "pubkey"
	FlagAmount                       = "amount"
	FlagSharesAmount                 = "shares-amount"
	FlagSharesPercent                = "shares-percent"

	FlagMoniker  = "moniker"
	FlagIdentity = "identity"
	FlagWebsite  = "website"
	FlagDetails  = "details"

	FlagCommissionRate          = "commission-rate"
	FlagCommissionMaxRate       = "commission-max-rate"
	FlagCommissionMaxChangeRate = "commission-max-change-rate"

	FlagGenesisFormat = "genesis-format"
	FlagOffline       = "offline"
	FlagNodeID        = "node-id"
	FlagIP            = "ip"

	FlagProposalID        = "proposal-id"
	FlagConsAddrValidator = "cons-addr-validator"
	FlagDeposit           = "deposit"
	FlagVotingPeriod      = "voting-period"

	FlagOutputDocument = "output-document" // inspired by wget -O

	FlagSideChainId  = "side-chain-id"
	FlagSideConsAddr = "side-cons-addr"
	FlagSideFeeAddr  = "side-fee-addr"
	FlagSideVoteAddr = "side-vote-addr"
	FlagBLSWalletDir = "bls-wallet"
	FlagBLSPassword  = "bls-password"
)

// common flagsets to add to various functions
var (
	fsPk                    = flag.NewFlagSet("", flag.ContinueOnError)
	fsAmount                = flag.NewFlagSet("", flag.ContinueOnError)
	fsShares                = flag.NewFlagSet("", flag.ContinueOnError)
	fsDescriptionCreate     = flag.NewFlagSet("", flag.ContinueOnError)
	fsCommissionCreate      = flag.NewFlagSet("", flag.ContinueOnError)
	fsCommissionUpdate      = flag.NewFlagSet("", flag.ContinueOnError)
	fsDescriptionEdit       = flag.NewFlagSet("", flag.ContinueOnError)
	fsValidator             = flag.NewFlagSet("", flag.ContinueOnError)
	fsDelegator             = flag.NewFlagSet("", flag.ContinueOnError)
	fsRedelegation          = flag.NewFlagSet("", flag.ContinueOnError)
	fsSideChainFull         = flag.NewFlagSet("", flag.ContinueOnError)
	fsSideChainEdit         = flag.NewFlagSet("", flag.ContinueOnError)
	fsSideChainId           = flag.NewFlagSet("", flag.ContinueOnError)
	fsSmartChainValidator   = flag.NewFlagSet("", flag.ContinueOnError)
	fsSmartChainBeneficiary = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	fsPk.String(FlagPubKey, "", "Go-Amino encoded hex PubKey of the validator. For Ed25519 the go-amino prepend hex is 1624de6220")
	fsAmount.String(FlagAmount, "", "Amount of coins to bond")
	fsShares.String(FlagSharesAmount, "", "Amount of source-shares to either unbond or redelegate as a positive integer or decimal")
	fsShares.String(FlagSharesPercent, "", "Percent of source-shares to either unbond or redelegate as a positive integer or decimal >0 and <=1")
	fsDescriptionCreate.String(FlagMoniker, "", "Validator name")
	fsDescriptionCreate.String(FlagIdentity, "", "Optional identity signature (ex. UPort or Keybase)")
	fsDescriptionCreate.String(FlagWebsite, "", "Optional website")
	fsDescriptionCreate.String(FlagDetails, "", "Optional details")
	fsCommissionUpdate.String(FlagCommissionRate, "", "The new commission rate percentage")
	fsCommissionCreate.String(FlagCommissionRate, "", "The initial commission rate percentage")
	fsCommissionCreate.String(FlagCommissionMaxRate, "", "The maximum commission rate percentage")
	fsCommissionCreate.String(FlagCommissionMaxChangeRate, "", "The maximum commission change rate percentage (per day)")
	fsDescriptionEdit.String(FlagMoniker, types.DoNotModifyDesc, "Validator name")
	fsDescriptionEdit.String(FlagIdentity, types.DoNotModifyDesc, "Optional identity signature (ex. UPort or Keybase)")
	fsDescriptionEdit.String(FlagWebsite, types.DoNotModifyDesc, "Optional website")
	fsDescriptionEdit.String(FlagDetails, types.DoNotModifyDesc, "Optional details")
	fsValidator.String(FlagAddressValidator, "", "Bech address of the validator")
	fsDelegator.String(FlagAddressDelegator, "", "Bech address of the delegator")
	fsRedelegation.String(FlagAddressValidatorSrc, "", "Bech address of the source validator")
	fsRedelegation.String(FlagAddressValidatorDst, "", "Bech address of the destination validator")
	fsSideChainFull.String(FlagSideChainId, "", "Chain-id of the side chain the validator belongs to")
	fsSideChainFull.String(FlagSideConsAddr, "", "Consensus address of the validator on side chain, please use hex format prefixed with 0x")
	fsSideChainFull.String(FlagSideFeeAddr, "", "Address that validator collects fee rewards on side chain, please use hex format prefixed with 0x")
	fsSideChainFull.String(FlagSideVoteAddr, "", "BLS public key that validator votes for block on side chain, please use hex format prefixed with 0x")
	fsSideChainFull.String(FlagBLSWalletDir, "", "Absolute path of BLS wallet, should be provided if the side vote address is provided")
	fsSideChainFull.String(FlagBLSPassword, "", "Password for BLS wallet")
	fsSideChainEdit.String(FlagSideChainId, "", "Chain-id of the side chain the validator belongs to")
	fsSideChainEdit.String(FlagSideFeeAddr, "", "Address that validator collects fee rewards on side chain, please use hex format prefixed with 0x")
	fsSideChainEdit.String(FlagSideConsAddr, "", "consensus address of the validator on side chain, please use hex format prefixed with 0x")
	fsSideChainEdit.String(FlagSideVoteAddr, "", "BLS public key that validator votes for block on side chain, please use hex format prefixed with 0x")
	fsSideChainEdit.String(FlagBLSWalletDir, "", "Absolute path of BLS wallet, should be provided if the side vote address is provided")
	fsSideChainEdit.String(FlagBLSPassword, "", "Password for BLS wallet")
	fsSideChainId.String(FlagSideChainId, "", "Chain-id of the side chain the validator belongs to")
	fsSmartChainValidator.String(FlagAddressSmartChainValidator, "", "Smart chain operator address of the validator")
	fsSmartChainBeneficiary.String(FlagAddressSmartChainBeneficiary, "", "Smart chain address of the delegation's beneficiary")
}
