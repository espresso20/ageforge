package ui

import (
	"fmt"
	"image"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// lifeAccount is an account that has reached the first nAges ages, with a
// fastest time for each but the first, the badges of every family with the
// first earn(i, n) of family i's n badges earned, and records set in runs
// of its own. Titles and points follow the badges.
func lifeAccount(ages []config.AgeDef, nAges int, earn func(fam, n int) int) *game.AccountStatsView {
	at := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	av := &game.AccountStatsView{DisplayName: "Ada Lovelace", TotalPrestiges: 3 * nAges / 2, CivilizationsStarted: 4 + nAges}
	if nAges <= 1 {
		av.TotalPrestiges, av.CivilizationsStarted = 0, 1
	}
	if nAges > 0 {
		av.HighestAge = ages[nAges-1].Key
		av.FastestAge = map[string]game.AgeRecord{}
		for i := 1; i < nAges; i++ {
			av.FastestAge[ages[i].Key] = game.AgeRecord{Ticks: 900 + 700*i*i, At: at.AddDate(0, 0, i), Run: "Rome"}
		}
		if nAges > 1 {
			av.BestRun = map[string]game.RunBest{
				game.RunBestBuilt:       {Value: 40 * nAges, At: at, Run: "Rome"},
				game.RunBestRecruited:   {Value: 21 * nAges, At: at.AddDate(0, 0, 3), Run: "Memphis"},
				game.RunBestResearched:  {Value: 11 * nAges, At: at.AddDate(0, 0, 5), Run: "Rome"},
				game.RunBestExpeditions: {Value: 2 + nAges, At: at.AddDate(0, 0, 9), Run: "Ur"},
			}
		}
	}
	tiers := []config.BadgeTier{config.BadgeBronze, config.BadgeBronze, config.BadgeSilver, config.BadgeGold, config.BadgePlatinum, config.BadgeLegendary}
	var sum game.BadgeSummary
	for fi, f := range badgeFamilyTitles {
		have := earn(fi, len(tiers))
		for i, tier := range tiers {
			v := game.BadgeView{
				Key: fmt.Sprintf("%s.%d", f.key, i), Family: f.key, Name: fmt.Sprintf("%s %d", f.key, i),
				Level: tier, Tier: tier.Name(), Points: tier.Points(), Earned: i < have,
			}
			if v.Earned {
				v.At, v.Run = at, "Rome"
				sum.Earned++
				sum.Points += v.Points
			}
			sum.Shown++
			av.Badges = append(av.Badges, v)
		}
	}
	sum.Title, sum.NextTitle, sum.NextTitleAt = "Settler", "Headman", max(sum.Points+60, 100)
	av.BadgeSummary = sum
	return av
}

// lifeStats is a Lifetime part laid out w by h for an account that reached
// the first nAges ages; edit may change the state before it is laid out.
func lifeStats(t *testing.T, nAges, w, h int, pics bool, earn func(fam, n int) int, edit func(*game.GameState)) (*statsPanel, game.GameState) {
	t.Helper()
	st := statsFixture(t, nAges, 12, 5*time.Hour)
	st.AccountStats = lifeAccount(config.Ages(), nAges, earn)
	if edit != nil {
		edit(&st)
	}
	p := laidStats(t, st, w, h, pics)
	keyOf(p, kit.KeyTab)
	p.Layout(p.Bounds())
	return p, st
}

func lifeNone(fam, n int) int { return 0 }
func lifeSome(fam, n int) int { return fam % (n + 1) }

// clickStep clicks the middle of step i.
func clickStep(p *statsPanel, i int) {
	r := p.page.steps[i].r
	pt := image.Pt(r.Min.X+r.Dx()/2, r.Min.Y+1)
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft,
		X: p.Bounds().Min.X + 2 + pt.X, Y: p.Bounds().Min.Y + 3 + pt.Y - p.scroll})
}

