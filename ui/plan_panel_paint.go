package ui

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/planboard"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// plan_panel_paint.go builds what the Plan panel shows: pictures from
// pkg/planboard, words over them, and the key buttons. Everything is made in
// relayout, once per state or layout change; Paint only replays it.

// Words and labels. Placeholders: the wording is written elsewhere.
var (
	txtPlanSmall   = rowid.E("ui.plan_panel.words.01", "The plan needs a larger window.")
	txtPlanOptim   = rowid.E("ui.plan_panel.words.02", "Optimize")
	txtPlanPause   = rowid.E("ui.plan_panel.words.03", "Pause")
	txtPlanResume  = rowid.E("ui.plan_panel.words.04", "Resume")
	txtPlanClear   = rowid.E("ui.plan_panel.words.05", "Clear")
	txtPlanSure    = rowid.E("ui.plan_panel.words.06", "Clear again")
	txtPlanClose   = rowid.E("ui.plan_panel.words.07", "Close")
	txtPlanKeep    = rowid.E("ui.plan_panel.words.08", "Keep")
	txtPlanDrop    = rowid.E("ui.plan_panel.words.09", "Discard")
	txtPlanUp      = rowid.E("ui.plan_panel.words.10", "Up")
	txtPlanDown    = rowid.E("ui.plan_panel.words.11", "Down")
	txtPlanTop     = rowid.E("ui.plan_panel.words.12", "To top")
	txtPlanBottom  = rowid.E("ui.plan_panel.words.13", "To bottom")
	txtPlanRemove  = rowid.E("ui.plan_panel.words.14", "Remove")
	txtPlanCopies  = rowid.E("ui.plan_panel.words.15", "Copies")
	txtPlanFewer   = rowid.E("ui.plan_panel.words.16", "fewer")
	txtPlanMore    = rowid.E("ui.plan_panel.words.17", "more")
	txtPlanNumber  = rowid.E("ui.plan_panel.words.18", "type a number")
	txtPlanName    = rowid.E("ui.plan_panel.words.19", "type a name")
	txtPlanHint    = rowid.E("ui.plan_panel.words.20", "click or type; drag a row to change the order")
	txtPlanStarts  = rowid.E("ui.plan_panel.words.21", "Starts in")
	txtPlanEnds    = rowid.E("ui.plan_panel.words.22", "Ends in")
	txtPlanHolding = rowid.E("ui.plan_panel.words.23", "Holding back")
	txtPlanWhy     = rowid.E("ui.plan_panel.words.24", "Why it waits")
	txtPlanAbove   = rowid.E("ui.plan_panel.words.25", "Above:")
	txtPlanBelow   = rowid.E("ui.plan_panel.words.26", "Below:")
	txtPlanWaits   = rowid.E("ui.plan_panel.words.27", "What the plan waits on")
	txtPlanNoInc   = rowid.E("ui.plan_panel.words.28", "no income")
	txtPlanLater   = rowid.E("ui.plan_panel.words.29", "Later")
	txtPlanPlan    = rowid.E("ui.plan_panel.words.30", "Plan")
	txtPlanSplit   = rowid.E("ui.plan_panel.words.31", "split")
	txtPlanSlowest = rowid.E("ui.plan_panel.words.32", "slowest")
	txtPlanCols    = [5]rowid.Entry{
		rowid.E("ui.plan_panel.words.33", "Needs"), rowid.E("ui.plan_panel.words.34", "Have"),
		rowid.E("ui.plan_panel.words.35", "Held"), rowid.E("ui.plan_panel.words.36", "Income"),
		rowid.E("ui.plan_panel.words.37", "Ready in"),
	}
	txtPlanTabs = [6]rowid.Entry{
		rowid.E("ui.plan_panel.words.38", "Build"), rowid.E("ui.plan_panel.words.39", "Upgrade"),
		rowid.E("ui.plan_panel.words.40", "Research"), rowid.E("ui.plan_panel.words.41", "Advance"),
		rowid.E("ui.plan_panel.words.42", "Expedition"), rowid.E("ui.plan_panel.words.43", "Deal"),
	}
)

