// Contains examples or tests which require external packages.
package godesim_test

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/soypat/godesim"
	"github.com/soypat/godesim/state"
)

func TestSimLoggerToCSV(t *testing.T) {
	logcfg := godesim.LoggerOptions{}
	logcfg.Results.Separator = ","
	logcfg.Results.AllStates = true
	logcfg.Results.FormatLen = 6

	cfg := godesim.Config{
		Domain: "time",
		Log:    logcfg,
	}
	cfg.Algorithm.Steps = 1
	sim := godesim.New()
	sim.SetConfig(cfg)
	sim.SetDiffFromMap(map[state.Symbol]state.Diff{
		"y": func(s state.State) float64 { return 0.1 },
		"x": func(s state.State) float64 { return 0.1 },
	})
	sim.SetX0FromMap(map[state.Symbol]float64{
		"y": 0,
		"x": 1,
	})
	const nsteps = 10
	sim.SetTimespan(0, 1, nsteps)
	var out = &strings.Builder{}
	sim.Logger.Output = out
	sim.Begin()
	lines := strings.Split(out.String(), "\n")
	lines = lines[1:] // first line are header
	for i := range lines {
		vals := strings.Split(lines[i], ",")
		if i > nsteps {
			break
		}
		for j := range vals {
			_, err := strconv.ParseFloat(strings.TrimSpace(vals[j]), 64)
			if err != nil {
				t.Error(err)
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

// Solve a stiff equation problem
// using the Newton-Raphson method.
// Equation being solved taken from https://en.wikipedia.org/wiki/Stiff_equation
// Do note that the accuracy for stiff problems is reduced greatly for conventional
// methods.
//  y'(t) = -15*y(t)
//  solution: y(t) = exp(-15*t)
func Example_implicit() {
	sim := godesim.New()
	tau := -15.
	solution := func(x []float64) []float64 {
		sol := make([]float64, len(x))
		for i := range x {
			sol[i] = math.Exp(tau * x[i])
		}
		return sol
	}
	sim.SetDiffFromMap(map[state.Symbol]state.Diff{
		"y": func(s state.State) float64 {
			return tau * s.X("y")
		},
	})
	sim.SetX0FromMap(map[state.Symbol]float64{
		"y": 1,
	})
	sim.SetTimespan(0.0, 0.5, 15)
	sim.Begin()
	fmt.Printf("domain  :%0.3f:\nresult  :%0.3f\nsolution:%0.3f", sim.Results("time"), sim.Results("y"), solution(sim.Results("time")))
	// Output:
	//domain  :[0.000 0.033 0.067 0.100 0.133 0.167 0.200 0.233 0.267 0.300 0.333 0.367 0.400 0.433 0.467 0.500]:
	//result  :[1.000 0.607 0.368 0.223 0.136 0.082 0.050 0.030 0.018 0.011 0.007 0.004 0.002 0.002 0.001 0.001]
	//solution:[1.000 0.607 0.368 0.223 0.135 0.082 0.050 0.030 0.018 0.011 0.007 0.004 0.002 0.002 0.001 0.001]
}
