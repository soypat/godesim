package godesim_test

import (
	"math"
	"testing"

	"github.com/soypat/godesim"
	"github.com/soypat/godesim/state"
)

var stiffDiff = map[state.Symbol]state.Diff{
	"x":  func(s state.State) float64 { return s.X("Dx") },
	"Dx": func(s state.State) float64 { return -50 * (s.X("x") - math.Cos(s.Time())) },
}
var stiffX0 = map[state.Symbol]float64{
	"x":  0,
	"Dx": -1,
}

func BenchmarkRK4(b *testing.B) {
	sim := godesim.New()
	sim.Algorithm.Steps = b.N
	sim.SetTimespan(0, 100., 1)

	sim.SetDiffFromMap(stiffDiff)
	sim.SetX0FromMap(stiffX0)
	sim.Begin()
}

func BenchmarkRKF45(b *testing.B) {
	sim := godesim.New()
	sim.Solver = godesim.RKF45Solver
	sim.Algorithm.Steps = b.N
	sim.SetTimespan(0, 100., 1)

	sim.SetDiffFromMap(stiffDiff)
	sim.SetX0FromMap(stiffX0)
	sim.Begin()
}

func BenchmarkRKF45Tableau(b *testing.B) {
	sim := godesim.New()
	sim.Solver = godesim.RKF45TableauSolver
	sim.Algorithm.Steps = b.N
	sim.SetTimespan(0, 100., 1)
	sim.SetDiffFromMap(stiffDiff)
	sim.SetX0FromMap(stiffX0)
	sim.Begin()
}
