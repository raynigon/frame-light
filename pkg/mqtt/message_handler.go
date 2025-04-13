package mqtt

import (
	"encoding/json"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

type SetStateMessageContent struct {
	State string `json:"state"`
}

func (m *MQTTServiceImpl) handleMessage(client mqtt.Client, msg mqtt.Message) {
	// Ignore all messages which are not written on a topic ending with /set
	if !strings.HasSuffix(msg.Topic(), "/set") {
		return
	}
	// Parse the device name from the topic by splitting the topic at the "/" and using the second part
	// Example: "home/device/set" -> "device"
	deviceName := strings.TrimPrefix(msg.Topic(), m.baseTopic)
	deviceName = strings.TrimPrefix(deviceName, "/")
	deviceName = strings.TrimSuffix(deviceName, "/set")
	// Check that the device topic has no slashes
	if strings.Contains(deviceName, "/") && len(deviceName) > 0 {
		log.Errorf("Invalid device topic: %s", deviceName)
		return
	}
	// Parse the message to a json object
	var content SetStateMessageContent
	if err := json.Unmarshal(msg.Payload(), &content); err != nil {
		log.Errorf("Error parsing message: %s", err)
		return
	}
	// Check if the state is valid
	if content.State != "on" && content.State != "off" {
		log.Errorf("Invalid state: %s", content.State)
		return
	}
	// Check if the device is valid
	states, err := m.gpioService.GetStateForAll()
	if err != nil {
		log.Errorf("Error getting state for all devices: %s", err)
		return
	}
	if _, ok := states[deviceName]; !ok {
		log.Errorf("Unknown device: %s", deviceName)
		return
	}
	// Set the state of the device
	err = m.gpioService.SetState(deviceName, strings.ToUpper(content.State))
	if err != nil {
		log.Errorf("Error setting state: %s", err)
		return
	}
}
