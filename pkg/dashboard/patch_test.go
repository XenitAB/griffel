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
	newTemplating, err := patchTemplating(templating, tplVars)
	require.NoError(t, err)
	require.Len(t, newTemplating.List, 2)
	require.Equal(t, "bar", newTemplating.List[0].Name)
	require.Equal(t, "query_2", newTemplating.List[0].Query)
	require.Equal(t, "foo", newTemplating.List[1].Name)
	require.Equal(t, "query_1{bar=~\"${bar}\"}", newTemplating.List[1].Query)
}

func TestPatchPanel(t *testing.T) {
	graphPanel := sdk.NewGraph("graph")
	graphPanel.SetTarget(&sdk.Target{
		RefID:  "A",
		Expr:   "sum(keycloak_registrations{instance=\"$instance\"})",
		Format: "time_series",
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
	require.Equal(t, "sum(keycloak_registrations{instance=\"$instance\", foo=~\"${foo}\"})", newPanels[0].GraphPanel.Targets[0].Expr)
	newCustomPanel := *newPanels[1].CustomPanel
	newTargets := newCustomPanel["targets"].([]map[string]interface{})
	require.Equal(t, "sum(keycloak_registrations{instance=\"$instance\", foo=~\"${foo}\"})", newTargets[0]["expr"])
}

func TestAppendFilterLabelValues(t *testing.T) {
	tplVars := []sdk.TemplateVar{
		{
			Name:  "foo",
			Label: "Foo",
		},
	}
	exprString, err := appendVariables("label_values(metric, label)", tplVars)
	require.NoError(t, err)
	require.Equal(t, "label_values(metric{foo=~\"${foo}\"}, label)", exprString)
}
