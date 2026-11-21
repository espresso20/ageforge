package shields

// The small emblems are drawn by hand for a field eight pixels wide: X is
// metal, h its light, s its shade, d the deep tone, g the warm spark.
var smallEmblems = [NumDomains][]string{
	Food: {
		"..h..",
		"h.X.h",
		"X.X.X",
		".XXX.",
		".ddd.",
		".XXX.",
		"X.X.X",
	},
	Lumber: {
		".hXX..",
		"hXXXs.",
		".XXXs.",
		"..Xs..",
		".Xs...",
		"Xs....",
		"s.....",
	},
	Masonry: {
		".hXXs.",
		"hX..Xs",
		"XX..XX",
		"Xd..dX",
		"XX..XX",
		"Xd..dX",
		"XXXXXX",
	},
	Faith: {
		"..h..",
		".hXs.",
		".XgX.",
		"XgggX",
		"XXgXX",
		".XXX.",
		"ddddd",
	},
	Trade: {
		"...h...",
		"XXXXXXX",
		".X.X.X.",
		".X.X.X.",
		"XXX.XXX",
		"...X...",
		"..XXX..",
	},
	Engineering: {
		".X.X.X.",
		"XXXXXXs",
		".XX.XXs",
		"XX...XX",
		".XX.XXs",
		"XXXXXXs",
		".X.X.Xs",
	},
	Energy: {
		"..XXX.",
		".XXX..",
		"XXXXX.",
		"..XX..",
		".XX...",
		".X....",
		"X.....",
	},
	Hacker: {
		"XXXXXXX",
		"Xs....X",
		"X.g...X",
		"X..g..X",
		"X.g.hhX",
		"XXXXXXX",
	},
	Astronaut: {
		"..h..",
		".hXs.",
		".XdX.",
		".XXXs",
		"XXXXX",
		"X.X.X",
		"..g..",
	},
	Military: {
		"..h..",
		"..X..",
		"..X..",
		"..X..",
		"XXXXX",
		"..d..",
		"..X..",
	},
	Metallurgy: {
		"..h.h..",
		"XXXXXXX",
		".XXXXXs",
		"..XXX..",
		".XXXXXs",
		"XXXXXXX",
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
