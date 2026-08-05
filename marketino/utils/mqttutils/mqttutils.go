// Package mqttutils holds the backend's MQTT client for talking to devices
// through the Mosquitto broker. Topic naming and payload shapes follow the
// contract in docs/architucture.md §8.
package mqttutils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var client mqtt.Client

// StateUpdate is a partial state report parsed off a device's `state`
// topic. A nil field means the message didn't mention it (e.g. a last-will
// "offline" message only carries Online) and any previously stored value
// for that field should be left as-is.
type StateUpdate struct {
	On     *bool
	Online *bool
	Ts     int64
}

// StateHandler persists a device's state update. Callers that need to read
// device state back out (e.g. the device package, over its own DB
// connection) register one via SetStateHandler; mqttutils itself doesn't
// store anything, it just parses topic/payload and forwards.
type StateHandler func(deviceGUID uuid.UUID, update StateUpdate)

var stateHandler StateHandler

// SetStateHandler registers the callback invoked for every message on a
// device's state topic. Only one handler is supported; call this once at
// startup.
func SetStateHandler(h StateHandler) {
	stateHandler = h
}

func init() {
	initClient()
}

func initClient() {
	broker := getEnv("MQTT_BROKER_URL", "tcp://localhost:1883")
	clientID := getEnv("MQTT_CLIENT_ID", "marketino-backend")

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(2 * time.Second).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Printf("mqtt: connected to %s as %s", broker, clientID)
			subscribeStateTopics(c)
		}).
		SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			log.Printf("mqtt: connection lost: %v", err)
		})

	client = mqtt.NewClient(opts)

	// Connecting can block for a while (or fail) if the broker isn't up yet;
	// don't let that stall backend startup like dbutils' DB connect does.
	go func() {
		token := client.Connect()
		token.Wait()
		if err := token.Error(); err != nil {
			log.Printf("mqtt: connect error: %v", err)
		}
	}()
}

// GetClient returns the shared MQTT client, connected in the background.
func GetClient() mqtt.Client {
	return client
}

// LightCmdTopic is where the backend publishes commands for a light device.
func LightCmdTopic(environmentID int64, deviceGUID uuid.UUID) string {
	return fmt.Sprintf("home/%d/device/%s/cmd", environmentID, deviceGUID)
}

// LightStateTopic is where a light device publishes its reported state.
func LightStateTopic(environmentID int64, deviceGUID uuid.UUID) string {
	return fmt.Sprintf("home/%d/device/%s/state", environmentID, deviceGUID)
}

// Publish sends a retained, QoS 1 message, matching the contract every
// device on the broker expects (see docs/architucture.md §8).
func Publish(topic string, payload []byte) error {
	c := GetClient()
	if c == nil || !c.IsConnectionOpen() {
		return fmt.Errorf("mqtt: not connected to broker")
	}

	token := c.Publish(topic, 1, true, payload)
	token.Wait()
	return token.Error()
}

// subscribeStateTopics listens to every device's state topic so the backend
// can answer "is this device on?" without round-tripping to the device.
// Re-subscribing on every connect (including reconnects) is required since
// paho drops subscriptions across a lost connection.
func subscribeStateTopics(c mqtt.Client) {
	token := c.Subscribe("home/+/device/+/state", 1, handleStateMessage)
	if token.Wait() && token.Error() != nil {
		log.Printf("mqtt: subscribe to state topics failed: %v", token.Error())
	}
}

func handleStateMessage(_ mqtt.Client, msg mqtt.Message) {
	if stateHandler == nil {
		return
	}

	// home/{environment_id}/device/{device_guid}/state
	parts := strings.Split(msg.Topic(), "/")
	if len(parts) != 5 {
		return
	}
	deviceGUID, err := uuid.Parse(parts[3])
	if err != nil {
		return
	}

	var raw struct {
		On     *bool `json:"on"`
		Online *bool `json:"online"`
		Ts     int64 `json:"ts"`
	}
	if err := json.Unmarshal(msg.Payload(), &raw); err != nil {
		log.Printf("mqtt: ignoring malformed state on %s: %v", msg.Topic(), err)
		return
	}

	stateHandler(deviceGUID, StateUpdate{On: raw.On, Online: raw.Online, Ts: raw.Ts})
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
