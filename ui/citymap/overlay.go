package citymap

import (
	"image/color"
	"math"
	"sort"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// overlay.go is the P3 text-overlay pass — the "crisp glyph" half of the hybrid
// render model (design doc §"Rendering model"). P1/P2 paint everything as pixels
// into an *image.RGBA streamed through '▄' half-blocks; labels and civ markers
// read terribly as pixels, so instead we compute their CELL positions here and the
// Build draw-closure stamps them with screen.SetContent AFTER the half-blocks are
// laid down, overwriting the '▄' with theme-colored characters.
//
// The crucial discipline: an overlayPlan stores GEOMETRY + a theme ROLE, never a
// resolved color. Geometry doesn't change when the theme switches (only the hue
// does), so the plan is cached alongside the image, and the draw closure resolves
// theme.Color(role) live every frame — a theme switch retints the overlay text too.
//
// This file owns the systems-weave from the design doc's D3:
//   - civ edge-markers — the discovered diplomacy civs as a relationship-colored
//     ring around the border (drawn as text here; "your neighbors").
//   - district labels — each lineage district named at its centroid.
//   - title — "Empire — <Age>" in a corner.
//   - the trade-route border labels (the lanes themselves are pixels, drawn into
//     the image in entities.go; the little end labels are text, planned here).

// labelKind tags an overlay label so the draw pass can apply small per-kind tweaks
// (bold the title, glyph-prefix a civ at war) without baking colors into the plan.
type labelKind int

const (
	labelDistrict labelKind = iota
	labelCiv
	labelTitle
	labelTrade
	labelCapital
)

// overlayLabel is one piece of text (or a single glyph) to stamp into the cell
// grid. cx/cy are CELL coordinates (not pixels). The color is resolved at DRAW time
// (so the overlay retints on a theme switch) from one of two sources: normally
// theme.Color(role); but if lineageColored is set, from lineageColor(lineageKey,
// category) — the same hue-rotated theme color the district's volume uses, so a
// district's banner matches its buildings and still retints. bright nudges the
// resolved color toward white (the at-war civ marker). align controls anchoring so
// edge labels on the right border don't spill off-screen.
type overlayLabel struct {
	cx, cy int
	text   string
	role   theme.Role
	kind   labelKind
	bright bool
	align  alignment

	// Lineage color source (overrides role when lineageColored is true).
	lineageColored bool
	lineageKey     string
	category       string
}

// alignment controls how a label's text is positioned relative to (cx,cy).
type alignment int

const (
	alignLeft   alignment = iota // text starts at cx, runs right
	alignRight                   // text ends at cx, runs left
	alignCenter                  // text centered on cx
)

// overlayPlan is the full set of text labels for one render, in cell space. It is
// computed in renderImage (so its geometry matches the pixel layout exactly) and
// cached next to the image. labels are stamped in slice order, so later entries win
// where two would overlap — callers append the title last so it sits on top.
type overlayPlan struct {
	labels []overlayLabel
}

// cellCols/cellRows: the image is (cols × rows*2) pixels because each cell is one
// half-block (two stacked pixels). The overlay works in cells, so converting a
// pixel y to a cell row is a halving. A label placed at cell row R lands on image
// pixel rows 2R / 2R+1; we anchor on the lower pixel (matches the eye, since the
// half-block's foreground — the glyph color — is the lower pixel).
func pxToCellY(py int) int { return py / 2 }

// buildOverlayPlan assembles every text label for a frame from the state snapshot
// plus the already-computed layout geometry (district centroids, palace cell, and
// the trade-lane border endpoints from the pixel pass). cols/rows are the terminal
// CELL dimensions. It is pure: no locks, no engine calls — it reads the snapshot
// and config-by-key data only, honoring the bus/lock rules.
func buildOverlayPlan(state game.GameState, cols, rows int, geo layoutGeometry) overlayPlan {
	var plan overlayPlan
	if cols <= 0 || rows <= 0 {
		return plan
	}

	// occupied tracks cell rows already claimed by a label on a given side so we can
	// skip/stagger collisions. Keyed by a packed (side<<20 | row) — cheap and avoids
	// a 2D grid alloc per frame. Districts and civs both consult it.
	occupied := map[int]bool{}

	// 1) District labels at their centroids — only districts with enough buildings
	//    to warrant a name, skipping any that would collide.
	plan.addDistrictLabels(geo, cols, rows, occupied)

	// 2) The capital label sits just under the palace.
	plan.addCapitalLabel(geo, cols, rows, occupied)

	// 3) Civ edge-markers — the discovered diplomacy ring around the border.
	plan.addCivMarkers(state, cols, rows, occupied)

	// 4) Trade-lane end labels at the border (the lanes are pixels; these name them).
	plan.addTradeLabels(geo, cols, rows, occupied)

	// 5) Title last, so it sits on top of anything in its corner.
	plan.addTitle(state, cols, rows)

	return plan
}

// addDistrictLabels names each district at the cell above its centroid. It labels
// only districts the geometry deemed worth naming (enough reps), truncates the name
// to fit, and skips a label whose row is already taken on the center band so two
// district names never stack on the same row illegibly.
func (p *overlayPlan) addDistrictLabels(geo layoutGeometry, cols, rows int, occupied map[int]bool) {
	// Sort by descending building count so when space is tight the bigger districts
	// win their labels (centroids are pre-sorted by lineage key otherwise).
	cents := make([]districtCentroid, len(geo.districts))
	copy(cents, geo.districts)
	sort.SliceStable(cents, func(i, j int) bool { return cents[i].count > cents[j].count })

	const sideCenter = 0
	for _, c := range cents {
		name := lineageLabel(c.lineageKey, c.category)
		if name == "" {
			continue
		}
		row := clampInt(pxToCellY(c.py)-1, 0, rows-1) // a cell above the cluster
		// Center the label on the centroid column, clamped so it stays on-screen.
		col := clampInt(pxToCellX(c.px), 0, cols-1)
		// Collision: if this row (center band) is taken near this column, nudge up a
		// row; if that's taken too, skip the label rather than overprint.
		key := packCell(sideCenter, row)
		if occupied[key] {
			row = clampInt(row-1, 0, rows-1)
			key = packCell(sideCenter, row)
			if occupied[key] {
				continue
			}
		}
		occupied[key] = true
		name = truncLabel(name, maxLabelLen(col, cols, alignCenter))
		if name == "" {
			continue
		}
		p.labels = append(p.labels, overlayLabel{
			cx: col, cy: row, text: name, kind: labelDistrict, align: alignCenter,
			lineageColored: true, lineageKey: c.lineageKey, category: c.category,
		})
	}
}

// addCapitalLabel stamps "Capital" just beneath the palace, in the accent role.
func (p *overlayPlan) addCapitalLabel(geo layoutGeometry, cols, rows int, occupied map[int]bool) {
	row := clampInt(pxToCellY(geo.palaceY)+2, 0, rows-1)
	col := clampInt(pxToCellX(geo.palaceX), 0, cols-1)
	occupied[packCell(0, row)] = true
	p.labels = append(p.labels, overlayLabel{
		cx: col, cy: row, text: "Capital", role: theme.RoleAccent, kind: labelCapital, align: alignCenter,
	})
}

// civMarker is a discovered civ resolved to a display name + a status role + flags.
type civMarker struct {
	name  string
	role  theme.Role
	atWar bool
}

// addCivMarkers places the discovered diplomacy civs around the border as a
// relationship-colored ring. Allies → Positive, rivals/embargo → Negative, at-war →
// bright Negative + a "!" mark, friendly → a softer Positive (via Label), neutral →
// Dim. Civs are spaced down the left and right borders (rows), top reserved for the
// title, so markers never overlap each other or the title.
func (p *overlayPlan) addCivMarkers(state game.GameState, cols, rows int, occupied map[int]bool) {
	civs := discoveredCivs(state)
	if len(civs) == 0 {
		return
	}

	// Two columns of slots: left edge (col 0, left-aligned) and right edge
	// (col cols-1, right-aligned). Leave the top 2 rows for the title and the very
	// bottom row clear, then space remaining civs evenly down each side.
	topPad, botPad := 2, 1
	usable := rows - topPad - botPad
	if usable < 1 {
		usable = 1
	}
	// Round-robin civs into left/right so both sides fill evenly.
	leftSlots, rightSlots := splitSlots(len(civs), usable)

	place := func(c civMarker, sideLeft bool, slotIdx, slotCount int) {
		if slotCount <= 0 {
			return
		}
		// Evenly distribute slots across the usable height.
		var row int
		if slotCount == 1 {
			row = topPad + usable/2
		} else {
			row = topPad + slotIdx*(usable-1)/(slotCount-1)
		}
		row = clampInt(row, 0, rows-1)
		side := 1
		col := 0
		al := alignLeft
		if !sideLeft {
			side = 2
			col = cols - 1
			al = alignRight
		}
		key := packCell(side, row)
		if occupied[key] {
			// Nudge down one row, then up one; give up if both taken.
			if r2 := row + 1; r2 < rows && !occupied[packCell(side, r2)] {
				row = r2
			} else if r3 := row - 1; r3 >= 0 && !occupied[packCell(side, r3)] {
				row = r3
			} else {
				return
			}
			key = packCell(side, row)
		}
		occupied[key] = true

		text := c.name
		if c.atWar {
			text = "!" + text
		}
		// Budget the text to the half-width so a left and a right label can't meet in
		// the middle; right labels run leftward from the edge.
		text = truncLabel(text, maxLabelLen(col, cols, al))
		if text == "" {
			return
		}
		p.labels = append(p.labels, overlayLabel{
			cx: col, cy: row, text: text, role: c.role, kind: labelCiv, bright: c.atWar, align: al,
		})
	}

	li, ri := 0, 0
	for i, c := range civs {
		if i%2 == 0 && li < leftSlots {
			place(c, true, li, leftSlots)
			li++
		} else if ri < rightSlots {
			place(c, false, ri, rightSlots)
			ri++
		} else if li < leftSlots {
			place(c, true, li, leftSlots)
			li++
		}
	}
}

// addTradeLabels names the active trade lanes with a tiny tag at their border end.
// The lanes themselves are pixel lines (drawn in the image pass); here we drop a
// short label near each border endpoint, skipping rows already claimed by a civ
// marker so a lane tag never lands on a neighbor's name.
func (p *overlayPlan) addTradeLabels(geo layoutGeometry, cols, rows int, occupied map[int]bool) {
	for _, t := range geo.tradeEnds {
		row := clampInt(pxToCellY(t.py), 0, rows-1)
		col := clampInt(pxToCellX(t.px), 0, cols-1)
		// Choose a side for collision bookkeeping by which border the end hugs.
		side := 3 // a dedicated "trade" band so it only collides with other trade tags
		al := alignCenter
		if col <= 1 {
			al = alignLeft
		} else if col >= cols-2 {
			al = alignRight
		}
		key := packCell(side, row)
		if occupied[key] {
			continue
		}
		// Also avoid sitting directly on a civ marker on the same side row.
		if (al == alignLeft && occupied[packCell(1, row)]) || (al == alignRight && occupied[packCell(2, row)]) {
			continue
		}
		occupied[key] = true
		name := truncLabel(t.name, maxLabelLen(col, cols, al))
		if name == "" {
			continue
		}
		p.labels = append(p.labels, overlayLabel{
			cx: col, cy: row, text: name, role: theme.RolePositive, kind: labelTrade, align: al,
		})
	}
}

// addTitle stamps the corner title "<civ> — <Age>" at the top-left. The civ name
// is the account display name (AccountStats), falling back to "Empire" when
// accountless or unnamed; the age name is the live one. Drawn last so it crowns
// its corner.
func (p *overlayPlan) addTitle(state game.GameState, cols, rows int) {
	age := state.AgeName
	if age == "" {
		age = state.Age
	}
	name := "Empire"
	if state.AccountStats != nil && state.AccountStats.DisplayName != "" {
		name = state.AccountStats.DisplayName
	}
	title := name
	if age != "" {
		title = name + " — " + age
	}
	title = truncLabel(title, cols-1)
	if title == "" {
		return
	}
	p.labels = append(p.labels, overlayLabel{
		cx: 1, cy: 0, text: title, role: theme.RoleAccent, kind: labelTitle, align: alignLeft,
	})
}

// stampOverlay draws an overlayPlan onto the screen at the box origin (offX,offY),
// resolving each label's role to a live theme color so a theme switch retints the
// text. It is panic-safe and clips every glyph to the box bounds — nothing is drawn
// outside [offX,offX+cols) × [offY,offY+rows). This runs in the Build draw-closure
// AFTER streamHalfBlocks, overwriting the '▄' cells with crisp characters.
func stampOverlay(screen tcell.Screen, plan overlayPlan, offX, offY, cols, rows int) {
	if screen == nil || cols <= 0 || rows <= 0 {
		return
	}
	for _, lb := range plan.labels {
		if lb.text == "" {
			continue
		}
		// Resolve color live (retints on theme switch) from either the lineage color
		// (so a district banner matches its volumes) or the flat theme role. Bright
		// nudges toward white so the at-war marker pops above the rest of the ring.
		var base tcell.Color
		if lb.lineageColored {
			base = tcellFromRGBA(lineageColor(lb.lineageKey, lb.category))
		} else {
			base = theme.Color(lb.role)
		}
		if lb.bright {
			base = brightenTcell(base, 0.35)
		}
		style := tcell.StyleDefault.Foreground(base)
		if lb.kind == labelTitle || lb.kind == labelCapital {
			style = style.Bold(true)
		}

		runes := []rune(lb.text)
		// Compute the starting column from the alignment.
		startCol := lb.cx
		switch lb.align {
		case alignRight:
			startCol = lb.cx - len(runes) + 1
		case alignCenter:
			startCol = lb.cx - len(runes)/2
		}
		row := lb.cy
		if row < 0 || row >= rows {
			continue
		}
		for i, r := range runes {
			col := startCol + i
			if col < 0 || col >= cols {
				continue // clip per-glyph so a long label degrades gracefully at edges
			}
			screen.SetContent(offX+col, offY+row, r, nil, style)
		}
	}
}

// tcellFromRGBA converts an image/color.RGBA to a true-RGB tcell.Color, so the
// overlay can reuse the image-space lineageColor for a district banner.
func tcellFromRGBA(c color.RGBA) tcell.Color {
	return tcell.NewRGBColor(int32(c.R), int32(c.G), int32(c.B))
}

// brightenTcell lightens a tcell color toward white by t in [0,1]. Used for the
// at-war civ marker; mirrors the image-space brighten but stays in tcell space so
// the overlay color math doesn't round-trip through image/color.
func brightenTcell(c tcell.Color, t float64) tcell.Color {
	r, g, b := c.RGB()
	lerp := func(v int32) int32 {
		f := float64(v) + (255-float64(v))*t
		if f < 0 {
			f = 0
		}
		if f > 255 {
			f = 255
		}
		return int32(f)
	}
	return tcell.NewRGBColor(lerp(r), lerp(g), lerp(b))
}

// discoveredCivs resolves the diplomacy snapshot to a stable, status-colored list
// of markers. Sorted by name so the ring is stable frame-to-frame (map iteration
// is otherwise random and would make the markers jump). Pure read of the snapshot.
func discoveredCivs(state game.GameState) []civMarker {
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
	out := make([]civMarker, 0, len(keys))
	for _, k := range keys {
		f := facs[k]
		out = append(out, civMarker{
			name:  f.Name,
			role:  civStatusRole(f),
			atWar: f.AtWar,
		})
	}
	return out
}

// civStatusRole maps a faction's diplomatic standing to a theme role. At-war and
// rival/embargo both read Negative (the bright/"!" treatment separates war); allied
// reads Positive; friendly a quieter Label (in-family but not a full ally green);
// neutral and anything else read Dim.
func civStatusRole(f game.FactionInfo) theme.Role {
	if f.AtWar {
		return theme.RoleNegative
	}
	switch f.Status {
	case "allied":
		return theme.RolePositive
	case "rival", "embargo":
		return theme.RoleNegative
	case "friendly":
		return theme.RoleLabel
	default: // "neutral" and unknown
		return theme.RoleDim
	}
}

// splitSlots divides n civs between two border columns within a usable height,
// returning (left, right) slot counts. Caps each side at the usable rows so we never
// try to place more markers than there are rows — extras are simply not placed.
func splitSlots(n, usable int) (left, right int) {
	left = (n + 1) / 2
	right = n - left
	if left > usable {
		left = usable
	}
	if right > usable {
		right = usable
	}
	return left, right
}

// packCell packs a side id and a row into a single int key for the collision set.
func packCell(side, row int) int { return side<<20 | (row & 0xfffff) }

// maxLabelLen returns how many characters a label at column col may use without
// running off the cols-wide grid, given its alignment. Center labels get the
// smaller of the left/right runway (so they stay centered and on-screen).
func maxLabelLen(col, cols int, al alignment) int {
	switch al {
	case alignRight:
		return col + 1
	case alignCenter:
		// Half-runway on the tighter side, doubled, so the centered text fits.
		leftRun := col
		rightRun := cols - 1 - col
		half := leftRun
		if rightRun < half {
			half = rightRun
		}
		return half*2 + 1
	default: // alignLeft
		return cols - col
	}
}

// truncLabel shortens s to at most max characters, dropping a trailing '…' if it had
// to cut (and there's room for the ellipsis). Returns "" if max <= 0.
func truncLabel(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return string(r[:1])
	}
	return string(r[:max-1]) + "…"
}

