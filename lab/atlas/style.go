package main

import (
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/lucasb-eyer/go-colorful"
)

// Plate is one epoch's cartographic medium: a small table of dials the one
// renderer reads. No plate has code of its own; a plate is data.
type Plate struct {
	Key      string
	Name     string // shown in the legend: "Portolan chart"
	Title    string // cartouche title, %s is the home region
	Sub      string // cartouche subtitle
	Unit     string // scale bar unit
	UnitPer  float64
	Model    string // "earth", "orrery", "stars", "galaxy"
	PaperHue string // hex tint of the paper
	PaperMix float64

	SeaGlyphs  []rune // sparse sea marks
	SeaDensity float64
	WaterLines []float64 // offshore distances drawn as coast-parallel lines
	Coast      string    // "round", "square", "none", "dim"
	CoastBold  bool
	Relief     string // "marks", "hachure", "contour", "shade", "none"
	LandMarks  map[string]rune
	MarkDens   float64
	Rhumbs     bool
	Graticule  float64 // spacing in world units, 0 for none
	GridRefs   bool
	Frame      string // "hide", "double", "neat", "survey", "hud", "chart"
	Fog        string // "dark", "dragons", "blank", "unsurveyed", "none"
	Night      bool   // land is dark, cities are lights
	Sweep      bool   // radar sweep over the capital
	Tracks     bool   // satellite ground tracks
	Scan       bool   // satellite downlink scanline
	Wash       float64
	BorderRune rune
	RiverH     rune
	RiverV     rune
	RouteSea   rune
	RouteLand  rune
	TrailRune  rune
	Capital    rune
	Town       rune
	CivCapital rune
	Wonder     rune
	Ship       rune
	Caravan    rune
	Front      rune
	Scar       rune
	Omen       rune
	Caps       bool // labels in spaced capitals
	Mini       bool // the sidebar locator: realm, peoples, routes, alerts only
}

// mini is the plate reduced for the sidebar: no furniture, no fine relief.
func (p *Plate) mini() *Plate {
	q := *p
	q.Mini = true
	q.Rhumbs = false
	q.Graticule = 0
	q.WaterLines = nil
	q.Sweep = false
	q.Tracks = false
	if q.Relief == "contour" || q.Relief == "hachure" {
		q.Relief = "marks"
		q.MarkDens = 0.25
		q.LandMarks = map[string]rune{"mountains": '^', "peaks": '^'}
	}
	return &q
}

