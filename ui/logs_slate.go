package ui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/window/kit"
)

// logs_slate.go is what the Logs panel's slate says: the town as it stands
// (what is under way, what is wrong, what comes next, the stores) and a tally
// of the sitting. slate() reads the state and nothing else; logs_paint.go
// chalks it.

var (
	txtSlateUnder   = rowid.E("ui.logs_slate.words.01", "Under way")
	txtSlateTrouble = rowid.E("ui.logs_slate.words.02", "Trouble")
	txtSlateComing  = rowid.E("ui.logs_slate.words.03", "Coming up")
	txtSlateStores  = rowid.E("ui.logs_slate.words.04", "Stores")
	txtSlateSitting = rowid.E("ui.logs_slate.words.05", "This sitting")
	txtSlateSince   = rowid.E("ui.logs_slate.words.07", "since %s, %s")
	txtSlateNow     = rowid.E("ui.logs_slate.words.08", "Right now")
	txtSlateTally   = [...]rowid.Entry{rowid.E("ui.logs_slate.tally.01", "Built"), rowid.E("ui.logs_slate.tally.02", "Techs"),
		rowid.E("ui.logs_slate.tally.03", "Milestones"), rowid.E("ui.logs_slate.tally.04", "Deals taken"),
		rowid.E("ui.logs_slate.tally.05", "Parties back"), rowid.E("ui.logs_slate.tally.06", "Workers gained"),
		rowid.E("ui.logs_slate.tally.07", "Gold spent"), rowid.E("ui.logs_slate.tally.08", "Biggest build")}
)

const (
	slateTroubleMax = 3
	slateStoresMax  = 12
	slateFoodWarn   = time.Hour
)

// slateKind says how a slate row is chalked.
type slateKind int

const (
	srText  slateKind = iota // text, with right-aligned text
	srBar                    // text, a bar at frac, right-aligned text
	srWarn                   // a warning mark and text
	srStore                  // text, an arrow at dir, right-aligned text
)

type slateRow struct {
	kind        slateKind
	text, right string
	frac        float64
	dir         int
}

// slateTally is one of "This sitting": a name and its figure.
type slateTally struct{ name, n string }

type slateData struct {
	under, trouble, coming []slateRow
	stores                 []slateRow
	tallies                []slateTally // two columns of four, column-major
	since                  string
}

var txtSlateBuilding = rowid.E("ui.logs_slate.words.06", "%s \u00b7 %d building")

