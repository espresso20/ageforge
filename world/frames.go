package world

import "github.com/espresso20/ageforge/mapmodel"

// cellLy puts a cell of the 240x96 galaxy raster in light-years from the
// galaxy's centre, where the seeded plan keeps the home star. The raster's
// 110 columns are 50000 ly; a row is twice a column, as on screen.
func cellLy(p Pt) Vec {
	return Vec{float64(p.X-CellsW/2) * 50000 / GalRX, float64(p.Y-CellsH/2) * 2 * 50000 / GalRX}
}

// galRadiusLy is the galaxy disc's radius.
const galRadiusLy = 50000

// solLy is Sol's place in the galaxy frame: out on the home arm, as the real
// one sits on the Orion spur, not at the centre.
func solLy(g *Galaxy) Vec { return cellLy(g.Home[len(g.Home)*55/100]) }

func polar(r float64, a int) Vec {
	return Vec{r * float64(Cos4096(a)) / 1024, r * float64(Sin4096(a)) / 1024}
}

// galaxyFrame is the home galaxy: its disc and home arm, and the people's
// territory on it (lit systems, civilizations met, the harbinger's rift).
func galaxyFrame(m *mapmodel.Model, seed int64, g *Galaxy, t Territory) *Frame {
	mw := lgVec(lgW*26/100, lgH*46/100)
	f := &Frame{Kind: FrameGalaxy, Key: "milky_way", Unit: UnitLy, Anchor: "milky_way", Pos: mw,
		Radius: galRadiusLy, Band: Band{Min: 15000, Max: 240000}}
	shape := g.Shape
	f.Objects = append(f.Objects, Object{Kind: KindGalaxy, Key: "disc", Label: "milky_way", Inspect: "milky_way",
		Size: galRadiusLy, Tone: ToneStar, Weight: 255, Shape: &shape, Seed: seed})
	path := make([]Vec, len(g.Home))
	for i, p := range g.Home {
		path[i] = cellLy(p)
	}
	f.Objects = append(f.Objects, Object{Kind: KindFeature, Key: "home_arm", Path: path, Tone: ToneArm, Weight: 200})
	f.Objects = append(f.Objects, Object{Kind: KindStar, Key: "sol", Label: "sol", Inspect: "sol", Pos: solLy(g), Tone: ToneSun, Weight: 255, Pick: true,
		Star: &Look{Tone: ToneSun, Size: 1, Weight: 255, Extra: ToneWhite}})
	// The Magellanic Clouds, pulled in from 160 kly to sit in the Galaxy band's
	// view, below the disc as in the Local Group; radii from their real diameters.
	for _, c := range [...]struct {
		key  string
		pos  Vec
		r    int
		look GalaxyLook
	}{{"large_magellanic_cloud", Vec{68600, 58300}, 7000, GalaxyLook{Q: 170, Stars: 14, Irr: true, Bar: 60}},
		{"small_magellanic_cloud", Vec{84600, 62300}, 3500, GalaxyLook{Q: 170, Stars: 14, Irr: true}}} {
		c.look.R = c.r
		f.Objects = append(f.Objects, Object{Kind: KindGalaxy, Key: c.key, Label: c.key, Inspect: c.key, Pos: c.pos, Size: float64(c.r),
			Tone: ToneCool, Weight: 255, Galaxy: &c.look, Pick: true})
	}
	cols := Colonies(m, t)
	for _, c := range cols {
		sec := t.Sectors[lineIndex[c.Line]]
		for i := 0; i < c.Systems; i++ {
			f.Objects = append(f.Objects, Object{Kind: KindSector, Key: c.Line + "_" + itoa(i), Pos: cellLy(sec.Cells[i]),
				Line: c.Line, Count: c.Count, Tier: c.Tier, Gated: "lineage", Tone: ToneGold, Weight: 255})
		}
	}
	var civs []*mapmodel.Faction
	var keys []string
	for i := range m.Factions {
		if c := &m.Factions[i]; c.Discovered {
			civs = append(civs, c)
			keys = append(keys, c.Key)
		}
	}
	for i, slot := range PlaceRivals(seed, keys, len(t.Rivals)) {
		if slot < 0 {
			continue
		}
		cells := t.Rivals[slot].Cells
		for j, c := range cells[:min(len(cells), RivalSystems(civs[i].Strength))] {
			f.Objects = append(f.Objects, Object{Kind: KindRival, Key: civs[i].Key + "_" + itoa(j), Pos: cellLy(c),
				Line: civs[i].Key, Gated: "civ met", Tone: ToneRed, Weight: 220, Pick: j == 0, Inspect: civs[i].Key})
		}
	}
	if m.Harbinger != nil {
		for i, c := range t.Rift {
			f.Objects = append(f.Objects, Object{Kind: KindRift, Key: "rift_" + itoa(i), Pos: cellLy(c), Gated: "harbinger", Tone: TonePink, Weight: 200})
		}
	}
	return f
}

