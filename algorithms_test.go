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

func TestConvergenceRKF45(t *testing.T) {
	sim := godesim.New()
	sim.Solver = godesim.RKF45Solver
	sim.SetTimespan(0, 42., 1)
	// set up adaptive timestep
	sim.Algorithm.Error.Max = 1e-4
	sim.Algorithm.Step.Max, sim.Algorithm.Step.Min = 100, 1e-6

	sim.SetDiffFromMap(stiffDiff)
	sim.SetX0FromMap(stiffX0)
	sim.Begin()
	// tm := sim.Results("time")
	// t.Errorf("%.2f\n", tm)
	// fmt.Printf("%.2f\n", tm)
}

/*
// Benchmarks
*/
func BenchmarkRK4(b *testing.B) {
	sim := godesim.New()
	sim.Algorithm.Steps = b.N
	sim.SetTimespan(0, 100., 1)

	sim.SetDiffFromMap(stiffDiff)
	sim.SetX0FromMap(stiffX0)
	sim.Begin()
}

func BenchmarkRK5(b *testing.B) {
	sim := godesim.New()
	sim.Solver = godesim.RKF45Solver
	sim.Algorithm.Steps = b.N
	sim.SetTimespan(0, 100., 1)
	// No adaptive timestepping, only 5th order RK
	sim.SetDiffFromMap(stiffDiff)
	sim.SetX0FromMap(stiffX0)
	sim.Begin()
}
func BenchmarkRKF45(b *testing.B) {
	sim := godesim.New()
	sim.Solver = godesim.RKF45Solver
	sim.Algorithm.Steps = b.N
	sim.SetTimespan(0, 100., 1)
	// set up adaptive timestep
	expectedRKStep := sim.Dt() / float64(b.N)
	sim.Algorithm.Error.Max = .1
	sim.Algorithm.Step.Min, sim.Algorithm.Step.Max = expectedRKStep, math.Max(expectedRKStep, 4.)

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

/*
func BenchmarkNewton(b *testing.B) {
	sim := godesim.New()
	sim.Solver = godesim.NewtonIterativeSolver
	sim.Algorithm.Steps = b.N
	sim.SetTimespan(0, 100., 1)
	sim.SetDiffFromMap(stiffDiff)
	sim.SetX0FromMap(stiffX0)
	sim.Begin()
	v := sim.Results("x")
	fmt.Printf("%v", v)
}
*/
