// Package skyline is the Skyline map style: the empire drawn side-on as an
// ANSI panorama. Each age lived through is a district (oldest west, the
// present east, the build queue as cranes on the frontier); every
// silhouette is a building owned; windows light and chimneys smoke only
// where workers are staffed; the sky follows the clock and the weather; the
// civs met stand as towns on the far ridge with the harbinger; and the
// traffic of each age weaves between the building rows, as dense as the
// trade, labour, soldiers and wealth behind it.
package skyline

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// Entry is the skyline's registry entry.
func Entry() mapstyle.Entry {
	return mapstyle.Entry{
		Name:  "skyline",
		Title: "Skyline",
		Blurb: "Your empire side-on as an ANSI panorama, one district per age, lit by the people working it.",
		New:   func() mapstyle.Style { return newView() },
	}
}

// view is one skyline view: the camera, the cursor, the toggles and the
// caches. A frame is a pure function of (model, anim, tier, this state, the
// active theme).
type view struct {
	cam     int
	follow  bool // the camera tracks the present
	inspect bool
	cur     target
	reveal  bool // scroll the cursor into view on the next draw
	flows   bool
	changes bool
	legend  bool

	// zoom is how close the view looks (zoom.go); lastW is the world's
	// columns the last frame showed (what the cursor maths assume), sw the
	// screen's.
	zoom  zoom
	lastW int
	sw    int

	lay     *layout
	clay    *layout // the compact view's
	slay    *layout // the sky arc's (orbit.go), full and compact,
	sclay   *layout
	glay    *layout        // and the ground skyline behind the Quantum Age's echo
	md      *mandalaLayout // the Transcendent Age's mandala (orbit_mandala.go)
	sprites map[spriteKey]*sprite
	fb      fb
	lfb     fb // the world's frame when it is not the screen's (zoom.go)
	vis     []int

	palM   *mapmodel.Model
	palKey [3]tcell.Color
	palT   string
	pal    *pal
	mp     *mapstyle.Palette
	mpKey  [2]tcell.Color
	mpEp   int

	ts    []tgt
	tsM   *mapmodel.Model
	tsW   int
	tsCam int

	noTraffic bool // tests: compose without traffic
}

func newView() *view { return &view{follow: true, changes: true} }

func (v *view) Name() string { return "skyline" }

// SceneInset: the full view keeps one row for its header and two below
// the scene for the strip and the key bar (chrome.go).
func (v *view) SceneInset(w, h int) (top, bottom int) { return 1, 2 }

// viewW is the width the cursor maths assume before the first draw.
func (v *view) viewW() int {
	if v.lastW > 0 {
		return v.lastW
	}
	return 160
}

func (v *view) SetOption(o mapstyle.Option, on bool) {
	switch o {
	case mapstyle.OptFlows:
		v.flows = on
	case mapstyle.OptInspect:
		v.inspect = on
		v.reveal = on
	case mapstyle.OptLegend:
		v.legend = on
	case mapstyle.OptChanges:
		v.changes = on
	}
}

