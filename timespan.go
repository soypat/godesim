package godesim

import "math"

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

// SetTimespan Set time domain (step domain) for simulation
func (ts *Timespan) SetTimespan(Start, End float64, Steps int) {
	(*ts) = newTimespan(Start, End, Steps)
}

// time returns time corresponding to step in Timespan.
func (ts Timespan) time(Step int) float64 {
	return float64(Step)*ts.stepLength + ts.start
}

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
	if dt == 0 {
		throwf("Timespan: Resulting time step is 0")
	}
	if dt <= 1e5*math.SmallestNonzeroFloat64 {
		warnf("warning: time step %e is very small", dt)
	}
	return Timespan{
		start:      Start,
		end:        End,
		steps:      Steps,
		stepLength: dt,
	}
}
