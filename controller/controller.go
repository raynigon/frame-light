package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/raynigon/frame-light/gpio"
)

// HandleAPICall handles an incomming call starting with the url "/api/"
func HandleAPICall(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path[1:], "/")
	if len(parts) < 3 {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Path '%v' is not implemented yet", parts)
		return
	}
	pinID, err := strconv.Atoi(parts[2])
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "PinID is not valid, given: %s", parts[2])
		return
	} else if len(parts) < 4 || parts[3] == "" {
		w.WriteHeader(200)
		fmt.Fprintf(w, "{\"status\": \"%s\"}", gpio.GetState(pinID))
		return
	}
	var pinState bool
	if parts[3] == "on" {
		pinState = true
	} else if parts[3] == "off" {
		pinState = false
	} else {
		w.WriteHeader(400)
		fmt.Fprintf(w, "PinState is not valid, given: %s", parts[3])
		return
	}
	if pinState {
		gpio.On(pinID)
	} else {
		gpio.Off(pinID)
	}
	fmt.Fprintf(w, "Pin with id '%d' was set to state '%v'", pinID, parts[3])
}
