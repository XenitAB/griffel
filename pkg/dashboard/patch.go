package dashboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/VictoriaMetrics/metricsql"
	"github.com/grafana-tools/sdk"
	"github.com/xenitab/griffel/pkg/util"
)

func patchTemplating(templating sdk.Templating, tplVars []sdk.TemplateVar, datasource *sdk.TemplateVar) (sdk.Templating, error) {
	ds := "${DS_PROMETHEUS}"
	newList := []sdk.TemplateVar{}
	if datasource != nil {
		newList = append(newList, *datasource)
	}
	newList = append(newList, tplVars...)

	// nolint:gocritic // can't affect SDK
	for _, template := range templating.List {
		// Remove datasource if datasource override is set
		if template.Type == "datasource" && datasource != nil {
			continue
		}
		template.Datasource = &ds
		if template.Type != "query" {
			newList = append(newList, template)
			continue
		}
		expr, err := appendVariables(template.Query.(string), tplVars)
		if err != nil {
			return sdk.Templating{}, err
		}
		template.Query = expr
		newList = append(newList, template)
	}
	templating.List = newList
	return templating, nil
}

func patchRows(rows []*sdk.Row, tplVars []sdk.TemplateVar) ([]*sdk.Row, error) {
	for i, row := range rows {
		panelPointers := util.PanelPointerSlice(row.Panels)
		panels, err := patchPanels(panelPointers, tplVars)
		if err != nil {
			return nil, err
		}
		rows[i].Panels = util.PanelSlice(panels)
	}
	return rows, nil
}

// nolint:gocognit // skip
func patchPanels(panels []*sdk.Panel, tplVars []sdk.TemplateVar) ([]*sdk.Panel, error) {
	for i, panel := range panels {
		// Override datasource for panel
		panels[i].Datasource = util.StringPointer("${DS_PROMETHEUS}")

		// If panel is a row iterate recurse through panels
		if panel.RowPanel != nil && len(panel.RowPanel.Panels) > 0 {
			panelPointers := util.PanelPointerSlice(panel.RowPanel.Panels)
			rowPanels, err := patchPanels(panelPointers, tplVars)
			if err != nil {
				return nil, err
			}
			panels[i].RowPanel.Panels = util.PanelSlice(rowPanels)
			continue
		}

		// Get all the targes (queries) for the panel
		targets, err := getTargets(panel)
		if err != nil {
			return nil, err
		}
		if targets == nil {
			continue
		}
		newTargets := []sdk.Target{}

		// Append variables to all of the targets
		// nolint:gocritic // can't affect SDK
		for _, target := range *targets {
			// Expr is only set for Prometheus targets
			if target.Expr == "" {
				continue
			}
			expr, err := appendVariables(target.Expr, tplVars)
			if err != nil {
				return nil, err
			}
			target.Expr = expr
			newTargets = append(newTargets, target)
		}

		// Write the targes back to the panels
		err = overrideTarget(panels[i], newTargets)
		if err != nil {
			return nil, err
		}
	}
	return panels, nil
}

func getTargets(panel *sdk.Panel) (*[]sdk.Target, error) {
	if panel.CustomPanel != nil {
		targets, err := getCustomTargets(panel.CustomPanel)
		if err != nil {
			return nil, err
		}
		return targets, nil
	}
	return panel.GetTargets(), nil
}

func getCustomTargets(customPanel *sdk.CustomPanel) (*[]sdk.Target, error) {
	// GetTargets() does not support custom panels so need to parse
	// TODO (Philip): There has to be a simpler way of casting to target list
	targetsMap, ok := (*customPanel)["targets"]
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

func overrideTarget(panel *sdk.Panel, targets []sdk.Target) error {
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

func appendVariables(exprStr string, tplVars []sdk.TemplateVar) (string, error) {
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

// These two functions are dirty but a quick work around until I can figure
// out how to parse promql that contains gragana variables in the interval.
// Variables in the label filter is ok because they look like normal strings.
// When used in the interval they are expected to be in a duration unit.

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

func unReplaceInterval(expr string, cache map[string]string) string {
	result := expr
	for k, v := range cache {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}
