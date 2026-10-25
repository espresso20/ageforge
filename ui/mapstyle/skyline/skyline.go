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
	"github.com/espresso20/ageforge/vibe"
)

// Entry is the skyline's registry entry.
func Entry() mapstyle.Entry {
	return mapstyle.Entry{
		Name:  "skyline",
		Title: vibe.E("ui.mapstyle.skyline.skyline.Entry.01", "Skyline"),
		Blurb: vibe.E("ui.mapstyle.skyline.skyline.Entry.02", "Your empire side-on as an ANSI panorama, one district per age, lit by the people working it."),
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
	// screen's, worldW the town's.
	zoom   zoom
	lastW  int
	sw     int
	worldW int
	// fitC is the world column the view was centred on before the far level
	// fitted the town on the screen (fit.go).
	fitC int
	// pan raises the close level's window over the scene by that many world
	// rows from where it starts; panLo and panHi are what the last frame
	// let it be (the foot of the scene up to its top).
	pan, panLo, panHi int
	// the mandala's close window pans along as well as up (orbit_mandala.go)
	panX, panXLo, panXHi int

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
// the present, ↑ ↓ between depth rows and the ridge while inspecting (at the
// close level, with the cursor away, they raise and lower the view over a
// tall tower, Shift for half a screen, and Home brings it back to the foot of
// the scene), Tab Shift-Tab through every target (the first Tab puts the
// cursor out), Esc puts it away. The flows overlay is the map flows command
// (SetOption). The Map panel hands this view the printable keys (+ - = [ ])
// only while its prompt is empty; the rest of the keys print nothing.
func (v *view) HandleKey(ev *tcell.EventKey, f mapstyle.Frame) bool {
	m := f.Model
	if m != nil && m.Sky() == mapmodel.SkyMandala {
		return v.mandalaKey(ev, f) // no panorama to scroll: orbit_mandala.go
	}
	w := v.viewW()
	key := ev.Key()
	shift := ev.Modifiers()&tcell.ModShift != 0
	fitting := v.fitting() // the whole town is in view: nothing to scroll
	scroll := func(dx int) {
		if fitting {
			return
		}
		v.follow = false
		v.cam += dx
	}
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
		if v.zoom == zClose && v.pan != 0 { // the view comes back to the foot of the scene first
			v.pan = 0
			return true
		}
		v.follow, v.cam = false, 0
	case key == tcell.KeyEnd:
		v.follow = true
	case key == tcell.KeyPgDn || key == tcell.KeyRune && (ev.Rune() == '+' || ev.Rune() == '='):
		v.setZoom(v.zoom.in(v.fitOK(v.sw)))
	case key == tcell.KeyPgUp || key == tcell.KeyRune && (ev.Rune() == '-' || ev.Rune() == '_'):
		v.setZoom(v.zoom.out(v.fitOK(v.sw)))
	case key == tcell.KeyRune && (ev.Rune() == '[' || ev.Rune() == ']'):
		if m == nil {
			return false
		}
		if !fitting {
			v.jumpDistrict(m, map[bool]int{true: -1, false: 1}[ev.Rune() == '['])
		}
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
		up := key == tcell.KeyUp
		if v.inspect && m != nil {
			v.step(m, f.VisitFrame(), 0, map[bool]int{true: 1, false: -1}[up], 0)
			return true
		}
		if v.zoom != zClose || v.panHi <= v.panLo {
			return false // nothing taller than the view to raise it over
		}
		step := panStep
		if shift {
			step = max(panStep, (v.panHi-v.panLo+1)/2)
		}
		v.pan = clampInt(v.pan+map[bool]int{true: step, false: -step}[up], v.panLo, v.panHi)
	default:
		return false
	}
	return true
}

// panStep is the world rows an up or down arrow raises or lowers the close
// level's window by: four screen rows.
const panStep = 2

// ZoomLevel and SetZoom are mapstyle.Zoomable: the level persists in the map
// settings.
func (v *view) ZoomLevel() string { return v.zoom.String() }

func (v *view) SetZoom(level string) {
	if z, ok := parseZoom(level); ok {
		v.setZoom(z)
	}
}

// setZoom changes the zoom level and keeps the view centred on the same
// spot of the town (a view that follows the present keeps following it, and
// one that leaves the far level, where the whole town was in view, comes
// back to the spot it left).
func (v *view) setZoom(z zoom) {
	if z == v.zoom {
		return
	}
	sw := v.sw
	if sw <= 0 {
		sw = 160
	}
	c := v.cam + v.viewCols(v.zoom, sw)/2
	if v.fitting() {
		c = v.fitC
	}
	into := z == zFar && v.fitOK(sw)
	if into {
		v.fitC = c
	}
	if !v.follow && !into {
		v.cam = c - v.viewCols(z, sw)/2
	}
	v.zoom = z
	v.pan = 0
	v.reveal = v.inspect // the cursor stays in view
}

// viewCols is the world columns a view at zoom z on a screen sw wide shows:
// the whole town at the far level, when that is wider than the wide level.
func (v *view) viewCols(z zoom, sw int) int {
	if z == zFar && v.fitOK(sw) {
		return v.worldW
	}
	return z.cols(sw)
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

// frameScene starts the scene of a frame: the zoom level (the far one falls
// back to the wide one when the town is too big to fit) and the frame the world
// is composed in (v.fb at mid, v.lfb otherwise).
func (v *view) frameScene(f mapstyle.Frame, W, H int) *scene {
	m := f.Model
	v.sw, v.worldW = W, m.Skyline.Width
	z := v.zoom
	fit := z == zFar && v.fitOK(W)
	if z == zFar && !fit {
		z = zWide // the wide level already shows all of a town this small
	}
	lw, lh := z.cols(W), z.frameRows(H)
	wcols, kx := lw, 1.0 // the world columns in view, and the share of them a frame column holds
	if fit {
		wcols, kx = m.Skyline.Width, float64(m.Skyline.Width)/float64(lw)
	}
	world := &v.fb
	if z != zMid {
		world = &v.lfb
	}
	v.lastW = wcols
	return &scene{v: v, m: m, fb: world, tier: f.Tier, anim: f.Anim, visit: f.VisitFrame(), W: lw, H: lh, S: lh - 3, top: 1, sel: -1,
		zm: z, lw: wcols, ls: lh - 3, sw: W, sh: H, fit: fit, kx: kx}
}

// placeCamera puts the camera and the cursor in a scene whose layout is set:
// it follows the present, reveals the cursor's target, clamps the camera and
// takes the pan (panDef rows up from the foot), the ridge, the selected lot
// and the grid-out flag of a catastrophe.
func (v *view) placeCamera(s *scene, f mapstyle.Frame, panDef int) {
	m, fit, lw, wcols := s.m, s.fit, s.W, s.lw
	if v.follow && !fit {
		v.cam = presentCam(m, lw)
	}
	if v.inspect {
		ts := v.targetsFor(m, wcols, v.camX())
		if i := v.resolve(m, ts, f.VisitFrame()); i >= 0 && v.reveal {
			v.revealTarget(m, ts[i], wcols)
		}
	}
	if !fit {
		v.cam = clampInt(v.cam, 0, max(0, m.Skyline.Width-lw))
	}
	s.cam = v.camX()
	s.pan = v.panFor(s.zm, s.S, panDef)
	if fit {
		s.ridge = ridgeItemsFar(m, lw)
	} else {
		s.ridge = ridgeItems(m, lw, s.cam)
	}
	if v.inspect && v.cur.kind == tLot {
		s.sel = s.lay.find(v.cur.key, v.cur.cp)
	}
	if ep := m.Catastrophe.Pending; ep != "" {
		e := m.Catalog.EpochIdx[ep]
		s.gridOut = e == 3 || e == 4 || e == 5
	}
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
	s := v.frameScene(f, W, H)
	fit := s.fit
	s.groundY = s.S - 3
	s.band = bandOf(m.AgeIdx)
	s.p, s.mp = v.palettes(m)
	s.lay = v.layoutSized(m, s.groundY, H-6)
	s.vis = v.vis
	s.city, s.inCity = mapmodel.CityLookAt(m.AgeIdx)
	defer func() { v.vis = s.vis }()
	v.placeCamera(s, f, 0)
	s.fb.reset(s.W, s.H)
	s.sky()
	s.celestial()
	s.skyStructures()
	s.clouds()
	s.farRidge()
	s.nearLayer()
	if fit {
		s.fitTown()
		s.fitGround()
	} else {
		s.lots()
		s.cityRoofs()
		s.ground()
		s.traffic()
	}
	s.visitor()
	if !fit {
		s.smoke()
	}
	s.weather()
	if !fit {
		s.catastrophe()
	} else {
		s.catastropheSky()
	}
	sc := s.present()
	if fit {
		sc.farFlags()
		sc.districtNames(1)
	}
	sc.markers()
	sc.panBar()
	sc.chrome()
	sc.clearCursorFill()
	return sc
}

// panFor is the rows the close level's window is raised by in a frame of S
// scene rows: the view's pan from where the scene wants the window to start
// (def rows up from the foot), kept inside the scene, and the range the
// view's pan may take is noted for the keys. Away from the close level
// there is no window to raise.
func (v *view) panFor(z zoom, S, def int) int {
	hi := S - 1 - (S-1)/2 // the window's top reaches the scene's top
	v.panLo, v.panHi = -def, hi-def
	v.panXLo, v.panXHi = 0, 0
	if z != zClose {
		return 0
	}
	v.pan = clampInt(v.pan, v.panLo, v.panHi)
	return def + v.pan
}

// revealTarget scrolls so the cursor's target is on screen.
func (v *view) revealTarget(m *mapmodel.Model, t tgt, W int) {
	v.reveal = false
	if v.fitting() {
		return // the whole town is in view
	}
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

// clearCursorFill keeps every cell that is not the cursor's own off the
// cursor's fill and its blink shade. A theme with few colours can land a
// scene's own shade on them exactly, and a cell that looks like the cursor
// points at nothing: it is moved a step toward the page's text.
func (s *scene) clearCursorFill() {
	if s.mp == nil {
		return
	}
	_, fill, _ := s.mp.Cursor.Decompose()
	_, alt, _ := s.mp.CursorAlt.Decompose()
	step := theme.Color(theme.RoleText)
	nudge := func(c tcell.Color) tcell.Color { return theme.Mix(c, step, 0.05) }
	f := s.fb
	for i := range f.c {
		c := &f.c[i]
		if c.d == dTop {
			continue
		}
		if c.ch == '█' {
			if c.fg == fill || c.fg == alt {
				c.fg = nudge(c.fg)
			}
			continue
		}
		if c.bg == fill || c.bg == alt {
			c.bg = nudge(c.bg)
		}
	}
}
