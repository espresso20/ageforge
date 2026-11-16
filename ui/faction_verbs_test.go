package ui

import (
	"image"
	"strings"
	"testing"

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
	if k := s.button(t, diplomacyCmd+"gift "+s.key); k.kind != keyPrimary {
		t.Errorf("Gift is %v, want primary", k.kind)
	}
	if k := s.button(t, diplomacyCmd+"tribute "+s.key); !k.off {
		t.Error("Tribute is on with no war")
	}
	w := stageFactions(t, 0, func(st *game.GameState, key string) {
		setFaction(st, key, func(f *game.FactionInfo) { f.AtWar = true })
	})
	if k := w.button(t, diplomacyCmd+"tribute "+w.key); k.kind != keyPrimary || k.off {
		t.Errorf("Tribute at war is %v off=%v, want primary and on", k.kind, k.off)
	}
	if k := w.button(t, diplomacyCmd+"gift "+w.key); k.kind != keyPlain {
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
	if k.kind != keyDanger || k.off {
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

func TestFactionDealButtonCarriesAcceptAndIsOffWhenTaken(t *testing.T) {
	s := stageFactions(t, 0, func(st *game.GameState, key string) {
		setFaction(st, key, func(f *game.FactionInfo) {
			f.Deals = []game.DealInfo{
				{Num: 1, Kind: game.DealWant, Give: "gold", GiveAmt: 1, Get: "gold", GetAmt: 2},
				{Num: 2, Kind: game.DealWant, Give: "gold", GiveAmt: 1, Get: "gold", GetAmt: 2, Taken: true},
			}
			f.DealsBlocked = ""
		})
	})
	one := s.button(t, diplomacyCmd+"accept "+s.key+" 1")
	two := s.button(t, diplomacyCmd+"accept "+s.key+" 2")
	if one.off || !two.off {
		t.Fatalf("offers: first off=%v, taken off=%v", one.off, two.off)
	}
	s.click(one)
	s.click(two)
	if len(s.ran) != 1 || s.ran[0] != one.cmd {
		t.Errorf("ran %q, want only %q", s.ran, one.cmd)
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
