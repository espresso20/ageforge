package mapgen

// The six eras of the ground and the city. An era is the look its ages
// share: the land, the art of the buildings, the colours. Each age is a row
// (ages.go) that sets it apart from its neighbours.

// natural is the open land, the same under every era; the city clears its
// own forest into lawns (frame.groundAt).
var natural = [numBiomes]Look{
	Deep:     {Glyphs: "≈~", Fg: P(HDeep, 235), Bg: P(HDeep, 42), Dense: 55},
	Shallows: {Glyphs: "~", Fg: P(HWater, 235), Bg: P(HWater, 48), Dense: 40},
	Sand:     {Glyphs: "·.", Fg: P(HSand, 225), Bg: P(HSand, 45), Dense: 30},
	Scrub:    {Glyphs: ",'.", Fg: P(HScrub, 225), Bg: P(HScrub, 26), Dense: 35},
	Grass:    {Glyphs: "\"',.", Fg: P(HGrass, 235), Bg: P(HGrass, 22), Dense: 30},
	Forest:   {Glyphs: "♣♣τ♠", Fg: P(HForest, 235), Bg: P(HForest, 26), Dense: 70},
	Hills:    {Glyphs: "^n∩", Fg: P(HHill, 235), Bg: P(HHill, 26), Dense: 40},
	Mountain: {Glyphs: "▲^", Fg: P(HPeak, 235), Bg: P(HPeak, 26), Dense: 70},
	River:    {Fg: P(HWater, 255), Bg: P(HWater, 70)},
}

// redPlanet is the red frontier: basalt and iron dust, with green oases under
// glass.
var redPlanet = [numBiomes]Look{
	Deep:     {Glyphs: "▒░", Fg: P(HRock, 200), Bg: P(HRock, 40), Dense: 50},
	Shallows: {Glyphs: "·∙", Fg: P(HAmber, 160), Bg: P(HRock, 48), Dense: 35},
	Sand:     {Glyphs: "~·", Fg: P(HAmber, 225), Bg: P(HRed, 34), Dense: 35},
	Scrub:    {Glyphs: ",'.", Fg: P(HRed, 210), Bg: P(HRock, 30), Dense: 35},
	Grass:    {Glyphs: "\"',.", Fg: P(HLawn, 215), Bg: P(HAmber, 24), Dense: 28},
	Forest:   {Glyphs: "♣τ", Fg: P(HForest, 235), Bg: P(HForest, 26), Dense: 70},
	Hills:    {Glyphs: "^∩", Fg: P(HRock, 235), Bg: P(HRock, 28), Dense: 45},
	Mountain: {Glyphs: "▲^", Fg: P(HRock, 245), Bg: P(HRock, 20), Dense: 70},
	River:    {Fg: P(HCyan, 210), Bg: P(HRock, 60)},
}

// alienPlanet is an exoplanet: glowing seas, crystal sand, flora that shines.
var alienPlanet = [numBiomes]Look{
	Deep:     {Glyphs: "≈~", Fg: P(HNeon, 210), Bg: P(HWater, 52), Dense: 55},
	Shallows: {Glyphs: "~", Fg: P(HWater, 235), Bg: P(HNeon, 32), Dense: 40},
	Sand:     {Glyphs: "·✧.", Fg: P(HWhite, 230), Bg: P(HSand, 40), Dense: 30},
	Scrub:    {Glyphs: ",'.", Fg: P(HPlasma, 215), Bg: P(HPlasma, 22), Dense: 35},
	Grass:    {Glyphs: "\"',.", Fg: P(HNeon, 205), Bg: P(HPlasma, 20), Dense: 30},
	Forest:   {Glyphs: "✦✧♠♣", Fg: P(HNeon, 235), Bg: P(HPlasma, 26), Dense: 70},
	Hills:    {Glyphs: "^▲", Fg: P(HSteel, 230), Bg: P(HSteel, 24), Dense: 45},
	Mountain: {Glyphs: "▲", Fg: P(HWhite, 245), Bg: P(HSteel, 30), Dense: 70},
	River:    {Fg: P(HCyan, 255), Bg: P(HNeon, 70)},
}

// transfiguredNatural is the world after it has been remade: seas that glow,
// groves of crystal and ground lit by stars.
var transfiguredNatural = [numBiomes]Look{
	Deep:     {Glyphs: "≈~", Fg: P(HPlasma, 215), Bg: P(HWater, 52), Dense: 55},
	Shallows: {Glyphs: "~", Fg: P(HPlasma, 235), Bg: P(HWater, 40), Dense: 40},
	Sand:     {Glyphs: "·✧.", Fg: P(HGold, 225), Bg: P(HSand, 36), Dense: 30},
	Scrub:    {Glyphs: ",'.", Fg: P(HGold, 205), Bg: P(HGold, 18), Dense: 35},
	Grass:    {Glyphs: "\"'.✧", Fg: P(HGold, 225), Bg: P(HWhite, 20), Dense: 30},
	Forest:   {Glyphs: "✦✧♣", Fg: P(HGold, 240), Bg: P(HNeon, 26), Dense: 70},
	Hills:    {Glyphs: "^n∩", Fg: P(HWhite, 230), Bg: P(HGold, 24), Dense: 45},
	Mountain: {Glyphs: "▲^", Fg: P(HWhite, 245), Bg: P(HGold, 30), Dense: 70},
	River:    {Glyphs: "≈~", Fg: P(HPlasma, 255), Bg: P(HNeon, 70)},
}

