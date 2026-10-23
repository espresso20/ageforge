package main

import (
	"testing"

	"github.com/espresso20/ageforge/words"
)

// TestFieldIDsAreTheTablesOwn: the id of every field the export reads from a
// definition comes from words.FieldIDs, the table the ruleset overlay reads
// too. A rule that is missing from it would export no id, and an id with no
// rule would be a row the export never writes.
func TestFieldIDsAreTheTablesOwn(t *testing.T) {
	for full, rule := range fieldRules {
		if rule.id != "" {
			t.Errorf("%s carries its own id template %q: put it in words.FieldIDs", full, rule.id)
		}
		if words.FieldIDs[full] == "" {
			t.Errorf("%s has a rule but no id in words.FieldIDs", full)
		}
	}
	for full := range words.FieldIDs {
		if _, ok := fieldRules[full]; !ok {
			t.Errorf("words.FieldIDs names %s, which the export has no rule for", full)
		}
	}
}
