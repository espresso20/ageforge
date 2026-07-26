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
//
// Three cosmic ages are wired, each reframing the SAME empire roster (strategicEmpires) at a
// bigger scale so the set reads coherent-but-distinct:
//
//	space_age        — HOME CLUSTER: one dominant command hub + spheres of influence (territory).
//	interstellar_age — SECTOR ROUTE NETWORK: a hyperspace lane WEB linking star systems; the hub
//	                   is one node in the lattice, not the whole show. Routes are the picture.
//	galactic_age     — GALACTIC DOMINION: a spiral galaxy carved into territorial SECTORS by owner,
//	                   capitals seated on the arms. Territory (bold wedges), not routes, is the read.
//
// The relationship SIGNAL colors (war=red / ally=green / merc=gold / neutral=steel) stay constant
// across all three via nodeColor/empireRoleFor, so standings read the same everywhere; only the
// background, dominant metaphor, and chrome vary. quantum_age + transcendent_age are still unwired
// (they keep falling through to the neutral atlas) — a later batch.
func cosmicWorldViewFor(ageKey string) (func(img *image.RGBA, state game.GameState, w, h int, seed uint32), bool) {
	switch ageKey {
	case "space_age":
		return drawSpaceStrategicView, true
	case "interstellar_age":
		return interstellarStrategicView, true
	case "galactic_age":
		return galacticStrategicView, true
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

// ---- interstellar_age strategic view ----------------------------------------

// interstellarStrategicView paints the interstellar_age World tab as a SECTOR ROUTE NETWORK: a
// hyperspace route map where the STORY is connectivity, not a single seat. Where space_age
// centered one dominant command hub ringed by spheres of influence, this reframes the SAME roster
// as a lattice of STAR SYSTEMS wired to each other — the network topology IS the picture. Layers
// back-to-front:
//
//  1. deep-space field + a faint diagonal GALACTIC-PLANE BAND — colder and denser than space's
//     plain starfield, so the backdrop alone says "we're out in the lanes now."
//  2. a very faint node aura per system — territory is DE-EMPHASIZED (routes, not spheres, own
//     the frame), so the big influence bubbles of space_age are gone.
//  3. the LANE WEB (dominant) — every system linked to its two nearest neighbours, faint route
//     lines with small WAYPOINT JUNCTION pips at the midpoints (the peer-to-peer lattice).
//  4. relationship SPOKES from the hub to each faction reuse the lane vocabulary (bright trade
//     with flow pips / red dashed conflict / faint contact) so standings still read at a glance —
//     these are the "+ the hub" links; the web itself stays system↔system.
//  5. STAR SYSTEMS — each roster node a star with a tight little planetary ring, sized by
//     strength, relationship-coloured; the hub is a prominent-but-MODEST command star (smaller
//     than space's multi-ring seat — interstellar is about the web, not one node).
//  6. a route-map HUD — corner brackets plus faint waypoint ticks along the frame (distinct from
//     space's concentric range-ring reticle).
//
// LABEL-FREE, deterministic (hash/seed only), panic-safe (all writes bounds-checked; a tiny /
// odd / zero canvas returns after the field). Empty roster → the lone hub star in the band.
func interstellarStrategicView(img *image.RGBA, state game.GameState, w, h int, seed uint32) {
	if w <= 0 || h <= 0 {
		return
	}
	p := cosmicStratPal()

	// 1) Field: the shared deep-space fill/starfield, then a colder/denser galactic-plane band.
	drawCosmicField(img, p, seed)
	drawInterstellarField(img, p, seed)

	// Roster once, then REPOSITIONED into a wide network spread (not the mid-radius home cluster)
	// so the lattice spans the whole frame.
	nodes := interstellarSystems(state, w, h, seed)
	if len(nodes) == 0 {
		return
	}
	home := nodes[0]

	// 2) De-emphasized territory: a very faint aura per system (routes dominate, not spheres).
	for _, n := range nodes {
		drawSystemAura(img, n, p)
	}

	// 3) The LANE WEB (dominant): system↔nearest-neighbour lattice with waypoint junction pips.
	drawLaneWeb(img, nodes, p)

	// 4) Relationship spokes hub→faction reuse the lane vocabulary on top of the web.
	for _, n := range nodes {
		if n.isHome {
			continue
		}
		drawStrategicLane(img, home, n, p, seed)
	}

	// Undiscovered factions → faint edge fog blips (same fog-of-war read as space).
	drawUndiscoveredFog(img, state, home, p, seed)

	// 5) Star-system nodes on top; the command star LAST so the hub still crowns the web (modestly).
	for _, n := range nodes {
		if n.isHome {
			continue
		}
		drawStarSystemNode(img, n, p)
	}
	drawCommandStar(img, home, p)

	// 6) Route-map HUD: brackets + waypoint ticks.
	drawRouteHUD(img, p)
}

// drawInterstellarField overlays a faint diagonal GALACTIC-PLANE BAND on top of the shared cosmic
// field: a subtle cool wash concentrated along a slanted axis through the frame plus a DENSER
// scatter of cold blue stars inside it. The extra density + blue tint read colder than space's
// plain starfield — the backdrop alone signals we've zoomed out past the home cluster into the
// lanes. All hash-driven, bounds-checked, panic-safe on any canvas.
func drawInterstellarField(img *image.RGBA, p cosmicStratPalette, seed uint32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return
	}
	cx := float64(w) / 2
	cy := float64(h) / 2
	// Band axis runs along a shallow down-right diagonal; nx,ny is its (unit) NORMAL, so the
	// distance to the band centreline is |(x-cx)*nx + (y-cy)*ny|.
	nx, ny := 0.45, -0.89
	half := float64(minInt(w, h)) * 0.42
	if half < 1 {
		half = 1
	}
	cool := blend(p.voidDeep, p.starCool, 0.5)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			d := math.Abs((float64(x)-cx)*nx + (float64(y)-cy)*ny)
			if d > half {
				continue
			}
			f := 1.0 - d/half
			f = f * f // concentrate toward the axis
			n := valueNoise(float64(x)*0.03, float64(y)*0.03, seed^0x8A17)
			blendPixel(img, x, y, cool, clamp01(0.10*f*(0.5+0.5*n)))
		}
	}
	// Denser cold stars inside the band: a fine hash grid, only cells near the axis carry a star,
	// biased toward the centreline.
	cell := 4
	gw := w / cell
	gh := h / cell
	for gy := 0; gy <= gh; gy++ {
		for gx := 0; gx <= gw; gx++ {
			sx := gx * cell
			sy := gy * cell
			d := math.Abs((float64(sx)-cx)*nx + (float64(sy)-cy)*ny)
			if d > half {
				continue
			}
			thresh := 0.5 * (1.0 - d/half) // denser than the base field, tightest on the axis
			if hashUnit(uint32(gx), uint32(gy), seed^0xB33D) > thresh {
				continue
			}
			ox := int(hashUnit(uint32(gx), uint32(gy), seed^0x51) * float64(cell))
			oy := int(hashUnit(uint32(gy), uint32(gx), seed^0x52) * float64(cell))
			bright := 0.25 + 0.55*hashUnit(uint32(gx), uint32(gy), seed^0x53)
			setPixel(img, sx+ox, sy+oy, blend(p.voidDeep, p.starCool, bright))
		}
	}
}

