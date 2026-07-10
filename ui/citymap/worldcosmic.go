package citymap

import (
	"image"
	"image/color"
	"math"

	"github.com/espresso20/ageforge/game"
)

// worldcosmic.go is Phase C of the world map: the COSMIC STRATEGIC VIEWS. Phases A/B built
// ONE seeded continent (worldmodel.go) and re-skinned it per age through a cartographic
// MEDIUM (worldmedium.go). For the space-and-beyond ages there IS no continent — the World
// tab should show a STRATEGIC star-map instead: YOUR empire vs the rival diplomacy factions
// competing for control of the local cluster.
//
// This is deliberately DIFFERENT from the Map-tab "cosmic scenes" (drawPlanetScene /
// drawGalaxyScene / …), which are SCENIC (a planet from orbit, a spiral galaxy). Those show
// where you ARE; this shows who you're up AGAINST. The World tab is STRATEGY: nodes (empires),
// lanes (trade / conflict / contact), spheres of influence (territory), and HUD chrome — no
// terrain, no continent.
//
// Dispatch mirrors topdown.go's cosmicSceneFor: cosmicWorldViewFor(ageKey) returns a
// (drawFn, ok) pair. renderWorldImage calls it FIRST — if a cosmic strategic view owns the
// age, that view paints the ENTIRE frame (background + territory + lanes + nodes + chrome)
// and renderWorldImage returns immediately, skipping BOTH the medium terrain draw AND the
// shared civ-dot / label overlay (the strategic nodes ARE the factions; drawing dots too
// would double-render them). This first Phase-C pass wires ONLY space_age as a
// proof-of-concept; the four higher cosmic ages (interstellar/galactic/quantum/transcendent)
// still fall through to the neutral atlas until a later phase.
//
// Determinism + safety, exactly like the mediums: a strategic view is a pure function of
// (state, size, seed). Faction node placement is seeded off a hash of the faction KEY (never
// map iteration order, never math/rand) so the same neighbour always sits in the same spot.
// Every write clips to the canvas (setPixel / bounds-checked helpers), so tiny / odd / zero
// canvases never panic; an empty roster still draws the home hub in the starfield. The view
// is LABEL-FREE — everything reads through color / size / glyph, like the Map-tab scenes.

// cosmicWorldViewFor returns the strategic star-map renderer for a cosmic age, mirroring the
// (func, bool) intercept shape of topdown.go's cosmicSceneFor. The signature matches what
// renderWorldImage already has in scope: (img, state, w, h in PIXELS, the uint32 age seed).
// Only space_age is wired in this proof-of-concept pass; every other age (including the four
// higher cosmic ages 19–22) returns (nil, false) so it keeps falling through to the medium
// path (neutral atlas for now).
func cosmicWorldViewFor(ageKey string) (func(img *image.RGBA, state game.GameState, w, h int, seed uint32), bool) {
	switch ageKey {
	case "space_age":
		return drawSpaceStrategicView, true
	}
	return nil, false
}

// ---- cosmic strategic palette -----------------------------------------------

// cosmicStratPalette is the fixed tone set the strategic views paint with. Like every
// cartographic medium it anchors HARD to its own identity — deep-indigo void, cyan command
// hues, relationship signal colors — independent of the active theme, because the whole point
// is that the cosmic ages read as a distinct STRATEGIC display, not a tinted atlas. Carried in
// one struct so future ages (19–22) can reuse the same signal vocabulary with a different frame.
type cosmicStratPalette struct {
	voidDeep  color.RGBA // near-black deep-space fill
	nebulaA   color.RGBA // low-freq nebula wash tone A (teal)
	nebulaB   color.RGBA // low-freq nebula wash tone B (violet)
	starWarm  color.RGBA // faint warm/white star
	starCool  color.RGBA // faint blue star
	home      color.RGBA // your empire — cyan/blue command hue
	homeCore  color.RGBA // your empire hot core (near-white cyan)
	ally      color.RGBA // allied / high-opinion — green
	merc      color.RGBA // mercantile — gold
	neutral   color.RGBA // neutral / isolationist — steel-blue
	war       color.RGBA // at-war — hot red
	laneTrade color.RGBA // trade lane / flow pips
	laneWar   color.RGBA // conflict lane (dashed red)
	laneDim   color.RGBA // faint contact line (grey)
	chrome    color.RGBA // HUD brackets + range rings (crisp cyan)
	fog       color.RGBA // undiscovered fog blip
}

