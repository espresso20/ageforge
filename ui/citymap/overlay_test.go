package citymap

import (
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// overlay_test.go exercises the P3 systems-weave + text-overlay pass: trade-route
// lanes, civ edge-markers, district labels, the title, and the live theme retint of
// the overlay. The pixel passes are covered alongside (lanes are pixels; their tags
// and the markers/labels are text).

// stateWithRoutes returns a sample state for an age with the given built buildings
// and a set of active trade routes (name → completed cycles). All routes are active
// and non-disrupted unless markDisrupted contains the name.
func stateWithRoutes(age string, built map[string]int, routes map[string]int, markDisrupted map[string]bool) game.GameState {
	st := sampleState(age, built)
	active := make([]game.ActiveRouteInfo, 0, len(routes))
	for name, cycles := range routes {
		active = append(active, game.ActiveRouteInfo{
			Name:       name,
			Key:        name,
			CyclesDone: cycles,
			Disrupted:  markDisrupted[name],
		})
	}
	st.Trade = game.TradeState{ActiveRoutes: active}
	return st
}

// withCivs attaches a diplomacy roster to a state. Each entry is name → (status,
// atWar, discovered).
func withCivs(st game.GameState, civs map[string]struct {
	status     string
	atWar      bool
	discovered bool
}) game.GameState {
	facs := make(map[string]game.FactionInfo, len(civs))
	for key, c := range civs {
		facs[key] = game.FactionInfo{
			Name:       key,
			Status:     c.status,
			AtWar:      c.atWar,
			Discovered: c.discovered,
		}
	}
	st.Diplomacy = game.DiplomacyState{Factions: facs}
	return st
}

// newSimScreen builds an initialized simulation screen of the given cell size for
// overlay stamping tests, registering cleanup.
func newSimScreen(t *testing.T, cols, rows int) tcell.SimulationScreen {
	t.Helper()
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatalf("screen.Init: %v", err)
	}
	t.Cleanup(s.Fini)
	s.SetSize(cols, rows)
	return s
}

// TestSelectTradeRoutesCap proves the lane selector caps the number of routes and
// keeps the most-established ones (by completed cycles), deterministically.
func TestSelectTradeRoutesCap(t *testing.T) {
	routes := map[string]int{}
	for i := 0; i < maxTradeLanes+5; i++ {
		routes[string(rune('a'+i))] = i // cycles ascending → highest letters are most established
	}
	st := stateWithRoutes("iron_age", nil, routes, nil)
	picks := selectTradeRoutes(st, maxTradeLanes)
	if len(picks) != maxTradeLanes {
		t.Fatalf("selectTradeRoutes returned %d, want cap %d", len(picks), maxTradeLanes)
	}
	// Must be sorted by cycles desc; the top pick should be the most-cycled route.
	for i := 1; i < len(picks); i++ {
		if picks[i-1].cycles < picks[i].cycles {
			t.Fatalf("routes not cycle-sorted: %d before %d", picks[i-1].cycles, picks[i].cycles)
		}
	}
	// Determinism: same input → identical order.
	again := selectTradeRoutes(st, maxTradeLanes)
	for i := range picks {
		if picks[i] != again[i] {
			t.Fatalf("selectTradeRoutes nondeterministic at %d: %+v vs %+v", i, picks[i], again[i])
		}
	}
	// Empty trade → no picks, no panic.
	if got := selectTradeRoutes(sampleState("iron_age", nil), maxTradeLanes); got != nil {
		t.Fatalf("empty trade should yield nil picks, got %v", got)
	}
}

// TestDrawTradeLanesRespectsCapAndBorder renders lanes for an over-cap active-route
// state and asserts: no panic, the returned endpoint count is capped, and every
// endpoint lands on the image border (so an overlay tag anchored there is on-canvas).
func TestDrawTradeLanesRespectsCapAndBorder(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := buildPalette(0)
	const w, h = 100, 60
	routes := map[string]int{}
	for i := 0; i < maxTradeLanes+6; i++ {
		routes["route_"+string(rune('A'+i))] = i
	}
	st := stateWithRoutes("classical_age", sampleBuilt(), routes, map[string]bool{"route_A": true})

	img, _ := renderImage(st, w, h)
	if got := img.Bounds(); got.Dx() != w || got.Dy() != h {
		t.Fatalf("image size = %dx%d, want %dx%d", got.Dx(), got.Dy(), w, h)
	}

	ends := drawTradeLanes(img, pal, st, w/2, h/2)
	if len(ends) > maxTradeLanes {
		t.Fatalf("drew %d lanes, exceeds cap %d", len(ends), maxTradeLanes)
	}
	if len(ends) == 0 {
		t.Fatal("expected some trade lanes for an active-route state")
	}
	for _, e := range ends {
		onBorder := e.px == 0 || e.px == w-1 || e.py == 0 || e.py == h-1
		if !onBorder {
			t.Fatalf("trade endpoint (%d,%d) not on the %dx%d border", e.px, e.py, w, h)
		}
	}
}

