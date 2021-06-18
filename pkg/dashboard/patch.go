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
)

func patchTemplating(templating sdk.Templating, tplVars []sdk.TemplateVar, datasource *sdk.TemplateVar) (sdk.Templating, error) {
	ds := "${DS_PROMETHEUS}"
	newList := []sdk.TemplateVar{}
	if datasource != nil {
		newList = append(newList, *datasource)
	}
	newList = append(newList, tplVars...)

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
		expr, err := appendVariables(template.Query, tplVars)
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
		panels, err := patchPanels(toPanelPointerSlice(row.Panels), tplVars)
		if err != nil {
			return nil, err
		}
		rows[i].Panels = fromPanelPointerSlice(panels)
	}
	return rows, nil
}

func patchPanels(panels []*sdk.Panel, tplVars []sdk.TemplateVar) ([]*sdk.Panel, error) {
	ds := "${DS_PROMETHEUS}"
	for i, panel := range panels {
		panels[i].Datasource = &ds

		if panel.RowPanel != nil && len(panel.RowPanel.Panels) > 0 {
			rowPanels, err := patchPanels(toPanelPointerSlice(panel.RowPanel.Panels), tplVars)
			if err != nil {
				return nil, err
			}
			panels[i].RowPanel.Panels = fromPanelPointerSlice(rowPanels)
			continue
		}

		targets, err := getTargets(panel)
		if err != nil {
			return nil, err
		}
		if targets == nil {
			continue
		}
		newTargets := []sdk.Target{}
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
		overrideTarget(panels[i], newTargets)
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
	if panel.CustomPanel != nil {
		b, err := json.Marshal(targets)
		if err != nil {
			return err
		}
		targetsMap := &[]map[string]interface{}{}
		err = json.Unmarshal(b, targetsMap)
		if err != nil {
			return err
		}
		(*panel.CustomPanel)["targets"] = *targetsMap
		return nil
	}

	switch panel.OfType {
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
	}
	return nil
}

func appendVariables(exprStr string, tplVars []sdk.TemplateVar) (string, error) {
	if exprStr == "" {
		return "", errors.New("expression string cannot be empty")
	}
	exprStr, replaceCache := replaceInterval(exprStr)
	expr, err := metricsql.Parse(exprStr)
	if err != nil {
		return "", fmt.Errorf("could not parse promql: %w", err)
	}

	labelFilters := []metricsql.LabelFilter{}
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
	if strings.HasPrefix(exprStr, "label_values(") {
		funcExpr := expr.(*metricsql.FuncExpr)
		if len(funcExpr.Args) == 0 {
			return "", errors.New("label_value cannot have zero arguments")
		}
		metricExpr := funcExpr.Args[0].(*metricsql.MetricExpr)
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

func toPanelPointerSlice(panels []sdk.Panel) []*sdk.Panel {
	newPanels := []*sdk.Panel{}
	for i := range panels {
		newPanels = append(newPanels, &panels[i])
	}
	return newPanels
}

func fromPanelPointerSlice(panels []*sdk.Panel) []sdk.Panel {
	newPanels := []sdk.Panel{}
	for _, p := range panels {
		newPanels = append(newPanels, *p)
	}
	return newPanels
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
