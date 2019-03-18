package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsLimitAddressLengthFork(t *testing.T) {
	config := UpgradeConfig{
		map[string]int64{
			UpgradeLimitAddressLength: 545000,
		},
	}
	UpgradeMgr = NewUpgradeManager(config)

	type testCase struct {
		config        UpgradeConfig
		height        int64
		upgradeResult bool
		heightResult  bool
	}

	testCases := []testCase{
		{
			config: UpgradeConfig{
				map[string]int64{},
			},
			height:        10000,
			upgradeResult: false,
			heightResult:  false,
		},
		{
			config: UpgradeConfig{
				map[string]int64{
					UpgradeLimitAddressLength: 545000,
				},
			},
			height:        10000,
			upgradeResult: false,
			heightResult:  false,
		}, {
			config: UpgradeConfig{
				map[string]int64{
					UpgradeLimitAddressLength: 545000,
				},
			},
			height:        545000,
			upgradeResult: true,
			heightResult:  true,
		}, {
			config: UpgradeConfig{
				map[string]int64{
					UpgradeLimitAddressLength: 545000,
				},
			},
			height:        545001,
			upgradeResult: true,
			heightResult:  false,
		},
	}

	for _, tc := range testCases {
		UpgradeMgr.SetHeight(tc.height)
		require.Equal(t, tc.upgradeResult, IsLimitAddressLengthUpgrade())
		require.Equal(t, tc.upgradeResult, IsUpgrade(UpgradeLimitAddressLength))
		require.Equal(t, tc.heightResult, IsUpgradeHeight(UpgradeLimitAddressLength))
	}
}
