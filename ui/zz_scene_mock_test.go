package ui

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// SCRATCH: design drafts of three panels as scenes. Not for commit as is.

func sceneFrame(title string, w, h int) *xgrid {
	g := newXGrid(w, h)
	g.box(0, 0, w, h, xBorder, title, xTitle, false)
	g.closeButton(w)
	return g
}

// ===== Trade: the road out of town =====

type mockRoute struct {
	name, out, in, note, cmd, act string
	state                         int // 0 running, 1 suspended, 2 ready, 3 locked
	prog                          float64
}

func sceneTrade(w, h, sel int) *xgrid {
	g := sceneFrame("Trade · Medieval Age", w, h)
	routes := []mockRoute{
		{"Local Barter", "10 food", "8 wood", "1m 12s left · 14 trips", "trade route stop local_barter", "stop", 0, 0.30},
		{"Iron Road", "40 stone", "12 iron", "iron is blockaded", "trade route stop iron_road", "stop", 1, 0.72},
		{"Stone Trade", "15 wood", "12 stone", "", "trade route start stone_trade", "start", 2, 0},
		{"Gold Caravan", "50 stone", "5 gold", "", "trade route start gold_caravan", "start", 2, 0},
		{"Silk Road", "30 gold", "80 culture", "needs 2 Markets (have 0)", "", "", 3, 0},
		{"Spice Fleet", "120 gold", "40 culture, 10 faith", "needs 1 Harbor (have 0)", "", "", 3, 0},
	}
	mapW := w * 2 / 3
	y0, y1 := 3, h-5
	g.segs(2, 1, xseg{"Caravans ", xDim}, xseg{"2 of 3 out", xBright}, xseg{"    Each trip brings ", xDim}, xseg{"+8 wood", xGood}, xseg{"  ", xDim}, xseg{"+12 iron", xBad}, xseg{" (held)", xDim})

	// The land: a dusting of ground, so the roads cross somewhere.
	for y := y0; y <= y1; y++ {
		for x := 2; x < mapW-32; x++ {
			switch v := (x*7919 + y*104729 + x*y*31) % 97; {
			case v == 0:
				g.set(x, y, '∩', xFaint)
			case v == 1 || v == 2:
				g.set(x, y, '"', xFaint)
			case v == 3:
				g.set(x, y, '.', xFaint)
			}
		}
	}
	// The town, at the left, in the middle of the map: a wall, a keep, roofs.
	ty := (y0 + y1) / 2
	town := []string{
		"     ▲       ",
		"  ∩  █  ∩ ∩  ",
		" ∩∩∩ █ ∩∩∩∩∩ ",
		"▐███████████▌",
		"▐█ ∩∩ ∩ ∩∩ █▌",
		" ▀▀▀▀▀▀▀▀▀▀▀ ",
	}
	for i, l := range town {
		for j, r := range []rune(l) {
			k := xHill
			switch r {
			case ' ':
				continue
			case '∩':
				k = xGold
			case '▀':
				k = xDim
			}
			g.set(3+j, ty-4+i, r, k)
		}
	}
	g.set(8, ty-4, '▲', xAccent)
	g.set(8, ty-3, '█', xTitle)
	g.set(8, ty-2, '█', xTitle)
	g.put(5, ty+2, " Galgate ", xLabel)
	gateX := 17

	// The destinations, down the right of the map, and a road to each.
	n := len(routes)
	labelX := mapW - 30
	rows := make([]int, n)
	for i := range routes {
		rows[i] = y0 + 1 + i*(y1-y0-2)/(n-1)
	}
	// The trunk: the road out of the gate, as far as the last turn.
	lastBend := gateX + 3 + 3*(n/2)
	for x := gateX; x <= lastBend; x++ {
		g.set(x, ty, '─', xDim)
	}
	for i, r := range routes {
		dy := rows[i]
		// The road farthest from the town's row turns off first.
		dist := dy - ty
		if dist < 0 {
			dist = -dist
		}
		step := (y1 - y0) / (n - 1)
		bend := lastBend - 3*((dist+step/2)/max(step, 1))/1
		bend = clamp(bend, gateX+2, lastBend)
		k, line := xDim, '─'
		switch r.state {
		case 0:
			k, line = xGood, '━'
		case 1:
			k, line = xBad, '━'
		case 3:
			k, line = xFaint, '┄'
		}
		if i == sel {
			k = xAccent
		}
		vert := '│'
		if line == '━' {
			vert = '┃'
		}
		if line == '┄' {
			vert = '┆'
		}
		switch {
		case dy < ty:
			for y := dy + 1; y < ty; y++ {
				g.set(bend, y, vert, k)
			}
			g.set(bend, ty, '┴', xDim)
			g.set(bend, dy, '╭', k)
		case dy > ty:
			for y := ty + 1; y < dy; y++ {
				g.set(bend, y, vert, k)
			}
			g.set(bend, ty, '┬', xDim)
			g.set(bend, dy, '╰', k)
		}
		for x := bend + 1; x < labelX-1; x++ {
			g.set(x, dy, line, k)
		}
		// A caravan on the road: out for the first half of a trip, home for
		// the second.
		if r.state <= 1 {
			span := labelX - 3 - (bend + 2)
			p := r.prog * 2
			mark := '▸'
			if p > 1 {
				p, mark = 2-p, '◂'
			}
			cx := bend + 2 + int(p*float64(span))
			load := r.out + " out"
			if mark == '◂' {
				load = r.in + " home"
			}
			g.put(cx-1, dy, "◆◆", xBright)
			g.put(cx-len([]rune(load))/2, dy-1, " "+load+" ", xLabel)
		}
		// The place at the end of the road.
		mk, nk := '◆', xBright
		chip, ck := "", xKeyName
		switch r.state {
		case 0:
			chip, ck = "running", xChipShort
		case 1:
			chip, ck = "held", xChipMil
		case 2:
			chip, ck = "ready", xChipCivic
		case 3:
			mk, nk, chip = '◇', xDim, "locked"
		}
		g.set(labelX, dy, mk, k)
		x := g.put(labelX+2, dy, r.name, nk)
		g.put(x+1, dy, " "+chip+" ", ck)
		g.hit(hitDeckCard, i, labelX, dy, 30, 2)
		if i == sel {
			g.put(labelX-2, dy, "▸", xAccent)
		}
		g.segs(labelX+2, dy+1, xseg{r.out, xText}, xseg{" for ", xDim}, xseg{r.in, xGood})
		if r.note != "" {
			nk := xDim
			if r.state != 0 {
				nk = xBad
			}
			if i != sel || r.cmd == "" {
				g.putFit(labelX+2, dy+2, 30, r.note, nk)
			}
		}
		if i == sel && r.cmd != "" {
			ex := g.keyButton(labelX+2, dy+2, "Enter", r.act)
			g.put(labelX+2, dy+3, r.cmd, xDim)
			_ = ex
		}
	}

	// The books, at the right: what the routes move.
	bx := mapW + 8
	g.box(bx, y0, w-bx-2, y1-y0+1, xFaint, "The books", xTitle, false)
	lines := [][]xseg{
		{{"Going out each trip", xDim}},
		{{"  −10 food", xBad}, {"   Local Barter", xDim}},
		{{"  −40 stone", xBad}, {"  Iron Road", xDim}},
		{},
		{{"Coming home", xDim}},
		{{"  +8 wood", xGood}, {"    Local Barter", xDim}},
		{{"  +12 iron", xBad}, {"   held at the border", xDim}},
		{},
		{{"Allies", xDim}},
		{{"  Riverlands Tribes ", xText}, {"+10% food", xGood}},
		{},
		{{"⚠ A war, or your embargo, holds iron.", xBad}},
		{{"  diplomacy neutral <civ>", xDim}},
		{},
		{{"More on other peoples: ", xDim}, {" factions ", xKeyName}},
	}
	for i, l := range lines {
		g.segsFit(bx+2, y0+1+i, w-bx-6, l...)
	}
	keys := [][2]string{{"↑↓", "pick a road"}, {"Enter", routes[sel].act}, {"Esc", "close"}}
	expKeyBar(g, expLayout{w: w, h: h, footY: h - 2}, keys, "trade route start/stop <key>")
	return g
}

