package citymap

import (
	"image"
	"image/color"
	"math"
	"sort"
	"sync"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// worldmap.go is the second map view — the "Known World." Where CityMap renders the
// player's own settlement (terrain + the buildings they've built), WorldMap zooms out
// to the wider world: a procedural, Game-of-Life-style field of anonymous settlements
// (the backdrop) with the player's civilization and the discovered diplomacy civs
// called out on top as distinct, labeled, relationship-colored dots.
//
// It deliberately reuses CityMap's render primitives rather than reinventing them:
//   - the half-block streamer (streamHalfBlocks) and the same (cols × rows*2) pixel
//     canvas / '▄' cell model,
//   - the noise field (fbm) for a muted terrain/ocean wash under the dots,
//   - the palette helpers (rgba/blend/brighten + theme roles) so every layer retints
//     on a theme switch,
//   - the overlay text-pass (overlayPlan / overlayLabel / stampOverlay): labels store
//     GEOMETRY + a theme ROLE, never a baked color, so a theme switch retints the text
//     without re-laying-out, exactly like CityMap.
//
// Lock discipline mirrors CityMap: Refresh only stores a snapshot, all rendering reads
// the stored snapshot under the mutex, nothing here touches the engine or acquires an
// engine lock. The diplomacy data is read straight off the snapshot's Diplomacy.Factions.
//
// A note on civ-dot SIZE: the design calls for sizing each civ by "Strength," but the
// UI projection game.FactionInfo carries no Strength field (that lives on the internal
// faction def, not the snapshot). We size from what the snapshot DOES expose — a
// strength proxy built from Opinion magnitude plus a standing weight (allies/at-war
// read as larger, more consequential neighbors) — so the dots still convey relative
// weight without reaching past the snapshot contract.

// WorldMap is the overlay widget renderer for the world view. Like CityMap it satisfies
// the OverlayManager widget contract via Build (returns the primitive) and Refresh
// (stores a fresh snapshot).
type WorldMap struct {
	mu    sync.Mutex
	state game.GameState

	// Cache of the last rendered image AND its overlay plan, plus the key that produced
	// them — same scheme as CityMap. The geometry depends on age/size/building-count/
	// discovered-civ-set (NOT the theme); the overlay's colors resolve live at draw
	// time, so a theme switch retints without invalidating the cached layout. We fold a
	// cheap signature of the discovered-civ set into the key so discovering/losing a civ
	// (or a civ's status flipping) re-lays-out the dots.
	cachedImg      *image.RGBA
	cachedPlan     overlayPlan
	cachedW        int
	cachedH        int
	cachedAge      string
	cachedBld      int
	cachedCivSig   uint64
	cachedThemeKey string
}

// NewWorldMap creates a WorldMap renderer.
func NewWorldMap() *WorldMap { return &WorldMap{} }

// Build returns the tview.Primitive that draws the world map. The draw closure reads
// the latest stored snapshot each frame, (re)generates the image on demand, streams it
// as half-blocks, then stamps the crisp text overlay on top — colors resolved live so a
// theme switch retints the labels too. Mirrors CityMap.Build.
func (wm *WorldMap) Build(state game.GameState) tview.Primitive {
	wm.mu.Lock()
	wm.state = state
	wm.mu.Unlock()

	box := tview.NewBox().SetBorder(false)
	box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		if width <= 0 || height <= 0 {
			return x, y, width, height
		}
		img, plan := wm.imageFor(width, height)
		streamHalfBlocks(screen, img, x, y, width, height)
		stampOverlay(screen, plan, x, y, width, height)
		return x, y, width, height
	})
	return box
}

// Refresh stores a fresh state snapshot. Safe from any goroutine; no rendering, no
// engine locks. Mirrors CityMap.Refresh.
func (wm *WorldMap) Refresh(state game.GameState) {
	wm.mu.Lock()
	wm.state = state
	wm.mu.Unlock()
}

