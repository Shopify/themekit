package commands

import (
	"fmt"
	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/theme"
	"os"
)

type RemoveOptions struct {
	BasicOptions
}

func RemoveCommand(args map[string]interface{}) chan bool {
	options := RemoveOptions{}
	extractThemeClient(&options.Client, args)
	extractEventLog(&options.EventLog, args)
	options.Filenames = extractStringSlice("filenames", args)

	return Remove(options)
}

func Remove(options RemoveOptions) chan bool {
	events := make(chan themekit.AssetEvent)
	done, logs := options.Client.Process(events)

	mergeEvents(options.getEventLog(), []chan themekit.ThemeEvent{logs})

	go func() {
		for _, filename := range options.Filenames {
			asset := theme.Asset{Key: filename}
			events <- themekit.NewRemovalEvent(asset)
			removeFile(filename)
		}
		close(events)
	}()

	return done
}

func removeFile(filename string) error {
	dir, err := os.Getwd()
	err = os.Remove(fmt.Sprintf("%s/%s", dir, filename))
	return err
}
