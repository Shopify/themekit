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

 For more documentation please see http://shopify.github.io/themekit/commands/#deploy
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForEachClient(flags, args, deploy)
	},
}

func deploy(ctx *cmdutil.Ctx) error {
	fmt.Printf("deploy...\n")
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
		fmt.Printf("path: %s\n", path)
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
	fmt.Printf("generateActions...\n")
	assetsActions := map[string]file.Op{}

	if len(ctx.Args) == 0 && !ctx.Flags.NoDelete {
		remoteFiles, err := ctx.Client.GetAllAssets()
		if err != nil {
			return assetsActions, err
		}
		for _, asset := range remoteFiles {
			assetsActions[asset.Key] = file.Remove
		}
	}

	localAssets, err := shopify.FindAssets(ctx.Env, ctx.Args...)
	if err != nil {
		return assetsActions, err
	}

	problemAssets := compileAssetFilenames(localAssets)
	if len(problemAssets) > 0 {
		return assetsActions, compiledAssetWarning(ctx.Env.Name, problemAssets)
	}

	for _, path := range localAssets {
		fmt.Printf("setting assectActions for '%s' to Update\n", path)
		// if checksums match delete assetsActions[path]
		assetsActions[path] = file.Update
	}
	return assetsActions, nil
}

func compileAssetFilenames(filenames []string) (problemAssets []string) {
	sort.Strings(filenames)
	for i := 0; i < len(filenames)-1; i += 2 {
		if filenames[i]+".liquid" == filenames[i+1] {
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
