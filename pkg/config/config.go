package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Dashboards []Dashboard    `yaml:"dashboards"`
	Patch      DashboardPatch `yaml:"patch"`
}

type DashboardPatch struct {
	Variables []DashboardVariable `yaml:"variables"`
}

type DashboardVariable struct {
	Name  string `yaml:"name"`
	Label string `yaml:"label"`
}

type Dashboard struct {
	Name        string               `yaml:"name"`
	Format      string               `yaml:"format"`
	Source      DashboardSource      `yaml:"source"`
	Destination DashboardDestination `yaml:"destination"`
}

type DashboardSource struct {
	Kind  string `yaml:"kind"` // GrafanaLabs, URL or Path
	Value string `yaml:"value"`
}

type DashboardDestination struct {
	Format string `yaml:"format"` // JSON or Operator
	Path   string `yaml:"path"`
}

func ParseConfig(path string) (*Config, error) {
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
