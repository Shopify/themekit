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

	ctx.Log.Printf("[%s] Watching for file changes on host %s ", colors.Green(ctx.Env.Name), colors.Yellow(ctx.Env.Domain))
	for {
		select {
		case event := <-events:
			if event.Path == ctx.Flags.ConfigPath {
				ctx.Log.Print("Reloading config changes")
				return cmdutil.ErrReload
			}

			ctx.Log.Printf("[%s] processing %s", colors.Green(ctx.Env.Name), colors.Blue(event.Path))
			if event.Op == file.Remove {
				cmdutil.DeleteAsset(ctx, shopify.Asset{Key: event.Path})
			} else {
				asset, err := shopify.ReadAsset(ctx.Env, event.Path)
				if err != nil {
					ctx.ErrLog.Printf("[%s] error loading %s: %s", colors.Green(ctx.Env.Name), colors.Green(event.Path), colors.Red(err))
					continue
				}
				cmdutil.UploadAsset(ctx, asset)
			}
		case <-sig:
			return nil
		}
	}
}
