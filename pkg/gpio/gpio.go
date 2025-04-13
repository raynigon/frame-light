package gpio

import (
	"fmt"

	"github.com/raynigon/frame-light/pkg/config"
	"github.com/stianeikeland/go-rpio"
)

// Enum for Pin states
const (
	Off = "OFF"
	On  = "ON"
)

type GpioService interface {
	On(name string) error
	Off(name string) error
	SetState(name string, state string) error
	GetState(name string) (string, error)
	GetStateForAll() (map[string]string, error)
	Close() error
	RegisterListener(callback func(name string, state string)) (int, error)
	UnregisterListener(id int) error
}

type GpioServiceImpl struct {
	pins           map[string]rpio.Pin
	listeners      map[int]func(name string, state string)
	nextListenerId int
	closed         bool
}

// NewGpioService creates a new GpioService instance
func NewGpioService(config *config.Config) GpioService {
	if config.Development {
		mock := GpioServiceMock{
			pins:           make(map[string]string),
			listeners:      make(map[int]func(name string, state string)),
			nextListenerId: 1,
			closed:         false,
		}
		for _, gpio := range config.GPIO {
			mock.pins[gpio.Name] = Off
		}
		return &mock
	}
	err := rpio.Open()
	if err != nil {
		panic(err)
	}
	pins := make(map[string]rpio.Pin)
	for _, gpio := range config.GPIO {
		pin := rpio.Pin(gpio.ID)
		pin.Output()
		pins[gpio.Name] = pin
	}
	return &GpioServiceImpl{
		pins:           pins,
		listeners:      make(map[int]func(name string, state string)),
		nextListenerId: 1,
		closed:         false,
	}
}

// Close ensures a valid state for all pins
func (g *GpioServiceImpl) Close() error {
	if g.closed {
		return nil
	}
	for _, pin := range g.pins {
		pin.Input()
	}
	g.closed = true
	return rpio.Close()
}

// On activates a single pin
func (g *GpioServiceImpl) On(name string) error {
	return g.SetState(name, On)
}

// Off deactivates a single pin
func (g *GpioServiceImpl) Off(name string) error {
	return g.SetState(name, Off)
}

// SetState sets the state of a given gpio pin
func (g *GpioServiceImpl) SetState(name string, state string) error {
	if pin, ok := g.pins[name]; ok {
		if state == On {
			pin.Low()
		} else if state == Off {
			pin.High()
		} else {
			return fmt.Errorf("invalid state %s", state)
		}
		for _, listener := range g.listeners {
			go listener(name, state)
		}
		return nil
	}
	return fmt.Errorf("pin %s not found", name)
}

// GetState returns the state of a given gpio pin
func (g *GpioServiceImpl) GetState(name string) (string, error) {
	if pin, ok := g.pins[name]; ok {
		if pin.Read() == rpio.Low {
			return On, nil
		}
		return Off, nil
	}
	return "", fmt.Errorf("pin %s not found", name)
}

// GetStateForAll returns the state of all gpio pins
func (g *GpioServiceImpl) GetStateForAll() (map[string]string, error) {
	states := make(map[string]string)
	for name, pin := range g.pins {
		if pin.Read() == rpio.Low {
			states[name] = On
		} else {
			states[name] = Off
		}
	}
	return states, nil
}

// RegisterListener registers a callback function to be called when a pin state changes
func (g *GpioServiceImpl) RegisterListener(callback func(name string, state string)) (int, error) {
	id := g.nextListenerId
	g.listeners[id] = callback
	g.nextListenerId++
	return id, nil
}

// UnregisterListener unregisters a callback function
func (g *GpioServiceImpl) UnregisterListener(id int) error {
	if _, ok := g.listeners[id]; ok {
		delete(g.listeners, id)
		return nil
	}
	return fmt.Errorf("listener with id %d not found", id)
}
