package assistant

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Message is one turn of the conversation. ToolCallID and ToolCalls are only
// used in server-to-provider traffic (tool results and the assistant message
// that requested them); NormalizeMessages rejects them from clients.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// ToolDefinition declares a function the model may call
// (docs/architucture.md §9).
type ToolDefinition struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction is the schema half of a ToolDefinition.
type ToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ToolCall is a call the model wants to make. Function.Arguments is a JSON
// string that the provider streams in arbitrary fragments; chatRound
// reassembles it before the call reaches an Executor.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction carries the tool name and the raw JSON arguments.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Emitter receives the streamed reply: text chunks via Delta and executed
// tool events via Action.
type Emitter interface {
	Delta(text string) error
	Action(event any) error
}

// Executor runs one tool call requested by the model. It returns the JSON
// content to feed back to the provider plus an event for the client UI.
type Executor interface {
	Execute(ctx context.Context, call ToolCall) (string, any, error)
}

const (
	maxMessages       = 50
	maxContentChars   = 8000
	maxUpstreamErrorB = 2048
	streamLineBufSize = 1024 * 1024
	upstreamTimeout   = 5 * time.Minute
	maxToolRounds     = 5
)

var validRoles = map[string]bool{
	"user":      true,
	"assistant": true,
}

var httpClient = &http.Client{}

type upstreamChatRequest struct {
	Model    string           `json:"model"`
	Messages []Message        `json:"messages"`
	Tools    []ToolDefinition `json:"tools,omitempty"`
	Stream   bool             `json:"stream"`
}

type upstreamChunk struct {
	Choices []struct {
		Delta struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NormalizeMessages validates and sanitizes client-supplied conversation
// history. The system prompt is injected server-side only, so system and
// tool roles are rejected here.
func NormalizeMessages(raw []Message) ([]Message, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("messages must not be empty")
	}

	messages := make([]Message, 0, len(raw))
	for _, m := range raw {
		if !validRoles[m.Role] {
			return nil, fmt.Errorf("invalid message role: %s", m.Role)
		}

		content := strings.TrimSpace(m.Content)
		if content == "" {
			continue
		}

		runes := []rune(content)
		if len(runes) > maxContentChars {
			content = string(runes[:maxContentChars])
		}

		messages = append(messages, Message{Role: m.Role, Content: content})
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("messages must not be empty")
	}

	if len(messages) > maxMessages {
		messages = messages[len(messages)-maxMessages:]
	}

	return messages, nil
}

// RunChat streams the assistant reply to the client while letting the model
// call tools: each round the model either produces text (relayed via
// Emitter.Delta, conversation done) or requests tool calls (executed
// immediately, their results fed back to the model as role:"tool" messages
// for another round). Executed tools are surfaced to the client through
// Emitter.Action. The loop runs at most maxToolRounds rounds.
func RunChat(ctx context.Context, systemPrompt string, rawMessages []Message, tools []ToolDefinition, exec Executor, emit Emitter) error {
	messages, err := NormalizeMessages(rawMessages)
	if err != nil {
		return err
	}
	messages = append([]Message{{Role: "system", Content: systemPrompt}}, messages...)

	for round := 0; round < maxToolRounds; round++ {
		result, err := chatRound(ctx, messages, tools, emit)
		if err != nil {
			return err
		}

		if len(result.toolCalls) == 0 {
			return nil
		}

		messages = append(messages, Message{
			Role:      "assistant",
			Content:   result.content,
			ToolCalls: result.toolCalls,
		})

		for _, call := range result.toolCalls {
			content, event, err := exec.Execute(ctx, call)
			if err != nil {
				return fmt.Errorf("tool %s failed: %w", call.Function.Name, err)
			}
			if err := emit.Action(event); err != nil {
				return err
			}
			messages = append(messages, Message{Role: "tool", ToolCallID: call.ID, Content: content})
		}
	}

	return fmt.Errorf("assistant exceeded %d tool rounds", maxToolRounds)
}

// roundResult is what one provider round produced.
type roundResult struct {
	content   string
	toolCalls []ToolCall
}

// chatRound performs one streaming chat-completions request, relaying text
// deltas to emit as they arrive and reassembling streamed tool-call
// fragments (the provider sends the arguments JSON in chunks).
func chatRound(ctx context.Context, messages []Message, tools []ToolDefinition, emit Emitter) (roundResult, error) {
	payload, err := json.Marshal(upstreamChatRequest{
		Model:    model(),
		Messages: messages,
		Tools:    tools,
		Stream:   true,
	})
	if err != nil {
		return roundResult{}, err
	}

	reqCtx, cancel := context.WithTimeout(ctx, upstreamTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, baseURL()+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return roundResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+APIKey())

	resp, err := httpClient.Do(req)
	if err != nil {
		return roundResult{}, fmt.Errorf("assistant provider request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxUpstreamErrorB))
		detail := strings.TrimSpace(string(body))
		if detail == "" {
			detail = resp.Status
		}
		return roundResult{}, fmt.Errorf("assistant provider returned %s: %s", resp.Status, detail)
	}

	var (
		result    roundResult
		fragments = map[int]*ToolCall{}
		streamErr error
	)

	reader := bufio.NewReaderSize(resp.Body, streamLineBufSize)
	for streamErr == nil {
		line, err := reader.ReadString('\n')
		if line != "" {
			done, fwdErr := handleUpstreamLine(line, emit, &result, fragments)
			if fwdErr != nil {
				streamErr = fwdErr
				break
			}
			if done {
				break
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			streamErr = fmt.Errorf("assistant provider stream failed: %w", err)
		}
	}

	result.toolCalls = collectToolCalls(fragments)
	return result, streamErr
}

// handleUpstreamLine processes one SSE line, updating result and forwarding
// text deltas to emit. It returns done=true on the stream terminator
// ([DONE]).
func handleUpstreamLine(line string, emit Emitter, result *roundResult, fragments map[int]*ToolCall) (bool, error) {
	line = strings.TrimRight(line, "\r\n")
	if line == "" || !strings.HasPrefix(line, "data:") {
		return false, nil
	}

	data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	if data == "[DONE]" {
		return true, nil
	}

	var chunk upstreamChunk
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return false, nil
	}

	if chunk.Error != nil && chunk.Error.Message != "" {
		return false, fmt.Errorf("assistant provider error: %s", chunk.Error.Message)
	}

	if len(chunk.Choices) == 0 {
		return false, nil
	}

	delta := chunk.Choices[0].Delta
	if delta.Content != "" {
		result.content += delta.Content
		if err := emit.Delta(delta.Content); err != nil {
			return false, err
		}
	}

	for _, tc := range delta.ToolCalls {
		existing, ok := fragments[tc.Index]
		if !ok {
			fragments[tc.Index] = &ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
			continue
		}
		if tc.ID != "" {
			existing.ID = tc.ID
		}
		if tc.Type != "" {
			existing.Type = tc.Type
		}
		if tc.Function.Name != "" {
			existing.Function.Name += tc.Function.Name
		}
		existing.Function.Arguments += tc.Function.Arguments
	}

	return false, nil
}

// collectToolCalls turns the reassembled fragments into an index-ordered
// slice (models may stream several tool calls in one round).
func collectToolCalls(fragments map[int]*ToolCall) []ToolCall {
	if len(fragments) == 0 {
		return nil
	}

	indexes := make([]int, 0, len(fragments))
	for index := range fragments {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)

	calls := make([]ToolCall, 0, len(indexes))
	for _, index := range indexes {
		calls = append(calls, *fragments[index])
	}
	return calls
}
