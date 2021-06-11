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

func patchPanels(panels []*sdk.Panel, tplVars []sdk.TemplateVar) ([]*sdk.Panel, error) {
	ds := "${DS_PROMETHEUS}"
	for i, panel := range panels {
		panels[i].Datasource = &ds
		targets, err := getTargets(panel)
		if err != nil {
			return nil, err
		}
		if targets == nil {
			continue
		}
		for _, target := range *targets {
			// Skip if target is not Prometheus
			if target.Expr == "" {
				continue
			}
			expr, err := appendVariables(target.Expr, tplVars)
			if err != nil {
				return nil, err
			}
			// TODO (Philip): Skip updating if expression has not changed
			target.Expr = expr
			setTarget(panels[i], &target)
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

func setTarget(panel *sdk.Panel, target *sdk.Target) error {
	if panel.CustomPanel != nil {
		targets, err := getCustomTargets(panel.CustomPanel)
		if err != nil {
			return err
		}
		setTarget := func(t *sdk.Target, targets *[]sdk.Target) {
			for i, target := range *targets {
				if t.RefID == target.RefID {
					(*targets)[i] = *t
					return
				}
			}
			(*targets) = append((*targets), *t)
		}
		setTarget(target, targets)
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
	panel.SetTarget(target)
	return nil
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

// These two functions are dirty but a quick work around until I can figure
// out how to parse promql that contains gragana variables in the interval.
// Variables in the label filter is ok because they look like normal strings.
// When used in the interval they are expected to be in a duration unit.

func replaceInterval(expr string) (string, map[string]string) {
	betweenRegex := regexp.MustCompile(`\[(.*?)\]`)
	between := betweenRegex.FindString(expr)

	variableRegex := regexp.MustCompile(`\$[a-z]+`)
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