// planNoChange says why an optimize left the plan alone.
func planNoChange(reason string) string {
	if reason == game.PlanOptimizeNothing {
		return rowid.Line("ui.plan_panel.noChange.01", "No building in the plan grows the town.")
	}
	return rowid.Line("ui.plan_panel.noChange.02", "No faster order found.")
}

type planWord struct {
	x, y int
	s    string
	st   kit.Style
}

type planPic struct {
	r   image.Rectangle
	img *image.RGBA
}

type planSchedKey struct {
	L planboard.Layout
	S planboard.Sched
}

func planInk(c color.RGBA) kit.Color { return theme.RGB(c.R, c.G, c.B) }

func (p *planPanel) say(x, y int, s string, c color.RGBA, bold bool) {
	p.words = append(p.words, planWord{x, y, s, kit.Style{}.Foreground(planInk(c)).Bold(bold)})
}

// sayR writes s so that it ends just before column xr.
func (p *planPanel) sayR(xr, y int, s string, c color.RGBA) {
	p.say(xr-runeLen(s), y, s, c, false)
}

func (p *planPanel) pic(r image.Rectangle, img *image.RGBA) {
	p.pics = append(p.pics, planPic{r, img})
}

// btn is a key button at (x, y); it returns the next free column.
func (p *planPanel) btn(x, y int, key, name string) int {
	c := planboard.PInk
	if y == 2 { // the header sits on the dark ground
		c = planboard.Ink
	}
	end := p.keys.keyButton(x, y, key, name)
	p.say(x+runeLen(key)+3, y, name, c, false)
	return end + 1
}

// entries lists what the tray's tab offers, for the name typed so far.
func (p *planPanel) entries() {
	st := p.st
	var es []planEntry
	kind := planTabs[p.tab]
	switch kind {
	case game.PlanBuild:
		for _, k := range plannableBuildingKeys(st) {
			es = append(es, planEntry{kind, k, game.BuildingName(k), planPriceWhole(st.Buildings[k].NextCost), 0})
		}
	case game.PlanUpgrade:
		if p.engine == nil {
			break
		}
		cost := map[string]string{}
		for _, u := range p.engine.GetAvailableUpgrades() {
			if _, ok := cost[u.FromKey]; !ok {
				cost[u.FromKey] = planPriceWhole(u.Cost)
			}
		}
		for _, k := range upgradeableBuildingKeys(p.engine) {
			es = append(es, planEntry{kind, k, game.BuildingName(k), cost[k], 0})
		}
	case game.PlanResearch:
		for _, k := range plannableTechKeys(st) {
			es = append(es, planEntry{kind, k, st.Research.Techs[k].Name, "", 0})
		}
	case game.PlanAdvance:
		if st.NextAgeName != "" {
			es = append(es, planEntry{kind, "", st.NextAgeName, "", 0})
		}
	case game.PlanExpedition:
		for _, s := range st.Military.Board {
			if s.Offer != nil {
				es = append(es, planEntry{kind, s.Offer.Name, s.Offer.Name, "", 0})
			}
		}
	case game.PlanDeal:
		for _, civ := range sortedKeysOf(st.Diplomacy.Factions) {
			f := st.Diplomacy.Factions[civ]
			if f.DealsBlocked != "" {
				continue
			}
			for _, d := range f.Deals {
				if !d.Taken {
					label := fmt.Sprintf("%s: %s / %s", game.CivName(civ), game.Amount(d.GiveAmt, d.Give), game.Amount(d.GetAmt, d.Get))
					es = append(es, planEntry{kind, civ, label, "", d.Num})
				}
			}
		}
	}
	p.all = es
}

// filtered is the tray's list for the name typed.
func (p *planPanel) filtered() []planEntry {
	if p.mode != planNameBox || p.input == "" {
		return p.all
	}
	names := make([]string, len(p.all))
	for i, e := range p.all {
		names[i] = e.label
	}
	var out []planEntry
	for _, i := range matchOffers(names, p.input) {
		out = append(out, p.all[i])
	}
	return out
}

func (p *planPanel) kindWord(kind string) string {
	for i, k := range planTabs {
		if k == kind {
			return txtPlanTabs[i].String()
		}
	}
	return kind
}

