package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
)

// caseTiers are the glyph tiers the case is checked in.
var caseTiers = []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII, mapmodel.TierNerd}

// blockAt is the w by h block of a grid at (x, y) as rows of runes.
func blockAt(g *tGrid, x, y, w, h int) []string {
	out := make([]string, h)
	for dy := 0; dy < h; dy++ {
		var sb strings.Builder
		for dx := 0; dx < w; dx++ {
			sb.WriteRune(g.at(x+dx, y+dy).r)
		}
		out[dy] = sb.String()
	}
	return out
}

// wantMini is an item's mini drawn on its own, folded for the tier.
func wantMini(it *caseItem, v caseView) []string {
	g := newTGrid(miniW, miniH)
	drawItemMini(g, 0, 0, it, v, atRest)
	if v.plain() {
		foldBadges(g)
	}
	return blockAt(g, 0, 0, miniW, miniH)
}

// wantMedal is an item's full art drawn on its own, folded for the tier.
func wantMedal(it *caseItem, v caseView) (rows []string, w, h int) {
	md := medalOf(it.v, v.tier)
	w, h = medalSize(md)
	g := newTGrid(w, h)
	drawMedal(g, 0, 0, md, atRest, v.plain())
	if v.plain() {
		foldBadges(g)
	}
	return blockAt(g, 0, 0, w, h), w, h
}

// TestBadgeCaseAtEverySize draws the case at the four sizes the game is
// checked at, in every glyph tier, on every tab and with every badge
// selected in turn, and holds the picture to what the panel promises:
//
//   - the selected badge's mini is on screen, whole, inside the grid;
//   - no badge is ever cut: every mini of a line that shows is drawn in
//     full, and a line that does not fit draws nothing;
//   - the selected badge's full art is beside the grid, whole, where the
//     layout puts it, with its name under it;
//   - nothing is drawn off the cell grid: every rune is one cell wide;
//   - the plain tier draws nothing but ASCII;
//   - the title, the tabs, the status line and the key bar are there, and
//     no row runs past the panel.
func TestBadgeCaseAtEverySize(t *testing.T) {
	views, sum := caseFixture()
	for _, size := range treeSizes {
		w, h := size[0], size[1]-promptRows
		for _, tier := range caseTiers {
			tabs := caseTabs(views, false)
			for _, tab := range tabs {
				m0 := buildCase(views, sum, true, tab.key, caseLayoutFor(w, h, tab.key))
				if len(m0.items) == 0 {
					t.Fatalf("%dx%d: the %s tab lists nothing", w, h, tab.title)
				}
				for _, sel := range m0.items {
					where := fmt.Sprintf("%dx%d %s tab %q, %s selected", w, h, tier, tab.title, sel.id)
					m, v, g := drawCase(views, sum, w, h, caseView{tab: tab.key, sel: sel.id, tier: tier}, atRest)
					if g.w != w || g.h != h {
						t.Fatalf("%s: rendered %dx%d", where, g.w, g.h)
					}
					lay := m.lay
					it := m.selected(v)
					if it == nil || it.id != sel.id {
						t.Fatalf("%s: the selection was lost", where)
					}
					for i, c := range g.c {
						if c.r != ' ' && uniseg.StringWidth(string(c.r)) != 1 {
							t.Fatalf("%s: %q at %d,%d is not one cell wide", where, c.r, i%w, i/w)
						}
						if tier == mapmodel.TierASCII && c.r >= 0x80 {
							t.Fatalf("%s: the plain tier draws %q at %d,%d", where, c.r, i%w, i/w)
						}
					}
					lines := strings.Split(g.String(), "\n")
					if !strings.Contains(lines[0], "BADGES") || !strings.Contains(lines[0], sum.Title) {
						t.Errorf("%s: the title bar has no title: %q", where, lines[0])
					}
					if !strings.Contains(lines[1], "["+tab.title+"]") {
						t.Errorf("%s: the open tab is not marked: %q", where, lines[1])
					}
					if !strings.Contains(lines[h-1], "Esc") {
						t.Errorf("%s: the key bar is missing: %q", where, lines[h-1])
					}
					if strings.TrimSpace(lines[h-2]) == "" {
						t.Errorf("%s: the status line is empty", where)
					}

					// Every line that shows is whole; the rest draw nothing.
					selShown := false
					for li, ln := range m.lines {
						if ln.kind != lineBadges {
							continue
						}
						y := lay.gy + ln.y - v.vy
						shown := m.shows(li, v.vy)
						if shown && (y < lay.gy || y+miniH > lay.gy+lay.gh) {
							t.Errorf("%s: a line of badges is drawn at rows %d to %d, outside the grid (%d to %d)", where, y, y+miniH-1, lay.gy, lay.gy+lay.gh-1)
						}
						for _, i := range ln.items {
							o := &m.items[i]
							x := lay.gx + o.col*casePitchX
							if !shown {
								continue
							}
							if x < lay.gx || x+miniW > lay.gx+lay.gw {
								t.Errorf("%s: %s is drawn at columns %d to %d, outside the grid (%d to %d)", where, o.id, x, x+miniW-1, lay.gx, lay.gx+lay.gw-1)
							}
							if got, want := blockAt(g, x, y, miniW, miniH), wantMini(o, v); strings.Join(got, "\n") != strings.Join(want, "\n") {
								t.Errorf("%s: %s is not drawn whole:\n%s\nwant\n%s", where, o.id, strings.Join(got, "\n"), strings.Join(want, "\n"))
							}
							if o.id == it.id {
								selShown = true
								mark := []rune{'▸', '◂'}
								if tier == mapmodel.TierASCII {
									mark = []rune{'>', '<'}
								}
								if g.at(x-1, y+1).r != mark[0] || g.at(x+miniW, y+1).r != mark[1] {
									t.Errorf("%s: the selection is not marked", where)
								}
							}
						}
					}
					if lay.list {
						for _, ln := range m.lines {
							if len(ln.items) > 0 && m.items[ln.items[0]].id == it.id && m.visible(ln, v.vy) {
								selShown = true
							}
						}
					}
					if !selShown {
						t.Errorf("%s: the selected badge is not on screen", where)
					}

					// The selected badge at full size, beside the grid.
					rows, aw, ah := wantMedal(it, v)
					ax, ay := lay.px+(lay.pw-aw)/2, lay.py
					if ax < lay.px || ax+aw > w || ay+ah > lay.py+lay.ph {
						t.Fatalf("%s: the full art (%dx%d) does not fit its pane (%dx%d)", where, aw, ah, lay.pw, lay.ph)
					}
					if got := blockAt(g, ax, ay, aw, ah); strings.Join(got, "\n") != strings.Join(rows, "\n") {
						t.Errorf("%s: the full art is not whole:\n%s\nwant\n%s", where, strings.Join(got, "\n"), strings.Join(rows, "\n"))
					}
					name := it.v.Name
					if it.group {
						name = game.BadgeHiddenName
					}
					pane := ""
					for y := lay.py; y < lay.py+lay.ph; y++ {
						pane += strings.Join(blockAt(g, lay.px, y, lay.pw, 1), "") + "\n"
					}
					if !strings.Contains(pane, name) {
						t.Errorf("%s: the pane does not name the badge:\n%s", where, pane)
					}
					if it.v.Earned && !strings.Contains(pane, "Earned") {
						t.Errorf("%s: the pane does not say the badge is earned:\n%s", where, pane)
					}
				}
			}
		}
	}
}