// ===== Army: the wall and the muster =====

func sceneArmy(w, h int) *xgrid {
	g := sceneFrame("Army · Medieval Age", w, h)
	inner := w - 4
	// The balance: your defense from the left, the threat from the right.
	def, thr := 1240.0, 2810.0
	barW := inner - 44
	split := int(def / (def + thr) * float64(barW))
	g.segs(2, 2, xseg{"DEFENSE ", xGood}, xseg{"1,240", xBright})
	bx := 20
	for i := 0; i < barW; i++ {
		if i < split {
			g.set(bx+i, 2, '█', xGood)
		} else {
			g.set(bx+i, 2, '█', xBad)
		}
	}
	g.set(bx+split, 2, '▌', xBright)
	g.segs(bx+barW+2, 2, xseg{"THREAT ", xBad}, xseg{"2.81K", xBright})
	g.segs(2, 3, xseg{"blunts 31% of a raid", xDim})
	g.put(bx+barW+2, 3, "raids in the Medieval Age", xDim)
	g.put(bx+split-9, 3, "the line holds here", xDim)

	// The wall, across the panel: towers, a gate, and what is outside.
	wy := 6
	for x := 2; x < w-2; x++ {
		top := '▄'
		if (x/2)%2 == 0 {
			top = '█'
		}
		g.set(x, wy, top, xDim)
		g.set(x, wy+1, '█', xHill)
		g.set(x, wy+2, '█', xHill)
		g.set(x, wy+3, '▀', xFaint)
	}
	for _, tx := range []int{6, 30, 54, w/2 - 8, w/2 + 6, w - 56, w - 32, w - 9} {
		for dx := 0; dx < 3; dx++ {
			g.set(tx+dx, wy-1, '█', xDim)
			g.set(tx+dx, wy, '█', xDim)
		}
		g.set(tx+1, wy-2, '▲', xAccent)
	}
	g.put(w/2-3, wy+1, " ▄███▄ ", xKeyName)
	g.put(w/2-3, wy+2, " █   █ ", xKeyName)
	// The raiders, outside: as many as the threat is to the defense.
	for i := 0; i < 34; i++ {
		g.set(w-6-i*2-(i%3), wy+5, '▲', xBad)
	}
	g.put(w-78, wy+5, "raiders", xBad)
	wy += 3

	// The muster: who stands behind it.
	g.put(2, wy+4, "The muster", xTitle)
	g.segs(14, wy+4, xseg{"204 soldiers", xBright}, xseg{"  ·  164 at their posts  ·  ", xDim}, xseg{"40 away", xLegend}, xseg{"  ·  each ▲ is 4 soldiers, gold ones are away", xDim})
	units := []struct {
		name       string
		home, away int
		post       string
	}{{"Men-at-arms", 120, 40, "Castle Keep ×2"}, {"Archers", 60, 0, "Watchtower ×3"}, {"Knights", 24, 0, "Great Hall"}}
	for i, u := range units {
		y := wy + 6 + i*2
		x := g.segs(2, y, xseg{fmt.Sprintf("%-12s", u.name), xText}, xseg{fmt.Sprintf("%4d  ", u.home+u.away), xBright})
		for j := 0; j < u.home/4; j++ {
			g.set(x, y, '▲', xArmy)
			x++
			if j%10 == 9 {
				x++
			}
		}
		for j := 0; j < u.away/4; j++ {
			g.set(x, y, '▲', xLegend)
			x++
		}
		g.put(max(x+3, 86), y, u.post, xDim)
	}

	// The orders: a campaign out, and the ones that could go.
	oy := wy + 13
	g.put(2, oy, "Orders", xTitle)
	for x := 9; x < w-2; x++ {
		g.set(x, oy, '─', xFaint)
	}
	g.segs(2, oy+2, xseg{"⚔ ", xLegend}, xseg{"Raid on the Coast", xBright}, xseg{"  40 soldiers out  ", xDim})
	px := 42
	for i := 0; i < 40; i++ {
		r, k := '━', xLegend
		if i >= 26 {
			r, k = '·', xFaint
		}
		g.set(px+i, oy+2, r, k)
	}
	g.put(px+25, oy+2, "◆◆", xBright)
	g.segs(px+42, oy+2, xseg{"3m 10s left", xText}, xseg{"   comes home with about 20 min of income", xDim})
	g.put(1, oy+4, "▸", xAccent)
	x := g.segs(2, oy+4, xseg{"⚔ ", xAccent}, xseg{"Border Skirmish", xBright}, xseg{"   needs 40 soldiers · odds ", xDim}, xseg{"72%", xGood}, xseg{" · about 20 min of income   ", xDim})
	ex := g.keyButton(x, oy+4, "Enter", "send")
	g.put(ex+1, oy+4, "expedition border_skirmish", xDim)
	g.segs(2, oy+6, xseg{"⚔ ", xFaint}, xseg{"Siege of Ashford", xDim}, xseg{"  needs 200 soldiers   ", xDim}, xseg{"needs 2 Barracks (have 0)", xBad})
	g.put(2, oy+9, "Lately", xTitle)
	for x := 9; x < w-2; x++ {
		g.set(x, oy+9, '─', xFaint)
	}
	g.segs(2, oy+10, xseg{"Raiders came over the east wall. ", xText}, xseg{"12 fell", xBad}, xseg{", the granary stands.", xText})
	g.segs(2, oy+11, xseg{"The Watchtower at the ford is manned: ", xText}, xseg{"+20 archers", xGood}, xseg{".", xText})
	expKeyBar(g, expLayout{w: w, h: h, footY: h - 2}, [][2]string{{"↑↓", "pick an order"}, {"Enter", "send"}, {"Esc", "close"}}, "expedition <name>")
	return g
}

