package world

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// zones.go: the Milky Way as the people's territory. The disc is cut into
// sectors (a core and three rings of eight wedges, the wedges turned with the
// arms, the cut phased by the seed) and the people hold a growing set of them.
//
// How far the people have spread is one number, the lit systems: the sum of
// Colony.Systems over the building lines. The held sectors are
//
//	held = clamp(1 + lit/4, 1, ZoneTotal)
//
// taken in order of how near they lie to Sol's own sector (with a seeded
// jitter), so more lit systems never lose a sector, and Sol's is always held.

const (
	zoneRings   = 3
	zoneWedges  = 8
	ZoneTotal   = 1 + zoneRings*zoneWedges // the core, then ring by ring
	zoneCore    = .22                      // the core's radius, in disc radii
	zoneTwist   = 1.1                      // radians the wedges turn from core to rim
	zoneStars   = 3                        // seeded stars in each sector
	zoneRingW   = (1 - zoneCore) / zoneRings
	genStarBase = 1000   // ids of the stars a place makes up for itself start here, clear of the sectors' own
	genIdx      = 100000 // a generated star's system rolls as star genIdx+id: clear of the catalogue's indexes
)

// Zones is the sector plan of one galaxy and which of it the people hold.
type Zones struct {
	Held   [ZoneTotal]bool
	Names  [ZoneTotal]string // the objects' keys, "~Name Reach"
	shape  GalaxyShape
	phase  float64
	Sol    int // the sector Sol is in
	mask   [zoneMask * zoneMask]uint8
	maskMu sync.Once
}

const zoneMask = 128 // the face-on raster of the territory the renderer tints

// ZonesHeld is how many sectors lit systems hold.
func ZonesHeld(lit int) int { return clamp(1+lit/4, 1, ZoneTotal) }

// face is the disc point (-1..1) under a galaxy-frame point in light-years:
// the inverse of how the renderer lays the disc (tilt, flattening, rows 0.8
// as tall as columns).
func (s GalaxyShape) face(p Vec) (u, v float64) {
	a := float64(s.Tilt) * 2 * math.Pi / Turn
	c, sn, q := math.Cos(a), math.Sin(a), float64(max(s.Flat, 8))/100
	fx, fy := p.X/galRadiusLy, p.Y/(galRadiusLy*.8)
	return fx*c + fy*sn, (-fx*sn + fy*c) / q
}

func (s GalaxyShape) unface(u, v float64) Vec {
	a := float64(s.Tilt) * 2 * math.Pi / Turn
	c, sn, q := math.Cos(a), math.Sin(a), float64(max(s.Flat, 8))/100
	w := v * q
	return Vec{(u*c - w*sn) * galRadiusLy, (u*sn + w*c) * galRadiusLy * .8}
}

// NewZones plans the sectors of the galaxy of seed and shape, with Sol at sol
// (galaxy-frame light-years), and marks the ones lit systems hold.
func NewZones(seed int64, sh GalaxyShape, sol Vec, lit int) *Zones {
	z := &Zones{shape: sh, phase: float64(Roll(Turn, seed, 9000)) * 2 * math.Pi / Turn}
	su, sv := sh.face(sol)
	z.Sol = z.At(su, sv)
	var rank [ZoneTotal]float64
	order := make([]int, ZoneTotal)
	for s := range rank {
		cu, cv := z.centre(s)
		rank[s] = math.Hypot(cu-su, cv-sv) + float64(Roll(100, seed, 9001, int64(s)))/400
		order[s] = s
	}
	rank[z.Sol] = -1
	sort.Slice(order, func(i, j int) bool { return rank[order[i]] < rank[order[j]] })
	for _, s := range order[:ZonesHeld(lit)] {
		z.Held[s] = true
	}
	used := map[string]bool{}
	for s := range z.Names {
		for k := 0; ; k++ {
			n := "~" + genName(seed, 9100+int64(s), int64(k)) + " Reach"
			if !used[n] {
				used[n], z.Names[s] = true, n
				break
			}
		}
	}
	return z
}

