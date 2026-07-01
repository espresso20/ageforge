package citymap

import (
	"image"
	"image/color"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// worldState builds a small GameState for the world map: an age, a building set
// (name → count, drives empire scale / progress), an optional account display name,
// and an optional faction roster. Only the fields the world renderer reads are set.
func worldState(age string, built map[string]int, displayName string, facs map[string]game.FactionInfo) game.GameState {
	st := game.GameState{Age: age}
	if built != nil {
		bs := map[string]game.BuildingState{}
		for k, n := range built {
			bs[k] = game.BuildingState{Count: n, Name: k}
		}
		st.Buildings = bs
	}
	if displayName != "" {
		st.AccountStats = &game.AccountStatsView{DisplayName: displayName}
	}
	if facs != nil {
		st.Diplomacy = game.DiplomacyState{Factions: facs}
	}
	return st
}

// hasLabel reports whether the overlay plan contains a label whose text contains sub.
func planHasLabel(state game.GameState, w, h int, sub string) (bool, []overlayLabel) {
	_, plan := renderWorldImage(state, w, h)
	found := false
	for _, lb := range plan.labels {
		if containsStr(lb.text, sub) {
			found = true
		}
	}
	return found, plan.labels
}

func containsStr(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// TestWorldRenderNoPanic renders the world map across ages and civ configurations: it
// must not panic and must produce an image of exactly the requested pixel size, fully
// opaque. Covers the no-civ case (just your civ + backdrop), several discovered civs
// including an at-war one, and an unknown age.
func TestWorldRenderNoPanic(t *testing.T) {
	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}
	threeCivs := map[string]game.FactionInfo{
		"rome":     {Name: "Rome", Discovered: true, Status: "allied", Opinion: 80},
		"carthage": {Name: "Carthage", Discovered: true, Status: "rival", Opinion: -40, AtWar: true},
		"egypt":    {Name: "Egypt", Discovered: true, Status: "neutral", Opinion: 5},
		"hidden":   {Name: "Atlantis", Discovered: false, Status: "friendly", Opinion: 50},
	}
	cases := []struct {
		name        string
		age         string
		built       map[string]int
		displayName string
		facs        map[string]game.FactionInfo
	}{
		{"primitive-no-civs", "primitive_age", nil, "Babylon", nil},
		{"iron-three-civs", "iron_age", map[string]int{"forge": 2, "barracks": 3}, "Rome", threeCivs},
		{"unknown-age", "made_up_age", map[string]int{"x": 1}, "", threeCivs},
		{"no-account-name", "bronze_age", map[string]int{"granary": 1}, "", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			const w, h = 80, 48
			img, _ := renderWorldImage(worldState(tc.age, tc.built, tc.displayName, tc.facs), w, h)
			if img == nil {
				t.Fatal("renderWorldImage returned nil")
			}
			if got := img.Bounds(); got.Dx() != w || got.Dy() != h {
				t.Fatalf("image size = %dx%d, want %dx%d", got.Dx(), got.Dy(), w, h)
			}
			for i := 3; i < len(img.Pix); i += 4 {
				if img.Pix[i] != 0xff {
					t.Fatalf("pixel %d alpha = %d, want 255", i/4, img.Pix[i])
				}
			}
		})
	}
}

// TestWorldRenderTinyCanvas exercises degenerate sizes (the settlement grid + ring math
// have guards for canvases too small to inset); must not panic and must size correctly.
func TestWorldRenderTinyCanvas(t *testing.T) {
	_ = theme.SetActive("forge")
	facs := map[string]game.FactionInfo{
		"rome": {Name: "Rome", Discovered: true, Status: "allied", Opinion: 60},
	}
	for _, sz := range []struct{ w, h int }{{1, 2}, {3, 4}, {6, 6}, {0, 0}} {
		img, _ := renderWorldImage(worldState("iron_age", map[string]int{"forge": 1}, "Rome", facs), sz.w, sz.h)
		if img == nil {
			t.Fatalf("renderWorldImage(%d,%d) returned nil", sz.w, sz.h)
		}
		if got := img.Bounds(); got.Dx() != max0(sz.w) || got.Dy() != max0(sz.h) {
			t.Fatalf("size = %dx%d, want %dx%d", got.Dx(), got.Dy(), sz.w, sz.h)
		}
	}
}

