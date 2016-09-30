package commands

import (
	"fmt"
	"os"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

// RemoveCommand removes file(s) from theme
func RemoveCommand(args Args, done chan bool) {
	foreman := args.ThemeClient.NewForeman()
	args.ThemeClient.Process(foreman.WorkerQueue, done)
	go func() {
		for _, filename := range args.Filenames {
			asset := theme.Asset{Key: filename}
			foreman.JobQueue <- kit.NewRemovalEvent(asset)
			removeFile(filename)
		}
		close(foreman.JobQueue)
	}()
}

func removeFile(filename string) error {
	dir, err := os.Getwd()
	err = os.Remove(fmt.Sprintf("%s/%s", dir, filename))
	return err
}
