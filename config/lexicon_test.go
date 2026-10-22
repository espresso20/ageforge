package config

import "testing"

// peopleNames is the owner's table of the four people terms, by age (plural
// forms), written out here so a typo in the lexicon fails.
var peopleNames = map[string][4]string{
	"primitive_age":    {"Gatherers", "Clanfolk", "Hunters", "Elders"},
	"stone_age":        {"Laborers", "Villagers", "Warriors", "Storytellers"},
	"bronze_age":       {"Craftsfolk", "Villagers", "Spearmen", "Scribes"},
	"iron_age":         {"Laborers", "Townsfolk", "Swordsmen", "Sages"},
	"classical_age":    {"Freemen", "Citizens", "Hoplites", "Philosophers"},
	"medieval_age":     {"Serfs", "Commoners", "Men-at-arms", "Monks"},
	"renaissance_age":  {"Guildsmen", "Burghers", "Musketeers", "Scholars"},
	"colonial_age":     {"Settlers", "Colonists", "Regulars", "Naturalists"},
	"industrial_age":   {"Mill hands", "Townsfolk", "Riflemen", "Engineers"},
	"victorian_age":    {"Laborers", "Citizens", "Infantry", "Scientists"},
	"electric_age":     {"Linemen", "Citizens", "Infantry", "Inventors"},
	"atomic_age":       {"Workers", "Civilians", "Troopers", "Physicists"},
	"modern_age":       {"Employees", "Civilians", "Soldiers", "Researchers"},
	"information_age":  {"Staff", "Civilians", "Operators", "Analysts"},
	"digital_age":      {"Coders", "Netizens", "Cyber troops", "Data scientists"},
	"cyberpunk_age":    {"Wage slaves", "Street folk", "Mercs", "Netrunners"},
	"fusion_age":       {"Technicians", "Residents", "Augments", "Theorists"},
	"space_age":        {"Crew", "Colonists", "Marines", "Astronomers"},
	"interstellar_age": {"Spacers", "Settlers", "Rangers", "Xenologists"},
	"galactic_age":     {"Hands", "Citizens", "Legionnaires", "Archivists"},
	"quantum_age":      {"Operators", "Entangled", "Wardens", "Quantum minds"},
	"transcendent_age": {"Shapers", "Souls", "Paragons", "Oracles"},
}

// TestLexiconPeopleTermsFilled holds all 22 ages to the owner's table, with a
// singular for every plural.
func TestLexiconPeopleTermsFilled(t *testing.T) {
	order := AgeOrder()
	if len(order) != 22 || len(peopleNames) != 22 {
		t.Fatalf("expected 22 ages, got %d in the game and %d in the table", len(order), len(peopleNames))
	}
	for _, age := range order {
		want, ok := peopleNames[age]
		if !ok {
			t.Errorf("%s: no row in the test's table", age)
			continue
		}
		for i, key := range PeopleTerms() {
			w, has := TermsFor(age)[key]
			if !has {
				t.Errorf("%s: the lexicon has no %s", age, key)
				continue
			}
			if w.Many != want[i] {
				t.Errorf("%s %s: plural %q, want %q", age, key, w.Many, want[i])
			}
			if w.One == "" {
				t.Errorf("%s %s: no singular", age, key)
			}
		}
	}
	if len(lexiconRows) != 22 {
		t.Errorf("the lexicon has %d rows, want one per age (22)", len(lexiconRows))
	}
}

// TestLexiconFallback: a term with no entry gives its plain word, an unknown
// age gives the plain people words, and an unknown key gives the key.
func TestLexiconFallback(t *testing.T) {
	if got := Term("bronze_age", TermScholars); got.One != "Scribe" || got.Many != "Scribes" {
		t.Errorf("bronze scholars = %+v", got)
	}
	if got := Term("no_such_age", TermScholars); got.One != "scholar" || got.Many != "scholars" {
		t.Errorf("an unknown age should give the plain word, got %+v", got)
	}
	if got := Term("bronze_age", "pigs"); got.One != "pigs" || got.Many != "pigs" {
		t.Errorf("an unknown term should give the key, got %+v", got)
	}
	if got := Term("", TermWorkers); got.One != "worker" || got.Many != "workers" {
		t.Errorf("no age should give the plain worker, got %+v", got)
	}
}
