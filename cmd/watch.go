package cmd

import (
	"sync"

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

		wg := sync.WaitGroup{}
		for _, client := range themeClients {
			config := client.GetConfiguration()
			client.Message("Spawning %d workers for %s", config.Concurrency, kit.GreenText(config.Domain))
			assetEvents := client.NewFileWatcher(notifyFile)
			for i := 0; i < config.Concurrency; i++ {
				wg.Add(1)
				go spawnWorker(assetEvents, client, &wg)
				client.Message("%s Worker #%d ready to upload local changes", kit.GreenText(config.Domain), i)
			}
		}
		wg.Wait()
		return nil
	},
}

func spawnWorker(assetEvents chan kit.AssetEvent, client kit.ThemeClient, wg *sync.WaitGroup) {
	for {
		event, more := <-assetEvents
		if !more {
			wg.Done()
			return
		}
		if event.Asset().IsValid() || event.Type() == kit.Remove {
			client.Message("Received %s event on %s", kit.GreenText(event.Type().String()), kit.BlueText(event.Asset().Key))
			client.Perform(event)
		}
	}
}
