package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/bucket"
)

// Args is a struct containing fields, set via CLI args, that are used by various themekit Commands
type Args struct {
	Console      *themekit.Console
	EventLog     chan themekit.ThemeEvent
	Environments themekit.Environments
	ThemeClient  themekit.ThemeClient
	ThemeClients []themekit.ThemeClient
	Filenames    []string
	AccessToken  string
	Password     string
	Environment  string
	Directory    string
	Domain       string
	NotifyFile   string
	Prefix       string
	Version      string
	ThemeID      string
	SetThemeID   bool
	BucketSize   int
	RefillRate   int
	Bucket       *bucket.LeakyBucket
	Timeout      time.Duration

	WorkingDirGetter WorkingDirGetterType
}

// DefaultArgs returns an instance of Args, initialized with defaults
func DefaultArgs() Args {
	currentDir, _ := os.Getwd()

	return Args{
		Domain:           "",
		AccessToken:      "",
		Directory:        currentDir,
		Environment:      themekit.DefaultEnvironment,
		BucketSize:       themekit.DefaultBucketSize,
		RefillRate:       themekit.DefaultRefillRate,
		Timeout:          themekit.DefaultTimeout,
		WorkingDirGetter: os.Getwd,
	}
}

// WorkingDirGetterType functions fulfills interface of os.Getwd(), used in testing
type WorkingDirGetterType func() (string, error)

// DefaultConfigurationOptions returns a default themekit.Configuration using fields from an Args instance
func (args Args) DefaultConfigurationOptions() themekit.Configuration {
	accessToken := args.AccessToken
	if args.AccessToken == "" {
		accessToken = args.Password
	}

	return themekit.Configuration{
		Domain:      args.Domain,
		AccessToken: accessToken,
		Password:    accessToken,
		BucketSize:  args.BucketSize,
		RefillRate:  args.RefillRate,
		Timeout:     args.Timeout,
	}
}

// ConfigurationErrors returns an error for the first invalid field detected on an Args
func (args Args) ConfigurationErrors() error {
	var errs = []string{}
	if len(args.Domain) <= 0 {
		errs = append(errs, "\t-domain cannot be blank")
	}
	if len(args.AccessToken) <= 0 && len(args.Password) <= 0 {
		errs = append(errs, "\t-password or access_token cannot be blank")
	}
	if len(errs) > 0 {
		fullPath := filepath.Join(args.Directory, "config.yml")
		return fmt.Errorf("Cannot create %s!\nErrors:\n%s", fullPath, strings.Join(errs, "\n"))
	}
	return nil
}
