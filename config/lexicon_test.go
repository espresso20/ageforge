package config

import (
	"strings"
	"testing"
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
