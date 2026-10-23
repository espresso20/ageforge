package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/vibe"
)

// wonders_view.go is the Wonders panel's second view, the gallery: the art
// of every wonder the town has built, one at a time, drawn by the skyline
// map's own wonder renderer (mapstyle.Portraitist) on a plain backdrop as
// large as the panel, with its name and age under it. G or Tab opens it from
// the list and closes it again; Left and Right (Up and Down too) step through
// the wonders built, oldest first; Esc goes back to the list, and the next
// Esc closes the panel. A wonder not built yet is not in it.
//
// The panel is the text overlay it always was: the picture is read back off
// an offscreen screen, cell by cell, as colour tags, so it follows the
// theme, the glyph tier and the terminal's size like the rest of the text.

// wondersView is the panel's state: which view it is on and which wonder the
// gallery shows.
type wondersView struct {
	gallery bool
	idx     int
	// key is the wonder the gallery is on, so a wonder built while the panel
	// is open does not move the picture.
	key string

	// size is the panel's inner size, tier the map's glyph tier, model the
	// map model of a state (the same builder the Map panel uses).
	size  func() (w, h int)
	tier  func() mapmodel.GlyphTier
	model func(state *game.GameState) *mapmodel.Model
	art   mapstyle.Portraitist
	reg   *mapstyle.Registry
}

// builtWonders lists the wonders the town has built, oldest first.
func builtWonders(state game.GameState) []wonderInfo {
	var out []wonderInfo
	for _, w := range getWonderList(state.Ruleset()) {
		if bs, ok := state.Buildings[w.key]; ok && bs.Count > 0 {
			out = append(out, w)
		}
	}
	return out
}

// provider is the panel's text: the list, or the gallery.
func (wv *wondersView) provider(state game.GameState, w int) string {
	if !wv.gallery {
		return wondersProvider(state, w)
	}
	return wv.galleryText(state)
}

// open starts the panel on the list.
func (wv *wondersView) open() { wv.gallery, wv.key = false, "" }

// back is Esc: out of the gallery, to the list. It reports whether it did
// (the next Esc closes the panel).
func (wv *wondersView) back() bool {
	if !wv.gallery {
		return false
	}
	wv.gallery = false
	return true
}

// handleKey applies a key to the panel's own views and reports whether it
// took it: the rest (scrolling the list) is the text view's.
func (wv *wondersView) handleKey(ev *tcell.EventKey, state game.GameState) bool {
	toggle := ev.Key() == tcell.KeyTab || ev.Key() == tcell.KeyRune && (ev.Rune() == 'g' || ev.Rune() == 'G')
	if !wv.gallery {
		if !toggle {
			return false
		}
		wv.gallery = true
		list := builtWonders(state)
		wv.idx = len(list) - 1 // the newest
		if wv.idx >= 0 {
			wv.key = list[wv.idx].key
		}
		return true
	}
	list := builtWonders(state)
	wv.sync(list)
	step := 0
	switch ev.Key() {
	case tcell.KeyLeft, tcell.KeyUp, tcell.KeyPgUp:
		step = -1
	case tcell.KeyRight, tcell.KeyDown, tcell.KeyPgDn:
		step = 1
	case tcell.KeyHome:
		wv.idx = 0
	case tcell.KeyEnd:
		wv.idx = len(list) - 1
	default:
		if !toggle {
			return false
		}
		wv.gallery = false
		return true
	}
	wv.idx = clamp(wv.idx+step, 0, max(0, len(list)-1))
	if wv.idx < len(list) {
		wv.key = list[wv.idx].key
	}
	return true
}

// sync keeps the gallery on the wonder it was showing when the list changes.
func (wv *wondersView) sync(list []wonderInfo) {
	for i, w := range list {
		if w.key == wv.key {
			wv.idx = i
			return
		}
	}
	wv.idx = clamp(wv.idx, 0, max(0, len(list)-1))
	if wv.idx < len(list) {
		wv.key = list[wv.idx].key
	}
}

// galleryKeys is the line of keys under the picture.
func galleryKeys(many bool) string {
	if !many {
		return vibe.Line("ui.wonders_view.galleryKeys.01", "[gray]Tab or G the list   Esc back[-]")
	}
	return vibe.Line("ui.wonders_view.galleryKeys.02", "[gray]◄ ► step through them   Home End first, last   Tab or G the list   Esc back[-]")
}

