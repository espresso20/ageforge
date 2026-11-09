package mapgen

import "github.com/espresso20/ageforge/world"

// toneHue maps the sky roles of world/ onto the painters' hues.
var toneHue = [...]Hue{
	world.ToneNone:  HNone,
	world.ToneSun:   HSun,
	world.ToneAmber: HAmber,
	world.ToneRed:   HRed,
	world.ToneCool:  HCool,
	world.ToneWhite: HWhite,
	world.ToneRock:  HRock,
	world.ToneCore:  HCore,
	world.ToneSand:  HSand,
	world.ToneGold:  HGold,
	world.TonePanel: HPanel,
	world.ToneCyan:  HCyan,
}

// hueOf is the painters' hue for a sky role.
func hueOf(t world.Tone) Hue { return toneHue[t] }

// paintOf is the painters' paint for a shade.
func paintOf(s world.Shade) Paint { return Paint{hueOf(s.Tone), s.W} }
