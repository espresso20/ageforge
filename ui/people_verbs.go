package ui

import (
	"fmt"
	"image"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/window/kit"
)

// people_verbs.go is what the People panel does by click: the stepper on each
// building of a work area, Dismiss on the picked one, Recruit with its count,
// the scholars' split and the cards' links to other tabs. Each is an xkey that
// carries the command line it stands for.

const (
	verbRecruit  = "recruit"
	verbDismiss  = "dismiss"
	verbScholars = "scholars"

	// goKeyBase is the first of the private runes the links to a tab press.
	goKeyBase = 0xE100
	// splitStep is how many percent a click on the split's chips moves it.
	splitStep = 10
	// recruitTall is the panel height from which Recruit is a three-row button.
	recruitTall = 26
)

// goKey is the key a click on a link to tab presses.
func goKey(tab int) kit.Key { return kit.Key{Code: kit.KeyRune, Rune: rune(goKeyBase + tab)} }

// goOf is the tab a key links to.
func goOf(k kit.Key) (int, bool) {
	t := int(k.Rune) - goKeyBase
	return t, k.Code == kit.KeyRune && t >= 0 && t < len(peopleTabs)
}

// The words the buttons say.
var (
	txtDismiss = vibe.E("ui.people_verbs.dismiss.01", "Dismiss")
	txtRecruit = vibe.E("ui.people_verbs.recruit.01", "Recruit")
	txtSplit   = vibe.E("ui.people_verbs.split.01", "Split")
)

// dismissAsk is the question before a worker is dismissed from a building.
func dismissAsk(name string) string {
	return vibe.Sprintf("ui.people_verbs.dismissAsk.01", "Dismiss a worker from %s?\n\nThey leave your population.", name)
}

// chip draws a one-row key of kind at (x, y) holding word, records that a click
// runs cmd (after ask, when it has one) and returns the next x.
func (g *xgrid) chip(x, y int, word string, kind keyKind, cmd, ask, yes string, off bool) int {
	word = " " + word + " "
	r := image.Rect(x, y, x+runeLen(word), y+1)
	g.put(x, y, word, xKeycap)
	set := g.keyset()
	set.keys = append(set.keys, xkey{r: r, cap: r, kind: kind, cmd: cmd, ask: ask, yes: yes, off: off})
	return r.Max.X
}

// pickedBuilding is the building the Work tab has picked, if there is one: the
// picked area's building at v.sub.
func pickedBuilding(st game.GameState, v peopleView) (peopleBuilding, bool) {
	rows := game.ShareRows(st)
	if len(rows) == 0 {
		return peopleBuilding{}, false
	}
	bs := peopleBody{st: st}.areaList(rows[clamp(v.row, 0, len(rows)-1)].Domain)
	if len(bs) == 0 {
		return peopleBuilding{}, false
	}
	return bs[clamp(v.sub, 0, len(bs)-1)], true
}

// workList draws the picked area's buildings from y to end: each with its
// stepper and its workers, the picked one marked and with Dismiss under it. It
// returns the next y.
func (c peopleBody) workList(rx, y, ri, end int, bs []peopleBuilding) int {
	g := c.g
	sub := clamp(c.v.sub, 0, len(bs)-1)
	for i, b := range bs {
		need := 1
		if i == sub {
			need = 2
		}
		if y+need > end {
			g.putFit(rx, y, ri, vibe.Sprintf("ui.people_tabs.work.09", "and %d more", len(bs)-i), xFaint)
			return y + 1
		}
		count := fmt.Sprintf("%s/%s", num(b.WorkersAssigned), num(b.WorkerCapacity*b.Count))
		minus, plus := buildingStep(b.key, b.BuildingState, c.st.Workers.TotalIdle)
		sx := rx + ri - runeLen(count) - 1 - stepperW
		pickName(g, rx, y, sx-rx-1, fmt.Sprintf("%s x%d", b.Name, b.Count), i == sub)
		g.hit(hitSub, i, rx, y, sx-rx, 1)
		g.stepper(sx, y, minus, plus)
		g.rput(rx+ri-1, y, count, xGood)
		y++
		if i == sub {
			g.chip(rx+2, y, txtDismiss.String(), keyDanger, command(verbDismiss, b.key), dismissAsk(b.Name), txtDismiss.String(), b.WorkersAssigned == 0)
			y++
		}
	}
	return y
}

// recruitRows is the rows the Recruit block takes.
func (c peopleBody) recruitRows() int {
	if c.h >= recruitTall {
		return actionRows
	}
	return 1
}

// recruitBlock draws Recruit and its count choice at (x, y): a primary button
// with the choices beside it when there is room, else a chip row. Recruit is
// off when the homes have no room for the count.
func (c peopleBody) recruitBlock(x, y, w int) {
	g := c.g
	free := max(0, c.st.Workers.MaxPop-c.st.Workers.TotalPop)
	need := 1
	if c.v.buy == buyTen {
		need = 10
	}
	cmd := command(verbRecruit, buyCounts[c.v.buy])
	off := free < need
	label := txtRecruit.String()
	cx, cy := x, y
	if c.recruitRows() == actionRows {
		cx = g.actionButton(x, y, actionSpec{kind: keyPrimary, label: label, cmd: cmd, off: off}) + 1
		cy = y + 1
	} else {
		cx = g.chip(x, y, label, keyAccent, cmd, "", "", off) + 1
	}
	for ch := buyChoice(0); ch < buyChoices; ch++ {
		word := " " + buyWords[ch] + " "
		r := image.Rect(cx, cy, cx+runeLen(word), cy+1)
		g.put(cx, cy, word, xKeycap)
		kind := keySteel
		if ch == c.v.buy {
			kind = keyAccent
		}
		g.keyset().keys = append(g.keyset().keys, xkey{key: buyKey(ch), r: r, cap: r, kind: kind})
		cx = r.Max.X + 1
	}
	if room := x + w - cx; room > 8 {
		g.putFit(cx+1, cy, room-1, vibe.Sprintf("ui.people_verbs.recruitBlock.01", "%d homes empty", free), xDim)
	}
}

// splitRow draws the scholars' split at (x, y): the stepper that moves the
// research share ten percent at a time, and what it is now.
func (c peopleBody) splitRow(x, y, w int) {
	g, s := c.g, c.st.Scholars
	g.put(x, y, fitText(txtSplit.String(), 9), xDim)
	step := func(to int) stepSide {
		return stepSide{command(verbScholars, fmt.Sprint(clamp(to, 0, 100))), s.Total <= 0}
	}
	minus, plus := step(s.Share-splitStep), step(s.Share+splitStep)
	minus.off = minus.off || s.Share <= 0
	plus.off = plus.off || s.Share >= 100
	nx := g.stepper(x+10, y, minus, plus) + 1
	g.putFit(nx, y, x+w-nx, vibe.Sprintf("ui.people_verbs.splitRow.01", "%d%% research", s.Share), xText)
}

// peopleLink draws text at (x, y) as a link to tab: the same words, in cells
// padded to a chip, and a click on it goes to the tab.
func (g *xgrid) peopleLink(x, y, w int, text string, tab int) {
	word := " " + fitText(text, w-2) + " "
	r := image.Rect(x, y, x+runeLen(word), y+1)
	g.put(x, y, word, xKeycap)
	g.keyset().keys = append(g.keyset().keys, xkey{key: goKey(tab), r: r, cap: r, kind: keySteel})
}
