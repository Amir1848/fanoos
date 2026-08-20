package firmware

import (
	"strings"
	"testing"
)

func TestRenderTemplates(t *testing.T) {
	data := templateData{
		ControllerName:      `کنترلر "1" \ /`,
		EnvironmentID:       4,
		ControllerGUID:      "e1095eea-0c1f-4b4f-8ed5-82b9cd464764",
		ClientID:            "esp32-e1095eea",
		ControllerButtonPin: ControllerButtonPin,
		GeneratedDate:       "2026-08-14",
		CircuitCount:        2,
		Circuits: []CircuitDTO{
			{Index: 1, GUID: "2728e8e8-9ab4-4bcc-9480-8b273b87e613", Name: cString("چراغ 1"), LightPin: 2, ButtonPin: 5, ActiveHigh: true, IsBuiltinLED: true},
			{Index: 2, GUID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", Name: cString(`چراغ "2"`), LightPin: 25, ButtonPin: 26, ActiveHigh: true},
		},
	}

	ino, err := renderTemplate("controller.ino.tmpl", data)
	if err != nil {
		t.Fatalf("render controller.ino: %v", err)
	}
	readme, err := renderTemplate("README.md.tmpl", data)
	if err != nil {
		t.Fatalf("render README.md: %v", err)
	}

	for _, want := range []string{
		`static const char* ENVIRONMENT_ID   = "4";`,
		`static const char* CONTROLLER_GUID  = "e1095eea-0c1f-4b4f-8ed5-82b9cd464764";`,
		`static const uint8_t CONTROLLER_BUTTON_PIN = 4;`,
		`static const char* MQTT_CLIENT_ID = "esp32-e1095eea";`,
		`"2728e8e8-9ab4-4bcc-9480-8b273b87e613"`,
		"چراغ 1",
		`\"2\"`,
		`false, false, false,`,
		`YOUR_WIFI_SSID`,
	} {
		if !strings.Contains(ino, want) {
			t.Errorf("controller.ino missing %q", want)
		}
	}
	for _, want := range []string{
		"GPIO 4",
		"GPIO 25",
		"2728e8e8-9ab4-4bcc-9480-8b273b87e613",
		"CIRCUITS[]",
		"115200",
	} {
		if !strings.Contains(readme, want) {
			t.Errorf("README.md missing %q", want)
		}
	}
}