// The stone era: huts of thatch and mud, dirt tracks, hearths.
var stoneEra = AgeStyle{
	Scale:   Ground,
	Visitor: onFoot,
	Caravan: byCart,
	Omen:    omen(HFire),
	Site:    siteFlag,
	Land:    natural,
	Roads: [3]RoadLook{
		{Joins: joinsLight, Fg: P(HPath, 235), Bg: P(HEarth, 40)},
		{Joins: "·∙", Fg: P(HEarth, 235), Bg: P(HEarth, 34)},
		{Joins: "·.", Fg: P(HEarth, 200), Bg: P(HEarth, 28)},
	},
	Deck:   P(HTimber, 255),
	Yard:   Look{Glyphs: "·", Fg: P(HEarth, 130), Bg: P(HEarth, 36), Dense: 10},
	Crops:  Look{Glyphs: "≡", Fg: P(HCrop, 235), Bg: P(HCrop, 34), Dense: 60},
	Plaza:  Look{Glyphs: "\"", Fg: P(HLawn, 190), Bg: P(HLawn, 26), Dense: 25},
	Rubble: Look{Glyphs: "░▒%", Fg: P(HRubble, 210), Bg: P(HEarth, 36), Dense: 70},
	Shapes: [numFamilies]Tiers{
		Home: {
			{art("/^\\\n|∩|"), art(" ▲ \n(∩)")},
			{art(" /^\\ \n/___\\\n|▫∩▫|"), art("/‾‾‾\\\n|∩▫▫|")},
			{art("/^\\/^\\\n|▫||▫|\n|∩||∩|"), art(" /^^\\ \n/____\\\n|▫▫∩▫|")},
			{art("/^^^^\\\n|▫▫▫▫|\n|▫∩∩▫|")},
		},
		Work: {
			{art(" ╷ \n[*]")},
			{art(" ╷ \n┌─┐\n│*│")},
			{art("  ╷  \n┌───┐\n│▫*▫│")},
			{art(" ╷  ╷ \n┌────┐\n│▫*∩▫│")},
		},
		Civic: {
			{art(" ▲ \n▐*▌")},
			{art(" ▲ \n/_\\\n│*│")},
			{art("  ▲  \n /_\\ \n/___\\\n│▫*▫│")},
			{art("  ▲▲  \n /__\\ \n/____\\\n│▫*∩▫│")},
		},
	},
	Filler:  []Art{art("/\\\n|▫"), art(" ▲ \n(∩)")},
	FillPer: 25,
	Wall:    HStone,
	Ink: map[rune]Paint{
		'/': P(HThatch, 255), '\\': P(HThatch, 255), '^': P(HThatch, 255), '_': P(HThatch, 255),
		'▲': P(HThatch, 255), '∩': P(HTimber, 255), '╷': P(HSmoke, 255),
	},
	Windows: "▫∩",
	Chimney: '╷',
	Smoke:   Look{Glyphs: "°∘", Fg: P(HSmoke, 255)},
	Scatter: []Scatter{
		{On: Grass, Glyphs: "♣τ", Fg: P(HForest, 240), Chance: 14},
		{On: Scrub, Glyphs: "♠", Fg: P(HScrub, 240), Chance: 10},
		{On: Hills, Glyphs: "▲", Fg: P(HPeak, 240), Chance: 10},
	},
	Agents: []AgentKind{
		{Name: "walker", Glyphs: "☺", Fg: P(HLife, 255), Per: 18, Max: 24, Slow: 3},
		{Name: "ox_cart", Glyphs: "o=", Fg: P(HThatch, 255), Per: 60, Max: 6, Slow: 5},
	},
	Night: Night{Tint: HNight, Dim: 165, Light: HLight},
	Tones: map[byte]Paint{'s': P(HStone, 230), 'a': P(HThatch, 255)},
}

