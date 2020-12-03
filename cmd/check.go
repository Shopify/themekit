package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/lint"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check theme files for common issues",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return check(flags)
	},
}

func check(flags cmdutil.Flags) error {
	config, err := env.Load(flags.ConfigPath)
	envName := env.Default.Name
	if len(flags.Environments) > 0 {
		envName = flags.Environments[0]
	}

	var e *env.Env
	flagEnv := env.Env{Directory: flags.Directory}
	if e, err = config.Get(envName, flagEnv); err != nil {
		return err
	}

	return lint.Lint(e.Directory)
}