// interstellarSystems returns the roster (via strategicEmpires) REPOSITIONED for the route
// network: the hub nudged a hair off dead-centre so it reads as ONE node in the web (not the seat
// everything orbits), and the factions spread WIDE across the frame on a golden-angle sweep in an
// outer band so the lattice fills the map instead of hugging one radius. Deterministic.
func interstellarSystems(state game.GameState, w, h int, seed uint32) []empireNode {
	nodes := strategicEmpires(state, w, h, seed)
	if len(nodes) == 0 {
		return nodes
	}
	// Hub: nudged off dead-centre (deterministic) — one node in the network, not THE seat.
	nudgeX := int((hashUnit(seed, 0x1AA, 0x9) - 0.5) * float64(w) * 0.10)
	nudgeY := int((hashUnit(seed, 0x2BB, 0x9) - 0.5) * float64(h) * 0.10)
	nodes[0].cx = clampInt(w/2+nudgeX, 0, w-1)
	nodes[0].cy = clampInt(h/2+nudgeY, 0, h-1)

	margin := clampInt(minInt(w, h)/9, 3, 32)
	maxRX := float64(w)/2 - float64(margin)
	maxRY := float64(h)/2 - float64(margin)
	if maxRX < 2 {
		maxRX = 2
	}
	if maxRY < 2 {
		maxRY = 2
	}
	const golden = 2.399963229728653
	rot := hashUnit(seed, 0x5C7, 0x271) * 2 * math.Pi
	fx := float64(nodes[0].cx)
	fy := float64(nodes[0].cy)
	fi := 0
	for i := range nodes {
		if nodes[i].isHome {
			continue
		}
		ks := factionKeySeed(nodes[i].key)
		ang := float64(fi)*golden + rot
		rr := 0.55 + 0.45*hashUnit(ks, 0x77, seed) // outer band → systems sit well apart
		px := fx + math.Cos(ang)*maxRX*rr
		py := fy + math.Sin(ang)*maxRY*rr
		nodes[i].cx = clampInt(int(math.Round(px)), margin, maxInt(w-1-margin, margin))
		nodes[i].cy = clampInt(int(math.Round(py)), margin, maxInt(h-1-margin, margin))
		fi++
	}
	return nodes
}

