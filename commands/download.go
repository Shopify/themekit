package commands

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/theme"
)

// DownloadCommand downloads file(s) from theme
func DownloadCommand(args Args, done chan bool) {
	eventLog := args.EventLog

	if len(args.Filenames) <= 0 {
		assets, errs := args.ThemeClient.AssetList()
		go drainErrors(errs)
		go downloadAllFiles(assets, done, eventLog)
	} else {
		go downloadFiles(args.ThemeClient.Asset, args.Filenames, done, eventLog)
	}
}

func downloadAllFiles(assets chan theme.Asset, done chan bool, eventLog chan themekit.ThemeEvent) {
	for {
		asset, more := <-assets
		if more {
			writeToDisk(asset, eventLog)
		} else {
			done <- true
			return
		}
	}
}

func downloadFiles(retrievalFunction themekit.AssetRetrieval, filenames []string, done chan bool, eventLog chan themekit.ThemeEvent) {
	for _, filename := range filenames {
		if asset, err := retrievalFunction(filename); err != nil {
			handleError(filename, err, eventLog)
		} else {
			writeToDisk(asset, eventLog)
		}
	}
	done <- true
	return
}

func writeToDisk(asset theme.Asset, eventLog chan themekit.ThemeEvent) {
	dir, err := os.Getwd()
	if err != nil {
		themekit.NotifyError(err)
		return
	}

	perms, err := os.Stat(dir)
	if err != nil {
		themekit.NotifyError(err)
		return
	}

	filename := fmt.Sprintf("%s/%s", dir, asset.Key)
	err = os.MkdirAll(filepath.Dir(filename), perms.Mode())
	if err != nil {
		themekit.NotifyError(err)
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		themekit.NotifyError(err)
		return
	}
	defer file.Sync()
	defer file.Close()

	var data []byte
	switch {
	case len(asset.Value) > 0:
		data = []byte(asset.Value)
	case len(asset.Attachment) > 0:
		data, err = base64.StdEncoding.DecodeString(asset.Attachment)
		if err != nil {
			themekit.NotifyError(fmt.Errorf("Could not decode %s. error: %s", asset.Key, err))
			return
		}
	}

	if len(data) > 0 {
		_, err = file.Write(data)
	}

	if err != nil {
		themekit.NotifyError(err)
	} else {
		event := basicEvent{
			Title:     "FS Event",
			EventType: "Write",
			Target:    filename,
			Etype:     "fsevent",
			Formatter: func(b basicEvent) string {
				return themekit.GreenText(fmt.Sprintf("Successfully wrote %s to disk", b.Target))
			},
		}
		logEvent(event, eventLog)
	}
}

// TODO do version of this that doesn't do channel stuff
// TODO generally, just leave DownloadCommand until last
func handleError(filename string, err error, eventLog chan themekit.ThemeEvent) {
	if nonFatal, ok := err.(themekit.NonFatalNetworkError); ok {
		event := basicEvent{
			Title:     "Non-Fatal Network Error",
			EventType: nonFatal.Verb,
			Target:    filename,
			Etype:     "fsevent",
			Formatter: func(b basicEvent) string {
				return fmt.Sprintf(
					"[%s] Could not complete %s for %s",
					themekit.RedText(fmt.Sprintf("%d", nonFatal.Code)),
					themekit.YellowText(b.EventType),
					themekit.BlueText(b.Target),
				)
			},
		}
		logEvent(event, eventLog)
	}
}
