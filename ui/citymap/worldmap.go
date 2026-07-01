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
//	            kept dim so the settlement marks read clearly over it.
//	backdrop  — the sparse procedural settlement field: a scatter of small, dim distant
//	            marks in a dim land tone; count scales modestly with progress.
//	your civ  — one prominent accent dot at a stable spot (center).
//	civ dots  — each discovered diplomacy civ, ringed around your civ deterministically,
//	            sized by a strength proxy, colored by its own faction hue with a standing
//	            ring, and — at war — a hot-red override + ⚔ mark.
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

// drawSettlementField paints the procedural backdrop: a SPARSE scatter of small, dim
// "settlement" marks — the wider world reading as a few far-off settlements in a big dark
// world, not a snowstorm of noise. Construction:
//
//  1. build a sparse boolean grid via settlementGrid (a low density threshold that climbs
//     only modestly with progress, then one de-clumping pass — see settlementGrid),
//  2. render each live cell as a SMALL mark (mostly a single pixel; a stable minority a
//     touch wider), drawn wider-than-tall via fillDotAspect so it reads round on the 1×2
//     half-block canvas instead of a fat cross, in a DIM land tone (RoleDim pulled a
//     touch toward the land) so it's unmistakably background behind the bright civ dots.
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

	// Render each live cell as a SMALL, dim mark — a distant settlement, not a blob. Most
	// are a single pixel; a stable minority get a 2-wide × 0-tall (i.e. 3px across, 1px
	// tall) speck so the field has a little variety without any cell reading as a fat
	// cross. Marks are drawn WIDER than tall on purpose: the half-block canvas is 1px wide
	// × 2px tall per cell, so an equal-radius dot would stretch vertically into a "+"; a
	// horizontal bias renders round-ish. Dim land tone keeps them clearly in the backdrop.
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			if !grid[gy*gw+gx] {
				continue
			}
			cx := gx*settlementCellPx + settlementCellPx/2
			cy := gy*settlementCellPx + settlementCellPx/2
			// Stable per-cell hash drives both a slight brightness jitter (so the field
			// isn't stamped-uniform) and the occasional wider speck.
			j := hashUnit(uint32(gx), uint32(gy), settlementSeed(seed)^0x5bd1e995)
			tone := blend(landDim, darken(landDim, 0.22), j)
			// ~30% of settlements are a touch wider (still 1px tall); the rest single dots.
			// A faint extra size lift at high progress reads as "more established" worlds
			// without ever becoming a cross.
			rx := 0
			if j > 0.70 || (progress > 0.6 && j > 0.45) {
				rx = 1
			}
			fillDotAspect(img, cx, cy, rx, 0, tone)
		}
	}
	return worldGeometry{}
}

// settlementCellPx is the pixel span of one backdrop cell. Widened from 3→6 so the
// coarse grid has FEWER cells (a quarter as many) and the settlements read as a sparse
// scatter of distant marks with dark space between them, not a dense field of noise.
const settlementCellPx = 6

// settlementSeed derives the backdrop's grid seed from the age seed with a 32-bit
// golden-ratio salt (the age seed is uint32, so no 64-bit constant). Stable per render.
func settlementSeed(ageSeed uint32) uint32 { return ageSeed ^ 0x9e3779b9 }

