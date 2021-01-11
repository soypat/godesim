package godesim_test

import (
	"math"
	"testing"

	"github.com/soypat/godesim"
	"github.com/soypat/godesim/state"
)

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
		sim.AddEventHandlers(func(s state.State) *godesim.Event {
			if s.Time() >= tswitch {
				ev := godesim.NewEvent("refine", godesim.EvStepLength)
				ev.SetStepLength(newStepLen)
				return ev
			}
			return nil
		})
		sim.Solver = solver
		sim.Begin()

		time, x_res := sim.Results("time"), sim.Results("theta")
		x_quad := applyFunc(time, func(v float64) float64 { return v /* solution is theta(t) = t*/ })

		if len(time) != len(x_res) {
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
			if math.Abs(x_quad[i]-x_res[i]) > math.Pow(sim.Dt()/float64(sim.Algorithm.Steps), 4) {
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

		sim := godesim.New()
		sim.SetChangeMap(map[state.Symbol]state.Changer{
			"theta":     func(s state.State) float64 { return s.X("theta-dot") },
			"theta-dot": Dtheta1,
		})
		sim.SetX0FromMap(map[state.Symbol]float64{
			"theta":     0,
			"theta-dot": 0,
		})
		const ti, tf, N_steps = 0.0, 2, 10
		sim.SetTimespan(ti, tf, N_steps)
		tswitch := 1.
		sim.AddEventHandlers(func(s state.State) *godesim.Event {
			if s.Time() >= tswitch {
				ev := godesim.NewEvent("change derivative", godesim.EvBehaviour)
				ev.SetBehaviour(map[state.Symbol]func(state.State) float64{
					"theta-dot": Dtheta2,
				})
				return ev
			}
			return nil
		})
		sim.Solver = solver
		sim.Begin()

		time, x_res := sim.Results("time"), sim.Results("theta")

		x_expected := applyFunc(time, func(v float64) float64 {
			if v > tswitch {
				return math.NaN() // I haven't figured out exact solution after switching equation
			}
			return math.Pow(v, 3)
		})

		if len(time) != len(x_res) {
			t.Error("length of time and theta vectors should be the same")
		}
		for i := range time {
			diff := x_res[i] - x_expected[i]
			if math.Abs(diff) > math.Pow(sim.Dt()/float64(sim.Algorithm.Steps), 4) {
				if time[i] > tswitch {
					continue // I haven't figured the exact solution after tswitch
				}
				t.Errorf("incorrect curve profile for test %s. t=%.2f Expected %.4f, got %.4f", t.Name(), time[i], x_expected[i], x_res[i])
			}
		}
	}
}

func TestMultiEvent(t *testing.T) {
	for _, solver := range gdsimSolvers {
		Dtheta1 := func(s state.State) float64 {
			return 6 * s.Time()
		}

		sim := godesim.New()
		sim.SetChangeMap(map[state.Symbol]state.Changer{
			"theta":     func(s state.State) float64 { return s.X("theta-dot") },
			"theta-dot": Dtheta1,
		})
		sim.SetX0FromMap(map[state.Symbol]float64{
			"theta":     0,
			"theta-dot": 0,
		})
		const ti, tf, N_steps = 0.0, 3, 15
		sim.SetTimespan(ti, tf, N_steps)
		stepOriginal := sim.Dt()
		stepNew := 0.5 * stepOriginal
		tStepRefine := 1.
		tNewEndSim := 2.
		sim.AddEventHandlers(
			func(s state.State) *godesim.Event {
				if s.Time() >= tNewEndSim {
					return godesim.NewEvent("time up", godesim.EvEndSimulation)
				}
				return nil
			},
			// Augment step size after vehicle runs out of fuel
			func(s state.State) *godesim.Event {
				if s.Time() >= tStepRefine {
					ev := godesim.NewEvent("refine", godesim.EvStepLength)
					ev.SetStepLength(stepNew)
					return ev
				}
				return nil
			},
		)
		sim.Solver = solver
		sim.Begin()

		time, x_res := sim.Results("time"), sim.Results("theta")

		x_expected := applyFunc(time, func(v float64) float64 { return math.Pow(v, 3) })
		if len(time) != len(x_res) {
			t.Error("length of time and theta vectors should be the same")
			t.FailNow()
		}
		if math.Abs(time[len(time)-1]-tNewEndSim) > 1e-12 {
			t.Errorf("simulation end event not triggered at domain point. Expected %.3f, got %.3f", tNewEndSim, time[len(time)-1])
		}
		for i := range time {
			diff := x_res[i] - x_expected[i]
			if math.Abs(diff) > math.Pow(sim.Dt()/float64(sim.Algorithm.Steps), 4) {
				t.Errorf("incorrect curve profile for test %s. t=%.2f Expected %.4f, got %.4f", t.Name(), time[i], x_expected[i], x_res[i])
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
