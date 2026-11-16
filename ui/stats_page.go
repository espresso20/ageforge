package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/window/kit"
)

// stats_page.go is the Stats panel: a record board of the run in two parts,
// "This run" (headline tiles, the ages as a band of blocks, ranked bars of
// what was gathered, tables of multipliers, prestige and epoch) and
// "Lifetime" (the account's tiles). stats_pix.go draws its pictures.

var (
	txtStatsTitle  = vibe.E("ui.stats_page.txtStatsTitle.01", "Statistics")
	txtThisRun     = vibe.E("ui.stats_page.txtThisRun.01", "This run")
	txtLifetime    = vibe.E("ui.stats_page.txtLifetime.01", "Lifetime")
	txtPlayTime    = vibe.E("ui.stats_page.txtPlayTime.01", "Play time")
	txtBuilt       = vibe.E("ui.stats_page.txtBuilt.01", "Buildings built")
	txtRecruited   = vibe.E("ui.stats_page.txtRecruited.01", "Workers recruited")
	txtTechs       = vibe.E("ui.stats_page.txtTechs.01", "Techs researched")
	txtMissions    = vibe.E("ui.stats_page.txtMissions.01", "Expeditions and campaigns done")
	txtPrestigeLvl = vibe.E("ui.stats_page.txtPrestigeLvl.01", "Prestige level")
	txtAgesRun     = vibe.E("ui.stats_page.txtAgesRun.01", "Ages this run")
	txtSoFar       = vibe.E("ui.stats_page.txtSoFar.01", "so far")
	txtGatheredHd  = vibe.E("ui.stats_page.txtGatheredHd.01", "Total gathered")
	txtGathered    = vibe.E("ui.stats_page.txtGathered.01", "gathered")
	txtPerSecond   = vibe.E("ui.stats_page.txtPerSecond.01", "per second")
	txtMultsHd     = vibe.E("ui.stats_page.txtMultsHd.01", "Active multipliers")
	txtNoMults     = vibe.E("ui.stats_page.txtNoMults.01", "No active multipliers.")
	txtPrestigeHd  = vibe.E("ui.stats_page.txtPrestigeHd.01", "Prestige and epoch")
	txtResearchHd  = vibe.E("ui.stats_page.txtResearchHd.01", "Research bonuses")
	txtAccount     = vibe.E("ui.stats_page.txtAccount.01", "Account")
	txtTitleOf     = vibe.E("ui.stats_page.txtTitleOf.01", "Title")
	txtPrestiges   = vibe.E("ui.stats_page.txtPrestiges.01", "Total prestiges")
	txtCivs        = vibe.E("ui.stats_page.txtCivs.01", "Civilizations started")
	txtHighestAge  = vibe.E("ui.stats_page.txtHighestAge.01", "Highest age ever")
	txtBadgesEarn  = vibe.E("ui.stats_page.txtBadgesEarn.01", "Badges earned")
	txtOfPoints    = vibe.E("ui.stats_page.txtOfPoints.01", "%d of %d points to %s")
	txtNoAccount   = vibe.E("ui.stats_page.txtNoAccount.01", "No account loaded.")
	txtSwitchPart  = vibe.E("ui.stats_page.txtSwitchPart.01", "switch part")
)

// hit kinds on the page.
const (
	hitAge = iota + 1
	hitBar
)

// the parts of the panel.
const (
	partRun = iota
	partLifetime
)

// statsWide is the screen width from which the page has two columns and the
// tiles stand six across.
const statsWide = 140

// statsTile is one tile: its label (and a second line under it), the value
// as numeral glyphs when the font has them, else as text, the sign it
// carries and the cells it takes on the page.
type statsTile struct {
	name, note, text string
	toks             []numTok
	kind             string
	r                image.Rectangle
}

// statsBlock is an age's block of the band.
type statsBlock struct {
	name, time, readout string
	open, named, timed  bool
	r                   image.Rectangle
}

// statsBar is a ranked bar. frac is its length as a share of the track.
type statsBar struct {
	name, readout, total, rate string
	value, frac                float64
	track                      image.Rectangle
}

// statsRow is a line of a table: a label, a value and the value's colour.
type statsRow struct {
	label, value string
	k            xk
}

// statsPic is a picture on the page, drawn when painted so it takes the
// window's cell size and the theme as they are then.
type statsPic struct {
	r        image.Rectangle
	key      string
	hit, n   int
	draw     func(cw, ch int, hot bool, in statsInk) *image.RGBA
	withTone bool
}

// statsLabel is a word over a picture; tint, when set, is the sign kind of
// the block under it, whose ink it takes.
type statsLabel struct {
	x, y int
	s    string
	k    xk
	tint string
}

// statsPage is the body of the page for one part, laid out for a width.
type statsPage struct {
	g        *xgrid // the cells, and the hits
	used     int    // rows the page takes
	pic      bool
	tiles    []statsTile
	blocks   []statsBlock
	bars     []statsBar
	ticks    []float64 // where each decade falls on a track, as a share of it
	mults    []multRow
	rows     []statsRow
	research []statsRow
	pics     []statsPic
	labels   []statsLabel
	heads    [2]image.Rectangle // the heading rows of the ages band and of the bars, as the page's cells
	ends     [2]int             // the row each column ends at (the left one holds the bars)
}

