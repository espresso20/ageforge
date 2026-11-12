package mapstyle

import (
	"strings"
	"sync"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/world"
)

// hosted.go: what the window's own map tells the ui. The window reports the
// frame it is in and the object selected (MapState); the scale label and the
// inspect lines follow it, and the names written by the objects come from here.

// MapState is the window map's frame and selection. Sel points into the world
// the host was shown and is read-only; it is nil when nothing is selected.
type MapState struct {
	Frame    world.FrameKind // the frame the camera is centred in
	SelFrame world.FrameKind // the frame Sel belongs to
	Sel      *world.Object
	Lock     string // the age that opens Sel's frame, while the player's age has not reached it
}

// hostState is the last MapState, kept for the next draw. The host reports on
// its own goroutines, so this is all a report does: no ui calls.
type hostState struct {
	mu sync.Mutex
	s  MapState
	ok bool
}

func (h *hostState) set(s MapState) {
	h.mu.Lock()
	h.s, h.ok = s, true
	h.mu.Unlock()
}

func (h *hostState) get() (MapState, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.s, h.ok
}

func (h *hostState) clear() {
	h.mu.Lock()
	h.s, h.ok = MapState{}, false
	h.mu.Unlock()
}

// The home system's names are rows shared with the terminal's inspector.
var (
	solName   = vibe.E("ui.mapstyle.space.sun.01", "Sol")
	moonName  = vibe.E("ui.mapstyle.space.moon.01", "The Moon")
	satName   = vibe.E("ui.mapstyle.space.satellite.01", "Satellite")
	stnName   = vibe.E("ui.mapstyle.space.station.name", "Orbital Space Station")
	plutoName = vibe.E("ui.mapstyle.space.planet.pluto", "Pluto")
)

// moonNames are the major moons by the world's keys.
var moonNames = map[string]vibe.Entry{
	"phobos":    vibe.E("ui.mapstyle.space.moon.phobos", "Phobos"),
	"deimos":    vibe.E("ui.mapstyle.space.moon.deimos", "Deimos"),
	"io":        vibe.E("ui.mapstyle.space.moon.io", "Io"),
	"europa":    vibe.E("ui.mapstyle.space.moon.europa", "Europa"),
	"ganymede":  vibe.E("ui.mapstyle.space.moon.ganymede", "Ganymede"),
	"callisto":  vibe.E("ui.mapstyle.space.moon.callisto", "Callisto"),
	"enceladus": vibe.E("ui.mapstyle.space.moon.enceladus", "Enceladus"),
	"rhea":      vibe.E("ui.mapstyle.space.moon.rhea", "Rhea"),
	"titan":     vibe.E("ui.mapstyle.space.moon.titan", "Titan"),
	"iapetus":   vibe.E("ui.mapstyle.space.moon.iapetus", "Iapetus"),
	"miranda":   vibe.E("ui.mapstyle.space.moon.miranda", "Miranda"),
	"titania":   vibe.E("ui.mapstyle.space.moon.titania", "Titania"),
	"oberon":    vibe.E("ui.mapstyle.space.moon.oberon", "Oberon"),
	"triton":    vibe.E("ui.mapstyle.space.moon.triton", "Triton"),
	"charon":    vibe.E("ui.mapstyle.space.moon.charon", "Charon"),
}

// lockLines say which age opens a frame the player's age has not reached.
// Placeholders: the words are rewritten in the sheets.
var lockLines = map[string]vibe.Entry{
	world.AgeInterstellar: vibe.E("ui.mapstyle.space.lock.interstellar_age", "Opens in the Interstellar Age."),
	world.AgeGalactic:     vibe.E("ui.mapstyle.space.lock.galactic_age", "Opens in the Galactic Age."),
	world.AgeQuantum:      vibe.E("ui.mapstyle.space.lock.quantum_age", "Opens in the Quantum Age."),
}

// heldLines say whether the people hold a sector or can go to a star. Placeholders.
type hl struct{ zone, held bool }

var heldLines = map[hl]vibe.Entry{
	{true, true}:   vibe.E("ui.mapstyle.space.sector.held", "Held by your people."),
	{true, false}:  vibe.E("ui.mapstyle.space.sector.free", "Not yet reached by your people."),
	{false, true}:  vibe.E("ui.mapstyle.space.star.held", "Inside your territory."),
	{false, false}: vibe.E("ui.mapstyle.space.star.out", "Outside your reach."),
}

var (
	namesOnce sync.Once
	nameOf    map[string]vibe.Entry
)