var plates = map[string]*Plate{
	"hide": {
		Key: "hide", Name: "Charcoal on hide", Title: "the lands we know", Sub: "drawn by the fire of %s",
		Unit: "days' walk", UnitPer: 8, Model: "earth", PaperHue: "#8a6a44", PaperMix: 0.16,
		SeaGlyphs: []rune{'~', '˜'}, SeaDensity: 0.07, Coast: "round",
		Relief: "marks", MarkDens: 0.22,
		LandMarks: map[string]rune{"mountains": '^', "peaks": '^', "hills": '∩', "forest": '♣', "desert": '∴', "tundra": '⁘', "ice": '*', "steppe": '"', "grassland": ','},
		Frame:     "hide", Fog: "dark", BorderRune: '·', RiverH: '~', RiverV: '≀',
		RouteSea: '·', RouteLand: '·', TrailRune: '˙', Capital: '⌂', Town: '⌂', CivCapital: 'Δ',
		Wonder: '✦', Ship: '◆', Caravan: '•', Front: '✕', Scar: '◌', Omen: '☄',
	},
	"portolan": {
		Key: "portolan", Name: "Portolan chart", Title: "Carta of %s", Sub: "with the winds and the harbours",
		Unit: "leagues", UnitPer: 6, Model: "earth", PaperHue: "#c9a86a", PaperMix: 0.13,
		SeaDensity: 0, Coast: "round", CoastBold: true, Relief: "marks", MarkDens: 0.16,
		LandMarks: map[string]rune{"mountains": '⋀', "peaks": '⋀', "hills": '⌒', "forest": '♣', "desert": '∴', "ice": '*', "tundra": '·'},
		Rhumbs:    true, Frame: "double", Fog: "dragons", BorderRune: '·', RiverH: '~', RiverV: '≀',
		RouteSea: '·', RouteLand: '·', TrailRune: '˙', Capital: '♜', Town: '▪', CivCapital: '♜',
		Wonder: '✦', Ship: '◆', Caravan: '•', Front: '✕', Scar: '◌', Omen: '☄', Caps: true,
	},
	"engraved": {
		Key: "engraved", Name: "Engraved map", Title: "A New & Accurate Map of %s", Sub: "from the latest surveys",
		Unit: "miles", UnitPer: 20, Model: "earth", PaperHue: "#e8dcc0", PaperMix: 0.06,
		WaterLines: []float64{1.3, 2.6, 4.2}, Coast: "round", CoastBold: true, Relief: "hachure", MarkDens: 0.5,
		LandMarks: map[string]rune{"forest": '♣', "peaks": '▲'},
		Graticule: 30, Frame: "neat", Fog: "blank", Wash: 0.12, BorderRune: '·', RiverH: '~', RiverV: '≀',
		RouteSea: '╌', RouteLand: '╌', TrailRune: '·', Capital: '◉', Town: '○', CivCapital: '◉',
		Wonder: '✦', Ship: '◆', Caravan: '•', Front: '✕', Scar: '⊗', Omen: '☄', Caps: true,
	},
	"survey": {
		Key: "survey", Name: "Survey sheet", Title: "Ordnance Survey of %s", Sub: "sheet %d · 1:250 000",
		Unit: "km", UnitPer: 40, Model: "earth", PaperHue: "#ffffff", PaperMix: 0.03,
		Coast: "square", Relief: "contour", MarkDens: 0.3,
		LandMarks: map[string]rune{"forest": '♧'},
		Graticule: 20, GridRefs: true, Frame: "survey", Fog: "unsurveyed", BorderRune: '╌', RiverH: '─', RiverV: '│',
		RouteSea: '┄', RouteLand: '┿', TrailRune: '┄', Capital: '■', Town: '▪', CivCapital: '■',
		Wonder: '✦', Ship: '◆', Caravan: '•', Front: '✕', Scar: '⊗', Omen: '✶', Caps: true,
	},
	"satellite": {
		Key: "satellite", Name: "Satellite mosaic", Title: "EARTH OBSERVATION · %s", Sub: "multispectral · cloud-free composite",
		Unit: "km", UnitPer: 40, Model: "earth", PaperHue: "#000000", PaperMix: 0.0,
		Coast: "none", Relief: "shade", Graticule: 30, Frame: "hud", Fog: "none", Scan: true,
		BorderRune: '┄', RiverH: '─', RiverV: '│', RouteSea: '·', RouteLand: '·', TrailRune: '·',
		Capital: '▣', Town: '□', CivCapital: '▣', Wonder: '✦', Ship: '◆', Caravan: '◆', Front: '✕', Scar: '⊗', Omen: '✶',
	},
	"radar": {
		Key: "radar", Name: "Orbital night pass", Title: "NIGHTSIDE · %s GRID", Sub: "orbital radar · city lights",
		Unit: "km", UnitPer: 40, Model: "earth", PaperHue: "#000814", PaperMix: 0.0,
		Coast: "dim", Relief: "none", Graticule: 45, Frame: "hud", Fog: "none", Night: true, Sweep: true, Tracks: true,
		BorderRune: '┄', RiverH: ' ', RiverV: ' ', RouteSea: '·', RouteLand: '·', TrailRune: '·',
		Capital: '◈', Town: '◇', CivCapital: '◈', Wonder: '✦', Ship: '◆', Caravan: '◆', Front: '✕', Scar: '⊗', Omen: '✶',
	},
	"orrery": {
		Key: "orrery", Name: "Orrery", Title: "THE SYSTEM OF %s", Sub: "planets, colonies and lanes",
		Unit: "AU", UnitPer: 0.1, Model: "orrery", Frame: "chart", Omen: '☄',
		Capital: '◉', Town: '•', CivCapital: '◈', Ship: '◆', Front: '✕', Scar: '⊗',
	},
	"stars": {
		Key: "stars", Name: "Star chart", Title: "STELLAR CHART · SOL NEIGHBOURHOOD", Sub: "within 30 light years",
		Unit: "ly", UnitPer: 0.25, Model: "stars", Frame: "chart", Omen: '☄',
		Capital: '☉', Town: '•', CivCapital: '✶', Ship: '◆', Front: '✕', Scar: '⊗',
	},
	"galaxy": {
		Key: "galaxy", Name: "Galactic atlas", Title: "THE GALAXY", Sub: "claimed arms and hyperlanes",
		Unit: "kly", UnitPer: 0.5, Model: "galaxy", Frame: "chart", Omen: '☄',
		Capital: '☉', Town: '•', CivCapital: '✶', Ship: '◆', Front: '✕', Scar: '⊗',
	},
}

