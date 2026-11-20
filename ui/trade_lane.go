package ui

import "math"

// laneW is the width of a trade lane, in cells.
const laneW = 9

// laneSpeed is how fast goods cross a lane, in cells a second, for an import
// flow of perSec: 1 + 1.5 * log10(1 + flow), held between 1 and 12.
func laneSpeed(perSec float64) float64 {
	l := math.Log10(1 + math.Max(perSec, 0))
	return math.Min(math.Max(1+float64(1.5*l), 1), 12)
}

// laneCell is the glyph of cell i of a lane at clock seconds: a pulse every
// third cell, moving right at speed cells a second. A pulse is kept with
// chance run (judged once, by its number, so it stays whole while it moves),
// which is what leaves gaps in a throttled lane.
func laneCell(i int, clock, speed, run float64) rune {
	m := int(math.Floor(float64(clock*speed))) - i
	if ((m%3)+3)%3 != 0 {
		return '·'
	}
	_, frac := math.Modf(float64(float64(m/3) * 0.6180339887))
	if frac < 0 {
		frac++
	}
	if frac >= run {
		return '·'
	}
	return '●'
}

// laneAt draws a lane at (x, y): goods in transit in k while moving, and a
// still dim row of dots when not.
func laneAt(g *xgrid, x, y int, moving bool, perSec, run, clock float64, k xk) {
	for i := 0; i < laneW; i++ {
		if r := '·'; !moving {
			g.set(x+i, y, r, xDim)
		} else if r = laneCell(i, clock, laneSpeed(perSec)*math.Max(run, 0.1), run); r == '·' {
			g.set(x+i, y, r, xDim)
		} else {
			g.set(x+i, y, r, k)
		}
	}
}
