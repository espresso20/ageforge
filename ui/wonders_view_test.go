package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// galleryTown is a dashboard over a fresh engine and a state in which the
// wonders of the first n ages are built (n = 0: none), the one after them
// banked but not built.
func galleryTown(t *testing.T, built int) (*Dashboard, *tview.Pages, game.GameState) {
	t.Helper()
	eng := game.NewGameEngine()
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), eng, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
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

// TestGalleryListsBuiltWonders: the gallery is the wonders the town has built,
// oldest first, and no other: opening it puts it on the newest, the arrows
// step through them and stop at the ends, Home and End go to the ends.
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
		if wv.handleKey(wkey(tcell.KeyLeft, 0), st) {
			t.Errorf("%d built: Left was taken by the list", built)
		}
		if !wv.handleKey(wkey(tcell.KeyRune, 'g'), st) || !wv.gallery {
			t.Fatalf("%d built: G did not open the gallery", built)
		}
		text := untag(wv.provider(st, 100))
		if built == 0 {
			if !strings.Contains(text, "No wonder built yet") {
				t.Errorf("an empty gallery says nothing about it:\n%s", text)
			}
			continue
		}
		if want := got[built-1]; !strings.Contains(text, want.name) || wv.key != want.key {
			t.Errorf("%d built: the gallery opens on %s, want the newest, %s:\n%s", built, wv.key, want.key, text)
		}
		if unbuilt := list[built]; strings.Contains(text, unbuilt.name) {
			t.Errorf("%d built: the gallery names %s, which is not built", built, unbuilt.name)
		}
		// to the first, and no further
		for i := 0; i < built+3; i++ {
			wv.handleKey(wkey(tcell.KeyLeft, 0), st)
		}
		if wv.idx != 0 || wv.key != got[0].key {
			t.Errorf("%d built: Left ended on %d (%s)", built, wv.idx, wv.key)
		}
		// and through every one to the last, naming each in turn
		seen := []string{wv.key}
		for i := 0; i < built+3; i++ {
			wv.handleKey(wkey(tcell.KeyRight, 0), st)
			if seen[len(seen)-1] != wv.key {
				seen = append(seen, wv.key)
			}
		}
		if len(seen) != built {
			t.Errorf("%d built: Right visited %d of them: %v", built, len(seen), seen)
		}
		for i, k := range seen {
			if k != got[i].key {
				t.Errorf("%d built: stop %d is %s, want %s", built, i, k, got[i].key)
			}
			wv.key = k
			if text := untag(wv.provider(st, 100)); !strings.Contains(text, got[i].name) || !strings.Contains(text, got[i].ageName) {
				t.Errorf("%d built: stop %d does not name %s of the %s:\n%s", built, i, got[i].name, got[i].ageName, text)
			}
		}
		wv.handleKey(wkey(tcell.KeyHome, 0), st)
		if wv.idx != 0 {
			t.Errorf("%d built: Home went to %d", built, wv.idx)
		}
		wv.handleKey(wkey(tcell.KeyEnd, 0), st)
		if wv.idx != built-1 {
			t.Errorf("%d built: End went to %d", built, wv.idx)
		}
		// a wonder built while it is open does not move the picture
		keep := wv.key
		if built < len(list) {
			bs := st.Buildings[list[built].key]
			bs.Count = 1
			st.Buildings[list[built].key] = bs
			wv.handleKey(wkey(tcell.KeyHome, 0), st)
			wv.sync(builtWonders(st))
			if wv.key != got[0].key || keep == "" {
				t.Errorf("%d built: the gallery moved off %s when another wonder was built: %s", built, got[0].key, wv.key)
			}
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
	// G opens it (the panel's key handler is the dashboard's: it asks the
	// engine for the state, so the keys are applied here)
	if !d.wonders.handleKey(wkey(tcell.KeyTab, 0), st) || !d.wonders.gallery {
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
	d.wonders.handleKey(wkey(tcell.KeyRune, 'G'), st)
	if !d.wonders.gallery {
		t.Errorf("G did not open the gallery again")
	}
	d.wonders.handleKey(wkey(tcell.KeyRune, 'g'), st)
	if d.wonders.gallery {
		t.Errorf("g did not close it")
	}
	d.wonders.handleKey(wkey(tcell.KeyRune, 'g'), st)
	press(wkey(tcell.KeyEscape, 0))
	press(wkey(tcell.KeyEscape, 0))
	if om.ActiveName() != "" {
		t.Errorf("the second Esc did not close the panel: %q", om.ActiveName())
	}
	d.wonders.handleKey(wkey(tcell.KeyRune, 'g'), st)
	open()
	if d.wonders.gallery {
		t.Errorf("the panel opened on the gallery")
	}
}

// galleryArt is the rows of a drawn page between the gallery's header and its
// caption: the picture.
func galleryArt(t *testing.T, screen, name string) []string {
	t.Helper()
	rows := strings.Split(screen, "\n")
	top, cap := -1, -1
	for i, r := range rows {
		switch {
		case strings.Contains(r, "Wonder gallery"):
			top = i
		case strings.Contains(r, "★ "+name):
			cap = i
		}
	}
	if top < 0 || cap < 0 || cap <= top+1 {
		t.Fatalf("no gallery with %q on the page (header row %d, caption row %d):\n%s", name, top, cap, screen)
	}
	art := rows[top+1 : cap]
	for i, r := range art { // between the box's own borders (the page shows through round it)
		cs := []rune(r)
		a, b := -1, -1
		for j, c := range cs {
			if c == '║' {
				if a < 0 {
					a = j
				}
				b = j
			}
		}
		if a < 0 || b <= a {
			t.Fatalf("no box round the picture's row %d: %q", i, r)
		}
		art[i] = string(cs[a+1 : b])
	}
	return art
}

// TestGalleryDrawsEveryWonder: every wonder of the game, built, is drawn in
// the gallery at 80x24 and larger, in every glyph tier and in several
// themes, with its name, its age and the keys under the picture, a picture
// with more than a handful of cells in it, and, in the plain tier, nothing
// outside ASCII in it.
func TestGalleryDrawsEveryWonder(t *testing.T) {
	restoreForge(t)
	d, pages, st := galleryTown(t, 22)
	list := builtWonders(st)
	if len(list) != 22 {
		t.Fatalf("staging: %d wonders built", len(list))
	}
	themes := theme.All()
	om := d.overlayMgr
	for _, sz := range [][2]int{{80, 24}, {100, 30}, {160, 45}} {
		w, h := sz[0], sz[1]
		om.screenW, om.screenH = w, h
		for ti, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
			d.wonders.tier = func() mapmodel.GlyphTier { return tier }
			for i, win := range list {
				must(t, theme.SetActive(themes[(i+ti)%len(themes)].Key))
				d.wonders.open()
				om.Show("wonders", st)
				d.wonders.gallery, d.wonders.idx, d.wonders.key = true, i, win.key
				om.Refresh(st)
				screen := renderText(t, pages, w, h)
				art := galleryArt(t, screen, win.name)
				if len(art) < 6 {
					t.Errorf("%s at %dx%d: the picture is %d rows", win.key, w, h, len(art))
				}
				ink := 0
				for _, r := range art {
					for _, c := range r {
						if c != ' ' {
							ink++
						}
						if tier == mapmodel.TierASCII && c >= 0x80 {
							t.Fatalf("%s at %dx%d plain: %q in the picture", win.key, w, h, c)
						}
						if c == '?' && tier != mapmodel.TierASCII {
							t.Fatalf("%s at %dx%d %v: a '?' fallback in the picture", win.key, w, h, tier)
						}
					}
				}
				// a plain backdrop is mostly colour, not glyphs: the wonder's own cells carry some
				if ink < 8 {
					t.Errorf("%s at %dx%d %v: the picture has %d glyphs", win.key, w, h, tier, ink)
				}
				wantAll(t, fmt.Sprintf("%s at %dx%d: the page", win.key, w, h), screen, win.name, win.ageName, "Esc back")
				if strings.Contains(screen, "—") {
					t.Errorf("%s: an em dash on the page", win.key)
				}
			}
		}
	}
}

// TestGalleryPictureFollowsTheTown: the picture is the skyline's own wonder
// art: the colour tags of one wonder's picture are not another's, and the
// same wonder twice is the same picture.
func TestGalleryPictureFollowsTheTown(t *testing.T) {
	restoreForge(t)
	d, _, st := galleryTown(t, 5)
	wv := d.wonders
	list := builtWonders(st)
	a := wv.picture(st, list[0].key, 70, 12)
	b := wv.picture(st, list[1].key, 70, 12)
	if a == b || strings.Count(a, "\n") != 12 {
		t.Errorf("two wonders draw alike, or a picture is not 12 rows: %d lines", strings.Count(a, "\n"))
	}
	if again := wv.picture(st, list[0].key, 70, 12); again != a {
		t.Errorf("the same wonder drew differently twice")
	}
	if !strings.Contains(a, "[#") {
		t.Errorf("the picture carries no colour: %q", a[:min(len(a), 80)])
	}
	if c := wv.picture(st, "no_such_wonder", 70, 12); strings.Contains(c, "[#") {
		t.Errorf("a wonder that does not exist drew something")
	}
}