// TestBadgeCaseDetail: Enter opens a badge's detail over the grid at every
// size: a frame inside the body, the full art whole inside it, the name
// in its title and the description in full, nothing cut.
func TestBadgeCaseDetail(t *testing.T) {
	views, sum := caseFixture()
	for _, size := range treeSizes {
		w, h := size[0], size[1]-promptRows
		for _, tier := range caseTiers {
			for _, bv := range views {
				if bv.Hidden {
					continue
				}
				where := fmt.Sprintf("%dx%d %s, the detail of %s", w, h, tier, bv.Key)
				m, v, g := drawCase(views, sum, w, h, caseView{sel: bv.Key, card: true, tier: tier}, atRest)
				if !v.card {
					t.Fatalf("%s: the detail did not open", where)
				}
				it := m.selected(v)
				text := g.String()
				lines := strings.Split(text, "\n")
				top, bottom := -1, -1
				tl, bl := "╭─ "+bv.Name, "╰──"
				if tier == mapmodel.TierASCII {
					tl, bl = ".- "+bv.Name, "'--"
				}
				for y, l := range lines {
					if strings.Contains(l, tl) {
						top = y
					}
					// The frame's foot is the last such row: a sprite may have
					// a rounded foot of its own inside it.
					if strings.Contains(l, bl) && top >= 0 && y > top {
						bottom = y
					}
				}
				if top < caseTop || bottom < 0 || bottom >= h-caseFoot {
					t.Fatalf("%s: the detail's frame is at rows %d to %d, outside the body (%d to %d):\n%s", where, top, bottom, caseTop, h-caseFoot-1, text)
				}
				rows, aw, ah := wantMedal(it, v)
				found := false
				for y := top + 1; y+ah <= bottom && !found; y++ {
					for x := 0; x+aw <= w && !found; x++ {
						found = strings.Join(blockAt(g, x, y, aw, ah), "\n") == strings.Join(rows, "\n")
					}
				}
				if !found {
					t.Errorf("%s: the full art is not whole inside the detail:\n%s", where, text)
				}
				flat := strings.Join(strings.Fields(strings.NewReplacer("│", " ", "|", " ").Replace(text)), " ")
				desc := bv.Desc
				if tier == mapmodel.TierASCII {
					desc = strings.Map(func(r rune) rune {
						if f, ok := badgeFold[r]; ok {
							return f
						}
						return r
					}, desc)
				}
				for _, word := range strings.Fields(desc) {
					if !strings.Contains(flat, word) {
						t.Errorf("%s: the description lost %q:\n%s", where, word, text)
						break
					}
				}
				if strings.Contains(text, "…") {
					t.Errorf("%s: the detail cuts something short:\n%s", where, text)
				}
				for i, c := range g.c {
					if c.r != ' ' && uniseg.StringWidth(string(c.r)) != 1 {
						t.Fatalf("%s: %q at %d,%d is not one cell wide", where, c.r, i%w, i/w)
					}
				}
			}
		}
	}
}