// statsKinds is the sign each resource's bar takes its colour from.
var statsKinds = map[string]string{
	"food": "food", "wood": "wood", "stone": "stone", "gold": "trade", "iron": "metal", "iron_ore": "metal",
	"coal": "energy", "marble": "monument", "faith": "faith", "steel": "metal", "culture": "culture", "oil": "energy",
	"electricity": "energy", "uranium": "energy", "data": "hacker", "nanobots": "engineering",
	"titanium_ore": "metal", "crypto": "hacker", "dark_matter_crystals": "wonder", "plasma": "energy",
	"titanium": "metal", "dark_matter": "wonder", "antimatter": "wonder", "quantum_flux": "knowledge",
}

// blockTone is an age block's colour: its sign's tint, quieted toward the
// ground.
func blockTone(kind string, in statsInk) color.RGBA {
	return toward(signTints[kind], in.ground, .35)
}

// commas writes a count in full, with a comma between thousands.
func commas(v float64) string {
	s := strconv.FormatFloat(math.Round(v), 'f', 0, 64)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		s = "-" + s
	}
	return s
}

// timeToks spells a time as its two largest units, "4h 12m": digits at full
// height, the units at half.
func timeToks(d time.Duration) ([]numTok, string) {
	secs := int(d / time.Second)
	vals := []int{secs / 86400, secs % 86400 / 3600, secs % 3600 / 60, secs % 60}
	i := 0
	for i < 3 && vals[i] == 0 {
		i++
	}
	var toks []numTok
	var text []string
	for j := i; j < min(i+2, 4); j++ {
		num, _ := numToks(strconv.Itoa(vals[j]), false)
		unit := []numTok{{rune("dhms"[j]), true}}
		if len(toks) > 0 {
			toks = append(toks, numTok{' ', false})
		}
		toks = append(toks, append(num, unit...)...)
		text = append(text, strconv.Itoa(vals[j])+string("dhms"[j]))
	}
	return toks, strings.Join(text, " ")
}

// countTile is a tile of a count.
func countTile(name, kind string, n int) statsTile {
	text := commas(float64(n))
	toks, _ := numToks(text, false)
	return statsTile{name: name, text: text, toks: toks, kind: kind}
}

// buildStats lays the page of a part out in w cells, a row at a time down
// the page, with pictures when pic is set. view is the rows the screen shows,
// which the bars share out. grow adds height to what is drawn (taller
// numerals and courses) for a page that would be bare.
func buildStats(st game.GameState, part, w, view int, wide, pic bool, grow int) *statsPage {
	pg := &statsPage{g: newXGrid(w, 400), pic: pic}
	y := 0
	if part == partLifetime {
		y = pg.lifetime(st, w, grow)
	} else {
		y = pg.run(st, w, view, wide, grow)
	}
	pg.used = y
	pg.g.h = y
	pg.g.c = pg.g.c[:w*y]
	return pg
}

// heading is a section's name with a rule across the rest of the row.
func (pg *statsPage) heading(x, y, w int, s string) {
	end := pg.g.put(x, y, s, xTitle)
	for ; end < x+w; end++ {
		pg.g.set(end, y, '─', xFaint)
	}
	_ = end
}

// tiles lays the tiles out n across from row y and returns the row after.
func (pg *statsPage) layTiles(tiles []statsTile, y, w, n, nr int) int {
	tw := (w - (n - 1)) / n
	lr := 1
	for _, t := range tiles {
		lr = max(lr, len(wrapWords(t.name, tw-2)))
		if t.note != "" {
			lr = max(lr, 2)
		}
	}
	th := nr + 1 + lr
	if !pg.pic {
		th = 2 + lr
	}
	texts := make([]string, len(tiles))
	for i, t := range tiles {
		texts[i] = t.text
	}
	for i := range tiles {
		t := &tiles[i]
		x, ty := i%n*(tw+1), y+i/n*(th+1)
		t.r = image.Rect(x, ty, x+tw, ty+th)
		lines := wrapWords(t.name, tw-2)
		if t.note != "" {
			lines = append(lines[:1], t.note)
		}
		if !pg.pic {
			for j, s := range lines {
				pg.g.put(x, ty+j, fitText(s, tw), xDim)
			}
			pg.g.put(x, ty+lr, fitText(t.text, tw), xBright)
			continue
		}
		for j, s := range lines {
			pg.labels = append(pg.labels, statsLabel{x + 1, ty + nr + 1 + j, fitText(s, tw-2), xText, ""})
		}
		if t.toks == nil {
			pg.labels = append(pg.labels, statsLabel{x + 1, ty + (nr+1)/2, fitText(t.text, tw-2), xBright, ""})
		}
		toks, kind := t.toks, t.kind
		pg.pics = append(pg.pics, statsPic{r: t.r, key: fmt.Sprint("tile", tw, th, nr, texts, i, kind),
			draw: func(cw, ch int, _ bool, in statsInk) *image.RGBA {
				return statsTilePic(tw, th, nr+1, cw, ch, toks, kind, tiles, in)
			}})
	}
	return y + (len(tiles)+n-1)/n*(th+1)
}

