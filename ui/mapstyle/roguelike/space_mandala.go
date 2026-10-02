package roguelike

import (
	"strconv"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// space_mandala.go is the Transcendent Age: beyond matter. The map becomes
// a mandala of light on deep indigo, built from the civilization's own
// history (the geometry is mapstyle.LayMandala, which the skyline draws
// too): a bright core; a crown of eight petals for the age's own
// buildings; and one ring for every era the player passed through, the
// Stone Era innermost, each a dotted line strung with its era's glyph in
// its era's muted colour. It is strictly symmetric about the core, with
// dark space between the rings. The light breathes outward from the core,
// ring by ring, about every twelve seconds; a mote or two drifts in from
// the dark and settles at the rim. Nothing else moves.

// Mandala layout on the plane.
const (
	mdX, mdY = skyW / 2, 21 // the core
	mdRY     = 20           // rows the mandala may reach above and below it
	mdRX     = 60           // and columns either side
	mdPetal  = 1023         // the ring number a petal's ref carries
)

// layMandala is the mandala's seed-only layer: a sparse starfield and the
// core as the hub. The rings follow the eras reached, so the scene lays
// them (mandala).
func (b *skyBase) layMandala() {
	b.hub = pt(mdX, mdY)
}

// mandala lays the mandala on the scene: the rings of the eras reached,
// the crown and the core, on pure indigo.
func (s *skyScene) mandala() {
	m := s.m
	s.rings = m.Mandala()
	s.crown = m.Crown()
	g := mapstyle.LayMandala(len(s.rings), mdRY, mdRX)
	s.md = g
	off := len(s.rings) - g.Rings // rings that did not fit: the oldest go
	// the stars keep to the dark round the mandala
	ext := g.Extent + 1.5
	for _, p := range discCells(mdX, mdY, mapstyle.MandalaAspect*ext, ext) {
		if c := s.cell(p.X, p.Y); c.k == skStar {
			*c = skyCell{r: ' ', k: skVoid}
		}
	}
	// the core's light grows with everything built in this last age
	total := 0
	for _, mk := range s.crown {
		total += mk.Count
	}
	if total > 0 && g.PetalR > 0 {
		r := min(g.PetalR+0.6, 0.8+0.32*mapmodel.Log2(1+float64(total)))
		for _, p := range discCells(mdX, mdY, mapstyle.MandalaAspect*r, r) {
			c := s.cell(p.X, p.Y)
			c.bg, c.soft = mapmodel.InkAccent, true
		}
	}
	singularity := false
	for _, w := range m.Wonders {
		singularity = singularity || w.Key == "singularity_core" && w.Built
	}
	for _, mc := range g.Cells {
		p := pt(mdX+mc.DX, mdY+mc.DY)
		c := s.cell(p.X, p.Y)
		bg, soft := c.bg, c.soft
		switch mc.Part {
		case mapstyle.MdCore:
			*c = skyCell{sym: mapmodel.SymCore, ink: mapmodel.InkGlow, lv: 3, k: skCore, bold: true, fx: fxBreathe}
			if singularity {
				c.sym, c.r = mapmodel.SymNone, '☼'
			}
		case mapstyle.MdPetal:
			*c = skyCell{sym: mapmodel.SymPetal, ink: mapmodel.InkAccent, lv: 2, k: skMark, ref: mdPetal<<10 | int32(mc.Idx),
				fx: fxBreathe, bold: len(s.crown) > 0}
			if len(s.crown) == 0 { // the crown not yet raised
				c.sym, c.r, c.ink, c.lv = mapmodel.SymNone, '◇', mapmodel.InkFrameDim, 1
			}
			if mc.Idx < len(s.crown) {
				s.anchorAt(s.crown[mc.Idx].Key, p)
			}
		case mapstyle.MdLine:
			k := mc.Ring + off
			*c = skyCell{r: '·', ink: inkEra + mapmodel.SkyInk(s.rings[k].Epoch), lv: 1, k: skRing, ref: int32(k),
				fx: fxBreathe, ph: uint8(k + 1)}
		case mapstyle.MdBead:
			k := mc.Ring + off
			r := s.rings[k]
			*c = skyCell{sym: mapmodel.EraSym(r.Epoch), ink: inkEra + mapmodel.SkyInk(r.Epoch), lv: 2, k: skMark,
				ref: int32(k<<10 | mc.Idx), fx: fxBreathe, ph: uint8(k + 1)}
			if mc.Idx < len(r.Marks) {
				s.anchorAt(r.Marks[mc.Idx].Key, p)
			}
			if mc.DY == 0 && mc.DX > 0 && r.Epoch < len(m.Catalog.EpochName) {
				s.labels = append(s.labels, skyLabel{p: p, text: m.Catalog.EpochName[r.Epoch], ink: mapmodel.InkFrame, far: true})
			}
		}
		c.bg, c.soft = bg, soft
	}
}

// anchorAt makes p the Tab target of building key, if it has none yet.
func (s *skyScene) anchorAt(key string, p mapmodel.Pt) {
	if _, ok := s.anchor[key]; !ok {
		s.anchor[key] = p
	}
}

// mandalaLife sets a mote or two drifting in from the dark, each settling
// at the rim: the mandala itself stays still and symmetric.
func (s *skyScene) mandalaLife() {
	ext := s.md.Extent + 2.5
	for i := 0; i < 2; i++ {
		x := 1
		if i%2 == 1 {
			x = skyW - 2
		}
		y := mdY - 8 + int(mapmodel.Hash(s.b.seed, 662, int64(i))%17)
		lane := lineCells(pt(x, y), pt(mdX, mdY))
		for k, p := range lane {
			if inEllipse(p.X, p.Y, mdX, mdY, mapstyle.MandalaAspect*ext, ext) {
				lane = lane[:k]
				break
			}
		}
		s.addMover(smMote, lane, i, true, 40+int(mapmodel.Hash(s.b.seed, 661, int64(i))%80))
	}
}

// breathe is the slow wave of light: it rolls out from the core ring by
// ring (the cell's phase is its ring plus one, the core and crown 0). Only
// the brightness changes; no glyph ever does.
func (v *skyView) breathe(s *skyScene, c skyCell, l look, x, y int) look {
	b := mapstyle.Breath(v.anim, int(c.ph)-1, len(s.rings))
	switch c.k {
	case skRing:
		l.lv = 0
		if b > 0.45 {
			l.lv = 1
		}
		if b > 0.9 {
			l.lv = 2
		}
	case skMark:
		if c.lv >= 2 {
			l.lv = 2
			if b > 0.7 {
				l.lv = 3
			}
		}
	case skCore:
		l.lv, l.bold = 3, b > 0.3
	}
	return l
}

// describeMark is a bead on a ring (one of the building types the player
// raised in that era, or a wonder, the beads taking them in turn), a petal
// of the crown (one of this age's own types), or the core.
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
	case ring == mdPetal && len(s.crown) == 0:
		return mapstyle.Inspection{Title: "A petal of the crown", Lines: []string{"it lights with what you build in this age"}}, true
	case ring == mdPetal:
		mk = s.crown[j%len(s.crown)]
	case ring < len(s.rings) && len(s.rings[ring].Marks) == 0:
		return describeRing(s, ring), true
	case ring < len(s.rings):
		mk = s.rings[ring].Marks[j%len(s.rings[ring].Marks)]
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

// describeRing is a ring's dotted line: its era and what the player raised
// in it.
func describeRing(s *skyScene, ring int) mapstyle.Inspection {
	r := s.rings[ring]
	name := "an era"
	if r.Epoch < len(s.m.Catalog.EpochName) {
		name = "the " + s.m.Catalog.EpochName[r.Epoch]
	}
	wonders := 0
	for _, mk := range r.Marks {
		if mk.Wonder {
			wonders++
		}
	}
	lines := []string{plural(len(r.Marks)-wonders, "kind of building", "kinds of building") + " still standing"}
	if wonders > 0 {
		lines = append(lines, plural(wonders, "wonder", "wonders"))
	}
	return mapstyle.Inspection{Title: "Ring of " + name, Lines: lines}
}

// compactMandala draws the mandala for the mini view at its own scale (a
// down-sampled one runs the rings together): the same geometry laid out
// for the space, with as many of the newest rings as it holds, breathing
// like the full one.
func (v *skyView) compactMandala(cv *mapstyle.Canvas, w, h int) {
	s := v.sc
	cx, cy := w/2, 1+(h-1)/2
	g := mapstyle.LayMandala(len(s.rings), float64(min(cy-1, h-cy)), float64(min(cx, w-1-cx)))
	off := len(s.rings) - g.Rings
	at := map[mapmodel.Pt]mapstyle.MandalaCell{}
	for _, mc := range g.Cells {
		at[pt(cx+mc.DX, cy+mc.DY)] = mc
	}
	ext := g.Extent + 1
	for y := 1; y < 1+h; y++ {
		for x := 0; x < w; x++ {
			c := starAt(s.b.seed, s.sky, x, y)
			if inEllipse(x, y, float64(cx), float64(cy), mapstyle.MandalaAspect*ext, ext) {
				c = skyCell{r: ' ', k: skVoid}
			}
			if mc, ok := at[pt(x, y)]; ok {
				switch k := mc.Ring + off; mc.Part {
				case mapstyle.MdCore:
					c = skyCell{sym: mapmodel.SymCore, ink: mapmodel.InkGlow, lv: 3, k: skCore, bold: true, fx: fxBreathe}
				case mapstyle.MdPetal:
					c = skyCell{sym: mapmodel.SymPetal, ink: mapmodel.InkAccent, lv: 2, k: skMark, fx: fxBreathe}
					if len(s.crown) == 0 {
						c.sym, c.r, c.ink, c.lv = mapmodel.SymNone, '◇', mapmodel.InkFrameDim, 1
					}
				case mapstyle.MdLine:
					c = skyCell{r: '·', ink: inkEra + mapmodel.SkyInk(s.rings[k].Epoch), lv: 1, k: skRing, fx: fxBreathe,
						ph: uint8(k + 1)}
				case mapstyle.MdBead:
					e := s.rings[k].Epoch
					c = skyCell{sym: mapmodel.EraSym(e), ink: inkEra + mapmodel.SkyInk(e), lv: 2, k: skMark, fx: fxBreathe,
						ph: uint8(k + 1)}
				}
			}
			l := v.resolve(s, c, x, y)
			cv.Put(x, y, l.r, v.style(l))
		}
	}
}
