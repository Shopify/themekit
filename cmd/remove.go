package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/Shopify/themekit/kit"
)

var removeCmd = &cobra.Command{
	Use:   "remove <filenames>",
	Short: "Remove theme file(s) from shopify",
	Long: `Remove will delete all specified files from shopify servers.

For more documentation please see http://shopify.github.io/themekit/commands/#remove
	`,
	RunE: arbiter.forEachClient(remove),
}

func remove(client kit.ThemeClient, filenames []string) error {
	if client.Config.ReadOnly {
		return fmt.Errorf("[%s] environment is reaonly", kit.GreenText(client.Config.Environment))
	} else if len(filenames) == 0 {
		return fmt.Errorf("[%s] please specify files to be removed", kit.GreenText(client.Config.Environment))
	}

	var removeGroup errgroup.Group
	bar := arbiter.newProgressBar(len(filenames), client.Config.Environment)
	for _, filename := range filenames {
		asset := kit.Asset{Key: filename}
		removeGroup.Go(func() error {
			if perform(client, asset, kit.Remove, bar, nil) {
				return os.Remove(filepath.Join(client.Config.Directory, asset.Key))
			}
			return nil
		})
	}

	return removeGroup.Wait()
}
