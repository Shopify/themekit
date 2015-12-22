package themekit

import (
	"fmt"
	"gopkg.in/yaml.v1"
	"io"
	"net/http"
	"os"
	"runtime"
)

type Configuration struct {
	ThemeId      int64    `yaml:"theme_id,omitempty"`
	AccessToken  string   `yaml:"access_token"`
	Domain       string   `yaml:"store"`
	Url          string   `yaml:"-"`
	IgnoredFiles []string `yaml:"ignore_files,omitempty"`
	BucketSize   int      `yaml:"bucket_size"`
	RefillRate   int      `yaml:"refill_rate"`
	Concurrency  int      `yaml:"concurrency,omitempty"`
	Proxy        string   `yaml:"proxy,omitempty"`
	Ignores      []string `yaml:"ignores,omitempty"`
}

const (
	DefaultBucketSize  int = 40
	DefaultRefillRate  int = 2
	DefaultConcurrency int = 2
)

func LoadConfiguration(contents []byte) (Configuration, error) {
	var conf Configuration
	if err := yaml.Unmarshal(contents, &conf); err != nil {
		return conf, err
	}
	return conf.Initialize()
}

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

	conf.Url = conf.AdminUrl()
	if conf.ThemeId != 0 {
		conf.Url = fmt.Sprintf("%s/themes/%d", conf.Url, conf.ThemeId)
	}

	if len(conf.Domain) == 0 {
		return conf, fmt.Errorf("missing domain")
	}
	if len(conf.AccessToken) == 0 {
		return conf, fmt.Errorf("missing access_token")
	}
	return conf, nil
}

func (conf Configuration) AdminUrl() string {
	return fmt.Sprintf("https://%s/admin", conf.Domain)
}

func (conf Configuration) Write(w io.Writer) error {
	bytes, err := yaml.Marshal(conf)
	if err == nil {
		_, err = w.Write(bytes)
	}
	return err
}

func (conf Configuration) Save(location string) error {
	file, err := os.OpenFile(location, os.O_WRONLY|os.O_CREATE, 0644)
	defer file.Close()
	if err == nil {
		err = conf.Write(file)
	}
	return err
}

func (conf Configuration) AssetPath() string {
	return fmt.Sprintf("%s/assets.json", conf.Url)
}

func (conf Configuration) AddHeaders(req *http.Request) {
	req.Header.Add("X-Shopify-Access-Token", conf.AccessToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("go/themekit (%s; %s)", runtime.GOOS, runtime.GOARCH))
}

func (c Configuration) String() string {
	return fmt.Sprintf("<token:%s domain:%s bucket:%d refill:%d url:%s>", c.AccessToken, c.Domain, c.BucketSize, c.RefillRate, c.Url)
}
