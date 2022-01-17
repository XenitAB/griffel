package util

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana-tools/sdk"
)

// GetTargets returns a list of targets for the given panel.
// Custom logic has to be implemented for panels of type custom
// as the sdk implementation does not support it.
func GetTargets(panel *sdk.Panel) (*[]sdk.Target, error) {
	if panel.CustomPanel == nil {
		return panel.GetTargets(), nil
	}

	// TODO (Philip): There has to be a simpler way of casting to target list
	targetsMap, ok := (*panel.CustomPanel)["targets"]
	if !ok {
		return nil, errors.New("targets not found in custom panel")
	}
	b, err := json.Marshal(targetsMap)
	if err != nil {
		return nil, fmt.Errorf("could not marshal custom panel targets: %w", err)
	}
	targets := &[]sdk.Target{}
	err = json.Unmarshal(b, targets)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal to target: %w", err)
	}
	return targets, nil
}

// OverrideTarget replaces the targets in the panel.
// This has to be implemented per panel type as the sdk method does not
// work properly for all the different types.
func OverrideTarget(panel *sdk.Panel, targets []sdk.Target) error {
	// nolint:exhaustive // only want to override panels with targets
	switch panel.OfType {
	case sdk.CustomType:
		b, err := json.Marshal(targets)
		if err != nil {
			return err
		}
		targetsMap := &[]map[string]interface{}{}
		err = json.Unmarshal(b, targetsMap)
		if err != nil {
			return err
		}
		customPanel := *panel.CustomPanel
		customPanel["targets"] = *targetsMap
		panel.CustomPanel = &customPanel
	case sdk.GraphType:
		panel.GraphPanel.Targets = targets
	case sdk.SinglestatType:
		panel.SinglestatPanel.Targets = targets
	case sdk.StatType:
		panel.StatPanel.Targets = targets
	case sdk.TableType:
		panel.TablePanel.Targets = targets
	case sdk.BarGaugeType:
		panel.BarGaugePanel.Targets = targets
	case sdk.HeatmapType:
		panel.HeatmapPanel.Targets = targets
	default:
		break
	}
	return nil
}
