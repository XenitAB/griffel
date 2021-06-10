package main

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	flag "github.com/spf13/pflag"

	"github.com/xenitab/griffel/pkg/config"
	"github.com/xenitab/griffel/pkg/dashboard"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "", "path to configuration file.")
	flag.Parse()
}

func main() {
	fs := afero.NewOsFs()
	cfg, err := config.ParseConfig(configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = dashboard.Patch(fs, cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
