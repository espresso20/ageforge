package ui

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/all"
)

// menuSizes are the terminals the menu is checked at: the least the game
// supports, the common ones up to 144x46, one well beyond, and two odd
// shapes (tall and narrow, wide and short) that cross the layouts'
// thresholds one way and not the other.
var menuSizes = [][2]int{{80, 24}, {100, 30}, {110, 36}, {114, 38}, {120, 40}, {132, 43}, {144, 46}, {200, 60}, {90, 44}, {180, 26}}

// stagedTown is a fixture game and the picture of its town.
type stagedTown struct {
	st   game.GameState
	town *menuTown
}

var stagedTowns = map[string]stagedTown{}

// stagedEngine is the engine a staged menu asks for its ruleset. The pages
// are staged from fixtures, so nothing else of it is read.
var stagedEngine = game.NewGameEngine()

// stagedMenuView is a menu's view as refresh would build it for a game in
// age, or for a first visit (age "").
func stagedMenuView(age string, forge bool, tier mapmodel.GlyphTier) (*menuView, *menuTown) {
	m := &mainMenu{version: "v4.0.0", reg: all.Registry(), engine: stagedEngine}
	m.set = defaultMapSettings(m.reg)
	m.set.Tier = tier
	var town *menuTown
	if age != "" {
		// One town per age for the whole run: laying a town out is the
		// costly part, and a town draws at any size and in any glyph set.
		staged, ok := stagedTowns[age]
		if !ok {
			st := fixture.State(fixture.Options{Age: age, Seed: 7})
			st.Workers.TotalPop = 216
			staged = stagedTown{st: st, town: newMenuTown(&st, m.reg, m.set.Style)}
			stagedTowns[age] = staged
		}
		town = staged.town
		m.cur, m.hasCur = game.CurrentGame{Save: game.SaveInfo{Name: "ashford"}, Why: game.CurrentLast}, true
		m.town = town
		m.view.captions = menuCaptions(town, &staged.st)
		m.continueDetails(&staged.st)
	}
	m.buildViewWith([]string{"38 of 120 · Survivor", "38 of 120", "38"})
	if forge {
		m.view.forge = true
	}
	return &m.view, town
}

// menuPage draws a staged page at w by h after frames more frames.
func menuPage(v *menuView, town *menuTown, w, h, frames int) (*mGrid, *menuScene) {
	sc := newMenuScene(menuLayoutFor(w, h, v))
	for i := 0; i < frames; i++ {
		sc.step()
	}
	return sc.render(newMenuPalette(theme.Active()), v, town, mapstyle.Frame{Anim: frames, Clock: frames, Tier: mapmodel.TierUnicode}), sc
}

