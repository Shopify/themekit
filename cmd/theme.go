package cmd

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/release"
)

const afterUpdateMessage = `Successfully updated to theme kit version %v, for more information on this release please see the change log
	https://github.com/Shopify/themekit/blob/master/changelog.txt

If you have troubles with this release please report them to
	https://github.com/Shopify/themekit/issues

If your troubles are preventing you from working you can roll back to the previous version using the command
	'theme update --version=v%s'
 `

var (
	flags cmdutil.Flags

	// ThemeCmd is the main entry point to the theme kit command line interface.
	ThemeCmd = &cobra.Command{
		Use:   "theme",
		Short: "Theme Kit is a tool kit for manipulating shopify themes",
		Long: `Theme Kit is a tool kit for manipulating shopify themes

Theme Kit is a fast and cross platform tool that enables you to build shopify themes with ease.

Complete documentation is available at https://shopify.github.io/themekit/`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !flags.DisableUpdateNotifier && release.IsUpdateAvailable() {
				colors.ColorStdOut.Print(colors.Yellow("An update for Themekit is available. To update please run `theme update`"))
			}
		},
	}

	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update Theme kit to the newest version.",
		Long: `Update will check for a new release, then if there is an applicable update it will download it and apply it.

 For more documentation please see http://shopify.github.io/themekit/commands/#update
 `,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			colors.ColorStdOut.Printf("Updating from %s to %s", colors.Yellow(release.ThemeKitVersion), colors.Yellow(flags.Version))
			if err = release.Install(flags.Version); err == nil {
				colors.ColorStdOut.Printf(afterUpdateMessage, colors.Green(flags.Version), colors.Yellow(release.ThemeKitVersion))
			}
			return
		},
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Theme Kit",
		Long:  `All software has versions. This is Theme Kit's version.`,
		Run: func(cmd *cobra.Command, args []string) {
			colors.ColorStdOut.Printf("ThemeKit %s %s/%s", release.ThemeKitVersion.String(), runtime.GOOS, runtime.GOARCH)
		},
	}
)

func init() {
	pwd, _ := os.Getwd()
	defaultConfigPath := filepath.Join(pwd, "config.yml")

	ThemeCmd.PersistentFlags().StringVarP(&flags.ConfigPath, "config", "c", defaultConfigPath, "path to config.yml")
	ThemeCmd.PersistentFlags().VarP(&flags.Environments, "env", "e", "environment to run the command")
	ThemeCmd.PersistentFlags().StringVarP(&flags.Directory, "dir", "d", "", "directory that command will take effect. (default current directory)")
	ThemeCmd.PersistentFlags().StringVarP(&flags.Password, "password", "p", "", "theme password. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVarP(&flags.ThemeID, "themeid", "t", "", "theme id. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVarP(&flags.Domain, "store", "s", "", "your shopify domain. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVar(&flags.Proxy, "proxy", "", "proxy for all theme requests. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().DurationVar(&flags.Timeout, "timeout", 0, "the timeout to kill any stalled processes. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Enable more verbose output from the running command.")
	ThemeCmd.PersistentFlags().BoolVarP(&flags.DisableUpdateNotifier, "no-update-notifier", "", false, "Stop theme kit from notifying about updates.")
	ThemeCmd.PersistentFlags().Var(&flags.IgnoredFiles, "ignored-file", "A single file to ignore, use the flag multiple times to add multiple.")
	ThemeCmd.PersistentFlags().Var(&flags.Ignores, "ignores", "A path to a file that contains ignore patterns.")
	ThemeCmd.PersistentFlags().BoolVar(&flags.DisableIgnore, "no-ignore", false, "Will disable config ignores so that all files can be changed")

	watchCmd.Flags().StringVarP(&flags.NotifyFile, "notify", "n", "", "file to touch when workers have gone idle")
	watchCmd.Flags().BoolVarP(&flags.AllEnvs, "allenvs", "a", false, "run command with all environments")
	removeCmd.Flags().BoolVarP(&flags.AllEnvs, "allenvs", "a", false, "run command with all environments")
	replaceCmd.Flags().BoolVarP(&flags.AllEnvs, "allenvs", "a", false, "run command with all environments")
	uploadCmd.Flags().BoolVarP(&flags.AllEnvs, "allenvs", "a", false, "run command with all environments")
	openCmd.Flags().BoolVarP(&flags.AllEnvs, "allenvs", "a", false, "run command with all environments")
	downloadCmd.Flags().BoolVarP(&flags.AllEnvs, "allenvs", "a", false, "run command with all environments")
	newCmd.Flags().StringVar(&flags.Version, "version", "latest", "version of Shopify Timber to use")
	bootstrapCmd.Flags().StringVar(&flags.Version, "version", "latest", "version of Shopify Timber to use")
	updateCmd.Flags().StringVar(&flags.Version, "version", "latest", "version of themekit to install")
	newCmd.Flags().StringVar(&flags.Prefix, "prefix", "", "prefix to the Timber theme being created")
	newCmd.Flags().StringVar(&flags.URL, "url", "", "a url to pull a project theme zip file from.")
	newCmd.Flags().StringVar(&flags.Name, "name", "", "a name to define your theme on your shopify admin")
	bootstrapCmd.Flags().StringVar(&flags.Prefix, "prefix", "", "prefix to the Timber theme being created")
	bootstrapCmd.Flags().StringVar(&flags.URL, "url", "", "a url to pull a project theme zip file from.")
	bootstrapCmd.Flags().StringVar(&flags.Name, "name", "", "a name to define your theme on your shopify admin")
	openCmd.Flags().BoolVarP(&flags.Edit, "edit", "E", false, "open the web editor for the theme.")
	openCmd.Flags().StringVarP(&flags.With, "browser", "b", "", "name of the browser to open the url. the name should match the name of browser on your system.")
	getCmd.Flags().BoolVarP(&flags.List, "list", "l", false, "list available themes.")

	ThemeCmd.AddCommand(openCmd, versionCmd, bootstrapCmd, newCmd, configureCmd, downloadCmd, removeCmd, updateCmd, uploadCmd, replaceCmd, watchCmd, getCmd)
}
