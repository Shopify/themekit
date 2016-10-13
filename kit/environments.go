package kit

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v1"
)

// DefaultEnvironment is the environment that will be loaded if no environment is specified.
const DefaultEnvironment string = "development"

// Environments is a map of configurations to their environment name.
type Environments map[string]Configuration

// LoadEnvironmentsFromFile will read in the file from the location provided and
// then unmarshal the data into environments.
func LoadEnvironments(location string) (env Environments, err error) {
	contents, err := ioutil.ReadFile(location)
	if err == nil {
		err = yaml.Unmarshal(contents, &env)
	}
	return
}

// SetConfiguration will update a config environment to the provided configuration.
func (e Environments) SetConfiguration(environmentName string, conf Configuration) {
	e[environmentName] = conf
}

// GetConfiguration will return the configuration for the environment. An error will
// be returned if the environment does not exist or the configuration is invalid.
func (e Environments) GetConfiguration(environmentName string) (conf Configuration, err error) {
	conf, exists := e[environmentName]
	if !exists {
		return conf, fmt.Errorf("%s does not exist in this environments list", environmentName)
	}

	validConfig, err := conf.Initialize()
	if err != nil {
		return conf, fmt.Errorf("could not load environment \"%s\": %s", environmentName, err)
	}

	return validConfig, nil
}

// Write will write out an environment to an io.Writer. It will return an error
// if there was an error marshalling the environment or writing.
func (e Environments) Write(w io.Writer) error {
	bytes, err := yaml.Marshal(e)
	if err == nil {
		_, err = w.Write(bytes)
	}
	return err
}

// String will return a formatted string of the environment.
func (e Environments) String() string {
	bytes, err := yaml.Marshal(e)
	if err != nil {
		return "environments: cannot serialize"
	}
	return string(bytes)
}

// Save will write out the environment to a file.
func (e Environments) Save(location string) error {
	file, err := os.OpenFile(location, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer file.Close()
	if err == nil {
		err = e.Write(file)
	}
	return err
}
