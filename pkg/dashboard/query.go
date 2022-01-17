package dashboard

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/VictoriaMetrics/metricsql"
	"github.com/grafana-tools/sdk"
)

type Query struct {
	queryString string
	queryMap    map[string]interface{}
}

func NewQuery(data interface{}) (*Query, error) {
	// A query can either be a string or a map containing the query string
	if queryString, ok := data.(string); ok {
		return &Query{queryString: queryString}, nil
	}
	if queryMap, ok := data.(map[string]interface{}); ok {
		return &Query{queryMap: queryMap, queryString: queryMap["query"].(string)}, nil
	}
	return nil, fmt.Errorf("unknown query data format")
}

func (q *Query) AppendVariables(tplVars []sdk.TemplateVar) (interface{}, error) {
	// Temporarly replace any interval variables in the query
	replacedExprStr, replaceCache := replaceInterval(q.queryString)

	expr, err := metricsql.Parse(replacedExprStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse promql: %w", err)
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
			return nil, errors.New("expr is not a function")
		}
		if len(funcExpr.Args) == 0 {
			return nil, errors.New("label_value cannot have zero arguments")
		}
		// Return early if there are no labels specified as the second argument
		if len(funcExpr.Args) == 1 {
			return q.getResult(q.queryString), nil
		}
		metricExpr, ok := funcExpr.Args[0].(*metricsql.MetricExpr)
		if !ok {
			return nil, errors.New("expr is not a metrics expression")
		}
		metricExpr.LabelFilters = append(metricExpr.LabelFilters, labelFilters...)
		return q.getResult(string(expr.AppendString(nil))), nil
	}

	metricsql.VisitAll(expr, func(expr metricsql.Expr) {
		metricExpr, ok := expr.(*metricsql.MetricExpr)
		if metricExpr != nil && ok {
			metricExpr.LabelFilters = append(metricExpr.LabelFilters, labelFilters...)
		}
	})
	result := unReplaceInterval(string(expr.AppendString(nil)), replaceCache)
	return q.getResult(result), nil
}

func (q *Query) getResult(input string) interface{} {
	if q.queryMap != nil {
		mapCopy := q.queryMap
		mapCopy["query"] = input
		return mapCopy
	}
	return input
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
