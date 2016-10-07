package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

var removeCmd = &cobra.Command{
	Use:   "remove <filenames>",
	Short: "Remove theme file(s) from shopify",
	Long:  `Remove will delete all specified files from shopify servers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd.Name(), true); err != nil {
			return err
		}

		wg := sync.WaitGroup{}
		for _, client := range themeClients {
			wg.Add(1)
			go remove(client, args, &wg)
		}
		wg.Wait()
		return nil
	},
}

func remove(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) {
	jobQueue := client.Process(wg)
	go func() {
		for _, filename := range filenames {
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
