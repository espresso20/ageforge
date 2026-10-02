package skyline

import (
	"math"

	"github.com/espresso20/ageforge/mapmodel"
)

// orbit_forms.go is the sky arc's building grammar: what every lot becomes
// in each scene, by lineage (mapmodel.SkyPartOf names the same parts). Each
// scene has its own family of shapes: station modules on the truss in the
// Space Age, small worlds and ships in formation in the Interstellar,
// starbase decks and docked starships in the Galactic, impossible geometry
// in the Quantum and pure light in the Transcendent. Like the ground's
// forms they are drawn in slots, so the scene's palette (skySlots) colours
// them and one drawing serves every theme.

// skyForm is a lot's sky form: an archetype, its nominal height in rows at
// scale 1 and the sprite cache key that names it.
type skyForm struct {
	fn     archFn
	height float64
	key    string
}

// keystoneWonder is the wonder that stands at the foot of a scene's
// signature structure while its age is the present: the gate, the
// starbase hub and the fractal landmark carry its look, so its own lot is
// a small keystone.
var keystoneWonder = [mapmodel.NumSkyScenes]string{
	mapmodel.SkyDeep: "warp_nexus", mapmodel.SkyGalaxy: "cosmic_beacon", mapmodel.SkyQuantum: "reality_anchor",
}

type formDef struct {
	fn     archFn
	height float64
	name   string
}

func fd(fn archFn, h float64, name string) formDef { return formDef{fn, h, name} }

// skyForms is the form table: a scene's form for each lineage.
var skyForms = [mapmodel.NumSkyScenes]map[string]formDef{
	mapmodel.SkyOrbit: {
		mapmodel.LinHousing: fd(oHabitat, 12, "o.hab"), mapmodel.LinFood: fd(oHydro, 7, "o.hydro"),
		mapmodel.LinStorage: fd(oDepot, 6, "o.depot"), mapmodel.LinWood: fd(oBioVat, 8, "o.vat"),
		mapmodel.LinMines: fd(oMiningRig, 8, "o.mine"), mapmodel.LinMetal: fd(oFoundry, 9, "o.foundry"),
		mapmodel.LinEngineer: fd(oShipyard, 10, "o.yard"), mapmodel.LinEnergy: fd(oSolar, 11, "o.solar"),
		mapmodel.LinHarbor: fd(oDockPort, 8, "o.dock"), mapmodel.LinHacker: fd(oRelay, 8, "o.relay"),
		mapmodel.LinKnowledge: fd(oObservatory, 10, "o.obs"), mapmodel.LinFaith: fd(oSanctuary, 14, "o.sanct"),
		mapmodel.LinCulture: fd(oGallery, 9, "o.gallery"), mapmodel.LinMonument: fd(oBeaconMast, 15, "o.beacon"),
		mapmodel.LinTrade: fd(oMarket, 9, "o.market"), mapmodel.LinMilitary: fd(oArmored, 7, "o.armor"),
		mapmodel.LinDiplomacy: fd(oEmbassy, 10, "o.embassy"),
	},
	mapmodel.SkyDeep: {
		mapmodel.LinHousing: fd(dArk, 6, "d.ark"), mapmodel.LinFood: fd(worldOf(wFarm), 6, "d.farm"),
		mapmodel.LinStorage: fd(worldOf(wDepot), 5, "d.depot"), mapmodel.LinWood: fd(worldOf(wGarden), 6, "d.garden"),
		mapmodel.LinMines: fd(worldOf(wMining), 7, "d.mining"), mapmodel.LinMetal: fd(worldOf(wForge), 7, "d.forge"),
		mapmodel.LinEngineer: fd(dGateFrame, 11, "d.gatef"), mapmodel.LinEnergy: fd(dCollector, 9, "d.collect"),
		mapmodel.LinHarbor: fd(worldOf(wPort), 7, "d.port"), mapmodel.LinHacker: fd(dRelayBeacon, 12, "d.relay"),
		mapmodel.LinKnowledge: fd(worldOf(wResearch), 8, "d.research"), mapmodel.LinFaith: fd(worldOf(wShrine), 9, "d.shrine"),
		mapmodel.LinCulture: fd(worldOf(wArts), 8, "d.arts"), mapmodel.LinMonument: fd(worldOf(wMemorial), 9, "d.memorial"),
		mapmodel.LinTrade: fd(worldOf(wTrade), 8, "d.trade"), mapmodel.LinMilitary: fd(dFrigate, 6, "d.frigate"),
		mapmodel.LinDiplomacy: fd(worldOf(wEmbassy), 8, "d.embassy"),
	},
	mapmodel.SkyGalaxy: {
		mapmodel.LinHousing: fd(deckOf(kHab), 15, "g.hab"), mapmodel.LinFood: fd(deckOf(kAgri), 10, "g.agri"),
		mapmodel.LinStorage: fd(deckOf(kCargo), 8, "g.cargo"), mapmodel.LinWood: fd(deckOf(kBio), 10, "g.bio"),
		mapmodel.LinMines: fd(gNeutronMine, 9, "g.mine"), mapmodel.LinMetal: fd(gForge, 10, "g.forge"),
		mapmodel.LinEngineer: fd(gPylon, 14, "g.pylon"), mapmodel.LinEnergy: fd(gQuasarTap, 14, "g.tap"),
		mapmodel.LinHarbor: fd(gBay, 8, "g.bay"), mapmodel.LinHacker: fd(deckOf(kUplink), 13, "g.uplink"),
		mapmodel.LinKnowledge: fd(deckOf(kResearch), 13, "g.research"), mapmodel.LinFaith: fd(deckOf(kShrine), 16, "g.shrine"),
		mapmodel.LinCulture: fd(deckOf(kArchive), 12, "g.archive"), mapmodel.LinMonument: fd(gBeaconSpire, 18, "g.spire"),
		mapmodel.LinTrade: fd(deckOf(kExchange), 12, "g.exchange"), mapmodel.LinMilitary: fd(gStarship, 6, "g.ship"),
		mapmodel.LinDiplomacy: fd(deckOf(kEmbassy), 12, "g.embassy"),
	},
	mapmodel.SkyQuantum: {
		mapmodel.LinHousing: fd(qStair, 13, "q.stair"), mapmodel.LinFood: fd(qOrb, 9, "q.orb"),
		mapmodel.LinStorage: fd(qTesseract, 8, "q.tess"), mapmodel.LinWood: fd(qShards, 10, "q.shards"),
		mapmodel.LinMines: fd(qShards, 9, "q.shards"), mapmodel.LinMetal: fd(qPenrose, 10, "q.penrose"),
		mapmodel.LinEngineer: fd(qTesseract, 12, "q.tess"), mapmodel.LinEnergy: fd(qOrb, 11, "q.orb"),
		mapmodel.LinHarbor: fd(qKnot, 8, "q.knot"), mapmodel.LinHacker: fd(qTesseract, 10, "q.tess"),
		mapmodel.LinKnowledge: fd(qPenrose, 12, "q.penrose"), mapmodel.LinFaith: fd(qStair, 15, "q.stair"),
		mapmodel.LinCulture: fd(qKnot, 10, "q.knot"), mapmodel.LinMonument: fd(qShards, 14, "q.shards"),
		mapmodel.LinTrade: fd(qOrb, 10, "q.orb"), mapmodel.LinMilitary: fd(qPenrose, 9, "q.penrose"),
		mapmodel.LinDiplomacy: fd(qKnot, 11, "q.knot"),
	},
	mapmodel.SkyMandala: {
		mapmodel.LinHousing: fd(mPillar, 8.1, "m.pillar"), mapmodel.LinFood: fd(mCircle, 4.3, "m.circle"),
		mapmodel.LinStorage: fd(mDiamond, 3.7, "m.diamond"), mapmodel.LinWood: fd(mCircle, 5, "m.circle"),
		mapmodel.LinMines: fd(mDiamond, 4.3, "m.diamond"), mapmodel.LinMetal: fd(mPillar, 6.2, "m.pillar"),
		mapmodel.LinEngineer: fd(mObelisk, 7.4, "m.obelisk"), mapmodel.LinEnergy: fd(mStar, 5.6, "m.star"),
		mapmodel.LinHarbor: fd(mCircle, 5, "m.circle"), mapmodel.LinHacker: fd(mDiamond, 5, "m.diamond"),
		mapmodel.LinKnowledge: fd(mObelisk, 8.1, "m.obelisk"), mapmodel.LinFaith: fd(mPillar, 9.3, "m.pillar"),
		mapmodel.LinCulture: fd(mStar, 5.6, "m.star"), mapmodel.LinMonument: fd(mObelisk, 9.9, "m.obelisk"),
		mapmodel.LinTrade: fd(mDiamond, 5.6, "m.diamond"), mapmodel.LinMilitary: fd(mCircle, 5.6, "m.circle"),
		mapmodel.LinDiplomacy: fd(mStar, 6.2, "m.star"),
	},
}

// skyFormFor is a lot's form in scene sc. Wonders become the scene's
// landmarks (the sky-era wonders keep their own drawing outside the
// Transcendent's pure light), and the present age's keystone wonder
// (keystoneWonder) a small keystone.
func skyFormFor(sc mapmodel.SkyScene, def *mapmodel.Def, ml *mapmodel.Lot, keystone bool) skyForm {
	scene := string(rune('0' + int(sc)))
	if def.Wonder {
		switch {
		case keystone:
			return skyForm{fn: keystoneFn[sc], height: 9, key: scene + "key:" + def.Key}
		case sc != mapmodel.SkyMandala && skyEraWonder[def.Key] > 0:
			k := def.Key
			return skyForm{fn: func(r *rnd, h int) *sprite { return wonderSprite(k, float64(h)/20) },
				height: 20 * skyEraWonder[def.Key], key: scene + "w:" + def.Key}
		}
		return skyForm{fn: landmarkFn[sc], height: landmarkH[sc], key: scene + "lm:" + def.Key}
	}
	if f, ok := skyForms[sc][def.Lineage]; ok {
		return skyForm{fn: f.fn, height: f.height, key: f.name}
	}
	f := skyForms[sc][mapmodel.LinHousing]
	if f.fn == nil {
		return skyForm{fn: oHabitat, height: 10, key: scene + "default"}
	}
	return skyForm{fn: f.fn, height: f.height, key: f.name}
}

