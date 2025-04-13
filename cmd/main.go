package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/raynigon/frame-light/pkg/config"
	"github.com/raynigon/frame-light/pkg/controller"
	"github.com/raynigon/frame-light/pkg/gpio"
	"github.com/raynigon/frame-light/pkg/mqtt"
	log "github.com/sirupsen/logrus"
)

func shutdownHook(done chan struct{}) {
	c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Info("Stopping Application")
	close(done)
}

func getConfigLocation(args []string) string {
	configLocation := "./config.json"
	for i, arg := range args {
		if arg == "-c" || arg == "--config" {
			if i+1 < len(args) {
				return args[i+1]
			}
			log.Fatal("No config file provided")
		}
	}
	return configLocation
}

func main() {
	// Check arguments for config file location
	configLocation := getConfigLocation(os.Args[1:])

	// Load Config
	config, err := config.LoadConfig(configLocation)
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	// Initialize GPIO service
	gpioService := gpio.NewGpioService(config)
	if err != nil {
		log.Fatal("Error initializing GPIO service: ", err)
	}
	defer gpioService.Close()

	// Check if Web UI is enabled
	var server http.Server
	if config.Web.Enabled {
		// Register Handlers
		controller.RegisterHandlers(gpioService)

		// Start Webserver
		server = http.Server{
			Addr: config.GetWebAddress(),
		}
		go server.ListenAndServe()
		log.Infof("Webserver started on http://%s", config.GetWebAddress())
	} else {
		log.Info("Webserver disabled")
	}

	// Check if MQTT is enabled
	var mqttService mqtt.MQTTService
	if config.MQTT.Enabled {
		// Start MQTT Service
		mqttService = mqtt.NewMQTTService(config, gpioService)
		defer mqttService.Close()
		err := mqttService.StartAndPublish()
		if err != nil {
			log.Fatal("Error starting MQTT service: ", err)
		}
		log.Info("MQTT service started")
	} else {
		log.Info("MQTT service disabled")
	}

	log.Info("Started Application")
	// Wait for shutdown signal
	done := make(chan struct{})
	go shutdownHook(done)
	<-done
	// Shutdown signal received, start shutdown process
	log.Info("Stopping Webserver")
	if config.Web.Enabled {
		server.Close()
	}
	log.Info("Stopping MQTT Service")
	if config.MQTT.Enabled {
		mqttService.Close()
		log.Info("MQTT Service stopped")
	}
	log.Info("Stopping GPIO Service")
	gpioService.Close()
	// Shutdown complete, all services are stopped
	log.Info("Application stopped")
	log.Info("Bye")
}
