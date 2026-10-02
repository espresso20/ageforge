package mapmodel_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
)

var skyCat = mapmodel.NewCatalog()

func skyModel(t testing.TB, o fixture.Options) *mapmodel.Model {
	t.Helper()
	st := fixture.State(o)
	return mapmodel.NewBuilder(skyCat).Build(&st, nil)
}

// TestSkyScenes: the ground lasts to the Fusion Age; each of the last five
// ages opens a scene of its own, in order.
func TestSkyScenes(t *testing.T) {
	want := map[string]mapmodel.SkyScene{"space_age": mapmodel.SkyOrbit, "interstellar_age": mapmodel.SkyDeep,
		"galactic_age": mapmodel.SkyGalaxy, "quantum_age": mapmodel.SkyQuantum, "transcendent_age": mapmodel.SkyMandala}
	ground := true
	for i, age := range config.AgeOrder() {
		got := skyCat.SkySceneAt(i)
		if s, ok := want[age]; ok {
			ground = false
			if got != s {
				t.Errorf("%s: scene %d, want %d", age, got, s)
			}
			if skyCat.SkySceneAge(s) != i {
				t.Errorf("scene %d opens at %d, want %d", s, skyCat.SkySceneAge(s), i)
			}
			continue
		}
		if ground && got != mapmodel.SkyGround {
			t.Errorf("%s: scene %d before the Space Age", age, got)
		}
	}
	m := skyModel(t, fixture.Options{Age: "galactic_age", Seed: 3})
	if m.Sky() != mapmodel.SkyGalaxy {
		t.Errorf("the Galactic model's scene is %d", m.Sky())
	}
	if (*mapmodel.Model)(nil).Sky() != mapmodel.SkyGround {
		t.Error("a nil model is not on the ground")
	}
}

// TestSkyParts: in every scene with parts each lineage becomes something
// with a name and a symbol, no two lineages share a name, and the names
// keep the house style.
func TestSkyParts(t *testing.T) {
	retired := regexp.MustCompile(`(?i)\b(colour|centre|harbour|favour|villager)`)
	for s := mapmodel.SkyOrbit; s <= mapmodel.SkyQuantum; s++ {
		names := map[string]string{}
		for _, lin := range mapmodel.LineageOrder {
			p := mapmodel.SkyPartOf(s, lin)
			if p.Name == "" || p.Sym == mapmodel.SymNone {
				t.Errorf("scene %d %s: %+v", s, lin, p)
			}
			if o, ok := names[p.Name]; ok {
				t.Errorf("scene %d: %s and %s are both %q", s, o, lin, p.Name)
			}
			names[p.Name] = lin
			if strings.ContainsAny(p.Name, "—!") || retired.MatchString(p.Name) {
				t.Errorf("scene %d %s: %q breaks the house style", s, lin, p.Name)
			}
		}
	}
	if p := mapmodel.SkyPartOf(mapmodel.SkyGround, mapmodel.LinTrade); p.Sym != mapmodel.SymTrade {
		t.Errorf("trade on the ground is %+v", p)
	}
}

// TestSkyUnitsGrow: a lineage's units are its town tiles, in slot order;
// building more never takes a unit away or moves one to another slot, and
// it adds units: more buildings, more scene.
func TestSkyUnitsGrow(t *testing.T) {
	for _, age := range []string{"space_age", "interstellar_age", "galactic_age", "quantum_age"} {
		st := fixture.State(fixture.Options{Age: age, Seed: 5})
		b := mapmodel.NewBuilder(skyCat)
		m0 := b.Build(&st, nil)
		slots := func(m *mapmodel.Model) map[string]map[int]bool {
			out := map[string]map[int]bool{}
			for _, l := range m.SkyLineages() {
				out[l.Key] = map[int]bool{}
				for i, u := range l.Units {
					tt := m.Town.Tiles[u.Tile]
					if tt.Lineage != l.Key || u.N != tt.Ord || i > 0 && u.N <= l.Units[i-1].N {
						t.Fatalf("%s %s: unit %d is tile %+v (slot %d)", age, l.Key, i, tt, u.N)
					}
					out[l.Key][u.N] = true
				}
			}
			return out
		}
		before := slots(m0)
		total := m0.SkyUnitCount()
		for step := 1; step <= 3; step++ {
			st = fixture.Grow(st, step)
			m := b.Build(&st, nil)
			after := slots(m)
			for lin, ns := range before {
				for n := range ns {
					if !after[lin][n] {
						t.Errorf("%s step %d: %s lost its unit in slot %d", age, step, lin, n)
					}
				}
			}
			if m.SkyUnitCount() < total {
				t.Errorf("%s step %d: %d units, down from %d", age, step, m.SkyUnitCount(), total)
			}
			before, total = after, m.SkyUnitCount()
		}
		if total <= m0.SkyUnitCount() {
			t.Errorf("%s: three steps of building added no units (%d)", age, total)
		}
		order := map[string]int{}
		for i, l := range mapmodel.LineageOrder {
			order[l] = i
		}
		ls := m0.SkyLineages()
		for i := 1; i < len(ls); i++ {
			if order[ls[i].Key] < order[ls[i-1].Key] {
				t.Errorf("%s: %s listed after %s", age, ls[i].Key, ls[i-1].Key)
			}
		}
	}
}

