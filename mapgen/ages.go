package mapgen

// The ground and city ages, one row each. A row names its era (eras.go) and
// sets its age apart: its roads, its material, its tallest buildings, its
// light. The space ages are the rows in styles_space.go.

var dirtRoads = [3]RoadLook{
	{Joins: "·∙", Fg: P(HEarth, 235), Bg: P(HEarth, 34)},
	{Joins: "·∙", Fg: P(HEarth, 210), Bg: P(HEarth, 30)},
	{Joins: "·.", Fg: P(HEarth, 190), Bg: P(HEarth, 26)},
}

var trackRoads = [3]RoadLook{
	{Joins: joinsLight, Fg: P(HEarth, 235), Bg: P(HEarth, 38)},
	{Joins: "·∙", Fg: P(HEarth, 215), Bg: P(HEarth, 32)},
	{Joins: "·.", Fg: P(HEarth, 190), Bg: P(HEarth, 26)},
}

var pavedRoads = [3]RoadLook{
	{Joins: joinsHeavy, Fg: P(HStone, 235), Bg: P(HStone, 40)},
	{Joins: joinsLight, Fg: P(HPath, 225), Bg: P(HEarth, 34)},
	{Joins: "·∙", Fg: P(HEarth, 200), Bg: P(HEarth, 28)},
}

var cobbledRoads = [3]RoadLook{
	{Joins: joinsLight, Fg: P(HStone, 235), Bg: P(HStone, 34)},
	{Joins: joinsLight, Fg: P(HStone, 205), Bg: P(HStone, 28)},
	{Joins: "·∙", Fg: P(HEarth, 200), Bg: P(HEarth, 28)},
}

var dockRoads = [3]RoadLook{
	{Joins: joinsHeavy, Fg: P(HEarth, 230), Bg: P(HEarth, 44)},
	{Joins: joinsLight, Fg: P(HStone, 205), Bg: P(HConcrete, 30)},
	{Joins: joinsLight, Fg: P(HEarth, 190), Bg: P(HEarth, 28)},
}

var railRoads = [3]RoadLook{
	{Joins: joinsHeavy, Fg: P(HConcrete, 230), Bg: P(HConcrete, 50)},
	{Joins: joinsHeavy, Fg: P(HStone, 205), Bg: P(HConcrete, 36)},
	{Joins: joinsLight, Fg: P(HEarth, 190), Bg: P(HEarth, 28)},
}

var highwayRoads = [3]RoadLook{
	{Joins: joinsDouble, Fg: P(HConcrete, 235), Bg: P(HAsphalt, 96)},
	{Joins: joinsHeavy, Fg: P(HConcrete, 215), Bg: P(HAsphalt, 78)},
	{Joins: joinsLight, Fg: P(HConcrete, 180), Bg: P(HAsphalt, 64)},
}

var dataRoads = [3]RoadLook{
	{Joins: joinsDouble, Fg: P(HCyan, 200), Bg: P(HAsphalt, 100)},
	{Joins: joinsHeavy, Fg: P(HConcrete, 215), Bg: P(HAsphalt, 80)},
	{Joins: joinsLight, Fg: P(HConcrete, 180), Bg: P(HAsphalt, 66)},
}

var plasmaRoads = [3]RoadLook{
	{Joins: joinsDouble, Fg: P(HPlasma, 220), Bg: P(HAsphalt, 112)},
	{Joins: joinsHeavy, Fg: P(HCyan, 200), Bg: P(HAsphalt, 98)},
	{Joins: joinsLight, Fg: P(HSteel, 170), Bg: P(HAsphalt, 86)},
}