// skyEraWonder lists the wonders whose own drawings already belong in
// space, with a size factor.
var skyEraWonder = map[string]float64{
	"stellar_cradle": 0.85, "dyson_scaffold": 0.85, "warp_nexus": 0.8, "cosmic_beacon": 0.95,
	"reality_anchor": 0.8, "singularity_core": 0.8,
}

var landmarkFn = [mapmodel.NumSkyScenes]archFn{
	mapmodel.SkyOrbit: oLandmark, mapmodel.SkyDeep: dLandmark, mapmodel.SkyGalaxy: gLandmark,
	mapmodel.SkyQuantum: qLandmark, mapmodel.SkyMandala: mLandmark,
}

var landmarkH = [mapmodel.NumSkyScenes]float64{
	mapmodel.SkyOrbit: 15, mapmodel.SkyDeep: 11, mapmodel.SkyGalaxy: 19, mapmodel.SkyQuantum: 15, mapmodel.SkyMandala: 10,
}

var keystoneFn = [mapmodel.NumSkyScenes]archFn{
	mapmodel.SkyOrbit: oLandmark, mapmodel.SkyDeep: dKeystone, mapmodel.SkyGalaxy: gKeystone,
	mapmodel.SkyQuantum: qKeystone, mapmodel.SkyMandala: mLandmark,
}

// ---------------------------------------------------------------- palettes

// slotSpec is what a sprite slot is made of in a scene: an ink, how it is
// lit, and a shade factor.
type slotSpec struct {
	ink  mapmodel.SkyInk
	mode int
	k    float64
}

func ss(ink mapmodel.SkyInk, mode int, k float64) slotSpec { return slotSpec{ink, mode, k} }

// skySlots maps the sprite slots onto each scene's inks: walls to the
// frame, windows to the lights (lit by staffing in the lot drawing), metal
// to the dim frame, glass to the glow, neon to the accents.
var skySlots = [mapmodel.NumSkyScenes][numSlots]slotSpec{
	mapmodel.SkyOrbit: {
		sWall: ss(mapmodel.InkFrame, iLit, 0.88), sWallLit: ss(mapmodel.InkFrame, iLit, 1.0),
		sWallShade: ss(mapmodel.InkFrameDim, iLit, 0.92), sWallDark: ss(mapmodel.InkEcho, iLit, 0.85),
		sRoof: ss(mapmodel.InkFrame, iLit, 0.95), sRoofShade: ss(mapmodel.InkFrameDim, iLit, 0.85),
		sTrim: ss(mapmodel.InkFrameDim, iLit, 1.08), sWin: ss(mapmodel.InkEcho, iLit, 0.5),
		sMetal: ss(mapmodel.InkFrameDim, iLit, 1.0), sMetalDark: ss(mapmodel.InkEcho, iLit, 1.0),
		sGlass: ss(mapmodel.InkGlow, iLit, 0.55), sGlassHi: ss(mapmodel.InkGlow, iEmit, 0.9),
		sNeon1: ss(mapmodel.InkAccent, iEmit, 1.0), sNeon2: ss(mapmodel.InkAccent2, iEmit, 1.0),
		sNeon3: ss(mapmodel.InkAccent3, iEmit, 1.0), sGlow: ss(mapmodel.InkLight, iEmit, 1.0),
		sBeacon: ss(mapmodel.InkAccent3, iEmit, 1.0), sLeaf: ss(mapmodel.InkAccent2, iEmit, 0.85),
		sLeafDark: ss(mapmodel.InkAccent2, iLit, 0.5), sTrunk: ss(mapmodel.InkRockDark, iLit, 1.0),
		sField1: ss(mapmodel.InkAccent2, iLit, 0.7), sField2: ss(mapmodel.InkAccent2, iLit, 0.5),
		sRock: ss(mapmodel.InkRock, iLit, 1.0), sRockDark: ss(mapmodel.InkRockDark, iLit, 1.0),
		sFire: ss(mapmodel.InkLight, iEmit, 1.0), sInk: ss(mapmodel.InkVoid, iBack, 1.0),
		sWater: ss(mapmodel.InkSurface, iLit, 1.4), sSmoke: ss(mapmodel.InkHaze, iBack, 1.0),
	},
	mapmodel.SkyDeep: {
		sWall: ss(mapmodel.InkMoon, iLit, 0.86), sWallLit: ss(mapmodel.InkMoon, iLit, 1.0),
		sWallShade: ss(mapmodel.InkFrame, iLit, 0.62), sWallDark: ss(mapmodel.InkFrameDim, iLit, 0.9),
		sRoof: ss(mapmodel.InkMoon, iLit, 0.9), sRoofShade: ss(mapmodel.InkFrame, iLit, 0.6),
		sTrim: ss(mapmodel.InkFrame, iLit, 0.95), sWin: ss(mapmodel.InkFrameDim, iLit, 0.6),
		sMetal: ss(mapmodel.InkFrame, iLit, 0.78), sMetalDark: ss(mapmodel.InkFrameDim, iLit, 1.0),
		sGlass: ss(mapmodel.InkGlow, iLit, 0.5), sGlassHi: ss(mapmodel.InkGlow, iEmit, 0.85),
		sNeon1: ss(mapmodel.InkAccent2, iEmit, 1.0), sNeon2: ss(mapmodel.InkGlow, iEmit, 1.0),
		sNeon3: ss(mapmodel.InkAccent3, iEmit, 1.0), sGlow: ss(mapmodel.InkAccent, iEmit, 1.0),
		sBeacon: ss(mapmodel.InkLight, iEmit, 1.0), sLeaf: ss(mapmodel.InkSurface2, iLit, 1.15),
		sLeafDark: ss(mapmodel.InkSurface2, iLit, 0.7), sTrunk: ss(mapmodel.InkRockDark, iLit, 1.0),
		sField1: ss(mapmodel.InkSurface2, iLit, 1.2), sField2: ss(mapmodel.InkAccent, iLit, 0.55),
		sRock: ss(mapmodel.InkRock, iLit, 1.0), sRockDark: ss(mapmodel.InkRockDark, iLit, 1.1),
		sFire: ss(mapmodel.InkAccent3, iEmit, 1.0), sInk: ss(mapmodel.InkVoid, iBack, 1.0),
		sWater: ss(mapmodel.InkSurface, iLit, 1.15), sSmoke: ss(mapmodel.InkHaze, iBack, 1.0),
	},
	mapmodel.SkyGalaxy: {
		sWall: ss(mapmodel.InkFrame, iLit, 0.9), sWallLit: ss(mapmodel.InkFrame, iLit, 1.0),
		sWallShade: ss(mapmodel.InkFrameDim, iLit, 0.95), sWallDark: ss(mapmodel.InkFrameDim, iLit, 0.55),
		sRoof: ss(mapmodel.InkFrame, iLit, 0.95), sRoofShade: ss(mapmodel.InkFrameDim, iLit, 0.85),
		sTrim: ss(mapmodel.InkAccent, iLit, 0.95), sWin: ss(mapmodel.InkFrameDim, iLit, 0.45),
		sMetal: ss(mapmodel.InkFrameDim, iLit, 1.0), sMetalDark: ss(mapmodel.InkFrameDim, iLit, 0.62),
		sGlass: ss(mapmodel.InkAccent2, iLit, 0.5), sGlassHi: ss(mapmodel.InkAccent2, iEmit, 0.85),
		sNeon1: ss(mapmodel.InkAccent2, iEmit, 1.0), sNeon2: ss(mapmodel.InkAccent, iEmit, 1.0),
		sNeon3: ss(mapmodel.InkAccent3, iEmit, 1.0), sGlow: ss(mapmodel.InkGlow, iEmit, 1.0),
		sBeacon: ss(mapmodel.InkAccent3, iEmit, 1.0), sLeaf: ss(mapmodel.InkAccent2, iEmit, 0.7),
		sLeafDark: ss(mapmodel.InkSurface, iLit, 1.0), sTrunk: ss(mapmodel.InkFrameDim, iLit, 0.5),
		sField1: ss(mapmodel.InkAccent, iLit, 0.7), sField2: ss(mapmodel.InkAccent, iLit, 0.5),
		sRock: ss(mapmodel.InkRock, iLit, 1.0), sRockDark: ss(mapmodel.InkRockDark, iLit, 1.0),
		sFire: ss(mapmodel.InkAccent, iEmit, 1.0), sInk: ss(mapmodel.InkVoid, iBack, 1.0),
		sWater: ss(mapmodel.InkSurface, iLit, 1.0), sSmoke: ss(mapmodel.InkHaze, iBack, 1.0),
	},
	mapmodel.SkyQuantum: {
		sWall: ss(mapmodel.InkFrame, iLit, 0.8), sWallLit: ss(mapmodel.InkFrame, iLit, 1.0),
		sWallShade: ss(mapmodel.InkFrameDim, iLit, 0.7), sWallDark: ss(mapmodel.InkFrameDim, iLit, 0.4),
		sRoof: ss(mapmodel.InkFrame, iLit, 0.9), sRoofShade: ss(mapmodel.InkFrameDim, iLit, 0.6),
		sTrim: ss(mapmodel.InkAccent, iLit, 0.9), sWin: ss(mapmodel.InkFrameDim, iLit, 0.35),
		sMetal: ss(mapmodel.InkAccent2, iLit, 0.8), sMetalDark: ss(mapmodel.InkAccent2, iLit, 0.45),
		sGlass: ss(mapmodel.InkGlow, iLit, 0.6), sGlassHi: ss(mapmodel.InkGlow, iEmit, 0.9),
		sNeon1: ss(mapmodel.InkAccent, iEmit, 1.0), sNeon2: ss(mapmodel.InkAccent2, iEmit, 1.0),
		sNeon3: ss(mapmodel.InkAccent3, iEmit, 1.0), sGlow: ss(mapmodel.InkLight, iEmit, 1.0),
		sBeacon: ss(mapmodel.InkAccent3, iEmit, 1.0), sLeaf: ss(mapmodel.InkRock, iEmit, 0.8),
		sLeafDark: ss(mapmodel.InkRock, iLit, 0.5), sTrunk: ss(mapmodel.InkFrameDim, iLit, 0.5),
		sField1: ss(mapmodel.InkRock, iLit, 0.7), sField2: ss(mapmodel.InkRock, iLit, 0.5),
		sRock: ss(mapmodel.InkRock, iLit, 0.8), sRockDark: ss(mapmodel.InkRockDark, iLit, 0.6),
		sFire: ss(mapmodel.InkLight, iEmit, 1.0), sInk: ss(mapmodel.InkVoid, iBack, 1.0),
		sWater: ss(mapmodel.InkGlow, iLit, 0.7), sSmoke: ss(mapmodel.InkHaze, iBack, 1.0),
	},
	mapmodel.SkyMandala: {
		sWall: ss(mapmodel.InkLight, iEmit, 0.85), sWallLit: ss(mapmodel.InkGlow, iEmit, 1.0),
		sWallShade: ss(mapmodel.InkFrameDim, iEmit, 1.0), sWallDark: ss(mapmodel.InkFrameDim, iEmit, 0.6),
		sRoof: ss(mapmodel.InkFrame, iEmit, 1.0), sRoofShade: ss(mapmodel.InkFrame, iEmit, 0.7),
		sTrim: ss(mapmodel.InkFrame, iEmit, 1.0), sWin: ss(mapmodel.InkFrameDim, iEmit, 0.55),
		sMetal: ss(mapmodel.InkFrame, iEmit, 0.85), sMetalDark: ss(mapmodel.InkFrameDim, iEmit, 0.8),
		sGlass: ss(mapmodel.InkLight, iEmit, 0.6), sGlassHi: ss(mapmodel.InkGlow, iEmit, 1.0),
		sNeon1: ss(mapmodel.InkAccent, iEmit, 1.0), sNeon2: ss(mapmodel.InkLight, iEmit, 1.0),
		sNeon3: ss(mapmodel.InkFrameDim, iEmit, 1.0), sGlow: ss(mapmodel.InkGlow, iEmit, 1.0),
		sBeacon: ss(mapmodel.InkGlow, iEmit, 1.0), sLeaf: ss(mapmodel.InkAccent, iEmit, 0.8),
		sLeafDark: ss(mapmodel.InkFrameDim, iEmit, 0.8), sTrunk: ss(mapmodel.InkFrameDim, iEmit, 0.6),
		sField1: ss(mapmodel.InkAccent, iEmit, 0.7), sField2: ss(mapmodel.InkFrameDim, iEmit, 0.7),
		sRock: ss(mapmodel.InkFrame, iEmit, 0.8), sRockDark: ss(mapmodel.InkFrameDim, iEmit, 0.7),
		sFire: ss(mapmodel.InkAccent, iEmit, 1.0), sInk: ss(mapmodel.InkVoid, iBack, 1.0),
		sWater: ss(mapmodel.InkLight, iEmit, 0.6), sSmoke: ss(mapmodel.InkHaze, iBack, 1.0),
	},
}

