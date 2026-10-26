package mapgen

// The colours below are the game's own map hues (theme/maphues.go), laid on
// the forge theme's near-black background.

var bronze = AgeStyle{
	From:  "primitive_age",
	Scene: Ground,
	Land: [numBiomes]Look{
		Deep:     {Glyphs: "≈~", Fg: rgb(0x2a6ab0), Bg: rgb(0x0a1d36), Dense: 55},
		Shallows: {Glyphs: "~", Fg: rgb(0x3a86c8), Bg: rgb(0x0e2c4a), Dense: 40},
		Sand:     {Glyphs: "·.", Fg: rgb(0x8a7a50), Bg: rgb(0x2a2619), Dense: 30},
		Scrub:    {Glyphs: ",'.", Fg: rgb(0x7a7a42), Bg: rgb(0x1a1c12), Dense: 35},
		Grass:    {Glyphs: "\"',.", Fg: rgb(0x4f8a3a), Bg: rgb(0x111b13), Dense: 30},
		Forest:   {Glyphs: "♣♣τ♠", Fg: rgb(0x3f9a4f), Bg: rgb(0x0e1c13), Dense: 70},
		Hills:    {Glyphs: "^n∩", Fg: rgb(0x9a8a5a), Bg: rgb(0x1b1c14), Dense: 40},
		Mountain: {Glyphs: "▲^", Fg: rgb(0x9a8a78), Bg: rgb(0x1a1a1c), Dense: 70},
		River:    {Fg: rgb(0x3a86c8), Bg: rgb(0x0e2c4a)},
	},
	Roads: [3]RoadLook{
		{Joins: joinsLight, Fg: rgb(0xb08a55), Bg: rgb(0x221c12)},
		{Joins: "·∙", Fg: rgb(0x8a7048), Bg: rgb(0x1d1810)},
		{Joins: "·.", Fg: rgb(0x75603f), Bg: rgb(0x1a160f)},
	},
	Deck:   rgb(0x6a4a2a),
	Yard:   Look{Glyphs: "·", Fg: rgb(0x3a3322), Bg: rgb(0x1d1a12), Dense: 10},
	Crops:  Look{Glyphs: "≡", Fg: rgb(0x9aaa3a), Bg: rgb(0x1b2410), Dense: 60},
	Plaza:  Look{Glyphs: "\"", Fg: rgb(0x3f6a3a), Bg: rgb(0x14201a), Dense: 25},
	Rubble: Look{Glyphs: "░▒%", Fg: rgb(0x6a5a48), Bg: rgb(0x1d1a12), Dense: 70},
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
	Wall:    rgb(0xa89a82),
	Ink:     map[rune]RGB{'/': rgb(0xc08a4a), '\\': rgb(0xc08a4a), '^': rgb(0xc08a4a), '_': rgb(0xc08a4a), '▲': rgb(0xc08a4a), '∩': rgb(0x6a4a2a), '╷': rgb(0x707880)},
	Windows: "▫∩",
	Chimney: '╷',
	Smoke:   Look{Glyphs: "°∘", Fg: rgb(0xb0b4bc)},
	Scatter: []Scatter{
		{On: Grass, Glyphs: "♣τ", Fg: rgb(0x3f9a4f), Chance: 14},
		{On: Scrub, Glyphs: "♠", Fg: rgb(0x6a8a42), Chance: 10},
		{On: Hills, Glyphs: "▲", Fg: rgb(0x9a8a78), Chance: 10},
	},
	Agents: []AgentKind{
		{Name: "walker", Glyphs: "☺", Fg: rgb(0xf0e0c0), Per: 18, Max: 24, Slow: 3},
		{Name: "cart", Glyphs: "o=", Fg: rgb(0xc08a4a), Per: 60, Max: 6, Slow: 5},
	},
	Night: Night{Tint: rgb(0x060a18), Dim: 165, Light: rgb(0xffc45a)},
	Tones: map[byte]RGB{'s': rgb(0x8a847a), 'a': rgb(0xb07a3a), 'g': rgb(0xffc45a), 'v': rgb(0x4f8a3a)},
}
