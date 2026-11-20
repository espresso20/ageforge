package ui

import (
	"slices"
	"strings"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/rules"
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

// slateTally is one line of "This sitting": a count drawn as strokes, or for
// the biggest build its name.
type slateTally struct {
	name string
	n    int
	text string
}

type slateData struct {
	under, trouble, coming []string
	stores                 []string
	tallies                []slateTally
}

// slate reads the state the panel holds.
func (p *logsPanel) slate() slateData {
	st, d := p.state, slateData{}
	for _, b := range st.BuildQueue {
		d.under = append(d.under, b.Name+"  "+formatTicks(b.TicksLeft, st))
	}
	if st.Research.CurrentTech != "" {
		d.under = append(d.under, rowid.Sprintf("ui.logs_slate.research.01", "%s  %d%%, %s", st.Research.CurrentTechName, researchPercent(st), researchLeft(st)))
	}
	if f, ok := st.Resources["food"]; ok && f.Rate < 0 {
		if left := time.Duration(f.Amount/-f.Rate) * tickInterval(st); left < slateFoodWarn {
			d.trouble = append(d.trouble, rowid.Sprintf("ui.logs_slate.food.01", "Food gone in %s", textfmt.Duration(left)))
		}
	}
	sitting := p.sitting()
	for i := len(sitting) - 1; i >= 0 && len(d.trouble) < slateTroubleMax; i-- {
		if e := sitting[i]; e.Trouble && !slices.Contains(d.trouble, e.Message) {
			d.trouble = append(d.trouble, e.Message)
		}
	}
	d.coming = queueLines(st)

	type store struct {
		name string
		rate float64
		text string
		fall bool
	}
	var ss []store
	for k, r := range st.Resources {
		if !r.Unlocked {
			continue
		}
		s := store{name: p.resName(k), rate: r.Rate, fall: r.Rate < 0}
		s.text = s.name + "  " + rateText(r.Rate, st.TickIntervalMs)
		if s.fall {
			s.text += rowid.Sprintf("ui.logs_slate.gone.01", ", gone in %s", textfmt.Duration(time.Duration(r.Amount/-r.Rate)*tickInterval(st)))
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
		d.stores = append(d.stores, s.text)
	}

	var n [8]int
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
		case e.Office == game.OfficeGame && e.Level == game.LevelHeadline:
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
		d.tallies = append(d.tallies, slateTally{name: txtSlateTally[i].String(), n: n[i]})
	}
	d.tallies[7] = slateTally{name: txtSlateTally[7].String(), text: bigName}
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