// cosmicStratPal builds the fixed strategic palette. Theme-independent by design.
func cosmicStratPal() cosmicStratPalette {
	return cosmicStratPalette{
		voidDeep:  color.RGBA{R: 0x05, G: 0x06, B: 0x0d, A: 0xff}, // deep indigo
		nebulaA:   color.RGBA{R: 0x18, G: 0x3a, B: 0x40, A: 0xff}, // dim teal
		nebulaB:   color.RGBA{R: 0x2a, G: 0x1e, B: 0x44, A: 0xff}, // dim violet
		starWarm:  color.RGBA{R: 0xe8, G: 0xe6, B: 0xe0, A: 0xff},
		starCool:  color.RGBA{R: 0xb8, G: 0xcf, B: 0xff, A: 0xff},
		home:      color.RGBA{R: 0x36, G: 0xc9, B: 0xe6, A: 0xff}, // command cyan
		homeCore:  color.RGBA{R: 0xe6, G: 0xfb, B: 0xff, A: 0xff}, // hot white-cyan core
		ally:      color.RGBA{R: 0x4c, G: 0xd6, B: 0x7a, A: 0xff}, // green
		merc:      color.RGBA{R: 0xe6, G: 0xbf, B: 0x3d, A: 0xff}, // gold
		neutral:   color.RGBA{R: 0x6f, G: 0x8a, B: 0xb0, A: 0xff}, // steel-blue
		war:       color.RGBA{R: 0xe8, G: 0x3f, B: 0x3f, A: 0xff}, // hot red
		laneTrade: color.RGBA{R: 0x38, G: 0xb8, B: 0xcf, A: 0xff}, // cyan trade lane
		laneWar:   color.RGBA{R: 0xd6, G: 0x3a, B: 0x3a, A: 0xff}, // red conflict lane
		laneDim:   color.RGBA{R: 0x53, G: 0x5c, B: 0x6e, A: 0xff}, // grey contact line
		chrome:    color.RGBA{R: 0x2f, G: 0xa8, B: 0xc4, A: 0xff}, // crisp cyan chrome
		fog:       color.RGBA{R: 0x3a, G: 0x44, B: 0x58, A: 0xff}, // faint fog blip
	}
}

// ---- empire layout (reusable) -----------------------------------------------

// empireRole tags a node's color-role so the reusable layout stays visual-framing-agnostic:
// the layout decides WHO sits WHERE and their relationship; each age's strategic view maps the
// role to its own palette. Ages 19–22 can reuse strategicEmpires and re-skin these roles.
type empireRole uint8

const (
	empireHome    empireRole = iota // the player's seat
	empireAlly                      // allied / high-opinion
	empireMerc                      // mercantile personality
	empireNeutral                   // neutral / isolationist / everything else
	empireWar                       // actively at war (overrides all others)
)

// empireNode is one placed actor in the strategic layout, in PIXEL space, carrying enough
// signal for any cosmic-age view to draw it: position, a size scalar (1–5 strength), the
// color-role, and the two lane flags (at-war / has-trade). The home node is flagged so a view
// can render it as the prominent command hub regardless of role.
type empireNode struct {
	key      string     // faction key ("" for home) — drives identity hue + placement seed
	cx, cy   int        // node center in pixels
	strength int        // 1–5 power rating (drives node + influence-bubble size)
	role     empireRole // color-role (home / ally / merc / neutral / war)
	atWar    bool       // conflict lane + alert ring
	trade    bool       // TradeCount>0 → trade lane
	isHome   bool       // the player's seat (drawn as the command hub)
}