// armForms is what each of the arm's named objects is, in SkyObjects order.
var armForms = [ArmObjects]Form{FormEmission, FormDark, FormCluster, FormCluster, FormRemnant, FormEmission, FormEmission, FormEmission}

// regionFrame is the Orion arm round Sol: the named nebulae and clusters at
// their true distance and bearing, the arm's pink knots, and Sol at the
// centre. The terminal map squeezes distance with a square root so the near
// ones are not on top of Sol; the renderer should do the same on screen.
func regionFrame(seed int64, colonies int) *Frame {
	f := &Frame{Kind: FrameRegion, Key: "orion_arm", Unit: UnitLy, Anchor: "sol", Radius: ArmReachLy,
		Band: Band{Min: 400, Max: 18000}}
	f.Objects = append(f.Objects, Object{Kind: KindFeature, Form: FormBand, Key: "arm_band", Size: ArmReachLy, Tone: ToneArm, Weight: 120,
		Pos: polar(0, 0), To: polar(ArmReachLy, ArmBandDeg*Turn/360)})
	for i, o := range SkyObjects[:ArmObjects] {
		a := (o.Lon - 60) * Turn / 360
		if o.Key == "horsehead_nebula" {
			a += ArmSkew * Turn / 360
		}
		pos := polar(float64(o.Ly), -a) // the screen's y runs down
		tone := ToneBlue
		if o.Key == "hyades" {
			tone = ToneAmber // the old cluster's stars are orange giants
		}
		f.Objects = append(f.Objects, Object{Kind: KindFeature, Form: armForms[i], Key: o.Key, Label: o.Key, Inspect: o.Key, Pos: pos,
			Size: float64(ArmRadii[i]) * ArmReachLy / 100, Tone: tone, Weight: 220, Pick: true})
	}
	wd := ArmReachLy / 5
	for i := 0; i < 11; i++ {
		u := float64((Roll(2*ArmReachLy, seed, 7051, int64(i)) - ArmReachLy) * 11 / 10)
		v := float64((Roll(2*wd, seed, 7052, int64(i)) - wd) / 2)
		c, s := float64(Cos4096(ArmBandDeg*Turn/360))/1024, float64(Sin4096(ArmBandDeg*Turn/360))/1024
		f.Objects = append(f.Objects, Object{Kind: KindFeature, Form: FormKnot, Key: "knot_" + itoa(i), Pos: Vec{u*c + v*s, -u*s + v*c},
			Size: float64(2+Roll(3, seed, 7053, int64(i))) * ArmReachLy / 100, Tone: TonePink, Weight: 90})
	}
	f.Objects = append(f.Objects, Object{Kind: KindFeature, Form: FormTerritory, Key: "territory", Size: float64(ArmReachLy * (10 + min(colonies, 17)) / 100),
		Tone: ToneGold, Weight: 24, Gated: "lineage"})
	f.Objects = append(f.Objects, Object{Kind: KindStar, Key: "sol", Label: "sol", Inspect: "sol", Tone: ToneSun, Weight: 255, Pick: true,
		Star: &Look{Tone: ToneSun, Size: 1, Weight: 255, Extra: ToneWhite}})
	return f
}

