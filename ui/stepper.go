package ui

import (
	"image"

	"github.com/espresso20/ageforge/game"
)

// stepper.go is the minus and plus chips that move one worker in or out of a
// building: the picked building's detail on the dashboard and the People
// panel's work areas draw the same pair.

// stepSide is one chip of a stepper: the command a click runs and whether it
// cannot run now.
type stepSide struct {
	cmd string
	off bool
}

// stepperW is the cells a stepper takes.
const stepperW = 7

// stepper draws the minus and plus chips at (x, y), records what a click on
// each runs and returns the next x.
func (g *xgrid) stepper(x, y int, minus, plus stepSide) int {
	set := g.keyset()
	for i, k := range [2]struct {
		sign string
		side stepSide
	}{{"-", minus}, {"+", plus}} {
		r := image.Rect(x+i*4, y, x+i*4+3, y+1)
		g.put(r.Min.X, y, " "+k.sign+" ", xKeycap)
		set.keys = append(set.keys, xkey{r: r, cap: r, kind: keySteel, cmd: k.side.cmd, off: k.side.off})
	}
	return x + stepperW
}

// buildingStep is the stepper of a building that takes workers, with idle
// people in the reserve: take one out unless none is in, put one in unless it
// is full, nothing is built or nobody is idle.
func buildingStep(key string, bs game.BuildingState, idle int) (minus, plus stepSide) {
	minus = stepSide{command(verbUnassign, key), bs.WorkersAssigned == 0}
	plus = stepSide{command(verbAssign, key), bs.Count == 0 || bs.WorkersAssigned >= bs.WorkerCapacity*bs.Count || idle == 0}
	return minus, plus
}
