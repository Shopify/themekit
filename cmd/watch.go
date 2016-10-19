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
		if err := initializeConfig(cmd.Name(), false); err != nil {
			return err
		}

		for _, client := range themeClients {
			config := client.GetConfiguration()
			kit.Printf("Watching for file changes for theme %v on host %s ", kit.GreenText(config.ThemeID), kit.YellowText(config.Domain))
			_, err := client.NewFileWatcher(notifyFile, handleWatchEvent)
			if err != nil {
				return err
			}
		}
		<-make(chan int)
		return nil
	},
}

func handleWatchEvent(client kit.ThemeClient, event kit.AssetEvent, err error) {
	if event.Asset.IsValid() || event.Type == kit.Remove {
		kit.Printf(
			"Received %s event on %s",
			kit.GreenText(event.Type.String()),
			kit.BlueText(event.Asset.Key),
		)
		client.Perform(event, func(resp *kit.ShopifyResponse, err kit.Error) {
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
		})
	}
}
