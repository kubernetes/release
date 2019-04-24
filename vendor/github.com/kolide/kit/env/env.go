/*
Package env provides utility functions for loading environment variables with defaults.

A common use of the env package is for combining flag with environment variables in a Go program.

Example:

	func main() {
		var (
			flProject = flag.String("http.addr", env.String("HTTP_ADDRESS", ":https"), "HTTP server address")
		)
		flag.Parse()
	}
*/
package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// String returns the environment variable value specified by the key parameter,
// otherwise returning a default value if set.
func String(key, def string) string {
	if env, ok := os.LookupEnv(key); ok {
		return env
	}
	return def
}

// Int returns the environment variable value specified by the key parameter,
// parsed as an integer. If the environment variable is not set, the default
// value is returned. If parsing the integer fails, Int will exit the program.
func Int(key string, def int) int {
	if env, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(env)
		if err != nil {
			fmt.Fprintf(os.Stderr, "env: parse int from flag: %s\n", err)
			os.Exit(1)
		}
		return i
	}
	return def
}

// Bool returns the environment variable value specified by the key parameter
// (parsed as a boolean), otherwise returning a default value if set.
func Bool(key string, def bool) bool {
	env, ok := os.LookupEnv(key)
	if !ok {
		return def
	}

	switch env {
	case "true", "T", "TRUE", "1":
		return true
	case "false", "F", "FALSE", "0":
		return false
	default:
		return def
	}
}

// Duration returns the environment variable value specified by the key parameter,
// otherwise returning a default value if set.
// If the time.Duration value cannot be parsed, Duration will exit the program
// with an error status.
func Duration(key string, def time.Duration) time.Duration {
	if env, ok := os.LookupEnv(key); ok {
		t, err := time.ParseDuration(env)
		if err != nil {
			fmt.Fprintf(os.Stderr, "env: parse time.Duration from flag: %s\n", err)
			os.Exit(1)
		}
		return t
	}
	return def
}
