package simulation_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-sim/simulation"
	"github.com/go-sim/simulation/state"
)

func TestQuadratic(t *testing.T) {

	Dtheta := func(s state.State) float64 {
		return s.X("Dtheta")
	}

	DDtheta := func(s state.State) float64 {
		return 1
	}
	sim := simulation.New()
	sim.SetChangeMap(map[state.Symbol]state.Changer{
		"theta":  Dtheta,
		"Dtheta": DDtheta,
	})
	sim.SetX0FromMap(map[state.Symbol]float64{
		"theta":  0,
		"Dtheta": 0,
	})
	sim.SetTimespan(0.0, 1.0, 10)
	sim.Begin()

	time, x_res := sim.TimeVector(), sim.XResults("theta")
	x_quad := applyFunc(time, func(v float64) float64 { return 1 / 2. * v * v })
	for i := range x_quad {
		if math.Abs(x_quad[i]-x_res[i]) > 0.000001 {
			t.Fail()
		}
	}
}

// Solves a simple
func Example_quadratic() {
	sim := simulation.New()
	sim.SetChangeMap(map[state.Symbol]state.Changer{
		"theta": func(s state.State) float64 {
			return s.X("Dtheta")
		},
		"Dtheta": func(s state.State) float64 {
			return 1
		},
	})
	sim.SetX0FromMap(map[state.Symbol]float64{
		"theta":  0,
		"Dtheta": 0,
	})
	sim.SetTimespan(0.0, 1.0, 10)
	sim.Begin()
	fmt.Printf("%v:\n%v", sim.TimeVector(), sim.XResults("theta"))
	// Output:
}

func applyFunc(sli []float64, f func(float64) float64) []float64 {
	res := make([]float64, len(sli))
	for i, v := range sli {
		res[i] = f(v)
	}
	return res
}