// drawSystemAura paints a VERY faint node aura for one system — the de-emphasized territory read
// of the route network (space_age's bold influence bubble, dialled right down so routes dominate).
// Additive, tiny radius, low alpha, tinted by relationship. Bounds-checked.
func drawSystemAura(img *image.RGBA, n empireNode, p cosmicStratPalette) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return
	}
	tint := nodeColor(n, p)
	base := float64(minInt(w, h))
	r := base * 0.045
	if n.isHome {
		r = base * 0.07
	}
	if r < 2 {
		r = 2
	}
	rY := r * 0.62
	peak := 0.10
	if n.isHome {
		peak = 0.14
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
			blendPixel(img, x, y, tint, peak*(1.0-d))
		}
	}
}

// drawLaneWeb strokes the DOMINANT lattice of the route network: each star SYSTEM linked to its
// two nearest peer systems by a faint route line with a small WAYPOINT JUNCTION pip at the
// midpoint. The hub is deliberately EXCLUDED from the lattice — its connections are the coloured
// relationship spokes drawn separately — so the web reads as a peer-to-peer network, not
// hub-and-spoke. Distances undo the half-block Y squash so "nearest" matches the on-screen look.
// Edges are de-duplicated so a mutual-nearest pair draws once. Deterministic (stable node order).
func drawLaneWeb(img *image.RGBA, nodes []empireNode, p cosmicStratPalette) {
	peers := make([]int, 0, len(nodes))
	for i := range nodes {
		if !nodes[i].isHome {
			peers = append(peers, i)
		}
	}
	if len(peers) < 2 {
		return
	}
	route := blend(p.laneDim, p.laneTrade, 0.35)
	seen := map[[2]int]bool{}
	addEdge := func(a, bIdx int) {
		if a == bIdx {
			return
		}
		key := [2]int{a, bIdx}
		if a > bIdx {
			key = [2]int{bIdx, a}
		}
		if seen[key] {
			return
		}
		seen[key] = true
		strokeSolidFaintLine(img, nodes[a].cx, nodes[a].cy, nodes[bIdx].cx, nodes[bIdx].cy, route, 0.5)
		drawWaypointPip(img, (nodes[a].cx+nodes[bIdx].cx)/2, (nodes[a].cy+nodes[bIdx].cy)/2, brighten(route, 0.30))
	}
	for _, i := range peers {
		n1, n2 := -1, -1
		d1, d2 := math.MaxFloat64, math.MaxFloat64
		for _, j := range peers {
			if j == i {
				continue
			}
			dx := float64(nodes[i].cx - nodes[j].cx)
			dy := float64(nodes[i].cy-nodes[j].cy) / 0.62 // undo half-block squash → true distance
			d := dx*dx + dy*dy
			if d < d1 {
				n2, d2 = n1, d1
				n1, d1 = j, d
			} else if d < d2 {
				n2, d2 = j, d
			}
		}
		if n1 >= 0 {
			addEdge(i, n1)
		}
		if n2 >= 0 {
			addEdge(i, n2)
		}
	}
}

