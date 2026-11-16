package ui

import (
	"image"
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// stopTrade opens the Trade panel of a Colonial Age game at w by h with the
// tech for the black market and plenty of culture.
func stopTrade(t *testing.T, w, h int) (*Dashboard, *nativePages, *game.GameEngine, *deckPanel) {
	t.Helper()
	d, pages, eng := stagedDashboard(t, "colonial_age")
	eng.GrantTechsForTest("mercantilism")
	eng.SetStockForTest("culture", 1e9)
	openPanel(t, d, "trade", w, h)
	paintAlone(pages, w, h)
	return d, pages, eng, d.decks["trade"]
}

func stopChips(p *deckPanel) []xkey {
	var out []xkey
	for _, k := range p.grid.keyset().keys {
		if strings.HasPrefix(k.cmd, "blackmarket ") {
			out = append(out, k)
		}
	}
	return out
}

func clickKey(p *deckPanel, k xkey) {
	pt := p.Bounds().Min.Add(k.r.Min).Add(image.Pt(1, 0))
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
	p.Mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
}

func stopIndex(t *testing.T, p *deckPanel) int {
	t.Helper()
	for n := 0; n < p.pg.count(); n++ {
		if c, _ := p.pg.card(n); c.Stop {
			return n
		}
	}
	t.Fatal("the page has no smugglers' stop")
	return -1
}

// pickStop picks the stop as a click on it does and draws again.
func pickStop(t *testing.T, p *deckPanel, pages *nativePages, w, h int) {
	t.Helper()
	p.pick(stopIndex(t, p))
	paintAlone(pages, w, h)
}

// TestBlackMarketStop_ChipRunsOnce: a click on a payout chip runs
// blackmarket <key> once, at once and without a question, and the culture
// falls by the cost.
func TestBlackMarketStop_ChipRunsOnce(t *testing.T) {
	for _, sz := range [][2]int{{140, 40}, {80, 50}} { // the map, then the cards
		w, h := sz[0], sz[1]
		d, pages, eng, p := stopTrade(t, w, h)
		pickStop(t, p, pages, w, h)
		chips := stopChips(p)
		st := eng.GetState().BlackMarket
		if len(chips) != len(st.Payouts) || len(chips) == 0 {
			t.Fatalf("%dx%d: %d chips for %d payouts", w, h, len(chips), len(st.Payouts))
		}
		for i, k := range chips {
			if k.cmd != "blackmarket "+st.Payouts[i].Resource || k.off {
				t.Fatalf("%dx%d: chip %d = %+v, want an on chip for %s", w, h, i, k, st.Payouts[i].Resource)
			}
		}
		before := eng.GetState().BlackMarket
		clickKey(p, chips[0])
		after := eng.GetState().BlackMarket
		if got, want := before.Culture-after.Culture, before.Cost; math.Abs(got-want) > 1e-3*want {
			t.Errorf("%dx%d: culture fell by %v, want the cost %v", w, h, got, want)
		}
		if after.Ready {
			t.Errorf("%dx%d: no wait after a run", w, h)
		}
		// During the wait every chip is off and a click does nothing.
		d.overlayMgr.Show("trade", eng.GetState())
		paintAlone(pages, w, h)
		p = d.decks["trade"]
		pickStop(t, p, pages, w, h)
		for _, k := range stopChips(p) {
			if !k.off {
				t.Errorf("%dx%d: chip %q is on during the wait", w, h, k.cmd)
			}
		}
		clickKey(p, stopChips(p)[0])
		if again := eng.GetState().BlackMarket; again.Culture != after.Culture {
			t.Errorf("%dx%d: a click on an off chip spent culture", w, h)
		}
	}
}

// TestBlackMarketStop_ShortOfCulture: one culture short turns every chip off.
func TestBlackMarketStop_ShortOfCulture(t *testing.T) {
	d, pages, eng, p := stopTrade(t, 140, 40)
	cost := eng.GetState().BlackMarket.Cost
	eng.SetStockForTest("culture", cost-1)
	d.overlayMgr.Show("trade", eng.GetState())
	paintAlone(pages, 140, 40)
	p = d.decks["trade"]
	pickStop(t, p, pages, 140, 40)
	chips := stopChips(p)
	if len(chips) == 0 {
		t.Fatal("no chips")
	}
	for _, k := range chips {
		if !k.off {
			t.Errorf("chip %q is on one culture short", k.cmd)
		}
	}
	c, _ := p.pg.card(stopIndex(t, p))
	if c.Chip != "not enough culture" {
		t.Errorf("status = %q", c.Chip)
	}
	clickKey(p, chips[0])
	if got := eng.GetState().BlackMarket.Culture; math.Abs(got-(cost-1)) > 1e-3*cost {
		t.Errorf("culture = %v after a click on an off chip, want %v", got, cost-1)
	}
	// With exactly the cost the chips are on.
	eng.SetStockForTest("culture", cost)
	d.overlayMgr.Show("trade", eng.GetState())
	paintAlone(pages, 140, 40)
	p = d.decks["trade"]
	pickStop(t, p, pages, 140, 40)
	for _, k := range stopChips(p) {
		if k.off {
			t.Errorf("chip %q is off with the cost in hand", k.cmd)
		}
	}
}

// TestBlackMarketStop_Gates: no stop before the Colonial Age; quiet with the
// reason and no chips in it without the tech.
func TestBlackMarketStop_Gates(t *testing.T) {
	early := tradeState() // medieval
	for _, sec := range tradePage(early).Sections {
		for _, c := range sec.Cards {
			if c.Stop {
				t.Fatal("a stop before the Colonial Age")
			}
		}
	}
	st := game.GameState{Age: "colonial_age"}
	st.BlackMarket = game.BlackMarketStatus{Available: true, Cost: 10, Culture: 100, Ready: true,
		Payouts: []game.BlackMarketHaul{{Resource: "gold", Amount: 5}}}
	pg := tradePage(st)
	c, _ := deckCardNamed(t, pg, "Black market")
	if len(c.Chips) != 1 || c.Chips[0].Command != "blackmarket gold" || c.Chips[0].Off || c.Quiet {
		t.Errorf("open stop = %+v", c)
	}
}

// TestBlackMarketStop_QuietWithoutTech: the engine's own state, tech missing.
func TestBlackMarketStop_QuietWithoutTech(t *testing.T) {
	d, pages, eng := stagedDashboard(t, "colonial_age")
	eng.SetStockForTest("culture", 1e9)
	st := eng.GetState()
	feats := map[string]game.FeatureState{}
	for k, v := range st.Features {
		feats[k] = v
	}
	f := feats[config.FeatureBlackMarket]
	f.Open = false
	feats[config.FeatureBlackMarket] = f
	st.Features = feats
	refusal, locked := featureRefusal(st, config.FeatureBlackMarket)
	if !locked || refusal == "" {
		t.Fatal("the feature should read as locked")
	}
	d.overlayMgr.screenW, d.overlayMgr.screenH = 140, 40
	d.overlayMgr.Show("trade", st)
	paintAlone(pages, 140, 40)
	p := d.decks["trade"]
	c, ok := p.pg.card(stopIndex(t, p))
	if !ok || !c.Quiet || c.Why != refusal || len(c.Chips) != 0 || c.Command != "" {
		t.Errorf("locked stop = %+v, want quiet with %q and nothing to press", c, refusal)
	}
	pickStop(t, p, pages, 140, 40)
	if len(stopChips(p)) != 0 {
		t.Error("a locked stop has chips")
	}
}

// TestBlackMarketStop_SelectionKeys: the arrow keys that move between the
// page's cards reach the stop, and the cards form holds its chips.
func TestBlackMarketStop_SelectionKeys(t *testing.T) {
	for _, sz := range [][2]int{{140, 40}, {80, 50}} {
		w, h := sz[0], sz[1]
		_, pages, _, p := stopTrade(t, w, h)
		want := stopIndex(t, p)
		p.pick(0)
		var seen []int
		for i := 0; i < p.pg.count()+1 && p.sel != want; i++ {
			p.routeKey(&kit.Key{Code: kit.KeyRight}, "")
			paintAlone(pages, w, h)
			seen = append(seen, p.sel)
		}
		if p.sel != want {
			t.Errorf("%dx%d: the right arrow never reached the stop (at %d, want %d, saw %v)", w, h, p.sel, want, seen)
		}
		if len(stopChips(p)) == 0 {
			t.Errorf("%dx%d: no chips with the stop picked", w, h)
		}
	}
}

// TestLayChips: the chips' rows are the rows the card's height counts, and
// no chip runs past the width or into another.
func TestLayChips(t *testing.T) {
	var chips []deckChip
	for _, w := range []string{"2.25B iron", "180B gold", "403M coal", "11.2B steel", "7 food"} {
		chips = append(chips, deckChip{Word: w})
	}
	for w := 12; w < 60; w += 5 {
		at, rows := layChips(chips, w)
		c := deckCard{Chips: chips}
		bare := deckCard{}
		if c.height(w+4, true)-bare.height(w+4, true) != rows {
			t.Errorf("w=%d: card height adds %d, layChips says %d rows", w, c.height(w+4, true)-bare.height(w+4, true), rows)
		}
		for i := range chips {
			wid := runeLen(chips[i].Word) + 2
			if at[i].X+min(wid, max(w, 1)) > w+1 && wid <= w {
				t.Errorf("w=%d: chip %d ends at %d", w, i, at[i].X+wid)
			}
			if i > 0 && at[i].Y == at[i-1].Y && at[i].X < at[i-1].X+runeLen(chips[i-1].Word)+2+1 {
				t.Errorf("w=%d: chips %d and %d touch", w, i-1, i)
			}
		}
	}
}
