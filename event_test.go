package godesim_test

import (
	"testing"

	"github.com/soypat/godesim"
	"github.com/soypat/godesim/state"
)

func TestStepLen(t *testing.T) {
	Dtheta := func(s state.State) float64 {
		return s.U("u")
	}

	inputVar := func(s state.State) float64 {
		return 1
	}
	sim := godesim.New()
	sim.SetChangeMap(map[state.Symbol]state.Changer{
		"theta": Dtheta,
	})
	sim.SetX0FromMap(map[state.Symbol]float64{
		"theta": 0,
	})
	sim.SetInputMap(map[state.Symbol]state.Input{
		"u": inputVar,
	})
	const N_steps, ti, tf = 10, 0.0, 1.0

	sim.SetTimespan(ti, tf, N_steps)
	initStepLen := sim.Dt()
	newStepLen := initStepLen * 0.25
	tswitch := 0.5
	sim.AddEvents(func(s state.State) *godesim.Event {
		if s.Time() >= tswitch {
			ev := godesim.NewEvent(godesim.EvStepLength)
			ev.SetStepLength(newStepLen)
			return ev
		}
		return godesim.NewEvent(0)
	})
	sim.Begin()

	time, x_res := sim.Results("time"), sim.Results("theta")
	if len(time) != len(x_res) {
		t.Error("length of time and theta vectors should be the same")
	}
	expectedLen := initStepLen
	for i, tm := range time[:len(time)-2] {
		if tm >= tswitch {
			expectedLen = newStepLen
		}
		StepLen := time[i+1] - tm
		if StepLen != expectedLen {
			t.Errorf("expected stepLength %.4f. Got %.4f", expectedLen, StepLen)
		}
	}

}
