package util

import (
	"testing"

	"github.com/grafana-tools/sdk"
	"github.com/stretchr/testify/require"
)

func TestGetTargetsBasic(t *testing.T) {
	panel := &sdk.Panel{
		CommonPanel: sdk.CommonPanel{
			OfType: sdk.StatType,
		},
		StatPanel: &sdk.StatPanel{
			Targets: []sdk.Target{
				{},
				{},
			},
		},
	}
	targets, err := GetTargets(panel)
	require.NoError(t, err)
	require.NotEmpty(t, targets)
}