func gridText(g *mGrid) string {
	var sb strings.Builder
	for y := 0; y < g.h; y++ {
		sb.WriteString(strings.TrimRight(g.row(y), " "))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// gridHTML is a page as a block of HTML: one span per run of cells of one
// colour, for a look at it in a browser.
func gridHTML(g *mGrid, pal *menuPalette, plain bool) string {
	hex := func(c tcell.Color) string {
		r, gr, b := c.RGB()
		return fmt.Sprintf("#%02x%02x%02x", r&0xff, gr&0xff, b&0xff)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "<pre style=\"background:%s;color:%s\">", hex(pal.bg), hex(pal.ink))
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; {
			c := g.c[y*g.w+x]
			bg := pal.bg
			if c.hasBg {
				bg = c.bg
			}
			var run strings.Builder
			for x < g.w {
				d := g.c[y*g.w+x]
				dbg := pal.bg
				if d.hasBg {
					dbg = d.bg
				}
				if d.fg != c.fg || dbg != bg || d.bold != c.bold {
					break
				}
				r := d.r
				if plain {
					r = mapmodel.Fold(r, mapmodel.TierASCII)
				}
				run.WriteString(html.EscapeString(string(r)))
				x++
			}
			weight := ""
			if c.bold {
				weight = ";font-weight:700"
			}
			fmt.Fprintf(&sb, "<span style=\"color:%s;background:%s%s\">%s</span>", hex(c.fg), hex(bg), weight, run.String())
		}
		sb.WriteByte('\n')
	}
	sb.WriteString("</pre>\n")
	return sb.String()
}

// TestMenuDump writes the staged pages when MENU_DUMP names a folder: each
// as text, and all of them in every theme as one HTML page (menu.html), for
// a look at the layout and the colours while working on them. MENU_DUMP_SIZES
// narrows the HTML to some sizes ("120x40,80x24").
func TestMenuDump(t *testing.T) {
	out := os.Getenv("MENU_DUMP")
	if out == "" {
		t.Skip("set MENU_DUMP to a folder to write the pages")
	}
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	cases := []struct {
		name, age string
		forge     bool
	}{{"forge", "", true}, {"town_bronze", "bronze_age", false}, {"town_quantum", "quantum_age", false}, {"forge_with_game", "medieval_age", true}}
	for _, size := range menuSizes {
		for _, c := range cases {
			v, town := stagedMenuView(c.age, c.forge, mapmodel.TierUnicode)
			g, _ := menuPage(v, town, size[0], size[1], 0)
			name := fmt.Sprintf("%s_%dx%d.txt", c.name, size[0], size[1])
			if err := os.WriteFile(filepath.Join(out, name), []byte(gridText(g)), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	sizes := os.Getenv("MENU_DUMP_SIZES")
	if sizes == "" {
		sizes = "120x40,80x24"
	}
	themes := strings.Split(os.Getenv("MENU_DUMP_THEMES"), ",")
	if themes[0] == "" {
		themes = nil
		for _, th := range theme.All() {
			themes = append(themes, th.Key)
		}
	}
	var page strings.Builder
	page.WriteString("<!doctype html><meta charset=\"utf-8\"><title>Menu pages</title><style>body{background:#222;color:#ddd;font:13px sans-serif}pre{font:10px/12px 'JetBrains Mono',Menlo,monospace;display:inline-block;margin:4px 12px 16px 0;vertical-align:top}h3{margin:10px 0 2px}</style>\n")
	for _, key := range themes {
		if err := theme.SetActive(key); err != nil {
			t.Fatal(err)
		}
		pal := newMenuPalette(theme.Active())
		for _, sz := range strings.Split(sizes, ",") {
			var w, h int
			if _, err := fmt.Sscanf(sz, "%dx%d", &w, &h); err != nil {
				t.Fatal(err)
			}
			for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
				fmt.Fprintf(&page, "<h3 id=\"%s-%s-%s\">%s, %s, %s glyphs</h3>\n", key, sz, tier, key, sz, tier)
				for _, c := range cases[:2] {
					v, town := stagedMenuView(c.age, c.forge, tier)
					g, _ := menuPage(v, town, w, h, 0)
					page.WriteString(gridHTML(g, pal, tier == mapmodel.TierASCII))
				}
				// And struck: the flare at its height.
				v, town := stagedMenuView("bronze_age", false, tier)
				sc := newMenuScene(menuLayoutFor(w, h, v))
				sc.hit()
				sc.step()
				page.WriteString(gridHTML(sc.render(pal, v, town, mapstyle.Frame{Anim: 1, Clock: 1, Tier: tier}), pal, tier == mapmodel.TierASCII))
			}
		}
	}
	if err := os.WriteFile(filepath.Join(out, "menu.html"), []byte(page.String()), 0644); err != nil {
		t.Fatal(err)
	}
}

// menuStage is one staged page of the menu.
type menuStage struct {
	name, age string
	forge     bool
}

// menuStages is the pages the render tests draw: a first visit, a town in
// an early, a middle and a late age (the late ones draw the sky), and the
// forge page with a game on it (what the menu shows when the town cannot
// be drawn).
func menuStages() []menuStage {
	out := []menuStage{{"a first visit", "", true}}
	for _, age := range []string{"primitive_age", "bronze_age", "medieval_age", "modern_age", "space_age", "quantum_age", "transcendent_age"} {
		out = append(out, menuStage{"a town in the " + ageDisplay(age), age, false})
	}
	return append(out, menuStage{"the forge page with a game", "medieval_age", true})
}

// rowHas reports whether row y of g holds s from column x.
func rowHas(g *mGrid, x, y int, s string) bool {
	for i, r := range []rune(s) {
		if !g.in(x+i, y) || g.c[y*g.w+x+i].r != r {
			return false
		}
	}
	return true
}

// pageHas reports whether any row of g holds s.
func pageHas(g *mGrid, s string) bool {
	for y := 0; y < g.h; y++ {
		if strings.Contains(g.row(y), s) {
			return true
		}
	}
	return false
}

// checkMenuPage holds a drawn page to its layout: nothing cut off, every
// entry whole with its key, the lines round the menu in place, the frames
// unbroken and the wordmark complete.
func checkMenuPage(t *testing.T, where string, sc *menuScene, v *menuView, town *menuTown) {
	t.Helper()
	pal := newMenuPalette(theme.Active())
	frame := mapstyle.Frame{Anim: 3, Clock: 3, Tier: mapmodel.TierUnicode}
	// The page is drawn with the sparks out of the way: a spark may pass
	// over a frame's line, and it is the layout that is on trial here. (A
	// spark is one glyph and cannot be cut off; TestMenuPagesFitWhileMoving
	// holds the words under a shower of them.)
	held := sc.sparks.a
	sc.sparks.a = nil
	g := sc.render(pal, v, town, frame)
	sc.sparks.a = held
	if len(g.clipped) > 0 {
		t.Errorf("%s: text cut off at the edge: %q", where, g.clipped)
	}
	L := sc.L

	// The entries: each on its row, marked when selected, its key at the
	// end of the row, and the Continue entry's line under it.
	row := L.my
	for i, it := range v.items {
		left := "  " + it.label
		if i == v.sel {
			left = "▸ " + it.label
		}
		if !rowHas(g, L.mx, row, left) {
			t.Errorf("%s: row %d should hold %q, holds %q", where, row, left, strings.TrimSpace(g.row(row)))
		}
		keyAt := L.mx + L.mw - 1
		if !L.forge {
			keyAt = L.mx + L.mw - 2
		}
		if !rowHas(g, keyAt, row, string(it.key)) {
			t.Errorf("%s: the key of %q is not at column %d of row %d: %q", where, it.label, keyAt, row, strings.TrimSpace(g.row(row)))
		}
		if it.id == miBadges && !strings.Contains(g.row(row), it.extras[len(it.extras)-1]) {
			t.Errorf("%s: the Badges row does not say how many are held: %q", where, strings.TrimSpace(g.row(row)))
		}
		row++
		if len(it.details) > 0 {
			found := false
			for _, d := range it.details {
				found = found || rowHas(g, L.mx+4, row, d)
			}
			if !found {
				t.Errorf("%s: the line under %q holds none of %q: %q", where, it.label, it.details, strings.TrimSpace(g.row(row)))
			}
			row++
		}
		row += L.gap - 1
	}

	if !pageHas(g, v.versions[len(v.versions)-1]) {
		t.Errorf("%s: the version is not on the page", where)
	}
	if !pageHas(g, "choose") {
		t.Errorf("%s: the line that says how the menu is driven is not on the page", where)
	}
	chromes := []string{"PgUp", "arrows move", "cursor hidden", "No map yet", "SETTLEMENT"}
	if town != nil && town.name != "" {
		chromes = append(chromes, "◆ "+town.name)
	}
	for _, chrome := range chromes {
		if pageHas(g, chrome) {
			t.Errorf("%s: the map's own bars are on the page (%q)", where, chrome)
		}
	}

	corners := func(what string, x, y, w, h int, c string) {
		rs := []rune(c)
		for i, at := range [][2]int{{x, y}, {x + w - 1, y}, {x, y + h - 1}, {x + w - 1, y + h - 1}} {
			if !g.in(at[0], at[1]) || g.c[at[1]*g.w+at[0]].r != rs[i] {
				t.Errorf("%s: %s is broken at its corner (%d,%d)", where, what, at[0], at[1])
			}
		}
		for j := y + 1; j < y+h-1; j++ {
			if g.c[j*g.w+x].r != rs[4] || g.c[j*g.w+x+w-1].r != rs[4] {
				t.Errorf("%s: %s is broken on row %d: %q", where, what, j, g.row(j))
			}
		}
	}
	lit := 0
	for py := 0; py < menuWordH; py++ {
		lit += strings.Count(menuWord[py], "#") * L.sx
	}
	block := "█"
	if v.plain {
		block = "#"
	}
	drawn := 0
	for y := L.wy; y < L.wy+menuWordH; y++ {
		drawn += strings.Count(string([]rune(g.row(y))[max(L.wx, 0):min(L.wx+menuWordW*L.sx, g.w)]), block)
	}
	if drawn != lit {
		t.Errorf("%s: the wordmark has %d of its %d cells", where, drawn, lit)
	}

	if L.forge {
		corners("the page's frame", 1, 0, L.w-2, L.h, "┏┓┗┛┃")
		corners("the plate", L.px, L.py, L.pw, L.ph, "┏┓┗┛┃")
		if L.mx+L.mw > L.px {
			t.Errorf("%s: the menu (to column %d) runs into the plate (from %d)", where, L.mx+L.mw, L.px)
		}
		for what, s := range map[string]string{"the plate's caption": "The forge", "the contents heading": " CONTENTS ", "the line under the wordmark": "begun at a campfire", "the line over it": "A G E S"} {
			if !pageHas(g, s) {
				t.Errorf("%s: %s is not on the page", where, what)
			}
		}
		if L.wx < 3 || L.wx+menuWordW*L.sx > L.w-3 {
			t.Errorf("%s: the wordmark (columns %d to %d) does not sit inside the frame", where, L.wx, L.wx+menuWordW*L.sx)
		}
		return
	}

	corners("the menu's box", L.px, L.py, L.pw, L.ph, "┌┐└┘│")
	if L.px < 0 || L.px+L.pw > L.w || L.py <= L.tag || L.py+L.ph > L.cap {
		t.Errorf("%s: the box (%d,%d %dx%d) does not sit between the sky and the caption (row %d)", where, L.px, L.py, L.pw, L.ph, L.cap)
	}
	if !strings.Contains(g.row(L.tag), "ages, one terminal") {
		t.Errorf("%s: the line under the wordmark is not on row %d: %q", where, L.tag, strings.TrimSpace(g.row(L.tag)))
	}
	found := false
	for _, c := range v.captions {
		found = found || rowHas(g, 2, L.cap, " "+c+" ")
	}
	if !found {
		t.Errorf("%s: the caption row holds none of %q: %q", where, v.captions, strings.TrimSpace(g.row(L.cap)))
	}
	// The town is there: the land under the sky is drawn, to the left of
	// the box, by the map's full view (its compact view would be a few
	// rows in a corner).
	land := 0
	for y := L.top + 3; y < L.h; y++ {
		for x := 0; x < min(L.px, L.landW); x++ {
			if c := g.c[y*g.w+x]; c.r != ' ' || c.hasBg {
				land++
			}
		}
	}
	if area := (L.h - L.top - 3) * min(L.px, L.landW); land*20 < area {
		t.Errorf("%s: the land is mostly empty (%d of %d cells drawn): the town is not behind the menu", where, land, area)
	}
}

// TestMenuPagesFitEverySize draws both pages from 80x24 to 144x46 (and at
// 200x60), in an early, a middle and a late age, in both glyph sets, with
// and without the lines that come and go (an update notice, the forge
// master's line), and fails on anything cut off, broken or out of place.
func TestMenuPagesFitEverySize(t *testing.T) {
	for _, size := range menuSizes {
		for _, st := range menuStages() {
			for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
				for _, more := range []bool{false, true} {
					if more && tier == mapmodel.TierASCII {
						continue // the lines that come and go are laid out the same in both glyph sets
					}
					v, town := stagedMenuView(st.age, st.forge, tier)
					where := fmt.Sprintf("%s at %dx%d, %s glyphs", st.name, size[0], size[1], tier)
					if more {
						v.versions = []string{"v4.0.0 · update available (u)", "v4.0.0 · update (u)", "v4.0.0"}
						v.elite = eliteLines[2]
						v.sel = len(v.items) - 1
						where += ", with an update and the forge master's line"
					}
					sc := newMenuScene(menuLayoutFor(size[0], size[1], v))
					checkMenuPage(t, where, sc, v, town)
					if more && !st.forge {
						g := sc.render(newMenuPalette(theme.Active()), v, town, mapstyle.Frame{Tier: tier})
						if !strings.Contains(g.row(sc.L.py), eliteLines[2][1]) {
							t.Errorf("%s: the forge master's line is not on the box", where)
						}
						if !strings.Contains(g.row(sc.L.py+sc.L.ph-1), "update") {
							t.Errorf("%s: the update notice is not on the box", where)
						}
					}
				}
			}
		}
	}
}

// TestMenuPagesFitWhileMoving: twenty-four frames of each page, with a
// strike on the way, and at every frame every entry is whole under the
// sparks (a spark never lands on the menu's words) and nothing is cut off.
func TestMenuPagesFitWhileMoving(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}, {144, 46}} {
		for _, st := range []menuStage{{"a first visit", "", true}, {"a town in the Bronze Age", "bronze_age", false}} {
			v, town := stagedMenuView(st.age, st.forge, mapmodel.TierUnicode)
			sc := newMenuScene(menuLayoutFor(size[0], size[1], v))
			pal := newMenuPalette(theme.Active())
			for n := 0; n < 24; n++ {
				if n == 3 {
					sc.hit()
				}
				sc.step()
				where := fmt.Sprintf("%s at %dx%d, frame %d", st.name, size[0], size[1], n)
				g := sc.render(pal, v, town, mapstyle.Frame{Anim: n, Clock: n, Tier: mapmodel.TierUnicode})
				if len(g.clipped) > 0 {
					t.Fatalf("%s: text cut off at the edge: %q", where, g.clipped)
				}
				if n == 4 && len(sc.sparks.a) < 40 {
					t.Fatalf("%s: the strike threw %d sparks", where, len(sc.sparks.a))
				}
				row := sc.L.my
				for i, it := range v.items {
					left := "  " + it.label
					if i == v.sel {
						left = "▸ " + it.label
					}
					if !rowHas(g, sc.L.mx, row, left) {
						t.Fatalf("%s: row %d should hold %q under the sparks, holds %q", where, row, left, strings.TrimSpace(g.row(row)))
					}
					row += sc.L.gap
					if len(it.details) > 0 {
						row++
					}
				}
			}
		}
	}
}

