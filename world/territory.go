package world

import (
	"cmp"
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
)

// The people's territory on the star map. Each building line owns a sector
// on the home arm, a hub with a cluster of systems round it; the cluster
// lights as the line grows, a log of its count. The civilizations met stand
// in smaller clusters on the other arms. Where everything can stand is fixed
// by the seed here; which of it is lit is the model's business (stars_paint.go).

const (
	SectorMax = 40 // systems in the biggest sector
	RivalMax  = 20 // and in the biggest rival
	// RivalSlots is how many places the other arms hold for rivals.
	RivalSlots = 14
)

// Sector is a hub and the cells round it where its systems light, in the
// order they light: Cells[0] is the hub.
type Sector struct {
	Hub   Pt
	Cells []Pt
}

// ArmSpread is n places on a path, evenly spaced by how far they are to the
// eye (a row counts double), the first and last half a step in from the ends.
func ArmSpread(path []Pt, n int) ([]Pt, int) {
	total := 0
	length := []int{0}
	for i := 1; i < len(path); i++ {
		total += Isqrt(Sq(path[i].X-path[i-1].X) + 4*Sq(path[i].Y-path[i-1].Y))
		length = append(length, total)
	}
	out := make([]Pt, 0, n)
	j := 0
	for k := 0; k < n; k++ {
		want := total * (2*k + 1) / (2 * n)
		for j < len(path)-1 && length[j] < want {
			j++
		}
		out = append(out, path[j])
	}
	return out, total
}

// ClusterOffsets are the cells round a hub, in the order they light, within an
// oval twice as wide as it is tall. They are not a grid: a few clumps stand
// off the hub, each cell is ranked by how near it is to the hub or a clump
// with a good deal of hashed disorder, and a third of the cells are left
// out, so the gaps come uneven and the cluster reads as a cluster of stars.
func ClusterOffsets(seed int64, salt int64, rows int) []Pt {
	type cand struct {
		p Pt
		d int
	}
	var clumps []Pt
	for k := 0; k < 2+Roll(3, seed, 351, salt); k++ {
		clumps = append(clumps, Pt{
			Roll(2*rows+1, seed, 352, salt, int64(k)) - rows,
			(Roll(rows+1, seed, 353, salt, int64(k)) - rows/2) * 2 / 3,
		})
	}
	var cs []cand
	for dy := -rows; dy <= rows; dy++ {
		for dx := -2 * rows; dx <= 2*rows; dx++ {
			d := dx*dx + 4*dy*dy
			if d == 0 || d > 4*rows*rows {
				continue
			}
			if Roll(100, seed, 354, salt, int64(dx), int64(dy)) < 34 {
				continue
			}
			near := d
			for _, c := range clumps {
				if dc := (Sq(dx-2*c.X) + 4*Sq(dy-c.Y)) * 3 / 2; dc < near {
					near = dc
				}
			}
			cs = append(cs, cand{Pt{dx, dy}, near*8 + d*2 + Roll(40*rows, seed, 350, salt, int64(dx), int64(dy))})
		}
	}
	slices.SortStableFunc(cs, func(a, b cand) int { return cmp.Compare(a.d, b.d) })
	out := make([]Pt, len(cs))
	for i, c := range cs {
		out[i] = c.p
	}
	return out
}

// LayClusters gives every hub its cells: in turn, each takes the nearest
// candidate not taken, with no two systems side by side in a row. Taking in
// turns keeps a crowded hub from starving the others, and the result depends
// only on the hubs, so it never changes while a game goes on.
func LayClusters(seed int64, hubs []Pt, want []int, rows int) []Sector {
	taken := map[Pt]bool{}
	for _, h := range hubs {
		taken[h] = true
	}
	free := func(p Pt) bool {
		return InWorld(p.X, p.Y) && !taken[p] && !taken[Pt{p.X - 1, p.Y}] && !taken[Pt{p.X + 1, p.Y}]
	}
	out := make([]Sector, len(hubs))
	offs := make([][]Pt, len(hubs))
	next := make([]int, len(hubs))
	for i, h := range hubs {
		out[i] = Sector{Hub: h, Cells: []Pt{h}}
		offs[i] = ClusterOffsets(seed, int64(i), rows)
	}
	for round := 1; ; round++ {
		more := false
		for i, h := range hubs {
			if len(out[i].Cells) >= want[i] {
				continue
			}
			more = true
			for next[i] < len(offs[i]) {
				p := Pt{h.X + offs[i][next[i]].X, h.Y + offs[i][next[i]].Y}
				next[i]++
				if free(p) {
					taken[p] = true
					out[i].Cells = append(out[i].Cells, p)
					break
				}
			}
		}
		if !more || round > 4*SectorMax {
			break
		}
	}
	return out
}

