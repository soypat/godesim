package godesim_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-sim/simulation"
	"github.com/soypat/godesim/state"
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
	const N_steps = 2
	sim.SetTimespan(0.0, 1, N_steps)
	sim.Begin()

	time, x_res := sim.Results("time"), sim.Results("theta")
	x_quad := applyFunc(time, func(v float64) float64 { return 1 / 2. * v * v })
	if len(time) != N_steps+1 {
		t.Errorf("Domain is not of length %d. got %d", N_steps+1, len(time))
	}
	for i := range x_quad {
		if math.Abs(x_quad[i]-x_res[i]) > 0.000001 {
			t.Errorf("Resulting curve not quadratic")
		}
	}
}

// Solves a simple system of equations of the form
//  Dtheta     = theta_dot
//  Dtheta_dot = 1
func Example_quadratic() {
	sim := simulation.New()
	sim.SetChangeMap(map[state.Symbol]state.Changer{
		"theta": func(s state.State) float64 {
			return s.X("theta-dot")
		},
		"theta-dot": func(s state.State) float64 {
			return 1
		},
	})
	sim.SetX0FromMap(map[state.Symbol]float64{
		"theta":     0,
		"theta-dot": 0,
	})
	sim.SetTimespan(0.0, 1.0, 10)
	sim.Begin()
	fmt.Printf("%0.3f:\n%0.3f", sim.Results("time"), sim.Results("theta"))
	// Output:
	//[0.000 0.100 0.200 0.300 0.400 0.500 0.600 0.700 0.800 0.900 1.000]:
	//[0.000 0.005 0.020 0.045 0.080 0.125 0.180 0.245 0.320 0.405 0.500]
}

func applyFunc(sli []float64, f func(float64) float64) []float64 {
	res := make([]float64, len(sli))
	for i, v := range sli {
		res[i] = f(v)
	}
	return res
}