// TestBadgeCaseSmallWindows: under the sizes the game is laid out for, the
// case still draws without cutting a badge: the selected badge's lines go
// under the grid below 80 columns, the grid becomes a list of names below
// 60, and a window too small for either says so.
func TestBadgeCaseSmallWindows(t *testing.T) {
	views, sum := caseFixture()
	for _, size := range [][2]int{{79, 21}, {70, 20}, {60, 18}, {59, 18}, {40, 12}, {30, 10}, {29, 9}, {10, 3}, {0, 0}} {
		w, h := size[0], size[1]
		for _, tier := range caseTiers {
			for _, tab := range []string{caseTabAll, "lineage", caseTabNext} {
				m, v, g := drawCase(views, sum, w, h, caseView{tab: tab, sel: "lineage.housing.2", tier: tier}, atRest)
				text := g.String()
				if w < caseMinW || h < caseMinH {
					if w >= 10 && h >= 1 && !strings.Contains(text, "Badges") {
						t.Errorf("%dx%d: a window too small for the case does not say so: %q", w, h, text)
					}
					continue
				}
				if (w < caseListW) != m.lay.list && tab != caseTabNext {
					t.Errorf("%dx%d: list layout is %v", w, h, m.lay.list)
				}
				if m.lay.side {
					t.Errorf("%dx%d: the selected badge is beside the grid in a window under %d columns", w, h, caseSideW)
				}
				// The name whole from 40 columns; under that a list's row may
				// cut it, and the start of it must still be there.
				it := m.selected(v)
				if it == nil {
					t.Fatalf("%dx%d %s tab %q: nothing is selected", w, h, tier, tab)
				}
				name := it.v.Name
				if w < 40 {
					name = clipRunes(name, 11)
				}
				if !strings.Contains(text, name) {
					t.Errorf("%dx%d %s tab %q: the selected badge is not named:\n%s", w, h, tier, tab, text)
				}
				for li, ln := range m.lines {
					if ln.kind != lineBadges || !m.shows(li, v.vy) {
						continue
					}
					y := m.lay.gy + ln.y - v.vy
					for _, i := range ln.items {
						o := &m.items[i]
						x := m.lay.gx + o.col*casePitchX
						if x+miniW > m.lay.gx+m.lay.gw || y+miniH > m.lay.gy+m.lay.gh {
							t.Errorf("%dx%d: %s is drawn outside the grid", w, h, o.id)
						}
						if got, want := blockAt(g, x, y, miniW, miniH), wantMini(o, v); strings.Join(got, "\n") != strings.Join(want, "\n") {
							t.Errorf("%dx%d: %s is not drawn whole", w, h, o.id)
						}
					}
				}
			}
		}
	}
}

// TestBadgeTiersDifferInShape: the five tiers, the locked frame, the
// hidden slab and the integrity frame are told apart with no colour at
// all, at both sizes and in the plain tier too, and a crossed badge
// differs from the same badge earned clean.
func TestBadgeTiersDifferInShape(t *testing.T) {
	type look struct {
		name string
		md   medal
	}
	looks := []look{
		{"bronze", medal{tier: config.BadgeBronze, emblem: 'x'}},
		{"silver", medal{tier: config.BadgeSilver, emblem: 'x'}},
		{"gold", medal{tier: config.BadgeGold, emblem: 'x'}},
		{"platinum", medal{tier: config.BadgePlatinum, emblem: 'x'}},
		{"legendary", medal{tier: config.BadgeLegendary, emblem: 'x'}},
		{"integrity", medal{tier: config.BadgeNoTier, emblem: 'x'}},
		{"locked", medal{tier: config.BadgeGold, emblem: 'x', state: medalLocked}},
		{"hidden", medal{tier: config.BadgeGold, emblem: 'x', state: medalHidden}},
		{"crossed bronze", medal{tier: config.BadgeBronze, emblem: 'x', crossed: true}},
	}
	for _, plain := range []bool{false, true} {
		minis, fulls := map[string]string{}, map[string]string{}
		for _, l := range looks {
			g := newTGrid(miniW, miniH)
			drawMini(g, 0, 0, l.md, atRest)
			w, h := medalSize(l.md)
			f := newTGrid(w, h)
			drawMedal(f, 0, 0, l.md, atRest, plain)
			if plain {
				foldBadges(g)
				foldBadges(f)
			}
			mini, full := g.String(), f.String()
			if strings.TrimSpace(mini) == "" || strings.TrimSpace(full) == "" {
				t.Fatalf("plain=%v: %s draws nothing", plain, l.name)
			}
			for other, s := range minis {
				if s == mini {
					t.Errorf("plain=%v: the %s and %s minis are the same shape:\n%s", plain, l.name, other, mini)
				}
			}
			for other, s := range fulls {
				// Locked takes the shape of its tier on purpose: only its
				// emblem and its ornaments differ.
				if s == full {
					t.Errorf("plain=%v: %s and %s are the same shape at full size:\n%s", plain, l.name, other, full)
				}
			}
			minis[l.name], fulls[l.name] = mini, full
		}
	}
	// Sizes: the table of the design.
	for tier, want := range map[config.BadgeTier][2]int{
		config.BadgeBronze: {5, 3}, config.BadgeSilver: {9, 5}, config.BadgeGold: {13, 7},
		config.BadgePlatinum: {13, 7}, config.BadgeLegendary: {17, 9},
	} {
		if w, h := medalSize(medal{tier: tier}); w != want[0] || h != want[1] {
			t.Errorf("%s is %dx%d, want %dx%d", tier.Name(), w, h, want[0], want[1])
		}
		sp := tierSprite(tier)
		for y, row := range sp.rows {
			if len([]rune(row)) != want[0] || len(sp.roles[y]) != want[0] {
				t.Errorf("%s: row %d of the sprite or its roles is not %d wide", tier.Name(), y, want[0])
			}
		}
		ex, ey := sp.emblemCell()
		if ex != want[0]/2 || ey != want[1]/2 {
			t.Errorf("%s: the emblem is at %d,%d, not the middle", tier.Name(), ex, ey)
		}
	}
}

