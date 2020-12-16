package simulation_test

import (
	"testing"

	"simulation"
)

func TestExponential(t *testing.T) {

	Dtheta := func(s simulation.State) float64 {
		return s.X("Dtheta")
	}

	DDtheta := func(s simulation.State) float64 {
		return 1
	}
	sim := simulation.New()
	sim.SetChangeMap(map[simulation.Symbol]simulation.StateChanger{
		"theta":  Dtheta,
		"Dtheta": DDtheta,
	})
	sim.SetX0FromMap(map[simulation.Symbol]float64{
		"theta":  0,
		"Dtheta": 0,
	})

	sim.Options(simulation.OptionPrintResults, simulation.OptionStepDelay)
	sim.Begin()

}
