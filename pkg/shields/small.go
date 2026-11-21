package shields

// The small emblems are drawn by hand for a field eight pixels wide: X is
// metal, h its light, s its shade, d the deep tone, g the warm spark.
var smallEmblems = [NumDomains][]string{
	Food: { // a wheat ear: a stalk with its grains stacked down both sides
		"...hh...",
		".h.XX.s.",
		"hXhXXsXs",
		"...XX...",
		".h.XX.s.",
		"hXhXXsXs",
		"...XX...",
		"...XX...",
	},
	Lumber: { // an axe: a haft running down left, the head up right with a bright edge
		"...hXXh.",
		"..hXXXXh",
		"..dXXXXh",
		"..hXXXh.",
		"..XXss..",
		".hXs....",
		"hXs.....",
		"Xs......",
	},
	Masonry: { // a keystone arch of bricks on a flat base
		"..hggs..",
		".hXXXXXs",
		"hXd..dXs",
		"XXX..XXX",
		"XdX..XdX",
		"XXX..XXX",
		"XdX..XdX",
		"hhhhhhhs",
	},
	Faith: { // a flame
		"...X....",
		"..XX....",
		"..XXX...",
		".hXXXXs.",
		".XXgXXs.",
		".XgggXs.",
		".XXgXXs.",
		"..sXXs..",
	},
	Trade: { // scales: beam, post, two hanging pans
		"...hh...",
		"XXXXXXXX",
		"X.XhsX.X",
		"X.XhsX.X",
		"XXXhsXXX",
		".X.hs.X.",
		"...hs...",
		"..XXXX..",
	},
	Engineering: { // a gear with a hole through it
		"...XX...",
		".X.XX.X.",
		"..hXXX..",
		"hXXddXXs",
		"XXXddXXs",
		"..XXXs..",
		".X.XX.X.",
		"...XX...",
	},
	Energy: { // a lightning bolt
		"....hXXX",
		"...hXXs.",
		"..hXXs..",
		".hXXXXXs",
		"..sXXXs.",
		"..hXXs..",
		"..XXs...",
		".XXs....",
	},
	Hacker: { // a terminal: prompt and cursor
		"hXXXXXXs",
		"XddddddX",
		"XgdddddX",
		"XdgddddX",
		"XgdhhddX",
		"XddddddX",
		"sXXXXXXs",
	},
	Astronaut: { // a rocket: window, fins, flame
		"...hX...",
		"..hXXs..",
		"..XddX..",
		"..XddX..",
		".XhXXsX.",
		"XX.XX.XX",
		"X..gg..X",
		"...gg...",
	},
	Military: { // a sword, point up
		"...hX...",
		"...hX...",
		"...hX...",
		"...hX...",
		"...hX...",
		".hXXXXs.",
		"...Xs...",
		"..dXXd..",
	},
	Metallurgy: { // an anvil with its horn to the left, sparks over it
		"...g.g..",
		"hhhhhhhh",
		".hXXXXXs",
		"..sXXXXs",
		"...XXXs.",
		"...XXXs.",
		"..XXXXXs",
		".XXXXXXX",
	},
}

func smallTone(c byte) uint8 {
	switch c {
	case 'X':
		return tNoBevl
	case 'h':
		return tLight
	case 's':
		return tShade
	case 'd':
		return tDeep
	case 'g':
		return tGem
	}
	return 0
}
