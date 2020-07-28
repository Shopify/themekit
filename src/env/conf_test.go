package env

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	conf := New("")
	assert.NotNil(t, conf.Envs)
	assert.NotNil(t, conf.osEnv)
}

func TestLoad(t *testing.T) {
	testcases := []struct {
		path string
		err  string
	}{
		{path: "_testdata/projectdir/valid_config.yml", err: ""},
		{path: "_testdata/projectdir/bad_format.yml", err: ""},
		{path: "_testdata/projectdir/invalid_yaml.yml", err: "Invalid yaml found while loading the config file"},
		{path: "_testdata/projectdir/config.json", err: ""},
		{path: "_testdata/projectdir/bad_config.json", err: "Invalid json found while loading the config file"},
		{path: "_testdata/projectdir/not_there.json", err: "file does not exist"},
	}

	for _, testcase := range testcases {
		conf, err := Load(testcase.path)
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.NotNil(t, conf)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}

	os.Setenv("STOREPASS", "abracadabra")
	os.Setenv("STOREDOMAIN", "magic.myshopify.com")
	defer os.Unsetenv("STOREPASS")
	defer os.Unsetenv("STOREDOMAIN")
	conf, err := Load("_testdata/projectdir/valid_config.yml")
	assert.Nil(t, err)
	assert.Equal(t, "abracadabra", conf.Envs["development"].Password)
	assert.Equal(t, "magic.myshopify.com", conf.Envs["development"].Domain)
}

func TestSearchConfigPath(t *testing.T) {
	testcases := []struct {
		path, ext string
		err       error
	}{
		{path: "_testdata/projectdir/valid_config.yml", err: nil, ext: "yml"},
		{path: "_testdata/projectdir/bad_format.yml", err: nil, ext: "yml"},
		{path: "_testdata/projectdir/config.json", err: nil, ext: "json"},
		{path: "_testdata/projectdir/not_there.json", err: os.ErrNotExist, ext: ""},
	}

	for _, testcase := range testcases {
		_, ext, err := searchConfigPath(testcase.path)
		assert.Equal(t, testcase.ext, ext)
		assert.Equal(t, testcase.err, err)
	}
}

func TestConf_Set(t *testing.T) {
	dir, _ := filepath.Abs(filepath.Join("..", "file"))

	testcases := []struct {
		name      string
		initial   Env
		expected  Env
		overrides []Env
		err       string
	}{
		{name: "", initial: Env{}, err: ErrInvalidEnvironmentName.Error()},
		{name: "development", initial: Env{}, err: "invalid environment"},
		{name: "development", initial: Env{ThemeID: "123", Domain: "yes.myshopify.com", Password: "abc123"}, expected: Env{ThemeID: "123", Name: "development", Domain: "yes.myshopify.com", Password: "abc123", Directory: Default.Directory, Timeout: Default.Timeout}},
		{name: "development", initial: Env{ThemeID: "123", Domain: "yes.myshopify.com", Password: "abc123", Directory: filepath.Join("..", "file")}, expected: Env{ThemeID: "123", Name: "development", Domain: "yes.myshopify.com", Password: "abc123", Directory: dir, Timeout: Default.Timeout}},
		{name: "development", initial: Env{Domain: "yes.myshopify.com", Password: "abc123"}, overrides: []Env{{ThemeID: "12345"}}, expected: Env{Name: "development", Domain: "yes.myshopify.com", Password: "abc123", ThemeID: "12345", Directory: Default.Directory, Timeout: Default.Timeout}},
	}

	for _, testcase := range testcases {
		conf := New("")
		env, err := conf.Set(testcase.name, testcase.initial, testcase.overrides...)
		if testcase.err == "" {
			assert.Nil(t, err)
			if assert.NotNil(t, env) {
				assert.Equal(t, testcase.expected, *env)
			}
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func TestConf_Get(t *testing.T) {
	os.Setenv("STOREPASS", "abracadabra")
	os.Setenv("STOREDOMAIN", "magic.myshopify.com")
	defer os.Unsetenv("STOREPASS")
	defer os.Unsetenv("STOREDOMAIN")

	testcases := []struct {
		path, toGet, themeid string
		err                  error
		overrides            []Env
	}{
		{path: "_testdata/projectdir/valid_config.yml", toGet: "development", err: nil, themeid: "123"},
		{path: "_testdata/projectdir/valid_config.yml", toGet: "development", err: nil, themeid: "12345", overrides: []Env{{ThemeID: "12345"}}},
		{path: "_testdata/projectdir/valid_config.yml", toGet: "nope", err: ErrEnvDoesNotExist},
		{path: "_testdata/projectdir/bad_format.yml", toGet: "other", err: ErrEnvNotDefined},
	}

	for _, testcase := range testcases {
		conf, _ := Load(testcase.path)
		env, err := conf.Get(testcase.toGet, testcase.overrides...)
		assert.Equal(t, testcase.err, err)
		if env != nil {
			assert.Equal(t, testcase.themeid, env.ThemeID)
		}
	}
}

func TestConf_Save(t *testing.T) {
	conf := New("")
	conf.Set("foobar", Env{
		Password: "password",
		Domain:   "nope.myshopify.com",
	})

	stringBuff := bytes.NewBufferString("")
	err := conf.save(stringBuff)
	assert.Nil(t, err)

	expected := `foobar:
  password: password
  store: nope.myshopify.com
`
	assert.Equal(t, expected, stringBuff.String())

	conf = New("")
	err = conf.save(stringBuff)
	assert.Equal(t, err, ErrNoEnvironmentsDefined)
}

func TestConf_SaveKeepsDirectoryRelativeIfItIsRelative(t *testing.T) {
	configDir, _ := os.Getwd()
	configPath := filepath.Join(configDir, "config.yml")
	conf := New(configPath)
	conf.Set("foobar", Env{
		Password:  "password",
		Domain:    "nope.myshopify.com",
		Directory: filepath.Join(configDir, "project"),
	})

	stringBuff := bytes.NewBufferString("")
	assert.Nil(t, conf.save(stringBuff))

	expected := `foobar:
  password: password
  store: nope.myshopify.com
  directory: project
`
	assert.Equal(t, expected, stringBuff.String())

	conf = New(configPath)
	conf.Set("foobar", Env{
		Password:  "password",
		Domain:    "nope.myshopify.com",
		Directory: filepath.Clean(filepath.Join(configDir, "..", "file")),
	})

	stringBuff = bytes.NewBufferString("")
	assert.Nil(t, conf.save(stringBuff))

	expected = `foobar:
  password: password
  store: nope.myshopify.com
  directory: ../file
`
	assert.Equal(t, expected, stringBuff.String())
}

func overWriteEnvVar(name, value string, fn func()) {
	originalValue := os.Getenv(name)
	os.Setenv(name, value)
	fn()
	os.Setenv(name, originalValue)
}
