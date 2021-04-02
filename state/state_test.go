package state

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/diff/fd"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
)

func TestEmptyStateError(t *testing.T) {
	sempty := NewFromXMap(map[Symbol]float64{})
	if sempty.XVector() == nil {
		t.Error("vector returned never should be nil")
	}

	syms := []Symbol{"x", "y"}
	consU, consX := sempty.ConsistencyU(syms), sempty.ConsistencyX(syms)
	for i := range syms {
		if !math.IsNaN(consU[i]) || !math.IsNaN(consX[i]) {
			t.Error("consistency should be bad")
		}
	}
	// depends on this creating symbols in ordered fashion (with XEqual)

}
func TestDumbFunctions(t *testing.T) {
	syms := []Symbol{"x", "y"}
	s := orderedState(syms...)
	usyms := []Symbol{"u", "v"}
	for j := range usyms {
		s.UEqual(usyms[j], rand.Float64())
	}

	if s.Len() != 2 {
		t.Error("bad statevar length")
	}
	if s.Time() != 0 {
		t.Error("statevar zero value for time must be 0")
	}
	s.SetTime(1.)
	if s.Time() != 1 {
		t.Error("set time not working")
	}
	sclone := s.CloneBlank(0.5)
	if sclone.Time() != 0.5 {
		t.Error("time cloning not working")
	}
	for _, v := range sclone.XVector() {
		if v != 0 {
			t.Error("cloneblank not set all values to 0")
		}
	}
	for i, sym := range sclone.XSymbols() {
		if sym != syms[i] {
			t.Errorf("symbol ordering changed? expected %s, got %s", syms[i], sym)
		}
	}
	consU, consX := sclone.ConsistencyU(usyms), sclone.ConsistencyX(syms)
	for i := range syms {
		if math.IsNaN(consX[i]) {
			t.Error("consistency should be all numbers")
		}
	}
	for i := range usyms {
		if math.IsNaN(consU[i]) {
			t.Error("consistency should be all numbers")
		}
	}

}

func TestBadSymbol(t *testing.T) {
	s := New()
	err1, err2 := recoverFromSymbol(s.U, "nonexistent"), recoverFromSymbol(s.X, "nonexistent")
	if err1 == nil || err2 == nil {
		t.Error("looking for non existent input/statevar should panic")
	}
}

func recoverFromSymbol(f func(Symbol) float64, sym Symbol) (i interface{}) {
	defer func() {
		i = recover()
	}()
	_ = f(sym)
	return nil
}

func TestStateCreation(t *testing.T) {
	var tests = []struct {
		x []Symbol
		u []Symbol
	}{
		{x: []Symbol{"x", "y", "z"}, u: []Symbol{"u", "v", "w"}},
	}
	for i := range tests {
		s := New()
		n, p := len(tests[i].x), len(tests[i].u)
		xvec, uvec := randVec(n, 2), randVec(p, 2)
		for j := range tests[i].x {
			s.XEqual(tests[i].x[j], xvec[j])
		}
		for j := range tests[i].u {
			s.UEqual(tests[i].u[j], uvec[j])
		}
		// check results
		assertVectorEqual(t, uvec, s.UVector())
		assertVectorEqual(t, xvec, s.XVector())
		assertUSymsOrdered(t, s, uvec, tests[i].u)
		assertUSymsOrdered(t, s, uvec, s.USymbols())
		assertXSymsOrdered(t, s, xvec, tests[i].x)
		// create new vectors
		xvec, uvec = randVec(n, 3), randVec(p, 3)

		for j := range tests[i].x {
			s.XSet(tests[i].x[j], xvec[j])
		}
		for j := range tests[i].u {
			s.USet(tests[i].u[j], uvec[j])
		}
		assertVectorEqual(t, uvec, s.UVector())
		assertVectorEqual(t, xvec, s.XVector())
		assertUSymsOrdered(t, s, uvec, tests[i].u)
		assertUSymsOrdered(t, s, uvec, s.USymbols())
		assertXSymsOrdered(t, s, xvec, tests[i].x)
	}
}