// TestMandala: one ring per era, each holding a mark for every type owned
// from its ages and nothing from the Transcendent Age, whose buildings are
// the crown; owning more types adds marks.
func TestMandala(t *testing.T) {
	m := skyModel(t, fixture.Options{Age: "transcendent_age", Seed: 7})
	rings := m.Mandala()
	if len(rings) != len(skyCat.Epochs) {
		t.Fatalf("%d rings for %d eras", len(rings), len(skyCat.Epochs))
	}
	last := skyCat.SkySceneAge(mapmodel.SkyMandala)
	marks := 0
	for e, r := range rings {
		if r.Epoch != e {
			t.Errorf("ring %d is era %d", e, r.Epoch)
		}
		for _, mk := range r.Marks {
			b := m.Building(mk.Key)
			switch {
			case b == nil || b.Count <= 0:
				t.Errorf("ring %d marks %s, which is not owned", e, mk.Key)
			case mk.Age >= last || skyCat.AgeEpoch[mk.Age] != e:
				t.Errorf("ring %d marks %s of age %d", e, mk.Key, mk.Age)
			case mk.Sym == mapmodel.SymNone:
				t.Errorf("ring %d: %s has no symbol", e, mk.Key)
			}
			marks++
		}
		if len(r.Marks) == 0 {
			t.Errorf("era %d has an empty ring in a built-up game", e)
		}
	}
	for _, mk := range m.Crown() {
		if mk.Age != last {
			t.Errorf("the crown holds %s of age %d", mk.Key, mk.Age)
		}
	}
	if len(m.Crown()) == 0 {
		t.Error("no crown")
	}
	st := fixture.State(fixture.Options{Age: "transcendent_age", Seed: 7})
	st = fixture.Grow(st, 1)
	for k, d := range config.BuildingByKey() {
		if _, ok := st.Buildings[k]; !ok && d.RequiredAge == "bronze_age" && d.Category != "wonder" {
			st.Buildings[k] = st.Buildings["hut"]
			bs := st.Buildings[k]
			bs.Count, bs.IsLegacy = 2, true
			st.Buildings[k] = bs
		}
	}
	more := mapmodel.NewBuilder(skyCat).Build(&st, nil)
	n := 0
	for _, r := range more.Mandala() {
		n += len(r.Marks)
	}
	if n <= marks {
		t.Errorf("owning more types left the mandala at %d marks (was %d)", n, marks)
	}
	if skyModel(t, fixture.Options{Age: "space_age", Seed: 7}).Mandala() == nil {
		t.Error("the mandala needs a catalogue, nothing else")
	}
}

