package godesim

import (
	"math"
	"testing"

	"github.com/soypat/godesim/state"
)

// add all explicit testable solvers here
var explicitSolvers = []func(*Simulation) []state.State{RK4Solver, RKF45Solver, RKF45TableauSolver}

func TestQuadratic(t *testing.T) {
	for _, solver := range explicitSolvers {
		Dtheta := func(s state.State) float64 {
			return s.X("Dtheta")
		}

		DDtheta := func(s state.State) float64 {
			return 1
		}
		sim := New()
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
	for _, solver := range explicitSolvers {
		Dtheta := func(s state.State) float64 {
			return s.U("u")
		}

		inputVar := func(s state.State) float64 {
			return 1
		}
		sim := New()
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

// Stiff equation example based off https://en.wikipedia.org/wiki/Stiff_equation
func TestNewtonRaphson_stiff(t *testing.T) {
	// TODO change tau to -15
	tau := -15.
	sim := New()
	sim.SetDiffFromMap(map[state.Symbol]state.Diff{
		"y": func(s state.State) float64 { return tau * s.X("y") },
	})
	sim.SetX0FromMap(map[state.Symbol]float64{
		"y": 1.,
	})
	sim.Solver = NewtonRaphsonSolver
	sim.Config.Algorithm.Error.Max = 1e-6
	const N_steps = 50
	sim.SetTimespan(0.0, 1, N_steps)
	sim.Begin()

	time, x_res := sim.Results("time"), sim.Results("y")
	solution := applyFunc(time, func(v float64) float64 { return math.Exp(tau * v) })
	if len(time) != N_steps+1 {
		t.Errorf("Domain is not of length %d. got %d", N_steps+1, len(time))
	}
	permissibleErr := sim.Dt() * 4
	for i := range solution {
		if math.Abs(solution[i]-x_res[i]) > permissibleErr {
			t.Errorf("incorrect curve profile for test %s. got %0.3f. want %0.3f +/-%0.4g", t.Name(), x_res[i], solution[i], permissibleErr)
		}
	}
}

func TestNewtonRaphson_chemistry(t *testing.T) {
	sim := New()
	sim.SetDiffFromMap(map[state.Symbol]state.Diff{
		"y1": func(s state.State) float64 { return -0.04*s.X("y1") + 1e4*s.X("y2")*s.X("y3") },
		"y2": func(s state.State) float64 {
			return 0.04*s.X("y1") - 1e4*s.X("y2")*s.X("y3") - 3e7*math.Pow(s.X("y2"), 2.)
		},
		"y3": func(s state.State) float64 { return 3e7 * math.Pow(s.X("y2"), 2.) },
	})
	sim.SetX0FromMap(map[state.Symbol]float64{
		"y1": 1.,
		"y2": 0,
		"y3": 0,
	})
	// approximate solution. Less than 2e-3 error on all points up to time=600
	y1approx := func(t float64) float64 {
		var B = [...]float64{9.442029890550312e-03, 8.192125814943391e-02, 1.345288005759563e-04, 4.686920668485861e-10, -1.435963210148673e-11, 2.449294560790408e+00, 1.077867031678799}
		return B[6] - B[0]*math.Sqrt(t) - B[1]*math.Log(B[5]+t) + B[2]*t + B[3]*math.Pow(t, 2) + B[4]*math.Pow(t, 3)
	}

	sim.Solver = NewtonRaphsonSolver
	sim.Config.Algorithm.Error.Max = 1e-6
	const N_steps = 600 * 2
	sim.SetTimespan(0.0, 600, N_steps) // ten minutes simulated in 0.5 steps
	sim.Begin()

	time, x_res := sim.Results("time"), sim.Results("y1")
	time, x_res = time[1:], x_res[1:] // exclude first point (is singularity for approximate solution)
	solution := applyFunc(time, y1approx)
	if len(time) != N_steps {
		t.Errorf("Domain is not of length %d. got %d", N_steps+1, len(time))
	}
	for i := range solution {
		if math.Abs(solution[i]-x_res[i]) > 5e-3 {
			t.Errorf("incorrect curve profile for test %s. got %0.3f. want %0.3f", t.Name(), x_res[i], solution[i])
		}
	}
}

func TestTimespanErrors(t *testing.T) {
	var tests = []struct {
		start, end float64
		steps      int
	}{
		{start: 1., end: 0, steps: 10},
		{start: 0, end: 1., steps: 0},
		{start: 20., end: 20., steps: 10},
	}
	for i := range tests {
		err := recoverTimespanTest(tests[i].start, tests[i].end, tests[i].steps)
		if err == nil {
			t.Errorf("timespan should have panic'd with %#v", tests[i])
		}
	}
	defer func() {
		err := recover()
		if err == nil {
			t.Error("timespan should have panic'd with no timespan")
		}
	}()
	sim := New()
	sim.SetDiffFromMap(map[state.Symbol]state.Diff{"x": nil})
	sim.SetX0FromMap(map[state.Symbol]float64{"x": 1})
	sim.Begin()
}

func recoverTimespanTest(Start, End float64, Steps int) (i interface{}) {
	defer func() {
		i = recover()
	}()
	_ = newTimespan(Start, End, Steps)
	return nil
}

func TestBadEquations(t *testing.T) {
	var id = func(state.State) float64 { return 1 }
	var tests = []struct {
		eq map[state.Symbol]state.Diff
		x0 map[state.Symbol]float64
		u  map[state.Symbol]state.Input
	}{
		{eq: map[state.Symbol]state.Diff{"x": id, "y": id}, x0: map[state.Symbol]float64{"u": 1}},
		{eq: map[state.Symbol]state.Diff{"x": id}, x0: map[state.Symbol]float64{"y": 1}},
		{eq: map[state.Symbol]state.Diff{"x": id}},
		{x0: map[state.Symbol]float64{"y": 1}},
	}

	for i := range tests {
		sim := New()
		sim.SetTimespan(0, 1, 10)
		sim.SetDiffFromMap(tests[i].eq)
		sim.SetX0FromMap(tests[i].x0)
		sim.SetInputFromMap(tests[i].u)
		err := recoverSimTest(sim)
		if err == nil {
			t.Errorf("sim should have panic'd with %#v", tests[i])
		}
	}
}

func recoverSimTest(sim *Simulation) (i interface{}) {
	defer func() {
		i = recover()
	}()
	sim.Begin()
	return nil
}

// Provides a simulation with identity differential
// equations for state `x` and input `u`
func newWorkingSim() *Simulation {
	sim := New()
	sim.SetDiffFromMap(map[state.Symbol]state.Diff{
		"x": func(state.State) float64 { return 1 },
	})
	sim.SetX0FromMap(map[state.Symbol]float64{
		"x": 1,
	})
	sim.SetInputFromMap(map[state.Symbol]state.Input{
		"u": func(state.State) float64 { return 1 },
	})
	sim.SetTimespan(0, 1., 10)
	return sim
}
func TestWorkingSim(t *testing.T) {
	sim := newWorkingSim()
	err := recoverSimTest(sim)
	if err != nil {
		t.Error("other tests depend on this not failing")
	}
}

func TestResultsNotEmpty(t *testing.T) {
	// create a simulation and run it succesfully
	sim := newWorkingSim()
	sim.Begin()
	// attempt to run it again
	err := recoverSimTest(sim)
	if err == nil {
		t.Error("simulation should have prevented a run when results not empty")
	}
}

func TestNilSolver(t *testing.T) {
	sim := newWorkingSim()
	sim.Solver = nil
	err := recoverSimTest(sim)
	if err == nil {
		t.Error("simulation got a nil solver and did not panic")
	}
}
func TestBadDomainName(t *testing.T) {
	sim := newWorkingSim()
	sim.Config.Domain = ""
	err := recoverSimTest(sim)
	if err == nil {
		t.Error("simulation got a empty domain name and did not panic")
	}
}
func TestTooFewAlgorithmSteps(t *testing.T) {
	sim := newWorkingSim()
	sim.Config.Algorithm.Steps = 0
	err := recoverSimTest(sim)
	if err == nil {
		t.Error("simulation got 0 algo steps and did not panic")
	}
}

func TestSymbolNotFoundInResults(t *testing.T) {
	sim := newWorkingSim()
	sim.Begin()
	err := recoverSimResults(sim, "u")
	if err != nil {
		t.Errorf("should have been able to find result in inputs")
	}
	err = recoverSimResults(sim, "x")
	if err != nil {
		t.Errorf("should have been able to find result in state vector")
	}
	err = recoverSimResults(sim, "unknown")
	if err == nil {
		t.Error("should panic if symbol not found in results")
	}
}

func recoverSimResults(sim *Simulation, resultname state.Symbol) (i interface{}) {
	defer func() {
		i = recover()
	}()
	sim.Results(resultname)
	return nil
}

func TestBadStateDiff(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("panic expected when StateDiff called on state/function of different lengths")
		}
	}()
	F := state.Diffs{
		func(s state.State) float64 { return 1 },
	}
	s := state.New()
	StateDiff(F, s)
}

func applyFunc(sli []float64, f func(float64) float64) []float64 {
	res := make([]float64, len(sli))
	for i, v := range sli {
		res[i] = f(v)
	}
	return res
}