// HandleKey: ← → scroll (or move the cursor while inspecting), Shift-← →
// half a screen, PgDn PgUp (and + and -, = too) zoom in and out, [ and ]
// jump to the previous and next district, Home End the oldest district and
// the present, ↑ ↓ between depth rows and the ridge while inspecting, Tab
// Shift-Tab through every target (the first Tab puts the cursor out), Esc
// puts it away. The flows overlay is the map flows command (SetOption). The
// Map panel hands this view the printable keys (+ - = [ ]) only while its
// prompt is empty; the rest of the keys print nothing.
func (v *view) HandleKey(ev *tcell.EventKey, f mapstyle.Frame) bool {
	m := f.Model
	if m != nil && m.Sky() == mapmodel.SkyMandala {
		return v.mandalaKey(ev, f) // no panorama to scroll: orbit_mandala.go
	}
	w := v.viewW()
	key := ev.Key()
	shift := ev.Modifiers()&tcell.ModShift != 0
	scroll := func(dx int) {
		v.follow = false
		v.cam += dx
	}
	ground := m != nil && m.Sky() == mapmodel.SkyGround // the sky arc has no zoom
	switch {
	case key == tcell.KeyTab || key == tcell.KeyBacktab:
		if m == nil {
			return false
		}
		d := 1
		if key == tcell.KeyBacktab || shift {
			d = -1
		}
		if !v.inspect {
			v.inspect = true
			v.step(m, f.VisitFrame(), 0, 0, 0)
		} else {
			v.step(m, f.VisitFrame(), 0, 0, d)
		}
	case key == tcell.KeyEscape:
		if !v.inspect {
			return false
		}
		v.inspect = false
	case key == tcell.KeyHome:
		v.follow, v.cam = false, 0
	case key == tcell.KeyEnd:
		v.follow = true
	case key == tcell.KeyPgDn || key == tcell.KeyRune && (ev.Rune() == '+' || ev.Rune() == '='):
		if !ground {
			return false
		}
		v.setZoom(v.zoom.in())
	case key == tcell.KeyPgUp || key == tcell.KeyRune && (ev.Rune() == '-' || ev.Rune() == '_'):
		if !ground {
			return false
		}
		v.setZoom(v.zoom.out())
	case key == tcell.KeyRune && (ev.Rune() == '[' || ev.Rune() == ']'):
		if m == nil {
			return false
		}
		v.jumpDistrict(m, map[bool]int{true: -1, false: 1}[ev.Rune() == '['])
	case key == tcell.KeyLeft && shift:
		scroll(-w / 2)
	case key == tcell.KeyRight && shift:
		scroll(w / 2)
	case key == tcell.KeyLeft || key == tcell.KeyRight:
		dx := map[bool]int{true: -1, false: 1}[key == tcell.KeyLeft]
		if v.inspect && m != nil {
			v.step(m, f.VisitFrame(), dx, 0, 0)
		} else {
			scroll(v.zoom.scrollStep() * dx)
		}
	case key == tcell.KeyUp || key == tcell.KeyDown:
		if !v.inspect || m == nil {
			return false
		}
		v.step(m, f.VisitFrame(), 0, map[bool]int{true: 1, false: -1}[key == tcell.KeyUp], 0)
	default:
		return false
	}
	return true
}

// ZoomLevel and SetZoom are mapstyle.Zoomable: the level persists in the map
// settings.
func (v *view) ZoomLevel() string { return v.zoom.String() }

func (v *view) SetZoom(level string) {
	if z, ok := parseZoom(level); ok {
		v.setZoom(z)
	}
}

// setZoom changes the zoom level and keeps the view centred on the same
// spot of the town (a view that follows the present keeps following it).
func (v *view) setZoom(z zoom) {
	if z == v.zoom {
		return
	}
	sw := v.sw
	if sw <= 0 {
		sw = 160
	}
	if !v.follow {
		c := v.cam + v.zoom.cols(sw)/2
		v.cam = c - z.cols(sw)/2
	}
	v.zoom = z
	v.reveal = v.inspect // the cursor stays in view
}

// jumpDistrict puts the left edge of the view on the start of the district
// before (dir < 0) or after (dir > 0) the one the view is at, and stops
// following the present. With none that way it stays where it is.
func (v *view) jumpDistrict(m *mapmodel.Model, dir int) {
	cam := v.cam
	to := -1
	for _, d := range m.Skyline.Districts {
		switch {
		case dir > 0 && d.X0 > cam:
			if to < 0 {
				to = d.X0
			}
		case dir < 0 && d.X0 < cam:
			to = d.X0
		}
	}
	if to < 0 {
		return
	}
	v.follow, v.cam = false, to
}

func (v *view) Inspect(f mapstyle.Frame) (mapstyle.Inspection, bool) {
	if !v.inspect || f.Model == nil {
		return mapstyle.Inspection{}, false
	}
	in, ok := v.inspection(f.Model, f.VisitFrame())
	if !ok {
		return mapstyle.Inspection{}, false
	}
	return mapstyle.Inspection{Title: in.title, Lines: in.lines, Command: in.cmd, Kind: in.kind}, true
}

// Draw renders the full view; below 60x16 it falls back to the compact one.
func (v *view) Draw(scr tcell.Screen, r mapstyle.Rect, f mapstyle.Frame) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	if f.Model == nil {
		blank(scr, r)
		return
	}
	if r.W < 60 || r.H < 16 {
		v.DrawCompact(scr, r, f)
		return
	}
	v.compose(f, r.W, r.H)
	v.fb.blit(scr, r, f.Tier)
}

func blank(scr tcell.Screen, r mapstyle.Rect) {
	cv := &mapstyle.Canvas{Scr: scr, R: r}
	st := tcell.StyleDefault.Foreground(theme.Color(theme.RoleDim)).Background(theme.Color(theme.RoleBackground))
	cv.Fill(0, 0, r.W, r.H, ' ', st)
}

