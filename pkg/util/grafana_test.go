package util

import (
	"testing"

	"github.com/grafana-tools/sdk"
	"github.com/stretchr/testify/require"
)

func TestGetTargetsBasic(t *testing.T) {
	panel := &sdk.Panel{
		CommonPanel: sdk.CommonPanel{
			OfType: sdk.StatType,
		},
		StatPanel: &sdk.StatPanel{
			Targets: []sdk.Target{
				{},
				{},
			},
		},
	}
	targets, err := GetTargets(panel)
	require.NoError(t, err)
	require.NotEmpty(t, targets)
}

func TestAppendFilterBasic(t *testing.T) {
	tplVars := []sdk.TemplateVar{
		{
			Name:  "foo",
			Label: "Foo",
		},
	}
	//nolint:lll // ignore
	exprString, err := AppendVariables("sum(gotk_reconcile_condition{namespace=~\"$namespace\", type=\"Ready\", status=\"False\", kind=~\"Kustomization|HelmRelease\"})", tplVars)
	require.NoError(t, err)
	//nolint:lll // ignore
	require.Equal(t, "sum(gotk_reconcile_condition{namespace=~\"$namespace\", type=\"Ready\", status=\"False\", kind=~\"Kustomization|HelmRelease\", foo=~\"$foo\"})", exprString)
}

func TestAppendFilterLabelValues(t *testing.T) {
	tplVars := []sdk.TemplateVar{
		{
			Name:  "foo",
			Label: "Foo",
		},
	}
	exprString, err := AppendVariables("label_values(metric, label)", tplVars)
	require.NoError(t, err)
	require.Equal(t, "label_values(metric{foo=~\"$foo\"}, label)", exprString)
}

func TestAppendFilterWithVariables(t *testing.T) {
	tplVars := []sdk.TemplateVar{
		{
			Name:  "foo",
			Label: "Foo",
		},
	}
	//nolint:lll // ignore
	exprString, err := AppendVariables("sort_desc(sum(irate(container_network_transmit_packets_dropped_total{cluster=\"$cluster\",namespace=~\".+\"}[$__interval:$resolution])) by (namespace))", tplVars)
	require.NoError(t, err)
	//nolint:lll // ignore
	require.Equal(t, "sort_desc(sum(irate(container_network_transmit_packets_dropped_total{cluster=\"$cluster\", namespace=~\".+\", foo=~\"$foo\"}[$__interval:$resolution])) by (namespace))", exprString)
}

func TestReplaceInterval(t *testing.T) {
	//nolint:lll // ignore
	result, cache := replaceInterval("container_network_transmit_packets_dropped_total{cluster=\"$cluster\",namespace=~\".+\"}[$__interval:$resolution]")
	require.NotEmpty(t, cache)
	require.NotContains(t, result, "$__interval")
	require.NotContains(t, result, "$resolution")
}

func TestUnReplaceInterval(t *testing.T) {
	cache := map[string]string{
		"123m": "$__interval",
		"456m": "$resolution",
	}

	result := unReplaceInterval("container_network_transmit_packets_dropped_total{cluster=\"$cluster\",namespace=~\".+\"}[123m:456m]", cache)
	//nolint:lll // ignore
	require.Equal(t, "container_network_transmit_packets_dropped_total{cluster=\"$cluster\",namespace=~\".+\"}[$__interval:$resolution]", result)
}