// TestBadgeFoldKeepsTheRim: the plain tier keeps what the map's own fold
// would lose. The three stops of a rim fold to three marks, a round
// corner to a point above and a tick below, a dashed edge and a solid one
// stay apart, and nothing badge art draws is left unfolded.
func TestBadgeFoldKeepsTheRim(t *testing.T) {
	for _, set := range [][]rune{{'▓', '▒', '░'}, {'─', '┄', '═'}, {'│', '┆'}, {'╭', '╰'}, {'▄', '▀'}} {
		seen := map[rune]rune{}
		for _, r := range set {
			f, ok := badgeFold[r]
			if !ok {
				t.Errorf("%q is not in the badge fold", r)
				continue
			}
			if other, dup := seen[f]; dup {
				t.Errorf("%q and %q both fold to %q", r, other, f)
			}
			seen[f] = r
		}
	}
	for r, f := range badgeFold {
		if f >= 0x80 || f < 0x20 {
			t.Errorf("%q folds to %q, which is not plain", r, f)
		}
	}
	// Everything the art draws, at every frame, folds to ASCII.
	for _, md := range allMedals() {
		for n := atRest; n < 80; n++ {
			w, h := medalSize(md)
			g := newTGrid(w, h)
			drawMedal(g, 0, 0, md, n, true)
			m := newTGrid(miniW, miniH)
			drawMini(m, 0, 0, md, n)
			for _, c := range append(g.c, m.c...) {
				if c.r < 0x80 {
					continue
				}
				if _, ok := badgeFold[c.r]; !ok {
					t.Fatalf("%+v frame %d draws %q, which the badge fold does not know", md, n, c.r)
				}
			}
		}
	}
}

// allMedals is a badge of every tier and state, and the three hand-drawn
// ones, as the plain tier and the others draw them.
func allMedals() []medal {
	var out []medal
	for _, tier := range []config.BadgeTier{config.BadgeBronze, config.BadgeSilver, config.BadgeGold, config.BadgePlatinum, config.BadgeLegendary} {
		for _, state := range []medalState{medalEarned, medalLocked, medalHidden} {
			out = append(out, medal{tier: tier, emblem: 'x', state: state, phase: 5})
		}
		out = append(out, medal{tier: tier, emblem: 'x', crossed: true})
	}
	for _, sp := range []string{specialCookieJar, specialLedger, specialSource} {
		out = append(out, medal{emblem: 'x', special: sp, phase: 3}, medal{emblem: 'x', special: sp, crossed: true})
	}
	return out
}

// TestBadgeFramesArePure: a frame is a function of the badge and the frame
// number and nothing else. Drawing the same frame twice gives the same
// cells; bronze, silver and gold never move; platinum, legendary and the
// hand-drawn badges do; and at rest nothing moves at all. Every frame
// stays inside the badge's own box and on the cell grid.
func TestBadgeFramesArePure(t *testing.T) {
	for _, md := range allMedals() {
		w, h := medalSize(md)
		draw := func(n int, plain bool) string {
			// A margin all round: a frame that draws outside its box shows.
			g := newTGrid(w+4, h+4)
			drawMedal(g, 2, 2, md, n, plain)
			for i, c := range g.c {
				x, y := i%g.w, i/g.w
				if c.r != ' ' && (x < 2 || x >= 2+w || y < 2 || y >= 2+h) {
					t.Fatalf("%+v frame %d draws %q at %d,%d, outside its %dx%d box", md, n, c.r, x-2, y-2, w, h)
				}
				if c.r != ' ' && uniseg.StringWidth(string(c.r)) != 1 {
					t.Fatalf("%+v frame %d draws %q, which is not one cell wide", md, n, c.r)
				}
			}
			var sb strings.Builder
			for _, c := range g.c {
				fmt.Fprintf(&sb, "%c%d.%d.%d ", c.r, c.st, c.ink, c.fl)
			}
			return sb.String()
		}
		for _, plain := range []bool{false, true} {
			rest := draw(atRest, plain)
			if rest != draw(atRest, plain) {
				t.Errorf("%+v: the frame at rest is not the same twice", md)
			}
			moved := false
			for n := 0; n < 96; n++ {
				f := draw(n, plain)
				if f != draw(n, plain) {
					t.Fatalf("%+v: frame %d is not the same twice", md, n)
				}
				moved = moved || f != rest
			}
			if want := md.animated(); moved != want {
				t.Errorf("%+v plain=%v: moves=%v, want %v", md, plain, moved, want)
			}
		}
	}
	// The mini: only the legendary one moves.
	for _, md := range allMedals() {
		rest := newTGrid(miniW, miniH)
		drawMini(rest, 0, 0, md, atRest)
		moved := false
		for n := 0; n < 96; n++ {
			g := newTGrid(miniW, miniH)
			drawMini(g, 0, 0, md, n)
			for i := range g.c {
				moved = moved || g.c[i] != rest.c[i]
			}
		}
		if want := md.state == medalEarned && md.tier == config.BadgeLegendary && md.special == ""; moved != want {
			t.Errorf("%+v: the mini moves=%v, want %v", md, moved, want)
		}
	}
}

