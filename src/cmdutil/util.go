package cmdutil

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ryanuber/go-glob"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"golang.org/x/sync/errgroup"

	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/shopify"
)

// ErrReload is an error to return from a command if you want to reload and run again
var ErrReload = errors.New("reloading config")

// Flags encapsulates all the possible flags that can be set in the themekit
// command line. Some of the values are used across different commands
type Flags struct {
	ConfigPath            string
	Environments          stringArgArray
	Directory             string
	Password              string
	ThemeID               string
	Domain                string
	Proxy                 string
	Timeout               time.Duration
	Verbose               bool
	DisableUpdateNotifier bool
	IgnoredFiles          stringArgArray
	Ignores               stringArgArray
	DisableIgnore         bool
	NotifyFile            string
	AllEnvs               bool
	Version               string
	Prefix                string
	URL                   string
	Name                  string
	Edit                  bool
	With                  string
}

// Ctx is a specific context that a command will run in
type Ctx struct {
	Conf     config
	Client   shopifyClient
	Flags    Flags
	Env      *env.Env
	Args     []string
	Log      *log.Logger
	ErrLog   *log.Logger
	progress *mpb.Progress
	Bar      *mpb.Bar
}

func createCtx(conf env.Conf, e *env.Env, flags Flags, args []string, progress *mpb.Progress) (Ctx, error) {
	client, err := shopify.NewClient(e)
	if err != nil {
		return Ctx{}, err
	}

	if flags.DisableIgnore {
		e.IgnoredFiles = []string{}
		e.Ignores = []string{}
	}

	return Ctx{
		Conf:     &conf,
		Client:   &client,
		Env:      e,
		Flags:    flags,
		Args:     args,
		progress: progress,
		Log:      colors.ColorStdOut,
		ErrLog:   colors.ColorStdErr,
	}, nil
}

// StartProgress will create a new progress bar for the running context with the
// total amount of tasks as the count
func (ctx *Ctx) StartProgress(count int) {
	if !ctx.Flags.Verbose && ctx.progress != nil {
		ctx.Bar = ctx.progress.AddBar(
			int64(count),
			mpb.PrependDecorators(decor.Name(fmt.Sprintf("[%s] ", ctx.Env.Name)), decor.Counters(0, "%d|%d")),
			mpb.AppendDecorators(decor.Percentage(decor.WCSyncSpace)),
		)
	}
}

// DoneTask will mark one unit of work complete. If the context has a progress bar
// then it will increment it.
func (ctx *Ctx) DoneTask() {
	if !ctx.Flags.Verbose && ctx.Bar != nil {
		ctx.Bar.Increment()
	}
}

func generateContexts(progress *mpb.Progress, flags Flags, args []string) ([]Ctx, error) {
	ctxs := []Ctx{}
	flagEnv := getFlagEnv(flags)

	config, err := env.Load(flags.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ctxs, fmt.Errorf("Could not find config file at %v", flags.ConfigPath)
		}
		return ctxs, err
	}

	for name := range config.Envs {
		if !shouldUseEnvironment(flags, name) {
			continue
		}

		e, err := config.Get(name, flagEnv)
		if err != nil {
			return ctxs, err
		}

		if e.ThemeID == "" {
			colors.ColorStdOut.Printf("[%s] Warning, this is your live theme.", colors.Yellow(name))
		}

		if e.Proxy != "" {
			colors.ColorStdOut.Printf(
				"[%s] Proxy URL detected from Configuration [%s] SSL Certificate Validation will be disabled!",
				colors.Green(name),
				colors.Yellow(e.Proxy),
			)
		}

		ctx, err := createCtx(config, e, flags, args, progress)
		if err != nil {
			return ctxs, err
		}

		ctxs = append(ctxs, ctx)
	}

	if len(ctxs) == 0 {
		return ctxs, fmt.Errorf("Could not load any valid environments")
	}

	return ctxs, nil
}