// status is the line under an item's title: when it starts and what holds it.
func planStatus(v game.PlanItemView, e *game.PlanItemEstimate) string {
	s := planWhen(v, e)
	switch v.Cause {
	case game.PlanCauseFunds:
		s += rowid.Sprintf("ui.plan_panel.status.01", ", short of %s", game.ResourceName(v.Bottleneck))
	case game.PlanCauseIncome:
		s += rowid.Sprintf("ui.plan_panel.status.02", ", no %s coming in", game.ResourceName(v.Bottleneck))
	default:
		if v.Status == game.PlanStatusBlocked && v.Note != "" {
			s = v.Note
		}
	}
	return s
}

// rows is what the schedule draws: the plan, or the optimize preview.
func (p *planPanel) rows() (rows []planboard.Row, titles, status []string) {
	if p.mode == planPreview {
		for j, r := range p.opt.Rows {
			if r.From < 1 || r.From > len(p.st.Plan) {
				continue
			}
			v := p.st.Plan[r.From-1]
			row := planboard.Row{Seed: v.Kind + v.Key, Sign: v.Kind, Mode: planboard.Timed, Start: r.Start, End: r.End, Gate: v.Kind == game.PlanAdvance}
			if r.Never {
				row.Mode = planboard.Parked
			}
			if o := r.From - 1; o < len(p.old) && p.old[o].Modeled && !p.old[o].Never {
				row.Old = &planboard.Span{Start: p.old[o].Start, End: p.old[o].End}
			}
			t := planItemTitle(v)
			if j < len(p.opt.Order) {
				t = p.opt.Order[j]
			}
			tag := planTimeText(r.Start)
			switch d := r.From - 1 - j; {
			case d > 0:
				tag += rowid.Sprintf("ui.plan_panel.rows.01", ", up %d", d)
			case d < 0:
				tag += rowid.Sprintf("ui.plan_panel.rows.02", ", down %d", -d)
			}
			for k, o := range p.opt.Rows {
				if k != j && o.From == r.From {
					tag += ", " + txtPlanSplit.String()
					break
				}
			}
			rows, titles, status = append(rows, row), append(titles, t), append(status, tag)
		}
		return
	}
	for i, v := range p.st.Plan {
		var e *game.PlanItemEstimate
		if len(p.est.Items) == len(p.st.Plan) {
			e = &p.est.Items[i]
		}
		row := planboard.Row{Seed: v.Kind + v.Key, Sign: v.Kind, Mode: planboard.Parked, Cause: v.Cause, Gate: v.Kind == game.PlanAdvance}
		switch {
		case v.Cause == game.PlanCauseAge:
			row.Mode = planboard.After
		case e != nil && e.Modeled && !e.Never && v.When == game.PlanWhenTimed:
			row.Mode, row.Start, row.End, row.Copies = planboard.Timed, e.Start, e.End, e.Copies
		}
		rows, titles, status = append(rows, row), append(titles, planItemTitle(v)), append(status, planStatus(v, e))
	}
	return
}

