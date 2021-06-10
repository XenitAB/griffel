package dashboard

import (
	"github.com/grafana-tools/sdk"
	"github.com/spf13/afero"

	"github.com/xenitab/griffel/pkg/config"
)

func Patch(fs afero.Fs, cfg *config.Config) error {
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

		err = writeDashboard(fs, dash.Destination.Directory, dash.Destination.Format, dash.Name, board)
		if err != nil {
			return err
		}
	}

	return nil
}
