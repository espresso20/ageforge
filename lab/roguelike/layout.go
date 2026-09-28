package main

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
)

// Plan is the static part of the layout, a pure function of the world
// seed: the street grid, the square, the buildable lots and where each
// lineage's quarter starts.
type Plan struct {
	Street []bool
	Plaza  []bool
	Lots   []Pt // buildable tiles, nearest the centre first
	Anchor map[string]Pt
	Rmax   float64

	nearWater, nearForest, nearHills []bool
}

// quarterOrder is the ring of quarters around the square; the seed rotates
// it, so every run's town is laid out differently.
var quarterOrder = []string{
	"knowledge", "faith", "culture_arts", "monument", "trade", "storage", "food", "organic_extraction",
	"metallurgy", "engineering", "energy", "military", "geological_extraction", "hacker", "astronaut", "harbor",
}

func mod(a, b int) int { return ((a % b) + b) % b }

func NewPlan(w *World) *Plan {
	n := w.W * w.H
	p := &Plan{Street: make([]bool, n), Plaza: make([]bool, n), Anchor: map[string]Pt{}, Rmax: 30,
		nearWater: make([]bool, n), nearForest: make([]bool, n), nearHills: make([]bool, n)}
	type lot struct {
		p Pt
		d float64
	}
	var lots []lot
	for y := 0; y < w.H; y++ {
		for x := 0; x < w.W; x++ {
			i := y*w.W + x
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					t := w.at(x+dx, y+dy)
					p.nearWater[i] = p.nearWater[i] || t.water() && w.in(x+dx, y+dy)
					p.nearForest[i] = p.nearForest[i] || t == TForest
					p.nearHills[i] = p.nearHills[i] || t == THills || t == TMountain
				}
			}
			d := rdist(x, y, w.CX, w.CY)
			if d > p.Rmax {
				continue
			}
			if d <= 1.6 {
				p.Plaza[i] = true
				continue
			}
			// Streets: a staggered grid of 4x2 blocks. Each band of blocks
			// shifts sideways a little, so the town reads as grown, not ruled.
			ry := y - w.CY + 1
			band := int(math.Floor(float64(ry) / 3))
			warp := int(hash(w.Seed, 31, int64(band))%3) - 1
			if mod(ry, 3) == 0 || mod(x-w.CX+warp+2, 5) == 0 {
				p.Street[i] = true
				continue
			}
			if t := w.T[i]; t.land() && t != TMountain {
				lots = append(lots, lot{Pt{x, y}, d})
			}
		}
	}
	sort.Slice(lots, func(a, b int) bool {
		if lots[a].d != lots[b].d {
			return lots[a].d < lots[b].d
		}
		return lots[a].p.Y*w.W+lots[a].p.X < lots[b].p.Y*w.W+lots[b].p.X
	})
	for _, l := range lots {
		p.Lots = append(p.Lots, l.p)
	}
	rot := hf(w.Seed, 61) * 2 * math.Pi
	for k, lin := range quarterOrder {
		a := rot + 2*math.Pi*float64(k)/float64(len(quarterOrder))
		r := 5.0
		if lin == "food" {
			r = 9
		}
		p.Anchor[lin] = Pt{w.CX + int(math.Round(math.Cos(a)*r*2)), w.CY + int(math.Round(math.Sin(a)*r))}
	}
	p.Anchor["housing"] = Pt{w.CX, w.CY}
	return p
}

// item is one tile of one building type in the replayed build history.
type item struct {
	tier, j int
	key     string
	bi      int
	ruin    bool
}

