package kit

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/imdario/mergo"
)

// Configuration is the structure of a configuration for an environment. This will
// get loaded into a theme client to dictate it's actions.
type Configuration struct {
	Environment  string        `yaml:"-" json:"-" env:"-"`
	Password     string        `yaml:"password,omitempty" json:"password,omitempty" env:"THEMEKIT_PASSWORD"`
	ThemeID      string        `yaml:"theme_id,omitempty" json:"theme_id,omitempty" env:"THEMEKIT_THEME_ID"`
	Domain       string        `yaml:"store" json:"store" env:"THEMEKIT_STORE"`
	Directory    string        `yaml:"-" json:"-" env:"THEMEKIT_DIRECTORY"`
	IgnoredFiles []string      `yaml:"ignore_files,omitempty" json:"ignore_files,omitempty" env:"THEMEKIT_IGNORE_FILES" envSeparator:":"`
	Proxy        string        `yaml:"proxy,omitempty" json:"proxy,omitempty" env:"THEMEKIT_PROXY"`
	Ignores      []string      `yaml:"ignores,omitempty" json:"ignores,omitempty" env:"THEMEKIT_IGNORES" envSeparator:":"`
	Timeout      time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" env:"THEMEKIT_TIMEOUT"`
}

// DefaultTimeout is the default timeout to kill any stalled processes.
const DefaultTimeout = 30 * time.Second

var (
	defaultConfig     = Configuration{}
	environmentConfig = Configuration{}
	flagConfig        = Configuration{}
)

func init() {
	pwd, _ := os.Getwd()

	defaultConfig = Configuration{
		Directory: pwd,
		Timeout:   DefaultTimeout,
	}

	env.Parse(&environmentConfig)
}

// SetFlagConfig will set the configuration that is set by your applications flags.
// Set the flag config before inializing any theme clients so that the loaded
// configurations will have the proper config precedence.
func SetFlagConfig(config Configuration) {
	flagConfig = config
}

// NewConfiguration will format a Configuration that combines the config from env variables,
// flags. Then it will validate that config. It will return the
// formatted configuration along with any validation errors. The config precedence
// is flags, environment variables, then the config file.
func NewConfiguration() (*Configuration, error) {
	return (&Configuration{}).compile()
}

func (conf *Configuration) compile() (*Configuration, error) {
	newConfig := &Configuration{}
	mergo.Merge(newConfig, &flagConfig)
	mergo.Merge(newConfig, &environmentConfig)
	mergo.Merge(newConfig, conf)
	mergo.Merge(newConfig, &defaultConfig)
	return newConfig, newConfig.Validate()
}

// Validate will check the configuration for any problems that will cause theme kit
// to function incorrectly.
func (conf Configuration) Validate() error {
	errors := []string{}

	if conf.ThemeID == "" {
		errors = append(errors, "missing theme_id")
	} else if !conf.IsLive() {
		if _, err := strconv.ParseInt(conf.ThemeID, 10, 64); err != nil {
			errors = append(errors, "invalid theme_id")
		}
	}

	if err := conf.validateNoThemeID(); err != nil {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("Invalid configuration: %v", strings.Join(errors, ","))
	}
	return nil
}

func (conf Configuration) validateNoThemeID() error {
	errors := []string{}

	if len(conf.Domain) == 0 {
		errors = append(errors, "missing store domain")
	} else if !strings.HasSuffix(conf.Domain, "myshopify.com") &&
		!strings.HasSuffix(conf.Domain, "myshopify.io") &&
		!strings.HasPrefix(conf.Domain, "http://127.0.0.1:") {
		errors = append(errors, "invalid store domain must end in '.myshopify.com'")
	}

	if len(conf.Password) == 0 {
		errors = append(errors, "missing password")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, ","))
	}
	return nil
}

// IsLive will return true if the configurations theme id is set to live
func (conf Configuration) IsLive() bool {
	return strings.ToLower(strings.TrimSpace(conf.ThemeID)) == "live"
}

// String will return a formatted string with the information about this configuration
func (conf Configuration) String() string {
	return fmt.Sprintf(`
Password     %v
ThemeID      %v
Domain       %v
Directory    %v
IgnoredFiles %v
Proxy        %v
Ignores      %v
Timeout      %v
	`,
		conf.Password,
		conf.ThemeID,
		conf.Domain,
		conf.Directory,
		conf.IgnoredFiles,
		conf.Proxy,
		conf.Ignores,
		conf.Timeout)
}
