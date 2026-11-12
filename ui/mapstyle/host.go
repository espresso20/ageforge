package mapstyle

import (
	"image"
	"image/color"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/world"
)

// A SpaceHost is where a window that can draw its own pictures takes the
// sky: the view hands it the world model and where it goes, and the window
// draws every space scale itself, with one continuous zoom. The cells under
// the picture are drawn with theme.Transparent as their ground so the
// picture shows through; the header, news and inspector stay cells. With no
// host (the terminal) the cell painters draw the sky as they always did.
type SpaceHost interface {
	// Show says the sky is w and goes in rect (screen cells). It is called on
	// every draw and is cheap: w is the same pointer until key changes, and
	// pal is the theme's colours resolved to plain RGBA.
	Show(rect image.Rectangle, w *world.World, key uint64, pal world.Palette)
	// Hide says the map no longer shows the sky (the panel closed, the rect
	// went, or the age has no sky).
	Hide()
	// Command moves the camera: the map's keys, while the sky is shown.
	Command(c MapCmd)
	// Watch gives the host the words for the names it writes by objects
	// (see SkyName) and where to report the frame and selection it is in.
	// fn runs on the host's goroutines: it must only store the state.
	Watch(names func(label string) string, fn func(MapState))
	// TakeSettled hands over the uncertain places the player settled since
	// the last call, for the game to keep in the save.
	TakeSettled() []string
}

// MapOp is what a map key asks the camera to do.
type MapOp uint8

const (
	MapPanLeft MapOp = iota
	MapPanRight
	MapPanUp
	MapPanDown
	MapZoomIn
	MapZoomOut
	MapNext // the next pickable object, centred
	MapPrev
	MapHome // back to the home object
)

// A MapCmd is one key for the camera. Fast is Shift: pan four times as far.
type MapCmd struct {
	Op   MapOp
	Fast bool
}

// keyOps are the camera ops the map's keys stand for.
var keyOps = map[tcell.Key]MapOp{
	tcell.KeyLeft: MapPanLeft, tcell.KeyRight: MapPanRight, tcell.KeyUp: MapPanUp, tcell.KeyDown: MapPanDown,
	tcell.KeyPgUp: MapZoomOut, tcell.KeyPgDn: MapZoomIn, tcell.KeyTab: MapNext, tcell.KeyBacktab: MapPrev, tcell.KeyHome: MapHome,
}

// cmdOf is the camera command a map key stands for.
func cmdOf(ev *tcell.EventKey) (MapCmd, bool) {
	op, ok := keyOps[ev.Key()]
	return MapCmd{Op: op, Fast: ev.Modifiers()&tcell.ModShift != 0}, ok
}

// toneHues says which of the cell painters' sky hues a tone is coloured by,
// so the window's sky and the terminal's wear the same theme.
var toneHues = [world.NumTones]mapgen.Hue{
	world.ToneSun: mapgen.HSun, world.ToneAmber: mapgen.HAmber, world.ToneRed: mapgen.HRed, world.ToneCool: mapgen.HCool,
	world.ToneWhite: mapgen.HWhite, world.ToneRock: mapgen.HRock, world.ToneCore: mapgen.HCore, world.ToneSand: mapgen.HSand,
	world.ToneGold: mapgen.HGold, world.TonePanel: mapgen.HPanel, world.ToneCyan: mapgen.HCyan, world.ToneStar: mapgen.HStar,
	world.TonePink: mapgen.HNeon, world.ToneTeal: mapgen.HHydro, world.ToneBlue: mapgen.HAir, world.ToneOrange: mapgen.HFire,
	world.ToneArm: mapgen.HArm, world.ToneDust: mapgen.HNebula,
}

func rgba(c tcell.Color) color.RGBA {
	r, g, b := c.RGB()
	return color.RGBA{R: uint8(max(r, 0)), G: uint8(max(g, 0)), B: uint8(max(b, 0)), A: 255}
}

// skyPalette is the sky's colours in the cell palette's theme.
func skyPalette(p *cellPalette) world.Palette {
	out := world.Palette{Bg: rgba(p.bg)}
	for t := range out.Tones {
		out.Tones[t] = rgba(p.hue[toneHues[t]])
	}
	return out
}

// hostSky gives the sky to the host when there is a host and an age that has
// one, clears the cells under it and reports whether it did. Otherwise the
// host is told to hide, and the cell painters draw.
func (v *view) hostSky(cv *Canvas, f Frame, p *cellPalette, x, y, w, h int) bool {
	if f.Space == nil || f.Model == nil || w <= 0 || h <= 0 {
		return false
	}
	seed := f.Model.Seed
	if key := world.ChangeKey(f.Model, seed); v.sky == nil || v.sky.Key != key {
		v.sky = world.Build(f.Model, seed)
	}
	if v.sky.Root == nil {
		return false
	}
	if p != v.skyFor {
		v.skyFor, v.skyPal = p, skyPalette(p)
	}
	if v.watched != f.Space {
		f.Space.Watch(SkyName, v.hst.set)
		v.watched = f.Space
	}
	key := v.sky.Key ^ v.skyPal.Hash()*0x100000001b3
	f.Space.Show(image.Rect(cv.R.X+x, cv.R.Y+y, cv.R.X+x+w, cv.R.Y+y+h), v.sky, key, v.skyPal)
	clear := tcell.StyleDefault.Foreground(p.bg).Background(theme.Transparent)
	for yy := 0; yy < h; yy++ {
		for xx := 0; xx < w; xx++ {
			cv.Put(x+xx, y+yy, ' ', clear)
		}
	}
	return true
}

// unhost tells the host the sky is gone, once.
func (v *view) unhost(f Frame) {
	if v.hosted && f.Space != nil {
		f.Space.Hide()
		v.hst.clear()
	}
	v.hosted = false
}
