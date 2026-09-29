package main

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// layout.go places the city. The panorama reads like a timeline: one
// district per age, oldest in the west, the frontier (and your build
// queue's scaffolding) in the east. Every lot is a pure function of
// (save seed, building type, copy index): the type's slot in its age's
// district is fixed by the catalogue, so building something new adds a
// silhouette and never moves an old one.

type Lot struct {
	Key      string
	Name     string
	Def      config.BuildingDef
	Count    int // how many the player owns of this type
	Copy     int // which silhouette of the type this is
	Age      int // the building's own period
	District int
	X        int // world column of the sprite's left edge
	Row      int // depth: 0 front, 1, 2 behind
	Spr      *Sprite
	Fam, Var int
	Staff    float64 // 0..1 how staffed the type is (drives light and smoke)
	Producer bool
	Wonder   bool
	New      bool // appeared since the last check-in
	Seed     int
}

type District struct {
	Age      int
	X0, W    int
	Bay      bool
	BayX, BW int // water inlet (world columns) when Bay
}

type World struct {
	St        game.GameState
	Prev      *game.GameState
	W         int
	AgeIdx    int
	Districts []District
	Lots      []*Lot
	FrontierX int
	Queue     []string // queued build names, drawn as scaffolding
	Scale     float64
	Seed      int
	NewCount  int
}

const (
	westPad    = 34
	frontierW  = 30
	eastPad    = 46
	lotSpacing = 3
	maxCopies  = 6
)

func ageIndex() map[string]int {
	m := map[string]int{}
	for i, a := range config.AgeOrder() {
		m[a] = i
	}
	return m
}

// copiesFor is sub-linear: 1:1 while small, then a silhouette per few.
func copiesFor(n int) int {
	if n <= 0 {
		return 0
	}
	c := int(math.Ceil(math.Sqrt(float64(n)) * 1.15))
	if n <= 3 {
		c = n
	}
	return min(c, maxCopies)
}

// ageTypes lists the catalogue's buildings per age in their fixed district
// order: tallest nominal form first, so each district peaks in its middle.
func ageTypes() map[int][]config.BuildingDef {
	idx := ageIndex()
	out := map[int][]config.BuildingDef{}
	for _, d := range config.BuildingByKey() {
		a := idx[d.RequiredAge]
		out[a] = append(out[a], d)
	}
	for a, ds := range out {
		sort.Slice(ds, func(i, j int) bool {
			hi, hj := formFor(ds[i], a).height, formFor(ds[j], a).height
			if hi != hj {
				return hi > hj
			}
			return ds[i].Key < ds[j].Key
		})
		out[a] = ds
	}
	return out
}

func hasBay(ds []config.BuildingDef) bool {
	for _, d := range ds {
		if d.LineageKey == "harbor" || d.Key == "port" || d.Key == "dockyard" || d.Key == "grand_lighthouse" {
			return true
		}
	}
	return false
}

func districtWidth(n int) int { return max(30, 12+n*10/3) }

