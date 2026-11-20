package ui

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/logroll"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// logs_paint.go lays the Logs panel out (pkg/logroll plans the rects) and
// paints it: pictures from pkg/logroll, the words over them as labels. The
// only cells the panel draws as a grid are the key row at the foot.

const logsSeed = 7

var (
	txtLogsHeads = [...]rowid.Entry{rowid.E("ui.logs_panel.heads.01", "When"), rowid.E("ui.logs_panel.heads.02", "What happened"),
		rowid.E("ui.logs_panel.heads.03", "Paid"), rowid.E("ui.logs_panel.heads.04", "Gained"), rowid.E("ui.logs_panel.heads.05", "Times")}
	txtLogsHint    = rowid.E("ui.logs_panel.words.10", "click or type: logs trade, logs find stone, logs everything")
	txtLogsFindBox = rowid.E("ui.logs_panel.words.11", "Find: type a word")
	txtLogsFindKey = rowid.E("ui.logs_panel.words.12", "Find")
	txtLogsNewKey  = rowid.E("ui.logs_panel.words.13", "Newest")
	txtLogsEscKey  = rowid.E("ui.logs_panel.words.14", "Close")
	txtLogsDetail  = rowid.E("ui.logs_panel.words.15", "Detail")
	txtLogsShow    = [...]rowid.Entry{rowid.E("ui.logs_panel.show.01", "Show"), rowid.E("ui.logs_panel.show.02", "all"), rowid.E("ui.logs_panel.show.03", "none")}
	txtLogsNotes   = [3][]rowid.Entry{
		{rowid.E("ui.logs_panel.notes.01", "ages, wonders,"), rowid.E("ui.logs_panel.notes.02", "disasters")},
		{rowid.E("ui.logs_panel.notes.03", "adds builds,"), rowid.E("ui.logs_panel.notes.04", "tech, trade,"), rowid.E("ui.logs_panel.notes.05", "trouble")},
		{rowid.E("ui.logs_panel.notes.06", "adds each copy,"), rowid.E("ui.logs_panel.notes.07", "every run,"), rowid.E("ui.logs_panel.notes.08", "every save")},
	}
)

func (p *logsPanel) Layout(r image.Rectangle) {
	p.Place(r)
	if r.Empty() {
		return
	}
	p.lay = logroll.Plan(r.Dx(), r.Dy())
	if p.lay.Mode == logroll.Placeholder {
		g := newXGrid(r.Dx(), r.Dy())
		s := txtLogsSmall.String()
		g.put(max(0, (r.Dx()-runeLen(s))/2), r.Dy()/2, s, xDim)
		p.chrome = g
		return
	}
	g := newXGrid(p.lay.Keys.W, 1)
	x := g.keyButton(0, 0, "F", txtLogsFindKey.String())
	x = g.keyButton(x+1, 0, "N", txtLogsNewKey.String())
	g.keyButton(x+1, 0, "Esc", txtLogsEscKey.String())
	p.chrome = g
	p.clamp()
}

func style(c color.RGBA, bold bool) kit.Style {
	return kit.Style{}.Foreground(theme.RGB(c.R, c.G, c.B)).Bold(bold)
}

// fit cuts s to w cells with an ellipsis.
func fit(s string, w int) string {
	if r := []rune(s); len(r) > w {
		return string(r[:max(0, w-1)]) + "…"
	}
	return s
}

// pic is a picture kept until the size of a cell or its inputs change.
func (p *logsPanel) pic(mk func() *image.RGBA, parts ...any) *image.RGBA {
	cw, ch := cellPixels()
	key := fmt.Sprint(cw, ch, parts)
	if img := p.cache[key]; img != nil {
		return img
	}
	if len(p.cache) > 900 {
		clear(p.cache)
	}
	img := mk()
	p.cache[key] = img
	return img
}

func solid(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c.R, c.G, c.B, 255
	}
	return img
}