var ageRows = []AgeRow{
	{
		Age: "primitive_age", Era: &stoneEra, Roads: dirtRoads, Wall: HTimber,
		Top:    [numFamilies][]Art{Home: {art(" /\\/\\ \n/∩\\/∩\\\n|▫||▫|")}},
		Filler: []Art{art("/\\\n||"), art(" ▲ \n(∩)")},
	},
	{
		Age: "stone_age", Era: &stoneEra, Roads: trackRoads, Wall: HStone,
		Top:    [numFamilies][]Art{Civic: {art("ΠΠΠΠΠΠ\nΠ ▲* Π\nΠΠΠΠΠΠ")}},
		Ink:    map[rune]Paint{'Π': P(HStone, 255)},
		Filler: []Art{art("/\\\n|▫"), art("Π\nΠ")},
	},
	{Age: "bronze_age", Era: &stoneEra, Wall: HEarth},
	{
		Age: "iron_age", Era: &ironEra, Wall: HStone,
		Ink:   map[rune]Paint{'◢': P(HTimber, 255), '◣': P(HTimber, 255), '▲': P(HTimber, 255)},
		Tones: map[byte]Paint{'s': P(HStone, 230)},
	},
	{
		Age: "classical_age", Era: &ironEra, Roads: pavedRoads, Wall: HStone,
		Top: [numFamilies][]Art{
			Home:  {art("◢▲▲▲▲◣\n▌▫▫▫▫▐\n▌▫∩∩▫▐\n══════")},
			Civic: {art("▄▄▄▄▄▄\n║║║║║║\n║║*║║║\n▀▀▀▀▀▀")},
		},
		Ink: map[rune]Paint{'║': P(HStone, 255), '═': P(HStone, 255)},
	},
	{
		Age: "medieval_age", Era: &ironEra, Roads: cobbledRoads, Wall: HStone,
		Top: [numFamilies][]Art{
			Home:  {art("◢▲▲▲▲◣\n▌╳▫╳▫▐\n▌▫∩▫╳▐")},
			Work:  {art("╷ ╷ ╷ \n◢▲▲▲▲◣\n▌╳▫╳▫▐\n▌▫*∩╳▐")},
			Civic: {art("  ╫   \n ▟█▙  \n▟▫*▫▙ \n██∩██ ")},
		},
		Ink: map[rune]Paint{'╫': P(HGold, 255)},
	},
	{
		Age: "renaissance_age", Era: &steelEra, Roads: cobbledRoads, Wall: HBrick,
		Top: [numFamilies][]Art{
			Civic: {art("  ◠◠  \n ▟██▙ \n▟▫*▫▙\n█∩██∩█")},
		},
	},
	{
		Age: "colonial_age", Era: &steelEra, Roads: dockRoads, Wall: HTimber,
		Top: [numFamilies][]Art{
			Work: {art("▄▄▄▄▄▄\n█▫▫▫▫█\n█▫*∩▫█\n▀▀▀▀▀▀")},
		},
		Ink: map[rune]Paint{'┃': P(HTimber, 255)},
	},
	{
		Age: "industrial_age", Era: &steelEra, Roads: railRoads, Wall: HBrick,
		Top: [numFamilies][]Art{
			Work: {art("┃ ┃ ┃ \n▟█████\n█▫▫▫▫▫\n█*▫▫▫▫")},
		},
		Night: Night{Tint: HNight, Dim: 180, Light: HLight},
	},
	{
		Age: "victorian_age", Era: &electricEra, Roads: railRoads, Wall: HBrick,
		Top: [numFamilies][]Art{
			Civic: {art("  ╽   \n ▄█▄  \n▟▫*▫▙ \n█∩██∩█")},
		},
		Ink:   map[rune]Paint{'▪': P(HGlass, 220), '╤': P(HSteel, 255)},
		Night: Night{Tint: HNight, Dim: 176, Light: HLight, Lamps: true},
	},
	{Age: "electric_age", Era: &electricEra, Wall: HConcrete},
	{
		Age: "atomic_age", Era: &electricEra, Roads: highwayRoads, Wall: HConcrete,
		Top: [numFamilies][]Art{
			Work: {art("  ◙◙  \n ▟██▙ \n ▐▪▪▌ \n▟█*█▙")},
		},
		Ink: map[rune]Paint{'◙': P(HSteel, 255)},
	},
	{Age: "modern_age", Era: &digitalEra, Wall: HConcrete},
	{
		Age: "information_age", Era: &digitalEra, Roads: dataRoads, Wall: HSteel,
		Top: [numFamilies][]Art{
			Work: {art("▤▤▤▤▤▤\n▤▤▤▤▤▤\n▤▤*▤▤▤\n▀▀▀▀▀▀")},
		},
		Ink:   map[rune]Paint{'▪': P(HCool, 255), '▤': P(HGlass, 230)},
		Night: Night{Tint: HNight, Dim: 186, Light: HCool, Lamps: true},
	},
	{
		Age: "digital_age", Era: &digitalEra, Roads: dataRoads, Wall: HSteel,
		Top: [numFamilies][]Art{
			Civic: {art("  ╽   \n ▟▪▙  \n ▐▪▌  \n▟▪▪▪▙ ")},
			Work:  {art("▤▤▤▤▤▤\n▤▤▤▤▤▤\n▤▤*▤▤▤\n▀▀▀▀▀▀")},
		},
		Ink:   map[rune]Paint{'▪': P(HCyan, 255), '▤': P(HGlass, 230)},
		Night: Night{Tint: HNight, Dim: 188, Light: HCool, Lamps: true},
	},
	{Age: "cyberpunk_age", Era: &neonEra, Pads: 2},
	{
		Age: "fusion_age", Era: &neonEra, Roads: plasmaRoads, Wall: HSteel, Pads: 5,
		Top: [numFamilies][]Art{
			Home:  {art("  ◢◣  \n ◢██◣ \n ▐▪▫▌ \n▐▪▫▪▫▌")},
			Civic: {art("  ╽   \n ◢██◣ \n ▐▪▫▌ \n ▐*▮▌ ")},
		},
		Ink:   map[rune]Paint{'▪': P(HPlasma, 255), '▫': P(HCyan, 255), '◢': P(HPlasma, 255), '◣': P(HPlasma, 255)},
		Night: Night{Tint: HNight, Dim: 190, Light: HPlasma, Lamps: true},
	},
}
