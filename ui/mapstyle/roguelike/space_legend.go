package roguelike

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// space_legend.go is the sky scenes' legend: like the town's, it lists only
// what the frame drew, by group, and a row's glyph is the one drawn. The
// rare visitor never takes a row; the Galactic Age's alien ships do, being
// ordinary traffic there.

type skyLgID uint16

const (
	slStars skyLgID = iota
	slWonder
	slRising
	slPlanet
	slLimb
	slCity
	slTether
	slHub
	slRing
	slArc
	slTruss
	slSlip
	slRock
	slMoon
	slNebula
	slSun
	slOrbitLine
	slGate
	slScaffold
	slField
	slColony
	slBeacon
	slRelay
	slSpiral
	slLane
	slSystem
	slStarbase
	slPylon
	slCloud
	slEcho
	slCrown
	slCore
	slCiv
	slRuin
	slLegacy
	slUnder
	slFresh
	slFlagged
	slSparkle
	slPulsar
	slDebris
	// slPart plus an index in mapmodel.LineageOrder is that lineage's part;
	// slMover plus a skyMoverKind is that mover's row.
	slPart
	slMover = slPart + 18
	// slEra plus an epoch is that era's ring on the mandala.
	slEra    = slMover + skyLgID(numSkyMoverKinds)
	numSkyLg = slEra + 7
)

var skyLgInfo = [slPart]struct {
	label string
	group uint8
}{
	slStars: {"stars", 0}, slWonder: {"a wonder's star", 1}, slRising: {"wonder rising", 1},
	slPlanet: {"the homeworld", 0}, slLimb: {"atmosphere", 0}, slCity: {"city lights", 0},
	slTether: {"space elevator", 2}, slHub: {"hub", 2}, slRing: {"station ring", 2}, slArc: {"outer ring", 2},
	slTruss: {"truss", 2}, slSlip: {"slip, hull rising", 2}, slRock: {"asteroid", 0}, slMoon: {"the moon", 0},
	slNebula: {"nebula", 0}, slSun: {"the home sun", 0}, slOrbitLine: {"orbit", 0}, slGate: {"warp gate", 2},
	slScaffold: {"gate scaffold", 2}, slField: {"gate field", 2}, slColony: {"colony world", 1},
	slBeacon: {"colony beacon", 1}, slRelay: {"relay chain", 2}, slSpiral: {"the galaxy", 0},
	slLane: {"trade lane", 3}, slSystem: {"star system", 3}, slStarbase: {"starbase ring", 2},
	slPylon: {"docking pylon", 2}, slCloud: {"probability cloud", 0}, slEcho: {"echo of the old town", 0},
	slCrown: {"the crown", 1}, slCore: {"the core", 1}, slCiv: {"civ", 3}, slRuin: {"ruins", 4},
	slLegacy: {"older section (dim)", 4}, slUnder: {"understaffed (dim)", 4}, slFresh: {"new since last visit", 4},
	slFlagged: {"short of hands", 4}, slSparkle: {"transporter beam", 5}, slPulsar: {"the pulsar", 0}, slDebris: {"debris", 0},
}

var skyLgGroups = [6]string{"space", "yours", "structures", "civs", "state", "traffic"}

// skyLgLabel is a row's label and group.
func (v *skyView) skyLgLabel(id skyLgID) (string, uint8) {
	switch {
	case id < slPart:
		return skyLgInfo[id].label, skyLgInfo[id].group
	case id < slMover:
		if i := int(id - slPart); i < len(mapmodel.LineageOrder) {
			sky := mapmodel.SkyOrbit
			if v.sc != nil {
				sky = v.sc.sky
			}
			return mapmodel.SkyPartOf(sky, mapmodel.LineageOrder[i]).Name, 1
		}
		return "", 1
	case id < slEra:
		return skyMoverKinds[id-slMover].Info().Name, 5
	case id < numSkyLg:
		if e := int(id - slEra); v.sc != nil && e < len(v.sc.m.Catalog.EpochName) {
			return v.sc.m.Catalog.EpochName[e], 1
		}
		return "", 1
	}
	return "", 0
}

type skyLgEntry struct {
	on bool
	r  rune
	st tcell.Style
}

