package dashboard

import (
	"fmt"

	"github.com/grafana-tools/sdk"

	"github.com/xenitab/griffel/pkg/util"
)

// patchTemplating updates the dashboards template variables and adds label filters.
// It will create a new list of template variables that contains the new tplVars and
// the existing variables with the new variables added as a label filter.
// If datasource is set it will override the existing datasource variable.
func patchTemplating(templating sdk.Templating, tplVars []sdk.TemplateVar, datasource *sdk.TemplateVar) (sdk.Templating, error) {
	// Create a new template variable list
	newList := []sdk.TemplateVar{}
	if datasource != nil {
		newList = append(newList, *datasource)
	}
	newList = append(newList, tplVars...)

	// Append exsting template variables with additional label filters
	// nolint:gocritic // can't affect SDK
	for _, template := range templating.List {
		// Set values that cant be nil
		if template.Options == nil {
			template.Options = []sdk.Option{}
		}

		// Remove datasource if datasource override is set
		if template.Type == "datasource" && datasource != nil {
			continue
		}
		template.Datasource = util.StringPointer("${DS_PROMETHEUS}")
		if template.Type != "query" {
			newList = append(newList, template)
			continue
		}

		// Append additional variable filters to query
		query, err := util.AppendVariables(template.Query, tplVars)
		if err != nil {
			return sdk.Templating{}, fmt.Errorf("could not template Grafana query: %w", err)
		}
		template.Query = query

		newList = append(newList, template)
	}

	// Set the new template variable list
	templating.List = newList
	return templating, nil
}

// patchRows iterates through the panels in the rows and patches them.
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

// patchPanels patches all queries in the list of panels and adds label filters to all of the queries.
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
		targets, err := util.GetTargets(panel)
		if err != nil {
			return nil, err
		}
		if targets == nil {
			continue
		}

		// Append variables to all of the targets
		newTargets := []sdk.Target{}
		// nolint:gocritic // can't affect SDK
		for _, target := range *targets {
			// Expr is only set for Prometheus targets
			if target.Expr == "" {
				continue
			}
			expr, err := util.AppendVariables(target.Expr, tplVars)
			if err != nil {
				return nil, err
			}
			exprStr, ok := expr.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected expression return type")
			}
			target.Expr = exprStr
			newTargets = append(newTargets, target)
		}

		// Write the targes back to the panels
		err = util.OverrideTarget(panels[i], newTargets)
		if err != nil {
			return nil, err
		}
	}
	return panels, nil
}
