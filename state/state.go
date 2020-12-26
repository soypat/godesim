package state

// Symbol is used to reference a simulation variable. It should be unique for each simulation
type Symbol string

// State ...
type State struct {
	varmap   map[Symbol]int
	x        []float64
	inputmap map[Symbol]int
	u        []float64
	time     float64
}

// Changer represents ODE change of a simulation X variable
type Changer func(State) float64

// New creates new empty state
func New() State {
	s := State{varmap: make(map[Symbol]int)}
	s.x, s.u = make([]float64, 0), make([]float64, 0)
	return s
}

// NewFromXMap Creates new state from a X var symbol to value map
func NewFromXMap(xm map[Symbol]float64) State {
	s := New()
	s.x = make([]float64, len(xm))
	for sym, v := range xm {
		s.XEqual(sym, v)
	}
	return s
}

// // NewFromTime creates new state at a given time
// func New(Time float64) State {
// 	s := New()
// 	s.time = Time
// 	return s
// }

// X get a state variable. if state variable not exist
// then function panics
func (s State) X(sym Symbol) float64 {
	idx, ok := s.varmap[sym]
	if !ok {
		throwf("%v Symbol does not exist in State variables", sym)
	}
	return s.x[idx]
}

// XEqual Set a State Symbol to a value.
// If Symbol does not exist then it is created
func (s *State) XEqual(sym Symbol, val float64) {
	s.xCreateIfNotExist(sym)
	s.x[s.varmap[sym]] = val
}

func (s *State) xCreateIfNotExist(sym Symbol) {
	if _, ok := s.varmap[sym]; !ok {
		s.x = append(s.x, 0)
		s.varmap[sym] = len(s.x) - 1
	}
}

// XSet sets an existing State Symbol to a value
func (s *State) XSet(sym Symbol, val float64) {
	if _, ok := s.varmap[sym]; !ok {
		throwf("%v Symbol does not exist in State variables", sym)
	}
	s.XEqual(sym, val)
}

// Clone makes a duplicate of a State.
func (s State) Clone() State {
	return State{
		varmap:   s.varmap,
		x:        s.XFloats(),
		inputmap: s.inputmap,
		u:        s.UFloats(),
		time:     s.time,
	}
}

// CloneBlank creates a duplicate of state
// with all X and U vectors set to zero value
func (s State) CloneBlank() State {
	return State{
		varmap:   s.varmap,
		x:        make([]float64, len(s.x)),
		inputmap: s.inputmap,
		u:        make([]float64, len(s.u)),
		time:     s.time,
	}
}

// XFloats returns state X vector
func (s State) XFloats() []float64 {
	cp := make([]float64, len(s.x))
	copy(cp, s.x)
	return cp
}

// UFloats returns state U vector
func (s State) UFloats() []float64 {
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

// ARITHMETIC

// func (s State) Add(s State)