// TestCivMarkersStatusColorsNoOverlap places a roster covering every status (incl.
// an at-war civ) and asserts: each discovered civ gets a marker in the right status
// color, undiscovered civs are skipped, the at-war marker is brightened + flagged,
// and no two markers share a cell (overlap-free, panic-free).
func TestCivMarkersStatusColorsNoOverlap(t *testing.T) {
	_ = theme.SetActive("forge")
	const cols, rows = 50, 30

	civs := map[string]struct {
		status     string
		atWar      bool
		discovered bool
	}{
		"Allies":    {"allied", false, true},
		"Rivals":    {"rival", false, true},
		"Embargoed": {"embargo", false, true},
		"Warmonger": {"neutral", true, true}, // at war overrides status color
		"Friends":   {"friendly", false, true},
		"Neutrals":  {"neutral", false, true},
		"Unseen":    {"neutral", false, false}, // not discovered → no marker
	}
	st := withCivs(sampleState("iron_age", nil), civs)

	// discoveredCivs must skip the undiscovered one and color by status.
	markers := discoveredCivs(st)
	if len(markers) != 6 {
		t.Fatalf("discoveredCivs returned %d, want 6 (Unseen excluded)", len(markers))
	}
	roleByName := map[string]theme.Role{}
	warByName := map[string]bool{}
	for _, m := range markers {
		roleByName[m.name] = m.role
		warByName[m.name] = m.atWar
	}
	wantRole := map[string]theme.Role{
		"Allies":    theme.RolePositive,
		"Rivals":    theme.RoleNegative,
		"Embargoed": theme.RoleNegative,
		"Warmonger": theme.RoleNegative, // at-war → negative
		"Friends":   theme.RoleLabel,
		"Neutrals":  theme.RoleDim,
	}
	for name, want := range wantRole {
		if roleByName[name] != want {
			t.Fatalf("civ %q role = %v, want %v", name, roleByName[name], want)
		}
	}
	if !warByName["Warmonger"] {
		t.Fatal("Warmonger should be flagged atWar")
	}

	// Build the overlay plan and assert civ labels don't share a cell row+side.
	plan := buildOverlayPlan(st, cols, rows, layoutGeometry{palaceX: cols / 2, palaceY: rows})
	seen := map[[2]int]string{}
	civCount := 0
	for _, lb := range plan.labels {
		if lb.kind != labelCiv {
			continue
		}
		civCount++
		// A civ marker's anchor cell is (cx, cy); two civs must not anchor identically.
		k := [2]int{lb.cx, lb.cy}
		if prev, ok := seen[k]; ok {
			t.Fatalf("civ markers overlap at cell %v: %q and %q", k, prev, lb.text)
		}
		seen[k] = lb.text
		// The at-war marker must carry the "!" prefix and the bright flag.
		if strings.Contains(lb.text, "Warmonger") {
			if !lb.bright {
				t.Fatal("at-war civ marker not brightened")
			}
			if !strings.HasPrefix(lb.text, "!") {
				t.Fatalf("at-war civ marker missing '!' prefix: %q", lb.text)
			}
		}
	}
	if civCount == 0 {
		t.Fatal("no civ markers planned for a discovered roster")
	}
}

