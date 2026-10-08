package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// The full catalog on the case: the hand-drawn legendary badges, the
// families' headings, the titles and the things only the UI can report.
// Every test runs in a temp data root, never data/.

// TestLegendSprites: each hand-drawn legendary badge is a drawing of 17 by
// 9 with a role for every inked cell and none for an empty one; it draws
// in every glyph tier, one cell a rune and nothing but ASCII in the plain
// tier; a frame is a function of its number; it moves when it may and
// holds still at rest.
func TestLegendSprites(t *testing.T) {
	if len(legendSprites) != 8 {
		t.Errorf("%d legendary sprites, want 8", len(legendSprites))
	}
	for key, sp := range legendSprites {
		if len(sp.rows) != specialH || len(sp.roles) != specialH {
			t.Fatalf("%s: %d rows and %d rows of roles, want %d", key, len(sp.rows), len(sp.roles), specialH)
		}
		inked := 0
		for y := 0; y < specialH; y++ {
			if len(sp.rows[y]) != specialW || len(sp.roles[y]) != specialW {
				t.Fatalf("%s row %d: %d cells and %d roles, want %d", key, y, len(sp.rows[y]), len(sp.roles[y]), specialW)
			}
			for x := 0; x < specialW; x++ {
				r, role := sp.rows[y][x], sp.roles[y][x]
				if (r == ' ') != (role == ' ') {
					t.Errorf("%s at %d,%d: the drawing has %q and the role map %q", key, x, y, r, role)
				}
				if r != ' ' {
					inked++
					if !strings.ContainsRune("pPqgGslbcmwdT", role) {
						t.Errorf("%s at %d,%d: the role %q is not one the painter knows", key, x, y, role)
					}
				}
			}
		}
		if inked < 30 {
			t.Errorf("%s has only %d inked cells: not much of a drawing", key, inked)
		}
		m := medal{tier: config.BadgeLegendary, emblem: '*', special: key}
		draw := func(n int, tier mapmodel.GlyphTier) string {
			g := newTGrid(specialW, specialH)
			drawMedal(g, 0, 0, m, n, tier == mapmodel.TierASCII)
			if tier == mapmodel.TierASCII {
				foldBadges(g)
			}
			for i, c := range g.c {
				if c.r != ' ' && uniseg.StringWidth(string(c.r)) != 1 {
					t.Fatalf("%s frame %d: %q at %d,%d is not one cell wide", key, n, c.r, i%g.w, i/g.w)
				}
				if tier == mapmodel.TierASCII && c.r >= 0x80 {
					t.Fatalf("%s frame %d: the plain tier draws %q", key, n, c.r)
				}
			}
			return fmt.Sprint(g.c)
		}
		if w, h := medalSize(m); w != specialW || h != specialH {
			t.Errorf("%s: its full art is %dx%d", key, w, h)
		}
		for _, tier := range caseTiers {
			rest := draw(atRest, tier)
			if rest != draw(atRest, tier) {
				t.Errorf("%s: two draws at rest differ", key)
			}
			moved := false
			for n := 0; n < 48; n++ {
				f := draw(n, tier)
				if f != draw(n, tier) {
					t.Fatalf("%s frame %d is not a function of its number", key, n)
				}
				moved = moved || f != rest
			}
			if !moved {
				t.Errorf("%s never moves in the %s tier", key, tier)
			}
		}
		// Not yet earned, it is the legendary shape, hollow: the drawing is
		// for the account that has it.
		m.state = medalLocked
		plain := m
		plain.special = ""
		a, b := newTGrid(specialW, specialH), newTGrid(specialW, specialH)
		drawMedal(a, 0, 0, m, atRest, false)
		drawMedal(b, 0, 0, plain, atRest, false)
		if a.String() != b.String() {
			t.Errorf("%s shows its drawing before it is earned", key)
		}
	}
}

// TestLegendSpritesInEveryTheme: the hand-drawn legendary badges paint in
// every theme with every cell legible on the canvas.
func TestLegendSpritesInEveryTheme(t *testing.T) {
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		pal := newGridPalette(0, nil)
		for key := range legendSprites {
			g := newTGrid(specialW, specialH)
			drawMedal(g, 0, 0, medal{tier: config.BadgeLegendary, emblem: '*', special: key}, 5, false)
			for i, c := range g.c {
				if c.r == ' ' {
					continue
				}
				fg, _, _ := pal.style(c).Decompose()
				if ratio := theme.ContrastRatio(fg, pal.bg); ratio < 2.9 {
					t.Errorf("theme %s, %s at %d,%d: %q is %06x on %06x, a contrast of %.2f", th.Key, key, i%g.w, i/g.w, c.r, fg.Hex(), pal.bg.Hex(), ratio)
				}
			}
		}
	}
}

