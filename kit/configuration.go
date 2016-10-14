package kit

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v1"
)

// Configuration is the structure of a configuration for an environment. This will
// get loaded into a theme client to dictate it's actions.
type Configuration struct {
	Password     string        `yaml:"password,omitempty"`
	ThemeID      string        `yaml:"theme_id,omitempty"`
	Domain       string        `yaml:"store"`
	Directory    string        `yaml:"directory,omitempty"`
	IgnoredFiles []string      `yaml:"ignore_files,omitempty"`
	BucketSize   int           `yaml:"bucket_size"`
	RefillRate   int           `yaml:"refill_rate"`
	Concurrency  int           `yaml:"concurrency,omitempty"`
	Proxy        string        `yaml:"proxy,omitempty"`
	Ignores      []string      `yaml:"ignores,omitempty"`
	Timeout      time.Duration `yaml:"timeout,omitempty"`
}

const (
	// DefaultBucketSize is the default maximum amount of processes to run at the same time.
	DefaultBucketSize int = 40
	// DefaultRefillRate is the rate in which processes are allowed to spawn at a time.
	DefaultRefillRate int = 2
	// DefaultConcurrency is the default amount of workers that will be spawned for any job.
	DefaultConcurrency int = 2
	// DefaultTimeout is the default timeout to kill any stalled processes.
	DefaultTimeout = 30 * time.Second
)

var (
	defaultConfig     = Configuration{}
	environmentConfig = Configuration{}
	flagConfig        = Configuration{}
)

func init() {
	pwd, _ := os.Getwd()

	defaultConfig = Configuration{
		Directory:   pwd,
		BucketSize:  DefaultBucketSize,
		RefillRate:  DefaultRefillRate,
		Concurrency: DefaultConcurrency,
		Timeout:     DefaultTimeout,
	}

	environmentConfig = Configuration{
		Password:  os.Getenv("THEMEKIT_PASSWORD"),
		ThemeID:   os.Getenv("THEMEKIT_THEME_ID"),
		Domain:    os.Getenv("THEMEKIT_STORE"),
		Directory: os.Getenv("THEMEKIT_DIR"),
		Proxy:     os.Getenv("THEMEKIT_PROXY"),
	}

	environmentConfig.BucketSize, _ = strconv.Atoi(os.Getenv("THEMEKIT_BUCKET_SIZE"))
	environmentConfig.RefillRate, _ = strconv.Atoi(os.Getenv("THEMEKIT_REFILL_RATE"))
	environmentConfig.Concurrency, _ = strconv.Atoi(os.Getenv("THEMEKIT_CONCURRENCY"))

	if ignoredFiles := os.Getenv("THEMEKIT_IGNORE_FILES"); len(ignoredFiles) > 0 {
		environmentConfig.IgnoredFiles = strings.Split(ignoredFiles, ",")
	}

	if ignores := os.Getenv("THEMEKIT_IGNORES"); len(ignores) > 0 {
		environmentConfig.Ignores = strings.Split(ignores, ",")
	}

	if timeout := os.Getenv("THEMEKIT_TIMEOUT"); timeout != "" {
		environmentConfig.Timeout, _ = time.ParseDuration(timeout)
	}
}

func SetFlagConfig(config Configuration) {
	flagConfig = config
}

// LoadConfiguration will build a configuration object form a raw byte array.
func LoadConfiguration(location string) (Configuration, error) {
	var conf Configuration
	contents, err := ioutil.ReadFile(location)
	if err != nil {
		return conf, err
	}

	err = yaml.Unmarshal(contents, &conf)
	if err != nil {
		return conf, err
	}

	return conf.Initialize()
}

// Initialize will format a Configuration that combines the config from env variables,
// flags and the config file. Then it will validate that config. It will return the
// formatted configuration along with any validation errors.
func (conf Configuration) Initialize() (Configuration, error) {
	newConfig := Configuration{}
	mergo.Merge(&newConfig, &flagConfig)
	mergo.Merge(&newConfig, &environmentConfig)
	mergo.Merge(&newConfig, &conf)
	mergo.Merge(&newConfig, &defaultConfig)
	fmt.Println(newConfig.IgnoredFiles)
	return newConfig, newConfig.Validate()
}

func (conf Configuration) Validate() error {
	errors := []string{}

	if _, err := strconv.ParseInt(conf.ThemeID, 10, 64); !conf.IsLive() && err != nil {
		errors = append(errors, "missing theme_id.")
	}

	if len(conf.Domain) == 0 {
		errors = append(errors, "missing domain")
	} else if !strings.HasSuffix(conf.Domain, "myshopify.com") && !strings.HasSuffix(conf.Domain, "myshopify.io") {
		errors = append(errors, "invalid domain, must end in '.myshopify.com'")
	}

	if len(conf.Password) == 0 {
		errors = append(errors, "missing password")
	}

	if len(errors) > 0 {
		return fmt.Errorf("Invalid configuration: %v", strings.Join(errors, ","))
	}
	return nil
}

// AdminURL will return the url to the shopify admin.
func (conf Configuration) AdminURL() string {
	url := fmt.Sprintf("https://%s/admin", conf.Domain)
	if themeID, err := strconv.ParseInt(conf.ThemeID, 10, 64); !conf.IsLive() && err == nil {
		url = fmt.Sprintf("%s/themes/%d", url, themeID)
	}
	return url
}

func (conf Configuration) IsLive() bool {
	return strings.ToLower(strings.TrimSpace(conf.ThemeID)) != "live"
}

// Write will write out a configuration to a writer.
func (conf Configuration) Write(w io.Writer) error {
	bytes, err := yaml.Marshal(conf)
	if err == nil {
		_, err = w.Write(bytes)
	}
	return err
}

// Save will write out the configuration to a file.
func (conf Configuration) Save(location string) error {
	file, err := os.OpenFile(location, os.O_WRONLY|os.O_CREATE, 0644)
	defer file.Close()
	if err == nil {
		err = conf.Write(file)
	}
	return err
}

// AssetPath will return the assets endpoint in the admin section of shopify.
func (conf Configuration) AssetPath() string {
	return fmt.Sprintf("%s/assets.json", conf.AdminURL())
}

// AddHeaders will add api headers to an http.Requests so that it is a valid request.
func (conf Configuration) AddHeaders(req *http.Request) {
	req.Header.Add("X-Shopify-Access-Token", conf.Password)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("go/themekit (%s; %s)", runtime.GOOS, runtime.GOARCH))
}

// String will return a formatted string with the information about this configuration
func (conf Configuration) String() string {
	return fmt.Sprintf("<token:%s domain:%s bucket:%d refill:%d url:%s>", conf.Password, conf.Domain, conf.BucketSize, conf.RefillRate, conf.AdminURL())
}