// plateOrder is the atlas's plates in the order an empire earns them.
var plateOrder = []string{"hide", "portolan", "engraved", "survey", "satellite", "radar", "orrery", "stars", "galaxy"}

// plateForAge is the epoch dial: which plate an age draws on.
func plateForAge(age string) string {
	switch age {
	case "primitive_age", "stone_age", "bronze_age":
		return "hide"
	case "iron_age", "classical_age", "medieval_age":
		return "portolan"
	case "renaissance_age", "colonial_age", "industrial_age":
		return "engraved"
	case "victorian_age", "electric_age", "atomic_age":
		return "survey"
	case "modern_age", "information_age", "digital_age":
		return "satellite"
	case "cyberpunk_age", "fusion_age":
		return "radar"
	case "space_age":
		return "orrery"
	case "interstellar_age":
		return "stars"
	}
	return "galaxy"
}

// ---------------------------------------------------------------------------
// colour classes

// Pal is the frame's colour classes. Everything on the map is one of these;
// most are theme roles outright, the rest are a fixed hue lifted or sunk to
// sit on the theme's background. Computed once per frame, never per cell.
type Pal struct {
	Light                                        bool
	Bg, Paper, Ink, Faint, Label, Accent, Border tcell.Color
	Pos, Neg, Warn, High, Chip, Dim              tcell.Color
	Water, WaterDeep, WaterWash, Relief, Forest  tcell.Color
	Sand, Snow, Grass, Rock, Light1, Light2      tcell.Color
	PlayerWash, FreshWash, NightLand             tcell.Color
	BiomeWash                                    map[string]tcell.Color
	RelWash                                      map[string]tcell.Color
	mixed                                        map[mixKey]tcell.Color
}

type mixKey struct {
	a, b tcell.Color
	t    float64
}

// mix is a colour class between two others, computed once per frame and
// then looked up: the frame has a few dozen classes, not a colour per cell.
func (p *Pal) mix(a, b tcell.Color, t float64) tcell.Color {
	k := mixKey{a, b, t}
	if c, ok := p.mixed[k]; ok {
		return c
	}
	c := mix(a, b, t)
	p.mixed[k] = c
	return c
}

func toColorful(c tcell.Color) colorful.Color {
	r, g, b := c.RGB()
	return colorful.Color{R: float64(r) / 255, G: float64(g) / 255, B: float64(b) / 255}
}

func fromColorful(c colorful.Color) tcell.Color {
	r, g, b := c.Clamped().RGB255()
	return tcell.NewRGBColor(int32(r), int32(g), int32(b))
}

func hex(s string) colorful.Color { c, _ := colorful.Hex(s); return c }

// ink lifts a hue to sit on a dark background or sinks it for a light one.
func ink(h string, light bool, strength float64) tcell.Color {
	hh, cc, _ := hex(h).Hcl()
	l := 0.58 + 0.2*strength
	if light {
		l = 0.52 - 0.22*strength
	}
	return fromColorful(colorful.Hcl(hh, cc, l))
}