// TestEveryFamilyHasAHeading: every family in the catalog has a heading of
// its own, in plain words, and the case lists the families in that order.
func TestEveryFamilyHasAHeading(t *testing.T) {
	titled := map[string]string{}
	for _, f := range badgeFamilyTitles {
		if titled[f.key] != "" {
			t.Errorf("the family %s has two headings", f.key)
		}
		titled[f.key] = f.title
	}
	for _, def := range rules.Core().Badges() {
		if titled[def.Family] == "" {
			t.Errorf("%s is in the family %q, which has no heading", def.Key, def.Family)
		}
	}
	for key := range titled {
		found := false
		for _, def := range rules.Core().Badges() {
			found = found || def.Family == key
		}
		if !found {
			t.Errorf("the heading for %q has no badge under it", key)
		}
	}
}

// TestCaseNamesOnlyFamiliesInSight: a family with nothing in sight has no
// tab and no heading, in the case or in the list: its name alone would say
// what is ahead. A secret with a hint is in sight. Once the account may
// know how many badges are withheld, every family is listed.
func TestCaseNamesOnlyFamiliesInSight(t *testing.T) {
	hidden := func(key, family string) game.BadgeView {
		return game.BadgeView{Key: key, Family: family, Name: game.BadgeHiddenName, Hidden: true, Level: config.BadgeBronze, Tier: "bronze"}
	}
	secret := hidden("special.s", "special")
	secret.Secret, secret.Desc = true, "Something about huts."
	views := []game.BadgeView{
		{Key: "age.stone_age", Family: "age", Name: "Rock Solid", Desc: "Reach the Stone Age.", Level: config.BadgeBronze, Tier: "bronze"},
		hidden("age.iron_age", "age"),
		hidden("civ.a.met", "civ"), hidden("civ.b.met", "civ"),
		hidden("doom.iron_era.endured", "doom"),
		secret,
	}
	sum := game.BadgeSummary{Shown: 1, Hidden: 5, Title: "Settler"}
	titles := func(tabs []caseTab) string {
		var out []string
		for _, tb := range tabs {
			out = append(out, tb.title)
		}
		return strings.Join(out, " ")
	}
	if got := titles(caseTabs(views, false)); got != "All Ages Specials Next" {
		t.Errorf("tabs for a new account: %q", got)
	}
	_, _, g := drawCase(views, sum, 120, 40, caseView{tier: mapmodel.TierUnicode}, atRest)
	text := g.String()
	for _, leak := range []string{"Civilizations", "Catastrophes"} {
		if strings.Contains(text, leak) {
			t.Errorf("the case names %q with nothing of it in sight", leak)
		}
	}
	list := strings.Join(badgeListLines(views, sum), "\n")
	for _, leak := range []string{"Civilizations", "Catastrophes"} {
		if strings.Contains(list, leak) {
			t.Errorf("the list names %q with nothing of it in sight", leak)
		}
	}
	if !strings.Contains(list, "1 secret badge, each with a hint in the badge case.") || strings.Contains(list, "Something about huts.") {
		t.Errorf("the list should count the secret and leave its hint to the case:\n%s", list)
	}
	// `badges civilizations` finds no such tab to open.
	if _, ok := findBadgeTab(views, false, "civilizations"); ok {
		t.Error("a family out of sight can be opened by name")
	}
	sum.HiddenCounted = true
	if got := titles(caseTabs(views, true)); got != "All Ages Civilizations Catastrophes Specials Next" {
		t.Errorf("tabs once the hidden badges are counted: %q", got)
	}
}

// TestPayrollLadderNamesItsNextRung: a ladder that is not climbed by a
// lifetime count has no count to show, so its label names the next rung.
func TestPayrollLadderNamesItsNextRung(t *testing.T) {
	rung := func(n int, name string) game.BadgeView {
		return game.BadgeView{Key: fmt.Sprintf("domain.food.%d", n), Family: "domain", Name: name, Desc: "Have workers at work.",
			Level: config.BadgeSilver, Tier: "silver", Emblem: "lineage.food", Ladder: "Food", Rung: n, Rungs: 2}
	}
	views := []game.BadgeView{rung(1, "Food Payroll"), rung(2, "Food Payroll II")}
	_, _, g := drawCase(views, game.BadgeSummary{Shown: 2, Title: "Settler"}, 120, 40, caseView{tab: "domain", tier: mapmodel.TierUnicode}, atRest)
	text := g.String()
	if !strings.Contains(text, "next: Food Payroll") {
		t.Errorf("the payroll ladder does not name its next rung:\n%s", text)
	}
	if strings.Contains(text, "0 of 0") {
		t.Error("the payroll ladder shows a count of nothing")
	}
}

