package controller

import (
	"net/http"

	"github.com/raynigon/frame-light/pkg/gpio"
)

func RegisterHandlers(gpioService gpio.GpioService) {
	// Create API controller
	apiController := &APIController{gpioService: gpioService}
	// API handler for GPIO
	http.HandleFunc("GET /api/devices", apiController.ListAllDevices)
	http.HandleFunc("GET /api/devices/{name}", apiController.GetDeviceEndpoint)
	http.HandleFunc("POST /api/devices/{name}/state", apiController.UpdateDeviceStateEndpoint)
	// Handler for UI requests
	http.HandleFunc("/ui/", UIHandler)
	// Redirect root to UI
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", "/ui/")
		w.WriteHeader(302)
	})
}
