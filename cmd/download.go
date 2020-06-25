package cmd

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/shopify"
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
		return cmdutil.ForEachClient(flags, args, download)
	},
}

func download(ctx *cmdutil.Ctx) error {
	var downloadGroup sync.WaitGroup

	assets, err := filesToDownload(ctx)
	if err != nil {
		return err
	}

	if len(assets) == 0 {
		return fmt.Errorf("No files to download")
	}

	ctx.StartProgress(len(assets))
	for _, asset := range assets {
		downloadGroup.Add(1)
		go func(requestAsset shopify.Asset) {
			defer ctx.DoneTask()
			defer downloadGroup.Done()

			localAsset, _ := shopify.ReadAsset(ctx.Env, requestAsset.Key)
			if localAsset.Checksum == requestAsset.Checksum && requestAsset.Checksum != "" {
				ctx.Log.Printf("[%s] Skipped %s (%s)", colors.Green(ctx.Env.Name), colors.Blue(requestAsset.Key), localAsset.Checksum)
			} else if asset, err := ctx.Client.GetAsset(requestAsset.Key); err != nil {
				ctx.Err("[%s] error downloading %s: %s", colors.Green(ctx.Env.Name), colors.Blue(requestAsset.Key), err)
			} else if err = asset.Write(ctx.Env.Directory); err != nil {
				ctx.Err("[%s] error writing %s: %s", colors.Green(ctx.Env.Name), colors.Blue(requestAsset.Key), err)
			} else if ctx.Flags.Verbose {
				var checksumOutput = ""
				if asset.Checksum != "" {
					checksumOutput = "Remote Checksum: " + requestAsset.Checksum + ", Local Checksum: " + localAsset.Checksum
				} else {
					checksumOutput = "No Checksum, Local Checksum " + localAsset.Checksum
				}
				ctx.Log.Printf("[%s] Successfully wrote %s to disk (%s)", colors.Green(ctx.Env.Name), colors.Blue(asset.Key), checksumOutput)
			}
		}(asset)
	}

	downloadGroup.Wait()
	return nil
}

func filesToDownload(ctx *cmdutil.Ctx) ([]shopify.Asset, error) {
	assets, err := ctx.Client.GetAllAssets()
	if err != nil {
		return []shopify.Asset{}, err
	}

	if len(ctx.Args) <= 0 {
		return assets, nil
	}

	fetchableFiles := []shopify.Asset{}
	for _, asset := range assets {
		for _, pattern := range ctx.Args {
			// These need to be converted to platform specific because filepath.Match
			// uses platform specific separators
			pattern = filepath.FromSlash(pattern)
			filename := filepath.FromSlash(asset.Key)

			globMatched, _ := filepath.Match(pattern, filename)
			dirMatched, _ := filepath.Match(pattern+string(filepath.Separator)+"*", filename)
			fileMatched := filename == pattern
			if globMatched || dirMatched || fileMatched {
				fetchableFiles = append(fetchableFiles, asset)
			}
		}
	}

	if len(fetchableFiles) == 0 {
		return fetchableFiles, fmt.Errorf("No file paths matched the inputted arguments")
	}

	return fetchableFiles, nil
}
