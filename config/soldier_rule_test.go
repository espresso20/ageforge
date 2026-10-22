package config

import (
	"math"
	"testing"
)

// refHousingOracle works out the reference town's housing at the end of each
// age by hand, from the tables, in whole numbers: every non-wonder,
// non-monument building of the age and before at its reference count (the
// gate's count where it asks for more than TownCopies; a knowledge building
// the gate's alone; storage StorageRuleCopies), times its housing.
func refHousingOracle(t *testing.T) map[string]int64 {
	t.Helper()
	ages := Ages()
	asked := map[string]map[string]int{} // age -> building -> copies, the gate out of it
	for i := 0; i+1 < len(ages); i++ {
		asked[ages[i].Key] = ages[i+1].BuildingReqs
	}
	out := map[string]int64{}
	running := int64(0)
	for _, a := range ages {
		for _, d := range BaseBuildings() {
			if d.RequiredAge != a.Key || d.Category == "wonder" || d.Category == "monument" {
				continue
			}
			copies := TownCopies
			switch {
			case d.Category == "storage":
				copies = StorageRuleCopies
			}
			if g := asked[a.Key][d.Key]; g > copies || IsKnowledgeBuilding(d) {
				copies = g
			}
			for _, e := range d.Effects {
				if e.Type == "capacity" && e.Target == "population" {
					if e.Value != math.Trunc(e.Value) {
						t.Fatalf("%s houses %v: not a whole number", d.Key, e.Value)
					}
					running += int64(copies) * int64(e.Value)
				}
			}
		}
		out[a.Key] = running
	}
	return out
}

// TestSoldierPostsFollowTheRule: every military building comes with
// ceil(14% of the reference town's housing at the end of its age / the copies
// of the age's military buildings the reference town holds) soldiers a copy,
// worked here in whole numbers.
func TestSoldierPostsFollowTheRule(t *testing.T) {
	housing := refHousingOracle(t)
	ages := Ages()
	next := map[string]AgeDef{}
	for i := 0; i+1 < len(ages); i++ {
		next[ages[i].Key] = ages[i+1]
	}
	military := 0
	for _, d := range BaseBuildings() {
		if d.LineageKey != "military" {
			continue
		}
		military++
		// The copies of the age's military buildings the reference town
		// holds: eight of each, or the gate's count where it asks for more.
		gate := 0
		for _, m := range BaseBuildings() {
			if m.LineageKey == "military" && m.RequiredAge == d.RequiredAge {
				gate += max(TownCopies, next[d.RequiredAge].BuildingReqs[m.Key])
			}
		}
		h := housing[d.RequiredAge]
		want := int((14*h + int64(100*gate) - 1) / int64(100*gate))
		if d.Soldiers != want {
			t.Errorf("%s (%s): %d soldiers a copy, want ceil(0.14 x %d / %d copies) = %d", d.Key, d.RequiredAge, d.Soldiers, h, gate, want)
		}
		if d.Soldiers < 1 {
			t.Errorf("%s comes with no soldiers", d.Key)
		}
	}
	if military != 22 {
		t.Errorf("%d military buildings, want 22", military)
	}
}

// TestSoldierPostsOfAnAgeAreEqual: the military buildings of one age share the
// age's number (Hunting Lodge and Legion Fort).
func TestSoldierPostsOfAnAgeAreEqual(t *testing.T) {
	byAge := map[string]int{}
	for _, d := range BaseBuildings() {
		if d.LineageKey != "military" {
			continue
		}
		if n, ok := byAge[d.RequiredAge]; ok && n != d.Soldiers {
			t.Errorf("%s: %d soldiers a copy, another military building of %s has %d", d.Key, d.Soldiers, d.RequiredAge, n)
		}
		byAge[d.RequiredAge] = d.Soldiers
	}
}

// TestSoldierPostsRoundUpOnlyWhenTheyMust: a share that comes out whole is not
// rounded up by the float form of 0.14.
func TestSoldierPostsRoundUpOnlyWhenTheyMust(t *testing.T) {
	for _, c := range []struct {
		housing float64
		gate    int
		want    int
	}{
		{100, 1, 14}, {700, 2, 49}, {701, 2, 50}, {1, 8, 1}, {1000, 0, 140}, {50, 7, 1},
	} {
		if got := SoldierPostsFor(c.housing, c.gate); got != c.want {
			t.Errorf("housing %v, gate %d: %d posts, want %d", c.housing, c.gate, got, c.want)
		}
	}
}

