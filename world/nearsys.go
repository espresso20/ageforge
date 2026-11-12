package world

import (
	"math"

	"github.com/espresso20/ageforge/mapmodel"
)

// The near stars' systems. Sol's is a literal; every other star of NearStars
// gets one laid out from the save's seed and the star's own key, the same each
// time. Real exoplanets are fixed data; the rest is rolled from the star's class.

// exoplanet is a planet of a near star that is well known, at the orbit and
// radius the catalogues give it.
type exoplanet struct {
	Star, Letter string
	AU, Earths   float64
	Kind         string
}

var exoplanets = [...]exoplanet{
	{"barnards_star", "b", .023, .37, "rocky"},
	{"epsilon_eridani", "b", 3.4, 11, "gas"},
	{"tau_ceti", "e", .54, 1.5, "rocky"},
	{"tau_ceti", "f", 1.33, 1.8, "rocky"},
	{"teegardens_star", "b", .0259, 1.02, "rocky"},
	{"teegardens_star", "c", .0455, 1.04, "rocky"},
	{"ross_128", "b", .0496, 1.1, "rocky"},
	{"epsilon_indi", "b", 2.9, 11, "gas"},
}

// pairAU is how far a companion star draws from the primary, in AU: the real
// separation, or 40 for the wide pairs, which would leave the frame.
var pairAU = map[string]float64{"alpha_centauri": 23, "sirius": 20, "procyon": 15, "luyten_726_8": 5.5}

// starSpec is how a star of the near catalogue's class draws in its own system:
// its tones, how big it shows, and what light it gives (solar units, which
// sets where the frost line and the temperate orbits lie).
func starSpec(l Look) (sp StarSpec, lum float64) {
	switch l.Tone {
	case ToneRed:
		return StarSpec{Core: ToneAmber, Mid: ToneRed, Rim: ToneCore, Size: 2}, .02
	case ToneAmber:
		return StarSpec{Core: ToneSun, Mid: ToneAmber, Rim: ToneCore, Size: 2}, .3
	case ToneCool:
		return StarSpec{Core: ToneWhite, Mid: ToneCool, Rim: ToneCore, Size: 4}, 25
	}
	return StarSpec{Core: ToneWhite, Mid: ToneSun, Rim: ToneCore, Size: 3}, 1.2
}