// strategicEmpires is the REUSABLE layout helper: it turns the diplomacy snapshot into the
// list of placed empire nodes the strategic views draw. The home node sits at (near) the
// canvas center; each DISCOVERED faction is placed at a DETERMINISTIC position — a golden-angle
// spiral seeded off a hash of the faction key for the angle, with a seeded/jittered radius —
// clamped inside a frame margin and pushed off the center so nodes don't stack on the hub.
// UNDISCOVERED factions are NOT placed here (they render as edge fog blips separately). Pure +
// deterministic from (state, size, seed); ages 19–22 reuse this and only change the visual
// framing. Returns the home node first, then discovered factions in stable key order.
//
// Signature (for the 19–22 follow-up):
//
//	strategicEmpires(state game.GameState, w, h int, seed uint32) []empireNode
func strategicEmpires(state game.GameState, w, h int, seed uint32) []empireNode {
	if w <= 0 || h <= 0 {
		return nil
	}

	// Home hub at the canvas center (nudged a hair up so lanes fan below it too). This is the
	// seat of power; every lane and range ring centers here.
	homeX := w / 2
	homeY := h / 2
	nodes := []empireNode{{
		cx: clampInt(homeX, 0, w-1), cy: clampInt(homeY, 0, h-1),
		strength: 5, role: empireHome, isHome: true,
	}}

	civs := worldCivs(state) // discovered-only, stable key order (shared with the medium path)
	if len(civs) == 0 {
		return nodes
	}

	// Placement band: keep faction nodes well outside the hub's halo but inside a frame margin
	// so nothing clips the chrome brackets. Rows are half-height on the terminal canvas, so the
	// vertical band is scaled a touch tighter than the horizontal to keep nodes on-screen.
	margin := clampInt(minInt(w, h)/8, 4, 40)
	maxRX := float64(w)/2 - float64(margin)
	maxRY := float64(h)/2 - float64(margin)
	if maxRX < 2 {
		maxRX = 2
	}
	if maxRY < 2 {
		maxRY = 2
	}
	// Inner ring so no node lands on the hub. Golden angle spreads them evenly around the sweep;
	// pushed well off the hub (innerFrac) so the roster claims the frame instead of hugging center.
	const golden = 2.399963229728653 // 2π / φ²
	innerFrac := 0.52
	// ONE global rotation seeded off the world orients the whole ring deterministically while the
	// golden angle keeps nodes EVENLY spaced. (A per-key random angle — the old approach — let
	// neighbours cluster; the golden sweep is what actually spreads them.)
	globalRot := hashUnit(seed, 0xC05, 0x1234) * 2 * math.Pi

	for i, c := range civs {
		keySeed := factionKeySeed(c.key)
		ang := float64(i)*golden + globalRot
		// Radius: seeded jitter within the OUTER band (pushed off the hub for spread), biased
		// outward with index so a crowded roster spirals out instead of piling on one ring.
		rj := innerFrac + (1.0-innerFrac)*(0.45+0.5*hashUnit(keySeed, 0x5EED, seed))
		spiralOut := 0.08 * (float64(i) / float64(maxInt(len(civs), 1)))
		rr := clamp01(rj + spiralOut)

		px := float64(homeX) + math.Cos(ang)*maxRX*rr
		py := float64(homeY) + math.Sin(ang)*maxRY*rr
		cx := clampInt(int(math.Round(px)), margin, maxInt(w-1-margin, margin))
		cy := clampInt(int(math.Round(py)), margin, maxInt(h-1-margin, margin))

		f := state.Diplomacy.Factions[c.key]
		nodes = append(nodes, empireNode{
			key:      c.key,
			cx:       cx,
			cy:       cy,
			strength: factionStrength(f),
			role:     empireRoleFor(f),
			atWar:    f.AtWar,
			trade:    f.TradeCount > 0,
		})
	}
	return nodes
}

// factionKeySeed hashes a faction key to a stable uint32 (FNV-1a) — the placement + jitter
// seed so each faction always lands in the same relative spot.
func factionKeySeed(key string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= 16777619
	}
	return h
}

// factionStrength returns the civ's 1–5 power rating, guarding an unset/legacy snapshot.
func factionStrength(f game.FactionInfo) int {
	s := f.Strength
	if s <= 0 {
		s = 1
	}
	return clampInt(s, 1, 5)
}

// empireRoleFor maps a faction's standing to a strategic color-role. At-war overrides
// everything (a threat reads red first). Otherwise: allied/high-opinion → ally (green),
// mercantile personality → merc (gold), and isolationist/neutral/anything else → neutral
// (steel-blue). Pure read of the snapshot fields.
func empireRoleFor(f game.FactionInfo) empireRole {
	if f.AtWar {
		return empireWar
	}
	if f.Status == "allied" || f.Opinion >= 60 {
		return empireAlly
	}
	if f.Personality == "mercantile" || f.TradeCount > 0 {
		return empireMerc
	}
	return empireNeutral
}

