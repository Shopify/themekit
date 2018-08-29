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

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch directory for changes and update remote theme",
	Long: `Watch is for running in the background while you are making changes to your project.

 run 'theme watch' while you are editing and it will detect create, update and delete events.

 For more documentation please see http://shopify.github.io/themekit/commands/#watch
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForEachClient(flags, args, func(ctx cmdutil.Ctx) error {
			watcher, err := file.NewWatcher(ctx.Env, ctx.Flags.ConfigPath)
			if err != nil {
				return err
			}
			watchEvents, err := watcher.Watch()
			if err != nil {
				return err
			}
			defer watcher.Stop()

			signalChan := make(chan os.Signal)
			signal.Notify(signalChan, os.Interrupt)
			return watch(ctx, watchEvents, signalChan)
		})
	},
}

func watch(ctx cmdutil.Ctx, events chan file.Event, sig chan os.Signal) error {
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
			perform(ctx, event.Path, event.Op)
		case <-sig:
			return nil
		}
	}
}

func perform(ctx cmdutil.Ctx, path string, op file.Op) {
	defer ctx.DoneTask()

	if op == file.Remove {
		if err := ctx.Client.DeleteAsset(shopify.Asset{Key: path}); err != nil {
			ctx.ErrLog.Printf("[%s] (%s) %s", colors.Green(ctx.Env.Name), colors.Blue(path), err)
		} else if ctx.Flags.Verbose {
			ctx.Log.Printf("[%s] Deleted %s", colors.Green(ctx.Env.Name), colors.Blue(path))
		}
	} else {
		asset, err := shopify.ReadAsset(ctx.Env, path)
		if err != nil {
			ctx.ErrLog.Printf("[%s] error loading %s: %s", colors.Green(ctx.Env.Name), colors.Green(path), colors.Red(err))
			return
		}

		if err := ctx.Client.UpdateAsset(asset); err != nil {
			ctx.ErrLog.Printf("[%s] (%s) %s", colors.Green(ctx.Env.Name), colors.Blue(asset.Key), err)
		} else if ctx.Flags.Verbose {
			ctx.Log.Printf("[%s] Updated %s", colors.Green(ctx.Env.Name), colors.Blue(asset.Key))
		}
	}
}
