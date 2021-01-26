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
func Jacobian(dst *mat.Dense, d Diffs, s State) *mat.Dense {
	n := len(d)
	f := func(y, x []float64) {
		sx := State{x: x}
		for i := 0; i < len(d); i++ {
			y[i] = d[i](sx)
		}
	}
	j := &mat.Dense{}
	fd.Jacobian(j, f, s.x, nil)
	mat.NewBandDense(n, n, n-1, n-1, j.RawMatrix().Data)
	return dst
}

func (d Diffs) with(s State) []float64 {
	y := make([]float64, len(s.x))
	for i := 0; i < len(d); i++ {
		y[i] = d[i](s)
	}
	return y
}
