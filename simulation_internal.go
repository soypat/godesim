package godesim

import (
	"fmt"
	"math"
	"sort"

	"github.com/soypat/godesim/state"
	"gonum.org/v1/gonum/floats"
)

const (
	escape   = "\x1b"
	yellow   = 33
	hiyellow = 93
)

// throwf terminate simulation run inmediately due to error
func throwf(format string, a ...interface{}) {
	panic(fmt.Errorf(format+"\n", a...))
}

func scolorf(color int, str string) string {
	return fmt.Sprintf("%s[%dm%s%s[0m", escape, color, str, escape)
}

// warnf Gives you option to terminate simulation run inmediately due to error
func warnf(format string, a ...interface{}) {
	fmt.Printf(scolorf(yellow, format)+"\n", a...) //33 or 93 for HiYellow
}

func (sim *Simulation) verifyPreBegin() {
	sim.verify()
	if len(sim.results) > 0 {
		throwf("Simulation.Begin(): Simulation results not empty")
	}
	if !sim.Symbols.NoOrdering {
		sim.State = orderedState(sim.State)
	}
	sim.setDiffs()
}

func (sim *Simulation) verify() {
	if len(sim.State.XSymbols()) == 0 {
		throwf("Simulation: no X Symbols defined")
	}
	if sim.Len() == 0 {
		throwf("Simulation: no step (time) vector defined")
	}
	if sim.Solver == nil {
		throwf("Simulation: expected Simulation.Solver. got nil")
	}
	symsX, symsU := sim.diffSymbols(), sim.inputSymbols()
	consX, consU := sim.State.ConsistencyX(symsX), sim.State.ConsistencyU(symsU)

	if floats.HasNaN(consX) {
		nanidx, _ := floats.Find([]int{}, math.IsNaN, consX, -1)
		if len(nanidx) == 1 {
			throwf("Simulation: : X State is inconsistent for %v. Match X Diff with State Symbols", symsX[nanidx[0]])
		} else {
			throwf("Simulation: X State is inconsistent for %v and %d cases. Match X Change with State Symbols", symsX[nanidx[0]], len(nanidx)-1)
		}
		panic("should be unreachable")
	}
	if floats.HasNaN(consU) {
		nanidx, _ := floats.Find([]int{}, math.IsNaN, consU, -1)
		if len(nanidx) == 1 {
			throwf("Simulation: : U State is inconsistent for %v. Match U Inputs with State Symbols", symsU[nanidx[0]])
		} else {
			throwf("Simulation: U State is inconsistent for %v and %d case(s). Match U Inputs with State Symbols", symsU[nanidx[0]], len(nanidx)-1)
		}
		panic("should be unreachable")
	}
}

// IsRunning returns true if Simulation has yet not run last iteration
func (sim *Simulation) IsRunning() bool {
	if sim.currentStep < 0 {
		return false
	}
	return sim.Timespan.end > sim.State.Time()+sim.Dt()*.9
}

func (sim *Simulation) diffSymbols() []state.Symbol {
	syms := make([]state.Symbol, 0, len(sim.change))
	for sym := range sim.change {
		syms = append(syms, sym)
	}
	return syms
}

func (sim *Simulation) inputSymbols() []state.Symbol {
	if len(sim.inputs) == 0 {
		return []state.Symbol{}
	}
	syms := make([]state.Symbol, 0, len(sim.inputs))
	for sym := range sim.inputs {
		syms = append(syms, sym)
	}
	return syms
}

func (sim *Simulation) setInputs() {
	if len(sim.inputs) == 0 {
		return
	}
	for sym, f := range sim.inputs {
		sim.State.UEqual(sym, f(sim.State))
	}
}

// generates a new state with ordered X symbols
func orderedState(s state.State) state.State {
	syms := s.XSymbols()
	str := make([]string, len(syms))
	for i := range syms {
		str[i] = string(syms[i])
	}
	sort.Strings(str)
	newS := state.New()
	for i := range str {
		newS.XEqual(state.Symbol(str[i]), s.X(state.Symbol(str[i])))
	}
	syms = s.USymbols()
	for i := range syms {
		newS.UEqual(syms[i], s.U(syms[i]))
	}
	newS.SetTime(s.Time())
	return newS
}

func (sim *Simulation) setDiffs() {
	sim.Diffs = make(state.Diffs, len(sim.change))
	for i, sym := range sim.State.XSymbols() {
		sim.Diffs[i] = sim.change[sym]
	}
}

func (sim *Simulation) handleEvents() {
	for i := 0; i < len(sim.eventers); i++ {
		handler := sim.eventers[i]
		ev := handler.Event(sim.State)
		if ev == nil { //no action
			continue
		}
		err := ev(sim)
		if err == nil { // add happened event to event list
			sim.events = append(sim.events, struct {
				Label string
				State state.State
			}{Label: handler.Label(), State: sim.State.Clone()})

		} else if err.Error() != ErrorRemove.Error() {
			fmt.Println("error in simulation event: ", err)
		}
		// we always remove event after applying it to simulation
		sim.eventers = append(sim.eventers[:i], sim.eventers[i+1:]...)
		i--
	}
}
