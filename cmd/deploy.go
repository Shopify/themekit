package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
)

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
		if len(args) > 0 || flags.Soft {
			return cmdutil.ForEachClient(flags, args, upload)
		}
		return cmdutil.ForEachClient(flags, args, replace)
	},
}