// TestSkyMoverRoster: the sky arc's movers come with the scene they belong
// to and travel open space, with the roster's house style.
func TestSkyMoverRoster(t *testing.T) {
	ages := config.AgeOrder()
	space := skyCat.SkySceneAge(mapmodel.SkyOrbit)
	keys := map[string]bool{}
	for k := mapmodel.Mover(1); k < mapmodel.NumMovers; k++ {
		keys[k.Info().Key] = true
	}
	for i := 0; i < mapmodel.NumSkyMovers; i++ {
		k := mapmodel.SkyMoverFirst + mapmodel.Mover(i)
		in := k.Info()
		switch {
		case in.Key == "" || in.Name == "" || in.Title == "" || len(in.Lines) == 0:
			t.Errorf("sky mover %d is missing a key, name, title or line: %+v", i, in)
		case keys[in.Key]:
			t.Errorf("mover key %q twice", in.Key)
		case in.From < space || in.From >= len(ages):
			t.Errorf("%s arrives at %d, before the Space Age or never", in.Key, in.From)
		case in.Until >= 0 && in.Until < in.From:
			t.Errorf("%s retires before it arrives", in.Key)
		case in.Way != mapmodel.WaySpace:
			t.Errorf("%s travels way %d, not open space", in.Key, in.Way)
		case in.Sym == mapmodel.SymNone:
			t.Errorf("%s has no symbol", in.Key)
		}
		keys[in.Key] = true
		for _, s := range append([]string{in.Name, in.Title}, in.Lines...) {
			if strings.ContainsAny(s, "—!") {
				t.Errorf("%s: %q breaks the house style", in.Key, s)
			}
		}
		if !mapmodel.Introduced(k, in.From) || mapmodel.Introduced(k, in.From-1) {
			t.Errorf("%s: Introduced disagrees with From %d", in.Key, in.From)
		}
		for a := 0; a < in.From; a++ {
			for _, mv := range mapmodel.MoversAt(a) {
				if mv == k {
					t.Errorf("%s is about in age %d, before it arrives", in.Key, a)
				}
			}
		}
	}
	for _, c := range []struct {
		k   mapmodel.Mover
		age string
	}{{mapmodel.MoverMiningDrone, "space_age"}, {mapmodel.MoverGenShip, "interstellar_age"},
		{mapmodel.MoverStarship, "galactic_age"}, {mapmodel.MoverAlienShip, "galactic_age"},
		{mapmodel.MoverPhaseShip, "quantum_age"}, {mapmodel.MoverMote, "transcendent_age"}} {
		if got := ages[c.k.Info().From]; got != c.age {
			t.Errorf("%s arrives in the %s, want the %s", c.k.Info().Key, got, c.age)
		}
	}
	if mapmodel.MoverAlienShip.Info().Sym != mapmodel.SymUFO {
		t.Error("the alien ship is not the visitor's saucer")
	}
}

// TestSkyGlyphs: the sky table has a range of its own past the ground
// table, and every symbol in it draws one cell in every tier and folds to
// ASCII.
func TestSkyGlyphs(t *testing.T) {
	for _, s := range []mapmodel.Sym{mapmodel.SymSkyHub, mapmodel.SymSkySolar, mapmodel.SymSkyYard, mapmodel.SymSkyCity,
		mapmodel.SymSkyAsteroid, mapmodel.SymSkyColony, mapmodel.SymSkyGate, mapmodel.SymMiningDrone,
		mapmodel.SymGenShip, mapmodel.SymStarship, mapmodel.SymPhaseShip, mapmodel.SymMote} {
		g := mapmodel.G(s)
		if g == mapmodel.G(mapmodel.SymNone) {
			t.Errorf("sym %d resolves to nothing", s)
			continue
		}
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
			r := mapmodel.R(s, tier)
			if uniseg.StringWidth(string(r)) != 1 {
				t.Errorf("sym %d in %s: %q is not one cell", s, tier, r)
			}
			if a := mapmodel.Fold(r, mapmodel.TierASCII); a < 0x20 || a > 0x7e {
				t.Errorf("sym %d in %s: %q folds to %q", s, tier, r, a)
			}
		}
		if g.ASCII < 0x21 || g.ASCII > 0x7e || g.Unicode < 0x80 {
			t.Errorf("sym %d: %+v", s, g)
		}
	}
	if mapmodel.G(mapmodel.SymBuild) == mapmodel.G(mapmodel.SymSkyHub) {
		t.Error("the sky table overlaps the ground table")
	}
}

// TestSkyPalettes: every sky scene has a colour for every ink, and next
// door scenes never share their void, their main structure or their glow.
func TestSkyPalettes(t *testing.T) {
	for s := mapmodel.SkyOrbit; s < mapmodel.NumSkyScenes; s++ {
		p := mapmodel.SkyPaletteOf(s)
		for ink, h := range p {
			if h == theme.SpaceNone {
				t.Errorf("scene %d ink %d has no colour", s, ink)
			}
		}
		if s == mapmodel.SkyOrbit {
			continue
		}
		q := mapmodel.SkyPaletteOf(s - 1)
		for _, ink := range []mapmodel.SkyInk{mapmodel.InkVoid, mapmodel.InkFrame, mapmodel.InkGlow} {
			if p[ink] == q[ink] {
				t.Errorf("scenes %d and %d share ink %d", s-1, s, ink)
			}
		}
	}
	for i, h := range mapmodel.SkyIridescent {
		if !theme.SpaceColor(h).Valid() || i > 0 && h == mapmodel.SkyIridescent[i-1] {
			t.Errorf("iridescent stop %d", i)
		}
	}
	for _, a := range mapmodel.Aliens {
		if a.Color == "" || len(a.Lines) == 0 || a.Hue == theme.SpaceNone {
			t.Errorf("alien %+v", a)
		}
	}
}