// TestLegendaryShimmerShowsInThePlainTier: the band of light that crosses
// a legendary rim is a change of colour, and in the plain tier a change of
// glyph, so it shows without colour too.
func TestLegendaryShimmerShowsInThePlainTier(t *testing.T) {
	md := medal{tier: config.BadgeLegendary, emblem: 'x'}
	band := func(plain bool) (cells int) {
		for n := 0; n < 24; n++ {
			g := newTGrid(17, 9)
			drawMedal(g, 0, 0, md, n, plain)
			for _, c := range g.c {
				if plain && c.r == '@' || !plain && c.r == '█' && c.st == tsBright && c.ink == inkNone {
					cells++
				}
			}
		}
		return cells
	}
	if band(false) == 0 {
		t.Error("no band crosses the legendary rim")
	}
	if band(true) == 0 {
		t.Error("the plain tier has no mark for the band")
	}
}

// TestBadgeArtPaintsInEveryTheme paints the case, the detail and every
// frame of the moving badges on a screen in every theme, and holds the
// colours to the rules:
//
//   - nothing is drawn in its own background colour;
//   - art and text clear a contrast of 3.0 against what they sit on (a
//     hidden badge's slab is the one thing that is quiet on purpose, and
//     it still stands off the canvas);
//   - a hidden slab never outshines an earned bronze badge beside it;
//   - the five tiers keep five colours, on light themes too, where the
//     pale metals would otherwise all darken to one grey (themes that draw
//     in one ink are told apart by shape alone).
func TestBadgeArtPaintsInEveryTheme(t *testing.T) {
	views, sum := caseFixture()
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		pal := newGridPalette(0, nil)
		check := func(where string, g *tGrid) {
			scr := tcell.NewSimulationScreen("UTF-8")
			if err := scr.Init(); err != nil {
				t.Fatal(err)
			}
			defer scr.Fini()
			scr.SetSize(g.w, g.h)
			paintGrid(scr, g, 0, 0, pal)
			scr.Show()
			cells, cw, _ := scr.GetContents()
			for i, c := range cells {
				if len(c.Runes) == 0 || c.Runes[0] == ' ' {
					continue
				}
				fg, bg, _ := c.Style.Decompose()
				if fg == bg {
					t.Fatalf("theme %s, %s: cell %d,%d (%q) is drawn in its own background colour", th.Key, where, i%cw, i/cw, c.Runes[0])
				}
				min := 3.0
				if g.c[i].st == tsSlab && g.c[i].ink == inkNone && g.c[i].fl&cfSel == 0 {
					min = 1.08
				}
				if r := theme.ContrastRatio(fg, bg); r < min-0.01 {
					t.Fatalf("theme %s, %s: cell %d,%d (%q, style %d, ink %d) has a contrast of %.2f, under %.2f", th.Key, where, i%cw, i/cw, c.Runes[0], g.c[i].st, g.c[i].ink, r, min)
				}
			}
		}
		for _, size := range treeSizes {
			w, h := size[0], size[1]-promptRows
			for _, sel := range []string{"age.stone_age", "ladder.prestiges.4", "special.small_world", "hidden.age", "special.creative_accounting"} {
				for _, card := range []bool{false, true} {
					_, _, g := drawCase(views, sum, w, h, caseView{sel: sel, card: card, tier: mapmodel.TierUnicode}, 7)
					check(fmt.Sprintf("the case at %dx%d on %s", w, h, sel), g)
				}
			}
		}
		for _, md := range allMedals() {
			for n := atRest; n < 48; n++ {
				w, h := medalSize(md)
				g := newTGrid(w, h)
				drawMedal(g, 0, 0, md, n, false)
				check(fmt.Sprintf("%+v frame %d", md, n), g)
			}
		}

		// The slab against bronze.
		slab, _, _ := pal.styles[tsSlab].Decompose()
		bronze := pal.inkOn(inkBronze, false)
		if theme.ContrastRatio(slab, pal.bg) >= theme.ContrastRatio(bronze, pal.bg) {
			t.Errorf("theme %s: a hidden slab (%.2f against the canvas) outshines bronze (%.2f)", th.Key,
				theme.ContrastRatio(slab, pal.bg), theme.ContrastRatio(bronze, pal.bg))
		}

		// The stops of a rim keep their order of emphasis: pale stands out
		// most, dark least. (A theme of one ink may bring two of them to
		// the least contrast art may have, but never turns them round.)
		for _, tier := range []config.BadgeTier{config.BadgeBronze, config.BadgeSilver, config.BadgeGold, config.BadgePlatinum} {
			d, b, p := tierInks(tier)
			cd, cb, cp := theme.ContrastRatio(pal.inkOn(d, false), pal.bg), theme.ContrastRatio(pal.inkOn(b, false), pal.bg), theme.ContrastRatio(pal.inkOn(p, false), pal.bg)
			if ordered := cd <= cb && cb <= cp && cd < cp; !ordered || !th.Duotone && !(cd < cb && cb < cp) {
				t.Errorf("theme %s: the %s rim's stops stand out %.2f, %.2f, %.2f (dark, base, pale): not in order", th.Key, tier.Name(), cd, cb, cp)
			}
		}

		// Five tiers, five colours.
		if th.Duotone {
			continue
		}
		base := map[string]tcell.Color{
			"bronze": pal.inkOn(inkBronze, false), "silver": pal.inkOn(inkSilver, false),
			"gold": pal.inkOn(inkGold, false), "platinum": pal.inkOn(inkPlatinum, false),
			"legendary": pal.inkOn(tierInk(config.BadgeLegendary), false),
		}
		names := []string{"bronze", "silver", "gold", "platinum", "legendary"}
		for i, a := range names {
			for _, b := range names[i+1:] {
				if d := colourDistance(base[a], base[b]); d < tierColourGap {
					t.Errorf("theme %s: %s (%06x) and %s (%06x) are %.3f apart, too close to tell apart (want %.2f)", th.Key, a, base[a].Hex(), b, base[b].Hex(), d, tierColourGap)
				}
			}
		}
	}
}

