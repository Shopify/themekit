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
func UploadCommand(args Args, done chan bool)  {
	args.Filenames = extractFilenames(args, args.Filenames)
	syncAssetEvents := ReadAndPrepareFilesSync(args)

	go func() {
		args.ThemeClient.ProcessSync(syncAssetEvents, args.EventLog)
		done <- true
	}()
}

// ReadAndPrepareFilesSync ... TODO
func ReadAndPrepareFilesSync(args Args) (results []themekit.AssetEvent) {
	for _, filename := range args.Filenames {
		asset, err := loadAsset(args, filename)

		if err == nil {
			results = append(results, themekit.NewUploadEvent(asset))
		} else if err.Error() != "File is a directory" {
			themekit.NotifyError(err)
		}
	}
	return
}

func loadAsset(args Args, filename string) (asset theme.Asset, err error) {
	root, err := args.WorkingDirGetter()
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
