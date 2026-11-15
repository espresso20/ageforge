package ui

import (
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

var listRows = []string{"[gold]● Ashford[-]  [gray]a1b2[-]", "Birch Hollow", "[dim]── Group ──[-]", "Cinder Reach", "Dunmere", "Eastwatch"}

func TestNativeListKeys(t *testing.T) {
	keepTheme(t)
	l := newNativeList().ShowSecondaryText(false)
	var moved, chosen []int
	l.SetChangedFunc(func(i int, _, _ string, _ rune) { moved = append(moved, i) })
	l.SetSelectedFunc(func(i int, _, _ string, _ rune) { chosen = append(chosen, i) })
	for i, r := range []rune{'a', 'b', 0, 'd', 0} {
		l.AddItem("row", "", r, nil)
		_ = i
	}
	if len(moved) != 1 || moved[0] != 0 {
		t.Errorf("the first row added announces itself: %v", moved)
	}
	drawPrimitive(t, l, 20, 3)
	press := func(k kit.KeyCode, r rune) {
		keyTo(l, tk(k, r, 0))
	}
	at := func(want int, what string) {
		t.Helper()
		if got := l.GetCurrentItem(); got != want {
			t.Errorf("%s: row %d, want %d", what, got, want)
		}
	}
	press(kit.KeyDown, 0)
	at(1, "Down")
	press(kit.KeyEnd, 0)
	at(4, "End")
	press(kit.KeyDown, 0)
	at(0, "Down at the end comes round")
	press(kit.KeyUp, 0)
	at(4, "Up at the top comes round")
	press(kit.KeyHome, 0)
	at(0, "Home")
	press(kit.KeyPgDn, 0)
	at(3, "PgDn is a screenful (3 rows)")
	press(kit.KeyPgUp, 0)
	at(0, "PgUp")
	press(kit.KeyTab, 0)
	at(1, "Tab")
	press(kit.KeyEnter, 0)
	if len(chosen) != 1 || chosen[0] != 1 {
		t.Errorf("Enter chooses the row: %v", chosen)
	}
	press(kit.KeyRune, 'd')
	at(3, "a shortcut selects")
	if len(chosen) != 2 || chosen[1] != 3 {
		t.Errorf("a shortcut chooses its row: %v", chosen)
	}
	press(kit.KeyRune, 'z')
	if len(chosen) != 2 {
		t.Error("a letter that is no shortcut does nothing")
	}
	press(kit.KeyRune, ' ')
	if len(chosen) != 3 || chosen[2] != 3 {
		t.Errorf("Space chooses the selected row: %v", chosen)
	}
	l.SetCurrentItem(-1)
	at(4, "SetCurrentItem counts back from the end")
	l.Clear()
	if l.GetItemCount() != 0 || l.GetCurrentItem() != 0 {
		t.Error("Clear empties the list")
	}
}

func TestNativeListMouse(t *testing.T) {
	keepTheme(t)
	l := newNativeList().ShowSecondaryText(false)
	for _, r := range listRows {
		l.AddItem(r, "", 0, nil)
	}
	var moved, chosen []int
	l.SetChangedFunc(func(i int, _, _ string, _ rune) { moved = append(moved, i) })
	l.SetSelectedFunc(func(i int, _, _ string, _ rune) { chosen = append(chosen, i) })
	newStage(l, 30, 6).click(3, 3)
	if l.GetCurrentItem() != 3 || len(moved) != 1 || moved[0] != 3 {
		t.Errorf("a click selects its row: at %d, moved %v", l.GetCurrentItem(), moved)
	}
	if len(chosen) != 1 || chosen[0] != 3 {
		t.Errorf("one click chooses: %v", chosen)
	}
}
