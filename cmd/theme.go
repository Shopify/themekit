package cmd

import (
	"github.com/spf13/cobra"
)

var ThemeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Theme Kit is a tool kit for manipulating shopify themes",
	Long: `Theme Kit is a tool kit for manipulating shopify themes

Theme Kit is a Fast and cross platform tool that enables you
to build shopify themes with ease.

Complete documentation is available at http://themekit.cat`,
	Run: func(cmd *cobra.Command, args []string) {},
}
