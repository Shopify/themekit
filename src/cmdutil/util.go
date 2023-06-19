package cmdutil

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ryanuber/go-glob"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"golang.org/x/sync/errgroup"

	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/file"
	"github.com/Shopify/themekit/src/shopify"
)

// ErrReload is an error to return from a command if you want to reload and run again
var (
	ErrReload        = errors.New("reloading config")
	ErrLiveTheme     = errors.New("cannot make changes to a live theme without an override")
	ErrDuringRuntime = errors.New("finished command with errors")
)

// Flags encapsulates all the possible flags that can be set in the themekit
// command line. Some of the values are used across different commands
type Flags struct {
	ConfigPath                    string
	VariableFilePath              string
	Environments                  []string
	Directory                     string
	Password                      string
	ThemeID                       string
	Domain                        string
	Proxy                         string
	Timeout                       time.Duration
	Verbose                       bool
	DisableUpdateNotifier         bool
	IgnoredFiles                  []string
	Ignores                       []string
	DisableIgnore                 bool
	Notify                        string
	AllEnvs                       bool
	Version                       string
	Prefix                        string
	URL                           string
	Name                          string
	Edit                          bool
	With                          string
	List                          bool
	NoDelete                      bool
	AllowLive                     bool
	Live                          bool
	HidePreviewBar                bool
	DisableThemeKitAccessNotifier bool
}

// Ctx is a specific context that a command will run in
type Ctx struct {
	Shop     shopify.Shop
	Conf     config
	Client   shopifyClient
	Flags    Flags
	Env      *env.Env
	Args     []string
	Log      *log.Logger
	ErrLog   *log.Logger
	progress *mpb.Progress
	Bar      *mpb.Bar
	mu       sync.RWMutex
	summary  cmdSummary
}

type clientFact func(*env.Env) (shopifyClient, error)

