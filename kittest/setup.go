package kittest

import (
	"os"
	"path/filepath"
)

// ProjectFiles are the files that are generated for project fixtures
var ProjectFiles = map[string]string{
	filepath.Join("assets", "application.js"):              "// this is js",
	filepath.Join("assets", "pixel.png"):                   "",
	filepath.Join("config", "settings_data.json"):          "",
	filepath.Join("layout", ".gitkeep"):                    "",
	filepath.Join("locales", "en.json"):                    "",
	filepath.Join("snippets", "snippet.js"):                "",
	filepath.Join("templates", "template.liquid"):          "",
	filepath.Join("templates", "customers", "test.liquid"): "",
	filepath.Join("templates", "customers", "test.liquid"): "",
}

// Setup will generate all the project files and directories needed for testing kit and cmd
func Setup() {
	os.MkdirAll(FixtureProjectPath, 0777)
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
	for filename, content := range ProjectFiles {
		if err := TouchFixtureFile(filename, content); err != nil {
			return err
		}
	}
	return nil
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
