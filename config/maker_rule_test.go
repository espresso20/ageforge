package config

import (
	"math"
	"sort"
	"testing"
)

// The maker rule (maker_rule.go): payback(A) is written down and recomputed;
// the holes are the ones the rules find and MakerFills fills; each filled
// output is what the standard gives.

// TestMakerPaybackTable recomputes payback(A) for every age from the building
// tables and holds MakerPaybackTicks to it. The table was worked out before
// the fills; the fills sit on the standard, so the median of the tables as
// they are moves by less than a fifth of a percent, and the healthy makers of
// every age cluster around it.
// The two ages that sit outside a tenth of their median are the two first
// makers an earlier change re-priced without their own material (the Server
// Farm in the Information Age, the Black Market Hub in the Cyberpunk Age):
// each pays back within a fifth of the median.
func TestMakerPaybackTable(t *testing.T) {
	defs := BaseBuildings()
	got := MakerPaybacks(defs, AgeOrder())
	for _, age := range AgeOrder() {
		want, ok := MakerPaybackTicks[age]
		if !ok {
			t.Errorf("%s has no entry in MakerPaybackTicks", age)
			continue
		}
		g := got[age]
		if g.Healthy == 0 {
			t.Errorf("%s has no healthy maker of a material it prices", age)
			continue
		}
		if math.Abs(g.Median/want-1) > 0.002 {
			t.Errorf("%s: with the fills the median payback is %.6g ticks, %.2f%% from MakerPaybackTicks' %.6g", age, g.Median, (g.Median/want-1)*100, want)
		}
		if g.Lo/g.Median < 0.9 || g.Hi/g.Median > 1.25 {
			t.Errorf("%s: healthy makers pay back between %.3f and %.3f of the median %.6g (want within 0.9 to 1.25)", age, g.Lo/g.Median, g.Hi/g.Median, g.Median)
		}
		t.Logf("%-17s payback %-9.6g healthy %d of %d makers, %.3f to %.3f of the median", age, g.Median, g.Healthy, g.Makers, g.Lo/g.Median, g.Hi/g.Median)
	}
	if len(MakerPaybackTicks) != len(AgeOrder()) {
		t.Errorf("MakerPaybackTicks has %d entries for %d ages", len(MakerPaybackTicks), len(AgeOrder()))
	}
}

// checkMakerHolesAreFilledByRule proves the fills against the rules. Taking the
// ages in order, each on the tables as the fills of the ages before it leave
// them and without its own, the rules find holes: every one
// the rules decide is in MakerFills with the same building and rule, every
// one they cannot decide is in MakerOpenHoles, and nothing else is in either
// list. Read the tables as they are, the only (age, material) left without a
// healthy maker a run can build is an open hole.
func checkMakerHolesAreFilledByRule(t *testing.T) {
	defs, ages, techs := BaseBuildings(), Ages(), Technologies()
	fills := map[string]MakerFill{}
	for _, f := range MakerFills() {
		fills[f.Age+"/"+f.Material] = f
	}
	open := map[string]MakerOpenHole{}
	for _, o := range MakerOpenHoles() {
		open[o.Age+"/"+o.Material] = o
	}
	seen := map[string]bool{}
	for _, h := range MakerHoles(defs, ages, techs, MakerFills()) {
		id := h.Age + "/" + h.Material
		seen[id] = true
		switch h.Rule {
		case 1, 2, 3:
			f, ok := fills[id]
			if !ok {
				t.Errorf("%s: rule %d gives %s, and MakerFills has no entry", id, h.Rule, h.Building)
			} else if f.Building != h.Building || f.Rule != h.Rule {
				t.Errorf("%s: the rules give %s by rule %d, MakerFills says %s by rule %d", id, h.Building, h.Rule, f.Building, f.Rule)
			}
		default:
			if _, ok := open[id]; !ok {
				t.Errorf("%s: the rules leave it open (%s), and MakerOpenHoles does not list it", id, h.Why)
			}
		}
	}
	for id := range fills {
		if !seen[id] {
			t.Errorf("MakerFills lists %s, which is not a hole", id)
		}
	}
	for id := range open {
		if !seen[id] {
			t.Errorf("MakerOpenHoles lists %s, which is not a hole", id)
		}
	}
	byKey := BuildingByKey()
	for _, f := range MakerFills() {
		d := byKey[f.Building]
		if d.RequiredAge != f.Age {
			t.Errorf("%s fills a hole of %s but is a building of %s", f.Building, f.Age, d.RequiredAge)
		}
		made := false
		for _, e := range d.Effects {
			if e.Type == "production" && e.Target == f.Material && e.Value > 0 {
				made = true
			}
		}
		if !made {
			t.Errorf("%s is to make %s in %s and does not", f.Building, f.Material, f.Age)
		}
	}
	var left []string
	for _, h := range MakerHoles(defs, ages, techs, nil) {
		if h.Rule == 2 {
			continue // a raw form's maker makes the material hole or not
		}
		if _, ok := open[h.Age+"/"+h.Material]; !ok {
			left = append(left, h.Age+"/"+h.Material)
		}
	}
	if len(left) > 0 {
		t.Errorf("holes with no healthy maker in the tables and no entry in MakerOpenHoles: %v", left)
	}
}

