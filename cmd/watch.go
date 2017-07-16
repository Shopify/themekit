package cmd

import (
	"errors"
	"fmt"
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
	arbiter.verbose = true
	if err := arbiter.generateThemeClients(cmd, args); err != nil {
		return err
	}
	if err := watch(); err == errReload {
		stdOut.Print("Reloading because of config changes")
		return startWatch(cmd, args)
	} else if err != nil {
		return err
	}
	return nil
}

func watch() error {
	watchers := []*kit.FileWatcher{}
	defer func() {
		if len(watchers) > 0 {
			stdOut.Print("Cleaning up watchers")
			for _, watcher := range watchers {
				// This is half assed because fsnotify sometimes deadlocks
				// if it finishes before exit great if not garbage collection will do it.
				go watcher.StopWatching()
			}
		}
	}()

	for _, client := range arbiter.activeThemeClients {
		if client.Config.ReadOnly {
			stdErr.Printf("[%s] environment is reaonly", green(client.Config.Environment))
			continue
		}

		stdOut.Printf(
			"[%s] Watching for file changes on host %s ",
			green(client.Config.Environment),
			yellow(client.Config.Domain),
		)
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

	if len(watchers) == 0 {
		return fmt.Errorf("no valid configuration to start watch")
	}

	signal.Notify(signalChan, os.Interrupt)
	select {
	case <-signalChan:
	case <-reloadSignal:
		return errReload
	}

	return nil
}

func handleWatchEvent(client kit.ThemeClient, asset kit.Asset, event kit.EventType) {
	stdOut.Printf(
		"[%s] Received %s event on %s",
		green(client.Config.Environment),
		green(event),
		blue(asset.Key),
	)
	perform(client, asset, event, nil)
}