// nodeColor resolves an empire node's role to its palette hue, folding in the faction's own
// identity tint for non-home nodes so two neutral neighbours are still tellable apart (a small
// nudge toward factionColor, which hashes the key to a hue). The home hub is pure command cyan.
func nodeColor(n empireNode, p cosmicStratPalette) color.RGBA {
	var base color.RGBA
	switch n.role {
	case empireHome:
		return p.home
	case empireWar:
		base = p.war
	case empireAlly:
		base = p.ally
	case empireMerc:
		base = p.merc
	default:
		base = p.neutral
	}
	// A whisper of the faction's identity hue so same-role neighbours don't read identical,
	// while the relationship color still dominates.
	if n.key != "" {
		base = blend(base, factionColor(n.key), 0.16)
	}
	return base
}

// ---- space_age strategic view -----------------------------------------------

// drawSpaceStrategicView paints the space_age World tab as a HOME CLUSTER COMMAND OVERLAY: a
// strategic star-map of your empire vs the discovered rival factions. Layers back-to-front:
//
//  1. deep-space field — near-black indigo fill, a deterministic seeded starfield (faint
//     white/blue stars of varied brightness), and a subtle low-freq nebula wash for depth.
//  2. spheres of influence — a soft translucent additive bubble per empire, tinted by owner
//     (yours cyan + large; ally green; merc gold; neutral steel-blue; at-war red), sized by
//     strength; drawn BEHIND the nodes so overlaps read as contested space.
//  3. lanes — from your hub to each discovered faction: a solid cyan/gold trade lane with flow
//     pips when TradeCount>0; a red DASHED conflict lane with a crossed midpoint glyph when at
//     war; a faint dotted grey contact line otherwise.
//  4. empire nodes — your hub: a bright multi-ring cyan-white command node at center, the
//     largest, unmistakably the seat of power; each faction: a node at its deterministic spot,
//     sized by strength, colored by relationship, with a crisp thin ring (at-war gets a broken
//     red alert ring). Undiscovered factions become faint "?" fog blips near the frame edge.
//  5. HUD chrome — crisp cyan corner brackets + a couple of faint concentric range rings on the
//     hub (a tactical reticle).
//
// LABEL-FREE, deterministic (hash/seed only), panic-safe (all writes bounds-checked; a tiny /
// odd / zero canvas returns after the fill). Empty roster → the hub alone in the starfield,
// still reading as a strategic command view.
func drawSpaceStrategicView(img *image.RGBA, state game.GameState, w, h int, seed uint32) {
	b := img.Bounds()
	if w <= 0 || h <= 0 {
		return
	}
	p := cosmicStratPal()

	// 1) Deep-space field: fill + nebula wash + starfield.
	drawCosmicField(img, p, seed)

	// Layout the empire nodes ONCE and thread them through every layer.
	nodes := strategicEmpires(state, w, h, seed)
	if len(nodes) == 0 {
		// Degenerate canvas produced no nodes (already guarded), but keep the field.
		return
	}
	home := nodes[0]

	// 2) Spheres of influence (behind nodes): additive bubbles so overlaps brighten as contested.
	for _, n := range nodes {
		drawInfluenceBubble(img, n, p)
	}

	// 3) Lanes from the hub to each faction.
	for _, n := range nodes {
		if n.isHome {
			continue
		}
		drawStrategicLane(img, home, n, p, seed)
	}

	// 4a) Undiscovered factions → faint fog blips near the frame edge (only if any exist).
	drawUndiscoveredFog(img, state, home, p, seed)

	// 4b) Empire nodes on top: faction nodes first, the command hub LAST so it crowns the map.
	for _, n := range nodes {
		if n.isHome {
			continue
		}
		drawFactionNode(img, n, p)
	}
	drawCommandHub(img, home, p)

	// 5) HUD chrome: corner brackets + range rings on the hub.
	drawStrategicChrome(img, home, p)

	_ = b
}