// buildWorld lays the state out for a scene of height sceneH.
func buildWorld(st game.GameState, prev *game.GameState, sceneH int) *World {
	idx := ageIndex()
	w := &World{St: st, Prev: prev, AgeIdx: idx[st.Age], Seed: int(st.Seed)}
	w.Scale = math.Max(0.42, math.Min(1.25, float64(sceneH-4)/36))
	types := ageTypes()

	x := westPad
	for a := 0; a <= w.AgeIdx; a++ {
		ds := types[a]
		d := District{Age: a, X0: x, W: districtWidth(len(ds))}
		if hasBay(ds) {
			d.Bay = true
			d.BW = 20
			d.BayX = x + d.W
			d.W += d.BW + 2
		}
		w.Districts = append(w.Districts, d)
		x += d.W
	}
	w.FrontierX = x
	w.W = x + frontierW + eastPad

	for _, q := range st.BuildQueue {
		w.Queue = append(w.Queue, q.Name)
	}

	for a := 0; a <= w.AgeIdx; a++ {
		d := w.Districts[a]
		ds := types[a]
		landW := d.W
		if d.Bay {
			landW -= d.BW + 2
		}
		capRow := max(4, landW/lotSpacing)
		cx := d.X0 + landW/2
		for t, def := range ds {
			bs, ok := st.Buildings[def.Key]
			if !ok || bs.Count <= 0 {
				continue
			}
			prevN := 0
			if prev != nil {
				prevN = prev.Buildings[def.Key].Count
			}
			n := copiesFor(bs.Count)
			pn := copiesFor(prevN)
			fm := formFor(def, a)
			staff := staffing(st, def, bs)
			for c := 0; c < n; c++ {
				lot := &Lot{Key: def.Key, Name: def.Name, Def: def, Count: bs.Count, Copy: c, Age: a,
					District: a, Staff: staff, Producer: fm.smoke, Seed: int(hash(w.Seed, a, t, c))}
				lot.Fam = familyOf(a)
				lot.Var = int(hash(lot.Seed, 7) % variants)
				lot.New = prev != nil && (c >= pn || (prevN == 0 && bs.Count > 0))
				if lot.New {
					w.NewCount++
				}
				if def.Category == "wonder" {
					lot.Wonder = true
					if f, ok := wonderFam[def.Key]; ok {
						lot.Fam = f // a landmark can wear a later period's finish
					}
					lot.Spr = wonderSprite(def.Key, w.Scale)
					lot.Row = 0
					lot.X = cx - lot.Spr.W/2
					if d.Bay && def.Key == "grand_lighthouse" {
						lot.X = d.BayX + d.BW - lot.Spr.W - 1
					}
					w.Lots = append(w.Lots, lot)
					break // one wonder, whatever the count
				}
				if def.LineageKey == "harbor" && d.Bay {
					// quays line the bay's shore
					lot.Spr = fm.fn(newRnd(lot.Seed), heightFor(fm, w.Scale, lot, sceneH))
					lot.X = d.BayX - lot.Spr.W + 2 + c*3
					lot.Row = c % 2
					w.Lots = append(w.Lots, lot)
					continue
				}
				L := c*len(ds) + t
				s := L % capRow
				lot.Row = min(2, L/capRow)
				off := (s + 1) / 2 * lotSpacing
				if s%2 == 1 {
					off = -off
				}
				off += int(hash(lot.Seed, 3)%3) - 1
				lot.Spr = fm.fn(newRnd(lot.Seed), heightFor(fm, w.Scale, lot, sceneH))
				if hash(lot.Seed, 5)%2 == 0 {
					lot.Spr = lot.Spr.mirror()
				}
				lot.X = cx + off - lot.Spr.W/2
				// keep inside the district
				lot.X = max(d.X0+1, min(d.X0+landW-lot.Spr.W-1, lot.X))
				w.Lots = append(w.Lots, lot)
			}
		}
	}
	// draw order: far rows first, then taller first, wonders last in front
	sort.SliceStable(w.Lots, func(i, j int) bool {
		a, b := w.Lots[i], w.Lots[j]
		if a.Wonder != b.Wonder {
			return b.Wonder
		}
		if a.Row != b.Row {
			return a.Row > b.Row
		}
		if a.Spr.H != b.Spr.H {
			return a.Spr.H > b.Spr.H
		}
		return a.Seed < b.Seed
	})
	return w
}

var wonderFam = map[string]int{"crystal_palace": 7, "great_monolith": 8, "hoover_dam": 7}

func heightFor(fm form, scale float64, lot *Lot, sceneH int) int {
	h := fm.height * scale
	h *= 0.82 + 0.36*hashf(lot.Seed, 11)
	h *= 1 + 0.05*math.Log2(1+float64(lot.Count))
	if lot.Row > 0 {
		h *= 1.12
	}
	return max(2, min(int(h+0.5), sceneH-7))
}

// staffing is how busy a building type is: assigned workers over slots for
// producers, the population's fill for housing, a steady glow otherwise.
func staffing(st game.GameState, def config.BuildingDef, bs game.BuildingState) float64 {
	switch {
	case def.Category == "housing":
		if st.Workers.MaxPop > 0 {
			return math.Min(1, 0.25+float64(st.Workers.TotalPop)/float64(st.Workers.MaxPop))
		}
		return 0.5
	case bs.WorkerCapacity > 0:
		return math.Min(1, float64(bs.WorkersAssigned)/float64(bs.Count*bs.WorkerCapacity))
	case def.Category == "wonder":
		return 1
	}
	return 0.45
}

// DistrictAt returns the district index under world column x (-1 none).
func (w *World) DistrictAt(x int) int {
	for i, d := range w.Districts {
		if x >= d.X0 && x < d.X0+d.W {
			return i
		}
	}
	return -1
}

// LotAt returns the front-most lot covering world column x at scene row y
// (y < 0 means any row: the tallest in that column).
func (w *World) LotAt(x, y, groundY int) *Lot {
	var best *Lot
	for i := len(w.Lots) - 1; i >= 0; i-- {
		l := w.Lots[i]
		if x < l.X || x >= l.X+l.Spr.W {
			continue
		}
		if y >= 0 {
			sy := y - (groundY - l.Spr.H)
			if sy < 0 || sy >= l.Spr.H || l.Spr.at(x-l.X, sy).Ch == 0 {
				continue
			}
			return l
		}
		if best == nil {
			best = l
		}
	}
	return best
}

// profile is the skyline height per world column (for the minimap and the
// compact view).
func (w *World) profile() []int {
	p := make([]int, w.W)
	for _, l := range w.Lots {
		for x := 0; x < l.Spr.W; x++ {
			wx := l.X + x
			if wx < 0 || wx >= w.W {
				continue
			}
			for y := 0; y < l.Spr.H; y++ {
				if l.Spr.at(x, y).Ch != 0 {
					if h := l.Spr.H - y; h > p[wx] {
						p[wx] = h
					}
					break
				}
			}
		}
	}
	return p
}
