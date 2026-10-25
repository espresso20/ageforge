package mapmodel

import (
	"github.com/espresso20/ageforge/theme"

	"github.com/espresso20/ageforge/vibe"
)

// sky.go is the sky arc: from the Space Age on, both map styles leave the
// ground. Each of the last five ages is a scene of its own (SkyScene): the
// Space Age looks down on the planet from orbit, the Interstellar Age sees
// the home system shrink to an orrery among colony worlds, the Galactic Age
// is a starbase hub and its fleets, the Quantum Age bends reality, and the
// Transcendent Age is a mandala of every age the player has lived through.
//
// The scene is still the player's civilization. Every lineage becomes one
// kind of thing in a scene (SkyPartOf: housing is a habitat module in orbit,
// a colony ark in deep space), and its units are its town tiles seen from
// space (SkyLineages): unit n of a lineage always takes the lineage's n-th
// slot in a scene, and units only ever grow, so the scene grows as the
// player builds and never reshuffles. Like the rest of the model it is a
// pure function of the snapshot and the seed.

// SkyScene is what the maps draw for an age.
type SkyScene uint8

const (
	SkyGround  SkyScene = iota // before the Space Age: the town on the ground
	SkyOrbit                   // the Space Age: the planet below, the station in orbit
	SkyDeep                    // the Interstellar Age: deep space, the home system an orrery
	SkyGalaxy                  // the Galactic Age: the starbase hub and its fleets
	SkyQuantum                 // the Quantum Age: reality bends
	SkyMandala                 // the Transcendent Age: the mandala of the ages
	NumSkyScenes
)

// skySceneAges are the age keys that open each scene.
var skySceneAges = [NumSkyScenes]string{"", "space_age", "interstellar_age", "galactic_age", "quantum_age", "transcendent_age"}

// SkySceneAge is the index of the age that opens scene s, or -1 (the ground
// opens with the first age).
func (c *Catalog) SkySceneAge(s SkyScene) int {
	if s == SkyGround {
		return 0
	}
	if s >= NumSkyScenes {
		return -1
	}
	if i, ok := c.AgeIdx[skySceneAges[s]]; ok {
		return i
	}
	return -1
}

// SkySceneAt is the scene of the age with index age.
func (c *Catalog) SkySceneAt(age int) SkyScene {
	s := SkyGround
	for k := SkyOrbit; k < NumSkyScenes; k++ {
		if i := c.SkySceneAge(k); i >= 0 && age >= i {
			s = k
		}
	}
	return s
}

// Sky is the scene the maps draw for this model.
func (m *Model) Sky() SkyScene {
	if m == nil || m.Catalog == nil {
		return SkyGround
	}
	return m.Catalog.SkySceneAt(m.AgeIdx)
}

// SkyPart is what a lineage becomes in a sky scene: its name on the legend
// and in inspect lines, and its symbol. Most lineages keep the symbol they
// have on the ground (§ is always knowledge, $ is always trade); housing,
// fields and stores re-skin, as they do by epoch.
type SkyPart struct {
	Name string
	Sym  Sym
}

type skyPartDef struct {
	name vibe.Entry
	sym  Sym // SymNone: the lineage's own symbol (LineageSym in the Cosmic Era)
}