// drawWaypointPip stamps a tiny + junction mark at a route midpoint — a hyperspace waypoint where
// lanes meet. Crisp, bounds-checked.
func drawWaypointPip(img *image.RGBA, cx, cy int, col color.RGBA) {
	setPixel(img, cx, cy, col)
	setPixel(img, cx+1, cy, col)
	setPixel(img, cx-1, cy, col)
	setPixel(img, cx, cy+1, col)
	setPixel(img, cx, cy-1, col)
}

// drawStarSystemNode paints one faction as a STAR SYSTEM: a small bright star body sized by
// strength with a lit core pixel and a tight planetary RING hugging it (the "system" read). An
// at-war system gets a BROKEN red alert ring instead of the solid identity ring. Crisp, no halo —
// the faint aura already carries the glow. Clearly smaller than space's faction blobs by design.
func drawStarSystemNode(img *image.RGBA, n empireNode, p cosmicStratPalette) {
	col := nodeColor(n, p)
	r := clampInt(1+n.strength/2, 1, 3)
	fillDot(img, n.cx, n.cy, r, brighten(col, 0.15))
	setPixel(img, n.cx, n.cy, brighten(col, 0.50)) // lit core
	if n.atWar {
		drawRingDashed(img, n.cx, n.cy, r+2, p.war, 2, 2)
	} else {
		drawRing(img, n.cx, n.cy, r+2, blend(p.voidDeep, col, 0.75))
	}
}

// drawCommandStar paints the player's seat for the route network: a prominent but MODEST command
// star (body + hot core + a SINGLE command ring) — deliberately smaller than space_age's multi-
// ring hub so your empire reads as one node in the web, not the seat everything orbits.
func drawCommandStar(img *image.RGBA, home empireNode, p cosmicStratPalette) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	r := clampInt((w+h)/48, 3, 6)
	fillDot(img, home.cx, home.cy, r, p.home)
	coreR := clampInt(r/2, 1, 2)
	fillDot(img, home.cx, home.cy, coreR, p.homeCore)
	drawRing(img, home.cx, home.cy, r+2, brighten(p.home, 0.20))
}

// drawRouteHUD lays the route-map HUD: the shared corner brackets plus faint WAYPOINT TICKS along
// the top and bottom edges (a route-ruler read), distinct from space's concentric range rings and
// galactic's core ring. Crisp, bounds-checked.
func drawRouteHUD(img *image.RGBA, p cosmicStratPalette) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 12 || h < 12 {
		return
	}
	drawHUDBrackets(img, p.chrome)
	tickCol := blend(p.voidDeep, p.chrome, 0.5)
	step := clampInt(w/12, 6, 40)
	for x := step; x < w-step/2; x += step {
		setPixel(img, x, 2, tickCol)
		setPixel(img, x, 3, tickCol)
		setPixel(img, x, h-3, tickCol)
		setPixel(img, x, h-4, tickCol)
	}
}

// ---- galactic_age strategic view --------------------------------------------

