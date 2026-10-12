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
// The band is the tightest every healthy maker of every age passes, at two
// figures: 0.99 to 1.01 (the tables measure 0.997 to 1.004). A maker whose price
// was re-authored, or whose age's price levels moved under it, has its output
// re-derived from its price (Rule 6 in MakerFills), so it stays inside it.
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
		if g.Lo/g.Median < 0.99 || g.Hi/g.Median > 1.01 {
			t.Errorf("%s: healthy makers pay back between %.3f and %.3f of the median %.6g (want within 0.99 to 1.01)", age, g.Lo/g.Median, g.Hi/g.Median, g.Median)
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
		if f.Rule < 5 {
			fills[f.Age+"/"+f.Material+"/"+f.Building] = f
		}
	}
	open := map[string]MakerOpenHole{}
	for _, o := range MakerOpenHoles() {
		open[o.Age+"/"+o.Material] = o
	}
	seenFill, seenOpen := map[string]bool{}, map[string]bool{}
	for _, h := range MakerHoles(defs, ages, techs, MakerFills()) {
		id := h.Age + "/" + h.Material
		switch h.Rule {
		case 1, 2, 3:
			fid := id + "/" + h.Building
			seenFill[fid] = true
			f, ok := fills[fid]
			if !ok {
				t.Errorf("%s: rule %d gives %s, and MakerFills has no entry", id, h.Rule, h.Building)
			} else if f.Rule != h.Rule {
				t.Errorf("%s: the rules give %s by rule %d, MakerFills says rule %d", id, h.Building, h.Rule, f.Rule)
			}
		default:
			seenOpen[id] = true
			if _, ok := open[id]; !ok {
				t.Errorf("%s: the rules leave it open (%s), and MakerOpenHoles does not list it", id, h.Why)
			}
		}
	}
	for id := range fills {
		if !seenFill[id] {
			t.Errorf("MakerFills lists %s, which the rules do not give", id)
		}
	}
	for id := range open {
		if !seenOpen[id] {
			t.Errorf("MakerOpenHoles lists %s, which is not an open hole", id)
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

// TestMainframeHallIsTheMedianMaker: the Modern Age's maker of data is priced
// in steel and electricity only, at a first-copy price in units equal to the
// median of the age's other healthy makers (against the levels it joins), in the
// proportion of steel to electricity the two middle makers use (their mean),
// and makes data at the standard (a Rule 5 fill in MakerFills).
func TestMainframeHallIsTheMedianMaker(t *testing.T) {
	d := BuildingByKey()["mainframe_hall"]
	if d.RequiredAge != "modern_age" || d.RequiredTech != "" || d.Category != "production" || d.LineageKey != "hacker" || d.LineageTier != 0 {
		t.Fatalf("mainframe_hall = %+v: want a Modern Age production building with no tech, tier 0 of the hacker lineage", d)
	}
	if len(d.BaseCost) != 2 || d.BaseCost["steel"] <= 0 || d.BaseCost["electricity"] <= 0 {
		t.Fatalf("mainframe_hall costs %v, want steel and electricity only", d.BaseCost)
	}
	lv := PriceLevels("modern_age")
	type maker struct{ u, steel, power float64 }
	var others []maker
	best := 0.0
	outs := map[string]float64{}
	for _, o := range BaseBuildings() {
		if o.RequiredAge != "modern_age" || o.Category == "wonder" || o.Key == d.Key {
			continue
		}
		price, out := makerUnits(o, lv, nil)
		outs[o.Key] = out
		best = math.Max(best, out)
		_ = price
	}
	for _, o := range BaseBuildings() {
		if o.RequiredAge != "modern_age" || o.Category == "wonder" || o.Key == d.Key || outs[o.Key] < MakerHealthyShare*best || outs[o.Key] <= 0 {
			continue
		}
		price, _ := makerUnits(o, lv, nil)
		others = append(others, maker{price, o.BaseCost["steel"] / lv["steel"], o.BaseCost["electricity"] / lv["electricity"]})
	}
	sort.Slice(others, func(i, j int) bool { return others[i].u < others[j].u })
	n := len(others)
	if n%2 != 0 {
		t.Fatalf("%d other healthy makers: the test reads the mean of the two middle ones", n)
	}
	a, b := others[n/2-1], others[n/2]
	median := (a.u + b.u) / 2
	share := (a.steel/(a.steel+a.power) + b.steel/(b.steel+b.power)) / 2
	price, _ := makerUnits(d, lv, nil)
	if math.Abs(price/median-1) > 0.01 {
		t.Errorf("mainframe_hall costs %.4f price units, the other healthy makers' median is %.4f", price, median)
	}
	steel, power := d.BaseCost["steel"]/lv["steel"], d.BaseCost["electricity"]/lv["electricity"]
	if got := steel / (steel + power); math.Abs(got-share) > 0.01 {
		t.Errorf("mainframe_hall's steel is %.3f of its units, the two middle makers use %.3f", got, share)
	}
	placed := false
	for _, f := range MakerFills() {
		if f.Building == d.Key && f.Material == "data" && f.Rule == 5 {
			placed = true
		}
	}
	if !placed {
		t.Errorf("mainframe_hall is not a Rule 5 fill of the Modern Age's data")
	}
}

// TestAcceptedShortagesAreRead: each shortage the owner accepts is one: the
// carried copies need more than one target of the age to make what the gate
// and the wonder ask of it, nothing in the age makes it at the standard, and
// the reading written in MakerAcceptedShortages is the tables' (to 1%).
func TestAcceptedShortagesAreRead(t *testing.T) {
	defs, ages := BaseBuildings(), Ages()
	for _, a := range MakerAcceptedShortages() {
		got := MakerShortageTargets(defs, ages, a.Age, a.Material)
		t.Logf("%s %s: %.4f targets", a.Age, a.Material, got)
		if got <= 1 {
			t.Errorf("%s %s reads %.2f targets: the carried copies cover it, it is no shortage to accept", a.Age, a.Material, got)
		}
		if math.Abs(got/a.Targets-1) > 0.01 {
			t.Errorf("%s %s reads %.3f targets, MakerAcceptedShortages says %.3f", a.Age, a.Material, got, a.Targets)
		}
		for _, d := range defs {
			if d.RequiredAge != a.Age || d.Category == "wonder" {
				continue
			}
			for _, e := range d.Effects {
				if e.Type == "production" && e.Target == a.Material && e.Value > 0 {
					t.Errorf("%s makes %s in %s, an accepted shortage: nothing should", d.Key, a.Material, a.Age)
				}
			}
		}
	}
}
