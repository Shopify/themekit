package commands

import (
	"errors"
	"fmt"
	"github.com/csaunders/phoenix"
)

func extractString(s *string, key string, args map[string]interface{}) {
	var ok bool
	if args[key] == nil {
		return
	}

	if *s, ok = args[key].(string); !ok {
		errMsg := fmt.Sprintf("%s is not of valid type", key)
		phoenix.HaltAndCatchFire(errors.New(errMsg))
	}
}

func extractInt(s *int, key string, args map[string]interface{}) {
	var ok bool
	if args[key] == nil {
		return
	}

	if *s, ok = args[key].(int); !ok {
		errMsg := fmt.Sprintf("%s is not of valid type", key)
		phoenix.HaltAndCatchFire(errors.New(errMsg))
	}
}

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
