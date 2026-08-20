package envutils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := "# comment\n" +
		"\n" +
		"ASSISTANT_API_KEY=test-key\n" +
		"  QUOTED_VALUE = \"hello world\"  \n" +
		"SINGLE='single'\n" +
		"MALFORMED_LINE\n" +
		"EMPTY=\n" +
		"PRESET=from-file\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	t.Setenv("PRESET", "from-env")

	if err := load(path); err != nil {
		t.Fatalf("load: %v", err)
	}

	for key, want := range map[string]string{
		"ASSISTANT_API_KEY": "test-key",
		"QUOTED_VALUE":      "hello world",
		"SINGLE":            "single",
		"EMPTY":             "",
		"PRESET":            "from-env",
	} {
		if got := os.Getenv(key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestLoadMissingFileIsNoop(t *testing.T) {
	if err := load(filepath.Join(t.TempDir(), "does-not-exist.env")); err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
}
