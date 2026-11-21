package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
)

// stats_lifetime.go is the Stats page's Lifetime part: the account as a
// record board. A plaque and four counts, the ladder of ages reached, the
// best run of each count and the badge families.

var (
	txtFamilies = rowid.E("ui.stats_lifetime.txtFamilies.01", "Badge families")
	txtBadgePts = rowid.E("ui.stats_lifetime.txtBadgePts.01", "Badge points")
	txtFastest  = rowid.E("ui.stats_lifetime.txtFastest.01", "Fastest")
	txtBestRun  = rowid.E("ui.stats_lifetime.txtBestRun.01", "Best run")
	txtPickAge  = rowid.E("ui.stats_lifetime.txtPickAge.01", "pick age")
)

// statsPlaque is the account's plaque: its name, title and progress toward
// the next title (no bar and no line at the last).
type statsPlaque struct {
	name, title, points string
	frac                float64
	bar                 bool
}

// statsStep is one age's step of the ladder. r is its cells; mark is the row
// its mark takes with no pictures.
type statsStep struct {
	readout, time string
	reached       bool
	r             image.Rectangle
	mark          int
}

// recordDate is how a record's date is written.
const recordDate = "2 Jan 2006"

// lifetime lays the part out and returns the page's rows.
func (pg *statsPage) lifetime(st game.GameState, w int, wide bool, grow int) int {
	as := st.AccountStats
	if as == nil {
		pg.g.put(0, 1, txtNoAccount.String(), xDim)
		return 3
	}
	bs := as.BadgeSummary
	total := bs.Shown
	if bs.HiddenCounted {
		total += bs.Hidden
	}
	pl := &statsPlaque{name: as.DisplayName, title: bs.Title}
	if bs.NextTitle != "" {
		pl.bar, pl.frac = true, math.Min(1, float64(bs.Points)/float64(max(1, bs.NextTitleAt)))
		pl.points = fmt.Sprintf(txtOfPoints.String(), bs.Points, bs.NextTitleAt, bs.NextTitle)
	}
	tiles := []statsTile{{span: 2, plaque: pl},
		countTile(txtPrestiges.String(), "wonder", as.TotalPrestiges),
		countTile(txtCivs.String(), "diplomacy", as.CivilizationsStarted),
		{name: txtBadgesEarn.String(), text: fmt.Sprintf("%d/%d", bs.Earned, total), kind: "trade"},
		countTile(txtBadgePts.String(), "culture", bs.Points)}
	n := 4 // narrower: the plaque takes a row, the four counts the next
	if wide {
		n = 6
	} else {
		tiles[0].span = n
	}
	y := pg.layTiles(tiles, 0, w, n, 3+grow)
	y = pg.ladder(st, as, y, w, grow)
	y, rec := pg.records(st, as, y, w, grow)
	pg.tiles = append(tiles, rec...)
	return pg.families(st, as, y, w)
}

// layPlaque puts the plaque's picture and words (or, with no pictures, its
// rows) in the tile's cells: name, title, the bar and the points line.
func (pg *statsPage) layPlaque(t *statsTile, nr int) {
	pl, r := t.plaque, t.r
	x, y, w := r.Min.X, r.Min.Y, r.Dx()
	acc, ttl := txtAccount.String(), txtTitleOf.String()
	if !pg.pic {
		pg.g.put(x, y, acc, xDim)
		pg.g.put(x+len(acc)+2, y, fitText(pl.name, w-len(acc)-2), xBright)
		pg.g.put(x, y+1, ttl, xDim)
		pg.g.put(x+len(ttl)+2, y+1, fitText(pl.title, w-len(ttl)-2), xAccent)
		if pl.bar {
			fill := int(pl.frac*float64(w) + .5)
			for i := 0; i < w; i++ {
				c, k := '.', xFaint
				if i < fill {
					c, k = '#', xAccent
				}
				pg.g.set(x+i, y+2, c, k)
			}
			pg.g.put(x, y+3, fitText(pl.points, w), xText)
		}
		return
	}
	th := r.Dy()
	pg.pics = append(pg.pics, statsPic{r: r, key: fmt.Sprint("plaque", w, th, nr),
		draw: func(cw, ch int, _ bool, in statsInk) *image.RGBA {
			return statsTilePic(w, th, nr+1, cw, ch, "", false, in)
		}})
	pg.labels = append(pg.labels,
		statsLabel{x + 1, y, acc, xDim, ""}, statsLabel{x + len(acc) + 3, y, fitText(pl.name, w-len(acc)-4), xBright, ""},
		statsLabel{x + 1, y + 1, ttl, xDim, ""}, statsLabel{x + len(ttl) + 3, y + 1, fitText(pl.title, w-len(ttl)-4), xAccent, ""})
	if !pl.bar {
		return
	}
	frac, tw := pl.frac, w-2
	pg.pics = append(pg.pics, statsPic{r: image.Rect(x+1, y+nr, x+w-1, y+nr+1), key: fmt.Sprint("plaquebar", tw, frac),
		draw: func(cw, ch int, _ bool, in statsInk) *image.RGBA {
			in.well = toward(in.well, in.shadow, .7)
			return statsBarPic(tw, 1, cw, ch, frac, nil, signTints["trade"], false, in)
		}})
	pg.labels = append(pg.labels, statsLabel{x + 1, y + nr + 1, fitText(pl.points, w-2), xText, ""})
}

