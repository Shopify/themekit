package kittest

import (
	"os"
)

// Setup will generate all the project files and directories needed for testing kit and cmd
func Setup() {
	os.MkdirAll(FixtureProjectPath, 0777)
}

// Cleanup should be called after any test that touches the fs
func Cleanup() {
	os.Remove("config.yml")
	os.Remove("config.json")
	os.RemoveAll(FixturesPath)
}
