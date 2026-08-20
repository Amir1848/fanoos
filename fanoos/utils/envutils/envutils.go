package envutils

import (
	"log"
	"os"
	"strings"
)

// Loads a .env file from the working directory into the process environment
// before anything else runs. Real environment variables always win: values
// already set are NOT overridden (container/CI env takes precedence over the
// file).
//
// Packages that read configuration from the environment inside their own
// init() — utils/dbutils (DB_*), utils/mqttutils (MQTT_*) — blank-import
// this package, and Go's dependency-ordered initialization guarantees the
// file is loaded before those inits run. Runtime-only readers (e.g. the
// assistant's ASSISTANT_* variables) pick the values up automatically.
func init() {
	if err := load(".env"); err != nil {
		log.Printf("envutils: loading .env failed: %v", err)
	}
}

func load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		// Strip matching surrounding quotes.
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return nil
}