// empireWedge is one empire's angular TERRITORY in the galactic dominion view: a contiguous
// sector [start,end) of the galaxy (radians, measured from a seeded global rotation), its centre
// angle `mid` (the capital direction), and the index of its owner in the node slice. Widths are
// proportional to strength so a strong empire claims a wider sector.
type empireWedge struct {
	start, end float64
	mid        float64
	idx        int
}

// galacticStrategicView paints the galactic_age World tab as GALACTIC DOMINION: a spiral galaxy
// carved into TERRITORIAL sectors by owner. Where space_age was one home cluster and
// interstellar_age was a route web, this reframes the SAME roster at galaxy scale — the dominant
// read is bold TERRITORY, not spheres or lanes. Layers back-to-front:
//
//  1. a SPIRAL GALAXY field — a warm crisp central-bulge CORE glow plus a few log-spiral ARMS of
//     stars sweeping outward over the void (built here; NOT the Map-tab scenic galaxy).
//  2. SECTORS (dominant) — the galaxy divided into angular wedges from the core, one per empire,
//     sized by strength, each filled with the owner's translucent relationship colour (bolder and
//     far LARGER than space's bubbles — territory is the point). The spiral arms show THROUGH the
//     tint. Adjacent HOSTILE sectors get a red contested BORDER SEAM along their shared radial edge.
//  3. faint trade lanes (subordinate) from the home capital to trade partners — territory, not
//     routes, is the read, so these stay dim.
//  4. CAPITAL stars — each empire's seat repositioned into its sector out along the galaxy, sized
//     by strength, relationship-coloured; the home capital commands the CORE (brightest, central).
//  5. a galactic-scale HUD — corner brackets + a single faint core range ring.
//
// LABEL-FREE, deterministic (hash/seed only), panic-safe (all writes bounds-checked; a tiny /
// odd / zero canvas returns after the field). Empty roster → the home capital alone on the core.
func galacticStrategicView(img *image.RGBA, state game.GameState, w, h int, seed uint32) {
	if w <= 0 || h <= 0 {
		return
	}
	p := cosmicStratPal()

	// 1) Spiral galaxy backdrop (core bulge + log-spiral arms).
	drawSpiralGalaxyField(img, p, seed)

	nodes := strategicEmpires(state, w, h, seed)
	if len(nodes) == 0 {
		return
	}

	// Sector assignment: contiguous angular wedges around the core, sized by strength, in stable
	// order; each wedge also carries its capital direction (the wedge centre).
	wedges := galacticWedges(nodes, seed)

	// 2) Territory (dominant): fill each wedge with the owner's translucent relationship colour;
	//    red seams on hostile boundaries.
	drawGalacticSectors(img, nodes, wedges, p)

	// Reposition capitals into their sectors (home on the core), then thread through the rest.
	caps := galacticCapitals(nodes, wedges, w, h, seed)
	home := caps[0]

	// 3) Subordinate faint trade lanes home→trade partners (territory is the read; lanes stay dim).
	for _, n := range caps {
		if n.isHome || n.atWar || !n.trade {
			continue
		}
		strokeSolidFaintLine(img, home.cx, home.cy, n.cx, n.cy, blend(p.voidDeep, p.laneTrade, 0.55), 0.35)
	}

	// Undiscovered factions → faint edge fog blips.
	drawUndiscoveredFog(img, state, home, p, seed)

	// 4) Capital stars; the core capital LAST so the home seat crowns the galaxy.
	for _, n := range caps {
		if n.isHome {
			continue
		}
		drawCapitalStar(img, n, p)
	}
	drawCoreCapital(img, home, p)

	// 5) Galactic HUD: brackets + a faint core range ring.
	drawGalacticHUD(img, home, p)
}

