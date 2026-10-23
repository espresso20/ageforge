package config

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/vibe"
)

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

// TestLexiconReadsTheVibeScript: with VibeMode on, a word the script has a
// line for reads as that line, per form and per age; a word it lacks reads as
// ever; with it off, the plain word comes back. A plain English fallback
// (an age with no entry of its own) is its own row.
func TestLexiconReadsTheVibeScript(t *testing.T) {
	defer vibe.UseScriptForTest(map[string]string{
		"lexicon.bronze_age.scholars.one":  "Big Brain",
		"lexicon.bronze_age.scholars.many": "Big Brains",
		"lexicon.stone_age.workers.one":    "Grinder",
		"lexicon.plain.army_rank_0.one":    "Fresh Spawn",
	})()
	defer vibe.Set(false)
	plain := Term("bronze_age", TermScholars)
	vibe.Set(false)
	if got := Term("bronze_age", TermScholars); got != plain || plain.One != "Scribe" {
		t.Fatalf("with VibeMode off the bronze scholar is %+v", got)
	}
	vibe.Set(true)
	if got := Term("bronze_age", TermScholars); got != (Word{"Big Brain", "Big Brains"}) {
		t.Errorf("the bronze scholar with VibeMode on is %+v", got)
	}
	// One form only: the other stays plain.
	if got, was := Term("stone_age", TermWorkers), (Word{"Laborer", "Laborers"}); got.One != "Grinder" || got.Many != was.Many {
		t.Errorf("the stone worker with only a singular in the script is %+v", got)
	}
	// An age the script does not name keeps the plain word.
	if got := Term("iron_age", TermScholars); got.One != "Sage" {
		t.Errorf("the iron scholar is %+v", got)
	}
	// The plain English words are rows of their own, for every age.
	if got := Term("bronze_age", "army_rank_0"); got.One != "Fresh Spawn" || got.Many != "Green" {
		t.Errorf("army_rank_0 in the bronze age is %+v", got)
	}
	// A key nobody knows is its own word, script or not.
	if got := Term("bronze_age", "nonesuch"); got != (Word{"nonesuch", "nonesuch"}) {
		t.Errorf("an unknown key reads %+v", got)
	}
	vibe.Set(false)
	if got := Term("bronze_age", TermScholars); got != plain {
		t.Errorf("turned off again, the bronze scholar is %+v", got)
	}
}

// TestLexiconCellsAreEveryWord: the rows the script may name, each once,
// with the words as the lexicon has them.
func TestLexiconCellsAreEveryWord(t *testing.T) {
	cells := LexiconCells()
	seen := map[string]bool{}
	for _, c := range cells {
		if seen[c.ID] {
			t.Errorf("%s is listed twice", c.ID)
		}
		seen[c.ID] = true
		if c.Text == "" {
			t.Errorf("%s has no text", c.ID)
		}
	}
	// 22 ages of 4 people terms, the plain English words, both forms.
	if want := 2 * (22*4 + len(plainTerms)); len(cells) != want {
		t.Errorf("%d cells, want %d", len(cells), want)
	}
	for _, id := range []string{"lexicon.bronze_age.scholars.one", "lexicon.bronze_age.scholars.many", "lexicon.plain.army_rank_4.one"} {
		if !seen[id] {
			t.Errorf("%s is not a row", id)
		}
	}
	// Every cell is what Term reads with VibeMode off.
	vibe.Set(false)
	for _, c := range cells {
		parts := strings.Split(c.ID, ".")
		scope, key, form := parts[1], parts[2], parts[3]
		age := scope
		if scope == "plain" {
			age = "no_such_age"
		}
		w := Term(age, key)
		if got := map[string]string{"one": w.One, "many": w.Many}[form]; got != c.Text {
			t.Errorf("%s is %q and Term reads %q", c.ID, c.Text, got)
		}
	}
}