// slate reads the state the panel holds.
func (p *logsPanel) slate() slateData {
	st, d := p.state, slateData{}
	type group struct {
		name    string
		n, soon int
		total   int
	}
	var gs []*group
	for _, b := range st.BuildQueue {
		i := slices.IndexFunc(gs, func(g *group) bool { return g.name == b.Name })
		if i < 0 {
			gs = append(gs, &group{name: b.Name, soon: b.TicksLeft, total: b.TotalTicks})
			i = len(gs) - 1
		}
		g := gs[i]
		g.n++
		if b.TicksLeft < g.soon {
			g.soon, g.total = b.TicksLeft, b.TotalTicks
		}
	}
	for _, g := range gs {
		r := slateRow{kind: srBar, text: g.name, right: formatTicks(g.soon, st)}
		if g.total > 0 {
			r.frac = 1 - float64(g.soon)/float64(g.total)
		}
		if g.n > 1 {
			r.text = fmt.Sprintf(txtSlateBuilding.String(), g.name, g.n)
		}
		d.under = append(d.under, r)
	}
	if st.Research.CurrentTech != "" {
		d.under = append(d.under, slateRow{kind: srBar, text: st.Research.CurrentTechName, right: researchLeft(st), frac: researchFraction(st)})
	}
	if f, ok := st.Resources["food"]; ok && f.Rate < 0 {
		if left := time.Duration(f.Amount/-f.Rate) * tickInterval(st); left < slateFoodWarn {
			d.trouble = append(d.trouble, slateRow{kind: srWarn, text: rowid.Sprintf("ui.logs_slate.food.01", "Food gone in %s", textfmt.Duration(left))})
		}
	}
	sitting := p.sitting()
	for i := len(sitting) - 1; i >= 0 && len(d.trouble) < slateTroubleMax; i-- {
		if e := sitting[i]; e.Trouble && !slices.ContainsFunc(d.trouble, func(r slateRow) bool { return r.text == e.Message }) {
			d.trouble = append(d.trouble, slateRow{kind: srWarn, text: kit.Strip(e.Message)})
		}
	}
	for _, s := range queueLines(st) {
		d.coming = append(d.coming, slateRow{text: kit.Strip(s)})
	}

	type store struct {
		name string
		rate float64
		row  slateRow
		fall bool
	}
	var ss []store
	for k, r := range st.Resources {
		if !r.Unlocked {
			continue
		}
		s := store{name: p.resName(k), rate: r.Rate, fall: r.Rate < 0}
		s.row = slateRow{kind: srStore, text: s.name + "  " + commas(r.Amount), right: rateText(r.Rate, st.TickIntervalMs), dir: 0}
		switch {
		case s.fall:
			s.row.dir = -1
			s.row.right += rowid.Sprintf("ui.logs_slate.gone.01", ", gone in %s", textfmt.Duration(time.Duration(r.Amount/-r.Rate)*tickInterval(st)))
		case r.Rate > 0:
			s.row.dir = 1
		}
		ss = append(ss, s)
	}
	slices.SortFunc(ss, func(a, b store) int {
		switch {
		case a.fall != b.fall:
			if a.fall {
				return -1
			}
			return 1
		case a.rate != b.rate:
			if a.rate > b.rate {
				return -1
			}
			return 1
		}
		return strings.Compare(a.name, b.name)
	})
	for _, s := range ss[:min(slateStoresMax, len(ss))] {
		d.stores = append(d.stores, s.row)
	}

	var n [7]int
	var big float64
	bigName := ""
	for _, e := range sitting {
		c := max(1, e.Count)
		if tot := paidTotal(e); e.Office == game.OfficeBuilding && e.Subject != "" && tot > big {
			big, bigName = tot, p.state.Ruleset().Name(rules.KindBuilding, e.Subject)
		}
		switch {
		case e.Office == game.OfficeBuilding && e.Type == "success":
			n[0] += c
		case e.Office == game.OfficeResearch && e.Type == "success":
			n[1] += c
		case e.Office == game.OfficeWorld && e.Type == "success" && e.Level == game.LevelHeadline && e.Subject != "":
			n[2] += c
		case e.Office == game.OfficeTrade && len(e.Paid) > 0 && len(e.Gained) > 0:
			n[3] += c
		case e.Office == game.OfficeWorld && len(e.Gained) > 0:
			n[4] += c
		}
		for _, a := range e.Gained {
			if a.Res == "workers" {
				n[5] += int(a.N)
			}
		}
		for _, a := range e.Paid {
			if a.Res == "gold" {
				n[6] += int(a.N)
			}
		}
	}
	for i := range n {
		d.tallies = append(d.tallies, slateTally{txtSlateTally[i].String(), commas(float64(n[i]))})
	}
	d.tallies = append(d.tallies, slateTally{txtSlateTally[7].String(), bigName})
	if len(sitting) > 0 {
		at := time.Unix(sitting[0].At, 0)
		d.since = fmt.Sprintf(txtSlateSince.String(), at.Format("15:04"), textfmt.Duration(time.Since(at)))
	}
	return d
}

func paidTotal(e game.LogEntry) (t float64) {
	for _, a := range e.Paid {
		t += a.N
	}
	return t
}

// sitting is the lines written since the last seam.
func (p *logsPanel) sitting() []game.LogEntry {
	log := p.state.Log
	i := len(log)
	for i > 0 && !log[i-1].Seam {
		i--
	}
	return log[i:]
}
