package godesim

import (
	"fmt"
	"math"

	"github.com/soypat/godesim/state"
)

// Eventer specifies an event to be applied to simulation
//
// Events which return error when applied to simulation will cause the simulation
// to panic and stop execution.
type Eventer interface {
	// Behaviour aspect of event. changing simulation
	//
	// A nil func(*Simulation) error corresponds to no event taking place (just skips it)
	Event(state.State) func(*Simulation) error
	// For identification
	Label() string
}

var (
	// ErrorRemove should be returned by a simulation modifier
	// if the Eventer is to be removed from Event handlers
	ErrorRemove error = fmt.Errorf("remove this Eventer")
)

// EventDone is the uneventful event. Changes absolutely nothing
// and is removed from Event handler list.
func EventDone(sim *Simulation) error { return nil }

// DiffChangeFromMap Event handler. Takes new state variable (X) functions and applies them
func DiffChangeFromMap(newDiff map[state.Symbol]func(state.State) float64) func(*Simulation) error {
	return func(sim *Simulation) error {
		applied := 0
		for i, sym := range sim.State.XSymbols() {
			if _, ok := newDiff[sym]; ok {
				sim.Diffs[i] = newDiff[sym]
				applied++
			}
		}
		if applied != len(newDiff) {
			return fmt.Errorf("%d symbol(s) were not found during DiffChange event", len(newDiff)-applied)
		}
		return nil
	}
}

// NewStepLength Event handler. Sets the new minimum step length
func NewStepLength(h float64) func(*Simulation) error {
	return func(sim *Simulation) error {
		if sim.IsRunning() {
			steps := math.Ceil((sim.End() - sim.CurrentTime()) / h)

			sim.SetTimespan(sim.CurrentTime(), sim.CurrentTime()+steps*h, int(steps))
		}
		return nil
	}
}

// EndSimulation Event handler. Ends simulation
func EndSimulation(sim *Simulation) error {
	sim.currentStep = -1
	return nil
}
