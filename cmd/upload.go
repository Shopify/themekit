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
	PreRunE:  arbiter.generateThemeClients,
	RunE:     arbiter.forEachClient(deploy(false)),
	PostRunE: arbiter.forEachClient(uploadSettingsData),
}

func deploy(destructive bool) arbitratedCommand {
	return func(client kit.ThemeClient, filenames []string) error {
		if client.Config.ReadOnly {
			return fmt.Errorf("[%s] environment is reaonly", kit.GreenText(client.Config.Environment))
		}

		actions, err := arbiter.generateAssetActions(client, filenames, destructive)
		if err != nil {
			return err
		}

		if err := arbiter.preflightCheck(actions, destructive); err != nil {
			return err
		}

		var wg sync.WaitGroup
		bar := arbiter.newProgressBar(len(actions), client.Config.Environment)
		for key, action := range actions {
			shouldPerform := arbiter.force || arbiter.manifest.Should(action.event, action.asset.Key, client.Config.Environment)
			// pretend we did the settings data and we will do it last
			if !shouldPerform || key == settingsDataKey {
				arbiter.cleanupAction(bar, nil)
				continue
			}
			wg.Add(1)
			go perform(client, action.asset, action.event, bar, &wg)
		}
		wg.Wait()
		return nil
	}
}

func perform(client kit.ThemeClient, asset kit.Asset, event kit.EventType, bar *mpb.Bar, wg *sync.WaitGroup) bool {
	defer arbiter.cleanupAction(bar, wg)

	var resp *kit.ShopifyResponse
	var err error

	if arbiter.force {
		resp, err = client.Perform(asset, event)
	} else {
		var version string
		version, err = arbiter.manifest.Get(asset.Key, client.Config.Environment)
		if err != nil {
			kit.LogErrorf("[%s] Cannot get file version %s", kit.GreenText(client.Config.Environment), err)
		}
		resp, err = client.PerformStrict(asset, event, version)
	}

	if err != nil {
		kit.LogErrorf("[%s]%s", kit.GreenText(client.Config.Environment), err)
		return false
	} else if arbiter.verbose {
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
	}
	for _, filename := range filenames {
		if filename == settingsDataKey {
			return doupload()
		}
	}
	return nil
}
