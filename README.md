# Simulation

Simulate complex systems with a simple API.
---

Wrangle non-linear differential equations while writing maintainable, simple code.

### ODE multivariable solver with super simple API


```go
// Declare your rate-of-change functions using state-space symbols
Dtheta := func(s state.State) float64 {
	return s.X("Dtheta")
}

DDtheta := func(s state.State) float64 {
    return 1
}
// Set the simulation's differential equations and initial values and hit Begin!
sim := simulation.New() // Configurable with .SetConfig(simulation.Config{...})
sim.SetChangeMap(map[state.Symbol]state.Changer{
    "theta":  Dtheta,
    "Dtheta": DDtheta,
})
sim.SetX0FromMap(map[state.Symbol]float64{
    "theta":  0,
    "Dtheta": 0,
})
sim.SetTimespan(0.0, 1.0, 10)
sim.Begin()
```

The above code solves the following differential equations:

![](_assets/quadratic_eq.png)

for the domain t=0 to t=1s in 10 steps where theta and theta-dot are the `X` variables. The resulting curve is quadratic as the solution for this equation (for theta and theta-dot equal to zero) is

![](_assets/quadratic_eq_sol.png)

### How to obtain results
```go
// one can then obtain simulation results as float slices 
t := sim.Results("time")
theta := sim.Results("theta")
```