package theme

import (
	"github.com/espresso20/ageforge/rowid"
)

// Flavor themes (the theming design §4) are cosmetic, curated palettes that an
// account badge gives (UnlockBadge): the badge of an age for the five age
// themes, and a badge of its own for each of the rest. They are all
// Accessible: false. Because they're not accessible, the
// colorblind-distinguishability guard does not apply to them, but every WCAG AA
// luminance floor in contrast_test.go DOES, and each theme below is tuned to clear
// those with margin (Text/Label/Positive/Negative/Highlight vs Background >= 4.5,
// Dim/Accent vs Background >= 3.0, Text-on-Selection >= 4.5).
//
// Flavor themes don't need colorblind separation, but the ▲/▼ glyphs are harmless
// and keep delta formatting consistent across every theme, so they're set anyway.
const (
	flavorGainGlyph = "▲"
	flavorLossGlyph = "▼"
)

// Parchment is the unlockable LIGHT-background theme: dark ink-brown text on warm cream,
// with sepia/leather accents. It's the real contrast-guard exercise (the theming design §4,
// §8) — on a light bg the role colors must be DARK enough to stay legible, so the
// usual bright accent/positive/negative become deep, saturated versions: forest
// green gains, wax-red losses, umber labels, sepia accent/highlight.
var Parchment = define(Theme{
	Key:        "parchment",
	Duotone:    true,
	Name:       rowid.E("theme.parchment.name", "Parchment"),
	Blurb:      rowid.E("theme.parchment.blurb", "Ink on warm parchment — a sepia manuscript page."),
	Accessible: false,
	Colors: [numRoles]Color{
		RoleBackground: NewRGBColor(0xf2, 0xe8, 0xd0), // warm cream
		RoleText:       NewRGBColor(0x3a, 0x2c, 0x1a), // dark ink brown
		RoleDim:        NewRGBColor(0x6e, 0x5c, 0x40), // faded sepia
		RoleLabel:      NewRGBColor(0x6b, 0x4a, 0x12), // deep umber
		RoleAccent:     NewRGBColor(0x7e, 0x50, 0x10), // sepia / leather
		RoleHighlight:  NewRGBColor(0x7e, 0x50, 0x10), // wax-amber (dark on light)
		RolePositive:   NewRGBColor(0x2c, 0x6e, 0x2c), // forest green
		RoleNegative:   NewRGBColor(0xa3, 0x2a, 0x1f), // wax red
		RoleSelection:  NewRGBColor(0xd8, 0xc4, 0x9a), // darker cream backing
		RoleComment:    NewRGBColor(0xad, 0x42, 0x08), // flavor sentences only: rust orange
		RoleEffect:     NewRGBColor(0x1f, 0x4e, 0x8c), // what a tech does: ink blue
	},
	GainGlyph: flavorGainGlyph,
	LossGlyph: flavorLossGlyph,
	// The Renaissance Age's badge gives it: the manuscript page for the age of
	// letters and printing.
	UnlockBadge: "age.renaissance_age",
	UnlockHint:  rowid.E("theme.parchment.unlock_hint", "Reach the Renaissance Age"),
})

// Bronze is a warm metallic look: copper/bronze accent, amber highlight, olive
// gains and burnt-orange losses on a dark warm brown-black background.
var Bronze = define(Theme{
	Key:        "bronze",
	Name:       rowid.E("theme.bronze.name", "Bronze"),
	Blurb:      rowid.E("theme.bronze.blurb", "Burnished copper and amber over dark, warm metal."),
	Accessible: false,
	Colors: [numRoles]Color{
		RoleBackground: NewRGBColor(0x1c, 0x14, 0x0c), // dark warm brown-black
		RoleText:       NewRGBColor(0xf0, 0xe2, 0xc8), // warm parchment text
		RoleDim:        NewRGBColor(0x9a, 0x82, 0x60), // patina'd bronze
		RoleLabel:      NewRGBColor(0xd9, 0xa8, 0x6a), // copper
		RoleAccent:     NewRGBColor(0xcd, 0x7f, 0x32), // bronze
		RoleHighlight:  NewRGBColor(0xf2, 0xc1, 0x60), // amber
		RolePositive:   NewRGBColor(0x8f, 0xb5, 0x4a), // olive gain
		RoleNegative:   NewRGBColor(0xe0, 0x6a, 0x3c), // burnt-orange loss
		RoleSelection:  NewRGBColor(0x3a, 0x2a, 0x18),
		RoleComment:    NewRGBColor(0xff, 0x8a, 0x1a), // flavor sentences only: bright orange, brighter than the bronze accent
		RoleEffect:     NewRGBColor(0x7f, 0xb2, 0xea), // what a tech does: steel blue on the brown canvas
	},
	GainGlyph: flavorGainGlyph,
	LossGlyph: flavorLossGlyph,
	// The Bronze Age's badge gives it: burnished metal for the age that first
	// worked it. The earliest of the gated themes.
	UnlockBadge: "age.bronze_age",
	UnlockHint:  rowid.E("theme.bronze.unlock_hint", "Reach the Bronze Age"),
})

