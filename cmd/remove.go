package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/shopify"
)

var removeCmd = &cobra.Command{
	Use:   "remove <filenames>",
	Short: "Remove theme file(s) from shopify",
	Long: `Remove will delete all specified files from shopify servers.

 For more documentation please see http://shopify.github.io/themekit/commands/#remove
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForEachClient(flags, args, func(ctx cmdutil.Ctx) error {
			return remove(ctx, os.Remove)
		})
	},
}

func remove(ctx cmdutil.Ctx, removeFile func(string) error) error {
	if ctx.Env.ReadOnly {
		return fmt.Errorf("[%s] environment is readonly", colors.Green(ctx.Env.Name))
	} else if len(ctx.Args) == 0 {
		return fmt.Errorf("[%s] please specify file(s) to be removed", colors.Green(ctx.Env.Name))
	}

	var removeGroup sync.WaitGroup
	ctx.StartProgress(len(ctx.Args))
	for _, filename := range ctx.Args {
		removeGroup.Add(1)
		asset := shopify.Asset{Key: filename}
		go func() {
			defer removeGroup.Done()
			cmdutil.DeleteAsset(ctx, asset)
			removeFile(filepath.Join(ctx.Env.Directory, asset.Key))
		}()
	}

	removeGroup.Wait()
	return nil
}
