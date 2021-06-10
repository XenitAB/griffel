package dashboard

import (
	"github.com/grafana-tools/sdk"

	"github.com/xenitab/griffel/pkg/config"
)

func Patch(cfg *config.Config) error {
	tplVars := []sdk.TemplateVar{}
	for _, template := range cfg.Patch.Variables {
		tplVars = append(tplVars, sdk.TemplateVar{
			Name:       template.Name,
			Label:      template.Label,
			IncludeAll: true,
		})
	}

	for _, dash := range cfg.Dashboards {
		board, err := readDashboard(dash.Source.Kind, dash.Source.Value)
		if err != nil {
			return err
		}

		templating, err := patchTemplating(board.Templating, tplVars)
		if err != nil {
			return err
		}
		board.Templating = templating

		panels, err := patchPanels(board.Panels, tplVars)
		if err != nil {
			return err
		}
		board.Panels = panels

		err = writeDashboard(dash.Destination.Format, dash.Destination.Path, board)
		if err != nil {
			return err
		}
	}

	return nil
}
