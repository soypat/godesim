package godesim

import (
	"math"

	"github.com/soypat/godesim/state"
	"gonum.org/v1/exp/linsolve"
	"gonum.org/v1/gonum/diff/fd"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
)

// RK4Solver Integrates simulation state for next timesteps
// using 4th order Runge-Kutta multivariable algorithm
func RK4Solver(sim *Simulation) []state.State {
	const overSix float64 = 1. / 6.
	states := make([]state.State, sim.Algorithm.Steps+1)
	h := sim.Dt() / float64(sim.Algorithm.Steps)
	states[0] = sim.State.Clone()
	for i := 0; i < len(states)-1; i++ {
		// create auxiliary states for calculation
		t := states[i].Time()
		b, c, d := states[i].CloneBlank(t+.5*h), states[i].CloneBlank(t+.5*h), states[i].CloneBlank(t+h)

		a := StateDiff(sim.Diffs, states[i])

		state.AddScaledTo(b, states[i], 0.5*h, a)
		b = StateDiff(sim.Diffs, b)

		state.AddScaledTo(c, states[i], 0.5*h, b)
		c = StateDiff(sim.Diffs, c)

		state.AddScaledTo(d, states[i], h, c)
		d = StateDiff(sim.Diffs, d)

		state.Add(a, d)
		state.Add(b, c)
		state.AddScaled(a, 2, b)
		states[i+1] = states[i].Clone()
		state.AddScaled(states[i+1], h*overSix, a)
		states[i+1].SetTime(h + states[i].Time())
	}
	return states
}

// RKF45Solver Runge-Kutta-Fehlberg of Orders 4 and 5 solver
//
// To enable adaptive stepping, Config.Algorithm.Step Min/Max values
// must be set and a Config.Error.Min must be specified in configuration.
func RKF45Solver(sim *Simulation) []state.State {
	// Butcher Tableau for Fehlbergs  4(5) method (Table III https://en.wikipedia.org/wiki/Runge%E2%80%93Kutta%E2%80%93Fehlberg_method)
	const c20, c21 = 1. / 4., 1. / 4.
	const c30, c31, c32 = 3. / 8., 3. / 32., 9. / 32.
	const c40, c41, c42, c43 = 12. / 13., 1932. / 2197., -7200. / 2197., 7296. / 2197
	const c50, c51, c52, c53, c54 = 1., 439. / 216., -8., 3680. / 513., -845. / 4104.
	const c60, c61, c62, c63, c64, c65 = .5, -8. / 27., 2., -3544. / 2565., 1859. / 4104., -11. / 40.
	// Fourth order
	const a1, a3, a4, a5 = 25. / 216., 1408. / 2565., 2197. / 4104., -1. / 5.
	// Fifth order
	const b1, b3, b4, b5, b6 = 16. / 135., 6656. / 12825., 28561. / 56430., -9. / 50., 2. / 55.
	adaptive := sim.Algorithm.Error.Max > 0 && sim.Algorithm.Step.Min > 0 && sim.Algorithm.Step.Max > sim.Algorithm.Step.Min
	states := make([]state.State, sim.Algorithm.Steps+1)
	h := sim.Dt() / float64(sim.Algorithm.Steps)
	states[0] = sim.State.Clone()
	for i := 0; i < len(states)-1; i++ {
		// create auxiliary states for calculation
		t := states[i].Time()
		k2, k3, k4, k5, k6, s4, s5, err45 := states[i].CloneBlank(t+c20*h), states[i].CloneBlank(t+c30*h), states[i].CloneBlank(t+c40*h),
			states[i].CloneBlank(t+c50*h), states[i].CloneBlank(t+c60*h), states[i].CloneBlank(t+h), states[i].CloneBlank(t+h), states[i].CloneBlank(t+h)

		k1 := StateDiff(sim.Diffs, states[i])
		state.Scale(h, k1)

		state.AddScaledTo(k2, states[i], c21, k1)
		k2 = StateDiff(sim.Diffs, k2)
		state.Scale(h, k2)

		state.AddScaledTo(k3, states[i], c31, k1)
		state.AddScaled(k3, c32, k2)
		k3 = StateDiff(sim.Diffs, k3)
		state.Scale(h, k3)

		state.AddScaledTo(k4, states[i], c41, k1)
		state.AddScaled(k4, c42, k2)
		state.AddScaled(k4, c43, k3)
		k4 = StateDiff(sim.Diffs, k4)
		state.Scale(h, k4)

		state.AddScaledTo(k5, states[i], c51, k1)
		state.AddScaled(k5, c52, k2)
		state.AddScaled(k5, c53, k3)
		state.AddScaled(k5, c54, k4)
		k5 = StateDiff(sim.Diffs, k5)
		state.Scale(h, k5)

		state.AddScaledTo(k6, states[i], c61, k1)
		state.AddScaled(k6, c62, k2)
		state.AddScaled(k6, c63, k3)
		state.AddScaled(k6, c64, k4)
		state.AddScaled(k6, c65, k5)
		k6 = StateDiff(sim.Diffs, k6)
		state.Scale(h, k6)

		// fifth order approximation calc
		state.AddScaledTo(s5, states[i], b1, k1)
		state.AddScaled(s5, b3, k3)
		state.AddScaled(s5, b4, k4)
		state.AddScaled(s5, b5, k5)
		state.AddScaled(s5, b6, k6)

		// assign solution
		states[i+1] = s5.Clone()
		// Adaptive timestep block. Modify step length if necessary
		if adaptive {
			// fourth order approximation calc
			state.AddScaledTo(s4, states[i], a1, k1)
			state.AddScaled(s4, a3, k3)
			state.AddScaled(s4, a4, k4)
			state.AddScaled(s4, a5, k5)
			// Error and adaptive timestep implementation
			state.Abs(state.SubTo(err45, s4, s5))
			errRatio := sim.Algorithm.Error.Max / state.Max(err45)
			hnew := math.Min(math.Max(0.9*h*math.Pow(errRatio, .2), sim.Algorithm.Step.Min), sim.Algorithm.Step.Max)
			sim.Algorithm.Steps = int(math.Max(float64(sim.Algorithm.Steps)*(h/hnew), 1.0))
			h = hnew
			// If we do not have desired error, and have not reached minimum timestep, repeat step
			if errRatio < 1 && h != sim.Algorithm.Step.Min {
				i--
				continue
			}
		}
	}
	return states
}