// relayout rebuilds every picture, word and key for the current state.
func (p *planPanel) relayout() {
	r := p.Bounds()
	p.words, p.pics, p.keys = nil, nil, nil
	if r.Empty() {
		return
	}
	cw, ch := cellPixels()
	p.l, p.ok = planboard.NewLayout(r.Dx(), r.Dy(), cw, ch)
	p.keys = newXGrid(r.Dx(), r.Dy())
	if !p.ok {
		p.say(2, 1, txtPlanSmall.String(), planboard.Ink, false)
		return
	}
	l := p.l
	p.sel = clamp(p.sel, 0, max(0, p.n()-1))
	rows, titles, status := p.rows()
	var adv []int
	for i, rw := range rows {
		if rw.Gate {
			adv = append(adv, i)
		}
	}
	p.first = clamp(p.first, 0, max(0, len(rows)-1))
	if p.sel < p.first {
		p.first = p.sel
	}
	for range rows {
		slots, _ := l.Slots(p.first, len(rows), adv)
		if p.sel < p.first || len(slots) == 0 || slots[len(slots)-1].Item >= p.sel || p.first >= p.sel {
			break
		}
		p.first++
	}
	drop := -1
	if p.drag.on {
		drop = p.drag.to
	}
	finish := ""
	if p.est.Finish > 0 {
		finish = rowid.Sprintf("ui.plan_panel.finish.01", "Finishes in %s", planTimeText(p.est.Finish))
	}
	untimed := ""
	if p.est.Untimed > 0 {
		untimed = rowid.Sprintf("ui.plan_panel.untimed.01", "%d not timed", p.est.Untimed)
	}
	s := planboard.Sched{Rows: rows, First: p.first, Sel: p.sel, Drop: drop, Paused: p.st.PlanPaused}
	if finish != "" {
		s.FinishW = runeLen(finish) + 2
		if untimed != "" {
			s.UntimedW = runeLen(untimed) + 2
		}
	}
	p.pic(image.Rect(0, 0, r.Dx(), r.Dy()), p.boardM.get(l, l.Board))
	p.drawn = p.schedM.get(planSchedKey{l, s}, func() planboard.Drawn { return l.Schedule(s) })
	p.pic(p.drawn.Rect, p.drawn.Img)

	// Title block and header.
	n := len(p.st.Plan)
	p.say(4, 2, txtPlanPlan.String(), planboard.Ink, true)
	p.say(14, 2, rowid.Sprintf("ui.plan_panel.count.01", "%d of %d", n, game.MaxPlanItems), planboard.Dim, false)
	p.say(4, 3, p.st.AgeName, planboard.Dim, false)
	x := l.Lanes.Min.X + 1
	if finish != "" {
		p.say(x+1, 2, finish, planboard.Ink, false)
		p.say(x+s.FinishW+2, 2, untimed, planboard.Wait, false)
	}
	p.header()
	for _, t := range p.drawn.Ticks {
		p.say(t.Cell, 4, t.Name, planboard.Dim, false)
	}
	if p.drawn.GateX >= 0 {
		p.say(p.drawn.GateX+1, 3, p.st.NextAgeName, planboard.Brass, true)
	}
	if p.drawn.SelX >= 0 && p.sel < len(rows) && rows[p.sel].Mode == planboard.Timed {
		p.say(p.drawn.SelX, 3, planTimeText(rows[p.sel].Start), planboard.Ink, false)
	}

	// The list.
	hdr := []string{strings.ToUpper(p.st.AgeName), strings.ToUpper(p.st.NextAgeName), txtPlanLater.String()}
	for _, sl := range p.drawn.Slots {
		if sl.Item < 0 {
			p.say(l.ListX+1, sl.Y, hdr[min(sl.Age, 2)], planboard.Brass, true)
			continue
		}
		p.say(l.ListX+1, sl.Y, fmt.Sprint(sl.Item+1), planboard.Dim, false)
		p.say(l.ListX+7, sl.Y, fitText(titles[sl.Item], l.ListW-12), planboard.Ink, true)
		if l.RowH > 1 {
			p.say(l.ListX+7, sl.Y+1, fitText(status[sl.Item], l.ListW-12), planboard.Dim, false)
		}
	}
	if p.drawn.More > 0 {
		p.say(l.ListX+1, l.SchedEnd-1, rowid.Sprintf("ui.plan_panel.more.01", "%d more below", p.drawn.More), planboard.Dim, false)
	}

	p.right(adv)
	p.waits()
	p.tray()
	hint, c := txtPlanHint.String(), planboard.HintInk
	if p.note != "" {
		hint = p.note
		if !p.good {
			c = planboard.Bad
		}
	}
	p.say(3, l.Hint, hint, c, p.note != "")
}

// header is the buttons at the top right.
func (p *planPanel) header() {
	type b struct{ key, name string }
	bs := []b{{"O", txtPlanOptim.String()}, {"P", txtPlanPause.String()}, {"C", txtPlanClear.String()}, {"Esc", txtPlanClose.String()}}
	if p.st.PlanPaused {
		bs[1].name = txtPlanResume.String()
	}
	if p.armed {
		bs[2].name = txtPlanSure.String()
	}
	if p.mode == planPreview {
		bs = []b{{"Enter", txtPlanKeep.String()}, {"Esc", txtPlanDrop.String()}}
	}
	w := 0
	for _, x := range bs {
		w += runeLen(x.key) + runeLen(x.name) + 5
	}
	x := p.l.Inner.Max.X - w
	for _, y := range bs {
		x = p.btn(x, 2, y.key, y.name)
	}
}