// run lays out "This run" and returns the page's rows.
func (pg *statsPage) run(st game.GameState, w, view int, wide bool, grow int) int {
	s := st.Stats
	pt, ptext := timeToks(s.PlayTime.Truncate(config.Billion))
	play := statsTile{name: txtPlayTime.String(), text: ptext, toks: pt, kind: "monument"}
	tiles := []statsTile{play,
		countTile(txtBuilt.String(), "housing", s.TotalBuilt),
		countTile(txtRecruited.String(), "culture", s.TotalRecruited),
		countTile(txtTechs.String(), "knowledge", st.Research.TotalResearched),
		countTile(txtMissions.String(), "harbor", st.Military.CompletedCount),
		countTile(txtPrestigeLvl.String(), "wonder", st.Prestige.Level)}
	n := 3
	if wide {
		n = 6
	}
	y := pg.layTiles(tiles, 0, w, n, 3+grow)
	pg.tiles = tiles
	y = pg.band(st, y, w, grow)
	lw, rw, xr := w, w, 0
	if wide {
		lw = (w - 2) * 11 / 20
		rw, xr = w-2-lw, lw+2
	}
	yl, yr := pg.gathered(st, y, lw, view), y
	pg.mults, pg.rows, pg.research = activeMultipliers(st), prestigeRows(st), researchRows(st)
	sections := []func(x, y, w int) int{
		func(x, y, w int) int { return pg.rowsTable(txtPrestigeHd.String(), pg.rows, x, y, w) },
		pg.multsTable,
		func(x, y, w int) int { return pg.rowsTable(txtResearchHd.String(), pg.research, x, y, w) },
	}
	if !wide {
		sections[0], sections[1] = sections[1], sections[0]
	}
	// the sections go under the shorter column, a tie to the right; the first
	// opens the right one, and narrow pages have the one column
	for i, sec := range sections {
		switch {
		case !wide:
			yl = sec(0, yl+1, w)
		case i > 0 && yl < yr:
			yl = sec(0, yl+1, lw)
		case i > 0:
			yr = sec(xr, yr+1, rw)
		default:
			yr = sec(xr, yr, rw)
		}
	}
	pg.ends = [2]int{yl, yr}
	return max(yl, yr)
}

// band lays the ages out as courses of blocks, as many courses as it takes
// for every name to fit whole.
func (pg *statsPage) band(st game.GameState, y, w, grow int) int {
	log := st.Stats.AgeLog
	if len(log) == 0 {
		return y
	}
	iv := tickInterval(st)
	n := len(log)
	blocks := make([]statsBlock, n)
	weight := make([]int, n)
	for i, e := range log {
		b := &blocks[i]
		end := st.Tick
		if i+1 < n {
			end = log[i+1].Tick
		}
		weight[i] = max(1, end-e.Tick)
		b.name, b.time = game.AgeName(e.Age), game.DurationText(end-e.Tick, iv)
		b.readout = b.name + ": " + b.time
		if b.open = i == n-1; b.open {
			b.time += " " + txtSoFar.String()
			b.readout = b.name + ": " + b.time
		}
	}
	minW := func(b statsBlock) int { return max(runeLen(b.name), runeLen(b.time)) + 2 }
	// the fewest courses in which every group of ages fits the width
	per := n
	for c := 1; c <= n; c++ {
		per = (n + c - 1) / c
		ok := true
		for i := 0; i < n; i += per {
			sum := 0
			for _, b := range blocks[i:min(i+per, n)] {
				sum += minW(b)
			}
			ok = ok && sum <= w
		}
		if ok {
			break
		}
	}
	br := 2 + grow
	gap := 0
	if !pg.pic {
		br, gap = 2, 1
	}
	pg.heads[0] = image.Rect(0, y, w, y+1)
	pg.heading(0, y, w, txtAgesRun.String())
	y++
	for i0 := 0; i0 < n; i0 += per {
		group := blocks[i0:min(i0+per, n)]
		sum, wsum := 0, 0
		for j, b := range group {
			sum += minW(b)
			wsum += weight[i0+j]
		}
		x := 0
		for j := range group {
			b := &group[j]
			bw := minW(*b) + (w-sum)*weight[i0+j]/wsum
			if j == len(group)-1 {
				bw = w - x
			}
			b.r = image.Rect(x, y, x+bw, y+br)
			b.named, b.timed = runeLen(b.name)+2 <= bw, runeLen(b.time)+2 <= bw
			pg.g.hit(hitAge, i0+j, x, y, bw, br)
			pg.layBlock(b, i0+j)
			x += bw
		}
		y += br + gap
	}
	pg.blocks = blocks
	return y + 1
}

