package commands

import (
	"fmt"
	"os"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

// RemoveCommand removes file(s) from theme
func RemoveCommand(args Args, done chan bool) {
	jobQueue := args.ThemeClient.Process(done)
	go func() {
		for _, filename := range args.Filenames {
			asset := theme.Asset{Key: filename}
			jobQueue <- kit.NewRemovalEvent(asset)
			removeFile(filename)
		}
		close(jobQueue)
	}()
}

func removeFile(filename string) error {
	dir, err := os.Getwd()
	err = os.Remove(fmt.Sprintf("%s/%s", dir, filename))
	return err
}
