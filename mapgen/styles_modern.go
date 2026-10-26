package mapgen

// The modern city: concrete, glass and asphalt, with the last parks.

var modern = AgeStyle{
	From:  "industrial_age",
	Scene: City,
	Land: [numBiomes]Look{
		Deep:     {Glyphs: "≈~", Fg: rgb(0x2a6ab0), Bg: rgb(0x0a1d36), Dense: 55},
		Shallows: {Glyphs: "~", Fg: rgb(0x3a86c8), Bg: rgb(0x0e2c4a), Dense: 40},
		Sand:     {Glyphs: "·.", Fg: rgb(0x8a8068), Bg: rgb(0x23221c), Dense: 25},
		Scrub:    {Glyphs: ",'.", Fg: rgb(0x6a7448), Bg: rgb(0x171a14), Dense: 25},
		Grass:    {Glyphs: "\"',.", Fg: rgb(0x4f8a3a), Bg: rgb(0x111a14), Dense: 22},
		Forest:   {Glyphs: "♣♣τ♠", Fg: rgb(0x3f9a4f), Bg: rgb(0x0e1c13), Dense: 60},
		Hills:    {Glyphs: "^n∩", Fg: rgb(0x8a8470), Bg: rgb(0x1a1b16), Dense: 35},
		Mountain: {Glyphs: "▲^", Fg: rgb(0x9a9488), Bg: rgb(0x1a1a1d), Dense: 70},
		River:    {Fg: rgb(0x3a86c8), Bg: rgb(0x0e2c4a)},
	},
	Roads: [3]RoadLook{
		{Joins: joinsHeavy, Fg: rgb(0x8a929c), Bg: rgb(0x23262c)},
		{Joins: joinsLight, Fg: rgb(0x7c848e), Bg: rgb(0x1f2228)},
		{Joins: joinsLight, Fg: rgb(0x666e78), Bg: rgb(0x1b1e24)},
	},
	Deck:   rgb(0x4a4e56),
	Yard:   Look{Glyphs: "·", Fg: rgb(0x34383f), Bg: rgb(0x1c1f25), Dense: 8},
	Crops:  Look{Glyphs: "≡", Fg: rgb(0x8a9a3a), Bg: rgb(0x1a2010), Dense: 55},
	Plaza:  Look{Glyphs: "\"♣", Fg: rgb(0x4fa64f), Bg: rgb(0x13201a), Dense: 22},
	Rubble: Look{Glyphs: "░▒%", Fg: rgb(0x8c7448), Bg: rgb(0x1c1f25), Dense: 70},
	Shapes: [numFamilies]Tiers{
		Home: {
			{art("▟▙\n▐▪▌")},
			{art(" ▄▄▄▄ \n▐▪▪▪▪▌\n▐▪▮▪▪▌")},
			{art("▄▄▄▄▄▄\n█▪▪▪▪█\n█▪▪▪▪█")},
			{art("  ▄▄  \n  ██  \n ▐▪▪▌ \n ▐▪▪▌ ")},
		},
		Work: {
			{art("┃ \n▐*▌")},
			{art("┃ ┃ \n▐███▌\n▐▪*▪▌")},
			{art("┃  ┃ \n▟████▙\n█▪*▪▪█")},
			{art("┃ ┃ ┃ \n▟█████\n█▪▪▪▪▪\n█▪*▪▪▪")},
		},
		Civic: {
			{art(" ▲ \n▐*▌")},
			{art(" ▲▲ \n▟██▙\n█▪*█")},
			{art("  ▲   \n ▟██▙ \n▟████▙\n█▪*▪▪█")},
			{art("  ╽   \n ▄██▄ \n ▐▪▪▌ \n ▐*▮▌ ")},
		},
	},
	Wall:    rgb(0x9096a0),
	Ink:     map[rune]RGB{'▪': rgb(0x6aa6dc), '▮': rgb(0x50545c), '▲': rgb(0xb4c0cc), '╽': rgb(0xdcf2ff), '┃': rgb(0x8c7c66), '▀': rgb(0xb4c0cc)},
	Windows: "▪",
	Chimney: '┃',
	Smoke:   Look{Glyphs: "∘○", Fg: rgb(0x8a8e96)},
	Scatter: []Scatter{
		{On: Grass, Glyphs: "♣", Fg: rgb(0x4fa64f), Chance: 10},
		{On: Scrub, Glyphs: "♠", Fg: rgb(0x6a9a58), Chance: 8},
		{On: Hills, Glyphs: "▲", Fg: rgb(0x9a9488), Chance: 8},
	},
	Agents: []AgentKind{
		{Name: "car", Glyphs: "■", Fg: rgb(0xe05a4a), Per: 140, Max: 18, Slow: 1},
		{Name: "car", Glyphs: "■", Fg: rgb(0xdcdcdc), Per: 170, Max: 18, Slow: 1},
		{Name: "taxi", Glyphs: "■", Fg: rgb(0xe0b83a), Per: 400, Max: 8, Slow: 1},
	},
	Night: Night{Tint: rgb(0x0a0814), Dim: 185, Light: rgb(0xffc45a), Lamps: true},
	Tones: map[byte]RGB{'s': rgb(0x8a929c), 'a': rgb(0xc09a3a), 'g': rgb(0x6aa6dc), 'm': rgb(0x9096a0), 'f': rgb(0xff9440), 'v': rgb(0x4fa64f)},
}
