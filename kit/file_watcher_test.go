package kit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/fsnotify.v1"
)

const (
	textFixturePath  = "../fixtures/file_watcher/whatever.txt"
	watchFixturePath = "../fixtures/file_watcher"
)

type FileWatcherTestSuite struct {
	suite.Suite
}

func (suite *FileWatcherTestSuite) TestNewFileReader() {
	client := ThemeClient{}
	newFileWatcher(client, watchFixturePath, true, eventFilter{}, func(ThemeClient, AssetEvent, error) {})
}

func (suite *FileWatcherTestSuite) TestConvertFsEvents() {
}

func (suite *FileWatcherTestSuite) TestStopWatching() {
}

func (suite *FileWatcherTestSuite) TestHandleEvent() {
	_, err := handleEvent(fsnotify.Event{Name: "gone/oath", Op: fsnotify.Remove})
	assert.NotNil(suite.T(), err)

	writes := map[fsnotify.Op]EventType{
		fsnotify.Chmod:  Update,
		fsnotify.Create: Update,
		fsnotify.Write:  Update,
		fsnotify.Remove: Remove,
	}
	for fsEvent, themekitEvent := range writes {
		event := fsnotify.Event{Name: textFixturePath, Op: fsEvent}
		assetEvent, err := handleEvent(event)
		assert.Equal(suite.T(), themekitEvent, assetEvent.Type)
		assert.Equal(suite.T(), "File not in project workspace.", err.Error())
	}
}

func (suite *FileWatcherTestSuite) TestExtractAssetKey() {
	tests := map[string]string{
		textFixturePath:                                 "",
		"/long/path/to/config.yml":                      "",
		"/long/path/to/assets/logo.png":                 "assets/logo.png",
		"/long/path/to/templates/customers/test.liquid": "templates/customers/test.liquid",
		"/long/path/to/config/test.liquid":              "config/test.liquid",
		"/long/path/to/layout/test.liquid":              "layout/test.liquid",
		"/long/path/to/snippets/test.liquid":            "snippets/test.liquid",
		"/long/path/to/templates/test.liquid":           "templates/test.liquid",
		"/long/path/to/locales/test.liquid":             "locales/test.liquid",
		"/long/path/to/sections/test.liquid":            "sections/test.liquid",
	}
	for input, expected := range tests {
		assert.Equal(suite.T(), expected, extractAssetKey(input))
	}
}

func (suite *FileWatcherTestSuite) TestfindDirectoriesToWatch() {
	expected := []string{
		watchFixturePath,
		watchFixturePath + "/assets",
		watchFixturePath + "/config",
		watchFixturePath + "/layout",
		watchFixturePath + "/locales",
		watchFixturePath + "/snippets",
		watchFixturePath + "/templates",
		watchFixturePath + "/templates/customers",
	}

	files, err := findDirectoriesToWatch(watchFixturePath, true, func(string) bool { return false })
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), expected, files)

	files, err = findDirectoriesToWatch(watchFixturePath, false, func(string) bool { return false })
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []string{watchFixturePath}, files)
}

func TestFileWatcherTestSuite(t *testing.T) {
	suite.Run(t, new(FileWatcherTestSuite))
}
