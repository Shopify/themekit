package phoenix

import (
	"fmt"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"net/http"
)

type Configuration struct {
	AccessToken  string `yaml:"access_token"`
	Domain       string `yaml:"store"`
	Url          string
	IgnoredFiles []string `yaml:"ignore_files,omitempty"`
	BucketSize   int      `yaml:"bucket_size"`
	RefillRate   int      `yaml:"refill_rate"`
}

const (
	DefaultBucketSize int = 40
	DefaultRefillRate int = 2
)

func LoadConfiguration(contents []byte) (conf Configuration, err error) {
	err = yaml.Unmarshal(contents, &conf)
	if err == nil {
		if conf.BucketSize <= 0 {
			conf.BucketSize = DefaultBucketSize
		}
		if conf.RefillRate <= 0 {
			conf.RefillRate = DefaultRefillRate
		}
		conf.Url = fmt.Sprintf("https://%s/admin", conf.Domain)
	}
	return
}

func LoadConfigurationFromFile(location string) (conf Configuration, err error) {
	contents, err := ioutil.ReadFile(location)
	if err == nil {
		conf, err = LoadConfiguration(contents)
	}

	return
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
	return fmt.Sprintf("<token:%s domain:%s bucket:%d refill:%d>", c.AccessToken, c.Domain, c.BucketSize, c.RefillRate)
}
