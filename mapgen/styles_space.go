package mapgen

// The space ages, seen from straight above and pulled back era by era: the
// home planet and its orbits (the Space Age), the solar system (the
// Interstellar Age), then the star map (the three ages after it).

// orbitEra: the home planet's disc with its rings of stations.
var orbitEra = AgeStyle{
	Scale:   Orbit,
	Visitor: byShip,
	Caravan: byFreight,
	Omen:    omen(HRed),
	Site:    siteStar,
	// the planet's land and sea: the world's own biomes, seen from afar
	Land: [numBiomes]Look{
		Deep:     {Glyphs: "≈~", Fg: P(HWater, 140), Bg: P(HSea, 255), Dense: 35},
		Shallows: {Glyphs: "~", Fg: P(HWater, 190), Bg: P(HSea, 235), Dense: 30},
		Sand:     {Glyphs: "·", Fg: P(HSand, 200), Bg: P(HLand, 255), Dense: 40},
		Scrub:    {Glyphs: "\"'.", Fg: P(HScrub, 190), Bg: P(HLand, 255), Dense: 40},
		Grass:    {Glyphs: "\"'.", Fg: P(HGrass, 200), Bg: P(HLand, 255), Dense: 40},
		Forest:   {Glyphs: "♠♣", Fg: P(HForest, 210), Bg: P(HLand, 255), Dense: 55},
		Hills:    {Glyphs: "^n", Fg: P(HHill, 220), Bg: P(HLand, 255), Dense: 50},
		Mountain: {Glyphs: "▲^", Fg: P(HPeak, 230), Bg: P(HLand, 230), Dense: 60},
		River:    {Glyphs: "~", Fg: P(HWater, 190), Bg: P(HSea, 235), Dense: 30},
	},
	Cloud: Look{Glyphs: "░▒", Fg: P(HCloud, 255), Bg: P(HCloud, 200), Dense: 70},
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
			{art("[▓]")},
			{art("╒▄╕\n╘▓╛")},
			{art("╒═▄═╕\n║▓▓▓║\n╘═══╛")},
			{art(" ▄█▄ \n╞███╡\n ▀█▀ ")},
		},
		Work: {
			{art("▤*▤")},
			{art("▤▤*▤▤")},
			{art("▤▤▤▤▤\n╞═*═╡")},
			{art("▤▤▤▤▤\n╞■*■╡\n▤▤▤▤▤")},
		},
		Civic: {
			{art(" ◠ \n[*]")},
			{art(" ◠◠ \n╞*▫╡")},
			{art("  Ψ  \n ◠◠◠ \n╞▫*▫╡")},
			{art("  Ψ  \n ◠█◠ \n╞▫*▫╡")},
		},
	},
	Filler:  []Art{art("[▫]"), art("◘▫◘")},
	FillPer: 45,
	Wall:    HSteel,
	Ink: map[rune]Paint{
		'▤': P(HPanel, 255), '■': P(HWhite, 255), '▫': P(HWhite, 255), 'Ψ': P(HRed, 255), '◠': P(HHydro, 255),
		'▓': P(HSteel, 255), '█': P(HSteel, 255), '▄': P(HSteel, 255), '▀': P(HSteel, 255),
	},
	Windows: "▫",
	Agents: []AgentKind{
		{Name: "shuttle", Glyphs: "►◄▼▲", Fg: P(HEngine, 255), Per: 200, Max: 14, Slow: 1},
		{Name: "freighter", Glyphs: "►◄▼▲", Fg: P(HAmber, 255), Per: 450, Max: 8, Slow: 2},
	},
	Night: Night{Tint: HNight, Dim: 70, Light: HLight},
	Tones: map[byte]Paint{'s': P(HSteel, 255), 'a': P(HSun, 255), 'g': P(HEngine, 255), 'm': P(HSteel, 235), 'f': P(HSun, 255)},
	Sky:   Look{Glyphs: "·.∙·✧", Fg: P(HStar, 255), Bg: P(HVoid, 255), Dense: 46},
}