// ladder lays the ages out as steps, each a little higher than the last, and
// the fastest time of each step reached under it where it fits whole.
func (pg *statsPage) ladder(st game.GameState, as *game.AccountStatsView, y, w, grow int) int {
	ages := st.Ruleset().Ages()
	n, hi := len(ages), 0
	for i, a := range ages {
		if a.Key == as.HighestAge {
			hi = i
		}
	}
	pg.top = hi
	pg.heads[0] = image.Rect(0, y, w, y+1)
	pg.heading(0, y, w, txtHighestAge.String())
	y++
	sw := max(1, w/n)
	x0 := (w - sw*n) / 2
	hc := 5 + grow
	if !pg.pic {
		hc = 2
	}
	for i, a := range ages {
		s := statsStep{reached: i <= hi, r: image.Rect(x0+i*sw, y, x0+(i+1)*sw, y+hc), mark: y + hc - 1}
		if s.reached {
			s.readout = a.Name
			if rec, ok := as.FastestAge[a.Key]; ok {
				s.time = formatTicks(rec.Ticks, st)
				s.readout += ": " + txtFastest.String() + " " + s.time
				if rec.Run != "" {
					s.readout += ", " + rec.Run
				}
				if !rec.At.IsZero() {
					s.readout += ", " + rec.At.Format(recordDate)
				}
			}
			pg.g.hit(hitStep, i, s.r.Min.X, y, sw, hc+1)
			if rl := runeLen(s.time); rl > 0 && rl < sw {
				pg.g.put(s.r.Min.X+(sw-rl)/2, y+hc, s.time, xText)
			}
		}
		if pg.pic {
			reached, crown := s.reached, i == hi
			pg.pics = append(pg.pics, statsPic{r: s.r, key: fmt.Sprint("step", sw, hc, i, n, reached, crown), n: i,
				draw: func(cw, ch int, hot bool, in statsInk) *image.RGBA {
					return statsStepPic(sw, hc, cw, ch, i, n, reached, crown, hot, blockTone(signKinds[i%len(signKinds)], in), in)
				}})
			if reached {
				pg.pics[len(pg.pics)-1].hit = hitStep
			}
		} else {
			c, k := strings.Repeat(".", sw-1), xFaint
			if s.reached {
				c, k = strings.Repeat("#", sw-1), xText
			}
			pg.g.put(s.r.Min.X, s.mark, c, k)
			if i == hi {
				pg.g.put(s.r.Min.X+(sw-1)/2, y, "^", xGold)
			}
		}
		pg.steps = append(pg.steps, s)
	}
	return y + hc + 2
}

// records lays the four best-run tiles out and returns the row after them
// and the tiles.
func (pg *statsPage) records(st game.GameState, as *game.AccountStatsView, y, w, grow int) (int, []statsTile) {
	pg.heading(0, y, w, txtBestRun.String())
	y++
	now := []int{st.Stats.TotalBuilt, st.Stats.TotalRecruited, st.Research.TotalResearched, st.Military.CompletedCount}
	names := []string{txtBuilt.String(), txtRecruited.String(), txtTechs.String(), txtMissions.String()}
	kinds := []string{"housing", "culture", "knowledge", "harbor"}
	var tiles []statsTile
	for i, key := range []string{game.RunBestBuilt, game.RunBestRecruited, game.RunBestResearched, game.RunBestExpeditions} {
		rec := as.BestRun[key]
		t := countTile(names[i], kinds[i], max(rec.Value, now[i]))
		if t.hold = now[i] >= rec.Value; t.hold {
			t.note = txtThisRun.String()
		} else {
			t.note = strings.TrimSpace(rec.Run + " " + rec.At.Format(recordDate))
		}
		tiles = append(tiles, t)
	}
	n := 2
	if w >= statsWide-4 {
		n = 4
	}
	return pg.layTiles(tiles, y, w, n, 3+grow), tiles
}

