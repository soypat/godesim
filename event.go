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
type EventHandler func(state.State) Event

// Belongs to the enum of Event types (EventKind)
const (
	EvUndefined EventKind = iota
	EvBehaviour
	EvMarker
	EvDomainChange
	// Not implemented
	EvDelay
)

// NewEvent Creates new event
func NewEvent(kind EventKind) *Event {
	ev := new(Event)
	ev.EventKind = kind
	switch kind {
	case EvBehaviour:
		ev.targets = make([]string, 0)
		ev.functions = make([]func(state.State) float64, 0)
	case EvMarker:
		// pass
	case EvDomainChange:
		// pass
	case EvDelay:
		throwf("Delayed event not implemented yet")
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
	ev.newDomain = newTimespan(start, end, steps)
}