// galleryRows is how many rows the picture gets in a panel h rows tall: what
// is left under the header, the three lines of caption and the keys.
func galleryRows(h int) int { return max(6, h-5) }

// galleryText is the gallery for the wonder the panel is on.
func (wv *wondersView) galleryText(state game.GameState) string {
	list := builtWonders(state)
	wv.sync(list)
	w, h := 80, 24
	if wv.size != nil {
		w, h = wv.size()
	}
	var sb strings.Builder
	sb.WriteString(vibe.Line("ui.wonders_view.galleryText.01", "[gold]═══ Wonder gallery ═══[-]"))
	if len(list) == 0 {
		fmt.Fprintf(&sb, "\n\n%s\n\n%s\n", vibe.Line("ui.wonders_view.galleryText.02", "[gray]No wonder built yet. Each one you build hangs here, drawn as the map draws it.[-]"), galleryKeys(false))
		return sb.String()
	}
	cur := list[wv.idx]
	fmt.Fprintf(&sb, "  %s\n", vibe.Sprintf("ui.wonders_view.galleryText.03", "[gray]%d of %d built[-]", wv.idx+1, len(list)))
	sb.WriteString(wv.picture(state, cur.key, w, galleryRows(h)))
	set := state.Ruleset()
	age := cur.ageName
	when := ""
	if yr, ok := set.AgeYear(cur.ageKey); ok {
		label := set.AgeYearLabel(cur.ageKey)
		if label == "" {
			label = config.YearLabel(yr)
		}
		when = vibe.Sprintf("ui.wonders_view.galleryText.04", "  [gray]the age began %s[-]", label)
	}
	fmt.Fprintf(&sb, "[gold]★ %s[-]  [gray]%s[-]%s\n", lit(cur.name), lit(age), when)
	desc := cur.def.Description
	if desc != "" {
		fmt.Fprintf(&sb, "[gray]%s[-]\n", lit(mapmodel.Clip(desc, max(10, w-1))))
	} else {
		sb.WriteString("\n")
	}
	sb.WriteString(galleryKeys(len(list) > 1))
	sb.WriteString("\n")
	return sb.String()
}

// picture is the wonder key drawn into w columns by rows rows, as colour
// tagged text, one line to a row.
func (wv *wondersView) picture(state game.GameState, key string, w, rows int) string {
	blank := strings.Repeat("\n", rows)
	if wv.art == nil && wv.reg != nil {
		if st, ok := wv.reg.New("skyline"); ok {
			wv.art, _ = st.(mapstyle.Portraitist)
		}
	}
	if wv.art == nil || wv.model == nil || w < 8 {
		return blank
	}
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		return blank
	}
	defer sim.Fini()
	sim.SetSize(w, rows)
	scr := theme.WrapScreen(sim)
	tier := mapmodel.TierUnicode
	if wv.tier != nil {
		tier = wv.tier()
	}
	if !wv.art.Portrait(scr, mapstyle.Rect{W: w, H: rows}, mapstyle.Frame{Model: wv.model(&state), Tier: tier}, key) {
		return blank
	}
	var sb strings.Builder
	for y := 0; y < rows; y++ {
		var run strings.Builder
		var fg, bg tcell.Color
		first := true
		flush := func() {
			if run.Len() > 0 {
				sb.WriteString(lit(run.String()))
				run.Reset()
			}
		}
		for x := 0; x < w; x++ {
			r, _, st, _ := sim.GetContent(x, y)
			cfg, cbg, _ := st.Decompose()
			if r == 0 {
				r = ' '
			}
			if first || cfg != fg || cbg != bg {
				flush()
				sb.WriteString(colorTag(cfg, cbg))
				fg, bg, first = cfg, cbg, false
			}
			run.WriteRune(r)
		}
		flush()
		sb.WriteString("[-:-]\n")
	}
	return sb.String()
}

// colorTag is the tview tag for a foreground and a background, "-" for one
// the screen leaves to the terminal.
func colorTag(fg, bg tcell.Color) string {
	hex := func(c tcell.Color) string {
		if !c.Valid() {
			return "-"
		}
		r, g, b := c.RGB()
		if r < 0 {
			return "-"
		}
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	return "[" + hex(fg) + ":" + hex(bg) + "]"
}