// officeFacts are each office's lines at the level and whether it wrote any
// in the last ten minutes: the ink of its tile and its red dot.
func (p *logsPanel) officeFacts() (n [11]int, wax [11]bool) {
	soon := p.now().Add(-10 * time.Minute).Unix()
	for _, e := range p.state.Log {
		if !e.Seam && e.Office >= 1 && e.Office <= 10 && e.Level <= p.level {
			n[e.Office]++
			wax[e.Office] = wax[e.Office] || e.At >= soon
		}
	}
	return n, wax
}

func (p *logsPanel) Paint(p2 *kit.Painter) {
	r := p.Bounds()
	if r.Empty() || p.chrome == nil {
		return
	}
	pal, org, l := newExpPalette(), r.Min, p.lay
	if l.Mode == logroll.Placeholder {
		paintExpeditionKit(p2, p.chrome, org, pal)
		return
	}
	cw, ch := cellPixels()
	cells := func(q logroll.Rect) image.Rectangle {
		return image.Rect(org.X+q.X, org.Y+q.Y, org.X+q.Right(), org.Y+q.Bottom())
	}
	say := func(x, y int, s string, c color.RGBA, bold bool) {
		p2.Label(org.X+x, org.Y+y, kit.Escape(s), style(c, bold))
	}
	p2.Picture(r, p.pic(func() *image.RGBA { return logroll.Desk(l, cw, ch, logsSeed) }, "desk", l.Key()))

	p2.Picture(cells(logroll.Rect{X: l.Title.X, Y: l.Title.Y + 1, W: 9, H: 1}), p.pic(func() *image.RGBA { return solid(9*cw, ch, logroll.Brass) }, "tab"))
	say(l.Title.X+2, 1, txtLogsTitle.String(), logroll.Light, true)
	say(l.Title.X+11, 1, p.state.AgeName, logroll.DeskInk, true)
	if l.Mode == logroll.Full {
		shown := 0
		for _, row := range p.rows {
			if row.kind == rowLine {
				shown++
			}
		}
		say(l.Title.X+27, 1, rowid.Sprintf("ui.logs_panel.words.08", "%s of %s lines kept, %d shown", commas(float64(p.kept())), commas(float64(game.MaxLogSize)), shown), logroll.Dim, false)
	}
	p2.Picture(cells(l.Find), p.pic(func() *image.RGBA { return solid(l.Find.W*cw, ch, logroll.Paper) }, "slip", l.Find))
	if p.find == "" && !p.typing {
		say(l.Find.X+1, l.Find.Y, fit(txtLogsFindBox.String(), l.Find.W-2), logroll.Dim, false)
	} else {
		cur := ""
		if p.typing {
			cur = "▌"
		}
		say(l.Find.X+1, l.Find.Y, fit(txtLogsFindKey.String()+": "+p.find+cur, l.Find.W-2), logroll.Ink, false)
	}
	say(l.Hint.X, l.Hint.Y, fit(txtLogsHint.String(), l.Hint.W), logroll.DeskInk, false)
	for i, q := range []logroll.Rect{l.Clock, l.Text, l.Paid, l.Gained, l.Times} {
		if q.W > 0 {
			say(q.X, l.Head.Y, txtLogsHeads[i].String(), logroll.Dim, true)
		}
	}

	counts, recent := p.officeFacts()
	if l.Mode == logroll.Full {
		for i, w := range []int{0, 7, 12} {
			say(l.Rack.X+1+w, l.Rack.Y, txtLogsShow[i].String(), logroll.DeskInk, i > 0)
		}
	}
	for i, q := range l.Stamps {
		o := i + 1
		ink := min(1, .2+float64(counts[o])/30)
		p2.Picture(cells(q), p.pic(func() *image.RGBA {
			return logroll.Tile(i, p.on[o], ink, recent[o], logsSeed, q.W*cw, q.H*ch)
		}, "tile", i, p.on[o], ink, recent[o], q))
		if l.Mode == logroll.Full {
			c := logroll.Ink
			if !p.on[o] {
				c = logroll.Dim
			}
			say(q.X+6, q.Y, fit(game.Office(o).Name(), q.W-8), c, p.on[o])
			say(q.X+6, q.Y+1, p.linesText(counts[o]), logroll.Dim, false)
		}
	}
	p2.Picture(cells(l.Lever), p.pic(func() *image.RGBA { return logroll.Lever(l, cw, ch, p.level-1) }, "lever", p.level, l.Key()))
	if l.Mode == logroll.Full {
		say(l.Lever.X+1, l.Lever.Y, txtLogsDetail.String(), logroll.DeskInk, true)
		for i, s := range l.Stops {
			ink, dim := logroll.DeskInk, logroll.DeskInk
			if i+1 == p.level {
				ink, dim = logroll.Ink, logroll.Dim
			}
			say(s.X+1, s.Y, fmt.Sprint(i+1), logroll.Light, true)
			say(s.X+4, s.Y, txtLogsLevels[i].String(), ink, i+1 == p.level)
			for j, n := range txtLogsNotes[i] {
				if j+1 < s.H {
					say(s.X+4, s.Y+1+j, n.String(), dim, false)
				}
			}
		}
	}

	p.paintRail(p2, cells(l.Rail), cw, ch)
	p.paintRoll(p2, org, cw, ch)
	if n := p.unseen(); n > 0 {
		q := p.plate()
		p2.Picture(cells(q), p.pic(func() *image.RGBA { return solid(q.W*cw, ch, logroll.Brass) }, "plate", q))
		say(q.X+1, q.Y, fit(rowid.Sprintf("ui.logs_panel.words.02", "%d new lines", n), q.W-2), logroll.Light, true)
	}
	p.paintSlate(p2, org, cw, ch)
	paintExpeditionKit(p2, p.chrome, p.keysOrigin(), pal)
	p.buttons.paint(p2, p.chrome, p.keysOrigin(), pal)
}