// right is the job sheet and the gate card.
func (p *planPanel) right(adv []int) {
	l := p.l
	v, has := p.view()
	gateAt := -1
	if has && v.Kind == game.PlanAdvance {
		gateAt = p.sel
	} else if len(adv) > 0 {
		gateAt = adv[0]
	}
	jobH := min(34, l.Right.Dy()-14)
	job, gate := l.Cards(jobH)
	if has && !(l.Compact && gateAt == p.sel) {
		p.job(v, job)
	}
	if gateAt >= 0 && gate.Dy() > 4 && (!l.Compact || gateAt == p.sel) {
		p.gate(p.st.Plan[gateAt], gate)
	}
}

func (p *planPanel) job(v game.PlanItemView, r image.Rectangle) {
	l := p.l
	var e *game.PlanItemEstimate
	if len(p.est.Items) == len(p.st.Plan) {
		e = &p.est.Items[p.sel]
	}
	tight := l.Compact || r.Dy() < 32
	row := func(full, small int) int { return r.Min.Y + map[bool]int{false: full, true: small}[tight] }
	step, gx, gy := 2, r.Min.X+57, row(14, 10)
	cols := [5]int{r.Min.X + 18, r.Min.X + 26, r.Min.X + 34, r.Min.X + 44, r.Min.X + 55}
	if tight {
		step = 1
	}
	if l.Compact {
		gx, cols = r.Min.X+38, [5]int{r.Min.X + 17, r.Min.X + 25, 0, 0, r.Min.X + 36}
	}
	jx := r.Min.X + 2
	p.say(r.Min.X+10, r.Min.Y+2, fitText(planItemTitle(v), r.Dx()-12), planboard.PInk, true)
	p.say(r.Min.X+10, r.Min.Y+3, p.kindWord(v.Kind), planboard.PDim, false)
	p.say(r.Min.X+10, r.Min.Y+4+map[bool]int{false: 1, true: 0}[tight], fitText(planStatus(v, e), r.Dx()-12), planboard.PInk, false)
	if v.Kind == game.PlanBuild || v.Kind == game.PlanUpgrade {
		y := row(7, 6)
		p.say(jx, y, txtPlanCopies.String(), planboard.PDim, false)
		cnt := fmt.Sprint(v.Count)
		if p.mode == planCountBox {
			cnt = p.input + "_"
		}
		p.say(jx+8, y, cnt, planboard.PInk, true)
		x := p.btn(jx+14, y, "-", txtPlanFewer.String())
		x = p.btn(x, y, "+", txtPlanMore.String())
		if !l.Compact {
			p.btn(x, y, "N", txtPlanNumber.String())
		}
	}
	at, until := planWhen(v, e), ""
	if e != nil && e.Modeled && !e.Never {
		until = planTimeText(e.End)
	}
	p.say(jx, row(9, 7), txtPlanStarts.String(), planboard.PDim, false)
	p.say(jx+12, row(9, 7), at, planboard.PInk, false)
	p.say(jx, row(10, 8), txtPlanEnds.String(), planboard.PDim, false)
	p.say(jx+12, row(10, 8), until, planboard.PInk, false)
	hy := row(12, 9)
	for i, c := range cols {
		if c > 0 {
			p.sayR(c+1, hy, txtPlanCols[i].String(), planboard.PDim)
		}
	}
	tail := 9
	if tight {
		tail = 6
	}
	var bars []planboard.Bar
	for i, rr := range v.Rows {
		if gy+i*step+tail > r.Max.Y {
			break
		}
		y, nm := gy+i*step, game.ResourceName(rr.Res)
		if rr.Res == v.Bottleneck {
			nm = "*" + nm
		}
		p.say(jx, y, fitText(nm, 11), planboard.PInk, false)
		vals := [5]string{textfmt.Number(rr.Need), textfmt.Number(rr.Have), textfmt.Number(rr.Held), textfmt.Number(rr.PerSecond) + "/s", ""}
		switch {
		case rr.Never:
			vals[4] = txtPlanNever.String()
		case rr.ReadyIn <= 0:
			vals[4] = txtPlanNow.String()
		default:
			vals[4] = textfmt.Duration(rr.ReadyIn)
		}
		for k, c := range cols {
			if c > 0 {
				p.sayR(c+1, y, vals[k], planboard.PInk)
			}
		}
		bars = append(bars, planboard.Bar{Have: rr.Have, Need: rr.Need, Held: rr.Held, Short: rr.Have < rr.Need})
	}
	y := gy + len(bars)*step + 1
	line := func(label, s string) {
		if y < r.Max.Y-2 && s != "" {
			p.say(jx, y, label, planboard.PDim, false)
			p.say(jx+len([]rune(label))+1, y, fitText(s, r.Dx()-len([]rune(label))-5), planboard.PInk, false)
			y++
		}
	}
	line(txtPlanHolding.String(), game.Amounts(v.HeldBack))
	if v.Cause != "" {
		line(txtPlanWhy.String(), v.Note)
		if v.Bottleneck != "" {
			line("", "* "+txtPlanSlowest.String()+": "+game.ResourceName(v.Bottleneck))
		}
	}
	if v.Count > 1 || v.Started > 0 {
		line("", rowid.Sprintf("ui.plan_panel.job.01", "Started %d of %d copies", v.Started, v.Count))
	}
	line("", rowid.Sprintf("ui.plan_panel.job.02", "Place %d of %d", p.sel+1, len(p.st.Plan)))
	if p.sel > 0 {
		line(txtPlanAbove.String(), planItemTitle(p.st.Plan[p.sel-1]))
	}
	if p.sel+1 < len(p.st.Plan) {
		line(txtPlanBelow.String(), planItemTitle(p.st.Plan[p.sel+1]))
	}
	x := jx
	x = p.btn(x, r.Max.Y-2, "U", txtPlanUp.String())
	x = p.btn(x, r.Max.Y-2, "D", txtPlanDown.String())
	if !l.Compact {
		x = p.btn(x, r.Max.Y-2, "T", txtPlanTop.String())
		x = p.btn(x, r.Max.Y-2, "B", txtPlanBottom.String())
	}
	p.btn(x, r.Max.Y-2, "X", txtPlanRemove.String())
	key := struct {
		R    image.Rectangle
		K    string
		B    []planboard.Bar
		G, Y int
		S    int
	}{r, v.Kind + v.Key, bars, gx, gy, step}
	p.pic(r, p.picM[0].get(key, func() *image.RGBA { return l.Slip(r, v.Kind, gx, gy, step, bars) }))
}