// drawCosmicField fills the deep-space background: a near-black indigo base, a subtle low-freq
// nebula wash (teal↔violet) for depth, then a deterministic starfield of faint white/blue stars
// with varied brightness. All hash-driven (no math/rand). Panic-safe on any canvas.
func drawCosmicField(img *image.RGBA, p cosmicStratPalette, seed uint32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return
	}

	// Base + nebula: a broad low-freq value-noise field blends a dim teal/violet wash into the
	// void so the black has depth without reading busy. Kept very subtle (max ~18% mix).
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n := valueNoise(float64(x)*0.020, float64(y)*0.020, seed^0x4E8)
			// Which nebula tone: a second broad field picks teal vs violet regionally.
			t := valueNoise(float64(x)*0.014+40, float64(y)*0.014+40, seed^0x9B)
			neb := blend(p.nebulaA, p.nebulaB, clamp01(t))
			// Only the brighter lobes of the noise show at all, so most of the frame stays void.
			mix := 0.0
			if n > 0.55 {
				mix = (n - 0.55) / 0.45 * 0.18
			}
			setPixel(img, x, y, blend(p.voidDeep, neb, clamp01(mix)))
		}
	}

	// Starfield: a sparse hash-placed scatter. A coarse cell grid keeps stars from clumping; a
	// per-cell hash decides IF a star lands and its sub-cell offset + brightness + hue. Faint —
	// this is a backdrop, not a light show.
	cell := 5
	gw := w / cell
	gh := h / cell
	for gy := 0; gy <= gh; gy++ {
		for gx := 0; gx <= gw; gx++ {
			hsel := hashUnit(uint32(gx), uint32(gy), seed^0x57A2)
			if hsel > 0.34 { // ~1/3 of cells carry a star
				continue
			}
			ox := int(hashUnit(uint32(gx), uint32(gy), seed^0x11) * float64(cell))
			oy := int(hashUnit(uint32(gy), uint32(gx), seed^0x22) * float64(cell))
			sx := gx*cell + ox
			sy := gy*cell + oy
			bright := 0.30 + 0.70*hashUnit(uint32(gx), uint32(gy), seed^0x33)
			star := p.starWarm
			if hashUnit(uint32(gx), uint32(gy), seed^0x44) > 0.6 {
				star = p.starCool
			}
			// Dim the star toward the void by (1-bright) so brightness genuinely varies.
			col := blend(p.voidDeep, star, bright)
			setPixel(img, sx, sy, col)
			// A rare bright star gets a 1px cross twinkle (kept crisp, no soft halo).
			if bright > 0.90 {
				twin := blend(p.voidDeep, star, bright*0.55)
				setPixel(img, sx-1, sy, twin)
				setPixel(img, sx+1, sy, twin)
				setPixel(img, sx, sy-1, twin)
				setPixel(img, sx, sy+1, twin)
			}
		}
	}
}

// drawInfluenceBubble paints a soft translucent sphere of influence for one empire: an additive
// radial disc, low alpha, tinted by the node's role, sized by strength (the home hub reads
// LARGE). Additive (blendPixel accumulates toward the tint) so overlapping bubbles brighten —
// contested space reads as a hot seam. Falls off smoothly from center to a crisp-ish edge (no
// muddy wide halo): the alpha ramps to zero by the radius so it stays a controlled glow.
func drawInfluenceBubble(img *image.RGBA, n empireNode, p cosmicStratPalette) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return
	}
	tint := nodeColor(n, p)
	// Radius scales with strength; the home hub gets a large dominant field. Tie to canvas so
	// it stays proportional across sizes. rY is squashed for the half-block aspect.
	base := float64(minInt(w, h))
	var r float64
	if n.isHome {
		r = base * 0.34
	} else {
		r = base * (0.075 + 0.032*float64(n.strength)) // ~0.11→0.24 of the short side
	}
	if r < 2 {
		r = 2
	}
	rY := r * 0.62 // half-block squash so the field reads round on screen
	// Peak alpha: bold enough that spheres of influence read as claimed TERRITORY, not a faint
	// wash; the home field strongest. Additive, so overlaps still brighten as contested seams.
	peak := 0.26
	if n.isHome {
		peak = 0.34
	}
	x0 := clampInt(n.cx-int(r)-1, 0, w-1)
	x1 := clampInt(n.cx+int(r)+1, 0, w-1)
	y0 := clampInt(n.cy-int(rY)-1, 0, h-1)
	y1 := clampInt(n.cy+int(rY)+1, 0, h-1)
	invR := 1.0 / r
	invRY := 1.0 / rY
	for y := y0; y <= y1; y++ {
		dy := float64(y-n.cy) * invRY
		for x := x0; x <= x1; x++ {
			dx := float64(x-n.cx) * invR
			d := dx*dx + dy*dy
			if d > 1.0 {
				continue
			}
			// Smooth falloff to zero at the rim: (1-d)^1.3 fills more of the disc so the territory
			// reads bold, while still ramping to zero at the edge (no hard-cut, no muddy halo).
			fall := math.Pow(1.0-d, 1.3)
			blendPixel(img, x, y, tint, peak*fall)
		}
	}
}

