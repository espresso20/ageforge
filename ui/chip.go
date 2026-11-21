package ui

import (
	"image"
	"image/color"
	"math"

	"github.com/espresso20/ageforge/window/kit"
)

// chip.go is the one-row button: a label and, when it has one, the key it
// names, on a single plate. key_bar.go holds where it is and what a click
// does, and the command line it may carry; button.go draws it.

// toward is a mixed t of the way from a to b.
func toward(a, b color.RGBA, t float64) color.RGBA {
	t = math.Max(0, math.Min(1, t))
	m := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return color.RGBA{R: m(a.R, b.R), G: m(a.G, b.G), B: m(a.B, b.B), A: 255}
}

// chipRows is the rows a button takes.
const chipRows = 1

// chipSpec is a button: its colour, its label, the key it names (empty for
// none), the command line a click runs (empty to press the key), and whether
// it is unavailable. ask, with yes, is the question a click on a command asks
// first (xkey.ask).
type chipSpec struct {
	kind       keyKind
	label, key string
	cmd        string
	ask, yes   string
	off        bool
	// why is the line that says why it is off (xkey.why).
	why string
}

// chipText is what the plate holds: the label, then the key.
func chipText(label, key string) string {
	if key == "" {
		return " " + label + " "
	}
	return " " + label + "  " + key + " "
}

// chipW is the cells a button takes.
func chipW(label, key string) int { return runeLen(chipText(label, key)) }

// chip records a button at (x, y) in the set, writes its text through put (the
// grid's own way of writing text) and returns the next x.
func (s *keySet) chip(x, y int, a chipSpec, put func(x, y int, text string)) int {
	text := chipText(a.label, a.key)
	r := image.Rect(x, y, x+runeLen(text), y+1)
	put(x, y, text)
	b := xkey{r: r, cap: r, kind: a.kind, cmd: a.cmd, ask: a.ask, yes: a.yes, off: a.off, why: a.why}
	if a.key != "" {
		if parts := capParts(a.key); len(parts) > 0 {
			b.key = parts[0].key
		}
	}
	b.still = b.key == (kit.Key{}) && a.cmd == ""
	s.keys = append(s.keys, b)
	return r.Max.X
}

// chip draws a button on the grid at (x, y) and returns the next x.
func (g *xgrid) chip(x, y int, a chipSpec) int {
	return g.keyset().chip(x, y, a, func(x, y int, text string) { g.put(x, y, text, xKeycap) })
}