// TestMenuOnATerminalTooSmall: under 80x24 the page is the list alone, and
// no size at all makes it fail.
func TestMenuOnATerminalTooSmall(t *testing.T) {
	for _, w := range []int{0, 1, 7, 20, 44, 60, 79, 80} {
		for _, h := range []int{0, 1, 5, 12, 23, 24} {
			for _, st := range []menuStage{{"", "", true}, {"", "bronze_age", false}} {
				v, town := stagedMenuView(st.age, st.forge, mapmodel.TierUnicode)
				g, sc := menuPage(v, town, w, h, 2)
				sc.hit()
				sc.step()
				if (w < menuMinW || h < menuMinH) != sc.L.compact {
					t.Errorf("%dx%d: compact is %v", w, h, sc.L.compact)
				}
				if sc.L.compact && w >= 44 && h >= 12 {
					for _, it := range v.items {
						if !pageHas(g, it.label) {
							t.Errorf("%dx%d: the list has no %q", w, h, it.label)
						}
					}
				}
			}
		}
	}
}

// TestMenuRampsInEveryTheme: in all the themes the wordmark's heat and the
// forge's fire come out of the theme's own roles and stay legible. Hotter
// always stands out more from the page than cooler; the coldest iron is
// still a letter; white heat is as strong as body text; and on a light
// theme the ramp runs the other way, hot iron being the darkest ink on the
// paper. The anvil shows against the flames, and the menu's own inks clear
// the text rule on the grounds they are drawn on.
func TestMenuRampsInEveryTheme(t *testing.T) {
	if n := len(theme.All()); n != 16 {
		t.Logf("the game has %d themes (the menu was drawn for 16): every one is checked", n)
	}
	for _, th := range theme.All() {
		p := newMenuPalette(th)
		con := func(c tcell.Color) float64 { return theme.ContrastRatio(c, p.bg) }
		for name, ramp := range map[string][6]tcell.Color{"heat": p.heatStops, "fire": p.fireStops} {
			for i := 1; i < len(ramp); i++ {
				if con(ramp[i]) <= con(ramp[i-1]) {
					t.Errorf("%s: the %s ramp's stop %d (%06x, %.2f) does not stand out more than stop %d (%06x, %.2f)",
						th.Key, name, i, ramp[i].Hex(), con(ramp[i]), i-1, ramp[i-1].Hex(), con(ramp[i-1]))
				}
			}
		}
		if p.fireStops[0] != p.bg {
			t.Errorf("%s: a fire gone out is not the page's own colour", th.Key)
		}
		if c := con(p.heatStops[0]); c < menuColdIron-0.02 {
			t.Errorf("%s: cold iron stands out %.2f from the page, under %.1f", th.Key, c, menuColdIron)
		}
		if top, text := con(p.heatStops[5]), con(th.Color(theme.RoleText)); top < 10 && top < text-0.05 {
			t.Errorf("%s: white heat stands out %.2f, less than body text (%.2f)", th.Key, top, text)
		}
		coldLum, hotLum := theme.RelativeLuminance(p.heatStops[0]), theme.RelativeLuminance(p.heatStops[5])
		if th.IsLight() && hotLum >= coldLum {
			t.Errorf("%s: on a light theme hot iron should be the darker ink (cold %.3f, hot %.3f)", th.Key, coldLum, hotLum)
		}
		if !th.IsLight() && hotLum <= coldLum {
			t.Errorf("%s: on a dark theme hot iron should be the brighter (cold %.3f, hot %.3f)", th.Key, coldLum, hotLum)
		}

		// The wordmark at rest, as a still menu holds it: it reads as a
		// word, and no pixel of it is lost in the page.
		sc := newMenuScene(menuLayoutFor(120, 40, &menuView{forge: true, items: make([]menuItem, 7)}))
		sum, n, lo := 0.0, 0, 99.0
		for py := 0; py < menuWordH; py++ {
			for px := 0; px < menuWordW; px++ {
				if menuWord[py][px] == '#' {
					c := con(rampAt(&p.heat, wordHeat(px, py, sc.t, 0, 0.5)))
					sum, n, lo = sum+c, n+1, min(lo, c)
				}
			}
		}
		if mean := sum / float64(n); mean < 3 || lo < menuColdIron-0.02 {
			t.Errorf("%s: the wordmark at rest stands out %.2f on average and %.2f at its coldest", th.Key, mean, lo)
		}

		if theme.ContrastRatio(p.anvil, p.fireStops[4]) < 3 {
			t.Errorf("%s: the anvil (%06x) is lost in the flames (%06x)", th.Key, p.anvil.Hex(), p.fireStops[4].Hex())
		}
		if theme.ContrastRatio(p.anvilInk, p.anvil) < 1.5 {
			t.Errorf("%s: the anvil's glyphs do not show on its body", th.Key)
		}
		for what, pair := range map[string][3]any{
			"dim text on the page":        {p.dim, p.bg, 3.0},
			"the accent on the page":      {p.accent, p.bg, 3.0},
			"the faint lines on the page": {p.faint, p.bg, 1.45},
			"text in the box":             {p.groundInk, p.ground, 7.0},
			"dim text in the box":         {p.groundDim, p.ground, 3.0},
			"the selected row's text":     {p.onBand, p.band, 4.5},
			"the selected row in the box": {p.band, p.ground, 3.0},
			"the elite line's second ink": {p.label, p.bg, 3.0},
		} {
			if c := theme.ContrastRatio(pair[0].(tcell.Color), pair[1].(tcell.Color)); c < pair[2].(float64)-0.02 {
				t.Errorf("%s: %s has a contrast of %.2f, under %.2f", th.Key, what, c, pair[2].(float64))
			}
		}
		if p.ground == p.bg {
			t.Errorf("%s: the box stands on the page's own colour", th.Key)
		}
	}
}

