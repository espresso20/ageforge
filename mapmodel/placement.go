package mapmodel

import (
	"cmp"
	"math"
	"slices"
)

// placement.go puts every building somewhere, under one rule: a thing's
// position is a function of (seed, its group, its ordinal in that group), and
// ordinals only ever grow. Nothing is placed by competing for space with what
// else happens to be built, so building something new never moves anything
// already drawn.
//
// A lineage's tiles fill its reserved quarter (Plan.Quarter) in order. The
// lineage draws TownTiles(copies) tiles, which only grows as copies are
// added; the tiles are shared out among its types oldest first (ruins, then
// legacy, then current), so an upgrade or a new tier re-skins tiles in place
// instead of moving them. Legacy types collapse into the old end of the
// quarter and draw in the legacy class, which keeps late ages from turning
// into glyph soup.
//
// The placement is a pure function of the seed and the snapshot, so no
// placement history has to be saved.

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

// Town is the settlement: the world, its plan and the tiles.
type Town struct {
	World *World
	Plan  *Plan
	Tiles []TownTile
	// WonderTiles are the built (or rising) wonders on their plots.
	Wonders []TownWonder
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
	t := Town{World: w, Plan: p}
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
	slices.SortStableFunc(rems, func(a, b rem) int { return cmp.Compare(b.f, a.f) })
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
