module github.com/xenitab/griffel

go 1.16

require (
	github.com/VictoriaMetrics/metricsql v0.15.0
	github.com/grafana-tools/sdk v0.0.0-20210521150820-354cd37a4b4e
	github.com/kr/pretty v0.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	golang.org/x/sys v0.0.0-20210603081109-ebe580a85c40 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/prometheus/tsdb => github.com/prometheus/tsdb v0.3.0