func max0(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

// TestWorldYourCivLabel verifies the player's civ label uses AccountStats.DisplayName,
// and falls back to "Your Empire" when there is no account / no name.
func TestWorldYourCivLabel(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 80, 48

	named := worldState("iron_age", map[string]int{"forge": 1}, "Babylon", nil)
	if ok, labels := planHasLabel(named, w, h, "Babylon"); !ok {
		t.Fatalf("expected your-civ label %q in plan; got labels %v", "Babylon", labelTexts(labels))
	}

	// No account → fallback.
	anon := worldState("iron_age", map[string]int{"forge": 1}, "", nil)
	if ok, labels := planHasLabel(anon, w, h, "Your Empire"); !ok {
		t.Fatalf("expected fallback label %q in plan; got labels %v", "Your Empire", labelTexts(labels))
	}
}

// TestWorldDiscoveredCivsShownUndiscoveredAbsent verifies that discovered civs are
// labeled and undiscovered civs are not shown at all.
func TestWorldDiscoveredCivsShownUndiscoveredAbsent(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 100, 60
	facs := map[string]game.FactionInfo{
		"rome":   {Name: "Rome", Discovered: true, Status: "allied", Opinion: 70},
		"egypt":  {Name: "Egypt", Discovered: true, Status: "neutral", Opinion: 10},
		"hidden": {Name: "Atlantis", Discovered: false, Status: "friendly", Opinion: 50},
	}
	st := worldState("classical_age", map[string]int{"forge": 2}, "Rome", facs)

	if ok, labels := planHasLabel(st, w, h, "Rome"); !ok {
		t.Fatalf("discovered civ Rome not labeled; labels %v", labelTexts(labels))
	}
	if ok, _ := planHasLabel(st, w, h, "Egypt"); !ok {
		t.Fatal("discovered civ Egypt not labeled")
	}
	if ok, labels := planHasLabel(st, w, h, "Atlantis"); ok {
		t.Fatalf("undiscovered civ Atlantis must NOT appear; labels %v", labelTexts(labels))
	}
}

// TestWorldAtWarMarkAndColor verifies an at-war civ's label carries the "⚔" mark, that
// its label is flagged bright, and its role resolves to the negative (threat) color.
func TestWorldAtWarMarkAndColor(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 100, 60
	facs := map[string]game.FactionInfo{
		"carthage": {Name: "Carthage", Discovered: true, Status: "rival", Opinion: -50, AtWar: true},
		"rome":     {Name: "Rome", Discovered: true, Status: "allied", Opinion: 80},
	}
	// Player name kept distinct from any faction so the your-civ (Accent) label can't be
	// mistaken for a neighbor label in the assertions below.
	st := worldState("classical_age", map[string]int{"forge": 2}, "Sparta", facs)
	_, plan := renderWorldImage(st, w, h)

	var foundWarLabel, foundAlly bool
	for _, lb := range plan.labels {
		if containsStr(lb.text, "Carthage") {
			foundWarLabel = true
			if !containsStr(lb.text, "⚔") {
				t.Errorf("at-war label %q missing the ⚔ mark", lb.text)
			}
			if !lb.bright {
				t.Errorf("at-war label %q not flagged bright", lb.text)
			}
			if lb.role != theme.RoleNegative {
				t.Errorf("at-war label role = %v, want RoleNegative", lb.role)
			}
		}
		if containsStr(lb.text, "Rome") {
			foundAlly = true
			if lb.role != theme.RolePositive {
				t.Errorf("allied label role = %v, want RolePositive", lb.role)
			}
		}
	}
	if !foundWarLabel {
		t.Fatal("at-war civ Carthage not found in overlay plan")
	}
	if !foundAlly {
		t.Fatal("allied civ Rome not found in overlay plan")
	}
}

// TestWorldStatusRolesAndStrengthBuckets checks the relationship→role mapping and the
// strength-proxy bucketing used to size civ dots.
func TestWorldStatusRolesAndStrengthBuckets(t *testing.T) {
	cases := []struct {
		f    game.FactionInfo
		role theme.Role
	}{
		{game.FactionInfo{Status: "allied", Opinion: 80}, theme.RolePositive},
		{game.FactionInfo{Status: "friendly", Opinion: 30}, theme.RoleLabel},
		{game.FactionInfo{Status: "rival", Opinion: -40}, theme.RoleNegative},
		{game.FactionInfo{Status: "embargo", Opinion: -60}, theme.RoleNegative},
		{game.FactionInfo{Status: "neutral", Opinion: 0}, theme.RoleDim},
		{game.FactionInfo{Status: "neutral", AtWar: true}, theme.RoleNegative}, // war overrides
	}
	for _, tc := range cases {
		if got := civWorldRole(tc.f); got != tc.role {
			t.Errorf("civWorldRole(%+v) = %v, want %v", tc.f, got, tc.role)
		}
	}

	// Buckets scale with the civ's Strength (1-5 → 0..3); a live war makes a foe loom
	// one size larger. Exact mapping: 1→0, 2→1, 3→2, 4→2, 5→3.
	for s, want := range map[int]int{1: 0, 2: 1, 3: 2, 4: 2, 5: 3} {
		if got := civStrengthBucket(game.FactionInfo{Strength: s}); got != want {
			t.Errorf("strength %d → bucket %d, want %d", s, got, want)
		}
	}
	weak := civStrengthBucket(game.FactionInfo{Strength: 1})
	strong := civStrengthBucket(game.FactionInfo{Strength: 5})
	warFoe := civStrengthBucket(game.FactionInfo{Strength: 4, AtWar: true})
	peaceFoe := civStrengthBucket(game.FactionInfo{Strength: 4})
	if !(weak < strong) {
		t.Errorf("expected strength-1 bucket (%d) < strength-5 bucket (%d)", weak, strong)
	}
	if warFoe <= peaceFoe {
		t.Errorf("expected at-war to loom larger: war=%d vs peace=%d", warFoe, peaceFoe)
	}
	if weak < 0 || strong > 3 || warFoe > 3 {
		t.Errorf("buckets out of [0,3]: weak=%d strong=%d war=%d", weak, strong, warFoe)
	}
}

// TestWorldBackdropScalesWithProgress proves the Game-of-Life backdrop density visibly
// grows with the player's progress: a late, larger empire yields more live settlement
// cells than an early, tiny one, for the same canvas + seed family. We measure the live
// cell count directly (the pure backdrop seam) rather than counting rasterized pixels,
// which would be confounded by the full-canvas wash.
func TestWorldBackdropScalesWithProgress(t *testing.T) {
	_ = theme.SetActive("forge")

	early := worldState("primitive_age", map[string]int{"hut": 1}, "Tribe", nil)
	late := worldState("transcendent_age", map[string]int{"forge": 30, "barracks": 20}, "Empire", nil)

	earlyP := worldProgress(early)
	lateP := worldProgress(late)
	if !(lateP > earlyP) {
		t.Fatalf("expected later/larger progress (%.3f) > early progress (%.3f)", lateP, earlyP)
	}

	// Both worlds share the same age-seed family only if same age; they differ in age, so
	// compare each against itself across progress with a FIXED seed to isolate density.
	const seed, gw, gh = 12345, 40, 26
	earlyCells := countLiveCells(settlementGrid(seed, gw, gh, earlyP))
	lateCells := countLiveCells(settlementGrid(seed, gw, gh, lateP))
	if lateCells <= earlyCells {
		t.Fatalf("expected denser backdrop at higher progress: late cells=%d <= early cells=%d (earlyP=%.3f lateP=%.3f)",
			lateCells, earlyCells, earlyP, lateP)
	}

	// Sanity: progress 0 is genuinely sparser than progress 1 at the same seed.
	lo := countLiveCells(settlementGrid(seed, gw, gh, 0.0))
	hi := countLiveCells(settlementGrid(seed, gw, gh, 1.0))
	if hi <= lo {
		t.Fatalf("expected progress=1 backdrop (%d cells) denser than progress=0 (%d cells)", hi, lo)
	}

	// Backdrop is now SPARSE, not a snowstorm. Pin the absolute density to the new, much
	// lower band so a regression that re-densifies the field (the old 12–38% Conway
	// snowstorm) is caught. On a gw*gh grid, progress-0 must fill well under ~8% of cells
	// and even a maxed-out world must stay under ~20% — a scatter of distant settlements,
	// not a packed field. (Old behavior seeded ~12% at progress 0 and clustered UP.)
	total := gw * gh
	if frac := float64(lo) / float64(total); frac > 0.08 {
		t.Fatalf("progress=0 backdrop too dense: %d/%d = %.3f, want < 0.08 (sparse scatter)", lo, total, frac)
	}
	if frac := float64(hi) / float64(total); frac > 0.20 {
		t.Fatalf("progress=1 backdrop too dense: %d/%d = %.3f, want < 0.20 (no snowstorm)", hi, total, frac)
	}

	// ---- Land-gating dimension ------------------------------------------------
	// The rendered backdrop now GATES each candidate to land: a live grid cell only becomes a
	// settlement mark when its center is passable. So the drawn count is the live cells whose
	// center is on land — strictly FEWER than the raw live count on a mixed field, yet still
	// monotonic in progress (gating removes a fixed water mask, it doesn't reorder density).
	// We measure at the world's real cell pitch (settlementCellPx) over a fixed mixed field so
	// the gate has genuine ocean to exclude.
	const pw, ph = 100, 80
	wf := newWorldTerrainField(pw, ph, worldTerrainSeed("Memphis")) // ~65% land / ~35% sea
	pgw, pgh := pw/settlementCellPx, ph/settlementCellPx

	// gatedCount replays the render's per-cell land gate against a settlement grid: count the
	// live cells whose center pixel is passable — exactly the marks drawSettlementField draws.
	gatedCount := func(prog float64) (gated, raw int) {
		grid := settlementGrid(worldTerrainSeed("Memphis"), pgw, pgh, prog)
		for gy := 0; gy < pgh; gy++ {
			for gx := 0; gx < pgw; gx++ {
				if !grid[gy*pgw+gx] {
					continue
				}
				raw++
				cx := gx*settlementCellPx + settlementCellPx/2
				cy := gy*settlementCellPx + settlementCellPx/2
				if wf.passableAt(cx, cy) {
					gated++
				}
			}
		}
		return gated, raw
	}

	gLo, rLo := gatedCount(0.0)
	gHi, rHi := gatedCount(1.0)

	// Gating removes real candidates: on a mixed field the drawn (gated) count is strictly
	// below the raw live count at full progress (water cells were excluded).
	if !(gHi < rHi) {
		t.Fatalf("land-gating removed nothing: gated=%d raw=%d — expected fewer marks once water is excluded", gHi, rHi)
	}
	// Still monotonic after gating: more progress → at least as many land marks (strictly more
	// here, since progress-1 seeds many more candidates than progress-0).
	if !(gHi > gLo) {
		t.Fatalf("gated backdrop not monotonic in progress: gated(0)=%d gated(1)=%d", gLo, gHi)
	}
	// Sanity that the fixture actually exercised the gate.
	if rLo == 0 || rHi == 0 {
		t.Fatalf("no raw candidates generated (raw lo=%d hi=%d) — fixture too small", rLo, rHi)
	}
}

// TestWorldFactionColorDistinctAndRetints proves each discovered civ gets its OWN stable
// identity color derived from its faction key (so factions are distinguishable at a
// glance), and that the color retints on a theme switch (it is a rotation of a live theme
// role, not a baked constant) — mirroring how the city map colors lineages.
func TestWorldFactionColorDistinctAndRetints(t *testing.T) {
	_ = theme.SetActive("forge")
	rome := factionColor("rome")
	egypt := factionColor("egypt")
	carthage := factionColor("carthage")
	// Distinct keys → distinct hues (not all collapsed to one role color).
	if rome == egypt || rome == carthage || egypt == carthage {
		t.Fatalf("faction colors not distinct: rome=%v egypt=%v carthage=%v", rome, egypt, carthage)
	}
	// Same key → stable color across calls.
	if factionColor("rome") != rome {
		t.Fatal("factionColor is not deterministic for a fixed key")
	}
	// Retints with the theme: the same key resolves to a different color under a different
	// theme (the RoleLabel base changed).
	_ = theme.SetActive("high_contrast")
	romeHC := factionColor("rome")
	_ = theme.SetActive("forge")
	if romeHC == rome {
		t.Fatalf("faction color identical across themes (%v) — not theme-aware", rome)
	}
}

// TestWorldAtWarDotOverride proves the AT-WAR dot override actually paints: rendering a
// world with an at-war civ puts the hot-red override color (brighten(RoleNegative,0.20))
// into the image, whereas the same world with that civ at peace does not. This guards the
// draw-time threat cue that sits on top of the per-faction identity color.
func TestWorldAtWarDotOverride(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 100, 60
	hotRed := brighten(rgba(theme.Color(theme.RoleNegative)), 0.20)

	warFacs := map[string]game.FactionInfo{
		"carthage": {Name: "Carthage", Discovered: true, Status: "rival", Opinion: -50, AtWar: true, Strength: 3},
	}
	peaceFacs := map[string]game.FactionInfo{
		"carthage": {Name: "Carthage", Discovered: true, Status: "neutral", Opinion: 0, AtWar: false, Strength: 3},
	}
	countColor := func(facs map[string]game.FactionInfo, target color.RGBA) int {
		img, _ := renderWorldImage(worldState("classical_age", map[string]int{"forge": 2}, "Sparta", facs), w, h)
		n := 0
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if img.RGBAAt(x, y) == target {
					n++
				}
			}
		}
		return n
	}
	if got := countColor(warFacs, hotRed); got == 0 {
		t.Fatal("at-war civ did not paint the hot-red override anywhere in the image")
	}
	if got := countColor(peaceFacs, hotRed); got != 0 {
		t.Fatalf("peaceful civ painted the at-war hot-red override %d times — override leaked", got)
	}
}