// TestMenuPaintsInEveryTheme draws both pages on a screen in every theme
// and both glyph sets and reads the colours back: every word of the menu
// clears the text rule against what it is drawn on, the page's ground is
// the theme's, and the plain glyph set puts nothing on the screen that is
// not ASCII.
func TestMenuPaintsInEveryTheme(t *testing.T) {
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		pal := newMenuPalette(th)
		for _, size := range [][2]int{{80, 24}, {120, 40}} {
			for _, st := range []menuStage{{"the first visit", "", true}, {"the town", "bronze_age", false}} {
				for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
					w, h := size[0], size[1]
					where := fmt.Sprintf("%s, %s at %dx%d, %s glyphs", th.Key, st.name, w, h, tier)
					v, town := stagedMenuView(st.age, st.forge, tier)
					v.elite = eliteLines[0]
					sc := newMenuScene(menuLayoutFor(w, h, v))
					sc.sparks.a = nil
					g := sc.render(pal, v, town, mapstyle.Frame{Anim: 3, Clock: 3, Tier: tier})

					sim := tcell.NewSimulationScreen("UTF-8")
					if err := sim.Init(); err != nil {
						t.Fatal(err)
					}
					sim.SetSize(w, h)
					g.flush(theme.WrapScreen(sim), 0, 0, pal, v.plain)
					rows := make([][]rune, h)
					for y := range rows {
						rows[y] = make([]rune, w)
						for x := range rows[y] {
							r, _, _, _ := sim.GetContent(x, y)
							rows[y][x] = r
							if v.plain && r > 0x7e {
								t.Fatalf("%s: cell (%d,%d) holds %q, which the plain glyph set does not have", where, x, y, r)
							}
						}
					}
					_, _, corner, _ := sim.GetContent(0, 0)
					if _, bg, _ := corner.Decompose(); bg != pal.bg || bg != th.Color(theme.RoleBackground) {
						t.Errorf("%s: the page's corner is not on the theme's ground", where)
					}

					words := []string{"New game", "Load game", "Badges", "Themes", "Accounts", "Check for updates", "Quit", "v4.0.0", "choose", "Enter open", "38 of 120", eliteLines[0][1]}
					if st.forge {
						words = append(words, "CONTENTS", "The forge", "begun at a campfire", "fourth edition")
					} else {
						words = append(words, "Continue", "ashford", "Bronze Age", "216 people", "ages, one terminal")
					}
					for _, word := range words {
						found := false
						for y := 0; y < h && !found; y++ {
							at := strings.Index(string(rows[y]), word)
							if at < 0 {
								continue
							}
							found = true
							x0 := len([]rune(string(rows[y])[:at]))
							for i, r := range []rune(word) {
								if r == ' ' {
									continue
								}
								_, _, cst, _ := sim.GetContent(x0+i, y)
								fg, bg, _ := cst.Decompose()
								if c := theme.ContrastRatio(fg, bg); c < 3-0.02 {
									t.Errorf("%s: %q is drawn at a contrast of %.2f (%06x on %06x)", where, word, c, fg.Hex(), bg.Hex())
									break
								}
							}
						}
						if !found {
							t.Errorf("%s: %q is not on the screen", where, word)
						}
					}
					sim.Fini()
				}
			}
		}
	}
}