var skyParts = [NumSkyScenes]map[string]skyPartDef{
	SkyOrbit: {
		LinHousing: {vibe.E("map.sky.SkyOrbit.housing", "habitat module"), SymHabitat}, LinFood: {vibe.E("map.sky.SkyOrbit.food", "hydroponic bay"), SymFieldHydro},
		LinStorage: {vibe.E("map.sky.SkyOrbit.storage", "depot pod"), SymDataStore}, LinWood: {vibe.E("map.sky.SkyOrbit.organic_extraction", "bio-vat"), SymNone},
		LinMines: {vibe.E("map.sky.SkyOrbit.geological_extraction", "asteroid mine"), SymNone}, LinMetal: {vibe.E("map.sky.SkyOrbit.metallurgy", "orbital foundry"), SymNone},
		LinEngineer: {vibe.E("map.sky.SkyOrbit.engineering", "shipyard"), SymSkyYard}, LinEnergy: {vibe.E("map.sky.SkyOrbit.energy", "solar array"), SymSkySolar},
		LinHarbor: {vibe.E("map.sky.SkyOrbit.harbor", "docking port"), SymNone}, LinHacker: {vibe.E("map.sky.SkyOrbit.hacker", "relay satellite"), SymNone},
		LinKnowledge: {vibe.E("map.sky.SkyOrbit.knowledge", "observatory"), SymNone}, LinFaith: {vibe.E("map.sky.SkyOrbit.faith", "sanctuary"), SymNone},
		LinCulture: {vibe.E("map.sky.SkyOrbit.culture_arts", "zero-g gallery"), SymNone}, LinMonument: {vibe.E("map.sky.SkyOrbit.monument", "beacon"), SymNone},
		LinTrade: {vibe.E("map.sky.SkyOrbit.trade", "market dock"), SymNone}, LinMilitary: {vibe.E("map.sky.SkyOrbit.military", "moon base"), SymNone},
		LinDiplomacy: {vibe.E("map.sky.SkyOrbit.diplomacy", "embassy"), SymNone},
	},
	SkyDeep: {
		LinHousing: {vibe.E("map.sky.SkyDeep.housing", "colony ark"), SymHabitat}, LinFood: {vibe.E("map.sky.SkyDeep.food", "farm world"), SymFieldHydro},
		LinStorage: {vibe.E("map.sky.SkyDeep.storage", "depot moon"), SymDataStore}, LinWood: {vibe.E("map.sky.SkyDeep.organic_extraction", "garden world"), SymNone},
		LinMines: {vibe.E("map.sky.SkyDeep.geological_extraction", "mining world"), SymNone}, LinMetal: {vibe.E("map.sky.SkyDeep.metallurgy", "forge world"), SymNone},
		LinEngineer: {vibe.E("map.sky.SkyDeep.engineering", "warp gate frame"), SymSkyGate}, LinEnergy: {vibe.E("map.sky.SkyDeep.energy", "stellar collector"), SymSkySolar},
		LinHarbor: {vibe.E("map.sky.SkyDeep.harbor", "port world"), SymNone}, LinHacker: {vibe.E("map.sky.SkyDeep.hacker", "relay beacon"), SymNone},
		LinKnowledge: {vibe.E("map.sky.SkyDeep.knowledge", "research world"), SymNone}, LinFaith: {vibe.E("map.sky.SkyDeep.faith", "shrine world"), SymNone},
		LinCulture: {vibe.E("map.sky.SkyDeep.culture_arts", "arts world"), SymNone}, LinMonument: {vibe.E("map.sky.SkyDeep.monument", "memorial world"), SymNone},
		LinTrade: {vibe.E("map.sky.SkyDeep.trade", "trade world"), SymNone}, LinMilitary: {vibe.E("map.sky.SkyDeep.military", "escort frigate"), SymNone},
		LinDiplomacy: {vibe.E("map.sky.SkyDeep.diplomacy", "embassy world"), SymNone},
	},
	SkyGalaxy: {
		LinHousing: {vibe.E("map.sky.SkyGalaxy.housing", "habitat deck"), SymHabitat}, LinFood: {vibe.E("map.sky.SkyGalaxy.food", "agri deck"), SymFieldHydro},
		LinStorage: {vibe.E("map.sky.SkyGalaxy.storage", "cargo hold"), SymDataStore}, LinWood: {vibe.E("map.sky.SkyGalaxy.organic_extraction", "bio-works"), SymNone},
		LinMines: {vibe.E("map.sky.SkyGalaxy.geological_extraction", "neutron star mine"), SymNone}, LinMetal: {vibe.E("map.sky.SkyGalaxy.metallurgy", "stellar forge"), SymNone},
		LinEngineer: {vibe.E("map.sky.SkyGalaxy.engineering", "repair pylon"), SymNone}, LinEnergy: {vibe.E("map.sky.SkyGalaxy.energy", "quasar tap"), SymNone},
		LinHarbor: {vibe.E("map.sky.SkyGalaxy.harbor", "docking bay"), SymNone}, LinHacker: {vibe.E("map.sky.SkyGalaxy.hacker", "uplink node"), SymNone},
		LinKnowledge: {vibe.E("map.sky.SkyGalaxy.knowledge", "research deck"), SymNone}, LinFaith: {vibe.E("map.sky.SkyGalaxy.faith", "shrine deck"), SymNone},
		LinCulture: {vibe.E("map.sky.SkyGalaxy.culture_arts", "archive deck"), SymNone}, LinMonument: {vibe.E("map.sky.SkyGalaxy.monument", "beacon spire"), SymNone},
		LinTrade: {vibe.E("map.sky.SkyGalaxy.trade", "exchange deck"), SymNone}, LinMilitary: {vibe.E("map.sky.SkyGalaxy.military", "fleet at dock"), SymStarship},
		LinDiplomacy: {vibe.E("map.sky.SkyGalaxy.diplomacy", "embassy deck"), SymNone},
	},
	SkyQuantum: {
		LinHousing: {vibe.E("map.sky.SkyQuantum.housing", "folded habitat"), SymHabitat}, LinFood: {vibe.E("map.sky.SkyQuantum.food", "probability garden"), SymFieldHydro},
		LinStorage: {vibe.E("map.sky.SkyQuantum.storage", "entangled vault"), SymDataStore}, LinWood: {vibe.E("map.sky.SkyQuantum.organic_extraction", "reality harvest"), SymNone},
		LinMines: {vibe.E("map.sky.SkyQuantum.geological_extraction", "possibility mine"), SymNone}, LinMetal: {vibe.E("map.sky.SkyQuantum.metallurgy", "quantum works"), SymNone},
		LinEngineer: {vibe.E("map.sky.SkyQuantum.engineering", "reality forge"), SymNone}, LinEnergy: {vibe.E("map.sky.SkyQuantum.energy", "zero-point well"), SymNone},
		LinHarbor: {vibe.E("map.sky.SkyQuantum.harbor", "phase dock"), SymNone}, LinHacker: {vibe.E("map.sky.SkyQuantum.hacker", "reality processor"), SymNone},
		LinKnowledge: {vibe.E("map.sky.SkyQuantum.knowledge", "academy of maybes"), SymNone}, LinFaith: {vibe.E("map.sky.SkyQuantum.faith", "hall of echoes"), SymNone},
		LinCulture: {vibe.E("map.sky.SkyQuantum.culture_arts", "art that might be"), SymNone}, LinMonument: {vibe.E("map.sky.SkyQuantum.monument", "fixed point"), SymNone},
		LinTrade: {vibe.E("map.sky.SkyQuantum.trade", "probability market"), SymNone}, LinMilitary: {vibe.E("map.sky.SkyQuantum.military", "war room"), SymNone},
		LinDiplomacy: {vibe.E("map.sky.SkyQuantum.diplomacy", "embassy of selves"), SymNone},
	},
	SkyMandala: {},
}