// drawStrategicLane strokes a relationship lane from the home hub to a faction node:
//
//	trade (TradeCount>0) — a solid soft cyan/gold line (gold if the node is mercantile) with a
//	                       few evenly-spaced brighter flow PIPS along it (goods moving).
//	at-war               — a red DASHED conflict lane, plus a small crossed/crosshair glyph at
//	                       the midpoint (a contested vector).
//	other discovered     — a faint dotted grey contact line (we know they're there).
//
// All strokes go through a bounds-checked put; the dashing/pips are index-gated so they stay
// deterministic. Trade takes precedence over the plain contact line; at-war always draws its
// conflict lane (a warring trade partner still reads as a threat).
func drawStrategicLane(img *image.RGBA, home, n empireNode, p cosmicStratPalette, seed uint32) {
	if n.atWar {
		strokeDashedLine(img, home.cx, home.cy, n.cx, n.cy, p.laneWar, 3, 2)
		drawCrossGlyph(img, (home.cx+n.cx)/2, (home.cy+n.cy)/2, p.laneWar)
		return
	}
	if n.trade {
		laneCol := p.laneTrade
		if n.role == empireMerc {
			laneCol = blend(p.laneTrade, p.merc, 0.6)
		}
		strokeSolidFaintLine(img, home.cx, home.cy, n.cx, n.cy, laneCol, 0.55)
		drawFlowPips(img, home.cx, home.cy, n.cx, n.cy, brighten(laneCol, 0.25), seed)
		return
	}
	// Plain discovered contact: a faint dotted grey line.
	strokeDashedLine(img, home.cx, home.cy, n.cx, n.cy, p.laneDim, 1, 3)
}

// strokeSolidFaintLine draws a solid line blended into the background at alpha t (a soft lane
// that doesn't overpower the nodes). Reuses the shared Bresenham engine via strokeThickLineFunc
// with a bounds-checked blend put.
func strokeSolidFaintLine(img *image.RGBA, x0, y0, x1, y1 int, col color.RGBA, t float64) {
	put := func(x, y int, c color.RGBA) { blendPixel(img, x, y, c, t) }
	strokeThickLineFunc(x0, y0, x1, y1, 0, col, put)
}

// strokeDashedLine draws a dashed line: `on` pixels drawn, then `off` pixels skipped, repeating.
// The dash counter walks with the Bresenham steps so the pattern is even along the line.
// Bounds-checked. Used for the conflict lane (bold dashes) and the faint contact line (fine dots).
func strokeDashedLine(img *image.RGBA, x0, y0, x1, y1 int, col color.RGBA, on, off int) {
	if on < 1 {
		on = 1
	}
	if off < 0 {
		off = 0
	}
	period := on + off
	step := 0
	put := func(x, y int, c color.RGBA) {
		if step%period < on {
			setPixel(img, x, y, c)
		}
		step++
	}
	strokeThickLineFunc(x0, y0, x1, y1, 0, col, put)
}

// drawFlowPips stamps a few evenly-spaced bright pips along a trade lane so goods read as
// moving. Positions are fractions of the segment; a seed nudge offsets the phase so different
// lanes don't pulse in lockstep. Crisp single pixels (no glow).
func drawFlowPips(img *image.RGBA, x0, y0, x1, y1 int, col color.RGBA, seed uint32) {
	const pips = 4
	phase := hashUnit(uint32(x0^y1), uint32(y0^x1), seed) // stable per-lane phase
	for i := 0; i < pips; i++ {
		f := (float64(i) + 0.5 + phase) / float64(pips)
		if f >= 1.0 {
			f -= 1.0
		}
		px := int(math.Round(float64(x0) + (float64(x1-x0))*f))
		py := int(math.Round(float64(y0) + (float64(y1-y0))*f))
		setPixel(img, px, py, col)
	}
}

