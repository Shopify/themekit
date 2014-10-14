package phoenix

import "fmt"

func RedText(s string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", s)
}