// drawSpiralGalaxyField builds the galactic-dominion backdrop: a near-black void, a warm crisp
// central-bulge CORE glow (tight falloff — bright, not a muddy smear), a few log-spiral ARMS of
// stars sweeping outward (denser + warmer near the core, thinning + cooling outward), and a sparse
// scatter of faint background stars in the void. Purpose-built here (NOT the Map-tab scenic
// drawGalaxyScene). Aspect-squashed on Y so the galaxy reads round on the half-block canvas. All
// hash-driven, bounds-checked, panic-safe on any canvas.
func drawSpiralGalaxyField(img *image.RGBA, p cosmicStratPalette, seed uint32) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return
	}
	cx := float64(w) / 2
	cy := float64(h) / 2
	const aspect = 0.62
	coreWarm := color.RGBA{R: 0xff, G: 0xe6, B: 0xb0, A: 0xff} // gold bulge
	coreHot := color.RGBA{R: 0xff, G: 0xf6, B: 0xe6, A: 0xff}  // near-white centre
	armStar := p.starCool
	base := float64(minInt(w, h))
	coreR := base * 0.30

	// 1) Void base + warm core bulge glow. One pass over the frame: start from the void, add the
	//    radial core glow with a tight pow-falloff so the bulge is bright and crisp, not smeared.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			col := p.voidDeep
			dx := float64(x) - cx
			dy := (float64(y) - cy) / aspect
			dr := math.Sqrt(dx*dx+dy*dy) / coreR
			if dr < 1.0 {
				g := math.Pow(1.0-dr, 2.2)
				col = blend(col, coreWarm, clamp01(g*0.9))
				if dr < 0.28 {
					hot := math.Pow(1.0-dr/0.28, 1.6)
					col = blend(col, coreHot, clamp01(hot*0.8))
				}
			}
			setPixel(img, x, y, col)
		}
	}

	// 2) Log-spiral arms r = a·e^(b·θ): stars stamped along each arm with seeded perpendicular
	//    scatter (growing with radius), brightness thinning outward, warm-tinted near the core.
	const arms = 2
	maxR := base * 0.5
	a := base * 0.05
	tight := 0.22
	for arm := 0; arm < arms; arm++ {
		armPhase := float64(arm) / float64(arms) * 2 * math.Pi
		steps := int(maxR * 3)
		if steps < 60 {
			steps = 60
		}
		for s := 0; s < steps; s++ {
			th := float64(s) / float64(steps) * (3.2 * math.Pi) // ~1.6 turns
			r := a * math.Exp(tight*th)
			if r > maxR {
				break
			}
			ang := th + armPhase
			for k := 0; k < 2; k++ {
				hs := seed ^ uint32(arm*911+s*17+k*7)
				jr := (hashUnit(hs, 0x1, seed) - 0.5) * r * 0.30
				ja := (hashUnit(hs, 0x2, seed) - 0.5) * 0.25
				rr := r + jr
				aa := ang + ja
				px := cx + math.Cos(aa)*rr
				py := cy + math.Sin(aa)*rr*aspect
				bright := clamp01(0.80 - rr/maxR*0.55)
				star := blend(p.voidDeep, armStar, bright)
				if rr < coreR {
					star = blend(star, coreWarm, 0.40)
				}
				setPixel(img, int(math.Round(px)), int(math.Round(py)), star)
			}
		}
	}

	// 3) Sparse faint background stars in the void so it isn't dead.
	cell := 7
	gw := w / cell
	gh := h / cell
	for gy := 0; gy <= gh; gy++ {
		for gx := 0; gx <= gw; gx++ {
			if hashUnit(uint32(gx), uint32(gy), seed^0x2C1) > 0.22 {
				continue
			}
			sx := gx*cell + int(hashUnit(uint32(gx), uint32(gy), seed^0x11)*float64(cell))
			sy := gy*cell + int(hashUnit(uint32(gy), uint32(gx), seed^0x22)*float64(cell))
			br := 0.20 + 0.40*hashUnit(uint32(gx), uint32(gy), seed^0x33)
			setPixel(img, sx, sy, blend(p.voidDeep, p.starWarm, br))
		}
	}
}

