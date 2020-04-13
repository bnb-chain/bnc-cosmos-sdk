package types

const (
	EventTypeCompleteUnbonding              = "complete_unbonding"
	EventTypeCompleteRedelegation           = "complete_redelegation"
	EventTypeCreateValidator                = "create_validator"
	EventTypeEditValidator                  = "edit_validator"
	EventTypeDelegate                       = "delegate"
	EventTypeUnbond                         = "unbond"
	EventTypeRedelegate                     = "redelegate"
	EventTypeSaveValidatorUpdatesIbcPackage = "save_val_updates_ibc_package"

	AttributeKeyValidator         = "validator"
	AttributeKeyCommissionRate    = "commission_rate"
	AttributeKeyMinSelfDelegation = "min_self_delegation"
	AttributeKeySrcValidator      = "source_validator"
	AttributeKeyDstValidator      = "destination_validator"
	AttributeKeyDelegator         = "delegator"
	AttributeKeyCompletionTime    = "completion_time"

	AttributeKeySideChainId                 = "side_chain_id"
	AttributeKeyValidatorUpdatesIbcSequence = "validator_updates_ibc_sequence"
)
