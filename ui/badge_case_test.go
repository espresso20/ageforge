package ui

import (
	"bytes"
	"fmt"
	"image"
	"strings"
	"testing"
	"time"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// caseTiers are the glyph tiers the case is checked in.
var caseTiers = []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII}

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

// coinAt is the room the case gave a badge's coin, and whether it drew one.
func coinAt(m *caseModel, id string) (image.Rectangle, bool) {
	for _, c := range m.coins {
		if c.it.id == id {
			return c.r, true
		}
	}
	return image.Rectangle{}, false
}

// TestBadgeCaseAtEverySize draws the case at the four sizes the game is
// checked at, in every glyph tier, on every tab and with every badge
// selected in turn, and holds the picture to what the panel promises:
//
//   - the selected badge's coin is on screen, whole, inside the grid;
//   - no badge is ever cut: every coin of a line that shows has its whole
//     room, and a line that does not fit draws nothing;
//   - the selected badge's medal is beside the grid, in its room, where the
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
					m, v, g := drawCase(views, sum, w, h, caseView{tab: tab.key, sel: sel.id, tier: tier})
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
					// The key bar is the chrome grid's foot row: buttons, whole
					// and inside the panel, with Esc among them.
					foot := m.chrome.Lines()[h-1]
					if !strings.Contains(foot, "Esc") {
						t.Errorf("%s: the key bar is missing: %q", where, foot)
					}
					if tier == mapmodel.TierASCII {
						for i, c := range m.chrome.c {
							if c.r >= 0x80 {
								t.Fatalf("%s: the plain tier draws %q at %d,%d of the buttons", where, c.r, i%w, i/w)
							}
						}
					}
					escKey := false
					for _, k := range m.chrome.keys {
						if k.r.Min.X < 0 || k.r.Max.X > w || k.r.Min.Y < 0 || k.r.Max.Y > h {
							t.Errorf("%s: a key button at %v is outside the panel", where, k.r)
						}
						escKey = escKey || (k.key.Code == kit.KeyEsc && k.r.Min.Y == h-1)
					}
					if !escKey {
						t.Errorf("%s: no key button of the key bar presses Esc", where)
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
						if shown && (y < lay.gy || y+lay.th > lay.gy+lay.gh) {
							t.Errorf("%s: a line of badges is drawn at rows %d to %d, outside the grid (%d to %d)", where, y, y+lay.th-1, lay.gy, lay.gy+lay.gh-1)
						}
						for _, i := range ln.items {
							o := &m.items[i]
							x := lay.gx + o.col*lay.pitchX()
							room, drawn := coinAt(m, o.id)
							if !shown {
								if drawn {
									t.Errorf("%s: %s has a coin on a line that does not show", where, o.id)
								}
								continue
							}
							if x < lay.gx || x+lay.tw > lay.gx+lay.gw {
								t.Errorf("%s: %s is drawn at columns %d to %d, outside the grid (%d to %d)", where, o.id, x, x+lay.tw-1, lay.gx, lay.gx+lay.gw-1)
							}
							if want := image.Rect(x, y, x+coinCols, y+coinRows); !drawn || room != want {
								t.Errorf("%s: %s has the room %v, want %v", where, o.id, room, want)
							}
							if o.id == it.id {
								selShown = true
								if g.at(x, y).fl&cfSel == 0 {
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

					// The selected badge's medal, beside the grid.
					aw, ah := medalPixCols, medalPixRows
					ax, ay := lay.px+(lay.pw-aw)/2, lay.py
					if ax < lay.px || ax+aw > w || ay+ah > lay.py+lay.ph {
						t.Fatalf("%s: the medal (%dx%d) does not fit its pane (%dx%d)", where, aw, ah, lay.pw, lay.ph)
					}
					if want := image.Rect(ax, ay, ax+aw, ay+ah); m.medalRoom != want || m.medalFor != it {
						t.Errorf("%s: the medal's room is %v for %v, want %v for %s", where, m.medalRoom, m.medalFor, want, it.id)
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
				m, v, g := drawCase(views, sum, w, h, caseView{sel: bv.Key, card: true, tier: tier})
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
				if room := m.medalRoom; m.medalFor != it || room.Dx() != medalPixCols || room.Dy() != medalPixRows || room.Min.Y <= top || room.Max.Y > bottom || !room.In(m.cardAt) {
					t.Errorf("%s: the medal's room is %v for %v, not whole inside the detail %v", where, room, m.medalFor, m.cardAt)
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
				m, v, g := drawCase(views, sum, w, h, caseView{tab: tab, sel: "lineage.housing.2", tier: tier})
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
						x := m.lay.gx + o.col*m.lay.pitchX()
						if x+m.lay.tw > m.lay.gx+m.lay.gw || y+m.lay.th > m.lay.gy+m.lay.gh {
							t.Errorf("%dx%d: %s is drawn outside the grid", w, h, o.id)
						}
						if room, ok := coinAt(m, o.id); !ok || room != image.Rect(x, y, x+coinCols, y+coinRows) {
							t.Errorf("%dx%d: %s has the room %v, want its whole room", w, h, o.id, room)
						}
					}
				}
			}
		}
	}
}