// imageFor returns the rendered image AND its overlay plan for a (width,height) cell
// area, regenerating only when the cache key changes. width/height are in cells; the
// image is (width × height*2) pixels (one half-block = two stacked pixels). Mirrors
// CityMap.imageFor but adds the discovered-civ signature to the key.
func (wm *WorldMap) imageFor(width, height int) (*image.RGBA, overlayPlan) {
	imgW := width
	imgH := height * 2

	wm.mu.Lock()
	defer wm.mu.Unlock()

	st := wm.state
	bld := builtBuildingCount(st)
	civSig := civSignature(st)
	themeKey := theme.Active().Key

	if wm.cachedImg != nil &&
		wm.cachedW == imgW && wm.cachedH == imgH &&
		wm.cachedAge == st.Age && wm.cachedBld == bld &&
		wm.cachedCivSig == civSig && wm.cachedThemeKey == themeKey {
		return wm.cachedImg, wm.cachedPlan
	}

	img, plan := renderWorldImage(st, imgW, imgH)
	wm.cachedImg = img
	wm.cachedPlan = plan
	wm.cachedW = imgW
	wm.cachedH = imgH
	wm.cachedAge = st.Age
	wm.cachedBld = bld
	wm.cachedCivSig = civSig
	wm.cachedThemeKey = themeKey
	return img, plan
}

// civSignature folds the discovered-civ set (names + status + at-war + a strength
// proxy bucket) into a stable hash so the cache invalidates when the diplomacy picture
// changes — a new neighbor discovered, a status flip to war, etc. Pure read of the
// snapshot; map iteration order doesn't matter because we sort first.
func civSignature(state game.GameState) uint64 {
	cs := worldCivs(state)
	var h uint64 = 1469598103934665603 // FNV-1a 64 offset
	mix := func(s string) {
		for i := 0; i < len(s); i++ {
			h ^= uint64(s[i])
			h *= 1099511628211
		}
		h ^= 0xff // separator
		h *= 1099511628211
	}
	for _, c := range cs {
		mix(c.name)
		mix(string(rune('0' + c.sizeBucket)))
		if c.atWar {
			mix("war")
		}
		mix(string(rune('A' + int(c.role)%26)))
	}
	return h
}

// renderWorldImage composites the full world map and returns the pixel image plus the
// cell-space overlay plan. Layers, in order:
//
//	wash      — a muted terrain/ocean field (reused fbm) so the canvas isn't flat black,
//	            kept dim so the settlement dots read clearly over it.
//	backdrop  — the procedural Game-of-Life settlement field: sparse, clustered dots of
//	            varied sizes in a dim land tone; density + sizes scale with progress.
//	your civ  — one prominent accent dot at a stable spot (center).
//	civ dots  — each discovered diplomacy civ, ringed around your civ deterministically,
//	            sized by a strength proxy, colored by relationship, at-war marked.
//
// The overlay plan (your-civ label, civ labels, title) is then computed from the same
// geometry so labels land on the dots. Pure: reads the snapshot + config-by-key only.
func renderWorldImage(state game.GameState, w, h int) (*image.RGBA, overlayPlan) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	if w <= 0 || h <= 0 {
		return img, overlayPlan{}
	}

	seed, hueShift := ageInfo(state.Age)
	pal := buildPalette(hueShift)
	prog := worldProgress(state)

	// 1) Muted wash: reuse the fbm field for a faint land/ocean undertone so the dots
	//    sit on something, but blend it heavily toward the background so it never
	//    competes with the foreground dots. Theme-aware via the palette band colors.
	drawWorldWash(img, pal, seed)

	// 2) Backdrop: the Game-of-Life settlement field — the wider world.
	geo := drawSettlementField(img, pal, seed, prog)

	// 3) Your civ: a single prominent dot at the stable anchor (center).
	geo.you = drawYourCiv(img, pal, w, h)

	// 4) Discovered diplomacy civs: ringed around your civ, sized/colored by standing.
	geo.civs = drawWorldCivs(img, pal, state, seed, geo.you.cx, geo.you.cy)

	// 5) Overlay plan (cell space): cols=w, rows=h/2.
	plan := buildWorldOverlay(state, w, h/2, geo)
	return img, plan
}

// ---- progress / scaling -----------------------------------------------------

