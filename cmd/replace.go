package cmd

import (
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/file"
	"github.com/Shopify/themekit/src/shopify"
)

var replaceCmd = &cobra.Command{
	Use:   "replace <filenames>",
	Short: "Overwrite theme file(s)",
	Long: `Replace will overwrite specific files if provided with file names.
 If replace is not provided with file names then it will replace all
 the files on shopify with your local files. Any files that do not
 exist on your local machine will be removed from shopify.

 For more documentation please see http://shopify.github.io/themekit/commands/#replace
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return cmdutil.ForEachClient(flags, args, upload)
		}
		return cmdutil.ForEachClient(flags, args, replace)
	},
}

func replace(ctx cmdutil.Ctx) error {
	assetsActions, err := generateActions(ctx)
	if err != nil {
		return err
	}

	var replaceGroup sync.WaitGroup
	ctx.StartProgress(len(assetsActions))
	for path, op := range assetsActions {
		if path == settingsDataKey {
			defer perform(ctx, path, op)
			continue
		}
		replaceGroup.Add(1)
		go func(path string, op file.Op) {
			defer replaceGroup.Done()
			perform(ctx, path, op)
		}(path, op)
	}

	replaceGroup.Wait()
	return nil
}

func generateActions(ctx cmdutil.Ctx) (map[string]file.Op, error) {
	assetsActions := map[string]file.Op{}

	remoteFiles, err := ctx.Client.GetAllAssets()
	if err != nil {
		return assetsActions, err
	}
	for _, filename := range remoteFiles {
		assetsActions[filename] = file.Remove
	}

	localAssets, err := shopify.FindAssets(ctx.Env)
	if err != nil {
		return assetsActions, err
	}

	for _, path := range localAssets {
		assetsActions[path] = file.Update
	}
	return assetsActions, nil
}
