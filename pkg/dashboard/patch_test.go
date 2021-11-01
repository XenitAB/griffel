package dashboard

import (
	"testing"

	"github.com/grafana-tools/sdk"
	"github.com/stretchr/testify/require"
)

func TestPatchTemplating(t *testing.T) {
	templating := sdk.Templating{
		List: []sdk.TemplateVar{
			{
				Type:  "datasource",
				Name:  "DS_PROMETHEUS",
				Label: "Datasource",
			},
			{
				Type:  "query",
				Name:  "foo",
				Label: "Foo",
				Query: "query_1",
			},
		},
	}
	tplVars := []sdk.TemplateVar{
		{
			Name:  "bar",
			Label: "Bar",
			Query: "query_2",
		},
	}
	newTemplating, err := patchTemplating(templating, tplVars, nil)
	require.NoError(t, err)
	require.Len(t, newTemplating.List, 3)
	require.Equal(t, "bar", newTemplating.List[0].Name)
	require.Equal(t, "query_2", newTemplating.List[0].Query)
	require.Equal(t, "DS_PROMETHEUS", newTemplating.List[1].Name)
	require.Nil(t, newTemplating.List[1].Query)
	require.Equal(t, "foo", newTemplating.List[2].Name)
	require.Equal(t, "query_1{bar=~\"$bar\"}", newTemplating.List[2].Query)
}

func TestPatchPanel(t *testing.T) {
	graphPanel := sdk.NewGraph("graph")
	graphPanel.SetTarget(&sdk.Target{
		RefID:        "A",
		Expr:         "sum(keycloak_registrations{instance=\"$instance\"})",
		LegendFormat: "foo",
		Format:       "time_series",
	})
	graphPanel.SetTarget(&sdk.Target{
		RefID:        "B",
		Expr:         "sum(keycloak_registrations{instance=\"$instance\"})",
		LegendFormat: "bar",
		Format:       "time_series",
	})
	graphPanel.SetTarget(&sdk.Target{
		RefID:        "C",
		Expr:         "sum(keycloak_registrations{instance=\"$instance\"})",
		LegendFormat: "baz",
		Format:       "time_series",
	})
	customPanel := sdk.NewCustom("custom")
	(*customPanel.CustomPanel)["targets"] = []map[string]string{
		{
			"refId": "A",
			"expr":  "sum(keycloak_registrations{instance=\"$instance\"})",
		},
	}
	panels := []*sdk.Panel{graphPanel, customPanel}
	tplVars := []sdk.TemplateVar{
		{
			Name:  "foo",
			Label: "Foo",
		},
	}

	newPanels, err := patchPanels(panels, tplVars)
	require.NoError(t, err)
	require.Equal(t, "A", newPanels[0].GraphPanel.Targets[0].RefID)
	require.Equal(t, "foo", newPanels[0].GraphPanel.Targets[0].LegendFormat)
	require.Equal(t, "sum(keycloak_registrations{instance=\"$instance\", foo=~\"$foo\"})", newPanels[0].GraphPanel.Targets[0].Expr)
	require.Equal(t, "B", newPanels[0].GraphPanel.Targets[1].RefID)
	require.Equal(t, "bar", newPanels[0].GraphPanel.Targets[1].LegendFormat)
	require.Equal(t, "sum(keycloak_registrations{instance=\"$instance\", foo=~\"$foo\"})", newPanels[0].GraphPanel.Targets[1].Expr)
	require.Equal(t, "C", newPanels[0].GraphPanel.Targets[2].RefID)
	require.Equal(t, "baz", newPanels[0].GraphPanel.Targets[2].LegendFormat)
	require.Equal(t, "sum(keycloak_registrations{instance=\"$instance\", foo=~\"$foo\"})", newPanels[0].GraphPanel.Targets[2].Expr)

	newCustomPanel := *newPanels[1].CustomPanel
	newTargets, ok := newCustomPanel["targets"].([]map[string]interface{})
	require.True(t, ok, "value not of expected type")
	require.Equal(t, "sum(keycloak_registrations{instance=\"$instance\", foo=~\"$foo\"})", newTargets[0]["expr"])
}
