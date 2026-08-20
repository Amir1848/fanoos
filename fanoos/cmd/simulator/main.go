// Command simulator stands in for real ESP32 firmware during development.
// It speaks the same MQTT topic/payload contract firmware will eventually
// use (see docs/architucture.md §8), so backend MQTT integration can be
// built and tested without physical hardware. It always plays the role of
// one ESP32 controller board: reporting its own press events, plus driving
// every light circuit wired to it (discovered from device_links, see
// loadCircuits) by applying `cmd` messages from the backend and reporting
// each circuit's resulting state.
//
// Unlike real firmware (which never talks to the database directly, per
// docs/architucture.md §7), this simulator reads the controller's circuits
// from Postgres at startup purely as a dev convenience.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/amir1848/fanoos/utils/dbutils"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Mirrors internal/app/device.TypeLight / TypeController. Not imported
// directly: that package pulls in utils/mqttutils, whose init() connects to
// the broker under a fixed client ID that would collide with this
// simulator's own connections.
const (
	dbDeviceTypeLight      = 1
	dbDeviceTypeController = 2
)

// linkTypeToggle marks a device_links row as "controller toggles light"
// (see devicelink.LinkTypeToggle). Today this is the only link type; the
// simulator drives every light linked to the controller this way.
const linkTypeToggle = 1

type controllerState struct {
	Type    string `json:"type"`
	Pressed bool   `json:"pressed"`
	Online  bool   `json:"online"`
	Ts      int64  `json:"ts"`
}

type lightState struct {
	Type   string `json:"type"`
	On     bool   `json:"on"`
	Online bool   `json:"online"`
	Ts     int64  `json:"ts"`
}

type lightCmd struct {
	On bool `json:"on"`
}

// circuit is one light wired to the simulated controller. Each circuit gets
// its own MQTT connection (and thus its own Last Will), matching the
// per-device contract in docs/architucture.md §8 exactly as if it were an
// independently connected light device.
type circuit struct {
	name       string
	guid       uuid.UUID
	cmdTopic   string
	stateTopic string

	client mqtt.Client
	mu     sync.Mutex
	on     bool
}

func main() {
	broker := flag.String("broker", "tcp://localhost:1883", "MQTT broker address")
	controllerGUIDFlag := flag.String("controller", "", "controller device GUID (devices.guid); its linked light circuits are simulated too")
	flag.Parse()

	if *controllerGUIDFlag == "" {
		log.Fatal("-controller is required")
	}
	controllerGUID, err := uuid.Parse(*controllerGUIDFlag)
	if err != nil {
		log.Fatalf("invalid -controller %q: not a GUID: %v", *controllerGUIDFlag, err)
	}

	envID, controllerName, circuits := loadCircuits(controllerGUID)

	log.Printf("controller %q (%s), env=%d, %d circuit(s):", controllerName, controllerGUID, envID, len(circuits))
	for _, c := range circuits {
		log.Printf("  - %s (%s)", c.name, c.guid)
	}

	controllerClient, controllerStateTopic := connectController(*broker, envID, controllerGUID)
	byName := make(map[string]*circuit, len(circuits))
	for _, c := range circuits {
		connectCircuit(*broker, c)
		byName[c.name] = c
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	log.Println(`type "press" to simulate the controller's button, or "<circuit-name> on|off" to toggle a circuit (Ctrl+C to quit)`)
	go readCommands(controllerClient, controllerStateTopic, byName)

	<-sigCh
	log.Println("shutting down")

	offlinePayload, _ := json.Marshal(map[string]any{"online": false, "ts": nowUnix()})
	publishRetained(controllerClient, controllerStateTopic, offlinePayload)
	controllerClient.Disconnect(250)
	for _, c := range circuits {
		publishRetained(c.client, c.stateTopic, offlinePayload)
		c.client.Disconnect(250)
	}
}

// loadCircuits looks up the controller device by GUID and every light
// linked to it via a device_links row with link_type=1 (toggle)
// (source_device_id=controller, target_device_id=light).
func loadCircuits(controllerGUID uuid.UUID) (envID int64, controllerName string, circuits []*circuit) {
	db := dbutils.GetDB()

	var ctrl struct {
		ID            int64
		EnvironmentID int64
		Name          string
		DeviceType    int
	}
	err := db.Table("devices").
		Select("id, environment_id, name, device_type").
		Where("guid = ?", controllerGUID).
		Take(&ctrl).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("controller device %s not found", controllerGUID)
	} else if err != nil {
		log.Fatalf("looking up controller %s failed: %v", controllerGUID, err)
	}
	if ctrl.DeviceType != dbDeviceTypeController {
		log.Fatalf("device %s (%q) is not a controller (device_type=%d)", controllerGUID, ctrl.Name, ctrl.DeviceType)
	}

	var rows []struct {
		GUID uuid.UUID
		Name string
	}
	err = db.Table("device_links dl").
		Select("d.guid AS guid, d.name AS name").
		Joins("JOIN devices d ON d.id = dl.target_device_id").
		Where("dl.source_device_id = ? AND dl.link_type = ? AND d.device_type = ?", ctrl.ID, linkTypeToggle, dbDeviceTypeLight).
		Order("d.id").
		Scan(&rows).Error
	if err != nil {
		log.Fatalf("looking up circuits for controller %s failed: %v", controllerGUID, err)
	}

	circuits = make([]*circuit, 0, len(rows))
	for _, r := range rows {
		circuits = append(circuits, &circuit{
			name:       r.Name,
			guid:       r.GUID,
			cmdTopic:   fmt.Sprintf("home/%d/device/%s/cmd", ctrl.EnvironmentID, r.GUID),
			stateTopic: fmt.Sprintf("home/%d/device/%s/state", ctrl.EnvironmentID, r.GUID),
		})
	}

	return ctrl.EnvironmentID, ctrl.Name, circuits
}

