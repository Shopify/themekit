package kit

import (
	"path/filepath"
	"strings"
)

var (
	assetLocations = []string{
		"templates/customers",
		"assets",
		"config",
		"layout",
		"snippets",
		"templates",
		"locales",
		"sections",
	}
)

func pathInProject(filename string) bool {
	return pathToProject(filename) != "" || isProjectDirectory(filename)
}

func isProjectDirectory(filename string) bool {
	for _, dir := range assetLocations {
		if directoriesEqual(filename, dir) {
			return true
		}
	}
	return false
}

func directoriesEqual(dir, other string) bool {
	return strings.HasSuffix(
		filepath.Clean(filepath.ToSlash(dir)+"/"),
		filepath.Clean(filepath.ToSlash(other)+"/"),
	)
}

func pathToProject(filename string) string {
	filename = filepath.ToSlash(filepath.Clean(filename))
	for _, dir := range assetLocations {
		split := strings.SplitAfterN(filename, dir+"/", 2)
		if len(split) > 1 {
			return filepath.ToSlash(filepath.Join(dir, split[len(split)-1]))
		}
	}
	return ""
}