// paintRail draws the whole kept roll with its headlines, troubles and seams
// and the stretch on show.
func (p *logsPanel) paintRail(p2 *kit.Painter, at image.Rectangle, cw, ch int) {
	n, h := len(p.rows), p.height()
	img := p.pic(func() *image.RGBA {
		var marks []logroll.Mark
		for i, row := range p.rows {
			switch {
			case row.kind == rowSeam:
				marks = append(marks, logroll.Mark{Pos: float64(i) / float64(max(1, n-1)), Kind: logroll.Seam})
			case row.kind == rowLine && row.e.Trouble:
				marks = append(marks, logroll.Mark{Pos: float64(i) / float64(max(1, n-1)), Kind: logroll.Trouble})
			case row.kind == rowLine && row.e.Level == game.LevelHeadline:
				marks = append(marks, logroll.Mark{Pos: float64(i) / float64(max(1, n-1)), Kind: logroll.Headline})
			}
		}
		return logroll.Rail(at.Dx()*cw, at.Dy()*ch, marks, float64(p.top)/float64(max(1, n)), float64(p.top+h)/float64(max(1, n)), logsSeed)
	}, "rail", n, p.top, h, p.level, p.on, p.find, len(p.opened))
	p2.Picture(at, img)
}

// paintRoll draws the visible rows, the newest at the foot, empty paper above
// a roll that is shorter than the window.
func (p *logsPanel) paintRoll(p2 *kit.Painter, org image.Point, cw, ch int) {
	l, n, h := p.lay, len(p.rows), p.height()
	say := func(x, y int, s string, c color.RGBA, bold bool) {
		p2.Label(org.X+x, org.Y+y, kit.Escape(s), style(c, bold))
	}
	if n == 0 {
		msg := txtLogsEmpty.String()
		if p.find != "" {
			msg = txtLogsNone.String()
		}
		say(l.Text.X, l.Body.Y+h/2, msg, logroll.Dim, false)
		return
	}
	for i := p.top; i < min(n, p.top+h); i++ {
		row, y := p.rows[i], l.Body.Y+i-p.top+max(0, h-n)
		strip := func(kind logroll.RowKind) {
			p2.Picture(image.Rect(org.X+l.Body.X, org.Y+y, org.X+l.Body.Right(), org.Y+y+1),
				p.pic(func() *image.RGBA { return logroll.RowBack(kind, l.Body.W*cw, ch, logsSeed+uint64(i%5)) }, "row", kind, l.Body.W))
		}
		switch row.kind {
		case rowBand:
			strip(logroll.RowBand)
			say(l.Clock.X, y, row.text, logroll.Ink, true)
			say(l.Times.Right()-runeLen(row.tally), y, row.tally, logroll.Dim, false)
		case rowSeam:
			gap := (runeLen(row.text) + 2) * cw
			p2.Picture(image.Rect(org.X+l.Body.X, org.Y+y, org.X+l.Body.Right(), org.Y+y+1),
				p.pic(func() *image.RGBA { return logroll.SeamRow(l.Body.W*cw, ch, (l.Body.W*cw-gap)/2, gap, logsSeed) }, "seam", gap, l.Body.W))
			say(l.Body.X+(l.Body.W-runeLen(row.text))/2, y, row.text, logroll.Dim, false)
		case rowDetail:
			strip(logroll.RowOpen)
			say(l.Text.X+2, y, row.label, logroll.Dim, true)
			say(l.Text.X+10, y, fit(row.text, l.Text.W-12), logroll.Ink, false)
		case rowGo:
			strip(logroll.RowOpen)
			plate := func(x int, key string) int {
				w := runeLen(key) + 2
				p2.Picture(image.Rect(org.X+x, org.Y+y, org.X+x+w, org.Y+y+1), p.pic(func() *image.RGBA { return solid(w*cw, ch, logroll.Brass) }, "key", w))
				say(x+1, y, key, logroll.Light, true)
				return x + w + 1
			}
			x := plate(l.Text.X+2, "Enter")
			say(x, y, fmt.Sprintf(txtLogsGoTo.String(), game.Office(row.e.Office).Name()), logroll.Ink, false)
			plate(p.goSplit(row), "O")
			say(p.goSplit(row)+4, y, txtLogsOnly.String(), logroll.Ink, false)
		default:
			if row.kind == rowLine && row.e.Banner {
				strip(logroll.RowBanner)
				say(l.Body.X+(l.Body.W-runeLen(row.text))/2, y, row.text, logroll.NavyInk, true)
				break
			}
			p.paintLine(p2, org, y, i, row, strip, cw, ch)
		}
		if i == p.sel && !p.follow {
			say(l.Body.X, y, "▸", logroll.Brass, true)
		}
	}
}

