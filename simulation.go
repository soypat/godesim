package simulation

import (
	"fmt"
	"time"
)

// Symbol is used to reference a simulation variable. It should be unique for each simulation
type Symbol string

// Simulation ...
type Simulation struct {
	x0 State
	Timespan
	currentStep int
	SolverSteps int
	results     []State
	Solver      func(sim *Simulation, s State) []State
	Change      map[Symbol]StateChanger
	config      struct {
		printResults bool
		rk4delay     time.Duration
	}
}

// New creates blank simulation
func New() *Simulation {
	sim := Simulation{
		Change:      make(map[Symbol]StateChanger),
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
	var states []State
	sim.results = make([]State, 0, sim.SolverSteps*sim.Len())
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
func RK4Solver(sim *Simulation, s State) []State {
	states := make([]State, sim.SolverSteps+1)
	dt := sim.Dt()
	states[0] = s
	t := sim.LastTime()
	for i := 0; i < len(states)-1; i++ {
		state := states[i]
		nextState := newState(t)
		// RK4 integration scheme
		var integrator State = newState(t)
		for i := 0; i < 4; i++ {
			for sym, change := range sim.Change {
				switch i {
				case 0:
					integrator.XEqual(sym, change(state))
				case 1, 2:
					integrator.XEqual(sym, change(state)+integrator.X(sym)*dt/2)
				case 3:
					integrator.XEqual(sym, change(state)+integrator.X(sym)*dt)
				}
			}
		}

		for sym, change := range sim.Change {
			a := change(state)
			b := change(state.XAdd(sym, dt/2*a))
			c := change(state.XAdd(sym, dt/2*b))
			d := change(state.XAdd(sym, dt*c))
			nextState.XEqual(sym, state.X(sym)+dt/6*(a+2*(b+c)+d))
		}
		states[i+1] = nextState
	}
	return states
}

// SetX0FromMap sets simulation's initial X values from a Symbol map
func (sim *Simulation) SetX0FromMap(m map[Symbol]float64) {
	sim.x0.variables = m
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
func (sim *Simulation) SetChangeMap(m map[Symbol]StateChanger) {
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
func (sim *Simulation) XResults(sym Symbol) []float64 {
	res := make([]float64, len(sim.results))

	if _, ok := sim.results[0].variables[sym]; !ok {
		throwf("%v Symbol not in state", sym)
	}
	for i, r := range sim.results {
		res[i] = r.variables[sym]
	}
	return res
}
