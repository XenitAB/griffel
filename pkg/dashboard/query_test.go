package dashboard

import (
	"testing"

	"github.com/grafana-tools/sdk"
	"github.com/stretchr/testify/require"
)

func TestAppendFilterBasic(t *testing.T) {
	tplVars := []sdk.TemplateVar{
		{
			Name:  "foo",
			Label: "Foo",
		},
	}
	// nolint:lll // ignore
	query, err := NewQuery("sum(gotk_reconcile_condition{namespace=~\"$namespace\", type=\"Ready\", status=\"False\", kind=~\"Kustomization|HelmRelease\"})")
	require.NoError(t, err)
	result, err := query.AppendVariables(tplVars)
	require.NoError(t, err)
	// nolint:lll // ignore
	require.Equal(t, "sum(gotk_reconcile_condition{namespace=~\"$namespace\", type=\"Ready\", status=\"False\", kind=~\"Kustomization|HelmRelease\", foo=~\"$foo\"})", result)
}

func TestAppendFilterLabelValues(t *testing.T) {
	tplVars := []sdk.TemplateVar{
		{
			Name:  "foo",
			Label: "Foo",
		},
	}
	query, err := NewQuery("label_values(metric, label)")
	require.NoError(t, err)
	result, err := query.AppendVariables(tplVars)
	require.NoError(t, err)
	require.Equal(t, "label_values(metric{foo=~\"$foo\"}, label)", result)
}

func TestAppendFilterWithVariables(t *testing.T) {
	tplVars := []sdk.TemplateVar{
		{
			Name:  "foo",
			Label: "Foo",
		},
	}
	// nolint:lll // ignore
	query, err := NewQuery("sort_desc(sum(irate(container_network_transmit_packets_dropped_total{cluster=\"$cluster\",namespace=~\".+\"}[$__interval:$resolution])) by (namespace))")
	require.NoError(t, err)
	result, err := query.AppendVariables(tplVars)
	require.NoError(t, err)
	// nolint:lll // ignore
	require.Equal(t, "sort_desc(sum(irate(container_network_transmit_packets_dropped_total{cluster=\"$cluster\", namespace=~\".+\", foo=~\"$foo\"}[$__interval:$resolution])) by (namespace))", result)
}

func TestReplaceInterval(t *testing.T) {
	// nolint:lll // ignore
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
	// nolint:lll // ignore
	require.Equal(t, "container_network_transmit_packets_dropped_total{cluster=\"$cluster\",namespace=~\".+\"}[$__interval:$resolution]", result)
}