// layBlock puts a block's picture and words (or, with no pictures, its words
// between bars) on the page.
func (pg *statsPage) layBlock(b *statsBlock, i int) {
	r := b.r
	kind := signKinds[i%len(signKinds)]
	ny := r.Min.Y + (r.Dy()-2)/2
	if pg.pic {
		open, w, h := b.open, r.Dx(), r.Dy()
		pg.pics = append(pg.pics, statsPic{r: r, key: fmt.Sprint("block", w, h, kind, open), hit: hitAge, n: i,
			draw: func(cw, ch int, hot bool, in statsInk) *image.RGBA {
				return statsBlockPic(w, h, cw, ch, blockTone(kind, in), open, hot, in)
			}})
		if b.named {
			pg.labels = append(pg.labels, statsLabel{r.Min.X + 1, ny, b.name, xBright, kind})
		}
		if b.timed {
			pg.labels = append(pg.labels, statsLabel{r.Min.X + 1, ny + 1, b.time, xText, kind})
		}
		return
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		pg.g.set(r.Min.X, y, '|', xFaint)
	}
	if b.named {
		pg.g.put(r.Min.X+1, r.Min.Y, b.name, xBright)
	}
	if b.timed {
		pg.g.put(r.Min.X+1, r.Min.Y+1, b.time, xText)
	}
}

// gathered lays the ranked bars of what the run has gathered out in the
// column of width w at row y, on a log10 scale to the next power of ten
// above the largest total. Each bar takes the rows the first screen has left
// for it, at most three: two thick and a blank row, one thick and a blank
// row, or one and no blank row.
func (pg *statsPage) gathered(st game.GameState, y, w, view int) int {
	pg.heads[1] = image.Rect(0, y, w, y+1)
	pg.heading(0, y, w, txtGatheredHd.String())
	y++
	var keys []string
	for k, v := range st.Stats.TotalGathered {
		if v > 0 {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := st.Stats.TotalGathered[keys[i]], st.Stats.TotalGathered[keys[j]]
		return a > b || a == b && keys[i] < keys[j]
	})
	if len(keys) == 0 {
		return y + 1
	}
	nameW, totW, rateW := 0, 9, 11
	for _, k := range keys {
		nameW = min(20, max(nameW, runeLen(game.ResourceName(k))))
	}
	trackX := nameW + 1
	trackW := w - trackX - 1 - totW - 1 - rateW
	if trackW < 12 {
		rateW, trackW = 0, w-trackX-1-totW
	}
	top := 1
	if mx := st.Stats.TotalGathered[keys[0]]; mx >= 1 {
		top = int(math.Floor(math.Log10(mx))) + 1
	}
	// the axis: a label at each decade, and the columns' names
	pg.ticks = pg.ticks[:0]
	last := -1
	for k := 0; k <= top; k++ {
		at := min(int(float64(k)/float64(top)*float64(trackW)), trackW-1)
		if k > 0 && k < top {
			pg.ticks = append(pg.ticks, float64(k)/float64(top))
		}
		if s := textfmt.Number(math.Pow10(k)); at > last+1 || k == 0 {
			if at+runeLen(s) <= trackW {
				pg.g.put(trackX+at, y, s, xDim)
				last = at + runeLen(s)
			}
		}
	}
	pg.g.rput(w-rateW-1, y, txtGathered.String(), xDim)
	if rateW > 0 {
		pg.g.rput(w-1, y, txtPerSecond.String(), xDim)
	}
	y++
	pitch := clamp((view-y)/len(keys), 1, 3)
	bh := max(1, pitch-1)
	gap := pitch - bh
	pg.bars = pg.bars[:0]
	for i, k := range keys {
		v := st.Stats.TotalGathered[k]
		b := statsBar{name: textfmt.Capitalize(game.ResourceName(k)), value: v, total: textfmt.Number(v)}
		if v >= 1 {
			b.frac = math.Log10(v) / float64(top)
		}
		b.readout = b.name + ": " + commas(v) + " " + txtGathered.String()
		rv, unit := perSecond(st.Resources[k].Rate, st.TickIntervalMs)
		b.rate = textfmt.Number(math.Abs(rv)) + unit
		rk := xDim
		switch {
		case rv < 0:
			b.rate, rk = "-"+b.rate, xBad
		case rv > 0:
			b.rate, rk = "+"+b.rate, xGood
		}
		b.track = image.Rect(trackX, y, trackX+trackW, y+bh)
		pg.g.put(0, y, fitText(b.name, nameW), xText)
		pg.g.rput(trackX+trackW+totW, y, b.total, xBright)
		if rateW > 0 {
			pg.g.rput(w-1, y, b.rate, rk)
		}
		pg.g.hit(hitBar, i, 0, y, w, bh)
		if pg.pic {
			frac, kind := b.frac, statsKinds[k]
			if kind == "" {
				kind = "storage"
			}
			ticks := pg.ticks
			pg.pics = append(pg.pics, statsPic{r: b.track, key: fmt.Sprint("bar", trackW, bh, frac, kind, ticks), hit: hitBar, n: i,
				draw: func(cw, ch int, hot bool, in statsInk) *image.RGBA {
					return statsBarPic(trackW, bh, cw, ch, frac, ticks, blockTone(kind, in), hot, in)
				}})
		} else {
			fill := int(b.frac*float64(trackW) + .5)
			for x := 0; x < trackW; x++ {
				ch, kk := '.', xFaint
				switch {
				case x < fill:
					ch, kk = '#', xAccent
				case isTick(x, trackW, pg.ticks):
					ch = ':'
				}
				pg.g.set(trackX+x, y, ch, kk)
			}
		}
		pg.bars = append(pg.bars, b)
		y += bh + gap
	}
	return y - gap
}

