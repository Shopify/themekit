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
		bin, uerr    string
		err, uploads bool
	}{
		{bin: "theme", uploads: true},
		{bin: "other", uploads: false, err: true},
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
		if !testcase.err {
			assert.Nil(t, err)
			assert.Equal(t, "platform", plat.Name)
			assert.Equal(t, "http://amazon.com/v0.1.1/platform/theme", plat.URL)
			assert.NotEqual(t, "", plat.Digest)
		} else {
			assert.NotNil(t, err)
		}

		if testcase.uploads {
			u.AssertExpectations(t)
		}
	}
}
