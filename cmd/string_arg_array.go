package cmd

import (
	"strings"
)

type stringArgArray struct {
	values []string
}

func (fa *stringArgArray) String() string {
	return strings.Join(fa.values, ",")
}

func (fa *stringArgArray) Set(value string) error {
	if len(value) > 0 {
		fa.values = append(fa.values, value)
	}
	return nil
}

func (fa *stringArgArray) Type() string {
	return "string"
}

func (fa *stringArgArray) Value() []string {
	if len(fa.values) == 0 {
		return nil
	}
	return fa.values
}
