package cmd

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/file"
	"github.com/Shopify/themekit/src/shopify"
)

const settingsDataKey = "config/settings_data.json"

var (
	deployCmd = &cobra.Command{
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

	replaceCmd = &cobra.Command{
		Use:   "replace <filenames>",
		Short: "Overwrite theme file(s)",
		Long: `Replace will overwrite specific files if provided with file names.
 If replace is not provided with file names then it will replace all
 the files on shopify with your local files. Any files that do not
 exist on your local machine will be removed from shopify.

  Deprecation Notice: This command is deprecated in v0.8.0 and will be removed in
	v0.8.1. Please use the 'deploy' command instead.

 For more documentation please see http://shopify.github.io/themekit/commands/#replace
 `,
		RunE: func(cmd *cobra.Command, args []string) error {
			colors.ColorStdOut.Printf("[%s] replace has been deprecated please use `deploy` instead", colors.Yellow("WARN"))
			return cmdutil.ForEachClient(flags, args, deploy)
		},
	}

	uploadCmd = &cobra.Command{
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
			flags.NoDelete = true
			return cmdutil.ForEachClient(flags, args, deploy)
		},
	}
)

func deploy(ctx cmdutil.Ctx) error {
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

func generateActions(ctx cmdutil.Ctx) (map[string]file.Op, error) {
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
		assetsActions[filepath.ToSlash(path)] = file.Update
	}
	return assetsActions, nil
}
