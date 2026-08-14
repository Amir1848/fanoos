// Package firmware compiles a controller's environment definition (its
// linked light circuits, see docs/architucture.md §3) into ready-to-flash
// Arduino source plus a wiring README. Generation is a pure render of Go
// text/templates over database state — the same input always yields the
// same output, so nothing is stored.
package firmware

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/amir1848/marketino/internal/app/device"
	"github.com/amir1848/marketino/utils/dbutils"
	"gorm.io/gorm"
)

//go:embed templates/controller.ino.tmpl templates/README.md.tmpl
var templateFS embed.FS

// ControllerButtonPin is the GPIO the controller's own momentary button is
// wired to (INPUT_PULLUP). Kept away from every circuit pin below.
const ControllerButtonPin = 4

// pinPlan is the default per-circuit allocation in wiring order: each entry
// is {lightPin, buttonPin}. Pairs stay deterministic so regenerating the
// source never reshuffles pins someone has already wired.
//
// Chosen for a standard ESP32-WROOM-32 DevKit: only output-capable pins with
// internal pull-ups; boot-strapping pins GPIO 0/12 (and 15) are avoided, and
// input-only pins 34–39 (no pull-ups) are never used. Circuit 1 defaults to
// the built-in LED (GPIO 2) so the board is testable with no wiring at all,
// mirroring /firmware/controller/controller.ino.
var pinPlan = [8][2]int{
	{2, 5},
	{25, 26},
	{27, 14},
	{13, 23},
	{22, 21},
	{19, 18},
	{17, 16},
	{33, 32},
}

// MaxCircuits is how many light circuits one generated board can host.
const MaxCircuits = len(pinPlan)

// CircuitDTO is one light circuit as rendered into the templates. Field
// order matters: it mirrors a CIRCUITS[] row in controller.ino.
type CircuitDTO struct {
	Index        int
	GUID         string
	Name         string
	LightPin     int
	ButtonPin    int
	ActiveHigh   bool
	IsBuiltinLED bool
}

// ControllerFirmwareDTO is the generated source pair served to the frontend.
type ControllerFirmwareDTO struct {
	INO    string `json:"ino"`
	Readme string `json:"readme"`
}

type templateData struct {
	ControllerName      string
	EnvironmentID       int64
	ControllerGUID      string
	ClientID            string
	ControllerButtonPin int
	GeneratedDate       string
	CircuitCount        int
	Circuits            []CircuitDTO
}

// GenerateForController renders controller.ino + README.md for the given
// controller device: it looks up the controller, finds every light linked
// to it via a toggle device_link (either direction — the API allows both),
// assigns each circuit a default pin pair, and renders the templates.
func GenerateForController(ctx context.Context, deviceID int64) (*ControllerFirmwareDTO, error) {
	var ctrl struct {
		ID            int64
		EnvironmentID int64
		GUID          string
		Name          string
		DeviceType    int
	}
	err := dbutils.GetDB().WithContext(ctx).
		Table("devices").
		Select("id, environment_id, guid, name, device_type").
		Where("id = ?", deviceID).
		Take(&ctrl).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("device %d not found", deviceID)
	}
	if err != nil {
		return nil, err
	}

	if ctrl.DeviceType != device.TypeController {
		return nil, fmt.Errorf("device %d is not a controller", deviceID)
	}

	var rows []struct {
		GUID string
		Name string
	}
	err = dbutils.GetDB().WithContext(ctx).
		Table("device_links dl").
		Select("d.guid AS guid, d.name AS name").
		Joins("JOIN devices d ON (d.id = dl.target_device_id AND dl.source_device_id = ?) OR (d.id = dl.source_device_id AND dl.target_device_id = ?)", ctrl.ID, ctrl.ID).
		Where("dl.link_type = ? AND d.device_type = ?", 1, device.TypeLight).
		Order("d.id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	if len(rows) > MaxCircuits {
		return nil, fmt.Errorf("controller has %d circuits, more than the %d a single ESP32 can host", len(rows), MaxCircuits)
	}

	circuits := make([]CircuitDTO, 0, len(rows))
	for i, r := range rows {
		pins := pinPlan[i]
		circuits = append(circuits, CircuitDTO{
			Index:        i + 1,
			GUID:         r.GUID,
			Name:         cString(r.Name),
			LightPin:     pins[0],
			ButtonPin:    pins[1],
			ActiveHigh:   true, // every Marketino light is active-HIGH (HIGH = light on)
			IsBuiltinLED: i == 0,
		})
	}

	data := templateData{
		ControllerName:      cString(ctrl.Name),
		EnvironmentID:       ctrl.EnvironmentID,
		ControllerGUID:      ctrl.GUID,
		ClientID:            "esp32-" + firstToken(ctrl.GUID),
		ControllerButtonPin: ControllerButtonPin,
		GeneratedDate:       time.Now().UTC().Format("2006-01-02"),
		CircuitCount:        len(circuits),
		Circuits:            circuits,
	}

	ino, err := renderTemplate("controller.ino.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("rendering controller.ino failed: %w", err)
	}
	readme, err := renderTemplate("README.md.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("rendering README.md failed: %w", err)
	}

	return &ControllerFirmwareDTO{INO: ino, Readme: readme}, nil
}

func renderTemplate(name string, data templateData) (string, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/"+name)
	if err != nil {
		return "", err
	}

	var out strings.Builder
	if err := tmpl.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}

// cString escapes a name (which may be any UTF-8 text, e.g. Persian) into a
// C string literal body.
func cString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, "\\%03o", r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// firstToken returns the dash-free prefix of a GUID ("e1095eea" for
// "e1095eea-0c1f-..."), used to build a unique MQTT client id.
func firstToken(guid string) string {
	if i := strings.IndexByte(guid, '-'); i >= 0 {
		return guid[:i]
	}
	return guid
}