// Cyberpunk is neon on near-black: hot magenta accent, neon-cyan highlight, neon
// green/pink for ±, over a near-black violet background (riffs on the cyberpunk_age
// palette, the theming design §4).
var Cyberpunk = define(Theme{
	Key:        "cyberpunk",
	Name:       rowid.E("theme.cyberpunk.name", "Cyberpunk"),
	Blurb:      rowid.E("theme.cyberpunk.blurb", "Hot magenta and neon cyan over near-black violet."),
	Accessible: false,
	Colors: [numRoles]Color{
		RoleBackground: NewRGBColor(0x0a, 0x06, 0x12), // near-black violet
		RoleText:       NewRGBColor(0xe6, 0xe6, 0xf5), // cool white
		RoleDim:        NewRGBColor(0x7a, 0x6e, 0x9e), // muted violet-grey
		RoleLabel:      NewRGBColor(0x3d, 0xe1, 0xe8), // cyan
		RoleAccent:     NewRGBColor(0xff, 0x2e, 0xc0), // hot magenta
		RoleHighlight:  NewRGBColor(0x4d, 0xf0, 0xff), // neon cyan
		RolePositive:   NewRGBColor(0x39, 0xff, 0x9e), // neon green gain
		RoleNegative:   NewRGBColor(0xff, 0x49, 0x6b), // neon pink-red loss
		RoleSelection:  NewRGBColor(0x2a, 0x12, 0x3a),
		RoleComment:    NewRGBColor(0xff, 0x8a, 0x1f), // flavor sentences only: neon orange
		RoleEffect:     NewRGBColor(0x5b, 0x8c, 0xff), // what a tech does: electric blue
	},
	GainGlyph: flavorGainGlyph,
	LossGlyph: flavorLossGlyph,
	// The Cyberpunk Age's badge gives it; the theme riffs on that age's palette.
	UnlockBadge: "age.cyberpunk_age",
	UnlockHint:  rowid.E("theme.cyberpunk.unlock_hint", "Reach the Cyberpunk Age"),
})

// Monochrome is a stylistic greyscale terminal: a single hue's shades. Accent and
// Highlight are bright greys/white; the ± distinction rides on LIGHTNESS (a lighter
// grey gains, a mid grey loses) rather than hue. It is NOT an accessibility theme —
// just a clean mono look — so it carries Accessible: false and skips the colorblind
// guard, but still clears every luminance floor.
var Monochrome = define(Theme{
	Key:        "monochrome",
	Duotone:    true,
	Name:       rowid.E("theme.monochrome.name", "Monochrome"),
	Blurb:      rowid.E("theme.monochrome.blurb", "Greyscale terminal — meaning carried by lightness, not hue."),
	Accessible: false,
	Colors: [numRoles]Color{
		RoleBackground: NewRGBColor(0x12, 0x12, 0x12), // near-black grey
		RoleText:       NewRGBColor(0xf5, 0xf5, 0xf5), // near-white
		RoleDim:        NewRGBColor(0x8a, 0x8a, 0x8a), // mid grey
		RoleLabel:      NewRGBColor(0xc8, 0xc8, 0xc8), // light grey
		RoleAccent:     NewRGBColor(0xe8, 0xe8, 0xe8), // bright grey
		RoleHighlight:  NewRGBColor(0xff, 0xff, 0xff), // white
		RolePositive:   NewRGBColor(0xd8, 0xd8, 0xd8), // lighter grey = gain
		RoleNegative:   NewRGBColor(0x9a, 0x9a, 0x9a), // mid grey = loss
		RoleSelection:  NewRGBColor(0x33, 0x33, 0x33),
		RoleComment:    NewRGBColor(0xe8, 0x88, 0x3a), // flavor sentences only: the one warm ink, for the voice
		RoleEffect:     NewRGBColor(0x9d, 0xb4, 0xcc), // what a tech does: the one cool ink; every other role is grey
	},
	GainGlyph: flavorGainGlyph,
	LossGlyph: flavorLossGlyph,
	// The Information Age's badge gives it: the retro terminal look for the age
	// that put a terminal on every desk.
	UnlockBadge: "age.information_age",
	UnlockHint:  rowid.E("theme.monochrome.unlock_hint", "Reach the Information Age"),
})

