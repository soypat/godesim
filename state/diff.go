package state

import (
	"gonum.org/v1/gonum/diff/fd"
	"gonum.org/v1/gonum/mat"
)

// Diff represents a single non-linear differential equation
// of the Simulation's system of differential equation. Specifically, it
// represents an X variable change.
type Diff func(State) float64

// Diffs represents a coupled non-linear algebraic system
type Diffs []Diff

// Input represents a time-varying or table look-up variable/coefficient
// of the Simulation's system of differential equations. Inputs can be used to
// model non-autonomous system of differential equations. Input functions are
// called after solver algorithm finishes on the resulting state.
type Input func(State) float64

// Jacobian approximates jacobian matrix for Diffs system
func Jacobian(dst *mat.Dense, d Diffs, s State, settings *fd.JacobianSettings) *mat.Dense {
	f := func(y, x []float64) {
		sx := s.Clone()
		sx.SetAllX(x)
		for i := 0; i < len(d); i++ {
			y[i] = d[i](sx)
		}
	}
	fd.Jacobian(dst, f, s.x, settings)
	return dst
}
