// equations from https://www.math24.net/double-pendulum/
package main

import (
	"fmt"
	"math"

	"github.com/soypat/godesim"
	"github.com/soypat/godesim/state"
)

// Declare simulation constants: gravity and pendulum length
const (
	g, l1, l2 = 9.8, 1., 1. // m/s2, m
	m1, m2    = 1., 1.
)

var sin, cos, pi = math.Sin, math.Cos, math.Pi

func main() {
	// equations from https://www.math24.net/double-pendulum/
	Dtheta1 := func(s state.State) float64 {
		return (s.X("mom1")*l2 - l1*s.X("mom2")*cos(s.X("theta1")-s.X("theta2"))) /
			(m1 * l1 * l1 * l2 * (m1 + m2*math.Pow(sin(s.X("theta1")-s.X("theta2")), 2.)))
	}
	Dtheta2 := func(s state.State) float64 {
		return (l1*s.X("mom2")*(m1+m2) - l2*m2*s.X("mom1")*cos(s.X("theta1")-s.X("theta2"))) /
			(m2 * l1 * l2 * l2 * (m1 + m2*math.Pow(sin(s.X("theta1")-s.X("theta2")), 2.)))
	}
	Dmom1 := func(s state.State) float64 {
		return -(m1+m2)*g*l1*sin(s.X("theta1")) - a1(s) + a2(s)
	}
	Dmom2 := func(s state.State) float64 {
		return -m2*g*l2*sin(s.X("theta2")) + a1(s) - a2(s)
	}
	sim := godesim.New()

	sim.SetChangeMap(map[state.Symbol]state.Changer{
		"theta1": Dtheta1,
		"mom1":   Dmom1,
		"theta2": Dtheta2,
		"mom2":   Dmom2,
	})

	sim.SetX0FromMap(map[state.Symbol]float64{
		"theta1": 20. * pi / 180., // convert angles to radians
		"mom1":   0,
		"theta2": 20 * pi / 180., // convert angles to radians
		"mom2":   0,
	})

	sim.SetTimespan(0., 8., 100)

	sim.Begin()

	time, theta := sim.Results("time"), sim.Results("theta1")
	fmt.Printf("%.2f\n\n%.2f\n", time, theta)
	// pixelgl.Run(run)
}

func a1(s state.State) float64 {
	return (s.X("mom1") * s.X("mom2") * sin(s.X("theta1")-s.X("theta2"))) /
		(l1 * l2 * (m1 + m2*math.Pow(s.X("theta1")-s.X("theta2"), 2)))
}

func a2(s state.State) float64 {
	p1, p2, alpha := s.X("mom1"), s.X("mom2"), s.X("theta1")-s.X("theta2")
	return (math.Pow(p1*l2, 2)*m2 - 2*p1*p2*m2*l1*l2*cos(alpha) + math.Pow(p2*l2, 2)*(m1+m2)) * sin(2*alpha) /
		2 * math.Pow(l1*l2*(m1+m2*math.Pow(sin(alpha), 2)), 2)
}

func run() {

}