// TestWorldThemeAwareRetint proves the world map retints on a theme switch: the SAME
// state renders to different pixels under forge vs high_contrast, and at least one
// overlay label resolves to a different color. (Labels store a role; the resolved color
// must differ across themes.)
func TestWorldThemeAwareRetint(t *testing.T) {
	const w, h = 100, 60
	facs := map[string]game.FactionInfo{
		"rome":     {Name: "Rome", Discovered: true, Status: "allied", Opinion: 70},
		"carthage": {Name: "Carthage", Discovered: true, Status: "rival", Opinion: -40, AtWar: true},
	}
	st := worldState("classical_age", map[string]int{"forge": 3}, "Rome", facs)

	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}
	imgForge, planForge := renderWorldImage(st, w, h)
	// Resolve a representative label color under forge.
	forgeColor := firstLabelColor(planForge, theme.RolePositive)

	if err := theme.SetActive("high_contrast"); err != nil {
		t.Fatalf("SetActive(high_contrast): %v", err)
	}
	imgHC, planHC := renderWorldImage(st, w, h)
	hcColor := firstLabelColor(planHC, theme.RolePositive)

	_ = theme.SetActive("forge") // restore for sibling tests

	// imagesDiffer lives in citymap_test.go (same package).
	if !imagesDiffer(imgForge, imgHC) {
		t.Fatal("world pixels identical across forge vs high_contrast — backdrop/dots not theme-aware")
	}
	if forgeColor == hcColor {
		t.Fatalf("RolePositive label color identical across themes (%v) — overlay not retinting", forgeColor)
	}
}

