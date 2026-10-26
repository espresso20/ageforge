package mapgen

import "cmp"

var (
	east  = [2]int{1, 0}
	west  = [2]int{-1, 0}
	south = [2]int{0, 1}
	north = [2]int{0, -1}
)

// layArms puts down the plaza and the four trunk arms that leave it.
func (b *planner) layArms() {
	p := b.plan
	p.Roads = append(p.Roads, Road{Cells: []Pt{p.Centre}, Parent: -1})
	b.claimRoad(p.Centre, 0)
	for i, arm := range [][2]int{east, west, south, north} {
		reach := 56 + roll(30, b.seed, 20, int64(i))
		if arm[0] == 0 {
			reach = 30 + roll(14, b.seed, 20, int64(i))
		}
		b.arms[i] = b.grow(0, 0, arm, reach, 0)
	}
}

// layStreets grows streets off the trunk arms and lanes off the streets.
func (b *planner) layStreets() {
	for _, arm := range b.arms {
		if arm >= 0 {
			b.branch(arm, 1)
		}
	}
	streets := len(b.plan.Roads)
	for road := 1; road < streets; road++ {
		if b.plan.Roads[road].Tier == 1 {
			b.branch(road, 2)
		}
	}
}

func (b *planner) claimRoad(c Pt, road int) {
	b.claim[at(c.X, c.Y)], b.roadOf[at(c.X, c.Y)] = roadCell, road
}

// branch spawns the children of a road: streets off a trunk, lanes off a
// street, on both sides at irregular gaps. A street across a trunk runs
// long; a lane stops where it meets the next street.
func (b *planner) branch(road, tier int) {
	r := b.plan.Roads[road]
	dir := b.heading(r)
	left, right := [2]int{dir[1], -dir[0]}, [2]int{-dir[1], dir[0]}
	// a child that runs across the screen needs a block of rows between
	// parents, a child that runs up and down a block of columns
	gap, span := 14, 3
	if dir[0] == 0 {
		gap, span = 10, 2
	}
	for pos := 4; pos < len(r.Cells); pos += gap + roll(span, b.seed, 44, int64(road), int64(pos)) {
		for s, side := range [][2]int{left, right} {
			if !chance(92, b.seed, 45, int64(road), int64(pos), int64(s)) {
				continue
			}
			length := 24 + roll(24, b.seed, 46, int64(road), int64(pos), int64(s))
			if side[0] != 0 {
				length *= 2 // a street running along the way the trunks run
			}
			b.grow(road, pos, side, length, tier)
		}
	}
}

// heading is the way a road runs, from its first cell to its last.
func (b *planner) heading(r Road) [2]int {
	first, last := r.Cells[0], r.Cells[len(r.Cells)-1]
	switch {
	case abs(last.X-first.X) >= abs(last.Y-first.Y) && last.X != first.X:
		return [2]int{cmp.Compare(last.X, first.X), 0}
	case last.Y != first.Y:
		return [2]int{0, cmp.Compare(last.Y, first.Y)}
	}
	return east
}

// grow lays one road out of cell `from` of road `parent` in direction dir and
// returns its index (-1 if it could not take a single cell). It goes
// straight, steps sideways round a snag, bridges a river, and stops at water,
// steep ground or another road.
func (b *planner) grow(parent, from int, dir [2]int, length, tier int) int {
	p := b.plan
	index := len(p.Roads)
	cur := p.Roads[parent].Cells[from]
	var cells []Pt
	for len(cells) < length {
		next := Pt{cur.X + dir[0], cur.Y + dir[1]}
		if inWorld(next.X, next.Y) && b.claim[at(next.X, next.Y)] == roadCell {
			break // it has met another road
		}
		if !b.canExtend(next, dir, index, parent) {
			next = b.nudge(cur, dir, index, parent)
			if next == cur {
				break
			}
		}
		if b.f.Biome[at(next.X, next.Y)] == River {
			p.Bridge[next] = true
		}
		b.claimRoad(next, index)
		cells = append(cells, next)
		cur = next
	}
	if len(cells) == 0 {
		return -1
	}
	p.Roads = append(p.Roads, Road{Cells: cells, Parent: parent, From: from, Tier: tier})
	return index
}

// canExtend says whether a road heading dir may take cell c: it must be
// passable, a river must have dry ground behind it, and no other road may run
// beside it.
func (b *planner) canExtend(c Pt, dir [2]int, road, parent int) bool {
	if !b.passable(c.X, c.Y) || b.claim[at(c.X, c.Y)] == roadCell {
		return false
	}
	if b.f.Biome[at(c.X, c.Y)] == River && !b.buildable(c.X+dir[0], c.Y+dir[1]) {
		return false
	}
	for _, side := range [][2]int{{dir[1], -dir[0]}, {-dir[1], dir[0]}} {
		s := Pt{c.X + side[0], c.Y + side[1]}
		if inWorld(s.X, s.Y) && b.claim[at(s.X, s.Y)] == roadCell && b.roadOf[at(s.X, s.Y)] != road && b.roadOf[at(s.X, s.Y)] != parent {
			return false
		}
	}
	return true
}

// nudge finds a cell one step to the side from which the road can carry on;
// it returns cur when there is none.
func (b *planner) nudge(cur Pt, dir [2]int, road, parent int) Pt {
	sides := [][2]int{{dir[1], -dir[0]}, {-dir[1], dir[0]}}
	if roll(2, b.seed, 47, int64(cur.X), int64(cur.Y)) == 1 {
		sides[0], sides[1] = sides[1], sides[0]
	}
	for _, side := range sides {
		c := Pt{cur.X + side[0], cur.Y + side[1]}
		ahead := Pt{c.X + dir[0], c.Y + dir[1]}
		if b.canExtend(c, dir, road, parent) && b.canExtend(ahead, dir, road, parent) {
			return c
		}
	}
	return cur
}
