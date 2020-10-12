package exchanges

import (
	"fmt"
)

var exchangeReg = map[string]Factory{}

// Factory is used by output plugins to build an output instance
type Factory func() (Exchange, error)

// RegisterType registers a new output type.
func RegisterType(name string, f Factory) {
	if exchangeReg[name] != nil {
		panic(fmt.Errorf("exchange type  '%v' exists already", name))
	}
	exchangeReg[name] = f
}

// findFactory finds an output type its factory if available.
func findFactory(name string) Factory {
	return exchangeReg[name]
}

func Load(name string) (Exchange, error) {
	factory := findFactory(name)
	if factory == nil {
		return nil, fmt.Errorf("exchange type %v undefined", name)
	}

	return factory()
}
