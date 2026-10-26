package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// wonders_view.go is the Wonders panel's second view, the gallery: a plain
// list of every wonder the town has built, oldest first, each with its age
// and its description. G or Tab opens it from the list and closes it again;
// the arrows and Page keys scroll it as they do the list; Esc goes back to
// the list, and the next Esc closes the panel. A wonder not built yet is not
// in it.

// wondersView is the panel's state: which view it is on.
type wondersView struct {
	gallery bool
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
func (wv *wondersView) open() { wv.gallery = false }

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
// took it: G or Tab swaps the list and the gallery, and the rest (scrolling)
// is the text view's.
func (wv *wondersView) handleKey(ev *tcell.EventKey) bool {
	toggle := ev.Key() == tcell.KeyTab || ev.Key() == tcell.KeyRune && (ev.Rune() == 'g' || ev.Rune() == 'G')
	if toggle {
		wv.gallery = !wv.gallery
	}
	return toggle
}

// footer is the keys on the panel's foot: the list's, or the gallery's.
func (wv *wondersView) footer() [][2]string {
	tab, esc := "gallery", "close"
	if wv.gallery {
		tab, esc = "list", "back"
	}
	return [][2]string{{"↑↓", "scroll"}, {"PgUp PgDn", vibe.Line("ui.wonders_view.footer.09", "page")}, {"Tab", tab}, {"Esc", esc}}
}

// galleryText is the gallery: each built wonder, oldest first, with the age
// it belongs to and its description.
func (wv *wondersView) galleryText(state game.GameState) string {
	var sb strings.Builder
	sb.WriteString(vibe.Line("ui.wonders_view.galleryText.01", "[gold]═══ Wonder gallery ═══[-]"))
	sb.WriteString("\n\n")
	set := state.Ruleset()
	for _, w := range builtWonders(state) {
		when := ""
		if yr, ok := set.AgeYear(w.ageKey); ok {
			label := set.AgeYearLabel(w.ageKey)
			if label == "" {
				label = config.YearLabel(yr)
			}
			when = vibe.Sprintf("ui.wonders_view.galleryText.04", "  [gray]the age began %s[-]", label)
		}
		fmt.Fprintf(&sb, "[gold]★ %s[-]  [gray]%s[-]%s\n", lit(w.name), lit(w.ageName), when)
		if w.def.Description != "" {
			fmt.Fprintf(&sb, "[gray]%s[-]\n", lit(w.def.Description))
		}
		sb.WriteString("\n")
	}
	sb.WriteString(vibe.Line("ui.wonders_view.galleryKeys.01", "[gray]Tab or G the list   Esc back[-]"))
	sb.WriteString("\n")
	return sb.String()
}
