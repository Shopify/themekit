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

	filenames, err := filesToDownload(ctx)
	if err != nil {
		return err
	}

	if len(filenames) == 0 {
		return fmt.Errorf("No files to download")
	}

	localAssets, err := shopify.FindAssets(ctx.Env)
	if err != nil {
		return err
	}

	ctx.StartProgress(len(filenames))
	for filename, remoteCheck := range filenames {
		downloadGroup.Add(1)
		go func(filename, remoteCheck string) {
			defer ctx.DoneTask()
			defer downloadGroup.Done()
			localCheck, ok := localAssets[filename]
			if ok && localCheck == remoteCheck {
				if ctx.Flags.Verbose {
					ctx.Log.Printf("[%s] No changes made to %s", colors.Green(ctx.Env.Name), colors.Blue(filename))
				}
			} else if asset, err := ctx.Client.GetAsset(filename); err != nil {
				ctx.Err("[%s] error downloading asset: %s", colors.Green(ctx.Env.Name), err)
			} else if err = asset.Write(ctx.Env.Directory); err != nil {
				ctx.Err("[%s] error writing asset: %s", colors.Green(ctx.Env.Name), err)
			} else if ctx.Flags.Verbose {
				ctx.Log.Printf("[%s] Successfully wrote %s to disk", colors.Green(ctx.Env.Name), colors.Blue(filename))
			}
		}(filename, remoteCheck)
	}

	downloadGroup.Wait()
	return nil
}

func filesToDownload(ctx *cmdutil.Ctx) (map[string]string, error) {
	allFilenames, err := ctx.Client.GetAllAssets()
	if err != nil {
		return allFilenames, err
	} else if len(ctx.Args) <= 0 {
		return allFilenames, nil
	}

	fetchableFilenames := map[string]string{}
	for filename, checksum := range allFilenames {
		for _, pattern := range ctx.Args {
			// These need to be converted to platform specific because filepath.Match
			// uses platform specific separators
			pattern = filepath.FromSlash(pattern)
			filename = filepath.FromSlash(filename)

			globMatched, _ := filepath.Match(pattern, filename)
			dirMatched, _ := filepath.Match(pattern+string(filepath.Separator)+"*", filename)
			fileMatched := filename == pattern
			if globMatched || dirMatched || fileMatched {
				fetchableFilenames[filename] = checksum
			}
		}
	}

	if len(fetchableFilenames) == 0 {
		return fetchableFilenames, fmt.Errorf("No file paths matched the inputted arguments")
	}

	return fetchableFilenames, nil
}