// isTick reports whether cell x of a track w cells long holds a decade tick.
func isTick(x, w int, ticks []float64) bool {
	for _, t := range ticks {
		if int(t*float64(w)) == x {
			return true
		}
	}
	return false
}

// multsTable lays the Active multipliers table out: a line of target, total
// and note, then the sources wrapped under it.
func (pg *statsPage) multsTable(x, y, w int) int {
	pg.heading(x, y, w, txtMultsHd.String())
	y++
	if len(pg.mults) == 0 {
		pg.g.put(x, y, txtNoMults.String(), xDim)
		return y + 1
	}
	for _, m := range pg.mults {
		k := xText
		switch {
		case m.pct > 0.5:
			k = xGood
		case m.pct < -0.5:
			k = xBad
		}
		pg.g.put(x, y, fitText(m.target, 22), xText)
		e := pg.g.put(x+30, y, fmt.Sprintf("%+.0f%%", m.pct), k)
		pg.g.put(e+1, y, fitText(m.note, x+w-e-1), xDim)
		y++
		cx := x + 2
		for i, s := range m.sources {
			txt := s.text()
			if i < len(m.sources)-1 {
				txt += ","
			}
			if cx+runeLen(txt) > x+w && cx > x+2 {
				cx, y = x+2, y+1
			}
			sk := xGood
			if s.neg {
				sk = xBad
			}
			cx = pg.g.put(cx, y, fitText(txt, x+w-cx), sk) + 1
		}
		y += 2
	}
	return y - 1
}

// rowsTable lays a table of label and value rows out under a heading.
func (pg *statsPage) rowsTable(head string, rows []statsRow, x, y, w int) int {
	pg.heading(x, y, w, head)
	y++
	lw := min(22, w/2)
	for _, r := range rows {
		vx, vw := x+lw+1, w-lw-1
		if r.label == "" {
			vx, vw = x, w
		}
		lines := wrapWords(r.value, vw)
		pg.g.put(x, y, fitText(r.label, lw), xDim)
		for i, s := range lines {
			pg.g.put(vx, y+i, s, r.k)
		}
		y += len(lines)
	}
	return y
}

