package assistant

import (
	"os"
	"strings"
)

const (
	defaultBaseURL    = "https://openrouter.ai/api/v1"
	defaultModel      = "deepseek/deepseek-v4-pro"
	defaultWhisperURL = "http://localhost:8176"
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

// WhisperURL is the base URL of the whisper.cpp server sidecar used for
// speech-to-text (docs/architucture.md §9).
func WhisperURL() string {
	return strings.TrimRight(getEnv("WHISPER_URL", defaultWhisperURL), "/")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