// galacticWedges divides the galaxy into contiguous angular TERRITORIES, one per empire, widths
// proportional to strength (the home empire gets a strength floor so its sector reads dominant),
// walked around the circle from a seeded global rotation in the stable node order. Deterministic.
func galacticWedges(nodes []empireNode, seed uint32) []empireWedge {
	if len(nodes) == 0 {
		return nil
	}
	weight := func(n empireNode) float64 {
		s := float64(n.strength)
		if s < 1 {
			s = 1
		}
		if n.isHome && s < 6 {
			s = 6 // home commands the widest sector
		}
		return s
	}
	total := 0.0
	for _, n := range nodes {
		total += weight(n)
	}
	if total <= 0 {
		total = 1
	}
	rot := hashUnit(seed, 0xC0D, 0x5EC) * 2 * math.Pi
	wedges := make([]empireWedge, 0, len(nodes))
	acc := rot
	for i, n := range nodes {
		span := weight(n) / total * 2 * math.Pi
		wedges = append(wedges, empireWedge{start: acc, end: acc + span, mid: acc + span/2, idx: i})
		acc += span
	}
	return wedges
}

// drawGalacticSectors fills the DOMINANT territory layer: every pixel's angle from the core picks
// its owning wedge, and the owner's translucent relationship colour is blended in — bold and
// covering the whole sector (far larger than space's bubbles), yet a blend so the spiral arms and
// core show THROUGH. The tint eases toward zero near the core so the bright bulge stays legible.
// Then each wedge boundary bordering an at-war sector gets a red dashed contested SEAM from the
// core outward. Deterministic, bounds-checked.
func drawGalacticSectors(img *image.RGBA, nodes []empireNode, wedges []empireWedge, p cosmicStratPalette) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || len(wedges) == 0 {
		return
	}
	cx := float64(w) / 2
	cy := float64(h) / 2
	const aspect = 0.62
	half := float64(minInt(w, h)) * 0.5
	if half < 1 {
		half = 1
	}
	tints := make([]color.RGBA, len(wedges))
	for i, wd := range wedges {
		tints[i] = nodeColor(nodes[wd.idx], p)
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - cx
			dy := (float64(y) - cy) / aspect
			owner := wedgeOwner(wedges, math.Atan2(dy, dx))
			if owner < 0 {
				continue
			}
			rr := math.Sqrt(dx*dx+dy*dy) / half
			alpha := 0.28 * clamp01(rr*1.3) // fade near the core so the bulge reads
			blendPixel(img, x, y, tints[owner], alpha)
		}
	}
	// Contested seams: a red radial edge wherever two adjacent sectors are hostile (either at war).
	for i := range wedges {
		j := (i + 1) % len(wedges)
		if !(nodes[wedges[i].idx].atWar || nodes[wedges[j].idx].atWar) {
			continue
		}
		x1 := int(math.Round(cx + math.Cos(wedges[i].end)*half))
		y1 := int(math.Round(cy + math.Sin(wedges[i].end)*half*aspect))
		strokeDashedLine(img, int(math.Round(cx)), int(math.Round(cy)), x1, y1, p.laneWar, 3, 2)
	}
}

// wedgeOwner returns the index of the wedge that contains angle `ang` (from atan2), normalising
// the query into the wedges' [start, start+2π) span. Returns the last wedge as a numerical-edge
// fallback so every pixel is owned. Deterministic.
func wedgeOwner(wedges []empireWedge, ang float64) int {
	if len(wedges) == 0 {
		return -1
	}
	const twoPi = 2 * math.Pi
	base := wedges[0].start
	a := ang
	for a < base {
		a += twoPi
	}
	for a >= base+twoPi {
		a -= twoPi
	}
	for i, wd := range wedges {
		if a >= wd.start && a < wd.end {
			return i
		}
	}
	return len(wedges) - 1
}

