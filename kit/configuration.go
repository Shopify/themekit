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

	"gopkg.in/yaml.v1"
)

// Configuration is the structure of a configuration for an environment. This will
// get loaded into a theme client to dictate it's actions.
type Configuration struct {
	Password     string        `yaml:"password,omitempty"`
	ThemeID      string        `yaml:"theme_id,omitempty"`
	Domain       string        `yaml:"store"`
	URL          string        `yaml:"-"`
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
	// DefaultTimeoutInt in integer is the default timeout to kill any stalled processes.
	DefaultTimeoutInt int = 30
	// DefaultTimeout is the default timeout to kill any stalled processes.
	DefaultTimeout time.Duration = time.Duration(DefaultTimeoutInt) * time.Second
)

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

// Initialize will format a Configuration that has been unmarshalled form json.
// It will set default values and validate settings.
func (conf Configuration) Initialize() (Configuration, error) {
	if conf.BucketSize <= 0 {
		conf.BucketSize = DefaultBucketSize
	}
	if conf.RefillRate <= 0 {
		conf.RefillRate = DefaultRefillRate
	}
	if conf.Concurrency <= 0 {
		conf.Concurrency = DefaultConcurrency
	}
	if conf.Timeout <= 0 {
		conf.Timeout = DefaultTimeout
	}

	conf.URL = conf.AdminURL()

	if !(strings.ToLower(strings.TrimSpace(conf.ThemeID)) == "live") {
		// theme_id may be specified as 'live', indicating that the user
		// is opting into always syncing to the current, production theme
		if themeID, err := strconv.ParseInt(conf.ThemeID, 10, 64); err == nil {
			conf.URL = fmt.Sprintf("%s/themes/%d", conf.URL, themeID)
		} else {
			return conf, fmt.Errorf("missing theme_id.")
		}
	}

	if len(conf.Domain) == 0 {
		return conf, fmt.Errorf("missing domain")
	} else if !strings.HasSuffix(conf.Domain, "myshopify.com") && !strings.HasSuffix(conf.Domain, "myshopify.io") {
		return conf, fmt.Errorf("invalid domain, must end in '.myshopify.com'")
	}

	if len(conf.Password) == 0 {
		return conf, fmt.Errorf("missing password")
	}
	return conf, nil
}

// AdminURL will return the url to the shopify admin.
func (conf Configuration) AdminURL() string {
	return fmt.Sprintf("https://%s/admin", conf.Domain)
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
	return fmt.Sprintf("%s/assets.json", conf.URL)
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
	return fmt.Sprintf("<token:%s domain:%s bucket:%d refill:%d url:%s>", conf.Password, conf.Domain, conf.BucketSize, conf.RefillRate, conf.URL)
}
