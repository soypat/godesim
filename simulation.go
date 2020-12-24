package simulation

import (
	"fmt"
	"time"

	"github.com/go-sim/simulation/state"
)

// Simulation ...
type Simulation struct {
	x0 state.State
	Timespan
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
		Timespan:    NewTimespan(0, 1, 10),
		Solver:      RK4Solver,
		SolverSteps: 1,
	}
	return &sim
}

// Begin starts simulation
func (sim *Simulation) Begin() {
	// This is step 0 of simulation
	state := sim.x0
	var states []state.State
	sim.results = make([]state.State, 0, sim.SolverSteps*sim.Len())
	sim.results = append(sim.results, state)
	for sim.isRunning() {
		sim.currentStep++
		states = sim.Solver(sim, state)
		sim.results = append(sim.results, states[1:]...)
		state = states[len(states)-1]
		if sim.config.printResults {
			fmt.Printf("%v\n", state)
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
	// dt := sim.Dt()
	states[0] = s
	// t := sim.LastTime()
	// syms := states[0].XSymbols()
	for i := 0; i < len(states)-1; i++ {
		a := ApplyFuncs(sim.Change, states[i])
		b := ApplyFuncs(sim.Change, states[i])
		print(a, b)
		// RK4 integration scheme
		// a := ApplyFuncs(s.Change, X)

		// states[i+1] = nextState
	}
	return states
}

// SetX0FromMap sets simulation's initial X values from a Symbol map
func (sim *Simulation) SetX0FromMap(m map[state.Symbol]float64) {
	sim.x0.varmap = m
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
	res := make([]float64, len(sim.results))

	if _, ok := sim.results[0].varmap[sym]; !ok {
		throwf("%v Symbol not in state", sym)
	}
	for i, r := range sim.results {
		res[i] = r.x[r.varmap[sym]]
	}
	return res
}

// ApplyFuncs obtain StateChanger results without modifying State
// Returns an ordered float slice according to State.XSymbols()
func ApplyFuncs(F map[state.Symbol]state.Changer, S state.State) []float64 {
	syms := S.XSymbols()
	if len(F) != len(syms) {
		throwf("length of func slice not equal to float slice (%v vs. %v)", len(F), len(syms))
	}
	dst := make([]float64, len(F))
	for i := 0; i < len(F); i++ {
		dst[i] = F[syms[i]](S)
	}
	return dst
}