// ---------------------------------------------------------------- helpers

// disc fills an ellipse in half-block pixels: centre (cx, cpy) in columns
// and half rows, radii rx and ry (columns and half rows are about the same
// size on a terminal, so rx == ry is round). sl picks the slot at a point
// (dx, dy in -1..1).
func disc(s *sprite, cx, cpy, rx, ry float64, sl func(dx, dy float64) slot) {
	for py := int(math.Floor(cpy - ry)); py <= int(math.Ceil(cpy+ry)); py++ {
		for x := int(math.Floor(cx - rx)); x <= int(math.Ceil(cx+rx)); x++ {
			dx := (float64(x) + 0.5 - cx) / rx
			dy := (float64(py) + 0.5 - cpy) / ry
			if float64(dx*dx)+float64(dy*dy) <= 1 {
				if v := sl(dx, dy); v != sNone {
					s.pset(x, py, v)
				}
			}
		}
	}
}

// ring draws an elliptical outline in half-block pixels, t pixels thick,
// over the arc where keep(angle in turns) holds.
func ring(s *sprite, cx, cpy, rx, ry, t float64, sl func(dx, dy float64) slot) {
	for py := int(math.Floor(cpy - ry - 1)); py <= int(math.Ceil(cpy+ry+1)); py++ {
		for x := int(math.Floor(cx - rx - 1)); x <= int(math.Ceil(cx+rx+1)); x++ {
			dx := (float64(x) + 0.5 - cx) / rx
			dy := (float64(py) + 0.5 - cpy) / ry
			d := math.Sqrt(float64(dx*dx) + float64(dy*dy))
			if d <= 1 && d >= 1-t/math.Min(rx, ry) {
				if v := sl(dx, dy); v != sNone {
					s.pset(x, py, v)
				}
			}
		}
	}
}

// lit is the usual light: lit on the left, shaded on the right.
func lit(a, b, c slot) func(dx, dy float64) slot {
	return func(dx, dy float64) slot {
		switch {
		case dx < -0.45:
			return a
		case dx > 0.35:
			return c
		}
		return b
	}
}

// capsule draws a horizontal cylinder module w wide over rows [y, y+n),
// n >= 3: rounded ends, a window row in the middle.
func capsule(s *sprite, x, y, w, n int) {
	s.hline(x+1, x+w-2, y, '▄', sWall, sNone)
	s.put(x+1, y, '▄', sWallLit, sNone)
	s.put(x+w-2, y, '▄', sWallShade, sNone)
	for yy := y + 1; yy < y+n-1; yy++ {
		s.body(x, yy, w, 1)
	}
	s.hline(x+1, x+w-2, y+n-1, '▀', sWall, sNone)
	s.put(x+w-2, y+n-1, '▀', sWallShade, sNone)
	mid := y + (n-1)/2
	for xx := x + 1; xx < x+w-1; xx += 2 {
		s.put(xx, mid, '▪', sWin, s.at(xx, mid).fg)
	}
	if n >= 5 {
		for xx := x + 2; xx < x+w-1; xx += 2 {
			s.put(xx, mid+1, '▪', sWin, s.at(xx, mid+1).fg)
		}
	}
}

// stem is a strut from row y0 down to row y1 at column x.
func stem(s *sprite, x, y0, y1 int) {
	for y := y0; y <= y1; y++ {
		s.put(x, y, '│', sMetal, sNone)
	}
}

// ---------------------------------------------------------------- the Space Age: station modules

// oHabitat: habitat cylinders stacked on a spine, windows lit by staffing.
func oHabitat(r *rnd, h int) *sprite {
	w := 7 + 2*r.n(2)
	n := 3
	if h >= 13 {
		n = 4
	}
	segs := clampInt((h-1)/(n+1), 1, 4)
	s := newSprite(w, segs*(n+1)+3)
	y := s.h - 1
	s.put(w/2, y, '█', sMetalDark, sNone)
	s.put(w/2-1, y, '▄', sMetal, sNone)
	s.put(w/2+1, y, '▄', sMetal, sNone)
	y--
	for i := 0; i < segs; i++ {
		capsule(s, 0, y-n+1, w, n)
		if i%2 == 1 {
			s.hline(1, w-2, y-n+1, '▄', sNeon1, sNone)
		}
		y -= n
		if i < segs-1 {
			s.text(w/2-1, y, "▐█▌", sMetal, sNone)
			s.put(w/2, y, '█', sMetalDark, sNone)
			y--
		}
	}
	s.put(w/2, y, '│', sMetal, sNone)
	s.beacons = append(s.beacons, pt{w / 2, y - 1})
	return s.trimTop()
}

// oHydro: a hydroponic bay, its glass walls glowing green in bands.
func oHydro(r *rnd, h int) *sprite {
	w := r.rng(7, 10)
	rows := clampInt(h-1, 3, 8)
	s := newSprite(w, rows+2)
	base := s.h - 1
	s.put(w/2, base, '▀', sMetalDark, sNone)
	top := base - rows
	s.hline(1, w-2, top, '▄', sWall, sNone)
	for y := top + 1; y < base; y++ {
		s.put(0, y, '▐', sWallLit, sNone)
		s.put(w-1, y, '▌', sWallShade, sNone)
		for x := 1; x < w-1; x++ {
			if (y-top)%2 == 1 {
				s.put(x, y, '▀', sLeaf, sWall)
			} else {
				s.put(x, y, '▄', sLeaf, sWall)
			}
		}
		if (y-top)%3 == 0 {
			s.put(w/2, y, '█', sWall, sNone)
		}
	}
	s.hline(1, w-2, base-1, '▀', sWall, sNone)
	s.smoke = append(s.smoke, pt{1, top})
	return s
}

// oDepot: boxy depot pods stacked a little askew, panel lines on each.
func oDepot(r *rnd, h int) *sprite {
	w := r.rng(8, 10)
	n := clampInt(h/2, 1, 4)
	s := newSprite(w+2, n*2+1)
	y := s.h - 1
	s.put(w/2, y, '█', sMetalDark, sNone)
	for i := 0; i < n; i++ {
		pw := w - r.n(3)
		px := r.n(w - pw + 2)
		y--
		fill := sWall
		if i%2 == 1 {
			fill = sMetal
		}
		for x := px; x < px+pw; x++ {
			sl := fill
			switch x {
			case px:
				sl = sWallLit
			case px + pw - 1:
				sl = sWallShade
			}
			s.put(x, y, '█', sl, sNone)
			if x > px && x < px+pw-1 && (x-px)%3 == 0 {
				s.put(x, y, '▬', sWallDark, fill)
			}
		}
		y--
		s.hline(px, px+pw-1, y, '▄', fill, sNone)
	}
	return s.trimTop()
}

// oBioVat: tall vats of living green on a frame, glowing vents on top.
func oBioVat(r *rnd, h int) *sprite {
	n := r.rng(2, 3)
	th := clampInt(h-1, 3, 10)
	s := newSprite(n*3+1, th+2)
	base := s.h - 1
	s.hline(0, s.w-1, base, '▀', sMetal, sNone)
	for i := 0; i < n; i++ {
		x := i*3 + 1
		vh := th - r.n(max(1, th/3))
		top := base - vh
		level := top + 1 + r.n(max(1, vh/2))
		for y := top + 1; y < base; y++ {
			sl := sGlass
			if y >= level {
				sl = sLeaf
			}
			s.put(x, y, '█', sl, sNone)
			s.put(x+1, y, '▌', sMetalDark, sNone)
			if y%3 == 0 {
				s.put(x, y, '▀', sMetal, sl)
			}
		}
		s.put(x, top, '▄', sMetal, sNone)
		s.smoke = append(s.smoke, pt{x, top - 1})
	}
	return s
}

