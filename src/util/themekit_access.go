package util

import "strings"

// IsThemeKitAccessPassword checks if the password is a Theme Kit Access password
func IsThemeKitAccessPassword(password string) bool {
	themeKitPasswordPrefix := "shptka_"
	return strings.HasPrefix(password, themeKitPasswordPrefix)
}
