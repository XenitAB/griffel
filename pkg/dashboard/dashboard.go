package dashboard

import (
	"fmt"

	"github.com/grafana-tools/sdk"
	"github.com/spf13/afero"

	"github.com/xenitab/griffel/pkg/config"
)

func Patch(fs afero.Fs, cfg *config.Config) error {
	refresh := int64(1)
	ds := "${DS_PROMETHEUS}"

	tplVars := []sdk.TemplateVar{}
	for _, template := range cfg.Patch.Variables {
		tplVars = append(tplVars, sdk.TemplateVar{
			Name:       template.Name,
			Label:      template.Label,
			IncludeAll: true,
			Refresh:    sdk.BoolInt{Value: &refresh},
			Datasource: &ds,
			Type:       "query",
			Query:      template.Query,
		})
	}

	var datasource *sdk.TemplateVar
	if cfg.Patch.Datasource != nil {
		hide := uint8(0)
		if cfg.Patch.Datasource.Hide {
			hide = 2
		}
		datasource = &sdk.TemplateVar{
			Name:    cfg.Patch.Datasource.Name,
			Label:   cfg.Patch.Datasource.Label,
			Type:    "datasource",
			Query:   "prometheus",
			Regex:   cfg.Patch.Datasource.Regex,
			Hide:    hide,
			Refresh: sdk.BoolInt{Value: &refresh},
		}
	}

	for _, dash := range cfg.Dashboards {
		fmt.Println(dash.Name)

		board, err := readDashboard(dash.Source.Kind, dash.Source.Value)
		if err != nil {
			return err
		}

		templating, err := patchTemplating(board.Templating, tplVars, datasource)
		if err != nil {
			return err
		}
		board.Templating = templating

		panels, err := patchPanels(board.Panels, tplVars)
		if err != nil {
			return err
		}
		board.Panels = panels

		rows, err := patchRows(board.Rows, tplVars)
		if err != nil {
			return err
		}
		board.Rows = rows

		err = writeDashboard(fs, dash.Destination.Directory, dash.Destination.Format, dash.Name, board)
		if err != nil {
			return err
		}
	}

	return nil
}