// worldProgress maps the player's advancement to a [0,1] scalar that drives backdrop
// density + dot sizes: an early world is sparse + small, a late world dense + large. It
// blends two signals — the age's index in the canonical age order (broad era progress)
// and the total built-building count (empire scale within an era) — so advancing an age
// OR sprawling a bigger empire both grow the world. Deterministic and pure.
func worldProgress(state game.GameState) float64 {
	order := config.AgeOrder()
	ageFrac := 0.0
	if n := len(order); n > 1 {
		idx := -1
		for i, k := range order {
			if k == state.Age {
				idx = i
				break
			}
		}
		if idx >= 0 {
			ageFrac = float64(idx) / float64(n-1)
		}
	}
	// Empire scale: total instances built across all distinct types, saturating so a
	// late hoard doesn't blow past 1. ~40 buildings reads as "fully built out."
	var total int
	for _, bs := range state.Buildings {
		if bs.Count > 0 {
			total += bs.Count
		}
	}
	bldFrac := math.Min(1.0, float64(total)/40.0)

	// Weight era progress a bit more than raw count — the age is the headline driver.
	p := 0.62*ageFrac + 0.38*bldFrac
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	return p
}

// ---- geometry carriers ------------------------------------------------------

// worldDot is a placed foreground dot (your civ or a neighbor) in PIXEL space, carried
// out of the pixel pass so the overlay can label it at the right cell.
type worldDot struct {
	cx, cy int    // dot center in pixels
	radius int    // dot radius in pixels (drives where the label sits)
	name   string // display name for the label
	role   theme.Role
	atWar  bool
}

// worldGeometry is the layout handed from the pixel pass to the overlay pass.
type worldGeometry struct {
	you  worldDot
	civs []worldDot
}

// ---- wash -------------------------------------------------------------------

// drawWorldWash paints a faint, muted land/ocean undertone from the same fbm field the
// city map uses, then pulls every pixel hard toward the background so the wash stays a
// quiet backdrop the dots read over. Theme-aware: the band colors come from the palette.
func drawWorldWash(img *image.RGBA, pal terrainPalette, seed uint32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return
	}
	bg := rgba(theme.Color(theme.RoleBackground))
	// A broad, low-frequency field so the wash reads as big continents/seas, not noise.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			e := fbm(float64(x), float64(y), seed)
			var band color.RGBA
			switch {
			case e < bandShallowWater:
				band = pal.bDeepWater
			case e < bandLowland:
				band = pal.bShallowWater
			default:
				band = pal.bGrass
			}
			// Mute hard: 78% background, 22% band — present but never loud.
			img.SetRGBA(b.Min.X+x, b.Min.Y+y, blend(bg, band, 0.22))
		}
	}
}

// ---- backdrop: Game-of-Life settlement field --------------------------------