// tierColourGap is the least CIEDE2000 distance between two tiers'
// colours on a theme (on go-colorful's scale, where 1.0 is black to
// white): about 7 of the usual units, a difference seen at a glance.
const tierColourGap = 0.07

func colourDistance(a, b tcell.Color) float64 {
	col := func(c tcell.Color) colorful.Color {
		r, g, bl := c.RGB()
		return colorful.Color{R: float64(r) / 255, G: float64(g) / 255, B: float64(bl) / 255}
	}
	return col(a).DistanceCIEDE2000(col(b))
}

// TestEveryInkHasAColour: every ink of badge art has a colour for a dark
// theme and one for a light theme, and the two differ wherever a pale
// colour would have to be darkened on a light canvas.
func TestEveryInkHasAColour(t *testing.T) {
	for k := inkNone + 1; k < inkPrism+3*prismHues; k++ {
		for _, light := range []bool{false, true} {
			if c := inkValue(k, light); !c.IsRGB() {
				t.Errorf("ink %d has no colour (light=%v)", k, light)
			}
		}
	}
	for _, tier := range []config.BadgeTier{config.BadgeBronze, config.BadgeSilver, config.BadgeGold, config.BadgePlatinum} {
		d, b, p := tierInks(tier)
		if d+1 != b || b+1 != p {
			t.Errorf("%s: its inks are %d, %d, %d, not a dark, a base and a pale stop in a row", tier.Name(), d, b, p)
		}
	}
}

// TestBadgeEmblemsAreMapGlyphs: every badge of the catalog names an emblem
// the case knows, and every emblem is a symbol the maps draw, in all three
// glyph tiers, one cell wide. The integrity badges have a sprite each.
func TestBadgeEmblemsAreMapGlyphs(t *testing.T) {
	onMap := map[rune]bool{}
	for _, g := range mapmodel.AllGlyphs() {
		onMap[g.Unicode] = true
	}
	check := func(name string, e badgeEmblem) {
		if !onMap[e.glyph.Unicode] {
			t.Errorf("the emblem %q is %q, which is not a symbol the maps draw", name, e.glyph.Unicode)
		}
		if e.glyph.ASCII < 0x21 || e.glyph.ASCII > 0x7e {
			t.Errorf("the emblem %q has no plain glyph (%q)", name, e.glyph.ASCII)
		}
		for _, tier := range caseTiers {
			if r := e.glyph.In(tier); uniseg.StringWidth(string(r)) != 1 && tier != mapmodel.TierNerd {
				t.Errorf("the emblem %q is %q in the %s tier, which is not one cell wide", name, r, tier)
			}
		}
	}
	for name, e := range badgeEmblems {
		check(name, e)
	}
	specials := map[string]bool{}
	integrity, drawn := 0, 0
	for _, def := range rules.Core().Badges() {
		e, ok := emblemOf(def.Emblem)
		if !ok {
			t.Errorf("%s names the emblem %q, which the case does not know", def.Key, def.Emblem)
			continue
		}
		check(def.Key, e)
		if e.special != "" {
			if specials[e.special] {
				t.Errorf("%s shares its sprite %q with another badge", def.Key, e.special)
			}
			specials[e.special] = true
		}
		switch {
		case def.Integrity():
			integrity++
			if e.special == "" {
				t.Errorf("%s is an integrity badge without a sprite", def.Key)
			}
		case e.special != "":
			// A hand-drawn sprite is for the rarest few: a legendary badge.
			drawn++
			if def.Tier != config.BadgeLegendary {
				t.Errorf("%s (%s) wears the sprite %q: the hand-drawn sprites are the legendary badges'", def.Key, def.Tier.Name(), e.special)
			}
			if _, ok := legendSprites[e.special]; !ok {
				t.Errorf("%s names the sprite %q, which is not drawn", def.Key, e.special)
			}
		case def.Tier == config.BadgeLegendary && def.Family == "special":
			t.Errorf("%s is a legendary special without a sprite of its own", def.Key)
		}
	}
	if integrity != 3 {
		t.Errorf("the catalog has %d integrity badges with sprites, want 3", integrity)
	}
	if drawn != len(legendSprites) {
		t.Errorf("%d badges wear a legendary sprite and %d are drawn: one each", drawn, len(legendSprites))
	}
	for _, lin := range mapmodel.LineageOrder {
		if _, ok := emblemOf(emblemLineage + lin); !ok {
			t.Errorf("the lineage %s has no emblem", lin)
		}
	}
	if _, ok := emblemOf("no_such_emblem"); ok {
		t.Error("an unknown emblem resolved")
	}
}