// settlementGrid builds the backdrop settlement field for a gw×gh coarse grid at a given
// progress. The earlier version seeded a dense field (12–38%) and ran Conway B3/S23
// generations, which over-clustered into a solid snowstorm of noise. We now use a SPARSE
// Poisson-ish scatter: each cell is live with a low probability that climbs only modestly
// with progress (~4% at progress 0 → ~15% at progress 1), sampled from a spatially
// uncorrelated hash so the live cells land as scattered distant settlements rather than a
// packed field. A single de-clumping pass then clears any cell with 3+ live Moore
// neighbours, so the rare coincidental clusters break apart and the field stays "a few
// far-off settlements in a big dark world." Pure + deterministic from (seed, progress);
// density is monotonic in progress. Extracted from the draw path so a test can count live
// cells across progress levels without rasterizing through the wash.
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
	// Sparse base density: a low floor plus a gentle progress ramp. Kept well under the
	// old 12–38% so early game reads as a handful of settlements, late game as a few dozen.
	seedDensity := 0.04 + 0.11*progress
	gseed := settlementSeed(seed)

	grid := make([]bool, gw*gh)
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			if hashUnit(uint32(gx), uint32(gy), gseed) < seedDensity {
				grid[gy*gw+gx] = true
			}
		}
	}
	// One light de-clumping pass: drop any live cell hemmed in by 3+ live neighbours so
	// coincidental clusters thin out to a scatter (the opposite of Conway growth). This
	// preserves the sparse, isolated look while still scaling count with progress.
	next := make([]bool, gw*gh)
	copy(next, grid)
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			if grid[gy*gw+gx] && settlementNeighbours(grid, gw, gh, gx, gy) >= 3 {
				next[gy*gw+gx] = false
			}
		}
	}
	return next
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

	// Size scales gently with canvas so your seat stays prominent on big and small maps,
	// clamped to a sensible band and aspect-corrected (wider than tall) so it reads as a
	// round accent mark on the 1×2 half-block canvas rather than a fat cross. It stays the
	// biggest, brightest dot so it dominates the sparse backdrop and the neighbor dots.
	rx := clampInt((w+h)/28, 3, 7)
	ry := rx/2 + 1

	// Halo: a dim accent ring a little wider than the dot, so your capital glows.
	halo := blend(rgba(theme.Color(theme.RoleBackground)), accent, 0.35)
	fillDotAspect(img, cx, cy, rx+2, ry+1, halo)
	// Body: bright accent, brightened a touch at the very center for a lit core.
	fillDotAspect(img, cx, cy, rx, ry, accent)
	fillDotAspect(img, cx, cy, rx/2, ry/2, brighten(accent, 0.30))

	return worldDot{cx: cx, cy: cy, radius: ry + 1, role: theme.RoleAccent}
}

// ---- foreground: diplomacy civs ---------------------------------------------

// worldCiv is a discovered civ resolved for the world view: its faction KEY (drives the
// stable per-faction color), a display name, a relationship role (drives the label + the
// standing ring), the at-war flag, and a size bucket (0..3) derived from strength.
type worldCiv struct {
	key        string
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
			key:        k,
			name:       f.Name,
			role:       civWorldRole(f),
			atWar:      f.AtWar,
			sizeBucket: civStrengthBucket(f),
		})
	}
	return out
}

