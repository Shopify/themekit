package static

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressData(t *testing.T) {
	buffer, err := compressData("_testdata/theme-template")
	assert.Nil(t, err)
	expected := "PK\\x03\\x04\\x14\\x00\\x08\\x00\\x08\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x15\\x00\\x00\\x00assets/application.js\\xd2\\xd7W\\x08(-Q\\xa8\\xcc/-RH,(\\xc8\\xc9LN\\xcc,\\xc9\\xcfS\\xc8J,K,N.\\xca,(Q\\xc8H-J\\xe5\\x02\\x04\\x00\\x00\\xff\\xffPK\\x07\\x08\\x11\\xad\\xdd#.\\x00\\x00\\x00(\\x00\\x00\\x00PK\\x01\\x02\\x14\\x00\\x14\\x00\\x08\\x00\\x08\\x00\\x00\\x00\\x00\\x00\\x11\\xad\\xdd#.\\x00\\x00\\x00(\\x00\\x00\\x00\\x15\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00assets/application.jsPK\\x05\\x06\\x00\\x00\\x00\\x00\\x01\\x00\\x01\\x00C\\x00\\x00\\x00q\\x00\\x00\\x00\\x00\\x00"
	assert.Equal(t, expected, buffer.String())

	_, err = compressData("_testdata/nope")
	assert.Error(t, err)
}

func TestWriteOutTemplate(t *testing.T) {
	path := "_testdata/out/foo.go"
	os.Mkdir(filepath.Dir(path), 0755)
	defer os.RemoveAll(filepath.Dir(path))
	err := writeOutTemplate(path, bytes.NewBuffer([]byte("a test file")))
	assert.Nil(t, err)
	f, _ := os.Open(path)
	defer f.Close()
	data, _ := ioutil.ReadAll(f)
	expected := `package out

import "github.com/Shopify/themekit/src/static"

func init() {
	static.Register("a test file")
}`
	assert.Equal(t, expected, string(data[:]))

	err = writeOutTemplate("_testdata/nope/no.go", bytes.NewBuffer([]byte("a test file")))
	assert.NotNil(t, err)
}
