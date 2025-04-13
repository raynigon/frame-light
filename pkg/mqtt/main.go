package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/raynigon/frame-light/pkg/config"
	"github.com/raynigon/frame-light/pkg/gpio"
)

type MQTTService interface {
	StartAndPublish() error
	Close() error
}

type MQTTServiceImpl struct {
	// config
	config *config.Config
	// gpio service
	gpioService gpio.GpioService
	// MQTT client
	client mqtt.Client
	// MQTT base topic
	baseTopic string
	// GPIO listener ID
	listenerId int
}

func createTLSConfig(config *config.Config) (*tls.Config, error) {
	if config.MQTT.TLS.CACertificate == "" || config.MQTT.TLS.ClientKey == "" || config.MQTT.TLS.ClientCertificate == "" {
		return nil, errors.New("TLS configuration is incomplete")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{},
		RootCAs:      x509.NewCertPool(),
	}

	// Load CA certificate
	caCert, err := os.ReadFile(config.MQTT.TLS.CACertificate)
	if err != nil {
		return nil, err
	}
	tlsConfig.RootCAs.AppendCertsFromPEM(caCert)

	// Load client certificate and key
	cert, err := tls.LoadX509KeyPair(config.MQTT.TLS.ClientCertificate, config.MQTT.TLS.ClientKey)
	if err != nil {
		return nil, err
	}
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	return tlsConfig, nil
}

func NewMQTTService(config *config.Config, gpioService gpio.GpioService) MQTTService {
	options := mqtt.NewClientOptions()
	options.AddBroker(config.GetBrokerAddress())
	if config.MQTT.Authentication.Username != "" && config.MQTT.Authentication.Password != "" {
		options.SetUsername(config.MQTT.Authentication.Username)
		options.SetPassword(config.MQTT.Authentication.Password)
	}
	if config.MQTT.TLS.Enabled {
		tlsConfig, err := createTLSConfig(config)
		if err != nil {
			panic(err)
		}
		options.SetTLSConfig(tlsConfig)
	}
	client := mqtt.NewClient(options)
	return &MQTTServiceImpl{
		config:      config,
		gpioService: gpioService,
		client:      client,
		baseTopic:   config.MQTT.Topic,
		listenerId:  -1,
	}
}

func (m *MQTTServiceImpl) StartAndPublish() error {
	// Subscribe to the gpio service
	listenerId, err := m.gpioService.RegisterListener(m.handleCallback)
	if err != nil {
		return err
	}
	m.listenerId = listenerId

	// Connect to the MQTT broker
	connectToken := m.client.Connect()
	if connectToken.WaitTimeout(30*time.Second) && connectToken.Error() != nil {
		return connectToken.Error()
	}

	// Subscribe to the base topic to receive write messages
	subscribeToken := m.client.Subscribe(m.baseTopic+"/#", 0, m.handleMessage)
	if subscribeToken.WaitTimeout(30*time.Second) && subscribeToken.Error() != nil {
		return subscribeToken.Error()
	}

	return nil
}

func (m *MQTTServiceImpl) Close() error {
	if m.client != nil {
		m.client.Disconnect(0)
		return nil
	}
	return nil
}