// families lays the worker domains' badges out in a grid: each domain's best
// tier earned (a blank mark where none is), its name and its count.
func (pg *statsPage) families(st game.GameState, as *game.AccountStatsView, y, w int) int {
	pg.heading(0, y, w, txtFamilies.String())
	y++
	doms := buildRanks(st.Ruleset(), as.Badges).doms
	cols := clamp(w/26, 2, 6)
	pw := w / cols
	for i, d := range doms {
		x, ty := i%cols*pw, y+i/cols*3
		count := fmt.Sprintf("%d / %d", d.earned, d.all)
		k := xDim
		if d.earned > 0 {
			k = xBright
		}
		mark := "[ ]"
		for t := config.BadgeLegendary; t >= config.BadgeBronze; t-- {
			if d.have[t] > 0 {
				mark = "[" + strings.ToUpper(t.Name()[:1]) + "]"
				break
			}
		}
		pg.g.put(x, ty, mark, xAccent)
		pg.g.put(x+4, ty, fitText(d.name, pw-5), xText)
		pg.g.put(x+4, ty+1, count, k)
	}
	return y + (len(doms)+cols-1)/cols*3
}

// statsCrown is the crown over the highest step, '#' lit.
var statsCrown = []string{"#...#...#", "##.###.##", "#########", "#########", "#########"}

// statsStepPic is one age's step of the ladder, wc by hc cells: a column in
// the age's tone, a pale lip along the top and a dark side, rising with the
// age; a step not reached is an empty recess. The highest reached step wears
// the crown; a hot one (under the pointer) is outlined.
func statsStepPic(wc, hc, cw, ch, i, n int, reached, crown, hot bool, tone color.RGBA, in statsInk) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, wc*cw, hc*ch))
	k := 2 * max(1, ch/21)
	a := art{img, k}
	W, H := (img.Rect.Dx()+k-1)/k, (img.Rect.Dy()+k-1)/k
	room := len(statsCrown) * 2
	lo, top := H/4, H-room-3
	h := lo + (top-lo)*i/max(1, n-1)
	y0 := H - h
	if !reached {
		a.fill(image.Rect(0, y0, W-1, H), 0, in.well)
		a.hline(0, W-1, y0, in.shadow)
		for y := y0; y < H; y++ {
			a.px(0, y, in.shadow)
		}
		a.hline(0, W-1, H-1, toward(in.well, in.face, .12))
		return img
	}
	R := rampOf(tone)
	a.fill(image.Rect(0, y0, W, H), 0, R.base)
	a.hline(0, W, y0, R.pale)
	a.hline(0, W, y0+1, R.pale)
	for y := y0; y < H; y++ {
		a.px(W-1, y, R.dark)
	}
	a.hline(0, W, H-1, R.dark)
	a.hline(0, W, H-2, R.dark)
	if hot {
		edge := toward(R.pale, keyWhite, .5)
		a.hline(0, W, y0, edge)
		a.hline(0, W, y0+1, edge)
		for y := y0; y < H; y++ {
			a.px(0, y, edge)
			a.px(1, y, edge)
			a.px(W-1, y, edge)
			a.px(W-2, y, edge)
		}
	}
	if crown {
		G := rampOf(signTints["trade"])
		cx, cy := (W-2*len(statsCrown[0]))/2, y0-room-1
		for r, row := range statsCrown {
			c := G.base
			switch r {
			case 0:
				c = G.pale
			case len(statsCrown) - 1:
				c = G.dark
			}
			for q, ch := range row {
				if ch == '#' {
					a.fill(image.Rect(cx+2*q, cy+2*r, cx+2*q+2, cy+2*r+2), 0, c)
				}
			}
		}
	}
	return img
}
