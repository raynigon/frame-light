package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/raynigon/frame-light/controller"
	"github.com/raynigon/frame-light/gpio"
	log "github.com/sirupsen/logrus"
)

var done chan struct{}
var server http.Server

// exists returns whether the given file or directory exists
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func shutdownHook() {
	c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Info("Stopping Application")
	close(done)
}

func startUp() {
	staticDirName := "./static"
	if !exists(staticDirName) {
		staticDirName = "../static"
	}
	// Register Handlers
	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir(staticDirName))))
	http.HandleFunc("/api/", controller.HandleAPICall)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", "/ui/")
		w.WriteHeader(302)
	})
	// Start Webserver
	server = http.Server{Addr: ":8080", Handler: nil}
	go server.ListenAndServe()
	log.Info("Started Application")
}

func main() {
	log.Info("Starting Application")
	done = make(chan struct{})
	go startUp()

	// Shutdown hook logic
	go shutdownHook()
	<-done
	defer gpio.Close()
	log.Info("Exit Application")
}
