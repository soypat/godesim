package godesim_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/soypat/godesim"
	"github.com/soypat/godesim/state"
)

// add all testable solvers here
var gdsimSolvers = []func(*godesim.Simulation) []state.State{godesim.RK4Solver, godesim.RKF45Solver, godesim.RKF45TableauSolver, godesim.NewtonIterativeSolver}

func TestQuadratic(t *testing.T) {
	for _, solver := range gdsimSolvers {
		Dtheta := func(s state.State) float64 {
			return s.X("Dtheta")
		}

		DDtheta := func(s state.State) float64 {
			return 1
		}
		sim := godesim.New()
		sim.SetDiffFromMap(map[state.Symbol]state.Diff{
			"theta":  Dtheta,
			"Dtheta": DDtheta,
		})
		sim.SetX0FromMap(map[state.Symbol]float64{
			"theta":  0,
			"Dtheta": 0,
		})
		const N_steps = 10
		sim.Solver = solver
		sim.SetTimespan(0.0, 1, N_steps)

		sim.Begin()

		time, x_res := sim.Results("time"), sim.Results("theta")
		x_quad := applyFunc(time, func(v float64) float64 { return 1 / 2. * v * v /* solution is theta(t) = 1/2*t^2 */ })
		if len(time) != N_steps+1 || sim.Len() != N_steps {
			t.Errorf("Domain is not of length %d. got %d", N_steps+1, len(time))
		}
		for i := range x_quad {
			if math.Abs(x_quad[i]-x_res[i]) > math.Pow(sim.Dt()/float64(sim.Algorithm.Steps), 4) {
				t.Errorf("incorrect curve profile for test %s", t.Name())
			}
		}
	}
}

func TestSimpleInput(t *testing.T) {
	for _, solver := range gdsimSolvers {
		Dtheta := func(s state.State) float64 {
			return s.U("u")
		}

		inputVar := func(s state.State) float64 {
			return 1
		}
		sim := godesim.New()
		sim.SetDiffFromMap(map[state.Symbol]state.Diff{
			"theta": Dtheta,
		})
		sim.SetX0FromMap(map[state.Symbol]float64{
			"theta": 0,
		})
		sim.SetInputFromMap(map[state.Symbol]state.Input{
			"u": inputVar,
		})
		sim.Solver = solver
		const N_steps = 5
		sim.SetTimespan(0.0, 1, N_steps)
		sim.Begin()

		time, x_res := sim.Results("time"), sim.Results("theta")
		x_quad := applyFunc(time, func(v float64) float64 { return v /* solution is theta(t) = t*/ })
		if len(time) != N_steps+1 {
			t.Errorf("Domain is not of length %d. got %d", N_steps+1, len(time))
		}
		for i := range x_quad {
			if math.Abs(x_quad[i]-x_res[i]) > math.Pow(sim.Dt()/float64(sim.Algorithm.Steps), 4) {
				t.Errorf("incorrect curve profile for test %s", t.Name())
			}
		}
	}

}

// Solves a simple system of equations of the form
//  Dtheta     = theta_dot
//  Dtheta_dot = 1
func Example_quadratic() {
	sim := godesim.New()
	sim.SetDiffFromMap(map[state.Symbol]state.Diff{
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
