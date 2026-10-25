package skyline

import "github.com/espresso20/ageforge/mapmodel"

// wonders.go: the landmarks. Each age's wonder is a short drawing on the
// same cell grammar as the buildings, the most detailed thing in its district
// (not the tallest: it stands as high as its form wants and may be shorter than
// a tower beside it): wonderRows sizes it, wonderArt draws it, and while it is
// still being built scaffoldArt shows how much of it is banked.

// wonderFamily lets a landmark wear a later period's finish.
var wonderFamily = map[string]int{"crystal_palace": 7, "great_monolith": 8, "hoover_dam": 32, "particle_accelerator": 30}

// wonderSpec is one wonder: its height in rows in a scene sized for a 45-row
// terminal at scale 1 (it is drawn this tall times the scene's scale, within
// the scene's own room: wonderRows) and the function that draws it.
type wonderSpec struct {
	nominal float64
	draw    func(rows int) *sprite
}

var wonderSpecs = map[string]wonderSpec{
	"sacred_grove":         {14, grove},
	"great_monolith":       {22, monolith},
	"stonehenge":           {22, henge},
	"colosseum":            {18, colosseum},
	"parthenon":            {20, parthenon},
	"great_library":        {26, library},
	"sistine_chapel":       {30, basilica},
	"grand_lighthouse":     {32, lighthouse},
	"crystal_palace":       {28, glassHall},
	"eiffel_tower":         {34, lattice},
	"hoover_dam":           {32, dam},
	"particle_accelerator": {30, accelerator},
	"space_program":        {34, rocket},
	"global_network":       {34, network},
	"world_simulation":     {32, simulation},
	"neon_citadel":         {34, citadel},
	"stellar_cradle":       {34, cradle},
	"dyson_scaffold":       {26, dyson},
	"warp_nexus":           {24, ringGate},
	"cosmic_beacon":        {32, beacon},
	"reality_anchor":       {24, anchor},
	"singularity_core":     {22, singularity},
}

// wonderMinRows is the least a landmark is drawn at: under it the shape is
// not a silhouette any more.
const wonderMinRows = 6

// wonderRows is how many rows the wonder key stands in a scene of the given
// scale whose room above the ground is ref rows: its nominal height scaled,
// never more than the scene holds and never under wonderMinRows.
func wonderRows(key string, scale float64, ref int) int {
	n := wonderSpecs[key].nominal
	if n == 0 {
		n = 14
	}
	h := int(n*scale + 0.5)
	return clampInt(h, wonderMinRows, max(wonderMinRows, ref-2))
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
	if w := wonderSpecs[key]; w.draw != nil {
		s = w.draw(rows)
	} else { // a wonder with no drawing of its own: a plain obelisk
		s = aObelisk(newRnd(hash(len(key), int(key[0]))), rows)
	}
	return fitRows(s, rows)
}

// scaffoldEra is how an age raises its scaffolding: lashed timber before the
// industrial age, riveted iron in it and the electric and atomic ages,
// tubular steel from the modern age on.
type scaffoldEra struct {
	slot         slot
	up, bar, tee rune // the upright, the plank, the joint at the top corners
}

func scaffoldEraOf(age int, c *mapmodel.Catalog) scaffoldEra {
	industrial, modern := c.AgeIdx["industrial_age"], c.AgeIdx["modern_age"]
	switch {
	case age < industrial:
		return scaffoldEra{sTrunk, '│', '─', '┬'}
	case age < modern:
		return scaffoldEra{sMetalDark, '┃', '━', '┳'}
	}
	return scaffoldEra{sMetal, '║', '═', '╦'}
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
func scaffoldArt(key string, rows, built int, era scaffoldEra) *sprite {
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
	pole, up, bar, tee := era.slot, era.up, era.bar, era.tee
	// a footing of trodden earth, in patches
	for x := x0; x <= x1; x++ {
		if s.at(x, h-1).ch == 0 && (x-x0)%3 != 1 {
			s.put(x, h-1, '▄', sRockDark, sNone)
		}
	}
	const top = 0 // the working deck, the topmost row of the frame
	posts := []int{x0, x1}
	for x := x0 + 6; x < x1-2; x += 6 {
		posts = append(posts, x)
	}
	for _, px := range posts {
		for y := top; y < h-1; y++ {
			if s.at(px, y).ch == 0 {
				s.put(px, y, up, pole, sNone)
			}
		}
	}
	// the working deck at the top, planks every few rows below it
	for x := x0; x <= x1; x++ {
		if s.at(x, top).ch == 0 {
			s.put(x, top, bar, pole, sNone)
		}
	}
	s.put(x0, top, tee, pole, sNone)
	s.put(x1, top, tee, pole, sNone)
	for y := top + 4; y < h-1; y += 4 {
		for x := x0; x <= x1; x++ {
			if s.at(x, y).ch == 0 {
				s.put(x, y, bar, pole, sNone)
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
