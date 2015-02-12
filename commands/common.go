package commands

import (
	"errors"
	"github.com/csaunders/phoenix"
)

func toClientAndFilesAsync(args map[string]interface{}, fn func(phoenix.ThemeClient, []string) chan bool) chan bool {
	var ok bool
	var themeClient phoenix.ThemeClient
	var filenames []string

	if themeClient, ok = args["themeClient"].(phoenix.ThemeClient); !ok {
		phoenix.HaltAndCatchFire(errors.New("themeClient is not of valid type"))
	}

	if filenames, ok = args["filenames"].([]string); !ok {
		phoenix.HaltAndCatchFire(errors.New("filenames are not of valid type"))
	}
	return fn(themeClient, filenames)
}
