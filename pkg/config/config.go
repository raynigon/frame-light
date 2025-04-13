package config

import (
	_ "embed"
	"encoding/json"
	"os"
	"strconv"
)

// Default configuration template
//
//go:embed config.json
var defaultConfig string

type Config struct {
	Development bool `json:"development"`
	Web         struct {
		Enabled bool `json:"enabled" validate:"required"`
		Port    int  `json:"port"`
	} `json:"web" validate:"required"`
	MQTT struct {
		Enabled  bool   `json:"enabled" validate:"required"`
		Broker   string `json:"broker"`
		ClientId string `json:"client_id"`
		// Authentication configuration
		Authentication struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"authentication"`
		// TLS configuration
		TLS struct {
			Enabled           bool   `json:"enabled"`
			CACertificate     string `json:"ca_certificate"`
			ClientKey         string `json:"client_key"`
			ClientCertificate string `json:"client_certificate"`
		} `json:"tls"`
		// Base topic for the MQTT client to publish and subscribe to
		Topic string `json:"topic"`
	} `json:"mqtt" validate:"required"`
	GPIO []struct {
		ID   int    `json:"id" validate:"required"`
		Name string `json:"name" validate:"required"`
	} `json:"gpio" validate:"required"`
}

// LoadConfig reads the configuration from a JSON file
func LoadConfig(filePath string) (*Config, error) {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create the file from the embedded template
		file, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		// Write the default configuration to the file
		_, err = file.WriteString(defaultConfig)
		if err != nil {
			return nil, err
		}
	}

	// Open the configuration file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return &config, err
	}

	return &config, nil
}

// GetWebAddress returns the web address in the format "host:port"
func (c *Config) GetWebAddress() string {
	if c.Web.Port == 0 {
		panic("Web server port is not set, must be greater than 0 and less than 65536")
	}
	return "localhost:" + strconv.Itoa(c.Web.Port)
}

// GetBrokerAddress returns the MQTT broker address in the format "tcp://host:port"
func (c *Config) GetBrokerAddress() string {
	// Parse the broker address to check if it is a valid URL
	if c.MQTT.Broker == "" {
		panic("MQTT broker address is not set")
	}
	// If TLS is enabled the scheme must be "ssl" or "wss"
	if c.MQTT.TLS.Enabled {
		if c.MQTT.Broker[:6] != "ssl://" && c.MQTT.Broker[:6] != "wss://" {
			panic("MQTT broker address must start with ssl:// or wss:// when TLS is enabled")
		}
	} else {
		if c.MQTT.Broker[:6] != "tcp://" && c.MQTT.Broker[:5] != "ws://" {
			panic("MQTT broker address must start with tcp:// or ws:// when TLS is disabled")
		}
	}
	return c.MQTT.Broker
}