// factionColor returns a stable, distinct color for a discovered civ derived from its
// faction KEY — the same idea the city map's lineageColor uses for districts: hash the
// key to a hue and rotate a live theme role base by it, so every faction gets its own
// recognizable hue AND the whole set retints when the theme changes. RoleLabel is the
// base (a mid-chroma role that rotates cleanly across the wheel). This is the civ's
// IDENTITY color; relationship (ally/rival/at-war) is layered on top as a ring / hot
// override at draw time, so you can read both "who" and "how they feel" at a glance.
func factionColor(key string) color.RGBA {
	var hsh uint32 = 2166136261
	for i := 0; i < len(key); i++ {
		hsh ^= uint32(key[i])
		hsh *= 16777619
	}
	deg := float64(hsh % 360)
	return rotateHue(rgba(theme.Color(theme.RoleLabel)), deg)
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
// place). Each dot carries TWO channels of meaning:
//
//	identity  — the body is the faction's own stable color (factionColor, hashed from the
//	            key so every civ is a distinct, theme-retinting hue you can tell apart).
//	standing  — layered on top: an ally gets a bright Positive ring, a (non-war) rival a
//	            Negative ring, and an AT-WAR civ overrides the body entirely with a hot
//	            bright red + a white-hot core so a threat pops instantly (the ⚔ mark rides
//	            on its label).
//
// Dots are drawn WIDER than tall (fillDotAspect) so they read round on the half-block
// canvas instead of as a cross, and clearly LARGER + brighter than the now-sparse, dim
// backdrop so neighbors unmistakably read as foreground actors. The label role stays the
// relationship role so the overlay colors names by standing. Returns the placed dots.
// Pure + theme-aware.
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

	hotRed := brighten(rgba(theme.Color(theme.RoleNegative)), 0.20) // at-war override body

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

		// Dot size from the strength bucket, bumped so even the smallest neighbor clearly
		// out-masses the 0–1px backdrop. x/y radii are aspect-corrected: wider than tall so
		// the mark reads round on the 1×2 half-block canvas rather than as a vertical cross.
		rx := 3 + c.sizeBucket // 3px .. 6px across
		ry := 1 + c.sizeBucket/2
		if ry < 1 {
			ry = 1
		}
		inset := rx
		if ry > inset {
			inset = ry
		}
		cx := clampInt(int(math.Round(px)), inset, w-1-inset)
		cy := clampInt(int(math.Round(py)), inset, h-1-inset)
		if w-1-inset < inset || h-1-inset < inset {
			// Canvas too small to inset; fall back to a clamped center-ish spot.
			cx = clampInt(int(math.Round(px)), 0, w-1)
			cy = clampInt(int(math.Round(py)), 0, h-1)
		}

		ident := factionColor(c.key) // the civ's own identity hue
		standing := rgba(theme.Color(c.role))

		// Standing ring: a slightly larger disc of the relationship color under the body,
		// so an ally/rival reads its standing as a colored halo around its identity hue.
		// (Neutral/friendly rings are quiet — Dim/Label — so they don't shout.)
		ringCol := blend(rgba(theme.Color(theme.RoleBackground)), standing, 0.55)
		fillDotAspect(img, cx, cy, rx+1, ry+1, ringCol)

		if c.atWar {
			// War overrides identity: a hot bright-red body with a white-hot core, the loud
			// threat read. The ring underneath is already Negative (civStatusRole), so the
			// whole dot glows red.
			fillDotAspect(img, cx, cy, rx, ry, hotRed)
			fillDotAspect(img, cx, cy, rx/2, ry/2, brighten(hotRed, 0.45))
		} else {
			// Peace: the faction's own color, brightened a touch at the core for a lit
			// center so it sits clearly above the dim backdrop.
			fillDotAspect(img, cx, cy, rx, ry, brighten(ident, 0.10))
			fillDotAspect(img, cx, cy, rx/2, ry/2, brighten(ident, 0.35))
		}

		out = append(out, worldDot{
			cx: cx, cy: cy, radius: ry + 1,
			name: c.name, role: c.role, atWar: c.atWar,
		})
	}
	return out
}

// ---- dot primitive ----------------------------------------------------------

// fillDotAspect paints a filled ellipse with independent x/y pixel radii centered at
// (cx,cy), clipped to the image bounds. rx==ry==0 paints a single pixel. This is the
// aspect-correcting primitive the world map uses: the half-block canvas is 1px wide × 2px
// TALL per cell, so an equal-radius disc looks like a vertical "+"; passing rx≥ry (wider
// than tall) renders a mark that reads round-ish on screen. Used by the backdrop and the
// foreground civ dots so nothing renders as a fat cross.
func fillDotAspect(img *image.RGBA, cx, cy, rx, ry int, col color.RGBA) {
	b := img.Bounds()
	if rx <= 0 && ry <= 0 {
		if cx >= b.Min.X && cx < b.Max.X && cy >= b.Min.Y && cy < b.Max.Y {
			img.SetRGBA(cx, cy, col)
		}
		return
	}
	if rx < 0 {
		rx = 0
	}
	if ry < 0 {
		ry = 0
	}
	for dy := -ry; dy <= ry; dy++ {
		for dx := -rx; dx <= rx; dx++ {
			// Normalized ellipse test; guard the degenerate 0-radius axis (treat as a line).
			var ex, ey float64
			if rx > 0 {
				ex = float64(dx) / float64(rx)
			}
			if ry > 0 {
				ey = float64(dy) / float64(ry)
			}
			if ex*ex+ey*ey > 1.0 {
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