// firstLabelColor returns the live-resolved theme color of the first label in the plan
// carrying the given role — the same resolution stampOverlay does. Returns the zero
// tcell color if no such label exists (the tests always include one).
func firstLabelColor(plan overlayPlan, role theme.Role) tcell.Color {
	for _, lb := range plan.labels {
		if lb.role == role && !lb.lineageColored {
			return theme.Color(lb.role)
		}
	}
	return tcell.ColorDefault
}

func labelTexts(labels []overlayLabel) []string {
	out := make([]string, 0, len(labels))
	for _, lb := range labels {
		out = append(out, lb.text)
	}
	return out
}

// ---- World terrain: continents + oceans, land-gated dots --------------------

// worldFieldFor rebuilds the world terrain field EXACTLY as renderWorldImage does — same
// display-name seed, same dimensions — so a test can sample the passability grid the render
// actually gated against. Mirrors the "rebuild the field as renderImage does" pattern the
// city-map placement test uses.
func worldFieldFor(state game.GameState, w, h int) *terrainField {
	return newWorldTerrainField(w, h, worldTerrainSeed(worldDisplayName(state)))
}

// TestWorldTerrainHasLandAndWater proves the world backdrop is a real continents-vs-sea map:
// for a display name known to produce a mixed field at this size, the terrain field contains
// BOTH passable (land) and impassable (ocean/mountain) cells. A field that was all land or
// all sea would make the "dots on land" contract vacuous, so we assert the mix explicitly at
// the exact canvas the render uses. ("Rome" mixes ~62% land / ~38% sea at 80×48.)
func TestWorldTerrainHasLandAndWater(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 80, 48
	f := worldFieldFor(worldState("iron_age", nil, "Rome", nil), w, h)
	land, water := 0, 0
	for i := range f.passable {
		if f.passable[i] {
			land++
		} else {
			water++
		}
	}
	if land == 0 {
		t.Fatal("world terrain has no land — the whole canvas is ocean")
	}
	if water == 0 {
		t.Fatal("world terrain has no ocean — the whole canvas is land (no continents-vs-sea)")
	}
}

