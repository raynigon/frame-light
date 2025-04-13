package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/raynigon/frame-light/pkg/gpio"
)

type APIController struct {
	gpioService gpio.GpioService
}

type Device struct {
	Name  string `json:"name"`
	State string `json:"state"`
	Type  string `json:"type"`
}

type APIResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (c *APIController) ListAllDevices(w http.ResponseWriter, r *http.Request) {
	states, err := c.gpioService.GetStateForAll()
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	var devices []Device
	for name, state := range states {
		devices = append(devices, Device{
			Name:  name,
			State: state,
			Type:  "gpio", // Assuming all devices are GPIO for now
		})
	}

	writeJSONResponse(w, http.StatusOK, APIResponse{
		Status: "ok",
		Data:   devices,
	})
}

func (c *APIController) GetDeviceEndpoint(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/"), "/")
	if len(pathParts) != 2 || pathParts[0] != "devices" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	name := pathParts[1]
	c.GetDevice(w, r, name)
}

func (c *APIController) GetDevice(w http.ResponseWriter, r *http.Request, name string) {
	state, err := c.gpioService.GetState(name)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	device := Device{
		Name:  name,
		State: state,
		Type:  "gpio", // Assuming all devices are GPIO for now
	}

	writeJSONResponse(w, http.StatusOK, APIResponse{
		Status: "ok",
		Data:   device,
	})
}

func (c *APIController) UpdateDeviceStateEndpoint(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/"), "/")
	if len(pathParts) != 3 || pathParts[0] != "devices" || pathParts[2] != "state" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	name := pathParts[1]
	c.UpdateDeviceState(w, r, name)
}

func (c *APIController) UpdateDeviceState(w http.ResponseWriter, r *http.Request, name string) {
	var requestBody struct {
		State string `json:"state"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, APIResponse{
			Status: "error",
			Error:  "Invalid request body",
		})
		return
	}

	switch strings.ToUpper(requestBody.State) {
	case "ON":
		err = c.gpioService.On(name)
	case "OFF":
		err = c.gpioService.Off(name)
	default:
		writeJSONResponse(w, http.StatusBadRequest, APIResponse{
			Status: "error",
			Error:  "Invalid state, must be 'ON' or 'OFF'",
		})
		return
	}

	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	writeJSONResponse(w, http.StatusOK, APIResponse{
		Status: "ok",
	})
}
