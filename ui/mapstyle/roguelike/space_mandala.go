package roguelike

import (
	"strconv"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// space_mandala.go is the Transcendent Age: beyond matter. The map becomes
// a slowly breathing mandala of light on deep indigo, built from the
// civilization's own history: one ring for every era, the Stone Era
// innermost, the Cosmic Era (to the Quantum Age) outermost, each carrying a
// mark for every kind of building the player raised in its ages, in the
// symbol it had then, and a star for each wonder. Twelve faint spokes join
// the rings; the crown of the age's own buildings rings the core. The light
// breathes outward from the centre, a slow wave; a mote or two drifts home.
// Nothing else moves.

// Mandala layout.
const (
	mdX, mdY    = skyW / 2, 21
	mdRing0     = 2.0  // the innermost ring's radius in rows
	mdRingStep  = 2.55 // and how much each ring adds
	mdAspect    = 2.05 // columns per row
	mdBreath    = 96   // frames per breath (about 12 seconds)
	mdSpokeRing = 1    // the spokes start at this ring
)

// mandalaRing is one ring's cells, clockwise from the top.
type mandalaRing struct {
	cells []mapmodel.Pt
}

// layMandala lays the rings, the spokes and the core's glow.
func (b *skyBase) layMandala() {
	cx, cy := float64(mdX), float64(mdY)
	n := 7
	b.rings = make([]mandalaRing, n)
	on := map[mapmodel.Pt]bool{}
	for k := 0; k < n; k++ {
		ry := mdRing0 + mdRingStep*float64(k)
		ps, rs := smoothRing(cx, cy, mdAspect*ry, ry)
		b.rings[k].cells = ps
		for i, p := range ps {
			b.put(p.X, p.Y, skyCell{r: rs[i], ink: mapmodel.InkFrameDim, lv: 1, k: skRing, ref: int32(k), fx: fxBreathe,
				ph: uint8(k)})
			on[p] = true
		}
	}
	// twelve spokes from the second ring out, faint, breathing with their
	// distance from the core
	inner := mdRing0 + mdRingStep*mdSpokeRing
	outer := mdRing0 + mdRingStep*float64(n-1)
	for s := 0; s < 12; s++ {
		t := float64(s) / 12
		for d := inner; d < outer; d += 0.5 {
			p := pt(int(cx+mdAspect*d*mapmodel.Sin(t)+0.5), int(cy-d*mapmodel.Cos(t)+0.5))
			if c := b.at(p.X, p.Y); !on[p] && (c.k == skVoid || c.k == skStar) {
				k := int((d - mdRing0) / mdRingStep)
				b.put(p.X, p.Y, skyCell{r: '·', ink: mapmodel.InkFrameDim, lv: 0, k: skRing, ref: int32(k), fx: fxBreathe,
					ph: uint8(k)})
			}
		}
	}
	for _, p := range discCells(cx, cy, 2.6, 1.1) {
		if c := b.at(p.X, p.Y); !on[p] {
			c.bg, c.soft = mapmodel.InkHaze, true
		}
	}
	b.hub = pt(mdX, mdY)
	b.put(mdX, mdY, skyCell{r: '✦', ink: mapmodel.InkGlow, lv: 3, k: skCore, bold: true, fx: fxBreathe, ph: 0})
}

// mandala puts the history on the rings: each era's marks spread evenly
// round its ring (as many as the ring has room for, the most prominent
// first), the wonders as stars, the crown round the core.
func (s *skyScene) mandala() {
	m, b := s.m, s.b
	rings := m.Mandala()
	s.marks = make([][]mapmodel.MandalaMark, len(b.rings))
	for k, r := range rings {
		if k >= len(b.rings) {
			break
		}
		cells := b.rings[k].cells
		marks := r.Marks
		if len(marks) > len(cells) {
			marks = marks[:len(cells)]
		}
		s.marks[k] = marks
		for j, mk := range marks {
			p := cells[(j*len(cells)+len(cells)/2)/max(1, len(marks))%len(cells)]
			c := skyCell{sym: mk.Sym, ink: mapmodel.InkLight, lv: 2, k: skMark, ref: int32(k<<10 | j), fx: fxBreathe,
				ph: uint8(k)}
			if mk.Wonder {
				c.ink, c.bold = mapmodel.InkAccent, true
			}
			*s.cell(p.X, p.Y) = c
			if s.anchor[mk.Key] == (mapmodel.Pt{}) {
				s.anchor[mk.Key] = p
			}
		}
		if e := r.Epoch; e < len(m.Catalog.EpochName) && len(cells) > 0 {
			right := cells[0]
			for _, p := range cells {
				if p.X > right.X {
					right = p
				}
			}
			s.labels = append(s.labels, skyLabel{p: right, text: m.Catalog.EpochName[e], ink: mapmodel.InkFrame, far: true})
		}
	}
	s.crown = m.Crown()
	// the core's light grows with everything built in this last age
	total := 0
	for _, mk := range s.crown {
		total += mk.Count
	}
	if total > 0 {
		r := min(3.6, 0.8+0.32*mapmodel.Log2(1+float64(total)))
		for _, p := range discCells(float64(mdX), float64(mdY), mdAspect*r, r) {
			c := s.cell(p.X, p.Y)
			c.bg, c.soft = mapmodel.InkAccent, true
			if c.k == skVoid || c.k == skStar {
				c.k, c.r = skCore, ' '
			}
		}
	}
	for j, mk := range s.crown {
		if j >= len(mdCrown) {
			break
		}
		p := pt(mdX+mdCrown[j][0], mdY+mdCrown[j][1])
		c := s.cell(p.X, p.Y)
		*c = skyCell{sym: mk.Sym, ink: mapmodel.InkAccent, lv: 3, k: skMark, ref: int32(1023<<10 | j),
			fx: fxBreathe, bold: true, bg: c.bg, soft: c.soft}
		if s.anchor[mk.Key] == (mapmodel.Pt{}) {
			s.anchor[mk.Key] = p
		}
	}
	for _, w := range m.Wonders {
		if w.Key == "singularity_core" && w.Built {
			c := s.cell(mdX, mdY)
			c.r, c.sym = '☼', mapmodel.SymNone
		}
	}
}

// mdCrown are the crown's places round the core.
var mdCrown = [8][2]int{{-2, 0}, {2, 0}, {0, -1}, {0, 1}, {-3, -1}, {3, -1}, {-3, 1}, {3, 1}}

// mandalaLife sets the motes drifting home along the spokes.
func (s *skyScene) mandalaLife() {
	cx, cy := float64(mdX), float64(mdY)
	outer := mdRing0 + mdRingStep*6 + 1
	for i := 0; i < 2; i++ {
		t := mapmodel.HashF(s.b.seed, 660, int64(i))
		from := pt(int(cx+mdAspect*outer*mapmodel.Sin(t)), int(cy-outer*mapmodel.Cos(t)))
		lane := lineCells(from, pt(mdX, mdY))
		lane = lane[:max(2, len(lane)-2)]
		s.addMover(smMote, lane, i, true, 40+int(mapmodel.Hash(s.b.seed, 661, int64(i))%80))
	}
}

// breathe is the slow wave of light: brightest at the core, rolling out
// ring by ring, once every mdBreath frames.
func (v *skyView) breathe(s *skyScene, c skyCell, l look, x, y int) look {
	ph := float64(v.anim)/mdBreath - float64(c.ph)/7
	b := 0.5 + 0.5*mapmodel.Sin(ph)
	switch c.k {
	case skRing:
		l.lv = uint8(int(b*2.2)) + c.lv/2
		if l.lv > 3 {
			l.lv = 3
		}
		if b > 0.85 && c.r != '·' {
			l.ink = mapmodel.InkFrame
		}
	case skMark:
		l.lv = 2
		if b > 0.6 {
			l.lv = 3
		}
		if b > 0.8 && l.ink == mapmodel.InkLight {
			l.ink = mapmodel.InkAccent
		}
	case skCore:
		l.lv, l.bold = 3, b > 0.3
	}
	return l
}

// describeMark is a mark on a ring (a building type the player raised in
// that era, or a wonder), a mark of the crown, or the core.
func (v *skyView) describeMark(s *skyScene, c skyCell, p mapmodel.Pt) (mapstyle.Inspection, bool) {
	m := s.m
	if c.k == skCore {
		in := mapstyle.Inspection{Title: "The core", Command: m.SquareCommand(),
			Lines: []string{"every age you lived through, in light", plural(m.Workers.Pop, "person", "people")}}
		return in, true
	}
	ring, j := int(c.ref>>10), int(c.ref&1023)
	var mk mapmodel.MandalaMark
	switch {
	case ring == 1023 && j < len(s.crown):
		mk = s.crown[j]
	case ring < len(s.marks) && j < len(s.marks[ring]):
		mk = s.marks[ring][j]
	default:
		return mapstyle.Inspection{}, false
	}
	if mk.Wonder {
		for i := range m.Wonders {
			if m.Wonders[i].Key == mk.Key {
				return describeSkyWonder(m, &m.Wonders[i]), true
			}
		}
	}
	b := m.Building(mk.Key)
	if b == nil {
		return mapstyle.Inspection{}, false
	}
	age := ""
	if mk.Age >= 0 && mk.Age < len(m.Catalog.AgeNames) {
		age = "raised in the " + m.Catalog.AgeNames[mk.Age]
	}
	in := mapstyle.Inspection{Title: b.Name + " ×" + strconv.Itoa(b.Count), Command: m.BuildingCommand(b),
		Lines: []string{lineageLabel(mk.Lineage), age}}
	if in.Command == "" {
		in.Command = mapmodel.CmdStatus
	}
	return in, true
}