// oMiningRig: a chunk of asteroid held off the truss, a rig on its back and
// a beacon on the rig.
func oMiningRig(r *rnd, h int) *sprite {
	w := r.rng(8, 11)
	rh := clampInt(h-3, 2, 6)
	s := newSprite(w, rh+5)
	base := s.h - 1
	stem(s, w/2, base-1, base)
	cpy := float64(2*(base-1)) - float64(rh)
	seed := r.n(1000)
	disc(s, float64(w)/2, cpy, float64(w)/2-0.3, float64(rh)+0.2, func(dx, dy float64) slot {
		n := mapmodel.Noise(int64(seed), (dx+1)*2.5, (dy+1)*2.5)
		if float64(dx*dx)+float64(dy*dy) > 0.55+0.45*n {
			return sNone
		}
		if dx > 0.25 || n < 0.35 {
			return sRockDark
		}
		return sRock
	})
	top := s.topAt(w / 2)
	if top < 0 {
		top = base - rh
	}
	s.put(w/2-1, top-1, '╱', sMetal, sNone)
	s.put(w/2, top-1, '█', sMetalDark, sNone)
	s.put(w/2+1, top-1, '╲', sMetal, sNone)
	s.put(w/2, top-2, '┼', sMetal, sNone)
	s.beacons = append(s.beacons, pt{w / 2, top - 3})
	s.smoke = append(s.smoke, pt{w/2 - 2, top})
	return s.trimTop()
}

// oFoundry: a foundry tank with a glowing mouth and pipes to its smelter.
func oFoundry(r *rnd, h int) *sprite {
	th := clampInt(h-1, 4, 10)
	tw := r.rng(5, 6)
	s := newSprite(tw+4, th+2)
	base := s.h - 1
	top := base - th
	for y := top + 1; y < base; y++ {
		s.body(0, y, tw, 1)
		if (y-top)%3 == 0 {
			s.hline(0, tw-1, y, '▀', sMetal, sWall)
		}
	}
	s.hline(1, tw-2, top, '▄', sWall, sNone)
	s.put(tw/2, base-2, '▄', sGlow, sWallDark)
	s.put(tw/2-1, base-2, '▄', sGlow, sWallDark)
	s.put(tw/2, base-1, '█', sGlow, sNone)
	s.put(tw/2-1, base-1, '▀', sGlow, sWallDark)
	s.hline(tw, tw+2, base-3, '═', sMetal, sNone)
	for y := base - 2; y < base; y++ {
		s.put(tw+2, y, '█', sMetal, sNone)
		s.put(tw+3, y, '▌', sMetalDark, sNone)
	}
	s.hline(0, s.w-1, base, '▀', sMetalDark, sNone)
	s.smoke = append(s.smoke, pt{1, top}, pt{tw + 2, base - 3})
	return s
}

// oShipyard: an open frame with a hull taking shape inside, half plating
// and half scaffold, sparks where the welders are.
func oShipyard(r *rnd, h int) *sprite {
	w := r.rng(11, 14)
	th := clampInt(h-1, 5, 11)
	s := newSprite(w, th+2)
	base := s.h - 1
	top := base - th
	s.put(0, top, '╔', sMetal, sNone)
	s.put(w-1, top, '╗', sMetal, sNone)
	s.hline(1, w-2, top, '═', sMetal, sNone)
	for y := top + 1; y < base; y++ {
		s.put(0, y, '║', sMetal, sNone)
		s.put(w-1, y, '║', sMetalDark, sNone)
	}
	s.put(0, base, '╩', sMetal, sNone)
	s.put(w-1, base, '╩', sMetalDark, sNone)
	s.hline(1, w-2, base, '═', sMetalDark, sNone)
	hy := top + th/2
	built := 2 + r.n(max(1, w-6))
	for x := 2; x < w-2; x++ {
		if x-2 < built {
			s.put(x, hy-1, '▄', sWall, sNone)
			s.put(x, hy, '█', sWall, sNone)
			s.put(x, hy+1, '▀', sWallShade, sNone)
			if x%2 == 0 {
				s.put(x, hy, '▪', sWin, sWall)
			}
		} else {
			s.put(x, hy-1, '┄', sMetal, sNone)
			s.put(x, hy+1, '┄', sMetal, sNone)
		}
	}
	s.put(1, hy, '◄', sWallShade, sNone)
	s.flame = append(s.flame, pt{2 + built, hy})
	s.hline(1, w-2, top+1, '┊', sMetalDark, sNone)
	s.put(w/2, top+1, '▼', sMetal, sNone)
	return s
}

// oSolar: blue solar wings on a mast.
func oSolar(r *rnd, h int) *sprite {
	w := r.rng(9, 13) | 1
	th := clampInt(h, 4, 14)
	s := newSprite(w, th+1)
	base := s.h - 1
	cx := w / 2
	stem(s, cx, base-th+1, base)
	gap := 3
	if th >= 9 {
		gap = 4
	}
	for y := base - 1; y > base-th+1; y -= gap {
		span := cx - 1 - r.n(2)
		for dx := 1; dx <= span; dx++ {
			sl := sNeon1
			if dx%3 == 0 {
				sl = sGlass
			}
			s.put(cx-dx, y, '▀', sl, sNone)
			s.put(cx+dx, y, '▀', sl, sNone)
		}
		s.put(cx, y, '┼', sMetal, sNone)
	}
	s.put(cx, base-th+1, '┴', sMetal, sNone)
	s.put(cx, base-th, '•', sGlow, sNone)
	return s.trimTop()
}

// oDockPort: a node with docking hatches either side and a capsule docked on
// top, its light on.
func oDockPort(r *rnd, h int) *sprite {
	w := 9
	th := clampInt(h, 5, 9)
	s := newSprite(w, th+1)
	base := s.h - 1
	stem(s, w/2, base-1, base)
	ny := base - 3
	s.body(2, ny, 5, 2)
	s.hline(2, 6, ny-1, '▄', sWall, sNone)
	s.put(0, ny, '◘', sMetal, sNone)
	s.put(1, ny, '═', sMetal, sNone)
	s.put(8, ny, '◘', sMetal, sNone)
	s.put(7, ny, '═', sMetal, sNone)
	s.put(4, ny+1, '▀', sGlow, s.at(4, ny+1).fg)
	cy := ny - 2
	for y := ny - th + 4; y < cy+1; y++ {
		if y >= 0 {
			s.put(3, y, '▐', sWallLit, sNone)
			s.put(4, y, '█', sWall, sNone)
			s.put(5, y, '▌', sWallShade, sNone)
		}
	}
	if t := ny - th + 4; t-1 >= 0 {
		s.put(4, t-1, '▲', sWall, sNone)
		s.put(4, t+1, '▪', sWin, sWall)
	}
	return s.trimTop()
}

// oRelay: an array of relay dishes on a frame.
func oRelay(r *rnd, h int) *sprite {
	n := r.rng(2, 3)
	s := newSprite(n*4+1, clampInt(h, 5, 9)+1)
	base := s.h - 1
	s.hline(0, s.w-1, base-1, '▄', sMetal, sNone)
	s.put(s.w/2, base, '│', sMetal, sNone)
	for i := 0; i < n; i++ {
		x := i * 4
		up := base - 2 - r.n(max(1, s.h-6))
		s.put(x+1, up, '▄', sMetal, sNone)
		s.put(x+2, up, '▄', sMetal, sNone)
		s.put(x, up+1, '▀', sMetal, sNone)
		s.put(x+1, up+1, '▄', sMetalDark, sNone)
		s.put(x+2, up+1, '▄', sMetalDark, sNone)
		s.put(x+3, up+1, '▀', sMetal, sNone)
		for y := up + 2; y < base-1; y++ {
			s.put(x+2, y, '│', sMetalDark, sNone)
		}
		s.beacons = append(s.beacons, pt{x + 2, up - 1})
	}
	return s.trimTop()
}

// oObservatory: a domed observatory, slit open and the telescope showing,
// with a dish beside it.
func oObservatory(r *rnd, h int) *sprite {
	w := r.rng(9, 11) | 1
	bh := clampInt(h-5, 2, 5)
	s := newSprite(w+3, bh+8)
	base := s.h - 1
	s.body(0, base-bh, w, bh)
	s.windows(1, base-bh+1, w-2, bh-1, 2, 2, 0, 0, '▪')
	d := s.dome(1, w-2, base-bh-1, sWall)
	cx := w / 2
	for y := base - bh - d; y <= base-bh-1; y++ {
		s.put(cx, y, '█', sWallDark, sNone)
	}
	s.put(cx+1, base-bh-d-1, '╱', sGlassHi, sNone)
	s.put(cx+2, base-bh-d-2, '╱', sGlassHi, sNone)
	s.put(w, base-bh-2, '▄', sMetal, sNone)
	s.put(w+1, base-bh-2, '▀', sMetal, sNone)
	s.put(w+1, base-bh-3, '▄', sMetal, sNone)
	stem(s, w, base-bh-1, base)
	s.beacons = append(s.beacons, pt{w + 2, base - bh - 4})
	return s.trimTop()
}

// oSanctuary: a slender spire over a small chapel module, a light at its tip.
func oSanctuary(r *rnd, h int) *sprite {
	w := 7
	th := clampInt(h, 7, 18)
	s := newSprite(w, th+1)
	base := s.h - 1
	cx := w / 2
	s.body(1, base-2, 5, 2)
	s.hline(1, 5, base-3, '▄', sWall, sNone)
	s.put(cx, base-2, '◘', sGlow, sWall)
	for y := base - 4; y > base-th+1; y-- {
		t := float64(base-4-y) / float64(max(1, th-5))
		if t < 0.55 {
			s.put(cx-1, y, '▐', sWallLit, sNone)
			s.put(cx+1, y, '▌', sWallShade, sNone)
		}
		s.put(cx, y, '█', sWall, sNone)
		if (base-y)%4 == 0 {
			s.put(cx, y, '▪', sWin, sWall)
		}
	}
	s.put(cx, base-th+1, '▲', sWall, sNone)
	s.put(cx, base-th, '•', sGlow, sNone)
	return s.trimTop()
}

