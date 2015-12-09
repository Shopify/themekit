package commands

import (
	"encoding/base64"
	"fmt"
	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/theme"
	"os"
	"path/filepath"
)

type DownloadOptions struct {
	BasicOptions
}

func DownloadCommand(args map[string]interface{}) chan bool {
	options := DownloadOptions{}
	extractThemeClient(&options.Client, args)
	extractEventLog(&options.EventLog, args)
	options.Filenames = extractStringSlice("filenames", args)

	return Download(options)
}

func Download(options DownloadOptions) (done chan bool) {
	done = make(chan bool)
	eventLog := options.getEventLog()

	if len(options.Filenames) <= 0 {
		assets, errs := options.Client.AssetList()
		go drainErrors(errs)
		go downloadAllFiles(assets, done, eventLog)
	} else {
		go downloadFiles(options.Client.Asset, options.Filenames, done, eventLog)
	}

	return done
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
			etype:     "fsevent",
			Formatter: func(b basicEvent) string {
				return themekit.GreenText(fmt.Sprintf("Successfully wrote %s to disk", b.Target))
			},
		}
		logEvent(event, eventLog)
	}
}

func handleError(filename string, err error, eventLog chan themekit.ThemeEvent) {
	if nonFatal, ok := err.(themekit.NonFatalNetworkError); ok {
		event := basicEvent{
			Title:     "Non-Fatal Network Error",
			EventType: nonFatal.Verb,
			Target:    filename,
			etype:     "fsevent",
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
