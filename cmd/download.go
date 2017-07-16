package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

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
	var downloadGroup errgroup.Group
	filenames = arbiter.manifest.FetchableFiles(filenames, client.Config.Environment)
	bar := arbiter.newProgressBar(len(filenames), client.Config.Environment)
	for _, filename := range filenames {
		filename := filename
		downloadGroup.Go(func() error {
			if err := downloadFile(client, filename); err != nil {
				stdErr.Printf("[%s] %s", green(client.Config.Environment), err)
			}
			incBar(bar)
			return nil
		})
	}
	return downloadGroup.Wait()
}

func downloadFile(client kit.ThemeClient, filename string) error {
	if !arbiter.force && !arbiter.manifest.NeedsDownloading(filename, client.Config.Environment) {
		if arbiter.verbose {
			stdOut.Print(green(fmt.Sprintf("[%s] skipping %s", client.Config.Environment, filename)))
		}
		return nil
	}

	asset, err := client.Asset(filename)
	if err != nil {
		return fmt.Errorf("error downloading asset: %v", err)
	}

	if err := asset.Write(client.Config.Directory); err != nil {
		return fmt.Errorf("error writing asset: %v", err)
	}

	if err := arbiter.manifest.Set(asset.Key, client.Config.Environment, asset.UpdatedAt); err != nil {
		return fmt.Errorf("error updating manifest: %v", err)
	}

	if arbiter.verbose {
		stdOut.Print(green(fmt.Sprintf("[%s] Successfully wrote %s to disk", client.Config.Environment, filename)))
	}
	return nil
}