// NewtonRaphsonSolver is an implicit solver which may calculate
// the jacobian several times on each algorithm step.
//
// sim.Algorithm.Error.Max should be set to a value above 0 for
// good run
func NewtonRaphsonSolver(sim *Simulation) []state.State {
	if sim.Algorithm.Error.Max <= 0 {
		sim.Algorithm.Error.Max = 1e-5 // throwf("set config Algorithm.Error.Max to a value above 0 to use NewtonRaphson method")
	}
	jacMult := 1 - sim.Algorithm.RelaxationFactor

	if sim.Algorithm.IterationMax <= 0 {
		sim.Algorithm.IterationMax = 10
	}

	adaptive := sim.Algorithm.Error.Max > 0
	n := len(sim.Diffs)

	states := make([]state.State, sim.Algorithm.Steps+1)
	states[0] = sim.State.Clone()
	h := sim.Dt() / float64(sim.Algorithm.Steps)

	residualers := make([]func(step float64, now state.State) func(next state.State) float64, n)
	for loopi, loopsym := range sim.State.XSymbols() {
		i, sym := loopi, loopsym // escape looping variables for closure
		residualers[i] = func(step float64, now state.State) func(next state.State) float64 {
			return func(next state.State) float64 {
				return next.X(sym) - now.X(sym) - step*sim.Diffs[i](next)
			}
		}
	}
	// initialize residual functions iteration storage
	F := make(state.Diffs, n)
	// Init guess
	guess := states[0].Clone()
	auxState := states[0].Clone()
	for i := 0; i < len(states)-1; i++ {

		old := guess.Clone()
		guess.SetTime(states[i].Time() + h)
		// iteration loop counter
		iter := 0
		ierr := 0.0
		// |X_(g) - X_(i)| < permissible error
		for iter == 0 || (adaptive && iter < sim.Algorithm.IterationMax && ierr > sim.Config.Algorithm.Error.Max) {
			// First propose residual functions such that
			// F(X_(i+1)) = 0 = X_(i+1) - X_(i) - step * f(X_(i+1))
			// where f is the vector of differential equations
			for i := range residualers {
				F[i] = residualers[i](h, old)
			}

			// We solve  J^-1 * b  where b = F(X_(g)) and J = J(X_(g))
			b := mat.NewVecDense(n, StateDiff(F, guess).XVector())
			Jaux := mat.NewDense(n, n, nil)
			var settings *fd.JacobianSettings = nil //&fd.JacobianSettings{Formula: fd.Forward, Step: 1e-6}
			state.Jacobian(Jaux, F, guess, settings)
			J := denseToBand(Jaux)

			result, err := linsolve.Iterative(J, b, &linsolve.GMRES{}, &linsolve.Settings{MaxIterations: 2})
			if err != nil {
				throwf("error in newton iterative solver: %s", err)
			}
			auxState.SetAllX(result.X.RawVector().Data)

			// X_(i+1) = X_(i) - alpha * F(X_(g)) / J(X_(g)) where g are guesses, and alpha is the relaxation factor
			state.AddScaledTo(auxState, guess, -jacMult, auxState)
			// error calculation
			errvec := guess.XVector()
			floats.Sub(errvec, auxState.XVector())
			for i := range errvec {
				errvec[i] = math.Abs(errvec[i])
			}
			ierr = floats.Max(errvec)
			guess.SetAllX(auxState.XVector())
			iter++
		}

		states[i+1] = guess.Clone()
	}

	return states
}

