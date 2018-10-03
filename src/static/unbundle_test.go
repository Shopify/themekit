package static

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/stretchr/testify/assert"
)

const testData = "PK\x03\x04\x14\x00\x08\x00\x08\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x15\x00\x00\x00assets/application.js\xd2\xd7W\x08(-Q\xa8\xcc/-RH,(\xc8\xc9LN\xcc,\xc9\xcfS\xc8J,K,N.\xca,(Q\xc8H-J\xe5\x02\x04\x00\x00\xff\xffPK\x07\x08\x11\xad\xdd#.\x00\x00\x00(\x00\x00\x00PK\x01\x02\x14\x00\x14\x00\x08\x00\x08\x00\x00\x00\x00\x00\x11\xad\xdd#.\x00\x00\x00(\x00\x00\x00\x15\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00assets/application.jsPK\x05\x06\x00\x00\x00\x00\x01\x00\x01\x00C\x00\x00\x00q\x00\x00\x00\x00\x00"

func TestGetZipContents(t *testing.T) {
	files, err := getZipContents(testData)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(files))
	assert.NotNil(t, files["assets"])
	assert.Equal(t, 1, len(files["assets"]))
	assert.NotNil(t, files["assets"]["assets/application.js"])

	_, err = getZipContents("not zip")
	assert.EqualError(t, err, "zip: not a valid zip file")
}

func TestUnbundle(t *testing.T) {
	testdirpath := filepath.Join("_testdata", "out")
	os.Mkdir(filepath.Dir(testdirpath), 0755)
	defer os.RemoveAll(testdirpath)

	stdOut := bytes.NewBufferString("")
	ctx := &cmdutil.Ctx{
		Flags: cmdutil.Flags{Directory: testdirpath},
		Log:   log.New(stdOut, "", 0),
	}

	Register(testData)
	assert.Nil(t, Unbundle(ctx))

	files := map[string][]string{}
	filepath.Walk(testdirpath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		dir := filepath.Dir(path)
		files[dir] = append(files[dir], path)
		return nil
	})

	fmt.Println(files)
	assert.Equal(t, 1, len(files))
	assert.Equal(t, 1, len(files[filepath.Join(testdirpath, "assets")]))
	assert.Equal(t,
		[]string{filepath.Join(testdirpath, "assets", "application.js")},
		files[filepath.Join(testdirpath, "assets")],
	)
}
