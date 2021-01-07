package godesim

import "github.com/soypat/godesim/state"

// Event Not implemented
type Event struct {
	EventKind
	targets   []string
	functions []func(state.State) float64
	newDomain Timespan
}

// EventKind Enum for type of event
type EventKind int

// EventHandler takes a state and returns events if any happened
type EventHandler func(state.State) *Event

// Belongs to the enum of Event types (EventKind)
const (
	// The None event has no effect on simulation
	EvNone EventKind = iota
	EvEndSimulation
	EvBehaviour
	EvMarker
	EvStepLength
	EvDomainChange
	EvError
	// Not implemented
	EvDelay
)

// NewEvent Creates new event
func NewEvent(kind EventKind) *Event {
	ev := new(Event)
	ev.EventKind = kind
	switch kind {
	case EvNone:
		return ev
	case EvBehaviour:
		ev.targets = make([]string, 0)
		ev.functions = make([]func(state.State) float64, 0)
	case EvError:
		ev.targets = make([]string, 0, 1)
	case EvMarker:
		// pass
	case EvEndSimulation:
		// pass
	case EvDomainChange, EvStepLength:
		// pass
	case EvDelay:
		throwf("NewEvent: delayed event not implemented yet")
	default:
		throwf("NewEvent: unexpected event kind")
	}
	return ev
}

// SetBehaviour for EvBehaviour: Takes new state input/variable functions
func (ev *Event) SetBehaviour(m map[state.Symbol]func(state.State) float64) {
	if ev.EventKind != EvBehaviour {
		throwf("Event.SetBehaviour: Event is not of type behaviour")
	}
	for k, v := range m {
		ev.targets = append(ev.targets, string(k))
		ev.functions = append(ev.functions, v)
	}
}

// SetDomain for EvDomainChange: Takes new Domain (timespan) for simulation
func (ev *Event) SetDomain(start, end float64, steps int) {
	throwf("SetDomain not implemented yet")
	ev.newDomain = newTimespan(start, end, steps)
}

// SetStepLength for EvStepLength: Set simulation steplength
func (ev *Event) SetStepLength(stepLength float64) {
	ev.newDomain = newTimespan(0, stepLength, 1)
}

// Error implements Error interface for user defined simulation errors
func (ev Event) Error() string {
	return ev.targets[0]
}

// IdleHandler does nothing. If in a running simulation an IdleHandler will be discarded
var IdleHandler EventHandler = func(s state.State) *Event {
	return NewEvent(EvNone)
}