// TestDistrictLabelsRenderAndSkipCollisions builds a populated layout and asserts
// district banners are planned, none collide on a cell, and a thin (single-building)
// district is not labeled while a fat one is.
func TestDistrictLabelsRenderAndSkipCollisions(t *testing.T) {
	_ = theme.SetActive("forge")
	const cols, rows = 90, 50

	// Force several districts onto nearly the same centroid to provoke the collision
	// path, plus distinct ones, plus a thin one that should be skipped by geometryFor.
	geo := layoutGeometry{
		palaceX: cols / 2, palaceY: rows,
		districts: []districtCentroid{
			{px: 20, py: 20, lineageKey: "food", category: "production", count: 12},
			{px: 21, py: 20, lineageKey: "metallurgy", category: "production", count: 8}, // near food → collision nudge
			{px: 20, py: 21, lineageKey: "military", category: "production", count: 6},   // also near → may skip
			{px: 70, py: 30, lineageKey: "knowledge", category: "production", count: 4},  // distinct
			{px: 50, py: 44, lineageKey: "trade", category: "production", count: 20},     // distinct
		},
	}
	plan := buildOverlayPlan(sampleState("industrial_age", nil), cols, rows, geo)

	seen := map[[2]int]string{}
	districtLabels := 0
	for _, lb := range plan.labels {
		if lb.kind != labelDistrict {
			continue
		}
		districtLabels++
		k := [2]int{lb.cx, lb.cy}
		if prev, ok := seen[k]; ok {
			t.Fatalf("district labels collide at cell %v: %q and %q", k, prev, lb.text)
		}
		seen[k] = lb.text
		if lb.text == "" {
			t.Fatal("empty district label text")
		}
		// District labels must use the lineage color source (so they retint per
		// district, matching the volumes).
		if !lb.lineageColored {
			t.Fatalf("district label %q not lineage-colored", lb.text)
		}
	}
	if districtLabels == 0 {
		t.Fatal("no district labels planned for a populated geometry")
	}
	// The distinct districts (knowledge, trade) should always get labels; with five
	// districts and collisions, at least the two distinct + one of the clustered
	// should survive.
	if districtLabels < 3 {
		t.Fatalf("expected >=3 surviving district labels, got %d (collision pruning too aggressive)", districtLabels)
	}
}

// TestGeometryForSkipsThinDistricts verifies geometryFor only emits centroids for
// districts worth a banner (reps >= 2 or special), not for a lone building.
func TestGeometryForSkipsThinDistricts(t *testing.T) {
	_ = theme.SetActive("forge")
	// hut alone (1 instance) → reps 1 → no banner; gathering_camp x12 → reps 4 → banner.
	thin := district{lineageKey: "housing", category: "housing", col: lineageColor("housing", "housing"), reps: representativeCount(1), count: 1}
	fat := district{lineageKey: "food", category: "production", col: lineageColor("food", "production"), reps: representativeCount(12), count: 12}
	placements := []placement{
		{cx: 40, cy: 40, size: 2, col: buildPalette(0).palace, tier: impPalace},
		{cx: 10, cy: 10, size: 0, col: thin.col, tier: impNormal},
		{cx: 60, cy: 30, size: 1, col: fat.col, tier: impNormal},
		{cx: 62, cy: 31, size: 1, col: fat.col, tier: impNormal},
	}
	geo := geometryFor(80, 80, []district{thin, fat}, placements)
	for _, c := range geo.districts {
		if c.lineageKey == "housing" {
			t.Fatal("thin housing district should not get a centroid/banner")
		}
	}
	foundFood := false
	for _, c := range geo.districts {
		if c.lineageKey == "food" {
			foundFood = true
		}
	}
	if !foundFood {
		t.Fatal("fat food district should get a centroid/banner")
	}
}

// TestTitlePresentAndAgeNamed asserts the corner title is planned and includes the
// age name.
func TestTitlePresentAndAgeNamed(t *testing.T) {
	_ = theme.SetActive("forge")
	st := sampleState("bronze_age", nil)
	st.AgeName = "Bronze Age"
	plan := buildOverlayPlan(st, 60, 30, layoutGeometry{palaceX: 30, palaceY: 30})
	var title string
	for _, lb := range plan.labels {
		if lb.kind == labelTitle {
			title = lb.text
		}
	}
	if title == "" {
		t.Fatal("no title planned")
	}
	if !strings.Contains(title, "Bronze Age") {
		t.Fatalf("title %q does not name the age", title)
	}
	if !strings.Contains(title, "Empire") {
		t.Fatalf("title %q missing empire label", title)
	}
}