// oGallery: a gallery sphere ringed with windows on a strut.
func oGallery(r *rnd, h int) *sprite {
	rr := clampInt(h-2, 3, 7)
	w := 2*rr + 1
	s := newSprite(w, rr+3)
	base := s.h - 1
	stem(s, w/2, base-1, base)
	cpy := float64(2*(base-1)) - float64(rr)
	disc(s, float64(w)/2, cpy, float64(rr), float64(rr), lit(sWallLit, sWall, sWallShade))
	mid := int(cpy) / 2
	for x := 1; x < w-1; x++ {
		if c := s.at(x, mid); c.ch != 0 {
			s.put(x, mid, '▬', sGlow, c.fg)
			if x%2 == 0 {
				s.put(x, mid, '▪', sWin, c.fg)
			}
		}
	}
	return s.trimTop()
}

// oBeaconMast: a tall mast of crossbars carrying blinking beacons.
func oBeaconMast(r *rnd, h int) *sprite {
	th := clampInt(h, 6, 22)
	s := newSprite(5, th+1)
	base := s.h - 1
	stem(s, 2, base-th+1, base)
	for y := base - 2; y > base-th+1; y -= 3 {
		s.put(1, y, '─', sMetal, sNone)
		s.put(2, y, '┼', sMetal, sNone)
		s.put(3, y, '─', sMetal, sNone)
		s.beacons = append(s.beacons, pt{0, y}, pt{4, y})
	}
	s.text(1, base, "▄█▄", sMetalDark, sNone)
	s.beacons = append(s.beacons, pt{2, base - th})
	return s.trimTop()
}

// oMarket: hexagonal market docks in a honeycomb, their bays lit.
func oMarket(r *rnd, h int) *sprite {
	hexes := 2
	if h >= 8 {
		hexes = 3
	}
	s := newSprite(11, 8)
	hex := func(x, y int) {
		s.text(x+1, y, "▄▄▄▄", sWall, sNone)
		s.put(x, y+1, '█', sWallLit, sNone)
		s.hline(x+1, x+4, y+1, '█', sWall, sNone)
		s.put(x+5, y+1, '█', sWallShade, sNone)
		s.put(x+2, y+1, '▪', sGlow, sWall)
		s.put(x+3, y+1, '▪', sWin, sWall)
		s.text(x+1, y+2, "▀▀▀▀", sWallShade, sNone)
	}
	hex(0, 5)
	hex(5, 5)
	if hexes == 3 {
		hex(2, 2)
		s.put(5, 1, '│', sMetal, sNone)
		s.beacons = append(s.beacons, pt{5, 0})
	}
	s.put(5, 7, '█', sMetalDark, sNone)
	return s.trimTop()
}

// oArmored: an armoured module, sloped plates, gun slits and a turret.
func oArmored(r *rnd, h int) *sprite {
	w := r.rng(9, 11)
	s := newSprite(w, 6)
	base := s.h - 1
	s.body(1, base-2, w-2, 2)
	s.put(0, base-1, '◢', sWallDark, sNone)
	s.put(w-1, base-1, '◣', sWallDark, sNone)
	s.put(0, base-2, '▗', sWallDark, sNone)
	s.put(w-1, base-2, '▖', sWallDark, sNone)
	for x := 2; x < w-2; x += 3 {
		s.put(x, base-1, '▬', sWallDark, s.at(x, base-1).fg)
	}
	s.hline(2, w-3, base-3, '▄', sWallShade, sNone)
	s.put(w/2, base-4, '▄', sMetal, sNone)
	s.put(w/2+1, base-4, '═', sMetal, sNone)
	s.put(w/2+2, base-4, '═', sMetal, sNone)
	s.put(2, base-4, '►', sBeacon, sNone)
	s.hline(0, w-1, base, '▀', sMetalDark, sNone)
	return s.trimTop()
}

// oEmbassy: a round module in a ring, the banners of the civs met on masts.
func oEmbassy(r *rnd, h int) *sprite {
	rr := clampInt(h/3, 2, 4)
	w := 2*rr + 5
	s := newSprite(w, rr+5)
	base := s.h - 1
	stem(s, w/2, base-1, base)
	cpy := float64(2*(base-1)) - float64(rr)
	disc(s, float64(w)/2, cpy, float64(rr), float64(rr), lit(sWallLit, sWall, sWallShade))
	mid := int(cpy) / 2
	s.hline(0, w-1, mid, '─', sTrim, sNone)
	for x := 2; x < w-2; x++ {
		if c := s.at(x, mid); c.ch != 0 && c.ch != '─' {
			s.put(x, mid, '▪', sWin, c.fg)
		}
	}
	top := s.topAt(w / 2)
	for i, sl := range []slot{sNeon1, sNeon2, sNeon3} {
		x := w/2 - 2 + i*2
		if top-2 >= 0 {
			s.put(x, top-1, '│', sMetal, sNone)
			s.put(x+1, top-2, '►', sl, sNone)
			s.put(x, top-2, '│', sMetal, sNone)
		}
	}
	return s.trimTop()
}

// oLandmark: the Space Age's landmark for a wonder of an earlier age: a
// wheel station on a pylon, its ring seen at a slant, spokes to a lit hub.
func oLandmark(r *rnd, h int) *sprite {
	th := clampInt(h, 8, 22)
	rx := float64(th) * 0.95
	w := int(rx*2) + 3
	s := newSprite(w, th+1)
	base := s.h - 1
	cx := float64(w) / 2
	ry := float64(th) * 0.42
	cpy := float64(2*base) - float64(th) - 1
	ring(s, cx, cpy, rx, ry, 1.6, func(dx, dy float64) slot {
		if dy < -0.2 {
			return sWallShade
		}
		if dx < -0.5 {
			return sWallLit
		}
		return sWall
	})
	hubY := int(cpy) / 2
	ic := int(cx)
	for x := ic - int(rx) + 2; x < ic+int(rx)-1; x++ {
		if s.at(x, hubY).ch == 0 {
			s.put(x, hubY, '─', sMetal, sNone)
		}
	}
	for y := hubY - int(ry/2) + 1; y < hubY+int(ry/2); y++ {
		if s.at(ic, y).ch == 0 {
			s.put(ic, y, '│', sMetal, sNone)
		}
	}
	s.put(ic-1, hubY, '▐', sWallLit, sNone)
	s.put(ic, hubY, '◘', sGlow, sWall)
	s.put(ic+1, hubY, '▌', sWallShade, sNone)
	// the ring's windows
	for x := 1; x < w-1; x += 2 {
		for y := 0; y < base; y++ {
			if c := s.at(x, y); c.ch == '█' && (x+y)%3 == 0 {
				s.put(x, y, '▪', sWin, c.fg)
			}
		}
	}
	for y := hubY + 1; y <= base; y++ {
		if s.at(ic, y).ch == 0 || s.at(ic, y).ch == '│' {
			s.put(ic, y, '█', sMetalDark, sNone)
		}
	}
	s.put(ic-1, base, '▄', sMetal, sNone)
	s.put(ic+1, base, '▄', sMetal, sNone)
	s.beacons = append(s.beacons, pt{ic, max(0, s.topAt(ic)-1)})
	return s.trimTop()
}

// ---------------------------------------------------------------- the Interstellar Age: worlds and ships

// dArk: a colony ark, a long hull with its turning ring sections, drive
// glowing aft. It floats a row above the formation line.
func dArk(r *rnd, h int) *sprite {
	w := r.rng(14, 18)
	rings := 2
	if w >= 16 {
		rings = 3
	}
	s := newSprite(w, 6)
	y := 3
	s.hline(1, w-2, y, '█', sWall, sNone)
	s.put(w-1, y, '►', sWallShade, sNone)
	s.put(0, y, '▐', sNeon2, sNone)
	s.hline(1, w-3, y+1, '▀', sWallShade, sNone)
	s.hline(2, w-3, y-1, '▄', sWall, sNone)
	for x := 2; x < w-2; x += 2 {
		s.put(x, y, '▪', sWin, sWall)
	}
	for i := 0; i < rings; i++ {
		x := 3 + i*(w-6)/max(1, rings)
		s.put(x, y-2, '▄', sMetal, sNone)
		s.put(x+1, y-2, '▄', sMetal, sNone)
		s.put(x, y-1, '█', sWallLit, sNone)
		s.put(x+1, y-1, '█', sWall, sNone)
		s.put(x, y+1, '█', sMetal, sNone)
		s.put(x+1, y+1, '█', sMetalDark, sNone)
		s.put(x, y+2, '▀', sMetal, sNone)
		s.put(x+1, y+2, '▀', sMetalDark, sNone)
	}
	s.smoke = append(s.smoke, pt{0, y - 1})
	return s.trimTop()
}

type worldKind uint8

const (
	wFarm worldKind = iota
	wGarden
	wMining
	wForge
	wResearch
	wShrine
	wArts
	wMemorial
	wTrade
	wEmbassy
	wDepot
	wPort
)

// worldOf is a small world of a kind, floating over the formation line:
// its face in the kind's colours, lit on one side, the colony's lights on
// its night side and the kind's mark on top.
func worldOf(k worldKind) archFn {
	return func(r *rnd, h int) *sprite { return dWorld(r, h, k) }
}