func (p *planPanel) gate(v game.PlanItemView, r image.Rectangle) {
	l := p.l
	name := fitText(p.st.NextAgeName, 20)
	p.say(r.Min.X+(r.Dx()-runeLen(name))/2, r.Min.Y-1, name, planboard.Ink, true)
	p.say(r.Min.X+2, r.Min.Y+1, v.Note, planboard.PInk, false)
	var bars []planboard.Bar
	for i, pl := range v.Pillars {
		y := r.Min.Y + 3 + 2*i
		if y+1 >= r.Max.Y {
			break
		}
		p.say(r.Min.X+1, y, fitText(pl.Name, 11), planboard.PInk, false)
		t := fmt.Sprintf("%d%%", pl.Percent)
		if pl.Blocker != "" {
			t += " " + pl.Blocker
		}
		p.say(r.Min.X+34, y, fitText(t, r.Dx()-35), planboard.PInk, false)
		bars = append(bars, planboard.Bar{Have: float64(pl.Percent), Need: 100, Short: !pl.Done})
	}
	key := struct {
		R image.Rectangle
		B []planboard.Bar
	}{r, bars}
	p.pic(image.Rect(r.Min.X, r.Min.Y-3, r.Max.X, r.Max.Y), p.picM[1].get(key, func() *image.RGBA {
		return l.Gate(r, r.Min.X+13, 20, r.Min.Y+3, 2, bars)
	}))
}

