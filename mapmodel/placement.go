package mapmodel

import (
	"math"
	"sort"
)

// placement.go puts every building somewhere, for both styles, under one
// rule: a thing's position is a function of (seed, its group, its ordinal
// in that group), and ordinals only ever grow. Nothing is placed by
// competing for space with what else happens to be built, so building
// something new never moves anything already drawn.
//
//   - The town (roguelike): a lineage's tiles fill its reserved quarter
//     (Plan.Quarter) in order. The lineage draws TownTiles(copies) tiles,
//     which only grows as copies are added; the tiles are shared out among
//     its types oldest first (ruins, then legacy, then current), so an
//     upgrade or a new tier re-skins tiles in place instead of moving them.
//     Legacy types collapse into the old end of the quarter and draw in the
//     legacy class, which keeps late ages from turning into glyph soup.
//   - The skyline: each age is a district of fixed width; a type's copies
//     take fixed-stride slots (copy c of the type in slot t takes lot
//     c·types+t), which depend on nothing but the catalogue.
//
// Both are pure functions of the seed and the snapshot, so no placement
// history has to be saved.

// TownTiles is how many town tiles a lineage with n copies (ruins
// included) draws: one per copy while small, then sub-linear. Fields draw
// twice as many.
func TownTiles(lin string, n int) int {
	if n <= 0 {
		return 0
	}
	t := n
	if n > 6 {
		t = 6 + int(math.Ceil(float64(5*Log2(float64(n)/6))))
		if t > n {
			t = n
		}
	}
	limit := quarterSize
	if lin == LinFood {
		t *= 2
		limit = foodSize
	}
	if t > limit {
		t = limit
	}
	return t
}

// Town is the roguelike's settlement: the world, its plan and the tiles.
type Town struct {
	World *World
	Plan  *Plan
	Tiles []TownTile
	// WonderTiles are the built (or rising) wonders on their plots.
	Wonders []TownWonder
	at      map[int]int // tile index by y*W+x
}

// TownTile is one building tile in the town.
type TownTile struct {
	Pt
	Lineage string
	Key     string // the building type this tile shows
	Ord     int    // position in the lineage's quarter
	Legacy  bool
	Ruin    bool
	Fresh   bool // added since the last visit
	Anchor  bool // the type's first tile: counts and names go here
}

// TownWonder is a wonder on its plot.
type TownWonder struct {
	Key   string
	At    Pt // top-left of the WonderW x WonderH plot
	Built bool
	Fresh bool
}

// TileAt returns the building tile at p, or nil.
func (t *Town) TileAt(p Pt) *TownTile {
	if i, ok := t.at[p.Y*WorldW+p.X]; ok {
		return &t.Tiles[i]
	}
	return nil
}

// WonderAt returns the wonder whose plot covers p, or nil.
func (t *Town) WonderAt(p Pt) *TownWonder {
	for i := range t.Wonders {
		w := &t.Wonders[i]
		if p.X >= w.At.X && p.X < w.At.X+WonderW && p.Y >= w.At.Y && p.Y < w.At.Y+WonderH {
			return w
		}
	}
	return nil
}

type share struct {
	b      *Building
	n      int // copies this share stands for
	ruin   bool
	legacy bool
	tiles  int
	fresh  int // how many of its tiles are new since the visit
}

func (b *Builder) town(m *Model) Town {
	w, p := b.seedParts(m.Seed)
	t := Town{World: w, Plan: p, at: map[int]int{}}
	for _, l := range m.Lineages {
		quarter := p.Quarter[l.Key]
		shares := lineageShares(l)
		total := 0
		for _, s := range shares {
			total += s.n
		}
		per := 1
		if l.Key == LinFood {
			per = 2 // a field is two tiles
		}
		allocate(shares, l.Tiles, per)
		ord := 0
		for _, s := range shares {
			for j := 0; j < s.tiles; j++ {
				if ord >= len(quarter) {
					break
				}
				pt := quarter[ord]
				tt := TownTile{Pt: pt, Lineage: l.Key, Key: s.b.Key, Ord: ord, Legacy: s.legacy, Ruin: s.ruin,
					Anchor: j == 0 && !s.ruin, Fresh: j >= s.tiles-s.fresh}
				t.at[pt.Y*WorldW+pt.X] = len(t.Tiles)
				t.Tiles = append(t.Tiles, tt)
				ord++
			}
		}
	}
	for _, wd := range m.Wonders {
		pt, ok := p.WonderPlots[wd.Key]
		if !ok {
			continue
		}
		t.Wonders = append(t.Wonders, TownWonder{Key: wd.Key, At: pt, Built: wd.Built, Fresh: wd.Delta})
	}
	return t
}

