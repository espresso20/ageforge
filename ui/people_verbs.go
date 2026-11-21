package ui

import (
	"fmt"
	"image"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
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
	verbRoster   = "roster"

	// goKeyBase is the first of the private runes the links to a tab press.
	goKeyBase = 0xE100
	// splitStep is how many percent a click on the split's chips moves it.
	splitStep = 10
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
	txtDismiss = rowid.E("ui.people_verbs.dismiss.01", "Dismiss")
	txtRecruit = rowid.E("ui.people_verbs.recruit.01", "Recruit")
	txtSplit   = rowid.E("ui.people_verbs.split.01", "Split")
	txtAuto    = rowid.E("ui.people_verbs.auto.01", "Auto")
)

// dismissAsk is the question before a worker is dismissed from a building.
func dismissAsk(name string) string {
	return rowid.Sprintf("ui.people_verbs.dismissAsk.01", "Dismiss a worker from %s?\n\nThey leave your population.", name)
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

// workList draws the picked area's buildings from y to end, one row each
// whichever is picked: the name, the stepper and the workers. The picked row
// also holds Dismiss, left of the stepper, and its name gives up the room. It
// returns the next y.
func (c peopleBody) workList(rx, y, ri, end int, bs []peopleBuilding) int {
	g := c.g
	sub := clamp(c.v.sub, 0, len(bs)-1)
	// The steppers share a column: the widest count sets it.
	cw := 0
	for _, b := range bs {
		cw = max(cw, runeLen(fmt.Sprintf("%s/%s", num(b.WorkersAssigned), num(b.WorkerCapacity*b.Count))))
	}
	for i, b := range bs {
		if y+1 > end {
			g.putFit(rx, y, ri, rowid.Sprintf("ui.people_verbs.workList.01", "and %d more", len(bs)-i), xFaint)
			return y + 1
		}
		count := fmt.Sprintf("%s/%s", num(b.WorkersAssigned), num(b.WorkerCapacity*b.Count))
		minus, plus := buildingStep(b.key, b.BuildingState, c.st.Workers.TotalIdle)
		sx := rx + ri - cw - 1 - stepperW
		nameEnd := sx
		if i == sub {
			word := txtDismiss.String()
			nameEnd = sx - 2 - (runeLen(word) + 2)
			g.chip(nameEnd+1, y, chipSpec{kind: keyAlert, label: word, cmd: command(verbDismiss, b.key), ask: dismissAsk(b.Name), yes: word, off: b.WorkersAssigned == 0})
		}
		pickName(g, rx, y, nameEnd-rx-1, fmt.Sprintf("%s x%d", b.Name, b.Count), i == sub)
		g.hit(hitSub, i, rx, y, nameEnd-rx, 1)
		g.stepper(sx, y, minus, plus)
		g.rput(rx+ri-1, y, count, xGood)
		y++
	}
	return y
}

// recruitBlock draws Recruit and its count choice at (x, y): a button with the
// choices beside it. Recruit is off when the homes have no room for the count.
func (c peopleBody) recruitBlock(x, y, w int) {
	g := c.g
	free := max(0, c.st.Workers.MaxPop-c.st.Workers.TotalPop)
	need := 1
	if c.buy == buyTen {
		need = 10
	}
	cmd := strings.TrimSpace(command(verbRecruit, buyCounts[c.buy]))
	off := free < need
	label := txtRecruit.String()
	cx, cy := g.chip(x, y, chipSpec{kind: keyAccent, label: label, cmd: cmd, off: off})+1, y
	for ch := buyChoice(0); ch < buyChoices; ch++ {
		word := " " + buyWords[ch] + " "
		r := image.Rect(cx, cy, cx+runeLen(word), cy+1)
		g.put(cx, cy, word, xKeycap)
		g.keyset().keys = append(g.keyset().keys, xkey{cmd: "buy " + buyArgs[ch], r: r, cap: r, kind: keySteel, picked: ch == c.buy})
		cx = r.Max.X + 1
	}
	if room := x + w - cx; room > 8 {
		g.putFit(cx+1, cy, room-1, rowid.Sprintf("ui.people_verbs.recruitBlock.01", "%d homes empty", free), xDim)
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
	g.putFit(nx, y, x+w-nx, rowid.Sprintf("ui.people_verbs.splitRow.01", "%d%% research", s.Share), xText)
}

// shareRow draws the picked area's Share line at (x, y): the stepper that
// moves the share ten points at a time, what it is now, and the Auto chip,
// accent while the area is on auto. From auto a step starts at the present
// share rounded to ten.
func (c peopleBody) shareRow(x, y, w int, r game.ShareRow) {
	g := c.g
	g.put(x, y, rowid.Line("ui.people_verbs.shareRow.01", "Share  "), xDim)
	base := int(r.Percent + 0.5)
	if !r.Set {
		base = (base + 5) / 10 * 10
	}
	step := func(to int) stepSide {
		return stepSide{command(verbRoster, r.Domain, fmt.Sprint(clamp(to, 0, 100))), false}
	}
	minus, plus := step(base-splitStep), step(base+splitStep)
	minus.off = base <= 0
	plus.off = base >= 100
	nx := g.stepper(x+8, y, minus, plus) + 1
	kind := keySteel
	if !r.Set {
		kind = keyAccent
	}
	nx = g.chip(nx, y, chipSpec{kind: kind, label: txtAuto.String(), cmd: command(verbRoster, r.Domain, "auto")}) + 1
	text := rowid.Line("ui.people_verbs.shareRow.02", "auto, by its buildings' slots")
	if r.Set {
		text = rowid.Line("ui.people_verbs.shareRow.03", "set to ") + game.SharePercent(float64(base))
	}
	g.putFit(nx, y, x+w-nx, text, xText)
}

// peopleLink draws text at (x, y) as a link to tab: the same words, in cells
// padded to a chip, and a click on it goes to the tab.
func (g *xgrid) peopleLink(x, y, w int, text string, tab int) {
	word := " " + fitText(text, w-2) + " "
	r := image.Rect(x, y, x+runeLen(word), y+1)
	g.put(x, y, word, xKeycap)
	g.keyset().keys = append(g.keyset().keys, xkey{key: goKey(tab), r: r, cap: r, kind: keySteel})
}
