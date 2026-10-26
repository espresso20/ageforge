package mapgen

import (
	"maps"
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
)

// Placement answers "which parcel does this building stand on". A building
// kind's n-th parcel is the parcel at ordinal n*kinds+rank in its line's list,
// where kinds is how many kinds the line has in all and rank is the kind's
// place among them, oldest first. That is a pure function of the kind, so a
// building never moves when others are added. The list runs nearest the
// line's heart first.

// growth turns a kind's count into parcels and a size tier. Both only grow
// with the count, so a parcel's tier rises in place.
var growth = []struct{ From, Parcels, Tier int }{
	{1, 1, 0}, {4, 1, 1}, {10, 1, 2}, {16, 2, 2}, {30, 2, 3},
}

func growthFor(count int) (parcels, tier int) {
	for _, g := range growth {
		if count >= g.From {
			parcels, tier = g.Parcels, g.Tier
		}
	}
	return parcels, tier
}

var lineOf = func() map[string]int {
	out := map[string]int{}
	for i, l := range mapmodel.LineageOrder {
		out[l] = i
	}
	return out
}()

// familyOf is the silhouette each building line is drawn with.
var familyOf = map[string]Family{
	mapmodel.LinHousing: Home, mapmodel.LinFood: Home,
	mapmodel.LinWood: Work, mapmodel.LinMines: Work, mapmodel.LinMetal: Work, mapmodel.LinEngineer: Work,
	mapmodel.LinEnergy: Work, mapmodel.LinHarbor: Work, mapmodel.LinHacker: Work, mapmodel.LinStorage: Work,
	mapmodel.LinKnowledge: Civic, mapmodel.LinFaith: Civic, mapmodel.LinCulture: Civic, mapmodel.LinMonument: Civic,
	mapmodel.LinTrade: Civic, mapmodel.LinMilitary: Civic, mapmodel.LinDiplomacy: Civic,
}

// lineTints colour each line's buildings a little, so districts read apart.
var lineTints = map[string]RGB{
	mapmodel.LinHousing: rgb(0xd8b890), mapmodel.LinFood: rgb(0xb8c860), mapmodel.LinWood: rgb(0xa0703c),
	mapmodel.LinMines: rgb(0x9a9078), mapmodel.LinMetal: rgb(0xe08848), mapmodel.LinEngineer: rgb(0x90a8c0),
	mapmodel.LinEnergy: rgb(0xe8d040), mapmodel.LinHarbor: rgb(0x48b0b8), mapmodel.LinHacker: rgb(0xd050c0),
	mapmodel.LinKnowledge: rgb(0x8a7ae0), mapmodel.LinFaith: rgb(0xf0e8c8), mapmodel.LinCulture: rgb(0xe07898),
	mapmodel.LinMonument: rgb(0xe0b040), mapmodel.LinTrade: rgb(0xe0c050), mapmodel.LinStorage: rgb(0x8a7a60),
	mapmodel.LinMilitary: rgb(0xc85050), mapmodel.LinDiplomacy: rgb(0x6a90b8),
}

// lineSigns are the signs hung on buildings, the symbols of the map legend:
// a * in art is drawn as its line's sign.
var lineSigns = map[string]rune{
	mapmodel.LinHousing: '∩', mapmodel.LinFood: '≡', mapmodel.LinWood: '♠', mapmodel.LinMines: '▼',
	mapmodel.LinMetal: '♦', mapmodel.LinEngineer: '¤', mapmodel.LinEnergy: '☼', mapmodel.LinHarbor: '╤',
	mapmodel.LinHacker: 'λ', mapmodel.LinKnowledge: '§', mapmodel.LinFaith: 'Ω', mapmodel.LinCulture: '♫',
	mapmodel.LinMonument: '◊', mapmodel.LinTrade: '$', mapmodel.LinStorage: '▣', mapmodel.LinMilitary: '†',
	mapmodel.LinDiplomacy: '≡',
}

// placed is one building kind standing on one parcel.
type placed struct {
	lot      int // index into Plan.Lots
	b        *mapmodel.Building
	instance int
	tier     int
	ruin     bool
}

// place puts every building of the model on its parcel. A line with fewer
// parcels than kinds wraps round: the newest kind takes the parcel of an
// older one, the way a town rebuilds its old streets, so the newest kinds
// always show. A line with no parcels shows nothing.
func (p *Plan) place(m *mapmodel.Model) []placed {
	type claim struct {
		pl   placed
		rank int
	}
	owner := map[int]claim{}
	stand := func(b *mapmodel.Building, instance, tier int, ruin bool) {
		line := lineOf[b.Lineage]
		lots := p.Order[line]
		if len(lots) == 0 {
			return
		}
		ordinal := instance*len(m.Catalog.ByLineage[b.Lineage]) + b.Rank
		lot := lots[ordinal%len(lots)]
		if prev, taken := owner[lot]; !taken || b.Rank > prev.rank {
			owner[lot] = claim{placed{lot: lot, b: b, instance: instance, tier: tier, ruin: ruin}, b.Rank}
		}
	}
	for _, b := range m.Buildings {
		if _, ok := lineOf[b.Lineage]; !ok || b.Wonder {
			continue
		}
		parcels, tier := growthFor(b.Count)
		for i := 0; b.Count > 0 && i < parcels; i++ {
			stand(b, i, tier, false)
		}
		if b.Ruins > 0 && parcels < 2 {
			stand(b, 1, tier, true)
		}
	}
	out := make([]placed, 0, len(owner))
	for _, lot := range slices.Sorted(maps.Keys(owner)) {
		out = append(out, owner[lot].pl)
	}
	return out
}