// lineageShares lists a lineage's types oldest first: ruins, legacy types,
// then current ones.
func lineageShares(l *Lineage) []*share {
	var ruins, legacy, cur []*share
	for _, b := range l.Buildings {
		if b.Ruins > 0 {
			ruins = append(ruins, &share{b: b, n: b.Ruins, ruin: true})
		}
		if b.Count == 0 {
			continue
		}
		s := &share{b: b, n: b.Count, legacy: b.Legacy}
		if b.Delta > 0 {
			s.fresh = b.Delta
		}
		if b.Legacy {
			legacy = append(legacy, s)
		} else {
			cur = append(cur, s)
		}
	}
	out := append(ruins, legacy...)
	return append(out, cur...)
}

// allocate shares n tiles among the shares in proportion to their copies
// (largest remainder), at least one each while there are tiles to go
// round (current types first), never more than per tiles a copy.
func allocate(shares []*share, n, per int) {
	total := 0
	for _, s := range shares {
		total += s.n
	}
	if total == 0 || n == 0 {
		return
	}
	given := 0
	// one each, newest first, so a new type always shows
	for i := len(shares) - 1; i >= 0 && given < n; i-- {
		shares[i].tiles = 1
		given++
	}
	type rem struct {
		i int
		f float64
	}
	var rems []rem
	for i, s := range shares {
		exact := float64(n) * float64(s.n) / float64(total)
		extra := int(exact) - s.tiles
		if extra < 0 {
			extra = 0
		}
		if extra > s.n*per-s.tiles {
			extra = s.n*per - s.tiles
		}
		if given+extra > n {
			extra = n - given
		}
		s.tiles += extra
		given += extra
		rems = append(rems, rem{i, exact - float64(int(exact))})
	}
	sort.SliceStable(rems, func(a, b int) bool { return rems[a].f > rems[b].f })
	for len(rems) > 0 && given < n {
		progressed := false
		for _, r := range rems {
			if given >= n {
				break
			}
			if s := shares[r.i]; s.tiles < s.n*per {
				s.tiles++
				given++
				progressed = true
			}
		}
		if !progressed {
			break
		}
	}
	for _, s := range shares {
		if s.fresh > s.tiles {
			s.fresh = s.tiles
		}
		// a share's new copies show as new tiles only in proportion
		if s.n > 0 && s.fresh > 0 {
			f := int(math.Ceil(float64(s.tiles) * float64(s.fresh) / float64(s.n)))
			if f < 1 {
				f = 1
			}
			if f > s.tiles {
				f = s.tiles
			}
			s.fresh = f
		}
	}
}

// Skyline is the panorama's layout: a district per age, oldest west, and
// the lots on them.
type Skyline struct {
	Districts []District
	Lots      []Lot
	// FrontierX is the first column past the newest district (cranes go
	// here); Width is the whole panorama's width in columns.
	FrontierX, Width int
}

// District is one age's stretch of the skyline.
type District struct {
	Age    int
	X0, W  int
	LandW  int // W without the bay
	Centre int // world column of the district's middle (on land)
	Bay    bool
	BayX   int // the inlet's first column when Bay
	BayW   int
}

// Lot is one silhouette on the skyline.
type Lot struct {
	Key     string
	Lineage string
	Age     int // the building's own period
	Copy    int
	Row     int // depth: 0 front, 1, 2 behind
	X       int // world column the silhouette is centred on
	Seed    uint64
	Wonder  bool
	Quay    bool // a harbour building on the bay's shore
	Legacy  bool
	New     bool
}

