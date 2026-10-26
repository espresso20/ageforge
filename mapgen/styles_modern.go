package mapgen

// The modern city: concrete, glass and asphalt, with the last parks.

var modern = AgeStyle{
	From:  "industrial_age",
	Scene: City,
	Land: [numBiomes]Look{
		Deep:     {Glyphs: "≈~", Fg: P(HDeep, 235), Bg: P(HDeep, 42), Dense: 55},
		Shallows: {Glyphs: "~", Fg: P(HWater, 235), Bg: P(HWater, 48), Dense: 40},
		Sand:     {Glyphs: "·.", Fg: P(HSand, 215), Bg: P(HSand, 38), Dense: 25},
		Scrub:    {Glyphs: ",'.", Fg: P(HScrub, 215), Bg: P(HScrub, 24), Dense: 25},
		Grass:    {Glyphs: "\"',.", Fg: P(HGrass, 235), Bg: P(HGrass, 22), Dense: 22},
		Forest:   {Glyphs: "♣♣τ♠", Fg: P(HForest, 235), Bg: P(HForest, 26), Dense: 60},
		Hills:    {Glyphs: "^n∩", Fg: P(HHill, 225), Bg: P(HHill, 24), Dense: 35},
		Mountain: {Glyphs: "▲^", Fg: P(HPeak, 235), Bg: P(HPeak, 24), Dense: 70},
		River:    {Fg: P(HWater, 255), Bg: P(HWater, 70)},
	},
	Roads: [3]RoadLook{
		{Joins: joinsHeavy, Fg: P(HConcrete, 235), Bg: P(HAsphalt, 90)},
		{Joins: joinsLight, Fg: P(HConcrete, 215), Bg: P(HAsphalt, 74)},
		{Joins: joinsLight, Fg: P(HConcrete, 180), Bg: P(HAsphalt, 64)},
	},
	Deck:   P(HAsphalt, 120),
	Yard:   Look{Glyphs: "·", Fg: P(HConcrete, 90), Bg: P(HAsphalt, 64), Dense: 8},
	Crops:  Look{Glyphs: "≡", Fg: P(HCrop, 225), Bg: P(HCrop, 30), Dense: 55},
	Plaza:  Look{Glyphs: "\"♣", Fg: P(HLawn, 225), Bg: P(HLawn, 28), Dense: 22},
	Rubble: Look{Glyphs: "░▒%", Fg: P(HRubble, 215), Bg: P(HAsphalt, 64), Dense: 70},
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
	Filler:  []Art{art("▟▙\n▐▪▌"), art("▄▄▄\n█▪█"), art("▗▖\n▐▌")},
	FillPer: 12,
	Wall:    HConcrete,
	Ink:     map[rune]Paint{'▪': P(HGlass, 255), '▮': P(HAsphalt, 255), '▲': P(HSteel, 255), '╽': P(HPlasma, 255), '┃': P(HRubble, 255), '▀': P(HSteel, 255)},
	Windows: "▪",
	Chimney: '┃',
	Smoke:   Look{Glyphs: "∘○", Fg: P(HSmoke, 255)},
	Scatter: []Scatter{
		{On: Grass, Glyphs: "♣", Fg: P(HLawn, 240), Chance: 10},
		{On: Scrub, Glyphs: "♠", Fg: P(HScrub, 240), Chance: 8},
		{On: Hills, Glyphs: "▲", Fg: P(HPeak, 240), Chance: 8},
	},
	Agents: []AgentKind{
		{Name: "car", Glyphs: "■", Fg: P(HRed, 255), Per: 140, Max: 18, Slow: 1},
		{Name: "car", Glyphs: "■", Fg: P(HWhite, 255), Per: 170, Max: 18, Slow: 1},
		{Name: "taxi", Glyphs: "■", Fg: P(HAmber, 255), Per: 400, Max: 8, Slow: 1},
	},
	Night: Night{Tint: HNight, Dim: 185, Light: HLight, Lamps: true},
	Tones: map[byte]Paint{'s': P(HSteel, 235), 'a': P(HGold, 255), 'g': P(HGlass, 255), 'm': P(HConcrete, 235), 'f': P(HFire, 255), 'v': P(HLawn, 255)},
}
