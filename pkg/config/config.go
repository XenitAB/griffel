package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

// Global Dashboard Patch

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

// Source

type SourceKind string

const (
	SourceKindGrafanaLabs SourceKind = "GrafanaLabs"
	SourceKindURL         SourceKind = "URL"
	SourceKindPath        SourceKind = "Path"
)

type Source struct {
	Kind  SourceKind `yaml:"kind"`
	Value string     `yaml:"value"`
}

// Output

type OutputFormat string

const (
	OutputFormatJson       OutputFormat = "JSON"
	OutputFormatKubernetes OutputFormat = "Kubernetes"
)

type Destination struct {
	Format    OutputFormat `yaml:"format"`
	Directory string       `yaml:"directory"`
}

// Dashboard

type DashboardPatch struct {
	Title    string   `yaml:"title"`
	Tags     []string `yaml:"tags"`
	Editable bool     `yaml:"editable"`
}

type Dashboard struct {
	Name        string         `yaml:"name"`
	Source      Source         `yaml:"source"`
	Destination Destination    `yaml:"destination"`
	Patch       DashboardPatch `yaml:"patch"`
}

// Rule

type Rule struct {
	Name        string      `yaml:"name"`
	Source      Source      `yaml:"source"`
	Destination Destination `yaml:"destination"`
}

// Root

type Config struct {
	Patch      GlobalDashboardPatch `yaml:"patch"`
	Dashboards []Dashboard          `yaml:"dashboards"`
	Rules      []Rule               `yaml:"rules"`
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