// waits is the box of what the plan waits on.
func (p *planPanel) waits() {
	l := p.l
	w := l.Waits
	p.say(w.Min.X+2, w.Min.Y, txtPlanWaits.String(), planboard.Ink, true)
	gw := 30
	if l.Compact {
		gw = 14
	}
	gx := w.Min.X + 12
	var rows []planboard.WaitRow
	for i, nd := range p.est.Needs {
		y := w.Min.Y + 2 + i*l.RowH
		if y+l.RowH > w.Max.Y {
			break
		}
		p.say(w.Min.X+1, y, fitText(game.ResourceName(nd.Res), 10), planboard.Ink, false)
		x := gx + gw + 1
		p.say(x, y, rowid.Sprintf("ui.plan_panel.waits.01", "%s of %s", textfmt.Number(nd.Have), textfmt.Number(nd.Need)), planboard.Dim, false)
		x += 17
		switch {
		case nd.Never:
			p.say(x, y, txtPlanNoInc.String(), planboard.Bad, false)
		case nd.CoverIn > 0 && !l.Compact:
			p.say(x, y, rowid.Sprintf("ui.plan_panel.waits.02", "covered in %s", textfmt.Duration(nd.CoverIn)), planboard.Dim, false)
		}
		if nd.Waiting > 0 {
			p.sayR(w.Max.X, y, rowid.Sprintf("ui.plan_panel.waits.03", "%d wait", nd.Waiting), planboard.Wait)
		}
		rows = append(rows, planboard.WaitRow{Have: nd.Have, Need: nd.Need})
	}
	p.pic(w, p.picM[2].get(struct {
		R    image.Rectangle
		W    []planboard.WaitRow
		X, G int
	}{w, rows, gx, gw}, func() *image.RGBA { return l.WaitsPic(w, gx, gw, w.Min.Y+2, l.RowH, rows) }))
}

// tray is the tabs, the search row and the entries.
func (p *planPanel) tray() {
	l := p.l
	t := l.Tray
	p.shown = p.filtered()
	per := l.EntryCount()
	p.trayOff = clamp(p.trayOff, 0, max(0, (len(p.shown)-per+1)/2*2))
	var es []planboard.Entry
	for k := 0; k < per && p.trayOff+k < len(p.shown); k++ {
		es = append(es, planboard.Entry{Sign: p.shown[p.trayOff+k].kind})
	}
	for i := range planTabs {
		tr := l.TabRect(i)
		c := planboard.Dim
		if i == p.tab {
			c = planboard.PInk
		} else {
			c = planboard.PDim
		}
		p.say(tr.Min.X+2, tr.Min.Y, fitText(txtPlanTabs[i].String(), l.TabW-4), c, i == p.tab)
	}
	label := txtPlanName.String() + "..."
	if p.mode == planNameBox {
		label = p.input + "_"
	}
	p.say(t.Min.X+2, t.Min.Y+1, label, planboard.PDim, false)
	p.btn(t.Max.X-runeLen(txtPlanName.String())-8, t.Min.Y+1, "A", txtPlanName.String())
	for k := range es {
		er, _ := l.EntryRect(k)
		e := p.shown[p.trayOff+k]
		off := 4
		if l.Compact {
			off = 2
		}
		p.say(er.Min.X+off, er.Min.Y, fitText(e.label, er.Dx()-off-4-map[bool]int{false: 25, true: 0}[l.Compact]), planboard.PInk, false)
		if !l.Compact && e.price != "" {
			p.say(er.Min.X+26, er.Min.Y, fitText(strings.ReplaceAll(e.price, glue, " "), er.Dx()-31), planboard.PDim, false)
		}
	}
	signs := make([]string, len(planTabs))
	copy(signs, planTabs[:])
	p.pic(t, p.picM[3].get(struct {
		R image.Rectangle
		T int
		E []planboard.Entry
	}{t, p.tab, es}, func() *image.RGBA { return l.TrayPic(t, signs, p.tab, es) }))
}

// Layout places the panel and rebuilds its pictures for the new size.
func (p *planPanel) Layout(r image.Rectangle) {
	p.Place(r)
	p.relayout()
}

// Paint replays what relayout made.
func (p *planPanel) Paint(pt *kit.Painter) {
	r := p.Bounds()
	if r.Empty() {
		return
	}
	if !p.ok {
		pt.Fill(r, planInk(planboard.Dark))
	}
	for _, q := range p.pics {
		pt.Picture(q.r.Add(r.Min), q.img)
	}
	for _, w := range p.words {
		pt.Label(r.Min.X+w.x, r.Min.Y+w.y, w.s, w.st)
	}
	if p.keys != nil {
		p.buttons.paint(pt, p.keys, r.Min, newExpPalette())
	}
}
