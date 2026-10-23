package skyline

// wonders.go: the landmarks. Each age's wonder is a short drawing on the
// same cell grammar as the buildings, the tallest and most detailed thing in
// its district: wonderRows sizes it, wonderArt draws it, and while it is
// still being built scaffoldArt shows how much of it is banked.

// wonderFamily lets a landmark wear a later period's finish.
var wonderFamily = map[string]int{"crystal_palace": 7, "great_monolith": 8, "hoover_dam": 7}

// wonderNominal is each wonder's height in rows in a scene sized for a
// 45-row terminal at scale 1. A wonder is drawn this tall times the scene's
// scale, within the scene's own room (wonderRows).
var wonderNominal = map[string]float64{
	"sacred_grove": 14, "great_monolith": 22, "stonehenge": 18, "colosseum": 18, "parthenon": 20,
	"great_library": 26, "sistine_chapel": 30, "grand_lighthouse": 32, "crystal_palace": 28,
	"eiffel_tower": 34, "hoover_dam": 32, "particle_accelerator": 30, "space_program": 34,
	"global_network": 34, "world_simulation": 32, "neon_citadel": 34, "stellar_cradle": 34,
	"dyson_scaffold": 26, "warp_nexus": 24, "cosmic_beacon": 32, "reality_anchor": 24, "singularity_core": 22,
}

// wonderMinRows is the least a landmark is drawn at: under it the shape is
// not a silhouette any more.
const wonderMinRows = 6

// wonderRows is how many rows the wonder key stands in a scene of the given
// scale whose room above the ground is ref rows: its nominal height scaled,
// never more than the scene holds and never under wonderMinRows.
func wonderRows(key string, scale float64, ref int) int {
	n := wonderNominal[key]
	if n == 0 {
		n = 14
	}
	h := int(n*scale + 0.5)
	return clampInt(h, wonderMinRows, max(wonderMinRows, ref-2))
}

// wonderSprite is the wonder key drawn for a ground scene of this scale and
// room.
func wonderSprite(key string, scale float64, ref int) *sprite {
	return wonderArt(key, wonderRows(key, scale, ref))
}

// fitRows makes s exactly rows tall: a taller one loses its top, a shorter
// one gains empty rows above.
func fitRows(s *sprite, rows int) *sprite {
	switch {
	case s.h > rows:
		return s.cropTop(rows)
	case s.h < rows:
		return s.raise(rows - s.h)
	}
	return s
}

// wonderArt draws the wonder key rows tall.
func wonderArt(key string, rows int) *sprite {
	var s *sprite
	switch key {
	case "sacred_grove":
		s = grove(rows)
	case "great_monolith":
		s = monolith(rows)
	case "stonehenge":
		s = henge(rows)
	case "colosseum":
		s = colosseum(rows)
	case "parthenon":
		s = parthenon(rows)
	case "great_library":
		s = library(rows)
	case "sistine_chapel":
		s = basilica(rows)
	case "grand_lighthouse":
		s = lighthouse(rows)
	case "crystal_palace":
		s = glassHall(rows)
	case "eiffel_tower":
		s = lattice(rows)
	case "hoover_dam":
		s = dam(rows)
	case "particle_accelerator":
		s = accelerator(rows)
	case "space_program":
		s = rocket(rows)
	case "global_network":
		s = network(rows)
	case "world_simulation":
		s = simulation(rows)
	case "neon_citadel":
		s = citadel(rows)
	case "stellar_cradle":
		s = cradle(rows)
	case "dyson_scaffold":
		s = dyson(rows)
	case "warp_nexus":
		s = ringGate(rows)
	case "cosmic_beacon":
		s = beacon(rows)
	case "reality_anchor":
		s = anchor(rows)
	case "singularity_core":
		s = singularity(rows)
	default: // a wonder with no drawing of its own: a plain obelisk
		s = aObelisk(newRnd(hash(len(key), int(key[0]))), rows)
	}
	return fitRows(s, rows)
}

// siteRows is how many rows of a rows-tall wonder are standing when share of
// its cost is banked: none only while nothing is, one at least once anything is.
func siteRows(share float64, rows int) int {
	if share <= 0 {
		return 0
	}
	return clampInt(int(share*float64(rows)+0.5), 1, rows)
}

// scaffoldArt is the wonder key under construction: the bottom built rows of
// the finished landmark, timber (or, from the industrial age, steel)
// scaffolding round them and two rows above, so the frame climbs as the
// wonder is banked. With nothing built it is a footing and two posts; with
// all of it built, the whole wonder in its frame.
func scaffoldArt(key string, rows, built int, steel bool) *sprite {
	fin := wonderArt(key, rows)
	const lift = 2
	h := min(rows, built+lift)
	s := newSprite(fin.w, h)
	for y := 0; y < built && y < h; y++ {
		for x := 0; x < fin.w; x++ {
			if c := fin.at(x, fin.h-1-y); c.ch != 0 {
				s.put(x, h-1-y, c.ch, c.fg, c.bg)
			}
		}
	}
	// the footprint: the columns the work stands on, one wider each side
	x0, x1 := fin.w, -1
	rowsToScan := max(built, 1)
	for y := 0; y < rowsToScan; y++ {
		for x := 0; x < fin.w; x++ {
			if fin.at(x, fin.h-1-y).ch != 0 {
				x0, x1 = min(x0, x), max(x1, x)
			}
		}
	}
	if x1 < 0 {
		x0, x1 = 0, fin.w-1
	}
	x0, x1 = max(0, x0-1), min(fin.w-1, x1+1)
	pole := sTrunk
	if steel {
		pole = sMetal
	}
	// a footing of trodden earth
	for x := x0; x <= x1; x++ {
		if s.at(x, h-1).ch == 0 {
			s.put(x, h-1, '▄', sRockDark, sNone)
		}
	}
	posts := []int{x0, x1}
	for x := x0 + 6; x < x1-2; x += 6 {
		posts = append(posts, x)
	}
	top := 0
	for _, px := range posts {
		for y := top; y < h-1; y++ {
			if s.at(px, y).ch == 0 {
				s.put(px, y, '│', pole, sNone)
			}
		}
	}
	// the working deck at the top, planks every few rows below it
	for x := x0; x <= x1; x++ {
		if s.at(x, top).ch == 0 {
			s.put(x, top, '─', pole, sNone)
		}
	}
	s.put(x0, top, '┬', pole, sNone)
	s.put(x1, top, '┬', pole, sNone)
	for y := top + 4; y < h-1; y += 4 {
		for x := x0; x <= x1; x++ {
			if s.at(x, y).ch == 0 {
				s.put(x, y, '─', pole, sNone)
			}
		}
	}
	// braces in the end bays
	for y := top + 1; y < h-1; y++ {
		d := (y - top) % 4
		if d == 0 {
			continue
		}
		if x := x0 + d; x < x1 && s.at(x, y).ch == 0 {
			s.put(x, y, '╲', pole, sNone)
		}
		if x := x1 - d; x > x0 && s.at(x, y).ch == 0 {
			s.put(x, y, '╱', pole, sNone)
		}
	}
	return s
}
