package simulation

import (
	"fmt"
	"time"
)

// Option is an argument for Options
type Option struct {
	key      string
	duration time.Duration
	boolVal  bool
}

// OptionPrintResults print simulation results
var OptionPrintResults = Option{key: "printResults", boolVal: true}

// OptionStepDelay adds artificial delay to each RK4 solve cycle
var OptionStepDelay = Option{key: "stepDelay", duration: 100 * time.Millisecond}

// Options sets simulation configuration
func (sim *Simulation) Options(opts ...Option) {
	for _, opt := range opts {
		switch opt.key {
		case "printResults":
			if opt.boolVal {
				sim.config.printResults = true
			}
		case "stepDelay":
			sim.config.rk4delay = opt.duration
		}
	}
}

// throwf Gives you option to terminate simulation run inmediately due to error
func throwf(format string, a ...interface{}) {
	panic(fmt.Errorf(format, a...))
}
