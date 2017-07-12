package kittest

import (
	"os"
	"path/filepath"
)

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
