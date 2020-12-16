package simulation

// State ...
type State struct {
	variables map[Symbol]float64
	inputs    map[Symbol]float64
	time      float64
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
	return State{variables: make(map[Symbol]float64), time: Time}
}

// X get a state variable
func (s State) X(sym Symbol) (val float64) {
	var ok bool
	if val, ok = s.variables[sym]; !ok {
		throwf("%v Symbol does not exist in State variables", sym)
	}
	return val
}

// XSet sets an existing State Symbol to a value
func (s *State) XSet(sym Symbol, val float64) {
	if _, ok := s.variables[sym]; !ok {
		throwf("%v Symbol does not exist in State variables", sym)
	}
	s.variables[sym] = val
}

// XEqual Set a State Symbol to a value.
// If Symbol does not exist then it is created
func (s *State) XEqual(sym Symbol, val float64) {
	s.variables[sym] = val
}

// XAdd adds a value to an existing symbol and returns
// a copy of state with value added
func (s State) XAdd(sym Symbol, val float64) State {
	if _, ok := s.variables[sym]; !ok {
		throwf("%v Symbol does not exist in State variables", sym)
	}
	s.variables[sym] += val
	return s
}

// Len how many iterations expected for RK4
func (ts Timespan) Len() int {
	return ts.steps
}

// Dt Obtains the step length of simulation
func (ts Timespan) Dt() float64 {
	return ts.stepLength
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