// TestBadgeCaseKeepsTheCountOfHiddenBadges: while the account may not know
// how many badges are withheld, the case does not let them be counted
// either. A family's withheld badges are one slab, whatever their number,
// and the header says "???". Once the number may be known, each has a slab
// of its own. A secret badge is listed by itself either way, with its
// hint, and never with its name.
func TestBadgeCaseKeepsTheCountOfHiddenBadges(t *testing.T) {
	views, sum := caseFixture()
	slabs := func(m *caseModel) (groups, singles int) {
		for _, it := range m.items {
			switch {
			case it.group:
				groups++
			case it.v.Hidden && !it.v.Secret:
				singles++
			}
		}
		return
	}
	m, _, g := drawCase(views, sum, 120, 37, caseView{tier: mapmodel.TierUnicode}, atRest)
	if groups, singles := slabs(m); groups != 2 || singles != 0 {
		t.Errorf("with the count withheld: %d family slabs and %d single ones, want 2 and 0", groups, singles)
	}
	text := g.String()
	if !strings.Contains(text, "??? hidden") {
		t.Errorf("the header does not say ??? hidden:\n%s", text)
	}
	more := append([]game.BadgeView(nil), views...)
	more = append(more, game.BadgeView{Key: "age.quantum_age", Family: "age", Name: game.BadgeHiddenName, Hidden: true})
	if _, _, g2 := drawCase(more, sum, 120, 37, caseView{tier: mapmodel.TierUnicode}, atRest); g2.String() != text {
		t.Error("one more withheld badge changes the picture: the case lets them be counted")
	}
	counted := sum
	counted.HiddenCounted = true
	m, _, g = drawCase(views, counted, 120, 37, caseView{tier: mapmodel.TierUnicode}, atRest)
	if groups, singles := slabs(m); groups != 0 || singles != 4 {
		t.Errorf("with the count known: %d family slabs and %d single ones, want 0 and 4", groups, singles)
	}
	if !strings.Contains(g.String(), fmt.Sprintf("%d hidden", sum.Hidden)) {
		t.Errorf("the header does not give the number once it may:\n%s", g.String())
	}
	// The secret: its hint, never its name.
	_, _, g = drawCase(views, sum, 120, 37, caseView{sel: "special.liquidation_sale", tier: mapmodel.TierUnicode}, atRest)
	if text := g.String(); !strings.Contains(text, "Something about a clearance.") || strings.Contains(text, "Liquidation") {
		t.Errorf("the secret badge's pane:\n%s", text)
	}
}

// TestBadgeCaseLadders: a ladder's rungs sit on one line in order, the
// line names the ladder and says how far its next rung is, and the status
// line says the same for the selected rung. The Next tab lists the badges
// nearest their count, nearest first.
func TestBadgeCaseLadders(t *testing.T) {
	views, sum := caseFixture()
	m, _, g := drawCase(views, sum, 100, 27, caseView{tab: "lineage", sel: "lineage.housing.2", tier: mapmodel.TierUnicode}, atRest)
	var housing *caseLine
	for i := range m.lines {
		if m.lines[i].ladder == "Housing" {
			housing = &m.lines[i]
		}
	}
	if housing == nil || len(housing.items) != 5 {
		t.Fatalf("the Housing ladder is not five rungs on a line: %+v", housing)
	}
	for i, idx := range housing.items {
		if want := fmt.Sprintf("lineage.housing.%d", i+1); m.items[idx].id != want {
			t.Errorf("rung %d of the line is %s, want %s", i+1, m.items[idx].id, want)
		}
	}
	if housing.next == nil || housing.next.Key != "lineage.housing.3" {
		t.Errorf("the ladder's next rung is %+v, want the third", housing.next)
	}
	text := g.String()
	for _, want := range []string{"Housing", "131 / 180", "Next rung: Housing Magnate", "Rung 2 of 5"} {
		if !strings.Contains(text, want) {
			t.Errorf("the case does not say %q:\n%s", want, text)
		}
	}
	m, _, g = drawCase(views, sum, 100, 27, caseView{tab: caseTabNext, tier: mapmodel.TierUnicode}, atRest)
	var order []string
	for _, it := range m.items {
		order = append(order, it.id)
	}
	if len(order) < 3 || order[0] != "lineage.housing.3" || order[1] != "lineage.military.1" {
		t.Errorf("the Next tab lists %v: want the Housing Magnate (131 of 180) first, then the Military Hobbyist (4 of 9)", order)
	}
	if text := g.String(); !strings.Contains(text, "Housing Magnate") || !strings.Contains(text, "131 / 180") {
		t.Errorf("the Next tab:\n%s", text)
	}
}