// LayTerritory fixes the sectors along the home arm, the places for rivals on
// the other arms and the rift.
// Territory is the laid-out sky of one seed: where the sectors, the rivals
// and the rift stand.
type Territory struct {
	Sectors []Sector // one per building line, in mapmodel.LineageOrder
	Rivals  []Sector
	Rift    []Pt
}

func LayTerritory(seed int64, g *Galaxy) Territory {
	s := g.Shape
	hubs, reach := ArmSpread(g.Home[GalArmN*8/100:GalArmN*90/100], len(mapmodel.LineageOrder))
	// a short arm gives smaller sectors, and the hubs take turns above and
	// below it
	room := clamp(reach*7/len(hubs), 14, SectorMax)
	for i := range hubs {
		hubs[i].Y = clamp(hubs[i].Y+(i%2*2-1)*2, 3, CellsH-4)
	}
	// the sectors and the rival places are laid together, so that none take
	// a cell another holds
	want := make([]int, 0, len(mapmodel.LineageOrder)+RivalSlots)
	for range hubs {
		want = append(want, room)
	}
	others := []int{}
	for k := 0; k < s.Arms+s.Faint; k++ {
		if k != s.Home {
			others = append(others, k)
		}
	}
	per := (RivalSlots + len(others) - 1) / len(others)
	for _, k := range others {
		for j := 0; j < per && len(hubs) < len(mapmodel.LineageOrder)+RivalSlots; j++ {
			rf := 260 + j*(660/per)
			swing := (2*(j%2) - 1) * 70
			c, _ := s.ArmAt(seed, k, rf, s.Spin*swing)
			c.X, c.Y = clamp(c.X, 6, CellsW-7), clamp(c.Y, 3, CellsH-4)
			hubs = append(hubs, c)
			want = append(want, RivalMax)
		}
	}
	all := LayClusters(seed, hubs, want, 5)
	return Territory{Sectors: all[:len(mapmodel.LineageOrder)], Rivals: all[len(mapmodel.LineageOrder):], Rift: RiftPath(seed, g)}
}

// RiftPath is the crack a harbinger opens: from the rim at the shape's angle,
// a jagged line toward the middle, one cell at a time.
func RiftPath(seed int64, g *Galaxy) []Pt {
	s := g.Shape
	r := GalRX * 94 / 100
	a := s.Rift
	gx, gy := r*Cos4096(a)/1024, r*Sin4096(a)/1024*s.Flat/100
	cs, sn := Cos4096(s.Tilt), Sin4096(s.Tilt)
	x := CellsW/2 + (gx*cs-gy*sn)/1024
	y := CellsH/2 + (gx*sn+gy*cs)/1024*GalRYBase/GalRX
	cur := Pt{clamp(x, 2, CellsW-3), clamp(y, 1, CellsH-2)}
	out := []Pt{cur}
	far := Isqrt(Sq(cur.X-CellsW/2) + 4*Sq(cur.Y-CellsH/2))
	for i := 1; i < 30; i++ {
		if Isqrt(Sq(cur.X-CellsW/2)+4*Sq(cur.Y-CellsH/2)) < far*62/100 {
			break // a crack in the rim, not a wound to the core
		}
		dx, dy := cmp.Compare(CellsW/2, cur.X), cmp.Compare(CellsH/2, cur.Y)
		switch Roll(7, seed, 360, int64(i)) {
		case 0, 1: // a step straight in
			if abs(CellsW/2-cur.X) >= 2*abs(CellsH/2-cur.Y) {
				cur.X += dx
			} else {
				cur.Y += dy
			}
		case 2, 3, 4: // a diagonal one
			cur.X, cur.Y = cur.X+dx, cur.Y+dy
		case 5: // a jog to one side
			cur.X += 2 * (2*Roll(2, seed, 361, int64(i)) - 1)
		default:
			cur.Y += dy
			cur.X -= dx
		}
		cur = Pt{clamp(cur.X, 1, CellsW-2), clamp(cur.Y, 1, CellsH-2)}
		out = append(out, cur)
	}
	return out
}

// SystemsFor is how many systems light for a count of buildings: a root scale
// from one system for a handful to a full cluster for a few hundred, steep
// enough that a young civilization is a spatter and an old one a settled arm.
func SystemsFor(count, most int) int {
	if count <= 0 {
		return 0
	}
	return clamp(2*Isqrt(count)-2, 1, most)
}

func clamp(v, lo, hi int) int { return min(max(v, lo), hi) }

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// InWorld says whether a cell is on the galaxy raster.
func InWorld(x, y int) bool { return x >= 0 && y >= 0 && x < CellsW && y < CellsH }
