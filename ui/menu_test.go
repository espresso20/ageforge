package ui

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/all"
)

// menuSizes are the terminals the menu is checked at: the least the game
// supports, the common ones, the largest the brief names and one beyond.
var menuSizes = [][2]int{{80, 24}, {100, 30}, {110, 36}, {114, 38}, {120, 40}, {132, 43}, {144, 46}, {200, 60}}

// stagedMenuView is a menu's view as refresh would build it for a game in
// age, or for a first visit (age "").
func stagedMenuView(age string, forge bool, tier mapmodel.GlyphTier) (*menuView, *menuTown) {
	m := &mainMenu{version: "v4.0.0", reg: all.Registry()}
	m.set = defaultMapSettings(m.reg)
	m.set.Tier = tier
	var town *menuTown
	if age != "" {
		st := fixture.State(fixture.Options{Age: age, Seed: 7})
		st.Workers.TotalPop = 216
		m.cur, m.hasCur = game.CurrentGame{Save: game.SaveInfo{Name: "ashford"}, Why: game.CurrentLast}, true
		town = newMenuTown(&st, m.reg, m.set.Style)
		m.town = town
		m.view.captions = menuCaptions(town, &st)
		m.continueDetails(&st)
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
