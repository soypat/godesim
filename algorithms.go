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
		b = [13]float64{41. / 840., 5: 34. / 105., 9. / 35., 9. / 35., 9. / 280., 9. / 280., 41. / 840.}
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
			errFactor := h * 41. / 840.
			state.AddScaled(err78, -errFactor, k[0])
			state.AddScaled(err78, -errFactor, k[10])
			state.AddScaled(err78, errFactor, k[11])
			state.AddScaled(err78, errFactor, k[12])
			errRatio := sim.Algorithm.Error.Max / state.Max(err78)
			scale := 0.8 * math.Abs(math.Pow(errRatio, 1./7.))
			scale = math.Min(math.Max(scale, .125), 4.0)
			hnew := math.Min(math.Max(h*scale, sim.Algorithm.Step.Min), sim.Algorithm.Step.Max)
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

func RKF10_12Solver(sim *Simulation) []state.State {
	const rkLen = 17
	// Table X. from Classical Fifth, Sixth, Seventh and Eight Order Runge Kutta Formulas with stepsize control by Erwin Fehlberg.
	var (
		cx0 = [rkLen]float64{0.0e0, 2.0e-2, 4.0e-2, 1.0e-1, 1.33333333333333333333333333333e-1, 1.6e-1, 5.0e-2, 2.0e-1, 2.5e-1, 3.33333333333333333333333333333e-1, 5.0e-1, 5.55555555555555555555555555556e-1, 7.5e-1, 8.57142857142857142857142857143e-1, 9.45216222272014340129957427739e-1, 1.0e0, 1.0e0}
		c   = [rkLen][rkLen]float64{
			1:  {0: 2e-4},
			2:  {2.66666666666666666666666666667e-4, 5.33333333333333333333333333333e-4},
			3:  {2.91666666666666666666666666667e-3, -4.16666666666666666666666666667e-3, 6.25e-3},
			4:  {1.64609053497942386831275720165e-3, 0, 5.48696844993141289437585733882e-3, 1.75582990397805212620027434842e-3},
			5:  {1.9456e-3, 0, 7.15174603174603174603174603175e-3, 2.91271111111111111111111111111e-3, 7.89942857142857142857142857143e-4},
			6:  {5.6640625e-4, 0, 8.80973048941798941798941798942e-4, -4.36921296296296296296296296296e-4, 3.39006696428571428571428571429e-4, -9.94646990740740740740740740741e-5},
			7:  {3.08333333333333333333333333333e-3, 0, 0, 1.77777777777777777777777777778e-3, 2.7e-3, 1.57828282828282828282828282828e-3, 1.08606060606060606060606060606e-2},
			8:  {3.65183937480112971375119150338e-3, 0, 3.96517171407234306617557289807e-3, 3.19725826293062822350093426091e-3, 8.22146730685543536968701883401e-3, -1.31309269595723798362013884863e-3, 9.77158696806486781562609494147e-3, 3.75576906923283379487932641079e-3},
			9:  {3.70724106871850081019565530521e-3, 0, 5.08204585455528598076108163479e-3, 1.17470800217541204473569104943e-3, -2.11476299151269914996229766362e-2, 6.01046369810788081222573525136e-2, 2.01057347685061881846748708777e-2, -2.83507501229335808430366774368e-2, 1.48795689185819327555905582479e-2},
			10: {3.51253765607334415311308293052e-2, 0, -8.61574919513847910340576078545e-3, -5.79144805100791652167632252471e-3, 1.94555482378261584239438810411e0, -3.43512386745651359636787167574e0, -1.09307011074752217583892572001e-1, 2.3496383118995166394320161088e0, -7.56009408687022978027190729778e-1, 1.09528972221569264246502018618e-1},
			11: {2.05277925374824966509720571672e-2, 0, -7.28644676448017991778247943149e-3, -2.11535560796184024069259562549e-3, 9.27580796872352224256768033235e-1, -1.65228248442573667907302673325e0, -2.10795630056865698191914366913e-2, 1.20653643262078715447708832536e0, -4.13714477001066141324662463645e-1, 9.07987398280965375956795739516e-2, 5.35555260053398504916870658215e-3},
			12: {-1.43240788755455150458921091632e-1, 0, 1.25287037730918172778464480231e-2, 6.82601916396982712868112411737e-3, -4.79955539557438726550216254291e0, 5.69862504395194143379169794156e0, 7.55343036952364522249444028716e-1, -1.27554878582810837175400796542e-1, -1.96059260511173843289133255423e0, 9.18560905663526240976234285341e-1, -2.38800855052844310534827013402e-1, 1.59110813572342155138740170963e-1},
			13: {8.04501920552048948697230778134e-1, 0, -1.66585270670112451778516268261e-2, -2.1415834042629734811731437191e-2, 1.68272359289624658702009353564e1, -1.11728353571760979267882984241e1, -3.37715929722632374148856475521e0, -1.52433266553608456461817682939e1, 1.71798357382154165620247684026e1, -5.43771923982399464535413738556e0, 1.38786716183646557551256778839e0, -5.92582773265281165347677029181e-1, 2.96038731712973527961592794552e-2},
			14: {-9.13296766697358082096250482648e-1, 0, 2.41127257578051783924489946102e-3, 1.76581226938617419820698839226e-2, -1.48516497797203838246128557088e1, 2.15897086700457560030782161561e0, 3.99791558311787990115282754337e0, 2.84341518002322318984542514988e1, -2.52593643549415984378843352235e1, 7.7338785423622373655340014114e0, -1.8913028948478674610382580129e0, 1.00148450702247178036685959248e0, 4.64119959910905190510518247052e-3, 1.12187550221489570339750499063e-2},
			15: {-2.75196297205593938206065227039e-1, 0, 3.66118887791549201342293285553e-2, 9.7895196882315626246509967162e-3, -1.2293062345886210304214726509e1, 1.42072264539379026942929665966e1, 1.58664769067895368322481964272e0, 2.45777353275959454390324346975e0, -8.93519369440327190552259086374e0, 4.37367273161340694839327077512e0, -1.83471817654494916304344410264e0, 1.15920852890614912078083198373e0, -1.72902531653839221518003422953e-2, 1.93259779044607666727649875324e-2, 5.20444293755499311184926401526e-3},
			16: {1.30763918474040575879994562983e0, 0, 1.73641091897458418670879991296e-2, -1.8544456454265795024362115588e-2, 1.48115220328677268968478356223e1, 9.38317630848247090787922177126e0, -5.2284261999445422541474024553e0, -4.89512805258476508040093482743e1, 3.82970960343379225625583875836e1, -1.05873813369759797091619037505e1, 2.43323043762262763585119618787e0, -1.04534060425754442848652456513e0, 7.17732095086725945198184857508e-2, 2.16221097080827826905505320027e-3, 7.00959575960251423699282781988e-3, 0},
		}
		// low order b.
		b = [rkLen]float64{1.70087019070069917527544646189e-2, 0.0e0, 0.0e0, 0.0e0, 0.0e0, 0.0e0, 7.22593359308314069488600038463e-2, 3.72026177326753045388210502067e-1, -4.01821145009303521439340233863e-1, 3.35455068301351666696584034896e-1, -1.31306501075331808430281840783e-1, 1.89431906616048652722659836455e-1, 2.68408020400290479053691655806e-2, 1.63056656059179238935180933102e-2, 3.79998835669659456166597387323e-3, 0.0e0, 0.0e0}
		_ = b
		//high order b (bhat)
		bhat = [rkLen]float64{1.21278685171854149768890395495e-2, 0.0e0, 0.0e0, 0.0e0, 0.0e0, 0.0e0, 8.62974625156887444363792274411e-2, 2.52546958118714719432343449316e-1, -1.97418679932682303358307954886e-1, 2.03186919078972590809261561009e-1, -2.07758080777149166121933554691e-2, 1.09678048745020136250111237823e-1, 3.80651325264665057344878719105e-2, 1.16340688043242296440927709215e-2, 4.65802970402487868693615238455e-3, 0.0e0, 0.0e0}
	)

	adaptive := sim.Algorithm.Error.Max > 0 && sim.Algorithm.Step.Min > 0 && sim.Algorithm.Step.Max > sim.Algorithm.Step.Min
	states := make([]state.State, sim.Algorithm.Steps+1)
	h := sim.Dt() / float64(sim.Algorithm.Steps)
	states[0] = sim.State.Clone()
	var k [rkLen]state.State
	var snext, err10_12 state.State
	for i := 0; i < len(states)-1; i++ {
		// create auxiliary states for calculation

		t := states[i].Time()

		err10_12 = states[i].CloneBlank(t + h)
		snext = states[i].CloneBlank(t + h)
		for ord := 1; ord < rkLen; ord++ {
			k[ord] = states[i].CloneBlank(t + cx0[ord]*h)
		}

		k[0] = StateDiff(sim.Diffs, states[i])
		state.Scale(h, k[0])

		for ord := 1; ord < rkLen; ord++ {
			state.AddScaledTo(k[ord], states[i], c[ord][0], k[0])
			for j := 1; j < ord; j++ {
				state.AddScaled(k[ord], c[ord][j], k[j])
			}
			k[ord] = StateDiff(sim.Diffs, k[ord])
			state.Scale(h, k[ord])
		}

		// tenth order approximation calc.
		state.AddScaledTo(snext, states[i], b[0], k[0])
		for ord := 1; ord < rkLen; ord++ {
			state.AddScaled(snext, b[ord], k[ord])
		}

		// assign solution
		states[i+1] = snext.Clone()
		// Adaptive timestep block. Modify step length if necessary
		if false && adaptive {
			state.AddScaledTo(err10_12, states[i], bhat[0], k[0])
			for ord := 1; ord < rkLen; ord++ {
				state.AddScaled(snext, bhat[ord], k[ord])
			}
			state.Sub(err10_12, snext)
			errRatio := sim.Algorithm.Error.Max / state.Max(err10_12)
			scale := 0.8 * math.Abs(math.Pow(errRatio, 1./11.))
			scale = math.Min(math.Max(scale, .125), 4.0)
			hnew := math.Min(math.Max(h*scale, sim.Algorithm.Step.Min), sim.Algorithm.Step.Max)
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
