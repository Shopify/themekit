package main

import (
	"bytes"
	"fmt"
	"github.com/csaunders/phoenix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/fsnotify.v1"
	"image"
	"image/png"
	"testing"
)

type FileWatcherSuite struct {
	suite.Suite
}

func (s *FileWatcherSuite) TearDownTest() {
	RestoreReader()
}

func (s *FileWatcherSuite) TestThatLoadAssetProperlyExtractsTheAssetKey() {
	var event fsnotify.Event
	WatcherFileReader = func(filename string) ([]byte, error) {
		assert.Equal(s.T(), event.Name, filename)
		return []byte("hello"), nil
	}

	for _, key := range assetLocations {
		event = fsnotify.Event{Name: fmt.Sprintf("/home/gopher/themes/simple/%s/thing.css", key)}
		asset := LoadAsset(event)
		assert.Equal(s.T(), asset.Key, fmt.Sprintf("%s/thing.css", key))
		assert.Equal(s.T(), "hello", asset.Value)
	}
}

func (s *FileWatcherSuite) TestDeterminingContentTypesOfFiles() {
	assert.Equal(s.T(), "binary", ContentTypeFor(makeRgbImage(8, 8)))
	assert.Equal(s.T(), "text", ContentTypeFor([]byte("Hello World")))
	assert.Equal(s.T(), "text", ContentTypeFor([]byte("<!DOCTYPE html><html><head></head><body></body></html>")))
}

func (s *FileWatcherSuite) TestThatLoadAssetProperlyExtractsAttachmentDataForBinaryFiles() {
	imageData := makeRgbImage(8, 8)
	encodedImageData := Encode64(imageData)
	WatcherFileReader = func(filename string) ([]byte, error) {
		return imageData, nil
	}
	event := fsnotify.Event{Name: "assets/image.png", Op: fsnotify.Chmod}
	assetEvent := HandleEvent(event)
	assert.Equal(s.T(), "", assetEvent.Asset().Value)
	assert.Equal(s.T(), encodedImageData, assetEvent.Asset().Attachment)
}

func (s *FileWatcherSuite) TestHandleEventConvertsFSNotifyEventsIntoAssetEvents() {
	WatcherFileReader = func(filename string) ([]byte, error) {
		return []byte("hello"), nil
	}
	writes := map[fsnotify.Op]phoenix.EventType{
		fsnotify.Chmod:  phoenix.Update,
		fsnotify.Create: phoenix.Update,
		fsnotify.Remove: phoenix.Remove,
	}
	for fsEvent, phoenixEvent := range writes {
		event := fsnotify.Event{Name: "assets/whatever.txt", Op: fsEvent}
		assetEvent := HandleEvent(event)
		assert.Equal(s.T(), phoenixEvent, assetEvent.Type())
	}
}

func makeRgbImage(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	buff := bytes.NewBuffer([]byte{})
	png.Encode(buff, img)
	return buff.Bytes()
}

func TestFileWatcherSuite(t *testing.T) {
	suite.Run(t, new(FileWatcherSuite))
}
