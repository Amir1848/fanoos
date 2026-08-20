// transcribe proxies short voice recordings to the whisper.cpp server
// sidecar (docs/architucture.md §9): the PWA records audio, the backend
// authenticates the user and forwards the file, and the returned text is
// meant to fill the chat draft for the user to review before sending.
package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/amir1848/fanoos/internal/app/assistant"
	"github.com/gin-gonic/gin"
)

const (
	maxAudioBytes  = 10 << 20
	whisperTimeout = 2 * time.Minute
)

var whisperClient = &http.Client{Timeout: whisperTimeout}

func RegisterTranscribeRoute(g *gin.RouterGroup) {
	g.POST("/transcribe", transcribe)
}

// transcribe accepts a multipart "audio" field (16 kHz mono WAV produced by
// the frontend), forwards it to the whisper sidecar's /inference endpoint,
// and returns {"text": "..."}.
func transcribe(ctx *gin.Context) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxAudioBytes)

	fileHeader, err := ctx.FormFile("audio")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "audio file is required"})
		return
	}
	if fileHeader.Size == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "audio file is empty"})
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read audio file"})
		return
	}
	defer src.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build transcription request"})
		return
	}
	if _, err := io.Copy(part, src); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read audio file"})
		return
	}
	if err := writer.WriteField("response_format", "json"); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build transcription request"})
		return
	}
	if err := writer.Close(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build transcription request"})
		return
	}

	req, err := http.NewRequestWithContext(ctx.Request.Context(), http.MethodPost, assistant.WhisperURL()+"/inference", &body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build transcription request"})
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := whisperClient.Do(req)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "transcription service is unreachable: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		message := strings.TrimSpace(string(detail))
		if message == "" {
			message = resp.Status
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("transcription failed: %s", message)})
		return
	}

	var out struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out); err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "invalid response from transcription service"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"text": strings.TrimSpace(out.Text)})
}