// System is the system of the near star of the given key: Sol's, or one
// generated from the seed and the colonies the people have lit there. It is
// built on first ask and kept; nil for a key that is not a star of the
// neighbourhood, or when the world holds no neighbourhood.
func (w *World) System(star string) *Frame {
	nb := w.Root.Find(FrameNeighbourhood)
	if nb == nil {
		return nil
	}
	if star == "sol" {
		if len(nb.Children) > 0 {
			return nb.Children[0]
		}
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if f := w.sys[star]; f != nil {
		return f
	}
	idx := -1
	for i, s := range NearStars {
		if s.Key == star {
			idx = i
		}
	}
	if idx < 0 {
		return nil
	}
	var cols []Object
	for _, o := range nb.Objects {
		if o.Kind == KindColony && o.Label == star {
			cols = append(cols, o)
		}
	}
	if w.sys == nil {
		w.sys = map[string]*Frame{}
	}
	f := starSystem(w.Seed, idx, cols)
	w.sys[star] = f
	return f
}

// starSystem lays out the system of NearStars[idx]: its star (and companion),
// the planets, a belt now and then, and the people's stations where colonies
// stand. cols are the colony markers of the neighbourhood at this star.
func starSystem(seed int64, idx int, cols []Object) *Frame {
	ns := NearStars[idx]
	sp, lum := starSpec(ns.Look)
	f := &Frame{Kind: FrameSystem, Key: ns.Key + "_system", Unit: UnitAU, Radius: GateAU * 1.3, Band: Band{Min: 0.5, Max: 140}, Anchor: ns.Key}
	f.Objects = append(f.Objects, Object{Kind: KindStar, Key: ns.Key, Label: ns.Key, Inspect: ns.Key, Size: 0.05, Tone: sp.Mid, Weight: 255, Pick: true,
		Star: &Look{Tone: sp.Mid, Size: sp.Size, Weight: 255, Extra: sp.Core}})
	reach := 32.0 // the outermost orbit
	if mate, ok := companion(ns.Look); ok {
		sep := 40.0
		if a, ok := pairAU[ns.Key]; ok {
			sep = a
		}
		reach = math.Min(reach, sep*.45)
		f.Objects = append(f.Objects, Object{Kind: KindStar, Key: ns.Key + "_B", Label: ns.Key + "_B", Inspect: ns.Key + "_B", Pos: polar(sep, Roll(Turn, seed, 6200, int64(idx))),
			Size: 0.03, Tone: mate, Weight: 255, Pick: true, Star: &Look{Tone: mate, Size: 2, Weight: 255}})
	}
	ps := genPlanets(seed, idx, lum, reach, len(cols))
	for i, p := range ps {
		ph := OrbitPhase(seed, 100+idx*10+i)
		f.Objects = append(f.Objects, Object{Kind: KindPlanet, Key: p.key, Label: p.key, Inspect: p.key, Pos: polar(p.au, ph), Size: float64(p.spec.Size),
			Tone: p.tone, Weight: 255, Pick: true,
			Body: &Body{Spec: p.spec, Orbit: p.au, Period: p.spec.OrbitPeriod(), Phase: ph, Rings: p.spec.Ring, Moons: p.spec.Moons, Kind: p.kind, Earths: p.earths}})
	}
	if len(ps) > 2 && Roll(100, seed, 6202, int64(idx)) < 40 { // a belt in the widest gap
		at, gap := 0, 0.0
		for i := 1; i < len(ps); i++ {
			if g := ps[i].au / ps[i-1].au; g > gap {
				at, gap = i, g
			}
		}
		if gap > 2 {
			au := math.Sqrt(ps[at].au * ps[at-1].au)
			f.Objects = append(f.Objects, Object{Kind: KindBelt, Key: ns.Key + "_belt", Label: "asteroid_belt", Inspect: "asteroid_belt",
				Size: au, Tone: ToneRock, Weight: 160, Pick: true, Body: &Body{Orbit: au, Kind: "belt"}})
		}
	}
	first := firstPlanet(f)
	for i, p := range ps {
		sp := p.spec
		sp.Close = clamp(sp.Size*28/10, 50, 110)
		cf := worldFrame(seed, p.key, sp, p.kind, p.tone, p.earths, 100+idx*10+i)
		for _, c := range cols {
			if Roll(len(ps), seed, 6400, int64(idx), int64(lineIndex[c.Line])) != i {
				continue
			}
			f.Objects[first+i].Body.Settled++
			cf.Objects = append(cf.Objects, Object{Kind: KindStation, Key: "module_" + c.Line, Label: c.Line, Inspect: c.Line, Line: c.Line, Count: c.Count,
				Tier: c.Tier, Gated: "lineage", Tone: ToneWhite, Weight: 255, Pick: true, Pos: polar(1.2, lineIndex[c.Line]*Turn/len(mapmodel.LineageOrder)),
				Body: &Body{Parent: p.key, Kind: "module"}})
		}
		f.Children = append(f.Children, cf)
	}
	return f
}

// firstPlanet is the index of the first planet object of the system.
func firstPlanet(f *Frame) int {
	for i := range f.Objects {
		if f.Objects[i].Kind == KindPlanet {
			return i
		}
	}
	return 0
}

// companion is the tone of the star beside the primary, if the catalogue has one:
// a pair, or a white dwarf (Alpha Centauri's third, Proxima, is too far to fit).
func companion(l Look) (Tone, bool) {
	switch {
	case l.Pair != ToneNone:
		return l.Pair, true
	case l.Extra == ToneWhite:
		return ToneWhite, true
	}
	return ToneNone, false
}

// A genPlanet is one planet before it is placed in the frame.
type genPlanet struct {
	key    string
	au     float64
	earths float64
	kind   string
	tone   Tone
	spec   PlanetSpec
}

const letters = "bcdefghi"

// genPlanets rolls the planets of NearStars[idx]: the known ones, then seeded
// worlds on orbits spaced by a ratio of the star's own, out to reach AU. A star
// the people have lit (lit > 0) always has at least one.
func genPlanets(seed int64, idx int, lum, reach float64, lit int) []genPlanet {
	ns := NearStars[idx]
	var ps []genPlanet
	used := map[byte]bool{}
	for _, e := range exoplanets {
		if e.Star == ns.Key {
			ps = append(ps, genPlanet{key: ns.Key + "_" + e.Letter, au: e.AU, earths: e.Earths, kind: e.Kind})
			used[e.Letter[0]] = true
		}
	}
	base, span := 1, 4 // planets rolled in addition: by class
	switch ns.Look.Tone {
	case ToneAmber:
		base, span = 2, 4
	case ToneSun:
		base, span = 3, 6
	case ToneCool:
		base, span = 1, 3
	}
	n := base + Roll(span, seed, 6203, int64(idx)) - 1 // -1: a star may have none
	if lit > 0 {
		n = max(n, 1)
	}
	root := math.Sqrt(lum)
	a := root * float64(15+Roll(40, seed, 6204, int64(idx))) / 100 / 2 // the innermost orbit, scaled by the star's light
	ratio := 1.45 + float64(Roll(50, seed, 6205, int64(idx)))/100
	for k := 0; len(ps) < len(letters) && k < n+8 && n > 0; k, a = k+1, a*ratio {
		if a > reach {
			break
		}
		near := false
		for _, p := range ps {
			near = near || (a < p.au*1.25 && a > p.au/1.25)
		}
		if near {
			continue
		}
		ps = append(ps, genPlanet{au: a})
		n--
	}
	sortPlanets(ps)
	next := 0
	for i := range ps {
		if ps[i].key == "" {
			for used[letters[next]] {
				next++
			}
			ps[i].key = ns.Key + "_" + letters[next:next+1]
			used[letters[next]] = true
		}
		rollFace(&ps[i], seed, idx, i, root)
	}
	return ps
}

// sortPlanets orders planets inner to outer (insertion: there are eight at most).
func sortPlanets(ps []genPlanet) {
	for i := 1; i < len(ps); i++ {
		for j := i; j > 0 && ps[j].au < ps[j-1].au; j-- {
			ps[j], ps[j-1] = ps[j-1], ps[j]
		}
	}
}

// rollFace fills in what a planet is and how it looks, from its orbit against
// the star's frost line and temperate zone (root is the square root of the
// star's light, in AU for a solar-type star).
func rollFace(p *genPlanet, seed int64, idx, i int, root float64) {
	r := func(n int, salt int64) int { return Roll(n, seed, 6210+salt, int64(idx), int64(i)) }
	if p.kind == "" {
		switch frost := 2.7 * root; {
		case p.au < frost && r(100, 0) < 90, p.au >= frost && r(100, 0) < 20:
			p.kind, p.earths = "rocky", float64(30+r(150, 1))/100
		case p.au < frost:
			p.kind, p.earths = "gas", float64(60+r(60, 1))/10
		case r(100, 2) < 60:
			p.kind, p.earths = "gas", float64(40+r(80, 1))/10
		default:
			p.kind, p.earths = "ice", float64(25+r(20, 1))/10
		}
	}
	sp := PlanetSpec{Size: 14 + int(12*math.Pow(p.earths, .6)), Orbit: artOrbit(p.au)}
	rock := [...]Tone{ToneRock, ToneRed, ToneSand, ToneRock, ToneAmber}
	switch p.kind {
	case "rocky":
		p.tone = rock[r(len(rock), 3)]
		sp.Base = S(p.tone, 255)
		sp.Cap = r(100, 4) < 30 && p.au > .9*root
		if r(100, 5) < 40 {
			sp.Blotch = S(rock[r(len(rock), 6)], 200)
		}
		sp.World = p.earths > .8 && p.earths < 1.7 && p.au > .95*root && p.au < 1.6*root
		if sp.World {
			p.tone = ToneBlue
		}
		sp.Moons = r(3, 7) % 2 * (r(2, 8) + 1)
	case "gas":
		pair := [...][2]Tone{{ToneRock, ToneAmber}, {ToneSand, ToneGold}, {ToneSun, ToneAmber}, {ToneCool, ToneBlue}, {ToneSand, ToneRock}}[r(5, 3)]
		p.tone = pair[1]
		for b := 0; b < 9; b++ {
			sp.Bands = append(sp.Bands, S(pair[b%2], uint8(205+r(50, int64(10+b)))))
		}
		sp.Spot, sp.Ring, sp.Moons = r(100, 4) < 30, r(100, 5) < 25, 2+r(7, 7)
	default: // ice
		p.tone = ToneCool
		sp.Base = S(ToneCool, 255)
		if r(100, 4) < 50 {
			sp.Bands = []Shade{S(TonePanel, 255), S(ToneCyan, 215), S(TonePanel, 255), S(TonePanel, 255), S(TonePanel, 255), S(ToneCyan, 215), S(TonePanel, 255), S(TonePanel, 255), S(TonePanel, 255)}
		}
		sp.Moons = 1 + r(5, 7)
	}
	p.spec = sp
}

// artOrbit is a planet's orbit as the art spacing of SolSpec has it, found by
// reading the AU between Sol's own planets, so the animation clock runs at the
// pace Sol's does at that distance.
func artOrbit(au float64) int {
	if au <= SolAU[0] {
		return SolSpec.Planets[0].Orbit
	}
	for i := 1; i < len(SolAU); i++ {
		if au <= SolAU[i] {
			a, b := SolSpec.Planets[i-1].Orbit, SolSpec.Planets[i].Orbit
			return a + int(float64(b-a)*(au-SolAU[i-1])/(SolAU[i]-SolAU[i-1]))
		}
	}
	return SolSpec.Planets[len(SolAU)-1].Orbit
}
