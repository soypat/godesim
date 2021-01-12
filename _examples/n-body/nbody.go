package main

import (
	"fmt"
	"math"

	"github.com/soypat/godesim"
	"github.com/soypat/godesim/state"
)

// Declare simulation constants: softening coefficient and big G
const softening, G float64 = 1.0, 6.6743e-11

var sin, pi = math.Sin, math.Pi

type body struct {
	name string
	// [kg]
	mass float64
	// initial positions [m]
	x0, y0, z0 float64
	// initial velocities [m/s]
	u0, v0, w0 float64
}

func (b body) sym(s string) state.Symbol { return state.Symbol(fmt.Sprintf("%s_%s", s, b.name)) }

type bodies []body

func (bds bodies) DiffMap() map[state.Symbol]state.Diff {
	m := make(map[state.Symbol]state.Diff)
	for i := range bds {
		bd1 := bds[i] // define new variable so closure escapes looping variable
		for _, x := range []string{"x", "y", "z"} {
			vars := x // escape looping variable
			sym1 := bd1.sym(vars)
			Dsym := "D" + sym1
			m[sym1] = func(s state.State) float64 { return s.X(Dsym) }
			m[Dsym] = func(s state.State) float64 {
				sum := 0.0
				for _, bd2 := range bds {
					if bd1.name == bd2.name {
						continue
					}
					diff := s.X(bd2.sym(vars)) - s.X(sym1)
					sum += bd2.mass * diff * math.Pow(math.Abs(diff)+softening, -3.0)
				}
				return G * sum
			}
		}
	}
	return m
}
func (bds bodies) X0Map() map[state.Symbol]float64 {
	m := make(map[state.Symbol]float64)
	for _, bd := range bds {
		for _, x := range []string{"x", "y", "z"} {
			sym := bd.sym(x)
			Dsym := "D" + sym
			switch x {
			case "x":
				m[sym], m[Dsym] = bd.x0, bd.u0
			case "y":
				m[sym], m[Dsym] = bd.y0, bd.v0
			case "z":
				m[sym], m[Dsym] = bd.z0, bd.w0
			}
		}
	}
	return m
}

func main() {

	system := bodies{
		body{name: "earth", mass: 5.972e24}, // what do you mean geocentric model not true?
		body{name: "moon", mass: 7.3477e22, x0: 384e6, v0: 1.022e3},
		body{name: "iss", mass: 420e3, x0: 408e3, v0: 7.66e3},
	}
	sim := godesim.New()

	sim.SetDiffFromMap(system.DiffMap())
	x0 := system.X0Map()
	sim.SetX0FromMap(x0)

	sim.SetTimespan(0., daysToSeconds(28.), 1000)

	sim.Begin()

	time, x := sim.Results("time"), sim.Results(system[1].sym("x"))
	fmt.Printf("%.2f\n\n%.2f\n", time, x)
}

func daysToSeconds(d float64) float64 {
	const d2s = 24. * 60. * 60.
	return d * d2s
}
