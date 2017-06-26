package cmd

import (
	"errors"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var (
	signalChan   = make(chan os.Signal)
	reloadSignal = make(chan bool)
	errReload    = errors.New("Reload Watcher")
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch directory for changes and update remote theme",
	Long: `Watch is for running in the background while you are making changes to your project.

run 'theme watch' while you are editing and it will detect create, update and delete events.

For more documentation please see http://shopify.github.io/themekit/commands/#watch
`,
	RunE: startWatch,
}

func startWatch(cmd *cobra.Command, args []string) error {
	if err := arbiter.generateThemeClients(); err != nil {
		return err
	}
	if err := watch(arbiter.activeThemeClients); err == errReload {
		kit.Print("Reloading because of config changes")
		return startWatch(cmd, args)
	}
	return nil
}

func watch(themeClients []kit.ThemeClient) error {
	watchers := []*kit.FileWatcher{}
	defer func() {
		if len(watchers) > 0 {
			kit.Print("Cleaning up watchers")
			for _, watcher := range watchers {
				// This is half assed because fsnotify sometimes deadlocks
				// if it finishes before exit great if not garbage collection will do it.
				go watcher.StopWatching()
			}
		}
	}()

	for _, client := range themeClients {
		if client.Config.ReadOnly {
			kit.LogErrorf("[%s]environment is reaonly", kit.GreenText(client.Config.Environment))
			continue
		}

		kit.Printf("[%s] Watching for file changes on host %s ", kit.GreenText(client.Config.Environment), kit.YellowText(client.Config.Domain))
		watcher, err := client.NewFileWatcher(arbiter.notifyFile, handleWatchEvent)
		if err != nil {
			return err
		}
		err = watcher.WatchConfig(arbiter.configPath, reloadSignal)
		if err != nil {
			return err
		}
		watchers = append(watchers, watcher)
	}

	if len(watchers) > 0 {
		signal.Notify(signalChan, os.Interrupt)
		select {
		case <-signalChan:
		case <-reloadSignal:
			return errReload
		}
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

	var resp *kit.ShopifyResponse
	var err error

	if arbiter.force {
		resp, err = client.Perform(asset, event)
	} else {
		var version string
		version, err = arbiter.manifest.Get(asset.Key, client.Config.Environment)
		if err != nil {
			kit.LogErrorf("[%s] Cannot get file version %s", kit.GreenText(client.Config.Environment), err)
		}
		resp, err = client.PerformStrict(asset, event, version)
	}

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