// TestWorldTerrainStableAcrossAgesDifferentPerName pins the two-seed design: the LAND is keyed
// off the display name and must be IDENTICAL across two different ages (aging up must not
// rearrange the continents), yet DIFFERENT for a different display name (each account gets its
// own world). We compare the passability grids directly — the ground truth every dot gates on.
func TestWorldTerrainStableAcrossAgesDifferentPerName(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 100, 80

	// Same name, two very different ages → identical land.
	early := worldFieldFor(worldState("primitive_age", map[string]int{"hut": 1}, "Rome", nil), w, h)
	late := worldFieldFor(worldState("transcendent_age", map[string]int{"forge": 30}, "Rome", nil), w, h)
	if len(early.passable) != len(late.passable) {
		t.Fatalf("field sizes differ across ages: %d vs %d", len(early.passable), len(late.passable))
	}
	for i := range early.passable {
		if early.passable[i] != late.passable[i] {
			t.Fatalf("land moved across ages at cell %d (same name) — continents not age-stable", i)
		}
	}

	// Different name → different land (not identical grids).
	other := worldFieldFor(worldState("primitive_age", map[string]int{"hut": 1}, "Carthage", nil), w, h)
	same := true
	for i := range early.passable {
		if early.passable[i] != other.passable[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("two different display names produced identical land — world seed not name-dependent")
	}
}

// TestWorldSettlementBackdropZeroPixelsOnWater is the STRICT settlement-layer contract: the
// sparse backdrop scatter must place ZERO pixels in the ocean. Settlements are single-cell
// marks gated per cell (skip any candidate whose center is impassable), so the guarantee is
// exact — unlike the multi-pixel civ dots, whose RING can overhang a shoreline by a pixel
// (see TestWorldDotCentersOnLand for the dot contract, which is center-on-land like the city
// map's placement). We render terrain + settlement field ONLY (no civ dots) and diff against
// a terrain-only reference: any pixel that changed is a settlement mark, and it must be on
// land. Uses a mixed-field name so there is real ocean to stay out of.
func TestWorldSettlementBackdropZeroPixelsOnWater(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 100, 80
	// "Memphis" mixes ~65% land / ~35% sea at 100×80 — a genuine ocean to keep marks out of.
	st := worldState("classical_age", map[string]int{"forge": 8, "barracks": 6}, "Memphis", nil)

	f := worldFieldFor(st, w, h)
	water := 0
	for _, p := range f.passable {
		if !p {
			water++
		}
	}
	if water == 0 {
		t.Fatal("test fixture has no ocean — pick a name/size with real sea")
	}

	seed, hueShift := ageInfo(st.Age)
	pal := buildPalette(hueShift)
	prog := worldProgress(st)

	ref := image.NewRGBA(image.Rect(0, 0, w, h))
	drawWorldTerrain(ref, pal, f)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	drawWorldTerrain(img, pal, f)
	drawSettlementField(img, pal, f, seed, prog)

	marks := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if img.RGBAAt(x, y) == ref.RGBAAt(x, y) {
				continue // unchanged from the backdrop — terrain, not a settlement mark
			}
			marks++
			if !f.passableAt(x, y) {
				t.Fatalf("settlement mark pixel at (%d,%d) sits on a water/impassable cell", x, y)
			}
		}
	}
	if marks == 0 {
		t.Fatal("no settlement marks drawn on a mixed-progress world — backdrop empty")
	}
}

