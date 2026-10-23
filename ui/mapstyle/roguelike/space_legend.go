package roguelike

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/vibe"
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
	label vibe.Entry
	group uint8
}{
	slStars: {vibe.Same("stars"), 0}, slWonder: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.02", "a wonder's star"), 1}, slRising: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.03", "wonder rising"), 1},
	slPlanet: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.04", "the homeworld"), 0}, slLimb: {vibe.Same("atmosphere"), 0}, slCity: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.06", "city lights"), 0},
	slTether: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.07", "space elevator"), 2}, slHub: {vibe.Same("hub"), 2}, slRing: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.09", "station ring"), 2}, slArc: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.10", "outer ring"), 2},
	slTruss: {vibe.Same("truss"), 2}, slSlip: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.12", "slip, hull rising"), 2}, slRock: {vibe.Same("asteroid"), 0}, slMoon: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.14", "the moon"), 0},
	slNebula: {vibe.Same("nebula"), 0}, slSun: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.16", "the home sun"), 0}, slOrbitLine: {vibe.Same("orbit"), 0}, slGate: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.18", "warp gate"), 2},
	slScaffold: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.19", "gate scaffold"), 2}, slField: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.20", "gate field"), 2}, slColony: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.21", "colony world"), 1},
	slBeacon: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.22", "colony beacon"), 1}, slRelay: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.23", "relay chain"), 2}, slSpiral: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.24", "the galaxy"), 0},
	slLane: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.25", "trade lane"), 3}, slSystem: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.26", "star system"), 3}, slStarbase: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.27", "starbase ring"), 2},
	slPylon: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.28", "docking pylon"), 2}, slCloud: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.29", "probability cloud"), 0}, slEcho: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.30", "echo of the old town"), 0},
	slCrown: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.31", "the crown"), 1}, slCore: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.32", "the core"), 1}, slCiv: {vibe.Same("civ"), 3}, slRuin: {vibe.Same("ruins"), 4},
	slLegacy: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.35", "older section (dim)"), 4}, slUnder: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.36", "understaffed (dim)"), 4}, slFresh: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.37", "new since last visit"), 4},
	slFlagged: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.38", "short of hands"), 4}, slSparkle: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.39", "transporter beam"), 5}, slPulsar: {vibe.E("ui.mapstyle.roguelike.space_legend.skyLgInfo.40", "the pulsar"), 0}, slDebris: {vibe.Same("debris"), 0},
}

var skyLgGroups = [6]string{"space", "yours", "structures", "civs", "state", "traffic"}

// skyLgLabel is a row's label and group.
func (v *skyView) skyLgLabel(id skyLgID) (string, uint8) {
	switch {
	case id < slPart:
		return skyLgInfo[id].label.String(), skyLgInfo[id].group
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
			v.reg(slCrown, look{r: mapmodel.R(mapmodel.SymPetal, v.tier), ink: c.ink, lv: 3, bold: true})
		} else if c.k == skMark && int(c.ref>>10) < len(s.rings) {
			e := s.rings[c.ref>>10].Epoch
			v.reg(slEra+skyLgID(e), look{r: l.r, ink: c.ink, lv: 3, bold: true})
		}
	case skCore:
		switch {
		case s.sky != mapmodel.SkyMandala:
			v.reg(slCore, l)
		case c.ref != mdRay:
			v.reg(slCore, look{r: l.r, ink: c.ink, lv: 3, bold: true})
		}
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
	cv.Text(x, y, w, vibe.Line("ui.mapstyle.roguelike.space_legend.drawSkyLegend.01", "LEGEND"), v.g.cls(mapmodel.CAccent).Bold(true))
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
