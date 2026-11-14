package kit

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// frame_test.go holds the kit to its frame budget: a screen as full as the
// dashboard lays out and paints in under 4 ms, with every text on it changed
// since the frame before (nothing cached).

// frameBudget is the time one layout and paint may take.
const frameBudget = 4 * time.Millisecond

// dashboardTree is a tree the size of the dashboard at 240 by 67 cells: a top
// bar, a resource bar of 24 cells, a ledger list, the map's hole, a tall
// wrapped text, a log and the command line, with a modal over it all. fresh
// rewrites every text, so the next paint lays each out again.
func dashboardTree() (root *Root, fresh func(n int)) {
	top := Label("")
	res := Row()
	var cells []*Text
	for i := 0; i < 24; i++ {
		c := Label("")
		cells = append(cells, c)
		res.Add(c, 0, 1)
	}
	ledger := NewList()
	side := NewText("")
	log := NewText("")
	logView := NewScroll(log)
	logView.Follow = true
	cmd := Label("")
	mid := Row().
		Add(NewPanel("[gold]Ledger[-]", ledger), 44, 0).
		Add(NewPanel("[gold]Map[-]", &Hole{}), 0, 1).
		Add(NewPanel("[gold]Advisor[-]", NewScroll(side)), 60, 0)
	base := Col().
		Add(top, 1, 0).
		Add(res, 2, 0).
		Add(mid, 0, 1).
		Add(NewPanel("[gold]Log[-]", logView), 10, 0).
		Add(cmd, 1, 0)
	root = NewRoot(base)
	root.Resize(240, 67)
	root.Push(&Modal{W: 60, H: 9, Child: NewPanel("[gold]Advance?[-]", Col().
		Add(NewText("[white]Leaving the [gold]Bronze Age[-] costs [red]1,200 bronze[-] and [red]400 tools[-]. The storehouses keep what they hold."), 0, 1).
		Add(Row().Add(NewButton("Advance", nil), 0, 0).Add(nil, 2, 0).Add(NewButton("Not yet", nil), 0, 0), 1, 0))})
	fresh = func(n int) {
		top.SetText(fmt.Sprintf("[gold::b]AgeForge[-:-:-]  [gray]Bronze Age[-]  [cyan]tick %d[-]  [green]+12.5/s[-]", n))
		for i, c := range cells {
			c.SetText(fmt.Sprintf("[cyan]res%02d[-]\n[white]%d[-] [green]+%d.%d[-]", i, n*i, i, n%10))
		}
		rows := make([]string, 60)
		for i := range rows {
			rows[i] = fmt.Sprintf("[white]Building %02d[-] [gray]x%d[-] [green]+%d.%d/s[-] [yellow]%d wood[-]", i, n+i, i, n%10, 100*i+n)
		}
		ledger.SetItems(rows)
		side.SetText(strings.Repeat(fmt.Sprintf("[white]The [gold]advisor[-] has a long thing to say about frame %d, and it [green]wraps[-] over a good many lines. ", n), 60))
		var l strings.Builder
		for i := 0; i < 200; i++ {
			fmt.Fprintf(&l, "[gray]%04d[-] [white]the mill made [green]%d[-] planks and the log grew a line\n", i, n+i)
		}
		log.SetText(l.String())
		cmd.SetText(fmt.Sprintf("[gold]>[-] build lumber_mill %d", n))
	}
	return root, fresh
}

func TestADashboardSizedFrameFitsTheBudget(t *testing.T) {
	if raceOn || testing.Short() {
		t.Skip("timing means nothing under the race detector")
	}
	root, fresh := dashboardTree()
	b := NewBuffer(240, 67)
	best := time.Hour
	for n := 0; n < 30; n++ {
		fresh(n)
		start := time.Now()
		root.Paint(b)
		best = min(best, time.Since(start))
	}
	t.Logf("best frame with every text new: %v", best)
	if best > frameBudget {
		t.Errorf("a dashboard-sized frame took %v, over the %v budget", best, frameBudget)
	}
	if got := b.Line(0); !strings.Contains(got, "AgeForge") {
		t.Errorf("the tree did not paint: %q", got)
	}
}

func BenchmarkDashboardFrame(b *testing.B) {
	root, fresh := dashboardTree()
	buf := NewBuffer(240, 67)
	fresh(0)
	b.Run("cached", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			root.Paint(buf)
		}
	})
	b.Run("every text new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			fresh(i)
			b.StartTimer()
			root.Paint(buf)
		}
	})
}
