package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// factionStage is the Factions panel's buttons over a snapshot, with what
// they ran and asked collected.
type factionStage struct {
	fv    *factionsView
	key   string
	name  string
	ran   []string
	asked [][3]string
	rect  image.Rectangle
}

// stageFactions meets the first civilization at opinion, lets mut change the
// snapshot, and lays the buttons.
func stageFactions(t *testing.T, opinion int, mut func(st *game.GameState, key string)) *factionStage {
	t.Helper()
	_, _, eng := stagedDashboard(t, "iron_age")
	def := eng.Rules().Factions()[0]
	must(t, eng.MeetFactionForTest(def.Key, opinion))
	st := eng.GetState()
	fs := make(map[string]game.FactionInfo, len(st.Diplomacy.Factions))
	for k, v := range st.Diplomacy.Factions {
		fs[k] = v
	}
	st.Diplomacy.Factions = fs
	res := make(map[string]game.ResourceState, len(st.Resources))
	for k, v := range st.Resources {
		res[k] = v
	}
	st.Resources = res
	feat := make(map[string]game.FeatureState, len(st.Features))
	for k, v := range st.Features {
		v.Open = true
		feat[k] = v
	}
	st.Features = feat
	stockSet(&st, "gold", 1e9)
	stockSet(&st, "culture", 1e9)
	if mut != nil {
		mut(&st, def.Key)
	}
	s := &factionStage{fv: &factionsView{}, key: def.Key, name: def.Name, rect: image.Rect(2, 1, 130, 400)}
	s.fv.pr.press = func(kit.Key) {}
	s.fv.pr.run = func(c string) { s.ran = append(s.ran, c) }
	s.fv.pr.confirm = func(q, yes, line string) { s.asked = append(s.asked, [3]string{q, yes, line}) }
	s.fv.provider(st, 130)
	s.fv.grid, s.fv.org = s.fv.layer(s.rect, 0), s.rect.Min
	return s
}

func stockSet(st *game.GameState, res string, v float64) {
	r := st.Resources[res]
	r.Amount = v
	st.Resources[res] = r
}

func setFaction(st *game.GameState, key string, f func(fi *game.FactionInfo)) {
	fi := st.Diplomacy.Factions[key]
	f(&fi)
	st.Diplomacy.Factions[key] = fi
}

func (s *factionStage) button(t *testing.T, cmd string) xkey {
	t.Helper()
	if s.fv.grid != nil {
		for _, k := range s.fv.grid.keyset().keys {
			if k.cmd == cmd {
				return k
			}
		}
	}
	t.Fatalf("no button carries %q", cmd)
	return xkey{}
}

