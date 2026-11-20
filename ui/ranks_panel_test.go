package ui

import (
	"image"
	"image/png"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
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
	m := buildRanks(rules.Core(), views)
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
		if m.doms[i-1].rank == 0 && m.doms[i].rank > 0 {
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

// TestRanksPicture writes the screen at the window's size when RANKS_SHOT
// names a file. The account is made up from the real rule set: about 212
// badges earned, the lowest tiers first, each domain's share by its size,
// Energy none, Hacker and Astronaut a few.
func TestRanksPicture(t *testing.T) {
	out := os.Getenv("RANKS_SHOT")
	if out == "" {
		t.Skip("RANKS_SHOT not set")
	}
	byDom := map[string][]config.BadgeDef{}
	var all []config.BadgeDef
	for _, d := range rules.Core().Badges() {
		if d.Integrity() {
			continue
		}
		byDom[d.Domain()] = append(byDom[d.Domain()], d)
		all = append(all, d)
	}
	want := map[string]int{"energy": 0, "hacker": 2, "astronaut": 4}
	left := 212
	for k, n := range want {
		left -= n
		_ = k
	}
	size := 0
	for k, l := range byDom {
		if _, ok := want[k]; !ok {
			size += len(l)
		}
	}
	for k, l := range byDom {
		if _, ok := want[k]; !ok {
			want[k] = min(left*len(l)/size, len(l))
		}
	}
	now := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	got := map[string]bool{}
	i := 0
	for k, l := range byDom {
		l := append([]config.BadgeDef(nil), l...)
		sort.SliceStable(l, func(a, b int) bool { return l[a].Tier < l[b].Tier })
		for j := 0; j < want[k]; j++ {
			got[l[j].Key] = true
			i++
		}
	}
	var views []game.BadgeView
	n := 0
	for _, d := range all {
		v := game.BadgeView{Key: d.Key, Name: d.Name, Desc: d.Desc, Tier: d.Tier.Name(), Points: d.Points(),
			Level: d.Tier, Domain: d.Domain(), Earned: got[d.Key]}
		if v.Earned {
			n++
			v.At = now.Add(-time.Duration(n*9) * time.Hour)
		} else if d.Threshold > 0 {
			v.Target = d.Threshold
			v.Progress = d.Threshold * float64((len(d.Key)*7)%90+5) / 100
		}
		views = append(views, v)
	}
	p := ranksPanelOn(views, 279, 74)
	p.model.now = now
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, ScreenImagePixel(paintAlone(p, 279, 74))); err != nil {
		t.Fatal(err)
	}
}

// TestRanksKeysAreButtons: the screen's typed keys are buttons; a click on
// one presses its key.
func TestRanksKeysAreButtons(t *testing.T) {
	p := ranksPanelOn(ranksViews(2), 120, 37)
	paintAlone(p, 120, 37)
	var got []kit.KeyCode
	p.buttons.press = func(k kit.Key) { got = append(got, k.Code) }
	for _, want := range []kit.KeyCode{kit.KeyEnter, kit.KeyTab, kit.KeyEsc} {
		hit := false
		for _, k := range p.keys.keyset().keys {
			if k.key.Code == want && !k.r.Empty() {
				p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: k.r.Min.X, Y: k.r.Min.Y, Clicks: 1})
				hit = true
				break
			}
		}
		if !hit {
			t.Errorf("no button for key %v", want)
		}
	}
	if len(got) != 3 {
		t.Errorf("clicks pressed %v, want Enter, Tab and Esc", got)
	}
}