func TestLifetimeLadderLightsReachedAges(t *testing.T) {
	const reached = 8
	p, _ := lifeStats(t, reached, 160, 50, true, lifeSome, nil)
	if len(p.page.steps) != len(config.Ages()) {
		t.Fatalf("steps: %d, want %d", len(p.page.steps), len(config.Ages()))
	}
	lit := 0
	for i, s := range p.page.steps {
		if s.reached != (i < reached) {
			t.Errorf("step %d lit %v", i, s.reached)
		}
		if s.reached {
			lit++
			continue
		}
		if s.readout != "" || s.time != "" {
			t.Errorf("unreached step %d names itself: %q %q", i, s.readout, s.time)
		}
		for _, l := range p.page.labels {
			if image.Pt(l.x, l.y).In(image.Rect(s.r.Min.X, s.r.Min.Y, s.r.Max.X, s.r.Max.Y+2)) {
				t.Errorf("unreached step %d paints %q", i, l.s)
			}
		}
	}
	if lit != reached || p.page.top != reached-1 {
		t.Errorf("lit %d, top %d; want %d, %d", lit, p.page.top, reached, reached-1)
	}
	// without pictures the crown is the one caret, over the highest step
	pl, _ := lifeStats(t, reached, 160, 50, false, lifeSome, nil)
	var carets []image.Point
	for y := 0; y < pl.page.g.h; y++ {
		for x := 0; x < pl.page.g.w; x++ {
			if pl.page.g.c[y*pl.page.g.w+x].r == '^' {
				carets = append(carets, image.Pt(x, y))
			}
		}
	}
	top := pl.page.steps[reached-1]
	if len(carets) != 1 || carets[0].Y != top.r.Min.Y || !carets[0].In(image.Rect(top.r.Min.X, top.r.Min.Y, top.r.Max.X, top.r.Min.Y+1)) {
		t.Errorf("carets %v, want one over step %v", carets, top.r)
	}
}

func TestLifetimePointerMovesAndStops(t *testing.T) {
	p, st := lifeStats(t, 8, 160, 50, true, lifeSome, nil)
	ages := config.Ages()
	if p.picked() != 7 {
		t.Fatalf("starts on %d, want the highest", p.picked())
	}
	for i := 0; i < 12; i++ {
		keyOf(p, kit.KeyLeft)
	}
	if p.picked() != 0 {
		t.Fatalf("Left stops at %d, want 0", p.picked())
	}
	keyOf(p, kit.KeyRight)
	if p.picked() != 1 {
		t.Fatalf("Right: %d", p.picked())
	}
	for i := 0; i < 30; i++ {
		keyOf(p, kit.KeyRight)
	}
	if p.picked() != 7 {
		t.Fatalf("Right stops at %d, want 7", p.picked())
	}
	clickStep(p, 3)
	if p.picked() != 3 {
		t.Fatalf("click: %d, want 3", p.picked())
	}
	clickStep(p, 15) // not reached: no step to take it to
	if p.picked() != 3 {
		t.Fatalf("click on an unreached step moved the pointer to %d", p.picked())
	}
	rec := st.AccountStats.FastestAge[ages[3].Key]
	for _, want := range []string{ages[3].Name, formatTicks(rec.Ticks, st)} {
		buf := paintAlone(overBar(p), 160, 50)
		row := screenRow(buf, 3+p.page.heads[0].Min.Y)
		if !strings.Contains(row, want) {
			t.Errorf("heading row %q lacks %q", row, want)
		}
	}
	// the first age has no time, so the readout is its name alone
	clickStep(p, 0)
	if got, _ := p.readout(); got != ages[0].Name {
		t.Errorf("readout %q, want %q", got, ages[0].Name)
	}
}

func TestLifetimeRecordsShowTheLargerOfRecordAndRun(t *testing.T) {
	const n = 8
	p, _ := lifeStats(t, n, 160, 50, true, lifeSome, func(st *game.GameState) {
		st.Stats.TotalBuilt = 40*n + 5     // beats its record
		st.Stats.TotalRecruited = 21 * n   // ties it
		st.Research.TotalResearched = 3    // far below
		st.Military.CompletedCount = 2 + n // ties
	})
	want := map[string]struct {
		text string
		hold bool
	}{
		txtBuilt.String():     {commas(float64(40*n + 5)), true},
		txtRecruited.String(): {commas(float64(21 * n)), true},
		txtTechs.String():     {commas(float64(11 * n)), false},
		txtMissions.String():  {commas(float64(2 + n)), true},
	}
	seen := 0
	for _, tl := range p.page.tiles {
		w, ok := want[tl.name]
		if !ok {
			continue
		}
		seen++
		if tl.text != w.text || tl.hold != w.hold {
			t.Errorf("%s: shows %q hold %v, want %q %v", tl.name, tl.text, tl.hold, w.text, w.hold)
		}
		if w.hold != (tl.note == txtThisRun.String()) {
			t.Errorf("%s: note %q", tl.name, tl.note)
		}
		if !w.hold && tl.note != "Rome 4 Oct 2026" {
			t.Errorf("%s: note %q names no holder", tl.name, tl.note)
		}
	}
	if seen != len(want) {
		t.Errorf("found %d record tiles, want %d", seen, len(want))
	}
}