// pxToCellX converts an image pixel x to a cell column. The image is cols pixels
// wide (one pixel per cell horizontally), so this is the identity — provided as a
// named counterpart to pxToCellY for symmetry and to localize the assumption.
func pxToCellX(px int) int { return px }

// lineageLabel maps a lineage/category key to a short, human label for a district
// banner. citymap owns its own presentation vocabulary (config has no display-name
// table for lineages), so this is the single source for the on-map names. Unknown
// lineages fall back to a title-cased key so a new lineage still labels sensibly.
func lineageLabel(lineageKey, category string) string {
	switch category {
	case "wonder":
		return "Wonders"
	case "monument":
		return "Monuments"
	case "storage":
		return "Stores"
	case "diplomacy":
		return "Embassy"
	}
	if n, ok := lineageLabels[lineageKey]; ok {
		return n
	}
	return titleCaseKey(lineageKey)
}

// lineageLabels is the on-map display name for each production lineage. Kept terse
// so the banner fits a cluster; mirrors the lineages in lineageRoleBase (palette.go)
// one-for-one. Adding a lineage there should add a label here.
var lineageLabels = map[string]string{
	"food":                  "Food",
	"organic_extraction":    "Forestry",
	"geological_extraction": "Mining",
	"metallurgy":            "Metalworks",
	"energy":                "Energy",
	"engineering":           "Industry",
	"knowledge":             "Academy",
	"faith":                 "Temple",
	"culture_arts":          "Arts",
	"military":              "Military",
	"harbor":                "Harbor",
	"trade":                 "Market",
	"hacker":                "Cyber",
	"astronaut":             "Spaceport",
	"housing":               "Housing",
}

// titleCaseKey turns a snake_case lineage key into a spaced, title-cased label as a
// fallback for any lineage missing from lineageLabels. e.g. "deep_mining" → "Deep
// Mining". Pure ASCII handling — lineage keys are snake_case ASCII.
func titleCaseKey(key string) string {
	if key == "" {
		return ""
	}
	r := []rune(key)
	out := make([]rune, 0, len(r))
	upNext := true
	for _, c := range r {
		if c == '_' {
			out = append(out, ' ')
			upNext = true
			continue
		}
		if upNext && c >= 'a' && c <= 'z' {
			c -= 32
		}
		upNext = false
		out = append(out, c)
	}
	return string(out)
}

// roundPx rounds a float pixel coordinate to the nearest int, for the trade-lane
// border math which works in float polar space before stamping. Defined here so the
// overlay/trade geometry shares one rounding rule.
func roundPx(f float64) int { return int(math.Round(f)) }