// TestTitleCommand: `title` lists the titles an account holds, wears one
// by its name or the start of it, and goes back to the score's with
// `title default`. The status bar, the badge case and the list show the
// one worn. A title the account does not hold is refused.
func TestTitleCommand(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	if err := eng.StartNewNamedGame("run"); err != nil {
		t.Fatal(err)
	}
	res := HandleCommand("title", eng)
	if res.Type != "info" || !strings.Contains(res.Message, "Settler (worn)") || strings.Contains(res.Message, "Survivor") {
		t.Fatalf("title on a new account: %+v", res)
	}
	if res := HandleCommand("title survivor", eng); res.Type != "error" || !strings.Contains(res.Message, "do not hold") {
		t.Errorf("a title the account does not hold: %+v", res)
	}

	// Fifteen catastrophes endured: the gold rung, which gives Survivor.
	for i := 0; i < 15; i++ {
		eng.ReportForTest(config.BadgeEvEndured, "iron_era")
	}
	d.refresh()
	_, sum := eng.Badges()
	if !sliceContains(sum.Titles, "Survivor") || wornTitle(sum) != "Settler" {
		t.Fatalf("after the gold rung: titles %v, worn %q", sum.Titles, wornTitle(sum))
	}
	if res := HandleCommand("title", eng); !strings.Contains(res.Message, "Survivor") || !strings.Contains(res.Message, "Settler (worn)") {
		t.Errorf("title with two held: %+v", res)
	}
	res = HandleCommand("title surv", eng)
	if res.Type == "error" || !strings.Contains(res.Message, "You now wear the title Survivor.") {
		t.Fatalf("title surv: %+v", res)
	}
	st := eng.GetState()
	if line := untag(statusLine(st, 200)); !strings.Contains(line, "Map Tester Survivor · ") {
		t.Errorf("the status bar does not wear the chosen title: %q", line)
	}
	views, sum := eng.Badges()
	if wornTitle(sum) != "Survivor" {
		t.Errorf("the summary wears %q", wornTitle(sum))
	}
	_, _, g := drawCase(views, sum, 120, 40, caseView{tier: mapmodel.TierUnicode}, atRest)
	if !strings.Contains(strings.Split(g.String(), "\n")[0], "Survivor") {
		t.Errorf("the badge case's title bar does not wear it: %q", strings.Split(g.String(), "\n")[0])
	}
	if list := strings.Join(badgeListLines(views, sum), "\n"); !strings.Contains(list, "Title:[-] Survivor") {
		t.Error("the list does not wear the chosen title")
	}
	// The choice is the account's own: it is there on the next load.
	if got := game.NewGameEngine(); got != nil {
		got.SetAccount(eng.Account())
		if _, sum := got.Badges(); wornTitle(sum) != "Survivor" {
			t.Errorf("another engine holding the account reads the title %q", wornTitle(sum))
		}
	}
	if res := HandleCommand("title default", eng); res.Type == "error" || !strings.Contains(res.Message, "Settler") {
		t.Errorf("title default: %+v", res)
	}
	if _, sum := eng.Badges(); wornTitle(sum) != "Settler" {
		t.Errorf("after title default the account wears %q", wornTitle(sum))
	}

	// With no account there are no titles.
	if res := HandleCommand("title", game.NewGameEngine()); res.Type != "warning" {
		t.Errorf("title with no account: %+v", res)
	}
}

// alienStyle is a map style whose cursor is on the rare visitor when on is
// set, and on nothing otherwise.
type alienStyle struct {
	mapstyle.Style
	on bool
}

func (s *alienStyle) Inspect(mapstyle.Frame) (mapstyle.Inspection, bool) {
	if s.on {
		return mapstyle.Inspection{Title: "Unknown craft", Kind: mapstyle.KindAlien}, true
	}
	return mapstyle.Inspection{Title: "Hut"}, true
}
func (s *alienStyle) SetOption(mapstyle.Option, bool)                {}
func (s *alienStyle) HandleKey(*tcell.EventKey, mapstyle.Frame) bool { return true }

