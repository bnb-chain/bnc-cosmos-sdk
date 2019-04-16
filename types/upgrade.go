package types

import "fmt"

const (
	AddDelegationAccountAddr      = "DelegationAccountAddr"
	AddAnteCache                  = "AnteCache"
	AddCreateValidatorMsgValidate = "CreateValidatorMsgValidate"
	ChangeGovFeeAddress           = "ChangeGovFeeAddress"
)

var UpgradeMgr = NewUpgradeManager(UpgradeConfig{
	HeightMap: map[string]int64{
		AddDelegationAccountAddr:      100,
		AddAnteCache:                  100,
		AddCreateValidatorMsgValidate: 100,
		ChangeGovFeeAddress:           100,
	},
})

var MainNetConfig = UpgradeConfig{
	HeightMap: map[string]int64{},
}

type UpgradeConfig struct {
	HeightMap     map[string]int64
	BeginBlockers map[int64][]func(ctx Context)
}

type UpgradeManager struct {
	Config UpgradeConfig
	Height int64
}

func NewUpgradeManager(config UpgradeConfig) *UpgradeManager {
	return &UpgradeManager{
		Config: config,
	}
}

func (mgr *UpgradeManager) AddConfig(config UpgradeConfig) {
	for name, height := range config.HeightMap {
		mgr.AddUpgradeHeight(name, height)
	}
}

func (mgr *UpgradeManager) SetHeight(height int64) {
	mgr.Height = height
}

func (mgr *UpgradeManager) GetHeight() int64 {
	return mgr.Height
}

// run in every ABCI BeginBlock.
func (mgr *UpgradeManager) BeginBlocker(ctx Context) {
	if beginBlockers, ok := mgr.Config.BeginBlockers[mgr.GetHeight()]; ok {
		for _, beginBlocker := range beginBlockers {
			beginBlocker(ctx)
		}
	}
}

func (mgr *UpgradeManager) RegisterBeginBlocker(name string, beginBlocker func(Context)) {
	height := mgr.GetUpgradeHeight(name)
	if height == 0 {
		panic(fmt.Errorf("no UpgradeHeight found for %s", name))
	}

	if mgr.Config.BeginBlockers == nil {
		mgr.Config.BeginBlockers = make(map[int64][]func(ctx Context))
	}

	if beginBlockers, ok := mgr.Config.BeginBlockers[height]; ok {
		beginBlockers = append(beginBlockers, beginBlocker)
		mgr.Config.BeginBlockers[height] = beginBlockers
	} else {
		mgr.Config.BeginBlockers[height] = []func(Context){beginBlocker}
	}
}

func (mgr *UpgradeManager) AddUpgradeHeight(name string, height int64) {
	if mgr.Config.HeightMap == nil {
		mgr.Config.HeightMap = map[string]int64{}
	}

	mgr.Config.HeightMap[name] = height
}

func (mgr *UpgradeManager) GetUpgradeHeight(name string) int64 {
	if mgr.Config.HeightMap == nil {
		return 0
	}
	return mgr.Config.HeightMap[name]
}

func IsUpgradeHeight(name string) bool {
	upgradeHeight := UpgradeMgr.GetUpgradeHeight(name)
	if upgradeHeight == 0 {
		return false
	}

	return upgradeHeight == UpgradeMgr.GetHeight()
}

func IsUpgrade(name string) bool {
	upgradeHeight := UpgradeMgr.GetUpgradeHeight(name)
	if upgradeHeight == 0 {
		return false
	}

	return UpgradeMgr.GetHeight() >= upgradeHeight
}

func Upgrade(name string, before func(), in func(), after func()) {
	// if no special logic for the UpgradeHeight, than apply the `after` logic
	if in == nil {
		in = after
	}

	if IsUpgradeHeight(name) {
		if in != nil {
			in()
		}
	} else if IsUpgrade(name) {
		if after != nil {
			after()
		}
	} else {
		if before != nil {
			before()
		}
	}
}

func FixAddDelegationAccountAddr(before func(), after func()) {
	Upgrade(AddDelegationAccountAddr, before, nil, after)
}

func FixAddAnteCache(before func(), after func()) {
	Upgrade(AddAnteCache, before, nil, after)
}

func FixAddCreateValidatorMsgValidate(before func(), after func()) {
	Upgrade(AddCreateValidatorMsgValidate, before, nil, after)
}
