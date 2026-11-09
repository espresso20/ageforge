package mapgen

import "github.com/espresso20/ageforge/world"

// The people's territory on the star map. Each building line owns a sector
// on the home arm, a hub with a cluster of systems round it; the cluster
// lights as the line grows, a log of its count. The civilizations met stand
// in smaller clusters on the other arms. Where everything can stand is fixed
// by the seed here; which of it is lit is the model's business (stars_paint.go).

const (
	sectorMax  = world.SectorMax
	rivalMax   = world.RivalMax
	rivalSlots = world.RivalSlots
)

type Sector = world.Sector

func systemsFor(count, most int) int { return world.SystemsFor(count, most) }

// layTerritory fixes the sectors along the home arm, the places for rivals on
// the other arms and the rift (world/ lays them).
func (p *Plan) layTerritory(seed int64) {
	t := world.LayTerritory(seed, p.Galaxy)
	p.Sectors, p.Rivals, p.Rift = t.Sectors, t.Rivals, t.Rift
}

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