// TestLookingAtAVisitorEarnsItsBadge: when the Map panel's cursor comes to
// stand on the rare visitor the account is told, once each time, and the
// secret badge is earned. It is the same with the motion setting off: the
// world's clock still runs under a picture that holds still.
func TestLookingAtAVisitorEarnsItsBadge(t *testing.T) {
	const badge = "special.we_are_not_alone"
	d, eng := mapTestDashboard(t, true)
	if err := eng.StartNewNamedGame("run"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Account().SetMotion(false); err != nil {
		t.Fatal(err)
	}
	p := d.mapPanel
	p.set = d.mapSettings()
	if p.set.Motion {
		t.Fatal("precondition: motion is still on")
	}
	stub := &alienStyle{}
	p.styles.views = map[string]mapstyle.Style{p.set.Style: stub}
	told := 0
	spotted := p.spotted
	p.spotted = func(kind string) { told++; spotted(kind) }

	p.update(eng.GetState())
	if told != 0 || sliceContains(eng.Account().EarnedBadges(), badge) {
		t.Fatal("a cursor on a hut told the account of a visitor")
	}
	stub.on = true
	p.update(eng.GetState())
	p.update(eng.GetState()) // still on it: told once
	if told != 1 {
		t.Errorf("the account was told %d times of one look", told)
	}
	if !sliceContains(eng.Account().EarnedBadges(), badge) {
		t.Fatalf("looking at a visitor with motion off did not earn %s: %v", badge, eng.Account().EarnedBadges())
	}
	// It leaves and another comes: a key that moves the cursor onto it tells again.
	stub.on = false
	p.update(eng.GetState())
	stub.on = true
	p.handleKey(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
	if told != 2 {
		t.Errorf("a second look was told %d times in all, want 2", told)
	}

	// The panel's frame carries the world's clock whatever the motion
	// setting says, so a visitor has a frame to arrive on.
	p.now = func() time.Time { return p.start.Add(1000 * mapAnimStep) }
	if f := p.frame(); f.Anim != 0 || f.Clock != 1000 {
		t.Errorf("with motion off the frame is anim %d, clock %d; want 0 and 1000", f.Anim, f.Clock)
	}
	p.set.Motion = true
	if f := p.frame(); f.Anim != 1000 || f.Clock != 1000 {
		t.Errorf("with motion on the frame is anim %d, clock %d; want 1000 and 1000", f.Anim, f.Clock)
	}
}

// TestBadgeThemesSayHowTheyAreEarned: the three themes the catalog adds
// are locked until their badges are earned, their hints give nothing away
// that the badge itself hides, and the hint of a theme that comes with an
// age names the age only once the player can see it.
func TestBadgeThemesSayHowTheyAreEarned(t *testing.T) {
	_, eng := mapTestDashboard(t, true)
	if err := eng.StartNewNamedGame("run"); err != nil {
		t.Fatal(err)
	}
	st := eng.GetState()
	acct := eng.Account()
	for key, badge := range map[string]string{
		"ashfall": "special.connoisseur_of_endings", "ledger": "ladder.deals.3", "prismatic": "special.museum_piece",
	} {
		th, ok := theme.ByKey(key)
		if !ok {
			t.Fatalf("no theme %s", key)
		}
		def, ok := rules.Core().Badge(badge)
		if !ok || def.Reward.Theme != key || th.UnlockBadge != badge {
			t.Errorf("%s and %s do not name each other: theme says %q, badge gives %q", key, badge, th.UnlockBadge, def.Reward.Theme)
		}
		if themeAvailable(acct, th) {
			t.Errorf("%s is unlocked on a new account", key)
		}
		hint := themeUnlockHint(th, st)
		if hint == "" || strings.Contains(hint, def.Name) {
			t.Errorf("the hint for %s is %q", key, hint)
		}
		if !strings.Contains(def.Desc, "Unlocks the "+th.Name+" theme.") {
			t.Errorf("%s does not say it unlocks %s: %q", badge, th.Name, def.Desc)
		}
	}
	// Museum Piece is in plain sight, so its theme's hint is the badge's
	// own count.
	museum, _ := rules.Core().Badge("special.museum_piece")
	if th, _ := theme.ByKey("prismatic"); !strings.Contains(th.UnlockHint, fmt.Sprintf("%d badges", int(museum.Threshold))) {
		t.Errorf("Prismatic's hint %q does not match Museum Piece's %v badges", th.UnlockHint, museum.Threshold)
	}
	// An age's theme: "a later age" until the age is in sight.
	cosmic, _ := theme.ByKey("cosmic")
	if got := themeUnlockHint(cosmic, st); got != "Reach a later age" {
		t.Errorf("Cosmic's hint in the first age is %q", got)
	}
	bronze, _ := theme.ByKey("bronze")
	if err := eng.EnterAgeForTest("stone_age"); err != nil {
		t.Fatal(err)
	}
	if got := themeUnlockHint(bronze, eng.GetState()); got != "Reach the Bronze Age" {
		t.Errorf("Bronze's hint with the Bronze Age next is %q", got)
	}
	// Earning the age's badge unlocks its theme at once, and says so once.
	if err := eng.EnterAgeForTest("bronze_age"); err != nil {
		t.Fatal(err)
	}
	if !acct.HasTheme("bronze") {
		t.Fatal("reaching the Bronze Age did not unlock Bronze")
	}
}
