package tools

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func postAudio(body []byte) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/assistant/transcribe", transcribe)

	req := httptest.NewRequest(http.MethodPost, "/api/assistant/transcribe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=fanoosboundary")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func audioMultipartBody(t *testing.T, payload []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writer.SetBoundary("fanoosboundary"); err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile("audio", "audio.wav")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestTranscribeForwardsToWhisper(t *testing.T) {
	var gotPath, gotFormat string
	var gotBytes []int64

	whisper := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := r.ParseMultipartForm(maxAudioBytes); err != nil {
			t.Errorf("parse multipart: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		gotFormat = r.FormValue("response_format")
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("missing file field: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()
		data, _ := io.ReadAll(file)
		gotBytes = append(gotBytes, int64(len(data)))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"text": "  چراغ آشپزخانه را روشن کن  "}`))
	}))
	defer whisper.Close()
	t.Setenv("WHISPER_URL", whisper.URL)

	body := audioMultipartBody(t, []byte("fake-wav-bytes"))
	rec := postAudio(body)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if gotPath != "/inference" {
		t.Errorf("path = %s, want /inference", gotPath)
	}
	if gotFormat != "json" {
		t.Errorf("response_format = %s, want json", gotFormat)
	}
	if len(gotBytes) != 1 || gotBytes[0] != int64(len("fake-wav-bytes")) {
		t.Errorf("forwarded bytes = %v, want %d", gotBytes, len("fake-wav-bytes"))
	}
	if !strings.Contains(rec.Body.String(), `"text":"چراغ آشپزخانه را روشن کن"`) {
		t.Errorf("body = %s, want trimmed transcript", rec.Body.String())
	}
}

func TestTranscribeWhisperError(t *testing.T) {
	whisper := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer whisper.Close()
	t.Setenv("WHISPER_URL", whisper.URL)

	rec := postAudio(audioMultipartBody(t, []byte("fake-wav-bytes")))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}

func TestTranscribeUnreachableWhisper(t *testing.T) {
	t.Setenv("WHISPER_URL", "http://127.0.0.1:1")

	rec := postAudio(audioMultipartBody(t, []byte("fake-wav-bytes")))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}

func TestTranscribeMissingAudio(t *testing.T) {
	t.Setenv("WHISPER_URL", "http://127.0.0.1:1")

	rec := postAudio([]byte{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
