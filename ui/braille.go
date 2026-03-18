package ui

import (
	"fmt"
	"math"
)

// braille dot bit positions: [col: 0=left, 1=right][row: 0=top .. 3=bottom]
var brailleDot = [2][4]rune{
	{0x01, 0x02, 0x04, 0x40},
	{0x08, 0x10, 0x20, 0x80},
}

// BrailleLine renders values as a braille line graph, returning height string rows each width
// braille characters wide. ageCols contains canvas column positions where age-advance markers
// ('│') are drawn in otherwise-empty cells.
func BrailleLine(values []float64, width, height int, ageCols []int) []string {
	empty := make([]string, height)
	if len(values) == 0 {
		return empty
	}

	cW, cH := width*2, height*4
	canvas := make([][]bool, cH)
	for i := range canvas {
		canvas[i] = make([]bool, cW)
	}

	minV, maxV := values[0], values[0]
	for _, v := range values {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	if maxV == minV {
		maxV = minV + 1
	}

	// build set of age marker columns
	ageset := make(map[int]bool, len(ageCols))
	for _, c := range ageCols {
		ageset[c] = true
	}

	// plot a connected line across the canvas
	n := len(values)
	prevX, prevY := -1, -1
	for i, v := range values {
		x := i * (cW - 1) / (n - 1)
		norm := (v - minV) / (maxV - minV)
		y := int((1.0-norm)*float64(cH-1) + 0.5)
		if y < 0 {
			y = 0
		}
		if y >= cH {
			y = cH - 1
		}
		if prevX >= 0 {
			steps := x - prevX
			if steps < 1 {
				steps = 1
			}
			for s := 0; s <= steps; s++ {
				lx := prevX + s
				ly := prevY + (y-prevY)*s/steps
				if lx >= 0 && lx < cW && ly >= 0 && ly < cH {
					canvas[ly][lx] = true
				}
			}
		}
		prevX, prevY = x, y
	}

	lines := make([]string, height)
	for row := 0; row < height; row++ {
		sb := make([]rune, 0, width)
		for col := 0; col < width; col++ {
			var ch rune = 0x2800
			for dr := 0; dr < 4; dr++ {
				for dc := 0; dc < 2; dc++ {
					cr, cc := row*4+dr, col*2+dc
					if cr < cH && cc < cW && canvas[cr][cc] {
						ch |= brailleDot[dc][dr]
					}
				}
			}
			if ageset[col] && ch == 0x2800 {
				sb = append(sb, '│')
			} else {
				sb = append(sb, ch)
			}
		}
		lines[row] = string(sb)
	}
	return lines
}

// AgeMarkerCols converts age marker tick positions to canvas column indices.
func AgeMarkerCols(markerTicks []int, firstTick, lastTick, width int) []int {
	if lastTick <= firstTick {
		return nil
	}
	cols := make([]int, 0, len(markerTicks))
	for _, t := range markerTicks {
		col := (t - firstTick) * (width - 1) / (lastTick - firstTick)
		if col >= 0 && col < width {
			cols = append(cols, col)
		}
	}
	return cols
}

// fmtVal formats a float64 for compact display (K/M/B suffixes).
func fmtVal(v float64) string {
	abs := math.Abs(v)
	if abs >= 1e9 {
		return fmt.Sprintf("%.1fB", v/1e9)
	}
	if abs >= 1e6 {
		return fmt.Sprintf("%.1fM", v/1e6)
	}
	if abs >= 1e3 {
		return fmt.Sprintf("%.1fK", v/1e3)
	}
	return fmt.Sprintf("%.1f", v)
}