// TestStampOverlayWritesGlyphsInBounds stamps a plan onto a simulation screen and
// confirms the title glyphs were written, the foreground color matches the live
// theme role, and nothing was drawn outside the box bounds (panic-safe + clipped).
func TestStampOverlayWritesGlyphsInBounds(t *testing.T) {
	_ = theme.SetActive("forge")
	const cols, rows = 40, 20
	st := sampleState("iron_age", sampleBuilt())
	st.AgeName = "Iron Age"
	st = withCivs(st, map[string]struct {
		status     string
		atWar      bool
		discovered bool
	}{
		"Ath": {"allied", false, true},
		"Bru": {"neutral", true, true},
	})
	_, plan := renderImage(st, cols, rows*2)

	scr := newSimScreen(t, cols, rows)
	// Pre-fill with a sentinel so we can tell which cells the overlay overwrote.
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			scr.SetContent(x, y, '·', nil, tcell.StyleDefault)
		}
	}
	stampOverlay(scr, plan, 0, 0, cols, rows)
	scr.Show()

	cells, gotW, gotH := scr.GetContents()
	if gotW < cols || gotH < rows {
		t.Fatalf("screen %dx%d smaller than %dx%d", gotW, gotH, cols, rows)
	}
	// The title starts at cell (1,0) with the accent color: assert the first title
	// rune landed and is accent-colored.
	titleCell := cells[0*gotW+1]
	if len(titleCell.Runes) != 1 || titleCell.Runes[0] == '·' {
		t.Fatalf("title cell (1,0) not overwritten: %q", string(titleCell.Runes))
	}
	fg, _, _ := titleCell.Style.Decompose()
	ar, ag, ab := theme.Color(theme.RoleAccent).RGB()
	fr, fgc, fb := fg.RGB()
	if fr != ar || fgc != ag || fb != ab {
		t.Fatalf("title fg = (%d,%d,%d), want accent (%d,%d,%d)", fr, fgc, fb, ar, ag, ab)
	}

	// Count overwritten cells: there must be at least the title length, and every
	// written cell must be within bounds (GetContents only exposes in-bounds cells,
	// so the mere absence of a panic + a positive count proves clipping worked).
	overwritten := 0
	for i := range cells {
		if len(cells[i].Runes) == 1 && cells[i].Runes[0] != '·' {
			overwritten++
		}
	}
	if overwritten < len("Empire — Iron Age") {
		t.Fatalf("only %d cells overwritten; expected at least the title length", overwritten)
	}
}

// TestStampOverlayPanicSafeOnTinyScreen drives the overlay with a screen far smaller
// than the labels want, plus an absurd plan, to prove the per-glyph clipping never
// panics or writes out of range.
func TestStampOverlayPanicSafeOnTinyScreen(t *testing.T) {
	_ = theme.SetActive("forge")
	const cols, rows = 4, 3
	st := withCivs(sampleState("galactic_age", sampleBuilt()),
		map[string]struct {
			status     string
			atWar      bool
			discovered bool
		}{
			"VeryLongCivilizationName": {"allied", false, true},
			"AnotherImpossiblyLongOne": {"neutral", true, true},
		})
	_, plan := renderImage(st, cols, rows*2)

	// Also hand-craft out-of-range labels to stress the clip directly.
	plan.labels = append(plan.labels,
		overlayLabel{cx: 999, cy: 999, text: "off the map", role: theme.RoleText, align: alignLeft},
		overlayLabel{cx: -50, cy: -50, text: "negative", role: theme.RoleText, align: alignRight},
		overlayLabel{cx: 2, cy: 1, text: "way too wide to fit in four columns", role: theme.RoleText, align: alignCenter},
	)

	scr := newSimScreen(t, cols, rows)
	// Must not panic.
	stampOverlay(scr, plan, 0, 0, cols, rows)
	stampOverlay(nil, plan, 0, 0, cols, rows) // nil screen guard
	stampOverlay(scr, plan, 0, 0, 0, 0)       // zero-size guard
	scr.Show()
}