// drawSettlementField paints the procedural backdrop: a sparse, clustered field of
// anonymous "settlement" dots of varied sizes — the wider world reading like a Conway/
// Game-of-Life cell field. Construction:
//
//  1. seed a coarse boolean grid with a deterministic noise + density threshold (density
//     scales up with progress: an early world is sparse, a late world dense),
//  2. run a couple of Conway generations so the live cells CLUSTER and gain organic
//     shape rather than reading as uniform static,
//  3. render each surviving cell as a dot whose RADIUS scales with both progress AND the
//     cell's local neighbour count (denser clusters → bigger settlements), in a DIM land
//     tone (RoleDim pulled a touch toward the land) so it's unmistakably background.
//
// Deterministic from seed+progress; theme-aware (the dim tone is a live theme role).
// Returns an (empty) worldGeometry the caller fills with the foreground dots — kept as
// the single geometry hand-off so future backdrop callouts can ride along.
func drawSettlementField(img *image.RGBA, pal terrainPalette, seed uint32, progress float64) worldGeometry {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return worldGeometry{}
	}

	// Coarse cell grid: each cell spans ~3px so the field is sparse and the dots have
	// room. Guard tiny canvases.
	gw := w / settlementCellPx
	gh := h / settlementCellPx
	if gw < 3 || gh < 3 {
		return worldGeometry{}
	}

	// Generate the Game-of-Life cell field (seed density scaled by progress, then a few
	// Conway generations to cluster it). Extracted so a test can count live cells across
	// progress levels without rasterizing through the wash.
	grid := settlementGrid(seed, gw, gh, progress)

	// Dim land tone for the backdrop dots: RoleDim nudged a touch toward the grass land
	// so the settlements read as faint inhabited specks, clearly behind the bright
	// foreground civ dots. Lives in the theme via RoleDim → retints on a switch.
	dim := rgba(theme.Color(theme.RoleDim))
	landDim := blend(dim, pal.bGrass, 0.30)

	// Render each live cell as a dot. Radius scales with progress and the local cluster
	// density: an isolated cell is a single pixel; a dense cluster center is a few px.
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			if !grid[gy*gw+gx] {
				continue
			}
			n := settlementNeighbours(grid, gw, gh, gx, gy)
			// Base radius grows with progress (0px..~1px); cluster bonus adds up to ~2px
			// for the densest centers. So late, clustered worlds get the big settlements.
			r := int(math.Round(progress*1.4)) + n/3
			if r < 0 {
				r = 0
			}
			if r > 3 {
				r = 3
			}
			cx := gx*settlementCellPx + settlementCellPx/2
			cy := gy*settlementCellPx + settlementCellPx/2
			// Vary the per-dot brightness slightly by a stable hash so the field doesn't
			// look stamped — still all in the dim land family.
			j := hashUnit(uint32(gx), uint32(gy), settlementSeed(seed)^0x5bd1e995)
			tone := blend(landDim, darken(landDim, 0.18), j)
			fillDot(img, cx, cy, r, tone)
		}
	}
	return worldGeometry{}
}

// settlementCellPx is the pixel span of one backdrop cell. ~3px keeps the field sparse
// with room between settlements.
const settlementCellPx = 3

// settlementSeed derives the backdrop's grid seed from the age seed with a 32-bit
// golden-ratio salt (the age seed is uint32, so no 64-bit constant). Stable per render.
func settlementSeed(ageSeed uint32) uint32 { return ageSeed ^ 0x9e3779b9 }

// settlementGrid builds the Game-of-Life cell field for a gw×gh coarse grid at a given
// progress: seed every cell live with a probability that climbs with progress (~12% at
// progress 0 → ~38% at progress 1), then run a few Conway B3/S23 generations so the
// live cells cluster into organic clumps rather than uniform speckle. Pure +
// deterministic from (seed, progress). Extracted from the draw path so a test can count
// live cells across progress levels without rasterizing through the wash.
func settlementGrid(seed uint32, gw, gh int, progress float64) []bool {
	if gw <= 0 || gh <= 0 {
		return nil
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	seedDensity := 0.12 + 0.26*progress
	gseed := settlementSeed(seed)

	grid := make([]bool, gw*gh)
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			if hashUnit(uint32(gx), uint32(gy), gseed) < seedDensity {
				grid[gy*gw+gx] = true
			}
		}
	}
	// Two Conway generations: enough to break the uniform speckle and grow recognizable
	// clusters without collapsing the whole field.
	for gen := 0; gen < 2; gen++ {
		next := make([]bool, gw*gh)
		for gy := 0; gy < gh; gy++ {
			for gx := 0; gx < gw; gx++ {
				n := settlementNeighbours(grid, gw, gh, gx, gy)
				alive := grid[gy*gw+gx]
				next[gy*gw+gx] = (alive && (n == 2 || n == 3)) || (!alive && n == 3)
			}
		}
		grid = next
	}
	return grid
}

// settlementNeighbours counts the live Moore-neighbours of cell (gx,gy) in a gw×gh grid
// (out-of-range neighbours are dead — the canvas edge bounds the world).
func settlementNeighbours(grid []bool, gw, gh, gx, gy int) int {
	n := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := gx+dx, gy+dy
			if nx < 0 || ny < 0 || nx >= gw || ny >= gh {
				continue
			}
			if grid[ny*gw+nx] {
				n++
			}
		}
	}
	return n
}

