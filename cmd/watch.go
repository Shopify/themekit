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

			kit.Logf("%s watching for file changes", kit.GreenText(config.Domain))
			err := client.NewFileWatcher(notifyFile, handleWatchEvent)
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
		kit.Logf(
			"Received %s event on %s",
			kit.GreenText(event.Type.String()),
			kit.BlueText(event.Asset.Key),
		)
		client.Perform(event, func(resp *kit.ShopifyResponse, err kit.Error) {
			if err != nil {
				kit.Errorf(err.Error())
			} else {
				kit.Logf(
					"Successfully performed %s operation for file %s to %s",
					kit.GreenText(resp.EventType),
					kit.BlueText(resp.Asset.Key),
					kit.YellowText(resp.Host),
				)
			}
		})
	}
}
