package main

import (
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/raynigon/frame-light/pkg/config"
	"github.com/raynigon/frame-light/pkg/controller"
	"github.com/raynigon/frame-light/pkg/gpio"
	"github.com/raynigon/frame-light/pkg/mqtt"
	log "github.com/sirupsen/logrus"
)

//go:embed version.txt
var embeddedVersion string

var version = strings.TrimSpace(embeddedVersion)

func printHelp() {
	fmt.Println("Usage: gpio2mqtt [OPTIONS]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -version          Display the application version and exit")
	fmt.Println("  -help             Display this help message and exit")
	fmt.Println("  -config <FILE>    Specify the configuration file location")
	fmt.Println("  -c <FILE>         Alias for --config")
}

func parseArguments() string {
	// Define flags
	versionFlag := flag.Bool("version", false, "Display the application version and exit")
	helpFlag := flag.Bool("help", false, "Display help and exit")
	configFlag := flag.String("config", "./config.json", "Specify the configuration file location")
	flag.StringVar(configFlag, "c", "./config.json", "Alias for --config")

	// Parse flags
	flag.Parse()

	// Handle --version
	if *versionFlag {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}

	// Handle --help
	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	// Return the config file location
	return *configFlag
}

func shutdownHook(done chan struct{}) {
	c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Info("Stopping Application")
	close(done)
}

func main() {
	// Parse arguments
	configLocation := parseArguments()

	// Load Config
	config, err := config.LoadConfig(configLocation)
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	// Initialize GPIO service
	gpioService := gpio.NewGpioService(config)
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