// neighbourhoodFrame is the stars within 13 ly of Sol, the colonies the
// people have lit among them and the warp lanes between those.
func neighbourhoodFrame(seed int64, cols []Colony) *Frame {
	f := &Frame{Kind: FrameNeighbourhood, Key: "sol_neighbourhood", Unit: UnitLy, Anchor: "sol", Radius: 14,
		Band: Band{Min: 3, Max: 60}}
	pos := make([]Vec, len(NearStars))
	for i, s := range NearStars {
		d := float64(s.Dist100) / 100
		pos[i] = polar(d, -(s.Lon * Turn / 360))
		l := s.Look
		f.Objects = append(f.Objects, Object{Kind: KindStar, Key: s.Key, Label: s.Key, Inspect: s.Key, Pos: pos[i],
			Tone: l.Tone, Weight: uint8(l.Weight), Star: &l, Pick: true})
	}
	f.Objects = append(f.Objects, Object{Kind: KindStar, Key: "sol", Label: "sol", Inspect: "sol", Tone: ToneSun, Weight: 255, Pick: true,
		Star: &Look{Tone: ToneSun, Size: 1, Weight: 255, Extra: ToneWhite}})
	if len(cols) == 0 {
		return f
	}
	at := ColonyStars(seed, cols)
	byStar := map[int]Colony{}
	for i, c := range cols {
		byStar[at[i]] = c
		f.Objects = append(f.Objects, Object{Kind: KindColony, Key: "colony_" + c.Line, Label: NearStars[at[i]].Key,
			Inspect: c.Line, Pos: pos[at[i]], Line: c.Line, Count: c.Count, Tier: c.Tier, Tone: ToneGold, Weight: 255,
			Gated: "lineage", Pick: true})
	}
	// lanes join the nearest node by the flattened distance the map draws
	pts := make([][2]int, len(NearStars))
	for i, p := range pos {
		pts[i] = [2]int{int(p.X * 100), int(p.Y * 50)}
	}
	for _, l := range Lanes([2]int{0, 0}, pts, byStar) {
		from := Vec{}
		if l.From >= 0 {
			from = pos[l.From]
		}
		f.Objects = append(f.Objects, Object{Kind: KindLane, Key: "lane_" + NearStars[l.To].Key, Pos: from, To: pos[l.To],
			Line: l.Line, Flow: l.Traffic, Gated: "lineage", Tone: ToneGold, Weight: 200})
	}
	return f
}

// systemFrame is Sol's system in AU: the star, the eight planets, the belts,
// the dwarf and the comet, with a close-up frame of the home world under it.
// star says whether the system hangs in the galaxy (it is anchored on Sol) or
// stands alone as the whole map (space and interstellar ages).
func systemFrame(m *mapmodel.Model, seed int64, star bool) *Frame {
	f := &Frame{Kind: FrameSystem, Key: "sol_system", Unit: UnitAU, Radius: GateAU * 1.3, Band: Band{Min: 0.5, Max: 140}}
	if star {
		f.Anchor = "sol"
	}
	sp := SolSpec
	f.Objects = append(f.Objects, Object{Kind: KindStar, Key: "sol", Label: "sol", Inspect: "sol", Size: 0.05, Tone: sp.Star.Mid,
		Weight: 255, Pick: true, Star: &Look{Tone: sp.Star.Mid, Size: 3, Weight: 255, Extra: sp.Star.Core}})
	for i, p := range sp.Planets {
		kind := "rocky"
		switch {
		case p.Ring || p.Spot:
			kind = "gas"
		case p.Orbit > 85:
			kind = "ice"
		}
		ph := OrbitPhase(seed, i)
		f.Objects = append(f.Objects, Object{Kind: KindPlanet, Key: SolNames[i], Label: SolNames[i], Inspect: SolNames[i],
			Pos: polar(SolAU[i], ph), Size: float64(p.Size), Tone: p.Base.Tone, Weight: 255, Pick: true,
			Body: &Body{Spec: p, Orbit: SolAU[i], Period: p.OrbitPeriod(), Phase: ph, Rings: p.Ring, Moons: p.Moons,
				Kind: kind, Earths: SolEarths[i], StationN: p.Stations}})
	}
	for i := range sp.Belts {
		f.Objects = append(f.Objects, Object{Kind: KindBelt, Key: "asteroid_belt", Label: "asteroid_belt", Inspect: "asteroid_belt",
			Size: BeltAU, Tone: ToneRock, Weight: 160, Pick: true, Body: &Body{Orbit: BeltAU, Kind: "belt"}})
		_ = i
	}
	if sp.Dwarf {
		ph := OrbitPhase(seed, len(sp.Planets))
		f.Objects = append(f.Objects, Object{Kind: KindPlanet, Key: "pluto", Label: "pluto", Inspect: "pluto", Pos: polar(PlutoAU, ph),
			Size: 8, Tone: ToneRock, Weight: 200, Pick: true, Body: &Body{Orbit: PlutoAU, Phase: ph, Period: 256 * 1200, Kind: "dwarf", Earths: 0.19}})
		f.Objects = append(f.Objects, Object{Kind: KindBelt, Key: "kuiper_belt", Label: "kuiper_belt", Inspect: "kuiper_belt",
			Size: KuiperAU, Tone: ToneCool, Weight: 90, Body: &Body{Orbit: KuiperAU, Kind: "belt"}})
	}
	if sp.Comet {
		f.Objects = append(f.Objects, Object{Kind: KindComet, Key: "comet", Label: "comet", Inspect: "comet", Tone: ToneWhite, Weight: 230,
			Pick: true, Body: &Body{Orbit: KuiperAU, Period: CometCycle, Phase: Roll(Turn, seed, 5600), Kind: "comet"}})
	}
	if star || gateRaised(m.Wonders) {
		f.Objects = append(f.Objects, Object{Kind: KindGate, Key: "warp_gate", Label: GateWonder, Inspect: GateWonder,
			Pos: polar(GateAU, Roll(Turn, seed, 5305)), Tone: ToneCyan, Weight: 255, Pick: true, Gated: "wonder " + GateWonder})
	}
	f.Children = []*Frame{homeFrame(m, seed)}
	return f
}