// TestGatesAskForMilitaryBuildingsAsBefore: the military buildings made
// something (soldiers) and the gate rule asked for them for that. They make
// nothing now, and the rule asks for them all the same: the derived gate list
// is the one recorded when they did.
func TestGatesAskForMilitaryBuildingsAsBefore(t *testing.T) {
	want := map[string]map[string]int{
		"bronze_age": {"war_camp": 8}, "classical_age": {"hunting_lodge": 8}, "medieval_age": {"military_academy": 8},
		"renaissance_age": {"castle_keep": 8}, "colonial_age": {"fortress": 8}, "industrial_age": {"fort": 8},
		"victorian_age": {"military_base": 8}, "electric_age": {"garrison": 8}, "atomic_age": {"command_post": 8},
		"modern_age": {"bunker_complex": 8}, "information_age": {"special_ops_hq": 8},
		"cyberpunk_age": {"drone_warfare_center": 8}, "fusion_age": {"combat_aug_center": 8}, "space_age": {"plasma_command": 8},
		"interstellar_age": {"space_force_base": 8}, "galactic_age": {"fleet_command": 8}, "quantum_age": {"stellar_armada_hq": 8},
		"transcendent_age": {"probability_war_room": 8},
	}
	defs := BaseBuildings()
	mil := map[string]bool{}
	for _, d := range defs {
		if d.LineageKey == "military" {
			mil[d.Key] = true
		}
	}
	got := map[string]map[string]int{}
	total := 0
	for age, asks := range GateTown(defs, Ages(), Technologies()) {
		total += len(asks)
		for key, n := range asks {
			if mil[key] {
				if got[age] == nil {
					got[age] = map[string]int{}
				}
				got[age][key] = n
			}
		}
	}
	for age, w := range want {
		if len(got[age]) != len(w) {
			t.Errorf("%s: the gate rule asks for %v, want %v", age, got[age], w)
		}
		for key, n := range w {
			if got[age][key] != n {
				t.Errorf("%s: the gate rule asks for %d %s, want %d", age, got[age][key], key, n)
			}
		}
	}
	if len(got) != 18 {
		t.Errorf("the gate rule asks for military buildings in %d gates, want 18", len(got))
	}
	// The whole derived list, military or not, has the size it had.
	if total != 201 {
		t.Errorf("the derived gate list has %d entries, want the 201 it had", total)
	}
}

// TestRefSoldiersAreTheReferenceTownsArmy: the reference town has eight War
// Camps in the Stone Age, and eight Barracks more in the Bronze Age.
func TestRefSoldiersAreTheReferenceTownsArmy(t *testing.T) {
	army := RefSoldiers(BaseBuildings(), Ages())
	idx := map[string]int{}
	for i, a := range Ages() {
		idx[a.Key] = i
	}
	byKey := BuildingByKey()
	stone := float64(8 * byKey["war_camp"].Soldiers)
	if got := army[idx["stone_age"]]; got != stone {
		t.Errorf("Stone Age army %v, want %v", got, stone)
	}
	if got, want := army[idx["bronze_age"]], stone+float64(8*byKey["barracks"].Soldiers); got != want {
		t.Errorf("Bronze Age army %v, want %v", got, want)
	}
}

// The parties by age, pinned: short, long, grand military and grand civil for
// each of the 22 ages, worked out by hand from the reference army (the
// soldiers a reference town has at its posts) at 4%, 10%, 30% and 12%, rounded
// up, at least one. Bronze comes to 8, 20, 60 and 24, the sizes the lab
// settled on.
func TestPartySizesByAge(t *testing.T) {
	army := []int{0, 72, 200, 430, 790, 1406, 2566, 4798, 15094, 29694, 52894, 93304, 262728, 500968, 876832, 2112856, 3899386, 7750298, 19858746, 48482266, 85913821, 140961453}
	want := [][4]int{
		{0, 0, 0, 0}, {3, 8, 22, 9}, {8, 20, 60, 24}, {18, 43, 129, 52}, {32, 79, 237, 95}, {57, 141, 422, 169},
		{103, 257, 770, 308}, {192, 480, 1440, 576}, {604, 1510, 4529, 1812}, {1188, 2970, 8909, 3564},
		{2116, 5290, 15869, 6348}, {3733, 9331, 27992, 11197}, {10510, 26273, 78819, 31528},
		{20039, 50097, 150291, 60117}, {35074, 87684, 263050, 105220}, {84515, 211286, 633857, 253543},
		{155976, 389939, 1169816, 467927}, {310012, 775030, 2325090, 930036}, {794350, 1985875, 5957624, 2383050},
		{1939291, 4848227, 14544680, 5817872}, {3436553, 8591383, 25774147, 10309659}, {5638459, 14096146, 42288436, 16915375},
	}
	ref := RefSoldiers(BaseBuildings(), Ages())
	if len(ref) != len(army) || len(want) != len(army) {
		t.Fatalf("%d ages, the table has %d armies and %d parties", len(ref), len(army), len(want))
	}
	for i := range army {
		if int(ref[i]) != army[i] {
			t.Errorf("age %d: reference army %v, the table says %d", i, ref[i], army[i])
		}
		got := [4]int{PartySize(PartyShort, i), PartySize(PartyLong, i), PartySize(PartyMilitary, i), PartySize(PartyCivil, i)}
		if got != want[i] {
			t.Errorf("age %d (%s): parties %v, want %v", i, Ages()[i].Key, got, want[i])
		}
	}
	if PartyShares[PartyShort] != 4 || PartyShares[PartyLong] != 10 || PartyShares[PartyMilitary] != 30 || PartyShares[PartyCivil] != 12 || len(PartyShares) != 4 {
		t.Errorf("shares %v, want 4, 10, 30 and 12 percent", PartyShares)
	}
}
