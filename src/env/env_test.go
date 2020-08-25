package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnvNew(t *testing.T) {
	env, _ := newEnv("", Env{})
	assert.Equal(t, Default, *env)
	pwd, _ := os.Getwd()

	osEnv := Env{
		Name:         "Foobar",
		Password:     "password",
		ThemeID:      "themeid",
		Domain:       "nope.myshopify.com",
		Directory:    filepath.Join(pwd, "env"),
		IgnoredFiles: []string{"one", "two", "three"},
		Proxy:        ":3000",
		Ignores:      []string{"four", "five", "six"},
		Timeout:      40 * time.Second,
	}

	env, _ = newEnv("", Env{}, osEnv)
	assert.Equal(t, osEnv, *env)

	env, _ = newEnv("", Env{Password: "file"}, Env{Password: "environment"})
	assert.Equal(t, "environment", env.Password)

	env, _ = newEnv("", Env{Password: "file"}, Env{Password: "flag"}, Env{Password: "environment"})
	assert.Equal(t, "flag", env.Password)
}

func TestEnv_Validate(t *testing.T) {
	testCases := []struct {
		env        Env
		err        string
		notwindows bool
	}{
		{env: Env{Password: "file", ThemeID: "123", Domain: "test.myshopify.com"}},
		{env: Env{Password: "file", ThemeID: "live", Domain: "test.myshopify.com"}, err: "invalid environment []: ('live' is no longer supported for theme_id. Please use an ID instead)"},
		{env: Env{ThemeID: "123", Domain: "test.myshopify.com"}, err: "missing password"},
		{env: Env{Password: "test", ThemeID: "123", Domain: "test.nope.com"}, err: "invalid store domain"},
		{env: Env{Password: "test", ThemeID: "123"}, err: "missing store domain"},
		{env: Env{Password: "test", Domain: "test.myshopify.com"}, err: "missing theme_id"},
		{env: Env{Password: "file", ThemeID: "abc", Domain: "test.myshopify.com"}, err: "invalid theme_id"},
		{notwindows: true, env: Env{Password: "abc123", Domain: "test.myshopify.com", ThemeID: "123", Directory: filepath.Join("_testdata", "symlink_projectdir")}},
		{notwindows: true, env: Env{Password: "abc123", Domain: "test.myshopify.com", Directory: filepath.Join("_testdata", "bad_symlink")}, err: "invalid project symlink"},
		{notwindows: true, env: Env{Password: "abc123", Domain: "test.myshopify.com", Directory: filepath.Join("_testdata", "symlink_file")}, err: "is not a directory"},
		{env: Env{Password: "abc123", Domain: "test.myshopify.com", Directory: "not_a_dir"}, err: "invalid project directory"},
		{env: Env{Password: "abc123", Domain: "test.myshopify.com", Directory: filepath.Join("_testdata", "projectdir", "bad_format.yml")}, err: "is not a directory"},
	}

	for i, testcase := range testCases {
		if testcase.notwindows && runtime.GOOS == "windows" {
			continue
		}
		err := testcase.env.validate()
		if testcase.err == "" {
			assert.Nil(t, err, fmt.Sprintf("Testcase: %v", i))
		} else if assert.Error(t, err, fmt.Sprintf("Testcase: %v %v", i, testcase.err)) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}