// TestLegendaryShimmerShowsInThePlainTier: a legendary coin is a picture in
// the plain tier too, and its picture changes with the clock while the
// picture of a coin that is not earned does not.
func TestLegendaryShimmerShowsInThePlainTier(t *testing.T) {
	p, paint, tick := picturePanel(t)
	p.view.tier = mapmodel.TierASCII
	first := paint()
	legend := -1
	for i, c := range p.model.coins {
		if p.look(c.it, coinArt).tier == config.BadgeLegendary && p.look(c.it, coinArt).earned() {
			legend = i
		}
	}
	if legend < 0 {
		t.Fatal("no earned legendary coin is in view")
	}
	if oneColour(first.Pics[legend].Img) {
		t.Error("the legendary coin is one flat colour in the plain tier")
	}
	moved := false
	for range 24 {
		tick(1)
		moved = moved || !bytes.Equal(paint().Pics[legend].Img.Pix, first.Pics[legend].Img.Pix)
	}
	if !moved {
		t.Error("the legendary coin's picture never changes in the plain tier")
	}
}

// TestBadgeArtPaintsInEveryTheme paints the case and the detail on a screen
// in every theme, and holds the colours to the rules:
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
	keepTheme(t)
	for _, th := range theme.All() {
		must(t, theme.SetActive(th.Key))
		pal := newGridPalette(0, nil)
		check := func(where string, g *tGrid) {
			scr := newSim()
			must(t, scr.Init())
			defer scr.Fini()
			scr.SetSize(g.w, g.h)
			paintGrid(scr, g, 0, 0, pal)
			scr.Show()
			cells, cw, _ := scr.GetContents()
			for i, c := range cells {
				if len(c.Runes) == 0 || c.Runes[0] == ' ' {
					continue
				}
				fg, bg, _ := decompose(c.Style)
				if fg == bg {
					t.Fatalf("theme %s, %s: cell %d,%d (%q) is drawn in its own background colour", th.Key, where, i%cw, i/cw, c.Runes[0])
				}
				min := 3.0
				if g.c[i].k == tsSlab && g.c[i].ink == inkNone && g.c[i].fl&cfSel == 0 {
					min = 1.08
				}
				if r := theme.ContrastRatio(fg, bg); r < min-0.01 {
					t.Fatalf("theme %s, %s: cell %d,%d (%q, style %d, ink %d) has a contrast of %.2f, under %.2f", th.Key, where, i%cw, i/cw, c.Runes[0], g.c[i].k, g.c[i].ink, r, min)
				}
			}
		}
		for _, size := range treeSizes {
			w, h := size[0], size[1]-promptRows
			for _, sel := range []string{"age.stone_age", "ladder.prestiges.4", "special.small_world", "hidden.age", "special.creative_accounting"} {
				for _, card := range []bool{false, true} {
					_, _, g := drawCase(views, sum, w, h, caseView{sel: sel, card: card, tier: mapmodel.TierUnicode})
					check(fmt.Sprintf("the case at %dx%d on %s", w, h, sel), g)
				}
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
		base := map[string]theme.Color{
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

func colourDistance(a, b theme.Color) float64 {
	col := func(c theme.Color) colorful.Color {
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
			if c := inkValue(k, light); !(c.Hex() >= 0) {
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
			if r := e.glyph.In(tier); uniseg.StringWidth(string(r)) != 1 {
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
			if _, ok := medalIcons[e.special]; !ok {
				t.Errorf("%s names the sprite %q, which is not drawn", def.Key, e.special)
			}
		case def.Tier == config.BadgeLegendary && def.Family == "special":
			t.Errorf("%s is a legendary special without a sprite of its own", def.Key)
		}
	}
	if integrity != 3 {
		t.Errorf("the catalog has %d integrity badges with sprites, want 3", integrity)
	}
	legend := 0
	for key := range medalIcons {
		if strings.HasPrefix(key, "legend.") {
			legend++
		}
	}
	if drawn != legend {
		t.Errorf("%d badges wear a legendary sprite and %d are drawn: one each", drawn, legend)
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
	m, _, g := drawCase(views, sum, 120, 37, caseView{tier: mapmodel.TierUnicode})
	if groups, singles := slabs(m); groups != 2 || singles != 0 {
		t.Errorf("with the count withheld: %d family slabs and %d single ones, want 2 and 0", groups, singles)
	}
	text := g.String()
	if !strings.Contains(text, "??? hidden") {
		t.Errorf("the header does not say ??? hidden:\n%s", text)
	}
	more := append([]game.BadgeView(nil), views...)
	more = append(more, game.BadgeView{Key: "age.quantum_age", Family: "age", Name: game.BadgeHiddenName, Hidden: true})
	if _, _, g2 := drawCase(more, sum, 120, 37, caseView{tier: mapmodel.TierUnicode}); g2.String() != text {
		t.Error("one more withheld badge changes the picture: the case lets them be counted")
	}
	counted := sum
	counted.HiddenCounted = true
	m, _, g = drawCase(views, counted, 120, 37, caseView{tier: mapmodel.TierUnicode})
	if groups, singles := slabs(m); groups != 0 || singles != 4 {
		t.Errorf("with the count known: %d family slabs and %d single ones, want 0 and 4", groups, singles)
	}
	if !strings.Contains(g.String(), fmt.Sprintf("%d hidden", sum.Hidden)) {
		t.Errorf("the header does not give the number once it may:\n%s", g.String())
	}
	// The secret: its hint, never its name.
	_, _, g = drawCase(views, sum, 120, 37, caseView{sel: "special.liquidation_sale", tier: mapmodel.TierUnicode})
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
	m, _, g := drawCase(views, sum, 100, 27, caseView{tab: "lineage", sel: "lineage.housing.2", tier: mapmodel.TierUnicode})
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
	wantAll(t, "the case", text, "Housing", "131 / 180", "Next rung: Housing Magnate", "Rung 2 of 5")
	m, _, g = drawCase(views, sum, 100, 27, caseView{tab: caseTabNext, tier: mapmodel.TierUnicode})
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
		scr := newSim()
		must(t, scr.Init())
		defer scr.Fini()
		scr.SetSize(100, 27)
		drawOn(scr, p)
		scr.Show()
		var sb strings.Builder
		cells, _, _ := scr.GetContents()
		for _, c := range cells {
			fg, bg, _ := decompose(c.Style)
			fmt.Fprintf(&sb, "%c%d.%d ", c.Runes[0], fg, bg)
		}
		return sb.String()
	}
	press := func(k kit.KeyCode) bool { return p.routeKey(newKey(k, 0, 0), "") }
	p.open(st)
	draw()
	if p.view.sel != "age.stone_age" || p.view.tab != caseTabAll {
		t.Fatalf("the case opens on %q in tab %q", p.view.sel, p.view.tab)
	}
	press(kit.KeyRight)
	if p.view.sel != "age.iron_age" {
		t.Errorf("Right selects %q", p.view.sel)
	}
	press(kit.KeyDown)
	if p.view.sel != "lineage.housing.2" {
		t.Errorf("Down selects %q, want the badge under it", p.view.sel)
	}
	press(kit.KeyUp)
	press(kit.KeyLeft)
	if p.view.sel != "age.stone_age" {
		t.Errorf("Up and Left come back to %q", p.view.sel)
	}
	press(kit.KeyEnd)
	if p.view.sel != "special.creative_accounting" {
		t.Errorf("End selects %q", p.view.sel)
	}
	press(kit.KeyHome)
	var tabs []string
	for range caseTabs(views, false) {
		press(kit.KeyTab)
		tabs = append(tabs, p.view.tab)
	}
	if got := strings.Join(tabs, ","); got != "age,lineage,ladder,special,next," {
		t.Errorf("Tab steps through %q", got)
	}
	press(kit.KeyBacktab)
	if p.view.tab != caseTabNext {
		t.Errorf("Shift-Tab from All opens %q", p.view.tab)
	}
	press(kit.KeyTab)
	if !press(kit.KeyEnter) || !p.view.card {
		t.Error("Enter does not open the detail")
	}
	if !p.closeCard() || p.view.card || p.closeCard() {
		t.Error("Esc closes the detail once")
	}
	// The prompt's keys stay the prompt's.
	if p.routeKey(newKey(kit.KeyEnter, 0, 0), "build hut") || p.routeKey(newKey(kit.KeyTab, 0, 0), "bu") {
		t.Error("with something typed, the panel took Enter or Tab from the prompt")
	}
	if p.routeKey(newKey(kit.KeyRune, 'b', 0), "") {
		t.Error("the panel took a letter")
	}

	// Motion: a legendary badge selected moves with the clock.
	paint := func() *kit.Buffer {
		b := kit.NewBuffer(100, 27)
		p.Layout(image.Rect(0, 0, 100, 27))
		p.Paint(kit.NewPainter(b))
		return b
	}
	p.request(badgeRequest{sel: "ladder.prestiges.4"})
	p.update(st)
	a := paint()
	if !p.moving.Load() {
		t.Error("a legendary badge is selected and the panel reports nothing moving")
	}
	clock = clock.Add(13 * mapAnimStep)
	if samePictures(a, paint()) {
		t.Error("the legendary badge did not move with the clock")
	}
	// Motion off: the pictures are the same whatever the clock says.
	set.Motion = false
	p.request(badgeRequest{sel: "ladder.prestiges.4"})
	p.update(st)
	a = paint()
	clock = clock.Add(17 * mapAnimStep)
	if !samePictures(a, paint()) || p.moving.Load() {
		t.Error("with motion off the case still moves")
	}
	for _, sel := range []string{"special.hand_in_the_cookie_jar", "special.touched_by_the_source", "special.creative_accounting", "special.small_world"} {
		p.request(badgeRequest{sel: sel, card: true})
		p.update(st)
		a = paint()
		clock = clock.Add(29 * mapAnimStep)
		if !samePictures(a, paint()) {
			t.Errorf("with motion off %s still moves", sel)
		}
	}
}

// samePictures reports whether two buffers hold the same pictures, pixel for
// pixel, in the same places.
func samePictures(a, b *kit.Buffer) bool {
	if len(a.Pics) != len(b.Pics) {
		return false
	}
	for i := range a.Pics {
		if a.Pics[i].Rect != b.Pics[i].Rect || !bytes.Equal(a.Pics[i].Img.Pix, b.Pics[i].Img.Pix) {
			return false
		}
	}
	return true
}

// TestBadgePanelMouse: a click selects the badge under the pointer, a double
// click opens its detail as Enter does, a click outside the detail closes it,
// and the wheel moves the selection.
func TestBadgePanelMouse(t *testing.T) {
	views, sum := caseFixture()
	st := game.GameState{AccountStats: &game.AccountStatsView{Badges: views, BadgeSummary: sum}}
	clock := time.Unix(1_700_000_000, 0)
	p := newBadgePanel()
	p.now, p.start = func() time.Time { return clock }, clock
	p.SetRect(0, 0, 100, 27)
	p.open(st)
	drawPrimitive(t, p, 100, 27)
	lay := p.model.lay
	at := func(id string) image.Point {
		it := p.model.items[p.model.by[id]]
		return image.Pt(lay.gx+it.col*lay.pitchX()+1, lay.gy+p.model.lines[it.line].y-p.view.vy+1)
	}
	press := func(pt image.Point, clicks int) bool {
		return p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: clicks})
	}
	if !press(at("age.iron_age"), 1) || p.view.sel != "age.iron_age" || p.view.card {
		t.Fatalf("a click: selected %q, detail open %v", p.view.sel, p.view.card)
	}
	if press(image.Pt(0, 0), 1) {
		t.Error("a click on the title bar changed something")
	}
	tabs, _, _, _ := p.model.tabRow(p.view)
	if p.PointerAt(at("age.iron_age")) != kit.PointerHand || p.PointerAt(image.Pt(tabs[1]+2, 1)) != kit.PointerHand ||
		p.PointerAt(image.Pt(tabs[0]+1, 1)) != kit.PointerArrow || p.PointerAt(image.Pt(0, 0)) != kit.PointerArrow {
		t.Error("the hand should show over a badge and a closed tab, and not over the open tab or the title bar")
	}
	if !press(at("age.iron_age"), 2) || !p.view.card {
		t.Error("a double click did not open the detail")
	}
	if p.PointerAt(at("age.iron_age")) != kit.PointerArrow {
		t.Error("with the detail open the hand still shows over the badge behind it")
	}
	drawPrimitive(t, p, 100, 27)
	if c := p.model.cardAt; press(c.Min.Add(image.Pt(c.Dx()/2, c.Dy()/2)), 1) {
		t.Error("a click on the detail closed it")
	}
	if !press(image.Pt(0, 0), 1) || p.view.card {
		t.Error("a click outside the detail did not close it")
	}
	if !p.Mouse(kit.Mouse{Kind: kit.MouseWheel, Wheel: 1}) || p.view.sel != "lineage.housing.2" {
		t.Errorf("the wheel down selects %q, want the badge under it", p.view.sel)
	}
}

// TestBadgePanelTabsTakeClicks: a click on a tab's title opens that tab, as
// Tab reaching it does; a click on the arrow at an end of the row opens the
// nearest tab out of sight that way; a click on the row between tabs, or on
// the open tab, changes nothing.
func TestBadgePanelTabsTakeClicks(t *testing.T) {
	views, sum := caseFixture()
	st := game.GameState{AccountStats: &game.AccountStatsView{Badges: views, BadgeSummary: sum}}
	clock := time.Unix(1_700_000_000, 0)
	p := newBadgePanel()
	p.now, p.start = func() time.Time { return clock }, clock
	const w = 44 // too narrow for every tab: the row has more to the right
	p.SetRect(0, 0, w, 27)
	p.open(st)
	drawPrimitive(t, p, w, 27)
	press := func(x int) bool {
		return p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: x, Y: 1, Clicks: 1})
	}
	at, open, left, right := p.model.tabRow(p.view)
	if open != 0 || left || !right || at[1] < 0 || at[len(at)-1] >= 0 {
		t.Fatalf("the fixture's tab row at %d columns: open %d, more left %v, more right %v, columns %v", w, open, left, right, at)
	}
	if press(at[0] + 1) {
		t.Error("a click on the open tab changed something")
	}
	if press(at[1] - 1) {
		t.Error("a click between two tabs changed something")
	}
	second := p.model.tabs[1].key
	p.view.vy = 3
	if !press(at[1]+2) || p.view.tab != second || p.view.vy != 0 {
		t.Errorf("a click on the second tab: open %q (want %q), scrolled to %d (want its top)", p.view.tab, second, p.view.vy)
	}
	// The arrow at the right end opens the first tab that did not fit.
	at, _, _, right = p.model.tabRow(p.view)
	hidden := -1
	for i := range at {
		if at[i] < 0 && i > 1 {
			hidden = i
			break
		}
	}
	if !right || hidden < 0 {
		t.Fatalf("after the click the row has nothing more to the right: %v", at)
	}
	if want := p.model.tabs[hidden].key; !press(w-1) || p.view.tab != want {
		t.Errorf("a click on the right arrow opened %q, want %q, the nearest tab out of sight", p.view.tab, want)
	}
	if at, _, _, _ = p.model.tabRow(p.view); at[hidden] < 0 {
		t.Error("the tab the arrow opened is still out of sight")
	}
}

// paintGrid puts a rendered grid on the screen in theme colours.
func paintGrid(scr *simScreen, g *tGrid, x0, y0 int, p *gridPalette) {
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			c := g.c[y*g.w+x]
			scr.SetContent(x0+x, y0+y, c.r, nil, p.style(c))
		}
	}
}
