package integrationtests

import (
	"os"
)

func getFromEnvWithDefault(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultValue
}
