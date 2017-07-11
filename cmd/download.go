package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb"

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
	PreRunE: arbiter.generateThemeClients,
	RunE:    arbiter.forSingleClient(download),
}

func download(client kit.ThemeClient, filenames []string) error {
	var wg sync.WaitGroup
	filenames = arbiter.manifest.FetchableFiles(filenames, client.Config.Environment)
	bar := arbiter.newProgressBar(len(filenames), client.Config.Environment)
	for _, filename := range filenames {
		wg.Add(1)
		go downloadFile(client, filename, bar, &wg)
	}
	wg.Wait()
	return nil
}

func downloadFile(client kit.ThemeClient, filename string, bar *mpb.Bar, wg *sync.WaitGroup) error {
	defer arbiter.cleanupAction(bar, wg)

	if !arbiter.force && !arbiter.manifest.NeedsDownloading(filename, client.Config.Environment) {
		if arbiter.verbose {
			kit.Print(kit.GreenText(fmt.Sprintf("[%s] no changes were made to %s so it was skipped", client.Config.Environment, filename)))
		}
		return fmt.Errorf("Skipping because versions match")
	}

	asset, err := client.Asset(filename)
	if err != nil {
		kit.LogErrorf("[%s]%s", kit.GreenText(client.Config.Environment), err)
		return fmt.Errorf("error downloading asset: %v", err)
	}

	if err := asset.Write(client.Config.Directory); err != nil {
		kit.LogErrorf("[%s]%s", kit.GreenText(client.Config.Environment), err)
		return fmt.Errorf("error writing asset: %v", err)
	}

	if err := arbiter.manifest.Set(asset.Key, client.Config.Environment, asset.UpdatedAt); err != nil {
		kit.LogErrorf("[%s] Could not update manifest %s", kit.GreenText(client.Config.Environment), err)
		return fmt.Errorf("error updating manifest: %v", err)
	}

	if arbiter.verbose {
		kit.Print(kit.GreenText(fmt.Sprintf("[%s] Successfully wrote %s to disk", client.Config.Environment, filename)))
	}
	return nil
}