// prestigeRows is the Prestige and epoch table's data: prestige, epoch,
// catastrophes, legacy bonuses and the events under way. A row with no
// label is a sentence.
func prestigeRows(st game.GameState) []statsRow {
	p := st.Prestige
	iv := tickInterval(st)
	rows := []statsRow{
		{vibe.Line("ui.stats_page.prestigeRows.01", "Level:"), textfmt.Int(p.Level), xBright},
		{vibe.Line("ui.stats_page.prestigeRows.02", "Prestige:"), vibe.Sprintf("ui.stats_page.prestigeRows.03", "x%.2f production", p.ProductionMultiplier), xGood},
	}
	if cue := st.PrestigeCue; cue.Ready {
		rows = append(rows, statsRow{"", vibe.Sprintf("ui.stats_page.prestigeRows.04", "A new run would pass this age in about %s; finishing it takes about %s at today's income.",
			game.DurationText(cue.FreshTicks, iv), game.DurationText(cue.FinishTicks, iv)), xDim})
	}
	if p.CanPrestige {
		rows = append(rows, statsRow{vibe.Line("ui.stats_page.prestigeRows.05", "After this prestige:"), vibe.Sprintf("ui.stats_page.prestigeRows.06", "x%.2f production", p.NextProductionMultiplier), xGood})
	} else if p.FloorBlocks && p.Floor != "" {
		floorName := st.Ruleset().Name(rules.KindAge, p.Floor)
		if !strings.HasPrefix(strings.ToLower(floorName), "the ") {
			floorName = "the " + floorName
		}
		rows = append(rows, statsRow{"", vibe.Sprintf("ui.stats_page.prestigeRows.07", "Prestige opens once you pass %s, the furthest an earlier run reached.", floorName), xGold})
	}
	rows = append(rows, statsRow{vibe.Line("ui.stats_page.prestigeRows.08", "Era Mastery:"), masteryNowText(st), xText})
	if next := masteryNextText(st); next != "" {
		rows = append(rows, statsRow{"", vibe.Sprintf("ui.stats_page.prestigeRows.09", "Next prestige: %s.", next), xDim})
	}
	rows = append(rows, statsRow{vibe.Line("ui.stats_page.prestigeRows.10", "Points earned:"), vibe.Sprintf("ui.stats_page.prestigeRows.11", "%d total", p.TotalEarned), xBright})
	switch {
	case p.CanPrestige:
		rows = append(rows, statsRow{"", vibe.Sprintf("ui.stats_page.prestigeRows.12", "Prestige now for %s.", textfmt.Count(p.PendingPoints, vibe.Line("ui.stats_page.prestigeRows.13", "point"), vibe.Line("ui.stats_page.prestigeRows.14", "points"))), xGood})
		if p.NextAge != "" {
			rows = append(rows, statsRow{"", vibe.Sprintf("ui.stats_page.prestigeRows.15", "From %s: %s.", ageRef(st, p.NextAge), textfmt.Count(p.NextAgePoints, vibe.Line("ui.stats_page.prestigeRows.16", "point"), vibe.Line("ui.stats_page.prestigeRows.17", "points"))), xDim})
		}
	case p.Level == 0:
		rows = append(rows, statsRow{"", vibe.Sprintf("ui.stats_page.prestigeRows.18", "Reach %s to prestige.", ageRef(st, game.PrestigeMinAge)), xDim})
	default:
		rows = append(rows, statsRow{"", vibe.Sprintf("ui.stats_page.prestigeRows.19", "Reach %s to prestige again.", ageRef(st, game.PrestigeMinAge)), xGold})
	}
	for _, line := range carryoverStatusLines(st) {
		rows = append(rows, statsRow{"", strings.TrimSpace(kit.Strip(line)), xText})
	}
	epoch := strings.TrimSpace(st.EpochIcon + " " + st.EpochName)
	if epoch == "" {
		epoch = vibe.Line("ui.stats_page.prestigeRows.20", "none")
	}
	rows = append(rows, statsRow{vibe.Line("ui.stats_page.prestigeRows.21", "Current epoch:"), epoch, xText},
		statsRow{vibe.Line("ui.stats_page.prestigeRows.22", "Catastrophes:"), vibe.Sprintf("ui.stats_page.prestigeRows.23", "%d  (endured %d, succumbed %d)",
			st.CatastrophesEndured+st.CatastrophesSuccumbed, st.CatastrophesEndured, st.CatastrophesSuccumbed), xText})
	if st.PendingCatastrophe != "" {
		rows = append(rows, statsRow{"", vibe.Line("ui.stats_page.prestigeRows.24", "One catastrophe is pending. Type catastrophe to choose."), xBad})
	} else if st.LastPassage.Pending {
		rows = append(rows, statsRow{"", vibe.Line("ui.stats_page.prestigeRows.25", "The Last Passage is pending. Type catastrophe to choose."), xBad})
	}
	set := st.Ruleset()
	rows = append(rows, statsRow{vibe.Line("ui.stats_page.prestigeRows.26", "Legacy bonuses:"), "", xGold})
	legacy := keysWhere(st.LegacyBonuses, func(_ string, on bool) bool { return on })
	for _, key := range legacy {
		label := key
		if def, ok := set.Era(key); ok {
			label = def.Name
		}
		rows = append(rows, statsRow{label + ":", kit.Strip(legacyBonusLine(st, set.LegacyBonus(key))), xText})
	}
	if f := st.SuccumbResearchFactor; f > 0 && f < 1 {
		rows = append(rows, statsRow{vibe.Line("ui.stats_page.prestigeRows.27", "Ancient Knowledge:"), vibe.Sprintf("ui.stats_page.prestigeRows.28", "research time %s (permanent, %s for each epoch succumbed in)", game.ResearchFactorText(f), game.ResearchFactorText(game.SuccumbResearchTimeFactor)), xText})
	}
	if st.LastPassage.CosmicLegacy {
		rows = append(rows, statsRow{vibe.Line("ui.stats_page.prestigeRows.29", "Cosmic Legacy:"), vibe.Sprintf("ui.stats_page.prestigeRows.30", "all production %s, counted in full after every other bonus (permanent, through every prestige)", textfmt.SignedPercent(game.CosmicLegacyProductionBonus)), xText})
	} else if len(legacy) == 0 && !(st.SuccumbResearchFactor > 0 && st.SuccumbResearchFactor < 1) {
		rows = append(rows, statsRow{"", vibe.Line("ui.stats_page.prestigeRows.31", "None"), xDim})
	}
	rows = append(rows, statsRow{vibe.Line("ui.stats_page.prestigeRows.32", "Active events"), "", xGold})
	if len(st.ActiveEvents) == 0 {
		return append(rows, statsRow{"", vibe.Line("ui.stats_page.prestigeRows.33", "No active events."), xDim})
	}
	for _, evt := range st.ActiveEvents {
		rows = append(rows, statsRow{"", vibe.Sprintf("ui.stats_page.prestigeRows.34", "%s (%s left)", evt.Name, formatTicks(evt.TicksLeft, st)), xGold})
		for _, eff := range evt.Effects {
			if plain, _ := effectMagnitude(eff); plain != "" {
				k := xGood
				if eff.Value < 0 {
					k = xBad
				}
				held := config.Effect{Type: eff.Type, Target: eff.Target, Value: eff.Value}
				rows = append(rows, statsRow{"", "  " + plain + kit.Strip(capTag(st, held, true, "-")), k})
			}
		}
	}
	return rows
}

