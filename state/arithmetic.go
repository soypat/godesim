package state

import (
	"math"

	"gonum.org/v1/gonum/floats"
)

// Abs takes absolute value of elements in X variables
// and stores the result in dst.
func Abs(dst State) {
	for i := range dst.x {
		dst.x[i] = math.Abs(dst.x[i])
	}
}

// Add adds, element-wise the X variables of state
// s and dst and stores result in dst.
// It panics if the slice argument lengths do not match.
func Add(dst, s State) {
	floats.Add(dst.x, s.x)
}

// AddTo adds, element-wise, the elements of s and t and
// stores the result in dst. Returns dst.
// It panics if the slice argument lengths do not match.
func AddTo(dst, s, t State) State {
	floats.AddTo(dst.x, s.x, t.x)
	return dst
}

// AddConst adds the scalar c to all of the values in dst.
func AddConst(c float64, dst State) {
	floats.AddConst(c, dst.x)
}

// AddScaled performs dst = dst + alpha * s.
// It panics if the slice argument lengths do not match.
func AddScaled(dst State, alpha float64, s State) {
	floats.AddScaled(dst.x, alpha, s.x)
}

// AddScaledTo performs elementwise dst = y + alpha * s, where alpha is a scalar,
// and dst, y and s are all slices.
// It panics if the slice argument lengths do not match.
func AddScaledTo(dst, y State, alpha float64, s State) State {
	dst.x = floats.AddScaledTo(dst.x, y.x, alpha, s.x)
	return dst
}

// Div performs element-wise division dst / s
// and stores the value in dst.
// It panics if the argument lengths do not match.
func Div(dst, s State) {
	floats.Div(dst.x, s.x)
}

// DivTo performs element-wise division s / t
// and stores the value in dst. Returns modified dst.
// It panics if the argument lengths do not match.
func DivTo(dst, s, t State) State {
	floats.DivTo(dst.x, s.x, t.x)
	return dst
}

// Max returns the maximum value of s
func Max(s State) float64 {
	return floats.Max(s.x)
}

// Max returns the minimum value of s
func Min(s State) float64 {
	return floats.Min(s.x)
}

// Mul performs element-wise multiplication between dst
// and s and stores the value in dst.
// It panics if the argument lengths do not match.
func Mul(dst, s State) {
	floats.Mul(dst.x, s.x)
}

// MulTo performs element-wise multiplication between s
// and t and stores the value in dst.
// It panics if the argument lengths do not match.
func MulTo(dst, s, t State) State {
	floats.MulTo(dst.x, s.x, t.x)
	return dst
}

// Scale multiplies every element in dst by the scalar c.
func Scale(c float64, dst State) {
	floats.Scale(c, dst.x)
}

// ScaleTo multiplies the elements in s by c and stores the result in dst.
// It panics if the slice argument lengths do not match.
func ScaleTo(dst State, c float64, s State) State {
	floats.ScaleTo(dst.x, c, s.x)
	return dst
}

// Sub subtracts, element-wise, the elements of s from dst.
// It panics if the argument lengths do not match.
func Sub(dst, s State) {
	floats.Sub(dst.x, s.x)
}

// SubTo subtracts, element-wise, the elements of t from s and
// stores the result in dst.
// It panics if the argument lengths do not match.
func SubTo(dst, s, t State) State {
	floats.SubTo(dst.x, s.x, t.x)
	return dst
}
