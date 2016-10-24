package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch directory for changes and update remote theme",
	Long: `Watch is for running in the background while you are making changes to your project.

run 'theme watch' while you are editing and it will detect create, update and delete events. `,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(); err != nil {
			return err
		}
		for _, client := range themeClients {
			kit.Printf("Watching for file changes for theme %v on host %s ", kit.GreenText(client.Config.ThemeID), kit.YellowText(client.Config.Domain))
			if _, err := client.NewFileWatcher(notifyFile, handleWatchEvent); err != nil {
				return err
			}
		}
		<-make(chan int)
		return nil
	},
}

func handleWatchEvent(client kit.ThemeClient, asset kit.Asset, event kit.EventType, err error) {
	kit.Printf(
		"Received %s event on %s",
		kit.GreenText(event),
		kit.BlueText(asset.Key),
	)
	resp, err := client.Perform(asset, event)
	if err != nil {
		kit.LogError(err)
	} else {
		kit.Printf(
			"Successfully performed %s operation for file %s to %s",
			kit.GreenText(resp.EventType),
			kit.BlueText(resp.Asset.Key),
			kit.YellowText(resp.Host),
		)
	}
}
