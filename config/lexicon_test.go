package config

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/vibe"
)

// TestLexiconPeopleTermsFilled holds every age to a full row: one lexicon row
// per age in age order, all four people terms present, each with a singular
// and a plural that are trimmed and capitalised, and no two terms of an age
// sharing a word.
func TestLexiconPeopleTermsFilled(t *testing.T) {
	order := AgeOrder()
	if len(lexiconRows) != len(order) {
		t.Fatalf("the lexicon has %d rows, want one per age (%d)", len(lexiconRows), len(order))
	}
	for i, age := range order {
		if lexiconRows[i].age != age {
			t.Errorf("lexicon row %d is %s, want %s (the order of ages)", i, lexiconRows[i].age, age)
		}
		seen := map[string]string{}
		for _, key := range PeopleTerms() {
			w, has := TermsFor(age)[key]
			if !has {
				t.Errorf("%s: the lexicon has no %s", age, key)
				continue
			}
			for form, word := range map[string]string{"singular": w.One, "plural": w.Many} {
				if word == "" || word != strings.TrimSpace(word) || word[0] < 'A' || word[0] > 'Z' {
					t.Errorf("%s %s: %s %q should be a trimmed, capitalised word", age, key, form, word)
				}
				if other, dup := seen[word]; dup && other != key {
					t.Errorf("%s: %s and %s share the word %q", age, other, key, word)
				}
				seen[word] = key
			}
		}
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
	// Every age's people terms, the plain English words, both forms.
	if want := 2 * (len(AgeOrder())*len(PeopleTerms()) + len(plainTerms)); len(cells) != want {
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

// TermsFor is every term the lexicon holds for age, by key; a term the age
// has no entry for is left out (Term answers it with the plain word).
func TermsFor(age string) map[string]Word {
	out := map[string]Word{}
	for _, row := range lexiconRows {
		if row.age == age {
			for k, w := range row.terms {
				out[k] = w
			}
		}
	}
	return out
}
