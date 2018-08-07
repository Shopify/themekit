package env

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnvNew(t *testing.T) {
	env, _ := newEnv("", Env{})
	assert.Equal(t, Default, *env)

	osEnv := Env{
		Name:         "Foobar",
		Password:     "password",
		ThemeID:      "themeid",
		Domain:       "nope.myshopify.com",
		Directory:    "../env",
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
		env Env
		err string
	}{
		{env: Env{Password: "file", ThemeID: "123", Domain: "test.myshopify.com"}},
		{env: Env{Password: "file", ThemeID: "live", Domain: "test.myshopify.com"}},
		{env: Env{ThemeID: "123", Domain: "test.myshopify.com"}, err: "missing password"},
		{env: Env{Password: "test", ThemeID: "123", Domain: "test.nope.com"}, err: "invalid store domain"},
		{env: Env{Password: "test", ThemeID: "123"}, err: "missing store domain"},
		{env: Env{Password: "test", Domain: "test.myshopify.com"}},
		{env: Env{Password: "file", ThemeID: "abc", Domain: "test.myshopify.com"}, err: "invalid theme_id"},
		{env: Env{Password: "abc123", Domain: "test.myshopify.com", Directory: "_testdata/symlink_projectdir"}},
		{env: Env{Password: "abc123", Domain: "test.myshopify.com", Directory: "_testdata/bad_symlink"}, err: "invalid project directory"},
		{env: Env{Password: "abc123", Domain: "test.myshopify.com", Directory: "_testdata/file_symlink"}, err: "invalid project directory"},
		{env: Env{Password: "abc123", Domain: "test.myshopify.com", Directory: "not_a_dir"}, err: "invalid project directory"},
	}

	for _, testcase := range testCases {
		err := testcase.env.validate()
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.Error(t, err, testcase.err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}
