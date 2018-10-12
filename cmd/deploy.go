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

var deployCmd = &cobra.Command{
	Use:   "deploy <filenames>",
	Short: "deploy files to shopify",
	Long: `Deploy will overwrite specific files if provided with file names.
 If deploy is not provided with file names then it will deploy all
 the files on shopify with your local files. Any files that do not
 exist on your local machine will be removed from shopify unless the --soft
 flag is passed

 For more documentation please see http://shopify.github.io/themekit/commands/#deploy
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForEachClient(flags, args, deploy)
	},
}

func deploy(ctx *cmdutil.Ctx) error {
	if ctx.Env.ReadOnly {
		return fmt.Errorf("[%s] environment is readonly", colors.Green(ctx.Env.Name))
	}

	assetsActions, err := generateActions(ctx)
	if err != nil {
		return err
	}

	var deployGroup sync.WaitGroup
	ctx.StartProgress(len(assetsActions))
	for path, op := range assetsActions {
		if path == settingsDataKey {
			defer perform(ctx, path, op)
			continue
		}
		deployGroup.Add(1)
		go func(path string, op file.Op) {
			defer deployGroup.Done()
			perform(ctx, path, op)
		}(path, op)
	}

	deployGroup.Wait()
	return nil
}

func generateActions(ctx *cmdutil.Ctx) (map[string]file.Op, error) {
	assetsActions := map[string]file.Op{}

	if len(ctx.Args) == 0 && !ctx.Flags.NoDelete {
		remoteFiles, err := ctx.Client.GetAllAssets()
		if err != nil {
			return assetsActions, err
		}
		for _, filename := range remoteFiles {
			assetsActions[filename] = file.Remove
		}
	}

	localAssets, err := shopify.FindAssets(ctx.Env, ctx.Args...)
	if err != nil {
		return assetsActions, err
	}

	for _, path := range localAssets {
		assetsActions[path] = file.Update
	}
	return assetsActions, nil
}