// TestBadgePanelKeysAndMotion drives the panel as the dashboard does: the
// arrows move, Tab steps through the families and wraps round, Enter opens
// and closes the detail and Esc closes it first, and keys that belong to
// the prompt are left to it. With the motion setting off the panel draws
// the same picture whatever the clock says and reports nothing moving.
func TestBadgePanelKeysAndMotion(t *testing.T) {
	views, sum := caseFixture()
	st := game.GameState{AccountStats: &game.AccountStatsView{Badges: views, BadgeSummary: sum}}
	set := mapSettings{Tier: mapmodel.TierUnicode, Motion: true}
	clock := time.Unix(1_700_000_000, 0)
	p := newBadgePanel()
	p.settings = func() mapSettings { return set }
	p.now = func() time.Time { return clock }
	p.start = clock
	p.SetRect(0, 0, 100, 27)
	draw := func() string {
		scr := tcell.NewSimulationScreen("UTF-8")
		if err := scr.Init(); err != nil {
			t.Fatal(err)
		}
		defer scr.Fini()
		scr.SetSize(100, 27)
		p.Draw(scr)
		scr.Show()
		var sb strings.Builder
		cells, _, _ := scr.GetContents()
		for _, c := range cells {
			fg, bg, _ := c.Style.Decompose()
			fmt.Fprintf(&sb, "%c%d.%d ", c.Runes[0], fg, bg)
		}
		return sb.String()
	}
	press := func(k tcell.Key) bool { return p.routeKey(tcell.NewEventKey(k, 0, tcell.ModNone), "") }
	p.open(st)
	draw()
	if p.view.sel != "age.stone_age" || p.view.tab != caseTabAll {
		t.Fatalf("the case opens on %q in tab %q", p.view.sel, p.view.tab)
	}
	press(tcell.KeyRight)
	if p.view.sel != "age.iron_age" {
		t.Errorf("Right selects %q", p.view.sel)
	}
	press(tcell.KeyDown)
	if p.view.sel != "lineage.housing.2" {
		t.Errorf("Down selects %q, want the badge under it", p.view.sel)
	}
	press(tcell.KeyUp)
	press(tcell.KeyLeft)
	if p.view.sel != "age.stone_age" {
		t.Errorf("Up and Left come back to %q", p.view.sel)
	}
	press(tcell.KeyEnd)
	if p.view.sel != "special.creative_accounting" {
		t.Errorf("End selects %q", p.view.sel)
	}
	press(tcell.KeyHome)
	var tabs []string
	for range caseTabs(views, false) {
		press(tcell.KeyTab)
		tabs = append(tabs, p.view.tab)
	}
	if got := strings.Join(tabs, ","); got != "age,lineage,ladder,special,next," {
		t.Errorf("Tab steps through %q", got)
	}
	press(tcell.KeyBacktab)
	if p.view.tab != caseTabNext {
		t.Errorf("Shift-Tab from All opens %q", p.view.tab)
	}
	press(tcell.KeyTab)
	if !press(tcell.KeyEnter) || !p.view.card {
		t.Error("Enter does not open the detail")
	}
	if !p.closeCard() || p.view.card || p.closeCard() {
		t.Error("Esc closes the detail once")
	}
	// The prompt's keys stay the prompt's.
	if p.routeKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), "build hut") || p.routeKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), "bu") {
		t.Error("with something typed, the panel took Enter or Tab from the prompt")
	}
	if p.routeKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone), "") {
		t.Error("the panel took a letter")
	}

	// Motion: a legendary badge selected moves with the clock.
	p.request(badgeRequest{sel: "ladder.prestiges.4"})
	p.update(st)
	a := draw()
	if !p.moving.Load() {
		t.Error("a legendary badge is selected and the panel reports nothing moving")
	}
	clock = clock.Add(13 * mapAnimStep)
	if a == draw() {
		t.Error("the legendary badge did not move with the clock")
	}
	// A bronze badge selected, no legendary on the page: nothing moves.
	p.request(badgeRequest{tab: "age", setTab: true})
	p.update(st)
	a = draw()
	clock = clock.Add(13 * mapAnimStep)
	if p.moving.Load() || a != draw() {
		t.Error("a page of badges that hold still is drawn as moving")
	}
	// Motion off: the picture is the same whatever the clock says.
	set.Motion = false
	p.request(badgeRequest{sel: "ladder.prestiges.4"})
	p.update(st)
	a = draw()
	clock = clock.Add(17 * mapAnimStep)
	if a != draw() || p.moving.Load() || p.frame() != atRest {
		t.Error("with motion off the case still moves")
	}
	for _, sel := range []string{"special.hand_in_the_cookie_jar", "special.touched_by_the_source", "special.creative_accounting", "special.small_world"} {
		p.request(badgeRequest{sel: sel, card: true})
		p.update(st)
		a = draw()
		clock = clock.Add(29 * mapAnimStep)
		if a != draw() {
			t.Errorf("with motion off %s still moves", sel)
		}
	}
}
