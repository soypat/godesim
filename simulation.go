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
	change        map[state.Symbol]state.Diff
	inputs        map[state.Symbol]state.Input
	eventHandlers []*EventHandler
	events        []*Event
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
		Step  struct {
			Max float64 `yaml:"max"`
			Min float64 `yaml:"min"`
		} `yaml:"step"`
		Error struct {
			Max float64 `yaml:"max"`
		} `yaml:"error"`
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
		change: make(map[state.Symbol]state.Diff),
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
	sim.events = make([]*Event, 0, len(sim.eventHandlers))

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
				if ev == nil || ev.EventKind == EvNone {
					continue
				}
				sim.events = append(sim.events, ev)
				if ev.EventKind == EvRemove {
					sim.eventHandlers = append(sim.eventHandlers[:i], sim.eventHandlers[i+1:]...)
					i--
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

// SetX0FromMap sets simulation's initial X values from a Symbol map
func (sim *Simulation) SetX0FromMap(m map[state.Symbol]float64) {
	sim.State = state.New()
	for sym, v := range m {
		sim.State.XEqual(sym, v)
	}
}

// SetDiffFromMap Sets the ODE equations (change in X) with a pre-built map
//
// i.e. theta(t) = 0.5 * t^2
//
//  sim.SetDiffFromMap(map[state.Symbol]state.Change{
//  	"theta":  func(s state.State) float64 {
//  		return s.X("Dtheta")
//  	},
//  	"Dtheta": func(s state.State) float64 {
//  		return 1
//  	},
//  })
func (sim *Simulation) SetDiffFromMap(m map[state.Symbol]state.Diff) {
	sim.change = m
}

// SetInputFromMap Sets Input (U) functions with pre-built map
func (sim *Simulation) SetInputFromMap(m map[state.Symbol]state.Input) {
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

// StateDiff obtain Change results without modifying State
// Returns state evolution (result of applying Changer functions to S)
func StateDiff(F map[state.Symbol]state.Diff, S state.State) state.State {
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

// AddEventHandlers add event handlers to simulation.
func (sim *Simulation) AddEventHandlers(evhand ...EventHandler) {
	if len(evhand) == 0 {
		throwf("AddEventHandlers: can't add 0 event handlers")
	}
	if sim.eventHandlers == nil {
		sim.eventHandlers = make([]*EventHandler, 0, len(evhand))
	}
	for i := range evhand {
		sim.eventHandlers = append(sim.eventHandlers, &evhand[i])
	}
}

// Events Returns a copy of all simulation events
func (sim *Simulation) Events() []Event {
	ev := make([]Event, len(sim.events))
	for i := range sim.events {
		ev[i] = *sim.events[i]
	}
	return ev
}
