// Package tools wires the assistant chat endpoint to real Marketino devices:
// it injects the user's light inventory into the system prompt, declares
// device tools to the LLM (docs/architucture.md §9), and executes the model's
// tool calls through the same service functions as the REST API.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amir1848/marketino/internal/app/assistant"
	"github.com/amir1848/marketino/internal/app/device"
	"github.com/amir1848/marketino/utils/dbutils"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	g := router.Group("/assistant")

	g.POST("/chat", chat)
}

// light is one entry of the compact inventory injected into the system
// prompt. On/Online are nil when the device has never reported.
type light struct {
	DeviceID    int64  `json:"deviceId"`
	Name        string `json:"name"`
	Environment string `json:"environment"`
	On          *bool  `json:"on,omitempty"`
	Online      *bool  `json:"online,omitempty"`
}

func loadLights(ctx context.Context) ([]light, map[int64]string, error) {
	var rows []struct {
		EnvName    string
		DeviceID   int64
		DeviceName string
		IsOn       *bool
		Online     *bool
	}

	err := dbutils.GetDB().WithContext(ctx).Raw(`
		SELECT e.name AS env_name, d.id AS device_id, d.name AS device_name, ds.is_on, ds.online
		FROM devices d
		JOIN environments e ON e.id = d.environment_id
		LEFT JOIN device_states ds ON ds.device_guid = d.guid
		WHERE d.device_type = ?
		ORDER BY e.id, d.id
	`, device.TypeLight).Scan(&rows).Error
	if err != nil {
		return nil, nil, err
	}

	lights := make([]light, 0, len(rows))
	names := make(map[int64]string, len(rows))
	for _, row := range rows {
		lights = append(lights, light{
			DeviceID:    row.DeviceID,
			Name:        row.DeviceName,
			Environment: row.EnvName,
			On:          row.IsOn,
			Online:      row.Online,
		})
		names[row.DeviceID] = row.DeviceName
	}

	return lights, names, nil
}

func systemPrompt(lights []light) string {
	inventory, _ := json.Marshal(lights)

	return "You are the assistant of Marketino, a smart home automation platform, " +
		"talking to the user in a mobile app.\n\n" +
		"The user's controllable devices (lights only for now), as compact JSON:\n" + string(inventory) + "\n\n" +
		"Rules:\n" +
		"- Control lights only through the provided tools; the deviceId argument must be an id from the inventory above.\n" +
		"- Lights are safe and reversible: call turn_light right away when asked, no confirmation needed.\n" +
		"- get_light_state returns the last state a light reported over MQTT; known=false means it has never reported.\n" +
		"- After calling a tool, briefly confirm the result to the user (e.g. \"چراغ آشپزخانه روشن شد ✅\").\n" +
		"- If the user mentions a device that is not in the inventory, say you can't find it.\n" +
		"- Answer in the same language the user writes in (usually Persian). Keep replies short and friendly."
}

func toolDefinitions() []assistant.ToolDefinition {
	return []assistant.ToolDefinition{
		{
			Type: "function",
			Function: assistant.ToolFunction{
				Name:        "turn_light",
				Description: "Turn a light on or off. Use device ids from the inventory.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"deviceId": map[string]any{
							"type":        "integer",
							"description": "Numeric id of the light device, from the inventory",
						},
						"on": map[string]any{
							"type":        "boolean",
							"description": "true to turn on, false to turn off",
						},
					},
					"required": []string{"deviceId", "on"},
				},
			},
		},
		{
			Type: "function",
			Function: assistant.ToolFunction{
				Name:        "get_light_state",
				Description: "Get a light's last reported state (on/off, online). Use device ids from the inventory.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"deviceId": map[string]any{
							"type":        "integer",
							"description": "Numeric id of the light device, from the inventory",
						},
					},
					"required": []string{"deviceId"},
				},
			},
		},
	}
}

// deviceTools executes the model's tool calls through the same service
// functions the REST API uses, and produces Persian action events for the
// chat UI.
type deviceTools struct {
	names map[int64]string
}

func (t *deviceTools) Execute(ctx context.Context, call assistant.ToolCall) (string, any, error) {
	switch call.Function.Name {
	case "turn_light":
		return t.turnLight(ctx, call)
	case "get_light_state":
		return t.lightState(ctx, call)
	default:
		return "", nil, fmt.Errorf("unknown tool: %s", call.Function.Name)
	}
}

