package mapgen

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
	sectorMax = 40 // systems in the biggest sector
	rivalMax  = 14 // and in the biggest rival
	// rivalSlots is how many places the other arms hold for rivals.
	rivalSlots = 14
)

// Sector is a hub and the cells round it where its systems light, in the
// order they light: Cells[0] is the hub.
type Sector struct {
	Hub   Pt
	Cells []Pt
}

// armSpread is n places on a path, evenly spaced by how far they are to the
// eye (a row counts double), the first and last half a step in from the ends.
func armSpread(path []Pt, n int) ([]Pt, int) {
	total := 0
	length := []int{0}
	for i := 1; i < len(path); i++ {
		total += isqrt(sq(path[i].X-path[i-1].X) + 4*sq(path[i].Y-path[i-1].Y))
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

// clusterOffsets are the cells round a hub, nearest first with a little hashed
// disorder, within an oval twice as wide as it is tall.
func clusterOffsets(seed int64, salt int64, rows int) []Pt {
	type cand struct {
		p Pt
		d int
	}
	var cs []cand
	for dy := -rows; dy <= rows; dy++ {
		for dx := -2 * rows; dx <= 2*rows; dx++ {
			d := dx*dx + 4*dy*dy
			if d == 0 || d > 4*rows*rows {
				continue
			}
			cs = append(cs, cand{Pt{dx, dy}, d*8 + roll(24, seed, 350, salt, int64(dx), int64(dy))})
		}
	}
	slices.SortStableFunc(cs, func(a, b cand) int { return cmp.Compare(a.d, b.d) })
	out := make([]Pt, len(cs))
	for i, c := range cs {
		out[i] = c.p
	}
	return out
}

// layClusters gives every hub its cells: in turn, each takes the nearest
// candidate not taken, with no two systems side by side in a row. Taking in
// turns keeps a crowded hub from starving the others, and the result depends
// only on the hubs, so it never changes while a game goes on.
func layClusters(seed int64, hubs []Pt, want []int, rows int) []Sector {
	taken := map[Pt]bool{}
	for _, h := range hubs {
		taken[h] = true
	}
	free := func(p Pt) bool {
		return inWorld(p.X, p.Y) && !taken[p] && !taken[Pt{p.X - 1, p.Y}] && !taken[Pt{p.X + 1, p.Y}]
	}
	out := make([]Sector, len(hubs))
	offs := make([][]Pt, len(hubs))
	next := make([]int, len(hubs))
	for i, h := range hubs {
		out[i] = Sector{Hub: h, Cells: []Pt{h}}
		offs[i] = clusterOffsets(seed, int64(i), rows)
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
		if !more || round > 4*sectorMax {
			break
		}
	}
	return out
}

// layTerritory fixes the sectors along the home arm, the places for rivals on
// the other arms and the rift.
func (p *Plan) layTerritory(seed int64) {
	g := p.Galaxy
	s := g.Shape
	hubs, reach := armSpread(g.Home[galArmN*8/100:galArmN*90/100], len(mapmodel.LineageOrder))
	// a short arm gives smaller sectors, and the hubs take turns above and
	// below it
	room := clamp(reach*7/len(hubs), 14, sectorMax)
	for i := range hubs {
		hubs[i].Y = clamp(hubs[i].Y+(i%2*2-1)*2, 3, worldH-4)
	}
	// the sectors and the rival places are laid together, so that none take
	// a cell another holds
	want := make([]int, 0, len(mapmodel.LineageOrder)+rivalSlots)
	for range hubs {
		want = append(want, room)
	}
	others := []int{}
	for k := 0; k < s.Arms+s.Faint; k++ {
		if k != s.Home {
			others = append(others, k)
		}
	}
	per := (rivalSlots + len(others) - 1) / len(others)
	for _, k := range others {
		for j := 0; j < per && len(hubs) < len(mapmodel.LineageOrder)+rivalSlots; j++ {
			rf := 260 + j*(660/per)
			swing := (2*(j%2) - 1) * 70
			c, _ := s.armAt(seed, k, rf, s.Spin*swing)
			c.X, c.Y = clamp(c.X, 6, worldW-7), clamp(c.Y, 3, worldH-4)
			hubs = append(hubs, c)
			want = append(want, rivalMax)
		}
	}
	all := layClusters(seed, hubs, want, 5)
	p.Sectors = all[:len(mapmodel.LineageOrder)]
	p.Rivals = all[len(mapmodel.LineageOrder):]
	p.Rift = riftPath(seed, g)
}

// riftPath is the crack a harbinger opens: from the rim at the shape's angle,
// a jagged line toward the middle, one cell at a time.
func riftPath(seed int64, g *Galaxy) []Pt {
	s := g.Shape
	r := galRX * 94 / 100
	a := s.Rift
	gx, gy := r*cos4096(a)/1024, r*sin4096(a)/1024*s.Flat/100
	cs, sn := cos4096(s.Tilt), sin4096(s.Tilt)
	x := worldW/2 + (gx*cs-gy*sn)/1024
	y := worldH/2 + (gx*sn+gy*cs)/1024*galRYBase/galRX
	cur := Pt{clamp(x, 2, worldW-3), clamp(y, 1, worldH-2)}
	out := []Pt{cur}
	for i := 1; i < 30; i++ {
		dx, dy := cmp.Compare(worldW/2, cur.X), cmp.Compare(worldH/2, cur.Y)
		switch roll(7, seed, 360, int64(i)) {
		case 0, 1: // a step straight in
			if abs(worldW/2-cur.X) >= 2*abs(worldH/2-cur.Y) {
				cur.X += dx
			} else {
				cur.Y += dy
			}
		case 2, 3, 4: // a diagonal one
			cur.X, cur.Y = cur.X+dx, cur.Y+dy
		case 5: // a jog to one side
			cur.X += 2 * (2*roll(2, seed, 361, int64(i)) - 1)
		default:
			cur.Y += dy
			cur.X -= dx
		}
		cur = Pt{clamp(cur.X, 1, worldW-2), clamp(cur.Y, 1, worldH-2)}
		out = append(out, cur)
	}
	return out
}

// systemsFor is how many systems light for a count of buildings: a log scale,
// five copies a few systems, five hundred a big dense cluster.
func systemsFor(count, most int) int {
	if count <= 0 {
		return 0
	}
	l := 0 // four times the log base two, floored
	for v := count * count * count * count; v > 1; v >>= 1 {
		l++
	}
	return clamp(l-5, 1, most)
}

var _ = mapmodel.LineageOrder

// reserved says whether a cell is one a sector or a rival can light, or the
// home star: the places a wonder's star keeps off.
func (p *Plan) reserved(c Pt) bool {
	if abs(c.X-p.Centre.X) <= 1 && c.Y == p.Centre.Y {
		return true
	}
	for _, list := range [][]Sector{p.Sectors, p.Rivals} {
		for _, s := range list {
			if abs(c.X-s.Hub.X) > 2*5+1 || abs(c.Y-s.Hub.Y) > 5+1 {
				continue
			}
			for _, x := range s.Cells {
				if abs(x.X-c.X) <= 1 && x.Y == c.Y {
					return true
				}
			}
		}
	}
	return false
}
