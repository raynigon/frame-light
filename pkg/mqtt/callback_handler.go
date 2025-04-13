package mqtt

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

type GpioMessageContent struct {
	State string `json:"state"`
}

func (m *MQTTServiceImpl) handleCallback(name string, state string) {
	topicName := m.baseTopic + "/" + name
	content := GpioMessageContent{
		State: state,
	}
	message, err := json.Marshal(content)
	if err != nil {
		log.Errorf("Error marshalling state: %s", err)
		return
	}
	token := m.client.Publish(topicName, 0, false, message)
	if token.WaitTimeout(30*time.Second) && token.Error() != nil {
		log.Errorf("Error publishing message: %s", token.Error())
		return
	}
	log.Infof("Published message to topic %s: %s", topicName, message)
}