func (s *factionStage) click(k xkey) {
	pt := s.rect.Min.Add(k.r.Min).Add(image.Pt(k.r.Dx()/2, k.r.Dy()/2))
	for _, kind := range []kit.MouseKind{kit.MousePress, kit.MouseRelease} {
		s.fv.mouse(kit.Mouse{Kind: kind, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
	}
}

func TestFactionVerbsRunTheirCommandOnceByClick(t *testing.T) {
	s := stageFactions(t, game.AllyOpinion, nil)
	for _, sub := range []string{"gift", "ally", "rival", "embargo"} {
		cmd := diplomacyCmd + sub + " " + s.key
		k := s.button(t, cmd)
		if k.off {
			t.Fatalf("%s is off", sub)
		}
		s.ran = nil
		s.click(k)
		if len(s.ran) != 1 || s.ran[0] != cmd {
			t.Errorf("%s ran %q, want %q once", sub, s.ran, cmd)
		}
	}
	if len(s.asked) != 0 {
		t.Errorf("a peaceful verb asked %q", s.asked)
	}
}

func TestFactionGiftLeadsAndTributeTakesOverAtWar(t *testing.T) {
	s := stageFactions(t, 0, nil)
	if k := s.button(t, diplomacyCmd+"gift "+s.key); k.kind != keyAccent {
		t.Errorf("Gift is %v, want primary", k.kind)
	}
	if k := s.button(t, diplomacyCmd+"tribute "+s.key); !k.off {
		t.Error("Tribute is on with no war")
	}
	w := stageFactions(t, 0, func(st *game.GameState, key string) {
		setFaction(st, key, func(f *game.FactionInfo) { f.AtWar = true })
	})
	if k := w.button(t, diplomacyCmd+"tribute "+w.key); k.kind != keyAccent || k.off {
		t.Errorf("Tribute at war is %v off=%v, want primary and on", k.kind, k.off)
	}
	if k := w.button(t, diplomacyCmd+"gift "+w.key); k.kind != keySteel {
		t.Errorf("Gift at war is %v, want plain", k.kind)
	}
}

func TestFactionVerbsAreOffWhenTheEngineWouldRefuse(t *testing.T) {
	for name, c := range map[string]struct {
		opinion int
		mut     func(st *game.GameState, key string)
		sub     string
	}{
		"ally without the opinion": {0, nil, "ally"},
		"ally when allied": {game.AllyOpinion, func(st *game.GameState, k string) {
			setFaction(st, k, func(f *game.FactionInfo) { f.Status = "allied" })
		}, "ally"},
		"rival when rival": {0, func(st *game.GameState, k string) {
			setFaction(st, k, func(f *game.FactionInfo) { f.Status = "rival" })
		}, "rival"},
		"embargo when embargoed": {0, func(st *game.GameState, k string) {
			setFaction(st, k, func(f *game.FactionInfo) { f.Status = "embargo" })
		}, "embargo"},
		"neutral when neutral": {0, func(st *game.GameState, k string) {
			setFaction(st, k, func(f *game.FactionInfo) { f.Status = "neutral" })
		}, "neutral"},
		"gift without gold": {0, func(st *game.GameState, _ string) { stockSet(st, "gold", 0) }, "gift"},
		"gift while locked": {0, func(st *game.GameState, _ string) {
			f := st.Features["diplomacy"]
			f.Open = false
			st.Features["diplomacy"] = f
		}, "gift"},
	} {
		s := stageFactions(t, c.opinion, c.mut)
		cmd := diplomacyCmd + c.sub + " " + s.key
		k := s.button(t, cmd)
		if !k.off {
			t.Errorf("%s: the button is on", name)
		}
		s.click(k)
		if len(s.ran)+len(s.asked) != 0 {
			t.Errorf("%s: an off button ran %q and asked %q", name, s.ran, s.asked)
		}
	}
}

func TestFactionRaidAsksFirstNamingTheCivAndRunsOnlyOnYes(t *testing.T) {
	s := stageFactions(t, 0, nil)
	cmd := diplomacyCmd + "raid " + s.key
	k := s.button(t, cmd)
	if k.kind != keyAlert || k.off {
		t.Fatalf("Raid is %v off=%v, want a danger button that is on", k.kind, k.off)
	}
	s.click(k)
	if len(s.ran) != 0 || len(s.asked) != 1 {
		t.Fatalf("a click ran %q and asked %d times, want a question only", s.ran, len(s.asked))
	}
	q := s.asked[0]
	if !strings.Contains(q[0], s.name) || q[1] != txtVerbRaid.String() || q[2] != cmd {
		t.Errorf("asked %q, want a question naming %q, the word %q and %q", q, s.name, txtVerbRaid.String(), cmd)
	}
}

// Tribute is off one gold short of the engine's price and on at it.
func TestFactionTributeIsOffOneGoldShortAndOnAtThePrice(t *testing.T) {
	for _, c := range []struct {
		name string
		diff float64
		off  bool
	}{{"one short", -1, true}, {"at the price", 0, false}} {
		s := stageFactions(t, 0, func(st *game.GameState, key string) {
			gold, _ := game.TributePrice(config.FactionByKey()[key])
			setFaction(st, key, func(f *game.FactionInfo) { f.AtWar = true })
			stockSet(st, "gold", gold+c.diff)
		})
		if k := s.button(t, diplomacyCmd+"tribute "+s.key); k.off != c.off {
			t.Errorf("%s: Tribute off=%v, want %v", c.name, k.off, c.off)
		}
	}
}

// A card gives its open offers one line and a button to Trade; it has no
// accept of its own, and a taken offer is not counted.
func TestFactionOffersLineOpensTrade(t *testing.T) {
	s := stageFactions(t, 0, func(st *game.GameState, key string) {
		setFaction(st, key, func(f *game.FactionInfo) {
			f.Deals = []game.DealInfo{
				{Num: 1, Kind: game.DealBoost, Give: "gold", GiveAmt: 1, Res: "gold", Amt: 2},
				{Num: 2, Kind: game.DealBoost, Give: "gold", GiveAmt: 1, Res: "gold", Amt: 2, Taken: true},
			}
			f.DealsBlocked = ""
		})
	})
	if !strings.Contains(s.fv.text, "1 offer open") {
		t.Errorf("the card lacks its offers line:\n%s", s.fv.text)
	}
	s.click(s.button(t, "trade"))
	if len(s.ran) != 1 || s.ran[0] != "trade" {
		t.Errorf("ran %q, want the Trade panel", s.ran)
	}
	for _, b := range s.fv.grid.keyset().keys {
		if strings.HasPrefix(b.cmd, diplomacyCmd+"accept") {
			t.Errorf("the card has an accept button: %q", b.cmd)
		}
	}
}

func TestFactionsTextNoLongerTypesWhatTheButtonsCarry(t *testing.T) {
	s := stageFactions(t, game.AllyOpinion, func(st *game.GameState, key string) {
		setFaction(st, key, func(f *game.FactionInfo) { f.Status = "allied" })
	})
	for _, line := range []string{"Commands: diplomacy ally", "diplomacy rival/embargo/neutral"} {
		if strings.Contains(s.fv.text, line) {
			t.Errorf("the panel still types %q", line)
		}
	}
}

// With the mouse alone, through the dashboard: a gift, then a raid answered
// by the danger button.
func TestFactionGiftAndRaidByMouseAlone(t *testing.T) {
	keepTheme(t)
	d, pages, eng := stagedDashboard(t, "iron_age")
	def := eng.Rules().Factions()[0]
	eng.GrantTechsForTest("envoys")
	must(t, eng.MeetFactionForTest(def.Key, 0))
	eng.SetStockForTest("gold", 1e9)
	d.overlayMgr.screenW = 130
	show := func() {
		d.overlayMgr.Show("factions", eng.GetState())
		drawPrimitive(t, pages, 130, 60)
	}
	show()
	press := func(cmd string) {
		t.Helper()
		var found *xkey
		for i, k := range d.factions.grid.keyset().keys {
			if k.cmd == cmd {
				found = &d.factions.grid.keyset().keys[i]
			}
		}
		if found == nil {
			t.Fatalf("no button carries %q", cmd)
		}
		text := d.overlayMgr.entries["factions"].tv.view.Bounds()
		pt := text.Min.Add(found.r.Min).Add(image.Pt(found.r.Dx()/2, found.r.Dy()/2))
		for _, kind := range []kit.MouseKind{kit.MousePress, kit.MouseRelease} {
			d.overlayMgr.entries["factions"].tv.Mouse(kit.Mouse{Kind: kind, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
		}
	}
	before := eng.GetState().Diplomacy.Factions[def.Key].Opinion
	press(diplomacyCmd + "gift " + def.Key)
	if after := eng.GetState().Diplomacy.Factions[def.Key].Opinion; after <= before {
		t.Fatalf("the gift did not raise opinion: %d -> %d", before, after)
	}
	show()
	op := eng.GetState().Diplomacy.Factions[def.Key].Opinion
	press(diplomacyCmd + "raid " + def.Key)
	if got := eng.GetState().Diplomacy.Factions[def.Key].Opinion; got != op {
		t.Fatal("the raid ran before the question was answered")
	}
	_, item := pages.GetFrontPage()
	dlg, ok := item.(*nativeDialog)
	if !ok {
		t.Fatalf("the front page is %T, not the question", item)
	}
	d.app.SetFocus(dlg)
	keyTo(dlg, tk(kit.KeyTab, 0, 0))
	keyTo(dlg, tk(kit.KeyEnter, 0, 0))
	if got := eng.GetState().Diplomacy.Factions[def.Key].Opinion; got >= op {
		t.Errorf("the raid did not lower opinion: %d -> %d", op, got)
	}
}

// raidStage is the dashboard with a civilization met and its Raid button
// found, the lines run through the command bar's path counted.
func raidStage(t *testing.T) (press func(), answer func(danger bool), opinion func() int, ran *[]string, front func() string) {
	t.Helper()
	keepTheme(t)
	d, pages, eng := stagedDashboard(t, "iron_age")
	def := eng.Rules().Factions()[0]
	eng.GrantTechsForTest("envoys")
	must(t, eng.MeetFactionForTest(def.Key, 0))
	d.overlayMgr.screenW = 130
	d.overlayMgr.Show("factions", eng.GetState())
	drawPrimitive(t, pages, 130, 60)
	cmd := diplomacyCmd + "raid " + def.Key
	press = func() {
		for _, k := range d.factions.grid.keyset().keys {
			if k.cmd != cmd {
				continue
			}
			tv := d.overlayMgr.entries["factions"].tv
			pt := tv.view.Bounds().Min.Add(k.r.Min).Add(image.Pt(k.r.Dx()/2, k.r.Dy()/2))
			for _, kind := range []kit.MouseKind{kit.MousePress, kit.MouseRelease} {
				tv.Mouse(kit.Mouse{Kind: kind, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
			}
			return
		}
		t.Fatalf("no button carries %q", cmd)
	}
	answer = func(danger bool) {
		_, item := pages.GetFrontPage()
		dlg, ok := item.(*nativeDialog)
		if !ok {
			t.Fatalf("the front page is %T, not the question", item)
		}
		d.app.SetFocus(dlg)
		if danger {
			keyTo(dlg, tk(kit.KeyTab, 0, 0))
		}
		keyTo(dlg, tk(kit.KeyEnter, 0, 0))
	}
	opinion = func() int { return eng.GetState().Diplomacy.Factions[def.Key].Opinion }
	front = func() string { name, _ := pages.GetFrontPage(); return name }
	return press, answer, opinion, &d.cmdHistory, front
}

// A click on Raid puts its question on screen and leaves the opinion alone;
// Cancel leaves it alone too.
func TestFactionRaidClickAsksAndCancelChangesNothing(t *testing.T) {
	press, answer, opinion, ran, front := raidStage(t)
	before := opinion()
	press()
	if front() != "ask_line" {
		t.Fatalf("the front page is %q, not the question", front())
	}
	if opinion() != before || len(*ran) != 0 {
		t.Fatalf("the click ran %q and moved opinion %d -> %d", *ran, before, opinion())
	}
	answer(false)
	if opinion() != before || len(*ran) != 0 || front() == "ask_line" {
		t.Errorf("Cancel ran %q, opinion %d -> %d, front %q", *ran, before, opinion(), front())
	}
}

// The danger button runs the raid exactly once and the opinion falls.
func TestFactionRaidDangerButtonRunsOnce(t *testing.T) {
	press, answer, opinion, ran, _ := raidStage(t)
	before := opinion()
	press()
	answer(true)
	if len(*ran) != 1 || !strings.HasPrefix((*ran)[0], diplomacyCmd+"raid ") {
		t.Errorf("ran %q, want one raid", *ran)
	}
	if got := opinion(); got >= before {
		t.Errorf("opinion %d -> %d, want it to fall", before, got)
	}
}
