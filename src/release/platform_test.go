package release

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Shopify/themekit/src/release/_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuildPlatform(t *testing.T) {
	testcases := []struct {
		bin, err, uerr string
		uploads        bool
	}{
		{bin: "theme", uploads: true},
		{bin: "other", uploads: false, err: "no such file or directory"},
	}

	for _, testcase := range testcases {
		u := new(mocks.Uploader)
		if testcase.uploads {
			expectation := u.On("File", "v0.1.1/platform/theme", mock.MatchedBy(func(r *os.File) bool { return true }))
			if testcase.uerr != "" {
				expectation.Return("", errors.New(testcase.uerr))
			} else {
				expectation.Return("http://amazon.com/v0.1.1/platform/theme", nil)
			}
		}

		plat, err := buildPlatform("v0.1.1", "platform", filepath.Join("_testdata", "dist"), testcase.bin, u)
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, platform{
				Name:   "platform",
				URL:    "http://amazon.com/v0.1.1/platform/theme",
				Digest: "641b84fb6d971219b19aaea227a77235",
			}, plat)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}

		if testcase.uploads {
			u.AssertExpectations(t)
		}
	}
}
