// Simulation can be described as a simple interface
// to solve a system of non-linear differential equations
// which can be defined as Go code.
//
// Why Simulation?
//
// ODE solvers seem to fill the niche of simple system solvers in
// your numerical packages such as scipy's odeint/solve_ivp. Among these integrators there
// seems to be room for a solver that offers simulation interactivity such as modifying
// the differential equations during simulation based on events such as a rocket stage separation.
package simulation

import (
	"fmt"
	"time"

	"github.com/go-sim/simulation/state"
	"gonum.org/v1/gonum/floats"
)

const overSix float64 = 0.166666666666666666666667

// Simulation contains dynamics of system and stores
// simulation results.
//
// Defines an object that can solve
// a non-autonomous, non-linear system
// of differential equations
type Simulation struct {
	Timespan
	State       state.State
	currentTime float64
	currentStep int
	SolverSteps int
	results     []state.State
	Solver      func(sim *Simulation) []state.State
	Change      map[state.Symbol]state.Changer
	Inputs      map[state.Symbol]state.Input
	Config
}

type Config struct {
	Domain state.Symbol `yaml:"domain"`
	Log    struct {
		Results bool `yaml:"results"`
	} `yaml:"log"`
	Behaviour struct {
		StepDelay time.Duration `yaml:"delay"`
	} `yaml:"behaviour"`
	Algorithm struct {
		Steps int `yaml:"steps"`
	} `yaml:"algorithm"`
}

// New creates blank simulation
//
// Default values:
//
// Domain is the integration variable. "time" is default value
//	simulation.Domain
// Solver used is fourth order Runge-Kutta multivariable integration.
//  simulation.Solver
// How many solver steps are run between Timespan steps. Set to 1
//  simulation.SolverSteps
func New() *Simulation {
	sim := Simulation{
		Change:      make(map[state.Symbol]state.Changer),
		Solver:      RK4Solver,
		SolverSteps: 1,
	}
	sim.Domain = "time"
	return &sim
}

func (sim *Simulation) SetConfig(cfg Config) *Simulation {
	sim.Config = cfg
	return sim
}

// Begin starts simulation
//
// Unrecoverable errors will be printed as
func (sim *Simulation) Begin() {
	// This is step 0 of simulation
	sim.verifyPreBegin()

	sim.results = make([]state.State, 0, sim.SolverSteps*sim.Len())
	sim.results = append(sim.results, sim.State)

	var states []state.State
	for sim.isRunning() {
		sim.currentStep++
		states = sim.Solver(sim)
		sim.results = append(sim.results, states[1:]...)
		sim.State = states[len(states)-1]
		if sim.Log.Results {
			fmt.Printf("%v\n", sim.State)
		}
		time.Sleep(sim.Behaviour.StepDelay)
	}
}

// RK4Solver Integrates simulation state for next timesteps
// using 4th order Runge-Kutta multivariable algorithm
func RK4Solver(sim *Simulation) []state.State {
	states := make([]state.State, sim.SolverSteps+1)
	dt := sim.Dt()
	states[0] = sim.State.Clone()
	for i := 0; i < len(states)-1; i++ {
		// create auxiliary states for calculation
		b, c, d := states[i].CloneBlank(), states[i].CloneBlank(), states[i].CloneBlank()
		a := StateDiff(sim.Change, states[i])
		aaux := a.Clone()
		b = StateDiff(sim.Change, state.AddTo(b, states[i],
			state.ScaleTo(aaux, 0.5*dt, aaux)))
		baux := b.Clone()
		c = StateDiff(sim.Change, state.AddTo(c, states[i],
			state.ScaleTo(baux, 0.5*dt, baux)))
		caux := c.Clone()
		d = StateDiff(sim.Change, state.AddTo(d, states[i],
			state.ScaleTo(caux, dt, caux)))
		state.Add(a, d)
		state.Add(b, c)
		state.AddScaled(a, 2, b)
		states[i+1] = states[i].Clone()
		state.AddScaled(states[i+1], dt*overSix, a)
		states[i+1].SetTime(dt + states[i].Time())
	}
	return states
}

// SetX0FromMap sets simulation's initial X values from a Symbol map
func (sim *Simulation) SetX0FromMap(m map[state.Symbol]float64) {
	sim.State = state.New()
	for sym, v := range m {
		sim.State.XEqual(sym, v)
	}
}

// SetChangeMap Sets the ODE equations (change in X) with a pre-built map
//
// i.e. theta(t) = 0.5 * t^2
//
//  sim.SetChangeMap(map[state.Symbol]state.Changer{
//  	"theta":  func(s state.State) float64 {
//  		return s.X("Dtheta")
//  	},
//  	"Dtheta": func(s state.State) float64 {
//  		return 1
//  	},
//  })
func (sim *Simulation) SetChangeMap(m map[state.Symbol]state.Changer) {
	sim.Change = m
}

// SetChangeMap Sets U functions with pre-built map
func (sim *Simulation) SetInputMap(m map[state.Symbol]state.Input) {
	sim.Inputs = m
}

// CurrentStep get number of steps done.
// Reaches maximum Simulation's Timespan's `Steps`
func (sim *Simulation) CurrentStep() int {
	return sim.currentStep
}

// CurrentTime obtain simulation step variable
func (sim *Simulation) CurrentTime() float64 {
	return sim.results[len(sim.results)-1].Time()
}

// LastTime Obtains last Simulation step time.
// Does not take into account Solver's steps
func (sim *Simulation) LastTime() float64 {
	return sim.stepLength * float64(sim.CurrentStep())
}

// Results get vector of simulation results for given symbol (X or U)
//
// Special case is the Simulation.Domain (default "time") symbol.
func (sim *Simulation) Results(sym state.Symbol) []float64 {
	vec := make([]float64, len(sim.results))
	// TODO verify simulation has run!
	if sym == sim.Domain {
		for i, r := range sim.results {
			vec[i] = r.Time()
		}
		return vec
	}
	symV := []state.Symbol{sym}
	consU, consX := !floats.HasNaN(sim.State.ConsistencyU(symV)), !floats.HasNaN(sim.State.ConsistencyX(symV))
	if consU {
		for i, r := range sim.results {
			vec[i] = r.U(sym)
		}
	}
	if consX {
		for i, r := range sim.results {
			vec[i] = r.X(sym)
		}
	}
	return vec
}

// StateDiff obtain StateChanger results without modifying State
// Returns state evolution (result of applying Changer functions to S)
func StateDiff(F map[state.Symbol]state.Changer, S state.State) state.State {
	diff := S.Clone()
	syms := S.XSymbols()
	if len(F) != len(syms) {
		throwf("length of func slice not equal to float slice (%v vs. %v)", len(F), len(syms))
	}
	for i := 0; i < len(F); i++ {
		diff.XEqual(syms[i], F[syms[i]](S))
	}
	return diff
}