func denseToBand(d *mat.Dense) *mat.BandDense {
	r, c := d.Caps()
	b := mat.NewBandDense(r, c, r-1, c-1, nil)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			b.SetBand(i, j, d.At(i, j))
		}
	}
	return b
}

// DormandPrinceSolver. Very similar to RKF45. Used by Matlab in ode45 solver
// and Simulink's system solver by default.
//
// To enable adaptive stepping, Config.Algorithm.Step Min/Max values
// must be set and a Config.Error.Min must be specified in configuration.
func DormandPrinceSolver(sim *Simulation) []state.State {
	// Butcher Tableau for Fehlbergs  4(5) method (Table III https://en.wikipedia.org/wiki/Runge%E2%80%93Kutta%E2%80%93Fehlberg_method)
	const c20, c21 = 1. / 5., 1. / 5.
	const c30, c31, c32 = 3. / 10., 3. / 40., 9. / 40.
	const c40, c41, c42, c43 = 4. / 5., 44. / 45., -56. / 15., 32. / 9
	const c50, c51, c52, c53, c54 = 8. / 9., 19372. / 6561., -25360. / 2187., 64448. / 6561., -212. / 729.
	const c60, c61, c62, c63, c64, c65 = 1., 9017. / 3168., -355. / 33., 46732. / 5247., 49. / 176., -5103. / 18656.
	const c70, c71, c72, c73, c74, c75, c76 = 1., 35. / 384., 0., 500. / 1113., 125. / 192., -2187. / 6784., 11. / 84.
	// Alternate solution for error calculation
	const a1, a3, a4, a5, a6, a7 = 5179. / 57600., 7571. / 16695., 393. / 640., -92097. / 339200., 187. / 2100., 1. / 40.
	// Fifth order
	const b1, b3, b4, b5, b6 = 35. / 384., 500. / 1113., 125. / 192., -2187. / 6784., 11. / 84.
	adaptive := sim.Algorithm.Error.Max > 0 && sim.Algorithm.Step.Min > 0 && sim.Algorithm.Step.Max > sim.Algorithm.Step.Min
	states := make([]state.State, sim.Algorithm.Steps+1)
	h := sim.Dt() / float64(sim.Algorithm.Steps)
	states[0] = sim.State.Clone()
	for i := 0; i < len(states)-1; i++ {
		// create auxiliary states for calculation
		t := states[i].Time()
		k2, k3, k4, k5, k6, k7, s4, s5, err45 := states[i].CloneBlank(t+c20*h), states[i].CloneBlank(t+c30*h), states[i].CloneBlank(t+c40*h),
			states[i].CloneBlank(t+c50*h), states[i].CloneBlank(t+c60*h), states[i].CloneBlank(t+c70*h), states[i].CloneBlank(t+h), states[i].CloneBlank(t+h), states[i].CloneBlank(t+h)

		k1 := StateDiff(sim.Diffs, states[i])
		state.Scale(h, k1)

		state.AddScaledTo(k2, states[i], c21, k1)
		k2 = StateDiff(sim.Diffs, k2)
		state.Scale(h, k2)

		state.AddScaledTo(k3, states[i], c31, k1)
		state.AddScaled(k3, c32, k2)
		k3 = StateDiff(sim.Diffs, k3)
		state.Scale(h, k3)

		state.AddScaledTo(k4, states[i], c41, k1)
		state.AddScaled(k4, c42, k2)
		state.AddScaled(k4, c43, k3)
		k4 = StateDiff(sim.Diffs, k4)
		state.Scale(h, k4)

		state.AddScaledTo(k5, states[i], c51, k1)
		state.AddScaled(k5, c52, k2)
		state.AddScaled(k5, c53, k3)
		state.AddScaled(k5, c54, k4)
		k5 = StateDiff(sim.Diffs, k5)
		state.Scale(h, k5)

		state.AddScaledTo(k6, states[i], c61, k1)
		state.AddScaled(k6, c62, k2)
		state.AddScaled(k6, c63, k3)
		state.AddScaled(k6, c64, k4)
		state.AddScaled(k6, c65, k5)
		k6 = StateDiff(sim.Diffs, k6)
		state.Scale(h, k6)

		// fifth order approximation calc
		state.AddScaledTo(s5, states[i], b1, k1)
		state.AddScaled(s5, b3, k3)
		state.AddScaled(s5, b4, k4)
		state.AddScaled(s5, b5, k5)
		state.AddScaled(s5, b6, k6)

		// assign solution
		states[i+1] = s5.Clone()
		// Adaptive timestep block. Modify step length if necessary
		if adaptive {
			state.AddScaledTo(k7, states[i], c71, k1)
			state.AddScaled(k7, c72, k2)
			state.AddScaled(k7, c73, k3)
			state.AddScaled(k7, c74, k4)
			state.AddScaled(k7, c75, k5)
			state.AddScaled(k7, c76, k6)
			k7 = StateDiff(sim.Diffs, k7)
			state.Scale(h, k7)
			// Alternate solution approximation calc
			state.AddScaledTo(s4, states[i], a1, k1)
			state.AddScaled(s4, a3, k3)
			state.AddScaled(s4, a4, k4)
			state.AddScaled(s4, a5, k5)
			state.AddScaled(s4, a6, k6)
			state.AddScaled(s4, a7, k7)
			// Error and adaptive timestep implementation
			state.Abs(state.SubTo(err45, s4, s5))
			errRatio := sim.Algorithm.Error.Max / state.Max(err45)
			hnew := math.Min(math.Max(0.9*h*math.Pow(errRatio, .2), sim.Algorithm.Step.Min), sim.Algorithm.Step.Max)
			sim.Algorithm.Steps = int(math.Max(float64(sim.Algorithm.Steps)*(h/hnew), 1.0))
			h = hnew
			// If we do not have desired error, and have not reached minimum timestep, repeat step
			if errRatio < 1 && h != sim.Algorithm.Step.Min {
				i--
				continue
			}
		}

	}
	return states
}

