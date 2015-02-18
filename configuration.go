package phoenix

import (
	"fmt"
	"gopkg.in/yaml.v1"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

func LoadConfigurationFromCurrentDirectory() (conf Configuration, err error) {
	fmt.Println(RedText("[Deprecated] LoadConfigurationFromCurrentDirectory will be removed in the next release. Use Environments instead."))
	dir, err := os.Getwd()
	if err != nil {
		return Configuration{}, err
	}

	config, err := LoadConfigurationFromFile(fmt.Sprintf("%s/config.yml", dir))
	return config, err
}

func LoadConfiguration(contents []byte) (conf Configuration, err error) {
	err = yaml.Unmarshal(contents, &conf)
	if err == nil {
		return conf.Initialize(), err
	}
	return
}

func (conf Configuration) Initialize() Configuration {
	if conf.BucketSize <= 0 {
		conf.BucketSize = DefaultBucketSize
	}
	if conf.RefillRate <= 0 {
		conf.RefillRate = DefaultRefillRate
	}
	if conf.Concurrency <= 0 {
		conf.Concurrency = DefaultConcurrency
	}

	conf.Url = fmt.Sprintf("https://%s/admin", conf.Domain)
	if conf.ThemeId != 0 {
		conf.Url = fmt.Sprintf("%s/themes/%d", conf.Url, conf.ThemeId)
	}
	return conf
}

func LoadConfigurationFromFile(location string) (conf Configuration, err error) {
	fmt.Println(RedText("[Deprecated] LoadConfigurationFromFile will be removed in the next release. Use Environments instead."))
	contents, err := ioutil.ReadFile(location)
	if err == nil {
		conf, err = LoadConfiguration(contents)
	}

	return
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
	req.Header.Add("User-Agent", "go/phoenix")
}

func (c Configuration) String() string {
	return fmt.Sprintf("<token:%s domain:%s bucket:%d refill:%d url:%s>", c.AccessToken, c.Domain, c.BucketSize, c.RefillRate, c.Url)
}