// SkyPartOf is what lineage lin becomes in scene s. A lineage the table
// does not list (and every lineage on the ground or in the mandala) keeps
// its own name and symbol.
func SkyPartOf(s SkyScene, lin string) SkyPart {
	sym := LineageSym(lin, 6)
	name := LineageName(lin)
	if s < NumSkyScenes {
		if d, ok := skyParts[s][lin]; ok {
			name = d.name.String()
			if d.sym != SymNone {
				sym = d.sym
			}
		}
	}
	if name == "" {
		name = lin
	}
	return SkyPart{Name: name, Sym: sym}
}

// SkyUnit is one unit of a lineage in a sky scene: a town tile seen from
// space.
type SkyUnit struct {
	Tile int // index into Model.Town.Tiles
	N    int // its ordinal in the lineage (TownTile.Ord): the slot it takes
}

// SkyLineage is one lineage's units in a sky scene, in slot order.
type SkyLineage struct {
	Key   string
	Part  SkyPart
	Units []SkyUnit
}

// SkyLineages lists every lineage with units, in LineageOrder, for the
// model's scene (the town lays its tiles out lineage by lineage, in that
// order). The units are the lineage's town tiles (ruins and legacy
// tiles included: an old section of the station, a dim colony), so a
// lineage has as many units as it has tiles (TownTiles of its copies),
// which grows sub-linearly with what is built and never shrinks while the
// buildings stand.
func (m *Model) SkyLineages() []SkyLineage {
	s := m.Sky()
	idx := map[string]int{}
	var out []SkyLineage
	for i, t := range m.Town.Tiles {
		j, ok := idx[t.Lineage]
		if !ok {
			j = len(out)
			idx[t.Lineage] = j
			out = append(out, SkyLineage{Key: t.Lineage, Part: SkyPartOf(s, t.Lineage)})
		}
		out[j].Units = append(out[j].Units, SkyUnit{Tile: i, N: t.Ord})
	}
	return out
}