func (t *deviceTools) turnLight(ctx context.Context, call assistant.ToolCall) (string, any, error) {
	var args struct {
		DeviceID int64 `json:"deviceId"`
		On       bool  `json:"on"`
	}
	if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
		return "", nil, fmt.Errorf("invalid arguments for turn_light: %w", err)
	}

	name := t.deviceName(args.DeviceID)

	err := device.SendLightCommand(ctx, &device.LightCommandDTO{DeviceID: args.DeviceID, On: args.On})
	if err != nil {
		return "", nil, err
	}

	state := "خاموش شد"
	if args.On {
		state = "روشن شد"
	}

	content, _ := json.Marshal(map[string]any{"ok": true})
	event := map[string]any{
		"tool":       "turn_light",
		"deviceId":   args.DeviceID,
		"deviceName": name,
		"on":         args.On,
		"ok":         true,
		"text":       fmt.Sprintf("💡 %s %s", name, state),
	}

	return string(content), event, nil
}

func (t *deviceTools) lightState(ctx context.Context, call assistant.ToolCall) (string, any, error) {
	var args struct {
		DeviceID int64 `json:"deviceId"`
	}
	if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
		return "", nil, fmt.Errorf("invalid arguments for get_light_state: %w", err)
	}

	state, err := device.GetLightState(ctx, args.DeviceID)
	if err != nil {
		return "", nil, err
	}

	name := t.deviceName(args.DeviceID)

	text := fmt.Sprintf("💡 %s هنوز گزارشی نداده است", name)
	if state.Known && state.Online {
		if state.On {
			text = fmt.Sprintf("💡 %s روشن است", name)
		} else {
			text = fmt.Sprintf("💡 %s خاموش است", name)
		}
	} else if state.Known {
		text = fmt.Sprintf("💡 %s در دسترس نیست", name)
	}

	content, _ := json.Marshal(state)
	event := map[string]any{
		"tool":       "get_light_state",
		"deviceId":   args.DeviceID,
		"deviceName": name,
		"on":         state.On,
		"online":     state.Online,
		"ok":         true,
		"text":       text,
	}

	return string(content), event, nil
}

func (t *deviceTools) deviceName(id int64) string {
	if name, ok := t.names[id]; ok {
		return name
	}
	return fmt.Sprintf("دستگاه %d", id)
}

// chat streams the assistant reply as Server-Sent Events:
//   - data: {"delta": "..."}   — a text chunk of the reply
//   - data: {"action": {...}}  — a tool was executed (e.g. a light was
//     turned on); the UI renders these as inline chips
//   - data: {"error": "..."}   — failure after the stream started
//   - data: [DONE]
func chat(ctx *gin.Context) {
	var dto struct {
		Messages []assistant.Message `json:"messages"`
	}
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if assistant.APIKey() == "" {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "assistant is not configured: set ASSISTANT_API_KEY"})
		return
	}

	lights, names, err := loadLights(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load device inventory"})
		return
	}

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("X-Accel-Buffering", "no")
	ctx.Status(http.StatusOK)

	if err := assistant.RunChat(
		ctx.Request.Context(),
		systemPrompt(lights),
		dto.Messages,
		toolDefinitions(),
		&deviceTools{names: names},
		&sseEmitter{ctx: ctx},
	); err != nil {
		emitSSE(ctx, map[string]string{"error": err.Error()})
	}

	emitSSE(ctx, nil) // data: [DONE]
}

type sseEmitter struct {
	ctx *gin.Context
}

func (e *sseEmitter) Delta(text string) error {
	return emitSSE(e.ctx, map[string]string{"delta": text})
}

func (e *sseEmitter) Action(event any) error {
	return emitSSE(e.ctx, map[string]any{"action": event})
}

func emitSSE(ctx *gin.Context, payload any) error {
	data := "[DONE]"
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		data = string(encoded)
	}

	if _, err := ctx.Writer.Write(append(append([]byte("data: "), data...), '\n', '\n')); err != nil {
		return err
	}
	ctx.Writer.Flush()
	return nil
}
