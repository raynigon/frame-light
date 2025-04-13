package gpio

import "errors"

type GpioServiceMock struct {
	pins           map[string]string
	listeners      map[int]func(name string, state string)
	nextListenerId int
	closed         bool
}

func (g *GpioServiceMock) On(name string) error {
	return g.SetState(name, On)
}

func (g *GpioServiceMock) Off(name string) error {
	return g.SetState(name, Off)
}

func (g *GpioServiceMock) SetState(name string, state string) error {
	if g.closed {
		return nil
	}
	if _, ok := g.pins[name]; ok {
		g.pins[name] = state
		for _, listener := range g.listeners {
			listener(name, state)
		}
		return nil
	}
	return errors.New("pin not found")
}

func (g *GpioServiceMock) GetState(name string) (string, error) {
	if state, ok := g.pins[name]; ok {
		return state, nil
	}
	return "", errors.New("pin not found")
}

func (g *GpioServiceMock) GetStateForAll() (map[string]string, error) {
	if g.closed {
		return nil, nil
	}
	return g.pins, nil
}

func (g *GpioServiceMock) Close() error {
	if g.closed {
		return nil
	}
	g.closed = true
	return nil
}

func (g *GpioServiceMock) RegisterListener(callback func(name string, state string)) (int, error) {
	if g.closed {
		return 0, nil
	}
	id := g.nextListenerId
	g.listeners[id] = callback
	g.nextListenerId++
	return id, nil
}

func (g *GpioServiceMock) UnregisterListener(id int) error {
	if g.closed {
		return nil
	}
	if _, ok := g.listeners[id]; ok {
		delete(g.listeners, id)
		return nil
	}
	return nil
}
