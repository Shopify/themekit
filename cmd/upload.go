package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/file"
	"github.com/Shopify/themekit/src/shopify"
)

const settingsDataKey = "config/settings_data.json"

var uploadCmd = &cobra.Command{
	Use:   "upload <filenames>",
	Short: "Upload theme file(s) to shopify",
	Long: `Upload will upload specific files to shopify servers if provided file names.
 If no filenames are provided then upload will upload every file in the project
 to shopify.

  Deprecation Notice: This command is deprecated in v0.8.0 and will be removed in
	v0.8.1. Please use the 'deploy' command instead.

 For more documentation please see http://shopify.github.io/themekit/commands/#upload
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		colors.ColorStdOut.Printf("[%s] upload has been deprecated please use `deploy` with the --nodelete flag instead", colors.Yellow("WARN"))
		return cmdutil.ForEachClient(flags, args, upload)
	},
}

func upload(ctx cmdutil.Ctx) error {
	if ctx.Env.ReadOnly {
		return fmt.Errorf("[%s] environment is readonly", colors.Green(ctx.Env.Name))
	}

	assetPaths, err := shopify.FindAssets(ctx.Env, ctx.Args...)
	if err != nil {
		return err
	}

	var uploadGroup sync.WaitGroup
	ctx.StartProgress(len(assetPaths))
	for _, path := range assetPaths {
		if path == settingsDataKey {
			defer perform(ctx, path, file.Update)
			continue
		}
		uploadGroup.Add(1)
		go func(path string) {
			defer uploadGroup.Done()
			perform(ctx, path, file.Update)
		}(path)
	}
	uploadGroup.Wait()
	return nil
}
