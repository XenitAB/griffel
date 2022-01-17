package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/VictoriaMetrics/metricsql"
	"github.com/grafana-tools/sdk"
)

func ConvertToStatPanel(panel *sdk.Panel) (*sdk.Panel, error) {
	if panel.OfType != sdk.SinglestatType {
		return nil, errors.New("can only convert panels of type single stat type")
	}

	newPanel := sdk.NewStat(panel.Title)
	newPanel.ID = panel.ID
	newPanel.Description = panel.Description
	newPanel.Datasource = panel.Datasource
	newPanel.Editable = panel.Editable
	newPanel.Error = panel.Error
	newPanel.GridPos = panel.GridPos
	newPanel.Height = panel.Height
	newPanel.HideTimeOverride = panel.HideTimeOverride
	newPanel.IsNew = panel.IsNew
	newPanel.Links = panel.Links
	newPanel.Renderer = panel.Renderer
	newPanel.Repeat = panel.Repeat
	newPanel.RepeatPanelID = panel.RepeatPanelID
	newPanel.StatPanel = &sdk.StatPanel{
		Colors:          panel.SinglestatPanel.Colors,
		ColorValue:      panel.SinglestatPanel.ColorValue,
		ColorBackground: panel.SinglestatPanel.ColorBackground,
		Decimals:        panel.SinglestatPanel.Decimals,
		Format:          panel.SinglestatPanel.Format,
		Gauge:           panel.SinglestatPanel.Gauge,
		MappingType:     panel.SinglestatPanel.MappingType,
		MappingTypes:    panel.SinglestatPanel.MappingTypes,
		MaxDataPoints:   panel.SinglestatPanel.MaxDataPoints,
		NullPointMode:   panel.SinglestatPanel.NullPointMode,
		Postfix:         panel.SinglestatPanel.Postfix,
		PostfixFontSize: panel.SinglestatPanel.PostfixFontSize,
		Prefix:          panel.SinglestatPanel.Postfix,
		PrefixFontSize:  panel.SinglestatPanel.PrefixFontSize,
		RangeMaps:       panel.SinglestatPanel.RangeMaps,
		SparkLine:       panel.SinglestatPanel.SparkLine,
		Targets:         panel.SinglestatPanel.Targets,
		Thresholds:      panel.SinglestatPanel.Thresholds,
		ValueFontSize:   panel.SinglestatPanel.ValueFontSize,
		ValueMaps:       panel.SinglestatPanel.ValueMaps,
		ValueName:       panel.SinglestatPanel.ValueName,
	}
	return newPanel, nil
}

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

func AppendVariables(exprStr string, tplVars []sdk.TemplateVar) (string, error) {
	if exprStr == "" {
		return "", errors.New("expression string cannot be empty")
	}
	replacedExprStr, replaceCache := replaceInterval(exprStr)
	expr, err := metricsql.Parse(replacedExprStr)
	if err != nil {
		return "", fmt.Errorf("could not parse promql: %w", err)
	}

	labelFilters := []metricsql.LabelFilter{}
	// nolint:gocritic // can't affect SDK
	for _, tplVar := range tplVars {
		labelFilters = append(labelFilters, metricsql.LabelFilter{
			Label:    tplVar.Name, // no this is not wrong
			Value:    fmt.Sprintf("$%s", tplVar.Name),
			IsRegexp: true,
		})
	}

	// TODO (Philip):
	// Need to handle label_values as a special case as
	// to not apply label filter to all parameters
	if strings.HasPrefix(replacedExprStr, "label_values(") {
		funcExpr, ok := expr.(*metricsql.FuncExpr)
		if !ok {
			return "", errors.New("expr is not a function")
		}
		if len(funcExpr.Args) == 0 {
			return "", errors.New("label_value cannot have zero arguments")
		}
		if len(funcExpr.Args) == 1 {
			return exprStr, nil
		}
		metricExpr, ok := funcExpr.Args[0].(*metricsql.MetricExpr)
		if !ok {
			return "", errors.New("expr is not a metrics expression")
		}
		metricExpr.LabelFilters = append(metricExpr.LabelFilters, labelFilters...)
		return string(expr.AppendString(nil)), nil
	}

	metricsql.VisitAll(expr, func(expr metricsql.Expr) {
		metricExpr, ok := expr.(*metricsql.MetricExpr)
		if metricExpr != nil && ok {
			metricExpr.LabelFilters = append(metricExpr.LabelFilters, labelFilters...)
		}
	})
	result := unReplaceInterval(string(expr.AppendString(nil)), replaceCache)
	return result, nil
}

/*
These two functions are dirty but a quick work around until I can figure
out how to parse promql that contains gragana variables in the interval.
Variables in the label filter is ok because they look like normal strings.
When used in the interval they are expected to be in a duration unit.
*/

// replaceInterval replaces interval values with actual real values.
// It also keeps a cache to that the real values can be "unreplaced"
// in the future.
func replaceInterval(expr string) (string, map[string]string) {
	betweenRegex := regexp.MustCompile(`\[(.*?)\]`)
	between := betweenRegex.FindString(expr)

	variableRegex := regexp.MustCompile(`\$[a-z_]+`)
	variables := variableRegex.FindAllString(between, -1)

	cache := map[string]string{}
	result := expr
	for _, v := range variables {
		// nolint:gosec // security is not an issue for this as the value is temporary
		duration := fmt.Sprintf("%vm", rand.Int())
		cache[duration] = v
		result = strings.ReplaceAll(result, v, duration)
	}

	return result, cache
}

// unReplaceInterval puts the variable back in place based on the
// cache content.
func unReplaceInterval(expr string, cache map[string]string) string {
	result := expr
	for k, v := range cache {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}
