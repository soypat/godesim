package godesim

import (
	"math"
	"testing"

	"github.com/soypat/godesim/state"
)

type TypicalEventer struct {
	action func(state.State) func(*Simulation) error
	label  string
}

func (ev TypicalEventer) Event(s state.State) func(*Simulation) error {
	return ev.action(s)
}

func (ev TypicalEventer) Label() string {
	if ev.label == "" {
		panic("empty Eventer label")
	}
	return ev.label
}

// TestStepLen changes steplength mid-simulation.
// Verifies change of steplength and accuracy of results for simpleInput
func TestStepLen(t *testing.T) {
	for _, solver := range gdsimSolvers {
		Dtheta := func(s state.State) float64 {
			return s.U("u")
		}

		inputVar := func(s state.State) float64 {
			return 1
		}
		sim := New()
		sim.SetDiffFromMap(map[state.Symbol]state.Diff{
			"theta": Dtheta,
		})
		sim.SetX0FromMap(map[state.Symbol]float64{
			"theta": 0,
		})
		sim.SetInputFromMap(map[state.Symbol]state.Input{
			"u": inputVar,
		})
		const NSteps, ti, tf = 10, 0.0, 1.0

		sim.SetTimespan(ti, tf, NSteps)
		initStepLen := sim.Dt()
		newStepLen := initStepLen * 0.25
		tswitch := 0.5
		var refiner Eventer = TypicalEventer{
			label: "refine",
			action: func(s state.State) func(*Simulation) error {
				if s.Time() >= tswitch {
					return NewStepLength(newStepLen)
				}
				return nil
			},
		}
		sim.AddEventHandlers(refiner)
		sim.Solver = solver.f
		sim.Begin()

		time, xResults := sim.Results("time"), sim.Results("theta")
		xQuad := applyFunc(time, func(v float64) float64 { return v /* solution is theta(t) = t*/ })

		if len(time) != len(xResults) {
			t.Error("length of time and theta vectors should be the same")
		}
		expectedLen := initStepLen
		for i, tm := range time[:len(time)-2] {
			if tm >= tswitch {
				expectedLen = newStepLen
			}
			StepLen := time[i+1] - tm
			if math.Abs(StepLen-expectedLen) > 1e-12 {
				t.Errorf("expected stepLength %.4f. Got %.4f", expectedLen, StepLen)
			}
			// Also test accuracy of results
			if math.Abs(xQuad[i]-xResults[i]) > solver.err(sim.Dt(), float64(i)) {
				t.Errorf("incorrect curve profile for test %s", t.Name())
			}
		}
	}
}

// TestBehaviourCubicToQuartic This one's solution is more complex.
// theta-dot's solution for the IVP theta-dot(t=0)=0 is  theta-dot=t^2
// thus theta's solution then is theta=1/3*t^3
func TestBehaviourCubicToQuartic(t *testing.T) {
	for _, solver := range gdsimSolvers {
		Dtheta1 := func(s state.State) float64 {
			return 6 * s.Time()
		}
		Dtheta2 := func(s state.State) float64 {
			return 12 * s.Time() * s.Time()
		}

		sim := New()
		sim.SetDiffFromMap(map[state.Symbol]state.Diff{
			"theta":     func(s state.State) float64 { return s.X("theta-dot") },
			"theta-dot": Dtheta1,
		})
		sim.SetX0FromMap(map[state.Symbol]float64{
			"theta":     0,
			"theta-dot": 0,
		})
		const ti, tf, NSteps = 0.0, 2, 10
		sim.SetTimespan(ti, tf, NSteps)
		tswitch := 1.
		var quartic Eventer = TypicalEventer{
			label: "change derivative",
			action: func(s state.State) func(*Simulation) error {
				if s.Time() >= tswitch {
					return DiffChangeFromMap(map[state.Symbol]func(state.State) float64{
						"theta-dot": Dtheta2,
					})
				}
				return nil
			},
		}
		sim.AddEventHandlers(quartic)
		sim.Solver = solver.f
		sim.Begin()

		time, xResults := sim.Results("time"), sim.Results("theta")

		xExpected := applyFunc(time, func(v float64) float64 {
			if v > tswitch {
				return math.NaN() // I haven't figured out exact solution after switching equation
			}
			return math.Pow(v, 3)
		})

		if len(time) != len(xResults) {
			t.Error("length of time and theta vectors should be the same")
		}
		for i := range time {
			diff := xResults[i] - xExpected[i]
			if math.Abs(diff) > solver.err(sim.Dt(), float64(i)) {
				if time[i] > tswitch {
					break // I haven't figured the exact solution after tswitch
				}
				t.Errorf("%s:curve expected %6.4g, got %6.4g", solver.name, xExpected[i], xResults[i])
			}
		}
	}
}

