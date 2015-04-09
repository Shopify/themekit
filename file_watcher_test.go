package themekit

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/fsnotify.v1"
	"io/ioutil"
	"testing"
)

type FileWatcherSuite struct {
	suite.Suite
}

func (s *FileWatcherSuite) TearDownTest() {
	RestoreReader()
}

func (s *FileWatcherSuite) TestThatLoadAssetProperlyExtractsTheAssetKey() {
	event := fsnotify.Event{Name: "fixtures/whatever.txt"}
	asset := fwLoadAsset(event)
	assert.Equal(s.T(), asset.Key, "fixtures/whatever.txt")
	assert.Equal(s.T(), "whatever\n", asset.Value)
}

func (s *FileWatcherSuite) TestDeterminingContentTypesOfFiles() {
	image, _ := ioutil.ReadFile("fixtures/image.png")
	assert.Equal(s.T(), "binary", ContentTypeFor(image))
	assert.Equal(s.T(), "text", ContentTypeFor([]byte("Hello World")))
	assert.Equal(s.T(), "text", ContentTypeFor([]byte("<!DOCTYPE html><html><head></head><body></body></html>")))
}

func (s *FileWatcherSuite) TestThatLoadAssetProperlyExtractsAttachmentDataForBinaryFiles() {
	imageData, _ := ioutil.ReadFile("fixtures/image.png")
	encodedImageData := Encode64(imageData)
	WatcherFileReader = func(filename string) ([]byte, error) {
		return imageData, nil
	}
	event := fsnotify.Event{Name: "fixtures/image.png", Op: fsnotify.Chmod}
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
		event := fsnotify.Event{Name: "fixtures/whatever.txt", Op: fsEvent}
		assetEvent := HandleEvent(event)
		assert.Equal(s.T(), themekitEvent, assetEvent.Type())
	}
}

func TestFileWatcherSuite(t *testing.T) {
	suite.Run(t, new(FileWatcherSuite))
}
