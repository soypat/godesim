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
	var states, results []State
	for sim.isRunning() {
		sim.currentStep++
		states = sim.Solver(sim, state)
		state = states[len(states)-1]
		results = append(results, state)

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