// Cosmic is deep-space: a dark indigo/violet background, starlight text, and
// nebula-pink / cyan accents with aurora-green gains (riffs on galactic_age,
// the theming design §4).
var Cosmic = define(Theme{
	Key:        "cosmic",
	Name:       rowid.E("theme.cosmic.name", "Cosmic"),
	Blurb:      rowid.E("theme.cosmic.blurb", "Deep indigo with starlight and nebula accents."),
	Accessible: false,
	Colors: [numRoles]Color{
		RoleBackground: NewRGBColor(0x0c, 0x0a, 0x1f), // deep indigo
		RoleText:       NewRGBColor(0xe8, 0xe6, 0xf8), // starlight
		RoleDim:        NewRGBColor(0x7e, 0x78, 0xa6), // dust violet
		RoleLabel:      NewRGBColor(0x6e, 0xd6, 0xe8), // cyan
		RoleAccent:     NewRGBColor(0xc8, 0x6e, 0xff), // nebula violet
		RoleHighlight:  NewRGBColor(0xff, 0x8a, 0xd8), // nebula pink
		RolePositive:   NewRGBColor(0x5a, 0xe0, 0xa8), // aurora green
		RoleNegative:   NewRGBColor(0xff, 0x6e, 0x8a), // nebula red
		RoleSelection:  NewRGBColor(0x24, 0x20, 0x4a),
		RoleComment:    NewRGBColor(0xff, 0x9a, 0x3c), // flavor sentences only: nebula orange
		RoleEffect:     NewRGBColor(0x7a, 0xa8, 0xff), // what a tech does: nebula blue
	},
	GainGlyph: flavorGainGlyph,
	LossGlyph: flavorLossGlyph,
	// The Galactic Age's badge gives it; the theme riffs on that age's palette.
	UnlockBadge: "age.galactic_age",
	UnlockHint:  rowid.E("theme.cosmic.unlock_hint", "Reach the Galactic Age"),
})

// Source is a phosphor terminal: green on green-black, with code falling down
// the empty columns. It is the reward of the Touched by the Source badge, so
// its hint does not say how it is earned.
var Source = define(Theme{
	Key:        "source",
	Name:       rowid.E("theme.source.name", "Source"),
	Blurb:      rowid.E("theme.source.blurb", "Phosphor green on black, with code rain in the empty columns."),
	Accessible: false,
	Colors: [numRoles]Color{
		RoleBackground: NewRGBColor(0x00, 0x0a, 0x03), // green-black
		RoleText:       NewRGBColor(0x6c, 0xff, 0x9a), // phosphor green
		RoleDim:        NewRGBColor(0x2f, 0x8a, 0x4f), // dim phosphor
		RoleLabel:      NewRGBColor(0x39, 0xd2, 0x6a), // mid green
		RoleAccent:     NewRGBColor(0xb6, 0xff, 0xca), // pale green
		RoleHighlight:  NewRGBColor(0xe9, 0xff, 0x7a), // yellow-green
		RolePositive:   NewRGBColor(0x5d, 0xff, 0xd0), // aqua gain
		RoleNegative:   NewRGBColor(0xff, 0x4d, 0x4d), // red loss
		RoleSelection:  NewRGBColor(0x0d, 0x2b, 0x17),
		RoleComment:    NewRGBColor(0xff, 0x9a, 0x3a), // flavor sentences only: amber against the greens
		RoleEffect:     NewRGBColor(0x7f, 0xd4, 0xff), // what a tech does: sky blue; the only cool hue on phosphor green
		RoleBright:     NewRGBColor(0xd9, 0xff, 0xe3), // near-white green
	},
	GainGlyph:   flavorGainGlyph,
	LossGlyph:   flavorLossGlyph,
	UnlockBadge: "special.touched_by_the_source",
	UnlockHint:  rowid.E("theme.source.unlock_hint", "Given by a secret badge"),
	Effect:      EffectRain,
})