func dWorld(r *rnd, h int, k worldKind) *sprite {
	rr := clampInt(h/2+r.n(2), 2, 6)
	w := 2*rr + 3
	s := newSprite(w, rr+5)
	base := s.h - 2 // a row of space under it: it floats
	cpy := float64(2*base) - float64(rr) + 1
	cx := float64(w) / 2
	seed := int64(r.n(1 << 16))
	face := func(dx, dy float64) slot {
		n := mapmodel.Noise(seed, (dx+1)*2.2, (dy+1)*2.2)
		shade := dx > 0.3
		switch k {
		case wFarm:
			if (int((dy+1)*4))%2 == 0 {
				return pickS(shade, sField1, sLeafDark)
			}
			return pickS(shade, sField2, sRockDark)
		case wGarden, wEmbassy:
			if n > 0.5 {
				return pickS(shade, sLeaf, sLeafDark)
			}
			return pickS(shade, sWater, sMetalDark)
		case wMining, wMemorial, wDepot:
			if n < 0.35 {
				return sRockDark
			}
			return pickS(shade, sRock, sRockDark)
		case wForge:
			if n > 0.7 {
				return sFire
			}
			return sRockDark
		case wShrine:
			return pickS(shade, sWall, sWallShade)
		}
		if n > 0.58 {
			return pickS(shade, sLeaf, sLeafDark)
		}
		return pickS(shade, sWater, sMetalDark)
	}
	disc(s, cx, cpy, float64(rr)+0.4, float64(rr)+0.4, face)
	// the colony's lights on the night side
	for y := 0; y <= base; y++ {
		for x := w / 2; x < w; x++ {
			if c := s.at(x, y); c.ch == '█' && hash(x, y, int(seed))%5 == 0 {
				s.put(x, y, '·', sWin, c.fg)
			}
		}
	}
	top := max(0, s.topAt(w/2))
	ic := w / 2
	switch k {
	case wMining:
		if top >= 2 {
			s.put(ic, top-1, '┼', sMetal, sNone)
			s.put(ic-1, top-1, '╱', sMetal, sNone)
			s.beacons = append(s.beacons, pt{ic, top - 2})
		}
	case wResearch:
		if top >= 2 {
			s.put(ic, top-1, '▄', sMetal, sNone)
			s.put(ic+1, top-1, '▀', sMetal, sNone)
		}
	case wShrine, wMemorial:
		for y := top - 1; y >= max(0, top-3); y-- {
			s.put(ic, y, '│', sGlow, sNone)
		}
		if top >= 4 {
			s.put(ic, top-4, '•', sGlow, sNone)
		}
	case wArts, wPort:
		mid := int(cpy) / 2
		for x := 0; x < w; x++ {
			if c := s.at(x, mid); c.ch == 0 {
				s.put(x, mid, '─', sTrim, sNone)
			} else if x == 0 || x == w-1 {
				s.put(x, mid, '─', sTrim, sNone)
			}
		}
	case wTrade:
		if top >= 2 {
			s.put(ic+2, top-1, '▬', sMetal, sNone)
			s.put(ic+3, top-1, '•', sGlow, sNone)
		}
	case wEmbassy:
		if top >= 2 {
			s.put(ic, top-1, '│', sMetal, sNone)
			s.put(ic+1, top-2, '►', sNeon1, sNone)
			s.put(ic, top-2, '│', sMetal, sNone)
		}
	case wDepot:
		if top >= 1 {
			s.put(ic-1, top, '▄', sGlass, s.at(ic-1, top).fg)
			s.put(ic+1, top, '▄', sGlass, s.at(ic+1, top).fg)
		}
	case wForge, wFarm, wGarden:
		s.smoke = append(s.smoke, pt{ic, top})
	}
	return s.trimTop()
}

func pickS(shade bool, a, b slot) slot {
	if shade {
		return b
	}
	return a
}

// dCollector: a stellar collector, a glowing sail of panels on a mast,
// turned to the far sun.
func dCollector(r *rnd, h int) *sprite {
	w := r.rng(5, 7)
	rows := clampInt(h/4, 1, 3)
	s := newSprite(w, rows+4)
	base := s.h - 2 // it floats a row over the line
	top := base - rows - 1
	s.put(0, top, '◢', sNeon1, sNone)
	s.hline(1, w-1, top, '▄', sNeon1, sNone)
	for y := top + 1; y <= top+rows-1; y++ {
		for x := 0; x < w; x++ {
			if x%3 == 2 {
				s.put(x, y, '┼', sMetal, sGlass)
			} else {
				s.put(x, y, '█', sGlass, sNone)
			}
		}
	}
	s.hline(0, w-2, top+rows, '▀', sNeon1, sNone)
	s.put(w-1, top+rows, '◤', sNeon1, sNone)
	s.put(w/2, base, '█', sMetal, sNone)
	s.put(w/2, base-1, '▀', sMetalDark, sNone)
	s.beacons = append(s.beacons, pt{0, top - 1})
	return s.trimTop()
}

// dGateFrame: a warp gate frame, a small ring on its pylon, plated as far
// as it is built and scaffold beyond, its lights on the plated arc.
func dGateFrame(r *rnd, h int) *sprite {
	rr := clampInt(h/2, 3, 7)
	R := float64(2 * rr)
	w := 2*int(R) + 3
	s := newSprite(w, rr*2+3)
	base := s.h - 1
	cx := float64(w) / 2
	cpy := float64(2*(base-1)) - R
	built := 0.3 + 0.6*float64(r.n(100))/100
	for py := 0; py < 2*base; py++ {
		for x := 0; x < w; x++ {
			dx, dy := float64(x)+0.5-cx, float64(py)+0.5-cpy
			d := math.Sqrt(float64(dx*dx) + float64(dy*dy))
			if d > R || d < R-1.7 {
				continue
			}
			from := math.Abs(turns(dx, dy) - 0.75)
			if from > 0.5 {
				from = 1 - from
			}
			switch {
			case from <= built/2:
				sl := sMetal
				if dx < -R*0.45 {
					sl = sWallLit
				} else if dx > R*0.45 {
					sl = sMetalDark
				}
				s.pset(x, py, sl)
			case int(from*R*7)%3 == 0:
				s.pset(x, py, sMetalDark) // scaffold, in dashes round the rest
			}
		}
	}
	ic := w / 2
	s.put(ic, base, '█', sMetalDark, sNone)
	s.put(ic-1, base, '▄', sMetal, sNone)
	s.put(ic+1, base, '▄', sMetal, sNone)
	for i := 0; i < 3; i++ {
		x := ic + int(math.Round((R-0.8)*mapmodel.Cos(0.75+float64(i-1)*built/3)))
		y := int(math.Round((cpy - (R-0.8)*mapmodel.Sin(0.75+float64(i-1)*built/3)) / 2))
		if c := s.at(x, y); c.ch != 0 {
			s.put(x, y, '▪', sNeon1, c.fg)
		}
	}
	return s.trimTop()
}

// dFrigate: an escort frigate, an angular hull with its drive lit.
func dFrigate(r *rnd, h int) *sprite {
	w := r.rng(10, 13)
	s := newSprite(w, 5)
	y := 2
	s.hline(2, w-3, y-1, '▄', sWallShade, sNone)
	s.put(w/2, y-1, '█', sWall, sNone)
	s.put(w/2-1, y-2, '▄', sMetal, sNone)
	s.hline(1, w-2, y, '█', sWall, sNone)
	s.put(w-1, y, '►', sWall, sNone)
	s.put(0, y, '▐', sNeon2, sNone)
	s.put(1, y, '▓', sNeon2, sWall)
	s.hline(2, w-4, y+1, '▀', sWallDark, sNone)
	for x := 3; x < w-2; x += 3 {
		s.put(x, y, '▬', sWallDark, sWall)
	}
	s.put(w-3, y, '▪', sNeon3, sWall)
	return s.trimTop()
}

// dRelayBeacon: a relay beacon, a needle hung with rings, its tip blinking.
func dRelayBeacon(r *rnd, h int) *sprite {
	th := clampInt(h, 6, 18)
	s := newSprite(5, th+2)
	base := s.h - 2
	for y := base - th + 1; y <= base; y++ {
		s.put(2, y, '│', sMetal, sNone)
	}
	for y := base - 2; y > base-th+1; y -= 3 {
		s.text(0, y, "═◘═", sNeon1, sNone)
		s.put(1, y, '═', sNeon1, sNone)
		s.put(2, y, '◘', sMetal, sNone)
		s.put(3, y, '═', sNeon1, sNone)
	}
	s.beacons = append(s.beacons, pt{2, base - th})
	return s.trimTop()
}

// dLandmark: a wonder of an earlier age in the Interstellar: a ringed
// memorial world, larger than the rest.
func dLandmark(r *rnd, h int) *sprite {
	rr := clampInt(h/2, 3, 8)
	w := 4*rr + 3
	s := newSprite(w, rr+4)
	base := s.h - 2
	cpy := float64(2*base) - float64(rr) + 1
	cx := float64(w) / 2
	disc(s, cx, cpy, float64(rr)+0.4, float64(rr)+0.4, func(dx, dy float64) slot {
		if (int((dy+1)*5))%2 == 0 {
			return pickS(dx > 0.3, sWall, sWallShade)
		}
		return pickS(dx > 0.3, sTrim, sMetalDark)
	})
	ring(s, cx, cpy, float64(2*rr)+1, float64(rr)/2+0.6, 1, func(dx, dy float64) slot {
		if dy < 0 && math.Abs(dx) < 0.45 {
			return sNone // behind the world
		}
		return sNeon1
	})
	return s.trimTop()
}

// dKeystone: the warp nexus core at the gate's foot: a lit node on a pylon.
func dKeystone(r *rnd, h int) *sprite {
	s := newSprite(7, 6)
	s.text(1, 5, "▄███▄", sMetalDark, sNone)
	s.put(3, 4, '█', sMetal, sNone)
	s.put(3, 3, '█', sMetal, sNone)
	s.text(1, 2, "═◆═", sNeon1, sNone)
	s.put(2, 2, '▐', sNeon1, sNone)
	s.put(3, 2, '◆', sGlow, sNone)
	s.put(4, 2, '▌', sNeon1, sNone)
	s.put(3, 1, '│', sNeon2, sNone)
	s.beacons = append(s.beacons, pt{3, 0})
	return s
}

// ---------------------------------------------------------------- the Galactic Age: starbase decks and ships

type deckKind uint8

const (
	kHab deckKind = iota
	kAgri
	kCargo
	kBio
	kUplink
	kResearch
	kShrine
	kArchive
	kExchange
	kEmbassy
)

// deckOf is a starbase deck tower of a kind.
func deckOf(k deckKind) archFn {
	return func(r *rnd, h int) *sprite { return gDeck(r, h, k) }
}