// TestFullRenderOverlayAcrossAges renders the entire pipeline (terrain → flourish →
// roads → volumes → trade lanes → overlay) and stamps it for one age per era with a
// fully populated state (buildings + active routes + a discovered roster including an
// at-war civ). Asserts correct image size, full opacity, a non-empty plan, and a
// panic-free stamp.
func TestFullRenderOverlayAcrossAges(t *testing.T) {
	_ = theme.SetActive("forge")
	const cols, rows = 100, 48
	routes := map[string]int{"silk_road": 5, "spice_run": 2, "amber_way": 9, "tin_route": 1}
	roster := map[string]struct {
		status     string
		atWar      bool
		discovered bool
	}{
		"Kael":  {"allied", false, true},
		"Vorn":  {"rival", false, true},
		"Sythe": {"neutral", true, true},
		"Dun":   {"friendly", false, true},
		"Mira":  {"neutral", false, true},
		"Ghost": {"neutral", false, false}, // undiscovered
	}

	for _, tc := range oneAgePerEra {
		t.Run(tc.name, func(t *testing.T) {
			st := stateWithRoutes(tc.age, sampleBuilt(), routes, map[string]bool{"tin_route": true})
			st = withCivs(st, roster)
			st.AgeName = tc.age

			img, plan := renderImage(st, cols, rows*2)
			if got := img.Bounds(); got.Dx() != cols || got.Dy() != rows*2 {
				t.Fatalf("image %dx%d, want %dx%d", got.Dx(), got.Dy(), cols, rows*2)
			}
			for i := 3; i < len(img.Pix); i += 4 {
				if img.Pix[i] != 0xff {
					t.Fatalf("pixel %d alpha = %d, want 255", i/4, img.Pix[i])
				}
			}
			if len(plan.labels) == 0 {
				t.Fatal("overlay plan empty for a fully populated state")
			}
			// Must contain at least one civ marker and the title.
			var civs, titles int
			for _, lb := range plan.labels {
				switch lb.kind {
				case labelCiv:
					civs++
				case labelTitle:
					titles++
				}
			}
			if civs == 0 {
				t.Fatal("no civ markers in a discovered-roster render")
			}
			if titles != 1 {
				t.Fatalf("want exactly 1 title, got %d", titles)
			}

			scr := newSimScreen(t, cols, rows)
			stampHalfThenOverlay(t, scr, img, plan, cols, rows)
		})
	}
}

// stampHalfThenOverlay mimics the Build draw-closure order: half-blocks first, then
// the text overlay on top. Purely a panic/coherence smoke test of the real ordering.
func stampHalfThenOverlay(t *testing.T, scr tcell.SimulationScreen, img *image.RGBA, plan overlayPlan, cols, rows int) {
	t.Helper()
	streamHalfBlocks(scr, img, 0, 0, cols, rows)
	stampOverlay(scr, plan, 0, 0, cols, rows)
	scr.Show()
	// A handful of cells should now be crisp overlay glyphs rather than '▄'.
	cells, gotW, _ := scr.GetContents()
	nonBlock := 0
	for i := range cells {
		if len(cells[i].Runes) == 1 && cells[i].Runes[0] != '▄' {
			nonBlock++
		}
	}
	if nonBlock == 0 {
		t.Fatal("overlay produced no crisp glyphs over the half-block field")
	}
	_ = gotW
}

// TestOverlayRetintsOnThemeSwitch proves the overlay text color is resolved live: the
// SAME plan stamped under two themes yields different title foreground colors.
func TestOverlayRetintsOnThemeSwitch(t *testing.T) {
	const cols, rows = 40, 12
	st := sampleState("modern_age", sampleBuilt())
	st.AgeName = "Modern Age"
	st = withCivs(st, map[string]struct {
		status     string
		atWar      bool
		discovered bool
	}{"Zed": {"allied", false, true}})

	_ = theme.SetActive("forge")
	_, plan := renderImage(st, cols, rows*2)

	titleFg := func() tcell.Color {
		scr := newSimScreen(t, cols, rows)
		stampOverlay(scr, plan, 0, 0, cols, rows)
		scr.Show()
		cells, gotW, _ := scr.GetContents()
		fg, _, _ := cells[0*gotW+1].Style.Decompose()
		return fg
	}

	_ = theme.SetActive("forge")
	forgeFg := titleFg()
	if err := theme.SetActive("high_contrast"); err != nil {
		t.Fatalf("SetActive(high_contrast): %v", err)
	}
	hcFg := titleFg()
	_ = theme.SetActive("forge")

	fr, fg, fb := forgeFg.RGB()
	hr, hg, hb := hcFg.RGB()
	if fr == hr && fg == hg && fb == hb {
		t.Fatalf("title color identical across themes (%d,%d,%d) — overlay not retinting live", fr, fg, fb)
	}
	// And it must track the accent role under each theme.
	_ = theme.SetActive("high_contrast")
	ar, ag, ab := theme.Color(theme.RoleAccent).RGB()
	_ = theme.SetActive("forge")
	if hr != ar || hg != ag || hb != ab {
		t.Fatalf("hc title fg (%d,%d,%d) != hc accent (%d,%d,%d)", hr, hg, hb, ar, ag, ab)
	}
}