func assertVectorEqual(t *testing.T, want, got []float64) {
	if len(want) != len(got) {
		t.Errorf("length of vectors not equal! want:%g, got:%g", want, got)
		return
	}
	for i := range want {
		if want[i] != got[i] {
			t.Errorf("value at idx %d unequal. want %g, got %g ", i, want[i], got[i])
		}
	}
}
func assertXSymsOrdered(t *testing.T, state State, want []float64, syms []Symbol) {
	for i, sym := range syms {
		if state.X(sym) != want[i] {
			t.Errorf("state var at idx %d, sym %s unequal. want %g, got %g ", i, sym, want[i], state.X(sym))
		}
	}
}
func assertUSymsOrdered(t *testing.T, state State, want []float64, syms []Symbol) {
	for i, sym := range syms {
		if state.U(sym) != want[i] {
			t.Errorf("input at idx %d, sym %s unequal. want %g, got %g ", i, sym, want[i], state.X(sym))
		}
	}
}

func randVec(n int, multiplier float64) []float64 {
	f := make([]float64, n)
	for i := range f {
		f[i] = rand.Float64() * multiplier
	}
	return f
}

func TestJac(t *testing.T) {
	var jacSettings *fd.JacobianSettings = nil
	var tests = []struct {
		x0 []float64
		f  []func([]float64) float64
	}{
		{x0: []float64{1, 1},
			f: []func([]float64) float64{
				func(f []float64) float64 { return 1 },
				func(f []float64) float64 { return 1 },
			},
		},
		{x0: []float64{1, 2, 3},
			f: []func([]float64) float64{
				func(f []float64) float64 { return 10 },
				func(f []float64) float64 { return 20 },
				func(f []float64) float64 { return -10 },
			},
		},
	}
	for i := range tests {
		n := len(tests[i].x0)
		syms := numberedSymbols(n)
		s := randomState(syms...)
		d := make(Diffs, n)
		for j := range d {
			jloop := j // escape loopvar for closure
			diff := func(st State) float64 {
				return tests[i].f[jloop](st.x)
			}
			d[j] = diff
		}
		s.x = tests[i].x0 // dont care about testing random data since jacobian may be singular and thats life
		ms := mat.NewDense(n, n, nil)
		Jacobian(ms, d, s, jacSettings)

		fm := mat.Formatted(ms)
		jStateString := fmt.Sprintf("%v", fm)
		f := func(y, x []float64) {
			for j := range x {
				y[j] = tests[i].f[j](x)
			}
		}
		mfd := mat.NewDense(n, n, nil)
		fd.Jacobian(mfd, f, tests[i].x0, jacSettings)
		fmfd := mat.Formatted(mfd)
		jFDString := fmt.Sprintf("%v", fmfd)
		if jStateString != jFDString || jStateString == "" {
			t.Errorf("jacobians not equal:\n%v\n!=\n%v", jStateString, jFDString)
		}
	}
}

func numberedSymbols(n int) []Symbol {
	s := make([]Symbol, n)
	for i := range s {
		s[i] = Symbol(fmt.Sprintf("x%d", i))
	}
	return s
}
func orderedState(syms ...Symbol) State {
	// TODO, randomize this further
	s := New()
	for _, sym := range syms {
		s.XEqual(sym, rand.Float64()-2*rand.Float64()+-500*rand.Float64()+500*rand.Float64())
	}
	return s
}
func randomState(syms ...Symbol) State {
	// TODO, randomize this further
	m := make(map[Symbol]float64)
	for _, sym := range syms {
		m[sym] = rand.Float64() - 2*rand.Float64() + -500*rand.Float64() + 500*rand.Float64()
	}
	return NewFromXMap(m)
}

