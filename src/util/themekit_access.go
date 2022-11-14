package util

import "strings"

// IsThemeAccessPassword checks if the password is a Theme Access password
func IsThemeAccessPassword(password string) bool {
	themeKitPasswordPrefix := "shptka_"
	return strings.HasPrefix(password, themeKitPasswordPrefix)
}
