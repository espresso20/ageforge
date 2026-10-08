package roguelike

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// space.go is where the roguelike leaves the ground. From the Space Age on
// (mapmodel/sky.go) the settlement and district zooms draw a sky scene
// instead of the town: the planet from orbit, deep space and its colony
// worlds, the starbase hub, a starbase in superposition, and the mandala of
// the ages. The region zoom stays the known world (the ground view's
// plate), so worldmap and the civs keep their map. skyStyle is the one
// hook: the registry entry wraps the ground view in it, and it hands each
// call to the ground view or the sky view.

// skyStyle routes a roguelike view between the ground and the sky.
type skyStyle struct {
	g  *view
	sv *skyView
}

func withSky(g *view) *skyStyle { return &skyStyle{g: g, sv: &skyView{g: g}} }

// onSky reports whether a frame of m draws the sky view: a sky age at the
// settlement or district zoom.
func (s *skyStyle) onSky(m *mapmodel.Model) bool {
	return m != nil && m.Sky() != mapmodel.SkyGround && s.g.zoom != zRegion
}

func (s *skyStyle) Name() string { return s.g.Name() }

// SceneInset: the full view keeps one row for its header and four below
// the map for the news, the inspector and the status lines, on the ground
// and in the sky alike (view.Draw, skyView.Draw).
func (s *skyStyle) SceneInset(w, h int) (top, bottom int) { return 1, 4 }

func (s *skyStyle) Draw(scr tcell.Screen, r mapstyle.Rect, f mapstyle.Frame) {
	if s.onSky(f.Model) {
		s.sv.Draw(scr, r, f)
		return
	}
	s.g.Draw(scr, r, f)
}

func (s *skyStyle) DrawCompact(scr tcell.Screen, r mapstyle.Rect, f mapstyle.Frame) {
	if f.Model != nil && f.Model.Sky() != mapmodel.SkyGround {
		s.sv.DrawCompact(scr, r, f)
		return
	}
	s.g.DrawCompact(scr, r, f)
}

// HandleKey: in a sky age PgUp from the sky climbs to the known world (the
// ground view's region zoom) and PgDn there comes back down to the sky.
func (s *skyStyle) HandleKey(ev *tcell.EventKey, f mapstyle.Frame) bool {
	m := f.Model
	if ev == nil || m == nil || m.Sky() == mapmodel.SkyGround {
		return s.g.HandleKey(ev, f)
	}
	if s.g.zoom == zRegion {
		if ev.Key() == tcell.KeyPgDn {
			s.g.zoom = zSettlement
			return true
		}
		return s.g.HandleKey(ev, f)
	}
	return s.sv.HandleKey(ev, f)
}

func (s *skyStyle) Inspect(f mapstyle.Frame) (mapstyle.Inspection, bool) {
	if s.onSky(f.Model) {
		return s.sv.Inspect(f)
	}
	return s.g.Inspect(f)
}

// SetOption: the options live on the ground view, which the sky view reads.
func (s *skyStyle) SetOption(o mapstyle.Option, on bool) { s.g.SetOption(o, on) }

// CompactShowsNews: both views end the compact view with the news line.
func (s *skyStyle) CompactShowsNews() bool { return true }

// skyView draws the sky scenes. Its options (flows, inspect, legend,
// changes) and its zoom are the ground view's; the cursor, the camera and
// the scene cache are its own.
type skyView struct {
	g      *view
	cur    mapmodel.Pt
	seed   int64
	placed bool
	scene  mapmodel.SkyScene
	cam    [3]mapmodel.Pt
	camKey [3][2]int
	snap   bool // put the camera on the cursor at the next draw

	sc   *skyScene
	scM  *mapmodel.Model
	base *skyBase
	sp   *skyPal

	// per-frame scratch
	sg      skyGeom
	anim    int
	visit   int // the frame the rare visitor is drawn at (Frame.VisitFrame)
	clock   int // the world's clock it came from (Frame.Clock)
	tier    mapmodel.GlyphTier
	occ     []bool
	seen    []skyLgEntry
	compact bool
	names   int
	movers  []skyMoverAt // where each mover is this frame (traffic)
}

// sceneFor returns the scene for a model, building it when the layout
// changes; a new seed, or a new scene, puts the cursor on the hub.
func (v *skyView) sceneFor(m *mapmodel.Model) *skyScene {
	if m == nil {
		return nil
	}
	sky := m.Sky()
	if sky == mapmodel.SkyGround {
		return nil
	}
	switch {
	case v.sc != nil && m == v.scM:
	case v.sc != nil && v.scM != nil && m.LayoutKey == v.scM.LayoutKey && sky == v.sc.sky:
		v.scM, v.sc.m = m, m // same layout, only the clock moved
	default:
		if v.base == nil || v.base.seed != m.Seed || v.base.sky != sky {
			v.base = newSkyBase(m, sky)
		}
		v.scM, v.sc = m, newSkyScene(m, v.base)
	}
	if !v.placed || v.seed != m.Seed || v.scene != sky {
		v.seed, v.scene, v.placed, v.snap = m.Seed, sky, true, true
		v.cur = v.sc.hub
	}
	return v.sc
}

// HandleKey takes only keys that print nothing: the arrows move the cursor
// (Shift by 8), Tab and Shift-Tab step through what can be inspected, Home
// goes back to the hub, PgDn zooms in to the close view, PgUp out to the
// sky and then (in skyStyle) up to the known world.
func (v *skyView) HandleKey(ev *tcell.EventKey, f mapstyle.Frame) bool {
	s := v.sceneFor(f.Model)
	if ev == nil || s == nil {
		return false
	}
	g := v.g
	switch ev.Key() {
	case tcell.KeyTab, tcell.KeyBacktab:
		dir := 1
		if ev.Key() == tcell.KeyBacktab {
			dir = -1
		}
		v.jump(s, dir)
		return true
	case tcell.KeyPgUp:
		if g.zoom == zDistrict {
			g.zoom = zSettlement
		} else {
			g.zoom = zRegion
		}
		return true
	case tcell.KeyPgDn:
		g.zoom = zDistrict
		return true
	case tcell.KeyHome:
		v.cur, v.snap = s.hub, true
		g.inspect = true
		return true
	}
	d, ok := keyMoves[ev.Key()]
	if !ok {
		return false
	}
	if ev.Modifiers()&tcell.ModShift != 0 {
		d = [2]int{d[0] * 8, d[1] * 8}
	}
	g.inspect = true
	v.cur.X = clamp(v.cur.X+d[0], 0, skyW-1)
	v.cur.Y = clamp(v.cur.Y+d[1], 0, skyH-1)
	return true
}

// jump moves the cursor to the next (dir 1) or previous inspect target.
func (v *skyView) jump(s *skyScene, dir int) {
	ts := s.targets
	if len(ts) == 0 {
		return
	}
	v.g.inspect = true
	n := len(ts) - 1
	if dir > 0 {
		n = 0
	}
	for i, t := range ts {
		if t == v.cur {
			n = ((i+dir)%len(ts) + len(ts)) % len(ts)
			break
		}
	}
	v.cur = ts[n]
}
