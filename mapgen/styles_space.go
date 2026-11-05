package mapgen

// The space ages, seen from straight above and pulled back era by era: the
// home planet and its orbits (the Space Age), the solar system (the
// Interstellar Age), then the star map (the three ages after it).

// orbitEra: the home planet's disc with its rings of stations.
var orbitEra = AgeStyle{
	Scale:   Orbit,
	Caravan: byFreight,
	Omen:    omen(HRed),
	Site:    siteStar,
	// the planet's land and sea: the world's own biomes, seen from afar
	Land: [numBiomes]Look{
		Deep:     {Glyphs: "≈", Fg: P(HWater, 150), Bg: P(HSea, 255), Dense: 40},
		Shallows: {Glyphs: "~", Fg: P(HCyan, 200), Bg: P(HCyan, 70), Dense: 40},
		Sand:     {Glyphs: "·", Fg: P(HSand, 200), Bg: P(HLand, 255), Dense: 40},
		Scrub:    {Glyphs: "\"'.", Fg: P(HScrub, 190), Bg: P(HLand, 255), Dense: 40},
		Grass:    {Glyphs: "\"'.", Fg: P(HGrass, 200), Bg: P(HLand, 255), Dense: 40},
		Forest:   {Glyphs: "♣♠", Fg: P(HForest, 230), Bg: P(HLand, 255), Dense: 60},
		Hills:    {Glyphs: "^n", Fg: P(HHill, 220), Bg: P(HLand, 255), Dense: 50},
		Mountain: {Glyphs: "▲^", Fg: P(HWhite, 255), Bg: P(HPeak, 200), Dense: 65},
		River:    {Glyphs: "~", Fg: P(HCyan, 200), Bg: P(HCyan, 70), Dense: 40},
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
			{art("▤═*═▤"), art("╞═◉═╡")},
			{art("  ╷  \n▤═*═▤\n  ╵  "), art(" ╭─╮ \n▤╡*╞▤")},
			{art(" ╭─╮ \n▤╣*╠▤\n ╰─╯ "), art(" ╭╦╮ \n▤╡*╞▤\n ╰╩╯ ")},
			{art(" ╭╦╮ \n▤╠*╣▤\n ╰╩╯ "), art("▤╭╦╮▤\n ╠*╣ \n▤╰╩╯▤")},
		},
		Work: {
			{art("▤─*─▤"), art("╞═▣═╡")},
			{art("▤   ▤\n╞═*═╡\n▤   ▤"), art("  ╷  \n▤═*═▤\n  ╵  ")},
			{art("▤ ╷ ▤\n╠═*═╣\n▤ ╵ ▤"), art("▤   ▤\n╠═*═╣\n▤   ▤")},
			{art("▤ ╦ ▤\n╠═*═╣\n▤ ╩ ▤"), art("▤╭─╮▤\n╠╡*╞╣\n▤╰─╯▤")},
		},
		Civic: {
			{art("▤─◈─▤"), art("──*──")},
			{art(" ╭╨╮ \n▤╡*╞▤\n  ╷  "), art("  ╷  \n▤╡*╞▤\n  ╵  ")},
			{art(" ╲╷╱ \n▤═*═▤\n ╱╵╲ "), art(" ╭╨╮ \n▤╡*╞▤\n ╰╥╯ ")},
			{art(" ╭╨╮ \n▤╣*╠▤\n ╰╥╯ "), art("╲ ╷ ╱\n▤╠*╣▤\n╱ ╵ ╲")},
		},
	},
	Filler:  []Art{art("─▫─"), art("·◈·")},
	FillPer: 45,
	Wall:    HSteel,
	Ink: map[rune]Paint{
		'▤': P(HGlass, 230), '■': P(HWhite, 255), '▫': P(HWhite, 255), '▪': P(HCyan, 255), 'Ψ': P(HRed, 255), '◠': P(HHydro, 255),
		'▓': P(HSteel, 255), '█': P(HSteel, 255), '▄': P(HSteel, 255), '▀': P(HSteel, 255),
	},
	Windows: "▫",
	Night:   Night{Tint: HNight, Dim: 70, Light: HLight},
	Tones:   map[byte]Paint{'s': P(HSteel, 255), 'a': P(HSun, 255), 'g': P(HEngine, 255), 'm': P(HSteel, 235), 'f': P(HSun, 255)},
	Sky:     Look{Glyphs: "·.∙·✧", Fg: P(HStar, 255), Bg: P(HVoid, 255), Dense: 46},
}

// systemEra: the sun, the planets on their orbits, lanes between them.
var systemEra = AgeStyle{
	Scale:   System,
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
		Home: {
			{art("╞*╡"), art("╭*╮")},
			{art("▤*▤"), art("╠*╣")},
			{art("╡*╞"), art("◈*◈")},
			{art("▣*▣"), art("▤*▤")},
		},
		Work: {
			{art("[*]"), art("╞*╡")},
			{art("▤*▤"), art("╠*╣")},
			{art("╫*╫"), art("▣*▣")},
			{art("◈*◈"), art("╠*╣")},
		},
		Civic: {
			{art("╞*╡"), art("╭*╮")},
			{art("╞*╡"), art("╠*╣")},
			{art("◈*◈"), art("▤*▤")},
			{art("▣*▣"), art("◈*◈")},
		},
	},
	Filler:  []Art{art("·▫·"), art("─◈─")},
	FillPer: 45,
	Wall:    HSteel,
	Hub:     art(" ▄█▄ \n█☼☼☼█\n ▀█▀ "),
	Bodies: []Body{
		{'●', HRock}, {'●', HAmber}, {'●', HWater}, {'●', HRed}, {'O', HAmber}, {'O', HSand}, {'●', HAir}, {'●', HPanel},
	},
	Ink:   map[rune]Paint{'█': P(HSun, 255), '▄': P(HSun, 255), '▀': P(HSun, 255), '☼': P(HFire, 255)},
	Tones: map[byte]Paint{'s': P(HSteel, 255), 'a': P(HSun, 255), 'g': P(HEngine, 255), 'm': P(HSteel, 235), 'f': P(HSun, 255)},
	Sky:   Look{Glyphs: "·.∙·", Fg: P(HStar, 255), Bg: P(HVoid, 255), Dense: 30},
}

// starsEra: the star map, systems joined by routes in the glow of the
// people's reach.
var starsEra = AgeStyle{
	Scale:   Stars,
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
	Wall:  HSteel,
	Hub:   art("☼"),
	Reach: P(HNebula, 34),
	Tones: map[byte]Paint{'s': P(HSteel, 255), 'a': P(HSun, 255), 'g': P(HEngine, 255), 'm': P(HSteel, 235), 'f': P(HSun, 255)},
	Sky:   Look{Glyphs: "·.∙·✧+", Fg: P(HStar, 255), Bg: P(HVoid, 255), Dense: 62},
}

var spaceRows = []AgeRow{
	{Age: "space_age", Era: &orbitEra},
	{Age: "interstellar_age", Era: &systemEra},
	{
		Age: "galactic_age", Era: &starsEra, Reach: P(HNebula, 34),
	},
	{
		Age: "quantum_age", Era: &starsEra, Reach: P(HNeon, 22), Ink: map[rune]Paint{'☼': P(HNeon, 255)},
	},
	{
		Age: "transcendent_age", Era: &starsEra, Reach: P(HPlasma, 20), Ink: map[rune]Paint{'☼': P(HPlasma, 255)},
	},
}