func TestArithmetic(t *testing.T) {
	var testsS2 = []struct {
		gonumF func(x, y []float64)
		stateF func(x, y State)
	}{
		{gonumF: floats.Mul, stateF: Mul},
		{gonumF: floats.Div, stateF: Div},
		{gonumF: floats.Add, stateF: Add},
		{gonumF: floats.Sub, stateF: Sub},
	}
	var testsS3to = []struct {
		gonumF func(x, y, z []float64) []float64
		stateF func(x, y, z State) State
	}{
		{gonumF: floats.AddTo, stateF: AddTo},
		{gonumF: floats.SubTo, stateF: SubTo},
		{gonumF: floats.DivTo, stateF: DivTo},
		{gonumF: floats.MulTo, stateF: MulTo},
	}
	var testsCs = []struct {
		gonumF func(c float64, dst []float64)
		stateF func(c float64, x State)
	}{
		{gonumF: floats.AddConst, stateF: AddConst},
		{gonumF: floats.Scale, stateF: Scale},
	}
	var testsScs = []struct {
		gonumF func(dst []float64, c float64, x []float64)
		stateF func(dst State, c float64, x State)
	}{
		{gonumF: floats.AddScaled, stateF: AddScaled},
	}
	var testsS2csto = []struct {
		gonumF func(dst, y []float64, c float64, x []float64) []float64
		stateF func(dst, y State, c float64, x State) State
	}{
		{gonumF: floats.AddScaledTo, stateF: AddScaledTo},
	}
	var testsS = []struct {
		gonumF func(x []float64) float64
		stateF func(x State) float64
	}{
		{gonumF: floats.Max, stateF: Max},
		{gonumF: floats.Min, stateF: Min},
	}
	var testsSc1 = []struct {
		gonumF func(x []float64, L float64) float64
		stateF func(x State, L float64) float64
	}{
		{gonumF: floats.Norm, stateF: Norm},
	}
	var testsScsto = []struct {
		gonumF func(x []float64, c float64, y []float64) []float64
		stateF func(x State, L float64, y State) State
	}{
		{gonumF: floats.ScaleTo, stateF: ScaleTo},
	}
	s1 := randomState("x", "y", "z", "xyz")
	s2 := randomState("x", "y", "z", "xyz")

	// stateRes := s1.Clone()
	vec1 := s1.XVector()
	vec2 := s2.XVector()
	// sresult := make([]float64, len(vec1))
	// vresult := make([]float64, len(vec1))

	for _, test := range testsS2 {
		test.gonumF(vec1, vec2)
		test.stateF(s1, s2)
		assertSliceStateEqual(t, s1, vec1)
	}

	for _, test := range testsS {
		test.gonumF(vec1)
		test.stateF(s1)
		assertSliceStateEqual(t, s1, vec1)
	}

	for i, test := range testsCs {
		c := float64(i + 2)
		test.gonumF(c, vec2)
		test.stateF(c, s2)
		assertSliceStateEqual(t, s2, vec2)
	}
	for i, test := range testsS2csto {
		c := float64(i + 2)
		vr := test.gonumF(vec1, vec1, c, vec2)
		sr := test.stateF(s1, s1, c, s2)
		assertSliceStateEqual(t, s1, vec1)
		assertSliceStateEqual(t, sr, vr)
	}
	for _, test := range testsS3to {
		vr := test.gonumF(vec2, vec1, vec2)
		sr := test.stateF(s2, s1, s2)
		assertSliceStateEqual(t, s2, vec2)
		assertSliceStateEqual(t, sr, vr)
	}
	for i, test := range testsSc1 {
		c := float64(i + 2)
		vr := test.gonumF(vec2, c)
		sr := test.stateF(s2, c)
		// TODO check out this NaN case
		if vr != sr && !math.IsNaN(vr) {
			t.Errorf("two float64 results not equal (state:%f == %f (vec)", sr, vr)
		}
		assertSliceStateEqual(t, s2, vec2)
	}
	for i, test := range testsScsto {
		c := float64(i + 2)
		vr := test.gonumF(vec2, c, vec1)
		sr := test.stateF(s2, c, s1)
		assertSliceStateEqual(t, s2, vec2)
		assertSliceStateEqual(t, sr, vr)
	}
	for i, test := range testsScs {
		c := float64(i + 2)
		test.gonumF(vec2, c, vec1)
		test.stateF(s2, c, s1)
		assertSliceStateEqual(t, s2, vec2)
	}
	s1.x[0] = -20
	Abs(s1)
	for _, v := range s1.x {
		if v < 0 {
			t.Errorf("absolute value fail, got %f", v)
		}
	}
}

func assertSliceStateEqual(t *testing.T, s State, fs []float64) {
	if !compareStateSlice(s, fs) {
		t.Errorf("expected vec %f to be equal to %f", fs, s.x)
	}
}

func compareStateSlice(s State, fs []float64) bool {
	for i, v := range s.XVector() {
		if v == fs[i] {
			return true
		}
	}
	return false
}
