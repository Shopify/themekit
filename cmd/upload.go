package cmd

import (
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <filenames>",
	Short: "Upload theme file(s) to shopify",
	Long: `Upload will upload specific files to shopify servers if provided file names.
If no filenames are provided then upload will upload every file in the project
to shopify.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd.Name(), true); err != nil {
			return err
		}

		wg := sync.WaitGroup{}
		for _, client := range themeClients {
			wg.Add(1)
			go upload(client, args, &wg)
		}
		wg.Wait()
		return nil
	},
}

func upload(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) {
	jobQueue := client.Process(wg)
	if len(filenames) == 0 {
		for _, asset := range client.LocalAssets() {
			if asset.IsValid() {
				jobQueue <- kit.NewUploadEvent(asset)
			}
		}
	} else {
		for _, filename := range filenames {
			asset, err := client.LocalAsset(filename)
			if err != nil {
				kit.Errorf(err.Error())
			} else if asset.IsValid() {
				jobQueue <- kit.NewUploadEvent(asset)
			}
		}
	}
	close(jobQueue)
}
