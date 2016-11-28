package kit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"encoding/json"
	"gopkg.in/yaml.v1"
)

// DefaultEnvironment is the environment that will be loaded if no environment is specified.
const DefaultEnvironment string = "development"

var supportedExts = []string{"json", "yml", "yaml"}

// Environments is a map of configurations to their environment name.
type Environments map[string]*Configuration

// LoadEnvironments will read in the file from the location provided and
// then unmarshal the data into environments.
func LoadEnvironments(location string) (env Environments, err error) {
	env = map[string]*Configuration{}
	path, ext, err := searchConfigPath(location)
	if err != nil {
		return env, err
	}

	contents, err := ioutil.ReadFile(path)
	if err == nil {
		switch ext {
		case "yml", "yaml":
			err = yaml.Unmarshal(contents, &env)
		case "json":
			err = json.Unmarshal(contents, &env)
		}
	}
	return
}

func searchConfigPath(configPath string) (string, string, error) {
	dir := filepath.Dir(configPath)
	filename := filepath.Base(configPath)
	name := filename[0 : len(filename)-len(filepath.Ext(filename))]
	for _, ext := range supportedExts {
		foundPath := filepath.Join(dir, name+"."+ext)
		if _, err := os.Stat(foundPath); err == nil {
			return foundPath, ext, nil
		}
	}
	return "", "", os.ErrNotExist
}

// SetConfiguration will update a config environment to the provided configuration.
func (e Environments) SetConfiguration(environmentName string, conf *Configuration) {
	e[environmentName] = conf
}

// GetConfiguration will return the configuration for the environment. An error will
// be returned if the environment does not exist or the configuration is invalid.
func (e Environments) GetConfiguration(environmentName string) (*Configuration, error) {
	conf, exists := e[environmentName]
	if !exists {
		return conf, fmt.Errorf("%s does not exist in this environments list", environmentName)
	}
	conf.Environment = environmentName
	return conf.compile()
}

// Save will write out the environment to a file.
func (e Environments) Save(location string) error {
	file, err := os.OpenFile(location, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer file.Close()
	if err == nil {
		var bytes []byte
		bytes, err = yaml.Marshal(e.asYAML())
		if err == nil {
			_, err = file.Write(bytes)
		}
	}
	return err
}

func (e Environments) asYAML() Environments {
	out := map[string]*Configuration{}
	for name, config := range e {
		out[name] = config.asYAML()
	}
	return out
}
