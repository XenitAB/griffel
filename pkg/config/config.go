package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

type DashboardDatasource struct {
	Name  string `yaml:"name"`
	Label string `yaml:"label"`
	Regex string `yaml:"regex"`
	Hide  bool   `yaml:"hide"`
}

type DashboardVariable struct {
	Name  string `yaml:"name"`
	Label string `yaml:"label"`
	Query string `yaml:"query"`
}

type GlobalDashboardPatch struct {
	Variables  []DashboardVariable  `yaml:"variables"`
	Datasource *DashboardDatasource `yaml:"datasource"`
}

type DashboardPatch struct {
	Title string   `yaml:"title"`
	Tags  []string `yaml:"tags"`
}

type Dashboard struct {
	Name        string               `yaml:"name"`
	Format      string               `yaml:"format"`
	Source      DashboardSource      `yaml:"source"`
	Destination DashboardDestination `yaml:"destination"`
	Patch       DashboardPatch       `yaml:"patch"`
}

type Config struct {
	Dashboards []Dashboard          `yaml:"dashboards"`
	Patch      GlobalDashboardPatch `yaml:"patch"`
}

type SourceKind string

const (
	SourceKindGrafanaLabs SourceKind = "GrafanaLabs"
	SourceKindURL                    = "URL"
	SourceKindPath                   = "Path"
)

type DashboardSource struct {
	Kind  SourceKind `yaml:"kind"`
	Value string     `yaml:"value"`
}

type OutputFormat string

const (
	OutputFormatJson       OutputFormat = "JSON"
	OutputFormatKubernetes              = "Kubernetes"
)

type DashboardDestination struct {
	Format    OutputFormat `yaml:"format"`
	Directory string       `yaml:"directory"`
}

func ParseConfig(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("config path cannot be empty")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
