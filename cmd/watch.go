package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/file"
	"github.com/Shopify/themekit/src/shopify"
)

// This is the limit of assets that can be uploaded at the same time. This may
// need to be tweaked in the future.
const assetLimit = 100

// This sets a hard limit on how many assets are loaded at a single time before
// being uploaded. This is to protect from memory errors when very large themes
// are uploaded.
var assetLimitSemaphore = make(chan struct{}, assetLimit)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch directory for changes and update remote theme",
	Long: `Watch is for running in the background while you are making changes to your project.

 run 'theme watch' while you are editing and it will detect create, update and delete events.

 For more information, refer to https://shopify.dev/tools/theme-kit/command-reference#watch.
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForEachClient(flags, args, func(ctx *cmdutil.Ctx) error {
			ctx.DisableSummary()

			checksums := map[string]string{}
			remoteFiles, err := ctx.Client.GetAllAssets()
			if err != nil {
				return fmt.Errorf("[%s] Error while fetching info from server: %v", colors.Green(ctx.Env.Name), err)
			}
			for _, remoteAsset := range remoteFiles {
				checksums[remoteAsset.Key] = remoteAsset.Checksum
			}

			watcher, err := file.NewWatcher(ctx.Env, ctx.Flags.ConfigPath, checksums)
			if err != nil {
				return err
			}
			watcher.Watch()
			defer watcher.Stop()

			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, os.Interrupt)

			notifier := newNotifyAdapter(ctx.Env.Notify)

			return watch(ctx, watcher.Events, signalChan, notifier)
		})
	},
}

func watch(ctx *cmdutil.Ctx, events chan file.Event, sig chan os.Signal, notifier notifyAdapter) error {
	// watch should output every action that it is taking and not use a progress bar
	ctx.Flags.Verbose = true
	ctx.Log.SetFlags(log.Ltime)

	if ctx.Env.ReadOnly {
		return fmt.Errorf("[%s] environment is reaonly", colors.Green(ctx.Env.Name))
	}

	ctx.Log.Printf(
		"[%s] %s: Watching for file changes to theme %v",
		colors.Green(ctx.Env.Name),
		colors.Yellow(ctx.Shop.Name),
		colors.Yellow(ctx.Env.ThemeID),
	)
	for {
		select {
		case event := <-events:
			if event.Path == ctx.Flags.ConfigPath {
				ctx.Log.Print("Reloading config changes")
				return cmdutil.ErrReload
			}
			ctx.Log.Printf("[%s] processing %s", colors.Green(ctx.Env.Name), colors.Blue(event.Path))
			perform(ctx, event.Path, event.Op, event.LastKnownChecksum)
			if event.Op != file.Skip {
				notifier.notify(ctx, event.Path)
			}
		case <-sig:
			return nil
		}
	}
}

func perform(ctx *cmdutil.Ctx, path string, op file.Op, checksum string) {
	defer ctx.DoneTask(op)

	switch op {
	case file.Skip:
		if ctx.Flags.Verbose {
			localAsset, _ := shopify.ReadAsset(ctx.Env, path)
			checksumOutput := "Checksum: " + localAsset.Checksum
			ctx.Log.Printf("[%s] %s %s (%s)", colors.Green(ctx.Env.Name), colors.Cyan("Skipped"), colors.Blue(path), checksumOutput)
		}
	case file.Remove:
		if err := ctx.Client.DeleteAsset(shopify.Asset{Key: path}); err != nil {
			ctx.Err("[%s] (%s) %s", colors.Green(ctx.Env.Name), colors.Blue(path), err)
		} else if ctx.Flags.Verbose {
			ctx.Log.Printf("[%s] Deleted %s", colors.Green(ctx.Env.Name), colors.Blue(path))
		}
	case file.Get:
		if asset, err := ctx.Client.GetAsset(path); err != nil {
			ctx.Err("[%s] error downloading %s: %s", colors.Green(ctx.Env.Name), colors.Blue(path), err)
		} else if err = asset.Write(ctx.Env.Directory); err != nil {
			ctx.Err("[%s] error writing %s: %s", colors.Green(ctx.Env.Name), colors.Blue(asset.Key), err)
		} else if ctx.Flags.Verbose {
			ctx.Log.Printf("[%s] Successfully wrote %s to disk", colors.Green(ctx.Env.Name), colors.Blue(asset.Key))
		}
	default:
		assetLimitSemaphore <- struct{}{}
		defer func() { <-assetLimitSemaphore }()

		asset, err := shopify.ReadAsset(ctx.Env, path)
		if err != nil {
			ctx.Err("[%s] error loading %s: %s", colors.Green(ctx.Env.Name), colors.Green(path), colors.Red(err))
			return
		}

		if err = ctx.Client.UpdateAsset(asset, checksum); err != nil {
			ctx.Err("[%s] (%s) %s", colors.Green(ctx.Env.Name), colors.Blue(asset.Key), err)
		} else if ctx.Flags.Verbose {
			ctx.Log.Printf("[%s] Updated %s", colors.Green(ctx.Env.Name), colors.Blue(asset.Key))
		}
	}
}
