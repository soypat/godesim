package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/soypat/godesim"
	"github.com/soypat/godesim/state"
)

// Declare simulation constants: gravity and pendulum length
const g, l float64 = 9.8, 1. // m/s2, m
var sin, pi = math.Sin, math.Pi

func main() {
	Dthetadot := func(s state.State) float64 {
		return -g / l * sin(s.X("theta"))
	}

	sim := godesim.New()

	sim.SetChangeMap(map[state.Symbol]state.Changer{
		"theta":     func(s state.State) float64 { return s.X("theta-dot") },
		"theta-dot": Dthetadot,
	})

	sim.SetX0FromMap(map[state.Symbol]float64{
		"theta":     20. * pi / 180., // convert angles to radians
		"theta-dot": 0,
	})

	sim.SetTimespan(0., 8., 100)

	sim.Begin()

	time, theta := sim.Results("time"), sim.Results("theta")
	fmt.Printf("%.2f\n\n%.2f\n", time, theta)
	pixelgl.Run(run)
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1024, 768),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	for !win.Closed() {
		win.SetClosed(win.JustPressed(pixelgl.KeyEscape) || win.JustPressed(pixelgl.KeyQ))

		win.Clear(color.White)

		imd := imdraw.New(nil)
		imd.Color = color.Black
		imd.Push(pixel.V(512, 768/2))
		imd.Circle(20, 0)

		win.Update()
	}
}
