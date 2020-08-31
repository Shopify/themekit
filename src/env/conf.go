package env

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"encoding/json"
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"github.com/shibukawa/configdir"
	"gopkg.in/yaml.v1"
)

const (
	vendor   = "Shopify"
	appname  = "Themekit"
	filename = "variables"
)

var (
	cdir          = configdir.New(vendor, appname)
	supportedExts = []string{"yml", "yaml", "json"}
	// ErrEnvDoesNotExist is returned when an environment that does not exist in the config is requested
	ErrEnvDoesNotExist = errors.New("environment does not exist in this environments list")
	// ErrEnvNotDefined is returned if the environment was found but had no config, this usually means poor conf
	ErrEnvNotDefined = errors.New("environment was found but not defined")
	// ErrNoEnvironmentsDefined is returned when trying to save an empty config
	ErrNoEnvironmentsDefined = errors.New("no environments defined, nothing to write")
	// ErrInvalidEnvironmentName is returned if an environment is trying to be set with a blank name
	ErrInvalidEnvironmentName = errors.New("environment name cannot be blank")
)

// Conf is a map of configurations to their environment name.
type Conf struct {
	Envs  map[string]*Env
	osEnv Env
	path  string
}

func init() {
	cdir.LocalPath, _ = os.Getwd()
}

// SourceVariables will find the environment files for configuration and import their
// environment variables into the current running process
func SourceVariables(flagpath string) error {
	if flagpath != "" {
		return godotenv.Load(flagpath)
	}
	if fldr := cdir.QueryFolderContainsFile(filename); fldr != nil {
		return godotenv.Load(filepath.Join(fldr.Path, filename))
	}
	return nil
}

// New will build a new blank config
func New(configPath string) Conf {
	conf := Conf{
		Envs:  make(map[string]*Env),
		osEnv: Env{},
		path:  configPath,
	}
	env.Parse(&conf.osEnv)
	return conf
}

// Load will read in the file from the configPath provided and
// then unmarshal the data into conf.
func Load(configPath string) (Conf, error) {
	conf := New(configPath)
	path, ext, err := searchConfigPath(configPath)
	if err != nil {
		return conf, err
	}

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return conf, err
	}

	contents = []byte(os.ExpandEnv(string(contents)))

	switch ext {
	case "yml", "yaml":
		if err = yaml.Unmarshal(contents, &conf.Envs); err != nil {
			return conf, fmt.Errorf("Invalid yaml found while loading the config file: %v", err)
		}
	case "json":
		if err = json.Unmarshal(contents, &conf.Envs); err != nil {
			return conf, fmt.Errorf("Invalid json found while loading the config file: %v", err)
		}
	}

	return conf, nil
}

// Set will set the environment value and then mixin any overrides passed in. The os
// overrides and defaults will also be mixed into the new environment
func (c *Conf) Set(name string, initial Env, overrides ...Env) (*Env, error) {
	if name == "" {
		return nil, ErrInvalidEnvironmentName
	}
	var err error
	c.Envs[name], err = newEnv(name, initial, append([]Env{c.osEnv}, overrides...)...)
	return c.Envs[name], err
}

// Get will check if an environment exists and then return it. If the environment
// does not exists it will return an error
func (c *Conf) Get(name string, overrides ...Env) (*Env, error) {
	env, exists := c.Envs[name]
	if !exists {
		return env, ErrEnvDoesNotExist
	} else if env == nil {
		return env, ErrEnvNotDefined
	}
	return newEnv(name, *env, append([]Env{c.osEnv}, overrides...)...)
}

// Save will write out the config to a file.
func (c Conf) Save() error {
	f, err := c.file()
	if err != nil {
		return err
	}
	defer f.Close()
	return c.save(f)
}

func (c Conf) save(w io.Writer) error {
	// clear defaults before writing, we don't need to save defaults
	for name, env := range c.Envs {
		if env == nil {
			delete(c.Envs, name)
			continue
		}
		if env.Directory == Default.Directory {
			env.Directory = ""
		} else if env.Directory != "" {
			configDir := filepath.Dir(c.path)
			if rel, err := filepath.Rel(configDir, env.Directory); err == nil {
				env.Directory = rel
			}
		}
		if env.Timeout == Default.Timeout {
			env.Timeout = 0
		}
		c.Envs[name] = env
	}

	if len(c.Envs) == 0 {
		return ErrNoEnvironmentsDefined
	}

	bytes, err := yaml.Marshal(c.Envs)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}

func (c Conf) file() (io.WriteCloser, error) {
	return os.OpenFile(c.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

func searchConfigPath(configPath string) (string, string, error) {
	dir := filepath.Dir(configPath)
	filename := filepath.Base(configPath)
	name := filename[0 : len(filename)-len(filepath.Ext(filename))]
	for _, ext := range supportedExts {
		foundPath := filepath.Join(dir, name+"."+ext)
		if _, err := os.Stat(foundPath); err == nil && "."+ext == filepath.Ext(filename) {
			return foundPath, ext, nil
		}
	}
	return "", "", os.ErrNotExist
}