// countLiveCells returns the number of live cells in a settlement grid — a pure measure
// of backdrop density used by the tests to show density scales with progress.
func countLiveCells(grid []bool) int {
	n := 0
	for _, c := range grid {
		if c {
			n++
		}
	}
	return n
}

// ---- foreground: your civ ---------------------------------------------------

// drawYourCiv paints the player's settlement as a single PROMINENT dot — larger than any
// backdrop dot and in the bright accent role — at a stable anchor (canvas center). A
// soft accent halo underneath lifts it off the dim field. Returns the placed dot so the
// overlay can label it. Theme-aware: accent resolves from the live theme.
func drawYourCiv(img *image.RGBA, pal terrainPalette, w, h int) worldDot {
	cx, cy := w/2, h/2
	accent := rgba(theme.Color(theme.RoleAccent))

	// Radius scales gently with canvas size so it stays prominent on big and small maps,
	// clamped to a sensible band. Always at least 2px so it dominates the 0–3px backdrop.
	r := clampInt((w+h)/40, 2, 6)

	// Halo: a dim accent ring a little wider than the dot, so your capital glows.
	halo := blend(rgba(theme.Color(theme.RoleBackground)), accent, 0.35)
	fillDot(img, cx, cy, r+2, halo)
	// Body: bright accent, brightened a touch at the very center for a lit core.
	fillDot(img, cx, cy, r, accent)
	fillDot(img, cx, cy, r/2, brighten(accent, 0.30))

	return worldDot{cx: cx, cy: cy, radius: r + 2, role: theme.RoleAccent}
}

// ---- foreground: diplomacy civs ---------------------------------------------

// worldCiv is a discovered civ resolved for the world view: a display name, a status
// role, the at-war flag, and a size bucket (0..3) derived from a strength proxy.
type worldCiv struct {
	name       string
	role       theme.Role
	atWar      bool
	sizeBucket int
}

