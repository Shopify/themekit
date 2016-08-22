package commands

import (
	"testing"

	"github.com/Shopify/themekit"
	"github.com/stretchr/testify/assert"
)

func FakeOsGetwd() (string, error) {
	return "../fixtures/local_assets/templates/", nil
}

func TestUploadSingleAsset(t *testing.T) {
	args := DefaultArgs()
	args.WorkingDirGetter = FakeOsGetwd
	args.Filenames = []string{"404.liquid"}

	results := ReadAndPrepareFilesSync(args)
	result := results[0]

	assert.Equal(t, "404.liquid", result.Asset().Key)
	assert.Equal(t, "404!\n", result.Asset().Value)

	assert.Equal(t, themekit.Update, result.Type())
}
