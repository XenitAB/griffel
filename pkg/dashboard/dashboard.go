package dashboard

import (
	"fmt"

	"github.com/grafana-tools/sdk"
	"github.com/spf13/afero"

	"github.com/xenitab/griffel/pkg/config"
	"github.com/xenitab/griffel/pkg/util"
)

func Patch(fs afero.Fs, cfg *config.Config) error {
	tplVars := []sdk.TemplateVar{}
	// nolint:gocritic // skip
	for _, template := range cfg.Patch.Variables {
		tplVars = append(tplVars, sdk.TemplateVar{
			Name:       template.Name,
			Label:      template.Label,
			IncludeAll: true,
			Refresh:    sdk.BoolInt{Value: util.Int64Pointer(1)},
			Datasource: util.StringPointer("${DS_PROMETHEUS}"),
			Type:       "query",
			Query:      template.Query,
			Options:    []sdk.Option{},
			Current: sdk.Current{
				Text: &sdk.StringSliceString{
					Value: []string{""},
					Valid: true,
				},
				Value: "",
			},
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
			Refresh: sdk.BoolInt{Value: util.Int64Pointer(1)},
			Options: []sdk.Option{},
			Current: sdk.Current{
				Text: &sdk.StringSliceString{
					Value: []string{""},
					Valid: true,
				},
				Value: "",
			},
		}
	}

	// nolint:gocritic // skip
	for _, dash := range cfg.Dashboards {
		fmt.Println(dash.Name)

		board, err := readDashboard(dash.Source.Kind, dash.Source.Value)
		if err != nil {
			return err
		}
		board.Editable = dash.Patch.Editable
		if dash.Patch.Title != "" {
			board.Title = dash.Patch.Title
		}
		if len(dash.Patch.Tags) > 0 {
			board.Tags = dash.Patch.Tags
		}
		board.Templating, err = patchTemplating(board.Templating, tplVars, datasource)
		if err != nil {
			return err
		}
		board.Panels, err = patchPanels(board.Panels, tplVars)
		if err != nil {
			return err
		}
		board.Rows, err = patchRows(board.Rows, tplVars)
		if err != nil {
			return err
		}
		err = writeDashboard(fs, dash.Destination.Directory, dash.Destination.Format, dash.Name, board)
		if err != nil {
			return err
		}
	}
	return nil
}
