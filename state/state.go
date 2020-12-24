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

func newState(Time float64) State {
	return State{varmap: make(map[Symbol]int), time: Time}
}

// X get a state variable
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
		s.varmap[sym] = len(s.varmap) - 1
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