// At is the sector under the disc point u, v, or -1 outside the disc.
func (z *Zones) At(u, v float64) int {
	r := math.Hypot(u, v)
	if r >= 1 {
		return -1
	}
	if r < zoneCore {
		return 0
	}
	ring := min(int((r-zoneCore)/zoneRingW), zoneRings-1)
	t := math.Mod(math.Atan2(v, u)-z.phase-zoneTwist*r, 2*math.Pi)
	if t < 0 {
		t += 2 * math.Pi
	}
	return 1 + ring*zoneWedges + min(int(t/(2*math.Pi/zoneWedges)), zoneWedges-1)
}

// at is the disc point in sector s at r (0..1 across its ring) and t (0..1
// across its wedge).
func (z *Zones) at(s int, r, t float64) (u, v float64) {
	if s == 0 {
		a := t * 2 * math.Pi
		return zoneCore * r * math.Cos(a), zoneCore * r * math.Sin(a)
	}
	ring, wedge := (s-1)/zoneWedges, (s-1)%zoneWedges
	rr := zoneCore + zoneRingW*(float64(ring)+r)
	a := z.phase + zoneTwist*rr + (float64(wedge)+t)*2*math.Pi/zoneWedges
	return rr * math.Cos(a), rr * math.Sin(a)
}

func (z *Zones) centre(s int) (u, v float64) { return z.at(s, .5, .5) }

// Mask is the territory as a face-on raster of zoneMask squared, row by row
// over -1..1: 0 for unheld or off the disc, 1 held, 2 held at the frontier.
// Built once per plan, so the renderer's tint is rebuilt only when the plan is.
func (z *Zones) Mask() []uint8 {
	z.maskMu.Do(func() {
		sec := make([]int, len(z.mask))
		for i := range sec {
			sec[i] = z.At((float64(i%zoneMask)+.5)/zoneMask*2-1, (float64(i/zoneMask)+.5)/zoneMask*2-1)
		}
		held := func(i int) bool { return sec[i] >= 0 && z.Held[sec[i]] }
		for i, s := range sec {
			if s < 0 || !z.Held[s] {
				continue
			}
			z.mask[i] = 1
			x, y := i%zoneMask, i/zoneMask
			for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if nx, ny := x+d[0], y+d[1]; nx >= 0 && ny >= 0 && nx < zoneMask && ny < zoneMask && sec[ny*zoneMask+nx] >= 0 && !held(ny*zoneMask+nx) {
					z.mask[i] = 2
				}
			}
		}
	})
	return z.mask[:]
}

// objects are the sectors' pick zones, at their centres, and the stars the
// seed puts in each; the people can enter a held sector's stars.
func (z *Zones) objects(seed int64) []Object {
	var out []Object
	for s := range z.Names {
		cu, cv := z.centre(s)
		out = append(out, Object{Kind: KindFeature, Form: FormZone, Key: z.Names[s], Label: z.Names[s], Inspect: z.Names[s],
			Pos: z.shape.unface(cu, cv), Tone: ToneGold, Weight: 255, Pick: true, Held: z.Held[s]})
		for j := 0; j < zoneStars; j++ {
			id := s*zoneStars + j
			u, v := z.at(s, .15+.7*float64(Roll(100, seed, 9200, int64(id)))/100, .15+.7*float64(Roll(100, seed, 9201, int64(id)))/100)
			st := GenStar(seed, id, z.shape.unface(u, v), z.Held[s])
			st.Star.Size = 0 // a point on the disc, not a sun
			if z.Held[s] {
				st.Weight, st.Star.Weight = 255, 255
			}
			out = append(out, st)
		}
	}
	return out
}

// ZoneAt is the sector the galaxy-frame point p (light-years) lies in.
func (z *Zones) ZoneAt(p Vec) int { return z.At(z.shape.face(p)) }

