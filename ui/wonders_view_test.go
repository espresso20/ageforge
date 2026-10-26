package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// galleryTown is a dashboard over a fresh engine and a state in which the
// wonders of the first n ages are built (n = 0: none), the one after them
// banked but not built.
func galleryTown(t *testing.T, built int) (*Dashboard, *tview.Pages, game.GameState) {
	t.Helper()
	eng := game.NewGameEngine()
	d, pages := newTestDashboard(eng)
	st := eng.GetState()
	list := getWonderList(st.Ruleset())
	if built > len(list) {
		t.Fatalf("staging: %d wonders asked of %d", built, len(list))
	}
	for i, w := range list {
		bs := st.Buildings[w.key]
		bs.Count, bs.Name = 0, w.name
		switch {
		case i < built:
			bs.Count, bs.Unlocked = 1, true
		case i == built:
			bs.Unlocked = true
		}
		st.Buildings[w.key] = bs
	}
	return d, pages, st
}

func wkey(k tcell.Key, r rune) *tcell.EventKey { return tcell.NewEventKey(k, r, 0) }

// TestGalleryListsBuiltWonders: the gallery is the wonders the town has
// built, oldest first, and no other, each with its age and its description;
// G or Tab swaps it with the list and the arrows stay the text view's.
func TestGalleryListsBuiltWonders(t *testing.T) {
	restoreForge(t)
	for _, built := range []int{0, 1, 3, 7} {
		d, _, st := galleryTown(t, built)
		list := getWonderList(st.Ruleset())
		got := builtWonders(st)
		if len(got) != built {
			t.Fatalf("%d built: the gallery lists %d", built, len(got))
		}
		for i, w := range got {
			if w.key != list[i].key {
				t.Errorf("%d built: place %d is %s, want %s", built, i, w.key, list[i].key)
			}
		}
		wv := d.wonders
		if wv.handleKey(wkey(tcell.KeyLeft, 0)) {
			t.Errorf("%d built: Left was taken by the panel", built)
		}
		if !wv.handleKey(wkey(tcell.KeyRune, 'g')) || !wv.gallery {
			t.Fatalf("%d built: G did not open the gallery", built)
		}
		if wv.handleKey(wkey(tcell.KeyRight, 0)) || wv.handleKey(wkey(tcell.KeyHome, 0)) {
			t.Errorf("%d built: the gallery took an arrow or Home; they scroll it", built)
		}
		text := untag(wv.provider(st, 100))
		wantAll(t, "the gallery", text, "Wonder gallery", "Tab or G the list")
		if built == 0 && strings.Contains(text, "★") {
			t.Errorf("an empty gallery lists a wonder:\n%s", text)
		}
		at := 0
		for _, w := range got {
			next := strings.Index(text[at:], "★ "+w.name)
			if next < 0 {
				t.Fatalf("%d built: the gallery does not list %s after place %d:\n%s", built, w.name, at, text)
			}
			at += next
			line, _, _ := strings.Cut(text[at:], "\n")
			if !strings.Contains(line, w.ageName) {
				t.Errorf("%d built: the line for %s does not name its age %s: %q", built, w.name, w.ageName, line)
			}
			if w.def.Description != "" && !strings.Contains(text[at:], w.def.Description) {
				t.Errorf("%d built: the gallery lacks the description of %s", built, w.name)
			}
			at++
		}
		if built < len(list) && strings.Contains(text, list[built].name) {
			t.Errorf("%d built: the gallery names %s, which is not built", built, list[built].name)
		}
		if strings.Contains(text, "—") {
			t.Errorf("%d built: an em dash in the gallery", built)
		}
	}
}

// TestGalleryKeysAndEsc: G or Tab open and close the gallery, Esc in it goes
// back to the list and the next Esc closes the panel; opening the panel
// again starts on the list.
func TestGalleryKeysAndEsc(t *testing.T) {
	restoreForge(t)
	d, _, st := galleryTown(t, 3)
	om := d.overlayMgr
	open := func() {
		if !om.Show("wonders", st) {
			t.Fatal("no wonders overlay")
		}
		om.Refresh(st)
	}
	open()
	tv := om.entries["wonders"].tv
	handler := tv.InputHandler()
	press := func(ev *tcell.EventKey) { handler(ev, func(tview.Primitive) {}) }
	list := untag(tv.GetText(false))
	if !strings.Contains(list, "Press G for a gallery") {
		t.Errorf("the list does not say how to open the gallery:\n%s", list)
	}
	// G opens it (the panel's key handler is the dashboard's, so the keys
	// are applied here)
	if !d.wonders.handleKey(wkey(tcell.KeyTab, 0)) || !d.wonders.gallery {
		t.Fatal("Tab did not open the gallery")
	}
	om.Refresh(st)
	if text := untag(tv.GetText(false)); !strings.Contains(text, "Wonder gallery") {
		t.Fatalf("the panel is not on the gallery:\n%s", text)
	}
	press(wkey(tcell.KeyEscape, 0))
	if om.ActiveName() != "wonders" || d.wonders.gallery {
		t.Fatalf("Esc in the gallery left active=%q gallery=%v: it goes back to the list", om.ActiveName(), d.wonders.gallery)
	}
	if text := untag(tv.GetText(false)); !strings.Contains(text, "Wonders: 3 /") || strings.Contains(text, "Wonder gallery") {
		t.Errorf("after Esc the panel is not the list:\n%s", text)
	}
	d.wonders.handleKey(wkey(tcell.KeyRune, 'G'))
	if !d.wonders.gallery {
		t.Errorf("G did not open the gallery again")
	}
	d.wonders.handleKey(wkey(tcell.KeyRune, 'g'))
	if d.wonders.gallery {
		t.Errorf("g did not close it")
	}
	d.wonders.handleKey(wkey(tcell.KeyRune, 'g'))
	press(wkey(tcell.KeyEscape, 0))
	press(wkey(tcell.KeyEscape, 0))
	if om.ActiveName() != "" {
		t.Errorf("the second Esc did not close the panel: %q", om.ActiveName())
	}
	d.wonders.handleKey(wkey(tcell.KeyRune, 'g'))
	open()
	if d.wonders.gallery {
		t.Errorf("the panel opened on the gallery")
	}
}

// TestGalleryListsEveryWonder: every wonder of the game, built, is in the
// gallery, in the game's order, with its name and its age.
func TestGalleryListsEveryWonder(t *testing.T) {
	restoreForge(t)
	d, _, st := galleryTown(t, 22)
	list := builtWonders(st)
	if len(list) != 22 {
		t.Fatalf("staging: %d wonders built", len(list))
	}
	d.wonders.gallery = true
	text := untag(d.wonders.provider(st, 100))
	at := 0
	for _, w := range list {
		next := strings.Index(text[at:], "★ "+w.name)
		if next < 0 {
			t.Fatalf("the gallery does not list %s after place %d:\n%s", w.key, at, text)
		}
		at += next + 1
	}
	if got := strings.Count(text, "★"); got != len(list) {
		t.Errorf("the gallery has %d stars for %d wonders", got, len(list))
	}
}
