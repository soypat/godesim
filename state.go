package simulation

// State ...
type State struct {
	varmap   map[Symbol]int
	x        []float64
	inputmap map[Symbol]int
	u        []float64
	time     float64
}

// StateChanger represents ODE change of a simulation X variable
type StateChanger func(State) float64

// Timespan ...
type Timespan struct {
	start      float64
	end        float64
	steps      int
	stepLength float64
	counter    int
	onNext     func()
}

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

// Len how many iterations expected for RK4
func (ts Timespan) Len() int {
	return ts.steps
}

// Dt Obtains the step length of simulation
func (ts Timespan) Dt() float64 {
	return ts.stepLength
}

// TimeVector is the ordered set of all Simulation time points
func (ts Timespan) TimeVector() []float64 {
	vec := make([]float64, ts.Len()+1)
	for i := 0; i < ts.Len()+1; i++ {
		vec[i] = float64(i)*ts.stepLength + ts.start
	}
	return vec
}

// NewTimespan generates a timespan object for simulation
// Steps must be minimum 1.
func NewTimespan(Start, End float64, Steps int) Timespan {
	nxt := func() {}
	if Start >= End {
		throwf("Timespan start cannot be greater or equal to Timespan end. got %v >= %v", Start, End)
	}
	if Steps < 1 {
		throwf("steps in Timespan must be greater or equal to 1. got %v", Steps)
	}
	return Timespan{
		start:      Start,
		end:        End,
		steps:      Steps,
		stepLength: (End - Start) / float64(Steps),
		onNext:     nxt,
	}
}

// ARITHMETIC

func (s State) Add(s State)
