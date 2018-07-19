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
	for _, action := range assetsActions {
		if action.asset.Key == settingsDataKey {
			action := action
			defer action.do(ctx)
			continue
		}
		replaceGroup.Add(1)
		go func(action replaceAction) {
			defer replaceGroup.Done()
			action.do(ctx)
		}(action)
	}

	replaceGroup.Wait()
	return nil
}

func generateActions(ctx cmdutil.Ctx) (map[string]replaceAction, error) {
	assetsActions := map[string]replaceAction{}

	remoteFiles, err := ctx.Client.GetAllAssets()
	if err != nil {
		return assetsActions, err
	}
	for _, filename := range remoteFiles {
		assetsActions[filename] = replaceAction{
			asset: shopify.Asset{Key: filename},
			op:    file.Remove,
		}
	}

	localAssets, err := shopify.ReadAssets(ctx.Env)
	if err != nil {
		return assetsActions, err
	}

	for _, asset := range localAssets {
		assetsActions[asset.Key] = replaceAction{asset: asset, op: file.Update}
	}
	return assetsActions, nil
}

type replaceAction struct {
	asset shopify.Asset
	op    file.Op
}

func (action *replaceAction) do(ctx cmdutil.Ctx) {
	if action.op == file.Remove {
		cmdutil.DeleteAsset(ctx, action.asset)
	} else {
		cmdutil.UploadAsset(ctx, action.asset)
	}
}
