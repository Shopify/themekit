package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/shopify"
)

const settingsDataKey = "config/settings_data.json"

var uploadCmd = &cobra.Command{
	Use:   "upload <filenames>",
	Short: "Upload theme file(s) to shopify",
	Long: `Upload will upload specific files to shopify servers if provided file names.
 If no filenames are provided then upload will upload every file in the project
 to shopify.

 For more documentation please see http://shopify.github.io/themekit/commands/#upload
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForEachClient(flags, args, upload)
	},
}

func upload(ctx cmdutil.Ctx) error {
	if ctx.Env.ReadOnly {
		return fmt.Errorf("[%s] environment is readonly", colors.Green(ctx.Env.Name))
	}

	assets, err := shopify.ReadAssets(ctx.Env, ctx.Args...)
	if err != nil {
		return err
	}

	var uploadGroup sync.WaitGroup
	ctx.StartProgress(len(assets))
	for _, asset := range assets {
		if asset.Key == settingsDataKey {
			asset := asset
			defer cmdutil.UploadAsset(ctx, asset)
			continue
		}
		uploadGroup.Add(1)
		go func(asset shopify.Asset) {
			defer uploadGroup.Done()
			cmdutil.UploadAsset(ctx, asset)
		}(asset)
	}

	uploadGroup.Wait()
	return nil
}