// The iron era: stone and timber, tile roofs, cobbles, towers and walls.
var ironEra = AgeStyle{
	Scale:   Ground,
	Visitor: onFoot,
	Caravan: byCart,
	Omen:    omen(HRed),
	Site:    siteFlag,
	Land:    natural,
	Roads: [3]RoadLook{
		{Joins: joinsLight, Fg: P(HPath, 235), Bg: P(HEarth, 40)},
		{Joins: joinsLight, Fg: P(HEarth, 220), Bg: P(HEarth, 34)},
		{Joins: "·∙", Fg: P(HEarth, 200), Bg: P(HEarth, 28)},
	},
	Deck:   P(HTimber, 255),
	Yard:   Look{Glyphs: "·", Fg: P(HStone, 100), Bg: P(HEarth, 34), Dense: 10},
	Crops:  Look{Glyphs: "≡", Fg: P(HCrop, 235), Bg: P(HCrop, 34), Dense: 60},
	Plaza:  Look{Glyphs: "\"", Fg: P(HLawn, 190), Bg: P(HLawn, 26), Dense: 25},
	Rubble: Look{Glyphs: "░▒%", Fg: P(HRubble, 210), Bg: P(HEarth, 36), Dense: 70},
	Shapes: [numFamilies]Tiers{
		Home: {
			{art("◢▲◣\n▌▫▐"), art(" ▲ \n▐∩▌")},
			{art(" ◢▲◣ \n◢███◣\n▌▫▫▫▐"), art("◢▲▲◣\n▌▫▫▐")},
			{art("◢▲◣◢▲◣\n▌▫▐▌▫▐\n▌∩▐▌∩▐"), art(" ◢▲▲◣ \n◢████◣\n▌▫▫▫▫▐")},
			{art(" ◢▲▲◣ \n◢████◣\n▌▫▫▫▫▐\n▌▫∩▫▫▐")},
		},
		Work: {
			{art(" ╷ \n▐*▌")},
			{art(" ╷ \n◢▲◣\n▌*▐")},
			{art("╷  ╷\n◢▲▲◣\n▌▫*▐")},
			{art("╷ ╷ ╷ \n◢▲▲▲▲◣\n▌▫▫▫▫▐\n▌▫*∩▫▐")},
		},
		Civic: {
			{art("▄▀▄\n▌*▐")},
			{art("▄▀▄▀▄\n▌▫*▫▐")},
			{art("▄▀▄▀▄▀\n▌▫▫▫▫▐\n▌▫*∩▫▐")},
			{art("▄▀▄▀▄▀\n█▫▫▫▫█\n█▫*▫▫█\n██∩███")},
		},
	},
	Filler:  []Art{art("◢◣\n▌▐"), art(" ▲ \n▐▌"), art("◢▲◣\n▌∩▐")},
	FillPer: 24,
	Wall:    HStone,
	Ink: map[rune]Paint{
		'◢': P(HTile, 255), '◣': P(HTile, 255), '▲': P(HTile, 255), '∩': P(HTimber, 255),
		'╷': P(HSmoke, 255), '▫': P(HGlass, 200), '╳': P(HTimber, 255), '▄': P(HSteel, 255), '▀': P(HSteel, 255),
	},
	Windows: "▫∩",
	Chimney: '╷',
	Smoke:   Look{Glyphs: "°∘", Fg: P(HSmoke, 255)},
	Scatter: []Scatter{
		{On: Grass, Glyphs: "♣τ", Fg: P(HForest, 240), Chance: 11},
		{On: Scrub, Glyphs: "♠", Fg: P(HScrub, 240), Chance: 9},
		{On: Hills, Glyphs: "▲", Fg: P(HPeak, 240), Chance: 10},
	},
	Agents: []AgentKind{
		{Name: "walker", Glyphs: "☺", Fg: P(HLife, 255), Per: 20, Max: 26, Slow: 3},
		{Name: "ox_cart", Glyphs: "o=", Fg: P(HTimber, 255), Per: 60, Max: 8, Slow: 4},
	},
	Night: Night{Tint: HNight, Dim: 170, Light: HLight},
	Tones: map[byte]Paint{'s': P(HStone, 240), 'a': P(HGold, 255)},
}

// The steel era: brick and slate, domes, docks, the first great chimneys.
var steelEra = AgeStyle{
	Scale:   City,
	Visitor: onFoot,
	Caravan: byCart,
	Omen:    omen(HRed),
	Site:    siteFlag,
	Land:    natural,
	Roads: [3]RoadLook{
		{Joins: joinsHeavy, Fg: P(HStone, 220), Bg: P(HConcrete, 36)},
		{Joins: joinsLight, Fg: P(HStone, 200), Bg: P(HConcrete, 30)},
		{Joins: joinsLight, Fg: P(HEarth, 190), Bg: P(HEarth, 28)},
	},
	Deck:   P(HTimber, 255),
	Yard:   Look{Glyphs: "·", Fg: P(HConcrete, 90), Bg: P(HConcrete, 36), Dense: 8},
	Crops:  Look{Glyphs: "≡", Fg: P(HCrop, 225), Bg: P(HCrop, 30), Dense: 55},
	Plaza:  Look{Glyphs: "\"♣", Fg: P(HLawn, 215), Bg: P(HLawn, 26), Dense: 20},
	Rubble: Look{Glyphs: "░▒%", Fg: P(HRubble, 215), Bg: P(HConcrete, 36), Dense: 70},
	Shapes: [numFamilies]Tiers{
		Home: {
			{art("▄▄▄\n█▫█"), art("◢▄◣\n█▫█")},
			{art("▄▄▄▄▄\n█▫█▫█\n█▫█∩█")},
			{art("▄▄▄▄▄▄\n█▫█▫█▫\n█▫█▫█▫\n█∩█∩█∩")},
			{art("▗▄▄▄▄▖\n█▫▫▫▫█\n█▫▫▫▫█\n█▫∩∩▫█")},
		},
		Work: {
			{art("┃  \n█*█")},
			{art("┃ ┃\n▟█▙\n█*█")},
			{art("┃  ┃ \n▟▀▀▀▙\n█▫*▫█")},
			{art("┃ ┃ ┃ \n▟█████\n█▫▫▫▫▫\n█▫*∩▫▫")},
		},
		Civic: {
			{art(" ◠ \n█*█")},
			{art(" ◠ \n▟█▙\n█*█")},
			{art("  ◠  \n ▟█▙ \n▟▫*▫▙")},
			{art("  ◠   \n ▄█▄  \n▟▫▫▫▙ \n█▫*∩█ ")},
		},
	},
	Filler:  []Art{art("▄▄▄\n█▫█"), art("▗▄▖\n▐▫▌"), art("▄▄\n█▫")},
	FillPer: 20,
	Wall:    HBrick,
	Ink: map[rune]Paint{
		'▄': P(HSteel, 255), '▀': P(HSteel, 255), '◢': P(HSteel, 255), '◣': P(HSteel, 255),
		'▫': P(HGlass, 210), '┃': P(HBrick, 255), '◠': P(HSteel, 255), '∩': P(HTimber, 255), '▗': P(HSteel, 255), '▖': P(HSteel, 255),
	},
	Windows: "▫∩",
	Chimney: '┃',
	Smoke:   Look{Glyphs: "∘○", Fg: P(HSmoke, 255)},
	Scatter: []Scatter{
		{On: Grass, Glyphs: "♣", Fg: P(HLawn, 240), Chance: 9},
		{On: Scrub, Glyphs: "♠", Fg: P(HScrub, 240), Chance: 8},
		{On: Hills, Glyphs: "▲", Fg: P(HPeak, 240), Chance: 8},
	},
	Agents: []AgentKind{
		{Name: "walker", Glyphs: "☺", Fg: P(HLife, 255), Per: 40, Max: 24, Slow: 3},
		{Name: "wagon", Glyphs: "o=", Fg: P(HBrick, 255), Per: 80, Max: 10, Slow: 3},
	},
	Night: Night{Tint: HNight, Dim: 175, Light: HLight},
	Tones: map[byte]Paint{'s': P(HBrick, 235), 'm': P(HSteel, 235)},
}