// connectController opens the controller's own MQTT connection and
// publishes its initial (unpressed) state. Controllers don't accept a cmd
// (see docs/architucture.md §8), so there's nothing to subscribe to.
func connectController(broker string, envID int64, guid uuid.UUID) (mqtt.Client, string) {
	stateTopic := fmt.Sprintf("home/%d/device/%s/state", envID, guid)
	offlinePayload, _ := json.Marshal(map[string]any{"online": false, "ts": nowUnix()})

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(fmt.Sprintf("sim-device-%s", guid)).
		SetAutoReconnect(true).
		SetWill(stateTopic, string(offlinePayload), 1, true).
		SetConnectRetry(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("controller: connect failed: %v", token.Error())
	}
	log.Printf("controller: state topic: %s", stateTopic)

	publishControllerState(client, stateTopic, false)
	return client, stateTopic
}

func publishControllerState(client mqtt.Client, stateTopic string, pressed bool) {
	payload, _ := json.Marshal(controllerState{Type: "controller", Pressed: pressed, Online: true, Ts: nowUnix()})
	publishRetained(client, stateTopic, payload)
}

// connectCircuit opens a light circuit's own MQTT connection, publishes its
// initial state, and subscribes to its cmd topic.
func connectCircuit(broker string, c *circuit) {
	offlinePayload, _ := json.Marshal(map[string]any{"online": false, "ts": nowUnix()})

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(fmt.Sprintf("sim-device-%s", c.guid)).
		SetAutoReconnect(true).
		SetWill(c.stateTopic, string(offlinePayload), 1, true).
		SetConnectRetry(true)

	c.client = mqtt.NewClient(opts)
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("circuit %s: connect failed: %v", c.name, token.Error())
	}
	log.Printf("%s: cmd topic:   %s", c.name, c.cmdTopic)
	log.Printf("%s: state topic: %s", c.name, c.stateTopic)

	c.publish()

	token := c.client.Subscribe(c.cmdTopic, 1, func(_ mqtt.Client, msg mqtt.Message) {
		var cmd lightCmd
		if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
			log.Printf("%s: ignoring malformed cmd on %s: %v", c.name, msg.Topic(), err)
			return
		}
		log.Printf("%s: received cmd: on=%v", c.name, cmd.On)
		c.setOn(cmd.On)
	})
	if token.Wait() && token.Error() != nil {
		log.Fatalf("circuit %s: subscribe to %s failed: %v", c.name, c.cmdTopic, token.Error())
	}
}

func (c *circuit) setOn(on bool) {
	c.mu.Lock()
	c.on = on
	c.mu.Unlock()
	c.publish()
}

func (c *circuit) publish() {
	c.mu.Lock()
	on := c.on
	c.mu.Unlock()
	payload, _ := json.Marshal(lightState{Type: "light", On: on, Online: true, Ts: nowUnix()})
	publishRetained(c.client, c.stateTopic, payload)
}

// readCommands lets the operator drive the simulated hardware by hand from
// stdin: "press" for the controller's own button, or "<circuit-name>
// on|off" to simulate a manual/physical toggle at that circuit.
func readCommands(controllerClient mqtt.Client, controllerStateTopic string, byName map[string]*circuit) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		if len(fields) == 1 && fields[0] == "press" {
			publishControllerState(controllerClient, controllerStateTopic, true)
			time.Sleep(150 * time.Millisecond)
			publishControllerState(controllerClient, controllerStateTopic, false)
			continue
		}

		if len(fields) >= 2 {
			name := strings.Join(fields[:len(fields)-1], " ")
			state := fields[len(fields)-1]
			c, ok := byName[name]
			if !ok {
				log.Printf("unknown circuit %q", name)
				continue
			}
			switch state {
			case "on":
				c.setOn(true)
				continue
			case "off":
				c.setOn(false)
				continue
			}
		}

		log.Println(`commands: "press", or "<circuit-name> on|off"`)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("stdin scan error: %v", err)
	}
}

func publishRetained(client mqtt.Client, topic string, payload []byte) {
	token := client.Publish(topic, 1, true, payload)
	token.Wait()
	if err := token.Error(); err != nil {
		log.Printf("publish to %s failed: %v", topic, err)
	}
}

func nowUnix() int64 {
	return time.Now().Unix()
}