// researchRows is the research bonuses, one text line a row.
func researchRows(st game.GameState) []statsRow {
	var rows []statsRow
	for _, l := range strings.Split(kit.Strip(researchSummaryLines(st)), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			rows = append(rows, statsRow{"", l, xText})
		}
	}
	return rows
}

// lifetime lays out the account's tiles: name, title and its progress,
// prestiges, civilizations started, the highest age and the badges earned.
func (pg *statsPage) lifetime(st game.GameState, w, grow int) int {
	as := st.AccountStats
	if as == nil {
		pg.g.put(0, 1, txtNoAccount.String(), xDim)
		return 3
	}
	age := vibe.Line("ui.stats_page.lifetime.01", "None")
	if as.HighestAge != "" {
		age = as.HighestAge
		if def, ok := st.Ruleset().Age(as.HighestAge); ok {
			age = def.Name
		}
	}
	bs := as.BadgeSummary
	title := statsTile{name: txtTitleOf.String(), text: bs.Title}
	if bs.NextTitle != "" {
		title.note = fmt.Sprintf(txtOfPoints.String(), bs.Points, bs.NextTitleAt, bs.NextTitle)
	}
	tiles := []statsTile{
		{name: txtAccount.String(), text: as.DisplayName}, title,
		countTile(txtPrestiges.String(), "", as.TotalPrestiges),
		countTile(txtCivs.String(), "", as.CivilizationsStarted),
		{name: txtHighestAge.String(), text: age},
		countTile(txtBadgesEarn.String(), "", bs.Earned),
	}
	pg.tiles = tiles
	n := 3
	if w < 90 {
		n = 2
	}
	return pg.layTiles(tiles, 1, w, n, 3+grow)
}

// statsPanel is the Stats overlay's primitive: the whole screen but the
// command bar. It leaves the keyboard with the command bar and takes only
// the keys that print nothing, and only while the bar is empty: Up, Down,
// PgUp and PgDn scroll and Tab switches the part. Esc closes it.
type statsPanel struct {
	barPanel
	kit.Base
	state     game.GameState
	part      int
	scroll    int
	hover     image.Point // the cell the pointer is over, on the page; (-1, -1) off it
	w, h      int
	chrome    *xgrid // the frame, tabs and keys, in screen cells
	page      *statsPage
	wide, pic bool
	cache     map[string]*image.RGBA

	// settings reads the glyph tier.
	settings func() mapSettings
}

func newStatsPanel() *statsPanel {
	p := &statsPanel{w: 100, h: 30, hover: image.Pt(-1, -1), cache: map[string]*image.RGBA{}}
	p.barPanel = newBarPanel(p.routeKey)
	return p
}

// open is the overlay's build: it runs on every Show.
func (p *statsPanel) open(state game.GameState) Primitive {
	p.state, p.part, p.scroll = state, partRun, 0
	return overBar(p)
}

func (p *statsPanel) update(state game.GameState) { p.state = state }

// viewRows is how many rows of the page show.
func (p *statsPanel) viewRows() int { return max(1, p.h-6) }

// Layout lays the frame and the page out in r.
func (p *statsPanel) Layout(r image.Rectangle) {
	p.Place(r)
	if r.Empty() {
		return
	}
	p.w, p.h = r.Dx(), r.Dy()
	plain := pictureCell == nil || p.settings != nil && p.settings().Tier == mapmodel.TierASCII
	p.pic = !plain
	p.wide = p.w >= statsWide
	bw, view := p.w-4, p.viewRows()
	for grow := 0; ; grow++ {
		p.page = buildStats(p.state, p.part, bw, view, p.wide, p.pic, grow)
		if grow >= 4 || p.page.used*5 >= view*4 {
			break
		}
	}
	p.scroll = clamp(p.scroll, 0, max(0, p.page.used-view))
	g := newXGrid(p.w, p.h)
	g.box(0, 0, p.w, p.h, xBorder, txtStatsTitle.String(), xTitle, false)
	g.closeButton(p.w)
	x := 2
	for i, name := range []string{txtThisRun.String(), txtLifetime.String()} {
		k := xDim
		if i == p.part {
			k = xSel
		}
		end := g.put(x, 1, " "+name+" ", k)
		g.tab(image.Rect(x, 1, end, 2), name, i, i == p.part)
		x = end + 1
	}
	expKeyBar(g, expLayout{w: p.w, h: p.h, footY: p.h - 2}, [][2]string{{"Up/Dn", "scroll"}, {"PgUp/PgDn", "page"},
		{"Tab", txtSwitchPart.String()}, escClose()}, "")
	if plain {
		g.fold()
		p.page.g.fold()
	}
	p.chrome = g
}