// gDeck: a starbase deck tower: decks stepping in as they rise, gold trim
// at every deck, lit windows, cyan running lights and a crown by kind.
func gDeck(r *rnd, h int, k deckKind) *sprite {
	w := r.rng(7, 9)
	if k == kExchange || k == kCargo {
		w += 2
	}
	th := clampInt(h, 5, 20)
	s := newSprite(w+2, th+3)
	base := s.h - 1
	y := base
	dw := w
	for y > base-th+1 && dw >= 3 {
		dh := 2 + r.n(2)
		if k == kCargo {
			dh = 2
		}
		x0 := 1 + (w-dw)/2
		for yy := y; yy > y-dh && yy > base-th; yy-- {
			s.body(x0, yy, dw, 1)
			switch k {
			case kAgri, kBio:
				for x := x0 + 1; x < x0+dw-1; x++ {
					s.put(x, yy, '▀', sLeaf, s.at(x, yy).fg)
				}
			case kCargo:
				for x := x0 + 1; x < x0+dw-1; x += 2 {
					s.put(x, yy, '▬', sWallDark, s.at(x, yy).fg)
				}
			default:
				for x := x0 + 1; x < x0+dw-1; x += 2 {
					s.put(x, yy, '▪', sWin, s.at(x, yy).fg)
				}
			}
		}
		y -= dh
		if y > base-th {
			s.hline(x0-1, x0+dw, y, '▄', sTrim, sNone)
			s.put(x0-1, y, '▗', sNeon1, sNone)
			s.put(x0+dw, y, '▖', sNeon1, sNone)
			y--
		}
		if r.n(3) > 0 {
			dw -= 2
		}
	}
	cx := s.w / 2
	if (k == kHab || k == kExchange) && r.n(3) == 0 && y >= 2 {
		// a disc crowning the tower, wider than its top deck
		s.hline(max(0, cx-dw/2-2), min(s.w-1, cx+dw/2+2), y, '▄', sWall, sNone)
		s.hline(max(0, cx-dw/2-1), min(s.w-1, cx+dw/2+1), y-1, '▄', sTrim, sNone)
		for x := max(0, cx-dw/2-1); x <= min(s.w-1, cx+dw/2+1); x += 2 {
			s.put(x, y, '▄', sNeon1, sNone)
		}
		y -= 2
	}
	switch k {
	case kResearch, kUplink:
		s.put(cx-1, y, '▄', sMetal, sNone)
		s.put(cx, y, '▀', sMetal, sNone)
		s.put(cx+1, y, '▄', sMetal, sNone)
		s.put(cx, y-1, '│', sMetal, sNone)
		s.beacons = append(s.beacons, pt{cx, y - 2})
	case kShrine:
		for i := 1; i <= 3; i++ {
			s.put(cx, y-i+1, '│', sTrim, sNone)
		}
		s.put(cx, y-3, '◆', sNeon2, sNone)
	case kArchive:
		s.dome(cx-2, 5, y, sGlass)
	case kEmbassy:
		s.put(cx-1, y, '►', sNeon3, sNone)
		s.put(cx+1, y, '►', sNeon1, sNone)
		s.put(cx, y, '│', sTrim, sNone)
	default:
		s.put(cx, y, '▲', sTrim, sNone)
		s.beacons = append(s.beacons, pt{cx, y - 1})
	}
	if k == kBio || k == kAgri {
		s.smoke = append(s.smoke, pt{1, base - 1})
	}
	return s.trimTop()
}

// gStarship: a starship at dock, saucer forward over the secondary hull,
// warp nacelles aft, its bussard collectors glowing.
func gStarship(r *rnd, h int) *sprite {
	big := h >= 7
	w := 15
	if big {
		w = 19
	}
	s := newSprite(w, 7)
	sx := 1
	sw := 9
	if big {
		sw = 11
	}
	// the saucer, edge on
	s.hline(sx+2, sx+sw-3, 1, '▄', sWall, sNone)
	s.hline(sx, sx+sw-1, 2, '█', sWall, sNone)
	s.put(sx, 2, '▐', sWallLit, sNone)
	s.put(sx+sw-1, 2, '▌', sWallShade, sNone)
	for x := sx + 1; x < sx+sw-1; x += 2 {
		s.put(x, 2, '▪', sWin, sWall)
	}
	s.hline(sx+1, sx+sw-2, 3, '▀', sTrim, sNone)
	s.put(sx+sw/2, 0, '▄', sGlass, sNone)
	// the neck and the secondary hull
	nx := sx + sw - 3
	s.put(nx, 3, '█', sWallShade, sNone)
	s.hline(nx, w-4, 4, '█', sWall, sNone)
	s.put(nx-1, 4, '▐', sNeon3, sNone)
	s.hline(nx, w-4, 5, '▀', sWallShade, sNone)
	// the nacelles, raised on pylons aft
	ny := 2
	s.put(w-5, 3, '╱', sMetal, sNone)
	s.hline(w-7, w-2, ny, '═', sMetal, sNone)
	s.put(w-7, ny, '◘', sNeon3, sNone)
	s.put(w-1, ny, '▌', sNeon1, sNone)
	return s.trimTop()
}

// gPylon: a repair pylon arching over a berth, its tip lit.
func gPylon(r *rnd, h int) *sprite {
	th := clampInt(h, 7, 20)
	w := 7
	s := newSprite(w, th+1)
	base := s.h - 1
	for y := base - 1; y > base-th+3; y-- {
		s.put(1, y, '█', sWallLit, sNone)
		s.put(2, y, '▌', sWallShade, sNone)
		if (base-y)%3 == 0 {
			s.put(1, y, '▀', sTrim, sWallLit)
		}
	}
	t := base - th + 3
	s.put(1, t, '▄', sWall, sNone)
	s.put(2, t-1, '▄', sWall, sNone)
	s.put(3, t-1, '▄', sWall, sNone)
	s.put(4, t, '▀', sWall, sNone)
	s.put(5, t, '▄', sWall, sNone)
	s.put(5, t+1, '│', sMetal, sNone)
	s.put(5, t+2, '◘', sNeon1, sNone)
	s.text(0, base, "▄██▄", sMetalDark, sNone)
	s.beacons = append(s.beacons, pt{3, t - 2})
	return s.trimTop()
}

// gQuasarTap: a coil tower drawing on a quasar, its core blazing.
func gQuasarTap(r *rnd, h int) *sprite {
	th := clampInt(h, 7, 20)
	s := newSprite(7, th+1)
	base := s.h - 1
	for y := base - 1; y > base-th+2; y-- {
		s.put(3, y, '█', sMetal, sNone)
		if (base-y)%2 == 0 {
			s.put(2, y, '▐', sNeon1, sNone)
			s.put(4, y, '▌', sNeon1, sNone)
		}
	}
	t := base - th + 2
	s.put(2, t, '▐', sGlow, sNone)
	s.put(3, t, '█', sGlow, sNone)
	s.put(4, t, '▌', sGlow, sNone)
	s.put(3, t-1, '│', sGlow, sNone)
	s.text(1, base, "▄███▄", sMetalDark, sNone)
	s.smoke = append(s.smoke, pt{3, t + 1})
	return s.trimTop()
}

// gForge: a stellar forge, a dome over a captive star, its heat showing.
func gForge(r *rnd, h int) *sprite {
	w := r.rng(9, 11) | 1
	s := newSprite(w, 8)
	base := s.h - 1
	s.body(0, base-1, w, 2)
	s.hline(0, w-1, base-1, '▀', sTrim, sWall)
	d := s.dome(1, w-2, base-2, sGlass)
	cx := w / 2
	s.put(cx, base-2, '◘', sGlow, sGlass)
	if d > 1 {
		s.put(cx, base-3, '▄', sGlow, sGlass)
	}
	s.smoke = append(s.smoke, pt{1, base - 2}, pt{w - 2, base - 2})
	return s.trimTop()
}

// gNeutronMine: a neutron star held in a rig, light pouring off it.
func gNeutronMine(r *rnd, h int) *sprite {
	s := newSprite(9, 7)
	base := s.h - 1
	s.put(4, 2, '█', sGlow, sNone)
	s.put(3, 2, '▐', sGlow, sNone)
	s.put(5, 2, '▌', sGlow, sNone)
	s.put(4, 1, '▄', sGlow, sNone)
	s.put(4, 3, '▀', sGlow, sNone)
	s.put(1, 1, '╲', sMetal, sNone)
	s.put(7, 1, '╱', sMetal, sNone)
	s.put(1, 3, '╱', sMetal, sNone)
	s.put(7, 3, '╲', sMetal, sNone)
	s.put(0, 2, '═', sMetal, sNone)
	s.put(8, 2, '═', sMetal, sNone)
	for y := 4; y < base; y++ {
		s.put(4, y, '│', sMetal, sNone)
	}
	s.text(2, base, "▄███▄", sMetalDark, sNone)
	s.smoke = append(s.smoke, pt{4, 0})
	return s
}

// gBay: a docking bay, its doors open on a lit hangar with a shuttle in it.
func gBay(r *rnd, h int) *sprite {
	w := r.rng(10, 12)
	s := newSprite(w, 6)
	base := s.h - 1
	s.hline(0, w-1, 1, '▄', sTrim, sNone)
	for y := 2; y < base; y++ {
		s.put(0, y, '█', sWallLit, sNone)
		s.put(w-1, y, '█', sWallShade, sNone)
		for x := 1; x < w-1; x++ {
			s.put(x, y, '█', sWallDark, sNone)
		}
	}
	s.hline(1, w-2, 2, '▀', sNeon1, sWallDark)
	s.put(w/2-1, 3, '▄', sWall, sWallDark)
	s.put(w/2, 3, '▄', sWall, sWallDark)
	s.put(w/2+1, 3, '►', sWall, sWallDark)
	s.hline(0, w-1, base, '▀', sTrim, sNone)
	return s.trimTop()
}

// gBeaconSpire: a monument spire, gold banded, its beacon high.
func gBeaconSpire(r *rnd, h int) *sprite {
	th := clampInt(h, 8, 24)
	s := newSprite(5, th+1)
	base := s.h - 1
	for y := base - 1; y > base-th+1; y-- {
		t := float64(base-y) / float64(th)
		if t < 0.45 {
			s.put(1, y, '▐', sWallLit, sNone)
			s.put(3, y, '▌', sWallShade, sNone)
		}
		s.put(2, y, '█', sWall, sNone)
		if (base-y)%4 == 0 {
			s.put(2, y, '▀', sTrim, sWall)
		}
	}
	s.put(2, base-th+1, '▲', sTrim, sNone)
	s.text(0, base, "▄███▄", sTrim, sNone)
	s.beacons = append(s.beacons, pt{2, base - th})
	return s.trimTop()
}

// gLandmark: a wonder of an earlier age in the Galactic: a grand tower of
// the starbase, gold banded, a ring of light round its crown.
func gLandmark(r *rnd, h int) *sprite {
	sp := gDeck(newRnd(uint64(h)*977+1), h, kHab)
	s := newSprite(sp.w+6, sp.h+2)
	s.blit(sp, 3, 2)
	cx := s.w / 2
	s.hline(0, s.w-1, 3, '▀', sNeon2, sNone)
	s.put(cx, 1, '│', sTrim, sNone)
	s.beacons = append(s.beacons, pt{cx, 0})
	return s.trimTop()
}

