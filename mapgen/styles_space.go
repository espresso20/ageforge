package mapgen

// The space age: a planet below, rings of orbit above it, stations on them.

var space = AgeStyle{
	From:  "space_age",
	Scene: Orbit,
	Land: [numBiomes]Look{
		Deep:  {Glyphs: "≈~", Fg: rgb(0x14406e), Bg: rgb(0x0b2648), Dense: 35},
		Grass: {Glyphs: "\"'.", Fg: rgb(0x2e5a40), Bg: rgb(0x1b3226), Dense: 40},
		Hills: {Glyphs: "░▒", Fg: rgb(0x6a86aa), Bg: rgb(0x46628a), Dense: 60},
		Sand:  {Glyphs: "░", Fg: rgb(0x4fb0ff), Bg: rgb(0x1f5a9a)},
	},
	Roads: [3]RoadLook{
		{Joins: joinsRounded, Fg: rgb(0x6a64a0), Bg: rgb(0x04060c)},
		{Joins: joinsHeavy, Fg: rgb(0xdcf2ff), Bg: rgb(0x04060c)},
		{Joins: joinsLight, Fg: rgb(0x4c4676), Bg: rgb(0x04060c)},
	},
	Deck:   rgb(0x58667a),
	Yard:   Look{Bg: rgb(0x04060c)},
	Plaza:  Look{Bg: rgb(0x04060c)},
	Rubble: Look{Glyphs: "·∙", Fg: rgb(0x8f8172), Bg: rgb(0x04060c), Dense: 60},
	Shapes: [numFamilies]Tiers{
		Home: {
			{art(" ▄ \n[▓]")},
			{art("╒═▄═╕\n╘▓▓▓╛")},
			{art("╒═▄▄═╕\n║▓▓▓▓║\n╘════╛")},
			{art(" ▄██▄ \n╞████╡\n ▀██▀ ")},
		},
		Work: {
			{art("▤■▤")},
			{art("▤▤■▤▤\n ╞═╡ ")},
			{art("▤▤▤▤▤▤\n╞═■■═╡\n ╘══╛ ")},
			{art("▤▤▤▤▤▤\n╞■■■■╡\n▤▤▤▤▤▤")},
		},
		Civic: {
			{art(" ◠ \n[▫]")},
			{art(" ◠◠ \n╞▫▫╡")},
			{art("  Ψ   \n ◠◠◠◠ \n╞▫▫▫▫╡")},
			{art("  Ψ   \n ◠██◠ \n╞▫██▫╡")},
		},
	},
	Wall:    rgb(0x9aa8bc),
	Ink:     map[rune]RGB{'▤': rgb(0x3f78d8), '■': rgb(0xeef3fa), '▫': rgb(0xffe6b8), 'Ψ': rgb(0xff4a3a), '◠': rgb(0x7fd88e), '▓': rgb(0xc6cee2)},
	Windows: "▫",
	Agents: []AgentKind{
		{Name: "shuttle", Glyphs: "►◄▼▲", Fg: rgb(0x78f0ff), Per: 200, Max: 14, Slow: 1},
		{Name: "freighter", Glyphs: "►◄▼▲", Fg: rgb(0xffb347), Per: 450, Max: 8, Slow: 2},
	},
	Night: Night{Tint: rgb(0x02030a), Dim: 120, Light: rgb(0xffdc8a)},
	Tones: map[byte]RGB{'s': rgb(0xc6cee2), 'a': rgb(0xffd27a), 'g': rgb(0x78f0ff), 'm': rgb(0x9aa8bc), 'f': rgb(0xffd27a), 'n': rgb(0x3a1f66)},
	Sky:   Look{Glyphs: "·.∙·✧", Fg: rgb(0xb8c6e6), Bg: rgb(0x04060c), Dense: 46},
}
