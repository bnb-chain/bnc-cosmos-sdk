package types

var UpgradeMgr = NewUpgradeManager(UpgradeConfig{})

const UpgradeLimitAddressLength = "UpgradeLimitAddressLength" // limit address length to 20 bytes
const UpgradeRunTx = "UpgradeRunTx" // Add cache for ante execution

var MainNetConfig = UpgradeConfig{
	map[string]int64{
		UpgradeLimitAddressLength: 554000,
		UpgradeRunTx: 100,
	},
}

type UpgradeConfig struct {
	HeightMap map[string]int64
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

func IsLimitAddressLengthUpgrade() bool {
	return IsUpgrade(UpgradeLimitAddressLength)
}
