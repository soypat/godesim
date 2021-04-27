[![Go Report Card](https://goreportcard.com/badge/github.com/soypat/godesim)](https://goreportcard.com/report/github.com/soypat/godesim)
[![go.dev reference](https://pkg.go.dev/badge/github.com/soypat/godesim)](https://pkg.go.dev/github.com/soypat/godesim)
[![codecov](https://codecov.io/gh/soypat/godesim/branch/main/graph/badge.svg)](https://codecov.io/gh/soypat/godesim/branch/main)

# godesim

Simulate complex systems with a simple API.
---

Wrangle non-linear differential equations while writing maintainable, simple code.

## Why Godesim?

ODE solvers seem to fill the niche of simple system solvers in
your numerical packages such as scipy's odeint/solve_ivp. 

Among these integrators there seems to be room for a solver that offers simulation interactivity such as modifying
the differential equations during simulation based on events such as a rocket stage separation.

## Installation

Requires Go.

```console
go get github.com/soypat/godesim
```

## Progress

Godesim is in early development and will naturally change as it is used more.
 The chart below shows some features that are planned or already part of godesim.

| Status legend | Planned | Started | Prototype | Stable | Mature |
| ------------- |:-------:|:-------:|:---------:|:------:|:------:|
| Legend symbol |    ✖️    |    🏗️   |     🐞️    |   🚦️   |   ✅️   |

| Features | Status | Notes |
| -------- |:------:| ----- |
| Non-linear solvers | 🚦️ | Suite of ODE solvers available. |
| Non-autonomous support | 🚦️ | `U` vector which need not a defined differential equation like `X` does.|
| Event driver | 🚦️ | Eventer interface implemented. |
| Stiff solver | 🚦️ | Newton-Raphson algorithm implemented and tested. |

<details><summary>Algorithms available and benchmarks</summary>

| Algorithm         |   Time/Operation| Memory/Op     | Allocations/Op    |
|-------------------|-----------------|---------------|-------------------|
|RK4             	|    1575 ns/op	  |   516 B/op	  |    12 allocs/op   |
|RK5             	|    2351 ns/op	  |   692 B/op	  |    21 allocs/op   |
|RKF45          	|    3229 ns/op	  |   780 B/op	  |    25 allocs/op   |
|Newton-Raphson     |    8616 ns/op	  |  4292 B/op	  |    92 allocs/op   |
|Dormand-Prince   	|    4365 ns/op	  |   926 B/op	  |    32 allocs/op   |

</details>

## [Examples](./_examples)

### Quadratic Solution

```go
// Declare your rate-of-change functions using state-space symbols
Dtheta := func(s state.State) float64 {
	return s.X("theta-dot")
}

DDtheta := func(s state.State) float64 {
    return 1
}
// Set the Simulation's differential equations and initial values and hit Begin!
sim := godesim.New() // Configurable with Simulation.SetConfig(godesim.Config{...})
sim.SetDiffFromMap(map[state.Symbol]state.Diff {
    "theta":  Dtheta,
    "theta-dot": DDtheta,
})
sim.SetX0FromMap(map[state.Symbol]float64{
    "theta":  0,
    "theta-dot": 0,
})
sim.SetTimespan(0.0, 1.0, 10) // One second simulated
sim.Begin()
```

The above code solves the following system:

![](_assets/quadratic_eq.png)

for the domain `t=0` to `t=1.0` in 10 steps where `theta` and `theta-dot` are the `X` variables. The resulting curve is quadratic as the solution for this equation (for theta and theta-dot equal to zero) is

![](_assets/quadratic_eq_sol.png)

### How to obtain results
```go
// one can then obtain simulation results as float slices 
t := sim.Results("time")
theta := sim.Results("theta")
```


### Other examples

To run an example, navigate to it's directory (under [`examples`](./_examples)) then type `go run .` in console.

There are three simple examples which have been cooked up and left in `_examples` directory.
I've been having problems running Pixel on my machine so the simulation animations are still under work.

* [Simple pendulum](./_examples/simplePendulum)
* [Double pendulum exhibiting chaotic motion](./_examples/doublePendulum)
* [N-Body simulation](./_examples/n-body)

## Contributing

Pull requests welcome!

This is my first library written for any programming language ever. I'll try to be fast on replying to pull-requests and issues. 

