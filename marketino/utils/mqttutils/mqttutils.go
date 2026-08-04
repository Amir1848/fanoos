// Package mqttutils holds the backend's MQTT client for talking to devices
// through the Mosquitto broker. Topic naming and payload shapes follow the
// contract in docs/architucture.md §8.
package mqttutils

import (
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var client mqtt.Client

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
		SetOnConnectHandler(func(mqtt.Client) {
			log.Printf("mqtt: connected to %s as %s", broker, clientID)
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

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