// readout is the line a hovered block or bar gives, and the page's heading
// row it is written on: the ages band's for a block, the bars' for a bar.
func (p *statsPanel) readout() (string, image.Rectangle) {
	if p.page == nil {
		return "", image.Rectangle{}
	}
	if n, ok := p.page.g.hitAt(hitAge, p.hover); ok && n < len(p.page.blocks) {
		return p.page.blocks[n].readout, p.page.heads[0]
	}
	if n, ok := p.page.g.hitAt(hitBar, p.hover); ok && n < len(p.page.bars) {
		return p.page.bars[n].readout, p.page.heads[1]
	}
	return "", image.Rectangle{}
}

// Paint puts the frame, then the page's visible rows with its pictures and
// the words over them, on the kit's painter.
func (p *statsPanel) Paint(p2 *kit.Painter) {
	r := p.Bounds()
	if r.Empty() || p.chrome == nil {
		return
	}
	pal := newExpPalette()
	paintExpeditionKit(p2, p.chrome, r.Min, pal)
	view := image.Rect(2, 3, p.w-2, p.h-3).Add(r.Min)
	bp := p2.Clip(view)
	org := image.Pt(view.Min.X, view.Min.Y-p.scroll)
	paintExpeditionKit(bp, p.page.g, org, pal)
	if p.pic {
		in := newStatsInk(pal)
		cw, ch := cellPixels()
		for _, pc := range p.page.pics {
			n, hot := p.page.g.hitAt(pc.hit, p.hover)
			hot = hot && pc.hit != 0 && n == pc.n
			key := fmt.Sprint(pc.key, hot, cw, ch, in.ground, in.face, in.plate)
			img := p.cache[key]
			if img == nil {
				if len(p.cache) > 600 {
					clear(p.cache)
				}
				img = pc.draw(cw, ch, hot, in)
				p.cache[key] = img
			}
			bp.Picture(pc.r.Add(org), img)
		}
		for _, l := range p.page.labels {
			col := pal.tone(l.k, false)
			if l.tint != "" {
				col = inkOn(blockTone(l.tint, in))
			}
			bp.Label(org.X+l.x, org.Y+l.y, l.s, kit.Style{}.Foreground(theme.RGB(col.R, col.G, col.B)))
		}
	}
	p2.Bar(r.Min.X+p.w-2, view.Min.Y, view.Dy(), p.scroll, view.Dy(), p.page.used)
	if s, head := p.readout(); s != "" {
		s = " " + s
		bp.Label(org.X+head.Max.X-runeLen(s), org.Y+head.Min.Y, s, pal.style(xcell{k: xBright}))
	}
	p.buttons.paint(p2, p.chrome, r.Min, pal)
}

// Mouse: a click on a part's tab switches to it, as Tab does; the wheel
// scrolls; the pointer over a block or a bar reads it out.
func (p *statsPanel) Mouse(m kit.Mouse) bool {
	switch {
	case p.buttons.mouse(p.chrome, p.Bounds().Min, m):
		return true
	case m.Kind == kit.MouseWheel:
		p.scrollBy(m.Wheel * 3)
		return m.Wheel != 0
	case m.Kind == kit.MouseLeave:
		p.hover = image.Pt(-1, -1)
		return false
	case m.Kind == kit.MouseMove:
		pt := m.Pt().Sub(p.Bounds().Min).Sub(image.Pt(2, 3))
		p.hover = image.Pt(pt.X, pt.Y+p.scroll)
		if pt.Y < 0 || pt.Y >= p.viewRows() {
			p.hover = image.Pt(-1, -1)
		}
		return false
	case m.Kind != kit.MousePress || m.Button != kit.ButtonLeft || p.chrome == nil:
		return false
	}
	if t, ok := p.chrome.tabAt(m.Pt().Sub(p.Bounds().Min)); ok {
		p.setPart(t)
		return true
	}
	return false
}

// PointerAt: the hand over a tab or a key.
func (p *statsPanel) PointerAt(at image.Point) kit.Pointer {
	if p.chrome == nil {
		return kit.PointerArrow
	}
	pt := at.Sub(p.Bounds().Min)
	_, tab := p.chrome.tabAt(pt)
	_, key := p.chrome.keyAt(pt)
	return handIf(tab || key)
}

// routeKey takes a key if it is one of the panel's. With something typed in
// the prompt every key belongs to the prompt.
func (p *statsPanel) routeKey(ev *kit.Key, prompt string) bool {
	if prompt != "" {
		return false
	}
	switch ev.Code {
	case kit.KeyUp:
		p.scrollBy(-1)
	case kit.KeyDown:
		p.scrollBy(1)
	case kit.KeyPgUp:
		p.scrollBy(-p.viewRows())
	case kit.KeyPgDn:
		p.scrollBy(p.viewRows())
	case kit.KeyTab:
		p.setPart(1 - p.part)
	default:
		return false
	}
	return true
}

func (p *statsPanel) scrollBy(n int) {
	top := 0
	if p.page != nil {
		top = max(0, p.page.used-p.viewRows())
	}
	p.scroll = clamp(p.scroll+n, 0, top)
}

func (p *statsPanel) setPart(t int) {
	if t != p.part {
		p.part, p.scroll = t, 0
		p.Layout(p.Bounds())
	}
}
