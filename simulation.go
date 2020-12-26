package simulation

import (
	"fmt"
	"time"

	"github.com/go-sim/simulation/state"
)

var overSix float64 = 0.166666666666666666666667

// Simulation ...
type Simulation struct {
	x0 state.State
	Timespan
	currentTime float64
	currentStep int
	SolverSteps int
	results     []state.State
	Solver      func(sim *Simulation, s state.State) []state.State
	Change      map[state.Symbol]state.Changer
	config      struct {
		printResults bool
		rk4delay     time.Duration
	}
}

// New creates blank simulation
func New() *Simulation {
	sim := Simulation{
		Change:      make(map[state.Symbol]state.Changer),
		Solver:      RK4Solver,
		SolverSteps: 1,
	}
	return &sim
}

// Begin starts simulation
func (sim *Simulation) Begin() {
	// This is step 0 of simulation
	st := sim.x0

	var states []state.State
	sim.results = make([]state.State, 0, sim.SolverSteps*sim.Len())
	sim.results = append(sim.results, st)
	for sim.isRunning() {
		sim.currentStep++
		states = sim.Solver(sim, st)
		sim.results = append(sim.results, states[1:]...)
		st = states[len(states)-1]
		if sim.config.printResults {
			fmt.Printf("%v\n", st)
		}
		time.Sleep(sim.config.rk4delay)
	}
}

// will contain heavy logic in future. event oriented stuff to come
func (sim *Simulation) isRunning() bool {
	return sim.CurrentStep() < sim.Len()
}

// RK4Solver Integrates simulation state for next timesteps
// using 4th order Runge-Kutta multivariable algorithm
func RK4Solver(sim *Simulation, s state.State) []state.State {
	states := make([]state.State, sim.SolverSteps+1)
	dt := sim.Dt()
	states[0] = s
	for i := 0; i < len(states)-1; i++ {
		// create auxiliary states for calculation
		b, c, d := states[i].CloneBlank(), states[i].CloneBlank(), states[i].CloneBlank()
		a := StateDiff(sim.Change, states[i])
		aaux := a.Clone()
		b = StateDiff(sim.Change, state.AddTo(b, states[i],
			state.ScaleTo(aaux, 0.5*dt, aaux)))
		baux := b.Clone()
		c = StateDiff(sim.Change, state.AddTo(c, states[i],
			state.ScaleTo(baux, 0.5*dt, baux)))
		caux := c.Clone()
		d = StateDiff(sim.Change, state.AddTo(d, states[i],
			state.ScaleTo(caux, dt, caux)))
		state.Add(a, d)
		state.Add(b, c)
		state.AddScaled(a, 2, b)
		states[i+1] = states[i].Clone()
		state.AddScaled(states[i+1], dt*overSix, a)
	}
	return states
}

// SetX0FromMap sets simulation's initial X values from a Symbol map
func (sim *Simulation) SetX0FromMap(m map[state.Symbol]float64) {
	sim.x0 = state.New()
	for sym, v := range m {
		sim.x0.XEqual(sym, v)
	}
}

// SetChangeMap Sets the ODE equations with a pre built map
//
// i.e.
//
//  sim.SetChangeMap(map[simulation.Symbol]simulation.StateChanger{
//  	"theta":  func(s simulation.State) float64 {
//  		return s.X("Dtheta")
//  	},
//  	"Dtheta": func(s simulation.State) float64 {
//  		return 1
//  	},
//  })
func (sim *Simulation) SetChangeMap(m map[state.Symbol]state.Changer) {
	sim.Change = m
}

// SetTimespan Set time domain (step domain) for simulation
func (sim *Simulation) SetTimespan(Start, End float64, Steps int) {
	sim.Timespan = NewTimespan(Start, End, Steps)
}

// CurrentStep get number of steps done.
// Reaches maximum Simulation's Timespan's `Steps`
func (sim *Simulation) CurrentStep() int {
	return sim.currentStep
}

// LastTime Obtains last Simulation step time.
// Does not take into account Solver's steps
func (sim *Simulation) LastTime() float64 {
	return sim.stepLength * float64(sim.CurrentStep())
}

// XResults get numerical slice of simulation results for given symbol
func (sim *Simulation) XResults(sym state.Symbol) []float64 {
	vec := make([]float64, len(sim.results))
	// TODO verify simulation has run!
	sim.results[0].X(sym) // Check if variable exists
	for i, r := range sim.results {
		vec[i] = r.X(sym)
	}
	return vec
}

// StateDiff obtain StateChanger results without modifying State
// Returns state evolution (result of applying Changer functions to S)
func StateDiff(F map[state.Symbol]state.Changer, S state.State) state.State {
	diff := S.Clone()
	syms := S.XSymbols()
	if len(F) != len(syms) {
		throwf("length of func slice not equal to float slice (%v vs. %v)", len(F), len(syms))
	}
	for i := 0; i < len(F); i++ {
		diff.XEqual(syms[i], F[syms[i]](S))
	}
	return diff
}