// ===== Milestones: the roads =====

type mockStone struct {
	name string
	done bool
	prog float64
}

func sceneMiles(w, h int) *xgrid {
	g := sceneFrame("Milestones · Medieval Age", w, h)
	g.segs(2, 1, xseg{"9 of 77 reached", xBright}, xseg{"    you are called ", xDim}, xseg{"\"Aspiring\"", xTitle}, xseg{"    next name at 12", xDim})
	tracks := []struct {
		name   string
		k      xk
		stones []mockStone
	}{
		{"Settlement", xWork, []mockStone{{"First Shelter", true, 1}, {"Small Village", false, 0.08}, {"Bustling Town", false, 0}, {"Growing City", false, 0}, {"Metropolis", false, 0}, {"Urban Sprawl", false, 0}, {"Megalopolis", false, 0}}},
		{"Builder", xLore, []mockStone{{"First Storehouse", true, 1}, {"Early Builder", true, 1}, {"Seasoned Builder", false, 0.6}, {"Master Builder", false, 0}, {"Grand Architect", false, 0}}},
		{"Scholar", xWater, []mockStone{{"First Words", true, 1}, {"Bronze Scholar", true, 1}, {"Classical Scholar", true, 1}, {"Medieval Scholar", false, 0.25}, {"Polymath", false, 0}, {"Enlightened", false, 0}}},
		{"Warlord", xBad, []mockStone{{"First Blood", false, 0.5}, {"Raider", false, 0}, {"Conqueror", false, 0}, {"Legend", false, 0}}},
		{"Merchant", xGold, []mockStone{{"First Deal", true, 1}, {"Caravan Master", false, 0.33}, {"Silk Baron", false, 0}, {"Trade Empire", false, 0}}},
		{"Faithful", xGood, []mockStone{{"First Altar", true, 1}, {"Shrine Keeper", true, 1}, {"High Priest", false, 0.1}, {"Living Saint", false, 0}}},
		{"Explorer", xArmy, []mockStone{{"First Steps", true, 1}, {"Pathfinder", false, 0.7}, {"Cartographer", false, 0}, {"World Walker", false, 0}}},
		{"Wonders", xBright, []mockStone{{"Great Monolith", true, 1}, {"Stonehenge", true, 1}, {"Colossus", false, 0.4}, {"Great Library", false, 0}, {"Sistine Chapel", false, 0}}},
		{"Ages", xLegend, []mockStone{{"Stone", true, 1}, {"Bronze", true, 1}, {"Iron", true, 1}, {"Classical", true, 1}, {"Medieval", true, 1}, {"Renaissance", false, 0}, {"Colonial", false, 0}, {"Industrial", false, 0}}},
	}
	selT, selS := 1, 2
	x0, trackW := 20, w-24
	y := 3
	for ti, t := range tracks {
		done := 0
		for _, s := range t.stones {
			if s.done {
				done++
			}
		}
		g.put(2, y, t.name, t.k)
		g.segs(2, y+1, xseg{fmt.Sprintf("%d of %d", done, len(t.stones)), xDim})
		step := min(trackW/len(t.stones), 22)
		for si, s := range t.stones {
			sx := x0 + si*step
			// The road to the next stone: walked where the stone behind is done.
			if si < len(t.stones)-1 {
				for dx := 1; dx < step; dx++ {
					r, k := '─', xFaint
					if s.done && t.stones[si+1].done {
						r, k = '━', t.k
					} else if s.done {
						if float64(dx) <= t.stones[si+1].prog*float64(step-1) {
							r, k = '━', t.k
						}
					}
					g.set(sx+dx, y, r, k)
				}
			}
			mark, mk, nk := '○', xFaint, xDim
			switch {
			case s.done:
				mark, mk, nk = '●', t.k, xText
			case si == 0 || t.stones[si-1].done:
				mark, mk, nk = '◉', xBright, xBright
			}
			if ti == selT && si == selS {
				g.put(sx-1, y, "[", xAccent)
				g.put(sx+1, y, "]", xAccent)
				nk = xAccent
			}
			g.set(sx, y, mark, mk)
			g.put(sx, y+1, fitText(s.name, step-1), nk)
		}
		y += 3
	}
	// The stone under the pick, in full.
	dy := h - 9
	g.box(2, dy, w-4, 6, xAccent, "Seasoned Builder", xBright, false)
	g.put(w-22, dy, " on the Builder road ", xLore)
	g.segs(4, dy+1, xseg{"Build 500 structures total.", xText})
	bw := 60
	for i := 0; i < bw; i++ {
		r, k := '█', xLore
		if i >= bw*6/10 {
			r, k = '░', xFaint
		}
		g.set(4+i, dy+2, r, k)
	}
	g.segs(6+bw, dy+2, xseg{"300 of 500", xBright}, xseg{" buildings built", xDim})
	g.segs(4, dy+3, xseg{"Reward ", xDim}, xseg{"-5% building costs", xGood}, xseg{"      Waits for ", xDim}, xseg{"the Iron Age, reached", xGood})
	g.segs(4, dy+4, xseg{"Before it ", xDim}, xseg{"Early Builder ✓", xText}, xseg{"      After it ", xDim}, xseg{"Master Builder", xText})
	expKeyBar(g, expLayout{w: w, h: h, footY: h - 2}, [][2]string{{"←→", "along a road"}, {"↑↓", "road"}, {"Esc", "close"}}, "")
	return g
}

func TestZZSceneMock(t *testing.T) {
	out := os.Getenv("SCENE_MOCK_DIR")
	if out == "" {
		t.Skip("set SCENE_MOCK_DIR")
	}
	keepTheme(t)
	for name, g := range map[string]*xgrid{"trade": sceneTrade(153, 38, 3), "army": sceneArmy(153, 38), "miles": sceneMiles(153, 38)} {
		b := kit.NewBuffer(g.w, g.h)
		paintExpeditionKit(kit.NewPainter(b), g, image.Point{}, newExpPalette())
		f, err := os.Create(filepath.Join(out, "SC-"+name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, ScreenImage(b, shotCellW, shotCellH)); err != nil {
			t.Fatal(err)
		}
		f.Close()
		if strings.Contains(os.Getenv("SCENE_MOCK_TEXT"), name) {
			t.Log("\n" + g.String())
		}
	}
}
