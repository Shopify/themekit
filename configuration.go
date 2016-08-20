package themekit

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v1"
)

// Configuration ... TODO
type Configuration struct {
	AccessToken  string        `yaml:"access_token,omitempty"`
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
	// DefaultBucketSize ... TODO
	DefaultBucketSize int = 40
	// DefaultRefillRate ... TODO
	DefaultRefillRate int = 2
	// DefaultConcurrency ... TODO
	DefaultConcurrency int = 2
	// DefaultTimeout ... TODO
	DefaultTimeout time.Duration = 30 * time.Second
)

// LoadConfiguration ... TODO
func LoadConfiguration(contents []byte) (Configuration, error) {
	var conf Configuration
	if err := yaml.Unmarshal(contents, &conf); err != nil {
		return conf, err
	}
	return conf.Initialize()
}

// Initialize ... TODO
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
			return conf, fmt.Errorf("missing theme_id. Error: \"%s\"", err)
		}
	}

	if len(conf.Domain) == 0 {
		return conf, fmt.Errorf("missing domain")
	} else if !strings.HasSuffix(conf.Domain, "myshopify.com") && !strings.HasSuffix(conf.Domain, "myshopify.io") {
		return conf, fmt.Errorf("invalid domain, must end in '.myshopify.com'")
	}

	if len(conf.AccessToken) == 0 && len(conf.Password) == 0 {
		return conf, fmt.Errorf("missing password or access_token (using 'password' is encouraged. 'access_token', which does the same thing will be deprecated soon)")
	}
	return conf, nil
}

// AdminURL ... TODO
func (conf Configuration) AdminURL() string {
	return fmt.Sprintf("https://%s/admin", conf.Domain)
}

func (conf Configuration) Write(w io.Writer) error {
	bytes, err := yaml.Marshal(conf)
	if err == nil {
		_, err = w.Write(bytes)
	}
	return err
}

// Save ... TODO
func (conf Configuration) Save(location string) error {
	file, err := os.OpenFile(location, os.O_WRONLY|os.O_CREATE, 0644)
	defer file.Close()
	if err == nil {
		err = conf.Write(file)
	}
	return err
}

// AssetPath ... TODO
func (conf Configuration) AssetPath() string {
	return fmt.Sprintf("%s/assets.json", conf.URL)
}

// AddHeaders ... TODO
func (conf Configuration) AddHeaders(req *http.Request) {
	var accessToken string
	if len(conf.Password) > 0 {
		accessToken = conf.Password
	} else {
		accessToken = conf.AccessToken
	}

	req.Header.Add("X-Shopify-Access-Token", accessToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("go/themekit (%s; %s)", runtime.GOOS, runtime.GOARCH))
}

func (conf Configuration) String() string {
	return fmt.Sprintf("<token:%s domain:%s bucket:%d refill:%d url:%s>", conf.AccessToken, conf.Domain, conf.BucketSize, conf.RefillRate, conf.URL)
}
