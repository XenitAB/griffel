package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/spf13/afero"

	"github.com/xenitab/griffel/pkg/config"
	"github.com/xenitab/griffel/pkg/dashboard"
)

func main() {
	var args struct {
		ConfigPath string `arg:"--config-path,required"`
	}
	arg.MustParse(&args)

	fs := afero.NewOsFs()
	cfg, err := config.ParseConfig(args.ConfigPath)
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
