package mapgen

// The space age: a planet below, rings of orbit above it, stations on them.

var space = AgeStyle{
	From:  "space_age",
	Scene: Orbit,
	Land: [numBiomes]Look{
		Deep:  {Glyphs: "≈~", Fg: P(HSea, 190), Bg: P(HSea, 255), Dense: 35},
		Grass: {Glyphs: "\"'.", Fg: P(HLand, 230), Bg: P(HLand, 255), Dense: 40},
		Hills: {Glyphs: "░▒", Fg: P(HCloud, 230), Bg: P(HCloud, 255), Dense: 60},
		Sand:  {Glyphs: "░", Fg: P(HAir, 255), Bg: P(HAir, 150)},
	},
	Roads: [3]RoadLook{
		{Joins: joinsRounded, Fg: P(HOrbit, 255), Bg: P(HVoid, 255)},
		{Joins: joinsHeavy, Fg: P(HWhite, 255), Bg: P(HVoid, 255)},
		{Joins: joinsLight, Fg: P(HOrbit, 200), Bg: P(HVoid, 255)},
	},
	Deck:   P(HSteel, 255),
	Yard:   Look{Bg: P(HVoid, 255)},
	Plaza:  Look{Bg: P(HVoid, 255)},
	Rubble: Look{Glyphs: "·∙", Fg: P(HRock, 255), Bg: P(HVoid, 255), Dense: 60},
	Shapes: [numFamilies]Tiers{
		Home: {
			{art(" ▄ \n[▓]")},
			{art("╒═▄═╕\n╘▓▓▓╛")},
			{art("╒═▄▄═╕\n║▓▓▓▓║\n╘════╛")},
			{art(" ▄██▄ \n╞████╡\n ▀██▀ ")},
		},
		Work: {
			{art("▤*▤")},
			{art("▤▤*▤▤\n ╞═╡ ")},
			{art("▤▤▤▤▤▤\n╞═*■═╡\n ╘══╛ ")},
			{art("▤▤▤▤▤▤\n╞■*■■╡\n▤▤▤▤▤▤")},
		},
		Civic: {
			{art(" ◠ \n[*]")},
			{art(" ◠◠ \n╞*▫╡")},
			{art("  Ψ   \n ◠◠◠◠ \n╞▫*▫▫╡")},
			{art("  Ψ   \n ◠██◠ \n╞▫*█▫╡")},
		},
	},
	Filler:  []Art{art("[▫]"), art("◘▫◘")},
	FillPer: 45,
	Wall:    HSteel,
	Ink:     map[rune]Paint{'▤': P(HPanel, 255), '■': P(HWhite, 255), '▫': P(HWhite, 255), 'Ψ': P(HRed, 255), '◠': P(HHydro, 255), '▓': P(HSteel, 255)},
	Windows: "▫",
	Agents: []AgentKind{
		{Name: "shuttle", Glyphs: "►◄▼▲", Fg: P(HEngine, 255), Per: 200, Max: 14, Slow: 1},
		{Name: "freighter", Glyphs: "►◄▼▲", Fg: P(HAmber, 255), Per: 450, Max: 8, Slow: 2},
	},
	Night: Night{Tint: HNight, Dim: 120, Light: HLight},
	Tones: map[byte]Paint{'s': P(HSteel, 255), 'a': P(HSun, 255), 'g': P(HEngine, 255), 'm': P(HSteel, 235), 'f': P(HSun, 255), 'n': P(HNebula, 255)},
	Sky:   Look{Glyphs: "·.∙·✧", Fg: P(HStar, 255), Bg: P(HVoid, 255), Dense: 46},
}