func (p *logsPanel) paintLine(p2 *kit.Painter, org image.Point, y, i int, row logRow, strip func(logroll.RowKind), cw, ch int) {
	l, e := p.lay, row.e
	say := func(x int, s string, c color.RGBA, bold bool) {
		p2.Label(org.X+x, org.Y+y, kit.Escape(s), style(c, bold))
	}
	ink, bg, head := logroll.Ink, logroll.Paper, e.Level == game.LevelHeadline && row.kind == rowLine
	switch {
	case row.kind != rowLine:
		strip(logroll.RowOpen)
		ink, bg = logroll.Dim, logroll.Band
	case head:
		ink = logroll.Navy
	case e.Trouble:
		ink = logroll.Rust
	case e.Office == game.OfficeGame && e.Level == game.LevelEverything:
		ink = logroll.Refuse
	}
	clock := "--:--"
	if row.at > 0 {
		clock = time.Unix(row.at, 0).Format("15:04")
	}
	indent := 0
	if row.kind != rowLine {
		indent = 2
	}
	if row.kind != rowEarlier {
		say(l.Clock.X, clock, ink, false)
	}
	if row.kind == rowLine {
		seed := uint64(e.At)<<16 ^ uint64(e.Tick)
		stamp := l.Stamp
		p2.Picture(image.Rect(org.X+stamp.X, org.Y+y, org.X+stamp.Right(), org.Y+y+1), p.pic(func() *image.RGBA {
			ic := logroll.Inks[int(e.Office)-1]
			if head {
				ic = logroll.NavyInk
			}
			return logroll.Impression(int(e.Office)-1, seed, stamp.W*cw, ch, ic, bg)
		}, "imp", e.Office, seed, head, stamp.W))
	}
	text := fit(row.text, l.Text.W-indent)
	for _, h := range row.hits {
		if h[1] <= runeLen(text) {
			p2.Picture(image.Rect(org.X+l.Text.X+indent+h[0], org.Y+y, org.X+l.Text.X+indent+h[1], org.Y+y+1),
				p.pic(func() *image.RGBA { return solid((h[1]-h[0])*cw, ch, logroll.Hit) }, "hit", h[1]-h[0]))
		}
	}
	say(l.Text.X+indent, text, ink, head)
	paid, gained := p.amounts(row.paid, "-"), p.amounts(row.gained, "+")
	if l.Gained.W == 0 && paid != "" && gained != "" {
		paid += " " + gained
	} else if l.Gained.W == 0 && gained != "" {
		paid = gained
	} else {
		say(l.Gained.X, fit(gained, l.Gained.W), ink, false)
	}
	say(l.Paid.X, fit(paid, l.Paid.W), ink, false)
	if c := e.Count; row.kind == rowLine && c > 1 {
		say(l.Times.X, "×"+textfmt.Number(float64(c)), ink, true)
		if l.Strokes {
			x := l.Times.X + 6
			p2.Picture(image.Rect(org.X+x, org.Y+y, org.X+l.Times.Right(), org.Y+y+1), p.pic(func() *image.RGBA {
				return logroll.Tally(min(c, 15), uint64(e.At), (l.Times.Right()-x)*cw, ch, ink, bg)
			}, "tally", min(c, 15), e.At, ink, head))
		}
	}
	_ = i
}