func createCtx(newClient clientFact, conf env.Conf, e *env.Env, flags Flags, args []string, progress *mpb.Progress) (*Ctx, error) {
	if e.Proxy != "" {
		colors.ColorStdOut.Printf(
			"[%s] Proxy URL detected from Configuration [%s] SSL Certificate Validation will be disabled!",
			colors.Green(e.Name),
			colors.Yellow(e.Proxy),
		)
	}

	if flags.DisableIgnore {
		e.IgnoredFiles = []string{}
		e.Ignores = []string{}
	}

	client, err := newClient(e)
	if err != nil {
		return &Ctx{}, err
	}

	shop, err := client.GetShop()
	if err != nil && err == shopify.ErrShopDomainNotFound {
		colors.ColorStdErr.Printf(
			"[%s] invalid credentials, the domain %s is not found",
			colors.Green(e.Name),
			colors.Yellow(e.Domain),
		)
		return &Ctx{}, fmt.Errorf("%s is an invalid domain", e.Domain)
	} else if err != nil {
		return &Ctx{}, err
	}

	themes, err := client.Themes() // this will make sure our token is correct
	if err != nil {
		return &Ctx{}, err
	}

	for _, theme := range themes {
		if theme.Role == "main" {
			if fmt.Sprintf("%v", theme.ID) == e.ThemeID && flags.AllowLive {
				colors.ColorStdOut.Printf(
					"[%s] Warning, this is the live theme on %s.",
					colors.Yellow(e.Name),
					colors.Yellow(shop.Name),
				)
			} else if fmt.Sprintf("%v", theme.ID) == e.ThemeID && !flags.AllowLive {
				colors.ColorStdOut.Printf(
					"[%s] This is the live theme on %s. If you wish to make changes to it, then you will have to pass the --allow-live flag",
					colors.Red(e.Name),
					colors.Yellow(shop.Name),
				)
				return &Ctx{}, ErrLiveTheme
			}
			break
		}
	}

	return &Ctx{
		Shop:     shop,
		Conf:     &conf,
		Client:   client,
		Env:      e,
		Flags:    flags,
		Args:     args,
		progress: progress,
		Log:      colors.ColorStdOut,
		ErrLog:   colors.ColorStdErr,
		summary:  cmdSummary{},
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

// Err acts like Printf but will display error messages better
func (ctx *Ctx) Err(msg string, inter ...interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.summary.err(fmt.Sprintf(msg, inter...))
	if ctx.progress == nil || ctx.Bar == nil {
		ctx.ErrLog.Printf(msg, inter...)
	}
}

// DoneTask will mark one unit of work complete. If the context has a progress bar
// then it will increment it.
func (ctx *Ctx) DoneTask(op file.Op) {
	if !ctx.Flags.Verbose && ctx.Bar != nil {
		ctx.Bar.Increment()
	}
	ctx.summary.completeOp(op)
}

// DisableSummary will ensure that the file operation summary will not output at
// the end of the operation
func (ctx *Ctx) DisableSummary() {
	ctx.summary.disable()
}

func generateContexts(newClient clientFact, progress *mpb.Progress, flags Flags, args []string) ([]*Ctx, error) {
	ctxs := []*Ctx{}
	flagEnv := getFlagEnv(flags)

	if err := env.SourceVariables(flags.VariableFilePath); err != nil {
		return ctxs, err
	}

	config, err := env.Load(flags.ConfigPath)
	if err != nil && os.IsNotExist(err) {
		colors.ColorStdOut.Printf(
			"[%s] Could not find config file at %v",
			colors.Yellow("warn"),
			colors.Yellow(flags.ConfigPath),
		)
	} else if err != nil {
		return ctxs, err
	}

	for _, name := range expandEnvironments(flags, config.Envs) {
		e, err := config.Get(name, flagEnv)
		if err != nil && err != env.ErrEnvDoesNotExist {
			return ctxs, err
		} else if e == nil {
			if e, err = config.Set(name, flagEnv); err != nil {
				return ctxs, err
			}
		}

		ctx, err := createCtx(newClient, config, e, flags, args, progress)
		if err != nil {
			return ctxs, err
		}

		ctxs = append(ctxs, ctx)
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
		Notify:    flags.Notify,
	}

	if !flags.DisableIgnore {
		flagEnv.IgnoredFiles = flags.IgnoredFiles
		flagEnv.Ignores = flags.Ignores
	}

	return flagEnv
}

func expandEnvironments(flags Flags, confEnvs map[string]*env.Env) []string {
	envs := []string{}

	if flags.AllEnvs {
		for env := range confEnvs {
			envs = append(envs, env)
		}
		return envs
	}

	for _, flagEnv := range flags.Environments {
		if strings.Contains(flagEnv, "*") {
			for confEnv := range confEnvs {
				if glob.Glob(flagEnv, confEnv) {
					envs = append(envs, confEnv)
				}
			}
		} else {
			envs = append(envs, flagEnv)
		}
	}

	return envs
}

// ForEachClient will generate a command context for all the available environments
// and run a command in each of those contexts
func ForEachClient(flags Flags, args []string, handler func(*Ctx) error) error {
	return forEachClient(shopifyThemeClientFactory, flags, args, handler)
}

func forEachClient(newClient clientFact, flags Flags, args []string, handler func(*Ctx) error) error {
	progressBarGroup := mpb.New(nil)
	ctxs, err := generateContexts(newClient, progressBarGroup, flags, args)
	if err != nil {
		return err
	}
	var handlerGroup errgroup.Group
	for _, ctx := range ctxs {
		ctx := ctx
		handlerGroup.Go(func() error { return handler(ctx) })
	}
	err = handlerGroup.Wait()
	if err == nil {
		progressBarGroup.Wait()
	}
	if err == ErrReload {
		return forEachClient(newClient, flags, args, handler)
	}
	hasErrors := false
	for _, ctx := range ctxs {
		ctx.summary.display(ctx)
		hasErrors = hasErrors || ctx.summary.hasErrors()
	}
	if err == nil && hasErrors {
		return ErrDuringRuntime
	}
	return err
}

// ForSingleClient will generate a command context for all the available environments,
// and run a command for the first context. If more than one environment was specified,
// then an error will be returned.
func ForSingleClient(flags Flags, args []string, handler func(*Ctx) error) error {
	return forSingleClient(shopifyThemeClientFactory, flags, args, handler)
}

func forSingleClient(newClient clientFact, flags Flags, args []string, handler func(*Ctx) error) error {
	progressBarGroup := mpb.New(nil)
	ctxs, err := generateContexts(newClient, progressBarGroup, flags, args)
	if err != nil {
		return err
	} else if len(ctxs) > 1 {
		return fmt.Errorf("more than one environment specified for a single environment command")
	}
	err = handler(ctxs[0])
	if err == nil {
		progressBarGroup.Wait()
	}
	if err == ErrReload {
		return forSingleClient(newClient, flags, args, handler)
	}
	ctxs[0].summary.display(ctxs[0])
	if err == nil && ctxs[0].summary.hasErrors() {
		return ErrDuringRuntime
	}
	return err
}

// ForDefaultClient will run in a context that runs of any available config including
// defaults
func ForDefaultClient(flags Flags, args []string, handler func(*Ctx) error) error {
	return forDefaultClient(shopifyThemeClientFactory, flags, args, handler)
}

func forDefaultClient(newClient clientFact, flags Flags, args []string, handler func(*Ctx) error) error {
	progressBarGroup := mpb.New(nil)

	if err := env.SourceVariables(flags.VariableFilePath); err != nil {
		return err
	}

	config, err := env.Load(flags.ConfigPath)
	if err != nil && os.IsNotExist(err) {
		config = env.New(flags.ConfigPath)
	} else if err != nil {
		return err
	}

	envName := env.Default.Name
	if len(flags.Environments) > 0 {
		envName = flags.Environments[0]
	}

	var e *env.Env
	flagEnv := getFlagEnv(flags)
	if e, err = config.Get(envName, flagEnv); err != nil {
		e, err = config.Set(envName, flagEnv)
		if err != nil {
			return err
		}
	}

	ctx, err := createCtx(newClient, config, e, flags, args, progressBarGroup)
	if err != nil {
		return err
	}

	err = handler(ctx)
	if err == nil {
		progressBarGroup.Wait()
	}

	ctx.summary.display(ctx)

	if err == nil && ctx.summary.hasErrors() {
		return ErrDuringRuntime
	}

	return err
}

func shopifyThemeClientFactory(e *env.Env) (shopifyClient, error) {
	client, err := shopify.NewClient(e)
	if err != nil {
		return nil, err
	}
	return &client, nil
}
