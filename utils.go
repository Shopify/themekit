package phoenix

import "fmt"

func RedText(s string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", s)
}

func YellowText(s string) string {
	return fmt.Sprintf("\033[33m%s\033[0m", s)
}

func BlueText(s string) string {
	return fmt.Sprintf("\033[34m%s\033[0m", s)
}
