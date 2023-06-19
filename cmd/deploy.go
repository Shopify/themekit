package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"sync"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/file"
	"github.com/Shopify/themekit/src/shopify"
)

const settingsDataKey = "config/settings_data.json"

var compiledFilenameWarning = template.Must(template.New("compiledFilenamesWarning").Parse(
	`[{{.EnvName}}] You have file names that will conflict with each other.
If you have files named [filename].js.liquid or [filename].scss.liquid,
they will be compiled to [filename].js and [filename].scss respectively
when they are uploaded to Shopify. Having both files uploaded to Shopify
will overwrite one or the other and cause unexpected behavior.

The files you will need to change are:
  {{- range .FileNames }}
	{{ . }}
	{{- end }}

To fix this, you will need to ignore, rename or delete one or both of
the files.
`))

var deployCmd = &cobra.Command{
	Use:   "deploy <filenames>",
	Short: "deploy files to shopify",
	Long: `Deploy will overwrite specific files if provided with file names.
 If deploy is not provided with file names then it will deploy all
 the files on shopify with your local files. Any files that do not
 exist on your local machine will be removed from shopify unless the --nodelete
 flag is passed

 For more information, refer to https://shopify.dev/tools/theme-kit/command-reference#deploy.
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForEachClient(flags, args, deploy)
	},
	PostRun: func(cmd *cobra.Command, args []string) {
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
			defer perform(ctx, path, op, "")
			continue
		}
		deployGroup.Add(1)
		go func(path string, op file.Op) {
			defer deployGroup.Done()
			perform(ctx, path, op, "")
		}(path, op)
	}

	deployGroup.Wait()

	return nil
}

func generateActions(ctx *cmdutil.Ctx) (map[string]file.Op, error) {
	assetsActions := map[string]file.Op{}
	pathsToChecksums := map[string]string{}

	remoteFiles, err := ctx.Client.GetAllAssets()
	if err != nil {
		return assetsActions, err
	}
	for _, remoteAsset := range remoteFiles {
		if len(ctx.Args) == 0 && !ctx.Flags.NoDelete {
			assetsActions[remoteAsset.Key] = file.Remove
		}
		pathsToChecksums[remoteAsset.Key] = remoteAsset.Checksum
	}

	localAssets, err := shopify.FindAssets(ctx.Env, ctx.Args...)
	if err != nil {
		return assetsActions, err
	}

	problemAssets := compileAssetFilenames(localAssets)
	if len(problemAssets) > 0 {
		return assetsActions, compiledAssetWarning(ctx.Env.Name, problemAssets)
	}

	for _, asset := range localAssets {
		var path = asset.Key
		if asset.Checksum != "" && (asset.Checksum == pathsToChecksums[asset.Key]) {
			assetsActions[path] = file.Skip
		} else {
			assetsActions[path] = file.Update
		}
	}
	return assetsActions, nil
}

func compileAssetFilenames(assets []shopify.Asset) (problemAssets []string) {
	var filenames []string
	for _, asset := range assets {
		filenames = append(filenames, asset.Key)
	}
	sort.Strings(filenames)
	for i, filename := range filenames {
		if i < len(filenames)-1 && filename+".liquid" == filenames[i+1] {
			problemAssets = append(problemAssets, colors.Yellow(filenames[i])+
				colors.Blue(" conflicts with ")+
				colors.Yellow(filenames[i+1]))
		}
	}
	return
}

func compiledAssetWarning(env string, filenames []string) error {
	var tpl bytes.Buffer
	compiledFilenameWarning.Execute(&tpl, struct {
		EnvName   string
		FileNames []string
	}{EnvName: colors.Yellow(env), FileNames: filenames})
	return errors.New(tpl.String())
}
