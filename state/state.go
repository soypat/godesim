package state

import "math"

// Symbol is used to reference a simulation variable.
//
// It should be unique for each simulation
type Symbol string

// State describes a discrete simulation configuration.
//
// Contains X and U vectors for a step instance (i.e. point in time)
// Can access step variable with Time() method.
type State struct {
	varmap   map[Symbol]int
	x        []float64
	inputmap map[Symbol]int
	u        []float64
	time     float64
}

// Changer represents a single non-linear differential equation
// of the Simulation's system of differential equation. Specifically, it
// represents an X variable change.
type Changer func(State) float64

// Input represents a time-varying or table look-up variable/coefficient
// of the Simulation's system of differential equations. Inputs can be used to
// model non-autonomous system of differential equations. Input functions are
// called after solver algorithm finishes on the resulting state.
type Input func(State) float64

// New creates empty state
func New() State {
	s := State{varmap: make(map[Symbol]int), inputmap: make(map[Symbol]int)}
	// s.x, s.u = make([]float64, 0), make([]float64, 0)
	return s
}

// NewFromXMap Creates state from a X var symbol to value map
func NewFromXMap(xm map[Symbol]float64) State {
	s := New()
	s.x = make([]float64, len(xm))
	for sym, v := range xm {
		s.XEqual(sym, v)
	}
	return s
}

// X get a State variable.
//
// If state variable does not exist then X panics
func (s State) X(sym Symbol) float64 {
	idx, ok := s.varmap[sym]
	if !ok {
		throwf("%v Symbol does not exist in State variables", sym)
	}
	return s.x[idx]
}

// U get a State input.
//
// If state input does not exist then U panics
func (s State) U(sym Symbol) float64 {
	idx, ok := s.inputmap[sym]
	if !ok {
		throwf("%v Symbol does not exist in State inputs", sym)
	}
	return s.u[idx]
}

// Time get State step variable (default time)
func (s State) Time() float64 {
	return s.time
}

// SetTime set domain variable (default time)
func (s *State) SetTime(t float64) {
	s.time = t
}

// XEqual Set a State Symbol to a value.
//
// If Symbol does not exist then it is created
func (s *State) XEqual(sym Symbol, val float64) {
	s.xCreateIfNotExist(sym)
	s.x[s.varmap[sym]] = val
}

// XSet set an existing State Symbol to a value
//
// If Symbol does not exist then XSet panics
func (s *State) XSet(sym Symbol, val float64) {
	if _, ok := s.varmap[sym]; !ok {
		throwf("%v Symbol does not exist in State variables", sym)
	}
	s.XEqual(sym, val)
}

// UEqual Set a State Input (U) Symbol to a value.
//
// If Symbol does not exist then it is created
func (s *State) UEqual(sym Symbol, val float64) {
	s.uCreateIfNotExist(sym)
	s.u[s.inputmap[sym]] = val
}

// USet sets an existing State Symbol to a value.
//
// If Symbol does not exist then USet panics
func (s *State) USet(sym Symbol, val float64) {
	if _, ok := s.inputmap[sym]; !ok {
		throwf("%v Symbol does not exist in State inputs", sym)
	}
	s.UEqual(sym, val)
}

// Clone creates a duplicate of a State.
func (s State) Clone() State {
	return State{
		varmap:   s.varmap,
		x:        s.XVector(),
		inputmap: s.inputmap,
		u:        s.UVector(),
		time:     s.time,
	}
}

// CloneBlank creates a duplicate of state
// with all X vector set to zero value
func (s State) CloneBlank() State {
	return State{
		varmap:   s.varmap,
		x:        make([]float64, len(s.x)),
		inputmap: s.inputmap,
		u:        s.UVector(),
		time:     s.time,
	}
}

// XVector returns state X vector
func (s State) XVector() []float64 {
	if len(s.x) == 0 {
		return make([]float64, 0)
	}
	cp := make([]float64, len(s.x))
	copy(cp, s.x)
	return cp
}

// UVector returns state U vector
func (s State) UVector() []float64 {
	if len(s.u) == 0 {
		return make([]float64, 0)
	}
	cp := make([]float64, len(s.u))
	copy(cp, s.u)
	return cp
}

// XSymbols returns ordered state Symbol slice
func (s State) XSymbols() []Symbol {
	syms := make([]Symbol, len(s.varmap))
	for sym, idx := range s.varmap {
		syms[idx] = sym
	}
	return syms
}

// USymbols returns ordered input Symbol slice
func (s State) USymbols() []Symbol {
	syms := make([]Symbol, len(s.inputmap))
	for sym, idx := range s.inputmap {
		syms[idx] = sym
	}
	return syms
}

// ConsistencyU can be used to determine if Symbols are present in U
// and if a symbol is missing.
//
// It takes a vector of Symbols and returns a vector
// of floats with each number corresponding to Symbol U state
//
// If a symbol is not present in U
// then an IEEE 754 “not-a-number” value will correspond to it.
func (s State) ConsistencyU(question []Symbol) []float64 {
	result := make([]float64, len(question))
	for i, sym := range question {

		if s.has("U", sym) {
			result[i] = s.U(sym)
			continue
		}
		result[i] = math.NaN()
	}
	return result
}

// ConsistencyX can be used to determine if Symbols are present in X
// and if a symbol is missing.
//
// It takes a vector of Symbols and returns a vector
// of floats with each number corresponding to Symbol X state
//
// If a symbol is not present in X
// then an IEEE 754 “not-a-number” value will correspond to it.
func (s State) ConsistencyX(question []Symbol) []float64 {
	result := make([]float64, len(question))
	for i, sym := range question {
		if s.has("X", sym) {
			result[i] = s.X(sym)
			continue
		}
		result[i] = math.NaN()
	}
	return result
}