// The made-up stars. A star's key is "~Name id": the id says everything the
// star is (its look, its system, its neighbours), so the key alone rebuilds it.

var (
	syl1 = [...]string{"Vel", "Kar", "Ost", "Mir", "Tal", "Sen", "Dra", "Ith", "Ora", "Zen", "Hal", "Bry", "Cor", "Lun", "Nys", "Ael"}
	syl2 = [...]string{"ta", "ra", "li", "no", "ve", "sa", "mi", "do", "ke", "ul", "an", "or"}
	syl3 = [...]string{"ra", "th", "ris", "on", "ex", "ia", "us", "ei", "ol", "ar"}
)

// genName is a short invented name, the same for the same seed and salts.
func genName(seed int64, salts ...int64) string {
	h := func(n int, i int64) int { return Roll(n, append([]int64{seed, 9300 + i}, salts...)...) }
	return syl1[h(len(syl1), 0)] + syl2[h(len(syl2), 1)] + syl3[h(len(syl3), 2)]
}

// StarKey is the key of generated star id.
func StarKey(seed int64, id int) string {
	return "~" + genName(seed, int64(id)) + " " + strconv.Itoa(id)
}

// StarID is the id in a generated star's key, and whether key is one.
func StarID(key string) (int, bool) {
	if !strings.HasPrefix(key, "~") {
		return 0, false
	}
	i := strings.LastIndexByte(key, ' ')
	n, err := strconv.Atoi(key[i+1:])
	return n, i > 0 && err == nil
}

// starLook is how generated star id shines, and so what its system is like.
func starLook(seed int64, id int) Look { return starLookV(seed, id, 0) }

// starLookV is the look of version n of star id; version 0 is the star itself.
func starLookV(seed int64, id, n int) Look {
	r := (Roll(100, seed, 9400, int64(id)) + 35*n) % 100 // each version a step along the classes
	switch {
	case r < 45:
		return Look{Tone: ToneRed, Size: 0, Weight: 170}
	case r < 70:
		return Look{Tone: ToneAmber, Size: 0, Weight: 200}
	case r < 92:
		return Look{Tone: ToneSun, Size: 1, Weight: 255}
	}
	return Look{Tone: ToneCool, Size: 1, Weight: 255}
}

// GenStar is the map object of generated star id at pos; held says the people
// may go in.
func GenStar(seed int64, id int, pos Vec, held bool) Object {
	l := starLook(seed, id)
	if !held {
		l.Weight = 120
	}
	k := StarKey(seed, id)
	return Object{Kind: KindStar, Key: k, Label: k, Inspect: k, Pos: pos, Tone: l.Tone, Weight: uint8(l.Weight), Star: &l, Pick: true, Held: held, Doubt: doubtOf(seed, id)}
}

// placeStars are the n stars about the star parent (0 for Sol), spread over
// the region's radius, for a Region frame to show as a place.
func placeStars(seed int64, parent, n int) []Object {
	var out []Object
	for j, id := range starIDs(seed, parent/galStride*galStride, parent, n) {
		a := Roll(Turn, seed, 9501, int64(parent), int64(j))
		r := float64(900 + Roll(6600, seed, 9502, int64(parent), int64(j)))
		out = append(out, GenStar(seed, id, polar(r, a), true))
	}
	return out
}

// starIDs are the ids of n stars about parent in the galaxy whose ids start at
// base, no two with the same name.
func starIDs(seed int64, base, parent, n int) []int {
	ids := make([]int, n)
	seen := map[string]bool{}
	for j := range ids {
		for t := int64(0); ; t++ {
			ids[j] = base + genStarBase + Roll(900000, seed, 9500, int64(parent), int64(j), t)
			if nm := genName(seed, int64(ids[j])); !seen[nm] || t > 50 {
				seen[nm] = true
				break
			}
		}
	}
	return ids
}
