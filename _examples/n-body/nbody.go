package main

import (
	"fmt"
	"math"

	"github.com/soypat/godesim"
	"github.com/soypat/godesim/state"
)

// Declare simulation constants: gravity and pendulum length
const softening, l float64 = 10, 1. // m/s2, m
var sin, pi = math.Sin, math.Pi

type body struct {
	name    string
	mass    float64
	x, y, z float64
}

type bodies []body

func (bds bodies) ChangeMap() map[state.Symbol]state.Diff {

}

func main() {
	Dthetadot := func(s state.State) float64 {
		return -g / l * sin(s.X("theta"))
	}

	sim := godesim.New()

	sim.SetDiffFromMap(map[state.Symbol]state.Diff{
		"theta":     func(s state.State) float64 { return s.X("theta-dot") },
		"theta-dot": Dthetadot,
	})

	sim.SetX0FromMap(map[state.Symbol]float64{
		"theta":     20. * pi / 180., // convert angles to radians
		"theta-dot": 0,
	})

	sim.SetTimespan(0., 8., 100)

	sim.Begin()

	time, theta := sim.Results("time"), sim.Results("theta")
	fmt.Printf("%.2f\n\n%.2f\n", time, theta)
}
