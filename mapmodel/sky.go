package mapmodel

import "github.com/espresso20/ageforge/theme"

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
	name string
	sym  Sym // SymNone: the lineage's own symbol (LineageSym in the Cosmic Era)
}

var skyParts = [NumSkyScenes]map[string]skyPartDef{
	SkyOrbit: {
		LinHousing: {"habitat module", SymHabitat}, LinFood: {"hydroponic bay", SymFieldHydro},
		LinStorage: {"depot pod", SymDataStore}, LinWood: {"bio-vat", SymNone},
		LinMines: {"asteroid mine", SymNone}, LinMetal: {"orbital foundry", SymNone},
		LinEngineer: {"shipyard", SymSkyYard}, LinEnergy: {"solar array", SymSkySolar},
		LinHarbor: {"docking port", SymNone}, LinHacker: {"relay satellite", SymNone},
		LinKnowledge: {"observatory", SymNone}, LinFaith: {"sanctuary", SymNone},
		LinCulture: {"zero-g gallery", SymNone}, LinMonument: {"beacon", SymNone},
		LinTrade: {"market dock", SymNone}, LinMilitary: {"moon base", SymNone},
		LinDiplomacy: {"embassy", SymNone},
	},
	SkyDeep: {
		LinHousing: {"colony ark", SymHabitat}, LinFood: {"farm world", SymFieldHydro},
		LinStorage: {"depot moon", SymDataStore}, LinWood: {"garden world", SymNone},
		LinMines: {"mining world", SymNone}, LinMetal: {"forge world", SymNone},
		LinEngineer: {"warp gate frame", SymSkyGate}, LinEnergy: {"stellar collector", SymSkySolar},
		LinHarbor: {"port world", SymNone}, LinHacker: {"relay beacon", SymNone},
		LinKnowledge: {"research world", SymNone}, LinFaith: {"shrine world", SymNone},
		LinCulture: {"arts world", SymNone}, LinMonument: {"memorial world", SymNone},
		LinTrade: {"trade world", SymNone}, LinMilitary: {"escort frigate", SymNone},
		LinDiplomacy: {"embassy world", SymNone},
	},
	SkyGalaxy: {
		LinHousing: {"habitat deck", SymHabitat}, LinFood: {"agri deck", SymFieldHydro},
		LinStorage: {"cargo hold", SymDataStore}, LinWood: {"bio-works", SymNone},
		LinMines: {"neutron star mine", SymNone}, LinMetal: {"stellar forge", SymNone},
		LinEngineer: {"repair pylon", SymNone}, LinEnergy: {"quasar tap", SymNone},
		LinHarbor: {"docking bay", SymNone}, LinHacker: {"uplink node", SymNone},
		LinKnowledge: {"research deck", SymNone}, LinFaith: {"shrine deck", SymNone},
		LinCulture: {"archive deck", SymNone}, LinMonument: {"beacon spire", SymNone},
		LinTrade: {"exchange deck", SymNone}, LinMilitary: {"fleet at dock", SymStarship},
		LinDiplomacy: {"embassy deck", SymNone},
	},
	SkyQuantum: {
		LinHousing: {"folded habitat", SymHabitat}, LinFood: {"probability garden", SymFieldHydro},
		LinStorage: {"entangled vault", SymDataStore}, LinWood: {"reality harvest", SymNone},
		LinMines: {"possibility mine", SymNone}, LinMetal: {"quantum works", SymNone},
		LinEngineer: {"reality forge", SymNone}, LinEnergy: {"zero-point well", SymNone},
		LinHarbor: {"phase dock", SymNone}, LinHacker: {"reality processor", SymNone},
		LinKnowledge: {"academy of maybes", SymNone}, LinFaith: {"hall of echoes", SymNone},
		LinCulture: {"art that might be", SymNone}, LinMonument: {"fixed point", SymNone},
		LinTrade: {"probability market", SymNone}, LinMilitary: {"war room", SymNone},
		LinDiplomacy: {"embassy of selves", SymNone},
	},
	SkyMandala: {},
}

// SkyPartOf is what lineage lin becomes in scene s. A lineage the table
// does not list (and every lineage on the ground or in the mandala) keeps
// its own name and symbol.
func SkyPartOf(s SkyScene, lin string) SkyPart {
	sym := LineageSym(lin, 6)
	name := LineageNames[lin]
	if s < NumSkyScenes {
		if d, ok := skyParts[s][lin]; ok {
			name = d.name
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
// model's scene. The units are the lineage's town tiles (ruins and legacy
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
	order := map[string]int{}
	for i, l := range LineageOrder {
		order[l] = i
	}
	// tiles come lineage by lineage already; keep LineageOrder explicit
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && order[out[j].Key] < order[out[j-1].Key]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
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
	Color string         // how the inspect line names it
	Hue   theme.SpaceHue // its colour
	Lines []string
}

// Aliens are the species of the Galactic Age's traffic, each in its own
// colour.
var Aliens = []AlienKind{
	{"green", theme.SpaceAlienGreen, []string{"A green saucer, waving its lights.", "A green saucer, trading in songs."}},
	{"violet", theme.SpaceAlienViolet, []string{"A violet saucer, humming a scale.", "A violet saucer, here for the view."}},
	{"amber", theme.SpaceAlienAmber, []string{"An amber saucer, late for something.", "An amber saucer, haggling by light."}},
}