func mix(a, b tcell.Color, t float64) tcell.Color {
	return fromColorful(toColorful(a).BlendLab(toColorful(b), t))
}

func newPal(p *Plate) *Pal {
	bg := theme.Color(theme.RoleBackground)
	light := theme.IsLight()
	pal := &Pal{Light: light, Bg: bg, mixed: map[mixKey]tcell.Color{},
		Ink: theme.Color(theme.RoleText), Faint: theme.Color(theme.RoleDim), Label: theme.Color(theme.RoleLabel),
		Accent: theme.Color(theme.RoleAccent), Border: theme.Color(theme.RoleBorder), Pos: theme.Color(theme.RolePositive),
		Neg: theme.Color(theme.RoleNegative), Warn: theme.Color(theme.RoleWarning), High: theme.Color(theme.RoleHighlight),
		Chip: theme.Color(theme.RoleChip), Dim: theme.Color(theme.RoleDim),
	}
	pal.Paper = bg
	if p.PaperMix > 0 {
		pal.Paper = mix(bg, fromColorful(hex(p.PaperHue)), p.PaperMix)
	}
	// night and satellite plates are photographs: a dark ground even on a
	// light theme would be a hole in the page, so they print lighter there.
	if p.Night && !light {
		pal.Paper = mix(bg, fromColorful(hex("#02101c")), 0.6)
	}
	pal.Water = ink("#5b8fd6", light, 0.3)
	pal.WaterDeep = ink("#34588f", light, -0.4)
	pal.WaterWash = mix(pal.Paper, fromColorful(hex("#3a78c8")), 0.14)
	pal.Relief = ink("#a07850", light, 0.2)
	pal.Forest = ink("#5e9a4a", light, 0.2)
	pal.Sand = ink("#c8a870", light, 0.4)
	pal.Snow = ink("#dfe8ef", light, 0.9)
	pal.Grass = ink("#8aa85a", light, 0.1)
	pal.Rock = ink("#8c8a86", light, 0.2)
	pal.Light1 = ink("#ffd27a", light, 0.9)
	pal.Light2 = ink("#ff9a3c", light, 0.4)
	pal.PlayerWash = mix(pal.Paper, pal.Accent, p.Wash)
	pal.NightLand = mix(pal.Paper, pal.Faint, 0.1)
	pal.BiomeWash = map[string]tcell.Color{}
	for b, c := range map[string]tcell.Color{"forest": pal.Forest, "grassland": pal.Grass, "steppe": pal.Grass, "tundra": pal.Rock,
		"desert": pal.Sand, "ice": pal.Snow, "peaks": pal.Snow, "hills": pal.Relief, "mountains": pal.Rock} {
		pal.BiomeWash[b] = mix(pal.Paper, c, 0.3)
	}
	pal.FreshWash = mix(pal.Paper, pal.High, 0.22)
	pal.RelWash = map[string]tcell.Color{}
	for _, rel := range []string{"allied", "friendly", "neutral", "rival", "war"} {
		pal.RelWash[rel] = mix(pal.Paper, pal.Rel(rel), p.Wash)
	}
	return pal
}

// relation is the civ's standing as one of five signal classes.
func relation(c *Civ) string {
	switch {
	case c.Info.AtWar:
		return "war"
	case c.Info.Status == "allied":
		return "allied"
	case c.Info.Status == "rival" || c.Info.Status == "embargo":
		return "rival"
	case c.Info.Status == "friendly" || c.Info.Opinion >= 25:
		return "friendly"
	}
	return "neutral"
}

// Rel is the fixed relationship signal colour: war red, ally green,
// friendly gold, rival amber, neutral steel (the label role).
func (p *Pal) Rel(rel string) tcell.Color {
	switch rel {
	case "war":
		return p.Neg
	case "allied":
		return p.Pos
	case "friendly":
		return p.High
	case "rival":
		return p.Warn
	}
	return p.Label
}