// homeFrame is the home world and what circles it, in planet radii: the
// moon, the seeded stations and swarm, the starship, and one module for
// every building line standing in the space age's ring city.
func homeFrame(m *mapmodel.Model, seed int64) *Frame {
	hp := SolSpec.Planets[Home]
	f := &Frame{Kind: FramePlanet, Key: "earth", Unit: UnitPlanetRadii, Anchor: SolNames[Home], Radius: 6,
		Band: Band{Min: 1.5, Max: 800}}
	for k := 0; k < hp.Moons; k++ {
		ph := Roll(Turn, seed, 5201, int64(k))
		f.Objects = append(f.Objects, Object{Kind: KindMoon, Key: "the_moon", Label: "the_moon", Inspect: "the_moon",
			Pos: polar(60, ph), Size: 0.27, Tone: ToneRock, Weight: 255, Pick: true,
			Body: &Body{Parent: "earth", Orbit: 60, Period: 256 * 273, Phase: ph, Kind: "moon", Earths: 0.27}})
	}
	n := hp.Stations + Roll(2, seed, 5206)
	for i := 0; i < n; i++ {
		kind := (Roll(4, seed, 5207) + i) % 4
		orbit := 1.4 + 0.15*float64(Roll(5, seed, 5208, int64(i)))
		ph := Roll(Turn, seed, 5209, int64(i))
		f.Objects = append(f.Objects, Object{Kind: KindStation, Key: "station_" + itoa(i), Label: "space_station", Inspect: "space_station",
			Pos: polar(orbit, ph), Count: kind, Tone: ToneWhite, Weight: 240, Pick: true,
			Body: &Body{Parent: "earth", Orbit: orbit, Phase: ph, Period: 256 * 40, Kind: "station"}})
	}
	for i := 0; i < 10; i++ { // the satellite swarm
		ph := Roll(Turn, seed, 5310, int64(i))
		f.Objects = append(f.Objects, Object{Kind: KindStation, Key: "satellite_" + itoa(i), Label: "satellite",
			Pos: polar(1.12+float64(Roll(45, seed, 5311, int64(i)))/450, ph), Tone: ToneWhite, Weight: 110,
			Body: &Body{Parent: "earth", Kind: "satellite", Phase: ph, Period: 256 * 20}})
	}
	for i, n := 0, 2+Roll(2, seed, 5440); i < n; i++ { // the mega stations
		ph := Roll(Turn, seed, 5442) + i*Turn/n + Roll(Turn/8, seed, 5443, int64(i))
		f.Objects = append(f.Objects, Object{Kind: KindStation, Key: "mega_" + itoa(i), Label: "space_station", Inspect: "space_station",
			Pos: polar(1.3+float64(Roll(25, seed, 5444, int64(i)))/100*1.2, ph), Count: (Roll(3, seed, 5441) + i) % 3, Size: 0.3, Tone: ToneWhite,
			Weight: 255, Pick: true, Body: &Body{Parent: "earth", Kind: "mega", Phase: ph}})
	}
	f.Objects = append(f.Objects, Object{Kind: KindShip, Key: "starship", Label: "starship", Inspect: "starship", Tone: ToneWhite, Weight: 255,
		Pick: true, Pos: polar(1.2, Roll(Turn, seed, 5320))})
	if ageOf(m) == AgeSpace {
		for _, l := range m.Lineages {
			if _, ok := lineIndex[l.Key]; !ok || l.Count <= 0 {
				continue
			}
			f.Objects = append(f.Objects, Object{Kind: KindStation, Key: "module_" + l.Key, Label: l.Key, Inspect: l.Key, Line: l.Key,
				Count: l.Count, Tier: tierOf(m, l), Gated: "lineage", Tone: ToneWhite, Weight: 255, Pick: true,
				Pos:  polar(1.2, lineIndex[l.Key]*Turn/len(mapmodel.LineageOrder)),
				Body: &Body{Parent: "earth", Kind: "module"}})
		}
	}
	return f
}
