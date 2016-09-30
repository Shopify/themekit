package kit

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/fsnotify.v1"

	"github.com/Shopify/themekit/theme"
)

type FileWatcherSuite struct {
	suite.Suite
}

func (s *FileWatcherSuite) TearDownTest() {
	RestoreReader()
}

func (s *FileWatcherSuite) TestThatLoadAssetProperlyExtractsTheAssetKey() {
	var tests = []struct {
		input    fsnotify.Event
		expected theme.Asset
	}{
		{fsnotify.Event{Name: "../fixtures/layout/theme.liquid"}, theme.Asset{Key: "layout/theme.liquid", Value: "Liquid Theme\n"}},
		{fsnotify.Event{Name: "../fixtures/templates/customers/account.liquid"}, theme.Asset{Key: "templates/customers/account.liquid", Value: "Account Page\n"}},
		{fsnotify.Event{Name: "../fixtures/snippets/layout-something.liquid"}, theme.Asset{Key: "snippets/layout-something.liquid", Value: "Something Liquid\n"}},
	}
	for _, test := range tests {
		actual := fwLoadAsset(test.input)
		assert.Equal(s.T(), test.expected.Key, actual.Key)
		assert.Equal(s.T(), test.expected.Value, actual.Value)
	}
}

func (s *FileWatcherSuite) TestDeterminingContentTypesOfFiles() {
	image, _ := ioutil.ReadFile("../fixtures/image.png")
	assert.Equal(s.T(), "binary", contentTypeFor(image))
	assert.Equal(s.T(), "text", contentTypeFor([]byte("Hello World")))
	assert.Equal(s.T(), "text", contentTypeFor([]byte("<!DOCTYPE html><html><head></head><body></body></html>")))
}

func (s *FileWatcherSuite) TestThatLoadAssetProperlyExtractsAttachmentDataForBinaryFiles() {
	imageData, _ := ioutil.ReadFile("../fixtures/image.png")
	encodedImageData := encode64(imageData)
	WatcherFileReader = func(filename string) ([]byte, error) {
		return imageData, nil
	}
	event := fsnotify.Event{Name: "../fixtures/image.png", Op: fsnotify.Chmod}
	assetEvent := HandleEvent(event)
	assert.Equal(s.T(), "", assetEvent.Asset().Value)
	assert.Equal(s.T(), encodedImageData, assetEvent.Asset().Attachment)
}

func (s *FileWatcherSuite) TestHandleEventConvertsFSNotifyEventsIntoAssetEvents() {
	WatcherFileReader = func(filename string) ([]byte, error) {
		return []byte("hello"), nil
	}
	writes := map[fsnotify.Op]EventType{
		fsnotify.Chmod:  Update,
		fsnotify.Create: Update,
		fsnotify.Remove: Remove,
	}
	for fsEvent, themekitEvent := range writes {
		event := fsnotify.Event{Name: "../fixtures/whatever.txt", Op: fsEvent}
		assetEvent := HandleEvent(event)
		assert.Equal(s.T(), themekitEvent, assetEvent.Type())
	}
}

func TestFileWatcherSuite(t *testing.T) {
	suite.Run(t, new(FileWatcherSuite))
}

func contentTypeFor(data []byte) string {
	contentType := http.DetectContentType(data)
	if strings.Contains(contentType, "text") {
		return "text"
	}
	return "binary"
}

func encode64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
