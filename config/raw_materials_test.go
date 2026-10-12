package config

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"
)

// The four raw materials (iron ore, marble, titanium ore, dark matter
// crystals) are ordinary materials in the ages that make them. In such an age
// a raw is paired with a worked form (RawPartner: iron, stone, titanium or
// steel, dark matter or electricity) and the pair share one price level. Every
// building of the age, and its wonder, that was priced in the worked form pays
// half of that part in the raw instead, one for one in amount; housing and
// storage do not, and neither do the raw's own makers of that age.
//
// testdata/raw_master_prices.json records the first-copy price of each building
// the rule names as it was before any raw was priced.

// rawMasterPrices are the prices of the buildings the rule names, before it.
func rawMasterPrices(t *testing.T) map[string]map[string]float64 {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "raw_master_prices.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]map[string]float64
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// rawSigFigs is the number of significant figures an amount is written in.
func rawSigFigs(v float64) int {
	for n := 1; n <= 17; n++ {
		p, _ := strconv.ParseFloat(strconv.FormatFloat(v, 'e', n-1, 64), 64)
		if p == v {
			return n
		}
	}
	return 17
}

// rawHalf is half of v, rounded down to the table's usual figures (three, or
// as many as v is written in): the raw's share. The worked form keeps the rest,
// so the two add up to v exactly.
func rawHalf(v float64) float64 {
	sf := max(3, rawSigFigs(v))
	mag := math.Pow(10, math.Floor(math.Log10(v/2))-float64(sf-1))
	return math.Floor(v/2/mag+1e-9) * mag
}

// TestRawMaterialsCostHalfTheirPartner proves the rule: every building it
// names carries its raw at half its partner's amount, nothing else is priced
// in a raw, and every price in units, and every price level, is what it was.
func TestRawMaterialsCostHalfTheirPartner(t *testing.T) {
	defs := BaseBuildings()
	master := rawMasterPrices(t)
	masterDefs := make([]BuildingDef, len(defs))
	copy(masterDefs, defs)
	for i, d := range masterDefs {
		if m, ok := master[d.Key]; ok {
			masterDefs[i].BaseCost = m
		}
	}
	isRaw := map[string]bool{}
	for _, r := range RawMaterials {
		isRaw[r] = true
	}
	makes := func(d BuildingDef, res string) bool {
		for _, e := range d.Effects {
			if e.Type == "production" && e.Value > 0 && e.Target == res {
				return true
			}
		}
		return false
	}
	// Who pays, from the master prices.
	pays := map[string]bool{}   // "building/raw"
	halved := map[string]bool{} // "building/partner": the part that was split
	var ages []string
	for _, a := range Ages() {
		ages = append(ages, a.Key)
	}
	for _, raw := range RawMaterials {
		for _, age := range ages {
			partner := RawPartner(masterDefs, age, raw)
			if partner == "" {
				continue
			}
			for _, d := range masterDefs {
				if _, left := RawLeftOut()[d.Key]; left || d.RequiredAge != age || d.Category == "housing" || d.Category == "storage" || makes(d, raw) || d.BaseCost[partner] <= 0 {
					continue
				}
				pays[d.Key+"/"+raw] = true
				halved[d.Key+"/"+partner] = true
				old := d.BaseCost[partner]
				live := BuildingByKey()[d.Key]
				got := live.BaseCost[raw]
				want := rawHalf(old)
				if got != want {
					t.Errorf("%s (%s) costs %v %s, half of its %v %s is %v", d.Key, age, got, raw, old, partner, want)
				}
				if math.Abs(got-old/2) > 0.001*old {
					t.Errorf("%s: %s %v is not half of %v", d.Key, raw, got, old)
				}
				if live.BaseCost[partner]+got != old {
					t.Errorf("%s: %v %s and %v %s add up to %v, not %v", d.Key, live.BaseCost[partner], partner, got, raw, live.BaseCost[partner]+got, old)
				}
			}
		}
	}
	if len(pays) == 0 {
		t.Fatal("the rule names no building")
	}
	// Nothing outside the rule is priced in a raw; the rest of a price is as it was.
	for _, d := range defs {
		for res := range d.BaseCost {
			if isRaw[res] && !pays[d.Key+"/"+res] {
				t.Errorf("%s is priced in %s and the rule does not name it", d.Key, res)
			}
		}
		if m, ok := master[d.Key]; ok {
			for res, v := range m {
				if halved[d.Key+"/"+res] {
					continue // split in two above
				}
				if d.BaseCost[res] != v {
					t.Errorf("%s: %s is %v, it was %v", d.Key, res, d.BaseCost[res], v)
				}
			}
		}
	}
	// Levels and units are what they were.
	live, was := priceLevels(defs), priceLevels(masterDefs)
	for age, m := range was {
		for res, v := range m {
			if live[age][res] != v {
				t.Errorf("%s %s level is %v, it was %v", age, res, live[age][res], v)
			}
		}
		for _, raw := range RawMaterials {
			if p := RawPartner(defs, age, raw); p != "" && live[age][raw] != live[age][p] {
				t.Errorf("%s: %s level %v is not its partner %s level %v", age, raw, live[age][raw], p, live[age][p])
			}
		}
	}
	for i, d := range defs {
		a := priceUnits(d.BaseCost, live[d.RequiredAge])
		b := priceUnits(masterDefs[i].BaseCost, was[d.RequiredAge])
		if b > 0 && math.Abs(a/b-1) > 1e-9 {
			t.Errorf("%s costs %.9g price units, it cost %.9g", d.Key, a, b)
		}
	}
	var names []string
	for k := range pays {
		names = append(names, k)
	}
	sort.Strings(names)
	t.Logf("%d (building, raw) prices carry a raw", len(names))
}

// TestRawsHaveNoLevelWithoutAMaker: a raw is priced only in an age with a
// maker of it, and has a level only there.
func TestRawsHaveNoLevelWithoutAMaker(t *testing.T) {
	defs := BaseBuildings()
	lv := priceLevels(defs)
	for _, raw := range RawMaterials {
		for _, a := range Ages() {
			made := ageMakesRaw(defs, a.Key, raw)
			if !made && lv[a.Key][raw] != 0 {
				t.Errorf("%s has a level for %s and no maker of it", a.Key, raw)
			}
			if made && lv[a.Key][raw] <= 0 {
				t.Errorf("%s makes %s and has no level for it", a.Key, raw)
			}
		}
	}
}

// TestRawLeftOutIsALoop: every building the raw rule leaves out is the only
// source, in its age, of a material the age's maker of a raw is priced in:
// the maker needs what the building makes, so the building cannot need the raw.
// The sources are the age's own buildings that a run can build and what every
// town carries into it (the gates' required copies and the earlier wonders).
func TestRawLeftOutIsALoop(t *testing.T) {
	defs, ages, techs := BaseBuildings(), Ages(), Technologies()
	held := GateTownHeldTechs(defs, ages, techs)
	byKey := BuildingByKey()
	makes := func(d BuildingDef, res string) bool {
		for _, e := range d.Effects {
			if e.Type == "production" && e.Value > 0 && e.Target == res {
				return true
			}
		}
		return false
	}
	for key, why := range RawLeftOut() {
		d, ok := byKey[key]
		if !ok || why == "" {
			t.Errorf("%s is left out of the raw rule and is not a building with a reason", key)
			continue
		}
		pos := AgePositions(ageKeys(ages))[d.RequiredAge]
		carried := makerCarried(defs, ages, pos)
		loop := false
		for _, mine := range defs {
			if mine.RequiredAge != d.RequiredAge || mine.Key == key {
				continue
			}
			rawMaker := false
			for _, raw := range RawMaterials {
				rawMaker = rawMaker || makes(mine, raw)
			}
			if !rawMaker {
				continue
			}
			for _, e := range d.Effects {
				mat := e.Target
				if e.Type != "production" || e.Value <= 0 || mine.BaseCost[mat] <= 0 || makerSkipped(mat) {
					continue
				}
				other := false
				for _, o := range defs {
					if o.Key == key || !makes(o, mat) || (o.Category == "wonder" && o.RequiredAge == d.RequiredAge) {
						continue // the age's own wonder is paid last: it starts nothing
					}
					if (o.RequiredAge == d.RequiredAge && (o.RequiredTech == "" || held[pos][o.RequiredTech])) || carried[o.Key] > 0 {
						other = true
					}
				}
				if !other {
					loop = true
				}
			}
		}
		if !loop {
			t.Errorf("%s is left out of the raw rule (%s) and no raw maker of its age is priced in something only it makes", key, why)
		}
		for _, raw := range RawMaterials {
			if d.BaseCost[raw] != 0 {
				t.Errorf("%s is left out and costs %s", key, raw)
			}
		}
	}
}
