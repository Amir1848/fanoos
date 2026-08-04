// Command simulator stands in for real ESP32 firmware during development.
// It speaks the same MQTT topic/payload contract firmware will eventually
// use (see docs/architucture.md §8), so backend MQTT integration can be
// built and tested without physical hardware.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	deviceTypeLight      = "light"
	deviceTypeController = "controller"
)

type lightState struct {
	Type   string `json:"type"`
	On     bool   `json:"on"`
	Online bool   `json:"online"`
	Ts     int64  `json:"ts"`
}

type controllerState struct {
	Type    string `json:"type"`
	Pressed bool   `json:"pressed"`
	Online  bool   `json:"online"`
	Ts      int64  `json:"ts"`
}

type lightCmd struct {
	On bool `json:"on"`
}

func main() {
	broker := flag.String("broker", "tcp://localhost:1883", "MQTT broker address")
	envID := flag.Int64("env", 0, "environment ID this device belongs to")
	deviceGUID := flag.String("device", "", "device GUID (matches the devices.guid column)")
	deviceType := flag.String("type", "", "device type: light|controller")
	clientID := flag.String("client-id", "", "MQTT client ID (default: sim-device-<guid>)")
	flag.Parse()

	if *envID == 0 || *deviceGUID == "" {
		log.Fatal("both -env and -device are required")
	}
	guid, err := uuid.Parse(*deviceGUID)
	if err != nil {
		log.Fatalf("invalid -device %q: not a GUID: %v", *deviceGUID, err)
	}
	if *deviceType != deviceTypeLight && *deviceType != deviceTypeController {
		log.Fatalf("invalid -type %q: must be %q or %q", *deviceType, deviceTypeLight, deviceTypeController)
	}
	if *clientID == "" {
		*clientID = fmt.Sprintf("sim-device-%s", guid)
	}

	cmdTopic := fmt.Sprintf("home/%d/device/%s/cmd", *envID, guid)
	stateTopic := fmt.Sprintf("home/%d/device/%s/state", *envID, guid)

	offlinePayload, _ := json.Marshal(map[string]any{"online": false, "ts": nowUnix()})

	opts := mqtt.NewClientOptions().
		AddBroker(*broker).
		SetClientID(*clientID).
		SetAutoReconnect(true).
		SetWill(stateTopic, string(offlinePayload), 1, true).
		SetConnectRetry(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("connect failed: %v", token.Error())
	}
	log.Printf("connected to %s as %s (env=%d device=%s type=%s)", *broker, *clientID, *envID, guid, *deviceType)
	log.Printf("cmd topic:   %s", cmdTopic)
	log.Printf("state topic: %s", stateTopic)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	switch *deviceType {
	case deviceTypeLight:
		runLight(client, cmdTopic, stateTopic)
	case deviceTypeController:
		runController(client, stateTopic)
	}

	<-sigCh
	log.Println("shutting down")
	publishRetained(client, stateTopic, offlinePayload)
	client.Disconnect(250)
}

func runLight(client mqtt.Client, cmdTopic, stateTopic string) {
	on := false

	publishState := func() {
		payload, _ := json.Marshal(lightState{Type: deviceTypeLight, On: on, Online: true, Ts: nowUnix()})
		publishRetained(client, stateTopic, payload)
	}

	token := client.Subscribe(cmdTopic, 1, func(_ mqtt.Client, msg mqtt.Message) {
		var cmd lightCmd
		if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
			log.Printf("ignoring malformed cmd on %s: %v", msg.Topic(), err)
			return
		}
		on = cmd.On
		log.Printf("cmd received: on=%v", on)
		publishState()
	})
	if token.Wait() && token.Error() != nil {
		log.Fatalf("subscribe failed: %v", token.Error())
	}

	publishState()
}

func runController(client mqtt.Client, stateTopic string) {
	publish := func(pressed bool) {
		payload, _ := json.Marshal(controllerState{Type: deviceTypeController, Pressed: pressed, Online: true, Ts: nowUnix()})
		publishRetained(client, stateTopic, payload)
	}
	publish(false)

	log.Println(`type "press" and Enter to simulate a button press (Ctrl+C to quit)`)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			switch strings.TrimSpace(scanner.Text()) {
			case "press":
				publish(true)
				time.Sleep(150 * time.Millisecond)
				publish(false)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("stdin scan error: %v", err)
		}
	}()
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
