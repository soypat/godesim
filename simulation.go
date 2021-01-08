// Package godesim can be described as a simple interface
// to solve a first-order system of non-linear differential equations
// which can be defined as Go code.
package godesim

import (
	"fmt"
	"time"

	"github.com/soypat/godesim/state"
	"gonum.org/v1/gonum/floats"
)

// Simulation contains dynamics of system and stores
// simulation results.
//
// Defines an object that can solve
// a non-autonomous, non-linear system
// of differential equations
type Simulation struct {
	Timespan
	State         state.State
	currentStep   int
	results       []state.State
	Solver        func(sim *Simulation) []state.State
	change        map[state.Symbol]state.Changer
	inputs        map[state.Symbol]state.Input
	eventHandlers []*EventHandler
	Config
}

// Config modifies Simulation behaviour/output.
// Set with simulation.SetConfig method
type Config struct {
	Domain state.Symbol `yaml:"domain"`
	Log    struct {
		Results bool `yaml:"results"`
	} `yaml:"log"`
	Behaviour struct {
		StepDelay time.Duration `yaml:"delay"`
	} `yaml:"behaviour"`
	Algorithm struct {
		Steps int `yaml:"steps"`
	} `yaml:"algorithm"`
}

// New creates blank simulation
//
// Default values:
//
// Domain is the integration variable. "time" is default value
//	simulation.Domain
// Solver used is fourth order Runge-Kutta multivariable integration.
//  simulation.Solver
// How many solver steps are run between Timespan steps. Set to 1
//  simulation.Algorithm.Steps
func New() *Simulation {
	sim := Simulation{
		change: make(map[state.Symbol]state.Changer),
		Solver: RK4Solver,
	}
	sim.Domain, sim.Algorithm.Steps = "time", 1
	return &sim
}

// SetConfig Set configuration to modify default Simulation values
func (sim *Simulation) SetConfig(cfg Config) *Simulation {
	sim.Config = cfg
	return sim
}

// Begin starts simulation
//
// Unrecoverable errors will panic. Warnings may be printed.
func (sim *Simulation) Begin() {
	// This is step 0 of simulation
	sim.setInputs()
	sim.verifyPreBegin()

	sim.results = make([]state.State, 0, sim.Algorithm.Steps*sim.Len())
	sim.results = append(sim.results, sim.State)

	var states []state.State
	for sim.isRunning() {
		sim.currentStep++
		states = sim.Solver(sim)
		sim.results = append(sim.results, states[1:]...)
		sim.State = states[len(states)-1]
		sim.setInputs()
		if sim.Log.Results {
			fmt.Printf("%v\n", sim.State)
		}
		time.Sleep(sim.Behaviour.StepDelay)
		if sim.eventHandlers != nil && len(sim.eventHandlers) > 0 {
			for i := 0; i < len(sim.eventHandlers); i++ {
				handler := sim.eventHandlers[i]
				if handler == &IdleHandler { // if idler, remove and continue
					sim.eventHandlers = append(sim.eventHandlers[:i], sim.eventHandlers[i+1:]...)
					i--
					continue
				}
				ev := (*handler)(sim.State)
				if ev.EventKind == EvNone {
					continue
				}
				if ev.EventKind == EvEndSimulation {
					sim.currentStep = -1
				}
				err := sim.applyEvent(ev)
				if err != nil {
					fmt.Println("error in simulation: ", err)
				}
				sim.eventHandlers[i] = &IdleHandler
			}
		}
	}
}

// RK4Solver Integrates simulation state for next timesteps
// using 4th order Runge-Kutta multivariable algorithm
func RK4Solver(sim *Simulation) []state.State {
	const overSix = 0.166666666666666666666667
	states := make([]state.State, sim.Algorithm.Steps+1)
	dt := sim.Dt() / float64(sim.Algorithm.Steps)
	states[0] = sim.State.Clone()
	for i := 0; i < len(states)-1; i++ {
		// create auxiliary states for calculation
		b, c, d := states[i].CloneBlank(), states[i].CloneBlank(), states[i].CloneBlank()
		b.SetTime(b.Time() + dt/2)
		c.SetTime(c.Time() + dt/2)
		d.SetTime(d.Time() + dt)
		a := StateDiff(sim.change, states[i])
		aaux := a.Clone()
		b = StateDiff(sim.change, state.AddTo(b, states[i],
			state.ScaleTo(aaux, 0.5*dt, aaux)))
		baux := b.Clone()
		c = StateDiff(sim.change, state.AddTo(c, states[i],
			state.ScaleTo(baux, 0.5*dt, baux)))
		caux := c.Clone()
		d = StateDiff(sim.change, state.AddTo(d, states[i],
			state.ScaleTo(caux, dt, caux)))
		state.Add(a, d)
		state.Add(b, c)
		state.AddScaled(a, 2, b)
		states[i+1] = states[i].Clone()
		state.AddScaled(states[i+1], dt*overSix, a)
		states[i+1].SetTime(dt + states[i].Time())
	}
	return states
}

// SetX0FromMap sets simulation's initial X values from a Symbol map
func (sim *Simulation) SetX0FromMap(m map[state.Symbol]float64) {
	sim.State = state.New()
	for sym, v := range m {
		sim.State.XEqual(sym, v)
	}
}

// SetChangeMap Sets the ODE equations (change in X) with a pre-built map
//
// i.e. theta(t) = 0.5 * t^2
//
//  sim.SetChangeMap(map[state.Symbol]state.Changer{
//  	"theta":  func(s state.State) float64 {
//  		return s.X("Dtheta")
//  	},
//  	"Dtheta": func(s state.State) float64 {
//  		return 1
//  	},
//  })
func (sim *Simulation) SetChangeMap(m map[state.Symbol]state.Changer) {
	sim.change = m
}

// SetInputMap Sets Input (U) functions with pre-built map
func (sim *Simulation) SetInputMap(m map[state.Symbol]state.Input) {
	sim.inputs = m
}

// CurrentTime obtain simulation step variable
func (sim *Simulation) CurrentTime() float64 {
	return sim.results[len(sim.results)-1].Time()
}

// Results get vector of simulation results for given symbol (X or U)
//
// Special case is the Simulation.Domain (default "time") symbol.
func (sim *Simulation) Results(sym state.Symbol) []float64 {
	vec := make([]float64, len(sim.results))
	// TODO verify simulation has run!
	if sym == sim.Domain {
		for i, r := range sim.results {
			vec[i] = r.Time()
		}
		return vec
	}
	symV := []state.Symbol{sym}
	consU, consX := !floats.HasNaN(sim.State.ConsistencyU(symV)), !floats.HasNaN(sim.State.ConsistencyX(symV))
	if consU {
		for i, r := range sim.results {
			vec[i] = r.U(sym)
		}
	}
	if consX {
		for i, r := range sim.results {
			vec[i] = r.X(sym)
		}
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

// AddEvents add event handlers to simulation.
func (sim *Simulation) AddEvents(evhand ...EventHandler) {
	if len(evhand) == 0 {
		throwf("AddEvents: can't have 0 event handlers")
	}
	if sim.eventHandlers == nil {
		sim.eventHandlers = make([]*EventHandler, 0, len(evhand))
	}
	for i := range evhand {
		sim.eventHandlers = append(sim.eventHandlers, &evhand[i])
	}

}
