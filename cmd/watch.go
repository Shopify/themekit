package cmd

import (
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var signalChan = make(chan os.Signal)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch directory for changes and update remote theme",
	Long: `Watch is for running in the background while you are making changes to your project.

run 'theme watch' while you are editing and it will detect create, update and delete events.

For more documentation please see http://shopify.github.io/themekit/commands/#watch
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		themeClients, err := generateThemeClients()
		if err != nil {
			return err
		}
		return watch(themeClients)
	},
}

func watch(themeClients []kit.ThemeClient) error {
	watchers := []*kit.FileWatcher{}
	defer func() {
		if len(watchers) > 0 {
			kit.Print("Cleaning up watchers")
			for _, watcher := range watchers {
				watcher.StopWatching()
			}
		}
	}()

	for _, client := range themeClients {
		if client.Config.ReadOnly {
			kit.LogErrorf("[%s]environment is reaonly", kit.GreenText(client.Config.Environment))
			continue
		}

		kit.Printf("[%s] Watching for file changes on host %s ", kit.GreenText(client.Config.Environment), kit.YellowText(client.Config.Domain))
		watcher, err := client.NewFileWatcher(notifyFile, handleWatchEvent)
		if err != nil {
			return err
		}
		watchers = append(watchers, watcher)
	}

	if len(watchers) > 0 {
		signal.Notify(signalChan, os.Interrupt)
		<-signalChan
	}

	return nil
}

func handleWatchEvent(client kit.ThemeClient, asset kit.Asset, event kit.EventType) {
	kit.Printf(
		"[%s] Received %s event on %s",
		kit.GreenText(client.Config.Environment),
		kit.GreenText(event),
		kit.BlueText(asset.Key),
	)
	resp, err := client.Perform(asset, event)
	if err != nil {
		kit.LogErrorf("[%s]%s", kit.GreenText(client.Config.Environment), err)
	} else {
		kit.Printf(
			"[%s] Successfully performed %s operation for file %s to %s",
			kit.GreenText(client.Config.Environment),
			kit.GreenText(resp.EventType),
			kit.BlueText(resp.Asset.Key),
			kit.YellowText(resp.Host),
		)
	}
}