func getFlagEnv(flags Flags) env.Env {
	flagEnv := env.Env{
		Directory: flags.Directory,
		Password:  flags.Password,
		ThemeID:   flags.ThemeID,
		Domain:    flags.Domain,
		Proxy:     flags.Proxy,
		Timeout:   flags.Timeout,
		Notify:    flags.NotifyFile,
	}

	if !flags.DisableIgnore {
		flagEnv.IgnoredFiles = flags.IgnoredFiles.Value()
		flagEnv.Ignores = flags.Ignores.Value()
	}

	return flagEnv
}

func shouldUseEnvironment(flags Flags, envName string) bool {
	flagEnvs := flags.Environments.Value()
	if flags.AllEnvs || (len(flagEnvs) == 0 && envName == env.Default.Name) {
		return true
	}
	for _, env := range flagEnvs {
		if env == envName || glob.Glob(env, envName) {
			return true
		}
	}
	return false
}

// ForEachClient will generate a command context for all the available environments
// and run a command in each of those contexts
func ForEachClient(flags Flags, args []string, handler func(Ctx) error) error {
	progressBarGroup := mpb.New(nil)
	ctxs, err := generateContexts(progressBarGroup, flags, args)
	if err != nil {
		return err
	}
	var handlerGroup errgroup.Group
	for _, ctx := range ctxs {
		handlerGroup.Go(func() error { return handler(ctx) })
	}
	err = handlerGroup.Wait()
	if err == ErrReload {
		return ForEachClient(flags, args, handler)
	}
	return err
}

// ForSingleClient will generate a command context for all the available environments,
// and run a command for the first context. If more than one environment was specified,
// then an error will be returned.
func ForSingleClient(flags Flags, args []string, handler func(Ctx) error) error {
	progressBarGroup := mpb.New(nil)
	ctxs, err := generateContexts(progressBarGroup, flags, args)
	if err != nil {
		return err
	} else if len(ctxs) > 1 {
		return fmt.Errorf("more than one environment specified for a single environment command")
	}
	err = handler(ctxs[0])
	if err == ErrReload {
		return ForEachClient(flags, args, handler)
	}
	return err
}

// ForDefaultClient will run in a context that runs of any available config including
// defaults
func ForDefaultClient(flags Flags, args []string, handler func(Ctx) error) error {
	progressBarGroup := mpb.New(nil)
	config, err := env.Load(flags.ConfigPath)
	if err != nil && os.IsNotExist(err) {
		config = env.New(flags.ConfigPath)
	} else if err != nil {
		return err
	}

	var e *env.Env
	if e, err = config.Get(env.Default.Name, getFlagEnv(flags)); err != nil {
		e, err = config.Set(env.Default.Name, getFlagEnv(flags))
		if err != nil {
			return err
		}
	}

	ctx, err := createCtx(config, e, flags, args, progressBarGroup)
	if err != nil {
		return err
	}

	return handler(ctx)
}

// UploadAsset will perform an upload for an asset and log success or failure
func UploadAsset(ctx Ctx, asset shopify.Asset) {
	defer ctx.DoneTask()
	if err := ctx.Client.UpdateAsset(asset); err != nil {
		ctx.ErrLog.Printf("[%s] %s", colors.Green(ctx.Env.Name), err)
	} else if ctx.Flags.Verbose {
		ctx.Log.Printf("[%s] Updated %s", colors.Green(ctx.Env.Name), colors.Blue(asset.Key))
	}
}

// DeleteAsset will perform a delete for an asset and log success or failure
func DeleteAsset(ctx Ctx, asset shopify.Asset) {
	defer ctx.DoneTask()
	if err := ctx.Client.DeleteAsset(asset); err != nil {
		ctx.ErrLog.Printf("[%s] %s", colors.Green(ctx.Env.Name), err)
	} else if ctx.Flags.Verbose {
		ctx.Log.Printf("[%s] Deleted %s", colors.Green(ctx.Env.Name), colors.Blue(asset.Key))
	}
}
