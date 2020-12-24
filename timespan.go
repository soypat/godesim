package simulation

// Timespan ...
type Timespan struct {
	start      float64
	end        float64
	steps      int
	stepLength float64
	counter    int
	onNext     func()
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
