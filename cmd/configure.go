package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Create a configuration file",
	Long: `Configure will create a new configuration file to
access shopify using the theme kit.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd.Name(), true); err != nil {
			return err
		}

		var errs = []string{}
		if len(domain) <= 0 {
			errs = append(errs, "\t-domain cannot be blank")
		}
		if len(password) <= 0 {
			errs = append(errs, "\t-password or access_token cannot be blank")
		}
		if len(errs) > 0 {
			return fmt.Errorf("Cannot create %s!\nErrors:\n%s", configPath, strings.Join(errs, "\n"))
		}

		config := kit.Configuration{
			Domain:     domain,
			Password:   password,
			BucketSize: bucketsize,
			RefillRate: refillrate,
			Timeout:    time.Duration(timeout) * time.Second,
			ThemeID:    themeid,
		}

		_, err := config.Initialize()
		if err != nil {
			return err
		}

		return addConfiguration(config)
	},
}

func addConfiguration(config kit.Configuration) error {
	env, err := kit.LoadEnvironments(configPath)
	if err != nil {
		return err
	}
	env.SetConfiguration(environment, config)
	return env.Save(configPath)
}
