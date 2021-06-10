package dashboard

import (
	"testing"

	"github.com/grafana-tools/sdk"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/xenitab/griffel/pkg/config"
)

const expectedJson = `{
  "id": 1,
  "slug": "",
  "title": "foobar",
  "originalTitle": "",
  "tags": null,
  "style": "dark",
  "timezone": "browser",
  "editable": true,
  "hideControls": false,
  "sharedCrosshair": false,
  "panels": null,
  "rows": [],
  "templating": {
    "list": null
  },
  "annotations": {
    "list": null
  },
  "schemaVersion": 0,
  "version": 0,
  "links": null,
  "time": {
    "from": "",
    "to": ""
  },
  "timepicker": {
    "refresh_intervals": null,
    "time_options": null
  }
}`

func TestWriteJson(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := writeDashboard(fs, "dashboard.json", config.OutputFormatJson, "test", sdk.NewBoard("foobar"))
	require.NoError(t, err)
	yaml, err := afero.ReadFile(fs, "dashboard.json")
	require.NoError(t, err)
	require.Equal(t, expectedJson, string(yaml))
}

const expectedYaml = `apiVersion: integreatly.org/v1alpha1
kind: GrafanaDashboard
metadata:
  creationTimestamp: null
  name: test
spec:
  json: |-
    {
      "id": 2,
      "slug": "",
      "title": "foobar",
      "originalTitle": "",
      "tags": null,
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "sharedCrosshair": false,
      "panels": null,
      "rows": [],
      "templating": {
        "list": null
      },
      "annotations": {
        "list": null
      },
      "schemaVersion": 0,
      "version": 0,
      "links": null,
      "time": {
        "from": "",
        "to": ""
      },
      "timepicker": {
        "refresh_intervals": null,
        "time_options": null
      }
    }
  jsonnet: ""
`

func TestWriteOperator(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := writeDashboard(fs, "dashboard.yaml", config.OutputFormatKubernetes, "test", sdk.NewBoard("foobar"))
	require.NoError(t, err)
	yaml, err := afero.ReadFile(fs, "dashboard.yaml")
	require.NoError(t, err)
	require.Equal(t, expectedYaml, string(yaml))
}
