package cmd

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/release"
	"github.com/Shopify/themekit/src/util"
)

const afterUpdateMessage = `Successfully updated to theme kit version %v, for more information on this release please see the change log
	https://github.com/Shopify/themekit/blob/main/changelog.txt

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

Complete documentation is available at https://shopify.dev/tools/theme-kit.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !flags.DisableUpdateNotifier && release.IsUpdateAvailable() {
				colors.ColorStdOut.Print(colors.Yellow("An update for Themekit is available. To update please run `theme update`"))
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			// env validation requires a theme id. setting a dummy one here if not provided
			if flags.ThemeID == "" {
				flags.ThemeID = "1337"
			}
			cmdutil.ForDefaultClient(flags, args, func(ctx *cmdutil.Ctx) error {
				if !flags.DisableThemeKitAccessNotifier && !util.IsThemeAccessPassword(ctx.Env.Password) {
					colors.ColorStdOut.Print(colors.Yellow("* Build themes without private apps. Learn more about the Theme Access app: https://shopify.dev/themes/tools/theme-access"))
				}
				return nil
			})
		},
	}

	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update Theme kit to the newest version.",
		Long: `Update will check for a new release, then if there is an applicable update it will download it and apply it.

 For more information, refer to https://shopify.dev/tools/theme-kit/troubleshooting.
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
	ThemeCmd.PersistentFlags().StringVar(&flags.VariableFilePath, "vars", "", "path to an file that defines environment variables")
	ThemeCmd.PersistentFlags().StringArrayVarP(&flags.Environments, "env", "e", []string{env.Default.Name}, "environment to run the command")
	ThemeCmd.PersistentFlags().StringVarP(&flags.Directory, "dir", "d", "", "directory that command will take effect. (default current directory)")
	ThemeCmd.PersistentFlags().StringVarP(&flags.Password, "password", "p", "", "theme password. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVarP(&flags.ThemeID, "themeid", "t", "", "theme id. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVarP(&flags.Domain, "store", "s", "", "your shopify domain. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVar(&flags.Proxy, "proxy", "", "proxy for all theme requests. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().DurationVar(&flags.Timeout, "timeout", 0, "the timeout to kill any stalled processes. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Enable more verbose output from the running command.")
	ThemeCmd.PersistentFlags().BoolVarP(&flags.DisableUpdateNotifier, "no-update-notifier", "", false, "Stop theme kit from notifying about updates.")
	ThemeCmd.PersistentFlags().StringArrayVar(&flags.IgnoredFiles, "ignored-file", []string{}, "A single file to ignore, use the flag multiple times to add multiple.")
	ThemeCmd.PersistentFlags().StringArrayVar(&flags.Ignores, "ignores", []string{}, "A path to a file that contains ignore patterns.")
	ThemeCmd.PersistentFlags().BoolVar(&flags.DisableIgnore, "no-ignore", false, "Will disable config ignores so that all files can be changed")
	ThemeCmd.PersistentFlags().BoolVar(&flags.AllowLive, "allow-live", false, "Will allow themekit to make changes to the live theme on the store.")
	ThemeCmd.PersistentFlags().BoolVarP(&flags.DisableThemeKitAccessNotifier, "no-theme-kit-access-notifier", "", false, "Stop theme kit from notifying about Theme Access.")

	watchCmd.Flags().StringVarP(&flags.Notify, "notify", "n", "", "file to touch or url to notify when a file has been changed")
	watchCmd.Flags().BoolVarP(&flags.AllEnvs, "allenvs", "a", false, "run command with all environments")
	removeCmd.Flags().BoolVarP(&flags.AllEnvs, "allenvs", "a", false, "run command with all environments")
	openCmd.Flags().BoolVarP(&flags.AllEnvs, "allenvs", "a", false, "run command with all environments")
	downloadCmd.Flags().BoolVarP(&flags.AllEnvs, "allenvs", "a", false, "run command with all environments")
	deployCmd.Flags().BoolVarP(&flags.AllEnvs, "allenvs", "a", false, "run command with all environments")
	updateCmd.Flags().StringVar(&flags.Version, "version", "latest", "version of themekit to install")
	newCmd.Flags().StringVarP(&flags.Name, "name", "n", "", "a name to define your theme on your shopify admin")
	openCmd.Flags().BoolVarP(&flags.Edit, "edit", "E", false, "open the web editor for the theme.")
	openCmd.Flags().StringVarP(&flags.With, "browser", "b", "", "name of the browser to open the url. the name should match the name of browser on your system.")
	getCmd.Flags().BoolVarP(&flags.List, "list", "l", false, "list available themes.")
	deployCmd.Flags().BoolVarP(&flags.NoDelete, "nodelete", "n", false, "do not delete files on shopify during deploy.")
	openCmd.Flags().BoolVar(&flags.HidePreviewBar, "hidepb", false, "run command with all environments")

	getCmd.Flags().BoolVar(&flags.Live, "live", false, "will allow themekit to autofill the theme ID as the currently published theme ID")
	downloadCmd.Flags().BoolVar(&flags.Live, "live", false, "will allow themekit to autofill the theme ID as the currently published theme ID")
	configureCmd.Flags().BoolVar(&flags.Live, "live", false, "will allow themekit to autofill the theme ID as the currently published theme ID")

	ThemeCmd.AddCommand(
		configureCmd,
		deployCmd,
		downloadCmd,
		getCmd,
		newCmd,
		openCmd,
		publishCmd,
		removeCmd,
		updateCmd,
		versionCmd,
		watchCmd,
	)
}