// drawCrossGlyph stamps a small crossed / crosshair mark (a contested vector) at (cx,cy): a
// short + overlaid with an × so it reads as a conflict flashpoint. Crisp, bounds-checked.
func drawCrossGlyph(img *image.RGBA, cx, cy int, col color.RGBA) {
	setPixel(img, cx, cy, col)
	for k := 1; k <= 2; k++ {
		setPixel(img, cx+k, cy, col)
		setPixel(img, cx-k, cy, col)
		setPixel(img, cx, cy+k, col)
		setPixel(img, cx, cy-k, col)
		setPixel(img, cx+k, cy+k, col)
		setPixel(img, cx-k, cy-k, col)
		setPixel(img, cx+k, cy-k, col)
		setPixel(img, cx-k, cy+k, col)
	}
}

// drawFactionNode paints one discovered faction as a strategic node: a filled disc in the
// relationship color sized by strength, with a crisp thin ring one pixel proud of the body so
// it reads as a marked contact (not a blob). An AT-WAR node gets a BROKEN red alert ring
// (dashed around the circumference) instead of the solid ring, so a threat reads instantly.
// The body carries a slightly brighter core (a touch of brighten, no inner disc → no "+"
// artifact). Crisp edges, no soft halo — the influence bubble already carries the glow.
func drawFactionNode(img *image.RGBA, n empireNode, p cosmicStratPalette) {
	col := nodeColor(n, p)
	// Radius by strength: 1→2px … 5→4px. Kept clearly smaller than the command hub (radius 5–7)
	// so your seat always dominates.
	r := 2 + n.strength/2
	r = clampInt(r, 2, 4)
	// Body: a filled disc, brightened a hair so it sits above the field.
	fillDot(img, n.cx, n.cy, r, brighten(col, 0.10))
	if n.atWar {
		// Broken red alert ring — dashed around the circumference.
		drawRingDashed(img, n.cx, n.cy, r+1, p.war, 2, 2)
	} else {
		// Crisp thin ring: the relationship color, one pixel proud of the body.
		drawRing(img, n.cx, n.cy, r+1, brighten(col, 0.18))
	}
}

// drawCommandHub paints the player's seat: a bright, prominent MULTI-RING cyan-white hub at the
// canvas center, the LARGEST and brightest mark so the seat of power is unmistakable. A
// hot-white core disc, a command-cyan body, and two crisp concentric rings proud of it (the
// "command" read). No soft halo — the influence bubble under it carries the glow; the hub
// itself is crisp.
func drawCommandHub(img *image.RGBA, home empireNode, p cosmicStratPalette) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	// Hub radius scales with canvas, clamped so it dominates the map as the seat of power.
	r := clampInt((w+h)/32, 6, 10)
	// Body: command cyan.
	fillDot(img, home.cx, home.cy, r, p.home)
	// Hot core: a smaller near-white disc. At these radii a same-center bright disc reads as a
	// lit core without a cross (the body underneath is a full disc, so no gap rasterizes a "+").
	coreR := clampInt(r/2, 1, 3)
	fillDot(img, home.cx, home.cy, coreR, p.homeCore)
	// Two crisp command rings proud of the body.
	drawRing(img, home.cx, home.cy, r+2, brighten(p.home, 0.20))
	drawRing(img, home.cx, home.cy, r+4, blend(p.voidDeep, p.home, 0.7))
}

// drawUndiscoveredFog stamps faint "?"-style fog blips near the FRAME EDGE for any undiscovered
// factions — the fog of war reading as "something is out there." Only drawn if undiscovered
// factions exist; each blip is a small dim cluster placed at a seeded edge position (off a hash
// of the key) so it's stable but clearly at the periphery, not among the known nodes.
func drawUndiscoveredFog(img *image.RGBA, state game.GameState, home empireNode, p cosmicStratPalette, seed uint32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 16 || h < 16 {
		return
	}
	// Collect undiscovered keys in stable order (sorted like worldCivs does for discovered).
	var undis []string
	for k, f := range state.Diplomacy.Factions {
		if !f.Discovered {
			undis = append(undis, k)
		}
	}
	if len(undis) == 0 {
		return
	}
	sortStringsStable(undis)

	margin := clampInt(minInt(w, h)/10, 2, 12)
	for _, k := range undis {
		ks := factionKeySeed(k)
		// Place along the frame perimeter: pick an edge and a position on it from the key hash.
		edge := int(hashUnit(ks, 0xED9E, seed) * 4) // 0..3
		t := hashUnit(ks, 0x7051, seed)
		var cx, cy int
		switch edge {
		case 0: // top
			cx = margin + int(t*float64(w-2*margin))
			cy = margin
		case 1: // bottom
			cx = margin + int(t*float64(w-2*margin))
			cy = h - 1 - margin
		case 2: // left
			cx = margin
			cy = margin + int(t*float64(h-2*margin))
		default: // right
			cx = w - 1 - margin
			cy = margin + int(t*float64(h-2*margin))
		}
		// A faint "?" fog blip: a small dim smudge (a couple of feathered pixels), no ring — it's
		// deliberately vague. Kept crisp-dim (no wide halo).
		blendPixel(img, cx, cy, p.fog, 0.55)
		blendPixel(img, cx+1, cy, p.fog, 0.32)
		blendPixel(img, cx-1, cy, p.fog, 0.32)
		blendPixel(img, cx, cy+1, p.fog, 0.32)
		blendPixel(img, cx, cy-1, p.fog, 0.32)
	}
}

