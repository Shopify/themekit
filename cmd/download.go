package cmd

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
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

	filenames, err := filesToDownload(ctx)
	if err != nil {
		return err
	}

	if len(filenames) == 0 {
		return fmt.Errorf("No files to download")
	}

	ctx.StartProgress(len(filenames))
	for _, filename := range filenames {
		downloadGroup.Add(1)
		go func(filename string) {
			defer ctx.DoneTask()
			defer downloadGroup.Done()
			if asset, err := ctx.Client.GetAsset(filename); err != nil {
				ctx.Err("[%s] error downloading %s: %s", colors.Green(ctx.Env.Name), colors.Blue(filename), err)
			} else if err = asset.Write(ctx.Env.Directory); err != nil {
				ctx.Err("[%s] error writing %s: %s", colors.Green(ctx.Env.Name), colors.Blue(filename), err)
			} else if ctx.Flags.Verbose {
				var checksumOutput = ""
				if asset.Checksum != "" {
					checksumOutput = "Checksum " + asset.Checksum
				} else {
					checksumOutput = "No checksum"
				}
				ctx.Log.Printf("[%s] Successfully wrote %s to disk (%s)", colors.Green(ctx.Env.Name), colors.Blue(filename), checksumOutput)
			}
		}(filename)
	}

	downloadGroup.Wait()
	return nil
}

func filesToDownload(ctx *cmdutil.Ctx) ([]string, error) {
	assets, err := ctx.Client.GetAllAssets()
	if err != nil {
		return []string{}, err
	}

	allFilenames := make([]string, len(assets))
	for i, asset := range assets {
		allFilenames[i] = asset.Key
	}

	if len(ctx.Args) <= 0 {
		return allFilenames, nil
	}

	fetchableFilenames := []string{}
	for _, filename := range allFilenames {
		for _, pattern := range ctx.Args {
			// These need to be converted to platform specific because filepath.Match
			// uses platform specific separators
			pattern = filepath.FromSlash(pattern)
			filename = filepath.FromSlash(filename)

			globMatched, _ := filepath.Match(pattern, filename)
			dirMatched, _ := filepath.Match(pattern+string(filepath.Separator)+"*", filename)
			fileMatched := filename == pattern
			if globMatched || dirMatched || fileMatched {
				fetchableFilenames = append(fetchableFilenames, filepath.ToSlash(filename))
			}
		}
	}

	if len(fetchableFilenames) == 0 {
		return fetchableFilenames, fmt.Errorf("No file paths matched the inputted arguments")
	}

	return fetchableFilenames, nil
}
