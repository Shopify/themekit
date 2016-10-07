package cmd

import (
	"os"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

var replaceCmd = &cobra.Command{
	Use:   "replace <filenames>",
	Short: "Overwrite theme file(s)",
	Long: `Replace will overwrite specific files if provided with file names.
If replace is not provided with file names then it will replace all
the files on shopify with your local files. Any files that do not
exist on your local machine will be removed from shopify.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd.Name(), true); err != nil {
			return err
		}

		wg := sync.WaitGroup{}
		for _, client := range themeClients {
			wg.Add(1)
			go replace(client, args, &wg)
		}
		wg.Wait()
		return nil
	},
}

func replace(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) {
	jobQueue := client.Process(wg)
	root, _ := os.Getwd()
	assetsActions := map[string]kit.AssetEvent{}
	if len(filenames) == 0 {
		for _, asset := range client.AssetList() {
			assetsActions[asset.Key] = kit.NewRemovalEvent(asset)
		}
		for _, asset := range client.LocalAssets(root) {
			assetsActions[asset.Key] = kit.NewUploadEvent(asset)
		}
	} else {
		for _, filename := range filenames {
			asset, err := theme.LoadAsset(root, filename)
			if err != nil {
				client.ErrorMessage(err.Error())
			} else if asset.IsValid() {
				assetsActions[asset.Key] = kit.NewUploadEvent(asset)
			}
		}
	}
	for _, event := range assetsActions {
		jobQueue <- event
	}
	close(jobQueue)
}
