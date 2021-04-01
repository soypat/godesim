// Contains examples or tests which require external packages.
package godesim_test

import (
	"fmt"
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
