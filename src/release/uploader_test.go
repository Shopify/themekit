package release

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewS3Uploader(t *testing.T) {
	assert.NotNil(t, newS3Uploader("key", "secret"))
}

func TestUploader_File(t *testing.T) {
	uf := func(name string, body io.Reader) (string, error) {
		return "", errors.New("nope")
	}
	location, err := (&s3Uploader{upload: uf}).File("name", strings.NewReader("body"))
	assert.Error(t, err)
	assert.Equal(t, "", location)

	uf = func(name string, body io.Reader) (string, error) {
		return "12*&^%*#&^*)$&^@#$*)%^@#$(*&$#", nil
	}
	location, err = (&s3Uploader{upload: uf}).File("name", strings.NewReader("body"))
	assert.Error(t, err)
	assert.Equal(t, "", location)

	uf = func(name string, body io.Reader) (string, error) {
		return "http://valid-url.com?test=1", nil
	}
	location, err = (&s3Uploader{upload: uf}).File("name", strings.NewReader("body"))
	assert.Nil(t, err)
	assert.Equal(t, "http://valid-url.com?test=1", location)
}

func TestUploader_JSON(t *testing.T) {
	uf := func(name string, body io.Reader) (string, error) {
		return "http://valid-url.com?test=1", nil
	}
	err := (&s3Uploader{upload: uf}).JSON("name", nil)
	assert.Nil(t, err)

	err = (&s3Uploader{upload: uf}).JSON("name", func() {})
	assert.Error(t, err)
}