// gKeystone: the cosmic beacon's base at the hub's foot: a lens on a plinth.
func gKeystone(r *rnd, h int) *sprite {
	s := newSprite(7, 5)
	s.text(0, 4, "▄█████▄", sTrim, sNone)
	s.text(2, 3, "███", sWall, sNone)
	s.put(3, 3, '▪', sNeon1, sWall)
	s.text(2, 2, "▐◆▌", sNeon2, sNone)
	s.put(3, 2, '◆', sGlow, sNone)
	s.put(3, 1, '│', sGlow, sNone)
	s.put(3, 0, '│', sGlow, sNone)
	return s
}

// ---------------------------------------------------------------- the Quantum Age: impossible geometry

// qTesseract: nested frames hanging in the air.
func qTesseract(r *rnd, h int) *sprite { return aTesseract(r, h) }

// qStair: an impossible stair of offset cubes.
func qStair(r *rnd, h int) *sprite {
	th := clampInt(h, 5, 18)
	w := 9
	s := newSprite(w, th+1)
	base := s.h - 1
	x := 0
	dir := 1
	for y := base; y > base-th+1; y -= 2 {
		s.put(x, y, '█', sWallLit, sNone)
		s.put(x+1, y, '█', sWall, sNone)
		s.put(x+2, y, '▌', sWallShade, sNone)
		s.put(x, y-1, '▄', sWallLit, sNone)
		s.put(x+1, y-1, '▄', sWall, sNone)
		x += dir * 2
		if x > w-3 || x < 0 {
			dir = -dir
			x += dir * 4
			x = clampInt(x, 0, w-3)
		}
	}
	s.put(w/2, base-th+1, '◊', sGlow, sNone)
	return s.trimTop()
}

// qOrb: a sphere in several orbits at once.
func qOrb(r *rnd, h int) *sprite {
	rr := clampInt(h/2, 2, 5)
	w := 4*rr + 1
	s := newSprite(w, rr+4)
	base := s.h - 1
	cpy := float64(2*(base-1)) - float64(rr)
	cx := float64(w) / 2
	disc(s, cx, cpy, float64(rr), float64(rr), lit(sWallLit, sWall, sWallShade))
	ring(s, cx, cpy, float64(2*rr), float64(rr)/2+0.5, 1, func(dx, dy float64) slot {
		if dy < 0 && math.Abs(dx) < 0.5 {
			return sNone
		}
		return sNeon1
	})
	s.put(w/2, base, '▀', sNeon2, sNone)
	return s.trimTop()
}

// qShards: shards of a crystal that is still deciding its shape.
func qShards(r *rnd, h int) *sprite { return aCrystal(r, h) }

// qPenrose: the impossible triangle.
func qPenrose(r *rnd, h int) *sprite {
	th := clampInt(h, 5, 14)
	w := th*2 - 1
	s := newSprite(w, th+1)
	base := s.h - 1
	for i := 0; i < th; i++ {
		y := base - i
		l, rr := i, w-1-i
		if l > rr {
			break
		}
		s.put(l, y, '◢', sWallLit, sNone)
		s.put(rr, y, '◣', sWallShade, sNone)
		if i == 0 {
			s.hline(l+1, rr-1, y, '█', sWall, sNone)
		}
		if i == 1 {
			s.hline(l+1, rr-1, y, '▀', sNeon1, sNone)
		}
	}
	s.put(w/2, base-th/2, '◊', sGlow, sNone)
	return s.trimTop()
}

// qKnot: a loop that passes through itself.
func qKnot(r *rnd, h int) *sprite {
	rr := clampInt(h/2, 2, 6)
	w := 2*rr*2 + 1
	s := newSprite(w, rr+3)
	base := s.h - 1
	cpy := float64(2*(base-1)) - float64(rr)
	ring(s, float64(w)*0.35, cpy, float64(rr), float64(rr), 1.2, func(dx, dy float64) slot { return sWall })
	ring(s, float64(w)*0.65, cpy, float64(rr), float64(rr), 1.2, func(dx, dy float64) slot { return sNeon1 })
	s.put(w/2, base, '▀', sMetal, sNone)
	return s.trimTop()
}

// qLandmark: a wonder of an earlier age in the Quantum: a great tesseract.
func qLandmark(r *rnd, h int) *sprite { return aTesseract(r, h+4) }

// qKeystone: the reality anchor's foot under the fractal: a small inverted
// pyramid pinned by two pillars.
func qKeystone(r *rnd, h int) *sprite {
	s := newSprite(9, 6)
	s.text(0, 0, "▀███████▀", sWall, sNone)
	s.text(1, 1, "▀█████▀", sWall, sNone)
	s.text(2, 2, "▀███▀", sWallShade, sNone)
	s.put(4, 3, '◊', sGlow, sNone)
	s.put(1, 3, '│', sMetal, sNone)
	s.put(7, 3, '│', sMetal, sNone)
	s.put(1, 4, '│', sMetal, sNone)
	s.put(7, 4, '│', sMetal, sNone)
	s.text(0, 5, "▄▀▄   ▄▀▄", sMetalDark, sNone)
	return s
}

// ---------------------------------------------------------------- the Transcendent Age: pure light

// mPillar: a luminous pillar, a capital and a foot of gold.
func mPillar(r *rnd, h int) *sprite {
	th := clampInt(h, 5, 22)
	s := newSprite(5, th+1)
	base := s.h - 1
	for y := base - 1; y > base-th+1; y-- {
		s.put(2, y, '█', sWallLit, sNone)
		s.put(1, y, '▐', sWallShade, sNone)
		s.put(3, y, '▌', sWallShade, sNone)
	}
	s.text(0, base-th+1, "▄███▄", sTrim, sNone)
	s.text(0, base, "▀▀▀▀▀", sTrim, sNone)
	s.put(2, base-th, '·', sGlow, sNone)
	return s.trimTop()
}

// mDiamond: diamonds of light stacked point to point.
func mDiamond(r *rnd, h int) *sprite {
	n := clampInt(h/3, 1, 4)
	s := newSprite(7, n*3+1)
	base := s.h - 1
	for i := 0; i < n; i++ {
		y := base - 1 - i*3
		s.put(3, y+1, '▀', sTrim, sNone)
		s.text(1, y, "◢█◣", sWallLit, sNone)
		s.put(2, y, '█', sWall, sNone)
		s.put(1, y, '◢', sWallLit, sNone)
		s.put(3, y, '█', sWallLit, sNone)
		s.put(4, y, '█', sWall, sNone)
		s.put(5, y, '◣', sWallShade, sNone)
		s.put(2, y-1, '◢', sWallLit, sNone)
		s.put(3, y-1, '█', sWallLit, sNone)
		s.put(4, y-1, '◣', sWall, sNone)
	}
	return s.trimTop()
}

// mCircle: a ring of light on a stem, bright at its heart.
func mCircle(r *rnd, h int) *sprite {
	rr := clampInt(h/2, 2, 6)
	w := 2*rr + 1
	s := newSprite(w, rr+3)
	base := s.h - 1
	cpy := float64(2*(base-1)) - float64(rr)
	cx := float64(w) / 2
	ring(s, cx, cpy, float64(rr), float64(rr), 1.2, func(dx, dy float64) slot { return sWall })
	mid := int(cpy) / 2
	s.put(w/2, mid, '✦', sGlow, sNone)
	for y := mid + rr/2 + 1; y <= base; y++ {
		if s.at(w/2, y).ch == 0 {
			s.put(w/2, y, '│', sTrim, sNone)
		}
	}
	return s.trimTop()
}

// mObelisk: a needle of light.
func mObelisk(r *rnd, h int) *sprite {
	th := clampInt(h, 6, 24)
	s := newSprite(3, th+1)
	base := s.h - 1
	s.put(1, base-th+1, '▲', sWallLit, sNone)
	for y := base - th + 2; y < base; y++ {
		s.put(1, y, '█', sWall, sNone)
		if y > base-th/2 {
			s.put(0, y, '▐', sWallShade, sNone)
			s.put(2, y, '▌', sWallShade, sNone)
		}
	}
	s.text(0, base, "▀█▀", sTrim, sNone)
	return s.trimTop()
}

// mStar: a star of light on a thin stem.
func mStar(r *rnd, h int) *sprite {
	th := clampInt(h, 4, 14)
	s := newSprite(7, th+1)
	base := s.h - 1
	cy := base - th + 2
	s.put(3, cy, '✦', sGlow, sNone)
	s.put(2, cy, '─', sWall, sNone)
	s.put(4, cy, '─', sWall, sNone)
	s.put(1, cy, '·', sTrim, sNone)
	s.put(5, cy, '·', sTrim, sNone)
	s.put(3, cy-1, '│', sWall, sNone)
	s.put(3, cy+1, '│', sWall, sNone)
	for y := cy + 2; y <= base; y++ {
		s.put(3, y, '│', sTrim, sNone)
	}
	return s.trimTop()
}

// mLandmark: a wonder in the Transcendent: a rose of light, a diamond in a
// ring on a pillar.
func mLandmark(r *rnd, h int) *sprite {
	rr := clampInt(h/3, 3, 7)
	w := 2*rr + 3
	th := clampInt(h, rr+4, 24)
	s := newSprite(w, th+1)
	base := s.h - 1
	cpy := float64(2*(base-th+rr+1)) - float64(rr)
	cx := float64(w) / 2
	ring(s, cx, cpy, float64(rr), float64(rr), 1.3, func(dx, dy float64) slot { return sTrim })
	disc(s, cx, cpy, float64(rr)/2, float64(rr)/2, func(dx, dy float64) slot {
		if math.Abs(dx)+math.Abs(dy) <= 1 {
			return sWallLit
		}
		return sNone
	})
	for y := int(cpy)/2 + rr/2 + 1; y < base; y++ {
		if s.at(w/2, y).ch == 0 {
			s.put(w/2, y, '║', sWall, sNone)
		}
	}
	s.text(w/2-2, base, "▀▀▀▀▀", sTrim, sNone)
	s.put(w/2, max(0, base-th), '✦', sGlow, sNone)
	return s.trimTop()
}
