package env

import (
	"fmt"
	"os"
)

// GetenvOrDie retrieves the value for the keyed environment variable, or it will panic
// with a message if the environment variable isn't set.
func GetenvOrDie(key, detail string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Environment variable `%s` must be set: %s", key, detail))
	}

	return value
}

// GetenvOrDefault returns the value of the environment variable, or the default value if the
// environment variable isn't set.
func GetenvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}