// palettes returns the frame palette and the chrome palette, cached per
// model and theme.
func (v *view) palettes(m *mapmodel.Model) (*pal, *mapstyle.Palette) {
	bg, tx := theme.Color(theme.RoleBackground), theme.Color(theme.RoleText)
	key := [3]tcell.Color{bg, tx, theme.Color(theme.RoleAccent)}
	tk := theme.Active().Key
	if v.pal == nil || v.palM != m || v.palKey != key || v.palT != tk {
		v.pal, v.palM, v.palKey, v.palT = newPal(m), m, key, tk
	}
	if v.mp == nil || v.mpKey != [2]tcell.Color{key[0], key[1]} || v.mpEp != m.Epoch {
		v.mp, v.mpKey, v.mpEp = mapstyle.NewPalette(m.Epoch), [2]tcell.Color{key[0], key[1]}, m.Epoch
	}
	return v.pal, v.mp
}

// presentCam frames the newest district and the frontier cranes.
func presentCam(m *mapmodel.Model, w int) int {
	return m.Skyline.FrontierX + 34 - w
}

// compose builds the frame into v.fb and returns the scene as the screen
// sees it. The world is composed at today's scale into a frame of its own
// (v.lfb, or v.fb itself at mid), and the zoom decides how that frame
// reaches the screen (zoom.go); the bars and the text are then drawn on the
// screen.
func (v *view) compose(f mapstyle.Frame, W, H int) *scene {
	if sky := f.Model.Sky(); sky != mapmodel.SkyGround {
		v.sw = W
		return v.composeSky(f, W, H, sky) // the sky arc: orbit.go
	}
	m := f.Model
	z := v.zoom
	lw, lh := z.cols(W), z.frameRows(H)
	world := &v.fb
	if z != zMid {
		world = &v.lfb
	}
	v.sw, v.lastW = W, lw
	s := &scene{v: v, m: m, fb: world, tier: f.Tier, anim: f.Anim, visit: f.VisitFrame(), W: lw, H: lh, S: lh - 3, top: 1, sel: -1,
		zm: z, lw: lw, ls: lh - 3, sw: W, sh: H}
	s.groundY = s.S - 3
	s.band = bandOf(m.AgeIdx)
	s.p, s.mp = v.palettes(m)
	s.lay = v.layoutSized(m, s.groundY, H-6)
	s.vis = v.vis
	s.city, s.inCity = mapmodel.CityLookAt(m.AgeIdx)
	defer func() { v.vis = s.vis }()
	if v.follow {
		v.cam = presentCam(m, lw)
	}
	if v.inspect {
		ts := v.targetsFor(m, lw, v.cam)
		if i := v.resolve(m, ts, f.VisitFrame()); i >= 0 && v.reveal {
			v.revealTarget(m, ts[i], lw)
		}
	}
	v.cam = clampInt(v.cam, 0, max(0, m.Skyline.Width-lw))
	s.cam = v.cam
	s.ridge = ridgeItems(m, lw, s.cam)
	if v.inspect && v.cur.kind == tLot {
		s.sel = s.lay.find(v.cur.key, v.cur.cp)
	}
	if ep := m.Catastrophe.Pending; ep != "" {
		e := m.Catalog.EpochIdx[ep]
		s.gridOut = e == 3 || e == 4 || e == 5
	}
	world.reset(lw, lh)
	s.sky()
	s.celestial()
	s.skyStructures()
	s.clouds()
	s.farRidge()
	s.nearLayer()
	s.lots()
	s.cityRoofs()
	s.ground()
	s.traffic()
	s.visitor()
	s.smoke()
	s.weather()
	s.catastrophe()
	sc := s.present()
	sc.markers()
	sc.chrome()
	return sc
}

// revealTarget scrolls so the cursor's target is on screen.
func (v *view) revealTarget(m *mapmodel.Model, t tgt, W int) {
	v.reveal = false
	x := t.x - v.cam
	if t.kind != tLot && t.kind != tTether { // the tether stands in the world, like a building
		x = t.x - v.cam // ridge targets are placed in screen terms already
		if x >= 4 && x < W-4 {
			return
		}
		v.follow = false
		v.cam += int(float64(x-W/2) / 0.15)
		return
	}
	if x >= 6 && x < W-6 {
		return
	}
	v.follow = false
	v.cam = t.x - W/2
}
