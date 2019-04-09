package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const UpgradeTest = "upgradeTest"

func TestUpgrade(t *testing.T) {
	UpgradeMgr = NewUpgradeManager(UpgradeConfig{})

	type testCase struct {
		config        UpgradeConfig
		height        int64
		upgradeResult bool
		heightResult  bool
	}

	testCases := []testCase{
		{
			config: UpgradeConfig{
				HeightMap: map[string]int64{},
			},
			height:        10000,
			upgradeResult: false,
			heightResult:  false,
		},
		{
			config: UpgradeConfig{
				HeightMap: map[string]int64{
					UpgradeTest: 545000,
				},
			},
			height:        10000,
			upgradeResult: false,
			heightResult:  false,
		}, {
			config: UpgradeConfig{
				HeightMap: map[string]int64{
					UpgradeTest: 545000,
				},
			},
			height:        545000,
			upgradeResult: true,
			heightResult:  true,
		}, {
			config: UpgradeConfig{
				HeightMap: map[string]int64{
					UpgradeTest: 545000,
				},
			},
			height:        545001,
			upgradeResult: true,
			heightResult:  false,
		},
	}

	for _, tc := range testCases {
		UpgradeMgr.AddConfig(tc.config)
		UpgradeMgr.SetHeight(tc.height)
		require.Equal(t, tc.upgradeResult, IsUpgrade(UpgradeTest))
		require.Equal(t, tc.heightResult, IsUpgradeHeight(UpgradeTest))
	}
}