// reg claims a legend row with what was drawn, the first time a frame
// draws it.
func (v *skyView) reg(id skyLgID, l look) {
	if int(id) < len(v.seen) && !v.seen[id].on {
		v.seen[id] = skyLgEntry{true, l.r, v.style(l)}
	}
}

// legendCell registers the legend row a drawn cell belongs to.
func (v *skyView) legendCell(s *skyScene, c skyCell, l look) {
	if l.r == ' ' && l.bg == mapmodel.InkVoid {
		return
	}
	if c.fx == fxSparkle && c.k == skVoid {
		v.reg(slSparkle, look{r: '✧', ink: l.ink, lv: 3, bold: true})
		return
	}
	switch c.k {
	case skStar:
		if c.lv >= 2 {
			v.reg(slStars, l)
		}
	case skWonder:
		if c.fx == fxBlink {
			v.reg(slRising, look{r: l.r, ink: l.ink, lv: 2})
		} else {
			v.reg(slWonder, l)
		}
	case skPlanet:
		v.reg(slPlanet, look{r: ' ', bg: c.bg})
	case skLimb:
		v.reg(slLimb, l)
	case skCity:
		v.reg(slCity, l)
	case skTether:
		v.reg(slTether, look{r: '║', ink: l.ink, lv: l.lv})
	case skHub:
		v.reg(slHub, l)
	case skFrame, skSlot:
		if int(c.ref) < len(s.b.frames) {
			if id := s.b.frames[c.ref].lg; id != 0 {
				v.reg(id, l)
			}
		}
	case skRock:
		v.reg(slRock, l)
	case skMoon:
		v.reg(slMoon, look{r: ' ', bg: c.bg})
	case skNebula:
		v.reg(slNebula, l)
	case skSun:
		if s.sky == mapmodel.SkyGalaxy {
			v.reg(slPulsar, look{r: '✦', ink: mapmodel.InkStarBright, lv: 3, bold: true})
		} else {
			v.reg(slSun, l)
		}
	case skOrbit:
		v.reg(slOrbitLine, l)
	case skColony:
		v.reg(slColony, l)
	case skBeacon:
		v.reg(slBeacon, look{r: l.r, ink: l.ink, lv: 2})
	case skSpiral:
		v.reg(slSpiral, l)
	case skLane:
		v.reg(slLane, l)
	case skCloud:
		if l.r != ' ' {
			v.reg(slCloud, l)
		}
	case skEcho:
		v.reg(slEcho, l)
	case skRing, skMark:
		if c.k == skMark && c.ref>>10 == mdPetal {
			v.reg(slCrown, l)
		} else if c.k == skMark && int(c.ref>>10) < len(s.rings) {
			e := s.rings[c.ref>>10].Epoch
			v.reg(slEra+skyLgID(e), look{r: l.r, ink: l.ink, lv: 2})
		}
	case skCore:
		v.reg(slCore, l)
	case skField:
		v.reg(slField, l)
	case skCiv:
		if !c.bold {
			break // a civ's seat stands for it, not its outskirts
		}
		if s.sky == mapmodel.SkyGalaxy {
			v.reg(slSystem, l)
		} else {
			v.reg(slCiv, l)
		}
	case skUnit:
		tt := &s.m.Town.Tiles[c.ref]
		v.reg(slPart+skyLgID(lineageIdx(tt.Lineage)), look{r: l.r, ink: l.ink, lv: 2})
	}
}

// drawSkyLegend lists what this frame drew, by group.
func (v *skyView) drawSkyLegend(cv *mapstyle.Canvas, x, y, w, h int) {
	cv.Text(x, y, w, "LEGEND", v.g.cls(mapmodel.CAccent).Bold(true))
	row, end := y+1, y+h
	for grp := uint8(0); grp < uint8(len(skyLgGroups)); grp++ {
		head := false
		for id := skyLgID(0); id < numSkyLg; id++ {
			label, g := v.skyLgLabel(id)
			if e := v.seen[id]; e.on && g == grp && label != "" {
				if !head && row+1 < end {
					cv.Text(x, row, w, skyLgGroups[grp], v.g.cls(mapmodel.CDim))
					row, head = row+1, true
				}
				if !head || row >= end {
					return
				}
				cv.Put(x+1, row, e.r, e.st)
				cv.Text(x+3, row, w-3, label, v.g.cls(mapmodel.CText))
				row++
			}
		}
	}
}
