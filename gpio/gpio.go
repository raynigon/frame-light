package gpio

import (
	"github.com/stianeikeland/go-rpio"
)

var pins map[int]rpio.Pin

func init() {
	err := rpio.Open()
	if err != nil {
		panic(err)
	}
	pins = make(map[int]rpio.Pin)
	myPin := rpio.Pin(17)
	myPin.Output()
	pins[0] = myPin
}

// Close ensures a valid state for all pins
func Close() {
	rpio.Close()
}

// On activates a single pin
func On(pinID int) {
	if pin, ok := pins[pinID]; ok {
		pin.Low()
	}
}

// Off deactivates a single pin
func Off(pinID int) {
	if pin, ok := pins[pinID]; ok {
		pin.High()
	}
}

// GetState returns the state of a given gpio pin
func GetState(pinID int) string {
	if pin, ok := pins[pinID]; ok {
		if pin.Read() == rpio.Low {
			return "on"
		}
		return "off"
	}
	return "NONE"
}
