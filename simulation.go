// Package godesim can be described as a simple interface
// to solve a first-order system of non-linear differential equations
// which can be defined as Go code.
package godesim

import (
	"os"
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
	Logger      Logger
	State       state.State
	currentStep int
	results     []state.State
	Solver      func(sim *Simulation) []state.State
	change      map[state.Symbol]state.Diff
	Diffs       state.Diffs
	inputs      map[state.Symbol]state.Input
	eventers    []Eventer
	events      []struct {
		Label string
		State state.State
	}
	Config
}

// Config modifies Simulation behaviour/output.
// Set with simulation.SetConfig method
type Config struct {
	// Domain is symbol name for step variable. Default is `time`
	Domain    state.Symbol `yaml:"domain"`
	Log       LoggerOptions
	Behaviour struct {
		StepDelay time.Duration `yaml:"delay"`
	} `yaml:"behaviour"`
	Algorithm struct {
		// Number of algorithm steps. Different to simulation Timespan.Len()
		Steps int `yaml:"steps"`
		// Step limits for adaptive algorithms
		Step struct {
			Max float64 `yaml:"max"`
			Min float64 `yaml:"min"`
		} `yaml:"step"`
		Error struct {
			// Sets max error before proceeding with adaptive iteration
			// Step.Min should override this
			Max float64 `yaml:"max"`
		} `yaml:"error"`
	} `yaml:"algorithm"`
	Symbols struct {
		// Sorts symbols for consistent logging and testing
		NoOrdering bool `yaml:"no_ordering"`
	} `yaml:"symbols"`
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
		Logger: newLogger(os.Stdout),
	}
	sim.Config = DefaultConfig()
	return &sim
}

// SetConfig Set configuration to modify default Simulation values
func (sim *Simulation) SetConfig(cfg Config) *Simulation {
	sim.Config = cfg
	return sim
}

// DefaultConfig returns configuration set for all new
// simulations by New()
func DefaultConfig() Config {
	cfg := Config{Domain: "time"}
	cfg.Log.Results.Precision = -1
	cfg.Algorithm.Steps = 1
	return cfg
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
	sim.events = make([]struct {
		Label string
		State state.State
	}, 0, len(sim.eventers))

	eventsOn := sim.eventers != nil && len(sim.eventers) > 0
	logging := sim.Log.Results.FormatLen > 0
	if logging {
		sim.logStates(sim.results[:1])
	}
	var states []state.State
	for sim.IsRunning() {
		sim.currentStep++
		states = sim.Solver(sim)
		sim.results = append(sim.results, states[1:]...)
		sim.State = states[len(states)-1]
		sim.setInputs()
		if logging {
			sim.logStates(states[1:])
		}
		time.Sleep(sim.Behaviour.StepDelay)
		if eventsOn {
			sim.handleEvents()
		}
	}
	if logging {
		sim.Logger.flush()
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
		return vec
	}
	if consX {
		for i, r := range sim.results {
			vec[i] = r.X(sym)
		}
		return vec
	}
	throwf("Simulation.Results: %s not found in X or U symbols", sym)
	return nil
}

// StateDiff obtain Change results without modifying State
// Returns state evolution (result of applying Changer functions to S)
func StateDiff(F state.Diffs, S state.State) state.State {
	diff := S.Clone()
	syms := S.XSymbols()
	if len(F) != len(syms) {
		throwf("length of func slice not equal to float slice (%v vs. %v)", len(F), len(syms))
	}
	for i := 0; i < len(F); i++ {
		diff.XEqual(syms[i], F[i](S))
	}
	return diff
}

// AddEventHandlers add event handlers to simulation.
//
// Events which return errors will stop the simulation and panic
func (sim *Simulation) AddEventHandlers(evhand ...Eventer) {
	if len(evhand) == 0 {
		throwf("AddEventHandlers: can't add 0 event handlers")
	}
	if sim.eventers == nil {
		sim.eventers = make([]Eventer, 0, len(evhand))
	}
	for i := range evhand {
		sim.eventers = append(sim.eventers, evhand[i])
	}
}

// Events Returns a copy of all simulation events
func (sim *Simulation) Events() []struct {
	Label string
	State state.State
} {
	return sim.events
}
