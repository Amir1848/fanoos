package assistant

import (
	"os"
	"strings"
)

const (
	defaultBaseURL = "https://openrouter.ai/api/v1"
	defaultModel   = "deepseek/deepseek-v4-pro"
)

func baseURL() string {
	return getEnv("ASSISTANT_BASE_URL", defaultBaseURL)
}

// APIKey returns the provider credentials; the chat endpoint returns 503
// until ASSISTANT_API_KEY is set.
func APIKey() string {
	return strings.TrimSpace(os.Getenv("ASSISTANT_API_KEY"))
}

func model() string {
	return getEnv("ASSISTANT_MODEL", defaultModel)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