// systemEra: the sun, the planets on their orbits, lanes between them.
var systemEra = AgeStyle{
	Scale:   System,
	Visitor: byShip,
	Caravan: byFreight,
	Omen:    omen(HRed),
	Site:    siteStar,
	Roads: [3]RoadLook{
		{Joins: joinsRounded, Fg: P(HEngine, 235), Bg: P(HVoid, 255)},
		{Joins: joinsLight, Fg: P(HOrbit, 255), Bg: P(HVoid, 255)},
		{Joins: joinsLight, Fg: P(HOrbit, 200), Bg: P(HVoid, 255)},
	},
	Deck:   P(HSteel, 255),
	Yard:   Look{Bg: P(HVoid, 255)},
	Plaza:  Look{Bg: P(HVoid, 255)},
	Rubble: Look{Glyphs: "·∙", Fg: P(HRock, 255), Bg: P(HVoid, 255), Dense: 60},
	Shapes: [numFamilies]Tiers{
		Home:  {{art("*")}, {art("*")}, {art("(*)")}, {art("[*]")}},
		Work:  {{art("*")}, {art("*")}, {art("(*)")}, {art("[*]")}},
		Civic: {{art("*")}, {art("*")}, {art("(*)")}, {art("[*]")}},
	},
	Wall: HSteel,
	Hub:  art(" ▄█▄ \n█☼☼☼█\n ▀█▀ "),
	Bodies: []Body{
		{'●', HRock}, {'●', HAmber}, {'●', HWater}, {'●', HRed}, {'O', HAmber}, {'O', HSand}, {'●', HAir}, {'●', HPanel},
	},
	Ink: map[rune]Paint{'█': P(HSun, 255), '▄': P(HSun, 255), '▀': P(HSun, 255), '☼': P(HFire, 255)},
	Agents: []AgentKind{
		{Name: "ship", Glyphs: "►◄▼▲", Fg: P(HEngine, 255), Per: 300, Max: 20, Slow: 2},
		{Name: "freighter", Glyphs: "►◄▼▲", Fg: P(HAmber, 255), Per: 600, Max: 10, Slow: 3},
	},
	Tones: map[byte]Paint{'s': P(HSteel, 255), 'a': P(HSun, 255), 'g': P(HEngine, 255), 'm': P(HSteel, 235), 'f': P(HSun, 255)},
	Sky:   Look{Glyphs: "·.∙·", Fg: P(HStar, 255), Bg: P(HVoid, 255), Dense: 30},
}

// starsEra: the star map, systems joined by routes in the glow of the
// people's reach.
var starsEra = AgeStyle{
	Scale:   Stars,
	Visitor: byShip,
	Caravan: byFleet,
	Omen:    omen(HNeon),
	Site:    siteStar,
	Roads: [3]RoadLook{
		{Joins: "·∙", Fg: P(HEngine, 215), Bg: P(HNebula, 30)},
		{Joins: "·∙", Fg: P(HEngine, 190), Bg: P(HNebula, 30)},
		{Joins: "·∙", Fg: P(HEngine, 170), Bg: P(HNebula, 30)},
	},
	Deck:   P(HSteel, 255),
	Yard:   Look{Bg: P(HVoid, 255)},
	Plaza:  Look{Bg: P(HVoid, 255)},
	Rubble: Look{Glyphs: "·∙", Fg: P(HRock, 255), Bg: P(HVoid, 255), Dense: 60},
	Shapes: [numFamilies]Tiers{
		Home:  {{art("*")}, {art("*")}, {art("*")}, {art("*")}},
		Work:  {{art("*")}, {art("*")}, {art("*")}, {art("*")}},
		Civic: {{art("*")}, {art("*")}, {art("*")}, {art("*")}},
	},
	Wall: HSteel,
	Hub:  art("☼"),
	Agents: []AgentKind{
		{Name: "fleet", Glyphs: "◆", Fg: P(HEngine, 255), Per: 400, Max: 16, Slow: 3},
	},
	Reach: P(HNebula, 34),
	Tones: map[byte]Paint{'s': P(HSteel, 255), 'a': P(HSun, 255), 'g': P(HEngine, 255), 'm': P(HSteel, 235), 'f': P(HSun, 255)},
	Sky:   Look{Glyphs: "·.∙·✧+", Fg: P(HStar, 255), Bg: P(HVoid, 255), Dense: 62},
}

var spaceRows = []AgeRow{
	{Age: "space_age", Era: &orbitEra},
	{Age: "interstellar_age", Era: &systemEra},
	{Age: "galactic_age", Era: &starsEra, Reach: P(HNebula, 34)},
	{Age: "quantum_age", Era: &starsEra, Reach: P(HNeon, 22), Ink: map[rune]Paint{'☼': P(HNeon, 255)}},
	{Age: "transcendent_age", Era: &starsEra, Reach: P(HPlasma, 20), Ink: map[rune]Paint{'☼': P(HPlasma, 255)}},
}
