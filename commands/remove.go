package commands

import (
	"fmt"
	"os"

	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/theme"
)

// RemoveCommand removes file(s) from theme
func RemoveCommand(args Args, done chan bool)  {
	// events := make(chan themekit.AssetEvent)
	// logs := args.ThemeClient.ProcessSync(events)

	// mergeEvents(args.EventLog, []chan themekit.ThemeEvent{logs})

	// events := []themekit.SimpleAssetEvent
	// go func() {
		for _, filename := range args.Filenames {
			asset := theme.Asset{Key: filename}
			event := themekit.NewRemovalEvent(asset)
			args.ThemeClient.ProcessSync([]themekit.AssetEvent{event})
			removeFile(filename)
		}
		done <- true
		// close(events)
	// }()
}

func removeFile(filename string) error {
	dir, err := os.Getwd()
	err = os.Remove(fmt.Sprintf("%s/%s", dir, filename))
	return err
}