// TestMenuGlyphsAreOneColumnWide: every glyph the menu itself puts on the
// page takes one column on a terminal (a wide one would push the rest of
// its row along), and in the plain glyph set the forge master's marks are
// glyphs it has, never a question mark.
func TestMenuGlyphsAreOneColumnWide(t *testing.T) {
	check := func(where string, g *mGrid) {
		for i, c := range g.c {
			if c.r > 0x7e && uniseg.StringWidth(string(c.r)) != 1 {
				t.Fatalf("%s: cell (%d,%d) holds %q, which is not one column wide", where, i%g.w, i/g.w, c.r)
			}
		}
	}
	pal := newMenuPalette(theme.Active())
	for i, line := range eliteLines {
		for _, st := range []menuStage{{"the first visit", "", true}, {"the town", "bronze_age", false}} {
			for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
				v, town := stagedMenuView(st.age, st.forge, tier)
				v.elite = line
				sc := newMenuScene(menuLayoutFor(120, 40, v))
				sc.hit()
				sc.step()
				g := sc.render(pal, v, town, mapstyle.Frame{Tier: tier})
				where := fmt.Sprintf("%s with the forge master's line %d, %s glyphs", st.name, i, tier)
				check(where, g)
				row := sc.L.py
				if st.forge {
					row = sc.L.rule
				}
				text := g.row(row)
				if !strings.Contains(text, line[1]) {
					t.Errorf("%s: the line is not on row %d: %q", where, row, strings.TrimSpace(text))
				}
				if tier == mapmodel.TierASCII {
					for _, r := range text {
						if mapmodel.Fold(r, mapmodel.TierASCII) == '?' {
							t.Errorf("%s: row %d holds %q, which the plain glyph set cannot show", where, row, r)
						}
					}
				}
			}
		}
	}
}