// Build lays a state out. It replays a plausible build history (older
// tiers first, then tile by tile) and gives each tile the best free lot:
// near its own quarter, near the square, on the ground its trade needs.
// A lot, once taken, is never re-assigned, so a new building only ever
// adds tiles at the edge of its quarter; nothing already drawn moves.
//
// note: the prototype replays history from the snapshot; production would
// persist the claims (an append-only list in the save) so even a same-tier
// out-of-order build can never nudge a neighbour.
func Build(w *World, p *Plan, v MapView, forceCatastrophe bool) *Scene {
	s := &Scene{W: w, P: p, V: v, D: dialFor(v.Epoch), Cells: make([]Cell, w.W*w.H), WallC: Pt{w.CX, w.CY}}
	if forceCatastrophe && s.V.PendingCatastrophe == "" {
		s.V.PendingCatastrophe = v.Epoch
		s.V.CatastropheName, _ = config.CatastropheInfo(v.Epoch)
	}
	for i, pl := range p.Plaza {
		if pl {
			s.Cells[i] = Cell{K: KPlaza}
		}
	}
	s.Cells[s.idx(w.CX, w.CY)] = Cell{K: KCentre, G: s.D.Centre, C: CWealth}

	var items []item
	for i, b := range v.Buildings {
		g, c := lineageGlyph(b.Lineage, s.D)
		bl := Bld{V: b, G: g, C: c}
		if v.Prev != nil {
			bl.Delta = b.Count - v.Prev[b.Key]
		}
		bl.Wonder = b.Category == "wonder" || b.Lineage == "wonder"
		s.Blds = append(s.Blds, bl)
		if bl.Wonder {
			if b.Count > 0 {
				items = append(items, item{tier: b.Tier, j: -1, key: b.Key, bi: i})
			}
			continue
		}
		for j := 0; j < tilesForLineage(b.Lineage, b.Count); j++ {
			items = append(items, item{tier: b.Tier, j: j, key: b.Key, bi: i})
		}
		for j := 0; j < tilesFor(b.Ruins); j++ {
			items = append(items, item{tier: b.Tier, j: 1000 + j, key: b.Key, bi: i, ruin: true})
		}
	}
	sort.Slice(items, func(a, b int) bool {
		ia, ib := items[a], items[b]
		if ia.tier != ib.tier {
			return ia.tier < ib.tier
		}
		if ia.j != ib.j {
			return ia.j < ib.j
		}
		return ia.key < ib.key
	})
	used := make([]bool, w.W*w.H)
	type acc struct{ sx, sy, n float64 }
	cent := map[string]*acc{}
	for _, it := range items {
		b := &s.Blds[it.bi]
		if b.Wonder {
			s.placeWonder(b, it.bi, used)
			continue
		}
		lin := b.V.Lineage
		target := p.Anchor[lin]
		if _, ok := p.Anchor[lin]; !ok {
			target = p.Anchor["housing"]
		}
		if a := cent[lin]; a != nil && a.n > 0 {
			target = Pt{int(a.sx / a.n), int(a.sy / a.n)}
		}
		best, bp := math.Inf(1), Pt{-1, -1}
		for _, l := range p.Lots {
			i := l.Y*w.W + l.X
			if used[i] {
				continue
			}
			dc := rdist(l.X, l.Y, w.CX, w.CY)
			if 0.7*dc > best {
				break
			}
			c := rdist(l.X, l.Y, target.X, target.Y) + 0.7*dc + 0.4*hf(w.Seed, 55, int64(i))
			switch lin {
			case "food":
				if dc < 5 {
					c += 8
				}
			case "harbor":
				if !p.nearWater[i] {
					c += 40
				}
			case "organic_extraction":
				if !p.nearForest[i] {
					c += 25
				}
			case "geological_extraction":
				if !p.nearHills[i] {
					c += 25
				}
			default:
				if t := w.T[i]; t == TForest || t == THills {
					c += 1.5
				}
			}
			if c < best {
				best, bp = c, l
			}
		}
		if bp.X < 0 {
			continue // the town is full; the count still shows in the inspector
		}
		i := bp.Y*w.W + bp.X
		used[i] = true
		if cent[lin] == nil {
			cent[lin] = &acc{}
		}
		a := cent[lin]
		a.sx, a.sy, a.n = a.sx+float64(bp.X), a.sy+float64(bp.Y), a.n+1
		b.Tiles = append(b.Tiles, bp)
		if it.ruin {
			s.Cells[i] = Cell{K: KRuin, G: '%', C: CRock, B: int16(it.bi + 1)}
			continue
		}
		n := tilesForLineage(lin, b.V.Count)
		fresh := b.Delta > 0 && it.j >= tilesForLineage(lin, b.V.Count-b.Delta)
		if b.Delta > 0 && it.j == n-1 {
			fresh = true
		}
		under := b.V.WorkCap > 0 && float64(b.V.Workers) < 0.5*float64(b.V.WorkCap)
		s.Cells[i] = Cell{K: KBuilding, G: b.G, C: b.C, B: int16(it.bi + 1), Anchor: it.j == 0, Fresh: fresh, Under: under}
	}
	s.streets()
	s.walls()
	s.trails()
	s.visibility()
	s.life()
	s.hazards()
	s.changes()
	return s
}

// placeWonder claims the best free 3x2 footprint near the square; the
// wonder keeps the look of the era it was raised in.
func (s *Scene) placeWonder(b *Bld, bi int, used []bool) {
	w := s.W
	best, bp := math.Inf(1), Pt{-1, -1}
	R := int(s.P.Rmax)
	for y := w.CY - R/2; y <= w.CY+R/2; y++ {
		for x := w.CX - R; x <= w.CX+R; x++ {
			d := rdist(x+1, y, w.CX, w.CY)
			if d+0 > best || d < 2 {
				continue
			}
			ok := true
			for dy := 0; dy < 2 && ok; dy++ {
				for dx := 0; dx < 3 && ok; dx++ {
					xx, yy := x+dx, y+dy
					i := yy*w.W + xx
					if !w.in(xx, yy) || used[i] || s.P.Plaza[i] || !w.T[i].land() || w.T[i] == TMountain {
						ok = false
					}
				}
			}
			if ok {
				c := d + 0.5*hf(w.Seed, 56, int64(x), int64(y))
				if c < best {
					best, bp = c, Pt{x, y}
				}
			}
		}
	}
	if bp.X < 0 {
		return
	}
	wd := dialFor(epochOfTier(b.V.Tier, 0)).Wonder
	for dy := 0; dy < 2; dy++ {
		for dx, r := range []rune(wd[dy]) {
			pt := Pt{bp.X + dx, bp.Y + dy}
			i := pt.Y*w.W + pt.X
			used[i] = true
			b.Tiles = append(b.Tiles, pt)
			if r == ' ' {
				s.Cells[i] = Cell{K: KPlaza}
				continue
			}
			s.Cells[i] = Cell{K: KWonder, G: r, C: CWealth, B: int16(bi + 1), Anchor: dx == 1 && dy == 1, Fresh: b.Delta > 0}
		}
	}
}
