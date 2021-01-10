package godesim

import "github.com/soypat/godesim/state"

var noneEvent, remEvent *Event = &Event{EventKind: EvNone}, &Event{EventKind: EvRemove}

// Event Not implemented
type Event struct {
	EventKind
	Label     string
	targets   []string
	functions []func(state.State) float64
	newDomain Timespan
}

// EventKind Enum for type of event
type EventKind uint8

// EventHandler checks if state indicates an
// ongoing or start of event. EventHandler is discarded
// after returning a non-nil event
//
// Returns nil if no event happened.
type EventHandler func(state.State) *Event

// Belongs to the enum of Event types (EventKind).
// If no comment present then the event has yet to be implemented.
const (
	// The None event has no effect on simulation.
	EvNone EventKind = iota
	// Removes EventHandler from simulation.
	EvRemove
	// Ends simulation immediately
	EvEndSimulation
	// Marks current time step with a string.
	EvMarker
	// Changes simulation state Change or Input. Must be set.
	EvBehaviour
	// Defines a new step length. Must be set.
	EvStepLength
	// Not implemented.
	EvDomainChange
	// Defines a user defined error in simulation. Event label is error.
	EvError
	// Triggers another event after a certain period. Must be set.
	EvDelay
)

// NewEvent Creates new event. After event is
// created it's effect must be set for certain EventKind.
//
// Event Label is set by first string argument and is used
// for Error and Marker kind.
func NewEvent(label string, kind EventKind) *Event {
	ev := new(Event)
	ev.EventKind, ev.Label = kind, label
	switch kind {
	case EvNone:
		return noneEvent
	case EvRemove, EvEndSimulation, EvMarker, EvError, EvBehaviour, EvStepLength:
		return ev
	case EvDomainChange:
		throwf("NewEvent: DomainChange event not implemented yet")
	case EvDelay:
		throwf("NewEvent: delayed event not implemented yet")
	default:
		throwf("NewEvent: unexpected event kind")
	}
	return ev
}

func EventNone() *Event { return noneEvent }

// SetBehaviour for EvBehaviour: Takes new state input/variable functions
func (ev *Event) SetBehaviour(m map[state.Symbol]func(state.State) float64) *Event {
	if ev.EventKind != EvBehaviour {
		throwf("Event.SetBehaviour: Event is not of kind behaviour")
	}
	ev.targets = make([]string, len(m))
	ev.functions = make([]func(state.State) float64, len(m))
	i := 0
	for k, v := range m {
		ev.targets[i] = string(k)
		ev.functions[i] = v
		i++
	}
	return ev
}

// SetDomain for EvDomainChange: Takes new Domain (timespan) for simulation.
//
// Not implemented
func (ev *Event) SetDomain(start, end float64, steps int) *Event {
	throwf("SetDomain not implemented yet")
	if ev.EventKind != EvDomainChange {
		throwf("Event.SetDomain: Event is not of kind EvDomainChange")
	}
	ev.newDomain = newTimespan(start, end, steps)
	return ev
}

// SetStepLength for EvStepLength: Set simulation steplength
func (ev *Event) SetStepLength(stepLength float64) *Event {
	if ev.EventKind != EvStepLength {
		throwf("Event.SetStepLength: Event is not of kind EvStepLength")
	}
	ev.newDomain = newTimespan(0, stepLength, 1)
	return ev
}

// Error implements Error interface for user defined simulation errors
func (ev Event) Error() string {
	if ev.EventKind != EvError {
		throwf("Event.Error(): Event is not of kind EvError")
	}
	return "EvError: " + ev.Label
}

// IdleHandler does nothing. If in a running simulation an IdleHandler will be discarded
var IdleHandler EventHandler = func(s state.State) *Event {
	return remEvent
}