// galacticCapitals returns a copy of the roster with each empire's capital repositioned into its
// sector: the home seat parked on the galaxy CORE (centre), every faction capital placed at its
// wedge's centre angle at a seeded radius out toward the rim (so capitals sit spread across the
// galaxy along their sectors, stronger empires no closer to the core than weaker ones — the seed
// decides). Deterministic, clamped inside a frame margin.
func galacticCapitals(nodes []empireNode, wedges []empireWedge, w, h int, seed uint32) []empireNode {
	out := make([]empireNode, len(nodes))
	copy(out, nodes)
	if w <= 0 || h <= 0 {
		return out
	}
	cx := float64(w) / 2
	cy := float64(h) / 2
	const aspect = 0.62
	maxR := float64(minInt(w, h)) * 0.46
	margin := clampInt(minInt(w, h)/12, 2, 20)
	for _, wd := range wedges {
		n := &out[wd.idx]
		if n.isHome {
			n.cx = clampInt(int(math.Round(cx)), 0, w-1)
			n.cy = clampInt(int(math.Round(cy)), 0, h-1)
			continue
		}
		ks := factionKeySeed(n.key)
		radius := maxR * (0.45 + 0.45*hashUnit(ks, 0x9A1, seed))
		px := cx + math.Cos(wd.mid)*radius
		py := cy + math.Sin(wd.mid)*radius*aspect
		n.cx = clampInt(int(math.Round(px)), margin, maxInt(w-1-margin, margin))
		n.cy = clampInt(int(math.Round(py)), margin, maxInt(h-1-margin, margin))
	}
	return out
}

// drawCapitalStar paints one faction's CAPITAL seat within its sector: a filled star body sized by
// strength with a lit core pixel and a crisp relationship-coloured ring (an at-war capital gets a
// broken red alert ring). Crisp, bounds-checked — the sector tint carries the territory glow.
func drawCapitalStar(img *image.RGBA, n empireNode, p cosmicStratPalette) {
	col := nodeColor(n, p)
	r := clampInt(2+n.strength/2, 2, 4)
	fillDot(img, n.cx, n.cy, r, brighten(col, 0.12))
	setPixel(img, n.cx, n.cy, brighten(col, 0.50))
	if n.atWar {
		drawRingDashed(img, n.cx, n.cy, r+1, p.war, 2, 2)
	} else {
		drawRing(img, n.cx, n.cy, r+1, brighten(col, 0.18))
	}
}

// drawCoreCapital paints the player's seat on the galaxy CORE: a command-cyan body over the warm
// bulge with a hot near-white core and two proud command rings — the brightest, most central mark,
// so the home empire unmistakably commands the core of the galaxy.
func drawCoreCapital(img *image.RGBA, home empireNode, p cosmicStratPalette) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	r := clampInt((w+h)/34, 5, 9)
	fillDot(img, home.cx, home.cy, r, p.home)
	coreR := clampInt(r/2, 1, 3)
	fillDot(img, home.cx, home.cy, coreR, p.homeCore)
	drawRing(img, home.cx, home.cy, r+2, brighten(p.home, 0.22))
	drawRing(img, home.cx, home.cy, r+4, blend(p.voidDeep, p.home, 0.65))
}

// drawGalacticHUD lays the galactic-scale HUD: the shared corner brackets plus a SINGLE faint core
// range ring (distinct from space's two reticle rings and interstellar's edge ticks). Crisp,
// bounds-checked.
func drawGalacticHUD(img *image.RGBA, home empireNode, p cosmicStratPalette) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 12 || h < 12 {
		return
	}
	drawHUDBrackets(img, p.chrome)
	ringCol := blend(p.voidDeep, p.chrome, 0.28)
	rr := float64(minInt(w, h)) * 0.44
	drawRingEllipseDashed(img, home.cx, home.cy, rr, rr*0.62, ringCol, 3, 4)
}
