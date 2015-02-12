package commands

import (
	"fmt"
	"github.com/csaunders/phoenix"
	"os"
)

func RemoveCommand(args map[string]interface{}) chan bool {
	return toClientAndFilesAsync(args, Remove)
}

func Remove(client phoenix.ThemeClient, filenames []string) (done chan bool) {
	var messages chan string
	events := make(chan phoenix.AssetEvent)
	done, messages = client.Process(events)

	go func() {
		for _, filename := range filenames {
			asset := phoenix.Asset{Key: filename}
			events <- phoenix.NewRemovalEvent(asset)
			removeFile(filename)
		}
		close(events)
	}()

	go func() {
		for {
			message, more := <-messages
			if !more {
				return
			}
			fmt.Println(message)
		}
	}()
	return done
}

func removeFile(filename string) error {
	dir, err := os.Getwd()
	err = os.Remove(fmt.Sprintf("%s/%s", dir, filename))
	return err
}