func TestLifetimeFamilyCoinsFollowTheBestTier(t *testing.T) {
	earn := func(fam, n int) int { return []int{0, 1, 2, 3, 4, 5, 6}[fam%7] }
	p, st := lifeStats(t, 8, 160, 80, true, earn, nil)
	as := st.AccountStats
	byKey := map[string]*statsPic{}
	for i := range p.page.pics {
		byKey[p.page.pics[i].key] = &p.page.pics[i]
	}
	for fi, f := range badgeFamilyTitles {
		var best config.BadgeTier
		got, have, total := false, 0, 0
		for _, v := range as.Badges {
			if v.Family != f.key {
				continue
			}
			total++
			if v.Earned {
				have++
				if !got || v.Level > best {
					best, got = v.Level, true
				}
			}
		}
		if have != earn(fi, 6) {
			t.Fatalf("fixture: family %s has %d", f.key, have)
		}
		pic := byKey[fmt.Sprint("coin", f.key, best, got)]
		if pic == nil {
			t.Errorf("family %s: no picture for tier %v, earned %v", f.key, best, got)
			continue
		}
		count := fmt.Sprintf("%d / %d", have, total)
		lines := gridLines(p.page.g)
		if !strings.Contains(lines[pic.r.Max.Y], f.title.String()) {
			t.Errorf("family %s: title %q not under its coin: %q", f.key, f.title.String(), lines[pic.r.Max.Y])
		}
		if !strings.Contains(lines[pic.r.Max.Y+1], count) {
			t.Errorf("family %s: count %q not under its title: %q", f.key, count, lines[pic.r.Max.Y+1])
		}
	}
	// a fresh account has every family as a recess
	q, _ := lifeStats(t, 1, 160, 80, true, lifeNone, nil)
	for _, pic := range q.page.pics {
		if strings.HasPrefix(pic.key, "coin") && !strings.HasSuffix(pic.key, "false") {
			t.Errorf("new account paints a coin: %s", pic.key)
		}
	}
}

func TestLifetimePaintsAreStable(t *testing.T) {
	p, _ := lifeStats(t, 12, 160, 50, true, lifeSome, nil)
	a := paintAlone(overBar(p), 160, 50)
	b := paintAlone(overBar(p), 160, 50)
	if !reflect.DeepEqual(a.Cells, b.Cells) || len(a.Pics) != len(b.Pics) {
		t.Fatal("two paints differ in their cells")
	}
	for i := range a.Pics {
		if a.Pics[i].Rect != b.Pics[i].Rect || !reflect.DeepEqual(a.Pics[i].Img.Pix, b.Pics[i].Img.Pix) {
			t.Fatalf("picture %d differs between paints", i)
		}
	}
}

func TestLifetimePlainTierIsASCII(t *testing.T) {
	p, _ := lifeStats(t, 12, 160, 50, false, lifeSome, nil)
	b := paintAlone(overBar(p), 160, 50)
	if len(b.Pics) != 0 {
		t.Fatalf("%d pictures in the plain tier", len(b.Pics))
	}
	for i, c := range b.Cells {
		if c.R > 127 {
			t.Fatalf("cell %d,%d holds %q", i%b.W, i/b.W, c.R)
		}
	}
	if txt := statsPageText(p.state, partLifetime, 160); !strings.Contains(txt, "[") {
		t.Error("plain families carry no marks")
	}
}

func TestLifetimeNothingOverlaps(t *testing.T) {
	for _, w := range []int{100, 160} {
		for _, n := range []int{1, 8, len(config.Ages())} {
			p, _ := lifeStats(t, n, w, 60, true, lifeSome, nil)
			var rects []image.Rectangle
			for _, tl := range p.page.tiles {
				rects = append(rects, tl.r)
			}
			for _, s := range p.page.steps {
				rects = append(rects, s.r)
			}
			for _, pc := range p.page.pics {
				if strings.HasPrefix(pc.key, "coin") {
					rects = append(rects, pc.r)
				}
			}
			for i := range rects {
				if rects[i].Max.X > w-4 {
					t.Errorf("w %d, %d ages: %v leaves the page", w, n, rects[i])
				}
				for j := i + 1; j < len(rects); j++ {
					if rects[i].Overlaps(rects[j]) {
						t.Errorf("w %d, %d ages: %v overlaps %v", w, n, rects[i], rects[j])
					}
				}
			}
		}
	}
}

// TestLifetimeStubIsGone: the old stub's painter and its word are not in the
// page file; the tab is laid out by stats_lifetime.go alone.
func TestLifetimeStubIsGone(t *testing.T) {
	src, err := os.ReadFile("stats_page.go")
	must(t, err)
	for _, s := range []string{"func (pg *statsPage) lifetime(", "ui.stats_page.lifetime.01"} {
		if strings.Contains(string(src), s) {
			t.Errorf("stats_page.go still holds %q", s)
		}
	}
}
