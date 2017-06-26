package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb"

	"github.com/Shopify/themekit/kit"
)

const settingsDataKey = "config/settings_data.json"

var uploadCmd = &cobra.Command{
	Use:   "upload <filenames>",
	Short: "Upload theme file(s) to shopify",
	Long: `Upload will upload specific files to shopify servers if provided file names.
If no filenames are provided then upload will upload every file in the project
to shopify.

For more documentation please see http://shopify.github.io/themekit/commands/#upload
`,
	RunE:     arbiter.forEachClient(upload),
	PostRunE: arbiter.forEachClient(uploadSettingsData),
}

func upload(client kit.ThemeClient, filenames []string) error {
	if client.Config.ReadOnly {
		return fmt.Errorf("[%s] environment is reaonly", kit.GreenText(client.Config.Environment))
	}

	localAssets, err := client.LocalAssets(filenames...)
	if err != nil {
		return fmt.Errorf("[%s] %v", kit.GreenText(client.Config.Environment), err)
	}

	var wg sync.WaitGroup
	bar := arbiter.newProgressBar(len(localAssets), client.Config.Environment)
	for _, asset := range localAssets {
		if asset.Key == settingsDataKey {
			arbiter.incBar(bar) //pretend we did this one we will do it later
			continue
		}
		wg.Add(1)
		go perform(client, asset, kit.Update, bar, &wg)
	}
	wg.Wait()

	return nil
}

func perform(client kit.ThemeClient, asset kit.Asset, event kit.EventType, bar *mpb.Bar, wg *sync.WaitGroup) bool {
	defer func() {
		if wg != nil {
			wg.Done()
		}
		arbiter.incBar(bar)
	}()

	if !arbiter.force && !arbiter.manifest.Should(event, settingsDataKey, client.Config.Environment) {
		if arbiter.verbose {
			kit.Printf(
				"[%s] Skipping %s on file %s because file versions match",
				kit.GreenText(client.Config.Environment),
				kit.GreenText(event),
				kit.GreenText(asset.Key),
			)
		}
		return true
	}

	resp, err := client.Perform(asset, event)
	if err != nil {
		kit.LogErrorf("[%s]%s", kit.GreenText(client.Config.Environment), err)
		return false
	}

	if arbiter.verbose {
		kit.Printf(
			"[%s] Successfully performed %s on file %s from %s",
			kit.GreenText(client.Config.Environment),
			kit.GreenText(resp.EventType),
			kit.GreenText(resp.Asset.Key),
			kit.YellowText(resp.Host),
		)
	}

	var storeErr error
	if event == kit.Remove {
		storeErr = arbiter.manifest.Delete(resp.Asset.Key, client.Config.Environment)
	} else {
		storeErr = arbiter.manifest.Set(resp.Asset.Key, client.Config.Environment, resp.Asset.UpdatedAt)
	}

	return storeErr == nil
}

func uploadSettingsData(client kit.ThemeClient, filenames []string) error {
	if client.Config.ReadOnly {
		return fmt.Errorf("[%s] environment is reaonly", kit.GreenText(client.Config.Environment))
	}

	doupload := func() error {
		asset, err := client.LocalAsset(settingsDataKey)
		if err != nil {
			return err
		}
		perform(client, asset, kit.Update, nil, nil)
		return nil
	}

	if len(filenames) == 0 {
		return doupload()
	} else {
		for _, filename := range filenames {
			if filename == settingsDataKey {
				return doupload()
			}
		}
	}
	return nil
}