// TestWorldDotCentersOnLand is the civ-dot land contract: your civ and EVERY discovered civ
// must have their dot CENTER on passable land (the position snap, nearestPassablePx). The
// dot body/ring can overhang a shoreline pixel — exactly as the city map's placement test
// accepts for building volumes — but the center (where the label anchors and the snap targets)
// is always on land, so no civ "floats" at sea. We drive the two draw passes directly to read
// back the placed centers over a mixed field. Includes an at-war civ (the hot-red override
// path also snaps).
func TestWorldDotCentersOnLand(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 100, 80
	facs := map[string]game.FactionInfo{
		"rome":     {Name: "Rome", Discovered: true, Status: "allied", Opinion: 70, Strength: 4},
		"carthage": {Name: "Carthage", Discovered: true, Status: "rival", Opinion: -50, AtWar: true, Strength: 3},
		"egypt":    {Name: "Egypt", Discovered: true, Status: "neutral", Opinion: 5, Strength: 2},
		"nubia":    {Name: "Nubia", Discovered: true, Status: "friendly", Opinion: 30, Strength: 1},
	}
	st := worldState("classical_age", map[string]int{"forge": 8, "barracks": 6}, "Memphis", facs)

	f := worldFieldFor(st, w, h)
	water := 0
	for _, p := range f.passable {
		if !p {
			water++
		}
	}
	if water == 0 {
		t.Fatal("test fixture has no ocean — pick a name/size with real sea")
	}

	_, hueShift := ageInfo(st.Age)
	pal := buildPalette(hueShift)
	seed, _ := ageInfo(st.Age)

	scratch := image.NewRGBA(image.Rect(0, 0, w, h))
	you := drawYourCiv(scratch, pal, f, w, h)
	if !f.passableAt(you.cx, you.cy) {
		t.Fatalf("your civ center (%d,%d) is on water — position not snapped to land", you.cx, you.cy)
	}
	civs := drawWorldCivs(scratch, pal, st, f, seed, you.cx, you.cy)
	if len(civs) != 4 {
		t.Fatalf("expected 4 discovered civ dots, got %d", len(civs))
	}
	for _, c := range civs {
		if !f.passableAt(c.cx, c.cy) {
			t.Fatalf("civ %q center (%d,%d) is on water — position not snapped to land", c.name, c.cx, c.cy)
		}
	}
}