// The electric era: concrete, wire and light; lamps line the streets.
var electricEra = AgeStyle{
	Scale:   City,
	Visitor: onFoot,
	Caravan: byTruck,
	Omen:    omen(HAmber),
	Site:    siteFlag,
	Land:    natural,
	Roads: [3]RoadLook{
		{Joins: joinsHeavy, Fg: P(HConcrete, 235), Bg: P(HAsphalt, 80)},
		{Joins: joinsLight, Fg: P(HConcrete, 215), Bg: P(HAsphalt, 68)},
		{Joins: joinsLight, Fg: P(HConcrete, 180), Bg: P(HAsphalt, 60)},
	},
	Deck:   P(HAsphalt, 120),
	Yard:   Look{Glyphs: "·", Fg: P(HConcrete, 90), Bg: P(HAsphalt, 60), Dense: 8},
	Crops:  Look{Glyphs: "≡", Fg: P(HCrop, 225), Bg: P(HCrop, 30), Dense: 55},
	Plaza:  Look{Glyphs: "\"♣", Fg: P(HLawn, 220), Bg: P(HLawn, 28), Dense: 22},
	Rubble: Look{Glyphs: "░▒%", Fg: P(HRubble, 215), Bg: P(HAsphalt, 60), Dense: 70},
	Shapes: [numFamilies]Tiers{
		Home: {
			{art("▟▙\n▐▪▌")},
			{art("▄▄▄▄▄\n▐▪▪▪▌\n▐▪∩▪▌")},
			{art("▄▄▄▄▄▄\n█▪▪▪▪█\n█▪▪▪▪█\n█▪∩∩▪█")},
			{art("▄▄▄▄▄▄\n█▪█▪█▪\n█▪█▪█▪\n█▪█▪█▪")},
		},
		Work: {
			{art("╤ \n▐*▌")},
			{art("╤ ╤\n▐█▌\n▐*▌")},
			{art("╤  ╤ \n▟███▙\n█▪*▪█")},
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
	FillPer: 14,
	Wall:    HConcrete,
	Ink: map[rune]Paint{
		'▪': P(HGlass, 255), '▮': P(HAsphalt, 255), '▲': P(HSteel, 255), '╽': P(HPlasma, 255),
		'┃': P(HRubble, 255), '╤': P(HGold, 255), '∩': P(HAsphalt, 255),
	},
	Windows: "▪",
	Chimney: '┃',
	Smoke:   Look{Glyphs: "∘○", Fg: P(HSmoke, 255)},
	Scatter: []Scatter{
		{On: Grass, Glyphs: "♣", Fg: P(HLawn, 240), Chance: 10},
		{On: Scrub, Glyphs: "♠", Fg: P(HScrub, 240), Chance: 8},
		{On: Hills, Glyphs: "▲", Fg: P(HPeak, 240), Chance: 8},
	},
	Agents: []AgentKind{
		{Name: "walker", Glyphs: "☺", Fg: P(HLife, 255), Per: 60, Max: 20, Slow: 3},
		{Name: "early_car", Glyphs: "■", Fg: P(HRed, 255), Per: 180, Max: 14, Slow: 1},
		{Name: "early_car", Glyphs: "■", Fg: P(HWhite, 255), Per: 220, Max: 12, Slow: 1},
	},
	Night: Night{Tint: HNight, Dim: 180, Light: HLight, Lamps: true},
	Tones: map[byte]Paint{'s': P(HConcrete, 235), 'a': P(HGold, 255), 'm': P(HSteel, 235)},
}

// The digital era: glass and asphalt, screens, the last parks.
var digitalEra = AgeStyle{
	Scale:   City,
	Visitor: onFoot,
	Caravan: byTruck,
	Omen:    omen(HCool),
	Site:    siteFlag,
	Land:    natural,
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
	Ink: map[rune]Paint{
		'▪': P(HGlass, 255), '▮': P(HAsphalt, 255), '▲': P(HSteel, 255), '╽': P(HPlasma, 255),
		'┃': P(HRubble, 255), '▀': P(HSteel, 255),
	},
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
		{Name: "car", Glyphs: "■", Fg: P(HAmber, 255), Per: 400, Max: 8, Slow: 1},
	},
	Night: Night{Tint: HNight, Dim: 185, Light: HLight, Lamps: true},
	Tones: map[byte]Paint{'s': P(HSteel, 235), 'a': P(HGold, 255), 'm': P(HConcrete, 235)},
}

// The neon era: dark towers hung with light, rain on black streets, and at
// the edge of the city a spaceport.
var neonEra = AgeStyle{
	Scale:   City,
	Visitor: onFoot,
	Caravan: byHover,
	Omen:    omen(HNeon),
	Site:    siteFlag,
	Land:    natural,
	Roads: [3]RoadLook{
		{Joins: joinsHeavy, Fg: P(HCyan, 200), Bg: P(HAsphalt, 110)},
		{Joins: joinsLight, Fg: P(HNeon, 190), Bg: P(HAsphalt, 96)},
		{Joins: joinsLight, Fg: P(HConcrete, 160), Bg: P(HAsphalt, 86)},
	},
	Deck:   P(HAsphalt, 140),
	Yard:   Look{Glyphs: "·", Fg: P(HCyan, 90), Bg: P(HAsphalt, 86), Dense: 10},
	Crops:  Look{Glyphs: "≡", Fg: P(HLawn, 225), Bg: P(HLawn, 30), Dense: 55},
	Plaza:  Look{Glyphs: "♣", Fg: P(HLawn, 225), Bg: P(HLawn, 34), Dense: 18},
	Rubble: Look{Glyphs: "░▒%", Fg: P(HRubble, 215), Bg: P(HAsphalt, 86), Dense: 70},
	Shapes: [numFamilies]Tiers{
		Home: {
			{art("▟▙\n▐▪▌")},
			{art("▄▄▄▄▄\n▐▪▫▪▌\n▐▪▪▫▌")},
			{art("▄▄▄▄▄▄\n█▪▫▪▫█\n█▫▪▫▪█\n█▪▫▪▫█")},
			{art("  ▄▄  \n ▐▪▪▌ \n ▐▫▪▌ \n ▐▪▫▌ ")},
		},
		Work: {
			{art("╤ \n▐*▌")},
			{art("╤ ╤\n▐█▌\n▐*▌")},
			{art("╤  ╤ \n▟███▙\n█▪*▫█")},
			{art("╤ ╤ ╤ \n▟█████\n█▪▫▪▫▪\n█▫*▪▫▪")},
		},
		Civic: {
			{art(" ▲ \n▐*▌")},
			{art(" ▲▲ \n▟██▙\n█▪*█")},
			{art("  ▲   \n ▟██▙ \n▟████▙\n█▫*▪▫█")},
			{art("  ╽   \n ▄██▄ \n ▐▪▫▌ \n ▐*▮▌ ")},
		},
	},
	Pad:     art("  ▲   \n ╱█╲  \n ╟█╢  \n╞╬╧╬╡ \n▄▄▄▄▄▄"),
	Filler:  []Art{art("▟▙\n▐▪▌"), art("▄▄▄\n█▫█"), art("▗▖\n▐▌")},
	FillPer: 10,
	Wall:    HAsphalt,
	Ink: map[rune]Paint{
		'▪': P(HNeon, 255), '▫': P(HCyan, 255), '▮': P(HAsphalt, 255), '▲': P(HCyan, 255), '╽': P(HPlasma, 255),
		'╤': P(HCyan, 255), '╱': P(HSteel, 255), '╲': P(HSteel, 255), '╟': P(HSteel, 255), '╢': P(HSteel, 255),
		'╞': P(HSteel, 255), '╡': P(HSteel, 255), '╬': P(HGold, 255), '╧': P(HSteel, 255), '▄': P(HSteel, 255),
	},
	Windows: "▪▫",
	Scatter: []Scatter{
		{On: Grass, Glyphs: "♣", Fg: P(HLawn, 235), Chance: 8},
		{On: Hills, Glyphs: "▲", Fg: P(HPeak, 235), Chance: 8},
	},
	Agents: []AgentKind{
		{Name: "hovercar", Glyphs: "◆", Fg: P(HNeon, 255), Per: 100, Max: 24, Slow: 1},
		{Name: "hovercar", Glyphs: "◆", Fg: P(HCyan, 255), Per: 130, Max: 24, Slow: 1},
		{Name: "walker", Glyphs: "☺", Fg: P(HLife, 255), Per: 200, Max: 12, Slow: 3},
	},
	Night: Night{Tint: HNight, Dim: 190, Light: HNeon, Lamps: true},
	Tones: map[byte]Paint{'s': P(HConcrete, 235), 'a': P(HNeon, 255), 'm': P(HSteel, 235)},
}

// The cosmic era: the launchport city of the first space age, arcologies and
// solar arrays under the rockets.
var cosmicEra = AgeStyle{
	Scale:   City,
	Visitor: byShip,
	Caravan: byHover,
	Omen:    omen(HCyan),
	Site:    siteStar,
	Land:    natural,
	Roads: [3]RoadLook{
		{Joins: joinsHeavy, Fg: P(HCyan, 210), Bg: P(HAsphalt, 112)},
		{Joins: joinsLight, Fg: P(HPlasma, 195), Bg: P(HAsphalt, 98)},
		{Joins: joinsLight, Fg: P(HConcrete, 165), Bg: P(HAsphalt, 86)},
	},
	Deck:   P(HAsphalt, 140),
	Yard:   Look{Glyphs: "·", Fg: P(HCyan, 90), Bg: P(HAsphalt, 86), Dense: 10},
	Crops:  Look{Glyphs: "≡", Fg: P(HLawn, 225), Bg: P(HLawn, 30), Dense: 55},
	Plaza:  Look{Glyphs: "♣", Fg: P(HLawn, 225), Bg: P(HLawn, 34), Dense: 18},
	Rubble: Look{Glyphs: "░▒%", Fg: P(HRubble, 215), Bg: P(HAsphalt, 86), Dense: 70},
	Shapes: [numFamilies]Tiers{
		Home: {
			{art("◢◣\n▐▪▌")},
			{art(" ▄▄▄▄ \n▐▪▫▪▌\n▐▪▪▫▌")},
			{art("  ◢◣  \n ◢██◣ \n▐▪▫▪▫▌")},
			{art("  ◢◣  \n ◢██◣ \n ▐▪▫▌ \n▐▪▫▪▫▌")},
		},
		Work: {
			{art("╽ \n▐*▌")},
			{art("▤▤▤▤\n▤▤*▤"), art("╽ ╽\n▐█▌\n▐*▌")},
			{art("╽  ╽ \n╞╬╬╬╡\n▤▤*▤▤"), art("▤▤▤▤▤\n▤▤*▤▤\n▤▤▤▤▤")},
			{art("╽ ╽ ╽ \n▤▤▤▤▤▤\n╞═╬╬═╡\n▤▤*▤▤▤")},
		},
		Civic: {
			{art(" ▲ \n▐*▌")},
			{art(" ▲▲ \n▟██▙\n█▪*█")},
			{art("  ▲   \n ▟██▙ \n▟████▙\n█▪*▪▪█")},
			{art("  ╽   \n  ╽   \n ▟██▙ \n█▪*▪▪█")},
		},
	},
	Pad:     art(" ╞▲╡  \n ╞█╡  \n ╞█╡  \n╞╬╧╬╡ \n▄▄▄▄▄▄"),
	Filler:  []Art{art("◢◣\n▐▌"), art("▄▄▄\n█▫█"), art("▗▖\n▐▌")},
	FillPer: 10,
	Wall:    HSteel,
	Ink: map[rune]Paint{
		'▪': P(HCyan, 255), '▫': P(HWhite, 255), '▤': P(HGlass, 230), '╽': P(HWhite, 255), '▲': P(HCyan, 255),
		'╤': P(HSteel, 255), '╞': P(HSteel, 255), '╡': P(HSteel, 255), '╬': P(HGold, 255), '╧': P(HSteel, 255),
		'▄': P(HSteel, 255),
	},
	Windows: "▪▫",
	Scatter: []Scatter{
		{On: Grass, Glyphs: "♣", Fg: P(HLawn, 235), Chance: 8},
		{On: Hills, Glyphs: "▲", Fg: P(HPeak, 235), Chance: 8},
	},
	Agents: []AgentKind{
		{Name: "shuttle", Glyphs: "►◄▼▲", Fg: P(HWhite, 255), Per: 140, Max: 14, Slow: 1},
		{Name: "hovercar", Glyphs: "◆", Fg: P(HCyan, 255), Per: 100, Max: 24, Slow: 1},
		{Name: "hovercar", Glyphs: "◆", Fg: P(HNeon, 255), Per: 130, Max: 20, Slow: 1},
	},
	Night: Night{Tint: HNight, Dim: 190, Light: HCyan, Lamps: true},
	Tones: map[byte]Paint{'s': P(HConcrete, 235), 'a': P(HCyan, 255), 'm': P(HSteel, 235)},
}

// The Mars era: pressurized domes on red ground, with a lander pad at the
// edge of town.
var marsEra = AgeStyle{
	Scale:   City,
	Visitor: byShip,
	Caravan: byHover,
	Omen:    omen(HRed),
	Site:    siteStar,
	Land:    redPlanet,
	Roads: [3]RoadLook{
		{Joins: joinsHeavy, Fg: P(HConcrete, 215), Bg: P(HRock, 70)},
		{Joins: joinsLight, Fg: P(HAmber, 195), Bg: P(HRock, 56)},
		{Joins: joinsLight, Fg: P(HConcrete, 165), Bg: P(HRock, 46)},
	},
	Deck:   P(HRock, 140),
	Yard:   Look{Glyphs: "·", Fg: P(HAmber, 90), Bg: P(HRock, 56), Dense: 10},
	Crops:  Look{Glyphs: "≡", Fg: P(HLawn, 225), Bg: P(HLawn, 30), Dense: 55},
	Plaza:  Look{Glyphs: "\"'", Fg: P(HLawn, 215), Bg: P(HAmber, 28), Dense: 22},
	Rubble: Look{Glyphs: "░▒%", Fg: P(HRock, 215), Bg: P(HRock, 56), Dense: 70},
	Shapes: [numFamilies]Tiers{
		Home: {
			{art(" ◠ \n▐▪▌")},
			{art(" ◠◠ \n▐▪▫▌")},
			{art(" ◠◠◠ \n▟▪▫▪▙\n█▫▪▫▪█")},
			{art("  ◠◠  \n ◠▪▪◠ \n▟▪▫▪▫▙\n▀▀▀▀▀▀")},
		},
		Work: {
			{art("┃ \n▐*▌")},
			{art("▤▤▤\n▤*▤"), art("┃ ┃\n▐█▌\n▐*▌")},
			{art("┃  ┃ \n▟███▙\n█▪*▪▪█"), art("▤▤▤▤▤\n▤▤*▤▤\n▤▤▤▤▤")},
			{art("┃ ┃ ┃ \n▟█████\n█▪▪▪▪▪\n█▪*▪▪▪"), art("▤▤▤▤▤▤\n▤▤▤▤▤▤\n▤▤*▤▤▤")},
		},
		Civic: {
			{art(" ◠ \n▐*▌")},
			{art(" ◠◠ \n▟██▙\n█▪*█")},
			{art("  ◠   \n ◠◠◠  \n▟████▙\n█▪*▪▪█")},
			{art("  ◠◠  \n ◠▫▫◠ \n▟████▙\n█▪*▪▪█")},
		},
	},
	Pad:     art("  ◢◣  \n ◢██◣ \n╱╟██╢╲\n╞╧╧╧╧╡\n▄▄▄▄▄▄"),
	Filler:  []Art{art("◠\n▐▌"), art("▄▄▄\n█▫█"), art("▗▖\n▐▌")},
	FillPer: 10,
	Wall:    HSteel,
	Ink: map[rune]Paint{
		'▪': P(HAmber, 255), '▫': P(HWhite, 255), '◠': P(HCyan, 255), '▤': P(HGlass, 230), '┃': P(HRock, 255),
		'╱': P(HSteel, 255), '╲': P(HSteel, 255), '╟': P(HSteel, 255), '╢': P(HSteel, 255), '╞': P(HSteel, 255),
		'╡': P(HSteel, 255), '╧': P(HSteel, 255), '◢': P(HRock, 255), '◣': P(HRock, 255), '▄': P(HRock, 255),
		'▀': P(HRock, 255),
	},
	Windows: "▪▫",
	Scatter: []Scatter{
		{On: Grass, Glyphs: "♣", Fg: P(HLawn, 235), Chance: 8},
		{On: Hills, Glyphs: "▲", Fg: P(HRock, 235), Chance: 8},
	},
	Agents: []AgentKind{
		{Name: "rover", Glyphs: "►◄▼▲", Fg: P(HAmber, 255), Per: 120, Max: 20, Slow: 1},
		{Name: "rover", Glyphs: "►◄▼▲", Fg: P(HWhite, 255), Per: 170, Max: 16, Slow: 1},
		{Name: "hauler", Glyphs: "■", Fg: P(HGold, 255), Per: 200, Max: 12, Slow: 2},
	},
	Night: Night{Tint: HNight, Dim: 190, Light: HAmber, Lamps: true},
	Tones: map[byte]Paint{'s': P(HRock, 235), 'a': P(HAmber, 255), 'm': P(HSteel, 235)},
}

// The alien era: spires and drydocks on a living world, the warp nexus at
// the heart.
var alienEra = AgeStyle{
	Scale:   City,
	Visitor: byShip,
	Caravan: byFleet,
	Omen:    omen(HNeon),
	Site:    siteStar,
	Land:    alienPlanet,
	Roads: [3]RoadLook{
		{Joins: joinsHeavy, Fg: P(HNeon, 210), Bg: P(HAsphalt, 112)},
		{Joins: joinsLight, Fg: P(HCyan, 195), Bg: P(HAsphalt, 98)},
		{Joins: joinsLight, Fg: P(HSteel, 165), Bg: P(HAsphalt, 86)},
	},
	Deck:   P(HAsphalt, 140),
	Yard:   Look{Glyphs: "·", Fg: P(HNeon, 90), Bg: P(HAsphalt, 86), Dense: 10},
	Crops:  Look{Glyphs: "≡", Fg: P(HNeon, 215), Bg: P(HPlasma, 30), Dense: 55},
	Plaza:  Look{Glyphs: "✧", Fg: P(HNeon, 225), Bg: P(HPlasma, 28), Dense: 18},
	Rubble: Look{Glyphs: "░▒%", Fg: P(HSteel, 215), Bg: P(HAsphalt, 86), Dense: 70},
	Shapes: [numFamilies]Tiers{
		Home: {
			{art(" ▲ \n▐▪▌")},
			{art(" ▲▲ \n▐▪▫▌\n▐▫▪▌")},
			{art("  ▲▲  \n ◢██◣ \n▐▪▫▪▫▌")},
			{art("  ╽   \n  ▲   \n ◢██◣ \n▐▪▫▪▫▌")},
		},
		Work: {
			{art("╤ \n▐*▌")},
			{art("╤══╤\n▐*▪▌")},
			{art("╤ ◆ ╤\n╟███╢\n▐▪*▪▌")},
			{art("╤ ▲▲ ╤\n╟████╢\n╞════╡\n▐*▪▪▪▌")},
		},
		Civic: {
			{art(" ◎ \n▐*▌")},
			{art(" ◎◎ \n▟██▙\n█▪*█")},
			{art("  ◎   \n ▟██▙ \n▟████▙\n█▪*▪▪█")},
			{art("  ◎◎  \n ◢██◣ \n ▐▪▫▌ \n ▐*▪▌ ")},
		},
	},
	Pad:     art("  ◆   \n ╱█╲  \n╞╟█╢╡ \n╞╬╧╬╡ \n▄▄▄▄▄▄"),
	Filler:  []Art{art("▲\n▐▌"), art("▄▄▄\n█▫█"), art("▗▖\n▐▌")},
	FillPer: 10,
	Wall:    HSteel,
	Ink: map[rune]Paint{
		'▪': P(HNeon, 255), '▫': P(HCyan, 255), '▲': P(HNeon, 255), '◎': P(HPlasma, 255), '◆': P(HCyan, 255),
		'╽': P(HPlasma, 255), '╤': P(HCyan, 255), '╱': P(HSteel, 255), '╲': P(HSteel, 255), '╟': P(HSteel, 255),
		'╢': P(HSteel, 255), '╞': P(HSteel, 255), '╡': P(HSteel, 255), '╬': P(HGold, 255), '╧': P(HSteel, 255),
		'▄': P(HSteel, 255),
	},
	Windows: "▪▫",
	Scatter: []Scatter{
		{On: Grass, Glyphs: "✧", Fg: P(HNeon, 235), Chance: 8},
		{On: Hills, Glyphs: "▲", Fg: P(HWhite, 235), Chance: 8},
	},
	Agents: []AgentKind{
		{Name: "starship", Glyphs: "◆", Fg: P(HNeon, 255), Per: 120, Max: 20, Slow: 1},
		{Name: "starship", Glyphs: "►◄▼▲", Fg: P(HCyan, 255), Per: 160, Max: 18, Slow: 1},
	},
	Night: Night{Tint: HNight, Dim: 190, Light: HNeon, Lamps: true},
	Tones: map[byte]Paint{'s': P(HSteel, 235), 'a': P(HNeon, 255), 'm': P(HSteel, 235)},
}

// The apotheosis era: crystal monoliths, folds in the world and roads of
// woven light.
var apotheosisEra = AgeStyle{
	Scale:   City,
	Visitor: byShip,
	Caravan: byFleet,
	Omen:    omen(HPlasma),
	Site:    siteStar,
	Land:    transfiguredNatural,
	Roads: [3]RoadLook{
		{Joins: joinsLight, Fg: P(HGold, 235), Bg: P(HPlasma, 60)},
		{Joins: joinsLight, Fg: P(HGold, 205), Bg: P(HPlasma, 50)},
		{Joins: joinsLight, Fg: P(HPlasma, 185), Bg: P(HPlasma, 40)},
	},
	Deck:   P(HPlasma, 140),
	Yard:   Look{Glyphs: "·", Fg: P(HGold, 90), Bg: P(HPlasma, 40), Dense: 10},
	Crops:  Look{Glyphs: "≡", Fg: P(HGold, 215), Bg: P(HNeon, 30), Dense: 55},
	Plaza:  Look{Glyphs: "✧", Fg: P(HGold, 225), Bg: P(HNeon, 28), Dense: 18},
	Rubble: Look{Glyphs: "░▒%", Fg: P(HWhite, 190), Bg: P(HPlasma, 40), Dense: 70},
	Shapes: [numFamilies]Tiers{
		Home: {
			{art("◢◣\n▐▌")},
			{art(" ◢◣ \n◢██◣\n▐▪▪▌")},
			{art("  ◢◣  \n ◢██◣ \n◢████◣\n▐▪▫▪▫▌")},
			{art("  ▲   \n ◢██◣ \n ▐▪▫▌ \n▐▪▫▪▫▌")},
		},
		Work: {
			{art("✧ \n▐*▌")},
			{art("✧ ✧\n▟█▙\n▐*▌")},
			{art("✧  ✧ \n◢████◣\n▐▪*▪▌")},
			{art("✧ ✧ ✧ \n◢█████\n▐▪▫▪▫▪\n▐▫*▪▫▪")},
		},
		Civic: {
			{art(" ✧ \n▐*▌")},
			{art(" ✧✧ \n▟██▙\n█▪*█")},
			{art("  ✧   \n ◢██◣ \n◢████◣\n█▪*▪▪█")},
			{art("  ☼   \n ◎☼◎  \n ▟██▙ \n ▐*▪▌ ")},
		},
	},
	Pad:     art(" ╔══╗ \n ║◎◎║ \n ║  ║ \n╞╧══╧╡\n▄▄▄▄▄▄"),
	Filler:  []Art{art("◢◣\n▐▌"), art("▄▄▄\n█▫█"), art("✧\n▐▌")},
	FillPer: 10,
	Wall:    HSteel,
	Ink: map[rune]Paint{
		'▪': P(HGold, 255), '▫': P(HWhite, 255), '✧': P(HNeon, 255), '☼': P(HGold, 255), '◎': P(HPlasma, 255),
		'▲': P(HWhite, 255), '◢': P(HNeon, 255), '◣': P(HNeon, 255), '█': P(HWhite, 255), '▟': P(HWhite, 255),
		'▙': P(HWhite, 255), '╔': P(HGold, 255), '╗': P(HGold, 255), '║': P(HGold, 255), '═': P(HGold, 255),
		'╞': P(HSteel, 255), '╡': P(HSteel, 255), '╧': P(HGold, 255), '▄': P(HSteel, 255),
	},
	Windows: "▪▫",
	Scatter: []Scatter{
		{On: Grass, Glyphs: "✧", Fg: P(HGold, 235), Chance: 8},
		{On: Hills, Glyphs: "▲", Fg: P(HWhite, 235), Chance: 8},
	},
	Agents: []AgentKind{
		{Name: "phantom", Glyphs: "☺", Fg: P(HGold, 255), Per: 120, Max: 20, Slow: 3},
		{Name: "phantom", Glyphs: "★", Fg: P(HGold, 255), Per: 220, Max: 12, Slow: 3},
		{Name: "mote", Glyphs: "∙", Fg: P(HWhite, 255), Per: 100, Max: 24, Slow: 2},
		{Name: "mote", Glyphs: "✧", Fg: P(HNeon, 255), Per: 140, Max: 18, Slow: 2},
	},
	Night: Night{Tint: HNight, Dim: 190, Light: HPlasma, Lamps: true},
	Tones: map[byte]Paint{'s': P(HWhite, 235), 'a': P(HGold, 255), 'm': P(HSteel, 235)},
}
