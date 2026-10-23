package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/game"
)

// expedition_field.go draws the Expedition Report's right side in the split
// layout (reportLayoutFor): the Out list, one line a party, and above the
// Came home box the field report for the party the list highlights. The
// compact layout keeps the card it always had (drawOutCard).

// roadTone is the colour of a road event's story: a good turn green, a bad
// one red, the rest plain.
func roadTone(kind string) xk {
	switch kind {
	case "helpers", "find", "fair":
		return xGood
	case "storm", "toll", "foes", "ambush":
		return xBad
	}
	return xText
}

// outListH is the height of the Out list: a line a party and its box, at
// least the empty box, and at most two fifths of the column.
func outListH(missions, h int) int {
	if missions == 0 {
		return 3
	}
	return min(missions+2, max(3, h*2/5))
}

// drawFieldReport draws the Out list and the field report beside the map.
func drawFieldReport(g *xgrid, state game.GameState, v expView, l expLayout, sc *expScene, hi int) {
	x, w := l.cardX, l.cardW
	cw := w - 2
	listH := outListH(len(sc.acts), l.mainH)
	drawOutList(g, state, x, l.mainY, w, listH, sc, hi)

	y, h := l.mainY+listH, l.mainH-listH
	g.box(x, y, w, h, xBorder, "Field report", xTitle, false)
	// With no legend row to hold it, the Society's line takes the box's foot.
	soc, socKind := societyLine(state)
	foot := 0
	if l.legendY < 0 {
		foot = 2
		for i, ln := range limitLines(wrapText(soc, cw), foot, cw) {
			g.putFit(x+1, y+h-1-foot+i, cw, ln, socKind)
		}
	}
	top, bottom := y+1, y+h-1-foot
	if len(sc.acts) == 0 {
		offers := len(sc.offers)
		for i, ln := range wrapText(fmt.Sprintf("Nobody is out. The Ready Room has %d %s waiting.", offers, pluralize("offer", offers)), cw) {
			if top+i < bottom {
				g.putFit(x+1, top+i, cw, ln, xDim)
			}
		}
		return
	}
	for i, ln := range fieldLines(state, sc.acts[hi].act, cw) {
		if top+i >= bottom {
			break
		}
		g.segsFit(x+1, top+i, cw, ln...)
	}
}

// drawOutList draws the Out list: each party under way on a line, its place
// and its progress and time left, with the highlighted one marked. It scrolls
// to keep the highlighted party in view.
func drawOutList(g *xgrid, state game.GameState, x, y, w, h int, sc *expScene, hi int) {
	cw := w - 2
	n := len(sc.acts)
	g.box(x, y, w, h, xBorder, fmt.Sprintf("Out · %d", n), xTitle, false)
	visible := h - 2
	if n == 0 || visible <= 0 {
		return
	}
	first := clamp(hi-visible+1, 0, max(0, n-visible))
	for i := first; i < n && i < first+visible; i++ {
		a := sc.acts[i].act
		row := y + 1 + (i - first)
		head, kind := "  ", xBright
		if i == hi {
			head, kind = "▸ ", xSel
		}
		progress := 0.0
		if a.TotalTicks > 0 {
			progress = 1 - float64(a.TicksLeft)/float64(a.TotalTicks)
		}
		bar := expBar(progress, 8, "▰", "▱")
		left := " " + expDur(a.TicksLeft, state) + " left"
		room := max(1, cw-1-runeLen(bar)-1-runeLen(left))
		g.put(x+1, row, fmt.Sprintf("%-*s", room, fitText(head+goingTo(a.Label, a.Place), room)), kind)
		g.segsFit(x+1+room+1, row, cw-room-1, xseg{" " + bar, xGood}, xseg{left, xText})
	}
}

// fieldLines is the field report for one party, wrapped to cw cells: its
// title and tier, its soldiers, odds, progress and expected haul, then every
// road event that has fired, newest first.
func fieldLines(state game.GameState, a *game.ExpeditionSnapshot, cw int) []xline {
	var out []xline
	add := func(parts ...xseg) { out = append(out, xline(parts)) }

	for _, ln := range wrapText(briefTitleOf(a.Name), cw) {
		add(xseg{ln, xBright})
	}
	tier := tierLabel(game.ExpeditionInfo{Label: a.Label, Tier: a.Tier, Grand: a.Grand})
	if a.Rare {
		add(xseg{tier, xAccent}, xseg{"  ◈ Legendary", xLegend})
	} else {
		add(xseg{tier, xAccent})
	}
	add()

	if a.Party > 0 {
		parts := []xseg{{"Party  ", xDim}, {fmt.Sprintf("%s %s", FormatNumber(float64(a.Party)), pluralize("soldier", a.Party)), xText}}
		if fell := a.Party - a.Soldiers; fell > 0 {
			parts = append(parts, xseg{fmt.Sprintf(" · %s fell so far", FormatNumber(float64(fell))), xBad})
		}
		add(parts...)
	}

	odds := []xseg{{"Odds   ", xDim}, {expBar(a.SuccessChance, 10, "█", "░"), oddsKind(a.SuccessChance)},
		{fmt.Sprintf(" %.0f%%", a.SuccessChance*100), oddsKind(a.SuccessChance)}}
	if a.LaunchChance > 0 && fmt.Sprintf("%.0f", a.LaunchChance*100) != fmt.Sprintf("%.0f", a.SuccessChance*100) {
		odds = append(odds, xseg{fmt.Sprintf(" (was %.0f%%)", a.LaunchChance*100), xDim})
	}
	add(odds...)

	progress := 0.0
	if a.TotalTicks > 0 {
		progress = 1 - float64(a.TicksLeft)/float64(a.TotalTicks)
	}
	add(xseg{"Going  ", xDim}, xseg{expBar(progress, 10, "▰", "▱"), xGood},
		xseg{fmt.Sprintf(" %s out · %s to go", expDur(a.TotalTicks-a.TicksLeft, state), expDur(a.TicksLeft, state)), xText})

	add(xseg{"Haul   ", xDim}, xseg{fmt.Sprintf("about %.0f min of income", a.HaulMinutes), xGood})
	var pays []string
	for _, k := range sortedKeysOf(a.SuccessHaul) {
		pays = append(pays, "+"+FormatNumber(a.SuccessHaul[k])+" "+resourceDisplayName(state, k))
	}
	for _, ln := range wrapText(strings.Join(pays, ", "), cw-7) {
		add(xseg{"       ", xDim}, xseg{ln, xGood})
	}
	if s := a.HaulScale; s > 0 && math.Abs(s-1) > 1e-9 {
		factor := strconv.FormatFloat(math.Round(s*100)/100, 'f', -1, 64)
		add(xseg{"       ", xDim}, xseg{"x" + factor + " from the road", xAccent})
	}
	add()

	add(xseg{"On the road", xAccent})
	if len(a.RoadStories) == 0 {
		add(xseg{"Nothing has happened yet.", xFaint})
	}
	for j := len(a.RoadStories) - 1; j >= 0; j-- {
		kind := xText
		if j < len(a.RoadKinds) {
			kind = roadTone(a.RoadKinds[j])
		}
		for i, ln := range wrapText(a.RoadStories[j], cw-2) {
			lead := "  "
			if i == 0 {
				lead = "· "
			}
			add(xseg{lead, kind}, xseg{ln, kind})
		}
	}
	return out
}
