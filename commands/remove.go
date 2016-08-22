package commands

import (
	"fmt"
	"os"

	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/theme"
)

// RemoveCommand removes file(s) from theme
func RemoveCommand(args Args, done chan bool)  {
	events := make(chan themekit.AssetEvent)
	logs := args.ThemeClient.Process(events, done)

	mergeEvents(args.EventLog, []chan themekit.ThemeEvent{logs})

	go func() {
		for _, filename := range args.Filenames {
			asset := theme.Asset{Key: filename}
			events <- themekit.NewRemovalEvent(asset)
			removeFile(filename)
		}
		close(events)
	}()
}

func removeFile(filename string) error {
	dir, err := os.Getwd()
	err = os.Remove(fmt.Sprintf("%s/%s", dir, filename))
	return err
}
