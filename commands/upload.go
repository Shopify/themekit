package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/theme"
)

type UploadOptions struct {
	BasicOptions
	Directory string
}

func UploadCommand(args map[string]interface{}) chan bool {
	currentDir, _ := os.Getwd()
	options := UploadOptions{Directory: currentDir}
	extractThemeClient(&options.Client, args)
	extractEventLog(&options.EventLog, args)
	extractString(&options.Directory, "directory", args)
	options.Filenames = extractFilenames(options, extractStringSlice("filenames", args))

	return Upload(options)
}

func Upload(options UploadOptions) chan bool {
	files := make(chan themekit.AssetEvent)
	go readAndPrepareFiles(options.Filenames, files)

	done, events := options.Client.Process(files)
	mergeEvents(options.getEventLog(), []chan themekit.ThemeEvent{events})
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

func extractFilenames(options UploadOptions, filenames []string) []string {
	if len(filenames) > 0 {
		return filenames
	}
	filepath.Walk(options.Directory, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			root := fmt.Sprintf("%s%s", options.Directory, string(filepath.Separator))
			name := strings.Replace(path, root, "", -1)
			filenames = append(filenames, name)
		}
		return nil
	})
	return filenames
}
