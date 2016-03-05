package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/theme"
)

// UploadCommand add file(s) to theme
func UploadCommand(args Args) chan bool {
	files := make(chan themekit.AssetEvent)
	go readAndPrepareFiles(args.Filenames, files)

	done, events := args.ThemeClient.Process(files)
	mergeEvents(args.EventLog, []chan themekit.ThemeEvent{events})
	return done
}

func readAndPrepareFiles(filenames []string, results chan themekit.AssetEvent) {
	for _, filename := range filenames {
		asset, err := loadAsset(filename)
		if err == nil {
			results <- themekit.NewUploadEvent(asset)
		} else if err.Error() != "File is a directory" {
			themekit.NotifyError(err)
		}
	}
	close(results)
}

func loadAsset(filename string) (asset theme.Asset, err error) {
	root, err := os.Getwd()
	if err != nil {
		return
	}

	return theme.LoadAsset(root, filename)
}

func extractFilenames(args Args, filenames []string) []string {
	if len(filenames) > 0 {
		return filenames
	}
	filepath.Walk(args.Directory, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			root := fmt.Sprintf("%s%s", args.Directory, string(filepath.Separator))
			name := strings.Replace(path, root, "", -1)
			filenames = append(filenames, name)
		}
		return nil
	})
	return filenames
}
