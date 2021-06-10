package dashboard

import (
	"fmt"

	"github.com/VictoriaMetrics/metricsql"
	"github.com/grafana-tools/sdk"

	"github.com/xenitab/griffel/pkg/config"
)

func Patch(cfg *config.Config) error {
	tplVars, labelFilters := convertToPatch(cfg.Patch)
	for _, dash := range cfg.Dashboards {
		board, err := readDashboard(dash.Source.Kind, dash.Source.Value)
		if err != nil {
			return err
		}

		templating, err := patchTemplating(board.Templating, tplVars, labelFilters)
		if err != nil {
			return err
		}
		board.Templating = templating

		panels, err := patchPanels(board.Panels, labelFilters)
		if err != nil {
			return err
		}
		board.Panels = panels

		err = writeDashboard(dash.Destination.Path, dash.Destination.Format, board)
		if err != nil {
			return err
		}
	}

	return nil
}

func convertToPatch(patch config.DashboardPatch) ([]sdk.TemplateVar, []metricsql.LabelFilter) {
	tplVars := []sdk.TemplateVar{}
	labelFilters := []metricsql.LabelFilter{}
	for _, template := range patch.Variables {
		tplVars = append(tplVars, sdk.TemplateVar{
			Name:       template.Name,
			Label:      template.Label,
			IncludeAll: true,
		})
		labelFilters = append(labelFilters, metricsql.LabelFilter{
			Label:    template.Name, // no this is not wrong
			Value:    fmt.Sprintf("${%s}", template.Name),
			IsRegexp: true,
		})
	}
	return tplVars, labelFilters
}