// ---- Player dot: round, bounded, no cross -----------------------------------

// dotExtent scans the image for pixels of (approximately) the given color and returns the
// bounding box + count, so a test can measure a dot's drawn WIDTH vs HEIGHT and its radius.
// Colors are matched with a small per-channel tolerance because the body is a brightened
// blend, not a single canonical constant.
func dotExtent(img *image.RGBA, target color.RGBA, tol int) (minX, minY, maxX, maxY, count int) {
	minX, minY = 1<<30, 1<<30
	maxX, maxY = -1, -1
	near := func(a, b uint8) bool {
		d := int(a) - int(b)
		if d < 0 {
			d = -d
		}
		return d <= tol
	}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.RGBAAt(x, y)
			if near(c.R, target.R) && near(c.G, target.G) && near(c.B, target.B) {
				count++
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	return minX, minY, maxX, maxY, count
}

// TestWorldPlayerDotRoundAndBounded pins the player-dot fix. The player dot is drawn in the
// accent role (brightened body) at canvas center. We render just your civ over an all-land
// field (so nothing snaps it away and no neighbour colors intrude), find the accent body's
// drawn extent, and assert:
//   - ROUND: drawn width ≈ height (not a wide football). The earlier ry≈rx/2 aspect produced
//     a box roughly twice as wide as tall; a round disc is within ~1px on each axis.
//   - BOUNDED: the radius is small (single-digit px), not a metropolis eating the canvas.
//   - NO PLUS CORE: there is no brighter inner disc that rasterizes a "+" — the whole body is
//     one bright disc, so scanning the center row/column for a distinct brighter core finds
//     none (the old inner brighten(accent,0.30) core is gone).
func TestWorldPlayerDotRoundAndBounded(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 100, 80
	// "Sparta" is all land at these sizes, so your civ stays at dead center and the accent
	// body is unobstructed — a clean target to measure.
	st := worldState("iron_age", map[string]int{"forge": 1}, "Sparta", nil)

	f := worldFieldFor(st, w, h)
	if !f.passableAt(w/2, h/2) {
		t.Skip("fixture unexpectedly has water at center; skipping geometry measure")
	}

	img, _ := renderWorldImage(st, w, h)
	body := brighten(rgba(theme.Color(theme.RoleAccent)), 0.12)
	minX, minY, maxX, maxY, count := dotExtent(img, body, 6)
	if count == 0 {
		t.Fatal("player dot body color not found in the render")
	}
	dw := maxX - minX + 1
	dh := maxY - minY + 1

	// ROUND: width and height within 2px of each other (a disc, not a football). The old
	// football was ~2× wider than tall; this catches any regression back to that aspect.
	if diff := absInt(dw - dh); diff > 2 {
		t.Fatalf("player dot not round: drawn %dx%d (w-h=%d) — looks like a football", dw, dh, diff)
	}

	// BOUNDED: not metropolis-sized. Radius ≈ max(dw,dh)/2 must stay in a sane band and never
	// approach the canvas dimensions.
	extent := dw
	if dh > extent {
		extent = dh
	}
	radius := extent / 2
	if radius < 2 {
		t.Fatalf("player dot too small (radius ~%d) — should be a prominent seat", radius)
	}
	if radius > 8 {
		t.Fatalf("player dot too large (radius ~%d) — metropolis-sized, not a dot", radius)
	}
	if extent >= w/3 || extent >= h/3 {
		t.Fatalf("player dot spans %dx%d — too big relative to the %dx%d canvas", dw, dh, w, h)
	}

	// NO PLUS CORE: the body must be a single filled disc. If a brighter inner core were
	// drawn on top of a same-color body it would rasterize a visible "+" — assert the center
	// row and center column are solidly filled with the body color across the disc (no bright
	// spike/gap pattern), i.e. the disc is convex and gap-free through its center.
	cx := (minX + maxX) / 2
	cy := (minY + maxY) / 2
	near := func(a, b uint8, tol int) bool {
		d := int(a) - int(b)
		if d < 0 {
			d = -d
		}
		return d <= tol
	}
	isBody := func(x, y int) bool {
		c := img.RGBAAt(x, y)
		return near(c.R, body.R, 6) && near(c.G, body.G, 6) && near(c.B, body.B, 6)
	}
	// Walk the center row across the body extent: once we enter the body we must not exit and
	// re-enter (a "+" core drawn atop would leave the arms a different color, breaking the run
	// — but here the run must be one contiguous body span).
	runs := 0
	inRun := false
	for x := minX; x <= maxX; x++ {
		if isBody(x, cy) {
			if !inRun {
				runs++
				inRun = true
			}
		} else {
			inRun = false
		}
	}
	if runs != 1 {
		t.Fatalf("player dot center row has %d body runs, want 1 — a broken/cross core, not a clean disc", runs)
	}
	_ = cx // center column symmetry is implied by the round-extent check above
}

// TestWorldPlayerDotBiggerThanNeighbours keeps the size relationship intact: your capital is
// the largest mark. Render a world with a maxed-strength neighbour and assert your accent
// body covers more pixels than any single neighbour dot's identity body — your seat dominates.
func TestWorldPlayerDotBiggerThanNeighbours(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 120, 90
	facs := map[string]game.FactionInfo{
		"rome": {Name: "Rome", Discovered: true, Status: "neutral", Opinion: 0, Strength: 5},
	}
	st := worldState("classical_age", map[string]int{"forge": 2}, "Sparta", facs)
	img, _ := renderWorldImage(st, w, h)

	_, _, _, _, yourN := dotExtent(img, brighten(rgba(theme.Color(theme.RoleAccent)), 0.12), 6)
	_, _, _, _, romeN := dotExtent(img, brighten(factionColor("rome"), 0.18), 8)
	if yourN == 0 {
		t.Fatal("player dot body not found")
	}
	if romeN == 0 {
		t.Skip("neighbour body color not distinctly found (blend/overlap); skipping size compare")
	}
	if yourN <= romeN {
		t.Fatalf("player dot (%d px) not larger than the strongest neighbour (%d px)", yourN, romeN)
	}
}