// worldCivs resolves the diplomacy snapshot to a stable, sorted list of discovered civs
// with status colors + size buckets. Sorted by name so the ring is stable frame-to-frame
// (map iteration is otherwise random and would make the dots jump). Pure read of the
// snapshot — no locks, no engine calls. Undiscovered civs are omitted entirely.
func worldCivs(state game.GameState) []worldCiv {
	facs := state.Diplomacy.Factions
	if len(facs) == 0 {
		return nil
	}
	keys := make([]string, 0, len(facs))
	for k, f := range facs {
		if f.Discovered {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	out := make([]worldCiv, 0, len(keys))
	for _, k := range keys {
		f := facs[k]
		out = append(out, worldCiv{
			name:       f.Name,
			role:       civWorldRole(f),
			atWar:      f.AtWar,
			sizeBucket: civStrengthBucket(f),
		})
	}
	return out
}

// civStrengthBucket derives a 0..3 dot-size bucket from the civ's real Strength — the
// 1-5 power rating on the diplomacy roster (FactionDef), now projected onto
// game.FactionInfo. Strength maps 1→0, 2→1, 3→2, 4→2, 5→3; a live war makes a neighbor
// loom one size larger. Pure.
func civStrengthBucket(f game.FactionInfo) int {
	s := f.Strength
	if s <= 0 {
		s = 1 // guard for an unset/legacy snapshot
	}
	b := int(math.Round(float64(s-1) * 3.0 / 4.0))
	if f.AtWar {
		b++
	}
	return clampInt(b, 0, 3)
}

// civWorldRole maps a faction's standing to a theme role for the world dot. It mirrors
// the city map's civStatusRole (allied→Positive, rival/embargo→Negative, at-war→bright
// Negative handled at draw, friendly→Label, neutral→Dim) so the two views agree on what
// a relationship looks like. Reused directly to keep one source of truth.
func civWorldRole(f game.FactionInfo) theme.Role { return civStatusRole(f) }

// drawWorldCivs places each discovered civ as a dot ringed around your civ at a stable,
// deterministic angle + radius (seeded so the same neighbor always sits in the same
// place), sized by its strength bucket, colored by relationship, and — when at war —
// brightened with a hot core so it stands out as a threat. Every civ dot is kept
// BRIGHTER than the dim backdrop so neighbors read as foreground. Returns the placed
// dots so the overlay can label each. Pure + theme-aware.
func drawWorldCivs(img *image.RGBA, pal terrainPalette, state game.GameState, seed uint32, yourX, yourY int) []worldDot {
	civs := worldCivs(state)
	if len(civs) == 0 {
		return nil
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	// Ring radius: keep neighbors well outside your civ's halo but inside the canvas, in
	// an elliptical band (the canvas is wider than tall in pixel terms only modestly, but
	// rows are half-height in cells, so we scale Y less to keep dots on-screen).
	baseRX := float64(w) * 0.34
	baseRY := float64(h) * 0.34

	out := make([]worldDot, 0, len(civs))
	for i, c := range civs {
		// Deterministic angle: spread civs around the circle by index, jittered by a
		// stable per-civ hash so they don't sit on a perfect mechanical clock.
		frac := float64(i) / float64(len(civs))
		jitter := (hashUnit(uint32(i), 0xC17A, seed) - 0.5) * 0.20
		ang := frac*2*math.Pi + jitter*2*math.Pi
		// Radius jittered per civ so the ring has organic depth, clamped into the band.
		rj := 0.82 + 0.30*hashUnit(uint32(i), 0x5EED, seed)
		px := float64(yourX) + math.Cos(ang)*baseRX*rj
		py := float64(yourY) + math.Sin(ang)*baseRY*rj

		// Dot radius from the strength bucket: bucket 0→2px .. bucket 3→5px. Always ≥2 so
		// neighbors out-mass the 0–3px backdrop and read as foreground actors.
		r := 2 + c.sizeBucket
		cx := clampInt(int(math.Round(px)), r, w-1-r)
		cy := clampInt(int(math.Round(py)), r, h-1-r)
		if w-1-r < r || h-1-r < r {
			// Canvas too small to inset; fall back to a clamped center-ish spot.
			cx = clampInt(int(math.Round(px)), 0, w-1)
			cy = clampInt(int(math.Round(py)), 0, h-1)
		}

		base := rgba(theme.Color(c.role))
		// Lift every neighbor above the dim backdrop: a faint background-blended halo,
		// then the body. At-war civs get a hot, brightened core so a threat pops.
		halo := blend(rgba(theme.Color(theme.RoleBackground)), base, 0.30)
		fillDot(img, cx, cy, r+1, halo)
		if c.atWar {
			fillDot(img, cx, cy, r, brighten(base, 0.25))
			fillDot(img, cx, cy, r/2, brighten(base, 0.55)) // white-hot core
		} else {
			fillDot(img, cx, cy, r, base)
		}

		out = append(out, worldDot{
			cx: cx, cy: cy, radius: r + 1,
			name: c.name, role: c.role, atWar: c.atWar,
		})
	}
	return out
}

// ---- dot primitive ----------------------------------------------------------

// fillDot paints a filled disc of radius r (pixels) centered at (cx,cy) into img with
// color col, clipped to the image bounds. r==0 paints a single pixel. Shared by the
// backdrop and both foreground passes so every dot uses one rasterizer.
func fillDot(img *image.RGBA, cx, cy, r int, col color.RGBA) {
	b := img.Bounds()
	if r <= 0 {
		if cx >= b.Min.X && cx < b.Max.X && cy >= b.Min.Y && cy < b.Max.Y {
			img.SetRGBA(cx, cy, col)
		}
		return
	}
	r2 := r * r
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy > r2 {
				continue
			}
			x, y := cx+dx, cy+dy
			if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
				continue
			}
			img.SetRGBA(x, y, col)
		}
	}
}

// ---- overlay ----------------------------------------------------------------

