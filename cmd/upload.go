package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb"
	"golang.org/x/sync/errgroup"

	"github.com/Shopify/themekit/cmd/ystore"
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

func deploy(destructive bool) arbitratedCmd {
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

		var deployGroup errgroup.Group
		bar := arbiter.newProgressBar(len(actions), client.Config.Environment)
		for key, action := range actions {
			shouldPerform := arbiter.force || arbiter.manifest.Should(action.event, action.asset.Key, client.Config.Environment)
			// pretend we did the settings data and we will do it last
			if !shouldPerform || key == settingsDataKey {
				incBar(bar)
				continue
			}
			action := action
			deployGroup.Go(func() error {
				if err := perform(client, action.asset, action.event, bar); err != nil {
					kit.LogErrorf("[%s] %s", kit.GreenText(client.Config.Environment), err)
				}
				return nil
			})
		}
		return deployGroup.Wait()
	}
}

func incBar(bar *mpb.Bar) {
	if bar != nil {
		defer bar.Incr(1)
	}
}

func perform(client kit.ThemeClient, asset kit.Asset, event kit.EventType, bar *mpb.Bar) error {
	defer incBar(bar)

	var (
		resp    *kit.ShopifyResponse
		err     error
		version string
	)

	if arbiter.force {
		resp, err = client.Perform(asset, event)
	} else if version, err = arbiter.manifest.Get(asset.Key, client.Config.Environment); err == nil {
		resp, err = client.PerformStrict(asset, event, version)
	}

	if err != nil {
		return err
	} else if arbiter.verbose {
		kit.Printf(
			"[%s] Successfully performed %s on file %s from %s",
			kit.GreenText(client.Config.Environment),
			kit.GreenText(resp.EventType),
			kit.GreenText(resp.Asset.Key),
			kit.YellowText(resp.Host),
		)
	}

	if event == kit.Remove {
		if err := arbiter.manifest.Delete(resp.Asset.Key, client.Config.Environment); err != nil && err != ystore.ErrorCollectionNotFound {
			return err
		}
	} else if err := arbiter.manifest.Set(resp.Asset.Key, client.Config.Environment, resp.Asset.UpdatedAt); err != nil {
		return err
	}

	return nil
}

func uploadSettingsData(client kit.ThemeClient, files []string) error {
	i := sort.Search(len(files), func(i int) bool { return files[i] == settingsDataKey })
	if len(files) > 0 && i == len(files) {
		return nil
	}
	asset, err := client.LocalAsset(settingsDataKey)
	if err != nil {
		return err
	}
	perform(client, asset, kit.Update, nil)
	return nil
}