// Glitch is a broken signal: cyan and magenta on violet black, and now and
// then a tear of static across the empty cells. It is the reward of the
// Creative Accounting badge, so its hint does not say how it is earned.
var Glitch = define(Theme{
	Key:        "glitch",
	Name:       rowid.E("theme.glitch.name", "Glitch"),
	Blurb:      rowid.E("theme.glitch.blurb", "Cyan and magenta on violet black, with a tear of static now and then."),
	Accessible: false,
	Colors: [numRoles]Color{
		RoleBackground: NewRGBColor(0x07, 0x06, 0x0d), // violet black
		RoleText:       NewRGBColor(0xe6, 0xe6, 0xff), // cool white
		RoleDim:        NewRGBColor(0x8a, 0x86, 0xa8), // violet-grey
		RoleLabel:      NewRGBColor(0x39, 0xe6, 0xff), // cyan
		RoleAccent:     NewRGBColor(0xff, 0x2e, 0x9a), // magenta
		RoleHighlight:  NewRGBColor(0xf8, 0xff, 0x5a), // acid yellow
		RolePositive:   NewRGBColor(0x39, 0xff, 0x88), // static green gain
		RoleNegative:   NewRGBColor(0xff, 0x3b, 0x3b), // red loss
		RoleSelection:  NewRGBColor(0x2a, 0x16, 0x40),
		RoleComment:    NewRGBColor(0xff, 0x8a, 0x2a), // flavor sentences only: neon orange
		RoleEffect:     NewRGBColor(0x6d, 0x8f, 0xff), // what a tech does: signal blue
		RoleBright:     NewRGBColor(0xff, 0xff, 0xff),
	},
	GainGlyph:   flavorGainGlyph,
	LossGlyph:   flavorLossGlyph,
	UnlockBadge: "special.creative_accounting",
	UnlockHint:  rowid.E("theme.glitch.unlock_hint", "Given by a secret badge"),
	Effect:      EffectGlitch,
})

// Ashfall is the morning after: ash grey and ember orange on soot, with
// sparks drifting up the empty columns. It is the reward of a badge that
// stays out of sight until the player has met an ending, so its hint does
// not name it.
var Ashfall = define(Theme{
	Key:        "ashfall",
	Name:       rowid.E("theme.ashfall.name", "Ashfall"),
	Blurb:      rowid.E("theme.ashfall.blurb", "Ember orange on soot, with sparks rising in the empty columns."),
	Accessible: false,
	Colors: [numRoles]Color{
		RoleBackground: NewRGBColor(0x15, 0x11, 0x0f), // soot
		RoleText:       NewRGBColor(0xd8, 0xcf, 0xc7), // ash
		RoleDim:        NewRGBColor(0x8a, 0x7f, 0x77), // cold ash
		RoleLabel:      NewRGBColor(0xe0, 0x89, 0x4a), // ember
		RoleAccent:     NewRGBColor(0xff, 0x6a, 0x2b), // live coal
		RoleHighlight:  NewRGBColor(0xff, 0xb3, 0x47), // flame
		RolePositive:   NewRGBColor(0xa9, 0xc4, 0x6c), // new growth
		RoleNegative:   NewRGBColor(0xff, 0x5a, 0x4d), // burn
		RoleSelection:  NewRGBColor(0x2b, 0x21, 0x1c),
		RoleComment:    NewRGBColor(0xff, 0xcf, 0x9e), // flavor sentences only: pale apricot: every orange here is taken
		RoleEffect:     NewRGBColor(0x8f, 0xb6, 0xe0), // what a tech does: cold steel blue; every warm hue here is taken
		RoleBright:     NewRGBColor(0xf3, 0xec, 0xe6),
	},
	GainGlyph:   flavorGainGlyph,
	LossGlyph:   flavorLossGlyph,
	UnlockBadge: "special.connoisseur_of_endings",
	UnlockHint:  rowid.E("theme.ashfall.unlock_hint", "Given by a legendary badge"),
	Effect:      EffectEmbers,
})

