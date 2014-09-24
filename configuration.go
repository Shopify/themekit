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
	AccessToken  string   `yaml:"access_token"`
	Domain       string   `yaml:"store"`
	Url          string   `yaml:"-"`
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
	return fmt.Sprintf("<token:%s domain:%s bucket:%d refill:%d>", c.AccessToken, c.Domain, c.BucketSize, c.RefillRate)
}