// drawStrategicChrome lays the HUD chrome: crisp cyan corner brackets around the frame (the
// targeting-reticle read, reusing the shared bracket look) and a couple of very faint
// concentric RANGE RINGS centered on the command hub (a tactical reticle). Clean and crisp, no
// heavy glow.
func drawStrategicChrome(img *image.RGBA, home empireNode, p cosmicStratPalette) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 12 || h < 12 {
		return
	}
	// Corner brackets (reuse the shared HUD bracket helper — crisp L-arms in the chrome cyan).
	drawHUDBrackets(img, p.chrome)

	// Faint concentric range rings on the hub — a couple of dim reticle circles. Kept dim
	// (blended toward the void) so they read as instrumentation, not clutter. Squashed on Y for
	// the half-block aspect so they read circular.
	ringCol := blend(p.voidDeep, p.chrome, 0.30)
	base := float64(minInt(w, h))
	for _, frac := range []float64{0.22, 0.36} {
		rr := base * frac
		drawRingEllipseDashed(img, home.cx, home.cy, rr, rr*0.62, ringCol, 2, 3)
	}
}

// ---- ring primitives --------------------------------------------------------

// drawRing strokes a crisp 1px circle of radius r centered at (cx,cy) — a thin unfilled ring.
// Squashed on Y (×0.62) so it reads round on the half-block canvas. Every pixel is
// bounds-checked. Used for the node rings and hub command rings.
func drawRing(img *image.RGBA, cx, cy, r int, col color.RGBA) {
	if r <= 0 {
		setPixel(img, cx, cy, col)
		return
	}
	drawRingEllipseDashed(img, cx, cy, float64(r), float64(r)*0.62, col, 1, 0)
}

// drawRingDashed strokes a dashed (broken) 1px ring — used for the at-war alert ring. `on`/`off`
// are the dash pattern measured in stepped points around the circumference.
func drawRingDashed(img *image.RGBA, cx, cy, r int, col color.RGBA, on, off int) {
	drawRingEllipseDashed(img, cx, cy, float64(r), float64(r)*0.62, col, on, off)
}

// drawRingEllipseDashed is the shared ring rasterizer: it walks an ellipse (radii rx,ry) by
// angle and stamps points, optionally dashed (on drawn / off skipped). off==0 → a solid ring.
// The step count scales with the radius so the ring stays continuous (no gaps) at any size, and
// every stamp is bounds-checked so it never panics off-canvas. Deterministic.
func drawRingEllipseDashed(img *image.RGBA, cx, cy int, rx, ry float64, col color.RGBA, on, off int) {
	if rx < 1 {
		rx = 1
	}
	if ry < 0.5 {
		ry = 0.5
	}
	if on < 1 {
		on = 1
	}
	if off < 0 {
		off = 0
	}
	// Enough steps that adjacent stamps are within ~1px: circumference-ish in pixels.
	steps := int(math.Max(rx, ry) * 6.5)
	if steps < 12 {
		steps = 12
	}
	period := on + off
	for i := 0; i < steps; i++ {
		if off > 0 && (i%period) >= on {
			continue
		}
		ang := float64(i) / float64(steps) * 2 * math.Pi
		x := cx + int(math.Round(math.Cos(ang)*rx))
		y := cy + int(math.Round(math.Sin(ang)*ry))
		setPixel(img, x, y, col)
	}
}

// sortStringsStable sorts a small string slice in place (insertion sort — the rosters are tiny
// and this avoids importing sort just for the fog-blip ordering; worldCivs already imports it
// for the discovered set). Deterministic ordering so the fog blips are stable frame-to-frame.
func sortStringsStable(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