// SkyName is how a label key of the world model reads, or "" for a key with no
// name. It is the host's lookup: the window draws these beside the objects.
func SkyName(key string) string {
	namesOnce.Do(func() {
		nameOf = map[string]vibe.Entry{
			"sol":           solName,
			"the_moon":      moonName,
			"satellite":     satName,
			"space_station": stnName,
			"pluto":         plutoName,
		}
		for k, e := range moonNames {
			nameOf[k] = e
		}
		for i, o := range mapgen.SkyObjects {
			nameOf[o.Key] = skyNames[i]
		}
		for i, s := range mapgen.NearStars {
			nameOf[s.Key] = starNames[i]
		}
		for i, k := range world.SolNames {
			nameOf[k] = planetNames[i]
		}
	})
	if e, ok := nameOf[key]; ok {
		return e.String()
	}
	if _, ok := mapmodel.LineageNames[key]; ok {
		return lineageLabel(key)
	}
	if i := strings.LastIndexByte(key, '_'); i > 0 && len(key)-i == 2 { // a generated planet or companion: its star's name and the letter
		if e, ok := nameOf[key[:i]]; ok {
			return e.String() + " " + key[i+1:]
		}
		if strings.HasPrefix(key, "~") {
			return key[1:i] + " " + key[i+1:]
		}
	}
	if strings.HasPrefix(key, "~") { // a name the seed made up carries its own text
		return key[1:]
	}
	return ""
}

// hostedZoom names the scale of a frame of the window's map with the labels
// the cell map already had: the star ages' rungs, or the ground ages' two.
func hostedZoom(star bool, k world.FrameKind) string {
	z := zCloseUp
	switch {
	case star && k == world.FrameLocalGroup:
		z = zCosmos
	case star && (k == world.FrameGalaxy || k == world.FrameRegion): // no rung of its own for the arm: it is still the galaxy's overview
		z = zOverview
	case star && k == world.FrameNeighbourhood:
		z = zInner
	case !star && k == world.FrameSystem:
		z = zOverview
	}
	return rungName(star, z)
}

// thingOf is the cell map's name for an object of the window's, so the
// inspector says the same words for it.
func thingOf(o *world.Object) mapgen.Thing {
	th := mapgen.Thing{Kind: mapgen.KindTerrain, Key: o.Inspect}
	for i, s := range mapgen.SkyObjects {
		if s.Key == o.Inspect {
			return mapgen.Thing{Kind: mapgen.KindTerrain, Key: "deepsky", Ref: i + 1}
		}
	}
	for i, s := range mapgen.NearStars {
		if s.Key == o.Label {
			return mapgen.Thing{Kind: mapgen.KindTerrain, Key: "neighbor", Ref: i + 1, Lineage: o.Line, Count: o.Count}
		}
	}
	for i, k := range world.SolNames {
		if k == o.Inspect {
			if k == world.SolNames[world.Home] {
				return mapgen.Thing{Kind: mapgen.KindTerrain, Key: "home"}
			}
			return mapgen.Thing{Kind: mapgen.KindTerrain, Key: "planet", Ref: i + 1}
		}
	}
	if o.Inspect == "sol" {
		th.Key = "sun"
	}
	return th
}

// hostedInspect says what the window's selected object is, in the cell map's
// words; ok is false while nothing is selected.
func (v *view) hostedInspect(m *mapmodel.Model) (Inspection, bool) {
	s, ok := v.hst.get()
	if !ok || s.Sel == nil {
		return Inspection{}, false
	}
	o := s.Sel
	if e, ok := lockLines[s.Lock]; ok { // a locked frame's object: its name and when it opens
		return Inspection{Title: SkyName(o.Label), Lines: []string{e.String()}}, true
	}
	if o.Inspect == world.SolNames[world.Home] { // the planet is the planet, whatever the age calls the town
		return describeHomeworld(m), true
	}
	if strings.HasPrefix(o.Label, "~") { // a sector or star the seed made up: its name, and whether the people hold it
		return Inspection{Title: SkyName(o.Label), Lines: []string{heldLines[hl{o.Form == world.FormZone, o.Held}].String()}}, true
	}
	switch o.Kind {
	case world.KindSector:
		return describeSector(m, mapgen.Thing{Kind: mapgen.KindSector, Key: o.Line, Lineage: o.Line, Count: o.Count}), true
	case world.KindRival:
		if civ := m.Faction(o.Inspect); civ != nil {
			return describeCiv(m, civ), true
		}
	}
	if in, ok := describeSpace(m, thingOf(o)); ok && in.Title != "" {
		return in, true
	}
	return Inspection{Title: SkyName(o.Label)}, true
}
