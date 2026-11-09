package world

import (
	"cmp"
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
)

// The near-star catalogue: the stars round Sol at their true distances. One
// scale serves every star, so a distance reads true; a star's direction is
// its galactic longitude. Moved here from mapgen so the terminal map and the
// native map read the same sky.

// NearStar is a star of the neighbourhood.
type NearStar struct {
	Key     string // the name's key, for the inspect line (vibe: mapstyle starNames)
	Dist100 int    // light-years from Sol, times 100
	Lon     int    // galactic longitude, degrees
	Look    Look
}

// Look is how a star is drawn: its colour, how big it shows (0 a single
// pixel, 1 a small disc, 2 a bright disc with a glow), how bright a dim one
// is, and the companion beside it, if any.
type Look struct {
	Tone   Tone
	Size   int
	Weight int
	Pair   Tone // a second star beside it
	Extra  Tone // a third, dim one, or a white dwarf
}

// NearStars are the stars of the neighbourhood, nearest first. Distances are
// the catalogue's; longitudes are rounded to a degree.
var NearStars = [...]NearStar{
	{"alpha_centauri", 437, 316, Look{ToneSun, 1, 255, ToneAmber, ToneRed}},
	{"barnards_star", 596, 31, Look{ToneRed, 0, 190, ToneNone, ToneNone}},
	{"wolf_359", 786, 244, Look{ToneRed, 0, 150, ToneNone, ToneNone}},
	{"lalande_21185", 831, 185, Look{ToneRed, 0, 190, ToneNone, ToneNone}},
	{"sirius", 860, 227, Look{ToneCool, 2, 255, ToneNone, ToneWhite}},
	{"luyten_726_8", 870, 175, Look{ToneRed, 0, 170, ToneRed, ToneNone}},
	{"ross_154", 970, 11, Look{ToneRed, 0, 160, ToneNone, ToneNone}},
	{"ross_248", 1030, 110, Look{ToneRed, 0, 140, ToneNone, ToneNone}},
	{"epsilon_eridani", 1050, 196, Look{ToneAmber, 1, 220, ToneNone, ToneNone}},
	{"lacaille_9352", 1070, 5, Look{ToneRed, 0, 200, ToneNone, ToneNone}},
	{"ross_128", 1100, 270, Look{ToneRed, 0, 150, ToneNone, ToneNone}},
	{"61_cygni", 1140, 82, Look{ToneAmber, 1, 190, ToneAmber, ToneNone}},
	{"procyon", 1150, 214, Look{ToneSun, 2, 230, ToneNone, ToneWhite}},
	{"struve_2398", 1150, 89, Look{ToneRed, 0, 170, ToneRed, ToneNone}},
	{"groombridge_34", 1160, 117, Look{ToneRed, 0, 160, ToneRed, ToneNone}},
	{"epsilon_indi", 1190, 336, Look{ToneAmber, 1, 190, ToneNone, ToneNone}},
	{"tau_ceti", 1190, 173, Look{ToneSun, 1, 210, ToneNone, ToneNone}},
	{"teegardens_star", 1250, 160, Look{ToneRed, 0, 130, ToneNone, ToneNone}},
	{"kapteyns_star", 1280, 251, Look{ToneRed, 0, 150, ToneNone, ToneNone}},
}

// RingLy is the gap between the faint distance rings, in light-years.
const RingLy = 5

// Colony is a building line with something standing, and so a lit star of the
// neighbourhood.
type Colony struct {
	Line    string
	Count   int // buildings of the line
	Systems int // systems of its sector that shine
	Tier    int // 0..3, the rank of the best building standing
}

// ColonyStars says which star each colony lights: the colonies keep a star for
// good, the earlier lines the nearer ones, with a seeded shuffle among stars
// of a like distance.
func ColonyStars(seed int64, cs []Colony) []int {
	order := make([]int, len(NearStars))
	for i := range order {
		order[i] = i
	}
	jitter := func(i int) int { return NearStars[i].Dist100 + Roll(300, seed, 6100, int64(i)) }
	slices.SortStableFunc(order, func(a, b int) int { return cmp.Compare(jitter(a), jitter(b)) })
	out := make([]int, len(cs))
	for i, c := range cs {
		j := lineIndex[c.Line]
		rank := 0 // where this line stands among all the lines, with its own jitter
		for k := 0; k < len(mapmodel.LineageOrder); k++ {
			if k*100+Roll(250, seed, 6101, int64(k)) < j*100+Roll(250, seed, 6101, int64(k)) {
				rank++
			}
		}
		out[i] = order[rank]
	}
	return out
}

// Lane joins a colony star to the node it hangs from.
type Lane struct {
	From, To int // indexes into NearStars; From -1 is Sol
	Line     string
	FromLine string // the line of the node it leaves ("home" for Sol)
	Traffic  int    // 1 to 3
}

// Lanes joins every lit star to Sol and to each other: each to the nearest
// of Sol and the colonies already joined (by the pts given, in any common
// unit), so the web holds together. colony maps a NearStars index to its
// colony; pts are the star positions with Sol at sol.
func Lanes(sol [2]int, pts [][2]int, colony map[int]Colony) []Lane {
	type node struct {
		idx, x, y int
		line      string
		count     int
	}
	nodes := []node{{-1, sol[0], sol[1], "home", 0}}
	var out []Lane
	for i := range pts { // nearest first
		c, ok := colony[i]
		if !ok {
			continue
		}
		best, gap := 0, 1<<60
		for j, n := range nodes {
			if g := Sq(n.x-pts[i][0]) + 4*Sq(n.y-pts[i][1]); g < gap {
				best, gap = j, g
			}
		}
		from := nodes[best]
		m := c.Count // the lane is as busy as the smaller colony it joins
		if from.count > 0 {
			m = min(m, from.count)
		}
		traffic := 1
		if m >= 40 {
			traffic = 3
		} else if m >= 8 {
			traffic = 2
		}
		out = append(out, Lane{From: from.idx, To: i, Line: c.Line, FromLine: from.line, Traffic: traffic})
		nodes = append(nodes, node{i, pts[i][0], pts[i][1], c.Line, c.Count})
	}
	return out
}

var lineIndex = func() map[string]int {
	out := map[string]int{}
	for i, l := range mapmodel.LineageOrder {
		out[l] = i
	}
	return out
}()