func RKF78Solver(sim *Simulation) []state.State {
	// Table X. from Classical Fifth, Sixth, Seventh and Eight Order Runge Kutta Formulas with stepsize control by Erwin Fehlberg.
	var (
		cx0 = [13]float64{0, 2. / 27., 1. / 9., 1. / 6., 5. / 12., //4
			1. / 2., 5. / 6., 1. / 6., 2. / 3., 1. / 3., 1, 0, 1}
		c = [13][12]float64{
			1:  {0: 2. / 27.},
			2:  {0: 1. / 36., 1: 1. / 12.},
			3:  {0: 1. / 24., 2: 1. / 8.},
			4:  {0: 5. / 12., 2: -25. / 16., 3: 25. / 16.},
			5:  {0: 1. / 20., 3: 1. / 4., 4: 1. / 5.},
			6:  {0: -25. / 108., 3: 125. / 108., 4: -65. / 27., 5: 125. / 54.},
			7:  {0: 31. / 300, 4: 61. / 225., 5: -2. / 9., 6: 13. / 900.},
			8:  {0: 2, 3: -53. / 6., 4: 704. / 45., 5: -107. / 9., 6: 67. / 90., 7: 3},
			9:  {0: -91. / 108., 3: 23. / 108., 4: -976. / 135., 5: 311. / 54., 6: -19. / 60., 7: 17. / 6., 8: -1. / 12.},
			10: {0: 2383. / 4100., 3: -341. / 164., 4: 4496. / 1025., -301. / 82., 2133. / 4100. /*neg?*/, 45. / 82., 45. / 164., 18. / 41.},
			11: {0: 3. / 205., 5: -6. / 41., -3. / 205., -3. / 41., 3. / 41., 6. / 41.},
			12: {0: -1777. / 4100., 3: -341. / 164., 4496. / 1025., -289. / 82., 2193. / 4100., 51. / 82., 33. / 164., 12. / 41., 0, 1},
		}
		b = [13]float64{81. / 840., 5: 34. / 105., 9. / 35., 9. / 35., 9. / 280., 9. / 280., 41. / 840.}
	)
	adaptive := sim.Algorithm.Error.Max > 0 && sim.Algorithm.Step.Min > 0 && sim.Algorithm.Step.Max > sim.Algorithm.Step.Min
	states := make([]state.State, sim.Algorithm.Steps+1)
	h := sim.Dt() / float64(sim.Algorithm.Steps)
	states[0] = sim.State.Clone()
	var k [13]state.State
	var snext, err78 state.State
	for i := 0; i < len(states)-1; i++ {
		// create auxiliary states for calculation
		t := states[i].Time()

		err78 = states[i].CloneBlank(t + h)
		snext = states[i].CloneBlank(t + h)
		for ord := 1; ord < 13; ord++ {
			k[ord] = states[i].CloneBlank(t + cx0[ord]*h)
		}

		k[0] = StateDiff(sim.Diffs, states[i])
		state.Scale(h, k[0])

		for ord := 1; ord < 13; ord++ {
			state.AddScaledTo(k[ord], states[i], c[ord][0], k[0])
			for j := 1; j < ord; j++ {
				state.AddScaled(k[ord], c[ord][j], k[j])
			}
			k[ord] = StateDiff(sim.Diffs, k[ord])
			state.Scale(h, k[ord])
		}

		// eight order approximation calc.
		state.AddScaledTo(snext, states[i], b[0], k[0])
		for ord := 1; ord < 13; ord++ {
			state.AddScaled(snext, b[ord], k[ord])
		}

		// assign solution
		states[i+1] = snext.Clone()
		// Adaptive timestep block. Modify step length if necessary
		if adaptive {
			state.AddScaled(err78, -41./840., k[0])
			state.AddScaled(err78, -41./840., k[10])
			state.AddScaled(err78, -41./840., k[11])
			state.AddScaled(err78, -41./840., k[12])
			errRatio := sim.Algorithm.Error.Max / state.Max(err78)
			hnew := math.Min(math.Max(0.9*h*math.Pow(errRatio, .2), sim.Algorithm.Step.Min), sim.Algorithm.Step.Max)
			sim.Algorithm.Steps = int(math.Max(float64(sim.Algorithm.Steps)*(h/hnew), 1.0))
			h = hnew
			// If we do not have desired error, and have not reached minimum timestep, repeat step
			if errRatio < 1 && h != sim.Algorithm.Step.Min {
				i--
				continue
			}
		}

	}
	return states
}

// DirectIntegrationSolver performs naive integration of ODEs. Should only be
// used to compare with other methods. Has the advantage of only performing one
// differentiation per step.
//  y_{n+1} = y_{n} + (dy/dt_{n} + dy/dt_{n-1})*step/2
func DirectIntegrationSolver(sim *Simulation) []state.State {
	states := make([]state.State, sim.Algorithm.Steps+1)
	t := sim.CurrentTime()
	h := sim.Dt() / float64(sim.Algorithm.Steps)
	states[0] = sim.State.Clone()
	y := sim.State.Clone()
	dydx := StateDiff(sim.Diffs, sim.State)
	for i := 0; i < len(states)-1; i++ {
		t += h
		ynext := sim.State.CloneBlank(t)
		dydxnext := StateDiff(sim.Diffs, y)
		aux := y.CloneBlank(y.Time())
		state.AddScaledTo(ynext, y, h/2, state.AddTo(aux, dydxnext, dydx))
		states[i+1] = ynext.Clone()
		dydx = dydxnext
	}
	return states
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
