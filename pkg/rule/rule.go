package rule

import (
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/spf13/afero"

	"github.com/xenitab/griffel/pkg/config"
)

func Patch(fs afero.Fs, cfg *config.Config) error {
  for _, _ = range cfg.Rules {
    promeRule := monitoringv1.PrometheusRule{}
    fmt.Println(promeRule)
  }
	return nil
}
