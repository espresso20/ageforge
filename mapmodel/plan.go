package mapmodel

import (
	"math"
	"sort"
)

// plan.go is the static half of the town layout: a pure function of the
// seed. It lays the street grid, the square and the buildable lots, and then
// reserves, for every lineage, the ordered list of lots its tiles will take
// (its "quarter"). Reserving per lineage is what makes placement strictly
// stable without any history: a lineage's n-th tile always lands on the n-th
// lot of its own list, whatever else the player has built.

// Plan is the town plan for one seed.
type Plan struct {
	Street []bool // grid streets (opened where something is built)
	Plaza  []bool // the central square
	// Quarter is each lineage's reserved lots, nearest its anchor first.
	Quarter map[string][]Pt
	// WonderPlots is the top-left of each wonder's 3x2 plot, by wonder key.
	WonderPlots map[string]Pt
	Anchor      map[string]Pt
	Rmax        float64

	NearWater, NearForest, NearHills []bool
}

// Plan sizing. A lineage reserves room for its tile budget (TownTiles).
const (
	planRmax    = 34
	quarterSize = 56
	foodSize    = 112
)

// WonderW and WonderH are a wonder plot's size in tiles.
const (
	WonderW = 3
	WonderH = 2
)

// NewPlan lays out the town for a world and catalogue.
func NewPlan(w *World, cat *Catalog) *Plan {
	n := w.W * w.H
	p := &Plan{Street: make([]bool, n), Plaza: make([]bool, n), Quarter: map[string][]Pt{},
		WonderPlots: map[string]Pt{}, Anchor: map[string]Pt{}, Rmax: planRmax,
		NearWater: make([]bool, n), NearForest: make([]bool, n), NearHills: make([]bool, n)}
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
					t := w.At(x+dx, y+dy)
					p.NearWater[i] = p.NearWater[i] || t.Water() && w.In(x+dx, y+dy)
					p.NearForest[i] = p.NearForest[i] || t == TForest
					p.NearHills[i] = p.NearHills[i] || t == THills || t == TMountain
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
			// A staggered grid of 4x2 blocks; each band shifts sideways a
			// little so the town reads as grown, not ruled.
			ry := y - w.CY + 1
			band := int(math.Floor(float64(ry) / 3))
			warp := int(Hash(w.Seed, 31, int64(band))%3) - 1
			if mod(ry, 3) == 0 || mod(x-w.CX+warp+2, 5) == 0 {
				p.Street[i] = true
				continue
			}
			if t := w.T[i]; t.Land() && t != TMountain {
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
	used := make([]bool, n)
	jit := make([]float64, len(lots))
	for li, l := range lots {
		jit[li] = float64(0.4 * HashF(w.Seed, 55, int64(l.p.Y*w.W+l.p.X)))
	}

	// Wonder plots first, on a ring just off the square, in age order.
	for _, d := range cat.Wonders {
		if pt, ok := p.wonderPlot(w, used); ok {
			p.WonderPlots[d.Key] = pt
			for dy := 0; dy < WonderH; dy++ {
				for dx := 0; dx < WonderW; dx++ {
					used[(pt.Y+dy)*w.W+pt.X+dx] = true
				}
			}
		}
	}

	// Anchors: the quarters sit on a ring round the square, rotated by seed.
	rot := HashF(w.Seed, 61)
	ring := LineageOrder[1:] // housing sits on the square itself
	for k, lin := range ring {
		a := rot + float64(k)/float64(len(ring))
		r := 6.0
		if lin == LinFood || lin == LinWood || lin == LinMines || lin == LinHarbor {
			r = 11 // fields, woods and mines sit further out
		}
		p.Anchor[lin] = Pt{w.CX + int(math.Round(float64(Cos(a)*r)*2)), w.CY + int(math.Round(float64(Sin(a)*r)))}
	}
	p.Anchor[LinHousing] = Pt{w.CX, w.CY}

	// Reserve each lineage's quarter: lineages take turns, food three lots
	// a turn, housing two, the rest one, each taking the free lot that best
	// suits it (near its anchor and the square, on the ground its trade
	// needs). The order of turns is fixed, so the result is too.
	type acc struct{ sx, sy, n float64 }
	cent := map[string]*acc{}
	want := map[string]int{}
	for _, lin := range LineageOrder {
		want[lin] = quarterSize
		cent[lin] = &acc{}
	}
	want[LinFood] = foodSize
	per := map[string]int{LinFood: 3, LinHousing: 2}
	for {
		progressed := false
		for _, lin := range LineageOrder {
			take := per[lin]
			if take == 0 {
				take = 1
			}
			for t := 0; t < take && len(p.Quarter[lin]) < want[lin]; t++ {
				a := cent[lin]
				target := p.Anchor[lin]
				if a.n > 0 {
					// Pull toward the quarter's own centre so it grows as a
					// clump, but keep a little of the anchor so it stays put.
					target = Pt{int((a.sx + float64(target.X)) / (a.n + 1)), int((a.sy + float64(target.Y)) / (a.n + 1))}
				}
				best, bi := math.Inf(1), -1
				for li, l := range lots {
					i := l.p.Y*w.W + l.p.X
					if used[i] {
						continue
					}
					dc := l.d
					if float64(0.7*dc) > best {
						break
					}
					c := rdist(l.p.X, l.p.Y, target.X, target.Y) + float64(0.7*dc) + jit[li]
					c += lotPenalty(lin, w, p, i, dc)
					if c < best {
						best, bi = c, li
					}
				}
				if bi < 0 {
					break
				}
				l := lots[bi].p
				used[l.Y*w.W+l.X] = true
				p.Quarter[lin] = append(p.Quarter[lin], l)
				a.sx, a.sy, a.n = a.sx+float64(l.X), a.sy+float64(l.Y), a.n+1
				progressed = true
			}
		}
		if !progressed {
			break
		}
	}
	return p
}

// lotPenalty is how badly a lot suits a lineage's trade.
func lotPenalty(lin string, w *World, p *Plan, i int, dc float64) float64 {
	switch lin {
	case LinFood:
		if dc < 5 {
			return 8
		}
	case LinHarbor:
		if !p.NearWater[i] {
			return 40
		}
	case LinWood:
		if !p.NearForest[i] {
			return 25
		}
	case LinMines:
		if !p.NearHills[i] {
			return 25
		}
	default:
		if t := w.T[i]; t == TForest || t == THills {
			return 1.5
		}
	}
	return 0
}

// wonderPlot finds the best free 3x2 plot just off the square.
func (p *Plan) wonderPlot(w *World, used []bool) (Pt, bool) {
	best, bp := math.Inf(1), Pt{-1, -1}
	R := int(p.Rmax)
	for y := w.CY - R/2; y <= w.CY+R/2; y++ {
		for x := w.CX - R; x <= w.CX+R; x++ {
			d := rdist(x+1, y, w.CX, w.CY)
			if d > best || d < 2.5 {
				continue
			}
			ok := true
			for dy := 0; dy < WonderH && ok; dy++ {
				for dx := 0; dx < WonderW && ok; dx++ {
					xx, yy := x+dx, y+dy
					// keep a one-tile margin from other wonders
					for my := -1; my <= 1 && ok; my++ {
						for mx := -1; mx <= 1 && ok; mx++ {
							if w.In(xx+mx, yy+my) && used[(yy+my)*w.W+xx+mx] {
								ok = false
							}
						}
					}
					if !ok {
						break
					}
					i := yy*w.W + xx
					if !w.In(xx, yy) || p.Plaza[i] || !w.T[i].Land() || w.T[i] == TMountain {
						ok = false
					}
				}
			}
			if ok {
				c := d + float64(0.5*HashF(w.Seed, 56, int64(x), int64(y)))
				if c < best {
					best, bp = c, Pt{x, y}
				}
			}
		}
	}
	return bp, bp.X >= 0
}

func mod(a, b int) int { return ((a % b) + b) % b }
