package roguelike

import (
	"math"

	"github.com/espresso20/ageforge/mapmodel"
)

// space_quantum.go is the Quantum Age: reality bends. The starbase is
// still there, but it repeats smaller inside itself, three times over, down
// to the anchor at its heart. Every structure is in superposition: its
// cells flicker between two states in slow bands that sweep across the
// plane, and colour runs along them iridescent. Probability clouds hang
// round the station, their interference fringes sliding. Behind it all, a
// faint echo of the old town on the ground, where every building once
// stood. Phase ships tunnel through, here and not here.

// Quantum layout, in plane cells.
const (
	qX, qY         = skyW / 2, 21
	qInRX, qInRY   = 13.5, 5.0
	qOutRX, qOutRY = 26.0, 9.0
)

// layQuantum lays the Quantum Age's base.
func (b *skyBase) layQuantum() {
	b.recursion()
	b.clouds()
}

// recursion draws the starbase and its smaller selves: each level is the
// same shape, a habitat ring, a docking ring and six pylons, scaled down to
// sit inside the level above's habitat ring. The outermost holds the
// player's buildings (laid out as the Galactic Age's base was); the inner
// levels are structure only.
func (b *skyBase) recursion() {
	cx, cy := float64(qX), float64(qY)
	fr := b.frame("Superposed starbase", "it is also, smaller, inside itself", slStarbase, "")
	used := map[mapmodel.Pt]bool{}
	ring := func(rx, ry float64, ink mapmodel.SkyInk) map[mapmodel.Pt]bool {
		ps, rs := smoothRing(cx, cy, rx, ry)
		on := map[mapmodel.Pt]bool{}
		for i, p := range ps {
			b.put(p.X, p.Y, skyCell{r: rs[i], alt: dashed(rs[i]), ink: ink, lv: 2, k: skFrame, ref: fr, fx: fxFlicker})
			on[p], used[p] = true, true
		}
		return on
	}
	pylons := func(rx, ry float64, n int, outer map[mapmodel.Pt]bool) (tips []mapmodel.Pt, cells []mapmodel.Pt) {
		for _, d := range [6][2]int{{0, -1}, {1, -1}, {1, 1}, {0, 1}, {-1, 1}, {-1, -1}} {
			p := pt(qX, qY)
			for inEllipse(p.X, p.Y, cx, cy, rx, ry) || outer[p] {
				p = pt(p.X+d[0]*(1+abs(d[1])*abs(d[0])), p.Y+d[1])
			}
			r := '│'
			if d[0]*d[1] < 0 {
				r = '╱'
			} else if d[0] != 0 {
				r = '╲'
			}
			for i := 0; i < n; i++ {
				if b.in(p.X, p.Y) && !used[p] {
					b.put(p.X, p.Y, skyCell{r: r, alt: '·', ink: mapmodel.InkFrameDim, lv: 2, k: skFrame, ref: fr, fx: fxFlicker})
					used[p] = true
					cells = append(cells, p)
				}
				if i == n-1 {
					tips = append(tips, p)
				}
				p = pt(p.X+d[0], p.Y+d[1])
			}
		}
		return tips, cells
	}
	// the outermost level, with the buildings
	inner := ring(qInRX, qInRY, mapmodel.InkFrame)
	outer := ring(qOutRX, qOutRY, mapmodel.InkAccent)
	tips, pyl := pylons(qOutRX, qOutRY, 6, outer)
	// the smaller selves, inside the habitat ring
	for _, k := range []float64{0.44, 0.2} {
		ring(qInRX*k, qInRY*k+0.4, mapmodel.InkFrame)
		out2 := ring(qOutRX*k, qOutRY*k+0.5, mapmodel.InkAccent)
		pylons(qOutRX*k, qOutRY*k+0.5, max(1, int(5*k)), out2)
	}
	b.hubAt(qX, qY)
	b.at(qX, qY).alt = '◈'
	b.at(qX, qY).fx = fxFlicker
	used[pt(qX, qY)] = true
	// the decks between the outer level's rings
	in := map[mapmodel.Pt]bool{}
	for p := range inner {
		in[p] = true
	}
	var decks [3][]mapmodel.Pt
	for k := range decks {
		for _, p := range halo(in, cx, cy, qInRX, qInRY) {
			in[p] = true
			if !used[p] && inEllipse(p.X, p.Y, cx, cy, qOutRX-0.6, qOutRY-0.6) {
				decks[k] = append(decks[k], p)
				used[p] = true
			}
		}
	}
	for i, lin := range galInner {
		for _, d := range decks {
			b.addBand(lin, centreOut(sector(rotate(d, len(d)-len(d)/12), i, len(galInner))))
		}
	}
	out := map[mapmodel.Pt]bool{}
	for p := range outer {
		out[p] = true
	}
	for k := 0; k < 3; k++ {
		var layer []mapmodel.Pt
		for _, p := range halo(out, cx, cy, qOutRX, qOutRY) {
			out[p] = true
			if !used[p] && b.in(p.X, p.Y) {
				layer = append(layer, p)
				used[p] = true
			}
		}
		lins := append(append([]string{}, galOuter...), mapmodel.LinMines, mapmodel.LinMetal, mapmodel.LinWood, mapmodel.LinEnergy)
		for i, lin := range lins {
			b.addBand(lin, centreOut(sector(rotate(layer, len(layer)-len(layer)/18), i, len(lins))))
		}
	}
	var eng, fleet []mapmodel.Pt
	for i, p := range pyl {
		if i%2 == 1 {
			eng = append(eng, p)
		}
	}
	b.addBand(mapmodel.LinEngineer, eng)
	for _, t := range tips {
		for _, o := range [6][2]int{{-2, 0}, {2, 0}, {-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
			if q := pt(t.X+o[0], t.Y+o[1]); b.in(q.X, q.Y) && !used[q] {
				fleet = append(fleet, q)
				used[q] = true
			}
		}
	}
	b.addBand(mapmodel.LinMilitary, fleet)
}

// dashed is a stroke's other state: the same line, unobserved.
func dashed(r rune) rune {
	switch r {
	case '─':
		return '┄'
	case '│':
		return '┆'
	}
	return '·'
}

// clouds hangs probability clouds round the station: fringes where the
// distances to the two ends of the docking ring differ by a whole number of
// wavelengths, sliding as the phase turns.
func (b *skyBase) clouds() {
	cx, cy := float64(qX), float64(qY)
	f1x, f2x := cx-qOutRX*0.8, cx+qOutRX*0.8
	for y := 0; y < skyH; y++ {
		for x := 0; x < skyW; x++ {
			c := b.at(x, y)
			if c.k != skVoid && c.k != skStar {
				continue
			}
			o := sq((float64(x)-cx)/(qOutRX+13)) + sq((float64(y)-cy)/(qOutRY+7))
			if o > 1 || inEllipse(x, y, cx, cy, qOutRX+6, qOutRY+3.5) {
				continue
			}
			dy := 2 * (float64(y) - cy)
			d1 := math.Sqrt(sq(float64(x)-f1x) + dy*dy)
			d2 := math.Sqrt(sq(float64(x)-f2x) + dy*dy)
			ph := int((d1-d2)*1.6) + 320
			*c = skyCell{r: ' ', ink: mapmodel.InkCloud, lv: 1, k: skCloud, fx: fxFringe, ph: uint8(ph % 32)}
		}
	}
}

// quantumLife puts what changes with play on the Quantum Age: the echo of
// the old town (each tile where it stood, faint, behind everything), the
// units' unobserved states, the anchor when it stands, and the traffic.
func (s *skyScene) quantumLife() {
	m := s.m
	w := m.Town.World
	if w != nil {
		for i, t := range m.Town.Tiles {
			x, y := qX+2*(t.X-w.CX), qY+(t.Y-w.CY)
			c := s.cell(x, y)
			if c.k != skVoid && c.k != skStar && c.k != skCloud || inEllipse(x, y, qX, qY, qOutRX+4, qOutRY+3) {
				continue
			}
			r := mapmodel.R(mapmodel.LineageSym(t.Lineage, 5), mapmodel.TierUnicode)
			*c = skyCell{r: r, ink: mapmodel.InkEcho, lv: 1, k: skEcho, ref: int32(i)}
		}
	}
	for i := range s.cells {
		if c := &s.cells[i]; c.k == skUnit {
			c.alt, c.fx = '○', fxFlicker
		}
	}
	for _, wd := range m.Wonders {
		if wd.Key == "reality_anchor" && wd.Built {
			c := s.cell(qX, qY)
			c.sym, c.r, c.alt = mapmodel.SymNone, '✦', '◈'
		}
	}
	busy := m.Activity.Traffic
	for k := 0; k < 2+int(busy*3); k++ {
		h := mapmodel.Hash(s.b.seed, 650, int64(k))
		y0, y1 := int(h%skyH), int(h>>8%skyH)
		lane := lineCells(pt(-2, y0), pt(skyW+1, y1))
		if k%2 == 1 {
			lane = reversed(lane)
		}
		s.addMover(smPhase, lane, k, true, 40+int(h>>16%120))
	}
}
