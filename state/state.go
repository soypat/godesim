package state

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

// Symbol is used to reference a simulation variable.
//
// It should be unique for each simulation
type Symbol string

// State describes a discrete simulation configuration.
//
// Contains X and U vectors for a step instance (i.e. point in time)
// Can access step variable with Time() method.
type State struct {
	varmap     map[Symbol]int
	x          []float64
	inputmap   map[Symbol]int
	u          []float64
	time       float64
	transposed bool
}

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

// Vector implementation

// Len returns amount of X variables in state
func (s State) Len() int {
	return len(s.x)
}

// AtVec gets X variable at position i
func (s State) AtVec(i int) float64 {
	return s.x[i]
}

// Matrix Implementation

// Dims returns X vector dimensions.
// States are column vectors by default
func (s State) Dims() (r, c int) {
	if s.transposed {
		return 1, s.Len()
	}
	return s.Len(), 1 // default status
}

// At returns X value at row i and column j
func (s State) At(i, j int) float64 {
	if s.transposed {
		return s.x[j]
	}
	return s.x[i]
}

// T transposes matrix and returns the State transposed (not a copy)
func (s State) T() mat.Matrix {
	s.transposed = !s.transposed
	return s
}

// Matrix is the basic matrix interface type.
type Matrix interface {
	// Dims returns the dimensions of a Matrix.
	Dims() (r, c int)

	// At returns the value of a matrix element at row i, column j.
	// It will panic if i or j are out of bounds for the matrix.
	At(i, j int) float64

	// T returns the transpose of the Matrix. Whether T returns a copy of the
	// underlying data is implementation dependent.
	// This method may be implemented using the Transpose type, which
	// provides an implicit matrix transpose.
	T() Matrix
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

// CloneBlank creates a duplicate of state at time `t`
// with all X vector set to zero value
func (s State) CloneBlank(t float64) State {
	return State{
		varmap:   s.varmap,
		x:        make([]float64, len(s.x)),
		inputmap: s.inputmap,
		u:        s.UVector(),
		time:     t,
	}
}

// XVector returns copy of state X vector
func (s State) XVector() []float64 {
	if len(s.x) == 0 {
		return make([]float64, 0)
	}
	cp := make([]float64, len(s.x))
	copy(cp, s.x)
	return cp
}

// UVector returns copy of state U vector
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
// of zero floats
//
// If a symbol is not present in U
// then an IEEE 754 “not-a-number” value will correspond to it.
func (s State) ConsistencyU(question []Symbol) []float64 {
	result := make([]float64, len(question))
	for i, sym := range question {
		if !s.has("U", sym) {
			result[i] = math.NaN()
		}
	}
	return result
}

// ConsistencyX can be used to determine if Symbols are present in X
// and if a symbol is missing.
//
// It takes a vector of Symbols and returns a vector
// of zero floats
//
// If a symbol is not present in X
// then an IEEE 754 “not-a-number” value will correspond to it.
func (s State) ConsistencyX(question []Symbol) []float64 {
	result := make([]float64, len(question))
	for i, sym := range question {
		if !s.has("X", sym) {
			result[i] = math.NaN()
		}

	}
	return result
}

func (s *State) SetAllX(new []float64) {
	if len(s.varmap) != len(new) {
		throwf("new slice length %d does not coincide with state X symbol length %d", len(new), len(s.varmap))
	}
	s.x = new
}