// Ledger is the accountant's page: ink and bookkeeper's green on pale
// green-bar paper, every other row of the empty space tinted. Light. It is the reward of the
// top rung of the deals ladder.
var Ledger = define(Theme{
	Key:        "ledger",
	Name:       rowid.E("theme.ledger.name", "Ledger"),
	Blurb:      rowid.E("theme.ledger.blurb", "Ink and bookkeeper's green on green-bar paper."),
	Accessible: false,
	Colors: [numRoles]Color{
		RoleBackground: NewRGBColor(0xf4, 0xf6, 0xee), // ledger paper
		RoleText:       NewRGBColor(0x1d, 0x2a, 0x22), // ink
		RoleDim:        NewRGBColor(0x5d, 0x6b, 0x61), // pencil
		RoleLabel:      NewRGBColor(0x1f, 0x6b, 0x45), // bookkeeper's green
		RoleAccent:     NewRGBColor(0x1f, 0x6b, 0x45),
		RoleHighlight:  NewRGBColor(0x8a, 0x5a, 0x00), // brass
		RolePositive:   NewRGBColor(0x14, 0x66, 0x3a), // in the black
		RoleNegative:   NewRGBColor(0xb3, 0x26, 0x1e), // in the red
		RoleSelection:  NewRGBColor(0xdf, 0xe9, 0xd6),
		RoleComment:    NewRGBColor(0xb5, 0x47, 0x09), // flavor sentences only: rust orange
		RoleEffect:     NewRGBColor(0x1f, 0x4f, 0x8f), // what a tech does: ledger blue
		RoleBright:     NewRGBColor(0x0b, 0x13, 0x0e),
	},
	GainGlyph:   flavorGainGlyph,
	LossGlyph:   flavorLossGlyph,
	UnlockBadge: "ledger.deals.3",
	UnlockHint:  rowid.E("theme.ledger.unlock_hint", "Given by the top rung of a tally"),
	Effect:      EffectGreenbar,
})

// Prismatic is white light split: near-white text on blue-black, with glints
// in the empty space that turn slowly through the spectrum. It is the reward
// of Museum Piece, the badge for earning most of the others.
var Prismatic = define(Theme{
	Key:        "prismatic",
	Name:       rowid.E("theme.prismatic.name", "Prismatic"),
	Blurb:      rowid.E("theme.prismatic.blurb", "Starlight on blue-black, with glints that turn through the spectrum."),
	Accessible: false,
	Colors: [numRoles]Color{
		RoleBackground: NewRGBColor(0x0b, 0x0b, 0x10), // blue-black
		RoleText:       NewRGBColor(0xec, 0xec, 0xf4), // starlight
		RoleDim:        NewRGBColor(0x8b, 0x8b, 0xa3), // haze
		RoleLabel:      NewRGBColor(0x8f, 0xd3, 0xff), // sky
		RoleAccent:     NewRGBColor(0xff, 0x7a, 0xd9), // rose
		RoleHighlight:  NewRGBColor(0xff, 0xe0, 0x66), // sunbeam
		RolePositive:   NewRGBColor(0x7d, 0xff, 0xb0), // spring green
		RoleNegative:   NewRGBColor(0xff, 0x5c, 0x5c), // red
		RoleSelection:  NewRGBColor(0x1d, 0x1b, 0x2c),
		RoleComment:    NewRGBColor(0xff, 0x9a, 0x3c), // flavor sentences only: orange
		RoleEffect:     NewRGBColor(0x8a, 0xa2, 0xff), // what a tech does: periwinkle
		RoleBright:     NewRGBColor(0xff, 0xff, 0xff),
	},
	GainGlyph:   flavorGainGlyph,
	LossGlyph:   flavorLossGlyph,
	UnlockBadge: "special.museum_piece",
	UnlockHint:  rowid.E("theme.prismatic.unlock_hint", "Earn 160 badges"),
	Effect:      EffectPrism,
})

var _ = register(
	Parchment,
	Bronze,
	Cyberpunk,
	Monochrome,
	Cosmic,
	Source,
	Glitch,
	Ashfall,
	Ledger,
	Prismatic,
)
