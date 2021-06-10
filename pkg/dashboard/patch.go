package dashboard

import (
	"errors"
	"strings"

	"github.com/VictoriaMetrics/metricsql"
	"github.com/grafana-tools/sdk"
)

func patchTemplating(templating sdk.Templating, tplVars []sdk.TemplateVar, labelFilters []metricsql.LabelFilter) (sdk.Templating, error) {
	for i, template := range templating.List {
		expr, err := appendLabelFilter(template.Query, labelFilters)
		if err != nil {
			return sdk.Templating{}, err
		}
		templating.List[i].Query = expr
	}
	templating.List = append(templating.List, tplVars...)
	return templating, nil
}

func patchPanels(panels []*sdk.Panel, labelFilters []metricsql.LabelFilter) ([]*sdk.Panel, error) {
	for i, panel := range panels {
		targets := panel.GetTargets()
		if targets == nil {
			continue
		}
		for _, target := range *targets {
			expr, err := appendLabelFilter(target.Expr, labelFilters)
			if err != nil {
				return nil, err
			}
			// TODO (Philip): Skip updating if expression has not changed
			target.Expr = expr
			panels[i].SetTarget(&target)
		}
	}
	return panels, nil
}

func appendLabelFilter(exprStr string, labelFilters []metricsql.LabelFilter) (string, error) {
	expr, err := metricsql.Parse(exprStr)
	if err != nil {
		return "", err
	}

	// Need to handle label_values as a special case as
	// to not apply label filter to all parameters
	if strings.HasPrefix(exprStr, "label_values(") {
		funcExpr := expr.(*metricsql.FuncExpr)
		if len(funcExpr.Args) != 2 {
			return "", errors.New("label_value does not have two parameters")
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
	return string(expr.AppendString(nil)), nil
}
