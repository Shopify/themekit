package cmd

import (
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var (
	signalChan  = make(chan os.Signal)
	isReloading = false
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch directory for changes and update remote theme",
	Long: `Watch is for running in the background while you are making changes to your project.

run 'theme watch' while you are editing and it will detect create, update and delete events.

For more documentation please see http://shopify.github.io/themekit/commands/#watch
`,
	RunE: reloadableWatch,
}

func reloadableWatch(cmd *cobra.Command, args []string) error {
	themeClients, err := generateThemeClients()
	if err != nil {
		return err
	}
	err = watch(themeClients)
	if err != nil {
		return err
	} else if isReloading {
		isReloading = false
		return reloadableWatch(cmd, args)
	}
	return nil
}

func watch(themeClients []kit.ThemeClient) error {
	watchers := []*kit.FileWatcher{}
	defer func() {
		kit.Print("Cleaning up watchers")
		for _, watcher := range watchers {
			watcher.StopWatching()
		}
	}()

	for _, client := range themeClients {
		kit.Printf("Watching for file changes for theme %v on host %s ", kit.GreenText(client.Config.ThemeID), kit.YellowText(client.Config.Domain))
		watcher, err := client.NewFileWatcher(notifyFile, handleWatchEvent)
		if err != nil {
			return err
		}
		watchers = append(watchers, watcher)
	}

	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	return nil
}

func handleWatchEvent(client kit.ThemeClient, asset kit.Asset, event kit.EventType, err error) {
	if err != nil && asset.Key != configPath {
		kit.LogError(err)
		return
	}

	if asset.Key == configPath {
		if !isReloading {
			isReloading = true
			kit.LogWarnf("Config.yml was changed, reloading environment.")
			signalChan <- os.Interrupt
		}
		return
	}

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