// checkMakerFilledOutputsAreTheStandard holds every output a fill touches to
// the standard: the building's first-copy price in units over payback(A),
// split by what the age's gate asks of each priced material it makes, turned
// back into the material by its level. The tables write three figures, so
// the tolerance is half a percent.
func checkMakerFilledOutputsAreTheStandard(t *testing.T) {
	defs := BaseBuildings()
	want := MakerStandard(defs, Ages(), MakerPaybackTicks, MakerFills())
	byKey := BuildingByKey()
	keys := make([]string, 0, len(want))
	for k := range want {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		made := map[string]float64{}
		for _, e := range byKey[key].Effects {
			if e.Type == "production" {
				made[e.Target] += e.Value
			}
		}
		for res, w := range want[key] {
			if math.Abs(made[res]/w-1) > 0.005 {
				t.Errorf("%s makes %.4g %s a tick, the standard is %.4g", key, made[res], res, w)
			}
		}
	}
}

// TestEveryHoleIsFilledByRule is the one test of the fills: the holes the
// rules find are the ones MakerFills fills and MakerOpenHoles lists, and every
// output a fill touches is what the standard gives.
func TestEveryHoleIsFilledByRule(t *testing.T) {
	t.Run("holes", checkMakerHolesAreFilledByRule)
	t.Run("outputs", checkMakerFilledOutputsAreTheStandard)
}

// tradeGoldUnpriced are the trade buildings whose age prices nothing in gold:
// their gold is left as it was and is listed, not put on the standard.
var tradeGoldUnpriced = []string{
	"crypto_exchange", "black_market", "energy_exchange", "asteroid_market",
	"galactic_trade_hub", "stellar_exchange", "probability_market", "omniversal_bazaar",
}

// TestTradeBuildingsMakeGold: every trade-lineage building makes gold, and in
// an age that prices gold it makes it at the standard (its payback is the
// age's, to half a percent; the Atomic Age's Corporate HQ was moved there when
// the Uranium Processing Works' re-price moved the age's price levels). The ones in an age that prices no gold are the listed
// ones. The Black Market Hub makes crypto as well.
func TestTradeBuildingsMakeGold(t *testing.T) {
	var unpriced []string
	for _, d := range BaseBuildings() {
		if d.LineageKey != "trade" {
			continue
		}
		gold := 0.0
		for _, e := range d.Effects {
			if e.Type == "production" && e.Target == "gold" {
				gold += e.Value
			}
		}
		if gold <= 0 {
			t.Errorf("%s is a trade building and makes no gold", d.Key)
			continue
		}
		lv := PriceLevels(d.RequiredAge)
		if lv["gold"] <= 0 {
			unpriced = append(unpriced, d.Key)
			continue
		}
		price, out := makerUnits(d, lv, nil)
		pb := MakerPaybackTicks[d.RequiredAge]
		if r := float64(price/out) / pb; math.Abs(r-1) > 0.005 {
			t.Errorf("%s pays back in %.4g of payback(%s): its gold is not on the standard", d.Key, r, d.RequiredAge)
		}
	}
	sort.Strings(unpriced)
	want := append([]string(nil), tradeGoldUnpriced...)
	sort.Strings(want)
	if len(unpriced) != len(want) {
		t.Fatalf("trade buildings in an age that prices no gold: %v, want %v", unpriced, want)
	}
	for i := range want {
		if unpriced[i] != want[i] {
			t.Errorf("trade buildings in an age that prices no gold: %v, want %v", unpriced, want)
			break
		}
	}
	bm := BuildingByKey()["black_market"]
	crypto := 0.0
	for _, e := range bm.Effects {
		if e.Type == "production" && e.Target == "crypto" {
			crypto += e.Value
		}
	}
	if crypto <= 0 {
		t.Errorf("black_market makes no crypto")
	}
}