func TestAddEventsError(t *testing.T) {
	sim := newWorkingSim()
	defer func() {
		err := recover()
		if err == nil {
			t.Error("should panic when adding 0 events")
		}
	}()
	sim.AddEventHandlers()
}

func TestMultiEvent(t *testing.T) {
	for _, solver := range gdsimSolvers {
		Dtheta1 := func(s state.State) float64 {
			return 6 * s.Time()
		}

		sim := New()
		sim.SetDiffFromMap(map[state.Symbol]state.Diff{
			"theta":     func(s state.State) float64 { return s.X("theta-dot") },
			"theta-dot": Dtheta1,
		})
		sim.SetX0FromMap(map[state.Symbol]float64{
			"theta":     0,
			"theta-dot": 0,
		})
		// The test is sensitive to these values since
		// it expects a discrete point around tNewEndSim and tStepRefine
		// to apply events. Of course, depending on domain subdivision, the
		// point at which the event is applied may be just under a step-length away
		const ti, tf, NSteps = 0.0, 3, 30
		sim.SetTimespan(ti, tf, NSteps)
		stepOriginal := sim.Dt()
		stepNew := 0.5 * stepOriginal
		tStepRefine := 1.
		tNewEndSim := 2.
		var endsim Eventer = TypicalEventer{
			label: "end sim",
			action: func(s state.State) func(*Simulation) error {
				if s.Time() >= tNewEndSim-1e-4 {
					return EndSimulation //NewEvent("time up", EvEndSimulation)
				}
				return nil
			},
		}
		var refiner Eventer = TypicalEventer{
			label: "refine",
			action: func(s state.State) func(*Simulation) error {
				if s.Time() >= tStepRefine-1e-4 {
					return NewStepLength(stepNew)
				}
				return nil
			},
		}
		sim.AddEventHandlers(endsim, refiner)
		sim.Solver = solver.f
		sim.Begin()
		evs := sim.Events()
		if len(evs) != 2 {
			t.Error("expected 2 events returned")
		}
		time, xResults := sim.Results("time"), sim.Results("theta")

		xExpected := applyFunc(time, func(v float64) float64 { return math.Pow(v, 3) })
		if len(time) != len(xResults) {
			t.Error("length of time and theta vectors should be the same")
			t.FailNow()
		}
		if math.Abs(time[len(time)-1]-tNewEndSim) > 1e-12 {
			t.Errorf("simulation end event not triggered at domain point. Expected %.3f, got %.3f", tNewEndSim, time[len(time)-1])
		}
		for i := range time {
			diff := xResults[i] - xExpected[i]
			if math.Abs(diff) > solver.err(sim.Dt(), float64(i)) {
				t.Errorf("%s:curve expected %6.4g, got %6.4g", solver.name, xExpected[i], xResults[i])
			}
			if i == 0 {
				continue
			}
			dt, tm := time[i]-time[i-1], time[i]
			if tm <= tStepRefine && math.Abs(dt-stepOriginal) > 1e-12 {
				t.Errorf("steplength event triggered before its time")
			}
			if tm > tStepRefine && math.Abs(dt-stepNew) > 1e-12 {
				t.Errorf("steplength event not applied. expected dt=%.4f, got dt=%.4f", stepNew, dt)
			}
		}
	}
}

func TestEventErrors(t *testing.T) {
	sim := New()
	sim.SetX0FromMap(map[state.Symbol]float64{
		"x": 1,
	})
	sim.SetDiffFromMap(map[state.Symbol]state.Diff{
		"x": func(s state.State) float64 { return 0 },
	})

	var badEvent Eventer = TypicalEventer{
		label: "change derivative",
		action: func(s state.State) func(*Simulation) error {
			err := DiffChangeFromMap(map[state.Symbol]func(state.State) float64{
				"x":     func(s state.State) float64 { return 0 },
				"extra": func(s state.State) float64 { return 0 },
			})
			if err != nil {
				panic(err)
			}
			return nil
		},
	}
	sim.SetTimespan(0, 1, 10)
	sim.AddEventHandlers(badEvent)
	if EventDone(sim) != nil {
		t.Error("expected nil return from EventDone")
	}
	defer func() {
		err := recover()
		if err == nil {
			t.Error("should have gotten an error from a bad event")
		}
	}()
	sim.Begin()

}