// paintSlate chalks the town's state, then the sitting's tally at the foot.
func (p *logsPanel) paintSlate(p2 *kit.Painter, org image.Point, cw, ch int) {
	l, d := p.lay, p.slate()
	b := l.SlateBody
	say := func(x, y int, s string, c color.RGBA, bold bool) {
		p2.Label(org.X+x, org.Y+y, kit.Escape(s), style(c, bold))
	}
	y, foot := b.Y, b.Bottom()-(1+len(d.tallies))
	section := func(title string, lines []string, max int, c color.RGBA) {
		if len(lines) == 0 || y+2 > foot {
			return
		}
		say(b.X+1, y, title, logroll.ChalkY, true)
		y++
		for _, s := range lines[:min(len(lines), max)] {
			if y < foot {
				say(b.X+2, y, fit(s, b.W-3), c, false)
				y++
			}
		}
		y++
	}
	section(txtSlateUnder.String(), d.under, 4, logroll.Chalk)
	section(txtSlateTrouble.String(), d.trouble, slateTroubleMax, logroll.Hit)
	section(txtSlateComing.String(), d.coming, 3, logroll.Chalk)
	section(txtSlateStores.String(), d.stores, slateStoresMax, logroll.ChalkD)
	if foot-y >= 2 {
		at := image.Rect(org.X+b.X, org.Y+y, org.X+b.Right(), org.Y+foot)
		p2.Picture(at, p.pic(func() *image.RGBA { return logroll.SlateBare(at.Dx()*cw, at.Dy()*ch, logsSeed) }, "bare", at.Dx(), at.Dy()))
	}
	say(b.X+1, foot, txtSlateSitting.String(), logroll.ChalkY, true)
	for i, t := range d.tallies {
		ty := foot + 1 + i
		say(b.X+2, ty, fit(t.name, 16), logroll.Chalk, false)
		if t.text != "" || i == 7 {
			say(b.X+19, ty, fit(t.text, b.W-21), logroll.ChalkD, false)
			continue
		}
		p2.Picture(image.Rect(org.X+b.X+19, org.Y+ty, org.X+b.X+19+14, org.Y+ty+1), p.pic(func() *image.RGBA {
			return logroll.Tally(min(t.n, 15), uint64(i), 14*cw, ch, logroll.Chalk, logroll.Slate)
		}, "stk", min(t.n, 15), i))
		say(b.X+34, ty, textfmt.Number(float64(t.n)), logroll.ChalkD, false)
	}
}
