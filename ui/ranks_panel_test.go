package ui

import (
	"image"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// ranksViews is a made-up account's badges: every badge in the rule set, the
// first earnedEvery-th earned, counts read from config and never fixed here.
func ranksViews(earnedEvery int) []game.BadgeView {
	var out []game.BadgeView
	day := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	for i, d := range config.Badges() {
		if d.Integrity() {
			continue
		}
		v := game.BadgeView{Key: d.Key, Name: d.Name, Tier: d.Tier.Name(), Points: d.Points(),
			Level: d.Tier, Domain: d.Domain(), Earned: i%earnedEvery == 0}
		if v.Earned {
			v.At = day.Add(time.Duration(i) * time.Hour)
		}
		out = append(out, v)
	}
	return out
}

func ranksPanelOn(views []game.BadgeView, w, h int) *ranksPanel {
	p := newRanksPanel()
	p.update(game.GameState{AccountStats: &game.AccountStatsView{Badges: views}})
	p.Layout(image.Rect(0, 0, w, h))
	return p
}

func TestRanksModelCountsFromConfig(t *testing.T) {
	views := ranksViews(3)
	m := buildRanks(views)
	var want, earned int
	for _, v := range views {
		want++
		if v.Earned {
			earned++
		}
	}
	if m.total != want || m.earned != earned {
		t.Fatalf("totals %d/%d, want %d/%d", m.earned, m.total, earned, want)
	}
	var sum int
	for _, d := range m.doms {
		sum += d.pts
		if d.rank != config.RankFor(d.pts) {
			t.Errorf("%s: rank %d for %d points", d.key, d.rank, d.pts)
		}
	}
	if sum != m.pts || len(m.doms) != len(config.WorkerDomains()) {
		t.Errorf("points %d (sum %d), domains %d", m.pts, sum, len(m.doms))
	}
	for i := 1; i < len(m.doms); i++ {
		if m.doms[i-1].rank < m.doms[i].rank {
			t.Fatalf("domains not ranked first: %s before %s", m.doms[i-1].key, m.doms[i].key)
		}
	}
}

func TestRanksKeysAndTabs(t *testing.T) {
	p := ranksPanelOn(ranksViews(2), 120, 37)
	first := p.view.dom
	p.handleKey(kit.KeyDown)
	if p.view.dom == first {
		t.Error("Down kept the domain")
	}
	p.handleKey(kit.KeyEnter)
	if p.view.tab != tabAll || !p.view.list {
		t.Fatalf("Enter: tab %d, list %v", p.view.tab, p.view.list)
	}
	for _, v := range p.rows {
		if v.Domain != p.view.dom {
			t.Fatalf("the domain's list holds %s of %q", v.Key, v.Domain)
		}
	}
	if !p.closeBack() || p.view.tab != tabRanks || p.closeBack() {
		t.Error("Esc should step back to Ranks once, then close")
	}
	p.handleKey(kit.KeyTab)
	p.handleKey(kit.KeyTab)
	if p.view.tab != tabRecent {
		t.Fatalf("two Tabs: tab %d", p.view.tab)
	}
	for i := 1; i < len(p.rows); i++ {
		if p.rows[i].At.After(p.rows[i-1].At) {
			t.Fatal("recent is not latest first")
		}
	}
	p.handleKey(kit.KeyBacktab)
	if p.view.tab != tabAll {
		t.Errorf("Backtab: tab %d", p.view.tab)
	}
	if p.routeKey(&kit.Key{Code: kit.KeyTab}, "adv") {
		t.Error("Tab with a typed prompt must stay with the prompt")
	}
}

func TestRanksMouse(t *testing.T) {
	p := ranksPanelOn(ranksViews(2), 120, 37)
	paintAlone(p, 120, 37)
	if len(p.hits) < 2 {
		t.Fatalf("no rows to click: %d", len(p.hits))
	}
	h := p.hits[1]
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: h.r.Min.X, Y: h.r.Min.Y, Clicks: 1})
	if p.view.dom != h.id {
		t.Errorf("a click picked %q, want %q", p.view.dom, h.id)
	}
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: h.r.Min.X, Y: h.r.Min.Y, Clicks: 2})
	if p.view.tab != tabAll || !p.view.list {
		t.Error("a double click should open the domain's list")
	}
	t1 := p.tabs[tabRecent]
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: t1.Min.X, Y: t1.Min.Y, Clicks: 1})
	if p.view.tab != tabRecent {
		t.Errorf("a click on the Recent tab left tab %d", p.view.tab)
	}
}

func TestRanksPaintsAtTheWindowSize(t *testing.T) {
	for _, tab := range []int{tabRanks, tabAll, tabRecent} {
		p := ranksPanelOn(ranksViews(2), 279, 74)
		p.setTab(tab)
		b := paintAlone(p, 279, 74)
		if b.W != 279 || b.H != 74 {
			t.Fatal("size")
		}
	}
}