// buildWorldOverlay assembles the world map's text labels from the geometry: your civ's
// name under your dot, each discovered neighbor's name under its dot (at-war prefixed
// with a "⚔" mark), and the corner title. Pure (snapshot + geometry only). Labels store
// geometry + a theme ROLE; stampOverlay resolves colors live so a theme switch retints
// them — the same overlay contract CityMap uses.
func buildWorldOverlay(state game.GameState, cols, rows int, geo worldGeometry) overlayPlan {
	var plan overlayPlan
	if cols <= 0 || rows <= 0 {
		return plan
	}
	occupied := map[int]bool{}

	// 1) Your civ label, just beneath your dot, in the accent role (bold via labelTitle
	//    treatment? no — keep it a capital-kind so stampOverlay bolds it like a capital).
	plan.addYourCivLabel(state, geo, cols, rows, occupied)

	// 2) Neighbor labels, beneath each civ dot, colored by relationship, "⚔" if at war.
	plan.addWorldCivLabels(geo, cols, rows, occupied)

	// 3) Title last so it crowns its corner.
	plan.addWorldTitle(state, cols, rows)

	return plan
}

// addYourCivLabel stamps the player's civ name centered just below your dot. The name is
// AccountStats.DisplayName, falling back to "Your Empire" when accountless/unnamed.
// Accent role + capital kind so stampOverlay bolds it, marking it as the player's seat.
func (p *overlayPlan) addYourCivLabel(state game.GameState, geo worldGeometry, cols, rows int, occupied map[int]bool) {
	name := "Your Empire"
	if state.AccountStats != nil && state.AccountStats.DisplayName != "" {
		name = state.AccountStats.DisplayName
	}
	col := clampInt(pxToCellX(geo.you.cx), 0, cols-1)
	// One cell below the bottom edge of the dot (radius is in pixels → /2 for cells).
	row := clampInt(pxToCellY(geo.you.cy)+geo.you.radius/2+1, 0, rows-1)
	occupied[packCell(0, row)] = true
	text := truncLabel(name, maxLabelLen(col, cols, alignCenter))
	if text == "" {
		return
	}
	p.labels = append(p.labels, overlayLabel{
		cx: col, cy: row, text: text, role: theme.RoleAccent, kind: labelCapital, align: alignCenter,
	})
}

// addWorldCivLabels names each discovered neighbor beneath its dot, colored by its
// relationship role (resolved live at draw), with a "⚔" prefix when at war (and the
// bright treatment so the warring neighbor's label pops). Collision-limited per cell so
// two close dots don't overprint each other's names. Labels nudge a row if taken.
func (p *overlayPlan) addWorldCivLabels(geo worldGeometry, cols, rows int, occupied map[int]bool) {
	for _, d := range geo.civs {
		if d.name == "" {
			continue
		}
		col := clampInt(pxToCellX(d.cx), 0, cols-1)
		baseRow := pxToCellY(d.cy) + d.radius/2 + 1
		// Try the row under the dot, then a small spread so a crowded cluster keeps more
		// of its labels instead of dropping them.
		row := -1
		for _, dr := range []int{0, 1, -1, 2, -2} {
			cand := clampInt(baseRow+dr, 0, rows-1)
			if !occupied[packCell(0, cand)] {
				row = cand
				break
			}
		}
		if row < 0 {
			continue
		}
		occupied[packCell(0, row)] = true

		text := d.name
		if d.atWar {
			text = "⚔ " + text
		}
		text = truncLabel(text, maxLabelLen(col, cols, alignCenter))
		if text == "" {
			continue
		}
		p.labels = append(p.labels, overlayLabel{
			cx: col, cy: row, text: text, role: d.role, kind: labelCiv, bright: d.atWar, align: alignCenter,
		})
	}
}

// addWorldTitle stamps the corner title "Known World — <Age>" at the top-left in the
// accent role, drawn last so it crowns its corner. Uses the live age name (falling back
// to the age key). Mirrors the city map's title pattern.
func (p *overlayPlan) addWorldTitle(state game.GameState, cols, rows int) {
	age := state.AgeName
	if age == "" {
		age = state.Age
	}
	title := "Known World"
	if age != "" {
		title = "Known World — " + age
	}
	title = truncLabel(title, cols-1)
	if title == "" {
		return
	}
	p.labels = append(p.labels, overlayLabel{
		cx: 1, cy: 0, text: title, role: theme.RoleAccent, kind: labelTitle, align: alignLeft,
	})
}
