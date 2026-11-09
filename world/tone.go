package world

// Tone is a colour role in the sky: the model names a role, the renderer
// picks the pixels from the theme. Never raw RGB. The names match the sky
// roles the terminal painters use, so mapgen converts one to one.
type Tone uint8

const (
	ToneNone Tone = iota
	ToneSun
	ToneAmber
	ToneRed
	ToneCool
	ToneWhite
	ToneRock
	ToneCore
	ToneSand
	ToneGold
	TonePanel
	ToneCyan
	ToneStar // a plain star's light
	TonePink // the knots of an arm, emission nebulae
	ToneTeal
	ToneBlue
	ToneOrange
	ToneArm  // the light of a spiral arm
	ToneDust // dark lanes
	numTones
)

// Shade is a tone at a weight: 255 is the full colour, less is dimmer. A
// zero Shade is nothing.
type Shade struct {
	Tone Tone
	W    uint8
}

// S is a shade of t at weight w.
func S(t Tone, w uint8) Shade { return Shade{t, w} }
