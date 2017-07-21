package kittest

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// ProjectFiles are the files that are generated for project fixtures
var ProjectFiles = []string{
	filepath.Join("assets", "application.js"),
	filepath.Join("assets", "pixel.png"),
	filepath.Join("config", "settings_data.json"),
	filepath.Join("layout", ".gitkeep"),
	filepath.Join("locales", "en.json"),
	filepath.Join("snippets", "snippet.js"),
	filepath.Join("templates", "template.liquid"),
	filepath.Join("templates", "customers", "test.liquid"),
}

// SymlinkProjectPath is the path the the symlink to test project symlinks
var SymlinkProjectPath = filepath.Join(FixturesPath, "sym_project")

// Setup will generate all the project files and directories needed for testing kit and cmd
func Setup() {
	os.MkdirAll(FixtureProjectPath, 0777)
	os.Create(UpdateFilePath)
}

// Cleanup should be called after any test that touches the fs
func Cleanup() {
	os.Remove("config.yml")
	os.Remove("config.json")
	os.Remove("theme.lock")
	os.RemoveAll(FixturesPath)
}

// GenerateProject will generate a fake project of fixtures for testing
func GenerateProject() error {
	for _, filename := range ProjectFiles {
		content := ""
		switch filepath.Ext(filename) {
		case ".png":
			buf := new(bytes.Buffer)
			img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
			png.Encode(buf, img)
			content = buf.String()
		case ".liquid":
			content = "this is liquid content"
		case ".js":
			content = "this is js content"
		default:
			content = "this is content"
		}
		if err := TouchFixtureFile(filename, content); err != nil {
			return err
		}
	}
	return os.Symlink("project", SymlinkProjectPath)
}

// TouchFixtureFile will create a fixture in the fixture project path.
func TouchFixtureFile(path, content string) error {
	path = filepath.Join(FixtureProjectPath, path)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return file.Sync()
}
