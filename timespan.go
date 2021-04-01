package godesim

// Timespan represents an iterable vector of evenly spaced time points.
// Does not store state information on steps done.
type Timespan struct {
	start      float64
	end        float64
	steps      int
	stepLength float64
}

// Len how many iterations expected for RK4
func (ts Timespan) Len() int {
	return ts.steps
}

// Dt Obtains the step length of simulation
func (ts Timespan) Dt() float64 {
	return ts.stepLength
}

// End returns greater limit of Timespan
func (ts Timespan) End() float64 {
	return ts.end
}

// SetTimespan Set time domain (step domain) for simulation.
// Step size is given by:
//   dt = (End - Start) / float64(Steps)
// since Steps is the amount of points to "solve".
func (ts *Timespan) SetTimespan(Start, End float64, Steps int) {
	(*ts) = newTimespan(Start, End, Steps)
}

const (
	// dlamchE is the machine epsilon. For IEEE this is 2^{-53}.
	dlamchE = 1.0 / (1 << 53)

	// dlamchB is the radix of the machine (the base of the number system).
	dlamchB = 2

	// dlamchP is base * eps.
	dlamchP = dlamchB * dlamchE
)

// newTimespan generates a timespan object for simulation.
// Steps must be minimum 1.
func newTimespan(Start, End float64, Steps int) Timespan {

	if Start >= End {
		throwf("Timespan: Start cannot be greater or equal to End. got %v >= %v", Start, End)
	}
	if Steps < 1 {
		throwf("Timespan: Steps must be greater or equal to 1. got %v", Steps)
	}

	dt := (End - Start) / float64(Steps)

	if dt <= 2*dlamchP {
		warnf("warning: time step %e is smaller than eps*2", dt)
	}
	return Timespan{
		start:      Start,
		end:        End,
		steps:      Steps,
		stepLength: dt,
	}
}
