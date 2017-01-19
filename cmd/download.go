package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var downloadCmd = &cobra.Command{
	Use:   "download <filenames>",
	Short: "Download one or all of the theme files",
	Long: `Download will download specific files from shopify servers if provided file names.
If no filenames are provided then download will download every file in the project
and write them to disk.

For more documentation please see http://shopify.github.io/themekit/commands/#download
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		themeClients, err := generateThemeClients()
		if err != nil {
			return err
		}
		return download(themeClients[0], args)
	},
}

func download(client kit.ThemeClient, filenames []string) error {
	wg := sync.WaitGroup{}

	if len(filenames) <= 0 {
		kit.Printf("[%s] Fetching asset list from %s",
			kit.GreenText(client.Config.Environment),
			kit.YellowText(client.Config.Domain))
		assets, err := client.AssetList()
		if err != nil {
			return err
		}
		for _, asset := range assets {
			filenames = append(filenames, asset.Key)
		}
	}

	for _, filename := range filenames {
		wg.Add(1)
		if err := downloadFile(client, filename, &wg); err != nil {
			return err
		}
	}

	wg.Wait()

	return nil
}

func downloadFile(client kit.ThemeClient, filename string, wg *sync.WaitGroup) error {
	defer wg.Done()

	kit.Printf("[%s] Downloading %s from %s",
		kit.GreenText(client.Config.Environment),
		filename,
		kit.YellowText(client.Config.Domain))

	asset, err := client.Asset(filename)
	if err != nil {
		return err
	}

	if err := asset.Write(client.Config.Directory); err != nil {
		return err
	}

	kit.Print(kit.GreenText(fmt.Sprintf("[%s] Successfully wrote %s to disk", client.Config.Environment, filename)))

	return nil
}