// SkyUnitCount is how many units every lineage has in all.
func (m *Model) SkyUnitCount() int { return len(m.Town.Tiles) }

// MandalaRing is one ring of the Transcendent mandala: an era of the
// player's history, the Stone Era innermost and the Cosmic Era (to the
// Quantum Age) outermost.
type MandalaRing struct {
	Epoch int
	Marks []MandalaMark
}

// MandalaMark is one building type the player raised in one of an era's
// ages, in the symbol it had then, or a wonder.
type MandalaMark struct {
	Age     int
	Key     string
	Lineage string
	Sym     Sym
	Wonder  bool
	Count   int
}

// Mandala lays the player's history out in rings, one for every era the
// player passed through this run (reached one of its ages before the
// mandala's own), oldest first. Each ring holds a mark for every building
// type owned from that era's ages, in the catalogue's skyline order (the
// most prominent first); an era with nothing still standing is a ring with
// no marks. The mandala's own age is not in it: its buildings are the
// Crown.
func (m *Model) Mandala() []MandalaRing {
	cat := m.Catalog
	if cat == nil {
		return nil
	}
	last := cat.SkySceneAge(SkyMandala)
	rings := make([]MandalaRing, len(cat.Epochs))
	passed := make([]bool, len(cat.Epochs))
	for e := range rings {
		rings[e].Epoch = e
	}
	for a := 0; a < len(cat.ByAge) && (last < 0 || a < last); a++ {
		e := cat.AgeEpoch[a]
		passed[e] = passed[e] || m.ReachedAge(a)
		for _, d := range cat.ByAge[a] {
			if mk, ok := m.mandalaMark(d, e); ok {
				rings[e].Marks = append(rings[e].Marks, mk)
			}
		}
	}
	out := make([]MandalaRing, 0, len(rings))
	for e, r := range rings {
		if passed[e] {
			out = append(out, r)
		}
	}
	return out
}

// Crown is the mandala's centre: a mark for every building type owned from
// the Transcendent Age itself.
func (m *Model) Crown() []MandalaMark {
	cat := m.Catalog
	if cat == nil {
		return nil
	}
	a := cat.SkySceneAge(SkyMandala)
	if a < 0 || a >= len(cat.ByAge) {
		return nil
	}
	var out []MandalaMark
	for _, d := range cat.ByAge[a] {
		if mk, ok := m.mandalaMark(d, cat.AgeEpoch[a]); ok {
			out = append(out, mk)
		}
	}
	return out
}

func (m *Model) mandalaMark(d *Def, epoch int) (MandalaMark, bool) {
	b := m.byKey[d.Key]
	if b == nil || b.Count <= 0 {
		return MandalaMark{}, false
	}
	mk := MandalaMark{Age: d.Age, Key: d.Key, Lineage: d.Lineage, Sym: LineageSym(d.Lineage, epoch), Wonder: d.Wonder,
		Count: b.Count}
	if d.Wonder {
		mk.Sym = SymWonder
	}
	return mk, true
}

// AlienKind is one of the species whose ships are ordinary traffic in the
// Galactic Age (the rare visitor of the earlier ages, graduated).
type AlienKind struct {
	Color vibe.Entry     // how the inspect line names it
	Hue   theme.SpaceHue // its colour
	Lines []vibe.Entry
}

// Aliens are the species of the Galactic Age's traffic, each in its own
// colour.
var Aliens = []AlienKind{
	{vibe.E("map.alien.01.color", "green"), theme.SpaceAlienGreen, []vibe.Entry{vibe.E("map.alien.01.line.01", "A green saucer, waving its lights."), vibe.E("map.alien.01.line.02", "A green saucer, trading in songs.")}},
	{vibe.E("map.alien.02.color", "violet"), theme.SpaceAlienViolet, []vibe.Entry{vibe.E("map.alien.02.line.01", "A violet saucer, humming a scale."), vibe.E("map.alien.02.line.02", "A violet saucer, here for the view.")}},
	{vibe.E("map.alien.03.color", "amber"), theme.SpaceAlienAmber, []vibe.Entry{vibe.E("map.alien.03.line.01", "An amber saucer, late for something."), vibe.E("map.alien.03.line.02", "An amber saucer, haggling by light.")}},
}