// Skyline sizing, in panorama columns.
const (
	SkyWestPad   = 34
	SkyFrontierW = 30
	SkyEastPad   = 46
	skyLotStep   = 3
	skyBayW      = 20
	// SkyMaxCopies caps the silhouettes one type draws.
	SkyMaxCopies = 6
)

// SkyCopies is how many silhouettes n copies draw: one each while small,
// then about √n, capped.
func SkyCopies(n int) int {
	if n <= 0 {
		return 0
	}
	if n <= 3 {
		return n
	}
	c := int(math.Ceil(float64(math.Sqrt(float64(n)) * 1.15)))
	if c > SkyMaxCopies {
		c = SkyMaxCopies
	}
	return c
}

func districtWidth(n int) int {
	w := 12 + n*10/3
	if w < 30 {
		w = 30
	}
	return w
}

func hasBay(ds []*Def) bool {
	for _, d := range ds {
		if d.Lineage == LinHarbor || d.Key == "grand_lighthouse" {
			return true
		}
	}
	return false
}

func skylineFor(m *Model) Skyline {
	cat := m.Catalog
	sk := Skyline{}
	x := SkyWestPad
	for a := 0; a <= m.AgeIdx && a < len(cat.ByAge); a++ {
		ds := cat.ByAge[a]
		d := District{Age: a, X0: x, W: districtWidth(len(ds))}
		d.LandW = d.W
		if hasBay(ds) {
			d.Bay, d.BayW, d.BayX = true, skyBayW, x+d.W
			d.W += skyBayW + 2
		}
		d.Centre = d.X0 + d.LandW/2
		sk.Districts = append(sk.Districts, d)
		x += d.W
	}
	sk.FrontierX = x
	sk.Width = x + SkyFrontierW + SkyEastPad

	for a, d := range sk.Districts {
		ds := cat.ByAge[a]
		nt := len(ds)
		capRow := d.LandW / skyLotStep
		if capRow < 4 {
			capRow = 4
		}
		for t, def := range ds {
			b := m.byKey[def.Key]
			if b == nil || b.Count <= 0 {
				continue
			}
			n := SkyCopies(b.Count)
			prevN := 0
			if b.Delta > 0 {
				prevN = SkyCopies(b.Count - b.Delta)
			}
			for c := 0; c < n; c++ {
				lot := Lot{Key: def.Key, Lineage: def.Lineage, Age: a, Copy: c, Legacy: b.Legacy,
					Seed: Hash(m.Seed, int64(a), int64(t), int64(c)), New: b.Delta > 0 && c >= prevN}
				switch {
				case def.Wonder:
					lot.Wonder, lot.X, lot.Row = true, d.Centre, 0
					if d.Bay && def.Key == "grand_lighthouse" {
						lot.X = d.BayX + d.BayW - 4
					}
				case def.Lineage == LinHarbor && d.Bay:
					lot.Quay, lot.X, lot.Row = true, d.BayX+2+c*skyLotStep, c%2
				default:
					L := c*nt + t
					s := L % capRow
					lot.Row = L / capRow
					if lot.Row > 2 {
						lot.Row = 2
					}
					off := (s + 1) / 2 * skyLotStep
					if s%2 == 1 {
						off = -off
					}
					off += int(Hash(int64(lot.Seed), 3)%3) - 1
					lot.X = d.Centre + off
					if lot.X < d.X0+2 {
						lot.X = d.X0 + 2
					}
					if lot.X > d.X0+d.LandW-3 {
						lot.X = d.X0 + d.LandW - 3
					}
				}
				sk.Lots = append(sk.Lots, lot)
				if def.Wonder {
					break // one landmark, whatever the count
				}
			}
		}
	}
	return sk
}

// DistrictAt returns the index of the district under panorama column x,
// or -1.
func (s *Skyline) DistrictAt(x int) int {
	for i, d := range s.Districts {
		if x >= d.X0 && x < d.X0+d.W {
			return i
		}
	}
	return -1
}
